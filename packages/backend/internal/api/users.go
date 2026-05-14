package api

import (
	"encoding/json"
	"net/http"
)

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
	if len(req.Password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// TODO: Hash password with bcrypt
	// TODO: INSERT INTO users (email, password_hash, name) VALUES (...)
	// TODO: Generate JWT
	// TODO: Handle duplicate email (UNIQUE constraint violation)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": "usr_TODO",
		"token":   "TODO_JWT",
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

	// TODO: SELECT user by email
	// TODO: bcrypt.CompareHashAndPassword
	// TODO: Generate JWT

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":      "TODO_JWT",
		"expires_at": "2025-06-01T00:00:00Z",
	})
}

// AuthMiddleware validates the JWT bearer token on protected routes.
func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Extract Authorization: Bearer <jwt> header
		// TODO: Parse and validate JWT with h.deps.Config.JWTSecret
		// TODO: Add user claims to request context
		// TODO: Return 401 if token is missing or invalid
		next.ServeHTTP(w, r)
	})
}

// AgentAuthMiddleware validates the agent JWT on agent-only routes.
func (h *Handlers) AgentAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Extract and validate agent JWT (has "agent" role claim)
		// TODO: Verify the agent_token_hash in the hosts table matches
		next.ServeHTTP(w, r)
	})
}
