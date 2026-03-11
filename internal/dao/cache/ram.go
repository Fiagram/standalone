package dao_cache

import (
	"context"
	"slices"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ramClient struct {
	cache      map[string]any
	cacheMutex *sync.Mutex
	logger     *zap.Logger
}

func NewRamClient(
	logger *zap.Logger,
) Client {
	return &ramClient{
		cache:      make(map[string]any),
		cacheMutex: new(sync.Mutex),
		logger:     logger,
	}
}

func (c ramClient) Set(_ context.Context, key string, data any, _ time.Duration) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cache[key] = data
	return nil
}

func (c ramClient) Get(_ context.Context, key string) (any, error) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	data, ok := c.cache[key]
	if !ok {
		return nil, ErrCacheMiss
	}

	return data, nil
}

func (c ramClient) AddToSet(_ context.Context, key string, data ...any) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	set := c.getSet(key)
	set = append(set, data...)
	c.cache[key] = set
	return nil
}

func (c ramClient) IsDataInSet(_ context.Context, key string, data any) (bool, error) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	set := c.getSet(key)

	if slices.Contains(set, data) {
		return true, nil
	}

	return false, nil
}

func (c ramClient) Del(ctx context.Context, keys ...string) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	for _, key := range keys {
		delete(c.cache, key)
	}

	return nil
}

func (c ramClient) getSet(key string) []any {
	setValue, ok := c.cache[key]
	if !ok {
		return make([]any, 0)
	}

	set, ok := setValue.([]any)
	if !ok {
		return make([]any, 0)
	}

	return set
}
