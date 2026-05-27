package service

import (
	"testing"
)

// Test getEstimatedQuota - unexported but same package
func TestGetEstimatedQuota(t *testing.T) {
	tests := []struct {
		promptTokens int
		ratio        float64
		expectedMin  int64
		expectedMax  int64
	}{
		{0, 1.0, 500, 500},  // minimum quota is 500
		{100, 2.0, 1, 10000},
		{1000, 0.5, 1, 10000},
		{100000, 1.0, 1, 500000},
	}

	for _, tt := range tests {
		result := getEstimatedQuota(tt.promptTokens, tt.ratio)
		if result < tt.expectedMin || result > tt.expectedMax {
			t.Errorf("getEstimatedQuota(%d, %f) = %d; expected [%d, %d]",
				tt.promptTokens, tt.ratio, result, tt.expectedMin, tt.expectedMax)
		}
	}
}

func TestGetEstimatedQuota_ZeroRatio(t *testing.T) {
	result := getEstimatedQuota(1000, 0)
	if result < 0 {
		t.Errorf("Expected non-negative for zero ratio, got %d", result)
	}
}

func TestQuotaToPrice(t *testing.T) {
	tests := []struct {
		quota       int64
		costPerUnit float64
		priceRate   float64
		desc        string
	}{
		{100000, 0.002, 1.0, "standard pricing"},
		{0, 0.002, 1.0, "zero quota"},
		{100000, 0, 1.0, "zero cost"},
		{100000, 0.002, 1.5, "with markup"},
	}

	for _, tt := range tests {
		result := quotaToPrice(tt.quota, tt.costPerUnit, tt.priceRate)
		if result < 0 {
			t.Errorf("quotaToPrice(%d, %f, %f) [%s] = %d; expected >= 0",
				tt.quota, tt.costPerUnit, tt.priceRate, tt.desc, result)
		}
	}
}

func TestQuotaToPrice_Markup(t *testing.T) {
	// With price rate 2.0, price should be roughly 2x
	base := quotaToPrice(100000, 0.002, 1.0)
	marked := quotaToPrice(100000, 0.002, 2.0)
	if marked < base {
		t.Errorf("Marked price (%d) should be >= base price (%d)", marked, base)
	}
}

func TestPreConsumeBalance(t *testing.T) {
	// PreConsumeBalance requires a full meta.Context - skip for unit test
	// This integration test needs a running DB
	t.Skip("Skipping integration test requiring DB")
}

func TestPostConsumeDeduct(t *testing.T) {
	t.Skip("Skipping integration test requiring DB")
}
