package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthEndpoint verifies the /health endpoint returns 200 OK.
func TestHealthEndpoint(t *testing.T) {
	// TODO: Set up test server with mocked DB and Redis
	// TODO: GET /health → expect 200 + { "status": "ok" }
	t.Skip("TODO: implement test server setup")
}

// TestCreateSessionRequiresAuth verifies POST /v1/sessions returns 401 without a token.
func TestCreateSessionRequiresAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions", nil)
	w := httptest.NewRecorder()
	// TODO: Call handler without auth token
	// TODO: Expect 401
	_ = req
	_ = w
	t.Skip("TODO: implement with test router")
}

// TestStripeWebhookRejectsInvalidSignature verifies webhooks with bad signatures are rejected.
func TestStripeWebhookRejectsInvalidSignature(t *testing.T) {
	// TODO: POST /v1/webhooks/stripe with invalid Stripe-Signature header
	// TODO: Expect 400
	t.Skip("TODO: implement")
}
