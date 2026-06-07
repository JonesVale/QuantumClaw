package monitor

import (
	"fmt"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ── 渠道-模型 冷期机制 ──
// 当渠道 A 的模型 M 失败且重试成功时，(A, M) 进入 30 分钟冷期
// 冷期内路由会跳过该渠道上的该模型（不影响该渠道的其他模型）

type penaltyKey struct {
	ChannelID int
	ModelName string
}

var (
	penaltyBox     = make(map[penaltyKey]int64) // key → 解禁时间戳
	penaltyBoxMu   sync.RWMutex
	penaltyPeriod  = 30 * time.Minute
)

// PenalizeModel 将(渠道ID, 模型名)标记为冷期
func PenalizeModel(channelID int, modelName string) {
	if channelID <= 0 || modelName == "" {
		return
	}
	key := penaltyKey{ChannelID: channelID, ModelName: modelName}
	releaseAt := time.Now().Add(penaltyPeriod).Unix()
	penaltyBoxMu.Lock()
	penaltyBox[key] = releaseAt
	penaltyBoxMu.Unlock()
	logger.SysWarnf("[PENALTY] (channel=%d, model=%s) penalized until %s",
		channelID, modelName, time.Unix(releaseAt, 0).Format("15:04:05"))
}

// IsModelPenalized 检查(渠道ID, 模型名)是否在冷期中
func IsModelPenalized(channelID int, modelName string) bool {
	if channelID <= 0 || modelName == "" {
		return false
	}
	key := penaltyKey{ChannelID: channelID, ModelName: modelName}
	penaltyBoxMu.RLock()
	releaseAt, ok := penaltyBox[key]
	penaltyBoxMu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().Unix() >= releaseAt {
		penaltyBoxMu.Lock()
		delete(penaltyBox, key)
		penaltyBoxMu.Unlock()
		return false
	}
	return true
}

// GetPenalizedKeys 返回当前所有冷期中的（运营看板用）
func GetPenalizedKeys() []string {
	penaltyBoxMu.RLock()
	defer penaltyBoxMu.RUnlock()
	now := time.Now().Unix()
	var result []string
	for key, releaseAt := range penaltyBox {
		if releaseAt > now {
			result = append(result, fmt.Sprintf("ch=%d model=%s until=%s",
				key.ChannelID, key.ModelName, time.Unix(releaseAt, 0).Format("15:04:05")))
		}
	}
	return result
}

// CleanupPenaltyBox 清理已过期的冷期记录
func CleanupPenaltyBox() int {
	penaltyBoxMu.Lock()
	defer penaltyBoxMu.Unlock()
	now := time.Now().Unix()
	count := 0
	for key, releaseAt := range penaltyBox {
		if releaseAt <= now {
			delete(penaltyBox, key)
			count++
		}
	}
	return count
}
