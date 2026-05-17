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

// VideoService 视频生成 API 服务
type VideoService struct {
	BaseURL string
	APIKey  string
	Model   string // kling/jimeng/runway/pika
}

// NewVideoService 创建视频生成服务实例
func NewVideoService(baseURL string, apiKey string, modelName string) *VideoService {
	return &VideoService{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
	}
}

// VideoSubmitRequest 视频生成提交请求
type VideoSubmitRequest struct {
	Model    string `json:"model,omitempty"`     // 模型名称
	Prompt   string `json:"prompt"`              // 提示词
	ImageURL string `json:"image_url,omitempty"`  // 参考图片
	Duration int    `json:"duration,omitempty"`   // 视频时长（秒）
	Ratio    string `json:"ratio,omitempty"`      // 宽高比 16:9/9:16/1:1
}

// VideoTaskResponse 视频任务响应
type VideoTaskResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	TaskID    string `json:"task_id"`     // 上游任务ID
	Status    string `json:"status"`       // pending/processing/success/failed
	Progress  int    `json:"progress"`     // 进度 0-100
	VideoURL  string `json:"video_url"`    // 视频URL（完成后）
}

// SubmitVideoTask 提交视频生成任务
func (s *VideoService) SubmitVideoTask(prompt string, imageURL string) (string, error) {
	reqBody := VideoSubmitRequest{
		Model:    s.Model,
		Prompt:   prompt,
		ImageURL: imageURL,
		Duration: 5, // 默认5秒
		Ratio:    "16:9",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 根据模型选择不同的 API 端点
	url := s.BaseURL + "/video/generate"
	
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

	var videoResp VideoTaskResponse
	err = json.Unmarshal(body, &videoResp)
	if err != nil {
		return "", fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if videoResp.Code != 0 {
		return "", fmt.Errorf("视频生成 API 错误: %s", videoResp.Message)
	}

	logger.Infof(nil, "[VideoService] 视频任务已提交: taskID=%s", videoResp.TaskID)
	return videoResp.TaskID, nil
}

// GetVideoTaskStatus 查询视频任务状态
func (s *VideoService) GetVideoTaskStatus(taskID string) (*VideoTaskResponse, error) {
	url := s.BaseURL + "/video/task/" + taskID

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

	var videoResp VideoTaskResponse
	err = json.Unmarshal(body, &videoResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	return &videoResp, nil
}

// ProcessVideoTask 处理视频生成任务的完整流程
func ProcessVideoTask(taskID string, baseURL string, apiKey string) error {
	// 获取视频任务记录
	videoTask, err := model.GetVideoTaskByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("获取视频任务记录失败: %w", err)
	}

	// 创建视频服务
	service := NewVideoService(baseURL, apiKey, videoTask.Model)

	// 标记任务开始处理
	taskService := NewTaskService()
	taskService.MarkTaskProcessing(taskID)

	// 提交任务
	platformTaskID, err := service.SubmitVideoTask(videoTask.Prompt, videoTask.ImageURL)
	if err != nil {
		taskService.MarkTaskFailed(taskID, err.Error())
		return err
	}

	// 更新任务记录中的外部ID
	err = model.UpdateVideoTask(taskID, map[string]interface{}{
		"platform_task_id": platformTaskID,
	})
	if err != nil {
		logger.Warnf(nil, "[ProcessVideoTask] 更新外部ID失败: %v", err)
	}

	// 轮询任务状态（最多轮询120次，每次间隔10秒，总共20分钟）
	for i := 0; i < 120; i++ {
		time.Sleep(10 * time.Second)

		statusResp, err := service.GetVideoTaskStatus(platformTaskID)
		if err != nil {
			logger.Warnf(nil, "[ProcessVideoTask] 轮询状态失败: %v", err)
			continue
		}

		// 更新进度
		taskService.UpdateTaskProgress(taskID, statusResp.Progress, "", statusResp.VideoURL, "")

		// 检查是否完成
		if statusResp.Status == "success" || statusResp.Status == "finished" {
			// 更新视频任务记录
			err = model.UpdateVideoTask(taskID, map[string]interface{}{
				"video_url":   statusResp.VideoURL,
				"finish_time": time.Now().Unix(),
			})
			if err != nil {
				logger.Warnf(nil, "[ProcessVideoTask] 更新视频URL失败: %v", err)
			}

			logger.Infof(nil, "[ProcessVideoTask] 任务完成: taskID=%s, videoURL=%s", taskID, statusResp.VideoURL)
			return nil
		}

		if statusResp.Status == "failed" {
			taskService.MarkTaskFailed(taskID, statusResp.Message)
			return fmt.Errorf("视频生成任务失败: %s", statusResp.Message)
		}
	}

	// 超时
	taskService.MarkTaskFailed(taskID, "视频生成任务处理超时")
	return fmt.Errorf("视频生成任务处理超时: %s", taskID)
}
