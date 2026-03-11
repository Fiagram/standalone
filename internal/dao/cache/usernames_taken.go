package dao_cache

import (
	"context"

	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

const (
	usernamesTakenKey = "usernames_taken"
)

type UsernamesTaken interface {
	Add(ctx context.Context, username string) error
	Has(ctx context.Context, username string) (bool, error)
}

type usernamesTaken struct {
	client Client
	logger *zap.Logger
}

func NewDaoCacheUsernamesTaken(
	client Client,
	logger *zap.Logger,
) UsernamesTaken {
	return &usernamesTaken{
		client: client,
		logger: logger,
	}
}

func (u *usernamesTaken) Add(ctx context.Context, username string) error {
	logger := logger.LoggerWithContext(ctx, u.logger).With(zap.String("username", username))
	if err := u.client.AddToSet(ctx, usernamesTakenKey, username); err != nil {
		logger.With(zap.Error(err)).Error("failed to add username to set in cache")
		return err
	}
	return nil
}

func (u *usernamesTaken) Has(ctx context.Context, username string) (bool, error) {
	logger := logger.LoggerWithContext(ctx, u.logger).With(zap.String("username", username))
	isTaken, err := u.client.IsDataInSet(ctx, usernamesTakenKey, username)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to check if username is in set in cache")
		return false, err
	}
	return isTaken, nil
}
