package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Env            string
	Port           int
	DatabaseURL    string
	RedisURL       string
	NATSURL        string
	JWTSecret      string
	JWTExpiryHours int
	StripeSecretKey          string
	StripeWebhookSecret      string
	StripePlatformFeePercent int
	BillingTickSeconds       int
	SessionWarningMinutes    int
	FrontendBaseURL          string
}

func Load() *Config {
	cfg := &Config{
		Env:                    getEnv("INDRANET_ENV", "development"),
		Port:                   getEnvInt("INDRANET_PORT", 8080),
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://indranet:changeme@localhost:5432/indranet?sslmode=disable"),
		RedisURL:               getEnv("REDIS_URL", "redis://localhost:6379"),
		NATSURL:                getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret:              getEnv("JWT_SECRET", ""),
		JWTExpiryHours:         getEnvInt("JWT_EXPIRY_HOURS", 24),
		StripeSecretKey:        getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:    getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripePlatformFeePercent: getEnvInt("STRIPE_PLATFORM_FEE_PERCENT", 20),
		BillingTickSeconds:       getEnvInt("SESSION_BILLING_TICK_SECONDS", 60),
		SessionWarningMinutes:    getEnvInt("SESSION_WARNING_MINUTES_REMAINING", 5),
		FrontendBaseURL:          getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
	}

	if cfg.JWTSecret == "" {
		if cfg.Env != "development" {
			slog.Error("JWT_SECRET is required in non-development environments")
			os.Exit(1)
		}
		slog.Warn("JWT_SECRET not set — using insecure default (dev only)")
		cfg.JWTSecret = "dev-secret-do-not-use-in-production"
	}

	if cfg.StripeSecretKey == "" && cfg.Env != "development" {
		slog.Error("STRIPE_SECRET_KEY is required in non-development environments")
		os.Exit(1)
	}

	if cfg.StripePlatformFeePercent < 0 || cfg.StripePlatformFeePercent > 100 {
		slog.Error("STRIPE_PLATFORM_FEE_PERCENT must be between 0 and 100",
			"value", cfg.StripePlatformFeePercent)
		os.Exit(1)
	}

	if cfg.BillingTickSeconds < 1 {
		slog.Warn("SESSION_BILLING_TICK_SECONDS must be ≥1, defaulting to 60",
			"value", cfg.BillingTickSeconds)
		cfg.BillingTickSeconds = 60
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("invalid int env var, using default", "key", key, "value", v, "default", fallback)
		return fallback
	}
	return n
}
