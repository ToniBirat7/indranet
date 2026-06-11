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
	pool           *pgxpool.Pool
	rdb            *redis.Client
	hub            HubNotifier
	tickEvery      time.Duration
	warningMinutes int
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

// HubNotifier is the minimal interface the billing engine needs from the signaling hub.
type HubNotifier interface {
	SendToSession(sessionID string, message interface{})
}

// NewEngine creates a new billing engine.
func NewEngine(pool *pgxpool.Pool, rdb *redis.Client, hub HubNotifier, tickEvery time.Duration, warningMinutes int) *Engine {
	return &Engine{
		pool:           pool,
		rdb:            rdb,
		hub:            hub,
		tickEvery:      tickEvery,
		warningMinutes: warningMinutes,
		stopCh:         make(chan struct{}),
	}
}

// Run starts the billing tick loop. Call in a goroutine.
func (e *Engine) Run() {
	e.wg.Add(1)
	defer e.wg.Done()

	ticker := time.NewTicker(e.tickEvery)
	// Sweep stuck ENDING sessions and stale hosts every 2 minutes.
	sweepTicker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	defer sweepTicker.Stop()

	slog.Info("billing engine running", "tick_interval", e.tickEvery)

	for {
		select {
		case <-ticker.C:
			e.tick()
		case <-sweepTicker.C:
			e.sweep()
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

// Tick runs one billing cycle immediately. Used in integration tests.
func (e *Engine) Tick() {
	e.tick()
}

type activeSession struct {
	SessionID          string
	UserID             string
	RatePerMinuteCents int64
	BalanceCents       int64
}

func (e *Engine) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := e.pool.Query(ctx, `
		SELECT s.id, s.user_id, s.rate_per_minute_cents, u.balance_cents
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.state = 'ACTIVE'
	`)
	if err != nil {
		slog.Error("billing: failed to query active sessions", "error", err)
		return
	}
	defer rows.Close()

	var sessions []activeSession
	for rows.Next() {
		var s activeSession
		if err := rows.Scan(&s.SessionID, &s.UserID, &s.RatePerMinuteCents, &s.BalanceCents); err != nil {
			slog.Error("billing: scan error", "error", err)
			continue
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		slog.Error("billing: rows error", "error", err)
		return
	}

	slog.Debug("billing tick", "active_sessions", len(sessions))
	for _, s := range sessions {
		e.processSessionTick(ctx, s.SessionID, s.UserID, s.RatePerMinuteCents, s.BalanceCents)
	}
}

// processSessionTick executes billing for a single session inside a DB transaction.
// Atomicity guarantee: balance deduction, tick record, and session total update either
// all succeed or all roll back — no partial billing.
func (e *Engine) processSessionTick(ctx context.Context, sessionID, userID string, ratePerMinuteCents, _ int64) {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		slog.Error("billing: begin tx failed", "session_id", sessionID, "error", err)
		return
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var newBalance int64
	if err := tx.QueryRow(ctx, `
		UPDATE users SET balance_cents = balance_cents - $1, updated_at = NOW()
		WHERE id = $2
		RETURNING balance_cents
	`, ratePerMinuteCents, userID).Scan(&newBalance); err != nil {
		slog.Error("billing: deduct balance failed", "session_id", sessionID, "error", err)
		return
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO billing_ticks (session_id, amount_cents) VALUES ($1, $2)`,
		sessionID, ratePerMinuteCents,
	); err != nil {
		slog.Error("billing: insert tick failed", "session_id", sessionID, "error", err)
		return
	}

	if _, err := tx.Exec(ctx, `
		UPDATE sessions
		SET total_charged_cents = total_charged_cents + $1, updated_at = NOW()
		WHERE id = $2
	`, ratePerMinuteCents, sessionID); err != nil {
		slog.Error("billing: update session total failed", "session_id", sessionID, "error", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("billing: commit failed", "session_id", sessionID, "error", err)
		return
	}

	slog.Debug("billing tick committed",
		"session_id", sessionID,
		"charged_cents", ratePerMinuteCents,
		"new_balance_cents", newBalance,
	)

	if newBalance <= 0 {
		e.killSession(ctx, sessionID)
		return
	}

	warningCents := int64(e.warningMinutes) * ratePerMinuteCents
	if newBalance < warningCents {
		minutesRemaining := int(newBalance / ratePerMinuteCents)
		e.sendWarning(sessionID, minutesRemaining)
	}
}

// killSession transitions a session to ENDING and notifies all connected clients.
func (e *Engine) killSession(ctx context.Context, sessionID string) {
	slog.Info("billing: killing session (balance exhausted)", "session_id", sessionID)

	if _, err := e.pool.Exec(ctx, `
		UPDATE sessions SET state = 'ENDING', updated_at = NOW()
		WHERE id = $1 AND state = 'ACTIVE'
	`, sessionID); err != nil {
		slog.Error("billing: failed to set session ENDING", "session_id", sessionID, "error", err)
	}

	e.hub.SendToSession(sessionID, map[string]string{
		"type":   "session_kill",
		"reason": "balance_exhausted",
	})
}

func (e *Engine) sendWarning(sessionID string, minutesRemaining int) {
	slog.Info("billing: low balance warning",
		"session_id", sessionID,
		"minutes_remaining", minutesRemaining,
	)
	e.hub.SendToSession(sessionID, map[string]interface{}{
		"type":              "session_warning",
		"minutes_remaining": minutesRemaining,
	})
}

// sweep finalizes sessions stuck in ENDING and marks stale host agents offline.
// Runs every 2 minutes as a safety net alongside normal billing ticks.
func (e *Engine) sweep() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ENDING sessions older than 5 minutes → ENDED; increment host session count.
	tag, err := e.pool.Exec(ctx, `
		WITH ended AS (
			UPDATE sessions
			SET state = 'ENDED', ended_at = NOW(), updated_at = NOW()
			WHERE state = 'ENDING' AND updated_at < NOW() - INTERVAL '5 minutes'
			RETURNING host_id
		)
		UPDATE hosts SET total_sessions = total_sessions + 1, updated_at = NOW()
		FROM ended WHERE hosts.id = ended.host_id
	`)
	if err != nil {
		slog.Error("billing: sweep ENDING→ENDED failed", "error", err)
	} else if tag.RowsAffected() > 0 {
		slog.Info("billing: swept ENDING→ENDED", "count", tag.RowsAffected())
	}

	// Hosts whose agent hasn't sent a heartbeat in 3 minutes → offline.
	tag, err = e.pool.Exec(ctx, `
		UPDATE hosts SET online = false, updated_at = NOW()
		WHERE online = true AND updated_at < NOW() - INTERVAL '3 minutes'
	`)
	if err != nil {
		slog.Error("billing: sweep stale hosts failed", "error", err)
	} else if tag.RowsAffected() > 0 {
		slog.Info("billing: marked stale hosts offline", "count", tag.RowsAffected())
	}
}
