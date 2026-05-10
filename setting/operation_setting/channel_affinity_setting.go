package operation_setting

import (
	"encoding/json"
	"sync"
)

// ChannelAffinityKeySource 亲和性键来源
type ChannelAffinityKeySource struct {
	Type string `json:"type"` // context_int / context_string / gjson
	Key  string `json:"key"`  // Gin Context key (for context_*)
	Path string `json:"path"` // gjson path (for gjson)
}

// ChannelAffinityRule 单条亲和性规则
type ChannelAffinityRule struct {
	Name string `json:"name"` // 规则名称（唯一标识）

	// 模型名正则匹配
	ModelRegex []string `json:"model_regex"`
	// 请求路径正则匹配
	PathRegex []string `json:"path_regex"`
	// User-Agent 包含关键字（不区分大小写）
	UserAgentInclude []string `json:"user_agent_include"`
	// 亲和性键来源（按顺序尝试，取第一个有值的）
	KeySources []ChannelAffinityKeySource `json:"key_sources"`
	// 亲和性值必须匹配的正则（空=不限制）
	ValueRegex string `json:"value_regex"`

	// 缓存TTL（秒），0=使用默认值
	TTLSeconds int `json:"ttl_seconds"`

	// 缓存命中后失败是否跳过重试（不尝试其他渠道）
	SkipRetryOnFailure bool `json:"skip_retry_on_failure"`

	// 以下字段控制缓存键中包含哪些维度
	IncludeRuleName   bool `json:"include_rule_name"`
	IncludeModelName  bool `json:"include_model_name"`
	IncludeUsingGroup bool `json:"include_using_group"`

	// 渠道参数覆盖模板（JSON对象），命中介入后合并到渠道参数
	ParamOverrideTemplate map[string]interface{} `json:"param_override_template"`
}

// ChannelAffinitySetting 渠道亲和性路由全局设置
type ChannelAffinitySetting struct {
	Enabled bool `json:"enabled"`

	// 缓存条目上限
	MaxEntries int `json:"max_entries"`
	// 默认TTL（秒）
	DefaultTTLSeconds int `json:"default_ttl_seconds"`

	// 路由成功后切换到新渠道并缓存
	SwitchOnSuccess bool `json:"switch_on_success"`

	// 规则列表
	Rules []ChannelAffinityRule `json:"rules"`
}

var (
	channelAffinitySetting   *ChannelAffinitySetting
	channelAffinitySettingMu sync.RWMutex
)

// GetChannelAffinitySetting 获取渠道亲和性设置（单例）
func GetChannelAffinitySetting() *ChannelAffinitySetting {
	channelAffinitySettingMu.RLock()
	defer channelAffinitySettingMu.RUnlock()
	if channelAffinitySetting == nil {
		channelAffinitySetting = &ChannelAffinitySetting{
			Enabled:          false,
			MaxEntries:       100_000,
			DefaultTTLSeconds: 3600,
			SwitchOnSuccess:  false,
			Rules:            []ChannelAffinityRule{},
		}
	}
	return channelAffinitySetting
}

// SetChannelAffinitySetting 更新渠道亲和性设置
func SetChannelAffinitySetting(s *ChannelAffinitySetting) {
	channelAffinitySettingMu.Lock()
	defer channelAffinitySettingMu.Unlock()
	if s == nil {
		s = &ChannelAffinitySetting{}
	}
	if s.MaxEntries <= 0 {
		s.MaxEntries = 100_000
	}
	if s.DefaultTTLSeconds <= 0 {
		s.DefaultTTLSeconds = 3600
	}
	channelAffinitySetting = s
}

// ParseChannelAffinitySetting 从 JSON 字符串解析设置
func ParseChannelAffinitySetting(data string) (*ChannelAffinitySetting, error) {
	var s ChannelAffinitySetting
	if data == "" {
		return &ChannelAffinitySetting{
			Enabled:          false,
			MaxEntries:       100_000,
			DefaultTTLSeconds: 3600,
			Rules:            []ChannelAffinityRule{},
		}, nil
	}
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}
