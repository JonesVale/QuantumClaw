package service

import (
	"testing"

	"github.com/quantumclaw/quantumclaw/setting/billing_setting"
	"github.com/stretchr/testify/assert"
)

func TestGetMatchingTieredBillingRule_Priority(t *testing.T) {
	// 验证：同优先级时选择第一个匹配（优先级高者优先）
	setting := &billing_setting.TieredBillingSetting{
		Enabled: true,
		Rules: []billing_setting.TieredBillingRule{
			{Name: "低优先级规则", Group: "group1", Priority: 1, Enabled: true, ModelRegex: []string{".*"}},
			{Name: "中优先级规则", Group: "group1", Priority: 2, Enabled: true, ModelRegex: []string{".*"}},
			{Name: "高优先级规则", Group: "group1", Priority: 3, Enabled: true, ModelRegex: []string{".*"}},
		},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{
		Group:     "group1",
		ModelName: "gpt-4o",
	}

	rule := GetMatchingTieredBillingRule(ctx)
	assert.NotNil(t, rule, "应有匹配的规则")
	assert.Equal(t, "高优先级规则", rule.Name, "高优先级规则应被优先匹配")
	assert.Equal(t, 3, rule.Priority, "优先级应为 3")
}

func TestGetMatchingTieredBillingRule_GroupMatch(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{
		Enabled: true,
		Rules: []billing_setting.TieredBillingRule{
			{Name: "VIP组专用", Group: "vip", Priority: 10, Enabled: true, ModelRegex: []string{".*"}},
			{Name: "通用规则", Group: "", Priority: 1, Enabled: true, ModelRegex: []string{".*"}},
		},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	tests := []struct {
		group     string
		wantRule  string
	}{
		{"vip", "VIP组专用"},
		{"free", "通用规则"},
		{"", "通用规则"},
	}

	for _, tt := range tests {
		t.Run(tt.group, func(t *testing.T) {
			ctx := &TieredBillingContext{Group: tt.group, ModelName: "gpt-4o"}
			rule := GetMatchingTieredBillingRule(ctx)
			assert.NotNil(t, rule)
			assert.Equal(t, tt.wantRule, rule.Name)
		})
	}
}

func TestGetMatchingTieredBillingRule_ModelRegex(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{
		Enabled: true,
		Rules: []billing_setting.TieredBillingRule{
			{Name: "GPT-4系列", ModelRegex: []string{"gpt-4.*"}, Priority: 10, Enabled: true},
			{Name: "Claude系列", ModelRegex: []string{"claude-.*"}, Priority: 10, Enabled: true},
			{Name: "其他模型", ModelRegex: []string{}, Priority: 1, Enabled: true},
		},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	tests := []struct {
		model   string
		wantRule string
	}{
		{"gpt-4o", "GPT-4系列"},
		{"gpt-4-turbo", "GPT-4系列"},
		{"claude-3-opus", "Claude系列"},
		{"claude-3-sonnet", "Claude系列"},
		{"deepseek-chat", "其他模型"},
		{"unknown-model", "其他模型"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			ctx := &TieredBillingContext{ModelName: tt.model}
			rule := GetMatchingTieredBillingRule(ctx)
			assert.NotNil(t, rule)
			assert.Equal(t, tt.wantRule, rule.Name)
		})
	}
}

func TestGetMatchingTieredBillingRule_ChannelId(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{
		Enabled: true,
		Rules: []billing_setting.TieredBillingRule{
			{Name: "渠道101专属", ChannelId: 101, Priority: 10, Enabled: true},
			{Name: "渠道202专属", ChannelId: 202, Priority: 10, Enabled: true},
			{Name: "通用渠道", ChannelId: 0, Priority: 1, Enabled: true},
		},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	tests := []struct {
		channelId int
		wantRule  string
	}{
		{101, "渠道101专属"},
		{202, "渠道202专属"},
		{999, "通用渠道"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			ctx := &TieredBillingContext{ChannelId: tt.channelId}
			rule := GetMatchingTieredBillingRule(ctx)
			assert.NotNil(t, rule)
			assert.Equal(t, tt.wantRule, rule.Name)
		})
	}
}

func TestGetMatchingTieredBillingRule_DisabledRule(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{
		Enabled: true,
		Rules: []billing_setting.TieredBillingRule{
			{Name: "已禁用规则", Priority: 100, Enabled: false, ModelRegex: []string{".*"}},
			{Name: "默认规则", Priority: 1, Enabled: true, ModelRegex: []string{".*"}},
		},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{ModelName: "gpt-4o"}
	rule := GetMatchingTieredBillingRule(ctx)

	assert.NotNil(t, rule)
	assert.Equal(t, "默认规则", rule.Name, "已禁用的规则不应被匹配")
}

func TestGetMatchingTieredBillingRule_NilContext(t *testing.T) {
	billing_setting.SetTieredBillingSetting(nil)
	defer billing_setting.SetTieredBillingSetting(nil)

	rule := GetMatchingTieredBillingRule(nil)
	assert.Nil(t, rule, "nil 上下文应返回 nil")
}

func TestGetMatchingTieredBillingRule_NoRules(t *testing.T) {
	billing_setting.SetTieredBillingSetting(nil)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{ModelName: "gpt-4o"}
	rule := GetMatchingTieredBillingRule(ctx)
	assert.Nil(t, rule)
}

func TestCalculateTieredQuota_UseRuleBillingExpr(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{
		Enabled: true,
		Rules: []billing_setting.TieredBillingRule{
			{
				Name:        "自定义费率",
				Priority:    10,
				Enabled:     true,
				BillingExpr: "input_tokens * 0.05 + output_tokens * 0.1",
			},
		},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{
		InputTokens:  1000,
		OutputTokens: 500,
	}

	quota, err := CalculateTieredQuota(ctx)

	assert.NoError(t, err)
	// 1000*0.05 + 500*0.1 = 50 + 50 = 100
	assert.Equal(t, int64(100_000_000), quota)
}

func TestCalculateTieredQuota_UseDefaultExpr(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{
		Enabled:     true,
		DefaultExpr: "input_tokens * 0.01",
		Rules:       []billing_setting.TieredBillingRule{},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{
		InputTokens:  1000,
		OutputTokens: 0,
	}

	quota, err := CalculateTieredQuota(ctx)

	assert.NoError(t, err)
	// 1000 * 0.01 = 10
	assert.Equal(t, int64(10_000_000), quota)
}

func TestCalculateTieredQuota_Disabled(t *testing.T) {
	billing_setting.SetTieredBillingSetting(nil)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{
		InputTokens:  1000,
		OutputTokens: 500,
	}

	quota, err := CalculateTieredQuota(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), quota, "禁用分层计费时应返回 0")
}

func TestCalculateTieredQuota_NilContext(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{Enabled: true}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	_, err := CalculateTieredQuota(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "billing context is nil")
}

func TestGetQuotaLimitForRequest_RuleQuotaLimit(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{
		Enabled:           true,
		DefaultQuotaLimit: 100_000_000,
		Rules: []billing_setting.TieredBillingRule{
			{
				Name:         "高价值模型",
				Priority:     10,
				Enabled:      true,
				QuotaLimit:   500_000_000, // 更高限额
				MaxConsumePerRequest: 200_000_000,
			},
		},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{ModelName: "gpt-4o"}
	limit := GetQuotaLimitForRequest(ctx)
	assert.Equal(t, int64(500_000_000), limit, "应使用规则特定限额")
}

func TestGetQuotaLimitForRequest_Unlimited(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{
		Enabled:           true,
		DefaultQuotaLimit: 100_000_000,
		Rules: []billing_setting.TieredBillingRule{
			{
				Name:        "无限额",
				Priority:    10,
				Enabled:     true,
				QuotaLimit:  -1, // 无限制
			},
		},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{ModelName: "gpt-4o"}
	limit := GetQuotaLimitForRequest(ctx)
	assert.Equal(t, int64(100_000_000), limit, "规则无限额时使用默认值")
}

func TestGetMaxConsumePerRequest(t *testing.T) {
	setting := &billing_setting.TieredBillingSetting{
		Enabled:           true,
		DefaultMaxConsume: 50_000_000,
		Rules: []billing_setting.TieredBillingRule{
			{
				Name:                "高价值模型",
				Priority:            10,
				Enabled:             true,
				MaxConsumePerRequest: 200_000_000,
			},
		},
	}
	billing_setting.SetTieredBillingSetting(setting)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{ModelName: "gpt-4o"}
	max := GetMaxConsumePerRequest(ctx)
	assert.Equal(t, int64(200_000_000), max, "应使用规则特定单次限额")
}

func TestGetMaxConsumePerRequest_Disabled(t *testing.T) {
	billing_setting.SetTieredBillingSetting(nil)
	defer billing_setting.SetTieredBillingSetting(nil)

	ctx := &TieredBillingContext{}
	max := GetMaxConsumePerRequest(ctx)
	assert.Equal(t, int64(-1), max, "禁用时应返回 -1")
}
