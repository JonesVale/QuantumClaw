package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ==================== 共享类型 ====================

type SubscriptionPlanDTO struct {
	Plan model.SubscriptionPlan `json:"plan"`
}

// ==================== 用户 API ====================

// GetSubscriptionPlans 获取所有已启用的订阅套餐（用户可见）
func GetSubscriptionPlans(c *gin.Context) {
	var plans []model.SubscriptionPlan
	if err := model.DB.Where("enabled = ?", true).
		Order("sort_order desc, id desc").Find(&plans).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	result := make([]SubscriptionPlanDTO, 0, len(plans))
	for _, p := range plans {
		result = append(result, SubscriptionPlanDTO{Plan: p})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetSubscriptionSelf 获取当前用户的订阅信息
func GetSubscriptionSelf(c *gin.Context) {
	userId := c.GetInt("id")

	allSubscriptions, err := model.GetAllUserSubscriptions(userId)
	if err != nil {
		allSubscriptions = []model.SubscriptionSummary{}
	}
	activeSubscriptions, err := model.GetAllActiveUserSubscriptions(userId)
	if err != nil {
		activeSubscriptions = []model.SubscriptionSummary{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subscriptions":     activeSubscriptions,
			"all_subscriptions": allSubscriptions,
		},
	})
}

// ==================== 管理员 API ====================

// AdminListSubscriptionPlans 管理员查看所有套餐
func AdminListSubscriptionPlans(c *gin.Context) {
	var plans []model.SubscriptionPlan
	if err := model.DB.Order("sort_order desc, id desc").Find(&plans).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	result := make([]SubscriptionPlanDTO, 0, len(plans))
	for _, p := range plans {
		result = append(result, SubscriptionPlanDTO{Plan: p})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

type AdminUpsertSubscriptionPlanRequest struct {
	Plan model.SubscriptionPlan `json:"plan"`
}

// AdminCreateSubscriptionPlan 管理员创建订阅套餐
func AdminCreateSubscriptionPlan(c *gin.Context) {
	var req AdminUpsertSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	req.Plan.Id = 0
	if strings.TrimSpace(req.Plan.Title) == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "套餐标题不能为空"})
		return
	}
	if req.Plan.PriceCents < 0 || req.Plan.PriceCents > 999900 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "价格需在 0~9999 USD 之间（单位：美分）"})
		return
	}
	if req.Plan.Currency == "" {
		req.Plan.Currency = "USD"
	}
	if req.Plan.DurationUnit == "" {
		req.Plan.DurationUnit = model.SubscriptionDurationMonth
	}
	if req.Plan.DurationValue <= 0 && req.Plan.DurationUnit != model.SubscriptionDurationCustom {
		req.Plan.DurationValue = 1
	}
	if req.Plan.MaxPurchasePerUser < 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "购买上限不能为负数"})
		return
	}
	if req.Plan.TotalAmount < 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "总额度不能为负数"})
		return
	}
	req.Plan.UpgradeGroup = strings.TrimSpace(req.Plan.UpgradeGroup)
	if req.Plan.UpgradeGroup != "" {
		if !ratio_setting.ContainsGroupRatio(req.Plan.UpgradeGroup) {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "升级分组不存在"})
			return
		}
	}
	req.Plan.QuotaResetPeriod = model.NormalizeResetPeriod(req.Plan.QuotaResetPeriod)
	if req.Plan.QuotaResetPeriod == model.SubscriptionResetCustom && req.Plan.QuotaResetCustomSeconds <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "自定义重置周期需大于0秒"})
		return
	}
	if err := model.DB.Create(&req.Plan).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": req.Plan})
}

// AdminUpdateSubscriptionPlan 管理员更新订阅套餐
func AdminUpdateSubscriptionPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的ID"})
		return
	}
	var req AdminUpsertSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if strings.TrimSpace(req.Plan.Title) == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "套餐标题不能为空"})
		return
	}
	if req.Plan.PriceCents < 0 || req.Plan.PriceCents > 999900 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "价格需在 0~9999 之间"})
		return
	}
	req.Plan.Id = id
	if req.Plan.Currency == "" {
		req.Plan.Currency = "USD"
	}
	if req.Plan.DurationUnit == "" {
		req.Plan.DurationUnit = model.SubscriptionDurationMonth
	}
	if req.Plan.DurationValue <= 0 && req.Plan.DurationUnit != model.SubscriptionDurationCustom {
		req.Plan.DurationValue = 1
	}
	req.Plan.UpgradeGroup = strings.TrimSpace(req.Plan.UpgradeGroup)
	if req.Plan.UpgradeGroup != "" {
		if !ratio_setting.ContainsGroupRatio(req.Plan.UpgradeGroup) {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "升级分组不存在"})
			return
		}
	}
	req.Plan.QuotaResetPeriod = model.NormalizeResetPeriod(req.Plan.QuotaResetPeriod)
	if req.Plan.QuotaResetPeriod == model.SubscriptionResetCustom && req.Plan.QuotaResetCustomSeconds <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "自定义重置周期需大于0秒"})
		return
	}
	updateMap := map[string]interface{}{
		"title":                      req.Plan.Title,
		"subtitle":                   req.Plan.Subtitle,
		"price_cents":                req.Plan.PriceCents,
		"currency":                   req.Plan.Currency,
		"duration_unit":              req.Plan.DurationUnit,
		"duration_value":             req.Plan.DurationValue,
		"custom_seconds":             req.Plan.CustomSeconds,
		"enabled":                    req.Plan.Enabled,
		"sort_order":                 req.Plan.SortOrder,
		"stripe_price_id":            req.Plan.StripePriceId,
		"creem_product_id":           req.Plan.CreemProductId,
		"max_purchase_per_user":      req.Plan.MaxPurchasePerUser,
		"total_amount":               req.Plan.TotalAmount,
		"upgrade_group":              req.Plan.UpgradeGroup,
		"quota_reset_period":         req.Plan.QuotaResetPeriod,
		"quota_reset_custom_seconds": req.Plan.QuotaResetCustomSeconds,
		"updated_at":                 common.GetTimestamp(),
	}
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&model.SubscriptionPlan{}).Where("id = ?", id).Updates(updateMap).Error
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type AdminUpdateSubscriptionPlanStatusRequest struct {
	Enabled *bool `json:"enabled"`
}

// AdminUpdateSubscriptionPlanStatus 管理员切换套餐启用状态
func AdminUpdateSubscriptionPlanStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的ID"})
		return
	}
	var req AdminUpdateSubscriptionPlanStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if err := model.DB.Model(&model.SubscriptionPlan{}).
		Where("id = ?", id).Update("enabled", *req.Enabled).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminDeleteSubscriptionPlan 管理员删除套餐
func AdminDeleteSubscriptionPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的ID"})
		return
	}
	if err := model.DB.Delete(&model.SubscriptionPlan{}, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type AdminBindSubscriptionRequest struct {
	UserId int `json:"user_id"`
	PlanId int `json:"plan_id"`
}

// AdminBindSubscription 管理员为用户绑定订阅（免支付）
func AdminBindSubscription(c *gin.Context) {
	var req AdminBindSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UserId <= 0 || req.PlanId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	msg, err := model.AdminBindSubscription(req.UserId, req.PlanId, "")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

// AdminListUserSubscriptions 管理员查看用户所有订阅
func AdminListUserSubscriptions(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Param("id"))
	if userId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的用户ID"})
		return
	}
	subs, err := model.GetAllUserSubscriptions(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": subs})
}

type AdminCreateUserSubscriptionRequest struct {
	PlanId int `json:"plan_id"`
}

// AdminCreateUserSubscription 管理员为用户创建订阅（无需支付）
func AdminCreateUserSubscription(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Param("id"))
	if userId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的用户ID"})
		return
	}
	var req AdminCreateUserSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	msg, err := model.AdminBindSubscription(userId, req.PlanId, "")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

// AdminInvalidateUserSubscription 管理员立即取消用户订阅
func AdminInvalidateUserSubscription(c *gin.Context) {
	subId, _ := strconv.Atoi(c.Param("id"))
	if subId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的订阅ID"})
		return
	}
	msg, err := model.AdminInvalidateUserSubscription(subId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

// AdminDeleteUserSubscription 管理员硬删除用户订阅记录
func AdminDeleteUserSubscription(c *gin.Context) {
	subId, _ := strconv.Atoi(c.Param("id"))
	if subId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的订阅ID"})
		return
	}
	msg, err := model.AdminDeleteUserSubscription(subId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}
