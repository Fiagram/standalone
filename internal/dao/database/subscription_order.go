package dao_database

import (
	"context"
	"time"

	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

type SubscriptionOrder struct {
	Id                 uint64
	OfAccountId        uint64
	Plan               string
	BillingPeriod      string
	Amount             float64
	Currency           string
	Status             string
	ReferenceCode      string
	SePayTransactionId *string
	PaymentExpiresAt   time.Time
	SubStartAt         *time.Time
	SubExpiresAt       *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SubscriptionOrderAccessor interface {
	CreateOrder(ctx context.Context, order SubscriptionOrder) (id uint64, err error)
	GetOrderById(ctx context.Context, id uint64) (SubscriptionOrder, error)
	GetOrderByReferenceCode(ctx context.Context, referenceCode string) (SubscriptionOrder, error)
	ListOrdersByAccountId(ctx context.Context, accountId uint64, limit, offset int) ([]SubscriptionOrder, error)
	UpdateOrderStatus(ctx context.Context, id uint64, status, transactionId string) error
	WithExecutor(exec Executor) SubscriptionOrderAccessor
}

type subscriptionOrderAccessor struct {
	exec   Executor
	logger *zap.Logger
}

func NewSubscriptionOrderAccessor(
	exec Executor,
	logger *zap.Logger,
) SubscriptionOrderAccessor {
	return &subscriptionOrderAccessor{
		exec:   exec,
		logger: logger,
	}
}

func (a *subscriptionOrderAccessor) CreateOrder(
	ctx context.Context,
	order SubscriptionOrder,
) (uint64, error) {
	if order.OfAccountId == 0 {
		return 0, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Uint64("account_id", order.OfAccountId))
	const query = `INSERT INTO subscription_orders
		(of_account_id, plan, billing_period, amount, currency, status, reference_code, payment_expires_at)
		VALUES (?, ?, ?, ?, ?, 'pending', ?, ?)`

	result, err := a.exec.ExecContext(ctx, query,
		order.OfAccountId,
		order.Plan,
		order.BillingPeriod,
		order.Amount,
		order.Currency,
		order.ReferenceCode,
		order.PaymentExpiresAt,
	)
	if err != nil {
		logger.Error("failed to create subscription order", zap.Error(err))
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (a *subscriptionOrderAccessor) GetOrderById(
	ctx context.Context,
	id uint64,
) (SubscriptionOrder, error) {
	if id == 0 {
		return SubscriptionOrder{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Uint64("order_id", id))
	const query = `SELECT id, of_account_id, plan, billing_period, amount, currency,
		status, reference_code, sepay_transaction_id, payment_expires_at,
		sub_start_at, sub_expires_at, created_at, updated_at
		FROM subscription_orders WHERE id = ?`

	var o SubscriptionOrder
	err := a.exec.QueryRowContext(ctx, query, id).Scan(
		&o.Id, &o.OfAccountId, &o.Plan, &o.BillingPeriod, &o.Amount, &o.Currency,
		&o.Status, &o.ReferenceCode, &o.SePayTransactionId, &o.PaymentExpiresAt,
		&o.SubStartAt, &o.SubExpiresAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		logger.Error("failed to get subscription order by id", zap.Error(err))
		return SubscriptionOrder{}, err
	}
	return o, nil
}

func (a *subscriptionOrderAccessor) GetOrderByReferenceCode(
	ctx context.Context,
	referenceCode string,
) (SubscriptionOrder, error) {
	if referenceCode == "" {
		return SubscriptionOrder{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.String("reference_code", referenceCode))
	const query = `SELECT id, of_account_id, plan, billing_period, amount, currency,
		status, reference_code, sepay_transaction_id, payment_expires_at,
		sub_start_at, sub_expires_at, created_at, updated_at
		FROM subscription_orders WHERE reference_code = ?`

	var o SubscriptionOrder
	err := a.exec.QueryRowContext(ctx, query, referenceCode).Scan(
		&o.Id, &o.OfAccountId, &o.Plan, &o.BillingPeriod, &o.Amount, &o.Currency,
		&o.Status, &o.ReferenceCode, &o.SePayTransactionId, &o.PaymentExpiresAt,
		&o.SubStartAt, &o.SubExpiresAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		logger.Error("failed to get subscription order by reference code", zap.Error(err))
		return SubscriptionOrder{}, err
	}
	return o, nil
}

func (a *subscriptionOrderAccessor) ListOrdersByAccountId(
	ctx context.Context,
	accountId uint64,
	limit, offset int,
) ([]SubscriptionOrder, error) {
	if accountId == 0 {
		return nil, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Uint64("account_id", accountId))
	const query = `SELECT id, of_account_id, plan, billing_period, amount, currency,
		status, reference_code, sepay_transaction_id, payment_expires_at,
		sub_start_at, sub_expires_at, created_at, updated_at
		FROM subscription_orders
		WHERE of_account_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := a.exec.QueryContext(ctx, query, accountId, limit, offset)
	if err != nil {
		logger.Error("failed to list subscription orders", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []SubscriptionOrder
	for rows.Next() {
		var o SubscriptionOrder
		if err := rows.Scan(
			&o.Id, &o.OfAccountId, &o.Plan, &o.BillingPeriod, &o.Amount, &o.Currency,
			&o.Status, &o.ReferenceCode, &o.SePayTransactionId, &o.PaymentExpiresAt,
			&o.SubStartAt, &o.SubExpiresAt, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			logger.Error("failed to scan subscription order row", zap.Error(err))
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		logger.Error("rows error listing subscription orders", zap.Error(err))
		return nil, err
	}
	return orders, nil
}

func (a *subscriptionOrderAccessor) UpdateOrderStatus(
	ctx context.Context,
	id uint64,
	status, transactionId string,
) error {
	if id == 0 {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(
		zap.Uint64("order_id", id),
		zap.String("status", status),
	)

	const query = `UPDATE subscription_orders
		SET status = ?,
		    sepay_transaction_id = CASE WHEN ? != '' THEN ? ELSE sepay_transaction_id END,
		    sub_start_at = CASE WHEN ? = 'paid' THEN NOW() ELSE sub_start_at END
		WHERE id = ?`

	_, err := a.exec.ExecContext(ctx, query,
		status,
		transactionId, transactionId,
		status,
		id,
	)
	if err != nil {
		logger.Error("failed to update subscription order status", zap.Error(err))
		return err
	}
	return nil
}

func (a *subscriptionOrderAccessor) WithExecutor(exec Executor) SubscriptionOrderAccessor {
	return &subscriptionOrderAccessor{
		exec:   exec,
		logger: a.logger,
	}
}
