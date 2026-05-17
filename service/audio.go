package service

import (
	"github.com/quantumclaw/quantumclaw/common/logger"
)

func InitAudioService() {
	logger.SysLog("audio service initialized")
}

func GetAudioDuration(data []byte) float64 {
	return 0
}

func EstimateAudioTokens(durationSeconds float64) int {
	return int(durationSeconds * 50)
}
