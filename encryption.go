package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	u "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对密码进行哈希
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password is required")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, hash string) error {
	if password == "" || hash == "" {
		return fmt.Errorf("password and hash are both required")
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// UUID 生成 UUID
func UUID() string {
	uuid, err := u.NewV7()
	if err != nil {
		return u.NewString()
	}
	return uuid.String()
}

// Sign 生成签名
// 使用 HMAC-SHA256
func Sign(appKey string, data ...interface{}) string {
	json := NewUnknownType(data)
	jsonBytes, err := json.JSONSort()
	if err != nil {
		return ""
	}
	stringToSign := string(jsonBytes) + appKey
	h := hmac.New(sha256.New, []byte(appKey))
	if _, err := h.Write([]byte(stringToSign)); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}
