package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHealthEndpoint verifies /health returns 200 with DB + Redis reachable.
func TestHealthEndpoint(t *testing.T) {
	d := newTestDeps(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

// TestCreateSessionRequiresAuth verifies POST /v1/sessions returns 401 without a token.
func TestCreateSessionRequiresAuth(t *testing.T) {
	d := newTestDeps(t)
	body := `{"host_id":"host_test","duration_minutes":15}`
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestStripeWebhookRejectsInvalidSignature verifies webhooks with bad signatures return 400.
func TestStripeWebhookRejectsInvalidSignature(t *testing.T) {
	d := newTestDeps(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/stripe",
		strings.NewReader(`{"type":"checkout.session.completed"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "t=invalid,v1=badsig")
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid stripe signature, got %d", w.Code)
	}
}

// TestRegisterAndLogin exercises the auth flow end-to-end with a real DB.
func TestRegisterAndLogin(t *testing.T) {
	d := newTestDeps(t)
	email := "test_register_api@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	regBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": "password123",
		"name":     "Test User",
	})

	// Register
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var regResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&regResp)
	if regResp["token"] == nil {
		t.Error("register: expected token in response")
	}

	// Duplicate registration must fail with 409
	req2 := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Errorf("duplicate register: expected 409, got %d", w2.Code)
	}

	// Login with correct credentials
	loginBody, _ := json.Marshal(map[string]string{"email": email, "password": "password123"})
	req3 := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(loginBody))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	d.router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %s", w3.Code, w3.Body.String())
	}
	var loginResp map[string]interface{}
	json.NewDecoder(w3.Body).Decode(&loginResp)
	if loginResp["token"] == nil {
		t.Error("login: expected token in response")
	}

	// Wrong password must return 401
	badBody, _ := json.Marshal(map[string]string{"email": email, "password": "wrongpassword"})
	req4 := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(badBody))
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	d.router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnauthorized {
		t.Errorf("bad password: expected 401, got %d", w4.Code)
	}
}

// TestCreateSessionDevAutoAuthorize verifies that in dev mode (no Stripe key),
// a session is immediately AUTHORIZED after creation.
func TestCreateSessionDevAutoAuthorize(t *testing.T) {
	d := newTestDeps(t)

	if d.cfg.StripeSecretKey != "" {
		t.Skip("skipping dev-mode test: STRIPE_SECRET_KEY is set")
	}

	email := "test_session_dev@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	// Create user
	regBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "name": "Session Test",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("setup register: %d %s", w.Code, w.Body.String())
	}
	var regResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&regResp)
	token := regResp["token"].(string)
	userID := regResp["user_id"].(string)

	// Create host directly in DB (all NOT NULL columns required)
	var hostID string
	if err := d.pool.QueryRow(context.Background(), `
		INSERT INTO hosts (user_id, display_name, gpu_model, vram_gb, cpu_model,
		                   ram_gb, os, price_per_hour_cents, online)
		VALUES ($1, 'Test Host', 'RTX 4090', 24, 'Intel i9', 64, 'Windows 11', 600, true)
		RETURNING id`,
		userID,
	).Scan(&hostID); err != nil {
		t.Fatalf("setup create host: %v", err)
	}
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostID) })

	// Create session
	sessBody, _ := json.Marshal(map[string]interface{}{
		"host_id": hostID, "duration_minutes": 15,
	})
	req2 := httptest.NewRequest(http.MethodPost, "/v1/sessions", bytes.NewReader(sessBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("create session: expected 201, got %d: %s", w2.Code, w2.Body.String())
	}
	var sessResp map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&sessResp)

	if sessResp["state"] != "AUTHORIZED" {
		t.Errorf("dev mode: expected state=AUTHORIZED, got %q", sessResp["state"])
	}
	if sessResp["session_id"] == "" {
		t.Error("expected session_id in response")
	}

	// Verify user's balance was credited (dev mode should fund wallet for the session duration)
	var balanceCents int64
	err := d.pool.QueryRow(context.Background(),
		`SELECT balance_cents FROM users WHERE id = $1`, userID,
	).Scan(&balanceCents)
	if err != nil {
		t.Fatalf("fetch balance: %v", err)
	}
	// 600 cents/hr = 10 cents/min; 15 min = 150 cents credited, then 0 deducted (not ACTIVE yet)
	expectedCredit := int64(10 * 15) // 150 cents
	if balanceCents < expectedCredit {
		t.Errorf("wallet not credited: expected ≥%d cents, got %d", expectedCredit, balanceCents)
	}
}

// TestSignalRejectsUnauthenticated verifies /v1/signal/{id} returns 401 without a valid token.
func TestSignalRejectsUnauthenticated(t *testing.T) {
	d := newTestDeps(t)

	// No token — must get 401 (HTTP response before WS upgrade)
	req := httptest.NewRequest(http.MethodGet, "/v1/signal/ses_fake?role=viewer", nil)
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w.Code)
	}

	// Invalid token — must get 401
	req2 := httptest.NewRequest(http.MethodGet, "/v1/signal/ses_fake?role=viewer&token=badtoken", nil)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid token, got %d", w2.Code)
	}
}

// TestConcurrentSessionGuardBlocksDoubleBooking verifies that a second POST /v1/sessions
// targeting the same host returns 409 when that host already has an AUTHORIZED session.
func TestConcurrentSessionGuardBlocksDoubleBooking(t *testing.T) {
	d := newTestDeps(t)

	if d.cfg.StripeSecretKey != "" {
		t.Skip("skipping dev-mode test: STRIPE_SECRET_KEY is set")
	}

	email := "test_concurrent@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	regBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "name": "Concurrent Test",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var regResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&regResp)
	token := regResp["token"].(string)
	userID := regResp["user_id"].(string)

	var hostID string
	if err := d.pool.QueryRow(context.Background(), `
		INSERT INTO hosts (user_id, display_name, gpu_model, vram_gb, cpu_model,
		                   ram_gb, os, price_per_hour_cents, online)
		VALUES ($1, 'H', 'RTX 4090', 24, 'CPU', 32, 'Windows 11', 600, true) RETURNING id`,
		userID,
	).Scan(&hostID); err != nil {
		t.Fatalf("create host: %v", err)
	}
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostID) })

	sessBody, _ := json.Marshal(map[string]interface{}{"host_id": hostID, "duration_minutes": 15})

	// First booking — should succeed
	req1 := httptest.NewRequest(http.MethodPost, "/v1/sessions", bytes.NewReader(sessBody))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	d.router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first session: expected 201, got %d: %s", w1.Code, w1.Body.String())
	}

	// Second booking against the same host — must return 409
	req2 := httptest.NewRequest(http.MethodPost, "/v1/sessions", bytes.NewReader(sessBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Errorf("double booking: expected 409, got %d: %s", w2.Code, w2.Body.String())
	}
}
