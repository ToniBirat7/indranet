package api

import (
	"net/http"

	"github.com/ToniBirat7/indranet/packages/backend/internal/billing"
	"github.com/ToniBirat7/indranet/packages/backend/internal/config"
	"github.com/ToniBirat7/indranet/packages/backend/internal/signaling"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// RouterDeps holds all dependencies injected into HTTP handlers.
type RouterDeps struct {
	Pool    *pgxpool.Pool
	Redis   *redis.Client
	Hub     *signaling.Hub
	Config  *config.Config
	Billing *billing.Engine
}

// NewRouter creates and configures the HTTP router with all API routes.
func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * 1000000000)) // 60s

	h := &Handlers{deps: deps}

	// ─── Public routes ────────────────────────────────────────────────────────
	r.Get("/health", h.Health)

	r.Route("/v1", func(r chi.Router) {
		// Auth
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)

		// Hosts (browse — no auth)
		r.Get("/hosts", h.ListHosts)
		r.Get("/hosts/{id}", h.GetHost)

		// ─── Authenticated routes ─────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)

			// User profile
			r.Get("/users/me", h.GetMe)

			// Host management
			r.Post("/hosts/register", h.RegisterHost)

			// Sessions
			r.Get("/sessions", h.ListSessions)
			r.Post("/sessions", h.CreateSession)
			r.Get("/sessions/{id}", h.GetSession)
			r.Delete("/sessions/{id}", h.EndSession)
		})

		// Agent-authenticated routes (host agent JWT)
		r.Group(func(r chi.Router) {
			r.Use(h.AgentAuthMiddleware)
			// Session lifecycle (agent-driven)
			r.Put("/sessions/{id}/start", h.StartSession)
			r.Put("/sessions/{id}/heartbeat", h.HeartbeatSession)
			// Host status management
			r.Put("/hosts/me/online", h.SetHostOnline)
			r.Put("/hosts/me/heartbeat", h.HostHeartbeat)
		})

		// WebSocket signaling
		r.Get("/signal/{sessionID}", h.Signal)

		// Stripe webhooks (no JWT — verified via Stripe-Signature header)
		r.Post("/webhooks/stripe", h.StripeWebhook)
	})

	return r
}

// Handlers groups all HTTP handler methods with their shared dependencies.
type Handlers struct {
	deps RouterDeps
}
