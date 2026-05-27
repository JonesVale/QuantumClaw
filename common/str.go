package common

import (
	"math/rand"
	"strings"
)

func StringContainsAny(s string, chars string) bool {
	return strings.ContainsAny(s, chars)
}

func RandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// MaskAPIKey 对 API Key 进行脱敏处理，显示前4位和后4位
func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		if len(key) > 0 {
			return key[:1] + "****"
		}
		return ""
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// MaskSensitiveInfo 对敏感信息进行脱敏，只显示前4位
func MaskSensitiveInfo(s string) string {
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-4)
}
