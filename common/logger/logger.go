package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/helper"
)

// sensitiveFields are field names whose values should be redacted from logs
var sensitiveFields = []string{
	"password",
	"secret",
	"token",
	"key",
	"authorization",
	"access_token",
	"refresh_token",
	"api_key",
	"private_key",
	"secret_key",
}

// sensitivePatterns are compiled regex patterns for efficient sanitization
var sensitivePatterns []*regexp.Regexp

func init() {
	for _, field := range sensitiveFields {
		// Match JSON keys: "field": "value" or "field":"value"
		pattern := fmt.Sprintf(`"%s"\s*:\s*"[^"]+"`, regexp.QuoteMeta(field))
		re, err := regexp.Compile(pattern)
		if err == nil {
			sensitivePatterns = append(sensitivePatterns, re)
		}
		// Match URL query params: field=value& or field=value
		pattern2 := fmt.Sprintf(`(%s=)[^&\s]+`, regexp.QuoteMeta(field))
		re2, err2 := regexp.Compile(pattern2)
		if err2 == nil {
			sensitivePatterns = append(sensitivePatterns, re2)
		}
		// Match form-encoded or Authorization header
		pattern3 := fmt.Sprintf(`(?i)(Authorization|Bearer|%s)\s*[:=]\s*\S+`, regexp.QuoteMeta(field))
		re3, err3 := regexp.Compile(pattern3)
		if err3 == nil {
			sensitivePatterns = append(sensitivePatterns, re3)
		}
	}
}

// SanitizeLogInput redacts sensitive information from log messages
func SanitizeLogInput(data string) string {
	result := data
	for _, re := range sensitivePatterns {
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			// Preserve the key/field name but replace the value
			idx := strings.LastIndexAny(match, "=:")
			if idx >= 0 && idx < len(match)-1 {
				return match[:idx+1] + "***"
			}
			return match
		})
	}
	return result
}

type loggerLevel string

const (
	loggerDEBUG loggerLevel = "DEBUG"
	loggerINFO  loggerLevel = "INFO"
	loggerWarn  loggerLevel = "WARN"
	loggerError loggerLevel = "ERROR"
	loggerFatal loggerLevel = "FATAL"
)

var setupLogOnce sync.Once

func SetupLogger() {
	setupLogOnce.Do(func() {
		if LogDir != "" {
			var logPath string
			if config.OnlyOneLogFile {
				logPath = filepath.Join(LogDir, "quantumclaw.log")
			} else {
				logPath = filepath.Join(LogDir, fmt.Sprintf("quantumclaw-%s.log", time.Now().Format("20060102")))
			}
			fd, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal("failed to open log file")
			}
			gin.DefaultWriter = io.MultiWriter(os.Stdout, fd)
			gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, fd)
		}
	})
}

func SysLog(s string) {
	logHelper(nil, loggerINFO, s)
}

func SysLogf(format string, a ...any) {
	logHelper(nil, loggerINFO, fmt.Sprintf(format, a...))
}

func SysWarn(s string) {
	logHelper(nil, loggerWarn, s)
}

func SysWarnf(format string, a ...any) {
	logHelper(nil, loggerWarn, fmt.Sprintf(format, a...))
}

func SysError(s string) {
	logHelper(nil, loggerError, s)
}

func SysErrorf(format string, a ...any) {
	logHelper(nil, loggerError, fmt.Sprintf(format, a...))
}

func Debug(ctx context.Context, msg string) {
	if !config.DebugEnabled {
		return
	}
	logHelper(ctx, loggerDEBUG, msg)
}

func Info(ctx context.Context, msg string) {
	logHelper(ctx, loggerINFO, msg)
}

func Warn(ctx context.Context, msg string) {
	logHelper(ctx, loggerWarn, msg)
}

func Error(ctx context.Context, msg string) {
	logHelper(ctx, loggerError, msg)
}

func Debugf(ctx context.Context, format string, a ...any) {
	if !config.DebugEnabled {
		return
	}
	logHelper(ctx, loggerDEBUG, fmt.Sprintf(format, a...))
}

func Infof(ctx context.Context, format string, a ...any) {
	logHelper(ctx, loggerINFO, fmt.Sprintf(format, a...))
}

func Warnf(ctx context.Context, format string, a ...any) {
	logHelper(ctx, loggerWarn, fmt.Sprintf(format, a...))
}

func Errorf(ctx context.Context, format string, a ...any) {
	logHelper(ctx, loggerError, fmt.Sprintf(format, a...))
}

func FatalLog(s string) {
	logHelper(nil, loggerFatal, s)
}

func FatalLogf(format string, a ...any) {
	logHelper(nil, loggerFatal, fmt.Sprintf(format, a...))
}

func logHelper(ctx context.Context, level loggerLevel, msg string) {
	writer := gin.DefaultErrorWriter
	if level == loggerINFO {
		writer = gin.DefaultWriter
	}
	var requestId string
	if ctx != nil {
		rawRequestId := helper.GetRequestID(ctx)
		if rawRequestId != "" {
			requestId = fmt.Sprintf(" | %s", rawRequestId)
		}
	}
	lineInfo, funcName := getLineInfo()
	now := time.Now()
	// Sanitize sensitive data from log messages
	sanitizedMsg := SanitizeLogInput(msg)
	_, _ = fmt.Fprintf(writer, "[%s] %v%s%s %s%s \n", level, now.Format("2006/01/02 - 15:04:05"), requestId, lineInfo, funcName, sanitizedMsg)
	SetupLogger()
	if level == loggerFatal {
		os.Exit(1)
	}
}

func getLineInfo() (string, string) {
	funcName := "[unknown] "
	pc, file, line, ok := runtime.Caller(3)
	if ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			parts := strings.Split(fn.Name(), ".")
			funcName = "[" + parts[len(parts)-1] + "] "
		}
	} else {
		file = "unknown"
		line = 0
	}
	parts := strings.Split(file, "quantumclaw/")
	if len(parts) > 1 {
		file = parts[1]
	}
	return fmt.Sprintf(" | %s:%d", file, line), funcName
}
