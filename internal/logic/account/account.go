package logic_account

import (
	"context"
	"database/sql"
	"fmt"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"go.uber.org/zap"
)

type Account interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error)

	CheckAccountValid(ctx context.Context, params CheckAccountValidParams) (CheckAccountValidOutput, error)

	IsUsernameTaken(ctx context.Context, params IsUsernameTakenParams) (IsUsernameTakenOutput, error)

	GetAccount(ctx context.Context, params GetAccountParams) (GetAccountOutput, error)
	GetAccountAll(ctx context.Context, params GetAccountAllParams) (GetAccountAllOutput, error)
	GetAccountList(ctx context.Context, params GetAccountListParams) (GetAccountListOutput, error)

	UpdateAccountInfo(ctx context.Context, params UpdateAccountInfoParams) (UpdateAccountInfoOutput, error)
	UpdateAccountPassword(ctx context.Context, params UpdateAccountPasswordParams) (UpdateAccountPasswordOutput, error)

	DeleteAccount(ctx context.Context, params DeleteAccountParams) error
	DeleteAccountByUsername(ctx context.Context, params DeleteAccountByUsernameParams) error
}

type account struct {
	db                      *sql.DB
	accountAccessor         dao_database.AccountAccessor
	accountPasswordAccessor dao_database.AccountPasswordAccessor
	subscriptionAccessor    dao_database.AccountSubscriptionAccessor
	hashLogic               Hash
	logger                  *zap.Logger
}

func NewAccount(
	db *sql.DB,
	accountAccessor dao_database.AccountAccessor,
	accountPasswordAccessor dao_database.AccountPasswordAccessor,
	subscriptionAccessor dao_database.AccountSubscriptionAccessor,
	hashLogic Hash,
	logger *zap.Logger,
) Account {
	return &account{
		db:                      db,
		accountAccessor:         accountAccessor,
		accountPasswordAccessor: accountPasswordAccessor,
		subscriptionAccessor:    subscriptionAccessor,
		hashLogic:               hashLogic,
		logger:                  logger,
	}
}

func (a account) CreateAccount(
	ctx context.Context,
	params CreateAccountParams,
) (CreateAccountOutput, error) {
	emptyOutput := CreateAccountOutput{}
	isUsernameTaken, err := a.accountAccessor.IsUsernameTaken(ctx, params.AccountInfo.Username)
	if err != nil {
		return emptyOutput, fmt.Errorf("failed to check if username taken: %w", err)
	} else if isUsernameTaken {
		return emptyOutput, fmt.Errorf("username has already been taken")
	}

	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return emptyOutput, ErrTxBeginFailed
	}
	defer tx.Rollback()

	id, err := a.accountAccessor.
		WithExecutor(tx).
		CreateAccount(ctx, dao_database.Account{
			Username:    params.AccountInfo.Username,
			Fullname:    params.AccountInfo.Fullname,
			Email:       params.AccountInfo.Email,
			PhoneNumber: params.AccountInfo.PhoneNumber,
			RoleId:      uint8(params.AccountInfo.Role),
		})
	if err != nil {
		return emptyOutput, fmt.Errorf("failed to create new account: %w", err)
	}

	hashedString, err := a.hashLogic.Hash(ctx, params.Password)
	if err != nil {
		return emptyOutput, fmt.Errorf("failed to hash password: %w", err)
	}

	err = a.accountPasswordAccessor.
		WithExecutor(tx).
		CreateAccountPassword(ctx, dao_database.AccountPassword{
			OfAccountId:  id,
			HashedString: hashedString,
		})
	if err != nil {
		return emptyOutput, fmt.Errorf("failed to create new password: %w", err)
	}

	err = a.subscriptionAccessor.
		WithExecutor(tx).
		CreateSubscription(ctx, dao_database.AccountSubscription{
			OfAccountId: id,
			Plan:        "free",
			Status:      "active",
		})
	if err != nil {
		return emptyOutput, fmt.Errorf("failed to create subscription: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return emptyOutput, ErrTxCommitFailed
	}

	return CreateAccountOutput{
		AccountId: id,
	}, nil
}

func (a account) DeleteAccount(
	ctx context.Context,
	params DeleteAccountParams,
) error {
	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ErrTxBeginFailed
	}
	defer tx.Rollback()

	err = a.accountPasswordAccessor.
		WithExecutor(tx).
		DeleteAccountPassword(ctx, params.AccountId)
	if err != nil {
		return fmt.Errorf("failed to delete password: %w", err)
	}
	err = a.accountAccessor.
		WithExecutor(tx).
		DeleteAccount(ctx, params.AccountId)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return ErrTxCommitFailed
	}

	return nil
}

func (a account) DeleteAccountByUsername(
	ctx context.Context,
	params DeleteAccountByUsernameParams,
) error {
	isExisted, err := a.accountAccessor.IsUsernameTaken(ctx, params.Username)
	if err != nil {
		return fmt.Errorf("failed to check if username existed: %w", err)
	} else if !isExisted {
		return fmt.Errorf("username does not existed")
	}

	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return ErrTxBeginFailed
	}
	defer tx.Rollback()
	acc, err := a.accountAccessor.
		WithExecutor(tx).
		GetAccountByUsername(ctx, params.Username)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}
	err = a.accountPasswordAccessor.
		WithExecutor(tx).
		DeleteAccountPassword(ctx, acc.Id)
	if err != nil {
		return fmt.Errorf("failed to delete password: %w", err)
	}
	err = a.accountAccessor.
		WithExecutor(tx).
		DeleteAccount(ctx, acc.Id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return ErrTxCommitFailed
	}

	return nil
}

func (a account) CheckAccountValid(
	ctx context.Context,
	params CheckAccountValidParams,
) (CheckAccountValidOutput, error) {
	emptyObj := CheckAccountValidOutput{}
	acc, err := a.accountAccessor.
		GetAccountByUsername(ctx, params.Username)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to get account: %w", err)
	}

	truly, err := a.accountPasswordAccessor.
		GetAccountPassword(ctx, acc.Id)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to get password: %w", err)
	}

	isValid, err := a.hashLogic.
		IsHashEqual(ctx, params.Password, truly.HashedString)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to check if hashes are equal: %w", err)
	}

	if !isValid {
		return emptyObj, nil
	}
	return CheckAccountValidOutput{
		AccountId: acc.Id,
	}, nil
}

func (a account) IsUsernameTaken(
	ctx context.Context,
	params IsUsernameTakenParams,
) (IsUsernameTakenOutput, error) {
	emptyObj := IsUsernameTakenOutput{}
	isTaken, err := a.accountAccessor.IsUsernameTaken(ctx, params.Username)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to check username is taken: %w", err)
	}
	return IsUsernameTakenOutput{
		IsTaken: isTaken,
	}, nil
}

func (a account) GetAccount(
	ctx context.Context,
	params GetAccountParams,
) (GetAccountOutput, error) {
	emptyObj := GetAccountOutput{}
	acc, err := a.accountAccessor.GetAccount(ctx, params.AccountId)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to get account: %w", err)
	}

	return GetAccountOutput{
		AccountId: acc.Id,
		AccountInfo: AccountInfo{
			Username:    acc.Username,
			Fullname:    acc.Fullname,
			Email:       acc.Email,
			PhoneNumber: acc.PhoneNumber,
			Role:        Role(acc.RoleId),
		},
	}, nil
}

func (a account) GetAccountAll(
	ctx context.Context,
	params GetAccountAllParams,
) (GetAccountAllOutput, error) {
	emptyObj := GetAccountAllOutput{}
	accs, err := a.accountAccessor.GetAccountAll(ctx)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to get all accounts: %w", err)
	}

	accountIds := make([]uint64, 0, len(accs))
	accountInfos := make([]AccountInfo, 0, len(accs))

	for _, acc := range accs {
		accountIds = append(accountIds, acc.Id)
		accountInfos = append(accountInfos, AccountInfo{
			Username:    acc.Username,
			Fullname:    acc.Fullname,
			Email:       acc.Email,
			PhoneNumber: acc.PhoneNumber,
			Role:        Role(acc.RoleId),
		})
	}

	return GetAccountAllOutput{
		AccountIds:   accountIds,
		AccountInfos: accountInfos,
	}, nil
}

func (a account) GetAccountList(
	ctx context.Context,
	params GetAccountListParams,
) (GetAccountListOutput, error) {
	emptyObj := GetAccountListOutput{}
	accs, err := a.accountAccessor.GetAccountList(ctx, params.AccountIds)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to get account list: %w", err)
	}

	accountIds := make([]uint64, 0, len(accs))
	accountInfos := make([]AccountInfo, 0, len(accs))

	for _, acc := range accs {
		accountIds = append(accountIds, acc.Id)
		accountInfos = append(accountInfos, AccountInfo{
			Username:    acc.Username,
			Fullname:    acc.Fullname,
			Email:       acc.Email,
			PhoneNumber: acc.PhoneNumber,
			Role:        Role(acc.RoleId),
		})
	}

	return GetAccountListOutput{
		AccountIds:   accountIds,
		AccountInfos: accountInfos,
	}, nil
}

func (a account) UpdateAccountInfo(
	ctx context.Context,
	params UpdateAccountInfoParams,
) (UpdateAccountInfoOutput, error) {
	emptyObj := UpdateAccountInfoOutput{}

	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return emptyObj, ErrTxBeginFailed
	}
	defer tx.Rollback()

	acc, err := a.accountAccessor.
		WithExecutor(tx).
		GetAccount(ctx, params.AccountId)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to get account: %w", err)
	}

	// Update fields
	acc.Fullname = params.UpdatedAccountInfo.Fullname
	acc.Email = params.UpdatedAccountInfo.Email
	acc.PhoneNumber = params.UpdatedAccountInfo.PhoneNumber
	acc.RoleId = uint8(params.UpdatedAccountInfo.Role)
	// Username usually shouldn't be updated, or special care taken if allowed.
	// Assuming params.UpdatedAccountInfo doesn't carry ID/Username for update target, but here we updating 'acc' found by ID.
	// The UpdateAccount accessor method uses Username to find record to update?
	// Let's check internal/dataaccess/database/account.go: UpdateAccount uses `WHERE username = ?` and takes `acc.Username`.
	// So we must ensure `acc.Username` is preserved. It is, since we fetched it.

	err = a.accountAccessor.
		WithExecutor(tx).
		UpdateAccount(ctx, acc)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to update account: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return emptyObj, ErrTxCommitFailed
	}

	return UpdateAccountInfoOutput{
		AccountId: acc.Id,
	}, nil
}

func (a account) UpdateAccountPassword(
	ctx context.Context,
	params UpdateAccountPasswordParams,
) (UpdateAccountPasswordOutput, error) {
	emptyObj := UpdateAccountPasswordOutput{}

	hashedString, err := a.hashLogic.Hash(ctx, params.Password)
	if err != nil {
		return emptyObj, fmt.Errorf("failed to hash password: %w", err)
	}

	err = a.accountPasswordAccessor.
		UpdateAccountPassword(ctx, dao_database.AccountPassword{
			OfAccountId:  params.AccountId,
			HashedString: hashedString,
		})
	if err != nil {
		return emptyObj, fmt.Errorf("failed to update account password: %w", err)
	}

	return UpdateAccountPasswordOutput{
		AccountId: params.AccountId,
	}, nil
}
