package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"net/http"
	"strconv"
	"strings"
)

// ChannelTypeDetail 渠道类型详细信息（含默认 URL 和可用模型列表）
type ChannelTypeDetail struct {
	Id        int      `json:"id"`
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	Models    []string `json:"models,omitempty"`
}

func GetChannelTypes(c *gin.Context) {
	typeNames := channeltype.ChannelTypeNames
	baseURLs := channeltype.ChannelBaseURLs

	// 查询 model_metadata 获取各提供商/品牌的模型列表
	var metadata []model.ModelMetadata
	model.DB.Select("DISTINCT model_name, provider").Where("languages_type = ?", "English").Find(&metadata)

	modelsByProvider := make(map[string][]string)
	for _, m := range metadata {
		if m.Provider != "" && m.ModelName != "" {
			modelsByProvider[m.Provider] = append(modelsByProvider[m.Provider], m.ModelName)
		}
	}

	result := make([]ChannelTypeDetail, 0, len(typeNames))
	for id, name := range typeNames {
		detail := ChannelTypeDetail{
			Id:   id,
			Name: name,
		}
		if id >= 0 && id < len(baseURLs) && baseURLs[id] != "" {
			detail.URL = baseURLs[id]
		}
		// 尝试通过名称匹配 model_metadata 中的 provider
		// 优先使用 ChannelTypeNameToProvider 映射表，再回退到直接名称匹配
		providerKey := name
		if mapped, ok := channeltype.ChannelTypeNameToProvider[name]; ok && mapped != "" {
			providerKey = mapped
		}
		if models, ok := modelsByProvider[providerKey]; ok && len(models) > 0 {
			detail.Models = models
		}
		result = append(result, detail)
	}

	c.JSON(http.StatusOK, result)
}

// GetChannelProfit 获取渠道利润分析
func GetChannelProfit(c *gin.Context) {
	channels, err := model.GetAllChannels(0, 0, "all")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	type ProfitItem struct {
		Id            int     `json:"id"`
		Name          string  `json:"name"`
		Type          int     `json:"type"`
		UsedQuota     int64   `json:"used_quota"`
		CostPerUnit   float64 `json:"cost_per_unit"`
		SellPriceRate float64 `json:"sell_price_rate"`
		SellPrice     float64 `json:"sell_price"`
		TotalCost     float64 `json:"total_cost"`
		TotalRevenue  float64 `json:"total_revenue"`
		Profit        float64 `json:"profit"`
		Margin        float64 `json:"margin"`
	}
	result := make([]ProfitItem, 0, len(channels))
	for _, ch := range channels {
		qp := config.QuotaPerUnit
		sellPrice := 1.0
		if qp > 0 {
			sellPrice = 1.0 / qp * 1000000
		}
		// Use per-channel SellPriceRate if set (> 0), otherwise use global conversion
		spr := ch.SellPriceRate
		revenueMultiplier := 1.0
		if spr > 0 {
			// SellPriceRate = multiplier on cost: revenue = cost * rate
			revenueMultiplier = spr * ch.CostPerUnit
			if revenueMultiplier <= 0 {
				revenueMultiplier = sellPrice
			}
		} else {
			revenueMultiplier = sellPrice
		}
		cost := float64(ch.UsedQuota) * ch.CostPerUnit / 1000000
		revenue := float64(ch.UsedQuota) * revenueMultiplier / 1000000
		profit := revenue - cost
		margin := float64(0)
		if revenue > 0 {
			margin = profit / revenue * 100
		}
		result = append(result, ProfitItem{
			Id:            ch.Id,
			Name:          ch.Name,
			Type:          ch.Type,
			UsedQuota:     ch.UsedQuota,
			CostPerUnit:   ch.CostPerUnit,
			SellPriceRate: spr,
			SellPrice:     sellPrice,
			TotalCost:     cost,
			TotalRevenue:  revenue,
			Profit:        profit,
			Margin:        margin,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func GetAllChannels(c *gin.Context) {
	role := c.GetInt("role")
	userId := c.GetInt("id")
	isAdmin := role >= model.RoleAdminUser

	// 管理员看全部，普通用户只看自己的
	query := model.DB.Model(&model.Channel{})
	if !isAdmin {
		query = query.Where("user_id = ?", userId)
	}

	scope := c.Query("scope")
	if scope == "" {
		scope = "limited"
	}

	var channels []*model.Channel
	var err error

	switch scope {
	case "all":
		if !isAdmin {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "无权查看全部渠道"})
			return
		}
		err = query.Order("id desc").Find(&channels).Error
	case "disabled":
		err = query.Where("status = ? or status = ?",
			model.ChannelStatusAutoDisabled, model.ChannelStatusManuallyDisabled).
			Order("id desc").Find(&channels).Error
	default:
		p, _ := strconv.Atoi(c.Query("p"))
		if p < 0 {
			p = 0
		}
		err = query.Order("id desc").Limit(config.ItemsPerPage).
			Offset(p * config.ItemsPerPage).Omit("key").Find(&channels).Error
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	// type_range 过滤
	typeRange := c.Query("type_range")
	if typeRange == "quantum" {
		filtered := make([]*model.Channel, 0, len(channels))
		for _, ch := range channels {
			if ch.Type >= 100 {
				filtered = append(filtered, ch)
			}
		}
		channels = filtered
	} else if typeRange == "ai" {
		filtered := make([]*model.Channel, 0, len(channels))
		for _, ch := range channels {
			if ch.Type < 100 {
				filtered = append(filtered, ch)
			}
		}
		channels = filtered
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": channels})
}

func SearchChannels(c *gin.Context) {
	keyword := c.Query("keyword")
	channels, err := model.SearchChannels(keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	// 普通用户只能搜到自己的渠道
	role := c.GetInt("role")
	if role < model.RoleAdminUser {
		userId := c.GetInt("id")
		filtered := make([]*model.Channel, 0, len(channels))
		for _, ch := range channels {
			if ch.UserId == userId {
				filtered = append(filtered, ch)
			}
		}
		channels = filtered
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": channels})
}

func GetChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	channel, err := model.GetChannelById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 普通用户只能查自己的渠道
	userId := c.GetInt("id")
	role := c.GetInt("role")
	if role < model.RoleAdminUser && channel.UserId != userId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无权查看此渠道"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": channel})
}

func AddChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	userId := c.GetInt("id")
	role := c.GetInt("role")

	// 普通用户必须是渠道商才能添加渠道
	if role < model.RoleAdminUser {
		user, err := model.GetUserById(userId, false)
		if err != nil || user.UserType != model.UserTypeProvider {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅渠道商可添加渠道，请在个人中心升级为渠道商"})
			return
		}
		channel.UserId = userId
	}

	channel.CreatedTime = helper.GetTimestamp()
	keys := strings.Split(channel.Key, "\n")
	channels := make([]model.Channel, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		localChannel := channel
		localChannel.Key = key
		channels = append(channels, localChannel)
	}
	err = model.BatchInsertChannels(channels)
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
	})
	return
}

// ChannelPricingRequest 渠道定价请求
// SellPriceRate: 售卖倍率（如 2.5 = 以成本的2.5倍售卖）
type ChannelPricingRequest struct {
	Id            int     `json:"id"`
	SellPriceRate float64 `json:"sell_price_rate"`
}

// SetChannelPricing 设置渠道售卖倍率
func SetChannelPricing(c *gin.Context) {
	var req ChannelPricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if req.Id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid channel id"})
		return
	}
	if req.SellPriceRate < 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "sell_price_rate must be >= 0"})
		return
	}
	if err := model.DB.Model(&model.Channel{}).Where("id = ?", req.Id).
		Update("sell_price_rate", req.SellPriceRate).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": req})
}

// SetChannelCategory 设置渠道分类（free / paid / custom）
func SetChannelCategory(c *gin.Context) {
	var req struct {
		Id       int    `json:"id"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if req.Id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid channel id"})
		return
	}
	validCategories := map[string]bool{"free": true, "paid": true, "custom": true, "": true}
	if !validCategories[req.Category] {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "category must be free, paid, or custom"})
		return
	}
	if err := model.DB.Model(&model.Channel{}).Where("id = ?", req.Id).
		Update("category", req.Category).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": req})
}

func DeleteChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// 普通用户只能删自己的渠道
	userId := c.GetInt("id")
	role := c.GetInt("role")
	if role < model.RoleAdminUser {
		existing, err := model.GetChannelById(id, false)
		if err != nil || existing.UserId != userId {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "无权删除此渠道"})
			return
		}
	}
	channel := model.Channel{Id: id}
	err := channel.Delete()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

func DeleteDisabledChannel(c *gin.Context) {
	rows, err := model.DeleteDisabledChannel()
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
		"data":    rows,
	})
	return
}

func UpdateChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	// 普通用户只能改自己的渠道
	userId := c.GetInt("id")
	role := c.GetInt("role")
	if role < model.RoleAdminUser {
		existing, err := model.GetChannelById(channel.Id, false)
		if err != nil || existing.UserId != userId {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "无权修改此渠道"})
			return
		}
		channel.UserId = userId
	}
	err = channel.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 如果配置了真实的 API Key，自动拉取供应商的最新模型列表
	if channel.Key != "" && !strings.HasPrefix(channel.Key, "PUT_YOUR") {
		go func(ch model.Channel) {
			_ = ch.UpdateModelsFromProvider()
		}(channel)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": channel})
}

// FetchChannelModels 获取渠道的供应商模型列表（不保存，仅查询）
// GET /api/channel/:id/fetch-models
func FetchChannelModels(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	channel, err := model.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	models, err := channel.FetchModelsFromProvider()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    models,
	})
}
