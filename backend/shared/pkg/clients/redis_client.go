package clients

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	mqClient    *asynq.Client
	cacheClient *redis.Client
	config      RedisConfig
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	mqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	cacheClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := cacheClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		mqClient:    mqClient,
		cacheClient: cacheClient,
		config:      cfg,
	}, nil
}

func (c *RedisClient) GetMQClient() *asynq.Client {
	return c.mqClient
}

func (c *RedisClient) GetCacheClient() *redis.Client {
	return c.cacheClient
}

func (c *RedisClient) GetRedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Password: c.config.Password,
		DB:       c.config.DB,
	}
}

func (c *RedisClient) Close() error {
	var err error
	if c.mqClient != nil {
		if e := c.mqClient.Close(); e != nil {
			err = e
		}
	}
	if c.cacheClient != nil {
		if e := c.cacheClient.Close(); e != nil {
			err = e
		}
	}
	return err
}
