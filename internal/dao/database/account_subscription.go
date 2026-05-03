package dao_database

import (
	"context"
	"time"

	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

type AccountSubscription struct {
	OfAccountId   uint64     `json:"of_account_id"`
	Plan          string     `json:"plan"`
	BillingPeriod string     `json:"billing_period"`
	Status        string     `json:"status"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AccountSubscriptionAccessor interface {
	GetSubscriptionByAccountId(ctx context.Context, accountId uint64) (AccountSubscription, error)
	CreateSubscription(ctx context.Context, sub AccountSubscription) error
	UpdateSubscription(ctx context.Context, sub AccountSubscription) error
	WithExecutor(exec Executor) AccountSubscriptionAccessor
}

type accountSubscriptionAccessor struct {
	exec   Executor
	logger *zap.Logger
}

func NewAccountSubscriptionAccessor(
	exec Executor,
	logger *zap.Logger,
) AccountSubscriptionAccessor {
	return &accountSubscriptionAccessor{
		exec:   exec,
		logger: logger,
	}
}

func (a accountSubscriptionAccessor) GetSubscriptionByAccountId(
	ctx context.Context,
	accountId uint64,
) (AccountSubscription, error) {
	if accountId == 0 {
		return AccountSubscription{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("account_id", accountId))
	const query = `SELECT of_account_id, plan, billing_period, status, expires_at, created_at, updated_at
		FROM account_subscriptions
		WHERE of_account_id = ?`

	var sub AccountSubscription
	err := a.exec.QueryRowContext(ctx, query, accountId).Scan(
		&sub.OfAccountId,
		&sub.Plan,
		&sub.BillingPeriod,
		&sub.Status,
		&sub.ExpiresAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get subscription by account id")
		return AccountSubscription{}, err
	}

	return sub, nil
}

func (a accountSubscriptionAccessor) CreateSubscription(
	ctx context.Context,
	sub AccountSubscription,
) error {
	if sub.OfAccountId == 0 {
		return ErrLackOfInfor
	}

	plan := sub.Plan
	if plan == "" {
		plan = "free"
	}
	status := sub.Status
	if status == "" {
		status = "active"
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("account_id", sub.OfAccountId))
	const query = `INSERT INTO account_subscriptions (of_account_id, plan, status) VALUES (?, ?, ?)`
	_, err := a.exec.ExecContext(ctx, query, sub.OfAccountId, plan, status)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create subscription")
		return err
	}

	return nil
}

func (a accountSubscriptionAccessor) WithExecutor(
	exec Executor,
) AccountSubscriptionAccessor {
	return &accountSubscriptionAccessor{
		exec:   exec,
		logger: a.logger,
	}
}

func (a accountSubscriptionAccessor) UpdateSubscription(
	ctx context.Context,
	sub AccountSubscription,
) error {
	if sub.OfAccountId == 0 {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Uint64("account_id", sub.OfAccountId))
	const query = `UPDATE account_subscriptions
		SET plan = ?, billing_period = ?, status = ?, expires_at = ?
		WHERE of_account_id = ?`
	_, err := a.exec.ExecContext(ctx, query,
		sub.Plan, sub.BillingPeriod, sub.Status, sub.ExpiresAt, sub.OfAccountId,
	)
	if err != nil {
		logger.Error("failed to update subscription", zap.Error(err))
		return err
	}
	return nil
}
