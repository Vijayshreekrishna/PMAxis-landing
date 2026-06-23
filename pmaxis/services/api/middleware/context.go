package middleware

type contextKey string

const (
	ContextKeyAPIKey    contextKey = "api_key"
	ContextKeyTier      contextKey = "tier"
	ContextKeyRateLimit contextKey = "rate_limit"
)
