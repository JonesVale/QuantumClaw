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
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "saved"})
}
