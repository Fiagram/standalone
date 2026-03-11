package logic_test

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/Fiagram/standalone/internal/configs"
	logic "github.com/Fiagram/standalone/internal/logic/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGenerateRefreshToken(t *testing.T) {
	logger := zap.NewNop()
	custom_config := configs.Token{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	tokenLogic := logic.NewTokenLogic(custom_config, logger)
	ctx := context.Background()

	// Generate refresh token
	token, expiresAt, err := tokenLogic.GenerateRefreshToken(ctx)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))
	assert.True(t, expiresAt.Before(time.Now().Add(custom_config.RefreshTokenTTL+time.Second)))
}

func TestGenerateRefreshTokenUniqueness(t *testing.T) {
	logger := zap.NewNop()
	custom_config := configs.Token{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	tokenLogic := logic.NewTokenLogic(custom_config, logger)
	ctx := context.Background()

	// Generate multiple refresh tokens
	token1, _, err1 := tokenLogic.GenerateRefreshToken(ctx)
	require.NoError(t, err1)

	token2, _, err2 := tokenLogic.GenerateRefreshToken(ctx)
	require.NoError(t, err2)

	// Tokens should be unique (different random bytes)
	assert.NotEqual(t, token1, token2)
}

func TestGenerateRefreshTokenFormat(t *testing.T) {
	logger := zap.NewNop()
	custom_config := configs.Token{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	tokenLogic := logic.NewTokenLogic(custom_config, logger)
	ctx := context.Background()

	token, _, err := tokenLogic.GenerateRefreshToken(ctx)
	require.NoError(t, err)

	// Verify token is valid base64 by attempting to decode
	decodedBytes, err := base64.RawURLEncoding.DecodeString(token)
	require.NoError(t, err)
	// Should decode to 64 bytes
	assert.Equal(t, 64, len(decodedBytes))
}
