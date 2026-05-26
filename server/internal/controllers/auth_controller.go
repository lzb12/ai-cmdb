package controllers

import (
	"errors"
	"net/http"

	"cmdb-server/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	result, err := ctl.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     result.Token,
		"expiresAt": result.ExpiresAt.Format(timeLayoutRFC3339),
		"user":      result.User,
	})
}

func (ctl *AuthController) Logout(c *gin.Context) {
	token := c.GetString("token")
	if err := ctl.authService.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "退出失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已退出登录"})
}

func (ctl *AuthController) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"userId":   c.GetUint("userID"),
			"username": c.GetString("username"),
		},
	})
}

const timeLayoutRFC3339 = "2006-01-02T15:04:05Z07:00"
