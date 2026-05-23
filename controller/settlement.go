package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetSettlementConfigs 获取所有结算配置
// GET /api/settlement/config
func GetSettlementConfigs(c *gin.Context) {
	var configs []model.SettlementConfig
	if err := model.DB.Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": configs})
}

// UpdateSettlementConfig 更新结算配置
// PUT /api/settlement/config/:id
func UpdateSettlementConfig(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ModelName       string  `json:"model_name"`
		UnifiedCost     float64 `json:"unified_cost"`
		CommissionRate  float64 `json:"commission_rate"`
		PlatformFeeRate float64 `json:"platform_fee_rate"`
		Enabled         *int    `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	updates := map[string]interface{}{
		"updated_time": time.Now().Unix(),
	}
	if req.ModelName != "" {
		updates["model_name"] = req.ModelName
	}
	if req.UnifiedCost > 0 {
		updates["unified_cost"] = req.UnifiedCost
	}
	if req.CommissionRate > 0 {
		updates["commission_rate"] = req.CommissionRate
	}
	if req.PlatformFeeRate > 0 {
		updates["platform_fee_rate"] = req.PlatformFeeRate
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := model.DB.Model(&model.SettlementConfig{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "updated"})
}

// CreateSettlementConfig 创建结算配置
// POST /api/settlement/config
func CreateSettlementConfig(c *gin.Context) {
	var req struct {
		ModelName       string  `json:"model_name" binding:"required"`
		UnifiedCost     float64 `json:"unified_cost"`
		CommissionRate  float64 `json:"commission_rate"`
		PlatformFeeRate float64 `json:"platform_fee_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "model_name is required"})
		return
	}

	cfg := model.SettlementConfig{
		ModelName:       req.ModelName,
		UnifiedCost:     req.UnifiedCost,
		CommissionRate:  req.CommissionRate,
		PlatformFeeRate: req.PlatformFeeRate,
		Enabled:         1,
		CreatedTime:     time.Now().Unix(),
		UpdatedTime:     time.Now().Unix(),
	}
	if cfg.UnifiedCost == 0 {
		cfg.UnifiedCost = 0.001000
	}
	if cfg.CommissionRate == 0 {
		cfg.CommissionRate = 0.2000
	}
	if cfg.PlatformFeeRate == 0 {
		cfg.PlatformFeeRate = 0.1000
	}

	if err := model.DB.Create(&cfg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

// DeleteSettlementConfig 删除结算配置
// DELETE /api/settlement/config/:id
func DeleteSettlementConfig(c *gin.Context) {
	id := c.Param("id")
	if err := model.DB.Delete(&model.SettlementConfig{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "deleted"})
}

// GetTransactions 获取交易流水（可按角色过滤）
// GET /api/transactions?user_id=&promoter_id=&channel_owner_id=&model=&page=&page_size=
func GetTransactions(c *gin.Context) {
	query := model.DB.Model(&model.TokenTransaction{})

	if userId := c.Query("user_id"); userId != "" {
		query = query.Where("user_id = ?", userId)
	}
	if promoterId := c.Query("promoter_id"); promoterId != "" {
		query = query.Where("promoter_id = ?", promoterId)
	}
	if ownerId := c.Query("channel_owner_id"); ownerId != "" {
		query = query.Where("channel_owner_id = ?", ownerId)
	}
	if modelName := c.Query("model"); modelName != "" {
		query = query.Where("model_name = ?", modelName)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 { page = 1 }
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 { pageSize = 20 }

	var total int64
	query.Count(&total)

	var txns []model.TokenTransaction
	query.Order("created_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&txns)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"transactions": txns,
			"total":        total,
			"page":         page,
			"page_size":    pageSize,
		},
	})
}
