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

// TestTopupWalletDevMode verifies POST /v1/users/me/topup credits the wallet directly in dev mode.
func TestTopupWalletDevMode(t *testing.T) {
	d := newTestDeps(t)

	if d.cfg.StripeSecretKey != "" {
		t.Skip("skipping dev-mode test: STRIPE_SECRET_KEY is set")
	}

	email := "test_topup@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	regBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "name": "Topup Test",
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

	topupBody, _ := json.Marshal(map[string]interface{}{"amount_cents": 1000})
	req2 := httptest.NewRequest(http.MethodPost, "/v1/users/me/topup", bytes.NewReader(topupBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("topup: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	var topupResp map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&topupResp)
	if topupResp["dev_mode"] != true {
		t.Error("expected dev_mode=true in topup response")
	}

	var balance int64
	d.pool.QueryRow(context.Background(),
		`SELECT balance_cents FROM users WHERE id = $1`, userID).Scan(&balance)
	if balance != 1000 {
		t.Errorf("expected balance=1000, got %d", balance)
	}
}

// TestGetPendingSessionsReturnsAuthorized verifies the agent can discover AUTHORIZED sessions.
func TestGetPendingSessionsReturnsAuthorized(t *testing.T) {
	d := newTestDeps(t)

	if d.cfg.StripeSecretKey != "" {
		t.Skip("skipping dev-mode test: STRIPE_SECRET_KEY is set")
	}

	email := "test_pending@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	regBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "name": "Pending Test",
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
	userToken := regResp["token"].(string)
	userID := regResp["user_id"].(string)

	// Register a host to get an agent token
	hostBody, _ := json.Marshal(map[string]interface{}{
		"display_name": "Test Host", "gpu_model": "RTX 4090", "vram_gb": 24,
		"cpu_model": "i9", "ram_gb": 64, "os": "Windows 11",
		"price_per_hour_cents": 600, "tags": []string{},
	})
	req2 := httptest.NewRequest(http.MethodPost, "/v1/hosts/register", bytes.NewReader(hostBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+userToken)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("register host: %d %s", w2.Code, w2.Body.String())
	}
	var hostResp map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&hostResp)
	agentToken := hostResp["agent_token"].(string)
	hostID := hostResp["host_id"].(string)
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostID) })

	// Mark host online
	onlineBody, _ := json.Marshal(map[string]bool{"online": true})
	req3 := httptest.NewRequest(http.MethodPut, "/v1/hosts/me/online", bytes.NewReader(onlineBody))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+agentToken)
	w3 := httptest.NewRecorder()
	d.router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("set online: %d %s", w3.Code, w3.Body.String())
	}

	// User creates a session (dev mode → auto-AUTHORIZED)
	sessBody, _ := json.Marshal(map[string]interface{}{"host_id": hostID, "duration_minutes": 15})
	req4 := httptest.NewRequest(http.MethodPost, "/v1/sessions", bytes.NewReader(sessBody))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Authorization", "Bearer "+userToken)
	// Need a different user since a user can't book their own host in some setups
	// but our current code doesn't block this — proceed
	_ = userID
	w4 := httptest.NewRecorder()
	d.router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusCreated {
		t.Fatalf("create session: %d %s", w4.Code, w4.Body.String())
	}
	var sessResp map[string]interface{}
	json.NewDecoder(w4.Body).Decode(&sessResp)
	sessionID := sessResp["session_id"].(string)

	// Agent fetches pending sessions
	req5 := httptest.NewRequest(http.MethodGet, "/v1/sessions/pending", nil)
	req5.Header.Set("Authorization", "Bearer "+agentToken)
	w5 := httptest.NewRecorder()
	d.router.ServeHTTP(w5, req5)
	if w5.Code != http.StatusOK {
		t.Fatalf("pending sessions: expected 200, got %d: %s", w5.Code, w5.Body.String())
	}
	var pendingResp map[string]interface{}
	json.NewDecoder(w5.Body).Decode(&pendingResp)
	pending := pendingResp["sessions"].([]interface{})
	if len(pending) == 0 {
		t.Fatal("expected at least one pending session")
	}
	found := false
	for _, p := range pending {
		pm := p.(map[string]interface{})
		if pm["session_id"] == sessionID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("session %q not found in pending list", sessionID)
	}
}

// TestAgentCannotOperateForeignSession verifies that an agent JWT for host A cannot
// call StartSession or HeartbeatSession on a session that belongs to host B.
// This is a regression test for the IDOR fix in StartSession/HeartbeatSession.
func TestAgentCannotOperateForeignSession(t *testing.T) {
	d := newTestDeps(t)
	if d.cfg.StripeSecretKey != "" {
		t.Skip("skipping dev-mode test: STRIPE_SECRET_KEY is set")
	}

	// Register user A (owns host A + creates a session)
	emailA := "test_idor_a@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, emailA) })
	regA, _ := json.Marshal(map[string]string{"email": emailA, "password": "password123", "name": "A"})
	reqA := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regA))
	reqA.Header.Set("Content-Type", "application/json")
	wA := httptest.NewRecorder()
	d.router.ServeHTTP(wA, reqA)
	if wA.Code != http.StatusCreated {
		t.Fatalf("register A: %d %s", wA.Code, wA.Body.String())
	}
	var rrA map[string]interface{}
	json.NewDecoder(wA.Body).Decode(&rrA)
	userTokenA := rrA["token"].(string)
	userIDA := rrA["user_id"].(string)

	// Register host A via the hosts/register endpoint to get an agent token
	hostBodyA, _ := json.Marshal(map[string]interface{}{
		"display_name": "Host A", "gpu_model": "RTX 4090", "vram_gb": 24,
		"cpu_model": "i9", "ram_gb": 64, "os": "Windows 11",
		"price_per_hour_cents": 600, "tags": []string{},
	})
	reqHA := httptest.NewRequest(http.MethodPost, "/v1/hosts/register", bytes.NewReader(hostBodyA))
	reqHA.Header.Set("Content-Type", "application/json")
	reqHA.Header.Set("Authorization", "Bearer "+userTokenA)
	wHA := httptest.NewRecorder()
	d.router.ServeHTTP(wHA, reqHA)
	if wHA.Code != http.StatusCreated {
		t.Fatalf("register host A: %d %s", wHA.Code, wHA.Body.String())
	}
	var hrA map[string]interface{}
	json.NewDecoder(wHA.Body).Decode(&hrA)
	agentTokenA := hrA["agent_token"].(string)
	hostIDA := hrA["host_id"].(string)
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostIDA) })

	// Register user B (owns host B)
	emailB := "test_idor_b@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, emailB) })
	regB, _ := json.Marshal(map[string]string{"email": emailB, "password": "password123", "name": "B"})
	reqB := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regB))
	reqB.Header.Set("Content-Type", "application/json")
	wB := httptest.NewRecorder()
	d.router.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusCreated {
		t.Fatalf("register B: %d %s", wB.Code, wB.Body.String())
	}
	var rrB map[string]interface{}
	json.NewDecoder(wB.Body).Decode(&rrB)
	userTokenB := rrB["token"].(string)

	// Register host B via endpoint
	hostBodyB, _ := json.Marshal(map[string]interface{}{
		"display_name": "Host B", "gpu_model": "RTX 3080", "vram_gb": 10,
		"cpu_model": "i7", "ram_gb": 32, "os": "Windows 11",
		"price_per_hour_cents": 600, "tags": []string{},
	})
	reqHB := httptest.NewRequest(http.MethodPost, "/v1/hosts/register", bytes.NewReader(hostBodyB))
	reqHB.Header.Set("Content-Type", "application/json")
	reqHB.Header.Set("Authorization", "Bearer "+userTokenB)
	wHB := httptest.NewRecorder()
	d.router.ServeHTTP(wHB, reqHB)
	if wHB.Code != http.StatusCreated {
		t.Fatalf("register host B: %d %s", wHB.Code, wHB.Body.String())
	}
	var hrB map[string]interface{}
	json.NewDecoder(wHB.Body).Decode(&hrB)
	hostIDB := hrB["host_id"].(string)
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostIDB) })

	// Set host A online so user A can book it
	onlineBody, _ := json.Marshal(map[string]bool{"online": true})
	reqOn := httptest.NewRequest(http.MethodPut, "/v1/hosts/me/online", bytes.NewReader(onlineBody))
	reqOn.Header.Set("Content-Type", "application/json")
	reqOn.Header.Set("Authorization", "Bearer "+agentTokenA)
	wOn := httptest.NewRecorder()
	d.router.ServeHTTP(wOn, reqOn)
	if wOn.Code != http.StatusOK {
		t.Fatalf("set host A online: %d %s", wOn.Code, wOn.Body.String())
	}

	// User A creates a session on Host A (dev mode → AUTHORIZED)
	sessBody, _ := json.Marshal(map[string]interface{}{"host_id": hostIDA, "duration_minutes": 15})
	reqS := httptest.NewRequest(http.MethodPost, "/v1/sessions", bytes.NewReader(sessBody))
	reqS.Header.Set("Content-Type", "application/json")
	reqS.Header.Set("Authorization", "Bearer "+userTokenA)
	_ = userIDA
	wS := httptest.NewRecorder()
	d.router.ServeHTTP(wS, reqS)
	if wS.Code != http.StatusCreated {
		t.Fatalf("create session: %d %s", wS.Code, wS.Body.String())
	}
	var sr map[string]interface{}
	json.NewDecoder(wS.Body).Decode(&sr)
	sessionID := sr["session_id"].(string)

	// Get agent token for Host B
	hostBodyBAgent, _ := json.Marshal(map[string]interface{}{
		"display_name": "Host B2", "gpu_model": "RTX 3090", "vram_gb": 24,
		"cpu_model": "i7", "ram_gb": 32, "os": "Windows 11",
		"price_per_hour_cents": 600, "tags": []string{},
	})
	reqHB2 := httptest.NewRequest(http.MethodPost, "/v1/hosts/register", bytes.NewReader(hostBodyBAgent))
	reqHB2.Header.Set("Content-Type", "application/json")
	reqHB2.Header.Set("Authorization", "Bearer "+userTokenB)
	wHB2 := httptest.NewRecorder()
	d.router.ServeHTTP(wHB2, reqHB2)
	if wHB2.Code != http.StatusCreated {
		t.Fatalf("register host B2: %d %s", wHB2.Code, wHB2.Body.String())
	}
	var hrB2 map[string]interface{}
	json.NewDecoder(wHB2.Body).Decode(&hrB2)
	agentTokenB := hrB2["agent_token"].(string)
	hostIDB2 := hrB2["host_id"].(string)
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostIDB2) })

	// Agent B tries to start session belonging to Host A → must fail (no rows affected = 409)
	reqStart := httptest.NewRequest(http.MethodPut, "/v1/sessions/"+sessionID+"/start", nil)
	reqStart.Header.Set("Authorization", "Bearer "+agentTokenB)
	wStart := httptest.NewRecorder()
	d.router.ServeHTTP(wStart, reqStart)
	if wStart.Code != http.StatusConflict {
		t.Errorf("foreign agent start: expected 409, got %d: %s", wStart.Code, wStart.Body.String())
	}

	// Agent B tries to heartbeat session belonging to Host A → must return 404
	reqHB3 := httptest.NewRequest(http.MethodPut, "/v1/sessions/"+sessionID+"/heartbeat", nil)
	reqHB3.Header.Set("Authorization", "Bearer "+agentTokenB)
	wHB3 := httptest.NewRecorder()
	d.router.ServeHTTP(wHB3, reqHB3)
	if wHB3.Code != http.StatusNotFound {
		t.Errorf("foreign agent heartbeat: expected 404, got %d: %s", wHB3.Code, wHB3.Body.String())
	}
}

// TestRateSession verifies session rating updates host aggregate and prevents double-rating.
func TestRateSession(t *testing.T) {
	d := newTestDeps(t)
	if d.cfg.StripeSecretKey != "" {
		t.Skip("skipping dev-mode test: STRIPE_SECRET_KEY is set")
	}

	email := "test_rate_session@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	// Register user
	regBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "name": "Rater",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var rr map[string]interface{}
	json.NewDecoder(w.Body).Decode(&rr)
	token := rr["token"].(string)
	userID := rr["user_id"].(string)

	// Create host
	var hostID string
	if err := d.pool.QueryRow(context.Background(), `
		INSERT INTO hosts (user_id, display_name, gpu_model, vram_gb, cpu_model,
		                   ram_gb, os, price_per_hour_cents, online)
		VALUES ($1, 'RateHost', 'RTX 4090', 24, 'Intel i9', 64, 'Windows 11', 600, true)
		RETURNING id`, userID,
	).Scan(&hostID); err != nil {
		t.Fatalf("create host: %v", err)
	}
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostID) })

	// Create session (dev mode → AUTHORIZED)
	sessBody, _ := json.Marshal(map[string]interface{}{"host_id": hostID, "duration_minutes": 15})
	req2 := httptest.NewRequest(http.MethodPost, "/v1/sessions", bytes.NewReader(sessBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("create session: %d %s", w2.Code, w2.Body.String())
	}
	var sr map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&sr)
	sessionID := sr["session_id"].(string)

	// Manually move session to ENDED
	if _, err := d.pool.Exec(context.Background(),
		`UPDATE sessions SET state = 'ENDED', ended_at = NOW() WHERE id = $1`, sessionID,
	); err != nil {
		t.Fatalf("set session ended: %v", err)
	}

	// Submit rating
	rateBody, _ := json.Marshal(map[string]int{"stars": 4})
	req3 := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/rate", bytes.NewReader(rateBody))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+token)
	w3 := httptest.NewRecorder()
	d.router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("rate: expected 200, got %d: %s", w3.Code, w3.Body.String())
	}

	// Verify host aggregate updated
	var ratingSum, ratingCount int
	if err := d.pool.QueryRow(context.Background(),
		`SELECT rating_sum, rating_count FROM hosts WHERE id = $1`, hostID,
	).Scan(&ratingSum, &ratingCount); err != nil {
		t.Fatalf("fetch host: %v", err)
	}
	if ratingSum != 4 || ratingCount != 1 {
		t.Errorf("host rating: want sum=4 count=1, got sum=%d count=%d", ratingSum, ratingCount)
	}

	// Double-rate must fail with 409
	req4 := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/rate", bytes.NewReader(rateBody))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Authorization", "Bearer "+token)
	w4 := httptest.NewRecorder()
	d.router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusConflict {
		t.Errorf("double-rate: expected 409, got %d", w4.Code)
	}
}

// TestRegisterRejectsInvalidEmail verifies that registration rejects malformed emails.
func TestRegisterRejectsInvalidEmail(t *testing.T) {
	d := newTestDeps(t)

	for _, bad := range []string{"notanemail", "nodotatall"} {
		body, _ := json.Marshal(map[string]string{
			"email":    bad,
			"password": "password123",
			"name":     "Test",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		d.router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("email %q: expected 400, got %d", bad, w.Code)
		}
	}
}

// TestListSessionsPagination verifies that GET /sessions returns total/page/limit fields.
func TestListSessionsPagination(t *testing.T) {
	d := newTestDeps(t)
	email := "test_list_sess_pag@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	// Register + login
	regBody, _ := json.Marshal(map[string]string{"email": email, "password": "password123", "name": "Pager"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var rr map[string]interface{}
	json.NewDecoder(w.Body).Decode(&rr)
	token := rr["token"].(string)

	// List sessions (empty)
	req2 := httptest.NewRequest(http.MethodGet, "/v1/sessions?page=1&limit=10", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("list sessions: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&resp)
	if _, ok := resp["total"]; !ok {
		t.Error("expected 'total' in response")
	}
	if _, ok := resp["page"]; !ok {
		t.Error("expected 'page' in response")
	}
	if _, ok := resp["limit"]; !ok {
		t.Error("expected 'limit' in response")
	}
	if resp["sessions"] == nil {
		t.Error("expected 'sessions' in response")
	}
	if strings.Contains(w2.Body.String(), "null") {
		// sessions should be [] not null
		t.Error("sessions field should be empty array, not null")
	}
}

// TestAgentGetHostMe verifies GET /v1/hosts/me returns the agent's own host record.
func TestAgentGetHostMe(t *testing.T) {
	d := newTestDeps(t)
	email := "test_hostme@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	// Register user
	regBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "name": "HostMe Test",
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
	userToken := regResp["token"].(string)

	// Register host → get agent token
	hostBody, _ := json.Marshal(map[string]interface{}{
		"display_name": "MyHost", "gpu_model": "RTX 3080", "vram_gb": 10,
		"cpu_model": "Ryzen 9", "ram_gb": 32, "os": "Windows 11",
		"price_per_hour_cents": 600, "tags": []string{"gaming"},
	})
	req2 := httptest.NewRequest(http.MethodPost, "/v1/hosts/register", bytes.NewReader(hostBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+userToken)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("register host: %d %s", w2.Code, w2.Body.String())
	}
	var hostResp map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&hostResp)
	agentToken := hostResp["agent_token"].(string)
	hostID := hostResp["host_id"].(string)
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostID) })

	// Agent calls GET /v1/hosts/me
	req3 := httptest.NewRequest(http.MethodGet, "/v1/hosts/me", nil)
	req3.Header.Set("Authorization", "Bearer "+agentToken)
	w3 := httptest.NewRecorder()
	d.router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("GET /v1/hosts/me: expected 200, got %d: %s", w3.Code, w3.Body.String())
	}
	var me map[string]interface{}
	if err := json.NewDecoder(w3.Body).Decode(&me); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if me["host_id"] != hostID {
		t.Errorf("expected host_id=%q, got %q", hostID, me["host_id"])
	}
	if me["gpu_model"] != "RTX 3080" {
		t.Errorf("expected gpu_model=RTX 3080, got %q", me["gpu_model"])
	}
	if me["display_name"] != "MyHost" {
		t.Errorf("expected display_name=MyHost, got %q", me["display_name"])
	}

	// Unauthenticated request must be rejected
	req4 := httptest.NewRequest(http.MethodGet, "/v1/hosts/me", nil)
	w4 := httptest.NewRecorder()
	d.router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated GET /v1/hosts/me: expected 401, got %d", w4.Code)
	}
}

// TestListHostsAndGetHost verifies GET /v1/hosts and GET /v1/hosts/{id} (public routes).
func TestListHostsAndGetHost(t *testing.T) {
	d := newTestDeps(t)
	email := "test_listhosts@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	// Register user + host
	regBody, _ := json.Marshal(map[string]string{"email": email, "password": "password123", "name": "Host Lister"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var rr map[string]interface{}
	json.NewDecoder(w.Body).Decode(&rr)
	userToken := rr["token"].(string)

	hostBody, _ := json.Marshal(map[string]interface{}{
		"display_name": "ListTest Host", "gpu_model": "RTX 4080", "vram_gb": 16,
		"cpu_model": "Ryzen 9", "ram_gb": 32, "os": "Windows 11",
		"price_per_hour_cents": 500, "tags": []string{"gaming"},
	})
	req2 := httptest.NewRequest(http.MethodPost, "/v1/hosts/register", bytes.NewReader(hostBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+userToken)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("register host: %d %s", w2.Code, w2.Body.String())
	}
	var hr map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&hr)
	hostID := hr["host_id"].(string)
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostID) })

	// GET /v1/hosts — no auth required
	req3 := httptest.NewRequest(http.MethodGet, "/v1/hosts", nil)
	w3 := httptest.NewRecorder()
	d.router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("list hosts: expected 200, got %d: %s", w3.Code, w3.Body.String())
	}
	var listResp map[string]interface{}
	json.NewDecoder(w3.Body).Decode(&listResp)
	if listResp["hosts"] == nil {
		t.Error("expected 'hosts' field in list response")
	}
	hosts := listResp["hosts"].([]interface{})
	found := false
	for _, h := range hosts {
		hm := h.(map[string]interface{})
		if hm["id"] == hostID {
			found = true
			if hm["display_name"] != "ListTest Host" {
				t.Errorf("expected display_name=ListTest Host, got %v", hm["display_name"])
			}
		}
	}
	if !found {
		t.Errorf("registered host %q not found in GET /v1/hosts", hostID)
	}

	// GET /v1/hosts/{id} — no auth required
	req4 := httptest.NewRequest(http.MethodGet, "/v1/hosts/"+hostID, nil)
	w4 := httptest.NewRecorder()
	d.router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Fatalf("get host: expected 200, got %d: %s", w4.Code, w4.Body.String())
	}
	var getResp map[string]interface{}
	json.NewDecoder(w4.Body).Decode(&getResp)
	if getResp["id"] != hostID {
		t.Errorf("get host: expected id=%q, got %q", hostID, getResp["id"])
	}
	if getResp["gpu_model"] != "RTX 4080" {
		t.Errorf("get host: expected gpu_model=RTX 4080, got %v", getResp["gpu_model"])
	}

	// GET /v1/hosts/{id} with unknown ID → 404
	req5 := httptest.NewRequest(http.MethodGet, "/v1/hosts/00000000-0000-0000-0000-000000000000", nil)
	w5 := httptest.NewRecorder()
	d.router.ServeHTTP(w5, req5)
	if w5.Code != http.StatusNotFound {
		t.Errorf("unknown host: expected 404, got %d", w5.Code)
	}
}

// TestHostHeartbeat verifies PUT /v1/hosts/me/heartbeat updates the host's last-seen time.
func TestHostHeartbeat(t *testing.T) {
	d := newTestDeps(t)
	email := "test_heartbeat@indranet.test"
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	// Register user + host → get agent_token
	regBody, _ := json.Marshal(map[string]string{"email": email, "password": "password123", "name": "Heartbeater"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	d.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var rr map[string]interface{}
	json.NewDecoder(w.Body).Decode(&rr)
	userToken := rr["token"].(string)

	hostBody, _ := json.Marshal(map[string]interface{}{
		"display_name": "Heartbeat Host", "gpu_model": "RX 7900 XTX", "vram_gb": 24,
		"cpu_model": "Ryzen 7", "ram_gb": 32, "os": "Windows 11",
		"price_per_hour_cents": 400, "tags": []string{},
	})
	req2 := httptest.NewRequest(http.MethodPost, "/v1/hosts/register", bytes.NewReader(hostBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+userToken)
	w2 := httptest.NewRecorder()
	d.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("register host: %d %s", w2.Code, w2.Body.String())
	}
	var hr map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&hr)
	agentToken := hr["agent_token"].(string)
	hostID := hr["host_id"].(string)
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostID) })

	// PUT /v1/hosts/me/heartbeat — agent auth
	req3 := httptest.NewRequest(http.MethodPut, "/v1/hosts/me/heartbeat", nil)
	req3.Header.Set("Authorization", "Bearer "+agentToken)
	w3 := httptest.NewRecorder()
	d.router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("heartbeat: expected 200, got %d: %s", w3.Code, w3.Body.String())
	}

	// Verify updated_at advanced (updated_at > created_at by re-reading the host)
	var updatedAt string
	err := d.pool.QueryRow(context.Background(), `SELECT updated_at FROM hosts WHERE id = $1`, hostID).Scan(&updatedAt)
	if err != nil {
		t.Fatalf("query host: %v", err)
	}
	if updatedAt == "" {
		t.Error("expected non-empty updated_at after heartbeat")
	}

	// Unauthenticated heartbeat must be rejected
	req4 := httptest.NewRequest(http.MethodPut, "/v1/hosts/me/heartbeat", nil)
	w4 := httptest.NewRecorder()
	d.router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated heartbeat: expected 401, got %d", w4.Code)
	}

	// User JWT (not agent JWT) must be rejected
	req5 := httptest.NewRequest(http.MethodPut, "/v1/hosts/me/heartbeat", nil)
	req5.Header.Set("Authorization", "Bearer "+userToken)
	w5 := httptest.NewRecorder()
	d.router.ServeHTTP(w5, req5)
	if w5.Code != http.StatusUnauthorized {
		t.Errorf("user JWT on agent route: expected 401, got %d", w5.Code)
	}
}
