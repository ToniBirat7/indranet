package models

import "time"

// Host represents a registered host machine available for rent.
type Host struct {
	ID                   string    `json:"host_id" db:"id"`
	UserID               string    `json:"user_id" db:"user_id"`
	DisplayName          string    `json:"display_name" db:"display_name"`
	GPUModel             string    `json:"gpu_model" db:"gpu_model"`
	VRAMgb               int       `json:"vram_gb" db:"vram_gb"`
	CPUModel             string    `json:"cpu_model" db:"cpu_model"`
	RAMgb                int       `json:"ram_gb" db:"ram_gb"`
	OS                   string    `json:"os" db:"os"`
	PricePerHourCents    int64     `json:"price_per_hour_cents" db:"price_per_hour_cents"`
	Tags                 []string  `json:"tags" db:"tags"`
	Online               bool      `json:"online" db:"online"`
	StripeAccountID      string    `json:"stripe_account_id,omitempty" db:"stripe_account_id"`
	PayoutsEnabled       bool      `json:"payouts_enabled" db:"payouts_enabled"`
	TotalSessions        int       `json:"total_sessions" db:"total_sessions"`
	RatingSum            int       `json:"-" db:"rating_sum"`
	RatingCount          int       `json:"-" db:"rating_count"`
	MachineFingerprint   string    `json:"-" db:"machine_fingerprint"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// Rating returns the average host rating (0-5).
func (h *Host) Rating() float64 {
	if h.RatingCount == 0 {
		return 0
	}
	return float64(h.RatingSum) / float64(h.RatingCount)
}

// PricePerMinuteCents returns the per-minute billing rate.
func (h *Host) PricePerMinuteCents() int64 {
	return h.PricePerHourCents / 60
}
