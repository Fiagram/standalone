package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Fiagram/standalone/internal/configs"
	oapi "github.com/Fiagram/standalone/internal/generated/openapi"
	"github.com/Fiagram/standalone/internal/logger"
	auth_logic "github.com/Fiagram/standalone/internal/logic/http"
	token_logic "github.com/Fiagram/standalone/internal/logic/token"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime"
	"go.uber.org/zap"
)

type HttpServer interface {
	Start(ctx context.Context) error
}

type httpServer struct {
	httpConfig configs.Http

	authLogic         auth_logic.AuthLogic
	usersLogic        auth_logic.ProfileLogic
	strategyLogic     auth_logic.StrategyLogic
	subscriptionLogic auth_logic.SubscriptionLogic
	tokenLogic        token_logic.Token

	logger *zap.Logger
}

func NewHttpServer(
	httpConfig configs.Http,
	authLogic auth_logic.AuthLogic,
	usersLogic auth_logic.ProfileLogic,
	strategyLogic auth_logic.StrategyLogic,
	subscriptionLogic auth_logic.SubscriptionLogic,
	tokenLogic token_logic.Token,
	logger *zap.Logger,
) HttpServer {
	return &httpServer{
		httpConfig:        httpConfig,
		authLogic:         authLogic,
		usersLogic:        usersLogic,
		strategyLogic:     strategyLogic,
		subscriptionLogic: subscriptionLogic,
		tokenLogic:        tokenLogic,
		logger:            logger,
	}
}

func (s *httpServer) Start(ctx context.Context) error {
	logger := logger.LoggerWithContext(ctx, s.logger)

	r := gin.Default()
	if s.httpConfig.CORS.IsEnable {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     s.httpConfig.CORS.AllowOrigins,
			AllowMethods:     s.httpConfig.CORS.AllowMethods,
			AllowHeaders:     s.httpConfig.CORS.AllowHeaders,
			ExposeHeaders:    s.httpConfig.CORS.ExposeHeaders,
			AllowCredentials: s.httpConfig.CORS.AllowCredentials,
			MaxAge:           s.httpConfig.CORS.MaxAge,
		}))
	}

	public := r.Group("/api/v1")
	public.POST("/auth/signup", s.authLogic.SignUp)
	public.POST("/auth/signin", s.authLogic.SignIn)
	public.POST("/auth/token/signout", s.authLogic.SignOut)
	public.POST("/auth/token/refresh", s.authLogic.RefreshToken)

	authorized := r.Group("/api/v1",
		verifyAccessToken(s.tokenLogic),
	)
	authorized.GET("/profile/me", s.usersLogic.GetProfileMe)
	authorized.PUT("/profile/me", s.usersLogic.UpdateProfileMe)
	authorized.PUT("/profile/me/password", s.usersLogic.UpdateProfilePassword)

	authorized.GET("/profile/webhooks", func(c *gin.Context) {
		var params oapi.GetProfileWebhooksParams
		_ = runtime.BindQueryParameterWithOptions("form", true, false, "limit", c.Request.URL.Query(), &params.Limit, runtime.BindQueryParameterOptions{Type: "integer", Format: ""})
		_ = runtime.BindQueryParameterWithOptions("form", true, false, "offset", c.Request.URL.Query(), &params.Offset, runtime.BindQueryParameterOptions{Type: "integer", Format: ""})
		s.usersLogic.GetProfileWebhooks(c, params)
	})
	authorized.POST("/profile/webhooks", s.usersLogic.CreateProfileWebhook)
	authorized.GET("/profile/webhooks/:webhookId", s.wrapWebhookId(s.usersLogic.GetProfileWebhook))
	authorized.PUT("/profile/webhooks/:webhookId", s.wrapWebhookId(s.usersLogic.UpdateProfileWebhook))
	authorized.DELETE("/profile/webhooks/:webhookId", s.wrapWebhookId(s.usersLogic.DeleteProfileWebhook))

	authorized.GET("/profile/subscription", s.subscriptionLogic.GetProfileSubscription)
	authorized.GET("/profile/subscription/plans", s.subscriptionLogic.GetProfileSubscriptionPlans)
	authorized.POST("/profile/subscription/purchase", s.subscriptionLogic.PurchaseSubscription)
	authorized.GET("/profile/subscription/orders/:orderId", s.wrapOrderId(s.subscriptionLogic.GetSubscriptionOrder))

	public.POST("/webhooks/payment/sepay", s.subscriptionLogic.SePayWebhook)

	authorized.GET("/strategy/alerts", func(c *gin.Context) {
		var params oapi.GetStrategyAlertsParams
		_ = runtime.BindQueryParameterWithOptions("form", true, false, "limit", c.Request.URL.Query(), &params.Limit, runtime.BindQueryParameterOptions{Type: "integer", Format: ""})
		_ = runtime.BindQueryParameterWithOptions("form", true, false, "offset", c.Request.URL.Query(), &params.Offset, runtime.BindQueryParameterOptions{Type: "integer", Format: ""})
		s.strategyLogic.GetStrategyAlerts(c, params)
	})
	authorized.POST("/strategy/alerts", s.strategyLogic.CreateStrategyAlert)
	authorized.GET("/strategy/alerts/:alertId", s.wrapAlertId(s.strategyLogic.GetStrategyAlert))
	authorized.PUT("/strategy/alerts/:alertId", s.wrapAlertId(s.strategyLogic.UpdateStrategyAlert))
	authorized.DELETE("/strategy/alerts/:alertId", s.wrapAlertId(s.strategyLogic.DeleteStrategyAlert))

	address := s.httpConfig.Address
	port := s.httpConfig.Port
	logger.With(zap.String("address", address)).
		With(zap.String("port", port)).
		Info("starting http server")

	return r.Run(address + ":" + port)
}

// wrapWebhookId parses the webhookId path parameter and delegates to the typed handler.
func (s *httpServer) wrapWebhookId(handler func(*gin.Context, oapi.WebhookId)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var webhookId oapi.WebhookId
		err := runtime.BindStyledParameterWithOptions("simple", "webhookId", c.Param("webhookId"), &webhookId, runtime.BindStyledParameterOptions{Explode: false, Required: true, Type: "integer", Format: "uint64"})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BadRequest", "message": fmt.Sprintf("invalid webhookId: %s", err)})
			return
		}
		handler(c, webhookId)
	}
}

// wrapOrderId parses the orderId path parameter and delegates to the typed handler.
func (s *httpServer) wrapOrderId(handler func(*gin.Context, oapi.OrderId)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderId oapi.OrderId
		err := runtime.BindStyledParameterWithOptions("simple", "orderId", c.Param("orderId"), &orderId, runtime.BindStyledParameterOptions{Explode: false, Required: true, Type: "integer", Format: "uint64"})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BadRequest", "message": fmt.Sprintf("invalid orderId: %s", err)})
			return
		}
		handler(c, orderId)
	}
}

// wrapAlertId extracts the alertId path parameter and delegates to the typed handler.
func (s *httpServer) wrapAlertId(handler func(*gin.Context, oapi.AlertId)) gin.HandlerFunc {
	return func(c *gin.Context) {
		alertId := c.Param("alertId")
		if alertId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": "BadRequest", "message": "missing alertId"})
			return
		}
		handler(c, alertId)
	}
}
