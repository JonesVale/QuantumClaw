package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// AdminGetFeeConfigs - list all tier fee configurations
func AdminGetFeeConfigs(c *gin.Context) {
	configs, err := model.GetAllFeeConfigs()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": configs})
}

// AdminUpdateFeeConfig - update a tier's fee configuration
func AdminUpdateFeeConfig(c *gin.Context) {
	adminID := c.GetInt("id")
	tier := c.Param("tier")

	var req struct {
		Rate    float64 `json:"rate"`
		MinSkip int64   `json:"min_skip"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid params"})
		return
	}
	if req.Rate <= 0 || req.Rate > 100 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "rate must be 0-100"})
		return
	}
	if req.MinSkip < 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "min_skip cannot be negative"})
		return
	}

	if err := model.UpdateFeeConfig(model.StoreTier(tier), req.Rate, req.MinSkip, adminID); err != nil {
		logger.Errorf(c.Request.Context(), "update fee config: %v", err)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "save failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "updated"})
}

// AdminGetStores - list all stores
func AdminGetStores(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	tier := c.Query("tier")
	status := c.Query("status")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := model.DB.Model(&model.Store{})
	if tier != "" {
		query = query.Where("tier = ?", tier)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		id, err := strconv.Atoi(keyword)
		if err == nil {
			query = query.Where("name LIKE ? OR id = ?", "%"+keyword+"%", id)
		} else {
			query = query.Where("name LIKE ?", "%"+keyword+"%")
		}
	}

	var total int64
	query.Count(&total)

	var stores []model.Store
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&stores).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if stores == nil {
		stores = []model.Store{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stores,
		"total":   total,
		"page":    page,
	})
}

// AdminGetStoreDetail - get store details
func AdminGetStoreDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	store, err := model.GetStoreByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "store not found"})
		return
	}
	listings, _ := model.GetListingsByStoreID(id)
	tierLogs, _ := model.GetStoreTierLogs(id, 20)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"store":     store,
			"listings":  listings,
			"tier_logs": tierLogs,
		},
	})
}

// AdminUpdateStoreTier - manually change store tier
func AdminUpdateStoreTier(c *gin.Context) {
	adminID := c.GetInt("id")
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Tier string `json:"tier" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid params"})
		return
	}

	store, err := model.GetStoreByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "store not found"})
		return
	}

	oldTier := store.Tier
	newTier := model.StoreTier(req.Tier)
	switch newTier {
	case model.StoreTierBasic, model.StoreTierGold, model.StoreTierFlagship:
		store.Tier = newTier
	default:
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid tier"})
		return
	}

	if err := model.UpdateStoreInfo(store); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "update failed"})
		return
	}
	model.CreateStoreTierLog(id, oldTier, newTier, "admin_set", adminID)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "tier updated"})
}

// AdminUpdateStoreStatus - suspend/activate store
func AdminUpdateStoreStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid params"})
		return
	}

	store, err := model.GetStoreByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "store not found"})
		return
	}
	switch req.Status {
	case "active", "suspended", "closed":
		store.Status = model.StoreStatus(req.Status)
	default:
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid status"})
		return
	}
	if err := model.UpdateStoreInfo(store); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "status updated"})
}

// AdminGetPlatformFeeSummary - platform fee summary report
func AdminGetPlatformFeeSummary(c *gin.Context) {
	var totalPending int64
	model.DB.Model(&model.PlatformFeeRecord{}).Where("status = ?", model.PlatformFeeStatusPending).
		Select("COALESCE(SUM(fee_amount), 0)").Scan(&totalPending)
	var totalDeducted int64
	model.DB.Model(&model.PlatformFeeRecord{}).Where("status = ?", model.PlatformFeeStatusDeducted).
		Select("COALESCE(SUM(fee_amount), 0)").Scan(&totalDeducted)

	var monthlyStats []struct {
		Period string `json:"period"`
		Amount int64  `json:"amount"`
		Count  int    `json:"count"`
	}
	model.DB.Model(&model.PlatformFeeRecord{}).
		Select("period, SUM(fee_amount) as amount, COUNT(*) as count").
		Where("status IN ?", []string{model.PlatformFeeStatusPending, model.PlatformFeeStatusDeducted}).
		Group("period").Order("period DESC").Limit(12).Scan(&monthlyStats)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_pending":  totalPending,
			"total_deducted": totalDeducted,
			"monthly":        monthlyStats,
		},
	})
}
