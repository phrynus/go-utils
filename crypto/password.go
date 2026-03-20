package crypto

import (
	"fmt"

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
