package dao_database_test

import (
	"context"
	"testing"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"

	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	input := RandomAccount()
	id, err := aAsor.CreateAccount(context.Background(), input)

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NotZero(t, id)

	errD := aAsor.DeleteAccountByUsername(context.Background(), input.Username)
	require.NoError(t, errD)
}

func TestGetAccountById(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	input := RandomAccount()
	id, err := aAsor.CreateAccount(context.Background(), input)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	acc, err := aAsor.GetAccount(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, acc)
	require.Equal(t, input.Username, acc.Username)
	require.Equal(t, input.Fullname, acc.Fullname)
	require.Equal(t, input.Email, acc.Email)
	require.Equal(t, input.PhoneNumber, acc.PhoneNumber)
	require.Equal(t, input.RoleId, acc.RoleId)

	errD := aAsor.DeleteAccountByUsername(context.Background(), input.Username)
	require.NoError(t, errD)
}

func TestGetAccountByUsername(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	input := RandomAccount()
	id, err := aAsor.CreateAccount(context.Background(), input)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	acc, err := aAsor.GetAccountByUsername(context.Background(), input.Username)
	require.NoError(t, err)
	require.NotEmpty(t, acc)
	require.Equal(t, input.Username, acc.Username)
	require.Equal(t, input.Fullname, acc.Fullname)
	require.Equal(t, input.Email, acc.Email)
	require.Equal(t, input.PhoneNumber, acc.PhoneNumber)
	require.Equal(t, input.RoleId, acc.RoleId)

	errD := aAsor.DeleteAccountByUsername(context.Background(), input.Username)
	require.NoError(t, errD)
}

func TestDeleteAccountById(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	input := RandomAccount()
	id, err := aAsor.CreateAccount(context.Background(), input)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	errD := aAsor.DeleteAccount(context.Background(), id)
	require.NoError(t, errD)
}

func TestDeleteAccountByUsername(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	input := RandomAccount()
	id, err := aAsor.CreateAccount(context.Background(), input)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	errD := aAsor.DeleteAccountByUsername(context.Background(), input.Username)
	require.NoError(t, errD)
}

func TestUpdateAccount(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	in1 := RandomAccount()
	id1, err1 := aAsor.CreateAccount(context.Background(), in1)
	require.NoError(t, err1)
	require.NotEmpty(t, id1)
	require.NotZero(t, id1)

	in2 := in1
	in2.Fullname = RandomVnPersonName()
	in2.Email = RandomGmailAddress()
	in2.PhoneNumber = RandomVnPhoneNum()
	in2.RoleId = 1
	err2 := aAsor.UpdateAccount(context.Background(), in2)
	require.NoError(t, err2)

	updatedAcc, errU := aAsor.GetAccountByUsername(context.Background(), in1.Username)
	require.NoError(t, errU)
	require.NotEmpty(t, updatedAcc)
	require.Equal(t, in2.Username, updatedAcc.Username)
	require.Equal(t, in2.Fullname, updatedAcc.Fullname)
	require.Equal(t, in2.Email, updatedAcc.Email)
	require.Equal(t, in2.PhoneNumber, updatedAcc.PhoneNumber)
	require.Equal(t, in2.RoleId, updatedAcc.RoleId)

	errD := aAsor.DeleteAccountByUsername(context.Background(), in1.Username)
	require.NoError(t, errD)
}

func TestIsUsernameTaken(t *testing.T) {
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	ctx := context.Background()

	input := RandomAccount()
	id, err := aAsor.CreateAccount(ctx, input)
	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NotZero(t, id)

	isTakenTrue, err := aAsor.IsUsernameTaken(ctx, input.Username)
	require.NoError(t, err)
	require.Equal(t, true, isTakenTrue)

	isTakenFalse, err := aAsor.IsUsernameTaken(ctx, RandomAccount().Username)
	require.NoError(t, err)
	require.Equal(t, false, isTakenFalse)

	errD := aAsor.DeleteAccountByUsername(ctx, input.Username)
	require.NoError(t, errD)
}
