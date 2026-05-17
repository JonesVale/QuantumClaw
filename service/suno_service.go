package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// SunoService Suno 音乐生成 API 服务
type SunoService struct {
	BaseURL string
	APIKey  string
	Model   string // bureaudeep/Breeze|Breeze-2
}

// NewSunoService 创建 Suno 服务实例
func NewSunoService(baseURL string, apiKey string, modelName string) *SunoService {
	return &SunoService{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
	}
}

// SunoSubmitRequest Suno 任务提交请求
type SunoSubmitRequest struct {
	Action string `json:"action"`           // song/lyrics/description-mode
	Title  string `json:"title,omitempty"`  // 歌曲标题
	Lyrics string `json:"lyrics,omitempty"` // 歌词
	Tags   string `json:"tags,omitempty"`    // 风格标签
	Model  string `json:"model,omitempty"`   // 模型名称
}

// SunoTaskResponse Suno 任务响应
type SunoTaskResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	TaskID   string `json:"task_id"`      // 上游任务ID
	Status   string `json:"status"`        // pending/processing/success/failed
	Progress int    `json:"progress"`      // 进度 0-100
	AudioURL string `json:"audio_url"`     // 音频URL（完成后）
}

// SubmitSunoTask 提交 Suno 音乐生成任务
func (s *SunoService) SubmitSunoTask(action string, title string, lyrics string) (string, error) {
	reqBody := SunoSubmitRequest{
		Action: action,
		Title:  title,
		Lyrics: lyrics,
		Model:  s.Model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	url := s.BaseURL + "/suno/submit"
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

	var sunoResp SunoTaskResponse
	err = json.Unmarshal(body, &sunoResp)
	if err != nil {
		return "", fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if sunoResp.Code != 0 {
		return "", fmt.Errorf("Suno API 错误: %s", sunoResp.Message)
	}

	logger.Infof(nil, "[SunoService] 音乐生成任务已提交: taskID=%s", sunoResp.TaskID)
	return sunoResp.TaskID, nil
}

// GetSunoTaskStatus 查询 Suno 任务状态
func (s *SunoService) GetSunoTaskStatus(taskID string) (*SunoTaskResponse, error) {
	url := s.BaseURL + "/suno/task/" + taskID

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

	var sunoResp SunoTaskResponse
	err = json.Unmarshal(body, &sunoResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	return &sunoResp, nil
}

// ProcessSunoTask 处理 Suno 音乐生成任务的完整流程
func ProcessSunoTask(taskID string, baseURL string, apiKey string) error {
	// 获取 Suno 任务记录
	sunoTask, err := model.GetSunoTaskByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("获取Suno任务记录失败: %w", err)
	}

	// 创建 Suno 服务
	service := NewSunoService(baseURL, apiKey, sunoTask.Model)

	// 标记任务开始处理
	taskService := NewTaskService()
	taskService.MarkTaskProcessing(taskID)

	// 提交任务
	sunoTaskID, err := service.SubmitSunoTask(sunoTask.Action, sunoTask.Title, sunoTask.Lyrics)
	if err != nil {
		taskService.MarkTaskFailed(taskID, err.Error())
		return err
	}

	// 更新任务记录中的外部ID
	err = model.UpdateSunoTask(taskID, map[string]interface{}{
		"suno_task_id": sunoTaskID,
	})
	if err != nil {
		logger.Warnf(nil, "[ProcessSunoTask] 更新外部ID失败: %v", err)
	}

	// 轮询任务状态（最多轮询120次，每次间隔10秒，总共20分钟）
	for i := 0; i < 120; i++ {
		time.Sleep(10 * time.Second)

		statusResp, err := service.GetSunoTaskStatus(sunoTaskID)
		if err != nil {
			logger.Warnf(nil, "[ProcessSunoTask] 轮询状态失败: %v", err)
			continue
		}

		// 更新进度
		taskService.UpdateTaskProgress(taskID, statusResp.Progress, "", "", statusResp.AudioURL)

		// 检查是否完成
		if statusResp.Status == "success" || statusResp.Status == "finished" {
			// 更新 Suno 任务记录
			err = model.UpdateSunoTask(taskID, map[string]interface{}{
				"audio_url":   statusResp.AudioURL,
				"finish_time": time.Now().Unix(),
			})
			if err != nil {
				logger.Warnf(nil, "[ProcessSunoTask] 更新音频URL失败: %v", err)
			}

			logger.Infof(nil, "[ProcessSunoTask] 任务完成: taskID=%s, audioURL=%s", taskID, statusResp.AudioURL)
			return nil
		}

		if statusResp.Status == "failed" {
			taskService.MarkTaskFailed(taskID, statusResp.Message)
			return fmt.Errorf("Suno 任务失败: %s", statusResp.Message)
		}
	}

	// 超时
	taskService.MarkTaskFailed(taskID, "Suno 任务处理超时")
	return fmt.Errorf("Suno 任务处理超时: %s", taskID)
}
