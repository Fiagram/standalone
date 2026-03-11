package dao_database

import (
	"context"

	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

type AccountRole struct {
	Id   uint8  `json:"id"`
	Name string `json:"name"`
}

type AccountRoleAccessor interface {
	GetRoleById(ctx context.Context, id uint8) (AccountRole, error)
	GetRoleByName(ctx context.Context, name string) (AccountRole, error)
	WithExecutor(exec Executor) AccountRoleAccessor
}

type accountRoleAccessor struct {
	exec   Executor
	logger *zap.Logger
}

func NewAccountRoleAccessor(
	exec Executor,
	logger *zap.Logger,
) AccountRoleAccessor {
	return &accountRoleAccessor{
		exec:   exec,
		logger: logger,
	}
}

func (a accountRoleAccessor) GetRoleById(
	ctx context.Context,
	id uint8,
) (AccountRole, error) {
	if id == 0 {
		return AccountRole{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("role_id", id))
	const query = `SELECT id, name FROM account_role WHERE id = ?`
	row := a.exec.QueryRowContext(ctx, query, id)

	var out AccountRole
	err := row.Scan(&out.Id, &out.Name)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account row by id")
		return AccountRole{}, err
	}

	return out, nil
}

func (a accountRoleAccessor) GetRoleByName(
	ctx context.Context,
	name string,
) (AccountRole, error) {
	if name == "" {
		return AccountRole{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("role_name", name))
	const query = `SELECT id, name FROM account_role WHERE name = ?`
	row := a.exec.QueryRowContext(ctx, query, name)

	var out AccountRole
	err := row.Scan(&out.Id, &out.Name)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account row by name")
		return AccountRole{}, err
	}

	return out, nil
}

func (a accountRoleAccessor) WithExecutor(
	exec Executor,
) AccountRoleAccessor {
	return &accountRoleAccessor{
		exec:   exec,
		logger: a.logger,
	}
}
