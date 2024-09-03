package redisclient

import (
	"context"
	"fmt"
	"time"

	"yk-dc-bot/internal/apperrors"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/logger"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
	log *logger.Logger
}

func NewRedisClient(cfg *config.Config, log *logger.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, apperrors.Wrap(err, "REDIS_CONNECTION_ERROR", "failed to connect to Redis")
	}

	log.Info("Successfully connected to Redis")
	return &Client{rdb: rdb, log: log}, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.rdb.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return apperrors.Wrap(err, "REDIS_SET_ERROR", fmt.Sprintf("Failed to set cache for key: %s", key))
	}
	return err
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	value, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", apperrors.New("REDIS_CACHE_MISS", fmt.Sprintf("Cache miss for key: %s", key))
	} else if err != nil {
		return "", apperrors.Wrap(err, "REDIS_GET_ERROR", fmt.Sprintf("Failed to get from cache for key: %s", key))
	}
	return value, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
