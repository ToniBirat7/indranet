package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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
	if req.DurationMinutes > 480 {
		http.Error(w, "duration_minutes cannot exceed 480 (8 hours)", http.StatusBadRequest)
		return
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

	// Check the host doesn't already have an active or pending session.
	var activeCount int
	if err := h.deps.Pool.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM sessions
		WHERE host_id = $1 AND state IN ('AUTHORIZED', 'ACTIVE')
	`, req.HostID).Scan(&activeCount); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if activeCount > 0 {
		http.Error(w, "host already has an active session", http.StatusConflict)
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
		// Dev mode: auto-authorize and credit wallet so billing engine works end-to-end.
		devTotalCents := ratePerMinuteCents * int64(req.DurationMinutes)
		if _, err := h.deps.Pool.Exec(r.Context(), `
			WITH s AS (
				UPDATE sessions SET state = 'AUTHORIZED', updated_at = NOW()
				WHERE id = $1 RETURNING user_id
			)
			UPDATE users SET balance_cents = balance_cents + $2, updated_at = NOW()
			FROM s WHERE users.id = s.user_id
		`, sessionID, devTotalCents); err != nil {
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
		SuccessURL: stripe.String(h.deps.Config.FrontendBaseURL + "/session/" + sessionID + "?payment=success"),
		CancelURL:  stripe.String(h.deps.Config.FrontendBaseURL + "/session/" + sessionID + "?payment=cancelled"),
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
// Supported query params: page (default 1), limit (default 20, max 100).
func (h *Handlers) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxKeyUserID).(string)

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, _ := strconv.Atoi(l); n > 0 && n <= 100 {
			limit = n
		}
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if n, _ := strconv.Atoi(p); n > 1 {
			offset = (n - 1) * limit
		}
	}

	var total int
	if err := h.deps.Pool.QueryRow(r.Context(),
		`SELECT COUNT(*) FROM sessions WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	rows, err := h.deps.Pool.Query(r.Context(), `
		SELECT s.id, s.host_id, h.display_name, s.state, s.rate_per_minute_cents,
		       s.total_charged_cents, s.started_at, s.created_at
		FROM sessions s
		JOIN hosts h ON h.id = s.host_id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type sessionSummary struct {
		ID                 string              `json:"session_id"`
		HostID             string              `json:"host_id"`
		HostName           string              `json:"host_name"`
		State              models.SessionState `json:"state"`
		RatePerMinuteCents int64               `json:"rate_per_minute_cents"`
		TotalChargedCents  int64               `json:"total_charged_cents"`
		StartedAt          interface{}         `json:"started_at"`
		CreatedAt          interface{}         `json:"created_at"`
	}

	var sessions []sessionSummary
	for rows.Next() {
		var s sessionSummary
		if err := rows.Scan(
			&s.ID, &s.HostID, &s.HostName, &s.State, &s.RatePerMinuteCents,
			&s.TotalChargedCents, &s.StartedAt, &s.CreatedAt,
		); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		sessions = append(sessions, s)
	}
	if rows.Err() != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if sessions == nil {
		sessions = []sessionSummary{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"total":    total,
		"page":     offset/limit + 1,
		"limit":    limit,
	})
}

// GetPendingSessions returns AUTHORIZED sessions assigned to this host agent.
// Called by the host agent on startup and periodically to discover new sessions.
// Uses agent JWT auth — the host ID is extracted from the token claim.
func (h *Handlers) GetPendingSessions(w http.ResponseWriter, r *http.Request) {
	hostID, _ := r.Context().Value(ctxKeyUserID).(string)
	if hostID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.deps.Pool.Query(r.Context(), `
		SELECT id, user_id, rate_per_minute_cents, pre_auth_minutes, created_at
		FROM sessions
		WHERE host_id = $1 AND state = 'AUTHORIZED'
		ORDER BY created_at ASC
	`, hostID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type pendingSession struct {
		SessionID          string      `json:"session_id"`
		UserID             string      `json:"user_id"`
		RatePerMinuteCents int64       `json:"rate_per_minute_cents"`
		PreAuthMinutes     int         `json:"pre_auth_minutes"`
		CreatedAt          interface{} `json:"created_at"`
	}

	var pending []pendingSession
	for rows.Next() {
		var s pendingSession
		if err := rows.Scan(
			&s.SessionID, &s.UserID, &s.RatePerMinuteCents,
			&s.PreAuthMinutes, &s.CreatedAt,
		); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		pending = append(pending, s)
	}
	if pending == nil {
		pending = []pendingSession{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"sessions": pending})
}
