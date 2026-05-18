package common

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ==================== SMTP 邮件服务 ====================

func InitEmailConfig() {
	if config.SMTPServer != "" && config.SMTPAccount != "" {
		logger.SysLog(fmt.Sprintf("email service configured: server=%s account=%s", config.SMTPServer, config.SMTPAccount))
	} else {
		logger.SysLog("email service not configured (SMTP settings empty)")
	}
}

// SendEmail 发送邮件
// 如果 SMTP 未配置，记录日志但不报错（开发模式友好）
func SendEmail(to string, subject string, body string) error {
	if config.SMTPServer == "" || config.SMTPAccount == "" {
		logger.SysLog(fmt.Sprintf("[email stub] to=%s subject=%s", to, subject))
		return nil
	}

	auth := smtp.PlainAuth("", config.SMTPAccount, config.SMTPToken, config.SMTPServer)

	from := config.SMTPFrom
	if from == "" {
		from = config.SMTPAccount
	}

	// 构建邮件头部
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", config.SMTPServer, config.SMTPPort)
	if err := smtp.SendMail(addr, auth, config.SMTPAccount, []string{to}, []byte(msg.String())); err != nil {
		logger.SysError(fmt.Sprintf("email send failed: to=%s error=%q", to, err.Error()))
		return fmt.Errorf("send email: %w", err)
	}

	logger.SysLog(fmt.Sprintf("email sent: to=%s subject=%s", to, subject))
	return nil
}
