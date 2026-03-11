package dao_database_test

import (
	"context"
	"testing"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"github.com/stretchr/testify/require"
)

func TestCreateAndDeleteAccountPassword(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	pAsor := dao_database.NewAccountPasswordAccessor(sqlDb, logger)
	ctx := context.Background()

	acc := RandomAccount()
	id, err := aAsor.CreateAccount(ctx, acc)
	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NotZero(t, id)

	input := dao_database.AccountPassword{
		OfAccountId:  id,
		HashedString: RandomString(128),
	}
	require.NoError(t, pAsor.CreateAccountPassword(ctx, input))

	require.NoError(t, pAsor.DeleteAccountPassword(ctx, id))
	require.NoError(t, aAsor.DeleteAccount(ctx, id))
}

func TestGetAccountPassword(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	pAsor := dao_database.NewAccountPasswordAccessor(sqlDb, logger)
	ctx := context.Background()

	acc := RandomAccount()
	id, err := aAsor.CreateAccount(ctx, acc)
	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NotZero(t, id)

	input := dao_database.AccountPassword{
		OfAccountId:  id,
		HashedString: RandomString(128),
	}
	require.NoError(t, pAsor.CreateAccountPassword(ctx, input))
	output, err := pAsor.GetAccountPassword(ctx, id)
	require.NoError(t, err)
	require.Equal(t, input.OfAccountId, output.OfAccountId)
	require.Equal(t, input.HashedString, output.HashedString)

	require.NoError(t, pAsor.DeleteAccountPassword(ctx, id))
	require.NoError(t, aAsor.DeleteAccount(ctx, id))
}

func TestUpdateAccountPassword(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	pAsor := dao_database.NewAccountPasswordAccessor(sqlDb, logger)
	ctx := context.Background()

	acc := RandomAccount()
	id, err := aAsor.CreateAccount(ctx, acc)
	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NotZero(t, id)

	input := dao_database.AccountPassword{
		OfAccountId:  id,
		HashedString: RandomString(128),
	}
	require.NoError(t, pAsor.CreateAccountPassword(ctx, input))

	updatedInput := input
	updatedInput.HashedString = RandomString(128)
	require.NoError(t, pAsor.UpdateAccountPassword(ctx, updatedInput))

	output, err := pAsor.GetAccountPassword(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedInput.OfAccountId, output.OfAccountId)
	require.Equal(t, updatedInput.HashedString, output.HashedString)

	require.NoError(t, pAsor.DeleteAccountPassword(ctx, id))
	require.NoError(t, aAsor.DeleteAccount(ctx, id))
}
