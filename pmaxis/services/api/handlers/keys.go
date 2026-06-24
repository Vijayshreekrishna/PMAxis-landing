package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pmaxis/pmaxis/services/api/middleware"
)

// APIKey represents a developer API key stored in Postgres and cached in Redis.
type APIKey struct {
	Key       string    `json:"key"`
	AppName   string    `json:"app_name"`
	Email     string    `json:"email"`
	Tier      string    `json:"tier"`
	RateLimit int       `json:"rate_limit"` // 0 = use tier default
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type keyMeta struct {
	Tier      string `json:"tier"`
	Active    bool   `json:"active"`
	RateLimit int    `json:"rate_limit"`
}

var tierDefaults = map[string]int{
	"free":       60,
	"pro":        600,
	"enterprise": 6000,
}

// MigrateAPIKeys creates the api_keys table if it does not exist.
func (h *APIHandler) MigrateAPIKeys(ctx context.Context) error {
	_, err := h.Postgres.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS api_keys (
			key        TEXT PRIMARY KEY,
			app_name   TEXT NOT NULL,
			email      TEXT NOT NULL,
			tier       TEXT NOT NULL DEFAULT 'free',
			rate_limit INTEGER NOT NULL DEFAULT 0,
			active     BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	return err
}

func generateKey() string {
	b := make([]byte, 18) // 18 bytes → 36 hex chars
	rand.Read(b)
	return "pmx_live_" + hex.EncodeToString(b)
}

func maskKey(key string) string {
	if len(key) <= 16 {
		return key
	}
	return key[:12] + "***" + key[len(key)-4:]
}

func (h *APIHandler) syncKeyToRedis(ctx context.Context, k *APIKey) error {
	meta := keyMeta{Tier: k.Tier, Active: k.Active, RateLimit: k.RateLimit}
	val, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return h.Redis.Set(ctx, "apikey:"+k.Key, string(val), 0).Err()
}

func (h *APIHandler) insertKeyToPostgres(ctx context.Context, k *APIKey) error {
	_, err := h.Postgres.Exec(ctx,
		`INSERT INTO api_keys (key, app_name, email, tier, rate_limit, active, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		k.Key, k.AppName, k.Email, k.Tier, k.RateLimit, k.Active, k.CreatedAt,
	)
	return err
}

func (h *APIHandler) scanKey(row interface {
	Scan(dest ...any) error
}) (*APIKey, error) {
	var k APIKey
	err := row.Scan(&k.Key, &k.AppName, &k.Email, &k.Tier, &k.RateLimit, &k.Active, &k.CreatedAt)
	return &k, err
}

// RegisterKey is the public endpoint — developer submits app name + email, receives a free tier key.
func (h *APIHandler) RegisterKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppName string `json:"app_name"`
		Email   string `json:"email"`
		UseCase string `json:"use_case"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	req.AppName = strings.TrimSpace(req.AppName)
	req.Email = strings.TrimSpace(req.Email)
	if req.AppName == "" || req.Email == "" {
		jsonResp(w, http.StatusBadRequest, map[string]string{"error": "app_name and email are required"})
		return
	}
	if !strings.Contains(req.Email, "@") {
		jsonResp(w, http.StatusBadRequest, map[string]string{"error": "invalid email address"})
		return
	}

	k := &APIKey{
		Key:       generateKey(),
		AppName:   req.AppName,
		Email:     req.Email,
		Tier:      "free",
		RateLimit: 0,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.insertKeyToPostgres(r.Context(), k); err != nil {
		h.Logger.Error("failed to save key to postgres", "error", err)
		jsonResp(w, http.StatusInternalServerError, map[string]string{"error": "failed to create key"})
		return
	}
	if err := h.syncKeyToRedis(r.Context(), k); err != nil {
		h.Logger.Warn("failed to cache key in redis", "error", err)
	}

	h.Logger.Info("new API key registered", "app", k.AppName, "email", k.Email)

	jsonResp(w, http.StatusCreated, map[string]any{
		"key":        k.Key,
		"app_name":   k.AppName,
		"tier":       k.Tier,
		"rate_limit": "60 req/min",
		"message":    "Save your key now — it will not be shown again",
	})
}

// AdminCreateKey creates a key for any tier (admin only, gated by DebugAuth).
func (h *APIHandler) AdminCreateKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppName   string `json:"app_name"`
		Email     string `json:"email"`
		Tier      string `json:"tier"`
		RateLimit int    `json:"rate_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.AppName == "" || req.Email == "" {
		jsonResp(w, http.StatusBadRequest, map[string]string{"error": "app_name and email are required"})
		return
	}
	if req.Tier == "" {
		req.Tier = "free"
	}

	k := &APIKey{
		Key:       generateKey(),
		AppName:   req.AppName,
		Email:     req.Email,
		Tier:      req.Tier,
		RateLimit: req.RateLimit,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.insertKeyToPostgres(r.Context(), k); err != nil {
		h.Logger.Error("failed to save key", "error", err)
		jsonResp(w, http.StatusInternalServerError, map[string]string{"error": "failed to create key"})
		return
	}
	if err := h.syncKeyToRedis(r.Context(), k); err != nil {
		h.Logger.Warn("failed to cache key in redis", "error", err)
	}

	jsonResp(w, http.StatusCreated, k)
}

// AdminListKeys returns all API keys ordered by creation date (admin only).
func (h *APIHandler) AdminListKeys(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Postgres.Query(r.Context(),
		`SELECT key, app_name, email, tier, rate_limit, active, created_at FROM api_keys ORDER BY created_at DESC`,
	)
	if err != nil {
		h.Logger.Error("failed to list keys", "error", err)
		jsonResp(w, http.StatusInternalServerError, map[string]string{"error": "failed to list keys"})
		return
	}
	defer rows.Close()

	keys := make([]APIKey, 0)
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.Key, &k.AppName, &k.Email, &k.Tier, &k.RateLimit, &k.Active, &k.CreatedAt); err != nil {
			continue
		}
		keys = append(keys, k)
	}
	jsonResp(w, http.StatusOK, keys)
}

// AdminGetKey returns a single key's full details (admin only).
func (h *APIHandler) AdminGetKey(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	row := h.Postgres.QueryRow(r.Context(),
		`SELECT key, app_name, email, tier, rate_limit, active, created_at FROM api_keys WHERE key=$1`, key,
	)
	var k APIKey
	if err := row.Scan(&k.Key, &k.AppName, &k.Email, &k.Tier, &k.RateLimit, &k.Active, &k.CreatedAt); err != nil {
		jsonResp(w, http.StatusNotFound, map[string]string{"error": "key not found"})
		return
	}
	jsonResp(w, http.StatusOK, k)
}

// AdminUpdateKey updates tier, rate_limit, and/or active status (admin only).
func (h *APIHandler) AdminUpdateKey(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]

	var req struct {
		Tier      *string `json:"tier"`
		RateLimit *int    `json:"rate_limit"`
		Active    *bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var k APIKey
	row := h.Postgres.QueryRow(r.Context(),
		`SELECT key, app_name, email, tier, rate_limit, active, created_at FROM api_keys WHERE key=$1`, key,
	)
	if err := row.Scan(&k.Key, &k.AppName, &k.Email, &k.Tier, &k.RateLimit, &k.Active, &k.CreatedAt); err != nil {
		jsonResp(w, http.StatusNotFound, map[string]string{"error": "key not found"})
		return
	}

	if req.Tier != nil {
		k.Tier = *req.Tier
	}
	if req.RateLimit != nil {
		k.RateLimit = *req.RateLimit
	}
	if req.Active != nil {
		k.Active = *req.Active
	}

	if _, err := h.Postgres.Exec(r.Context(),
		`UPDATE api_keys SET tier=$1, rate_limit=$2, active=$3 WHERE key=$4`,
		k.Tier, k.RateLimit, k.Active, key,
	); err != nil {
		h.Logger.Error("failed to update key", "error", err)
		jsonResp(w, http.StatusInternalServerError, map[string]string{"error": "failed to update key"})
		return
	}

	if err := h.syncKeyToRedis(r.Context(), &k); err != nil {
		h.Logger.Warn("failed to sync key to redis", "error", err)
	}

	jsonResp(w, http.StatusOK, k)
}

// AdminRevokeKey sets active=false immediately (admin only).
func (h *APIHandler) AdminRevokeKey(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]

	if _, err := h.Postgres.Exec(r.Context(), `UPDATE api_keys SET active=false WHERE key=$1`, key); err != nil {
		jsonResp(w, http.StatusInternalServerError, map[string]string{"error": "failed to revoke key"})
		return
	}

	// Reflect revocation in Redis immediately so the next request is blocked
	val, _ := h.Redis.Get(r.Context(), "apikey:"+key).Result()
	var meta keyMeta
	if val != "" {
		json.Unmarshal([]byte(val), &meta)
	}
	meta.Active = false
	if b, err := json.Marshal(meta); err == nil {
		h.Redis.Set(r.Context(), "apikey:"+key, string(b), 0)
	}

	jsonResp(w, http.StatusOK, map[string]string{"message": "key revoked"})
}

// AdminActivateKey sets active=true (admin only).
func (h *APIHandler) AdminActivateKey(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]

	if _, err := h.Postgres.Exec(r.Context(), `UPDATE api_keys SET active=true WHERE key=$1`, key); err != nil {
		jsonResp(w, http.StatusInternalServerError, map[string]string{"error": "failed to activate key"})
		return
	}

	val, _ := h.Redis.Get(r.Context(), "apikey:"+key).Result()
	var meta keyMeta
	if val != "" {
		json.Unmarshal([]byte(val), &meta)
	}
	meta.Active = true
	if b, err := json.Marshal(meta); err == nil {
		h.Redis.Set(r.Context(), "apikey:"+key, string(b), 0)
	}

	jsonResp(w, http.StatusOK, map[string]string{"message": "key activated"})
}

// AdminDeleteKey permanently deletes a revoked key from Postgres and Redis (admin only).
// Only keys with active=false can be deleted as a safety guard.
func (h *APIHandler) AdminDeleteKey(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	ctx := r.Context()

	var active bool
	if err := h.Postgres.QueryRow(ctx, `SELECT active FROM api_keys WHERE key=$1`, key).Scan(&active); err != nil {
		jsonResp(w, http.StatusNotFound, map[string]string{"error": "key not found"})
		return
	}
	if active {
		jsonResp(w, http.StatusBadRequest, map[string]string{"error": "key must be revoked before deletion"})
		return
	}

	if _, err := h.Postgres.Exec(ctx, `DELETE FROM api_keys WHERE key=$1`, key); err != nil {
		h.Logger.Error("failed to delete key", "error", err)
		jsonResp(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete key"})
		return
	}

	h.Redis.Del(ctx, "apikey:"+key)
	h.Redis.Del(ctx, "rl:"+key)
	h.Redis.Del(ctx, "usage:total:"+key)

	h.Logger.Info("API key deleted", "key", maskKey(key))
	jsonResp(w, http.StatusOK, map[string]string{"message": "key deleted"})
}

// AdminResetUsage clears the rate limit counter for a key in Redis (admin only).
func (h *APIHandler) AdminResetUsage(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	h.Redis.Del(r.Context(), fmt.Sprintf("rl:%s", key))
	jsonResp(w, http.StatusOK, map[string]string{"message": "usage reset"})
}

// AdminGetUsage returns current usage stats for a key (admin only).
func (h *APIHandler) AdminGetUsage(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	ctx := r.Context()

	// Current rate-limit window counter (resets every minute)
	countStr, _ := h.Redis.Get(ctx, "rl:"+key).Result()
	count := 0
	fmt.Sscanf(countStr, "%d", &count)

	// Lifetime total requests
	totalStr, _ := h.Redis.Get(ctx, "usage:total:"+key).Result()
	total := 0
	fmt.Sscanf(totalStr, "%d", &total)

	// Last 7 days daily breakdown
	daily := make(map[string]int, 7)
	now := time.Now().UTC()
	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, -i).Format("20060102")
		dStr, _ := h.Redis.Get(ctx, "usage:daily:"+key+":"+date).Result()
		d := 0
		fmt.Sscanf(dStr, "%d", &d)
		daily[date] = d
	}

	var tier string
	var customLimit int
	h.Postgres.QueryRow(ctx, `SELECT tier, rate_limit FROM api_keys WHERE key=$1`, key).Scan(&tier, &customLimit)

	limit := tierDefaults[tier]
	if limit == 0 {
		limit = 60
	}
	if customLimit > 0 {
		limit = customLimit
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	jsonResp(w, http.StatusOK, map[string]any{
		"key":            maskKey(key),
		"current":        count,
		"limit":          limit,
		"tier":           tier,
		"remaining":      remaining,
		"total_requests": total,
		"daily":          daily,
	})
}

// AdminGetStats returns aggregate dashboard stats (admin only).
func (h *APIHandler) AdminGetStats(w http.ResponseWriter, r *http.Request) {
	var total, active, free, pro int

	h.Postgres.QueryRow(r.Context(), `SELECT COUNT(*) FROM api_keys`).Scan(&total)
	h.Postgres.QueryRow(r.Context(), `SELECT COUNT(*) FROM api_keys WHERE active=true`).Scan(&active)
	h.Postgres.QueryRow(r.Context(), `SELECT COUNT(*) FROM api_keys WHERE tier='free'`).Scan(&free)
	h.Postgres.QueryRow(r.Context(), `SELECT COUNT(*) FROM api_keys WHERE tier='pro'`).Scan(&pro)

	jsonResp(w, http.StatusOK, map[string]any{
		"total_keys":  total,
		"active_keys": active,
		"free_keys":   free,
		"pro_keys":    pro,
	})
}

// RotateKey generates a new API key, invalidates the old one, and returns the new key.
// The caller authenticates with their current key (via X-API-Key header); Auth middleware
// puts that key in context. The old key is marked inactive in Postgres and deleted from Redis.
func (h *APIHandler) RotateKey(w http.ResponseWriter, r *http.Request) {
	oldKey, ok := r.Context().Value(middleware.ContextKeyAPIKey).(string)
	if !ok || oldKey == "" {
		jsonResp(w, http.StatusUnauthorized, map[string]string{"error": "missing API key context"})
		return
	}

	// Look up old key details
	var k APIKey
	row := h.Postgres.QueryRow(r.Context(),
		`SELECT key, app_name, email, tier, rate_limit, active, created_at FROM api_keys WHERE key=$1`, oldKey,
	)
	if err := row.Scan(&k.Key, &k.AppName, &k.Email, &k.Tier, &k.RateLimit, &k.Active, &k.CreatedAt); err != nil {
		jsonResp(w, http.StatusNotFound, map[string]string{"error": "key not found"})
		return
	}

	// Generate new key with same metadata
	newKey := &APIKey{
		Key:       generateKey(),
		AppName:   k.AppName,
		Email:     k.Email,
		Tier:      k.Tier,
		RateLimit: k.RateLimit,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.insertKeyToPostgres(r.Context(), newKey); err != nil {
		h.Logger.Error("failed to insert rotated key", "error", err)
		jsonResp(w, http.StatusInternalServerError, map[string]string{"error": "failed to rotate key"})
		return
	}
	if err := h.syncKeyToRedis(r.Context(), newKey); err != nil {
		h.Logger.Warn("failed to cache rotated key", "error", err)
	}

	// Invalidate old key immediately
	h.Postgres.Exec(r.Context(), `UPDATE api_keys SET active=false WHERE key=$1`, oldKey)
	h.Redis.Del(r.Context(), "apikey:"+oldKey)

	h.Logger.Info("API key rotated", "app", k.AppName, "email", k.Email)

	jsonResp(w, http.StatusOK, map[string]any{
		"key":     newKey.Key,
		"message": "Old key is now invalid. Save your new key — it will not be shown again.",
	})
}

func jsonResp(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
