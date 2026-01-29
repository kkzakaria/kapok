package security

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	// RequestsPerMinute is the maximum requests allowed per minute
	RequestsPerMinute int
	// BurstSize allows short bursts above the rate limit
	BurstSize int
	// BlockDuration is how long to block an IP after excessive requests
	BlockDuration time.Duration
}

// DefaultRateLimitConfig returns the default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 1000,
		BurstSize:         100,
		BlockDuration:     5 * time.Minute,
	}
}

// AuthRateLimitConfig returns stricter limits for authentication endpoints
func AuthRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 100,
		BurstSize:         10,
		BlockDuration:     15 * time.Minute,
	}
}

// RateLimiter implements distributed rate limiting using Redis
type RateLimiter struct {
	redis  *redis.Client
	logger zerolog.Logger
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client, logger zerolog.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		logger: logger,
	}
}

// Allow checks if a request should be allowed based on rate limits
func (rl *RateLimiter) Allow(ctx context.Context, key string, config RateLimitConfig) (bool, *RateLimitInfo, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/60) // 1-minute window
	blockKey := fmt.Sprintf("ratelimit:block:%s", key)

	// Check if IP is blocked
	blocked, err := rl.redis.Exists(ctx, blockKey).Result()
	if err != nil {
		rl.logger.Error().Err(err).Msg("failed to check block status")
		// Fail open: allow request if Redis is down
		return true, nil, err
	}

	if blocked > 0 {
		ttl, _ := rl.redis.TTL(ctx, blockKey).Result()
		info := &RateLimitInfo{
			Allowed:       false,
			Limit:         config.RequestsPerMinute,
			Remaining:     0,
			ResetAt:       now.Add(ttl),
			RetryAfter:    ttl,
			BlockedUntil:  now.Add(ttl),
		}
		return false, info, nil
	}

	// Increment request count
	pipe := rl.redis.Pipeline()
	incrCmd := pipe.Incr(ctx, windowKey)
	pipe.Expire(ctx, windowKey, 2*time.Minute) // Keep for 2 minutes for sliding window
	_, err = pipe.Exec(ctx)
	if err != nil {
		rl.logger.Error().Err(err).Msg("failed to increment rate limit counter")
		return true, nil, err
	}

	count := incrCmd.Val()
	limit := int64(config.RequestsPerMinute + config.BurstSize)
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	resetAt := now.Add(time.Minute).Truncate(time.Minute)

	info := &RateLimitInfo{
		Allowed:    count <= limit,
		Limit:      int(limit),
		Remaining:  int(remaining),
		ResetAt:    resetAt,
		RetryAfter: time.Until(resetAt),
	}

	// If significantly over limit, block the IP
	if count > limit*2 {
		err = rl.redis.Set(ctx, blockKey, "1", config.BlockDuration).Err()
		if err != nil {
			rl.logger.Error().Err(err).Msg("failed to set block key")
		} else {
			info.BlockedUntil = now.Add(config.BlockDuration)
			rl.logger.Warn().
				Str("key", key).
				Int64("count", count).
				Dur("block_duration", config.BlockDuration).
				Msg("IP blocked due to excessive requests")
		}
	}

	return info.Allowed, info, nil
}

// RateLimitInfo contains information about the current rate limit status
type RateLimitInfo struct {
	Allowed      bool
	Limit        int
	Remaining    int
	ResetAt      time.Time
	RetryAfter   time.Duration
	BlockedUntil time.Time
}

// RateLimitMiddleware creates HTTP middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware(config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client identifier (IP address or tenant ID)
			clientIP := getClientIP(r)
			key := clientIP

			// If tenant_id is in context, use that for more granular limiting
			if tenantID := getTenantIDFromContext(r.Context()); tenantID != "" {
				key = fmt.Sprintf("tenant:%s:%s", tenantID, clientIP)
			}

			allowed, info, err := rl.Allow(r.Context(), key, config)
			if err != nil {
				// Log error but allow request (fail open)
				rl.logger.Error().Err(err).Msg("rate limiter error")
			}

			// Set rate limit headers
			if info != nil {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", info.ResetAt.Unix()))
			}

			if !allowed {
				if info != nil && !info.BlockedUntil.IsZero() {
					w.Header().Set("Retry-After", fmt.Sprintf("%d", int(info.RetryAfter.Seconds())))
					http.Error(w, fmt.Sprintf("Too many requests. Blocked until %s", info.BlockedUntil.Format(time.RFC3339)), http.StatusTooManyRequests)
				} else {
					w.Header().Set("Retry-After", fmt.Sprintf("%d", int(info.RetryAfter.Seconds())))
					http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				}

				rl.logger.Warn().
					Str("client_ip", clientIP).
					Str("path", r.URL.Path).
					Int("remaining", info.Remaining).
					Msg("rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxy/load balancer scenarios)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if idx := len(xff); idx > 0 {
			return xff[:idx]
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// getTenantIDFromContext extracts tenant ID from context if available
func getTenantIDFromContext(ctx context.Context) string {
	// Try to get from JWT claims
	claims, ok := ctx.Value("jwt_claims").(map[string]interface{})
	if !ok {
		return ""
	}

	tenantID, ok := claims["tenant_id"].(string)
	if !ok {
		return ""
	}

	return tenantID
}
