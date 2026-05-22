package controller

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
	relay "github.com/quantumclaw/quantumclaw/relay"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/openai"
	"github.com/quantumclaw/quantumclaw/relay/apitype"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	relaymodel "github.com/quantumclaw/quantumclaw/relay/model"
	"net/http"
	"sort"
	"strings"
	"time"
)

// https://platform.openai.com/docs/api-reference/models/list

type OpenAIModelPermission struct {
	Id                 string  `json:"id"`
	Object             string  `json:"object"`
	Created            int     `json:"created"`
	AllowCreateEngine  bool    `json:"allow_create_engine"`
	AllowSampling      bool    `json:"allow_sampling"`
	AllowLogprobs      bool    `json:"allow_logprobs"`
	AllowSearchIndices bool    `json:"allow_search_indices"`
	AllowView          bool    `json:"allow_view"`
	AllowFineTuning    bool    `json:"allow_fine_tuning"`
	Organization       string  `json:"organization"`
	Group              *string `json:"group"`
	IsBlocking         bool    `json:"is_blocking"`
}

type OpenAIModels struct {
	Id         string                  `json:"id"`
	Object     string                  `json:"object"`
	Created    int                     `json:"created"`
	OwnedBy    string                  `json:"owned_by"`
	Permission []OpenAIModelPermission `json:"permission"`
	Root       string                  `json:"root"`
	Parent     *string                 `json:"parent"`
}

var models []OpenAIModels
var modelsMap map[string]OpenAIModels
var channelId2Models map[int][]string

func init() {
	var permission []OpenAIModelPermission
	permission = append(permission, OpenAIModelPermission{
		Id:                 "modelperm-LwHkVFn8AcMItP432fKKDIKJ",
		Object:             "model_permission",
		Created:            1626777600,
		AllowCreateEngine:  true,
		AllowSampling:      true,
		AllowLogprobs:      true,
		AllowSearchIndices: false,
		AllowView:          true,
		AllowFineTuning:    false,
		Organization:       "*",
		Group:              nil,
		IsBlocking:         false,
	})
	// https://platform.openai.com/docs/models/model-endpoint-compatibility
	for i := 0; i < apitype.Dummy; i++ {
		if i == apitype.AIProxyLibrary {
			continue
		}
		adaptor := relay.GetAdaptor(i)
		channelName := adaptor.GetChannelName()
		modelNames := adaptor.GetModelList()
		for _, modelName := range modelNames {
			models = append(models, OpenAIModels{
				Id:         modelName,
				Object:     "model",
				Created:    1626777600,
				OwnedBy:    channelName,
				Permission: permission,
				Root:       modelName,
				Parent:     nil,
			})
		}
	}
	for _, channelType := range openai.CompatibleChannels {
		if channelType == channeltype.Azure {
			continue
		}
		channelName, channelModelList := openai.GetCompatibleChannelMeta(channelType)
		for _, modelName := range channelModelList {
			models = append(models, OpenAIModels{
				Id:         modelName,
				Object:     "model",
				Created:    1626777600,
				OwnedBy:    channelName,
				Permission: permission,
				Root:       modelName,
				Parent:     nil,
			})
		}
	}
	// ==================== 量子算力后端注册 ====================
	for qt := channeltype.IonQ; qt < channeltype.QuantumDummy; qt++ {
		qAdaptor, err := relay.GetQuantumAdaptor(qt)
		if err != nil {
			continue
		}
		backends, _ := qAdaptor.ListBackends(context.Background())
		provider := qAdaptor.ProviderName()
		for _, backend := range backends {
			models = append(models, OpenAIModels{
				Id:         backend,
				Object:     "model",
				Created:    1626777600,
				OwnedBy:    provider,
				Permission: permission,
				Root:       backend,
				Parent:     nil,
			})
		}
	}

	modelsMap = make(map[string]OpenAIModels)
	for _, model := range models {
		modelsMap[model.Id] = model
	}
	channelId2Models = make(map[int][]string)
	for i := 1; i < channeltype.Dummy; i++ {
		adaptor := relay.GetAdaptor(channeltype.ToAPIType(i))
		meta := &meta.Meta{
			ChannelType: i,
		}
		adaptor.Init(meta)
		channelId2Models[i] = adaptor.GetModelList()
	}
}

// ModelInfo — 模型完整信息（含渠道定价和提供商）
type ModelInfo struct {
	Name          string  `json:"name"`
	ChannelID     int     `json:"channel_id"`
	ChannelName   string  `json:"channel_name"`
	Provider      string  `json:"provider"`
	ProviderType  int     `json:"provider_type"`
	CostPerUnit   float64 `json:"cost_per_unit"`
	SellPriceRate float64 `json:"sell_price_rate"`
	InputPrice    float64 `json:"input_price"`
	OutputPrice   float64 `json:"output_price"`
	Status        int     `json:"status"`
	Group         string  `json:"group"`
}

func DashboardListModels(c *gin.Context) {
	configuredChannels, _ := model.GetAllChannels(0, 0, "all")
	var result []ModelInfo
	channelTypeNames := channeltype.ChannelTypeNames

	for _, ch := range configuredChannels {
		if ch.Key == "" || strings.HasPrefix(ch.Key, "PUT_YOUR") {
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

		provider := ""
		if name, ok := channelTypeNames[ch.Type]; ok {
			provider = name
		}

		// 计算每 token 价格
		// CostPerUnit 是以 1K tokens 为单位的成本，转换为每 token 价格
		perTokenCost := ch.CostPerUnit / 1000.0
		inputPrice := perTokenCost
		outputPrice := perTokenCost * ch.SellPriceRate

		for _, modelName := range modelNames {
			result = append(result, ModelInfo{
				Name:          modelName,
				ChannelID:     ch.Id,
				ChannelName:   ch.Name,
				Provider:      provider,
				ProviderType:  ch.Type,
				CostPerUnit:   ch.CostPerUnit,
				SellPriceRate: ch.SellPriceRate,
				InputPrice:    inputPrice,
				OutputPrice:   outputPrice,
				Status:        ch.Status,
				Group:         ch.Group,
			})
		}
	}

	if result == nil {
		result = []ModelInfo{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    result,
	})
}

func ListAllModels(c *gin.Context) {
	// 只返回已配置渠道的模型（有真实 API Key 的渠道）
	configuredChannels, _ := model.GetAllChannels(0, 0, "all")
	configuredModelSet := make(map[string]bool)
	for _, ch := range configuredChannels {
		if ch.Key == "" || strings.HasPrefix(ch.Key, "PUT_YOUR") {
			continue
		}
		// 用该渠道的 Models 字段直接判断
		if ch.Models == "" {
			// 型号为空时，用该渠道 type 对应的默认模型列表
			if modelList, ok := channelId2Models[ch.Type]; ok {
				for _, m := range modelList {
					configuredModelSet[m] = true
				}
			}
		} else {
			for _, modelName := range strings.Split(ch.Models, ",") {
				modelName = strings.TrimSpace(modelName)
				if modelName != "" {
					configuredModelSet[modelName] = true
				}
			}
		}
	}

	var filteredModels []OpenAIModels
	for _, m := range models {
		if configuredModelSet[m.Id] {
			filteredModels = append(filteredModels, m)
		}
	}

	if filteredModels == nil {
		filteredModels = []OpenAIModels{}
	}

	c.JSON(200, gin.H{
		"object": "list",
		"data":   filteredModels,
	})
}

func ListModels(c *gin.Context) {
	ctx := c.Request.Context()
	var availableModels []string
	if c.GetString(ctxkey.AvailableModels) != "" {
		availableModels = strings.Split(c.GetString(ctxkey.AvailableModels), ",")
	} else {
		userId := c.GetInt(ctxkey.Id)
		userGroup, _ := model.CacheGetUserGroup(userId)
		availableModels, _ = model.CacheGetGroupModels(ctx, userGroup)
	}
	modelSet := make(map[string]bool)
	for _, availableModel := range availableModels {
		modelSet[availableModel] = true
	}
	availableOpenAIModels := make([]OpenAIModels, 0)
	for _, model := range models {
		if _, ok := modelSet[model.Id]; ok {
			modelSet[model.Id] = false
			availableOpenAIModels = append(availableOpenAIModels, model)
		}
	}
	for modelName, ok := range modelSet {
		if ok {
			availableOpenAIModels = append(availableOpenAIModels, OpenAIModels{
				Id:      modelName,
				Object:  "model",
				Created: 1626777600,
				OwnedBy: "custom",
				Root:    modelName,
				Parent:  nil,
			})
		}
	}
	c.JSON(200, gin.H{
		"object": "list",
		"data":   availableOpenAIModels,
	})
}

func RetrieveModel(c *gin.Context) {
	modelId := c.Param("model")
	if model, ok := modelsMap[modelId]; ok {
		c.JSON(200, model)
	} else {
		Error := relaymodel.Error{
			Message: fmt.Sprintf("The model '%s' does not exist", modelId),
			Type:    "invalid_request_error",
			Param:   "model",
			Code:    "model_not_found",
		}
		c.JSON(200, gin.H{
			"error": Error,
		})
	}
}

func GetUserAvailableModels(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.GetInt(ctxkey.Id)
	userGroup, err := model.CacheGetUserGroup(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	models, err := model.CacheGetGroupModels(ctx, userGroup)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    models,
	})
	return
}

// ListModelRankings — 模型排行榜（近7天请求量 + 趋势）
func ListModelRankings(c *gin.Context) {
	now := time.Now()
	endOfToday := now.Truncate(24*time.Hour).Add(24*time.Hour - time.Second).Unix()
	currentWeekStart := endOfToday - 7*24*3600
	previousWeekStart := currentWeekStart - 7*24*3600

	type RankingItem struct {
		Model           string  `json:"model"`
		Provider        string  `json:"provider"`
		ChannelName     string  `json:"channel_name"`
		Tokens7d        int64   `json:"tokens_7d"`
		TrendPercent    float64 `json:"trend_percent"`
		AvgSpeedMs      int     `json:"avg_speed_ms"`
		PricePer1k      float64 `json:"price_per_1k"`
		RequestCount7d  int64   `json:"request_count_7d"`
	}

	// 获取近7天各 model 在各 channel 上的统计数据
	stats, err := model.GetModelChannelStats(currentWeekStart, endOfToday)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取统计数据",
			"data":    nil,
		})
		return
	}

	// 获取上7天的统计数据用于趋势计算
	prevStats, _ := model.GetModelChannelStats(previousWeekStart, currentWeekStart)
	prevTokensByModel := make(map[string]int64)
	for _, ps := range prevStats {
		prevTokensByModel[ps.ModelName] += ps.TotalTokens
	}

	// 按 model 聚合（可能有多个 channel 服务于同一个 model）
	type aggregated struct {
		totalTokens   int64
		requestCount  int64
		avgSpeedSum   float64
		avgSpeedCount int
		pricePer1k    float64
		channelName   string
		provider      string
	}
	aggMap := make(map[string]*aggregated)
	// 按 model 获取首个 channel 信息
	channelTypeNames := channeltype.ChannelTypeNames
	channelCache := make(map[int]*model.Channel) // id → channel

	for _, ch := range getValidChannels() {
		channelCache[ch.Id] = ch
	}

	for _, s := range stats {
		key := s.ModelName
		if a, ok := aggMap[key]; ok {
			a.totalTokens += s.TotalTokens
			a.requestCount += s.RequestCount
			a.avgSpeedSum += s.AvgSpeedMs * float64(s.RequestCount)
			a.avgSpeedCount += int(s.RequestCount)
		} else {
			aggMap[key] = &aggregated{
				totalTokens:   s.TotalTokens,
				requestCount:  s.RequestCount,
				avgSpeedSum:   s.AvgSpeedMs * float64(s.RequestCount),
				avgSpeedCount: int(s.RequestCount),
			}
			// 填充渠道信息
			if ch, ok := channelCache[s.ChannelId]; ok {
				aggMap[key].pricePer1k = ch.CostPerUnit * ch.SellPriceRate
				aggMap[key].channelName = ch.Name
				if p, ok := channelTypeNames[ch.Type]; ok {
					aggMap[key].provider = p
				}
			}
		}
	}

	// 构建结果
	var result []RankingItem
	for modelName, a := range aggMap {
		avgSpeed := 0
		if a.avgSpeedCount > 0 {
			avgSpeed = int(a.avgSpeedSum / float64(a.avgSpeedCount))
		}

		trend := 0.0
		if prev, ok := prevTokensByModel[modelName]; ok && prev > 0 {
			trend = float64(a.totalTokens-prev) / float64(prev) * 100
		}

		result = append(result, RankingItem{
			Model:           modelName,
			Provider:        a.provider,
			ChannelName:     a.channelName,
			Tokens7d:        a.totalTokens,
			TrendPercent:    trend,
			AvgSpeedMs:      avgSpeed,
			PricePer1k:      a.pricePer1k,
			RequestCount7d:  a.requestCount,
		})
	}

	// 按请求量降序排列
	sort.Slice(result, func(i, j int) bool {
		return result[i].RequestCount7d > result[j].RequestCount7d
	})

	if result == nil {
		result = []RankingItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    result,
	})
}

// getValidChannels — 获取有真实 API Key 的渠道列表（懒加载，每次调用从 DB 读取）
func getValidChannels() []*model.Channel {
	allCh, _ := model.GetAllChannels(0, 0, "all")
	var valid []*model.Channel
	for _, ch := range allCh {
		if ch.Key != "" && !strings.HasPrefix(ch.Key, "PUT_YOUR") {
			valid = append(valid, ch)
		}
	}
	return valid
}
