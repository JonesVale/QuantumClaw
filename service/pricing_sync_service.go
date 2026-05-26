package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/billing/ratio"
	"gorm.io/gorm/clause"
)

// ── HTTP client for scrapers ──

var scrapingClient = &http.Client{
	Timeout: 30 * time.Second,
}

func newScrapingRequest(rawURL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/json,*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8")
	return req, nil
}

// ── Global upsert helper ──

// upsertReferencePrice creates or updates a reference_price entry.
func upsertReferencePrice(rp *model.ReferencePrice) error {
	return model.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "model_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"input_price", "output_price", "source", "fetched_at"}),
	}).Create(rp).Error
}

// ── Seed from ModelRatio (baseline + monthly refresh) ──

// SeedReferencePrices populates reference_prices from ModelRatio + ModelMetadata.
// Changed to full upsert so the monthly cron always has synchronized data.
func SeedReferencePrices() {
	now := time.Now().Unix()
	inserted := 0

	var metadata []model.ModelMetadata
	model.DB.Where("languages_type = ?", "English").Find(&metadata)

	for _, m := range metadata {
		normName := normalizeForRatio(m.ModelName)
		modelRatioVal := ratio.GetModelRatio(normName, 0)
		compRatio := ratio.GetCompletionRatio(normName, 0)

		if modelRatioVal <= 0 {
			modelRatioVal = ratio.GetModelRatio(m.ModelName, 0)
			compRatio = ratio.GetCompletionRatio(m.ModelName, 0)
		}

		inputPrice := modelRatioVal * 0.002
		outputPrice := inputPrice * compRatio

		provider := m.Provider
		if provider == "" {
			provider = "Unknown"
		}

		rp := &model.ReferencePrice{
			ModelName:   m.ModelName,
			Provider:    provider,
			InputPrice:  inputPrice,
			OutputPrice: outputPrice,
			Currency:    "USD",
			Source:      "modelratio",
			FetchedAt:   now,
		}

		// Insert only if missing — never overwrite existing (preserves admin edits)
		var existingPrice model.ReferencePrice
		if model.DB.Where("model_name = ?", m.ModelName).First(&existingPrice).Error != nil {
			if err := model.DB.Create(rp).Error; err != nil {
				logger.SysWarn(fmt.Sprintf("SeedReferencePrices: insert failed for %s: %v", m.ModelName, err))
			} else {
				inserted++
			}
		}
	}

	logger.SysLog(fmt.Sprintf("SeedReferencePrices: %d inserted (missing rows only, existing preserved)", inserted))
}

// ── HuggingFace pricing scraper ──

// HuggingFace API response for model info
type hfModelInfo struct {
	ID          string `json:"id"`
	PipelineTag string `json:"pipeline_tag"`
	ModelID     string `json:"modelId"`
	CardData    *struct {
		BaseModel       string            `json:"base_model"`
		License         string            `json:"license"`
		ModelIndex      []hfModelIndex    `json:"model-index"`
		Pricing         map[string]string `json:"pricing"`
		Provider        map[string]string `json:"provider"`
		CostPerMillion  map[string]float64 `json:"-"`
	} `json:"cardData"`
}

type hfModelIndex struct {
	Name      string           `json:"name"`
	Results   []hfModelResult  `json:"results"`
	Task      *hfTask          `json:"task"`
	Longtext  string           `json:"long_text"`
}

type hfModelResult struct {
	Task  *hfTask         `json:"task"`
	Metrics []hfMetric   `json:"metrics"`
}

type hfTask struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type hfMetric struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Name  string  `json:"name"`
}

// fetchFromHuggingFace scrapes model pricing data from HuggingFace API.
// Many proprietary models have pricing info in their card data.
func fetchFromHuggingFace() error {
	logger.SysLog("HuggingFace pricing: fetching model list...")

	// Get popular models (sorted by downloads)
	modelsURL := "https://huggingface.co/api/models?sort=downloads&direction=-1&limit=50"
	req, err := newScrapingRequest(modelsURL)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	resp, err := scrapingClient.Do(req)
	if err != nil {
		return fmt.Errorf("hf request failed: %w", err)
	}
	defer func() { io.Copy(io.Discard, resp.Body); resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return fmt.Errorf("hf status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return fmt.Errorf("hf read: %w", err)
	}

	var models []hfModelInfo
	if err := json.Unmarshal(body, &models); err != nil {
		return fmt.Errorf("hf parse: %w", err)
	}

	found := 0
	for _, m := range models {
		if m.CardData == nil || m.CardData.Pricing == nil {
			continue
		}
		// Parse pricing info from card data
		// Format typically: {"input": "$X.XX per 1M tokens", "output": "$X.XX per 1M tokens"}
		inputPrice := parseHfPricingValue(m.CardData.Pricing["input"])
		outputPrice := parseHfPricingValue(m.CardData.Pricing["output"])
		if inputPrice <= 0 && outputPrice <= 0 {
			continue
		}

		// Determine provider and model name
		modelName := m.ID
		provider := extractProviderFromModelID(modelName)

		rp := &model.ReferencePrice{
			ModelName:   modelName,
			Provider:    provider,
			InputPrice:  inputPrice,
			OutputPrice: outputPrice,
			Currency:    "USD",
			Source:      "huggingface",
			FetchedAt:   time.Now().Unix(),
		}
		if err := upsertReferencePrice(rp); err != nil {
			logger.SysWarn(fmt.Sprintf("HuggingFace pricing upsert failed for %s: %v", modelName, err))
		} else {
			found++
		}
	}

	logger.SysLog(fmt.Sprintf("HuggingFace pricing: %d models with pricing data processed", found))
	return nil
}

// parseHfPricingValue extracts a per-1K-token price from a HuggingFace pricing string.
// Expected format: "$5.00 per 1M tokens" or "$0.015 per 1K tokens"
func parseHfPricingValue(s string) float64 {
	if s == "" {
		return 0
	}
	s = strings.TrimSpace(s)

	// Match $X.XX per Y (K|M) tokens
	re := regexp.MustCompile(`\$(\d+\.?\d*)\s*/\s*(\d+(?:\.\d+)?)\s*(K|M|k|m)?\s*tokens?`)
	matches := re.FindStringSubmatch(s)
	if len(matches) >= 4 {
		val, _ := strconv.ParseFloat(matches[1], 64)
		qty, _ := strconv.ParseFloat(matches[2], 64)
		unit := strings.ToUpper(matches[3])

		// Convert to per-1K tokens
		var divisor float64
		switch unit {
		case "K":
			divisor = qty * 1000 / 1000 // e.g., $5/1K tokens = $5 per 1K tokens
		case "M":
			divisor = qty * 1000000 / 1000 // e.g., $5/1M tokens = $0.005 per 1K tokens
		default:
			divisor = qty
		}

		if divisor > 0 {
			return val / (qty * 1000 / 1000) // normalize to per 1K
		}
	}
	return 0
}

// extract provider from huggingface model ID (before the /)
func extractProviderFromModelID(id string) string {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) >= 2 {
		return parts[0]
	}
	return "Unknown"
}

// ── OpenAI pricing from API docs ──

// openAIModelEntry represents a model entry from OpenAI's model list
type openAIModelEntry struct {
	ID     string `json:"id"`
	Owned  string `json:"owned_by"`
}

// fetchOpenAIModelPricing fetches model names from OpenAI API and matches with ModelRatio for pricing.
// This doesn't give us exact prices, but gives us model names and ownership info.
func fetchOpenAIModelPricing() error {
	logger.SysLog("OpenAI pricing: fetching model list...")

	req, err := newScrapingRequest("https://api.openai.com/v1/models")
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	resp, err := scrapingClient.Do(req)
	if err != nil {
		return fmt.Errorf("openai api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("openai api status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return fmt.Errorf("openai api read: %w", err)
	}

	var result struct {
		Data []openAIModelEntry `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("openai api parse: %w", err)
	}

	// For each model, try to match with ModelRatio for pricing
	now := time.Now().Unix()
	found := 0
	for _, m := range result.Data {
		ratioVal := ratio.GetModelRatio(m.ID, 0)
		if ratioVal <= 0 {
			continue
		}

		compRatio := ratio.GetCompletionRatio(m.ID, 0)
		if compRatio <= 0 {
			compRatio = 1
		}

		provider := "OpenAI"
		if m.Owned != "system" && m.Owned != "" {
			provider = m.Owned
		}

		rp := &model.ReferencePrice{
			ModelName:   m.ID,
			Provider:    provider,
			InputPrice:  ratioVal * 0.002,
			OutputPrice: ratioVal * 0.002 * compRatio,
			Currency:    "USD",
			Source:      "openai-models-api",
			FetchedAt:   now,
		}
		if err := upsertReferencePrice(rp); err != nil {
			logger.SysWarn(fmt.Sprintf("OpenAI pricing upsert failed for %s: %v", m.ID, err))
		} else {
			found++
		}
	}

	logger.SysLog(fmt.Sprintf("OpenAI pricing: %d models with pricing data from models API", found))
	return nil
}

// ── Main entry point ──

// FetchOfficialPricing attempts to fetch official pricing from provider APIs.
// Uses a fallback chain:
// 1. HuggingFace API (model card pricing data)
// 2. OpenAI Models API (model names + ModelRatio pricing)
// 3. ModelRatio seed (always succeeds)
func FetchOfficialPricing() time.Duration {
	start := time.Now()
	logger.SysLog("FetchOfficialPricing: starting monthly official pricing update...")

	// Tier 1: Try external APIs
	hasExternalData := false

	if err := fetchFromHuggingFace(); err != nil {
		logger.SysWarn(fmt.Sprintf("FetchOfficialPricing: HuggingFace API failed: %v", err))
	} else {
		hasExternalData = true
	}

	if err := fetchOpenAIModelPricing(); err != nil {
		logger.SysWarn(fmt.Sprintf("FetchOfficialPricing: OpenAI API failed: %v", err))
	} else {
		hasExternalData = true
	}

	// Tier 2: Always refresh from ModelRatio as full baseline
	SeedReferencePrices()

	if hasExternalData {
		logger.SysLog("FetchOfficialPricing: external pricing data merged successfully")
	} else {
		logger.SysLog("FetchOfficialPricing: using ModelRatio baseline only (external sources unavailable)")
	}

	elapsed := time.Since(start)
	logger.SysLog(fmt.Sprintf("FetchOfficialPricing: completed in %v", elapsed))
	return elapsed
}

// normalizeForRatio converts a display name to ModelRatio key format.
func normalizeForRatio(name string) string {
	b := []byte(name)
	var out []byte
	for _, c := range b {
		if c >= 'A' && c <= 'Z' {
			out = append(out, c+'a'-'A')
		} else if c == ' ' {
			out = append(out, '-')
		} else {
			out = append(out, c)
		}
	}
	return string(out)
}
