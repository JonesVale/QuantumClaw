package service

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// FetchBrandRankings fetches industry-wide brand usage rankings.
// For MVP: seeds from a curated baseline. Future: fetch from HuggingFace/GitHub APIs.
func FetchBrandRankings() time.Duration {
	start := time.Now()
	logger.SysLog("FetchBrandRankings: starting monthly brand ranking update...")

	// Seed baseline rankings if the table is empty
	var count int64
	model.DB.Model(&model.BrandRanking{}).Count(&count)
	if count == 0 {
		seedBaselineRankings()
	}

	// TODO: Fetch from external sources:
	//   - HuggingFace API: GET huggingface.co/api/trending -> model download rankings
	//   - GitHub API: GET api.github.com/search/repositories?q=llm+stars:>1000
	//   - Artificial Analysis (if accessible)

	elapsed := time.Since(start)
	logger.SysLog(fmt.Sprintf("FetchBrandRankings: completed in %v", elapsed))
	return elapsed
}

// seedBaselineRankings creates an initial curated ranking of major AI/quantum brands.
func seedBaselineRankings() {
	now := time.Now().Unix()
	brands := []struct {
		name  string
		rank  int
		score float64
	}{
		{"OpenAI", 1, 100},
		{"Anthropic", 2, 85},
		{"Google", 3, 80},
		{"Meta", 4, 72},
		{"DeepSeek", 5, 70},
		{"Alibaba", 6, 65},
		{"Mistral", 7, 60},
		{"Baidu", 8, 55},
		{"Zhipu AI", 9, 50},
		{"Tencent", 10, 45},
		{"xAI", 11, 42},
		{"Cohere", 12, 38},
		{"Together AI", 13, 35},
		{"AWS", 14, 33},
		{"IonQ", 15, 30},
		{"IBM", 16, 28},
		{"Rigetti", 17, 25},
		{"Azure Quantum", 18, 23},
		{"Google Quantum", 19, 22},
	}

	for _, b := range brands {
		br := &model.BrandRanking{
			BrandName: b.name,
			Rank:      b.rank,
			Score:     b.score,
			Metric:    "composite",
			Source:    "baseline",
			FetchedAt: now,
		}
		if err := model.DB.Create(br).Error; err != nil {
			logger.SysWarn(fmt.Sprintf("seedBaselineRankings: create failed for %s: %v", b.name, err))
		}
	}
	logger.SysLog(fmt.Sprintf("seedBaselineRankings: %d brands seeded", len(brands)))
}
