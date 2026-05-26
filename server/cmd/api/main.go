package main

import (
	"log"
	"time"

	"cmdb-server/internal/bootstrap"
	"cmdb-server/internal/config"
	"cmdb-server/internal/global"
	"cmdb-server/internal/logger"
	"cmdb-server/internal/router"
	"cmdb-server/internal/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}
	global.CONFIG = cfg

	zapLogger, err := logger.Init(cfg.Log)
	if err != nil {
		log.Fatalf("init logger failed: %v", err)
	}
	defer func() {
		_ = zapLogger.Sync()
	}()
	global.LOGGER = zapLogger
	global.SUGAR = zapLogger.Sugar()

	ttlHours := cfg.Auth.SessionTTLHours
	if ttlHours <= 0 {
		ttlHours = 12
	}
	global.SESSION_TTL = time.Duration(ttlHours) * time.Hour

	encryptor, err := utils.NewEncryptor(cfg.Auth.Secret)
	if err != nil {
		global.SUGAR.Fatalf("init encryptor failed: %v", err)
	}
	global.ENCRYPTOR = encryptor

	if err := bootstrap.InitMySQL(); err != nil {
		global.SUGAR.Fatalf("connect mysql failed: %v", err)
	}
	if err := bootstrap.InitRedis(); err != nil {
		global.SUGAR.Fatalf("connect redis failed: %v", err)
	}
	if err := bootstrap.AutoMigrate(); err != nil {
		global.SUGAR.Fatalf("auto migrate failed: %v", err)
	}
	if err := bootstrap.SeedDefaultData(); err != nil {
		global.SUGAR.Fatalf("seed default data failed: %v", err)
	}

	gin.SetMode(cfg.App.Mode)
	r := router.New()

	global.SUGAR.Infow("cmdb api started", "port", cfg.App.Port, "name", cfg.App.Name)
	global.SUGAR.Infow("default login", "username", "admin", "password", "Admin@123456")

	if err := r.Run(":" + cfg.App.Port); err != nil {
		global.SUGAR.Fatalf("server startup failed: %v", err)
	}
}
