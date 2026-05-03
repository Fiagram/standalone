package logic_http_test

import (
"context"
"database/sql"
"encoding/json"
"fmt"
"net/http"
"testing"
"time"

"github.com/Fiagram/standalone/internal/configs"
dao_database "github.com/Fiagram/standalone/internal/dao/database"
oapi "github.com/Fiagram/standalone/internal/generated/openapi"
logic_http "github.com/Fiagram/standalone/internal/logic/http"
"github.com/stretchr/testify/require"
"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Mock: dao_database.SubscriptionPlanAccessor
// ---------------------------------------------------------------------------

type mockSubscriptionPlanAccessor struct {
getPlanFn   func(ctx context.Context, plan, billingPeriod string) (dao_database.SubscriptionPlan, error)
listPlansFn func(ctx context.Context) ([]dao_database.SubscriptionPlan, error)
}

func (m *mockSubscriptionPlanAccessor) GetPlan(ctx context.Context, plan, billingPeriod string) (dao_database.SubscriptionPlan, error) {
if m.getPlanFn != nil {
return m.getPlanFn(ctx, plan, billingPeriod)
}
return dao_database.SubscriptionPlan{Plan: plan, BillingPeriod: billingPeriod, Price: 99000, Currency: "VND"}, nil
}
func (m *mockSubscriptionPlanAccessor) ListPlans(ctx context.Context) ([]dao_database.SubscriptionPlan, error) {
if m.listPlansFn != nil {
return m.listPlansFn(ctx)
}
return []dao_database.SubscriptionPlan{
{Plan: "pro", BillingPeriod: "monthly", Price: 99000, Currency: "VND"},
{Plan: "pro", BillingPeriod: "yearly", Price: 990000, Currency: "VND"},
{Plan: "max", BillingPeriod: "monthly", Price: 199000, Currency: "VND"},
{Plan: "max", BillingPeriod: "yearly", Price: 1990000, Currency: "VND"},
}, nil
}
func (m *mockSubscriptionPlanAccessor) WithExecutor(_ dao_database.Executor) dao_database.SubscriptionPlanAccessor {
return m
}

// ---------------------------------------------------------------------------
// Mock: dao_database.SubscriptionOrderAccessor
// ---------------------------------------------------------------------------

type mockSubscriptionOrderAccessor struct {
createOrderFn             func(ctx context.Context, order dao_database.SubscriptionOrder) (uint64, error)
getOrderByIdFn            func(ctx context.Context, id uint64) (dao_database.SubscriptionOrder, error)
getOrderByReferenceCodeFn func(ctx context.Context, referenceCode string) (dao_database.SubscriptionOrder, error)
listOrdersByAccountIdFn   func(ctx context.Context, accountId uint64, limit, offset int) ([]dao_database.SubscriptionOrder, error)
updateOrderStatusFn       func(ctx context.Context, id uint64, status, transactionId string) error
}

func (m *mockSubscriptionOrderAccessor) CreateOrder(ctx context.Context, order dao_database.SubscriptionOrder) (uint64, error) {
if m.createOrderFn != nil {
return m.createOrderFn(ctx, order)
}
return 1, nil
}
func (m *mockSubscriptionOrderAccessor) GetOrderById(ctx context.Context, id uint64) (dao_database.SubscriptionOrder, error) {
if m.getOrderByIdFn != nil {
return m.getOrderByIdFn(ctx, id)
}
return dao_database.SubscriptionOrder{}, sql.ErrNoRows
}
func (m *mockSubscriptionOrderAccessor) GetOrderByReferenceCode(ctx context.Context, referenceCode string) (dao_database.SubscriptionOrder, error) {
if m.getOrderByReferenceCodeFn != nil {
return m.getOrderByReferenceCodeFn(ctx, referenceCode)
}
return dao_database.SubscriptionOrder{}, sql.ErrNoRows
}
func (m *mockSubscriptionOrderAccessor) ListOrdersByAccountId(ctx context.Context, accountId uint64, limit, offset int) ([]dao_database.SubscriptionOrder, error) {
if m.listOrdersByAccountIdFn != nil {
return m.listOrdersByAccountIdFn(ctx, accountId, limit, offset)
}
return nil, nil
}
func (m *mockSubscriptionOrderAccessor) UpdateOrderStatus(ctx context.Context, id uint64, status, transactionId string) error {
if m.updateOrderStatusFn != nil {
return m.updateOrderStatusFn(ctx, id, status, transactionId)
}
return nil
}
func (m *mockSubscriptionOrderAccessor) WithExecutor(_ dao_database.Executor) dao_database.SubscriptionOrderAccessor {
return m
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func defaultSePayConfig() configs.SePayConfig {
return configs.SePayConfig{
APIKey:        "testapikey",
AccountNumber: "100123456789",
BankCode:      "MB",
WebhookSecret: "webhooksecret",
PaymentExpiry: 15 * time.Minute,
}
}

func newTestSubscriptionLogic(
subAccessor dao_database.AccountSubscriptionAccessor,
planAccessor dao_database.SubscriptionPlanAccessor,
orderAccessor dao_database.SubscriptionOrderAccessor,
roleAccessor dao_database.AccountRoleAccessor,
) logic_http.SubscriptionLogic {
var db *sql.DB // nil is safe for tests that do not reach s.db.BeginTx
return logic_http.NewSubscriptionLogic(
subAccessor, planAccessor, orderAccessor, roleAccessor,
defaultSePayConfig(), db, zap.NewNop(),
)
}

func freeSubscriptionMock() *mockAccountSubscriptionAccessor {
return &mockAccountSubscriptionAccessor{
getSubscriptionByAccountIdFn: func(_ context.Context, accountId uint64) (dao_database.AccountSubscription, error) {
return dao_database.AccountSubscription{
OfAccountId:   accountId,
Plan:          "free",
BillingPeriod: "free",
Status:        "active",
}, nil
},
}
}

func memberRoleMock() *mockAccountRoleAccessor {
return &mockAccountRoleAccessor{
getRoleByAccountIdFn: func(_ context.Context, _ uint64) (dao_database.AccountRole, error) {
return dao_database.AccountRole{Id: 2, Name: "member"}, nil
},
}
}

// ---------------------------------------------------------------------------
// Tests: GetProfileSubscription
// ---------------------------------------------------------------------------

func TestGetProfileSubscription_Active(t *testing.T) {
expiresAt := time.Now().Add(30 * 24 * time.Hour)
sub := dao_database.AccountSubscription{
Plan:          "pro",
BillingPeriod: "monthly",
Status:        "active",
ExpiresAt:     &expiresAt,
}
subAccessor := &mockAccountSubscriptionAccessor{
getSubscriptionByAccountIdFn: func(_ context.Context, _ uint64) (dao_database.AccountSubscription, error) {
return sub, nil
},
}
sl := newTestSubscriptionLogic(subAccessor, &mockSubscriptionPlanAccessor{}, &mockSubscriptionOrderAccessor{}, &mockAccountRoleAccessor{})

c, w := newGinContext(http.MethodGet, "/profile/subscription", nil)
setAccountId(c, 42)
sl.GetProfileSubscription(c)

require.Equal(t, http.StatusOK, w.Code)
var resp oapi.SubscriptionStatus
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
require.Equal(t, oapi.SubscriptionPlan("pro"), resp.Plan)
require.Equal(t, oapi.SubscriptionBillingPeriod("monthly"), resp.BillingPeriod)
require.Equal(t, oapi.SubscriptionStatusStatus("active"), resp.Status)
require.NotNil(t, resp.ExpiresAt)
}

func TestGetProfileSubscription_LazyDowngrade(t *testing.T) {
expired := time.Now().Add(-24 * time.Hour)
sub := dao_database.AccountSubscription{
Plan:          "pro",
BillingPeriod: "monthly",
Status:        "active",
ExpiresAt:     &expired,
}
subAccessor := &mockAccountSubscriptionAccessor{
getSubscriptionByAccountIdFn: func(_ context.Context, _ uint64) (dao_database.AccountSubscription, error) {
return sub, nil
},
}
sl := newTestSubscriptionLogic(subAccessor, &mockSubscriptionPlanAccessor{}, &mockSubscriptionOrderAccessor{}, &mockAccountRoleAccessor{})

c, w := newGinContext(http.MethodGet, "/profile/subscription", nil)
setAccountId(c, 42)
sl.GetProfileSubscription(c)

require.Equal(t, http.StatusOK, w.Code)
var resp oapi.SubscriptionStatus
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
require.Equal(t, oapi.SubscriptionPlan("free"), resp.Plan)
require.Nil(t, resp.ExpiresAt)
}

// ---------------------------------------------------------------------------
// Tests: GetProfileSubscriptionPlans
// ---------------------------------------------------------------------------

func TestGetProfileSubscriptionPlans(t *testing.T) {
sl := newTestSubscriptionLogic(
&mockAccountSubscriptionAccessor{},
&mockSubscriptionPlanAccessor{},
&mockSubscriptionOrderAccessor{},
&mockAccountRoleAccessor{},
)

c, w := newGinContext(http.MethodGet, "/profile/subscription/plans", nil)
setAccountId(c, 42)
sl.GetProfileSubscriptionPlans(c)

require.Equal(t, http.StatusOK, w.Code)
var resp []oapi.SubscriptionPlanInfo
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
require.Len(t, resp, 4)
}

// ---------------------------------------------------------------------------
// Tests: PurchaseSubscription
// ---------------------------------------------------------------------------

func TestPurchaseSubscription_FreeToProMonthly(t *testing.T) {
var capturedOrder dao_database.SubscriptionOrder

sl := newTestSubscriptionLogic(
freeSubscriptionMock(),
&mockSubscriptionPlanAccessor{},
&mockSubscriptionOrderAccessor{
createOrderFn: func(_ context.Context, order dao_database.SubscriptionOrder) (uint64, error) {
capturedOrder = order
return 7, nil
},
},
memberRoleMock(),
)

c, w := newGinContext(http.MethodPost, "/profile/subscription/purchase", oapi.PurchaseSubscriptionRequest{
Plan:          "pro",
BillingPeriod: "monthly",
})
setAccountId(c, 42)
sl.PurchaseSubscription(c)

require.Equal(t, http.StatusCreated, w.Code)
var resp oapi.PurchaseSubscriptionResponse
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
require.NotNil(t, resp.OrderId)
require.Equal(t, uint64(7), *resp.OrderId)
require.NotEmpty(t, resp.QrCodeUrl)
require.NotEmpty(t, resp.ReferenceCode)
require.Equal(t, "VND", resp.Currency)

require.Equal(t, "pro", capturedOrder.Plan)
require.Equal(t, "monthly", capturedOrder.BillingPeriod)
require.Equal(t, uint64(42), capturedOrder.OfAccountId)
}

func TestPurchaseSubscription_FreeToMaxYearly(t *testing.T) {
sl := newTestSubscriptionLogic(
freeSubscriptionMock(),
&mockSubscriptionPlanAccessor{},
&mockSubscriptionOrderAccessor{},
memberRoleMock(),
)

c, w := newGinContext(http.MethodPost, "/profile/subscription/purchase", oapi.PurchaseSubscriptionRequest{
Plan:          "max",
BillingPeriod: "yearly",
})
setAccountId(c, 42)
sl.PurchaseSubscription(c)

require.Equal(t, http.StatusCreated, w.Code)
}

func TestPurchaseSubscription_ProToMax(t *testing.T) {
subMock := &mockAccountSubscriptionAccessor{
getSubscriptionByAccountIdFn: func(_ context.Context, _ uint64) (dao_database.AccountSubscription, error) {
return dao_database.AccountSubscription{Plan: "pro", Status: "active"}, nil
},
}
sl := newTestSubscriptionLogic(subMock, &mockSubscriptionPlanAccessor{}, &mockSubscriptionOrderAccessor{}, memberRoleMock())

c, w := newGinContext(http.MethodPost, "/profile/subscription/purchase", oapi.PurchaseSubscriptionRequest{
Plan:          "max",
BillingPeriod: "monthly",
})
setAccountId(c, 42)
sl.PurchaseSubscription(c)

require.Equal(t, http.StatusCreated, w.Code)
}

func TestPurchaseSubscription_ProToFree_InvalidPlan(t *testing.T) {
subMock := &mockAccountSubscriptionAccessor{
getSubscriptionByAccountIdFn: func(_ context.Context, _ uint64) (dao_database.AccountSubscription, error) {
return dao_database.AccountSubscription{Plan: "pro", Status: "active"}, nil
},
}
sl := newTestSubscriptionLogic(subMock, &mockSubscriptionPlanAccessor{}, &mockSubscriptionOrderAccessor{}, memberRoleMock())

c, w := newGinContext(http.MethodPost, "/profile/subscription/purchase", oapi.PurchaseSubscriptionRequest{
Plan:          "free",
BillingPeriod: "monthly",
})
setAccountId(c, 42)
sl.PurchaseSubscription(c)

// "free" is not a valid PurchaseSubscriptionRequestPlan value, so Valid() returns false -> 400
require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseSubscription_MaxToPro_Rejected(t *testing.T) {
subMock := &mockAccountSubscriptionAccessor{
getSubscriptionByAccountIdFn: func(_ context.Context, _ uint64) (dao_database.AccountSubscription, error) {
return dao_database.AccountSubscription{Plan: "max", Status: "active"}, nil
},
}
sl := newTestSubscriptionLogic(subMock, &mockSubscriptionPlanAccessor{}, &mockSubscriptionOrderAccessor{}, memberRoleMock())

c, w := newGinContext(http.MethodPost, "/profile/subscription/purchase", oapi.PurchaseSubscriptionRequest{
Plan:          "pro",
BillingPeriod: "monthly",
})
setAccountId(c, 42)
sl.PurchaseSubscription(c)

require.Equal(t, http.StatusUnprocessableEntity, w.Code)
var resp oapi.UnprocessableEntity
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
require.Equal(t, "PLAN_DOWNGRADE_NOT_ALLOWED", resp.Code)
}

func TestPurchaseSubscription_SamePlan_Rejected(t *testing.T) {
subMock := &mockAccountSubscriptionAccessor{
getSubscriptionByAccountIdFn: func(_ context.Context, _ uint64) (dao_database.AccountSubscription, error) {
return dao_database.AccountSubscription{Plan: "pro", Status: "active"}, nil
},
}
sl := newTestSubscriptionLogic(subMock, &mockSubscriptionPlanAccessor{}, &mockSubscriptionOrderAccessor{}, memberRoleMock())

c, w := newGinContext(http.MethodPost, "/profile/subscription/purchase", oapi.PurchaseSubscriptionRequest{
Plan:          "pro",
BillingPeriod: "monthly",
})
setAccountId(c, 42)
sl.PurchaseSubscription(c)

require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestPurchaseSubscription_NonMember_Forbidden(t *testing.T) {
roleMock := &mockAccountRoleAccessor{
getRoleByAccountIdFn: func(_ context.Context, _ uint64) (dao_database.AccountRole, error) {
return dao_database.AccountRole{Id: 1, Name: "admin"}, nil
},
}
sl := newTestSubscriptionLogic(freeSubscriptionMock(), &mockSubscriptionPlanAccessor{}, &mockSubscriptionOrderAccessor{}, roleMock)

c, w := newGinContext(http.MethodPost, "/profile/subscription/purchase", oapi.PurchaseSubscriptionRequest{
Plan:          "pro",
BillingPeriod: "monthly",
})
setAccountId(c, 42)
sl.PurchaseSubscription(c)

require.Equal(t, http.StatusForbidden, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GetSubscriptionOrder
// ---------------------------------------------------------------------------

func TestGetSubscriptionOrder_Success(t *testing.T) {
expiresAt := time.Now().Add(30 * 24 * time.Hour)
orderMock := &mockSubscriptionOrderAccessor{
getOrderByIdFn: func(_ context.Context, id uint64) (dao_database.SubscriptionOrder, error) {
require.Equal(t, uint64(7), id)
return dao_database.SubscriptionOrder{
Id:               7,
OfAccountId:      42,
Plan:             "pro",
BillingPeriod:    "monthly",
Amount:           99000,
Currency:         "VND",
Status:           "pending",
ReferenceCode:    "FIA4299999",
PaymentExpiresAt: expiresAt,
}, nil
},
}
sl := newTestSubscriptionLogic(freeSubscriptionMock(), &mockSubscriptionPlanAccessor{}, orderMock, &mockAccountRoleAccessor{})

c, w := newGinContext(http.MethodGet, "/profile/subscription/orders/7", nil)
setAccountId(c, 42)
sl.GetSubscriptionOrder(c, 7)

require.Equal(t, http.StatusOK, w.Code)
var resp oapi.SubscriptionOrderStatus
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
require.NotNil(t, resp.OrderId)
require.Equal(t, uint64(7), *resp.OrderId)
require.Equal(t, oapi.SubscriptionPlan("pro"), resp.Plan)
require.Equal(t, oapi.SubscriptionOrderStatusStatus("pending"), resp.Status)
}

func TestGetSubscriptionOrder_NotFound(t *testing.T) {
sl := newTestSubscriptionLogic(
freeSubscriptionMock(),
&mockSubscriptionPlanAccessor{},
&mockSubscriptionOrderAccessor{},
&mockAccountRoleAccessor{},
)

c, w := newGinContext(http.MethodGet, "/profile/subscription/orders/999", nil)
setAccountId(c, 42)
sl.GetSubscriptionOrder(c, 999)

require.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSubscriptionOrder_ForbiddenOtherAccount(t *testing.T) {
orderMock := &mockSubscriptionOrderAccessor{
getOrderByIdFn: func(_ context.Context, id uint64) (dao_database.SubscriptionOrder, error) {
return dao_database.SubscriptionOrder{
Id:               id,
OfAccountId:      99,
Plan:             "pro",
BillingPeriod:    "monthly",
Status:           "pending",
PaymentExpiresAt: time.Now().Add(15 * time.Minute),
}, nil
},
}
sl := newTestSubscriptionLogic(freeSubscriptionMock(), &mockSubscriptionPlanAccessor{}, orderMock, &mockAccountRoleAccessor{})

c, w := newGinContext(http.MethodGet, "/profile/subscription/orders/7", nil)
setAccountId(c, 42)
sl.GetSubscriptionOrder(c, 7)

require.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetSubscriptionOrder_AutoExpire(t *testing.T) {
past := time.Now().Add(-1 * time.Hour)
var updatedStatus string
orderMock := &mockSubscriptionOrderAccessor{
getOrderByIdFn: func(_ context.Context, id uint64) (dao_database.SubscriptionOrder, error) {
return dao_database.SubscriptionOrder{
Id:               id,
OfAccountId:      42,
Plan:             "pro",
BillingPeriod:    "monthly",
Status:           "pending",
PaymentExpiresAt: past,
ReferenceCode:    "FIA42123",
}, nil
},
updateOrderStatusFn: func(_ context.Context, _ uint64, status, _ string) error {
updatedStatus = status
return nil
},
}
sl := newTestSubscriptionLogic(freeSubscriptionMock(), &mockSubscriptionPlanAccessor{}, orderMock, &mockAccountRoleAccessor{})

c, w := newGinContext(http.MethodGet, "/profile/subscription/orders/7", nil)
setAccountId(c, 42)
sl.GetSubscriptionOrder(c, 7)

require.Equal(t, http.StatusOK, w.Code)
var resp oapi.SubscriptionOrderStatus
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
require.Equal(t, oapi.SubscriptionOrderStatusStatus("expired"), resp.Status)
require.Equal(t, "expired", updatedStatus)
}

// ---------------------------------------------------------------------------
// Tests: SePayWebhook — auth and idempotency (no DB transaction needed)
// ---------------------------------------------------------------------------

func TestSePayWebhook_InvalidAuth(t *testing.T) {
sl := newTestSubscriptionLogic(
freeSubscriptionMock(),
&mockSubscriptionPlanAccessor{},
&mockSubscriptionOrderAccessor{},
&mockAccountRoleAccessor{},
)

c, w := newGinContext(http.MethodPost, "/webhooks/payment/sepay", oapi.SePayWebhookPayload{
Content: "FIA42123",
Id:      1,
})
c.Request.Header.Set("Authorization", "Apikey wrongsecret")
sl.SePayWebhook(c)

require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSePayWebhook_ReferenceCodeNotFound_Ignored(t *testing.T) {
sl := newTestSubscriptionLogic(
freeSubscriptionMock(),
&mockSubscriptionPlanAccessor{},
&mockSubscriptionOrderAccessor{},
&mockAccountRoleAccessor{},
)

c, w := newGinContext(http.MethodPost, "/webhooks/payment/sepay", oapi.SePayWebhookPayload{
Content: "UNKNOWNREF",
Id:      2,
})
c.Request.Header.Set("Authorization", fmt.Sprintf("Apikey %s", defaultSePayConfig().WebhookSecret))
sl.SePayWebhook(c)

require.Equal(t, http.StatusOK, w.Code)
}

func TestSePayWebhook_AlreadyPaid_Idempotent(t *testing.T) {
orderMock := &mockSubscriptionOrderAccessor{
getOrderByReferenceCodeFn: func(_ context.Context, _ string) (dao_database.SubscriptionOrder, error) {
return dao_database.SubscriptionOrder{
Id:            1,
OfAccountId:   42,
Plan:          "pro",
BillingPeriod: "monthly",
Status:        "paid",
ReferenceCode: "FIA42PAID",
}, nil
},
}
sl := newTestSubscriptionLogic(freeSubscriptionMock(), &mockSubscriptionPlanAccessor{}, orderMock, &mockAccountRoleAccessor{})

c, w := newGinContext(http.MethodPost, "/webhooks/payment/sepay", oapi.SePayWebhookPayload{
Content: "FIA42PAID",
Id:      3,
})
c.Request.Header.Set("Authorization", fmt.Sprintf("Apikey %s", defaultSePayConfig().WebhookSecret))
sl.SePayWebhook(c)

require.Equal(t, http.StatusOK, w.Code)
var resp map[string]string
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
require.Equal(t, "already processed", resp["message"])
}
