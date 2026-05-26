package middleware

import (
	"errors"
	"net/http"
	"strings"

	"cmdb-server/internal/services"
	"cmdb-server/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthRequired(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := utils.BearerToken(c.GetHeader("Authorization"))
		if token == "" {
			token = strings.TrimSpace(c.Query("token"))
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录或 token 缺失"})
			return
		}

		payload, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, services.ErrInvalidSession) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态失效，请重新登录"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "鉴权失败"})
			return
		}

		c.Set("token", token)
		c.Set("userID", payload.UserID)
		c.Set("username", payload.Username)
		c.Next()
	}
}
