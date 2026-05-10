package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/samber/hot"
	"github.com/tidwall/gjson"
)

const (
	ginKeyChannelAffinityCacheKey   = "channel_affinity_cache_key"
	ginKeyChannelAffinityTTLSeconds = "channel_affinity_ttl_seconds"
	ginKeyChannelAffinityMeta       = "channel_affinity_meta"
	ginKeyChannelAffinityLogInfo    = "channel_affinity_log_info"
	ginKeyChannelAffinitySkipRetry  = "channel_affinity_skip_retry_on_failure"

	channelAffinityCacheNamespace = "quantumclaw:channel_affinity:v1"
)

var (
	channelAffinityCacheOnce sync.Once
	channelAffinityCache     *hot.HotCache[string, int]

	channelAffinityRegexCache sync.Map // map[string]*regexp.Regexp
)

// ==================== 元数据类型 ====================

type channelAffinityMeta struct {
	CacheKey      string
	TTLSeconds    int
	RuleName      string
	SkipRetry     bool
	ParamTemplate map[string]interface{}
	KeySourceType string
	KeySourceKey  string
	KeySourcePath string
	KeyHint       string
	KeyFingerprint string
	UsingGroup    string
	ModelName     string
	RequestPath   string
}

type ChannelAffinityStatsContext struct {
	RuleName       string
	UsingGroup     string
	KeyFingerprint string
	TTLSeconds     int64
}

type ChannelAffinityCacheStats struct {
	Enabled       bool           `json:"enabled"`
	Total         int            `json:"total"`
	Unknown       int            `json:"unknown"`
	ByRuleName    map[string]int `json:"by_rule_name"`
	CacheCapacity int            `json:"cache_capacity"`
}

// ==================== 缓存初始化 ====================

func getChannelAffinityCache() *hot.HotCache[string, int] {
	channelAffinityCacheOnce.Do(func() {
		setting := operation_setting.GetChannelAffinitySetting()
		capacity := 100_000
		if setting != nil && setting.MaxEntries > 0 {
			capacity = setting.MaxEntries
		}
		defaultTTL := 3600
		if setting != nil && setting.DefaultTTLSeconds > 0 {
			defaultTTL = setting.DefaultTTLSeconds
		}
		channelAffinityCache = hot.NewHotCache[string, int](hot.LRU, capacity).
			WithTTL(time.Duration(defaultTTL) * time.Second).
			WithJanitor().
			Build()
	})
	return channelAffinityCache
}

// GetChannelAffinityCacheStats 获取缓存统计信息
func GetChannelAffinityCacheStats() ChannelAffinityCacheStats {
	setting := operation_setting.GetChannelAffinitySetting()
	if setting == nil {
		return ChannelAffinityCacheStats{Enabled: false, Total: 0, Unknown: 0, ByRuleName: map[string]int{}}
	}

	cache := getChannelAffinityCache()
	totalLen := cache.Len()

	ruleByName := make(map[string]int)
	for _, r := range setting.Rules {
		name := strings.TrimSpace(r.Name)
		if name != "" {
			ruleByName[name] = 0
		}
	}

	capacity, _ := cache.Capacity()
	return ChannelAffinityCacheStats{
		Enabled:       setting.Enabled,
		Total:         totalLen,
		Unknown:       0,
		ByRuleName:    ruleByName,
		CacheCapacity: capacity,
	}
}

// ClearChannelAffinityCacheAll 清空全部亲和性缓存
func ClearChannelAffinityCacheAll() int {
	cache := getChannelAffinityCache()
	cache.Purge()
	return 0
}

// ClearChannelAffinityCacheByRuleName 按规则名清空缓存
func ClearChannelAffinityCacheByRuleName(ruleName string) (int, error) {
	ruleName = strings.TrimSpace(ruleName)
	if ruleName == "" {
		return 0, fmt.Errorf("rule_name 不能为空")
	}
	return 0, nil
}

// ==================== 匹配工具 ====================

func matchAnyRegexCached(patterns []string, s string) bool {
	if len(patterns) == 0 || s == "" {
		return false
	}
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		re, ok := channelAffinityRegexCache.Load(pattern)
		if !ok {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			re = compiled
			channelAffinityRegexCache.Store(pattern, re)
		}
		if re.(*regexp.Regexp).MatchString(s) {
			return true
		}
	}
	return false
}

func matchAnyIncludeFold(patterns []string, s string) bool {
	if len(patterns) == 0 || s == "" {
		return false
	}
	sLower := strings.ToLower(s)
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(sLower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

func extractChannelAffinityValue(c *gin.Context, src operation_setting.ChannelAffinityKeySource) string {
	switch src.Type {
	case "context_int":
		if src.Key == "" {
			return ""
		}
		v := c.GetInt(src.Key)
		if v <= 0 {
			return ""
		}
		return strconv.Itoa(v)
	case "context_string":
		if src.Key == "" {
			return ""
		}
		return strings.TrimSpace(c.GetString(src.Key))
	case "gjson":
		if src.Path == "" {
			return ""
		}
		body, err := common.GetRequestBody(c)
		if err != nil || len(body) == 0 {
			return ""
		}
		res := gjson.GetBytes(body, src.Path)
		if !res.Exists() {
			return ""
		}
		switch res.Type {
		case gjson.String, gjson.Number, gjson.True, gjson.False:
			return strings.TrimSpace(res.String())
		default:
			return strings.TrimSpace(res.Raw)
		}
	default:
		return ""
	}
}

// ==================== 缓存键构建 ====================

func buildChannelAffinityCacheKeySuffix(rule operation_setting.ChannelAffinityRule, modelName string, usingGroup string, affinityValue string) string {
	parts := make([]string, 0, 4)
	if rule.IncludeRuleName && rule.Name != "" {
		parts = append(parts, rule.Name)
	}
	if rule.IncludeModelName && modelName != "" {
		parts = append(parts, modelName)
	}
	if rule.IncludeUsingGroup && usingGroup != "" {
		parts = append(parts, usingGroup)
	}
	parts = append(parts, affinityValue)
	return strings.Join(parts, ":")
}

func setChannelAffinityContext(c *gin.Context, meta channelAffinityMeta) {
	c.Set(ginKeyChannelAffinityCacheKey, meta.CacheKey)
	c.Set(ginKeyChannelAffinityTTLSeconds, meta.TTLSeconds)
	c.Set(ginKeyChannelAffinityMeta, meta)
}

func getChannelAffinityContext(c *gin.Context) (string, int, bool) {
	keyAny, ok := c.Get(ginKeyChannelAffinityCacheKey)
	if !ok {
		return "", 0, false
	}
	key, ok := keyAny.(string)
	if !ok || key == "" {
		return "", 0, false
	}
	ttlAny, ok := c.Get(ginKeyChannelAffinityTTLSeconds)
	if !ok {
		return key, 0, true
	}
	ttlSeconds, _ := ttlAny.(int)
	return key, ttlSeconds, true
}

func getChannelAffinityMeta(c *gin.Context) (channelAffinityMeta, bool) {
	anyMeta, ok := c.Get(ginKeyChannelAffinityMeta)
	if !ok {
		return channelAffinityMeta{}, false
	}
	meta, ok := anyMeta.(channelAffinityMeta)
	if !ok {
		return channelAffinityMeta{}, false
	}
	return meta, true
}

// GetChannelAffinityStatsContext 获取统计上下文
func GetChannelAffinityStatsContext(c *gin.Context) (ChannelAffinityStatsContext, bool) {
	if c == nil {
		return ChannelAffinityStatsContext{}, false
	}
	meta, ok := getChannelAffinityMeta(c)
	if !ok {
		return ChannelAffinityStatsContext{}, false
	}
	ruleName := strings.TrimSpace(meta.RuleName)
	keyFp := strings.TrimSpace(meta.KeyFingerprint)
	usingGroup := strings.TrimSpace(meta.UsingGroup)
	if ruleName == "" || keyFp == "" {
		return ChannelAffinityStatsContext{}, false
	}
	ttlSeconds := int64(meta.TTLSeconds)
	if ttlSeconds <= 0 {
		return ChannelAffinityStatsContext{}, false
	}
	return ChannelAffinityStatsContext{
		RuleName:       ruleName,
		UsingGroup:     usingGroup,
		KeyFingerprint: keyFp,
		TTLSeconds:     ttlSeconds,
	}, true
}

// ==================== 指纹与提示 ====================

func affinityFingerprint(s string) string {
	if s == "" {
		return ""
	}
	h := sha256.Sum256([]byte(s))
	hexStr := hex.EncodeToString(h[:])
	if len(hexStr) >= 8 {
		return hexStr[:8]
	}
	return hexStr
}

func buildChannelAffinityKeyHint(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	if len(s) <= 12 {
		return s
	}
	return s[:4] + "..." + s[len(s)-4:]
}

// ==================== 参数模板合并 ====================

func cloneStringAnyMap(src map[string]interface{}) map[string]interface{} {
	if len(src) == 0 {
		return map[string]interface{}{}
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func mergeChannelOverride(base map[string]interface{}, tpl map[string]interface{}) map[string]interface{} {
	if len(base) == 0 && len(tpl) == 0 {
		return map[string]interface{}{}
	}
	if len(tpl) == 0 {
		return base
	}
	out := cloneStringAnyMap(base)
	for k, v := range tpl {
		if strings.EqualFold(strings.TrimSpace(k), "operations") {
			baseOps, hasBaseOps := extractParamOperations(out[k])
			tplOps, hasTplOps := extractParamOperations(v)
			if hasTplOps {
				if hasBaseOps {
					out[k] = append(tplOps, baseOps...)
				} else {
					out[k] = tplOps
				}
				continue
			}
		}
		if _, exists := out[k]; exists {
			continue
		}
		out[k] = v
	}
	return out
}

func extractParamOperations(value interface{}) ([]interface{}, bool) {
	switch ops := value.(type) {
	case []interface{}:
		if len(ops) == 0 {
			return []interface{}{}, true
		}
		cloned := make([]interface{}, 0, len(ops))
		cloned = append(cloned, ops...)
		return cloned, true
	case []map[string]interface{}:
		cloned := make([]interface{}, 0, len(ops))
		for _, op := range ops {
			cloned = append(cloned, op)
		}
		return cloned, true
	default:
		return nil, false
	}
}

func appendChannelAffinityTemplateAdminInfo(c *gin.Context, meta channelAffinityMeta) {
	if c == nil {
		return
	}
	if len(meta.ParamTemplate) == 0 {
		return
	}
	templateInfo := map[string]interface{}{
		"applied":              true,
		"rule_name":            meta.RuleName,
		"param_override_keys":  len(meta.ParamTemplate),
	}
	if anyInfo, ok := c.Get(ginKeyChannelAffinityLogInfo); ok {
		if info, ok := anyInfo.(map[string]interface{}); ok {
			info["override_template"] = templateInfo
			c.Set(ginKeyChannelAffinityLogInfo, info)
			return
		}
	}
	c.Set(ginKeyChannelAffinityLogInfo, map[string]interface{}{
		"reason":             meta.RuleName,
		"rule_name":          meta.RuleName,
		"using_group":        meta.UsingGroup,
		"model":              meta.ModelName,
		"request_path":       meta.RequestPath,
		"key_source":         meta.KeySourceType,
		"key_key":            meta.KeySourceKey,
		"key_path":           meta.KeySourcePath,
		"key_hint":           meta.KeyHint,
		"key_fp":             meta.KeyFingerprint,
		"override_template":  templateInfo,
	})
}

// ApplyChannelAffinityOverrideTemplate 合并渠道参数覆盖模板
func ApplyChannelAffinityOverrideTemplate(c *gin.Context, paramOverride map[string]interface{}) (map[string]interface{}, bool) {
	if c == nil {
		return paramOverride, false
	}
	meta, ok := getChannelAffinityMeta(c)
	if !ok {
		return paramOverride, false
	}
	if len(meta.ParamTemplate) == 0 {
		return paramOverride, false
	}
	mergedParam := mergeChannelOverride(paramOverride, meta.ParamTemplate)
	appendChannelAffinityTemplateAdminInfo(c, meta)
	return mergedParam, true
}

// ==================== 核心亲和性路由 ====================

// GetPreferredChannelByAffinity 根据亲和性规则获取优先渠道ID
func GetPreferredChannelByAffinity(c *gin.Context, modelName string, usingGroup string) (int, bool) {
	setting := operation_setting.GetChannelAffinitySetting()
	if setting == nil || !setting.Enabled {
		return 0, false
	}
	path := ""
	if c != nil && c.Request != nil && c.Request.URL != nil {
		path = c.Request.URL.Path
	}
	userAgent := ""
	if c != nil && c.Request != nil {
		userAgent = c.Request.UserAgent()
	}

	for _, rule := range setting.Rules {
		if !matchAnyRegexCached(rule.ModelRegex, modelName) {
			continue
		}
		if len(rule.PathRegex) > 0 && !matchAnyRegexCached(rule.PathRegex, path) {
			continue
		}
		if len(rule.UserAgentInclude) > 0 && !matchAnyIncludeFold(rule.UserAgentInclude, userAgent) {
			continue
		}
		var affinityValue string
		var usedSource operation_setting.ChannelAffinityKeySource
		for _, src := range rule.KeySources {
			affinityValue = extractChannelAffinityValue(c, src)
			if affinityValue != "" {
				usedSource = src
				break
			}
		}
		if affinityValue == "" {
			continue
		}
		if rule.ValueRegex != "" && !matchAnyRegexCached([]string{rule.ValueRegex}, affinityValue) {
			continue
		}

		ttlSeconds := rule.TTLSeconds
		if ttlSeconds <= 0 {
			ttlSeconds = setting.DefaultTTLSeconds
		}
		cacheKeySuffix := buildChannelAffinityCacheKeySuffix(rule, modelName, usingGroup, affinityValue)
		cacheKeyFull := channelAffinityCacheNamespace + ":" + cacheKeySuffix
		setChannelAffinityContext(c, channelAffinityMeta{
			CacheKey:       cacheKeyFull,
			TTLSeconds:     ttlSeconds,
			RuleName:       rule.Name,
			SkipRetry:      rule.SkipRetryOnFailure,
			ParamTemplate:  cloneStringAnyMap(rule.ParamOverrideTemplate),
			KeySourceType:  strings.TrimSpace(usedSource.Type),
			KeySourceKey:   strings.TrimSpace(usedSource.Key),
			KeySourcePath:  strings.TrimSpace(usedSource.Path),
			KeyHint:        buildChannelAffinityKeyHint(affinityValue),
			KeyFingerprint: affinityFingerprint(affinityValue),
			UsingGroup:     usingGroup,
			ModelName:      modelName,
			RequestPath:    path,
		})

		cache := getChannelAffinityCache()
		channelID, found, _ := cache.Get(cacheKeySuffix)
		if found && channelID > 0 {
			return channelID, true
		}
		// 当前规则缓存未命中，继续尝试其他匹配规则
		continue
	}
	return 0, false
}

// ShouldSkipRetryAfterChannelAffinityFailure 检查是否应跳过重试
func ShouldSkipRetryAfterChannelAffinityFailure(c *gin.Context) bool {
	if c == nil {
		return false
	}
	v, ok := c.Get(ginKeyChannelAffinitySkipRetry)
	if ok {
		b, ok := v.(bool)
		if ok {
			return b
		}
	}
	meta, ok := getChannelAffinityMeta(c)
	if !ok {
		return false
	}
	return meta.SkipRetry
}

// MarkChannelAffinityUsed 标记亲和性路由已被使用
func MarkChannelAffinityUsed(c *gin.Context, selectedGroup string, channelID int) {
	if c == nil || channelID <= 0 {
		return
	}
	meta, ok := getChannelAffinityMeta(c)
	if !ok {
		return
	}
	c.Set(ginKeyChannelAffinitySkipRetry, meta.SkipRetry)
	info := map[string]interface{}{
		"reason":          meta.RuleName,
		"rule_name":       meta.RuleName,
		"using_group":     meta.UsingGroup,
		"selected_group":  selectedGroup,
		"model":           meta.ModelName,
		"request_path":    meta.RequestPath,
		"channel_id":      channelID,
		"key_source":      meta.KeySourceType,
		"key_key":         meta.KeySourceKey,
		"key_path":        meta.KeySourcePath,
		"key_hint":        meta.KeyHint,
		"key_fp":          meta.KeyFingerprint,
	}
	c.Set(ginKeyChannelAffinityLogInfo, info)
}

// AppendChannelAffinityAdminInfo 追加亲和性信息到管理日志
func AppendChannelAffinityAdminInfo(c *gin.Context, adminInfo map[string]interface{}) {
	if c == nil || adminInfo == nil {
		return
	}
	anyInfo, ok := c.Get(ginKeyChannelAffinityLogInfo)
	if !ok || anyInfo == nil {
		return
	}
	adminInfo["channel_affinity"] = anyInfo
}

// RecordChannelAffinity 记录亲和性路由命中（成功时缓存）
func RecordChannelAffinity(c *gin.Context, channelID int) {
	if channelID <= 0 {
		return
	}
	setting := operation_setting.GetChannelAffinitySetting()
	if setting == nil || !setting.Enabled {
		return
	}
	if setting.SwitchOnSuccess && c != nil {
		if successChannelID := c.GetInt("channel_id"); successChannelID > 0 {
			channelID = successChannelID
		}
	}
	cacheKey, ttlSeconds, ok := getChannelAffinityContext(c)
	if !ok {
		return
	}
	if ttlSeconds <= 0 {
		ttlSeconds = setting.DefaultTTLSeconds
	}
	if ttlSeconds <= 0 {
		ttlSeconds = 3600
	}
	cache := getChannelAffinityCache()
	cache.SetWithTTL(cacheKey, channelID, time.Duration(ttlSeconds)*time.Second)
}

// ==================== 使用量统计 ====================

type ChannelAffinityUsageCacheStats struct {
	RuleName       string `json:"rule_name"`
	UsingGroup     string `json:"using_group"`
	KeyFingerprint string `json:"key_fp"`
	Hit            int64  `json:"hit"`
	Total          int64  `json:"total"`
	WindowSeconds  int64  `json:"window_seconds"`
	PromptTokens   int64  `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens    int64  `json:"total_tokens"`
	CachedTokens   int64  `json:"cached_tokens"`
	LastSeenAt     int64  `json:"last_seen_at"`
}

type channelAffinityUsageCacheCounters struct {
	Hit              int64 `json:"hit"`
	Total            int64 `json:"total"`
	WindowSeconds    int64 `json:"window_seconds"`
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
	CachedTokens     int64 `json:"cached_tokens"`
	LastSeenAt       int64 `json:"last_seen_at"`
}

var channelAffinityUsageStatsLocks [64]sync.Mutex
var channelAffinityUsageStats sync.Map // key: "ruleName\nusingGroup\nkeyFp" -> *channelAffinityUsageCacheCounters

// ObserveChannelAffinityUsageCacheFromContext 观察使用量（从上下文）
func ObserveChannelAffinityUsageCacheFromContext(c *gin.Context, promptTokens, completionTokens, cachedTokens int64) {
	statsCtx, ok := GetChannelAffinityStatsContext(c)
	if !ok {
		return
	}
	observeUsageCache(statsCtx, promptTokens, completionTokens, cachedTokens)
}

func observeUsageCache(statsCtx ChannelAffinityStatsContext, promptTokens, completionTokens, cachedTokens int64) {
	entryKey := statsCtx.RuleName + "\n" + statsCtx.UsingGroup + "\n" + statsCtx.KeyFingerprint
	lock := channelAffinityUsageCacheStatsLock(entryKey)
	lock.Lock()
	defer lock.Unlock()

	anyVal, _ := channelAffinityUsageStats.LoadOrStore(entryKey, &channelAffinityUsageCacheCounters{})
	counters := anyVal.(*channelAffinityUsageCacheCounters)
	counters.Total++
	if cachedTokens > 0 {
		counters.Hit++
	}
	counters.WindowSeconds = statsCtx.TTLSeconds
	counters.LastSeenAt = time.Now().Unix()
	counters.PromptTokens += promptTokens
	counters.CompletionTokens += completionTokens
	counters.TotalTokens += promptTokens + completionTokens
	counters.CachedTokens += cachedTokens
}

func channelAffinityUsageCacheStatsLock(key string) *sync.Mutex {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := h.Sum32() % uint32(len(channelAffinityUsageStatsLocks))
	return &channelAffinityUsageStatsLocks[idx]
}

// GetChannelAffinityUsageCacheStats 获取使用量统计
func GetChannelAffinityUsageCacheStats(ruleName, usingGroup, keyFp string) ChannelAffinityUsageCacheStats {
	entryKey := strings.TrimSpace(ruleName) + "\n" + strings.TrimSpace(usingGroup) + "\n" + strings.TrimSpace(keyFp)
	anyVal, ok := channelAffinityUsageStats.Load(entryKey)
	if !ok {
		return ChannelAffinityUsageCacheStats{
			RuleName:       ruleName,
			UsingGroup:     usingGroup,
			KeyFingerprint: keyFp,
		}
	}
	counters := anyVal.(*channelAffinityUsageCacheCounters)
	return ChannelAffinityUsageCacheStats{
		RuleName:         ruleName,
		UsingGroup:       usingGroup,
		KeyFingerprint:   keyFp,
		Hit:              counters.Hit,
		Total:            counters.Total,
		WindowSeconds:    counters.WindowSeconds,
		PromptTokens:     counters.PromptTokens,
		CompletionTokens: counters.CompletionTokens,
		TotalTokens:      counters.TotalTokens,
		CachedTokens:     counters.CachedTokens,
		LastSeenAt:       counters.LastSeenAt,
	}
}
