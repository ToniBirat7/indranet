package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// CreateSession creates a new session and initiates Stripe payment.
// Session state machine: CREATED → AUTHORIZED → ACTIVE → ENDING → ENDED
//
// Flow:
//  1. Validate request (host exists, user has sufficient balance or allow Stripe checkout)
//  2. Lock the pre-auth amount in user balance
//  3. Create session record (state=CREATED)
//  4. Create Stripe Checkout session
//  5. Return session_id + checkout_url to client
func (h *Handlers) CreateSession(w http.ResponseWriter, r *http.Request) {
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
		req.DurationMinutes = 15 // Minimum session
	}

	// TODO: Fetch host from DB, verify it's online
	// TODO: Calculate total cost: host.PricePerMinuteCents * req.DurationMinutes
	// TODO: Create Stripe Checkout session
	// TODO: Insert session record (state=CREATED)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"session_id":   "ses_TODO",
		"state":        "CREATED",
		"checkout_url": "https://checkout.stripe.com/pay/TODO",
	})
}

// StartSession transitions a session from AUTHORIZED to ACTIVE.
// Called by the host agent after the sandbox is ready and streaming.
// Requires agent JWT authentication (not user JWT).
func (h *Handlers) StartSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	// TODO: Verify session exists and is in AUTHORIZED state
	// TODO: UPDATE sessions SET state='ACTIVE', started_at=NOW() WHERE id=$1 AND state='AUTHORIZED'
	// TODO: Notify signaling hub to send session_ready to viewer

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"state": "ACTIVE"})
}

// EndSession transitions a session to ENDING state.
// Called by the user client when they want to disconnect.
// The host agent will confirm teardown via heartbeat response.
func (h *Handlers) EndSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	// TODO: Verify the requesting user owns this session
	// TODO: UPDATE sessions SET state='ENDING' WHERE id=$1 AND user_id=$2
	// TODO: Notify signaling hub to send session_kill to host

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"state": "ENDING"})
}

// GetSession returns the current status of a session.
func (h *Handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	// TODO: SELECT session + compute balance_remaining_minutes from user balance

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":                sessionID,
		"state":                     "TODO",
		"balance_remaining_minutes": 0,
	})
}

// HeartbeatSession handles the host agent's regular heartbeat.
// Returns action "continue" or "kill" based on session state.
// If the session is ENDING (balance exhausted or user disconnect), returns "kill".
func (h *Handlers) HeartbeatSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	// TODO: SELECT state FROM sessions WHERE id=$1
	// TODO: UPDATE sessions SET last_heartbeat_at=NOW() WHERE id=$1
	// If state is ENDING or FAILED, return action=kill

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"action": "continue"})
}
