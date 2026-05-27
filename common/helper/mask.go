package helper

import "strings"

// MaskAPIKey 对 API Key 进行统一脱敏处理，显示前4位和后4位
// 所有返回 API Key 的端点都应使用此函数
func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		if len(key) > 0 {
			return key[:1] + "****"
		}
		return ""
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

// MaskSensitive 对任意敏感字符串脱敏，仅显示前4位
func MaskSensitive(s string) string {
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-4)
}
