package model

import (
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ── 运行期失败追踪（滑动窗口）──
type ChannelFailureTracker struct {
	mu              sync.RWMutex
	ChannelID       int
	ConsecutiveFail int           // 连续失败次数（成功时归零）
	RecentFailures  []int64       // 最近 1h 内失败的时间戳
	RecentSuccesses []int64       // 最近 1h 内成功的时间戳
	ObservationMode bool          // 观察期标记
	FirstSeenAt     int64         // 首次进入路由的时间
}

var (
	failureTrackers     = make(map[int]*ChannelFailureTracker)
	failureTrackerMu    sync.RWMutex
)

const (
	// 连续失败 N 次后自动降权
	ConsecutiveFailThreshold = 5
	// 滑动窗口内失败率超过此值降权
	FailureRateThreshold = 0.15
	// 滑动窗口时间
	SlidingWindowMinutes = 60
	// 新 Provider 观察期（小时）
	ObservationPeriodHours = 48
	// 观察期内允许的最大失败率
	ObservationMaxFailureRate = 0.10
)

// GetOrCreateFailureTracker 获取或创建渠道失败追踪器
func GetOrCreateFailureTracker(channelID int) *ChannelFailureTracker {
	failureTrackerMu.Lock()
	defer failureTrackerMu.Unlock()

	if t, ok := failureTrackers[channelID]; ok {
		return t
	}
	t := &ChannelFailureTracker{
		ChannelID:       channelID,
		ObservationMode: true,
		FirstSeenAt:     time.Now().Unix(),
	}
	failureTrackers[channelID] = t
	// 观察期：前 48 小时或前 100 次请求
	go func() {
		time.Sleep(ObservationPeriodHours * time.Hour)
		t.mu.Lock()
		t.ObservationMode = false
		t.mu.Unlock()
	}()
	return t
}

// RecordSuccess 记录一次成功调用
func (t *ChannelFailureTracker) RecordSuccess() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ConsecutiveFail = 0
	now := time.Now().Unix()
	t.RecentSuccesses = append(t.RecentSuccesses, now)
	// 裁剪窗口
	t.pruneLocked(now)
}

// RecordFailure 记录一次失败调用，返回是否需要降权/禁用
func (t *ChannelFailureTracker) RecordFailure() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ConsecutiveFail++
	now := time.Now().Unix()
	t.RecentFailures = append(t.RecentFailures, now)
	t.pruneLocked(now)

	// 连续失败阈值检查
	if t.ConsecutiveFail >= ConsecutiveFailThreshold {
		logger.SysWarnf("[PROTECT] channel %d: %d consecutive failures, deprioritizing",
			t.ChannelID, t.ConsecutiveFail)
		return true
	}

	// 滑动窗口失败率检查
	failures := len(t.RecentFailures)
	total := failures + len(t.RecentSuccesses)
	if total > 10 { // 至少有 10 次请求才统计
		rate := float64(failures) / float64(total)
		threshold := FailureRateThreshold
		if t.ObservationMode {
			threshold = ObservationMaxFailureRate
		}
		if rate > threshold {
			logger.SysWarnf("[PROTECT] channel %d: failure rate %.1f%% exceeds %.0f%% threshold",
				t.ChannelID, rate*100, threshold*100)
			return true
		}
	}

	return false
}

// GetHealthScore 返回渠道健康分（0.0 ~ 1.0），用于路由权重
func (t *ChannelFailureTracker) GetHealthScore() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	failures := len(t.RecentFailures)
	total := failures + len(t.RecentSuccesses)
	if total == 0 {
		if t.ObservationMode {
			return 0.5 // 观察期内无数据，给中低分
		}
		return 1.0 // 老渠道无数据 = 大概率正常
	}
	rate := float64(failures) / float64(total)
	score := 1.0 - rate*2 // 0% -> 1.0, 10% -> 0.8, 50% -> 0.0
	if score < 0.1 {
		score = 0.1 // 最低保底 0.1，不完全杀死
	}
	if t.ObservationMode {
		score *= 0.8 // 观察期再加 20% 折扣
	}
	return score
}

// IsInObservation 是否在观察期
func (t *ChannelFailureTracker) IsInObservation() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ObservationMode
}

// GetStats 返回当前统计（调试/监控用）
func (t *ChannelFailureTracker) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return map[string]interface{}{
		"channel_id":        t.ChannelID,
		"consecutive_fail":  t.ConsecutiveFail,
		"recent_failures":   len(t.RecentFailures),
		"recent_successes":  len(t.RecentSuccesses),
		"observation_mode":  t.ObservationMode,
		"first_seen_at":     t.FirstSeenAt,
	}
}

// pruneLocked 清理窗口外的记录
func (t *ChannelFailureTracker) pruneLocked(now int64) {
	cutoff := now - SlidingWindowMinutes*60
	t.RecentFailures = pruneSlice(t.RecentFailures, cutoff)
	t.RecentSuccesses = pruneSlice(t.RecentSuccesses, cutoff)
}

func pruneSlice(s []int64, cutoff int64) []int64 {
	for i, v := range s {
		if v > cutoff {
			return s[i:]
		}
	}
	return s[:0]
}

// ── 全局清理 ──

// CleanupStaleTrackers 清理超过 24 小时无活跃的追踪器
func CleanupStaleTrackers() {
	failureTrackerMu.Lock()
	defer failureTrackerMu.Unlock()
	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	for id, t := range failureTrackers {
		t.mu.RLock()
		lastActive := t.FirstSeenAt
		t.mu.RUnlock()
		if lastActive < cutoff {
			delete(failureTrackers, id)
		}
	}
}
