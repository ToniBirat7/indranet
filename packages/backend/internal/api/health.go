package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Health returns service health status and dependency connectivity.
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	pgStatus := "ok"
	if err := h.deps.Pool.Ping(ctx); err != nil {
		pgStatus = "unavailable"
	}

	redisStatus := "ok"
	if err := h.deps.Redis.Ping(ctx).Err(); err != nil {
		redisStatus = "unavailable"
	}

	status := "ok"
	if pgStatus != "ok" || redisStatus != "ok" {
		status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   status,
		"postgres": pgStatus,
		"redis":    redisStatus,
		"version":  "0.1.0",
	})
}
