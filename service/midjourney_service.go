package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// MidjourneyService Midjourney API 代理服务
type MidjourneyService struct {
	BaseURL string
	APIKey  string
	ChannelID int
}

// NewMidjourneyService 创建 Midjourney 服务实例
func NewMidjourneyService(baseURL string, apiKey string, channelID int) *MidjourneyService {
	return &MidjourneyService{
		BaseURL: baseURL,
		APIKey:  apiKey,
		ChannelID: channelID,
	}
}

// MidjourneyRequest Midjourney API 通用请求结构
type MidjourneyRequest struct {
	Prompt      string   `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Style        string   `json:"style,omitempty"`         // raw/4a/4b
	Quality      string   `json:"quality,omitempty"`       // 0.25/0.5/1
	AspectRatio  string   `json:"aspect_ratio,omitempty"` // 1:1/16:9/2:3/3:2/4:5/5:4/9:16
	Seed         int      `json:"seed,omitempty"`
	ProcessMode  string   `json:"process_mode,omitempty"` // fast/relax/turbo
	Upscale     bool     `json:"upscale,omitempty"`
	Index       int      `json:"index,omitempty"`         // Upscale/Vary 时用的序号 1-4
	TaskID      string   `json:"task_id,omitempty"`      // Upscale/Vary 时的父任务ID
}

// MidjourneyResponse Midjourney API 通用响应结构
type MidjourneyResponse struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	TaskID    string          `json:"task_id"`             // 上游返回的任务ID
	Status    string          `json:"status"`
	Progress  int             `json:"progress"`
	ImageURL  string          `json:"image_url"`
	VideoURL  string          `json:"video_url"`
	Buttons   json.RawMessage `json:"buttons,omitempty"`  // 操作按钮JSON
}

// SubmitImagine 提交 Imagine 任务
func (s *MidjourneyService) SubmitImagine(prompt string, promptEn string, options map[string]interface{}) (string, error) {
	reqBody := MidjourneyRequest{
		Prompt: promptEn, // 使用英文提示词调用API
	}

	// 填充可选参数
	if v, ok := options["style"]; ok {
		reqBody.Style = v.(string)
	}
	if v, ok := options["quality"]; ok {
		reqBody.Quality = v.(string)
	}
	if v, ok := options["aspect_ratio"]; ok {
		reqBody.AspectRatio = v.(string)
	}
	if v, ok := options["seed"]; ok {
		reqBody.Seed = v.(int)
	}
	if v, ok := options["process_mode"]; ok {
		reqBody.ProcessMode = v.(string)
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	url := s.BaseURL + "/mj/submit/imagine"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var mjResp MidjourneyResponse
	err = json.Unmarshal(body, &mjResp)
	if err != nil {
		return "", fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if mjResp.Code != 0 {
		return "", fmt.Errorf("Midjourney API 错误: %s", mjResp.Message)
	}

	logger.Infof(nil, "[MidjourneyService] Imagine 任务已提交: mjTaskID=%s", mjResp.TaskID)
	return mjResp.TaskID, nil
}

// SubmitUpscale 提交 Upscale 任务
func (s *MidjourneyService) SubmitUpscale(parentTaskID string, index int) (string, error) {
	reqBody := MidjourneyRequest{
		TaskID: parentTaskID,
		Index:  index,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	url := s.BaseURL + "/mj/submit/upscale"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var mjResp MidjourneyResponse
	err = json.Unmarshal(body, &mjResp)
	if err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if mjResp.Code != 0 {
		return "", fmt.Errorf("Midjourney API 错误: %s", mjResp.Message)
	}

	logger.Infof(nil, "[MidjourneyService] Upscale 任务已提交: mjTaskID=%s, parent=%s, index=%d", mjResp.TaskID, parentTaskID, index)
	return mjResp.TaskID, nil
}

// SubmitVary 提交 Vary 任务（区域/缩放/变换）
func (s *MidjourneyService) SubmitVary(parentTaskID string, index int, varyType string) (string, error) {
	reqBody := MidjourneyRequest{
		TaskID: parentTaskID,
		Index:  index,
	}

	url := s.BaseURL + "/mj/submit/vary"
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var mjResp MidjourneyResponse
	err = json.Unmarshal(body, &mjResp)
	if err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if mjResp.Code != 0 {
		return "", fmt.Errorf("Midjourney API 错误: %s", mjResp.Message)
	}

	logger.Infof(nil, "[MidjourneyService] Vary 任务已提交: mjTaskID=%s, parent=%s", mjResp.TaskID, parentTaskID)
	return mjResp.TaskID, nil
}

// GetTaskStatus 查询 Midjourney 任务状态
func (s *MidjourneyService) GetTaskStatus(mjTaskID string) (*MidjourneyResponse, error) {
	url := s.BaseURL + "/mj/task/" + mjTaskID + "/fetch"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var mjResp MidjourneyResponse
	err = json.Unmarshal(body, &mjResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	return &mjResp, nil
}

// ProcessMidjourneyTask 处理 Midjourney 任务的完整流程
// 1. 根据 AsyncTask 记录提交到 Midjourney API
// 2. 轮询任务状态直到完成
// 3. 更新数据库中的任务状态和结果
func ProcessMidjourneyTask(task *model.AsyncTask, mjTask *model.MidjourneyTask, channelConfig map[string]string) error {
	ctx := context.Background()

	service := NewMidjourneyService(
		channelConfig["base_url"],
		channelConfig["api_key"],
		task.ChannelID,
	)

	// 标记任务开始处理
	taskService := NewTaskService()
	taskService.MarkTaskProcessing(task.TaskID)

	var mjTaskID string
	var err error

	switch task.Action {
	case "imagine":
		mjTaskID, err = service.SubmitImagine(mjTask.Prompt, mjTask.PromptEn, nil)
	case "upscale":
		// 从 RequestData 中解析父任务ID和index
		var reqData map[string]interface{}
		if task.GetRequestData(&reqData) == nil {
			parentTaskID := reqData["parent_task_id"].(string)
			index := int(reqData["index"].(float64))
			mjTaskID, err = service.SubmitUpscale(parentTaskID, index)
		}
	case "vary":
		var reqData map[string]interface{}
		if task.GetRequestData(&reqData) == nil {
			parentTaskID := reqData["parent_task_id"].(string)
			index := int(reqData["index"].(float64))
			mjTaskID, err = service.SubmitVary(parentTaskID, index, "")
		}
	default:
		return fmt.Errorf("未知的动作: %s", task.Action)
	}

	if err != nil {
		taskService.MarkTaskFailed(task.TaskID, err.Error())
		return err
	}

	// 更新 MJ 任务记录中的 MJ ID
	model.UpdateMidjourneyTask(mjTaskID, map[string]interface{}{
		"mj_id": mjTaskID,
	})

	// 轮询任务状态（最多轮询60次，每次间隔10秒）
	for i := 0; i < 60; i++ {
		time.Sleep(10 * time.Second)

		statusResp, err := service.GetTaskStatus(mjTaskID)
		if err != nil {
			logger.Warnf(ctx, "[ProcessMidjourneyTask] 轮询状态失败: %v", err)
			continue
		}

		// 更新进度
		taskService.UpdateTaskProgress(task.TaskID, statusResp.Progress, statusResp.ImageURL, "", "")

		// 检查是否完成
		if statusResp.Status == "finished" || statusResp.Status == "success" {
			// 更新 MJ 任务记录
			model.UpdateMidjourneyTask(mjTaskID, map[string]interface{}{
				"image_url":   statusResp.ImageURL,
				"video_url":   statusResp.VideoURL,
				"buttons":     string(statusResp.Buttons),
				"finish_time": time.Now().Unix(),
			})
			logger.Infof(ctx, "[ProcessMidjourneyTask] 任务完成: taskID=%s, imageURL=%s", task.TaskID, statusResp.ImageURL)
			return nil
		}

		if statusResp.Status == "failed" {
			taskService.MarkTaskFailed(task.TaskID, statusResp.Message)
			return fmt.Errorf("Midjourney 任务失败: %s", statusResp.Message)
		}
	}

	// 超时
	taskService.MarkTaskFailed(task.TaskID, "任务处理超时")
	return fmt.Errorf("任务处理超时: %s", task.TaskID)
}
