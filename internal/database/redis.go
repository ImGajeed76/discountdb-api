package database

import (
	"context"
	"crypto/tls"
	"discountdb-api/internal/config"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.REDISHost, cfg.REDISPort),
		Username: cfg.REDISUser,
		Password: cfg.REDISPassword,
		DB:       0,

		// Connection Pool
		PoolSize:        30,
		MinIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		PoolTimeout:     4 * time.Second,

		// Timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,

		// Retry Strategy
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,

		// TLS Config
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})

	// Test the connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	return rdb, nil
}
