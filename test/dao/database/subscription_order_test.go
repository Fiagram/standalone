package dao_database_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"github.com/stretchr/testify/require"
)

// helper: create a test order and return its ID + cleanup func
func createTestOrder(t *testing.T, accountId uint64) (uint64, func()) {
	t.Helper()
	oAsor := dao_database.NewSubscriptionOrderAccessor(sqlDb, logger)
	refCode := fmt.Sprintf("TESTREF%d%d", accountId, time.Now().UnixNano())
	if len(refCode) > 50 {
		refCode = refCode[:50]
	}
	id, err := oAsor.CreateOrder(context.Background(), dao_database.SubscriptionOrder{
		OfAccountId:      accountId,
		Plan:             "pro",
		BillingPeriod:    "monthly",
		Amount:           99000,
		Currency:         "VND",
		ReferenceCode:    refCode,
		PaymentExpiresAt: time.Now().Add(15 * time.Minute),
	})
	require.NoError(t, err)
	require.NotZero(t, id)
	return id, func() {
		_, _ = sqlDb.ExecContext(context.Background(), "DELETE FROM subscription_orders WHERE id = ?", id)
	}
}

func TestCreateOrder(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	orderId, cleanupOrder := createTestOrder(t, accountId)
	defer cleanupOrder()

	require.NotZero(t, orderId)
}

func TestGetOrderById(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	orderId, cleanupOrder := createTestOrder(t, accountId)
	defer cleanupOrder()

	oAsor := dao_database.NewSubscriptionOrderAccessor(sqlDb, logger)
	order, err := oAsor.GetOrderById(context.Background(), orderId)
	require.NoError(t, err)
	require.Equal(t, orderId, order.Id)
	require.Equal(t, accountId, order.OfAccountId)
	require.Equal(t, "pro", order.Plan)
	require.Equal(t, "monthly", order.BillingPeriod)
	require.Equal(t, float64(99000), order.Amount)
	require.Equal(t, "VND", order.Currency)
	require.Equal(t, "pending", order.Status)
	require.NotEmpty(t, order.ReferenceCode)
}

func TestGetOrderById_NotFound(t *testing.T) {
	oAsor := dao_database.NewSubscriptionOrderAccessor(sqlDb, logger)
	_, err := oAsor.GetOrderById(context.Background(), 999999999)
	require.Error(t, err)
	require.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestGetOrderByReferenceCode(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	refCode := fmt.Sprintf("TESTREFCODE%d", time.Now().UnixNano())
	if len(refCode) > 50 {
		refCode = refCode[:50]
	}
	oAsor := dao_database.NewSubscriptionOrderAccessor(sqlDb, logger)
	orderId, err := oAsor.CreateOrder(context.Background(), dao_database.SubscriptionOrder{
		OfAccountId:      accountId,
		Plan:             "max",
		BillingPeriod:    "yearly",
		Amount:           1990000,
		Currency:         "VND",
		ReferenceCode:    refCode,
		PaymentExpiresAt: time.Now().Add(15 * time.Minute),
	})
	require.NoError(t, err)
	defer func() {
		_, _ = sqlDb.ExecContext(context.Background(), "DELETE FROM subscription_orders WHERE id = ?", orderId)
	}()

	order, err := oAsor.GetOrderByReferenceCode(context.Background(), refCode)
	require.NoError(t, err)
	require.Equal(t, orderId, order.Id)
	require.Equal(t, refCode, order.ReferenceCode)
	require.Equal(t, "max", order.Plan)
	require.Equal(t, "yearly", order.BillingPeriod)
}

func TestGetOrderByReferenceCode_NotFound(t *testing.T) {
	oAsor := dao_database.NewSubscriptionOrderAccessor(sqlDb, logger)
	_, err := oAsor.GetOrderByReferenceCode(context.Background(), "no-such-reference-code-xyz")
	require.Error(t, err)
	require.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestUpdateOrderStatus_Paid(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	orderId, cleanupOrder := createTestOrder(t, accountId)
	defer cleanupOrder()

	oAsor := dao_database.NewSubscriptionOrderAccessor(sqlDb, logger)
	err := oAsor.UpdateOrderStatus(context.Background(), orderId, "paid", "TX123")
	require.NoError(t, err)

	order, err := oAsor.GetOrderById(context.Background(), orderId)
	require.NoError(t, err)
	require.Equal(t, "paid", order.Status)
	require.NotNil(t, order.SePayTransactionId)
	require.Equal(t, "TX123", *order.SePayTransactionId)
	require.NotNil(t, order.SubStartAt)
}

func TestUpdateOrderStatus_Expired(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	orderId, cleanupOrder := createTestOrder(t, accountId)
	defer cleanupOrder()

	oAsor := dao_database.NewSubscriptionOrderAccessor(sqlDb, logger)
	err := oAsor.UpdateOrderStatus(context.Background(), orderId, "expired", "")
	require.NoError(t, err)

	order, err := oAsor.GetOrderById(context.Background(), orderId)
	require.NoError(t, err)
	require.Equal(t, "expired", order.Status)
	require.Nil(t, order.SePayTransactionId)
}
