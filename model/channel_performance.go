package model

import (
	"math"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ── Channel Performance Metrics ──

// ChannelPerformance stores sliding-window performance data for a channel.
type ChannelPerformance struct {
	ChannelID        int
	LastLatencyMs    int64            `json:"last_latency_ms"`
	AvgLatencyMs     float64          `json:"avg_latency_ms"`
	RequestCount     int64            `json:"request_count"`
	SuccessCount     int64            `json:"success_count"`
	FailureCount     int64            `json:"failure_count"`
	ConsecutiveFails int              `json:"consecutive_fails"`
	LastSuccessAt    int64            `json:"last_success_at"`
	LastFailureAt    int64            `json:"last_failure_at"`
	SuccessRate      float64          `json:"success_rate"`

	// Smoothed moving average data
	latencyBuffer   []int64
	latencyBufIdx   int
	mu              sync.RWMutex
}

const (
	perfWindowSize     = 20    // sliding window of last N requests
	consecutiveFailMax = 5     // auto-reject after this many consecutive failures
	recoveryLatencyMs  = 60000 // recover after 60 seconds of no failures
	latencyWeight      = 0.3   // weight of latency in combined score
	successRateWeight  = 0.4   // weight of success rate
	costWeight         = 0.3   // weight of cost (inverse of price)
)

var (
	channelPerfMap  = make(map[int]*ChannelPerformance)
	channelPerfMu   sync.RWMutex
)

// GetOrCreateChannelPerformance returns the performance tracker for a channel.
func GetOrCreateChannelPerformance(channelID int) *ChannelPerformance {
	channelPerfMu.RLock()
	p, ok := channelPerfMap[channelID]
	channelPerfMu.RUnlock()
	if ok {
		return p
	}

	channelPerfMu.Lock()
	defer channelPerfMu.Unlock()

	// Double-check after acquiring write lock
	if p, ok := channelPerfMap[channelID]; ok {
		return p
	}

	p = &ChannelPerformance{
		ChannelID:     channelID,
		latencyBuffer: make([]int64, 0, perfWindowSize),
	}
	channelPerfMap[channelID] = p
	return p
}

// RecordSuccess records a successful API call for the channel.
func (p *ChannelPerformance) RecordSuccess(latencyMs int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.RequestCount++
	p.SuccessCount++
	p.LastSuccessAt = time.Now().UnixMilli()
	p.ConsecutiveFails = 0

	// Update sliding-window latency buffer
	if len(p.latencyBuffer) < perfWindowSize {
		p.latencyBuffer = append(p.latencyBuffer, latencyMs)
	} else {
		p.latencyBuffer[p.latencyBufIdx%perfWindowSize] = latencyMs
	}
	p.latencyBufIdx++

	// Calculate smoothed average
	var sum int64
	for _, v := range p.latencyBuffer {
		sum += v
	}
	p.AvgLatencyMs = float64(sum) / float64(len(p.latencyBuffer))
	p.LastLatencyMs = latencyMs
	p.SuccessRate = float64(p.SuccessCount) / float64(p.RequestCount)
}

// RecordFailure records a failed API call for the channel.
func (p *ChannelPerformance) RecordFailure(latencyMs int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.RequestCount++
	p.FailureCount++
	p.LastFailureAt = time.Now().UnixMilli()
	p.ConsecutiveFails++

	p.LastLatencyMs = latencyMs

	// Still update latency buffer for failures (slow channels also matter)
	if len(p.latencyBuffer) < perfWindowSize {
		p.latencyBuffer = append(p.latencyBuffer, latencyMs)
	} else {
		p.latencyBuffer[p.latencyBufIdx%perfWindowSize] = latencyMs
	}
	p.latencyBufIdx++

	var sum int64
	for _, v := range p.latencyBuffer {
		sum += v
	}
	p.AvgLatencyMs = float64(sum) / float64(len(p.latencyBuffer))
	p.SuccessRate = float64(p.SuccessCount) / float64(p.RequestCount)
}

// IsHealthy checks if the channel should be used for routing.
func (p *ChannelPerformance) IsHealthy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.RequestCount < 5 {
		// Not enough data yet, assume healthy
		return true
	}

	if p.ConsecutiveFails >= consecutiveFailMax {
		// Check if enough time has passed for recovery
		elapsed := time.Now().UnixMilli() - p.LastFailureAt
		if elapsed < int64(recoveryLatencyMs) {
			return false
		}
		// Grace period elapsed, allow recovery
		return true
	}

	return true
}

// CombinedScore returns a [0,1] score for weighted selection.
// Higher = better. Factors: success rate, latency (inverse), cost (inverse).
func (p *ChannelPerformance) CombinedScore(price float64, minPrice float64, maxPrice float64) float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.RequestCount == 0 {
		return 1.0 // unknown channel gets default score
	}

	// Success rate score [0,1]
	scoreSR := p.SuccessRate

	// Latency score: inverse, normalized to [0,1]
	// Assume 0ms = 1.0, 5000ms+ = 0.0
	latencyScore := 1.0 - math.Min(p.AvgLatencyMs/5000.0, 1.0)

	// Cost score: cheaper is better, normalized to [0,1]
	costScore := 1.0
	if maxPrice > minPrice && price > 0 {
		costScore = 1.0 - (price-minPrice)/(maxPrice-minPrice)
	}

	// Combined weighted score
	score := successRateWeight*scoreSR + latencyWeight*latencyScore + costWeight*costScore
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// ── Router Stats ──

func RecordChannelLatency(channelID int, latencyMs int64, success bool) {
	p := GetOrCreateChannelPerformance(channelID)
	if success {
		p.RecordSuccess(latencyMs)
	} else {
		p.RecordFailure(latencyMs)
	}
}

func IsChannelHealthy(channelID int) bool {
	p := GetOrCreateChannelPerformance(channelID)
	return p.IsHealthy()
}

// GetChannelPerformanceSnapshot returns a copy of all channel performance data.
func GetChannelPerformanceSnapshot() map[int]ChannelPerformance {
	channelPerfMu.RLock()
	defer channelPerfMu.RUnlock()

	snapshot := make(map[int]ChannelPerformance, len(channelPerfMap))
	for id, p := range channelPerfMap {
		p.mu.Lock()
		snapshot[id] = ChannelPerformance{
			ChannelID:        p.ChannelID,
			LastLatencyMs:    p.LastLatencyMs,
			AvgLatencyMs:     p.AvgLatencyMs,
			RequestCount:     p.RequestCount,
			SuccessCount:     p.SuccessCount,
			FailureCount:     p.FailureCount,
			ConsecutiveFails: p.ConsecutiveFails,
			LastSuccessAt:    p.LastSuccessAt,
			LastFailureAt:    p.LastFailureAt,
			SuccessRate:      p.SuccessRate,
		}
		p.mu.Unlock()
	}
	return snapshot
}

// ResetChannelPerformance clears all stored performance data.
func ResetChannelPerformance() {
	channelPerfMu.Lock()
	defer channelPerfMu.Unlock()
	logger.SysLog("intelligent router: resetting all channel performance data")
	channelPerfMap = make(map[int]*ChannelPerformance)
}
