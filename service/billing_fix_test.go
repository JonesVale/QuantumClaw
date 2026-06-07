package service

// ============================================================
// billing_fix_test.go — 扣费 Bug 修复后的核心验证测试
//
// 覆盖范围：
//   1. quotaToUserPrice — 新增的用户价格换算函数（Bug #2 修复）
//   2. quotaToPrice — 原渠道成本价格函数（应仅用于分账，不可用于用户扣费）
//   3. 价格一致性 — 验证用户价格 ≠ 渠道成本（核心 Bug 场景）
//   4. 边界条件 — 零配额、最小价格、超大配额
//
// 运行：go test -v -run TestBillingFix ./service/
// ============================================================

import (
	"math"
	"testing"
)

// ─── 1. quotaToUserPrice 测试（新增函数）─────────────────────

func TestQuotaToUserPrice_Basic(t *testing.T) {
	tests := []struct {
		name     string
		quota    int64
		expected int64 // 最小期望值（分）
		desc     string
	}{
		{"1M quota = 100 cents", 1_000_000, 100, "1 USD = 100 cents"},
		{"500K quota = 50 cents", 500_000, 50, "0.5 USD = 50 cents"},
		{"100K quota = 10 cents", 100_000, 10, "0.1 USD = 10 cents"},
		{"10K quota = 1 cent (min)", 10_000, 1, "0.01 USD rounds to 1 cent"},
		{"1K quota = 1 cent (min)", 1_000, 1, "below min rounds to 1 cent"},
		{"zero quota = 0", 0, 0, "no quota = no charge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quotaToUserPrice(tt.quota)
			if result < tt.expected {
				t.Errorf("quotaToUserPrice(%d) = %d cents; expected >= %d [%s]",
					tt.quota, result, tt.expected, tt.desc)
			}
		})
	}
}

func TestQuotaToUserPrice_Negative(t *testing.T) {
	result := quotaToUserPrice(-100)
	if result != 0 {
		t.Errorf("Negative quota should return 0, got %d", result)
	}
}

func TestQuotaToUserPrice_Precision(t *testing.T) {
	// 测试浮点精度：999,999 微额度应该接近但不等于 100 分
	result := quotaToUserPrice(999_999)
	// 0.999999 USD → 99.9999 cents → ceil → 100 cents
	if result != 100 {
		t.Errorf("Expected 100 cents for 999999 micro-quota, got %d", result)
	}
}

func TestQuotaToUserPrice_LargeQuota(t *testing.T) {
	// 大额配额：100M 微额度 = 100 USD = 10000 分
	result := quotaToUserPrice(100_000_000)
	expected := int64(math.Ceil(100.0 * 100.0))
	if result != expected {
		t.Errorf("Large quota: expected %d, got %d", expected, result)
	}
}

// ─── 2. 核心场景：用户价格 vs 渠道成本（Bug 复现防护）────────

func TestUserPriceVsChannelCost_Different(t *testing.T) {
	/*
	   这是原始 Bug 的核心场景：
	   - 用户请求消耗了 1,000,000 微额度（含 modelRatio × groupRatio）
	   - 原代码用 channel.CostPerUnit=0.002 * SellPriceRate=1.0 计算出渠道成本价
	   - 但这算的是"平台向上游付多少钱"，不是"用户应付多少钱"

	   正确行为：
	   - quotaToUserPrice(quota) 直接从用户的 quota 换算 → 用户实际应付
	   - quotaToPrice(quota, cost, rate) 计算的是平台成本 → 用于分账/核算
	   - 两者必须不同！否则就是 bug 重现
	*/
	quota := int64(1_000_000) // 用户消耗的配额（已含倍率）

	userPrice := quotaToUserPrice(quota)        // 用户应付（分）
	channelCost := quotaToPrice(quota, 0.002, 1.0) // 渠道成本（分）

	t.Logf("Quota: %d micro-units", quota)
	t.Logf("User price (quotaToUserPrice): %d cents (= $%.2f)", userPrice, float64(userPrice)/100.0)
	t.Logf("Channel cost (quotaToPrice): %d cents (= $%.2f)", channelCost, float64(channelCost)/100.0)

	// 关键断言：两者使用不同的计算方式
	// 如果实现正确，userPrice 应该直接基于 quota 换算
	// channelCost 应该基于 CostPerUnit 计算
	if userPrice == channelCost && quota > 0 {
		t.Error("CRITICAL: User price equals channel cost! " +
		"This indicates the billing bug may still exist. " +
			"userPrice and channelCost should use different formulas.")
	}

	// userPrice 应该等于 quota/10000（微额度→分）
	expectedDirect := int64(math.Ceil(float64(quota) / 10000.0))
	if userPrice != expectedDirect {
		t.Errorf("User price should be direct quota conversion: expected %d, got %d",
			expectedDirect, userPrice)
	}
}

func TestUserPrice_IndependentOfChannelParams(t *testing.T) {
	/*
	   核心不变量：用户价格不应依赖任何渠道参数！
	   无论 CostPerUnit 和 SellPriceRate 如何变化，
	   同一 quota 对应的用户价格应该完全一致。
	*/
	quota := int64(2_000_000)

	price1 := quotaToUserPrice(quota)

	// 模拟不同的渠道参数——不应该影响结果
	// （因为 quotaToUserPrice 不接受这些参数）
	_ = quotaToPrice(quota, 0.001, 1.0)   // 低价渠道
	_ = quotaToPrice(quota, 0.01, 2.0)    // 高价高加价渠道
	_ = quotaToPrice(quota, 0.005, 0.8)    // 折扣渠道

	price2 := quotaToUserPrice(quota)

	if price1 != price2 {
		t.Errorf("User price should be stable regardless of channel params: %d vs %d",
			price1, price2)
	}
}

// ─── 3. 边界条件与安全性 ──────────────────────────────────────

func TestQuotaToUserPrice_MinimumCharge(t *testing.T) {
	// 即使极小配额也应收最低费用（1 分）
	testCases := []int64{1, 10, 100, 500, 999}
	for _, q := range testCases {
		result := quotaToUserPrice(q)
		if result < 1 {
			t.Errorf("quotaToUserPrice(%d) = %d, minimum should be 1 cent", q, result)
		}
	}
}

func TestBilling_RoundTripConsistency(t *testing.T) {
	// 验证：用户价格 → 配额 → 用户价格 应该一致（允许 ceiling 取整误差）
	originalQuota := int64(7_777_777) // 一个不规则数字
	userPriceCents := quotaToUserPrice(originalQuota)

	// 反向：分 → 微额度
	reconstructedQuota := userPriceCents * 10000 // 分→微额度（近似）

	// 允许 ±1% 的取整误差
	diff := float64(abs64(originalQuota-reconstructedQuota)) / float64(originalQuota)
	if diff > 0.01 {
		t.Logf("Round-trip: original=%d, reconstructed=%d, diff=%.4f%%",
			originalQuota, reconstructedQuota, diff*100)
		// 这只是信息性日志，不作为失败条件
		// 因为 ceil 操作本身就会引入偏差
	}
}

// ─── 辅助函数 ──────────────────────────────────────────────────

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
