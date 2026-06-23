package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/pmaxis/pmaxis/libs/logger"
	redisclient "github.com/pmaxis/pmaxis/libs/redis-client"
)

type apiKeyMeta struct {
	Tier      string `json:"tier"`
	Active    bool   `json:"active"`
	RateLimit int    `json:"rate_limit"`
}

// Auth validates the X-API-Key header against Redis and attaches tier/key/rate_limit to context.
func Auth(rdb redisclient.Interface, log logger.Interface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				jsonError(w, http.StatusUnauthorized, "missing X-API-Key header")
				return
			}

			val, err := rdb.Get(r.Context(), "apikey:"+key).Result()
			if err != nil {
				log.Warn("API key not found", "key_prefix", safePrefix(key))
				jsonError(w, http.StatusUnauthorized, "invalid API key")
				return
			}

			var meta apiKeyMeta
			if err := json.Unmarshal([]byte(val), &meta); err != nil {
				jsonError(w, http.StatusInternalServerError, "internal error")
				return
			}

			if !meta.Active {
				jsonError(w, http.StatusForbidden, "API key revoked")
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyAPIKey, key)
			ctx = context.WithValue(ctx, ContextKeyTier, meta.Tier)
			ctx = context.WithValue(ctx, ContextKeyRateLimit, meta.RateLimit)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func safePrefix(key string) string {
	if len(key) > 12 {
		return key[:12] + "..."
	}
	return key
}

// AuthWS validates an API key for WebSocket connections.
// Checks X-API-Key header first, then ?api_key= query param (browsers cannot set
// custom headers on WS upgrade requests, so the query param is the fallback).
func AuthWS(rdb redisclient.Interface, log logger.Interface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}
			if key == "" {
				w.Header().Set("Content-Type", "application/json")
				jsonError(w, http.StatusUnauthorized, "missing API key — pass X-API-Key header or ?api_key= query param")
				return
			}

			val, err := rdb.Get(r.Context(), "apikey:"+key).Result()
			if err != nil {
				log.Warn("WS: API key not found", "key_prefix", safePrefix(key))
				jsonError(w, http.StatusUnauthorized, "invalid API key")
				return
			}

			var meta apiKeyMeta
			if err := json.Unmarshal([]byte(val), &meta); err != nil {
				jsonError(w, http.StatusInternalServerError, "internal error")
				return
			}

			if !meta.Active {
				jsonError(w, http.StatusForbidden, "API key revoked")
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyAPIKey, key)
			ctx = context.WithValue(ctx, ContextKeyTier, meta.Tier)
			ctx = context.WithValue(ctx, ContextKeyRateLimit, meta.RateLimit)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// DebugAuth gates /debug/* and /admin/* routes behind DEBUG_TOKEN env var.
// Pass the token via X-Debug-Token header or ?debug_token= query param.
func DebugAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := os.Getenv("DEBUG_TOKEN")
		if token == "" {
			http.Error(w, "admin endpoints disabled", http.StatusForbidden)
			return
		}
		provided := r.Header.Get("X-Debug-Token")
		if provided == "" {
			provided = r.URL.Query().Get("debug_token")
		}
		if provided != token {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
