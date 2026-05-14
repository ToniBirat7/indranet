package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

// StripeWebhook handles incoming Stripe webhook events.
// CRITICAL: Stripe-Signature header must be verified before processing any event.
// See research/06-payment/stripe-connect.md for the list of handled events.
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
		if err := h.handleCheckoutComplete(event); err != nil {
			slog.Error("stripe: checkout.session.completed handler failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

	case "payment_intent.payment_failed":
		// TODO: Mark session as FAILED, notify user

	case "account.updated":
		// TODO: Update host's payouts_enabled status in DB

	default:
		slog.Debug("stripe webhook: unhandled event type", "type", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

// handleCheckoutComplete processes a successful checkout and transitions the session to AUTHORIZED.
func (h *Handlers) handleCheckoutComplete(event stripe.Event) error {
	var checkoutSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
		return err
	}

	internalSessionID := checkoutSession.Metadata["indranet_session_id"]
	if internalSessionID == "" {
		slog.Warn("stripe: checkout completed with no indranet_session_id in metadata")
		return nil
	}

	slog.Info("stripe: payment complete, authorizing session",
		"stripe_session_id", checkoutSession.ID,
		"indranet_session_id", internalSessionID,
	)

	// TODO: UPDATE sessions SET state='AUTHORIZED', stripe_checkout_id=$1 WHERE id=$2 AND state='CREATED'
	// TODO: Notify host agent via signaling hub: session_authorized
	// TODO: Start a timeout goroutine: if agent doesn't confirm ready within 5min, mark FAILED

	return nil
}

// Signal handles WebSocket connections for WebRTC signaling.
// Clients connect with ?role=host|viewer and exchange SDP offer/answer and ICE candidates.
func (h *Handlers) Signal(w http.ResponseWriter, r *http.Request) {
	// TODO: Validate JWT token from query param
	// TODO: Extract session_id from URL path
	// TODO: Upgrade to WebSocket
	// TODO: Create signaling.Client and start ReadPump/WritePump goroutines
	http.Error(w, "TODO: implement WebSocket signaling", http.StatusNotImplemented)
}
