package handler

import (
	"context"

	"github.com/Fiagram/standalone/internal/configs"
	"github.com/Fiagram/standalone/internal/handler/middlewares"
	"github.com/Fiagram/standalone/internal/logger"
	auth_logic "github.com/Fiagram/standalone/internal/logic/http"
	token_logic "github.com/Fiagram/standalone/internal/logic/token"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HttpServer interface {
	Start(ctx context.Context) error
}

type httpServer struct {
	httpConfig configs.Http

	authLogic  auth_logic.AuthLogic
	usersLogic auth_logic.UsersLogic
	tokenLogic token_logic.Token

	logger *zap.Logger
}

func NewHttpServer(
	httpConfig configs.Http,
	authLogic auth_logic.AuthLogic,
	usersLogic auth_logic.UsersLogic,
	tokenLogic token_logic.Token,
	logger *zap.Logger,
) HttpServer {
	return &httpServer{
		httpConfig: httpConfig,
		authLogic:  authLogic,
		usersLogic: usersLogic,
		tokenLogic: tokenLogic,
		logger:     logger,
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
		middlewares.VerifyAccessToken(s.tokenLogic),
	)
	authorized.GET("/users/me", s.usersLogic.GetMe)

	address := s.httpConfig.Address
	port := s.httpConfig.Port
	logger.With(zap.String("address", address)).
		With(zap.String("port", port)).
		Info("starting http server")

	return r.Run(address + ":" + port)
}
