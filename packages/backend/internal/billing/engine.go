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
	stripe "github.com/stripe/stripe-go/v76"
	stripetransfer "github.com/stripe/stripe-go/v76/transfer"
)

// Engine is the billing tick loop. It runs in a dedicated goroutine.
type Engine struct {
	pool               *pgxpool.Pool
	rdb                *redis.Client
	hub                HubNotifier
	tickEvery          time.Duration
	warningMinutes     int
	stripeKey          string
	platformFeePercent int
	stopCh             chan struct{}
	wg                 sync.WaitGroup
	// warnedSessions tracks sessions that have already received a low-balance
	// warning so we don't spam the client on every subsequent tick.
	warnedSessions sync.Map
}

// HubNotifier is the minimal interface the billing engine needs from the signaling hub.
type HubNotifier interface {
	SendToSession(sessionID string, message interface{})
}

// NewEngine creates a new billing engine.
func NewEngine(pool *pgxpool.Pool, rdb *redis.Client, hub HubNotifier, tickEvery time.Duration, warningMinutes int, stripeKey string, platformFeePercent int) *Engine {
	return &Engine{
		pool:               pool,
		rdb:                rdb,
		hub:                hub,
		tickEvery:          tickEvery,
		warningMinutes:     warningMinutes,
		stripeKey:          stripeKey,
		platformFeePercent: platformFeePercent,
		stopCh:             make(chan struct{}),
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

// Sweep runs one maintenance sweep immediately. Used in integration tests.
func (e *Engine) Sweep() {
	e.sweep()
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

	// Deduct at most the user's current balance to prevent negative balance.
	// LEAST ensures the final tick never over-charges when balance < rate.
	var newBalance, actualCharge int64
	if err := tx.QueryRow(ctx, `
		WITH charge AS (
			SELECT LEAST(balance_cents, $1::BIGINT) AS amount FROM users WHERE id = $2
		)
		UPDATE users
		SET balance_cents = balance_cents - (SELECT amount FROM charge),
		    updated_at = NOW()
		WHERE id = $2
		RETURNING balance_cents, (SELECT amount FROM charge)
	`, ratePerMinuteCents, userID).Scan(&newBalance, &actualCharge); err != nil {
		slog.Error("billing: deduct balance failed", "session_id", sessionID, "error", err)
		return
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO billing_ticks (session_id, amount_cents) VALUES ($1, $2)`,
		sessionID, actualCharge,
	); err != nil {
		slog.Error("billing: insert tick failed", "session_id", sessionID, "error", err)
		return
	}

	if _, err := tx.Exec(ctx, `
		UPDATE sessions
		SET total_charged_cents = total_charged_cents + $1, updated_at = NOW()
		WHERE id = $2
	`, actualCharge, sessionID); err != nil {
		slog.Error("billing: update session total failed", "session_id", sessionID, "error", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("billing: commit failed", "session_id", sessionID, "error", err)
		return
	}

	slog.Debug("billing tick committed",
		"session_id", sessionID,
		"charged_cents", actualCharge,
		"new_balance_cents", newBalance,
	)

	if newBalance <= 0 {
		e.warnedSessions.Delete(sessionID) // clean up on kill
		e.killSession(ctx, sessionID)
		return
	}

	warningCents := int64(e.warningMinutes) * ratePerMinuteCents
	if newBalance < warningCents {
		if _, alreadyWarned := e.warnedSessions.LoadOrStore(sessionID, true); !alreadyWarned {
			minutesRemaining := int(newBalance / ratePerMinuteCents)
			e.sendWarning(sessionID, minutesRemaining)
		}
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

type endedSession struct {
	sessionID          string
	hostID             string
	totalChargedCents  int64
	stripeAccountID    string
	payoutsEnabled     bool
}

// sweep finalizes sessions stuck in ENDING and marks stale host agents offline.
// Runs every 2 minutes as a safety net alongside normal billing ticks.
func (e *Engine) sweep() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Transition ENDING→ENDED (older than 5 min), collect session data for payouts.
	rows, err := e.pool.Query(ctx, `
		WITH ended AS (
			UPDATE sessions
			SET state = 'ENDED', ended_at = NOW(), updated_at = NOW()
			WHERE state = 'ENDING' AND updated_at < NOW() - INTERVAL '5 minutes'
			RETURNING id, host_id, total_charged_cents
		),
		_ AS (
			UPDATE hosts SET total_sessions = total_sessions + 1, updated_at = NOW()
			FROM ended WHERE hosts.id = ended.host_id
		)
		SELECT e.id, e.host_id, e.total_charged_cents,
		       COALESCE(h.stripe_account_id, ''), h.payouts_enabled
		FROM ended e
		JOIN hosts h ON h.id = e.host_id
	`)
	if err != nil {
		slog.Error("billing: sweep ENDING→ENDED failed", "error", err)
	} else {
		var ended []endedSession
		for rows.Next() {
			var s endedSession
			if err := rows.Scan(&s.sessionID, &s.hostID, &s.totalChargedCents, &s.stripeAccountID, &s.payoutsEnabled); err == nil {
				ended = append(ended, s)
			}
		}
		rows.Close()
		if len(ended) > 0 {
			slog.Info("billing: swept ENDING→ENDED", "count", len(ended))
			for _, s := range ended {
				e.warnedSessions.Delete(s.sessionID)
				e.transferHostPayout(s)
			}
		}
	}

	// Hosts whose agent hasn't sent a heartbeat in 3 minutes → offline.
	if tag, err2 := e.pool.Exec(ctx, `
		UPDATE hosts SET online = false, updated_at = NOW()
		WHERE online = true AND updated_at < NOW() - INTERVAL '3 minutes'
	`); err2 != nil {
		slog.Error("billing: sweep stale hosts failed", "error", err2)
	} else if tag.RowsAffected() > 0 {
		slog.Info("billing: marked stale hosts offline", "count", tag.RowsAffected())
	}

	// CREATED sessions older than 30 minutes → FAILED (payment never completed).
	if tag2, err2 := e.pool.Exec(ctx, `
		UPDATE sessions SET state = 'FAILED', updated_at = NOW()
		WHERE state = 'CREATED' AND created_at < NOW() - INTERVAL '30 minutes'
	`); err2 != nil {
		slog.Error("billing: sweep abandoned CREATED→FAILED failed", "error", err2)
	} else if tag2.RowsAffected() > 0 {
		slog.Info("billing: swept abandoned CREATED sessions to FAILED", "count", tag2.RowsAffected())
	}

	// AUTHORIZED sessions older than 10 minutes → FAILED (host agent never came up).
	// Backstop for awaitAgentReady goroutines lost across backend restarts.
	if tag3, err3 := e.pool.Exec(ctx, `
		UPDATE sessions SET state = 'FAILED', updated_at = NOW()
		WHERE state = 'AUTHORIZED' AND updated_at < NOW() - INTERVAL '10 minutes'
	`); err3 != nil {
		slog.Error("billing: sweep stale AUTHORIZED→FAILED failed", "error", err3)
	} else if tag3.RowsAffected() > 0 {
		slog.Info("billing: swept stale AUTHORIZED sessions to FAILED", "count", tag3.RowsAffected())
	}
}

// transferHostPayout creates a Stripe Transfer sending 80% (platform keeps 20%) of
// total_charged_cents to the host's Connect account. No-op in dev (no Stripe key).
func (e *Engine) transferHostPayout(s endedSession) {
	if e.stripeKey == "" {
		return // dev mode — no real payments
	}
	if !s.payoutsEnabled || s.stripeAccountID == "" {
		slog.Warn("billing: skipping payout — host payouts not enabled",
			"session_id", s.sessionID, "host_id", s.hostID)
		return
	}
	if s.totalChargedCents <= 0 {
		return
	}

	hostPercent := int64(100 - e.platformFeePercent)
	payoutCents := s.totalChargedCents * hostPercent / 100

	stripe.Key = e.stripeKey
	params := &stripe.TransferParams{
		Amount:        stripe.Int64(payoutCents),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Destination:   stripe.String(s.stripeAccountID),
		TransferGroup: stripe.String("session_" + s.sessionID),
	}
	params.AddMetadata("session_id", s.sessionID)
	params.AddMetadata("host_id", s.hostID)

	t, err := stripetransfer.New(params)
	if err != nil {
		slog.Error("billing: Stripe transfer failed",
			"session_id", s.sessionID,
			"host_id", s.hostID,
			"amount_cents", payoutCents,
			"error", err,
		)
		return
	}
	slog.Info("billing: host payout transferred",
		"session_id", s.sessionID,
		"host_id", s.hostID,
		"transfer_id", t.ID,
		"amount_cents", payoutCents,
	)
}
