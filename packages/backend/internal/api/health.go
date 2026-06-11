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
	httpStatus := http.StatusOK
	if pgStatus != "ok" || redisStatus != "ok" {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	poolStat := h.deps.Pool.Stat()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   status,
		"postgres": pgStatus,
		"redis":    redisStatus,
		"version":  "0.1.0",
		"pool": map[string]int32{
			"total_conns":      poolStat.TotalConns(),
			"idle_conns":       poolStat.IdleConns(),
			"acquired_conns":   poolStat.AcquiredConns(),
			"constructing_conns": poolStat.ConstructingConns(),
			"max_conns":        poolStat.MaxConns(),
		},
	})
}
