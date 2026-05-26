package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
)

func BearerToken(v string) string {
	parts := strings.Fields(strings.TrimSpace(v))
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func SessionRedisKey(token string) string {
	return "cmdb:session:" + token
}

func RandomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func ParseID(raw string) (uint, error) {
	v, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || v == 0 {
		return 0, errors.New("invalid id")
	}
	return uint(v), nil
}

func SplitTags(raw string) []string {
	items := strings.Split(raw, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		tag := strings.TrimSpace(item)
		if tag != "" {
			out = append(out, tag)
		}
	}
	return out
}

func JoinTags(items []string) string {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		tag := strings.TrimSpace(item)
		if tag != "" {
			filtered = append(filtered, tag)
		}
	}
	return strings.Join(filtered, ",")
}
