package setting

import "github.com/quantumclaw/quantumclaw/common/logger"

var AutoGroupEnabled = false

func InitAutoGroup() {
	logger.SysLog("auto group settings initialized")
}
