package billing_setting

import (
	"encoding/json"
	"sync"
)

// TieredBillingRule 单条分层计费规则
type TieredBillingRule struct {
	// 规则名称（唯一标识）
	Name string `json:"name"`
	// 所属分组（匹配 token.go 中的 group_ratio）
	Group string `json:"group"`
	// 优先级（越大优先级越高）
	Priority int `json:"priority"`

	// 模型名正则匹配（空=全部匹配）
	ModelRegex []string `json:"model_regex"`

	// 关联渠道ID（0=所有渠道）
	ChannelId int `json:"channel_id"`

	// 额度上限（-1=不限）
	QuotaLimit int64 `json:"quota_limit"`
	// 单次最大消费（-1=不限）
	MaxConsumePerRequest int64 `json:"max_consume_per_request"`

	// 计费表达式（expr-lang 格式）
	// 可用变量: input_tokens, output_tokens, cache_hits, cache_misses
	// 示例: "input_tokens * 0.03 + output_tokens * 0.06"
	// 或分层:
	// "tier(input_tokens) * 0.03"
	BillingExpr string `json:"billing_expr"`

	// 是否启用
	Enabled bool `json:"enabled"`
}

// TieredBillingSetting 分层计费全局设置
type TieredBillingSetting struct {
	Enabled bool `json:"enabled"`

	// 默认计费表达式（无规则匹配时使用）
	DefaultExpr string `json:"default_expr"`

	// 默认额度上限
	DefaultQuotaLimit int64 `json:"default_quota_limit"`

	// 默认单次最大消费
	DefaultMaxConsume int64 `json:"default_max_consume"`

	// 规则列表（按 priority 倒序匹配）
	Rules []TieredBillingRule `json:"rules"`
}

var (
	tieredBillingSetting     *TieredBillingSetting
	tieredBillingSettingMu   sync.RWMutex
)

// GetTieredBillingSetting 获取分层计费设置（单例）
func GetTieredBillingSetting() *TieredBillingSetting {
	tieredBillingSettingMu.RLock()
	defer tieredBillingSettingMu.RUnlock()
	if tieredBillingSetting == nil {
		tieredBillingSetting = &TieredBillingSetting{
			Enabled:             false,
			DefaultExpr:         "",
			DefaultQuotaLimit:   -1,
			DefaultMaxConsume:   -1,
			Rules:               []TieredBillingRule{},
		}
	}
	return tieredBillingSetting
}

// SetTieredBillingSetting 更新分层计费设置
func SetTieredBillingSetting(s *TieredBillingSetting) {
	tieredBillingSettingMu.Lock()
	defer tieredBillingSettingMu.Unlock()
	if s == nil {
		s = &TieredBillingSetting{}
	}
	tieredBillingSetting = s
}

// ParseTieredBillingSetting 从 JSON 解析
func ParseTieredBillingSetting(data string) (*TieredBillingSetting, error) {
	var s TieredBillingSetting
	if data == "" {
		return &TieredBillingSetting{
			Enabled:           false,
			DefaultQuotaLimit: -1,
			DefaultMaxConsume: -1,
			Rules:             []TieredBillingRule{},
		}, nil
	}
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}
