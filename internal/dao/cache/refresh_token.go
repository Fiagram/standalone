package dao_cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

type RefreshToken interface {
	Set(ctx context.Context, key string, id uint64, ttl time.Duration) error
	Get(ctx context.Context, key string) (id uint64, err error)
	Del(ctx context.Context, key string) (bool, error)
}

type refreshToken struct {
	client Client
	logger *zap.Logger
}

func NewDaoCacheRefreshToken(
	client Client,
	logger *zap.Logger,
) RefreshToken {
	return &refreshToken{
		client: client,
		logger: logger,
	}
}

func (r *refreshToken) getRefreshTokenCacheKey(token string) string {
	return fmt.Sprintf("refresh_token:%s", token)
}

func (r *refreshToken) Set(ctx context.Context, key string, accountID uint64, ttl time.Duration) error {
	logger := logger.LoggerWithContext(ctx, r.logger)

	cacheKey := r.getRefreshTokenCacheKey(key)
	if err := r.client.Set(ctx, cacheKey, accountID, ttl); err != nil {
		logger.With(zap.Error(err)).Error("failed to insert token key to cache")
		return err
	}

	return nil
}

func (r *refreshToken) Get(ctx context.Context, key string) (data uint64, err error) {
	logger := logger.LoggerWithContext(ctx, r.logger)

	cacheKey := r.getRefreshTokenCacheKey(key)
	cacheEntry, err := r.client.Get(ctx, cacheKey)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get token key to cache")
		return 0, err
	}

	if cacheEntry == nil {
		return 0, ErrCacheMiss
	}

	// Try to convert from string (Redis returns data as string)
	var accountID uint64
	if strData, ok := cacheEntry.(string); ok {
		id, err := strconv.ParseUint(strData, 10, 64)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to parse account id from cache")
			return 0, err
		}
		accountID = id
	} else if id, ok := cacheEntry.(uint64); ok {
		// In case it's already uint64 (from in-memory cache)
		accountID = id
	} else {
		logger.Error("cache entry is not of type string or uint64")
		return 0, fmt.Errorf("invalid cache entry type")
	}

	return accountID, nil
}

func (r *refreshToken) Del(ctx context.Context, key string) (bool, error) {
	logger := logger.LoggerWithContext(ctx, r.logger)

	cacheKey := r.getRefreshTokenCacheKey(key)
	err := r.client.Del(ctx, cacheKey)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to del token key to cache")
		return false, err
	}

	return true, nil
}
