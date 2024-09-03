package redisclient

import (
	"context"
	"fmt"
	"time"

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
		log.Error("Failed to connect to Redis", "error", err)
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	fmt.Println("hdev api key", cfg.HdevApiKey)

	log.Info("Successfully connected to Redis")
	return &Client{rdb: rdb, log: log}, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.rdb.Set(ctx, key, value, expiration).Err()
	if err != nil {
		c.log.Error("Failed to set cache", "key", key, "error", err)
	} else {
		c.log.Debug("Successfully set cache", "key", key, "expiration", expiration)
	}
	return err
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	value, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		c.log.Debug("Cache miss", "key", key)
		return "", err
	} else if err != nil {
		c.log.Error("Failed to get from cache", "key", key, "error", err)
		return "", err
	}
	c.log.Debug("Cache hit", "key", key)
	return value, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
