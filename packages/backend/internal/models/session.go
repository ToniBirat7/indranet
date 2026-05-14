package models

import "time"

// SessionState represents the current state in the session lifecycle.
// Transitions: CREATED → AUTHORIZED → ACTIVE → ENDING → ENDED
//              CREATED → FAILED  (payment timeout or failure)
//              AUTHORIZED → FAILED  (agent timeout)
//              ACTIVE → FAILED  (agent heartbeat timeout)
//              ENDING → FAILED  (sandbox teardown timeout)
type SessionState string

const (
	SessionStateCreated    SessionState = "CREATED"
	SessionStateAuthorized SessionState = "AUTHORIZED"
	SessionStateActive     SessionState = "ACTIVE"
	SessionStateEnding     SessionState = "ENDING"
	SessionStateEnded      SessionState = "ENDED"
	SessionStateFailed     SessionState = "FAILED"
)

// Session represents an active or historical compute session.
type Session struct {
	ID                   string       `json:"session_id" db:"id"`
	UserID               string       `json:"user_id" db:"user_id"`
	HostID               string       `json:"host_id" db:"host_id"`
	State                SessionState `json:"state" db:"state"`
	RatePerMinuteCents   int64        `json:"rate_per_minute_cents" db:"rate_per_minute_cents"`
	PreAuthMinutes       int          `json:"pre_auth_minutes" db:"pre_auth_minutes"`
	TotalChargedCents    int64        `json:"total_charged_cents" db:"total_charged_cents"`
	StripeCheckoutID     string       `json:"stripe_checkout_id,omitempty" db:"stripe_checkout_id"`
	StripePaymentIntentID string      `json:"stripe_payment_intent_id,omitempty" db:"stripe_payment_intent_id"`
	StartedAt            *time.Time   `json:"started_at,omitempty" db:"started_at"`
	EndedAt              *time.Time   `json:"ended_at,omitempty" db:"ended_at"`
	CreatedAt            time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at" db:"updated_at"`
}
