package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis creates a Redis client and verifies connectivity.
func ConnectRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	rdb := redis.NewClient(opts)

	const maxRetries = 10
	const retryDelay = 2 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if lastErr = rdb.Ping(ctx).Err(); lastErr == nil {
			return rdb, nil
		}
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	rdb.Close()
	return nil, fmt.Errorf("could not connect to redis after %d attempts: %w", maxRetries, lastErr)
}
