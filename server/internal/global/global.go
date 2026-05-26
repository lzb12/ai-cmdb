package global

import (
	"time"

	"cmdb-server/internal/config"
	"cmdb-server/internal/utils"
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	CONFIG      *config.Config
	DB          *gorm.DB
	REDIS       *redis.Client
	LOGGER      *zap.Logger
	SUGAR       *zap.SugaredLogger
	ENCRYPTOR   *utils.Encryptor
	SESSION_TTL time.Duration
)
