package logic_http

import (
	"net/http"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	oapi "github.com/Fiagram/standalone/internal/generated/openapi"
	webhook_handler "github.com/Fiagram/standalone/internal/handler/chatbot"
	"github.com/Fiagram/standalone/internal/logger"
	logic_account "github.com/Fiagram/standalone/internal/logic/account"
	"github.com/Fiagram/standalone/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProfileLogic interface {
	GetProfileMe(c *gin.Context)
	UpdateProfileMe(c *gin.Context)
	UpdateProfilePassword(c *gin.Context)

	GetProfileWebhooks(c *gin.Context, params oapi.GetProfileWebhooksParams)
	CreateProfileWebhook(c *gin.Context)
	GetProfileWebhook(c *gin.Context, webhookId oapi.WebhookId)
	UpdateProfileWebhook(c *gin.Context, webhookId oapi.WebhookId)
	DeleteProfileWebhook(c *gin.Context, webhookId oapi.WebhookId)
}

var _ ProfileLogic = (oapi.ServerInterface)(nil)

type profileLogic struct {
	webhookAccessor     dao_database.ChatbotWebhookAccessor
	accountRoleAccessor dao_database.AccountRoleAccessor
	createdWebhookChan  webhook_handler.CreatedWebhookChan
	accountLogic        logic_account.Account
	logger              *zap.Logger
}

func NewProfileLogic(
	webhookAccessor dao_database.ChatbotWebhookAccessor,
	accountRoleAccessor dao_database.AccountRoleAccessor,
	createdWebhookChan webhook_handler.CreatedWebhookChan,
	accountLogic logic_account.Account,
	logger *zap.Logger,
) ProfileLogic {
	return &profileLogic{
		webhookAccessor:     webhookAccessor,
		accountRoleAccessor: accountRoleAccessor,
		createdWebhookChan:  createdWebhookChan,
		accountLogic:        accountLogic,
		logger:              logger,
	}
}

// getAccountIdFromContext extracts the accountId set by the auth middleware.
func getAccountIdFromContext(c *gin.Context, logger *zap.Logger) (uint64, bool) {
	accountId, exists := c.Get("accountId")
	if !exists || accountId.(uint64) == 0 {
		errMsg := "accountId not existed in context"
		logger.Error(errMsg)
		c.JSON(http.StatusUnauthorized, oapi.Unauthorized{
			Code:    "Unauthorized",
			Message: errMsg,
		})
		return 0, false
	}
	return accountId.(uint64), true
}

func (u *profileLogic) GetProfileMe(c *gin.Context) {
	logger := logger.LoggerWithContext(c, u.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	account, err := u.accountLogic.GetAccount(c,
		logic_account.GetAccountParams{
			AccountId: accountId,
		})
	if err != nil {
		errMsg := "failed to get account from account service"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	accountRole, err := u.accountRoleAccessor.GetRoleById(c, uint8(account.AccountInfo.Role))
	if err != nil {
		errMsg := "failed to get account role from database"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, oapi.ProfileMeResponse{
		Account: oapi.Account{
			Username:    account.AccountInfo.Username,
			Fullname:    account.AccountInfo.Fullname,
			Email:       account.AccountInfo.Email,
			PhoneNumber: &account.AccountInfo.PhoneNumber,
			Role:        oapi.Role(accountRole.Name),
		},
	})
}

func (u *profileLogic) UpdateProfileMe(c *gin.Context) {
	logger := logger.LoggerWithContext(c, u.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	var req oapi.UpdateProfileMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid request body",
		})
		return
	}

	// Fetch the current account to merge partial updates
	account, err := u.accountLogic.GetAccount(c,
		logic_account.GetAccountParams{
			AccountId: accountId,
		})
	if err != nil {
		errMsg := "failed to get account"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	updatedInfo := account.AccountInfo
	if req.Fullname != nil {
		updatedInfo.Fullname = *req.Fullname
	}
	if req.Email != nil {
		updatedInfo.Email = *req.Email
	}
	if req.PhoneNumber != nil {
		updatedInfo.PhoneNumber = *req.PhoneNumber
	}

	_, err = u.accountLogic.UpdateAccountInfo(c,
		logic_account.UpdateAccountInfoParams{
			AccountId:          accountId,
			UpdatedAccountInfo: updatedInfo,
		})
	if err != nil {
		errMsg := "failed to update account info"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	accountRole, err := u.accountRoleAccessor.GetRoleById(c, uint8(updatedInfo.Role))
	if err != nil {
		errMsg := "failed to get account role from database"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, oapi.ProfileMeResponse{
		Account: oapi.Account{
			Username:    updatedInfo.Username,
			Fullname:    updatedInfo.Fullname,
			Email:       updatedInfo.Email,
			PhoneNumber: &updatedInfo.PhoneNumber,
			Role:        oapi.Role(accountRole.Name),
		},
	})
}

func (u *profileLogic) UpdateProfilePassword(c *gin.Context) {
	logger := logger.LoggerWithContext(c, u.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	var req oapi.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid request body",
		})
		return
	}

	if req.OldPassword == nil || req.NewPassword == nil {
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "oldPassword and newPassword are required",
		})
		return
	}

	// Fetch account to get username for credential validation
	account, err := u.accountLogic.GetAccount(c,
		logic_account.GetAccountParams{
			AccountId: accountId,
		})
	if err != nil {
		errMsg := "failed to get account"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Verify old password
	validResult, err := u.accountLogic.CheckAccountValid(c,
		logic_account.CheckAccountValidParams{
			Username: account.AccountInfo.Username,
			Password: *req.OldPassword,
		})
	if err != nil {
		errMsg := "failed to validate old password"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}
	if validResult.AccountId == 0 {
		c.JSON(http.StatusForbidden, oapi.Forbidden{
			Code:    "Forbidden",
			Message: "old password is incorrect",
		})
		return
	}

	// Update to new password
	_, err = u.accountLogic.UpdateAccountPassword(c,
		logic_account.UpdateAccountPasswordParams{
			AccountId: accountId,
			Password:  *req.NewPassword,
		})
	if err != nil {
		errMsg := "failed to update password"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (u *profileLogic) GetProfileWebhooks(c *gin.Context, params oapi.GetProfileWebhooksParams) {
	logger := logger.LoggerWithContext(c, u.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	webhooks, err := u.webhookAccessor.GetWebhooksByAccountId(c, accountId, limit, offset)
	if err != nil {
		errMsg := "failed to get webhooks"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	out := make([]oapi.Webhook, 0, len(webhooks))
	for _, w := range webhooks {
		out = append(out, oapi.Webhook{
			Id:   utils.Ptr(int64(w.Id)),
			Name: w.Name,
			Url:  w.Url,
		})
	}

	c.JSON(http.StatusOK, out)
}

func (u *profileLogic) CreateProfileWebhook(c *gin.Context) {
	logger := logger.LoggerWithContext(c, u.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	var req oapi.Webhook
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid request body",
		})
		return
	}

	id, err := u.webhookAccessor.CreateWebhook(c, dao_database.ChatbotWebhook{
		OfAccountId: accountId,
		Name:        req.Name,
		Url:         req.Url,
	})
	if err != nil {
		errMsg := "failed to create webhook"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	u.createdWebhookChan <- webhook_handler.CreatedWebhookSignal{
		OfWebhookId: id,
	}

	c.JSON(http.StatusCreated, oapi.Webhook{
		Id:   utils.Ptr(int64(id)),
		Name: req.Name,
		Url:  req.Url,
	})
}

func (u *profileLogic) GetProfileWebhook(c *gin.Context, webhookId oapi.WebhookId) {
	logger := logger.LoggerWithContext(c, u.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	webhook, err := u.webhookAccessor.GetWebhook(c, uint64(webhookId))
	if err != nil {
		errMsg := "webhook not found"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusNotFound, oapi.NotFound{
			Code:    "NotFound",
			Message: errMsg,
		})
		return
	}

	if webhook.OfAccountId != accountId {
		c.JSON(http.StatusForbidden, oapi.Forbidden{
			Code:    "Forbidden",
			Message: "not allowed to access this webhook",
		})
		return
	}

	c.JSON(http.StatusOK, oapi.Webhook{
		Id:   utils.Ptr(int64(webhook.Id)),
		Name: webhook.Name,
		Url:  webhook.Url,
	})
}

func (u *profileLogic) UpdateProfileWebhook(c *gin.Context, webhookId oapi.WebhookId) {
	logger := logger.LoggerWithContext(c, u.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	// Verify ownership
	existing, err := u.webhookAccessor.GetWebhook(c, uint64(webhookId))
	if err != nil {
		c.JSON(http.StatusNotFound, oapi.NotFound{
			Code:    "NotFound",
			Message: "webhook not found",
		})
		return
	}
	if existing.OfAccountId != accountId {
		c.JSON(http.StatusForbidden, oapi.Forbidden{
			Code:    "Forbidden",
			Message: "not allowed to access this webhook",
		})
		return
	}

	var req oapi.Webhook
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid request body",
		})
		return
	}

	err = u.webhookAccessor.UpdateWebhook(c, dao_database.ChatbotWebhook{
		Id:   uint64(webhookId),
		Name: req.Name,
		Url:  req.Url,
	})
	if err != nil {
		errMsg := "failed to update webhook"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, oapi.Webhook{
		Id:   utils.Ptr(webhookId),
		Name: req.Name,
		Url:  req.Url,
	})
}

func (u *profileLogic) DeleteProfileWebhook(c *gin.Context, webhookId oapi.WebhookId) {
	logger := logger.LoggerWithContext(c, u.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	// Verify ownership
	existing, err := u.webhookAccessor.GetWebhook(c, uint64(webhookId))
	if err != nil {
		c.JSON(http.StatusNotFound, oapi.NotFound{
			Code:    "NotFound",
			Message: "webhook not found",
		})
		return
	}
	if existing.OfAccountId != accountId {
		c.JSON(http.StatusForbidden, oapi.Forbidden{
			Code:    "Forbidden",
			Message: "not allowed to access this webhook",
		})
		return
	}

	err = u.webhookAccessor.DeleteWebhook(c, uint64(webhookId))
	if err != nil {
		errMsg := "failed to delete webhook"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	c.Status(http.StatusNoContent)
}
