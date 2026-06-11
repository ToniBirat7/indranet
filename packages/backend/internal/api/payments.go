package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	stripe "github.com/stripe/stripe-go/v76"
	stripecs "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/ToniBirat7/indranet/packages/backend/internal/signaling"
)

// StripeWebhook handles incoming Stripe webhook events.
// CRITICAL: Stripe-Signature header must be verified before processing any event.
func (h *Handlers) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	const maxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "body read error", http.StatusBadRequest)
		return
	}

	// Verify webhook signature — NEVER skip this (see CLAUDE.md security invariants)
	event, err := webhook.ConstructEvent(
		payload,
		r.Header.Get("Stripe-Signature"),
		h.deps.Config.StripeWebhookSecret,
	)
	if err != nil {
		slog.Warn("stripe webhook: signature verification failed", "error", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	slog.Info("stripe webhook received", "type", event.Type, "id", event.ID)

	switch event.Type {
	case "checkout.session.completed":
		if err := h.handleCheckoutComplete(r.Context(), event); err != nil {
			slog.Error("stripe: checkout.session.completed handler failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

	case "payment_intent.payment_failed":
		if err := h.handlePaymentFailed(r.Context(), event); err != nil {
			slog.Error("stripe: payment_intent.payment_failed handler failed", "error", err)
		}

	case "account.updated":
		if err := h.handleAccountUpdated(r.Context(), event); err != nil {
			slog.Error("stripe: account.updated handler failed", "error", err)
		}

	default:
		slog.Debug("stripe webhook: unhandled event type", "type", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

// handleCheckoutComplete dispatches checkout.session.completed events by their type metadata.
func (h *Handlers) handleCheckoutComplete(ctx context.Context, event stripe.Event) error {
	var cs stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &cs); err != nil {
		return fmt.Errorf("unmarshal checkout session: %w", err)
	}

	switch cs.Metadata["type"] {
	case "wallet_topup":
		return h.handleWalletTopup(ctx, cs)
	default:
		return h.handleSessionCheckoutComplete(ctx, cs)
	}
}

// handleSessionCheckoutComplete authorizes a session and credits the wallet atomically.
func (h *Handlers) handleSessionCheckoutComplete(ctx context.Context, cs stripe.CheckoutSession) error {
	internalSessionID := cs.Metadata["indranet_session_id"]
	if internalSessionID == "" {
		slog.Warn("stripe: checkout completed with no indranet_session_id in metadata")
		return nil
	}

	// Authorize session and credit user wallet in one transaction so billing
	// engine never sees a zero balance between AUTHORIZED and ACTIVE.
	tx, err := h.deps.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var paymentIntentID string
	if cs.PaymentIntent != nil {
		paymentIntentID = cs.PaymentIntent.ID
	}

	var userID string
	var totalCents int64
	err = tx.QueryRow(ctx, `
		UPDATE sessions
		SET state = 'AUTHORIZED', stripe_checkout_id = $1, stripe_payment_intent_id = $2, updated_at = NOW()
		WHERE id = $3 AND state = 'CREATED'
		RETURNING user_id, rate_per_minute_cents * pre_auth_minutes
	`, cs.ID, paymentIntentID, internalSessionID).Scan(&userID, &totalCents)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("stripe: session not in CREATED state, ignoring duplicate webhook",
				"session_id", internalSessionID,
				"stripe_session_id", cs.ID,
			)
			return nil
		}
		return fmt.Errorf("authorize session: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users SET balance_cents = balance_cents + $1, updated_at = NOW()
		WHERE id = $2
	`, totalCents, userID); err != nil {
		return fmt.Errorf("credit wallet: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	slog.Info("stripe: session authorized and wallet credited",
		"stripe_session_id", cs.ID,
		"indranet_session_id", internalSessionID,
		"credited_cents", totalCents,
	)

	h.deps.Hub.SendToSession(internalSessionID, map[string]string{
		"type":       "session_authorized",
		"session_id": internalSessionID,
	})
	go h.awaitAgentReady(internalSessionID, 5*time.Minute)
	return nil
}

// handleWalletTopup credits the user's wallet directly from a top-up checkout.
func (h *Handlers) handleWalletTopup(ctx context.Context, cs stripe.CheckoutSession) error {
	userID := cs.Metadata["user_id"]
	if userID == "" {
		slog.Warn("stripe: wallet_topup checkout has no user_id metadata")
		return nil
	}

	amountCents := cs.AmountTotal
	if _, err := h.deps.Pool.Exec(ctx, `
		UPDATE users SET balance_cents = balance_cents + $1, updated_at = NOW()
		WHERE id = $2
	`, amountCents, userID); err != nil {
		return fmt.Errorf("credit wallet topup: %w", err)
	}

	slog.Info("stripe: wallet top-up credited",
		"user_id", userID,
		"amount_cents", amountCents,
	)
	return nil
}

// TopUpWallet creates a Stripe Checkout session to add funds to the user's wallet.
// POST /v1/users/me/topup — body: {"amount_cents": 1000}  (min $1.00)
func (h *Handlers) TopUpWallet(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxKeyUserID).(string)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		AmountCents int64 `json:"amount_cents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.AmountCents < 100 || req.AmountCents > 50_000 {
		http.Error(w, "amount_cents must be between 100 and 50000 ($1.00–$500.00)", http.StatusBadRequest)
		return
	}

	if h.deps.Config.StripeSecretKey == "" {
		// Dev mode: credit wallet directly without Stripe
		if _, err := h.deps.Pool.Exec(r.Context(), `
			UPDATE users SET balance_cents = balance_cents + $1, updated_at = NOW()
			WHERE id = $2
		`, req.AmountCents, userID); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"credited_cents": req.AmountCents,
			"dev_mode":       true,
		})
		return
	}

	stripe.Key = h.deps.Config.StripeSecretKey
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("IndraNet Wallet Top-Up"),
					},
					UnitAmount: stripe.Int64(req.AmountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"type":    "wallet_topup",
			"user_id": userID,
		},
		SuccessURL: stripe.String(h.deps.Config.FrontendBaseURL + "/dashboard?topup=success"),
		CancelURL:  stripe.String(h.deps.Config.FrontendBaseURL + "/dashboard"),
	}

	session, err := stripecs.New(params)
	if err != nil {
		slog.Error("stripe: wallet topup checkout creation failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"checkout_url": session.URL})
}

// handlePaymentFailed marks a session FAILED when Stripe payment fails or is declined.
func (h *Handlers) handlePaymentFailed(ctx context.Context, event stripe.Event) error {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		return fmt.Errorf("unmarshal payment intent: %w", err)
	}

	sessionID := pi.Metadata["indranet_session_id"]
	if sessionID == "" {
		slog.Debug("stripe: payment_failed with no indranet_session_id — ignoring")
		return nil
	}

	tag, err := h.deps.Pool.Exec(ctx, `
		UPDATE sessions SET state = 'FAILED', updated_at = NOW()
		WHERE id = $1 AND state IN ('CREATED', 'AUTHORIZED')
	`, sessionID)
	if err != nil {
		return fmt.Errorf("mark session failed: %w", err)
	}
	if tag.RowsAffected() > 0 {
		slog.Warn("stripe: payment failed, session marked FAILED", "session_id", sessionID)
		h.deps.Hub.SendToSession(sessionID, map[string]string{
			"type":   "session_failed",
			"reason": "payment_failed",
		})
	}
	return nil
}

// handleAccountUpdated syncs Stripe Connect account payout status to the host record.
func (h *Handlers) handleAccountUpdated(ctx context.Context, event stripe.Event) error {
	var account stripe.Account
	if err := json.Unmarshal(event.Data.Raw, &account); err != nil {
		return fmt.Errorf("unmarshal account: %w", err)
	}

	if _, err := h.deps.Pool.Exec(ctx, `
		UPDATE hosts SET payouts_enabled = $1, updated_at = NOW()
		WHERE stripe_account_id = $2
	`, account.PayoutsEnabled, account.ID); err != nil {
		return fmt.Errorf("update host payouts_enabled: %w", err)
	}
	slog.Info("stripe: host payout status updated",
		"account_id", account.ID,
		"payouts_enabled", account.PayoutsEnabled,
	)
	return nil
}

// awaitAgentReady marks a session FAILED if the host agent never transitions it to ACTIVE.
func (h *Handlers) awaitAgentReady(sessionID string, timeout time.Duration) {
	time.Sleep(timeout)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tag, err := h.deps.Pool.Exec(ctx, `
		UPDATE sessions SET state = 'FAILED', updated_at = NOW()
		WHERE id = $1 AND state = 'AUTHORIZED'
	`, sessionID)
	if err != nil {
		slog.Error("billing: awaitAgentReady: DB update failed",
			"session_id", sessionID, "error", err)
		return
	}
	if tag.RowsAffected() > 0 {
		slog.Warn("billing: agent never confirmed ready, session marked FAILED",
			"session_id", sessionID, "timeout", timeout)
		h.deps.Hub.SendToSession(sessionID, map[string]string{
			"type":   "session_failed",
			"reason": "agent_timeout",
		})
	}
}

// Signal handles WebSocket connections for WebRTC signaling.
// ?role=host|viewer&token=<jwt> — token is validated before upgrade.
// Security invariant: every WebSocket connection must carry a valid JWT.
func (h *Handlers) Signal(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	role := r.URL.Query().Get("role")
	if role != "host" && role != "viewer" {
		http.Error(w, "role must be 'host' or 'viewer'", http.StatusBadRequest)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		// Also accept Bearer header for host agents that use HTTP conventions
		token = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}
	if token == "" {
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}

	claims := &jwtClaims{}
	if _, err := jwt.ParseWithClaims(token, claims, h.jwtKeyFunc); err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// Role-specific authorization: viewer must own the session, host agent must own the host.
	if role == "viewer" {
		if claims.Role == "agent" {
			http.Error(w, "agents cannot connect as viewer", http.StatusForbidden)
			return
		}
		var ownerID string
		err := h.deps.Pool.QueryRow(r.Context(),
			`SELECT user_id FROM sessions WHERE id = $1`, sessionID,
		).Scan(&ownerID)
		if err != nil || ownerID != claims.UserID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	} else { // host
		if claims.Role != "agent" {
			http.Error(w, "only agent tokens may connect as host", http.StatusForbidden)
			return
		}
		var hostID string
		err := h.deps.Pool.QueryRow(r.Context(),
			`SELECT host_id FROM sessions WHERE id = $1`, sessionID,
		).Scan(&hostID)
		if err != nil || hostID != claims.UserID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	conn, err := h.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "session_id", sessionID, "error", err)
		return
	}

	client := signaling.NewClient(h.deps.Hub, conn, sessionID, role)
	go client.WritePump()
	client.ReadPump()
}
