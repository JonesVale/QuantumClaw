package service

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// StartDailyModelSync 每天零点自动从上游 API 同步模型列表到数据库
func StartDailyModelSync() {
	now := time.Now()
	// 计算到下一个零点的时间
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(24 * time.Hour)
	duration := nextMidnight.Sub(now)

	logger.SysLog(fmt.Sprintf("daily model sync scheduled: first run at %s (in %v)", nextMidnight.Format("15:04:05"), duration))

	// 等待到第一个零点
	time.Sleep(duration)

	// 每天执行一次
	for {
		syncAllChannels()
		time.Sleep(24 * time.Hour)
	}
}

func syncAllChannels() {
	logger.SysLog("daily model sync: starting...")

	var channels []model.Channel
	model.DB.Find(&channels)

	successCount := 0
	failCount := 0

	for _, ch := range channels {
		// 跳过没有有效 API Key 的渠道
		if ch.Key == "" || len(ch.Key) < 8 || ch.Status != model.ChannelStatusEnabled {
			continue
		}
		err := ch.UpdateModelsFromProvider()
		if err != nil {
			logger.SysWarn(fmt.Sprintf("daily model sync: channel #%d %s sync failed: %v", ch.Id, ch.Name, err))
			failCount++
		} else {
			successCount++
		}
	}

	logger.SysLog(fmt.Sprintf("daily model sync: completed (%d success, %d failed, %d skipped)",
		successCount, failCount, len(channels)-successCount-failCount))
}
