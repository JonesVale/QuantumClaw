package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEvaluateBillingExpr_Basic 计算：input_tokens * 0.03 + output_tokens * 0.06
// 标准 GPT-4o 费率（每 1M tokens 计算）
func TestEvaluateBillingExpr_Basic(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  1000,
		OutputTokens: 500,
	}

	expr := "input_tokens * 0.03 + output_tokens * 0.06"
	quota, err := EvaluateBillingExpr(ctx, expr)

	assert.NoError(t, err)
	// 1000*0.03 + 500*0.06 = 30 + 30 = 60 微额度
	assert.Equal(t, int64(60_000_000), quota, "1000 input + 500 output 应消耗 6000万微额度 (60美元 * 1M)")
}

// TestEvaluateBillingExpr_CacheHits 带缓存命中的计费：cache_hits 享受折扣
// Anthropic/Groq 等支持 cache_hits
func TestEvaluateBillingExpr_CacheHits(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  10000,
		OutputTokens: 2000,
		CacheHits:    6000, // 60% 命中缓存
		CacheMisses:  4000,
	}

	// cache_misses 全价，cache_hits 享受 1折
	expr := "(cache_hits * 0.001) + (cache_misses * 0.01)"
	quota, err := EvaluateBillingExpr(ctx, expr)

	assert.NoError(t, err)
	// 6000*0.001 + 4000*0.01 = 6 + 40 = 46
	assert.Equal(t, int64(46_000_000), quota, "带缓存命中的计费应享受折扣")
}

// TestEvaluateBillingExpr_OutputOnly 只有输出 token（纯补全场景）
func TestEvaluateBillingExpr_OutputOnly(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  0,
		OutputTokens: 1000,
	}

	expr := "output_tokens * 0.12"
	quota, err := EvaluateBillingExpr(ctx, expr)

	assert.NoError(t, err)
	// 1000 * 0.12 = 120
	assert.Equal(t, int64(120_000_000), quota, "纯输出应按 output_tokens 计价")
}

// TestEvaluateBillingExpr_TotalTokens 用 total_tokens 计算
func TestEvaluateBillingExpr_TotalTokens(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  3000,
		OutputTokens: 2000,
	}

	expr := "(input_tokens + output_tokens) * 0.05"
	quota, err := EvaluateBillingExpr(ctx, expr)

	assert.NoError(t, err)
	// (3000 + 2000) * 0.05 = 5000 * 0.05 = 250
	assert.Equal(t, int64(250_000_000), quota)
}

// TestEvaluateBillingExpr_EmptyExpr 空表达式返回 0
func TestEvaluateBillingExpr_EmptyExpr(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  1000,
		OutputTokens: 500,
	}

	quota, err := EvaluateBillingExpr(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), quota, "空表达式应返回 0")
}

// TestEvaluateBillingExpr_NilContext nil 上下文返回 0
func TestEvaluateBillingExpr_NilContext(t *testing.T) {
	quota, err := EvaluateBillingExpr(nil, "input_tokens * 0.01")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), quota, "nil 上下文应返回 0")
}

// TestEvaluateBillingExpr_InvalidExpr 非法表达式返回错误
func TestEvaluateBillingExpr_InvalidExpr(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  1000,
		OutputTokens: 500,
	}

	_, err := EvaluateBillingExpr(ctx, "input_tokens * * 0.01")
	assert.Error(t, err, "非法表达式应返回错误")
	assert.Contains(t, err.Error(), "billing expr error")
}

// TestEvaluateBillingExpr_UnknownVariable 未知变量返回错误
func TestEvaluateBillingExpr_UnknownVariable(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  1000,
		OutputTokens: 500,
	}

	_, err := EvaluateBillingExpr(ctx, "unknown_var * 0.01")
	assert.Error(t, err, "未知变量应返回错误")
}

// TestEvaluateBillingExpr_IntResult 整数表达式返回正确结果
func TestEvaluateBillingExpr_IntResult(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  100,
		OutputTokens: 50,
	}

	// 整数除法结果
	expr := "input_tokens / 10"
	quota, err := EvaluateBillingExpr(ctx, expr)

	assert.NoError(t, err)
	// 100 / 10 = 10 (int64)
	assert.Equal(t, int64(10), quota)
}

// TestEvaluateBillingExpr_ZeroTokens 零 token 场景
func TestEvaluateBillingExpr_ZeroTokens(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  0,
		OutputTokens: 0,
	}

	expr := "input_tokens * 0.03 + output_tokens * 0.06"
	quota, err := EvaluateBillingExpr(ctx, expr)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), quota, "零 token 应返回 0")
}

// TestEvaluateBillingExpr_LargeTokens 大 token 数（避免溢出）
func TestEvaluateBillingExpr_LargeTokens(t *testing.T) {
	ctx := &TieredBillingContext{
		InputTokens:  1_000_000,  // 1M
		OutputTokens: 500_000,   // 500K
	}

	expr := "input_tokens * 0.001"
	quota, err := EvaluateBillingExpr(ctx, expr)

	assert.NoError(t, err)
	// 1M * 0.001 = 1000 微额度（正确计算，无溢出）
	assert.Equal(t, int64(1_000_000_000), quota)
}

// TestEvaluateBillingExpr_CacheDiscountTier 缓存分层折扣表达式
func TestEvaluateBillingExpr_CacheDiscountTier(t *testing.T) {
	tests := []struct {
		name        string
		cacheHits   int
		cacheMisses int
		wantQuota   int64
	}{
		{"全缓存命中", 1000, 0, 1_000_000},     // 1000 * 0.001 * 1000000
		{"50%缓存命中", 500, 500, 3_000_000},  // (500*0.001 + 500*0.005)*1000000
		{"全缓存未命中", 0, 1000, 5_000_000},  // 1000 * 0.005 * 1000000
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &TieredBillingContext{
				InputTokens:  tt.cacheHits + tt.cacheMisses,
				OutputTokens: 0,
				CacheHits:    tt.cacheHits,
				CacheMisses:  tt.cacheMisses,
			}
			// cache_hits 0.1折，cache_misses 0.5折（每M token）
			expr := "(cache_hits * 0.001) + (cache_misses * 0.005)"
			quota, err := EvaluateBillingExpr(ctx, expr)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantQuota, quota)
		})
	}
}
