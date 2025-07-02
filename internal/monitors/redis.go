package monitors

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sxwebdev/sentinel/internal/config"
)

// RedisMonitor monitors Redis instances
type RedisMonitor struct {
	BaseMonitor
	client *redis.Client
}

// NewRedisMonitor creates a new Redis monitor
func NewRedisMonitor(cfg config.ServiceConfig) (*RedisMonitor, error) {
	password := getConfigString(cfg.Config, "password", "")
	db := getConfigInt(cfg.Config, "db", 0)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Endpoint,
		Password:     password,
		DB:           db,
		DialTimeout:  cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	return &RedisMonitor{
		BaseMonitor: NewBaseMonitor(cfg),
		client:      client,
	}, nil
}

// Check performs the Redis health check
func (r *RedisMonitor) Check(ctx context.Context) error {
	// Perform PING command
	result, err := r.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	// Check if response is expected "PONG"
	if result != "PONG" {
		return fmt.Errorf("unexpected ping response: %s", result)
	}

	return nil
}

// Close closes the Redis connection
func (r *RedisMonitor) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
