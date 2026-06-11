package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/stripe/stripe-go/v76"
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

// handleCheckoutComplete processes a successful checkout and transitions the session to AUTHORIZED.
func (h *Handlers) handleCheckoutComplete(ctx context.Context, event stripe.Event) error {
	var checkoutSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
		return fmt.Errorf("unmarshal checkout session: %w", err)
	}

	internalSessionID := checkoutSession.Metadata["indranet_session_id"]
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

	var userID string
	var totalCents int64
	err = tx.QueryRow(ctx, `
		UPDATE sessions
		SET state = 'AUTHORIZED', stripe_checkout_id = $1, updated_at = NOW()
		WHERE id = $2 AND state = 'CREATED'
		RETURNING user_id, rate_per_minute_cents * pre_auth_minutes
	`, checkoutSession.ID, internalSessionID).Scan(&userID, &totalCents)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("stripe: session not in CREATED state, ignoring duplicate webhook",
				"session_id", internalSessionID,
				"stripe_session_id", checkoutSession.ID,
			)
			return nil
		}
		return fmt.Errorf("authorize session: %w", err)
	}

	// Credit the user's wallet so the billing engine has funds to deduct.
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
		"stripe_session_id", checkoutSession.ID,
		"indranet_session_id", internalSessionID,
		"credited_cents", totalCents,
	)

	// Notify the host agent that it should prepare the sandbox.
	h.deps.Hub.SendToSession(internalSessionID, map[string]string{
		"type":       "session_authorized",
		"session_id": internalSessionID,
	})

	// If agent doesn't confirm ACTIVE within 5 minutes, mark FAILED.
	go h.awaitAgentReady(internalSessionID, 5*time.Minute)

	return nil
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

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Signal handles WebSocket connections for WebRTC signaling.
// ?role=host|viewer — the hub relays messages between the two participants.
func (h *Handlers) Signal(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	role := r.URL.Query().Get("role")
	if role != "host" && role != "viewer" {
		http.Error(w, "role must be 'host' or 'viewer'", http.StatusBadRequest)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "session_id", sessionID, "error", err)
		return
	}

	client := signaling.NewClient(h.deps.Hub, conn, sessionID, role)
	go client.WritePump()
	client.ReadPump()
}
