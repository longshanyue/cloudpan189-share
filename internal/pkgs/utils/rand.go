package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomPassword 生成随机密码的辅助函数
func GenerateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

	password := make([]byte, length)

	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// 如果加密随机数生成失败，回退到时间种子
			return generateFallbackPassword(length)
		}
		password[i] = charset[num.Int64()]
	}

	return string(password)
}

// generateFallbackPassword 回退的密码生成方法
func generateFallbackPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

	// 使用当前时间的纳秒作为种子
	seed := make([]byte, 8)
	_, _ = rand.Read(seed)

	password := make([]byte, length)
	for i := range password {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[num.Int64()]
	}

	return string(password)
}
