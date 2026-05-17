package setting

import (
	"strings"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

var sensitiveWords []string

func InitSensitiveFilter() {
	logger.SysLog("sensitive filter initialized (basic mode)")
}

func AddSensitiveWord(word string) {
	sensitiveWords = append(sensitiveWords, strings.ToLower(word))
}

func CheckSensitive(content string) (bool, string) {
	lower := strings.ToLower(content)
	for _, word := range sensitiveWords {
		if strings.Contains(lower, word) {
			return true, word
		}
	}
	return false, ""
}
