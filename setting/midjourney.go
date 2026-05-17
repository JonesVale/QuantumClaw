package setting

import "github.com/quantumclaw/quantumclaw/common/logger"

var (
	MidjourneyEnabled   = false
	MidjourneyApiUrl    = ""
	MidjourneyNotifyUrl = ""
)

func InitMidjourneySettings() {
	logger.SysLog("midjourney settings initialized")
}
