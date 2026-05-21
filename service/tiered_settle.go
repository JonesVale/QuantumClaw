package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/quantumclaw/quantumclaw/setting/billing_setting"
)

// ==================== 额度估算上下文 ====================

type TieredBillingContext struct {
	Group           string
	ModelName       string
	ChannelId       int
	InputTokens     int
	OutputTokens    int
	CacheHits       int // 缓存命中 token 数
	CacheMisses     int // 缓存未命中 token 数

	// 量子算力计费变量
	Qubits int `json:"qubits"`
	Shots  int `json:"shots"`
	Gates  int `json:"gates"`
}

// ==================== 分层计费解析 ====================

// EvaluateBillingExpr 使用 expr-lang 表达式计算费用
// 已在 go.mod 添加 github.com/expr-lang/expr 依赖
func EvaluateBillingExpr(ctx *TieredBillingContext, expr string) (int64, error) {
	if ctx == nil || strings.TrimSpace(expr) == "" {
		return 0, nil
	}
	// 构建表达式环境变量
	env := map[string]interface{}{
		"input_tokens":  int64(ctx.InputTokens),
		"output_tokens": int64(ctx.OutputTokens),
		"cache_hits":    int64(ctx.CacheHits),
		"cache_misses":  int64(ctx.CacheMisses),
		// 量子算力
		"qubits": int64(ctx.Qubits),
		"shots":  int64(ctx.Shots),
		"gates":  int64(ctx.Gates),
	}
	output, err := EvaluateExpr(expr, env)
	if err != nil {
		return 0, fmt.Errorf("billing expr error: %w", err)
	}
	if val, ok := output.(float64); ok {
		return int64(val * 1000000), nil // 转换为微额度
	}
	if val, ok := output.(int64); ok {
		return val, nil
	}
	if val, ok := output.(int); ok {
		return int64(val), nil
	}
	return 0, fmt.Errorf("billing expr returned non-numeric type: %T", output)
}

// EvaluateExpr 执行 expr-lang 表达式（延迟导入以避免循环依赖）
// 实际实现在 billing_expr.go 中
var EvaluateExpr func(expr string, env map[string]interface{}) (interface{}, error)

// ==================== 分层规则匹配 ====================

func matchAnyRegex(patterns []string, s string) bool {
	if len(patterns) == 0 || s == "" {
		return false
	}
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		matched, err := regexp.MatchString(pattern, s)
		if err != nil || !matched {
			continue
		}
		return true
	}
	return false
}

// GetMatchingTieredBillingRule 根据上下文找到匹配的分层计费规则
func GetMatchingTieredBillingRule(ctx *TieredBillingContext) *billing_setting.TieredBillingRule {
	if ctx == nil {
		return nil
	}
	setting := billing_setting.GetTieredBillingSetting()
	if !setting.Enabled || len(setting.Rules) == 0 {
		return nil
	}

	// 按 priority 倒序匹配（高优先级优先）
	maxPriority := -1
	var matched *billing_setting.TieredBillingRule

	for i := range setting.Rules {
		rule := &setting.Rules[i]
		if !rule.Enabled {
			continue
		}
		// 检查分组
		if rule.Group != "" && rule.Group != ctx.Group {
			continue
		}
		// 检查渠道
		if rule.ChannelId > 0 && rule.ChannelId != ctx.ChannelId {
			continue
		}
		// 检查模型正则
		if len(rule.ModelRegex) > 0 && !matchAnyRegex(rule.ModelRegex, ctx.ModelName) {
			continue
		}
		// 优先选择更高优先级的
		if rule.Priority > maxPriority {
			maxPriority = rule.Priority
			matched = rule
		}
	}
	return matched
}

// CalculateTieredQuota 计算分层计费后的消耗额度
func CalculateTieredQuota(ctx *TieredBillingContext) (int64, error) {
	if ctx == nil {
		return 0, fmt.Errorf("billing context is nil")
	}
	setting := billing_setting.GetTieredBillingSetting()
	if !setting.Enabled {
		return 0, nil
	}

	rule := GetMatchingTieredBillingRule(ctx)
	if rule == nil || rule.BillingExpr == "" {
		// 使用默认表达式
		if setting.DefaultExpr == "" {
			// 量子算力默认计费：每 qubit × shots 消耗 1 额度，最少 10
			if ctx.Qubits > 0 || ctx.Shots > 0 {
				qubits := ctx.Qubits
				if qubits <= 0 {
					qubits = 1
				}
				shots := ctx.Shots
				if shots <= 0 {
					shots = 1000
				}
				gates := ctx.Gates
				base := qubits * shots / 100
				if gates > 0 {
					base += gates * 10
				}
				if base < 10 {
					base = 10
				}
				return int64(base), nil
			}
			return 0, nil
		}
		return EvaluateBillingExpr(ctx, setting.DefaultExpr)
	}
	return EvaluateBillingExpr(ctx, rule.BillingExpr)
}

// GetQuotaLimitForRequest 获取单次请求的额度上限
func GetQuotaLimitForRequest(ctx *TieredBillingContext) int64 {
	if ctx == nil {
		return -1
	}
	setting := billing_setting.GetTieredBillingSetting()
	if !setting.Enabled {
		return setting.DefaultQuotaLimit
	}
	rule := GetMatchingTieredBillingRule(ctx)
	if rule == nil {
		return setting.DefaultQuotaLimit
	}
	if rule.QuotaLimit < 0 {
		return setting.DefaultQuotaLimit
	}
	return rule.QuotaLimit
}

// GetMaxConsumePerRequest 获取单次最大消费
func GetMaxConsumePerRequest(ctx *TieredBillingContext) int64 {
	if ctx == nil {
		return -1
	}
	setting := billing_setting.GetTieredBillingSetting()
	if !setting.Enabled {
		return setting.DefaultMaxConsume
	}
	rule := GetMatchingTieredBillingRule(ctx)
	if rule == nil {
		return setting.DefaultMaxConsume
	}
	if rule.MaxConsumePerRequest < 0 {
		return setting.DefaultMaxConsume
	}
	return rule.MaxConsumePerRequest
}
