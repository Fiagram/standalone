package dao_cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Fiagram/standalone/internal/configs"
	"github.com/Fiagram/standalone/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRedisClient(
	config configs.CacheClient,
	logger *zap.Logger,
) Client {
	logger.With(zap.Any("cache_config", config))

	aObj := redis.NewClient(&redis.Options{
		Addr:     config.Address + ":" + config.Port,
		Username: config.Username,
		Password: config.Password,
	})
	pong, err := aObj.Ping(context.Background()).Result()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to ping the redis server")
	}
	if pong == "PONG" {
		fmt.Println("The redis server connected!")
	}
	return &redisClient{
		accessObject: aObj,
		logger:       logger,
	}
}

type redisClient struct {
	accessObject *redis.Client
	logger       *zap.Logger
}

func (c *redisClient) AddToSet(ctx context.Context, key string, data ...any) error {
	logger := logger.LoggerWithContext(ctx, c.logger).
		With(zap.String("key", key)).
		With(zap.Any("data", data))

	if err := c.accessObject.SAdd(ctx, key, data...).Err(); err != nil {
		logger.With(zap.Error(err)).Error("failed to set data into set inside cache")
		return fmt.Errorf("failed to set data into set inside cache: %w", err)
	}

	return nil
}

func (c *redisClient) Get(ctx context.Context, key string) (any, error) {
	logger := logger.LoggerWithContext(ctx, c.logger).
		With(zap.String("key", key))

	data, err := c.accessObject.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}

		logger.With(zap.Error(err)).Error("failed to get data from cache")
		return nil, fmt.Errorf("failed to get data from cache: %w", err)
	}

	return data, nil
}

func (c *redisClient) IsDataInSet(ctx context.Context, key string, data any) (bool, error) {
	logger := logger.LoggerWithContext(ctx, c.logger).
		With(zap.String("key", key)).
		With(zap.Any("data", data))

	result, err := c.accessObject.SIsMember(ctx, key, data).Result()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to check if data is member of set inside cache")
		return false, fmt.Errorf("failed to check if data is member of set inside cache: %w", err)
	}

	return result, nil
}

func (c *redisClient) Set(ctx context.Context, key string, data any, ttl time.Duration) error {
	logger := logger.LoggerWithContext(ctx, c.logger).
		With(zap.String("key", key)).
		With(zap.Any("data", data)).
		With(zap.Duration("ttl", ttl))

	if err := c.accessObject.Set(ctx, key, data, ttl).Err(); err != nil {
		fmt.Println(err)
		logger.With(zap.Error(err)).Error("failed to set data into cache")
		return fmt.Errorf("failed to set data into cache: %w", err)
	}

	return nil
}

func (c *redisClient) Del(ctx context.Context, key ...string) error {
	logger := logger.LoggerWithContext(ctx, c.logger).
		With(zap.Any("keys", key))

	if err := c.accessObject.Del(ctx, key...).Err(); err != nil {
		logger.With(zap.Error(err)).Error("failed to delete keys in cache")
		return fmt.Errorf("failed to delete data: %w", err)
	}

	return nil
}
