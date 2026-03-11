package dao_cache_test

import (
	"context"
	"testing"
	"time"

	dao_cache "github.com/Fiagram/standalone/internal/dao/cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRefreshTokenSet(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	refreshToken := dao_cache.NewDaoCacheRefreshToken(client, logger)

	testToken := "test_token_set"
	testAccountID := uint64(12345)

	// Set the refresh token
	err := refreshToken.Set(ctx, testToken, testAccountID, time.Minute)
	assert.NoError(t, err, "Set should not return an error")

	t.Cleanup(func() {
		_, _ = refreshToken.Del(ctx, testToken)
	})
}

func TestRefreshTokenGet(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	refreshToken := dao_cache.NewDaoCacheRefreshToken(client, logger)

	testToken := "test_token_get"
	testAccountID := uint64(67890)

	// Set the refresh token first
	err := refreshToken.Set(ctx, testToken, testAccountID, time.Minute)
	assert.NoError(t, err, "Set should not return an error")

	// Get the refresh token
	data, err := refreshToken.Get(ctx, testToken)
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, testAccountID, data, "Retrieved account ID should match the set value")

	t.Cleanup(func() {
		_, _ = refreshToken.Del(ctx, testToken)
	})
}

func TestRefreshTokenGetCacheMiss(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	refreshToken := dao_cache.NewDaoCacheRefreshToken(client, logger)

	testToken := "test_token_nonexistent"

	// Get non-existent refresh token
	_, err := refreshToken.Get(ctx, testToken)
	assert.Equal(t, dao_cache.ErrCacheMiss, err, "Get should return ErrCacheMiss for non-existent token")
}

func TestRefreshTokenDel(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	refreshToken := dao_cache.NewDaoCacheRefreshToken(client, logger)

	testToken := "test_token_del"
	testAccountID := uint64(11111)

	// Set the refresh token first
	err := refreshToken.Set(ctx, testToken, testAccountID, time.Minute)
	assert.NoError(t, err, "Set should not return an error")

	// Delete the refresh token
	deleted, err := refreshToken.Del(ctx, testToken)
	assert.NoError(t, err, "Del should not return an error")
	assert.True(t, deleted, "Del should return true on successful deletion")

	// Verify it's deleted
	_, err = refreshToken.Get(ctx, testToken)
	assert.Equal(t, dao_cache.ErrCacheMiss, err, "Get should return ErrCacheMiss after deletion")
}

func TestRefreshTokenSetAndOverwrite(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	refreshToken := dao_cache.NewDaoCacheRefreshToken(client, logger)

	testToken := "test_token_overwrite"
	accountID1 := uint64(11111)
	accountID2 := uint64(22222)

	// Set initial refresh token
	err := refreshToken.Set(ctx, testToken, accountID1, time.Minute)
	assert.NoError(t, err, "First Set should not return an error")

	// Overwrite with new account ID
	err = refreshToken.Set(ctx, testToken, accountID2, time.Minute)
	assert.NoError(t, err, "Second Set should not return an error")

	// Verify the new value is stored
	data, err := refreshToken.Get(ctx, testToken)
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, accountID2, data, "Retrieved account ID should be the overwritten value")

	// Cleanup
	t.Cleanup(func() {
		_, _ = refreshToken.Del(ctx, testToken)
	})
}

func TestRefreshTokenDelNonExistent(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	refreshToken := dao_cache.NewDaoCacheRefreshToken(client, logger)

	testToken := "test_token_nonexistent_del"

	// Delete non-existent refresh token
	deleted, err := refreshToken.Del(ctx, testToken)
	// Del should still return true even if the key doesn't exist (Redis behavior)
	assert.NoError(t, err, "Del should not return an error")
	assert.True(t, deleted, "Del should return true")
}

func TestRefreshTokenMultipleKeys(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	refreshToken := dao_cache.NewDaoCacheRefreshToken(client, logger)

	testToken1 := "test_token_multi_1"
	testToken2 := "test_token_multi_2"
	accountID1 := uint64(99999)
	accountID2 := uint64(88888)

	// Set multiple tokens
	err := refreshToken.Set(ctx, testToken1, accountID1, time.Minute)
	assert.NoError(t, err, "First Set should not return an error")

	err = refreshToken.Set(ctx, testToken2, accountID2, time.Minute)
	assert.NoError(t, err, "Second Set should not return an error")

	// Verify both tokens are stored correctly
	data1, err := refreshToken.Get(ctx, testToken1)
	assert.NoError(t, err, "First Get should not return an error")
	assert.Equal(t, accountID1, data1, "First token should have correct account ID")

	data2, err := refreshToken.Get(ctx, testToken2)
	assert.NoError(t, err, "Second Get should not return an error")
	assert.Equal(t, accountID2, data2, "Second token should have correct account ID")

	// Cleanup
	t.Cleanup(func() {
		_, _ = refreshToken.Del(ctx, testToken1)
		_, _ = refreshToken.Del(ctx, testToken2)
	})
}
