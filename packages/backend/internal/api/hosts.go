package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	stripe "github.com/stripe/stripe-go/v76"
	stripeaccount "github.com/stripe/stripe-go/v76/account"
	stripeaccountlink "github.com/stripe/stripe-go/v76/accountlink"

	"github.com/ToniBirat7/indranet/packages/backend/internal/models"
)

// ListHosts returns online host machines, sorted by rating then price.
// Supported query params: min_vram (int GB), max_price_cents (int), online (1 = online only), page, limit.
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

	// Build optional filter clauses
	minVRAM, _ := strconv.Atoi(r.URL.Query().Get("min_vram"))
	maxPriceCents, _ := strconv.ParseInt(r.URL.Query().Get("max_price_cents"), 10, 64)
	onlineOnly := r.URL.Query().Get("online") == "1"

	args := []interface{}{}
	argIdx := 1
	filters := "TRUE"
	if onlineOnly {
		filters = "online = true"
	}
	if minVRAM > 0 {
		args = append(args, minVRAM)
		filters += " AND vram_gb >= $" + strconv.Itoa(argIdx)
		argIdx++
	}
	if maxPriceCents > 0 {
		args = append(args, maxPriceCents)
		filters += " AND price_per_hour_cents <= $" + strconv.Itoa(argIdx)
		argIdx++
	}
	// Count args are separate from LIMIT/OFFSET args
	countArgs := args[:len(args)] // same filter args, without limit/offset

	var totalCount int
	if err := h.deps.Pool.QueryRow(r.Context(),
		`SELECT COUNT(*) FROM hosts WHERE `+filters, countArgs...,
	).Scan(&totalCount); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	args = append(args, limit, offset)

	query := `
		SELECT id, user_id, display_name, gpu_model, vram_gb, cpu_model, ram_gb,
		       os, price_per_hour_cents, tags, online, payouts_enabled,
		       total_sessions, rating_sum, rating_count, created_at
		FROM hosts
		WHERE ` + filters + `
		ORDER BY
		  CASE WHEN rating_count > 0 THEN rating_sum::float / rating_count ELSE 0 END DESC,
		  price_per_hour_cents ASC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)

	rows, err := h.deps.Pool.Query(r.Context(), query, args...)
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
			"rating":                 host.Rating(),
			"total_sessions":         host.TotalSessions,
		})
	}
	if rows.Err() != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hosts": hosts,
		"total": totalCount,
		"page":  offset/limit + 1,
		"limit": limit,
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
	if len(req.DisplayName) > 80 || len(req.GPUModel) > 80 || len(req.CPUModel) > 80 || len(req.OS) > 80 {
		http.Error(w, "string fields must be ≤80 characters", http.StatusBadRequest)
		return
	}
	if req.PricePerHourCents <= 0 || req.PricePerHourCents > 100_000 { // max $1000/hr
		http.Error(w, "price_per_hour_cents must be between 1 and 100000", http.StatusBadRequest)
		return
	}
	if req.VRAMgb < 0 || req.VRAMgb > 1024 || req.RAMgb < 0 || req.RAMgb > 4096 {
		http.Error(w, "vram_gb and ram_gb out of range", http.StatusBadRequest)
		return
	}
	if len(req.Tags) > 20 {
		http.Error(w, "too many tags (max 20)", http.StatusBadRequest)
		return
	}
	for _, tag := range req.Tags {
		if len(tag) > 40 {
			http.Error(w, "each tag must be ≤40 characters", http.StatusBadRequest)
			return
		}
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

// SetHostOnline marks a host as online or offline. Called by the host agent on startup/shutdown.
// Uses agent JWT auth (hostID is extracted from the token claim, not a URL param).
func (h *Handlers) SetHostOnline(w http.ResponseWriter, r *http.Request) {
	hostID, _ := r.Context().Value(ctxKeyUserID).(string)
	if hostID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Online bool `json:"online"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tag, err := h.deps.Pool.Exec(r.Context(), `
		UPDATE hosts SET online = $1, updated_at = NOW() WHERE id = $2
	`, req.Online, hostID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "host not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"online": req.Online})
}

// HostHeartbeat records agent activity and returns the host's current status.
// The agent should call this every 60 seconds while running. The billing sweep
// marks hosts offline after 3 minutes without a heartbeat.
func (h *Handlers) HostHeartbeat(w http.ResponseWriter, r *http.Request) {
	hostID, _ := r.Context().Value(ctxKeyUserID).(string)
	if hostID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.deps.Pool.Exec(r.Context(), `
		UPDATE hosts SET online = true, updated_at = NOW() WHERE id = $1
	`, hostID); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ConnectStripeAccount creates or retrieves a Stripe Connect Express account for the host
// and returns an onboarding URL. After completion, Stripe sends account.updated webhook.
// POST /v1/hosts/me/stripe/connect
func (h *Handlers) ConnectStripeAccount(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(ctxKeyUserID).(string)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if h.deps.Config.StripeSecretKey == "" {
		http.Error(w, "stripe not configured", http.StatusNotImplemented)
		return
	}

	// Fetch host for this user
	var hostID, stripeAccountID string
	err := h.deps.Pool.QueryRow(r.Context(), `
		SELECT id, COALESCE(stripe_account_id, '') FROM hosts WHERE user_id = $1 LIMIT 1
	`, userID).Scan(&hostID, &stripeAccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "no host found for this user", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	stripe.Key = h.deps.Config.StripeSecretKey

	// Create a new Express account if the host doesn't have one yet
	if stripeAccountID == "" {
		acct, err := stripeaccount.New(&stripe.AccountParams{
			Type: stripe.String(string(stripe.AccountTypeExpress)),
			Capabilities: &stripe.AccountCapabilitiesParams{
				Transfers: &stripe.AccountCapabilitiesTransfersParams{
					Requested: stripe.Bool(true),
				},
			},
			BusinessType: stripe.String("individual"),
			Metadata:     map[string]string{"indranet_host_id": hostID},
		})
		if err != nil {
			slog.Error("stripe: create Connect account failed", "host_id", hostID, "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		stripeAccountID = acct.ID
		if _, err := h.deps.Pool.Exec(r.Context(), `
			UPDATE hosts SET stripe_account_id = $1, updated_at = NOW() WHERE id = $2
		`, stripeAccountID, hostID); err != nil {
			slog.Error("stripe: failed to store account_id", "host_id", hostID, "error", err)
		}
	}

	// Create onboarding link
	link, err := stripeaccountlink.New(&stripe.AccountLinkParams{
		Account:    stripe.String(stripeAccountID),
		RefreshURL: stripe.String(fmt.Sprintf("%s/dashboard/host?stripe=refresh", h.deps.Config.FrontendBaseURL)),
		ReturnURL:  stripe.String(fmt.Sprintf("%s/dashboard/host?stripe=success", h.deps.Config.FrontendBaseURL)),
		Type:       stripe.String("account_onboarding"),
	})
	if err != nil {
		slog.Error("stripe: create account link failed", "account_id", stripeAccountID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"onboarding_url": link.URL})
}
