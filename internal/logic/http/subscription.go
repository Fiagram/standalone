package logic_http

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Fiagram/standalone/internal/configs"
	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	oapi "github.com/Fiagram/standalone/internal/generated/openapi"
	"github.com/Fiagram/standalone/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// planRank maps plan name → numeric rank for upgrade-only enforcement.
var planRank = map[string]int{
	"free": 0,
	"pro":  1,
	"max":  2,
}

// SubscriptionLogic is the interface for all subscription-related HTTP handlers.
type SubscriptionLogic interface {
	GetProfileSubscription(c *gin.Context)
	GetProfileSubscriptionPlans(c *gin.Context)
	PurchaseSubscription(c *gin.Context)
	GetSubscriptionOrder(c *gin.Context, orderId oapi.OrderId)
	SePayWebhook(c *gin.Context)
}

var _ SubscriptionLogic = (oapi.ServerInterface)(nil)

type subscriptionLogicImpl struct {
	subscriptionAccessor dao_database.AccountSubscriptionAccessor
	planAccessor         dao_database.SubscriptionPlanAccessor
	orderAccessor        dao_database.SubscriptionOrderAccessor
	accountRoleAccessor  dao_database.AccountRoleAccessor
	sePayConfig          configs.SePayConfig
	db                   *sql.DB
	logger               *zap.Logger
}

func NewSubscriptionLogic(
	subscriptionAccessor dao_database.AccountSubscriptionAccessor,
	planAccessor dao_database.SubscriptionPlanAccessor,
	orderAccessor dao_database.SubscriptionOrderAccessor,
	accountRoleAccessor dao_database.AccountRoleAccessor,
	sePayConfig configs.SePayConfig,
	db *sql.DB,
	logger *zap.Logger,
) SubscriptionLogic {
	return &subscriptionLogicImpl{
		subscriptionAccessor: subscriptionAccessor,
		planAccessor:         planAccessor,
		orderAccessor:        orderAccessor,
		accountRoleAccessor:  accountRoleAccessor,
		sePayConfig:          sePayConfig,
		db:                   db,
		logger:               logger,
	}
}

// ──────────────────────────────────────────────
// Core business method (used by StrategyLogic)
// ──────────────────────────────────────────────

// GetSubscription returns the account's current subscription, applying lazy downgrade
// if the paid plan has expired (resets to free/free/active).
func (s *subscriptionLogicImpl) GetSubscription(
	c *gin.Context,
	accountId uint64,
) (dao_database.AccountSubscription, error) {
	sub, err := s.subscriptionAccessor.GetSubscriptionByAccountId(c, accountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dao_database.AccountSubscription{
				OfAccountId:   accountId,
				Plan:          "free",
				BillingPeriod: "free",
				Status:        "active",
			}, nil
		}
		return dao_database.AccountSubscription{}, err
	}

	// Lazy downgrade: paid plan whose expiry has passed → reset to free.
	if sub.Status == "active" && sub.Plan != "free" &&
		sub.ExpiresAt != nil && time.Now().After(*sub.ExpiresAt) {
		downgraded := dao_database.AccountSubscription{
			OfAccountId:   accountId,
			Plan:          "free",
			BillingPeriod: "free",
			Status:        "active",
			ExpiresAt:     nil,
		}
		_ = s.subscriptionAccessor.UpdateSubscription(c, downgraded)
		return downgraded, nil
	}

	return sub, nil
}

// ──────────────────────────────────────────────
// HTTP handlers
// ──────────────────────────────────────────────

func (s *subscriptionLogicImpl) GetProfileSubscription(c *gin.Context) {
	log := logger.LoggerWithContext(c, s.logger)

	accountId, ok := getAccountIdFromContext(c, log)
	if !ok {
		return
	}

	sub, err := s.GetSubscription(c, accountId)
	if err != nil {
		log.Error("failed to get subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to get subscription",
		})
		return
	}

	c.JSON(http.StatusOK, oapi.SubscriptionStatus{
		Plan:          oapi.SubscriptionPlan(sub.Plan),
		BillingPeriod: oapi.SubscriptionBillingPeriod(sub.BillingPeriod),
		Status:        oapi.SubscriptionStatusStatus(sub.Status),
		ExpiresAt:     sub.ExpiresAt,
	})
}

func (s *subscriptionLogicImpl) GetProfileSubscriptionPlans(c *gin.Context) {
	log := logger.LoggerWithContext(c, s.logger)

	plans, err := s.planAccessor.ListPlans(c)
	if err != nil {
		log.Error("failed to list subscription plans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to list subscription plans",
		})
		return
	}

	resp := make([]oapi.SubscriptionPlanInfo, 0, len(plans))
	for _, p := range plans {
		resp = append(resp, oapi.SubscriptionPlanInfo{
			Plan:          oapi.SubscriptionPlan(p.Plan),
			BillingPeriod: oapi.SubscriptionPlanInfoBillingPeriod(p.BillingPeriod),
			Price:         p.Price,
			Currency:      p.Currency,
		})
	}
	c.JSON(http.StatusOK, resp)
}

func (s *subscriptionLogicImpl) PurchaseSubscription(c *gin.Context) {
	log := logger.LoggerWithContext(c, s.logger)

	accountId, ok := getAccountIdFromContext(c, log)
	if !ok {
		return
	}

	// 1. Role check — only members can purchase.
	role, err := s.accountRoleAccessor.GetRoleByAccountId(c, accountId)
	if err != nil {
		log.Error("failed to get account role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to verify account role",
		})
		return
	}
	if role.Name != "member" {
		c.JSON(http.StatusForbidden, oapi.Forbidden{
			Code:    "Forbidden",
			Message: "only member accounts can purchase a subscription",
		})
		return
	}

	// 2. Parse request body.
	var req oapi.PurchaseSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid request body: " + err.Error(),
		})
		return
	}
	if !req.Plan.Valid() || !req.BillingPeriod.Valid() {
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid plan or billingPeriod value",
		})
		return
	}

	// 3. Upgrade-only enforcement.
	currentSub, err := s.GetSubscription(c, accountId)
	if err != nil {
		log.Error("failed to get current subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to verify current subscription",
		})
		return
	}
	requestedPlan := string(req.Plan)
	currentPlan := currentSub.Plan
	if planRank[requestedPlan] <= planRank[currentPlan] {
		c.JSON(http.StatusUnprocessableEntity, oapi.UnprocessableEntity{
			Code:    "PLAN_DOWNGRADE_NOT_ALLOWED",
			Message: fmt.Sprintf("cannot change from %s to %s: only upgrades are allowed", currentPlan, requestedPlan),
		})
		return
	}

	// 4. Price lookup.
	billingPeriod := string(req.BillingPeriod)
	planInfo, err := s.planAccessor.GetPlan(c, requestedPlan, billingPeriod)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusBadRequest, oapi.BadRequest{
				Code:    "BadRequest",
				Message: fmt.Sprintf("plan %s/%s not found", requestedPlan, billingPeriod),
			})
			return
		}
		log.Error("failed to get plan price", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to retrieve plan pricing",
		})
		return
	}

	// 5. Generate unique reference code (max 50 chars for bank memo field).
	referenceCode := fmt.Sprintf("FIA%d%d", accountId, time.Now().UnixNano())
	if len(referenceCode) > 50 {
		referenceCode = referenceCode[:50]
	}

	// 6. Insert pending order.
	paymentExpiresAt := time.Now().Add(s.sePayConfig.PaymentExpiry)
	orderId, err := s.orderAccessor.CreateOrder(c, dao_database.SubscriptionOrder{
		OfAccountId:      accountId,
		Plan:             requestedPlan,
		BillingPeriod:    billingPeriod,
		Amount:           planInfo.Price,
		Currency:         planInfo.Currency,
		ReferenceCode:    referenceCode,
		PaymentExpiresAt: paymentExpiresAt,
	})
	if err != nil {
		log.Error("failed to create subscription order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to create subscription order",
		})
		return
	}

	// 7. Build SePay VietQR URL.
	qrCodeUrl := fmt.Sprintf(
		"https://qr.sepay.vn/img?acc=%s&bank=%s&amount=%.0f&des=%s&template=compact",
		url.QueryEscape(s.sePayConfig.AccountNumber),
		url.QueryEscape(s.sePayConfig.BankCode),
		planInfo.Price,
		url.QueryEscape(referenceCode),
	)

	c.JSON(http.StatusCreated, oapi.PurchaseSubscriptionResponse{
		OrderId:       &orderId,
		ReferenceCode: referenceCode,
		Amount:        planInfo.Price,
		Currency:      planInfo.Currency,
		BankCode:      s.sePayConfig.BankCode,
		AccountNumber: s.sePayConfig.AccountNumber,
		QrCodeUrl:     qrCodeUrl,
		ExpiresAt:     paymentExpiresAt,
	})
}

func (s *subscriptionLogicImpl) GetSubscriptionOrder(c *gin.Context, orderId oapi.OrderId) {
	log := logger.LoggerWithContext(c, s.logger)

	accountId, ok := getAccountIdFromContext(c, log)
	if !ok {
		return
	}

	order, err := s.orderAccessor.GetOrderById(c, orderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, oapi.NotFound{
				Code:    "NotFound",
				Message: "order not found",
			})
			return
		}
		log.Error("failed to get subscription order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to get subscription order",
		})
		return
	}

	// Ownership check.
	if order.OfAccountId != accountId {
		c.JSON(http.StatusForbidden, oapi.Forbidden{
			Code:    "Forbidden",
			Message: "access denied",
		})
		return
	}

	// Auto-expire if pending and payment window has passed.
	if order.Status == "pending" && time.Now().After(order.PaymentExpiresAt) {
		_ = s.orderAccessor.UpdateOrderStatus(c, order.Id, "expired", "")
		order.Status = "expired"
	}

	c.JSON(http.StatusOK, oapi.SubscriptionOrderStatus{
		OrderId:          &order.Id,
		Plan:             oapi.SubscriptionPlan(order.Plan),
		BillingPeriod:    oapi.SubscriptionOrderStatusBillingPeriod(order.BillingPeriod),
		Amount:           order.Amount,
		Currency:         order.Currency,
		Status:           oapi.SubscriptionOrderStatusStatus(order.Status),
		ReferenceCode:    order.ReferenceCode,
		CreatedAt:        &order.CreatedAt,
		PaymentExpiresAt: order.PaymentExpiresAt,
		SubExpiresAt:     order.SubExpiresAt,
	})
}

func (s *subscriptionLogicImpl) SePayWebhook(c *gin.Context) {
	log := logger.LoggerWithContext(c, s.logger)

	// 1. Verify SePay Apikey header.
	apiKey := c.GetHeader("Authorization")
	expectedKey := "Apikey " + s.sePayConfig.WebhookSecret
	if apiKey != expectedKey {
		c.JSON(http.StatusUnauthorized, oapi.Unauthorized{
			Code:    "Unauthorized",
			Message: "invalid webhook secret",
		})
		return
	}

	// 2. Parse payload.
	var payload oapi.SePayWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid payload: " + err.Error(),
		})
		return
	}
	if payload.Content == "" {
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "content field is required",
		})
		return
	}

	// 3. Find order by reference code embedded in the memo content.
	order, err := s.orderAccessor.GetOrderByReferenceCode(c, payload.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Reference code not found — could be an unrelated transfer; acknowledge silently.
			c.JSON(http.StatusOK, gin.H{"message": "reference code not matched; ignored"})
			return
		}
		log.Error("failed to find order by reference code", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to process webhook",
		})
		return
	}

	// 4. Idempotency — if already paid, return 200 with no side effects.
	if order.Status == "paid" {
		c.JSON(http.StatusOK, gin.H{"message": "already processed"})
		return
	}

	// 5. Compute subscription period end date.
	now := time.Now()
	var subExpiresAt time.Time
	switch order.BillingPeriod {
	case "yearly":
		subExpiresAt = now.AddDate(1, 0, 0)
	default: // monthly
		subExpiresAt = now.AddDate(0, 1, 0)
	}

	// 6. Transactionally update order + subscription.
	tx, err := s.db.BeginTx(c, nil)
	if err != nil {
		log.Error("failed to begin transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to process payment",
		})
		return
	}
	defer tx.Rollback()

	transactionId := strconv.Itoa(payload.Id)
	if err := s.orderAccessor.WithExecutor(tx).UpdateOrderStatus(c, order.Id, "paid", transactionId); err != nil {
		log.Error("failed to update order status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to confirm payment",
		})
		return
	}

	updatedSub := dao_database.AccountSubscription{
		OfAccountId:   order.OfAccountId,
		Plan:          order.Plan,
		BillingPeriod: order.BillingPeriod,
		Status:        "active",
		ExpiresAt:     &subExpiresAt,
	}
	if err := s.subscriptionAccessor.WithExecutor(tx).UpdateSubscription(c, updatedSub); err != nil {
		log.Error("failed to update subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to activate subscription",
		})
		return
	}

	// Update sub_start_at and sub_expires_at on the order row for record-keeping.
	if _, err := tx.ExecContext(c,
		`UPDATE subscription_orders SET sub_start_at = ?, sub_expires_at = ? WHERE id = ?`,
		now, subExpiresAt, order.Id,
	); err != nil {
		log.Error("failed to set sub dates on order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to finalize order",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Error("failed to commit transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to finalize payment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment confirmed"})
}
