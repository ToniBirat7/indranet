package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	stripe "github.com/stripe/stripe-go/v76"
	stripecs "github.com/stripe/stripe-go/v76/checkout/session"

	"github.com/ToniBirat7/indranet/packages/backend/internal/models"
)

// CreateSession creates a new session and (in production) a Stripe Checkout session.
// In development (no STRIPE_SECRET_KEY), the session is auto-authorized for easy testing.
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

	resp := map[string]interface{}{
		"session_id":            sessionID,
		"state":                 "CREATED",
		"rate_per_minute_cents": ratePerMinuteCents,
		"pre_auth_minutes":      req.DurationMinutes,
	}

	if h.deps.Config.StripeSecretKey == "" {
		// Dev mode: auto-authorize so the full flow is testable without Stripe
		if _, err := h.deps.Pool.Exec(r.Context(), `
			UPDATE sessions SET state = 'AUTHORIZED', updated_at = NOW() WHERE id = $1
		`, sessionID); err != nil {
			slog.Error("dev auto-authorize failed", "session_id", sessionID, "error", err)
		} else {
			resp["state"] = "AUTHORIZED"
			h.deps.Hub.SendToSession(sessionID, map[string]string{
				"type":       "session_authorized",
				"session_id": sessionID,
			})
			go h.awaitAgentReady(sessionID, 5*time.Minute)
		}
	} else {
		checkoutURL, err := h.createStripeCheckout(sessionID, req.DurationMinutes, ratePerMinuteCents)
		if err != nil {
			slog.Error("stripe: checkout creation failed", "session_id", sessionID, "error", err)
			// Don't block session creation — return without checkout_url; user can retry payment
		} else {
			resp["checkout_url"] = checkoutURL
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) createStripeCheckout(sessionID string, durationMinutes int, ratePerMinuteCents int64) (string, error) {
	stripe.Key = h.deps.Config.StripeSecretKey

	totalCents := ratePerMinuteCents * int64(durationMinutes)
	productName := fmt.Sprintf("IndraNet Session — %d minutes", durationMinutes)

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(productName),
					},
					UnitAmount: stripe.Int64(totalCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		// Metadata on both the checkout session AND the payment intent so that
		// payment_intent.payment_failed webhook also carries indranet_session_id.
		Metadata: map[string]string{"indranet_session_id": sessionID},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{"indranet_session_id": sessionID},
		},
		SuccessURL: stripe.String(h.deps.Config.FrontendBaseURL + "/sessions/" + sessionID + "?payment=success"),
		CancelURL:  stripe.String(h.deps.Config.FrontendBaseURL + "/sessions/" + sessionID + "?payment=cancelled"),
	}

	cs, err := stripecs.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe checkout session: %w", err)
	}
	return cs.URL, nil
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

// ListSessions returns the authenticated user's session history (newest first).
func (h *Handlers) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxKeyUserID).(string)

	rows, err := h.deps.Pool.Query(r.Context(), `
		SELECT id, host_id, state, rate_per_minute_cents, total_charged_cents,
		       started_at, created_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 20
	`, userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type sessionSummary struct {
		ID                 string             `json:"session_id"`
		HostID             string             `json:"host_id"`
		State              models.SessionState `json:"state"`
		RatePerMinuteCents int64              `json:"rate_per_minute_cents"`
		TotalChargedCents  int64              `json:"total_charged_cents"`
		StartedAt          interface{}        `json:"started_at"`
		CreatedAt          interface{}        `json:"created_at"`
	}

	var sessions []sessionSummary
	for rows.Next() {
		var s sessionSummary
		if err := rows.Scan(
			&s.ID, &s.HostID, &s.State, &s.RatePerMinuteCents,
			&s.TotalChargedCents, &s.StartedAt, &s.CreatedAt,
		); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		sessions = append(sessions, s)
	}
	if sessions == nil {
		sessions = []sessionSummary{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"sessions": sessions})
}
