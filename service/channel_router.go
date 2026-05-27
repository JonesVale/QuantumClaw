package service

import (
	"math/rand"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// ── Intelligent Router Configuration ──

type RouterConfig struct {
	Enabled                  bool    `json:"enabled"`
	Strategy                 string  `json:"strategy"`                    // "weighted" | "lowest_latency" | "lowest_cost" | "best_success"
	LatencyWeight            float64 `json:"latency_weight"`              // 0.0 - 1.0
	SuccessRateWeight        float64 `json:"success_rate_weight"`         // 0.0 - 1.0
	CostWeight               float64 `json:"cost_weight"`                 // 0.0 - 1.0
	MinObservations          int     `json:"min_observations"`            // minimum requests before scoring
	UnhealthyChannelPenalty  float64 `json:"unhealthy_channel_penalty"`   // score multiplier for unhealthy channels
	ConsecutiveFailThreshold int     `json:"consecutive_fail_threshold"`  // auto-skip after N consecutive failures
	RecoveryIntervalMs       int64   `json:"recovery_interval_ms"`        // ms to wait before retrying a failed channel
	EnableAutoReject         bool    `json:"enable_auto_reject"`          // automatically exclude unhealthy channels
	EnableCostAware          bool    `json:"enable_cost_aware"`           // prefer cheaper channels
	EnableLatencyAware       bool    `json:"enable_latency_aware"`        // prefer faster channels
}

var DefaultRouterConfig = RouterConfig{
	Enabled:                  true,
	Strategy:                 "weighted",
	LatencyWeight:            0.3,
	SuccessRateWeight:        0.4,
	CostWeight:               0.3,
	MinObservations:          5,
	UnhealthyChannelPenalty:  0.1,
	ConsecutiveFailThreshold: 5,
	RecoveryIntervalMs:       60000,
	EnableAutoReject:         true,
	EnableCostAware:          true,
	EnableLatencyAware:       true,
}

var (
	routerConfig     = DefaultRouterConfig
	routerConfigMu   sync.RWMutex
)

func GetRouterConfig() RouterConfig {
	routerConfigMu.RLock()
	defer routerConfigMu.RUnlock()
	return routerConfig
}

func UpdateRouterConfig(cfg RouterConfig) {
	routerConfigMu.Lock()
	defer routerConfigMu.Unlock()
	routerConfig = cfg
	logger.SysLog("intelligent router: config updated")
}

// ── Intelligent Channel Selection ──

// SelectChannelWithWeights selects the best channel from a candidate list
// using weighted random selection based on performance + cost.
// Returns the selected channel id.
func SelectChannelWithWeights(channels []*model.Channel) *model.Channel {
	if len(channels) == 0 {
		return nil
	}
	if len(channels) == 1 {
		return channels[0]
	}

	cfg := GetRouterConfig()
	if !cfg.Enabled {
		// Fallback to random selection
		return channels[rand.Intn(len(channels))]
	}

	// Build scored candidates
	type scoredChannel struct {
		channel *model.Channel
		score   float64
	}
	candidates := make([]scoredChannel, 0, len(channels))

	// Find price range for cost normalization
	minPrice, maxPrice := 0.0, 0.0
	if cfg.EnableCostAware {
		for _, ch := range channels {
			// Use CostPrice directly as price estimate
			price := ch.CostPrice
			if price <= 0 {
				price = 1.0 // default price for channels without cost info
			}
			if minPrice == 0 || price < minPrice {
				minPrice = price
			}
			if price > maxPrice {
				maxPrice = price
			}
		}
	}
	if maxPrice == minPrice {
		minPrice = 0
		maxPrice = 1
	}

	for _, ch := range channels {
		perf := model.GetOrCreateChannelPerformance(ch.Id)

		// Skip unhealthy channels if auto-reject is enabled
		if cfg.EnableAutoReject && !perf.IsHealthy() {
			continue
		}

		price := ch.CostPrice
		if price <= 0 {
			price = 1.0
		}

		score := perf.CombinedScore(price, minPrice, maxPrice)

		// Adjust for strategy preference
		switch cfg.Strategy {
		case "lowest_latency":
			score = perf.CombinedScore(price, minPrice, maxPrice)*0.3 + latencyHeuristic(perf)
		case "lowest_cost":
			score = costHeuristic(price, minPrice, maxPrice)
		case "best_success":
			score = perf.CombinedScore(price, minPrice, maxPrice)*0.7 + successRateHeuristic(perf)*0.3
		}

		candidates = append(candidates, scoredChannel{channel: ch, score: score})
	}

	if len(candidates) == 0 {
		// All channels unhealthy, fallback to original list
		candidates = make([]scoredChannel, len(channels))
		for i, ch := range channels {
			candidates[i] = scoredChannel{channel: ch, score: 1.0}
		}
	}

	// Weighted random selection
	totalScore := 0.0
	for _, c := range candidates {
		if c.score < 0 {
			continue
		}
		totalScore += c.score
	}

	if totalScore <= 0 {
		return candidates[rand.Intn(len(candidates))].channel
	}

	pick := rand.Float64() * totalScore
	cumulative := 0.0
	for _, c := range candidates {
		cumulative += c.score
		if pick <= cumulative {
			return c.channel
		}
	}

	return candidates[len(candidates)-1].channel
}

// latencyHeuristic returns [0,1] based on latency performance.
func latencyHeuristic(perf *model.ChannelPerformance) float64 {
	// Fastest channels get highest score
	latency := perf.AvgLatencyMs
	if latency <= 0 {
		return 1.0
	}
	score := 1.0 - latency/5000.0
	if score < 0 {
		score = 0.1
	}
	return score
}

// costHeuristic returns [0,1] based on cost, cheaper = higher.
func costHeuristic(price, minPrice, maxPrice float64) float64 {
	if maxPrice <= minPrice {
		return 1.0
	}
	return 1.0 - (price-minPrice)/(maxPrice-minPrice)
}

// successRateHeuristic returns [0,1] based on success rate.
func successRateHeuristic(perf *model.ChannelPerformance) float64 {
	return perf.SuccessRate
}

// ── Channel Price Helper ──

// ChannelPriceCache is a simple price cache to avoid repeated DB queries.
var channelPriceCache = struct {
	sync.RWMutex
	prices map[int]float64
}{prices: make(map[int]float64)}

// (p *Channel) GetModelPrice is a helper but we can't extend model.Channel;
// let's use a simple price estimate.
func init() {
	// Start a cleanup goroutine for the price cache
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			channelPriceCache.Lock()
			channelPriceCache.prices = make(map[int]float64)
			channelPriceCache.Unlock()
		}
	}()
}
