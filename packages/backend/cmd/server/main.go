// IndraNet Backend Server
// Entry point: reads config, connects to dependencies, registers routes, starts HTTP server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ToniBirat7/indranet/packages/backend/internal/api"
	"github.com/ToniBirat7/indranet/packages/backend/internal/billing"
	"github.com/ToniBirat7/indranet/packages/backend/internal/config"
	"github.com/ToniBirat7/indranet/packages/backend/internal/db"
	"github.com/ToniBirat7/indranet/packages/backend/internal/signaling"
)

func main() {
	// ─── Config ──────────────────────────────────────────────────────────────
	cfg := config.Load()

	// ─── Logger ──────────────────────────────────────────────────────────────
	logLevel := slog.LevelInfo
	if cfg.Env == "development" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Info("IndraNet backend starting",
		"env", cfg.Env,
		"version", "0.1.0",
	)

	// ─── Database ────────────────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.ConnectPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("postgres connected")

	if err := db.RunMigrations(ctx, pool); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	rdb, err := db.ConnectRedis(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	slog.Info("redis connected")

	// ─── Startup reset ───────────────────────────────────────────────────────
	// Mark all hosts offline so agents must re-send heartbeats after a backend
	// restart. Without this, stale online=true persists until the 3-minute
	// heartbeat sweep, creating a window where unavailable hosts appear bookable.
	if _, err := pool.Exec(ctx, `UPDATE hosts SET online = false, updated_at = NOW() WHERE online = true`); err != nil {
		slog.Warn("startup: failed to reset online hosts", "error", err)
	} else {
		slog.Info("startup: all hosts marked offline (agents will re-register)")
	}

	// ─── Services ────────────────────────────────────────────────────────────
	hub := signaling.NewHub()
	go hub.Run()
	slog.Info("signaling hub started")

	billingEngine := billing.NewEngine(
		pool, rdb, hub,
		time.Duration(cfg.BillingTickSeconds)*time.Second,
		cfg.SessionWarningMinutes,
		cfg.StripeSecretKey,
		cfg.StripePlatformFeePercent,
	)
	go billingEngine.Run()
	slog.Info("billing engine started")

	// ─── HTTP Router ─────────────────────────────────────────────────────────
	router := api.NewRouter(api.RouterDeps{
		Pool:    pool,
		Redis:   rdb,
		Hub:     hub,
		Config:  cfg,
		Billing: billingEngine,
	})

	// ─── HTTP Server ─────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	go func() {
		slog.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// ─── Graceful Shutdown ───────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received")

	// Drain in-flight HTTP requests first, then stop background services.
	// Stopping the hub before srv.Shutdown() would deadlock any in-flight
	// handler that calls hub.SendToSession on a full/closed broadcast channel.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced shutdown", "error", err)
	}

	billingEngine.Stop()
	hub.Stop()

	slog.Info("server exited cleanly")
}
