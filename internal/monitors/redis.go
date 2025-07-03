package monitors

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// RedisMonitor monitors Redis endpoints
type RedisMonitor struct {
	BaseMonitor
	password string
	db       int
}

// NewRedisMonitor creates a new Redis monitor
func NewRedisMonitor(cfg storage.Service) (*RedisMonitor, error) {
	// Extract Redis config
	var redisConfig *storage.RedisConfig
	if cfg.Config.Redis != nil {
		redisConfig = cfg.Config.Redis
	}

	monitor := &RedisMonitor{
		BaseMonitor: NewBaseMonitor(cfg),
		db:          0,
	}

	// Apply Redis-specific config if available
	if redisConfig != nil {
		monitor.password = redisConfig.Password
		monitor.db = redisConfig.DB
	}

	return monitor, nil
}

// Check performs the Redis health check
func (r *RedisMonitor) Check(ctx context.Context) error {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         r.config.Endpoint,
		Password:     r.password,
		DB:           r.db,
		DialTimeout:  r.config.Timeout,
		ReadTimeout:  r.config.Timeout,
		WriteTimeout: r.config.Timeout,
	})
	defer client.Close()

	// Test connection with PING command
	ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	return nil
}
