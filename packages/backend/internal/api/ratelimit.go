package api

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	maxReqs  int
	window   time.Duration
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
	// Prune stale entries every 10 minutes
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

// authRateLimitMiddleware limits /auth/* requests: 20 per minute per IP.
func (h *Handlers) authRateLimitMiddleware(next http.Handler) http.Handler {
	rl := newRateLimiter(20, time.Minute)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff
		}
		if !rl.allow(ip) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
