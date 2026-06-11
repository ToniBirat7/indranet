package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ToniBirat7/indranet/packages/backend/internal/models"
)

// CreateSession creates a new session record.
// Payment gating (Stripe) is wired in packages/backend/internal/api/payments.go.
// For Phase 0: creates session in CREATED state without Stripe (direct auth flow).
func (h *Handlers) CreateSession(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxKeyUserID).(string)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		HostID          string `json:"host_id"`
		DurationMinutes int    `json:"duration_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.HostID == "" {
		http.Error(w, "host_id is required", http.StatusBadRequest)
		return
	}
	if req.DurationMinutes < 15 {
		req.DurationMinutes = 15
	}

	// Fetch host to get rate and verify it's online
	var pricePerHourCents int64
	var online bool
	err := h.deps.Pool.QueryRow(r.Context(),
		`SELECT price_per_hour_cents, online FROM hosts WHERE id = $1`,
		req.HostID,
	).Scan(&pricePerHourCents, &online)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "host not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if !online {
		http.Error(w, "host is not online", http.StatusConflict)
		return
	}

	ratePerMinuteCents := pricePerHourCents / 60

	var sessionID string
	err = h.deps.Pool.QueryRow(r.Context(), `
		INSERT INTO sessions (user_id, host_id, state, rate_per_minute_cents, pre_auth_minutes)
		VALUES ($1, $2, 'CREATED', $3, $4)
		RETURNING id
	`, userID, req.HostID, ratePerMinuteCents, req.DurationMinutes,
	).Scan(&sessionID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":           sessionID,
		"state":                "CREATED",
		"rate_per_minute_cents": ratePerMinuteCents,
		"pre_auth_minutes":     req.DurationMinutes,
	})
}

// StartSession transitions a session from AUTHORIZED to ACTIVE.
// Called by the host agent after the sandbox is ready and streaming has begun.
func (h *Handlers) StartSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	tag, err := h.deps.Pool.Exec(r.Context(), `
		UPDATE sessions
		SET state = 'ACTIVE', started_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND state = 'AUTHORIZED'
	`, sessionID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "session not found or not in AUTHORIZED state", http.StatusConflict)
		return
	}

	h.deps.Hub.SendToSession(sessionID, map[string]string{
		"type":  "session_state",
		"state": "ACTIVE",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"state": "ACTIVE"})
}

// EndSession transitions a session to ENDING state (user-initiated disconnect).
func (h *Handlers) EndSession(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxKeyUserID).(string)
	sessionID := chi.URLParam(r, "id")

	tag, err := h.deps.Pool.Exec(r.Context(), `
		UPDATE sessions
		SET state = 'ENDING', updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND state IN ('AUTHORIZED', 'ACTIVE')
	`, sessionID, userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "session not found or already ended", http.StatusConflict)
		return
	}

	h.deps.Hub.SendToSession(sessionID, map[string]string{
		"type":   "session_kill",
		"reason": "user_disconnect",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"state": "ENDING"})
}

// GetSession returns current session state and remaining balance.
func (h *Handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxKeyUserID).(string)
	sessionID := chi.URLParam(r, "id")

	var session models.Session
	err := h.deps.Pool.QueryRow(r.Context(), `
		SELECT id, user_id, host_id, state, rate_per_minute_cents,
		       pre_auth_minutes, total_charged_cents, started_at, created_at
		FROM sessions
		WHERE id = $1 AND user_id = $2
	`, sessionID, userID,
	).Scan(
		&session.ID, &session.UserID, &session.HostID, &session.State,
		&session.RatePerMinuteCents, &session.PreAuthMinutes,
		&session.TotalChargedCents, &session.StartedAt, &session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Compute remaining balance in minutes from user wallet
	var userBalanceCents int64
	_ = h.deps.Pool.QueryRow(r.Context(),
		`SELECT balance_cents FROM users WHERE id = $1`, userID,
	).Scan(&userBalanceCents)

	var remainingMinutes int64
	if session.RatePerMinuteCents > 0 {
		remainingMinutes = userBalanceCents / session.RatePerMinuteCents
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":                session.ID,
		"host_id":                   session.HostID,
		"state":                     session.State,
		"rate_per_minute_cents":     session.RatePerMinuteCents,
		"total_charged_cents":       session.TotalChargedCents,
		"balance_remaining_minutes": remainingMinutes,
		"started_at":                session.StartedAt,
		"created_at":                session.CreatedAt,
	})
}

// HeartbeatSession handles the host agent's 60-second heartbeat.
// Returns action: "continue" or "kill".
func (h *Handlers) HeartbeatSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var state models.SessionState
	err := h.deps.Pool.QueryRow(r.Context(),
		`SELECT state FROM sessions WHERE id = $1`, sessionID,
	).Scan(&state)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	action := "continue"
	if state == models.SessionStateEnding || state == models.SessionStateEnded ||
		state == models.SessionStateFailed {
		action = "kill"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"action": action})
}
