package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetDistributors — 管理员获取所有分销商
func GetDistributors(c *gin.Context) {
	ds, err := model.GetAllDistributors()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ds})
}

// GetMyDistributor — 当前用户获取自己的分销商信息
func GetMyDistributor(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	d, err := model.GetDistributorByUserId(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "not a distributor"})
		return
	}
	rules, _ := model.GetDistributorPricingRules(d.Id)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"distributor": d,
		"pricing":     rules,
	}})
}

// CreateDistributor — 创建分销商
func CreateDistributor(c *gin.Context) {
	var req struct {
		UserId       int     `json:"user_id"`
		Name         string  `json:"name"`
		ContactEmail string  `json:"contact_email"`
		MarkupRate   float64 `json:"markup_rate"`
		ProfitSplit  float64 `json:"profit_split"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	// 检查用户是否存在
	user, err := model.GetUserById(req.UserId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "user not found"})
		return
	}
	// 生成 API key
	apiKey := common.GetRandomString(32)
	d := model.Distributor{
		UserId:       req.UserId,
		Name:         req.Name,
		ContactEmail: req.ContactEmail,
		MarkupRate:   req.MarkupRate,
		ProfitSplit:  req.ProfitSplit,
		Status:       1,
		ApiKey:       apiKey,
	}
	if err := model.DB.Create(&d).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	// 提升用户角色为分销商（1000 = 分销商角色）
	model.DB.Model(&user).Update("role", 1000)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": d})
}

// UpdateDistributor — 更新分销商
func UpdateDistributor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid id"})
		return
	}
	var req struct {
		Name         string  `json:"name"`
		ContactEmail string  `json:"contact_email"`
		MarkupRate   float64 `json:"markup_rate"`
		ProfitSplit  float64 `json:"profit_split"`
		Status       int     `json:"status"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ContactEmail != "" {
		updates["contact_email"] = req.ContactEmail
	}
	updates["markup_rate"] = req.MarkupRate
	updates["profit_split"] = req.ProfitSplit
	updates["status"] = req.Status
	if err := model.DB.Model(&model.Distributor{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "updated"})
}

// GetDistributorPricing — 获取分销商定价规则列表
func GetDistributorPricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid id"})
		return
	}
	rules, err := model.GetDistributorPricingRules(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rules})
}

// SetDistributorPricing — 设置分销商定价规则
func SetDistributorPricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid id"})
		return
	}
	var req struct {
		Rules []struct {
			ModelName      string  `json:"model_name"`
			PriceMultiplier float64 `json:"price_multiplier"`
			FixedPrice     int64   `json:"fixed_price"`
		} `json:"rules"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	// 删除旧规则
	model.DB.Where("distributor_id = ?", id).Delete(&model.DistributorPricingRule{})
	// 插入新规则
	for _, r := range req.Rules {
		rule := model.DistributorPricingRule{
			DistributorId:  id,
			ModelName:      r.ModelName,
			PriceMultiplier: r.PriceMultiplier,
			FixedPrice:     r.FixedPrice,
		}
		model.DB.Create(&rule)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "pricing rules updated"})
}

// GetDistributorRevenue — 分销商营收报表
func GetDistributorRevenue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid id"})
		return
	}
	var d model.Distributor
	if err := model.DB.First(&d, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "distributor not found"})
		return
	}
	// 查询该分销商下所有用户的消费统计
	type RevenueStat struct {
		UserId      int    `json:"user_id"`
		Username    string `json:"username"`
		QuotaUsed   int64  `json:"quota_used"`
		Revenue     int64  `json:"revenue"`
		RequestCount int64 `json:"request_count"`
	}
	var stats []RevenueStat
	model.DB.Raw(`
		SELECT u.id as user_id, u.username, u.used_quota as quota_used,
			   u.used_quota as revenue, u.request_count
		FROM users u
		WHERE u.inviter_id = ?
		ORDER BY u.used_quota DESC
		LIMIT 100`, d.UserId).Scan(&stats)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"distributor": d,
			"downline":    stats,
		},
	})
}
