package api

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	maxReqs int
	window  time.Duration
	done    chan struct{}
}

type bucket struct {
	count     int
	windowEnd time.Time
}

func newRateLimiter(maxReqs int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		buckets: make(map[string]*bucket),
		maxReqs: maxReqs,
		window:  window,
		done:    make(chan struct{}),
	}
	go func() {
		t := time.NewTicker(10 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				rl.prune()
			case <-rl.done:
				return
			}
		}
	}()
	return rl
}

func (rl *rateLimiter) stop() { close(rl.done) }

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok || now.After(b.windowEnd) {
		rl.buckets[key] = &bucket{count: 1, windowEnd: now.Add(rl.window)}
		return true
	}
	if b.count >= rl.maxReqs {
		return false
	}
	b.count++
	return true
}

func (rl *rateLimiter) prune() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for k, b := range rl.buckets {
		if now.After(b.windowEnd) {
			delete(rl.buckets, k)
		}
	}
}

// clientIP returns the client's IP without the port suffix.
// chi's RealIP middleware may have already set r.RemoteAddr to the forwarded IP
// (no port), but for direct connections it remains "IP:PORT". We always strip
// the port so that all requests from the same IP share a rate limit bucket.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // already bare IP (set by chi RealIP from header)
	}
	return host
}

// authRateLimitMiddleware limits /auth/* requests: 20 per minute per IP.
// Uses a shared limiter stored on Handlers so all auth routes share state.
func (h *Handlers) authRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.authRL.allow(clientIP(r)) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// sessionCreateRateLimitMiddleware limits POST /sessions: 10 per hour per authenticated user.
func (h *Handlers) sessionCreateRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(ctxKeyUserID).(string)
		key := userID
		if key == "" {
			key = clientIP(r)
		}
		if !h.sessionRL.allow(key) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// topupRateLimitMiddleware limits POST /users/me/topup: 10 per hour per authenticated user.
func (h *Handlers) topupRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(ctxKeyUserID).(string)
		key := userID
		if key == "" {
			key = clientIP(r)
		}
		if !h.topupRL.allow(key) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
