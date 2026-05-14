// Package billing implements the per-session billing tick engine.
// Every N seconds, the engine queries all ACTIVE sessions, deducts the per-minute
// rate from each user's balance, and emits events when balance is low or exhausted.
package billing

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Engine is the billing tick loop. It runs in a dedicated goroutine.
type Engine struct {
	pool      *pgxpool.Pool
	rdb       *redis.Client
	hub       HubNotifier // interface to send WebSocket events to session clients
	tickEvery time.Duration
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// HubNotifier is the minimal interface the billing engine needs from the signaling hub.
type HubNotifier interface {
	SendToSession(sessionID string, message interface{})
}

// NewEngine creates a new billing engine.
// tickEvery is the interval between billing ticks (typically 60s in production, shorter in tests).
func NewEngine(pool *pgxpool.Pool, rdb *redis.Client, hub HubNotifier) *Engine {
	return &Engine{
		pool:      pool,
		rdb:       rdb,
		hub:       hub,
		tickEvery: 60 * time.Second,
		stopCh:    make(chan struct{}),
	}
}

// Run starts the billing tick loop. Call in a goroutine.
func (e *Engine) Run() {
	e.wg.Add(1)
	defer e.wg.Done()

	ticker := time.NewTicker(e.tickEvery)
	defer ticker.Stop()

	slog.Info("billing engine running", "tick_interval", e.tickEvery)

	for {
		select {
		case <-ticker.C:
			e.tick()
		case <-e.stopCh:
			slog.Info("billing engine stopped")
			return
		}
	}
}

// Stop signals the engine to stop and waits for it to exit.
func (e *Engine) Stop() {
	close(e.stopCh)
	e.wg.Wait()
}

// tick processes one billing cycle: find all ACTIVE sessions and deduct their rate.
func (e *Engine) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// TODO: Query all ACTIVE sessions with their rate and user balance
	// SELECT s.id, s.user_id, s.rate_per_minute_cents, u.balance_cents
	// FROM sessions s
	// JOIN users u ON u.id = s.user_id
	// WHERE s.state = 'ACTIVE'

	type activeSession struct {
		SessionID          string
		UserID             string
		RatePerMinuteCents int64
		BalanceCents       int64
	}

	// TODO: Replace with real DB query
	var sessions []activeSession
	_ = ctx

	for _, s := range sessions {
		e.processSessionTick(ctx, s.SessionID, s.UserID, s.RatePerMinuteCents, s.BalanceCents)
	}
}

// processSessionTick handles the billing logic for a single session tick.
// This is the core billing invariant — no side effects if the DB update fails.
func (e *Engine) processSessionTick(ctx context.Context, sessionID, userID string, ratePerMinuteCents, currentBalanceCents int64) {
	newBalance := currentBalanceCents - ratePerMinuteCents

	// TODO: Execute in a DB transaction:
	// 1. Deduct ratePerMinuteCents from user.balance_cents (UPDATE users SET balance_cents = balance_cents - $1 WHERE id = $2)
	// 2. Insert billing_tick record (INSERT INTO billing_ticks ...)
	// 3. Update session.total_charged_cents (UPDATE sessions SET total_charged_cents = total_charged_cents + $1)

	if newBalance <= 0 {
		// Balance exhausted — kill the session
		e.killSession(ctx, sessionID)
		return
	}

	// Check if warning threshold reached
	warningThresholdMinutes := 5
	warningThresholdCents := int64(warningThresholdMinutes) * ratePerMinuteCents
	if newBalance < warningThresholdCents {
		minutesRemaining := int(newBalance / ratePerMinuteCents)
		e.sendWarning(sessionID, minutesRemaining)
	}

	slog.Debug("billing tick",
		"session_id", sessionID,
		"charged_cents", ratePerMinuteCents,
		"new_balance_cents", newBalance,
	)
}

// killSession transitions a session to ENDING state and notifies the host agent.
func (e *Engine) killSession(ctx context.Context, sessionID string) {
	slog.Info("billing: killing session (balance exhausted)", "session_id", sessionID)

	// TODO: UPDATE sessions SET state = 'ENDING', updated_at = NOW() WHERE id = $1

	// Notify session participants via WebSocket
	e.hub.SendToSession(sessionID, map[string]string{
		"type": "session_kill",
		"reason": "balance_exhausted",
	})
}

// sendWarning sends a low-balance warning to the user client.
func (e *Engine) sendWarning(sessionID string, minutesRemaining int) {
	slog.Info("billing: sending low balance warning",
		"session_id", sessionID,
		"minutes_remaining", minutesRemaining,
	)

	e.hub.SendToSession(sessionID, map[string]interface{}{
		"type":              "session_warning",
		"minutes_remaining": minutesRemaining,
	})
}
