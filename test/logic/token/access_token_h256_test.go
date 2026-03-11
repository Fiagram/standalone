package logic_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/Fiagram/standalone/internal/configs"
	logic "github.com/Fiagram/standalone/internal/logic/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTokenGenerate(t *testing.T) {
	logger := zap.NewNop()
	custom_config := configs.Token{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	tokenLogic := logic.NewTokenLogic(custom_config, logger)

	tests := []struct {
		name      string
		accountID uint64
	}{
		{
			name:      "generate token with valid account id",
			accountID: 12345,
		},
		{
			name:      "generate token with zero account id",
			accountID: 0,
		},
		{
			name:      "generate token with large account id",
			accountID: 9223372036854775807,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := logic.TokenPayload{
				AccountId: tt.accountID,
			}

			ctx := context.Background()
			token, expiresAt, err := tokenLogic.GenerateAccessToken(ctx, payload)

			require.NoError(t, err)
			assert.NotEmpty(t, token)
			assert.True(t, expiresAt.After(time.Now()))
			assert.True(t, expiresAt.Before(time.Now().Add(custom_config.AccessTokenTTL+time.Second)))
		})
	}
}

func TestTokenGetPayload(t *testing.T) {
	logger := zap.NewNop()
	custom_config := configs.Token{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	tokenLogic := logic.NewTokenLogic(custom_config, logger)
	ctx := context.Background()

	tests := []struct {
		name      string
		accountID uint64
	}{
		{
			name:      "extract payload with valid account id",
			accountID: 12345,
		},
		{
			name:      "extract payload with zero account id",
			accountID: 0,
		},
		{
			name:      "extract payload with large account id",
			accountID: math.MaxUint32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a token
			payload := logic.TokenPayload{
				AccountId: tt.accountID,
			}
			token, _, err := tokenLogic.GenerateAccessToken(ctx, payload)
			require.NoError(t, err)

			// Extract payload from token
			retrievedPayload, expiresAt, err := tokenLogic.GetPayloadFromAccessToken(ctx, token)

			require.NoError(t, err)
			assert.Equal(t, tt.accountID, retrievedPayload.AccountId)
			assert.True(t, expiresAt.After(time.Now()))
		})
	}
}

func TestTokenInvalidToken(t *testing.T) {
	logger := zap.NewNop()
	custom_config := configs.Token{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	tokenLogic := logic.NewTokenLogic(custom_config, logger)
	ctx := context.Background()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "invalid token format",
			token: "invalid.token.format",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "malformed token",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
		},
		{
			name:  "token with wrong secret",
			token: generateTokenWithDifferentSecret(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := tokenLogic.GetPayloadFromAccessToken(ctx, tt.token)
			assert.Error(t, err)
		})
	}
}

func TestTokenRoundTrip(t *testing.T) {
	logger := zap.NewNop()
	custom_config := configs.Token{
		Secret:          "test-secret-key-123",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	tokenLogic := logic.NewTokenLogic(custom_config, logger)
	ctx := context.Background()

	originalPayload := logic.TokenPayload{
		AccountId: 98765,
	}

	// Generate token
	token, expiresAt, err := tokenLogic.GenerateAccessToken(ctx, originalPayload)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Extract payload
	retrievedPayload, retrievedExpiresAt, err := tokenLogic.GetPayloadFromAccessToken(ctx, token)
	require.NoError(t, err)

	// Verify payload matches
	assert.Equal(t, originalPayload.AccountId, retrievedPayload.AccountId)
	assert.True(t, expiresAt.After(time.Now()))
	assert.Equal(t, expiresAt.Unix(), retrievedExpiresAt.Unix())
}

func TestTokenDifferentSecrets(t *testing.T) {
	logger := zap.NewNop()
	config1 := configs.Token{
		Secret:          "secret-key-1",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	config2 := configs.Token{
		Secret:          "secret-key-2",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	tokenLogic1 := logic.NewTokenLogic(config1, logger)
	tokenLogic2 := logic.NewTokenLogic(config2, logger)
	ctx := context.Background()

	payload := logic.TokenPayload{
		AccountId: 11111,
	}

	// Generate token with config1
	token, _, err := tokenLogic1.GenerateAccessToken(ctx, payload)
	require.NoError(t, err)

	// Try to extract payload with config2 (different secret)
	_, _, err = tokenLogic2.GetPayloadFromAccessToken(ctx, token)
	assert.Error(t, err, "should fail when using different secret")
}

// Helper function to generate a token with a different secret
func generateTokenWithDifferentSecret() string {
	logger := zap.NewNop()
	custom_config := configs.Token{
		Secret:          "different-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	tokenLogic := logic.NewTokenLogic(custom_config, logger)
	ctx := context.Background()

	payload := logic.TokenPayload{
		AccountId: 12345,
	}

	token, _, _ := tokenLogic.GenerateAccessToken(ctx, payload)
	return token
}
