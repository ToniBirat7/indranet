package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/ToniBirat7/indranet/packages/backend/internal/billing"
	"github.com/ToniBirat7/indranet/packages/backend/internal/config"
	"github.com/ToniBirat7/indranet/packages/backend/internal/signaling"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
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

// allowedOrigins returns the set of origins permitted for CORS and WebSocket upgrades.
func allowedOrigins(frontendURL string) []string {
	return []string{
		frontendURL,
		"http://localhost:3000",
		"http://127.0.0.1:3000",
	}
}

// corsMiddleware adds CORS headers for the web frontend.
// Allowed origins are the frontend base URL and localhost variants for dev.
func corsMiddleware(frontendURL string) func(http.Handler) http.Handler {
	allowed := allowedOrigins(frontendURL)
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

// securityHeadersMiddleware sets defensive HTTP headers on every response.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// maxBodyMiddleware caps request bodies at 1 MB to prevent large-body DoS.
// Stripe webhook bodies are typically <8KB; SDP offers are <4KB; all API
// payloads are well under 1 MB.
func maxBodyMiddleware(next http.Handler) http.Handler {
	const maxBytes = 1 << 20 // 1 MB
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next.ServeHTTP(w, r)
	})
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
	r.Use(securityHeadersMiddleware)
	r.Use(maxBodyMiddleware)

	origins := allowedOrigins(deps.Config.FrontendBaseURL)
	h := &Handlers{
		deps:      deps,
		authRL:    newRateLimiter(20, time.Minute),
		sessionRL: newRateLimiter(10, time.Hour),
		topupRL:   newRateLimiter(10, time.Hour),
		wsUpgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, o := range origins {
					if strings.EqualFold(origin, o) {
						return true
					}
				}
				return false
			},
		},
	}

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
			r.With(h.topupRateLimitMiddleware).Post("/users/me/topup", h.TopUpWallet)

			// Host management
			r.Post("/hosts/register", h.RegisterHost)
			r.Post("/hosts/me/stripe/connect", h.ConnectStripeAccount)

			// Sessions
			r.Get("/sessions", h.ListSessions)
			r.With(h.sessionCreateRateLimitMiddleware).Post("/sessions", h.CreateSession)
			r.Get("/sessions/{id}", h.GetSession)
			r.Delete("/sessions/{id}", h.EndSession)
			r.Post("/sessions/{id}/rate", h.RateSession)
		})

		// Agent-authenticated routes (host agent JWT)
		r.Group(func(r chi.Router) {
			r.Use(h.AgentAuthMiddleware)
			// Session discovery + lifecycle (agent-driven)
			r.Get("/sessions/pending", h.GetPendingSessions)
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
	deps       RouterDeps
	authRL     *rateLimiter
	sessionRL  *rateLimiter
	topupRL    *rateLimiter
	wsUpgrader websocket.Upgrader
}
