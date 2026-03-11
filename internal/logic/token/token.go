package logic_token

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/Fiagram/standalone/internal/configs"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type TokenPayload struct {
	AccountId uint64
}

type Token interface {
	GenerateAccessToken(ctx context.Context, payload TokenPayload) (token string, expiresAt time.Time, err error)
	GetPayloadFromAccessToken(ctx context.Context, token string) (payload TokenPayload, expiresAt time.Time, err error)
	GenerateRefreshToken(ctx context.Context) (token string, expiresAt time.Time, err error)
}

func NewTokenLogic(
	config configs.Token,
	logger *zap.Logger,
) Token {
	return &token{
		config: config,
		logger: logger,
	}
}

type token struct {
	config configs.Token
	logger *zap.Logger
}

func (t *token) GenerateRefreshToken(ctx context.Context) (string, time.Time, error) {
	expiresAt := time.Now().Add(t.config.RefreshTokenTTL)

	randomBytes := make([]byte, 64)
	_, err := rand.Read(randomBytes)
	if err != nil {
		t.logger.Error("Failed to generate random bytes", zap.Error(err))
		return "", time.Time{}, err
	}

	tokenString := base64.RawURLEncoding.EncodeToString(randomBytes)

	return tokenString, expiresAt, nil
}

func (t *token) GenerateAccessToken(ctx context.Context, payload TokenPayload) (string, time.Time, error) {
	createAt := time.Now()
	expiresAt := createAt.Add(t.config.AccessTokenTTL)

	claims := jwt.MapClaims{
		"id":  payload.AccountId,
		"exp": expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(t.config.Secret))
	if err != nil {
		t.logger.Error("Failed to sign token", zap.Error(err))
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (t *token) GetPayloadFromAccessToken(ctx context.Context, tokenString string) (TokenPayload, time.Time, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(t.config.Secret), nil
	})

	if err != nil {
		t.logger.Error("Failed to parse token", zap.Error(err))
		return TokenPayload{}, time.Time{}, err
	}

	if !token.Valid {
		t.logger.Error("Invalid token")
		return TokenPayload{}, time.Time{}, errors.New("invalid token")
	}

	accountID, ok := claims["id"].(float64)
	if !ok {
		t.logger.Error("Failed to extract id from token")
		return TokenPayload{}, time.Time{}, errors.New("invalid id in token")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		t.logger.Error("Failed to extract exp from token")
		return TokenPayload{}, time.Time{}, errors.New("invalid exp in token")
	}
	expiresAt := time.Unix(int64(exp), 0)

	return TokenPayload{
		AccountId: uint64(accountID),
	}, expiresAt, nil
}
