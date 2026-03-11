package dao_cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRedisSetAndGet(t *testing.T) {

	ctx := context.Background()

	key := "key9"
	expected := "value9"
	err := client.Set(ctx, key, expected, 1*time.Second)
	require.NoError(t, err)

	actual, err := client.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestRedisDel(t *testing.T) {
	ctx := context.Background()

	key := "key9f"
	expected := "sdsd"
	err := client.Set(ctx, key, expected, 0)
	require.NoError(t, err)

	err = client.Del(ctx, key)
	require.NoError(t, err)

	_, err = client.Get(ctx, key)
	require.Error(t, err)
}

func TestAddToSet(t *testing.T) {
	ctx := context.Background()

	key := "key38"
	err := client.AddToSet(ctx, key, "v1", "v2", "v3")
	require.NoError(t, err)

	isTrue, err := client.IsDataInSet(ctx, key, "v3")
	require.NoError(t, err)
	require.Equal(t, true, isTrue)

	err = client.Del(ctx, key)
	require.NoError(t, err)
}
