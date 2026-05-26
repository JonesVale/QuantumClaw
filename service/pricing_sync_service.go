package service

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/billing/ratio"
)

// SeedReferencePrices populates reference_prices from two sources:
// 1. ModelRatio data (baseline, all models in model_metadata)
// 2. ModelMetadata table (for model names and provider info)
// Only inserts rows that don't already exist (idempotent).
func SeedReferencePrices() {
	now := time.Now().Unix()
	inserted := 0
	skipped := 0

	// Collect existing entries to avoid duplicates
	var existing []string
	model.DB.Model(&model.ReferencePrice{}).Pluck("model_name", &existing)
	existingMap := make(map[string]bool, len(existing))
	for _, n := range existing {
		existingMap[n] = true
	}

	// Get all unique models from model_metadata (English lang as baseline)
	var metadata []model.ModelMetadata
	model.DB.Where("languages_type = ?", "English").Find(&metadata)

	for _, m := range metadata {
		if existingMap[m.ModelName] {
			skipped++
			continue
		}

		// Normalize name to match ModelRatio keys
		normName := normalizeForRatio(m.ModelName)
		modelRatioVal := ratio.GetModelRatio(normName, 0)
		compRatio := ratio.GetCompletionRatio(normName, 0)

		if modelRatioVal <= 0 {
			// Try case-insensitive fuzzy match
			modelRatioVal = ratio.GetModelRatio(m.ModelName, 0)
			compRatio = ratio.GetCompletionRatio(m.ModelName, 0)
		}

		inputPrice := modelRatioVal * 0.002 // per 1K tokens in USD
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
		if err := model.DB.Create(rp).Error; err != nil {
			logger.SysWarn(fmt.Sprintf("SeedReferencePrices: create failed for %s: %v", m.ModelName, err))
		} else {
			inserted++
		}
	}

	logger.SysLog(fmt.Sprintf("SeedReferencePrices: %d inserted, %d skipped", inserted, skipped))
}

// FetchOfficialPricing attempts to fetch official pricing from provider pages.
// For MVP: reseed from ModelRatio (which already has official prices).
// Future: add per-provider HTTP scrapers.
func FetchOfficialPricing() time.Duration {
	start := time.Now()
	logger.SysLog("FetchOfficialPricing: starting monthly official pricing update...")

	// Re-seed from ModelRatio for any new models
	SeedReferencePrices()

	// TODO: Add per-provider scrapers:
	//   - OpenAI:   GET platform.openai.com/docs/models → parse HTML
	//   - Anthropic: GET anthropic.com/pricing → parse HTML
	//   - Google:    GET ai.google.dev/pricing → parse HTML
	//   - DeepSeek:  GET api-docs.deepseek.com → parse HTML

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
