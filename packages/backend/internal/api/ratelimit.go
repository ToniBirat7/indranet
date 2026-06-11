package api

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	maxReqs int
	window  time.Duration
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
	}
	go func() {
		for range time.Tick(10 * time.Minute) {
			rl.prune()
		}
	}()
	return rl
}

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

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	return r.RemoteAddr
}

// rateLimitMiddleware is a generic factory: keyFn extracts the rate-limit key from the request.
func rateLimitMiddleware(maxReqs int, window time.Duration, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	rl := newRateLimiter(maxReqs, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(keyFn(r)) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// authRateLimitMiddleware limits /auth/* requests: 20 per minute per IP.
func (h *Handlers) authRateLimitMiddleware(next http.Handler) http.Handler {
	return rateLimitMiddleware(20, time.Minute, clientIP)(next)
}

// sessionCreateRateLimitMiddleware limits POST /sessions: 10 per hour per authenticated user.
func (h *Handlers) sessionCreateRateLimitMiddleware(next http.Handler) http.Handler {
	rl := newRateLimiter(10, time.Hour)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(ctxKeyUserID).(string)
		key := userID
		if key == "" {
			key = clientIP(r) // fallback (shouldn't happen behind AuthMiddleware)
		}
		if !rl.allow(key) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// topupRateLimitMiddleware limits POST /users/me/topup: 10 per hour per authenticated user.
func (h *Handlers) topupRateLimitMiddleware(next http.Handler) http.Handler {
	rl := newRateLimiter(10, time.Hour)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(ctxKeyUserID).(string)
		key := userID
		if key == "" {
			key = clientIP(r)
		}
		if !rl.allow(key) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
