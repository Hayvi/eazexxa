package services

import (
	"context"
	"fmt"
	"time"

	"github.com/betpro/server/internal/config"
	"github.com/betpro/server/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client  *redis.Client
	enabled bool
}

func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	if !cfg.RedisEnabled {
		logger.Info("Redis disabled")
		return &RedisClient{enabled: false}, nil
	}

	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	logger.Info("Redis connected", "url", cfg.RedisURL)

	return &RedisClient{
		client:  client,
		enabled: true,
	}, nil
}

func (r *RedisClient) Client() *redis.Client {
	if !r.enabled {
		return nil
	}
	return r.client
}

func (r *RedisClient) IsEnabled() bool {
	return r.enabled
}

func (r *RedisClient) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if !r.enabled {
		return "", fmt.Errorf("redis not enabled")
	}
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !r.enabled {
		return fmt.Errorf("redis not enabled")
	}
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	if !r.enabled {
		return fmt.Errorf("redis not enabled")
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	if !r.enabled {
		return 0, fmt.Errorf("redis not enabled")
	}
	return r.client.Exists(ctx, keys...).Result()
}

func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	if !r.enabled {
		return 0, fmt.Errorf("redis not enabled")
	}
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if !r.enabled {
		return fmt.Errorf("redis not enabled")
	}
	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	if !r.enabled {
		return fmt.Errorf("redis not enabled")
	}
	return r.client.Publish(ctx, channel, message).Err()
}

func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if !r.enabled {
		return nil
	}
	return r.client.Subscribe(ctx, channels...)
}
