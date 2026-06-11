package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"github.com/ToniBirat7/indranet/packages/backend/internal/models"
)

type contextKey string

const (
	ctxKeyUserID contextKey = "user_id"
	ctxKeyRole   contextKey = "role"
)

type jwtClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Register creates a new user account.
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "email, password, and name are required", http.StatusBadRequest)
		return
	}
	if !strings.Contains(req.Email, "@") || len(req.Email) > 254 {
		http.Error(w, "invalid email address", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 80 {
		http.Error(w, "name must be ≤80 characters", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		http.Error(w, "password must be 8–72 characters", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var user models.User
	err = h.deps.Pool.QueryRow(r.Context(),
		`INSERT INTO users (email, password_hash, name)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, name, role, balance_cents, created_at`,
		req.Email, string(hash), req.Name,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.BalanceCents, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "email already registered", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, err := h.generateUserJWT(user.ID, user.Role)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"token":   token,
	})
}

// Login authenticates a user and returns a JWT.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	var user models.User
	err := h.deps.Pool.QueryRow(r.Context(),
		`SELECT id, email, password_hash, name, role FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.generateUserJWT(user.ID, user.Role)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.deps.Config.JWTExpiryHours) * time.Hour)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      token,
		"user_id":    user.ID,
		"expires_at": expiresAt,
	})
}

// AuthMiddleware validates the JWT bearer token on protected routes.
func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}
		claims := &jwtClaims{}
		_, err := jwt.ParseWithClaims(strings.TrimPrefix(authHeader, "Bearer "), claims, h.jwtKeyFunc)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxKeyRole, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AgentAuthMiddleware validates the agent JWT on agent-only routes.
func (h *Handlers) AgentAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}
		claims := &jwtClaims{}
		_, err := jwt.ParseWithClaims(strings.TrimPrefix(authHeader, "Bearer "), claims, h.jwtKeyFunc)
		if err != nil || claims.Role != "agent" {
			http.Error(w, "invalid agent token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxKeyRole, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handlers) generateUserJWT(userID, role string) (string, error) {
	expiry := time.Duration(h.deps.Config.JWTExpiryHours) * time.Hour
	claims := &jwtClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.deps.Config.JWTSecret))
}

func (h *Handlers) generateAgentJWT(hostID string) (string, error) {
	claims := &jwtClaims{
		UserID: hostID,
		Role:   "agent",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   hostID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.deps.Config.JWTSecret))
}

func (h *Handlers) jwtKeyFunc(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}
	return []byte(h.deps.Config.JWTSecret), nil
}

// GetMe returns the authenticated user's profile, balance, and host ID (if registered as host).
func (h *Handlers) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxKeyUserID).(string)

	var id, email, name, role string
	var balanceCents int64
	var hostID *string
	err := h.deps.Pool.QueryRow(r.Context(), `
		SELECT u.id, u.email, u.name, u.role, u.balance_cents,
		       (SELECT id FROM hosts WHERE user_id = u.id LIMIT 1)
		FROM users u WHERE u.id = $1
	`, userID).Scan(&id, &email, &name, &role, &balanceCents, &hostID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"user_id":       id,
		"email":         email,
		"name":          name,
		"role":          role,
		"balance_cents": balanceCents,
	}
	if hostID != nil {
		resp["host_id"] = *hostID
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
