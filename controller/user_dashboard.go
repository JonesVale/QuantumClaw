package controller

// ============================================================
// user_dashboard.go — 统计面板（Dashboard + 维度统计）
// 从原 controller/user.go 拆分，包含 DailyStat/ModelStat/ProviderStat
// ============================================================

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
)

// DailyStat 每日请求统计
type DailyStat struct {
	Date         string `json:"date"`
	RequestCount int    `json:"request_count"`
	TokenCount   int    `json:"token_count"`
	QuotaUsed    int    `json:"quota_used"`
}

// ModelStat 模型维度统计
type ModelStat struct {
	ModelName    string `json:"model_name"`
	RequestCount int    `json:"request_count"`
	TokenCount   int    `json:"token_count"`
	QuotaUsed    int    `json:"quota_used"`
}

// ProviderStat 提供商维度统计
type ProviderStat struct {
	Provider     string `json:"provider"`
	RequestCount int    `json:"request_count"`
	TokenCount   int    `json:"token_count"`
	QuotaUsed    int    `json:"quota_used"`
}

func GetUserDashboard(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	now := time.Now()
	startOfDay := now.Truncate(24*time.Hour).AddDate(0, 0, -6).Unix()
	endOfDay := now.Truncate(24 * time.Hour).Add(24*time.Second - time.Second).Unix()

	dashboards, err := model.SearchLogsByDayAndModel(id, int(startOfDay), int(endOfDay))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取统计信息",
			"data":    nil,
		})
		return
	}

	dayMap := make(map[string]*DailyStat)
	modelMap := make(map[string]*ModelStat)
	modelProvider := buildModelProviderMap()
	providerMap := make(map[string]*ProviderStat)

	for _, d := range dashboards {
		tokens := d.PromptTokens + d.CompletionTokens

		if _, ok := dayMap[d.Day]; !ok {
			dayMap[d.Day] = &DailyStat{Date: d.Day}
		}
		dayMap[d.Day].RequestCount += d.RequestCount
		dayMap[d.Day].TokenCount += tokens
		dayMap[d.Day].QuotaUsed += d.Quota

		if _, ok := modelMap[d.ModelName]; !ok {
			modelMap[d.ModelName] = &ModelStat{ModelName: d.ModelName}
		}
		modelMap[d.ModelName].RequestCount += d.RequestCount
		modelMap[d.ModelName].TokenCount += tokens
		modelMap[d.ModelName].QuotaUsed += d.Quota

		provider := modelProvider[d.ModelName]
		if provider == "" {
			provider = "其他"
		}
		if _, ok := providerMap[provider]; !ok {
			providerMap[provider] = &ProviderStat{Provider: provider}
		}
		providerMap[provider].RequestCount += d.RequestCount
		providerMap[provider].TokenCount += tokens
		providerMap[provider].QuotaUsed += d.Quota
	}

	var dailyRequests []DailyStat
	for _, v := range dayMap {
		dailyRequests = append(dailyRequests, *v)
	}
	sort.Slice(dailyRequests, func(i, j int) bool {
		return dailyRequests[i].Date < dailyRequests[j].Date
	})

	var modelBreakdown []ModelStat
	for _, v := range modelMap {
		modelBreakdown = append(modelBreakdown, *v)
	}

	var providerBreakdown []ProviderStat
	for _, v := range providerMap {
		providerBreakdown = append(providerBreakdown, *v)
	}

	if dailyRequests == nil {
		dailyRequests = []DailyStat{}
	}
	if modelBreakdown == nil {
		modelBreakdown = []ModelStat{}
	}
	if providerBreakdown == nil {
		providerBreakdown = []ProviderStat{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"logs":               dashboards,
			"daily_requests":     dailyRequests,
			"model_breakdown":    modelBreakdown,
			"provider_breakdown": providerBreakdown,
		},
	})
	return
}

// buildModelProviderMap 从已配置渠道构建 model_name → provider 名称的映射
func buildModelProviderMap() map[string]string {
	result := make(map[string]string)
	allCh, _ := model.GetAllChannels(0, 0, "all")
	channelTypeNames := channeltype.ChannelTypeNames

	for _, ch := range allCh {
		if ch.Key == "" || strings.HasPrefix(ch.Key, "PUT_YOUR") {
			continue
		}
		provider := ""
		if name, ok := channelTypeNames[ch.Type]; ok {
			provider = name
		}
		if provider == "" {
			continue
		}

		var modelNames []string
		if ch.Models == "" {
			if modelList, ok := channelId2Models[ch.Type]; ok {
				modelNames = modelList
			}
		} else {
			for _, m := range strings.Split(ch.Models, ",") {
				m = strings.TrimSpace(m)
				if m != "" {
					modelNames = append(modelNames, m)
				}
			}
		}
		for _, m := range modelNames {
			if _, exists := result[m]; !exists {
				result[m] = provider
			}
		}
	}
	return result
}
