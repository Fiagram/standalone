package dao_database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"github.com/stretchr/testify/require"
)

func TestGetPlan_HappyPath(t *testing.T) {
	asor := dao_database.NewSubscriptionPlanAccessor(sqlDb, logger)

	plan, err := asor.GetPlan(context.Background(), "pro", "monthly")
	require.NoError(t, err)
	require.Equal(t, "pro", plan.Plan)
	require.Equal(t, "monthly", plan.BillingPeriod)
	require.Greater(t, plan.Price, float64(0))
	require.NotEmpty(t, plan.Currency)
}

func TestGetPlan_NotFound(t *testing.T) {
	asor := dao_database.NewSubscriptionPlanAccessor(sqlDb, logger)

	_, err := asor.GetPlan(context.Background(), "nonexistent", "monthly")
	require.Error(t, err)
	require.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestListPlans(t *testing.T) {
	asor := dao_database.NewSubscriptionPlanAccessor(sqlDb, logger)

	plans, err := asor.ListPlans(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, plans)
	// Seeded rows: pro/monthly, pro/yearly, max/monthly, max/yearly
	require.GreaterOrEqual(t, len(plans), 4)
	for _, p := range plans {
		require.NotEmpty(t, p.Plan)
		require.NotEmpty(t, p.BillingPeriod)
		require.Greater(t, p.Price, float64(0))
		require.NotEmpty(t, p.Currency)
	}
}
