package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// FetchModelsFromProvider 用渠道的 API Key 查询供应商的 /v1/models 端点
// 获取该供应商支持的所有模型列表
func (channel *Channel) FetchModelsFromProvider() ([]string, error) {
	if channel.Key == "" || strings.HasPrefix(channel.Key, "PUT_YOUR") {
		return nil, fmt.Errorf("无效的 API Key")
	}
	baseURL := channel.GetBaseURL()
	if baseURL == "" {
		return nil, fmt.Errorf("未设置 Base URL")
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.Contains(baseURL, "/v1") {
		baseURL += "/v1"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+channel.Key)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回 %d: %s", resp.StatusCode, string(body))
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

// UpdateModelsFromProvider 从供应商拉取最新模型列表并更新到数据库
// 配置 API Key 后自动调用，确保模型列表始终是最新的
func (channel *Channel) UpdateModelsFromProvider() error {
	models, err := channel.FetchModelsFromProvider()
	if err != nil {
		logger.SysWarn(fmt.Sprintf("从供应商拉取模型失败（channel #%d %s）: %v", channel.Id, channel.Name, err))
		return err
	}
	if len(models) == 0 {
		return fmt.Errorf("供应商返回的模型列表为空")
	}

	channel.Models = strings.Join(models, ",")
	err = DB.Model(channel).Update("models", channel.Models).Error
	if err != nil {
		return err
	}
	logger.SysLog(fmt.Sprintf("从供应商拉取模型成功（channel #%d %s）: %d 个模型", channel.Id, channel.Name, len(models)))
	return nil
}
