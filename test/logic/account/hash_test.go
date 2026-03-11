package logic_account_test

import (
	"context"
	"testing"

	logic_account "github.com/Fiagram/standalone/internal/logic/account"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	ctx := context.Background()
	hashLogic := logic_account.NewHash(config.Auth.Hash)

	hs1, err1 := hashLogic.Hash(ctx, RandomString(72))
	require.NoError(t, err1)
	require.Equal(t, 60, len(hs1))

	hs2, err2 := hashLogic.Hash(ctx, RandomString(73))
	require.Error(t, err2)
	require.Equal(t, "", hs2)
}

func TestIsHashEqual(t *testing.T) {
	ctx := context.Background()
	hashLogic := logic_account.NewHash(config.Auth.Hash)

	input := RandomString(72)
	hs1, err1 := hashLogic.Hash(ctx, input)
	require.NoError(t, err1)
	require.Equal(t, 60, len(hs1))

	isEqualTrue, errT := hashLogic.IsHashEqual(ctx, input, hs1)
	require.NoError(t, errT)
	require.Equal(t, true, isEqualTrue)

	isEqualFalse, errF := hashLogic.IsHashEqual(ctx, RandomString(70), hs1)
	require.NoError(t, errF)
	require.Equal(t, false, isEqualFalse)
}
