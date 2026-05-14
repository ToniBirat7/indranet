package models

import "time"

// Payment records a completed payment event.
type Payment struct {
	ID                    string    `json:"payment_id" db:"id"`
	UserID                string    `json:"user_id" db:"user_id"`
	SessionID             string    `json:"session_id,omitempty" db:"session_id"`
	AmountCents           int64     `json:"amount_cents" db:"amount_cents"`
	Type                  string    `json:"type" db:"type"` // topup | session | refund
	StripePaymentIntentID string    `json:"-" db:"stripe_payment_intent_id"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
}
