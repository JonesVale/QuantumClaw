package common

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	Port         = flag.Int("port", 3666, "the listening port")
	PrintVersion = flag.Bool("version", false, "print version and exit")
	PrintHelp    = flag.Bool("help", false, "print help and exit")
	LogDir       = flag.String("log-dir", "./logs", "specify the log directory")
)

// CleanupOldLogs 清理 logs/ 目录下超过 7 天的 .log 文件
func CleanupOldLogs() {
	logDir := *LogDir
	if logDir == "" {
		logDir = "./logs"
	}
	entries, err := os.ReadDir(logDir)
	if err != nil {
		logger.SysError("failed to read log directory for cleanup: " + err.Error())
		return
	}
	now := time.Now()
	deletedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".log") {
			continue
		}
		// 从文件名提取日期: quantumclaw-YYYYMMDD.log 或 oneapi-YYYYMMDD.log
		var dateStr string
		if strings.HasPrefix(name, "quantumclaw-") || strings.HasPrefix(name, "oneapi-") {
			dateStr = strings.TrimSuffix(strings.TrimPrefix(name, "quantumclaw-"), ".log")
			dateStr = strings.TrimSuffix(strings.TrimPrefix(dateStr, "oneapi-"), ".log")
		}
		if len(dateStr) != 8 {
			continue
		}
		logDate, err := time.Parse("20060102", dateStr)
		if err != nil {
			continue
		}
		if now.Sub(logDate).Hours() > 7*24 {
			fullPath := filepath.Join(logDir, name)
			if rmErr := os.Remove(fullPath); rmErr != nil {
				logger.SysError("failed to remove old log: " + fullPath + ": " + rmErr.Error())
			} else {
				deletedCount++
				logger.SysLog("removed old log: " + fullPath)
			}
		}
	}
	if deletedCount > 0 {
		logger.SysLogf("cleaned up %d old log file(s) (older than 7 days)", deletedCount)
	}
}

func printHelp() {
	fmt.Println("QuantumClaw " + Version + " - AI API Gateway & Management Platform.")
	fmt.Println("Copyright (C) 2023 JustSong. All rights reserved.")
	fmt.Println("GitHub: https://github.com/quantumclaw/quantumclaw")
	fmt.Println("Usage: quantumclaw [--port <port>] [--log-dir <log directory>] [--version] [--help]")
}

func Init() {
	flag.Parse()

	if *PrintVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	if *PrintHelp {
		printHelp()
		os.Exit(0)
	}

	// 1. 优先使用环境变量（最高优先级）
	if envSecret := os.Getenv("SESSION_SECRET"); envSecret != "" {
		if envSecret == "random_string" {
			logger.SysError("SESSION_SECRET is set to an example value, please change it to a random string.")
		} else {
			config.SessionSecret = envSecret
		}
	} else {
		// 2. 尝试从 .session_secret 文件读取
		if data, err := os.ReadFile(".session_secret"); err == nil {
			secret := strings.TrimSpace(string(data))
			if secret != "" {
				config.SessionSecret = secret
			}
		}
	}

	// 3. 如果仍然是默认的随机 UUID（每次重启不同），生成持久化密钥
	// config.SessionSecret 的 var 默认值为 uuid.New().String()（每次不同）
	// 这里检测到是默认值（首次启动且无 .env/.session_secret）则持久化
	if _, statErr := os.Stat(".session_secret"); os.IsNotExist(statErr) && os.Getenv("SESSION_SECRET") == "" {
		// 首次启动：持久化当前 secret
		if err := os.WriteFile(".session_secret", []byte(config.SessionSecret), 0600); err != nil {
			logger.SysError("failed to persist session_secret: " + err.Error())
		} else {
			logger.SysLog("generated new session secret, persisted to .session_secret")
		}
	}

	// 4. Enhance SessionSecret with quantum random data if available
	if QRNGEnabled {
		quantumBytes, err := GetQuantumRandomBytes(16)
		if err == nil {
			hasher := sha256.New()
			hasher.Write([]byte(config.SessionSecret))
			hasher.Write(quantumBytes)
			config.SessionSecret = hex.EncodeToString(hasher.Sum(nil))
			logger.SysLog("session secret enhanced with quantum random data")
		}
	}
	if os.Getenv("SQLITE_PATH") != "" {
		SQLitePath = os.Getenv("SQLITE_PATH")
	}
	if *LogDir != "" {
		var err error
		*LogDir, err = filepath.Abs(*LogDir)
		if err != nil {
			logger.SysError("failed to resolve log directory: " + err.Error())
			*LogDir = "./logs"
		} else {
			if _, statErr := os.Stat(*LogDir); os.IsNotExist(statErr) {
				if mkErr := os.Mkdir(*LogDir, 0777); mkErr != nil {
					logger.SysError("failed to create log directory: " + mkErr.Error())
					*LogDir = "./logs"
				}
			}
		}
		logger.LogDir = *LogDir
	}

	// 启动时清理超过 7 天的旧日志
	CleanupOldLogs()

	// 启动验证码过期清理任务
	StartVerificationCleanupTask()
}
