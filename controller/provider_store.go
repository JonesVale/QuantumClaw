package controller

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
)

// ==================== 渠道商自服务 ====================

// GetMyStore 获取/创建我的店铺
func GetMyStore(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	store, err := model.GetOrCreateStore(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取店铺失败"})
		return
	}
	modelCount, _ := model.CountStoreModels(store.Id)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"store":       store,
			"model_count": modelCount,
		},
	})
}

// SaveMyStore 保存/更新我的店铺信息
func SaveMyStore(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	store, err := model.GetOrCreateStore(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取店铺失败"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Logo        string `json:"logo"`
		BannerUrl   string `json:"banner_url"`
		ContactInfo string `json:"contact_info"`
		StoreSlug   string `json:"store_slug"`
		Status      int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	store.Name = req.Name
	store.Description = req.Description
	store.Logo = req.Logo
	store.BannerUrl = req.BannerUrl
	store.ContactInfo = req.ContactInfo
	store.Status = req.Status

	// slug 唯一性检查
	if req.StoreSlug != "" && req.StoreSlug != store.StoreSlug {
		var existing model.ProviderStore
		if model.DB.Where("store_slug = ? AND id != ?", req.StoreSlug, store.Id).First(&existing).Error == nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "该店铺路径已被占用"})
			return
		}
		store.StoreSlug = req.StoreSlug
	}

	if err := model.UpdateStore(store); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "保存失败"})
		return
	}

	model.RecordLog(c.Request.Context(), userId, model.LogTypeSystem, "更新店铺信息")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "店铺信息已保存", "data": store})
}

// ToggleStoreStatus 切换营业/歇业
func ToggleStoreStatus(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	store, err := model.GetOrCreateStore(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取店铺失败"})
		return
	}
	newStatus := 1
	if store.Status == 1 {
		newStatus = 0
	}
	model.DB.Model(store).Update("status", newStatus)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    gin.H{"status": newStatus},
	})
}

// ==================== 店铺商品管理 ====================

// GetMyStoreModels 获取我的店铺上架模型
func GetMyStoreModels(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	store, err := model.GetOrCreateStore(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取店铺失败"})
		return
	}
	activeOnly := c.Query("active") == "1"
	models, err := model.GetStoreModels(store.Id, activeOnly)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取模型列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": models})
}

// AddStoreModel 上架模型到店铺
func AddStoreModel(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	store, err := model.GetOrCreateStore(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取店铺失败"})
		return
	}

	var req struct {
		ChannelId       int     `json:"channel_id" binding:"required"`
		ModelName       string  `json:"model_name" binding:"required"`
		DisplayName     string  `json:"display_name"`
		Description     string  `json:"description"`
		Tags            string  `json:"tags"`
		InputPrice      int64   `json:"input_price"`
		OutputPrice     int64   `json:"output_price"`
		CacheReadPrice  int64   `json:"cache_read_price"`
		PriceMultiplier float64 `json:"price_multiplier"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	// 校验渠道归属
	var ch model.Channel
	if err := model.DB.First(&ch, req.ChannelId).Error; err != nil || ch.UserId != userId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "渠道不存在或不属于您"})
		return
	}

	sm := model.StoreModel{
		StoreId:         store.Id,
		ChannelId:       req.ChannelId,
		ModelName:       req.ModelName,
		DisplayName:     req.DisplayName,
		Description:     req.Description,
		Tags:            req.Tags,
		InputPrice:      req.InputPrice,
		OutputPrice:     req.OutputPrice,
		CacheReadPrice:  req.CacheReadPrice,
		PriceMultiplier: req.PriceMultiplier,
		IsActive:        true,
	}
	if sm.DisplayName == "" {
		sm.DisplayName = sm.ModelName
	}
	if sm.PriceMultiplier <= 0 {
		sm.PriceMultiplier = 1.0
	}

	if err := model.AddStoreModel(&sm); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "上架失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "模型已上架", "data": sm})
}

// UpdateStoreModel 编辑店铺模型信息
func UpdateStoreModel(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效ID"})
		return
	}

	// 校验归属
	var sm model.StoreModel
	if err := model.DB.First(&sm, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "模型不存在"})
		return
	}
	var store model.ProviderStore
	if err := model.DB.First(&store, sm.StoreId).Error; err != nil || store.UserId != userId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无权操作"})
		return
	}
	_ = store

	var req struct {
		DisplayName     string  `json:"display_name"`
		Description     string  `json:"description"`
		Tags            string  `json:"tags"`
		InputPrice      int64   `json:"input_price"`
		OutputPrice     int64   `json:"output_price"`
		CacheReadPrice  int64   `json:"cache_read_price"`
		PriceMultiplier float64 `json:"price_multiplier"`
		IsActive        *bool   `json:"is_active"`
		SortOrder       int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if req.DisplayName != "" {
		sm.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		sm.Description = req.Description
	}
	if req.Tags != "" {
		sm.Tags = req.Tags
	}
	if req.InputPrice > 0 {
		sm.InputPrice = req.InputPrice
	}
	if req.OutputPrice > 0 {
		sm.OutputPrice = req.OutputPrice
	}
	if req.CacheReadPrice > 0 {
		sm.CacheReadPrice = req.CacheReadPrice
	}
	if req.PriceMultiplier > 0 {
		sm.PriceMultiplier = req.PriceMultiplier
	}
	if req.IsActive != nil {
		sm.IsActive = *req.IsActive
	}
	sm.SortOrder = req.SortOrder

	if err := model.UpdateStoreModel(&sm); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已更新", "data": sm})
}

// DeleteStoreModel 下架模型
func DeleteStoreModel(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效ID"})
		return
	}

	var sm model.StoreModel
	if err := model.DB.First(&sm, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "模型不存在"})
		return
	}
	var store model.ProviderStore
	if err := model.DB.First(&store, sm.StoreId).Error; err != nil || store.UserId != userId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无权操作"})
		return
	}

	if err := model.DeleteStoreModel(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已下架"})
}

// ==================== 公开展示 ====================

// ListActiveStores 店铺广场 — 所有营业中的店铺
func ListActiveStores(c *gin.Context) {
	stores, err := model.GetAllActiveStores()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stores})
}

// GetPublicStore 单个公开店铺页
func GetPublicStore(c *gin.Context) {
	slug := c.Param("slug")
	store, err := model.GetStoreBySlug(slug)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "店铺不存在"})
		return
	}

	tag := c.Query("tag")
	var models []model.StoreModel
	if tag != "" {
		models, err = model.ListStoreModelsByTag(store.Id, tag)
	} else {
		models, err = model.GetStoreModels(store.Id, true)
	}
	if err != nil {
		models = []model.StoreModel{}
		_ = err
	}

	var owner model.User
	model.DB.Select("username, display_name, avatar_url").First(&owner, store.UserId)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"store":  store,
			"owner":  owner,
			"models": models,
		},
	})
}

// SearchStores 搜索店铺
func SearchStores(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []model.StoreWithOwner{}})
		return
	}
	// 模糊匹配店铺名
	var result []model.StoreWithOwner
	model.DB.Raw(`
		SELECT s.*, u.username,
			(SELECT COUNT(*) FROM store_models sm WHERE sm.store_id = s.id AND sm.is_active = 1) AS active_model,
			(SELECT COUNT(*) FROM store_models sm WHERE sm.store_id = s.id) AS model_count
		FROM provider_stores s
		JOIN users u ON u.id = s.user_id
		WHERE s.status = 1 AND (s.name LIKE ? OR s.store_slug LIKE ?)
		ORDER BY s.rating DESC
		LIMIT 20
	`, "%"+keyword+"%", "%"+keyword+"%").Scan(&result)
	if result == nil {
		result = []model.StoreWithOwner{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// PreviewStoreSlug 检查 slug 是否可用
func PreviewStoreSlug(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "slug不能为空"})
		return
	}
	valid, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`, slug)
	if !valid {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "slug格式无效（仅支持小写字母、数字和连字符）"})
		return
	}
	var existing model.ProviderStore
	if model.DB.Where("store_slug = ?", slug).First(&existing).Error == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "slug已被占用", "data": gin.H{"available": false}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"available": true}})
}

// 辅助：格式化为货币字符串
func formatPrice(price int64) string {
	if price == 0 {
		return "按倍率"
	}
	return fmt.Sprintf("¥%.2f", float64(price)/100)
}
