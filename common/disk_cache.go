package common

import (
	"os"
	"path/filepath"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

func GetDiskSpaceInfo() DiskSpaceInfo {
	var info DiskSpaceInfo
	return info
}

func CleanupOldCacheFiles() {
	cacheDir := "./cache"
	maxAge := 7 * 24 * time.Hour

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		logger.SysLog("Failed to read cache directory: " + err.Error())
		return
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(cacheDir, entry.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			if err := os.Remove(fullPath); err != nil {
				logger.SysLog("Failed to remove old cache file: " + fullPath)
			}
		}
	}
}