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

type accountRoleList []AccountRole

var accountRoleCache accountRoleList

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

func (c *accountRoleList) fetchAccountRoles(ctx context.Context, a accountRoleAccessor) error {
	logger := logger.LoggerWithContext(ctx, a.logger)
	*c = (*c)[:0]
	const query = `SELECT id, name FROM account_role`
	rows, err := a.exec.QueryContext(ctx, query)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to fetch account role list")
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item AccountRole
		err = rows.Scan(&item.Id, &item.Name)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to scan account role row")
			return err
		}
		*c = append(*c, item)
	}

	return nil
}

func (a accountRoleAccessor) GetRoleById(
	ctx context.Context,
	id uint8,
) (AccountRole, error) {
	if accountRoleCache == nil {
		err := accountRoleCache.fetchAccountRoles(ctx, a)
		if err != nil {
			return AccountRole{}, err
		}
	}

	if id == 0 {
		return AccountRole{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("role_id", id))
	for _, item := range accountRoleCache {
		if item.Id == id {
			return item, nil
		}
	}
	logger.With(zap.Error(ErrAccRoleNotFound))
	return AccountRole{}, ErrAccRoleNotFound
}

func (a accountRoleAccessor) GetRoleByName(
	ctx context.Context,
	name string,
) (AccountRole, error) {
	if accountRoleCache == nil {
		err := accountRoleCache.fetchAccountRoles(ctx, a)
		if err != nil {
			return AccountRole{}, err
		}
	}

	if name == "" {
		return AccountRole{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("role_name", name))
	for _, item := range accountRoleCache {
		if name == item.Name {
			return item, nil
		}
	}
	logger.With(zap.Error(ErrAccRoleNotFound))
	return AccountRole{}, ErrAccRoleNotFound
}

func (a accountRoleAccessor) WithExecutor(
	exec Executor,
) AccountRoleAccessor {
	return &accountRoleAccessor{
		exec:   exec,
		logger: a.logger,
	}
}
