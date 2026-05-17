package system_setting

import "github.com/quantumclaw/quantumclaw/common/logger"

var DiscordEnabled = false

func InitDiscordSettings() {
	logger.SysLog("discord settings initialized")
}
