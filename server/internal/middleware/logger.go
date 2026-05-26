package middleware

import (
	"net/http"
	"time"

	"cmdb-server/internal/global"
	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if global.SUGAR != nil {
			global.SUGAR.Infow("http_request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"latency", time.Since(start).String(),
				"client_ip", c.ClientIP(),
			)
		}
	}
}

func RecoveryWithZap() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		if global.SUGAR != nil {
			global.SUGAR.Errorw("panic recovered",
				"error", recovered,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
