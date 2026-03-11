package dao_database_test

import (
	"context"
	"testing"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"github.com/stretchr/testify/require"
)

func TestGetRoleByName(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	ar, err := a.GetRoleByName(context.Background(), "admin")

	require.NoError(t, err)
	require.NotEmpty(t, ar)
	require.Equal(t, ar.Id, uint8(1))
	require.Equal(t, ar.Name, "admin")
}

func TestGetRoleById(t *testing.T) {
	a := dao_database.NewAccountRoleAccessor(sqlDb, logger)
	ar, err := a.GetRoleById(context.Background(), 2)

	require.NoError(t, err)
	require.NotEmpty(t, ar)
	require.Equal(t, ar.Id, uint8(2))
	require.Equal(t, ar.Name, "member")
}
