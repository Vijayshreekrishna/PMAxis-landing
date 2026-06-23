package middleware

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pmaxis/pmaxis/libs/logger"
	redisclient "github.com/pmaxis/pmaxis/libs/redis-client"
)

var tierLimits = map[string]int{
	"free":       60,
	"pro":        600,
	"enterprise": 6000,
}

// RateLimit is a Redis-backed rate limiter. When an API key is present in context
// (set by Auth middleware), limits per key based on tier. Falls back to per-IP at 60/min.
// On Redis failure it fails open so the service stays available.
func RateLimit(rdb redisclient.Interface, log logger.Interface) func(http.Handler) http.Handler {
	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			limit := 60
			var rlKey string
			var apiKey string

			if k, ok := ctx.Value(ContextKeyAPIKey).(string); ok && k != "" {
				apiKey = k
				rlKey = fmt.Sprintf("rl:%s", apiKey)

				if custom, ok := ctx.Value(ContextKeyRateLimit).(int); ok && custom > 0 {
					limit = custom
				} else if tier, ok := ctx.Value(ContextKeyTier).(string); ok {
					if l, found := tierLimits[tier]; found {
						limit = l
					}
				}
			} else {
				ip := extractClientIP(r, trustedProxies)
				rlKey = fmt.Sprintf("rl:ip:%s", ip)
			}

			count, err := rdb.Incr(ctx, rlKey).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if count == 1 {
				rdb.Expire(ctx, rlKey, time.Minute)
			}

			// Track persistent usage counters for authenticated keys
			if apiKey != "" {
				date := time.Now().UTC().Format("20060102")
				dailyKey := "usage:daily:" + apiKey + ":" + date
				rdb.Incr(ctx, "usage:total:"+apiKey)
				if n, _ := rdb.Incr(ctx, dailyKey).Result(); n == 1 {
					rdb.Expire(ctx, dailyKey, 30*24*time.Hour)
				}
			}

			remaining := limit - int(count)
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if int(count) > limit {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"rate limit exceeded","retry_after":"60s"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func parseTrustedProxies(raw string) []*net.IPNet {
	if raw == "" {
		return nil
	}
	var nets []*net.IPNet
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !strings.Contains(s, "/") {
			s += "/32"
		}
		_, ipNet, err := net.ParseCIDR(s)
		if err == nil {
			nets = append(nets, ipNet)
		}
	}
	return nets
}

func extractClientIP(r *http.Request, trustedProxies []*net.IPNet) string {
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	if remoteIP == "" {
		remoteIP = r.RemoteAddr
	}

	if len(trustedProxies) > 0 && isTrusted(remoteIP, trustedProxies) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.SplitN(xff, ",", 2)
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	return remoteIP
}

func isTrusted(ipStr string, nets []*net.IPNet) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
