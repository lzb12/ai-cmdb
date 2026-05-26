package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cmdb-server/internal/global"
	"cmdb-server/internal/models"
	redis "github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	defaultAdminUser = "admin"
	defaultAdminPass = "Admin@123456"
)

func InitMySQL() error {
	dsn, err := buildMySQLDSN()
	if err != nil {
		return err
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(global.CONFIG.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(global.CONFIG.MySQL.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(global.CONFIG.MySQL.ConnMaxLifetimeMinutes) * time.Minute)

	global.DB = db
	return nil
}

func buildMySQLDSN() (string, error) {
	cfg := global.CONFIG.MySQL
	username := strings.TrimSpace(cfg.Username)
	dbName := strings.TrimSpace(cfg.DB)
	if username == "" || dbName == "" {
		return "", errors.New("mysql.username 和 mysql.db 不能为空")
	}

	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port <= 0 {
		port = 3306
	}
	charset := strings.TrimSpace(cfg.Charset)
	if charset == "" {
		charset = "utf8mb4"
	}
	loc := strings.TrimSpace(cfg.Loc)
	if loc == "" {
		loc = "Local"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		username,
		cfg.Password,
		host,
		port,
		dbName,
		charset,
		cfg.ParseTime,
		url.QueryEscape(loc),
	)
	return dsn, nil
}

func InitRedis() error {
	cfg := global.CONFIG.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return err
	}
	global.REDIS = client
	return nil
}

func AutoMigrate() error {
	return global.DB.AutoMigrate(&models.User{}, &models.Credential{}, &models.HostAsset{})
}

func SeedDefaultData() error {
	if err := seedAdmin(); err != nil {
		return err
	}
	if err := seedAsset(); err != nil {
		return err
	}
	return nil
}

func seedAdmin() error {
	var count int64
	if err := global.DB.Model(&models.User{}).Where("username = ?", defaultAdminUser).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return global.DB.Create(&models.User{
		Username:     defaultAdminUser,
		PasswordHash: string(hash),
	}).Error
}

func seedAsset() error {
	var count int64
	if err := global.DB.Model(&models.HostAsset{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	asset := models.HostAsset{
		Hostname:    "prd-web-01",
		Address:     "10.20.1.11",
		Port:        22,
		OS:          "Ubuntu 22.04",
		Environment: "prod",
		Owner:       "平台运维",
		TagsText:    "nginx,entry",
	}
	if err := global.DB.Create(&asset).Error; err != nil {
		return fmt.Errorf("seed asset failed: %w", err)
	}
	return nil
}
