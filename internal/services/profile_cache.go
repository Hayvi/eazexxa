package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/betpro/server/internal/models"
)

type RedisProfileCache struct {
	redis *RedisClient
	ttl   time.Duration
}

func NewRedisProfileCache(redis *RedisClient) *RedisProfileCache {
	return &RedisProfileCache{
		redis: redis,
		ttl:   5 * time.Minute,
	}
}

func (c *RedisProfileCache) Get(ctx context.Context, userID string) (*models.Profile, error) {
	if !c.redis.IsEnabled() {
		return nil, nil
	}

	key := "profile:" + userID
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, nil
	}

	var profile models.Profile
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (c *RedisProfileCache) Set(ctx context.Context, userID string, profile *models.Profile) error {
	if !c.redis.IsEnabled() {
		return nil
	}

	key := "profile:" + userID
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, c.ttl)
}

func (c *RedisProfileCache) Invalidate(ctx context.Context, userID string) error {
	if !c.redis.IsEnabled() {
		return nil
	}

	key := "profile:" + userID
	return c.redis.Del(ctx, key)
}
