package setting

import "github.com/quantumclaw/quantumclaw/common/logger"

var (
	ChatContextEnabled = false
	ChatModel          = "gpt-3.5-turbo"
	ChatMaxTokens      = 4096
)

func InitChatSettings() {
	logger.SysLog("chat settings initialized")
}
