package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	oapi "github.com/Fiagram/standalone/internal/generated/openapi"
	logic "github.com/Fiagram/standalone/internal/logic/token"
	"github.com/gin-gonic/gin"
)

func VerifyAccessToken(tokenLogic logic.Token) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Try Authorization header
		authHeader := c.GetHeader("Authorization")
		if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			token = after
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, oapi.Unauthorized{
				Code:    "Unauthorized",
				Message: "access token is required",
			})
			return
		}

		claims, expiresAt, err := tokenLogic.GetPayloadFromAccessToken(c, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, oapi.Unauthorized{
				Code:    "Unauthorized",
				Message: "failed to verify access token",
			})
			return
		} else if time.Now().After(expiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, oapi.Unauthorized{
				Code:    "Unauthorized",
				Message: "the access token has expired",
			})
			return
		} else if claims.AccountId == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, oapi.Unauthorized{
				Code:    "Unauthorized",
				Message: "invalid access token",
			})
			return
		}

		c.Set("accountId", claims.AccountId)
		c.Next()
	}
}

func LogWithFormatter() gin.HandlerFunc {
	return gin.LoggerWithFormatter(
		func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		},
	)
}
