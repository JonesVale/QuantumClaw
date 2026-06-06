package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetPlatformConfigs 获取所有平台配置
// GET /api/platform/config
func GetPlatformConfigs(c *gin.Context) {
	var configs []model.PlatformConfig
	if err := model.DB.Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Key] = cfg.Value
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// UpdatePlatformConfig 更新平台配置
// PUT /api/platform/config
func UpdatePlatformConfig(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	// 敏感配置项需要 Root(>=100) 权限
	rootKeys := map[string]bool{
		"platform_fee_rate_percent":      true,
		"platform_fee_min_revenue_cents": true,
		"transaction_fee_domestic":       true,
		"transaction_fee_foreign":        true,
		"transaction_fee_foreign_min":    true,
		"new_user_trial_balance_cents":   true,
		"quota_for_new_user":             true,
	}
	role := c.GetInt("role")
	for key := range req {
		if rootKeys[key] && role < model.RoleRootUser {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "敏感配置项(" + key + ")需要超级管理员权限"})
			return
		}
	}

	now := time.Now().Unix()
	for key, value := range req {
		var existing model.PlatformConfig
		if model.DB.Where("`key` = ?", key).First(&existing).Error == nil {
			model.DB.Model(&model.PlatformConfig{}).Where("`key` = ?", key).Updates(map[string]interface{}{
				"value":        value,
				"updated_time": now,
			})
		} else {
			model.DB.Create(&model.PlatformConfig{
				Key:         key,
				Value:       value,
				UpdatedTime: now,
			})
		}

		// 同步关键配置到运行时变量
		switch key {
		case "server_address":
			config.ServerAddress = value
		case "footer_html":
			config.Footer = value
		case "system_name":
			config.SystemName = value
		case "logo":
			config.Logo = value
		case "top_up_link":
			config.TopUpLink = value
		case "chat_link":
			config.ChatLink = value
		case "quota_per_unit":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				config.QuotaPerUnit = v
			}
		case "platform_fee_rate_percent":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				config.PlatformFeeRate = v
			}
		case "quota_for_new_user":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				config.QuotaForNewUser = v
			}
		case "new_user_trial_balance_cents":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				config.NewUserTrialBalance = v
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "saved"})
}
