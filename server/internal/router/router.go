package router

import (
	"net/http"
	"time"

	"cmdb-server/internal/controllers"
	"cmdb-server/internal/middleware"
	"cmdb-server/internal/services"
	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	authService := services.NewAuthService()
	assetService := services.NewAssetService()
	credentialService := services.NewCredentialService()
	sshService := services.NewSSHService(credentialService)

	authController := controllers.NewAuthController(authService)
	assetController := controllers.NewAssetController(assetService)
	credentialController := controllers.NewCredentialController(credentialService)
	sshController := controllers.NewSSHController(sshService)

	r := gin.New()
	r.Use(middleware.RequestLogger(), middleware.RecoveryWithZap(), middleware.CORS())

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	r.POST("/api/auth/login", authController.Login)

	authGroup := r.Group("/api", middleware.AuthRequired(authService))
	{
		authGroup.POST("/auth/logout", authController.Logout)
		authGroup.GET("/auth/me", authController.Me)

		authGroup.GET("/assets", assetController.List)
		authGroup.POST("/assets", assetController.Create)
		authGroup.PUT("/assets/:id", assetController.Update)
		authGroup.DELETE("/assets/:id", assetController.Delete)

		authGroup.GET("/credentials", credentialController.List)
		authGroup.POST("/credentials", credentialController.Create)
		authGroup.PUT("/credentials/:id", credentialController.Update)
		authGroup.DELETE("/credentials/:id", credentialController.Delete)

		authGroup.POST("/ssh/test", sshController.Test)
	}

	return r
}
