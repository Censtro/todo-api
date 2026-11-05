package statistics_cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Censtro/todo-api/internal/core/domain"
	statistics_service "github.com/Censtro/todo-api/internal/features/statistics/service"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, ttl time.Duration) (*Cache, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("cache TTL must be positive")
	}

	return &Cache{client: client, ttl: ttl}, nil
}

func (c *Cache) Get(
	ctx context.Context,
	key string,
) (domain.Statistics, bool, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.Statistics{}, false, nil
	}
	if err != nil {
		return domain.Statistics{}, false, fmt.Errorf("get statistics from Redis: %w", err)
	}

	var statistics domain.Statistics
	if err := json.Unmarshal(data, &statistics); err != nil {
		return domain.Statistics{}, false, fmt.Errorf("decode cached statistics: %w", err)
	}

	return statistics, true, nil
}

func (c *Cache) Set(
	ctx context.Context,
	key string,
	value domain.Statistics,
) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode statistics for Redis: %w", err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("set statistics in Redis: %w", err)
	}

	return nil
}

var _ statistics_service.StatisticsCache = (*Cache)(nil)
