package logic_http

import (
	"net/http"

	oapi "github.com/Fiagram/standalone/internal/generated/openapi"
	"github.com/Fiagram/standalone/internal/logger"
	logic_account "github.com/Fiagram/standalone/internal/logic/account"
	"github.com/Fiagram/standalone/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UsersLogic interface {
	GetMe(c *gin.Context)
}

var _ UsersLogic = (oapi.ServerInterface)(nil)

type usersLogic struct {
	accountLogic logic_account.Account
	logger       *zap.Logger
}

func NewUsersLogic(
	accountLogic logic_account.Account,
	logger *zap.Logger,
) UsersLogic {
	return &usersLogic{
		accountLogic: accountLogic,
		logger:       logger,
	}
}

func (u *usersLogic) GetMe(c *gin.Context) {
	logger := logger.LoggerWithContext(c, u.logger)

	accountId, exists := c.Get("accountId")
	if !exists || accountId.(uint64) == 0 {
		errMsg := "accountId not existed in context"
		logger.Error(errMsg)
		c.JSON(http.StatusUnauthorized, oapi.Unauthorized{
			Code:    "Unauthorized",
			Message: errMsg,
		})
		return
	}

	account, err := u.accountLogic.GetAccount(c,
		logic_account.GetAccountParams{
			AccountId: accountId.(uint64),
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

	c.JSON(http.StatusOK, oapi.UsersMeResponse{
		Account: oapi.Account{
			Username: account.AccountInfo.Username,
			Fullname: account.AccountInfo.Fullname,
			Email:    account.AccountInfo.Email,
			PhoneNumber: &oapi.PhoneNumber{
				CountryCode: utils.Ptr("none"), // TODO: fill proper country code
				Number:      utils.Ptr(account.AccountInfo.PhoneNumber),
			},
			Role: "member", // TODO: fill proper role with converters int <-> string
		},
	})

}
