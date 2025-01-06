package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type RateLimiterConfig struct {
	// Maximum number of requests allowed within the window
	Max int

	// Duration of the sliding window
	Window time.Duration

	// Redis client instance
	Redis *redis.Client

	// Optional prefix for Redis keys
	KeyPrefix string

	// Optional response when rate limit is exceeded
	LimitExceededHandler fiber.Handler

	// Optional key generation function
	KeyFunc func(c *fiber.Ctx) string
}

// Default config for rate limiter
var ConfigDefault = RateLimiterConfig{
	Max:       60,
	Window:    time.Minute,
	KeyPrefix: "ratelimit:",
	LimitExceededHandler: func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "Too many requests",
		})
	},
	KeyFunc: func(c *fiber.Ctx) string {
		return c.IP()
	},
}

// Helper function to get configuration with defaults
func configDefault(config ...RateLimiterConfig) RateLimiterConfig {
	if len(config) < 1 {
		return ConfigDefault
	}

	cfg := config[0]

	if cfg.Max <= 0 {
		cfg.Max = ConfigDefault.Max
	}
	if cfg.Window == 0 {
		cfg.Window = ConfigDefault.Window
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = ConfigDefault.KeyPrefix
	}
	if cfg.LimitExceededHandler == nil {
		cfg.LimitExceededHandler = ConfigDefault.LimitExceededHandler
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = ConfigDefault.KeyFunc
	}

	return cfg
}

// NewRateLimiter creates a new rate limiter middleware
func NewRateLimiter(config ...RateLimiterConfig) fiber.Handler {
	cfg := configDefault(config...)

	if cfg.Redis == nil {
		panic("Redis client is required for rate limiter")
	}

	// Return the middleware handler
	return func(c *fiber.Ctx) error {
		// Get IP address
		key := fmt.Sprintf("%s%s", cfg.KeyPrefix, cfg.KeyFunc(c))

		// Get current timestamp in milliseconds
		now := time.Now().UnixNano() / int64(time.Millisecond)
		windowStart := now - int64(cfg.Window/time.Millisecond)

		// Create a Redis pipeline for atomic operations
		pipe := cfg.Redis.Pipeline()
		ctx := context.Background()

		// Remove old requests outside the window
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

		// Add current request
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: now,
		})

		// Get the count of requests within the window
		countCmd := pipe.ZCard(ctx, key)

		// Set key expiration to match the window
		pipe.Expire(ctx, key, cfg.Window)

		// Execute pipeline
		_, err := pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("rate limiter redis error: %w", err)
		}

		// Check if limit is exceeded
		count := countCmd.Val()
		if count > int64(cfg.Max) {
			c.Set("X-RateLimit-RetryAfter", fmt.Sprintf("%d", cfg.Window/time.Second))
			return cfg.LimitExceededHandler(c)
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", cfg.Max-int(count)))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", now+int64(cfg.Window/time.Millisecond)))

		return c.Next()
	}
}
