package dao_database

import (
	"context"
	"time"

	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

type SubscriptionPlan struct {
	Plan          string
	BillingPeriod string
	Price         float64
	Currency      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SubscriptionPlanAccessor interface {
	GetPlan(ctx context.Context, plan, billingPeriod string) (SubscriptionPlan, error)
	ListPlans(ctx context.Context) ([]SubscriptionPlan, error)
	WithExecutor(exec Executor) SubscriptionPlanAccessor
}

type subscriptionPlanAccessor struct {
	exec   Executor
	logger *zap.Logger
}

func NewSubscriptionPlanAccessor(
	exec Executor,
	logger *zap.Logger,
) SubscriptionPlanAccessor {
	return &subscriptionPlanAccessor{
		exec:   exec,
		logger: logger,
	}
}

func (a *subscriptionPlanAccessor) GetPlan(
	ctx context.Context,
	plan, billingPeriod string,
) (SubscriptionPlan, error) {
	logger := logger.LoggerWithContext(ctx, a.logger).With(
		zap.String("plan", plan),
		zap.String("billing_period", billingPeriod),
	)

	const query = `SELECT plan, billing_period, price, currency, created_at, updated_at
		FROM subscription_plans
		WHERE plan = ? AND billing_period = ?`

	var p SubscriptionPlan
	err := a.exec.QueryRowContext(ctx, query, plan, billingPeriod).Scan(
		&p.Plan,
		&p.BillingPeriod,
		&p.Price,
		&p.Currency,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		logger.Error("failed to get subscription plan", zap.Error(err))
		return SubscriptionPlan{}, err
	}

	return p, nil
}

func (a *subscriptionPlanAccessor) ListPlans(ctx context.Context) ([]SubscriptionPlan, error) {
	logger := logger.LoggerWithContext(ctx, a.logger)

	const query = `SELECT plan, billing_period, price, currency, created_at, updated_at
		FROM subscription_plans
		ORDER BY plan, billing_period`

	rows, err := a.exec.QueryContext(ctx, query)
	if err != nil {
		logger.Error("failed to list subscription plans", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var plans []SubscriptionPlan
	for rows.Next() {
		var p SubscriptionPlan
		if err := rows.Scan(
			&p.Plan,
			&p.BillingPeriod,
			&p.Price,
			&p.Currency,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			logger.Error("failed to scan subscription plan row", zap.Error(err))
			return nil, err
		}
		plans = append(plans, p)
	}
	if err := rows.Err(); err != nil {
		logger.Error("rows error listing subscription plans", zap.Error(err))
		return nil, err
	}

	return plans, nil
}

func (a *subscriptionPlanAccessor) WithExecutor(exec Executor) SubscriptionPlanAccessor {
	return &subscriptionPlanAccessor{
		exec:   exec,
		logger: a.logger,
	}
}
