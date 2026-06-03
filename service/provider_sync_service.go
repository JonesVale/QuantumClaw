package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// ProviderSyncService 定期从各供应商拉取最新模型列表
// 更新 channel_providers 表的 Models 字段 + last_synced 时间戳
type ProviderSyncService struct {
	interval time.Duration
}

// NewProviderSyncService 创建同步服务
// interval: 同步周期（默认 24h）
func NewProviderSyncService(interval time.Duration) *ProviderSyncService {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	return &ProviderSyncService{interval: interval}
}

// Start 启动后台同步循环
func (s *ProviderSyncService) Start() {
	go func() {
		// 启动时立即执行一次
		s.syncOnce()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for range ticker.C {
			s.syncOnce()
		}
	}()
	logger.SysLog(fmt.Sprintf("provider sync service started (interval: %v)", s.interval))
}

// syncOnce 执行一次全量同步
func (s *ProviderSyncService) syncOnce() {
	var providers []model.ChannelProvider
	if err := model.DB.Where("auto_sync = ? AND base_url != ''", true).Find(&providers).Error; err != nil {
		logger.SysError("provider sync: failed to query providers: " + err.Error())
		return
	}

	synced := 0
	for _, p := range providers {
		if err := s.syncProvider(&p); err != nil {
			logger.SysWarn(fmt.Sprintf("provider sync failed: type_id=%d (%s): %v", p.TypeID, p.Name, err))
			continue
		}
		synced++
	}

	if synced > 0 {
		logger.SysLog(fmt.Sprintf("provider sync completed: %d/%d providers synced", synced, len(providers)))
	}
}

// syncProvider 同步单个供应商的模型列表
func (s *ProviderSyncService) syncProvider(p *model.ChannelProvider) error {
	// FetchModelsFromProvider 需要渠道对象（有 Key 和 BaseURL）
	// 这里我们只更新 model 列表，不需要 API Key
	// 使用供应商的 BaseURL 查询 /v1/models
	models, err := fetchProviderModels(p.BaseURL)
	if err != nil {
		return fmt.Errorf("fetch models: %w", err)
	}

	if len(models) == 0 {
		return fmt.Errorf("empty model list from %s", p.BaseURL)
	}

	// 序列化模型列表
	modelsJSON := "[]"
	if b, err := json.Marshal(models); err == nil {
		modelsJSON = string(b)
	}

	now := time.Now().Unix()
	updates := map[string]interface{}{
		"models":      modelsJSON,
		"last_synced": now,
		"updated_at":  now,
	}
	if err := model.DB.Model(p).Updates(updates).Error; err != nil {
		return fmt.Errorf("update models: %w", err)
	}

	logger.SysLog(fmt.Sprintf("provider sync: %s (%s) -> %d models", p.Name, p.BaseURL, len(models)))
	return nil
}

// fetchProviderModels 向供应商的 /v1/models 发请求获取模型列表
// 不需要 API Key（公共端点或使用供应商公开信息）
func fetchProviderModels(baseURL string) ([]string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := strings.TrimRight(baseURL, "/")
	if !strings.Contains(url, "/v1") {
		url += "/v1"
	}
	url += "/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "QuantumClaw/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	return models, nil
}
