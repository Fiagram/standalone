package logic_http

import (
	"net/http"
	"strings"
	"time"

	"github.com/Fiagram/standalone/internal/configs"
	dao_cache "github.com/Fiagram/standalone/internal/dao/cache"
	oapi "github.com/Fiagram/standalone/internal/generated/openapi"
	"github.com/Fiagram/standalone/internal/logger"
	logic_account "github.com/Fiagram/standalone/internal/logic/account"
	logic_token "github.com/Fiagram/standalone/internal/logic/token"
	"github.com/Fiagram/standalone/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthLogic interface {
	SignIn(c *gin.Context)
	SignUp(c *gin.Context)
	RefreshToken(c *gin.Context)
	SignOut(c *gin.Context)
}

var _ AuthLogic = (oapi.ServerInterface)(nil)

type authLogic struct {
	authConfig          configs.Auth
	cookieConfig        configs.Cookie
	usernamesTakenCache dao_cache.UsernamesTaken
	refreshTokenCache   dao_cache.RefreshToken
	accountLogic        logic_account.Account
	tokenLogic          logic_token.Token
	logger              *zap.Logger
}

func NewAuthLogic(
	authConfig configs.Auth,
	cookieConfig configs.Cookie,
	usernamesTakenCache dao_cache.UsernamesTaken,
	refreshTokenCache dao_cache.RefreshToken,
	accountLogic logic_account.Account,
	tokenLogic logic_token.Token,
	logger *zap.Logger,
) AuthLogic {
	return &authLogic{
		authConfig:          authConfig,
		cookieConfig:        cookieConfig,
		usernamesTakenCache: usernamesTakenCache,
		refreshTokenCache:   refreshTokenCache,
		accountLogic:        accountLogic,
		tokenLogic:          tokenLogic,
		logger:              logger,
	}
}

func (o *authLogic) RefreshToken(c *gin.Context) {
	logger := logger.LoggerWithContext(c, o.logger)

	// Extract refresh token from header
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		errMsg := "refresh token is required"
		logger.Error(errMsg)
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: errMsg,
		})
		return
	}

	// Get account ID from refresh token cache
	accountId, err := o.refreshTokenCache.Get(c, refreshToken)
	if err != nil || accountId == 0 {
		errMsg := "invalid or expired refresh token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusUnauthorized, oapi.Unauthorized{
			Code:    "Unauthorized",
			Message: errMsg,
		})
		return
	}

	// Create a new access token
	accessToken, accessTokenExpiresAt, err := o.tokenLogic.GenerateAccessToken(c, logic_token.TokenPayload{
		AccountId: accountId,
	})
	if err != nil {
		errMsg := "failed to generate access token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Generate new refresh token (rotation)
	newRefreshToken, newRefreshTokenExpiresAt, err := o.tokenLogic.GenerateRefreshToken(c)
	if err != nil {
		errMsg := "failed to generate refresh token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Save the new refresh token to cache
	err = o.refreshTokenCache.Set(c, newRefreshToken, accountId, o.authConfig.Token.RefreshTokenTTL)
	if err != nil {
		errMsg := "failed to save refresh token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Revoke old refresh token
	if _, err := o.refreshTokenCache.Del(c, refreshToken); err != nil {
		logger.With(zap.Error(err)).Error("failed to revoke old refresh token")
		// Continue anyway, not a critical error
	}

	// Return the refresh token to cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     o.cookieConfig.Path,
		Domain:   o.cookieConfig.Domain,
		Expires:  newRefreshTokenExpiresAt,
		MaxAge:   int(time.Until(newRefreshTokenExpiresAt).Seconds()),
		Secure:   o.cookieConfig.Secure,
		HttpOnly: o.cookieConfig.HttpOnly,
		SameSite: o.cookieConfig.SameSite(),
	})

	// Return the new access token in response
	c.JSON(http.StatusOK, oapi.RefreshResponse{
		AccessToken: oapi.AccessTokenResponse{
			Token: accessToken,
			Exp:   accessTokenExpiresAt.Unix(),
		},
	})
}

func (o *authLogic) SignIn(c *gin.Context) {
	logger := logger.LoggerWithContext(c, o.logger)

	// Decode the incoming JSON object
	var req oapi.SigninRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := "failed to bind JSON object"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: errMsg,
		})
		return
	}

	// Verify data input is not empty
	username := req.Username
	password := *req.Password
	isRememberMe := *req.IsRememberMe
	logger.With(zap.String("username", username))
	if username == "" || password == "" {
		errMsg := "invalid username or password"
		logger.Error(errMsg)
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: errMsg,
		})
		return
	}

	// Checking account valid
	validResp, err := o.accountLogic.CheckAccountValid(c,
		logic_account.CheckAccountValidParams{
			Username: username,
			Password: password,
		})
	if err != nil {
		errMsg := "failed to check account valid"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	} else if validResp.AccountId == 0 {
		c.JSON(http.StatusUnauthorized, oapi.Unauthorized{
			Code:    "Unauthorized",
			Message: "invalid username or password",
		})
		return
	}

	// Get account info
	account, err := o.accountLogic.GetAccount(c, logic_account.GetAccountParams{
		AccountId: validResp.AccountId,
	})
	if err != nil {
		errMsg := "failed to get account info"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Create a new access token
	accessToken, accessTokenExpiresAt, err := o.tokenLogic.GenerateAccessToken(c, logic_token.TokenPayload{
		AccountId: validResp.AccountId,
	})
	if err != nil {
		errMsg := "failed to gen access token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Create refresh token
	refreshToken, refreshTokenExpiresAt, err := o.tokenLogic.GenerateRefreshToken(c)
	if err != nil {
		errMsg := "failed to gen refresh token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Save the refresh token to the redis
	err = o.refreshTokenCache.Set(c,
		refreshToken, validResp.AccountId,
		utils.If(isRememberMe,
			o.authConfig.Token.RefreshTokenLongTTL,
			o.authConfig.Token.RefreshTokenTTL),
	)
	if err != nil {
		errMsg := "failed to save refresh token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Return the refresh token to cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     o.cookieConfig.Path,
		Domain:   o.cookieConfig.Domain,
		Expires:  refreshTokenExpiresAt,
		MaxAge:   int(time.Until(refreshTokenExpiresAt).Seconds()),
		Secure:   o.cookieConfig.Secure,
		HttpOnly: o.cookieConfig.HttpOnly,
		SameSite: o.cookieConfig.SameSite(),
	})

	// Return the access token to the response
	phoneNumberRaw := account.AccountInfo.PhoneNumber
	countryCode := strings.Split(phoneNumberRaw, " ")[0]
	phoneNumber := strings.Split(phoneNumberRaw, " ")[1]
	c.JSON(http.StatusOK, oapi.SigninResponse{
		AccessToken: oapi.AccessTokenResponse{
			Token: accessToken,
			Exp:   accessTokenExpiresAt.Unix(),
		},
		Account: oapi.Account{
			Username: username,
			Fullname: account.AccountInfo.Fullname,
			Email:    account.AccountInfo.Email,
			PhoneNumber: &oapi.PhoneNumber{
				CountryCode: utils.Ptr(countryCode),
				Number:      utils.Ptr(phoneNumber),
			},
			Role: "member",
		},
	})
}

func (o *authLogic) SignOut(c *gin.Context) {
	logger := logger.LoggerWithContext(c, o.logger)

	// Extract refresh token from header
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		errMsg := "refresh token is required"
		logger.Error(errMsg)
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: errMsg,
		})
		return
	}

	// Revoke the refresh token from cache
	isDone, err := o.refreshTokenCache.Del(c, refreshToken)
	if err != nil || isDone == false {
		errMsg := "failed to revoke refresh token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Return the refresh token to cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     o.cookieConfig.Path,
		Domain:   o.cookieConfig.Domain,
		MaxAge:   -1,
		Secure:   o.cookieConfig.Secure,
		HttpOnly: o.cookieConfig.HttpOnly,
		SameSite: o.cookieConfig.SameSite(),
	})

	c.Status(http.StatusNoContent)
}

func (o *authLogic) SignUp(c *gin.Context) {
	logger := logger.LoggerWithContext(c, o.logger)

	// Decode the incoming JSON object
	var req oapi.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.With(zap.Error(err)).Error("failed to decode incoming JSON")
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "failed to bind JSON object",
		})
		return
	}

	// Check whether the username is taken
	username := req.Account.Username
	isTaken, err := o.usernamesTakenCache.Has(c, username)
	if isTaken {
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "username is taken",
		})
		return
	}
	takenResp, err := o.accountLogic.IsUsernameTaken(c,
		logic_account.IsUsernameTakenParams{
			Username: username,
		})
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to call grpc method")
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to call grpc method",
		})
		return
	}
	if takenResp.IsTaken {
		if err := o.usernamesTakenCache.Add(c, username); err != nil {
			logger.With(zap.Error(err)).
				With(zap.Any("account", req.Account)).
				Error("failed to add username to cache")
		}
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "username is taken",
		})
		return
	}

	// Process the incoming request
	accResp, err := o.accountLogic.CreateAccount(c, logic_account.CreateAccountParams{
		AccountInfo: logic_account.AccountInfo{
			Username:    username,
			Fullname:    req.Account.Fullname,
			Email:       req.Account.Email,
			PhoneNumber: *req.Account.PhoneNumber.CountryCode + " " + *req.Account.PhoneNumber.Number,
			Role:        2,
		},
		Password: *req.Password,
	})
	if err != nil || accResp.AccountId == 0 {
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "failed to process grpc method",
		})
		return
	}

	// Create a new access token
	accessToken, accessTokenExpiresAt, err := o.tokenLogic.GenerateAccessToken(c, logic_token.TokenPayload{
		AccountId: accResp.AccountId,
	})
	if err != nil {
		errMsg := "failed to gen access token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Create refresh token
	refreshToken, refreshTokenExpiresAt, err := o.tokenLogic.GenerateRefreshToken(c)
	if err != nil {
		errMsg := "failed to gen refresh token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Save the refresh token to the redis
	err = o.refreshTokenCache.Set(c,
		refreshToken, accResp.AccountId,
		o.authConfig.Token.RefreshTokenTTL)
	if err != nil {
		errMsg := "failed to save refresh token"
		logger.With(zap.Error(err)).Error(errMsg)
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	// Return the refresh token to cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     o.cookieConfig.Path,
		Domain:   o.cookieConfig.Domain,
		Expires:  refreshTokenExpiresAt,
		MaxAge:   int(time.Until(refreshTokenExpiresAt).Seconds()),
		Secure:   o.cookieConfig.Secure,
		HttpOnly: o.cookieConfig.HttpOnly,
		SameSite: o.cookieConfig.SameSite(),
	})

	// Return the access token to the response
	c.JSON(http.StatusOK, oapi.SignupResponse{
		AccessToken: oapi.AccessTokenResponse{
			Token: accessToken,
			Exp:   accessTokenExpiresAt.Unix(),
		},
		Account: oapi.Account{
			Username: username,
			Fullname: req.Account.Fullname,
			Email:    req.Account.Email,
			PhoneNumber: &oapi.PhoneNumber{
				CountryCode: req.Account.PhoneNumber.CountryCode,
				Number:      req.Account.PhoneNumber.Number,
			},
			Role: "member",
		},
	})
}
