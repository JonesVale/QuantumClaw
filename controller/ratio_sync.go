package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/setting/ratio_setting"
	"gorm.io/gorm"
)

// RatioSyncSetting 汇率同步设置
type RatioSyncSetting struct {
	Enabled       bool              `json:"enabled"`
	SourceURL     string            `json:"source_url"`
	SyncInterval  int               `json:"sync_interval"` // 分钟
	LastSyncTime  int64             `json:"last_sync_time"`
	Overrides     map[string]float64 `json:"overrides"`
}

var ratioSyncSetting = &RatioSyncSetting{
	Enabled:      false,
	SourceURL:    "https://api.exchangerate-api.com/v4/latest/USD",
	SyncInterval: 60,
	Overrides:    map[string]float64{},
}

// ==================== 汇率同步 API ====================

// GetRatioSyncSetting 获取汇率同步设置
func GetRatioSyncSetting(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ratioSyncSetting,
	})
}

// SaveRatioSyncSetting 保存汇率同步设置
func SaveRatioSyncSetting(c *gin.Context) {
	var req RatioSyncSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误: " + err.Error()})
		return
	}
	ratioSyncSetting.Enabled = req.Enabled
	ratioSyncSetting.SourceURL = req.SourceURL
	ratioSyncSetting.SyncInterval = req.SyncInterval
	if req.Overrides != nil {
		ratioSyncSetting.Overrides = req.Overrides
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// TriggerRatioSync 手动触发汇率同步
func TriggerRatioSync(c *gin.Context) {
	if !ratioSyncSetting.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "汇率同步未启用"})
		return
	}
	rates, err := fetchExchangeRates(ratioSyncSetting.SourceURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取汇率失败: " + err.Error()})
		return
	}
	for k, v := range ratioSyncSetting.Overrides {
		rates[k] = v
	}
	saved := applyRatioOverrides(rates)
	ratioSyncSetting.LastSyncTime = time.Now().Unix()
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"data":          rates,
		"saved_entries": saved,
	})
}

// ==================== 汇率获取 ====================

type exchangeRateResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

func fetchExchangeRates(sourceURL string) (map[string]float64, error) {
	if sourceURL == "" {
		return nil, fmt.Errorf("source URL is empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "QuantumClaw/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result exchangeRateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// 转换为 "BASE/QUOTE" 格式
	rates := make(map[string]float64)
	for _, pair := range []struct{ from, to string }{
		{"CNY", "USD/CNY"}, {"JPY", "USD/JPY"}, {"EUR", "USD/EUR"},
		{"GBP", "USD/GBP"}, {"KRW", "USD/KRW"}, {"AUD", "USD/AUD"},
		{"CAD", "USD/CAD"}, {"SGD", "USD/SGD"}, {"HKD", "USD/HKD"},
		{"TWD", "USD/TWD"}, {"INR", "USD/INR"},
	} {
		if rate, ok := result.Rates[pair.from]; ok {
			rates[pair.to] = rate
		}
	}

	return rates, nil
}

// ==================== 应用汇率到 ratio_setting ====================

var currencyRegex = regexp.MustCompile(`USD[A-Z]{3}|CNY[A-Z]{3}|EUR[A-Z]{3}|JPY[A-Z]{3}`)

func applyRatioOverrides(rates map[string]float64) int {
	if rates["USD/CNY"] > 0 {
		ratio_setting.SetUSDCNY(rates["USD/CNY"])
	}
	if rates["USD/JPY"] > 0 {
		ratio_setting.SetUSDJPY(rates["USD/JPY"])
	}
	if rates["USD/EUR"] > 0 {
		ratio_setting.SetUSDEUR(rates["USD/EUR"])
	}
	if rates["USD/GBP"] > 0 {
		ratio_setting.SetUSDGBP(rates["USD/GBP"])
	}
	saved := 0
	for k := range rates {
		if strings.HasPrefix(k, "USD/") {
			saved++
		}
	}
	logger.SysLogf("汇率同步完成: USD/CNY=%.4f USD/JPY=%.4f",
		rates["USD/CNY"], rates["USD/JPY"])
	return saved
}

// StartRatioSyncCron 启动汇率同步定时任务
func StartRatioSyncCron() {
	if !ratioSyncSetting.Enabled || ratioSyncSetting.SyncInterval <= 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(ratioSyncSetting.SyncInterval) * time.Minute)
	go func() {
		for range ticker.C {
			if !ratioSyncSetting.Enabled {
				continue
			}
			rates, err := fetchExchangeRates(ratioSyncSetting.SourceURL)
			if err != nil {
				logger.SysError("汇率同步失败: " + err.Error())
				continue
			}
			for k, v := range ratioSyncSetting.Overrides {
				rates[k] = v
			}
			applyRatioOverrides(rates)
			ratioSyncSetting.LastSyncTime = time.Now().Unix()
		}
	}()
}

// ==================== 辅助: 通用 option 读写 ====================

func getOptionValue(key string) string {
	var opt model.Option
	if err := model.DB.Where("`key` = ?", key).First(&opt).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.SysError("getOptionValue error: " + err.Error())
		}
		return ""
	}
	return opt.Value
}

func setOptionValue(key, value string) {
	model.DB.Exec("INSERT INTO `options` (`key`, `value`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `value` = ?",
		key, value, value)
}
