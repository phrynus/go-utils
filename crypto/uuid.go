package crypto

import (
	u "github.com/google/uuid"
)

// UUID 生成 UUID v7
func UUID() string {
	uuid, err := u.NewV7()
	if err != nil {
		return u.NewString()
	}
	return uuid.String()
}
