package setting

import "github.com/quantumclaw/quantumclaw/common/logger"

var (
	GlobalRateLimitEnabled     = true
	GlobalRateLimitPerMinute   = 60
	TokenRateLimitPerMinute    = 30
)

func InitRateLimitSettings() {
	logger.SysLog("rate limit settings initialized")
}
