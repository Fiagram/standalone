package dao_database

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

type AccountPassword struct {
	OfAccountId  uint64    `json:"of_account_id"`
	HashedString string    `json:"hashed_string"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AccountPasswordAccessor interface {
	CreateAccountPassword(ctx context.Context, ap AccountPassword) error
	GetAccountPassword(ctx context.Context, id uint64) (AccountPassword, error)
	UpdateAccountPassword(ctx context.Context, ap AccountPassword) error
	DeleteAccountPassword(ctx context.Context, id uint64) error
	WithExecutor(exec Executor) AccountPasswordAccessor
}

type accountPasswordAccessor struct {
	exec   Executor
	logger *zap.Logger
}

func NewAccountPasswordAccessor(
	exec Executor,
	logger *zap.Logger,
) AccountPasswordAccessor {
	return &accountPasswordAccessor{
		exec:   exec,
		logger: logger,
	}
}

func (a accountPasswordAccessor) CreateAccountPassword(
	ctx context.Context,
	ap AccountPassword,
) error {
	if ap.OfAccountId == 0 {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("of_account_id", ap.OfAccountId))
	if ap.OfAccountId == 0 {
		return ErrLackOfInfor
	}
	const query = `INSERT INTO account_passwords 
		(of_account_id, hashed_string)
		VALUES (?, ?)`

	result, err := a.exec.ExecContext(ctx, query,
		ap.OfAccountId,
		strings.TrimSpace(ap.HashedString),
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to insert account password")
		return err
	}

	rowEfNum, err := result.RowsAffected()
	if rowEfNum != 1 || err != nil {
		errMsg := "failed to effect row"
		logger.With(zap.Int64("rowEfNum", rowEfNum)).
			With(zap.Error(err)).
			Error(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (a accountPasswordAccessor) GetAccountPassword(
	ctx context.Context,
	id uint64,
) (AccountPassword, error) {
	if id == 0 {
		return AccountPassword{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("of_account_id", id))
	const query = `SELECT * FROM account_passwords WHERE of_account_id = ?`
	row := a.exec.QueryRowContext(ctx, query, id)
	var out AccountPassword
	err := row.Scan(&out.OfAccountId,
		&out.HashedString,
		&out.CreatedAt,
		&out.UpdatedAt)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get password")
		return AccountPassword{}, err
	}

	return out, nil
}

func (a accountPasswordAccessor) DeleteAccountPassword(
	ctx context.Context,
	id uint64,
) error {
	if id == 0 {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("of_account_id", id))
	if id == 0 {
		return ErrLackOfInfor
	}
	const query = `DELETE FROM account_passwords 
			WHERE of_account_id = ?`
	result, err := a.exec.ExecContext(ctx, query, id)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to delete password")
		return err
	}

	rowEfNum, err := result.RowsAffected()
	if rowEfNum != 1 || err != nil {
		errMsg := "failed to effect row"
		logger.With(zap.Int64("rowEfNum", rowEfNum)).
			With(zap.Error(err)).
			Error(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (a accountPasswordAccessor) UpdateAccountPassword(
	ctx context.Context,
	ap AccountPassword,
) error {
	if ap.OfAccountId == 0 && ap.HashedString == "" {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("of_account_id", ap.OfAccountId))
	const query = `UPDATE account_passwords SET 
			hashed_string = ?  
			WHERE of_account_id = ?`
	result, err := a.exec.ExecContext(ctx, query,
		strings.TrimSpace(ap.HashedString),
		ap.OfAccountId,
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update password")
		return err
	}

	rowEfNum, err := result.RowsAffected()
	if rowEfNum != 1 || err != nil {
		errMsg := "failed to effect row"
		logger.With(zap.Int64("rowEfNum", rowEfNum)).
			With(zap.Error(err)).
			Error(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (a accountPasswordAccessor) WithExecutor(
	exec Executor,
) AccountPasswordAccessor {
	return &accountPasswordAccessor{
		exec:   exec,
		logger: a.logger,
	}
}
