package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/pkg/logger"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.L().Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)
				response.Error(c, http.StatusInternalServerError, 500, "Internal Server Error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
