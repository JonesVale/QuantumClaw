package controller

import (
	"encoding/json"
	"net/http"

	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
	"github.com/quantumclaw/quantumclaw/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

// ==================== 渠道亲和性设置管理 ====================

// GetChannelAffinitySettings 获取渠道亲和性设置
func GetChannelAffinitySettings(c *gin.Context) {
	setting := operation_setting.GetChannelAffinitySetting()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    setting,
	})
}

// SaveChannelAffinitySettings 保存渠道亲和性设置
func SaveChannelAffinitySettings(c *gin.Context) {
	var req operation_setting.ChannelAffinitySetting
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误: " + err.Error()})
		return
	}
	operation_setting.SetChannelAffinitySetting(&req)
	// 保存到数据库
	saveSettingToDB("ChannelAffinitySetting", &req)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
}

// ClearChannelAffinityCache 清空渠道亲和性缓存
func ClearChannelAffinityCache(c *gin.Context) {
	ruleName := c.Query("rule_name")
	var count int
	var err error
	if ruleName != "" {
		count, err = service.ClearChannelAffinityCacheByRuleName(ruleName)
	} else {
		count = service.ClearChannelAffinityCacheAll()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "清空失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"cleared": count}})
}

// GetChannelAffinityCacheStatsHandler 获取缓存统计
func GetChannelAffinityCacheStatsHandler(c *gin.Context) {
	stats := service.GetChannelAffinityCacheStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// saveSettingToDB 将设置序列化后存入 Option 表
func saveSettingToDB(key string, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	if err := model.UpdateOption(key, string(data)); err != nil {
		// 日志由 model.UpdateOption 输出
	}
}
