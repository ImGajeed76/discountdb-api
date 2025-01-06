package middleware

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimiterConfig holds the configuration for the rate limiter middleware
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

// ConfigDefault provides default configuration
var ConfigDefault = RateLimiterConfig{
	Max:       60,
	Window:    time.Minute,
	KeyPrefix: "ratelimit",
	LimitExceededHandler: func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "Too many requests",
		})
	},
	KeyFunc: func(c *fiber.Ctx) string {
		return c.IP()
	},
}

// validateConfig ensures the configuration is valid
func validateConfig(cfg *RateLimiterConfig) error {
	if cfg.Max <= 0 {
		return fmt.Errorf("max requests must be greater than 0")
	}
	if cfg.Window < time.Second {
		return fmt.Errorf("window must be at least 1 second")
	}
	if cfg.Window > 24*time.Hour {
		return fmt.Errorf("window must not exceed 24 hours")
	}
	return nil
}

// configDefault returns a config with default values for unset fields
func configDefault(config ...RateLimiterConfig) (RateLimiterConfig, error) {
	cfg := ConfigDefault

	if len(config) > 0 {
		if config[0].Max > 0 {
			cfg.Max = config[0].Max
		}
		if config[0].Window != 0 {
			cfg.Window = config[0].Window
		}
		if config[0].KeyPrefix != "" {
			cfg.KeyPrefix = config[0].KeyPrefix
		}
		if config[0].LimitExceededHandler != nil {
			cfg.LimitExceededHandler = config[0].LimitExceededHandler
		}
		if config[0].KeyFunc != nil {
			cfg.KeyFunc = config[0].KeyFunc
		}
		if config[0].Redis != nil {
			cfg.Redis = config[0].Redis
		}
	}

	if err := validateConfig(&cfg); err != nil {
		return cfg, fmt.Errorf("invalid rate limiter config: %w", err)
	}

	return cfg, nil
}

// getRemainingTime calculates the time remaining until the rate limit resets
func getRemainingTime(ctx context.Context, redis *redis.Client, key string) (int64, error) {
	ttl, err := redis.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int64(math.Ceil(ttl.Seconds())), nil
}

// NewRateLimiter creates a new rate limiter middleware
func NewRateLimiter(config ...RateLimiterConfig) fiber.Handler {
	cfg, err := configDefault(config...)
	if err != nil {
		panic(err)
	}

	if cfg.Redis == nil {
		panic("Redis client is required for rate limiter")
	}

	// Return the middleware handler
	return func(c *fiber.Ctx) error {
		// Use request context for proper cancellation
		ctx := c.Context()

		// Generate Redis key with proper separator
		key := fmt.Sprintf("%s:%s", cfg.KeyPrefix, cfg.KeyFunc(c))

		// Use Redis MULTI/EXEC for atomic operations
		pipe := cfg.Redis.TxPipeline()

		// Increment counter and set expiration in single transaction
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, cfg.Window)

		// Execute transaction
		_, err := pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("rate limiter redis error: %w", err)
		}

		// Get the current count
		count := incr.Val()

		// If this is the first request, ensure the key expires
		if count == 1 {
			cfg.Redis.Expire(ctx, key, cfg.Window)
		}

		// Calculate remaining attempts
		remaining := int64(cfg.Max) - count
		if remaining < 0 {
			remaining = 0
		}

		// Get remaining time until reset
		remainingTime, err := getRemainingTime(ctx, cfg.Redis, key)
		if err != nil {
			return fmt.Errorf("error getting remaining time: %w", err)
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", remainingTime))

		// Check if limit is exceeded
		if count > int64(cfg.Max) {
			c.Set("X-RateLimit-RetryAfter", fmt.Sprintf("%d", remainingTime))
			return cfg.LimitExceededHandler(c)
		}

		return c.Next()
	}
}
