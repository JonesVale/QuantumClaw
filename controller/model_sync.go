package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"gorm.io/gorm"
)

// 模型同步设置
type ModelSyncSetting struct {
	Enabled        bool     `json:"enabled"`
	SourceURL      string   `json:"source_url"`
	SyncInterval   int      `json:"sync_interval"`   // 分钟
	LastSyncTime   int64    `json:"last_sync_time"`
	SyncOnStartup  bool     `json:"sync_on_startup"`
	AllowedChannels []string `json:"allowed_channels"` // 只同步这些渠道的模型
}

var modelSyncSetting = &ModelSyncSetting{
	Enabled:       false,
	SourceURL:     "https://basellm.github.io/llm-metadata/models.json",
	SyncInterval:  1440, // 每天一次
	SyncOnStartup: true,
}

// ModelMetadataEntry 从上游获取的模型元数据条目
type ModelMetadataEntry struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Provider  string  `json:"provider"`
	Pricing   *PricingEntry `json:"pricing,omitempty"`
	Contexts  int     `json:"context_length,omitempty"`
}

type PricingEntry struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

type ModelMetadataResponse struct {
	Models []ModelMetadataEntry `json:"models"`
}

// GetModelSyncSetting 获取模型同步设置
func GetModelSyncSetting(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    modelSyncSetting,
	})
}

// SaveModelSyncSetting 保存模型同步设置
func SaveModelSyncSetting(c *gin.Context) {
	var req ModelSyncSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误: " + err.Error()})
		return
	}
	modelSyncSetting.Enabled = req.Enabled
	modelSyncSetting.SourceURL = strings.TrimSpace(req.SourceURL)
	modelSyncSetting.SyncInterval = req.SyncInterval
	modelSyncSetting.SyncOnStartup = req.SyncOnStartup
	if req.AllowedChannels != nil {
		modelSyncSetting.AllowedChannels = req.AllowedChannels
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": modelSyncSetting})
}

// SyncModels 手动同步模型元数据
func SyncModels(c *gin.Context) {
	if !modelSyncSetting.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "模型同步未启用"})
		return
	}
	res, err := doModelSync()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "同步失败: " + err.Error()})
		return
	}
	modelSyncSetting.LastSyncTime = time.Now().Unix()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("同步完成，新增 %d，更新 %d，跳过 %d", res.Added, res.Updated, res.Skipped),
		"data":    res,
	})
}

// ModelSyncResult 同步结果
type ModelSyncResult struct {
	Added   int `json:"added"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
}

func doModelSync() (*ModelSyncResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelSyncSetting.SourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "QuantumClaw/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("上游返回状态码 %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var metaResp ModelMetadataResponse
	if err := json.Unmarshal(body, &metaResp); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	result := &ModelSyncResult{}
	for _, entry := range metaResp.Models {
		r, err := syncSingleModel(entry)
		if err != nil {
			logger.SysLogf("模型同步跳过 %s: %v", entry.ID, err)
			result.Skipped++
			continue
		}
		if r == "added" {
			result.Added++
		} else if r == "updated" {
			result.Updated++
		}
	}
	return result, nil
}

func syncSingleModel(entry ModelMetadataEntry) (string, error) {
	if strings.TrimSpace(entry.ID) == "" {
		return "", fmt.Errorf("模型 ID 为空")
	}

	// 检查是否需要过滤
	if len(modelSyncSetting.AllowedChannels) > 0 {
		allowed := false
		for _, ch := range modelSyncSetting.AllowedChannels {
			if strings.EqualFold(strings.TrimSpace(ch), strings.TrimSpace(entry.Provider)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("provider %s 不在允许列表中", entry.Provider)
		}
	}

	// 查找或创建模型
	modelName := strings.TrimSpace(entry.ID)
	var existing model.Channel
	err := model.DB.Where("name = ?", modelName).First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 新增
		priority := int64(1)
		newChannel := model.Channel{
			Name:     modelName,
			Group:    "default",
			Models:   modelName,
			Key:      fmt.Sprintf("sync://%s", entry.Provider),
			Status:   model.ChannelStatusEnabled,
			Priority: &priority,
			TestTime: time.Now().Unix(),
		}
		if err := model.DB.Create(&newChannel).Error; err != nil {
			return "", err
		}
		return "added", nil
	} else if err != nil {
		return "", err
	}

	// 定价由 ratio 包全局管理，这里仅记录 key/source 信息变化
	updated := false
	newKey := fmt.Sprintf("sync://%s", entry.Provider)
	if existing.Key != newKey {
		existing.Key = newKey
		updated = true
	}
	if updated {
		if err := model.DB.Save(&existing).Error; err != nil {
			return "", err
		}
		return "updated", nil
	}
	return "", nil
}

// GetModelSyncStatus 获取模型同步状态
func GetModelSyncStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled":        modelSyncSetting.Enabled,
			"last_sync_time": modelSyncSetting.LastSyncTime,
			"source_url":     modelSyncSetting.SourceURL,
			"sync_interval":  modelSyncSetting.SyncInterval,
		},
	})
}

// DEPRECATED: 曾被设计为由 main.go 的 cron 调用，但从未实际接入。
// 未来如需开启，请在 main.go 中添加调用。当前为死代码。
func StartModelSyncCron() {
	if !modelSyncSetting.Enabled || modelSyncSetting.SyncInterval <= 0 {
		return
	}
	interval := time.Duration(modelSyncSetting.SyncInterval) * time.Minute
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if modelSyncSetting.Enabled {
				res, err := doModelSync()
				if err != nil {
					logger.SysError("模型同步定时任务失败: " + err.Error())
				} else if res.Added > 0 || res.Updated > 0 {
					logger.SysLogf("模型同步定时任务完成: 新增=%d 更新=%d", res.Added, res.Updated)
				}
			}
		}
	}()
}

// ==================== 模型别名管理 ====================

type ModelAlias struct {
	Alias string `json:"alias"` // 用户请求的模型名
	Real  string `json:"real"`  // 实际路由到的模型
	Group string `json:"group"` // 所属分组
}

// GetModelAliasList 获取模型别名列表
func GetModelAliasList(c *gin.Context) {
	// 从缓存或数据库获取
	aliases := getModelAliasesFromDB()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    aliases,
	})
}

type ModelAliasRequest struct {
	Alias string `json:"alias"`
	Real  string `json:"real"`
	Group string `json:"group"`
}

// UpsertModelAlias 创建或更新模型别名
func UpsertModelAlias(c *gin.Context) {
	var req ModelAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	req.Alias = strings.TrimSpace(req.Alias)
	req.Real = strings.TrimSpace(req.Real)
	if req.Alias == "" || req.Real == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "别名和真实模型名不能为空"})
		return
	}
	upsertModelAlias(req.Alias, req.Real, req.Group)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func upsertModelAlias(alias, real, group string) {
	// 存储到 Option 表
	data, _ := json.Marshal(map[string]string{
		"alias": alias,
		"real":  real,
		"group": group,
	})
	model.DB.Exec("INSERT OR REPLACE INTO options (key, value) VALUES (?, ?)",
		"model_alias:"+alias, string(data))
}

func getModelAliasesFromDB() []ModelAlias {
	var opts []model.Option
	model.DB.Where("`key` LIKE ?", "model_alias:%").Find(&opts)
	result := make([]ModelAlias, 0, len(opts))
	for _, opt := range opts {
		var m ModelAlias
		if err := json.Unmarshal([]byte(opt.Value), &m); err == nil {
			result = append(result, m)
		}
	}
	return result
}

// ResolveModelAlias 解析模型别名
func ResolveModelAlias(modelName string) string {
	if modelName == "" {
		return ""
	}
	var opt model.Option
	if err := model.DB.Where("`key` = ?", "model_alias:"+modelName).First(&opt).Error; err != nil {
		return modelName // 未找到别名，原样返回
	}
	var m ModelAlias
	if err := json.Unmarshal([]byte(opt.Value), &m); err != nil {
		return modelName
	}
	return m.Real
}

// ==================== 模型搜索（支持正则）====================

// SearchModels 搜索模型
func SearchModels(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	provider := strings.TrimSpace(c.Query("provider"))
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var channels []model.Channel
	db := model.DB.Where("status = ?", model.ChannelStatusEnabled)
	if query != "" {
		if isRegexPattern(query) {
			// 正则搜索
			db = db.Where("name REGEXP ?", query)
		} else {
			// 模糊搜索
			db = db.Where("name LIKE ?", "%"+query+"%")
		}
	}
	if provider != "" {
		db = db.Where("type = ?", provider)
	}
	db = db.Order("priority DESC, id DESC").Limit(limit).Find(&channels)

	result := make([]map[string]interface{}, 0, len(channels))
	for _, ch := range channels {
		result = append(result, map[string]interface{}{
			"id":     ch.Id,
			"name":   ch.Name,
			"type":   ch.Type,
			"group":  ch.Group,
			"status": ch.Status,
			"models": ch.Models,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"count":   len(result),
	})
}

func isRegexPattern(s string) bool {
	_, err := regexp.Compile(s)
	return err == nil
}
