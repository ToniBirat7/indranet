package api

import (
	"net/http"
	"strings"

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

// corsMiddleware adds CORS headers for the web frontend.
// Allowed origins are the frontend base URL and localhost variants for dev.
func corsMiddleware(frontendURL string) func(http.Handler) http.Handler {
	allowed := []string{
		frontendURL,
		"http://localhost:3000",
		"http://127.0.0.1:3000",
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, o := range allowed {
				if strings.EqualFold(origin, o) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
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
	r.Use(corsMiddleware(deps.Config.FrontendBaseURL))

	h := &Handlers{deps: deps}

	// ─── Public routes ────────────────────────────────────────────────────────
	r.Get("/health", h.Health)

	r.Route("/v1", func(r chi.Router) {
		// Auth (rate-limited: 20 req/min per IP)
		r.Group(func(r chi.Router) {
			r.Use(h.authRateLimitMiddleware)
			r.Post("/auth/register", h.Register)
			r.Post("/auth/login", h.Login)
		})

		// Hosts (browse — no auth)
		r.Get("/hosts", h.ListHosts)
		r.Get("/hosts/{id}", h.GetHost)

		// ─── Authenticated routes ─────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)

			// User profile and wallet
			r.Get("/users/me", h.GetMe)
			r.Post("/users/me/topup", h.TopUpWallet)

			// Host management
			r.Post("/hosts/register", h.RegisterHost)
			r.Post("/hosts/me/stripe/connect", h.ConnectStripeAccount)

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
