package models

import "time"

// User represents an IndraNet user account (can be a user, host, or both).
type User struct {
	ID               string    `json:"user_id" db:"id"`
	Email            string    `json:"email" db:"email"`
	PasswordHash     string    `json:"-" db:"password_hash"`
	Name             string    `json:"name" db:"name"`
	Role             string    `json:"role" db:"role"` // user | host | admin
	BalanceCents     int64     `json:"balance_cents" db:"balance_cents"`
	StripeCustomerID string    `json:"-" db:"stripe_customer_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
