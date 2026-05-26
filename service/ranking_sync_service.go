package service

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"gorm.io/gorm/clause"
)

// ── Brand→keyword mapping for external API queries ──

// brandQuery maps known brand names to keywords used in GitHub/HuggingFace searches.
var brandQuery = map[string]string{
	"OpenAI":       "openai",
	"Anthropic":    "anthropic claude",
	"Google":       "gemini google",
	"Meta":         "llama meta",
	"DeepSeek":     "deepseek",
	"Alibaba":      "qwen alibaba",
	"Mistral":      "mistral",
	"Baidu":        "ernie baidu",
	"Zhipu AI":     "glm zhipu",
	"Tencent":      "hunyuan tencent",
	"xAI":          "grok xai",
	"Cohere":       "cohere",
	"Together AI":  "together",
	"AWS":          "bedrock aws",
	"IonQ":         "ionq",
	"IBM":          "ibm quantum",
	"Rigetti":      "rigetti",
	"Azure Quantum":"azure quantum",
	"Google Quantum":"google quantum",
}

// ── HuggingFace trending API ──

type hfTrendingModel struct {
	ID              string `json:"id"`
	Downloads       int    `json:"downloads"`
	Likes           int    `json:"likes"`
	PipelineTag     string `json:"pipeline_tag"`
}

// fetchFromHuggingFaceRankings pulls trending model data from HuggingFace and scores brands by download count.
func fetchFromHuggingFaceRankings() (map[string]float64, error) {
	logger.SysLog("HuggingFace rankings: fetching trending models...")

	req, err := newScrapingRequest("https://huggingface.co/api/trending?limit=50")
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := scrapingClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hf trending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("hf trending status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("hf trending read: %w", err)
	}

	var result struct {
		Data []hfTrendingModel `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("hf trending parse: %w", err)
	}

	// Score brands by download counts from their top models
	brandScores := make(map[string]float64)
	for _, m := range result.Data {
		for brand, keyword := range brandQuery {
			parts := strings.Split(keyword, " ")
			matched := true
			for _, p := range parts {
				if !strings.Contains(strings.ToLower(m.ID), strings.ToLower(p)) {
					matched = false
					break
				}
			}
			if matched {
				brandScores[brand] += float64(m.Downloads)
			}
		}
	}

	logger.SysLog(fmt.Sprintf("HuggingFace rankings: scored %d brands from %d trending models", len(brandScores), len(result.Data)))
	return brandScores, nil
}

// ── GitHub trending repos API ──

type ghRepo struct {
	FullName    string `json:"full_name"`
	Stars       int    `json:"stargazers_count"`
	Description string `json:"description"`
	Language    string `json:"language"`
}

type ghSearchResult struct {
	Items []ghRepo `json:"items"`
}

// fetchFromGitHubRankings searches GitHub for AI/LLM repos and scores brands by star count.
func fetchFromGitHubRankings() (map[string]float64, error) {
	logger.SysLog("GitHub rankings: searching trending AI repos...")

	// Search: AI model repos with most stars
	searchURL := "https://api.github.com/search/repositories?q=llm+OR+language-model+OR+ai-model+OR+transformer&sort=stars&order=desc&per_page=50"
	req, err := newScrapingRequest(searchURL)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := scrapingClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("github read: %w", err)
	}

	var result ghSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("github parse: %w", err)
	}

	// Score brands by star counts
	brandScores := make(map[string]float64)
	for _, repo := range result.Items {
		for brand, keyword := range brandQuery {
			parts := strings.Split(keyword, " ")
			matched := true
			for _, p := range parts {
				fullText := strings.ToLower(repo.FullName + " " + repo.Description)
				if !strings.Contains(fullText, strings.ToLower(p)) {
					matched = false
					break
				}
			}
			if matched {
				brandScores[brand] += float64(repo.Stars)
			}
		}
	}

	logger.SysLog(fmt.Sprintf("GitHub rankings: scored %d brands from %d repos", len(brandScores), len(result.Items)))
	return brandScores, nil
}

// ── Aggregation & upsert ──

// aggregateBrandScores combines multiple score sources into a single weighted score.
func aggregateBrandScores(sources ...map[string]float64) map[string]float64 {
	aggregated := make(map[string]float64)
	maxScores := make(map[string]float64)

	// For each source, track max score for normalization
	for _, source := range sources {
		for brand, score := range source {
			if score > maxScores[brand] {
				maxScores[brand] = score
			}
		}
	}

	// Use maximum score across sources (each source independently confirms popularity)
	for brand, score := range maxScores {
		aggregated[brand] = score
	}

	return aggregated
}

// upsertBrandRanking creates or updates a brand ranking entry.
func upsertBrandRanking(br *model.BrandRanking) error {
	return model.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "brand_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"rank", "score", "metric", "source", "fetched_at"}),
	}).Create(br).Error
}

// ── Main entry point ──

// FetchBrandRankings fetches industry-wide brand usage rankings from external APIs.
// Uses a fallback chain:
// 1. HuggingFace trending API
// 2. GitHub search API
// 3. Baseline seed (always succeeds)
func FetchBrandRankings() time.Duration {
	start := time.Now()
	logger.SysLog("FetchBrandRankings: starting monthly brand ranking update...")

	// Ensure baseline exists first
	var count int64
	model.DB.Model(&model.BrandRanking{}).Count(&count)
	if count == 0 {
		seedBaselineRankings()
	}

	// Try external sources
	hasExternalData := false
	var hfScores, ghScores map[string]float64

	if scores, err := fetchFromHuggingFaceRankings(); err != nil {
		logger.SysWarn(fmt.Sprintf("FetchBrandRankings: HuggingFace API failed: %v", err))
	} else {
		hfScores = scores
		hasExternalData = true
	}

	if scores, err := fetchFromGitHubRankings(); err != nil {
		logger.SysWarn(fmt.Sprintf("FetchBrandRankings: GitHub API failed: %v", err))
	} else {
		ghScores = scores
		hasExternalData = true
	}

	if hasExternalData {
		// Aggregate scores from all sources
		aggregated := aggregateBrandScores(hfScores, ghScores)

		// Sort brands by score descending
		type brandScore struct {
			name  string
			score float64
		}
		var sorted []brandScore
		for brand, score := range aggregated {
			sorted = append(sorted, brandScore{brand, score})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].score > sorted[j].score
		})

		// Normalize scores to 0-100 scale and assign ranks
		maxScore := 100.0
		if len(sorted) > 0 && sorted[0].score > 0 {
			maxScore = sorted[0].score
		}

		now := time.Now().Unix()
		for i, bs := range sorted {
			normScore := (bs.score / maxScore) * 100
			if normScore < 1 {
				normScore = 1
			}
			br := &model.BrandRanking{
				BrandName: bs.name,
				Rank:      i + 1,
				Score:     normScore,
				Metric:    "composite",
				Source:    "huggingface+github",
				FetchedAt: now,
			}
			if err := upsertBrandRanking(br); err != nil {
				logger.SysWarn(fmt.Sprintf("FetchBrandRankings: upsert failed for %s: %v", bs.name, err))
			}
		}

		logger.SysLog(fmt.Sprintf("FetchBrandRankings: %d brands updated from external sources", len(sorted)))
	} else {
		logger.SysLog("FetchBrandRankings: using baseline seed (external sources unavailable)")
	}

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
