package utils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Load 加载配置，从环境变量读取
func EnvLoad() error {
	// 尝试加载.env文件（可选）
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("Failed to load .env file: %w", err)
	}
	return nil
}

// Get 获取环境变量，如果不存在则返回默认值
func EnvGet(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetInt 获取环境变量，如果不存在则返回默认值
func EnvGetInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetBool 获取环境变量，如果不存在则返回默认值
func EnvGetBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetFloat64 获取环境变量，如果不存在则返回默认值
func EnvGetFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// Set 设置环境变量
func EnvSet(key string, value string) {
	os.Setenv(key, value)
}

// Unset 删除环境变量
func EnvUnset(key string) {
	os.Unsetenv(key)
}
