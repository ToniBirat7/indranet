package api

import (
	"encoding/json"
	"net/http"
)

// ListHosts returns available host machines with optional filtering.
func (h *Handlers) ListHosts(w http.ResponseWriter, r *http.Request) {
	// Query params: gpu_min_vram_gb, price_max_per_hour, tags, page, limit
	// TODO: Build dynamic WHERE clause from query params
	// TODO: SELECT from hosts WHERE online=true ORDER BY rating DESC, price ASC

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hosts": []interface{}{},
		"total": 0,
		"page":  1,
	})
}

// GetHost returns detailed information about a specific host.
func (h *Handlers) GetHost(w http.ResponseWriter, r *http.Request) {
	// TODO: Fetch host by ID, include rating calculation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "TODO"})
}

// RegisterHost registers a new host machine for the authenticated user.
func (h *Handlers) RegisterHost(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Insert into hosts table
	// TODO: Generate agent JWT (signed with JWT_SECRET, includes host_id claim)
	// TODO: Return host_id + agent_token

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"host_id":     "hst_TODO",
		"agent_token": "TODO_JWT",
	})
}
