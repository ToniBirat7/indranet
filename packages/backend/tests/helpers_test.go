package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ToniBirat7/indranet/packages/backend/internal/api"
	"github.com/ToniBirat7/indranet/packages/backend/internal/billing"
	"github.com/ToniBirat7/indranet/packages/backend/internal/config"
	"github.com/ToniBirat7/indranet/packages/backend/internal/db"
	"github.com/ToniBirat7/indranet/packages/backend/internal/signaling"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testDeps holds all live dependencies for integration tests.
type testDeps struct {
	pool   *pgxpool.Pool
	router http.Handler
	cfg    *config.Config
	hub    *signaling.Hub
	engine *billing.Engine
}

// newTestDeps connects to a real postgres + redis instance.
// Skips the test if DATABASE_URL is unreachable.
func newTestDeps(t *testing.T) *testDeps {
	t.Helper()

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.ConnectPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("postgres unavailable (%v) — set DATABASE_URL to run integration tests", err)
	}
	t.Cleanup(pool.Close)

	if err := db.RunMigrations(ctx, pool); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	rdb, err := db.ConnectRedis(ctx, cfg.RedisURL)
	if err != nil {
		t.Skipf("redis unavailable (%v) — set REDIS_URL to run integration tests", err)
	}
	t.Cleanup(func() { rdb.Close() })

	hub := signaling.NewHub()
	go hub.Run()

	eng := billing.NewEngine(pool, rdb, hub, time.Minute, 5, "", 20)

	router := api.NewRouter(api.RouterDeps{
		Pool:    pool,
		Redis:   rdb,
		Hub:     hub,
		Config:  cfg,
		Billing: eng,
	})

	return &testDeps{
		pool:   pool,
		router: router,
		cfg:    cfg,
		hub:    hub,
		engine: eng,
	}
}

// cleanupTestUser removes a user and all related sessions/billing data created in a test.
func cleanupTestUser(t *testing.T, pool *pgxpool.Pool, email string) {
	t.Helper()
	ctx := context.Background()
	// Delete in dependency order
	pool.Exec(ctx, `DELETE FROM billing_ticks WHERE session_id IN (SELECT id FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email = $1))`, email)
	pool.Exec(ctx, `DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email = $1)`, email)
	pool.Exec(ctx, `DELETE FROM users WHERE email = $1`, email)
}

// cleanupTestHost removes a host record created in a test.
func cleanupTestHost(t *testing.T, pool *pgxpool.Pool, hostID string) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM billing_ticks WHERE session_id IN (SELECT id FROM sessions WHERE host_id = $1)`, hostID)
	pool.Exec(ctx, `DELETE FROM sessions WHERE host_id = $1`, hostID)
	pool.Exec(ctx, `DELETE FROM hosts WHERE id = $1`, hostID)
}
