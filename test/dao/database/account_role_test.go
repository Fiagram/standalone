package dao_database_test

import (
	"context"
	"testing"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"github.com/stretchr/testify/require"
)

func TestGetRoleByName_Admin(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	ar, err := a.GetRoleByName(context.Background(), "admin")

	require.NoError(t, err)
	require.NotEmpty(t, ar)
	require.Equal(t, uint8(1), ar.Id)
	require.Equal(t, "admin", ar.Name)
}

func TestGetRoleByName_Member(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	ar, err := a.GetRoleByName(context.Background(), "member")

	require.NoError(t, err)
	require.NotEmpty(t, ar)
	require.Equal(t, uint8(2), ar.Id)
	require.Equal(t, "member", ar.Name)
}

func TestGetRoleByName_None(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	ar, err := a.GetRoleByName(context.Background(), "none")

	require.NoError(t, err)
	require.Equal(t, uint8(0), ar.Id)
	require.Equal(t, "none", ar.Name)
}

func TestGetRoleByName_EmptyString(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	_, err := a.GetRoleByName(context.Background(), "")

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
}

func TestGetRoleByName_NotFound(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	_, err := a.GetRoleByName(context.Background(), "nonexistent_role")

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrAccRoleNotFound)
}

func TestGetRoleById_Admin(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	ar, err := a.GetRoleById(context.Background(), 1)

	require.NoError(t, err)
	require.NotEmpty(t, ar)
	require.Equal(t, uint8(1), ar.Id)
	require.Equal(t, "admin", ar.Name)
}

func TestGetRoleById_Member(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	ar, err := a.GetRoleById(context.Background(), 2)

	require.NoError(t, err)
	require.NotEmpty(t, ar)
	require.Equal(t, uint8(2), ar.Id)
	require.Equal(t, "member", ar.Name)
}

func TestGetRoleById_Zero(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	_, err := a.GetRoleById(context.Background(), 0)

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
}

func TestGetRoleById_NotFound(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	_, err := a.GetRoleById(context.Background(), 255)

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrAccRoleNotFound)
}

func TestAccountRoleWithExecutor(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	a2 := a.WithExecutor(sqlDb)

	require.NotNil(t, a2)

	// Verify the new accessor still works correctly
	ar, err := a2.GetRoleByName(context.Background(), "admin")
	require.NoError(t, err)
	require.Equal(t, uint8(1), ar.Id)
	require.Equal(t, "admin", ar.Name)
}

func TestGetAllSeededRoles(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)

	expectedRoles := []struct {
		id   uint8
		name string
	}{
		{0, "none"},
		{1, "admin"},
		{2, "member"},
	}

	for _, expected := range expectedRoles {
		ar, err := a.GetRoleByName(context.Background(), expected.name)
		require.NoError(t, err)
		require.Equal(t, expected.id, ar.Id)
		require.Equal(t, expected.name, ar.Name)

		arById, err := a.GetRoleById(context.Background(), expected.id)
		if expected.id == 0 {
			// GetRoleById returns ErrLackOfInfor for id=0
			require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
			continue
		}
		require.NoError(t, err)
		require.Equal(t, expected.id, arById.Id)
		require.Equal(t, expected.name, arById.Name)
	}
}
