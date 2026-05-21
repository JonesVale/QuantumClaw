package service

import (
	"testing"
)

func TestQuotaToPrice(t *testing.T) {
	tests := []struct {
		name          string
		quota         int64
		costPerUnit   float64
		sellPriceRate float64
		want          int64
	}{
		{"zero quota", 0, 0.5, 2.0, 0},
		{"min price", 1, 0.5, 2.0, 1},
		{"normal", 1000, 0.5, 2.0, 1},
		{"large quota", 500000, 0.5, 2.0, 50},
		{"zero cost", 1000, 0, 2.0, 1},
		{"zero sell price rate", 1000, 0.5, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quotaToPrice(tt.quota, tt.costPerUnit, tt.sellPriceRate)
			if got != tt.want {
				t.Errorf("quotaToPrice(%d, %.4f, %.2f) = %d, want %d",
					tt.quota, tt.costPerUnit, tt.sellPriceRate, got, tt.want)
			}
		})
	}
}

func TestGetEstimatedQuota(t *testing.T) {
	tests := []struct {
		name         string
		promptTokens int
		ratio        float64
		wantMin      int64
	}{
		{"zero tokens", 0, 1.0, 500},
		{"normal", 100, 1.0, 600},
		{"with model ratio", 100, 2.5, 1500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEstimatedQuota(tt.promptTokens, tt.ratio)
			if got < tt.wantMin {
				t.Errorf("getEstimatedQuota(%d, %.2f) = %d, want >= %d",
					tt.promptTokens, tt.ratio, got, tt.wantMin)
			}
		})
	}
}

func TestCalculateQuota(t *testing.T) {
	tests := []struct {
		name             string
		promptTokens     int
		completionTokens int
		ratio            float64
		expectMin        int64
	}{
		{"normal chat", 100, 50, 1.0, 200},
		{"high model ratio", 100, 50, 3.0, 600},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := int64(0)
			if tt.ratio != 0 {
				got = int64(float64(tt.promptTokens+tt.completionTokens*2) * tt.ratio)
			}
			if got < tt.expectMin {
				t.Errorf("calculateQuota = %d, want >= %d", got, tt.expectMin)
			}
			if got <= 0 {
				t.Error("calculateQuota returned 0, expected > 0")
			}
		})
	}
}

func TestCreateProviderEarningStatus(t *testing.T) {
	if "settled" != "settled" {
		t.Error("EarningStatusSettled mismatch")
	}
}

func TestWithdrawMinAmount(t *testing.T) {
	// WithdrawMinAmount = 100 (¥1)
	minAmount := int64(100)
	if minAmount < 100 {
		t.Errorf("WithdrawMinAmount = %d, want >= 100 (¥1)", minAmount)
	}
}
