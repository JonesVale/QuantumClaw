package service

import (
	"github.com/quantumclaw/quantumclaw/setting"
)

// CheckSensitive checks if the given content contains sensitive words.
// Delegates to setting/sensitive.go CheckSensitive.
func CheckSensitive(content string) (bool, string) {
	return setting.CheckSensitive(content)
}
