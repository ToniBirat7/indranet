// PoC 04: Stripe Payment Gate Server
// Demonstrates: Stripe Checkout → webhook → JWT → stream access control
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
)

var (
	jwtSecret      = []byte(getEnv("JWT_SECRET", "poc-secret-change-in-prod"))
	stripeKey      = getEnv("STRIPE_SECRET_KEY", "")
	webhookSecret  = getEnv("STRIPE_WEBHOOK_SECRET", "")
	serverPort     = getEnv("PORT", "8080")
	successURL     = getEnv("SUCCESS_URL", fmt.Sprintf("http://localhost:%s/success", getEnv("PORT", "8080")))
	cancelURL      = getEnv("CANCEL_URL", fmt.Sprintf("http://localhost:%s/cancel", getEnv("PORT", "8080")))
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type SessionClaims struct {
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// issueSessionJWT generates a JWT granting access to a specific session.
func issueSessionJWT(sessionID string) (string, error) {
	claims := SessionClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "stream_access",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// verifySessionJWT validates a JWT and returns the session ID.
func verifySessionJWT(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*SessionClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	return claims.SessionID, nil
}

// POST /create-checkout — creates a Stripe Checkout session for $1
func handleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: In production, read host_id and duration from request body
	// and calculate the actual price. For PoC, hardcode $1.

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("IndraNet Session (1 hour)"),
					},
					UnitAmount: stripe.Int64(100), // $1.00 in cents
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(cancelURL),
		// Metadata carries our internal session ID into the webhook
		Metadata: map[string]string{
			"indranet_session_id": fmt.Sprintf("ses_%d", time.Now().UnixMilli()),
		},
	}

	checkoutSession, err := session.New(params)
	if err != nil {
		log.Printf("Stripe checkout error: %v", err)
		http.Error(w, "checkout creation failed", http.StatusInternalServerError)
		return
	}

	log.Printf("Checkout session created: %s", checkoutSession.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"checkout_url": checkoutSession.URL,
		"session_id":   checkoutSession.Metadata["indranet_session_id"],
	})
}

// POST /webhook — Stripe sends events here (verify signature, issue JWT on payment)
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "request body read error", http.StatusBadRequest)
		return
	}

	// Verify Stripe signature — CRITICAL: never skip this
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	log.Printf("Webhook received: %s", event.Type)

	switch event.Type {
	case "checkout.session.completed":
		var checkoutSession stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
			log.Printf("Failed to parse checkout session: %v", err)
			http.Error(w, "parse error", http.StatusBadRequest)
			return
		}

		internalSessionID := checkoutSession.Metadata["indranet_session_id"]
		log.Printf("Payment complete for session: %s", internalSessionID)

		// Issue JWT for stream access
		token, err := issueSessionJWT(internalSessionID)
		if err != nil {
			log.Printf("JWT issuance failed: %v", err)
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		// TODO: In production, store the JWT in Redis keyed by session ID
		// and notify the waiting client via WebSocket.
		// For PoC, just log it.
		log.Printf("Stream access token issued: %s...", token[:20])
	}

	w.WriteHeader(http.StatusOK)
}

// GET /verify — validate a JWT and return stream connection info
func handleVerify(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	sessionID, err := verifySessionJWT(tokenStr)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_id":    sessionID,
		"signaling_url": fmt.Sprintf("ws://localhost:8765/signal?room=%s&role=viewer", sessionID),
		"status":        "authorized",
	})
}

// GET /success — Stripe redirects here after payment
func handleSuccess(w http.ResponseWriter, r *http.Request) {
	stripeSessionID := r.URL.Query().Get("session_id")
	log.Printf("User returned from checkout: stripe_session=%s", stripeSessionID)
	fmt.Fprintln(w, `<!DOCTYPE html><html><body>
		<h1>Payment complete!</h1>
		<p>Your session is being set up. The stream token will appear shortly.</p>
		<p>Stripe session: `+stripeSessionID+`</p>
		<a href="/">Back</a>
	</body></html>`)
}

func main() {
	if stripeKey == "" {
		log.Fatal("STRIPE_SECRET_KEY environment variable is required")
	}
	stripe.Key = stripeKey

	http.HandleFunc("/create-checkout", handleCreateCheckout)
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/verify", handleVerify)
	http.HandleFunc("/success", handleSuccess)
	http.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Payment cancelled.")
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	addr := ":" + serverPort
	log.Printf("IndraNet PoC 04 Payment Gate server on %s", addr)
	log.Printf("Stripe webhook: POST /webhook (use: stripe listen --forward-to localhost%s/webhook)", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
