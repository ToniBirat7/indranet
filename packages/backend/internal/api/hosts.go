package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ToniBirat7/indranet/packages/backend/internal/models"
)

// ListHosts returns online host machines, sorted by rating then price.
func (h *Handlers) ListHosts(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, _ := strconv.Atoi(l); n > 0 && n <= 100 {
			limit = n
		}
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if n, _ := strconv.Atoi(p); n > 1 {
			offset = (n - 1) * limit
		}
	}

	rows, err := h.deps.Pool.Query(r.Context(), `
		SELECT id, user_id, display_name, gpu_model, vram_gb, cpu_model, ram_gb,
		       os, price_per_hour_cents, tags, online, payouts_enabled,
		       total_sessions, rating_sum, rating_count, created_at
		FROM hosts
		WHERE online = true
		ORDER BY
		  CASE WHEN rating_count > 0 THEN rating_sum::float / rating_count ELSE 0 END DESC,
		  price_per_hour_cents ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	hosts := []map[string]interface{}{}
	for rows.Next() {
		var host models.Host
		if err := rows.Scan(
			&host.ID, &host.UserID, &host.DisplayName, &host.GPUModel, &host.VRAMgb,
			&host.CPUModel, &host.RAMgb, &host.OS, &host.PricePerHourCents,
			&host.Tags, &host.Online, &host.PayoutsEnabled,
			&host.TotalSessions, &host.RatingSum, &host.RatingCount, &host.CreatedAt,
		); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		hosts = append(hosts, map[string]interface{}{
			"host_id":              host.ID,
			"display_name":         host.DisplayName,
			"gpu_model":            host.GPUModel,
			"vram_gb":              host.VRAMgb,
			"cpu_model":            host.CPUModel,
			"ram_gb":               host.RAMgb,
			"os":                   host.OS,
			"price_per_hour_cents": host.PricePerHourCents,
			"price_per_minute_cents": host.PricePerMinuteCents(),
			"tags":                 host.Tags,
			"rating":               host.Rating(),
			"total_sessions":       host.TotalSessions,
		})
	}
	if rows.Err() != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hosts": hosts,
		"total": len(hosts),
		"page":  offset/limit + 1,
	})
}

// GetHost returns detailed information about a specific host.
func (h *Handlers) GetHost(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "id")

	var host models.Host
	err := h.deps.Pool.QueryRow(r.Context(), `
		SELECT id, user_id, display_name, gpu_model, vram_gb, cpu_model, ram_gb,
		       os, price_per_hour_cents, tags, online, payouts_enabled,
		       total_sessions, rating_sum, rating_count, created_at
		FROM hosts WHERE id = $1
	`, hostID).Scan(
		&host.ID, &host.UserID, &host.DisplayName, &host.GPUModel, &host.VRAMgb,
		&host.CPUModel, &host.RAMgb, &host.OS, &host.PricePerHourCents,
		&host.Tags, &host.Online, &host.PayoutsEnabled,
		&host.TotalSessions, &host.RatingSum, &host.RatingCount, &host.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "host not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"host_id":                host.ID,
		"display_name":           host.DisplayName,
		"gpu_model":              host.GPUModel,
		"vram_gb":                host.VRAMgb,
		"cpu_model":              host.CPUModel,
		"ram_gb":                 host.RAMgb,
		"os":                     host.OS,
		"price_per_hour_cents":   host.PricePerHourCents,
		"price_per_minute_cents": host.PricePerMinuteCents(),
		"tags":                   host.Tags,
		"online":                 host.Online,
		"rating":                 host.Rating(),
		"total_sessions":         host.TotalSessions,
		"payouts_enabled":        host.PayoutsEnabled,
	})
}

// RegisterHost registers a new host machine for the authenticated user.
func (h *Handlers) RegisterHost(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxKeyUserID).(string)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		DisplayName       string   `json:"display_name"`
		GPUModel          string   `json:"gpu_model"`
		VRAMgb            int      `json:"vram_gb"`
		CPUModel          string   `json:"cpu_model"`
		RAMgb             int      `json:"ram_gb"`
		OS                string   `json:"os"`
		PricePerHourCents int64    `json:"price_per_hour_cents"`
		Tags              []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.DisplayName == "" || req.GPUModel == "" {
		http.Error(w, "display_name and gpu_model are required", http.StatusBadRequest)
		return
	}
	if req.PricePerHourCents <= 0 {
		http.Error(w, "price_per_hour_cents must be positive", http.StatusBadRequest)
		return
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}

	var hostID string
	err := h.deps.Pool.QueryRow(r.Context(), `
		INSERT INTO hosts (user_id, display_name, gpu_model, vram_gb, cpu_model,
		                   ram_gb, os, price_per_hour_cents, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, userID, req.DisplayName, req.GPUModel, req.VRAMgb, req.CPUModel,
		req.RAMgb, req.OS, req.PricePerHourCents, req.Tags,
	).Scan(&hostID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	agentToken, err := h.generateAgentJWT(hostID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"host_id":     hostID,
		"agent_token": agentToken,
	})
}
