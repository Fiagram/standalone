package dao_database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

type Account struct {
	Id          uint64    `json:"id"`
	Username    string    `json:"username"`
	Fullname    string    `json:"fullname"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	RoleId      uint8     `json:"role_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AccountAccessor interface {
	CreateAccount(ctx context.Context, account Account) (uint64, error)

	GetAccount(ctx context.Context, id uint64) (Account, error)
	GetAccountByUsername(ctx context.Context, username string) (Account, error)

	UpdateAccount(ctx context.Context, account Account) error

	DeleteAccount(ctx context.Context, id uint64) error
	DeleteAccountByUsername(ctx context.Context, username string) error

	IsUsernameTaken(ctx context.Context, username string) (bool, error)

	GetAccountAll(ctx context.Context) ([]Account, error)
	GetAccountList(ctx context.Context, ids []uint64) ([]Account, error)

	WithExecutor(exec Executor) AccountAccessor
}

type accountAccessor struct {
	exec   Executor
	logger *zap.Logger
}

func NewAccountAccessor(
	exec Executor,
	logger *zap.Logger,
) AccountAccessor {
	return &accountAccessor{
		exec:   exec,
		logger: logger,
	}
}

func (a accountAccessor) CreateAccount(
	ctx context.Context,
	acc Account,
) (uint64, error) {
	if acc.Username == "" &&
		acc.RoleId == 0 {
		return 0, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("account", acc))
	const query = `INSERT INTO accounts 
			(username, fullname, email, phone_number, role_id) 
			VALUES (?, ?, ?, ?, ?)`
	result, err := a.exec.ExecContext(ctx, query,
		strings.TrimSpace(acc.Username),
		strings.TrimSpace(acc.Fullname),
		strings.TrimSpace(acc.Email),
		strings.TrimSpace(acc.PhoneNumber),
		acc.RoleId,
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create account")
		return 0, err
	}

	rowEfNum, err := result.RowsAffected()
	if rowEfNum != 1 || err != nil {
		errMsg := "failed to effect row"
		logger.With(zap.Int64("rowEfNum", rowEfNum)).
			With(zap.Error(err)).
			Error(errMsg)
		return 0, errors.New(errMsg)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get last inserted id")
		return 0, err
	}

	return uint64(lastInsertedId), nil
}

func (a accountAccessor) GetAccount(
	ctx context.Context,
	id uint64,
) (Account, error) {
	if id == 0 {
		return Account{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("account_id", id))
	const query = `SELECT * FROM accounts WHERE id = ?`
	row := a.exec.QueryRowContext(ctx, query, id)

	var out Account
	err := row.Scan(&out.Id,
		&out.Username,
		&out.Fullname,
		&out.Email,
		&out.PhoneNumber,
		&out.RoleId,
		&out.CreatedAt,
		&out.UpdatedAt)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account by id")
		return Account{}, err
	}

	return out, nil
}

func (a accountAccessor) GetAccountByUsername(
	ctx context.Context,
	username string,
) (Account, error) {
	if username == "" {
		return Account{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("username", username))
	const query = `SELECT * FROM accounts WHERE username = ?`
	row := a.exec.QueryRowContext(ctx, query, username)

	var out Account
	err := row.Scan(&out.Id,
		&out.Username,
		&out.Fullname,
		&out.Email,
		&out.PhoneNumber,
		&out.RoleId,
		&out.CreatedAt,
		&out.UpdatedAt)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account by username")
		return Account{}, err
	}

	return out, nil
}

func (a accountAccessor) DeleteAccount(
	ctx context.Context,
	id uint64,
) error {
	if id == 0 {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("account_id", id))
	const query = `DELETE FROM accounts WHERE id = ?`
	result, err := a.exec.ExecContext(ctx, query, id)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to delete account")
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

func (a accountAccessor) DeleteAccountByUsername(
	ctx context.Context,
	username string,
) error {
	if username == "" {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("username", username))
	const query = `DELETE FROM accounts WHERE username = ?`
	result, err := a.exec.ExecContext(ctx, query, username)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to delete account")
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

func (a accountAccessor) UpdateAccount(
	ctx context.Context,
	acc Account,
) error {
	if acc.Username == "" {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("account", acc))
	const query = `UPDATE accounts SET 
			fullname = ?, 
			email = ?, 
			phone_number = ?, 
			role_id = ? 
			WHERE username = ?`

	result, err := a.exec.ExecContext(ctx, query,
		strings.TrimSpace(acc.Fullname),
		strings.TrimSpace(acc.Email),
		strings.TrimSpace(acc.PhoneNumber),
		acc.RoleId,
		strings.TrimSpace(acc.Username),
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update account")
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

func (a accountAccessor) IsUsernameTaken(
	ctx context.Context,
	username string,
) (bool, error) {
	if username == "" {
		return false, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("username", username))
	const query = `SELECT EXISTS(SELECT 1 FROM accounts WHERE username = ?) AS is_taken`
	var isTaken int
	err := a.exec.QueryRowContext(ctx, query, username).Scan(&isTaken)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to check username taken")
		return false, err
	}

	if isTaken == 1 {
		return true, nil
	}
	return false, nil
}

func (a accountAccessor) GetAccountAll(
	ctx context.Context,
) ([]Account, error) {
	logger := logger.LoggerWithContext(ctx, a.logger)
	const query = `SELECT * FROM accounts`
	rows, err := a.exec.QueryContext(ctx, query)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get all accounts")
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		err := rows.Scan(&acc.Id,
			&acc.Username,
			&acc.Fullname,
			&acc.Email,
			&acc.PhoneNumber,
			&acc.RoleId,
			&acc.CreatedAt,
			&acc.UpdatedAt)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to scan account")
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}

func (a accountAccessor) GetAccountList(
	ctx context.Context,
	ids []uint64,
) ([]Account, error) {
	if len(ids) == 0 {
		return []Account{}, fmt.Errorf("ids is empty")
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("ids", ids))
	query := `SELECT * FROM accounts WHERE id IN (?` + strings.Repeat(",?", len(ids)-1) + `)`

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := a.exec.QueryContext(ctx, query, args...)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account list")
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		err := rows.Scan(&acc.Id,
			&acc.Username,
			&acc.Fullname,
			&acc.Email,
			&acc.PhoneNumber,
			&acc.RoleId,
			&acc.CreatedAt,
			&acc.UpdatedAt)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to scan account")
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}

func (a accountAccessor) WithExecutor(
	exec Executor,
) AccountAccessor {
	return &accountAccessor{
		exec:   exec,
		logger: a.logger,
	}
}
