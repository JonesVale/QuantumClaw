package service

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
)

// TaskService 异步任务服务
type TaskService struct {
	// 可扩展：添加任务队列、Worker池等
}

// NewTaskService 创建任务服务实例
func NewTaskService() *TaskService {
	return &TaskService{}
}

// CreateMidjourneyTask 创建 Midjourney 异步任务
// 返回：内部任务ID、外部任务ID、错误
func (s *TaskService) CreateMidjourneyTask(userID int, channelID int, group string, action string, prompt string, promptEn string) (int64, string, error) {
	// 生成内部任务ID
	taskID := generateTaskID()

	// 创建通用任务记录
	task := &model.AsyncTask{
		TaskID:     taskID,
		Platform:   model.PlatformMidjourney,
		UserID:     userID,
		ChannelID:  channelID,
		Group:      group,
		Action:     action,
		Status:     model.TaskStatusPending,
		Progress:   0,
		SubmitTime: time.Now().Unix(),
	}
	err := model.CreateAsyncTask(task)
	if err != nil {
		return 0, "", fmt.Errorf("创建任务失败: %w", err)
	}

	// 创建 Midjourney 专用任务记录
	mjTask := &model.MidjourneyTask{
		TaskID:   taskID,
		Prompt:   prompt,
		PromptEn: promptEn,
	}
	err = model.CreateMidjourneyTask(mjTask)
	if err != nil {
		return 0, "", fmt.Errorf("创建MJ任务失败: %w", err)
	}

	logger.Infof(nil, "[TaskService] 创建Midjourney任务: taskID=%s, userID=%d, action=%s", taskID, userID, action)
	return task.ID, taskID, nil
}

// CreateVideoTask 创建视频生成异步任务
func (s *TaskService) CreateVideoTask(userID int, channelID int, group string, modelName string, prompt string, imageURL string) (int64, string, error) {
	taskID := generateTaskID()

	task := &model.AsyncTask{
		TaskID:     taskID,
		Platform:   model.PlatformVideo,
		UserID:     userID,
		ChannelID:  channelID,
		Group:      group,
		Action:     "generate",
		Status:     model.TaskStatusPending,
		Progress:   0,
		SubmitTime: time.Now().Unix(),
	}
	err := model.CreateAsyncTask(task)
	if err != nil {
		return 0, "", fmt.Errorf("创建视频任务失败: %w", err)
	}

	// 创建视频专用任务记录
	videoTask := &model.VideoTask{
		TaskID:    taskID,
		Model:     modelName,
		Prompt:    prompt,
		ImageURL:  imageURL,
	}
	err = model.CreateVideoTask(videoTask)
	if err != nil {
		return 0, "", fmt.Errorf("创建视频任务记录失败: %w", err)
	}

	logger.Infof(nil, "[TaskService] 创建视频任务: taskID=%s, userID=%d, model=%s", taskID, userID, modelName)
	return task.ID, taskID, nil
}

// CreateSunoTask 创建 Suno 音乐生成异步任务
func (s *TaskService) CreateSunoTask(userID int, channelID int, group string, action string, title string, lyrics string, modelName string) (int64, string, error) {
	taskID := generateTaskID()

	task := &model.AsyncTask{
		TaskID:     taskID,
		Platform:   model.PlatformSuno,
		UserID:     userID,
		ChannelID:  channelID,
		Group:      group,
		Action:     action,
		Status:     model.TaskStatusPending,
		Progress:   0,
		SubmitTime: time.Now().Unix(),
	}
	err := model.CreateAsyncTask(task)
	if err != nil {
		return 0, "", fmt.Errorf("创建Suno任务失败: %w", err)
	}

	// 创建Suno专用任务记录
	sunoTask := &model.SunoTask{
		TaskID: taskID,
		Action: action,
		Title:  title,
		Lyrics: lyrics,
		Model:  modelName,
	}
	err = model.CreateSunoTask(sunoTask)
	if err != nil {
		return 0, "", fmt.Errorf("创建Suno任务记录失败: %w", err)
	}

	logger.Infof(nil, "[TaskService] 创建Suno任务: taskID=%s, userID=%d, action=%s", taskID, userID, action)
	return task.ID, taskID, nil
}

// GetTaskStatus 获取任务状态
func (s *TaskService) GetTaskStatus(taskID string) (*model.AsyncTask, error) {
	return model.GetAsyncTaskByTaskID(taskID)
}

// GetUserTasks 获取用户的任务列表
func (s *TaskService) GetUserTasks(userID int, platform string, status string, page int, pageSize int) ([]*model.AsyncTask, error) {
	startIdx := (page - 1) * pageSize
	return model.GetUserTasks(userID, platform, status, startIdx, pageSize)
}

// UpdateTaskProgress 更新任务进度
func (s *TaskService) UpdateTaskProgress(taskID string, progress int, imageURL string, videoURL string, audioURL string) error {
	updates := map[string]interface{}{
		"progress": progress,
	}
	if imageURL != "" {
		updates["image_url"] = imageURL
	}
	if videoURL != "" {
		updates["video_url"] = videoURL
	}
	if audioURL != "" {
		updates["audio_url"] = audioURL
	}

	// 如果进度达到100%，自动标记成功
	if progress >= 100 {
		updates["status"] = model.TaskStatusSuccess
		updates["finish_time"] = time.Now().Unix()
	}

	return model.UpdateAsyncTaskStatus(taskID, model.AsyncTaskStatus(updates["status"].(string)), progress, "")
}

// MarkTaskFailed 标记任务失败
func (s *TaskService) MarkTaskFailed(taskID string, failReason string) error {
	return model.UpdateAsyncTaskStatus(taskID, model.TaskStatusFailed, 0, failReason)
}

// MarkTaskProcessing 标记任务开始处理
func (s *TaskService) MarkTaskProcessing(taskID string) error {
	return model.UpdateAsyncTaskStatus(taskID, model.TaskStatusProcessing, 0, "")
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(taskID string) error {
	return model.DeleteAsyncTask(taskID)
}

// PollPendingTasks 轮询未完成的任务（供后台goroutine调用）
// 返回处理中的任务数量
func (s *TaskService) PollPendingTasks() int {
	tasks, err := model.GetAllUnfinishedTasks()
	if err != nil {
		logger.Errorf(nil, "[TaskService] 获取未完成任务失败: %v", err)
		return 0
	}

	count := 0
	for _, task := range tasks {
		// 根据平台调用不同的轮询逻辑
		switch task.Platform {
		case model.PlatformMidjourney:
			s.pollMidjourneyTask(task)
		case model.PlatformVideo:
			s.pollVideoTask(task)
		case model.PlatformSuno:
			s.pollSunoTask(task)
		default:
			logger.Warnf(nil, "[TaskService] 未知平台: %s", task.Platform)
		}
		count++
	}
	return count
}

// pollMidjourneyTask 轮询Midjourney任务状态
func (s *TaskService) pollMidjourneyTask(task *model.AsyncTask) {
	// 获取渠道配置
	channel, err := model.GetChannelById(task.ChannelID, true)
	if err != nil {
		logger.Warnf(nil, "[TaskService] 获取渠道配置失败: %v", err)
		return
	}

	// 创建 Midjourney 服务
	mjService := NewMidjourneyService(*channel.BaseURL, channel.Key, task.ChannelID)

	// 获取 MJ 任务记录
	mjTask, err := model.GetMidjourneyTaskByTaskID(task.TaskID)
	if err != nil {
		logger.Warnf(nil, "[TaskService] 获取MJ任务记录失败: %v", err)
		return
	}

	// 如果还没有上游任务ID，跳过
	if mjTask.MJID == "" {
		logger.Debugf(nil, "[TaskService] MJ任务还没有上游ID，跳过轮询: taskID=%s", task.TaskID)
		return
	}

	// 查询任务状态
	statusResp, err := mjService.GetTaskStatus(mjTask.MJID)
	if err != nil {
		logger.Warnf(nil, "[TaskService] 轮询MJ任务状态失败: %v", err)
		return
	}

	// 更新进度
	s.UpdateTaskProgress(task.TaskID, statusResp.Progress, statusResp.ImageURL, statusResp.VideoURL, "")

	// 检查是否完成
	if statusResp.Status == "finished" || statusResp.Status == "success" {
		// 更新 MJ 任务记录
		model.UpdateMidjourneyTask(mjTask.MJID, map[string]interface{}{
			"image_url":   statusResp.ImageURL,
			"video_url":   statusResp.VideoURL,
			"buttons":     string(statusResp.Buttons),
			"finish_time": time.Now().Unix(),
		})
		logger.Infof(nil, "[TaskService] MJ任务完成: taskID=%s, imageURL=%s", task.TaskID, statusResp.ImageURL)
		return
	}

	if statusResp.Status == "failed" {
		s.MarkTaskFailed(task.TaskID, statusResp.Message)
		return
	}
}

// pollVideoTask 轮询视频生成任务状态
func (s *TaskService) pollVideoTask(task *model.AsyncTask) {
	// 获取渠道配置
	channel, err := model.GetChannelById(task.ChannelID, true)
	if err != nil {
		logger.Warnf(nil, "[TaskService] 获取渠道配置失败: %v", err)
		return
	}

	// 获取视频任务记录
	videoTask, err := model.GetVideoTaskByTaskID(task.TaskID)
	if err != nil {
		logger.Warnf(nil, "[TaskService] 获取视频任务记录失败: %v", err)
		return
	}

	// 如果还没有上游任务ID，跳过
	if videoTask.PlatformTaskID == "" {
		logger.Debugf(nil, "[TaskService] 视频任务还没有上游ID，跳过轮询: taskID=%s", task.TaskID)
		return
	}

	// 创建视频服务
	videoService := NewVideoService(*channel.BaseURL, channel.Key, videoTask.Model)

	// 查询任务状态
	statusResp, err := videoService.GetVideoTaskStatus(videoTask.PlatformTaskID)
	if err != nil {
		logger.Warnf(nil, "[TaskService] 轮询视频任务状态失败: %v", err)
		return
	}

	// 更新进度
	s.UpdateTaskProgress(task.TaskID, statusResp.Progress, "", statusResp.VideoURL, "")

	// 检查是否完成
	if statusResp.Status == "finished" || statusResp.Status == "success" {
		// 更新视频任务记录
		model.UpdateVideoTask(task.TaskID, map[string]interface{}{
			"video_url":   statusResp.VideoURL,
			"finish_time": time.Now().Unix(),
		})
		logger.Infof(nil, "[TaskService] 视频任务完成: taskID=%s, videoURL=%s", task.TaskID, statusResp.VideoURL)
		return
	}

	if statusResp.Status == "failed" {
		s.MarkTaskFailed(task.TaskID, statusResp.Message)
		return
	}
}

// pollSunoTask 轮询Suno任务状态
func (s *TaskService) pollSunoTask(task *model.AsyncTask) {
	// 获取渠道配置
	channel, err := model.GetChannelById(task.ChannelID, true)
	if err != nil {
		logger.Warnf(nil, "[TaskService] 获取渠道配置失败: %v", err)
		return
	}

	// 获取 Suno 任务记录
	sunoTask, err := model.GetSunoTaskByTaskID(task.TaskID)
	if err != nil {
		logger.Warnf(nil, "[TaskService] 获取Suno任务记录失败: %v", err)
		return
	}

	// 如果还没有上游任务ID，跳过
	if sunoTask.SunoTaskID == "" {
		logger.Debugf(nil, "[TaskService] Suno任务还没有上游ID，跳过轮询: taskID=%s", task.TaskID)
		return
	}

	// 创建 Suno 服务
	sunoService := NewSunoService(*channel.BaseURL, channel.Key, sunoTask.Model)

	// 查询任务状态
	statusResp, err := sunoService.GetSunoTaskStatus(sunoTask.SunoTaskID)
	if err != nil {
		logger.Warnf(nil, "[TaskService] 轮询Suno任务状态失败: %v", err)
		return
	}

	// 更新进度
	s.UpdateTaskProgress(task.TaskID, statusResp.Progress, "", "", statusResp.AudioURL)

	// 检查是否完成
	if statusResp.Status == "finished" || statusResp.Status == "success" {
		// 更新 Suno 任务记录
		model.UpdateSunoTask(task.TaskID, map[string]interface{}{
			"audio_url":   statusResp.AudioURL,
			"finish_time": time.Now().Unix(),
		})
		logger.Infof(nil, "[TaskService] Suno任务完成: taskID=%s, audioURL=%s", task.TaskID, statusResp.AudioURL)
		return
	}

	if statusResp.Status == "failed" {
		s.MarkTaskFailed(task.TaskID, statusResp.Message)
		return
	}
}

// generateTaskID 生成唯一的任务ID
func generateTaskID() string {
	// 使用时间戳+随机数生成唯一ID
	// 格式：task_时间戳_随机数
	return fmt.Sprintf("task_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// GetTaskByID 根据数据库ID获取任务
func (s *TaskService) GetTaskByID(id int64) (*model.AsyncTask, error) {
	return model.GetAsyncTaskByID(id)
}

// GetMidjourneyTaskDetail 获取Midjourney任务详情
func (s *TaskService) GetMidjourneyTaskDetail(taskID string) (*model.MidjourneyTask, error) {
	// 这里需要根据taskID查询midjourney_tasks表
	// 暂时返回nil，需要补充具体实现
	return nil, fmt.Errorf("暂未实现")
}

// EstimateTaskQuota 估算任务消耗的配额
func (s *TaskService) EstimateTaskQuota(platform model.AsyncTaskPlatform, action string) int {
	// 根据平台和动作估算配额消耗
	switch platform {
	case model.PlatformMidjourney:
		switch action {
		case "imagine":
			return 10 // 假设imagine消耗10配额
		case "upscale":
			return 5
		case "vary":
			return 5
		default:
			return 5
		}
	case model.PlatformVideo:
		return 50 // 视频生成消耗更多
	case model.PlatformSuno:
		return 20
	default:
		return 5
	}
}

// UpdateTaskQuota 更新任务消耗的配额
func (s *TaskService) UpdateTaskQuota(taskID string, quota int) error {
	return model.DB.Model(&model.AsyncTask{}).Where("task_id = ?", taskID).Update("quota", quota).Error
}

// GetRelayModeForPlatform 根据平台获取对应的relay模式
func (s *TaskService) GetRelayModeForPlatform(platform model.AsyncTaskPlatform) int {
	switch platform {
	case model.PlatformMidjourney:
		return relaymode.Midjourney
	case model.PlatformVideo:
		return relaymode.VideoGeneration
	case model.PlatformSuno:
		return relaymode.Suno
	default:
		return relaymode.Unknown
	}
}
