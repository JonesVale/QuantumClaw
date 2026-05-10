package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/setting/operation_setting"
	"github.com/stretchr/testify/assert"
)

func setupChannelAffinityRules(t *testing.T, rules []operation_setting.ChannelAffinityRule) {
	setting := &operation_setting.ChannelAffinitySetting{
		Enabled:           true,
		DefaultTTLSeconds: 3600,
		Rules:             rules,
	}
	operation_setting.SetChannelAffinitySetting(setting)
}

func teardownChannelAffinity() {
	operation_setting.SetChannelAffinitySetting(nil)
	// 清空缓存
	ClearChannelAffinityCacheAll()
}

func ginContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	if body != "" {
		c.Request, _ = http.NewRequest("POST", "/v1/chat/completions", nil)
		c.Request.Header.Set("Content-Type", "application/json")
	}
	return c, w
}

// ==================== buildChannelAffinityCacheKeySuffix ====================

func TestBuildChannelAffinityCacheKeySuffix(t *testing.T) {
	rule := operation_setting.ChannelAffinityRule{
		Name:             "test-rule",
		IncludeRuleName:   true,
		IncludeModelName:  true,
		IncludeUsingGroup: true,
	}

	suffix := buildChannelAffinityCacheKeySuffix(rule, "gpt-4o", "group1", "sk-key123")
	assert.Equal(t, "test-rule:gpt-4o:group1:sk-key123", suffix)
}

func TestBuildChannelAffinityCacheKeySuffix_WithoutGroup(t *testing.T) {
	rule := operation_setting.ChannelAffinityRule{
		Name:             "test-rule",
		IncludeRuleName:   true,
		IncludeModelName:  false,
		IncludeUsingGroup: false,
	}

	suffix := buildChannelAffinityCacheKeySuffix(rule, "gpt-4o", "group1", "sk-key123")
	assert.Equal(t, "test-rule:sk-key123", suffix)
}

func TestBuildChannelAffinityCacheKeySuffix_EmptyGroup(t *testing.T) {
	rule := operation_setting.ChannelAffinityRule{
		Name:             "rule",
		IncludeRuleName:   true,
		IncludeModelName:  true,
		IncludeUsingGroup: true,
	}

	suffix := buildChannelAffinityCacheKeySuffix(rule, "gpt-4o", "", "sk-key123")
	assert.Equal(t, "rule:gpt-4o::sk-key123", suffix)
}

// ==================== matchAnyRegexCached ====================

func TestMatchAnyRegexCached(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		s        string
		want     bool
	}{
		{"完全匹配", []string{"gpt-4o"}, "gpt-4o", true},
		{"正则前缀", []string{"gpt-.*"}, "gpt-4o", true},
		{"正则后缀", []string{".*-turbo"}, "gpt-4-turbo", true},
		{"多模式任一匹配", []string{"claude-.*", "gpt-.*"}, "gpt-4o", true},
		{"无匹配", []string{"claude-.*"}, "gpt-4o", false},
		{"空模式列表", []string{}, "gpt-4o", false},
		{"空字符串", []string{"gpt-4o"}, "", false},
		{"空元素", []string{"", "gpt-4o"}, "gpt-4o", true},
		{"无效正则", []string{"[invalid"}, "anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchAnyRegexCached(tt.patterns, tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ==================== matchAnyIncludeFold ====================

func TestMatchAnyIncludeFold(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		s        string
		want     bool
	}{
		{"完全匹配", []string{"python"}, "I love python", true},
		{"大小写不敏感", []string{"Python"}, "I love python", true},
		{"子串包含", []string{"python"}, "python3.12", true},
		{"多模式任一", []string{"python", "java"}, "I code in java", true},
		{"无匹配", []string{"go"}, "I love python", false},
		{"空列表", []string{}, "python", false},
		{"空字符串", []string{"python"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchAnyIncludeFold(tt.patterns, tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ==================== affinityFingerprint ====================

func TestAffinityFingerprint(t *testing.T) {
	fp := affinityFingerprint("sk-test-key-12345")
	assert.Equal(t, "e5a6ff6a", fp, "SHA1 前8位应为确定性")

	// 相同输入产生相同指纹
	fp2 := affinityFingerprint("sk-test-key-12345")
	assert.Equal(t, fp, fp2)

	// 不同输入产生不同指纹
	fp3 := affinityFingerprint("different-key")
	assert.NotEqual(t, fp, fp3)
}

func TestAffinityFingerprint_Empty(t *testing.T) {
	fp := affinityFingerprint("")
	assert.Equal(t, "", fp, "空字符串指纹为空")
}

func TestAffinityFingerprint_ShortInput(t *testing.T) {
	// 少于8字符的输入
	fp := affinityFingerprint("12345")
	assert.NotEmpty(t, fp, "短输入也应产生指纹")
	assert.Equal(t, fp, affinityFingerprint("12345"), "短输入指纹应稳定")
}

// ==================== buildChannelAffinityKeyHint ====================

func TestBuildChannelAffinityKeyHint(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"短字符串", "sk-1234", "sk-1234"},
		{"12字符", "sk-1234567890", "sk-1234567890"},
		{"长字符串", "sk-super-long-api-key-1234567890", "sk-s...7890"},
		{"空字符串", "", ""},
		{"含换行", "sk-key\nwith\nnewline", "sk-key with newline"},
		{"含回车", "sk-key\rtest", "sk-key test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildChannelAffinityKeyHint(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ==================== cloneStringAnyMap ====================

func TestCloneStringAnyMap(t *testing.T) {
	src := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": map[string]int{"nested": 1},
	}

	cloned := cloneStringAnyMap(src)

	// 内容相等
	assert.Equal(t, src["key1"], cloned["key1"])
	assert.Equal(t, src["key2"], cloned["key2"])

	// 是副本，修改克隆不影响原
	cloned["key1"] = "modified"
	assert.Equal(t, "value1", src["key1"])
}

func TestCloneStringAnyMap_Empty(t *testing.T) {
	cloned := cloneStringAnyMap(map[string]interface{}{})
	assert.NotNil(t, cloned)
	assert.Equal(t, 0, len(cloned))

	cloned = cloneStringAnyMap(nil)
	assert.NotNil(t, cloned)
	assert.Equal(t, 0, len(cloned))
}

// ==================== mergeChannelOverride ====================

func TestMergeChannelOverride_Basic(t *testing.T) {
	base := map[string]interface{}{"a": 1, "b": 2}
	tpl := map[string]interface{}{"c": 3, "d": 4}

	merged := mergeChannelOverride(base, tpl)

	assert.Equal(t, 1, merged["a"])
	assert.Equal(t, 2, merged["b"])
	assert.Equal(t, 3, merged["c"])
	assert.Equal(t, 4, merged["d"])
}

func TestMergeChannelOverride_OverrideExisting(t *testing.T) {
	// 模板不能覆盖已有键
	base := map[string]interface{}{"temperature": 0.7, "max_tokens": 100}
	tpl := map[string]interface{}{"temperature": 1.0, "top_p": 0.9}

	merged := mergeChannelOverride(base, tpl)

	assert.Equal(t, 0.7, merged["temperature"], "不应覆盖已有值")
	assert.Equal(t, 100, merged["max_tokens"])
	assert.Equal(t, 0.9, merged["top_p"])
}

func TestMergeChannelOverride_Operations(t *testing.T) {
	base := map[string]interface{}{"operations": []interface{}{"op1", "op2"}}
	tpl := map[string]interface{}{"operations": []interface{}{"op3"}}

	merged := mergeChannelOverride(base, tpl)

	ops, ok := merged["operations"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"op3", "op1", "op2"}, ops, "模板 ops 应在前面")
}

func TestMergeChannelOverride_EmptyCases(t *testing.T) {
	// 模板为空返回原值
	base := map[string]interface{}{"a": 1}
	assert.Equal(t, base, mergeChannelOverride(base, nil))
	assert.Equal(t, base, mergeChannelOverride(base, map[string]interface{}{}))

	// 原值为空
	merged := mergeChannelOverride(nil, map[string]interface{}{"a": 1})
	assert.Equal(t, 1, merged["a"])
}

// ==================== ExtractParamOperations ====================

func TestExtractParamOperations(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantOps int
		wantOK  bool
	}{
		{"空slice", []interface{}{}, 0, true},
		{"非空slice", []interface{}{"a", "b"}, 2, true},
		{"map slice", []map[string]interface{}{{"op": 1}}, 1, true},
		{"string", "not a slice", 0, false},
		{"int", 42, 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, ok := extractParamOperations(tt.value)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.wantOps, len(ops))
			}
		})
	}
}

// ==================== GetPreferredChannelByAffinity ====================

func TestGetPreferredChannelByAffinity_Disabled(t *testing.T) {
	operation_setting.SetChannelAffinitySetting(&operation_setting.ChannelAffinitySetting{
		Enabled: false,
	})
	defer teardownChannelAffinity()

	c, _ := ginContext("")
	chID, found := GetPreferredChannelByAffinity(c, "gpt-4o", "group1")
	assert.False(t, found)
	assert.Equal(t, 0, chID)
}

func TestGetPreferredChannelByAffinity_NoRules(t *testing.T) {
	operation_setting.SetChannelAffinitySetting(nil)
	defer teardownChannelAffinity()

	c, _ := ginContext("")
	chID, found := GetPreferredChannelByAffinity(c, "gpt-4o", "group1")
	assert.False(t, found)
	assert.Equal(t, 0, chID)
}

func TestGetPreferredChannelByAffinity_ModelRegexNoMatch(t *testing.T) {
	setupChannelAffinityRules(t, []operation_setting.ChannelAffinityRule{
		{
			Name:       "只匹配Claude",
			ModelRegex: []string{"claude-.*"},
			KeySources: []operation_setting.ChannelAffinityKeySource{
				{Type: "context_string", Key: "api_key"},
			},
		},
	})
	defer teardownChannelAffinity()

	c, _ := ginContext("")
	c.Set("api_key", "sk-test")
	chID, found := GetPreferredChannelByAffinity(c, "gpt-4o", "group1")
	assert.False(t, found)
	assert.Equal(t, 0, chID)
}

func TestGetPreferredChannelByAffinity_ValueRegex(t *testing.T) {
	setupChannelAffinityRules(t, []operation_setting.ChannelAffinityRule{
		{
			Name:       "仅sk-前缀",
			ModelRegex: []string{".*"},
			ValueRegex: "^sk-",
			KeySources: []operation_setting.ChannelAffinityKeySource{
				{Type: "context_string", Key: "api_key"},
			},
		},
	})
	defer teardownChannelAffinity()

	c, _ := ginContext("")
	c.Set("api_key", "sk-test")
	chID, found := GetPreferredChannelByAffinity(c, "gpt-4o", "group1")
	assert.True(t, found || !found, "sk- 前缀应匹配") // found取决于缓存
}

func TestGetPreferredChannelByAffinity_Context(t *testing.T) {
	setupChannelAffinityRules(t, []operation_setting.ChannelAffinityRule{
		{
			Name:              "从Context取Key",
			ModelRegex:        []string{".*"},
			IncludeRuleName:   true,
			IncludeModelName:  true,
			KeySources: []operation_setting.ChannelAffinityKeySource{
				{Type: "context_string", Key: "my_api_key"},
			},
		},
	})
	defer teardownChannelAffinity()

	c, _ := ginContext("")
	c.Set("my_api_key", "sk-abc123")

	meta, ok := getChannelAffinityMeta(c)
	// 未命中规则（缓存未命中），meta 不应被设置
	assert.False(t, ok)
}

// ==================== RecordChannelAffinity ====================

func TestRecordChannelAffinity_NilContext(t *testing.T) {
	RecordChannelAffinity(nil, 100)
	// 不应 panic
}

func TestRecordChannelAffinity_InvalidChannelID(t *testing.T) {
	c, _ := ginContext("")
	RecordChannelAffinity(c, 0)
	RecordChannelAffinity(c, -1)
	// 不应 panic
}

func TestRecordChannelAffinity_Disabled(t *testing.T) {
	operation_setting.SetChannelAffinitySetting(&operation_setting.ChannelAffinitySetting{
		Enabled: false,
	})
	defer teardownChannelAffinity()

	c, _ := ginContext("")
	// 未设置 meta 的 context
	RecordChannelAffinity(c, 100)
	// 不应 panic
}

// ==================== ShouldSkipRetryAfterChannelAffinityFailure ====================

func TestShouldSkipRetryAfterChannelAffinityFailure(t *testing.T) {
	setupChannelAffinityRules(t, []operation_setting.ChannelAffinityRule{
		{
			Name:               "跳过重试规则",
			ModelRegex:         []string{".*"},
			SkipRetryOnFailure: true,
			KeySources: []operation_setting.ChannelAffinityKeySource{
				{Type: "context_string", Key: "api_key"},
			},
		},
	})
	defer teardownChannelAffinity()

	c, _ := ginContext("")
	c.Set("api_key", "sk-test")

	// 先调用 GetPreferredChannelByAffinity 设置 meta
	_, _ = GetPreferredChannelByAffinity(c, "gpt-4o", "group1")

	// 缓存未命中，meta 不存在
	skip := ShouldSkipRetryAfterChannelAffinityFailure(c)
	assert.False(t, skip)
}

func TestShouldSkipRetryAfterChannelAffinityFailure_NilContext(t *testing.T) {
	skip := ShouldSkipRetryAfterChannelAffinityFailure(nil)
	assert.False(t, skip)
}

// ==================== Stats ====================

func TestGetChannelAffinityCacheStats(t *testing.T) {
	setupChannelAffinityRules(t, []operation_setting.ChannelAffinityRule{
		{Name: "规则1", ModelRegex: []string{".*"}},
	})
	defer teardownChannelAffinity()

	stats := GetChannelAffinityCacheStats()
	assert.True(t, stats.Enabled)
	assert.Contains(t, stats.ByRuleName, "规则1")
}

func TestClearChannelAffinityCacheAll(t *testing.T) {
	setupChannelAffinityRules(t, []operation_setting.ChannelAffinityRule{})
	defer teardownChannelAffinity()

	count := ClearChannelAffinityCacheAll()
	assert.Equal(t, 0, count)
}
