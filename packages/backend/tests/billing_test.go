package tests

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// billingTestSetup creates a user + ACTIVE session seeded with specific balance/rate.
// Returns (userID, sessionID). Registers cleanup via t.Cleanup.
func billingTestSetup(t *testing.T, pool *pgxpool.Pool, balanceCents, ratePerMinuteCents int64) (userID, sessionID string) {
	t.Helper()
	ctx := context.Background()
	email := "billing_" + t.Name() + "@indranet.test"

	if err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name, balance_cents)
		 VALUES ($1, 'testhash', 'Billing Test', $2) RETURNING id`,
		email, balanceCents,
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { cleanupTestUser(t, pool, email) })

	var hostID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO hosts (user_id, display_name, gpu_model, vram_gb, cpu_model,
		                   ram_gb, os, price_per_hour_cents, online)
		VALUES ($1, 'Test Host', 'RTX 4090', 24, 'Intel i9', 64, 'Windows 11', $2, true)
		RETURNING id`,
		userID, ratePerMinuteCents*60,
	).Scan(&hostID); err != nil {
		t.Fatalf("create host: %v", err)
	}
	t.Cleanup(func() { cleanupTestHost(t, pool, hostID) })

	if err := pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, host_id, state, rate_per_minute_cents, pre_auth_minutes, started_at)
		VALUES ($1, $2, 'ACTIVE', $3, 15, NOW()) RETURNING id`,
		userID, hostID, ratePerMinuteCents,
	).Scan(&sessionID); err != nil {
		t.Fatalf("create session: %v", err)
	}
	return userID, sessionID
}

// TestBillingTickDeductsBalance verifies one tick deducts exactly rate from balance,
// inserts one billing_tick row, and increments session.total_charged_cents.
func TestBillingTickDeductsBalance(t *testing.T) {
	d := newTestDeps(t)

	const rate int64 = 100    // 100 cents/min = $1/min
	const initial int64 = 1000 // $10

	userID, sessionID := billingTestSetup(t, d.pool, initial, rate)

	d.engine.Tick()

	var newBalance, totalCharged int64
	var tickCount int
	d.pool.QueryRow(context.Background(),
		`SELECT balance_cents FROM users WHERE id = $1`, userID).Scan(&newBalance)
	d.pool.QueryRow(context.Background(),
		`SELECT total_charged_cents FROM sessions WHERE id = $1`, sessionID).Scan(&totalCharged)
	d.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM billing_ticks WHERE session_id = $1`, sessionID).Scan(&tickCount)

	if newBalance != initial-rate {
		t.Errorf("balance: want %d, got %d", initial-rate, newBalance)
	}
	if totalCharged != rate {
		t.Errorf("total_charged_cents: want %d, got %d", rate, totalCharged)
	}
	if tickCount != 1 {
		t.Errorf("billing_ticks count: want 1, got %d", tickCount)
	}
}

// TestBillingKillsSessionOnZeroBalance verifies the engine transitions a session to
// ENDING when balance drops to zero after a tick.
func TestBillingKillsSessionOnZeroBalance(t *testing.T) {
	d := newTestDeps(t)

	const rate int64 = 100
	// Balance equals exactly one tick — after deduction: 0 → ENDING
	_, sessionID := billingTestSetup(t, d.pool, rate, rate)

	d.engine.Tick()

	var state string
	d.pool.QueryRow(context.Background(),
		`SELECT state FROM sessions WHERE id = $1`, sessionID).Scan(&state)

	if state != "ENDING" {
		t.Errorf("expected ENDING after balance exhausted, got %q", state)
	}
}

// TestBillingWarningAtThreshold verifies the session stays ACTIVE and billing_ticks
// accumulate correctly when balance drops below the 5-minute warning threshold.
func TestBillingWarningAtThreshold(t *testing.T) {
	d := newTestDeps(t)

	const rate int64 = 100
	// 6 min balance:
	//   tick 1: balance=500 → 500 < warningCents(500) is false → no warning
	//   tick 2: balance=400 → 400 < 500 → warning sent (via hub — can't assert here)
	_, sessionID := billingTestSetup(t, d.pool, rate*6, rate)

	d.engine.Tick()
	d.engine.Tick()

	var state string
	var tickCount int
	d.pool.QueryRow(context.Background(),
		`SELECT state FROM sessions WHERE id = $1`, sessionID).Scan(&state)
	d.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM billing_ticks WHERE session_id = $1`, sessionID).Scan(&tickCount)

	if state != "ACTIVE" {
		t.Errorf("session should still be ACTIVE after warning tick, got %q", state)
	}
	if tickCount != 2 {
		t.Errorf("expected 2 billing_ticks, got %d", tickCount)
	}
}

// TestBillingFinalTickCapped verifies that when balance < rate, the charged amount
// is capped at balance_cents (not rate), leaving balance at exactly 0.
// This exercises the LEAST(balance_cents, rate) CTE in processSessionTick.
func TestBillingFinalTickCapped(t *testing.T) {
	d := newTestDeps(t)

	const rate int64 = 100
	const balance int64 = 40 // less than one full tick

	userID, sessionID := billingTestSetup(t, d.pool, balance, rate)

	d.engine.Tick()

	var newBalance, totalCharged int64
	var tickAmount int64
	d.pool.QueryRow(context.Background(),
		`SELECT balance_cents FROM users WHERE id = $1`, userID).Scan(&newBalance)
	d.pool.QueryRow(context.Background(),
		`SELECT total_charged_cents FROM sessions WHERE id = $1`, sessionID).Scan(&totalCharged)
	d.pool.QueryRow(context.Background(),
		`SELECT amount_cents FROM billing_ticks WHERE session_id = $1`, sessionID).Scan(&tickAmount)

	if newBalance != 0 {
		t.Errorf("balance: want 0, got %d", newBalance)
	}
	if totalCharged != balance {
		t.Errorf("total_charged_cents: want %d (balance), got %d", balance, totalCharged)
	}
	if tickAmount != balance {
		t.Errorf("billing_tick amount_cents: want %d (capped), got %d", balance, tickAmount)
	}

	var state string
	d.pool.QueryRow(context.Background(),
		`SELECT state FROM sessions WHERE id = $1`, sessionID).Scan(&state)
	if state != "ENDING" {
		t.Errorf("session should be ENDING after balance exhausted, got %q", state)
	}
}

// TestBillingTransactionAtomicity verifies balance deduction, tick insertion, and
// session total update all commit together (happy path).
func TestBillingTransactionAtomicity(t *testing.T) {
	d := newTestDeps(t)

	const rate int64 = 100
	const balance int64 = 500
	userID, sessionID := billingTestSetup(t, d.pool, balance, rate)

	d.engine.Tick()

	var newBalance, totalCharged int64
	var tickCount int
	d.pool.QueryRow(context.Background(),
		`SELECT balance_cents FROM users WHERE id = $1`, userID).Scan(&newBalance)
	d.pool.QueryRow(context.Background(),
		`SELECT total_charged_cents FROM sessions WHERE id = $1`, sessionID).Scan(&totalCharged)
	d.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM billing_ticks WHERE session_id = $1`, sessionID).Scan(&tickCount)

	if newBalance != balance-rate || totalCharged != rate || tickCount != 1 {
		t.Errorf("atomicity: balance=%d (want %d), total_charged=%d (want %d), ticks=%d (want 1)",
			newBalance, balance-rate, totalCharged, rate, tickCount)
	}
}

// TestSweepAbandonedCreatedToFailed verifies sessions in CREATED state older than
// 30 minutes are transitioned to FAILED by the billing sweep.
func TestSweepAbandonedCreatedToFailed(t *testing.T) {
	d := newTestDeps(t)
	ctx := context.Background()

	email := "billing_sweep_abandoned@indranet.test"
	var userID string
	if err := d.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, 'h', 'S') RETURNING id`, email,
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	var hostID string
	if err := d.pool.QueryRow(ctx, `
		INSERT INTO hosts (user_id, display_name, gpu_model, vram_gb, cpu_model,
		                   ram_gb, os, price_per_hour_cents)
		VALUES ($1, 'H', 'RTX', 8, 'CPU', 16, 'Win', 600) RETURNING id`, userID,
	).Scan(&hostID); err != nil {
		t.Fatalf("create host: %v", err)
	}
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostID) })

	// Insert a CREATED session with created_at well in the past (>30 min ago)
	var sessionID string
	if err := d.pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, host_id, state, rate_per_minute_cents, pre_auth_minutes, created_at, updated_at)
		VALUES ($1, $2, 'CREATED', 10, 15, NOW() - INTERVAL '31 minutes', NOW() - INTERVAL '31 minutes')
		RETURNING id`, userID, hostID,
	).Scan(&sessionID); err != nil {
		t.Fatalf("create session: %v", err)
	}

	d.engine.Sweep()

	var state string
	if err := d.pool.QueryRow(ctx, `SELECT state FROM sessions WHERE id = $1`, sessionID).Scan(&state); err != nil {
		t.Fatalf("fetch state: %v", err)
	}
	if state != "FAILED" {
		t.Errorf("expected FAILED for abandoned CREATED session, got %q", state)
	}
}

// TestSweepStaleAuthorizedToFailed verifies sessions stuck in AUTHORIZED for >10 minutes
// are swept to FAILED (backstop for lost awaitAgentReady goroutines).
func TestSweepStaleAuthorizedToFailed(t *testing.T) {
	d := newTestDeps(t)
	ctx := context.Background()

	email := "billing_sweep_authorized@indranet.test"
	var userID string
	if err := d.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, 'h', 'S') RETURNING id`, email,
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { cleanupTestUser(t, d.pool, email) })

	var hostID string
	if err := d.pool.QueryRow(ctx, `
		INSERT INTO hosts (user_id, display_name, gpu_model, vram_gb, cpu_model,
		                   ram_gb, os, price_per_hour_cents)
		VALUES ($1, 'H', 'RTX', 8, 'CPU', 16, 'Win', 600) RETURNING id`, userID,
	).Scan(&hostID); err != nil {
		t.Fatalf("create host: %v", err)
	}
	t.Cleanup(func() { cleanupTestHost(t, d.pool, hostID) })

	// AUTHORIZED session with updated_at > 10 minutes ago
	var sessionID string
	if err := d.pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, host_id, state, rate_per_minute_cents, pre_auth_minutes, updated_at)
		VALUES ($1, $2, 'AUTHORIZED', 10, 15, NOW() - INTERVAL '11 minutes')
		RETURNING id`, userID, hostID,
	).Scan(&sessionID); err != nil {
		t.Fatalf("create session: %v", err)
	}

	d.engine.Sweep()

	var state string
	if err := d.pool.QueryRow(ctx, `SELECT state FROM sessions WHERE id = $1`, sessionID).Scan(&state); err != nil {
		t.Fatalf("fetch state: %v", err)
	}
	if state != "FAILED" {
		t.Errorf("expected FAILED for stale AUTHORIZED session, got %q", state)
	}
}

