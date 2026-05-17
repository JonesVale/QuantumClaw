package setting

import "github.com/quantumclaw/quantumclaw/common/logger"

var (
	UserUsableGroupEnabled = false
	DefaultUsableGroups    = []string{"default"}
)

func InitUserUsableGroup() {
	logger.SysLog("user usable group settings initialized")
}
