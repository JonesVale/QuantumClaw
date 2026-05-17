package common

import (
	"fmt"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

var EmailFrom string
var EmailSMTPHost string

func InitEmailConfig() {
	logger.SysLog("email config initialized")
}

func SendEmail(to string, subject string, body string) error {
	logger.SysLog(fmt.Sprintf("email would be sent to %s: %s", to, subject))
	return nil
}
