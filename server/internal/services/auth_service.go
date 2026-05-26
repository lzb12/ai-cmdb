package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"cmdb-server/internal/global"
	"cmdb-server/internal/models"
	"cmdb-server/internal/utils"
	redis "github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSession     = errors.New("invalid session")
)

type AuthService struct{}

type SessionPayload struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
}

type LoginResult struct {
	Token     string
	User      SessionPayload
	ExpiresAt time.Time
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(c context.Context, username, password string) (*LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	var user models.User
	if err := global.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := utils.RandomToken(32)
	if err != nil {
		return nil, err
	}
	payload := SessionPayload{
		UserID:   user.ID,
		Username: user.Username,
	}
	raw, _ := json.Marshal(payload)
	if err := global.REDIS.Set(c, utils.SessionRedisKey(token), raw, global.SESSION_TTL).Err(); err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:     token,
		User:      payload,
		ExpiresAt: time.Now().Add(global.SESSION_TTL),
	}, nil
}

func (s *AuthService) ValidateToken(c context.Context, token string) (*SessionPayload, error) {
	raw, err := global.REDIS.Get(c, utils.SessionRedisKey(token)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrInvalidSession
	}
	if err != nil {
		return nil, err
	}

	var payload SessionPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, ErrInvalidSession
	}
	return &payload, nil
}

func (s *AuthService) Logout(c context.Context, token string) error {
	if token == "" {
		return nil
	}
	return global.REDIS.Del(c, utils.SessionRedisKey(token)).Err()
}
