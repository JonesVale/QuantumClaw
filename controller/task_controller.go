package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
	"gorm.io/gorm"
)

// ==================== Midjourney 任务接口 ====================

// CreateMidjourneyTask 创建 Midjourney 任务
// POST /api/task/midjourney
func CreateMidjourneyTask(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt(ctxkey.Id)

	var req struct {
		Action   string `json:"action" binding:"required"`   // imagine/upscale/vary/describe/blend
		Prompt   string `json:"prompt" binding:"required"`
		PromptEn string `json:"prompt_en"`                    // 英文提示词（可选，不传则自动翻译）
		// Upscale/Vary 专用
		TaskID   string `json:"task_id"`                      // 父任务ID（upscale/vary时用）
		Index    int    `json:"index"`                         // 哪张图（1-4）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 获取渠道信息
	channelID := c.GetInt(ctxkey.ChannelId)
	group := c.GetString(ctxkey.Group)
	if group == "" {
		group = "default"
	}

	// 创建任务
	taskService := service.NewTaskService()
	_, taskID, err := taskService.CreateMidjourneyTask(userID, channelID, group, req.Action, req.Prompt, req.PromptEn)
	if err != nil {
		logger.Errorf(ctx, "[TaskController] 创建MJ任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 异步调用 Midjourney API 提交任务
	go func() {
		// 获取渠道配置
		channel, err := model.GetChannelById(channelID, true)
		if err != nil {
			logger.Errorf(nil, "[TaskController] 获取渠道配置失败: %v", err)
			return
		}
		
		// 获取任务记录
		task, err := model.GetAsyncTaskByTaskID(taskID)
		if err != nil {
			logger.Errorf(nil, "[TaskController] 获取任务记录失败: %v", err)
			return
		}
		
		// 获取 MJ 任务记录
		mjTask, err := model.GetMidjourneyTaskByMJID(taskID)
		if err != nil {
			logger.Errorf(nil, "[TaskController] 获取MJ任务记录失败: %v", err)
			return
		}
		
		// 构建渠道配置 map
		channelConfig := map[string]string{
			"base_url": *channel.BaseURL,
			"api_key":  channel.Key,
		}
		
		// 处理任务
		err = service.ProcessMidjourneyTask(task, mjTask, channelConfig)
		if err != nil {
			logger.Errorf(nil, "[TaskController] 处理MJ任务失败: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  string(model.TaskStatusPending),
		"message": "任务已创建，正在处理中",
	})
}

// GetMidjourneyTask 获取 Midjourney 任务状态
// GET /api/task/midjourney/:task_id
func GetMidjourneyTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id 不能为空"})
		return
	}

	taskService := service.NewTaskService()
	task, err := taskService.GetTaskStatus(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 获取 Midjourney 专用信息
	mjTask, err := model.GetMidjourneyTaskByMJID(taskID)
	if err != nil {
		// 如果不关联，只返回通用任务信息
		c.JSON(http.StatusOK, gin.H{
			"task_id":  task.TaskID,
			"status":   string(task.Status),
			"progress": task.Progress,
			"fail_reason": task.FailReason,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":     task.TaskID,
		"mj_id":       mjTask.MJID,
		"status":       string(task.Status),
		"progress":     task.Progress,
		"prompt":       mjTask.Prompt,
		"prompt_en":    mjTask.PromptEn,
		"image_url":    task.ImageURL,
		"fail_reason":  task.FailReason,
		"created_at":   task.CreatedAt,
		"finished_at":  task.FinishTime,
	})
}

// ==================== 通用任务接口 ====================

// GetTaskStatus 获取任意任务状态
// GET /api/task/:task_id
func GetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id 不能为空"})
		return
	}

	taskService := service.NewTaskService()
	task, err := taskService.GetTaskStatus(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":      task.TaskID,
		"platform":     string(task.Platform),
		"status":       string(task.Status),
		"progress":     task.Progress,
		"image_url":    task.ImageURL,
		"video_url":    task.VideoURL,
		"audio_url":    task.AudioURL,
		"fail_reason":  task.FailReason,
		"created_at":   task.CreatedAt,
		"finished_at":  task.FinishTime,
	})
}

// ListUserTasks 获取当前用户的任务列表
// GET /api/tasks?platform=&status=&page=1&page_size=20
func ListUserTasks(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt(ctxkey.Id)

	platform := c.Query("platform")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	taskService := service.NewTaskService()
	tasks, err := taskService.GetUserTasks(userID, platform, status, page, pageSize)
	if err != nil {
		logger.Errorf(ctx, "[TaskController] 获取用户任务列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为响应格式
	var result []gin.H
	for _, task := range tasks {
		result = append(result, gin.H{
			"id":           task.ID,
			"task_id":      task.TaskID,
			"platform":     string(task.Platform),
			"action":       task.Action,
			"status":       string(task.Status),
			"progress":     task.Progress,
			"quota":        task.Quota,
			"image_url":    task.ImageURL,
			"video_url":    task.VideoURL,
			"audio_url":    task.AudioURL,
			"fail_reason":  task.FailReason,
			"created_at":   task.CreatedAt,
			"finished_at":  task.FinishTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       result,
		"page":       page,
		"page_size":  pageSize,
	})
}

// CancelTask 取消任务
// POST /api/task/:task_id/cancel
func CancelTask(c *gin.Context) {
	ctx := c.Request.Context()
	taskID := c.Param("task_id")
	userID := c.GetInt(ctxkey.Id)

	task, err := model.GetAsyncTaskByTaskID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 检查权限：只能取消自己的任务
	if task.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此任务"})
		return
	}

	// 只能取消未完成的任务
	if task.IsFinished() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务已完成，无法取消"})
		return
	}

	taskService := service.NewTaskService()
	err = taskService.MarkTaskFailed(taskID, "用户取消")
	if err != nil {
		logger.Errorf(ctx, "[TaskController] 取消任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已取消"})
}

// DeleteTask 删除任务
// DELETE /api/task/:task_id
func DeleteTask(c *gin.Context) {
	ctx := c.Request.Context()
	taskID := c.Param("task_id")
	userID := c.GetInt(ctxkey.Id)

	task, err := model.GetAsyncTaskByTaskID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 检查权限：只能删除自己的任务
	if task.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此任务"})
		return
	}

	taskService := service.NewTaskService()
	err = taskService.DeleteTask(taskID)
	if err != nil {
		logger.Errorf(ctx, "[TaskController] 删除任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已删除"})
}

// ==================== 管理员接口 ====================

// AdminGetAllTasks 管理员获取所有任务
// GET /api/admin/task?platform=&status=&page=1&page_size=20
func AdminGetAllTasks(c *gin.Context) {
	ctx := c.Request.Context()

	platform := c.Query("platform")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	startIdx := (page - 1) * pageSize

	// 构建查询
	query := model.DB.Model(&model.AsyncTask{})
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var tasks []*model.AsyncTask
	err := query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&tasks).Error
	if err != nil {
		logger.Errorf(ctx, "[TaskController] 管理员获取任务列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       tasks,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// AdminPollTasks 管理员触发任务轮询（手动）
// POST /api/admin/task/poll
func AdminPollTasks(c *gin.Context) {
	ctx := c.Request.Context()

	taskService := new(service.TaskService)
	count := taskService.PollPendingTasks()

	logger.Infof(ctx, "[TaskController] 管理员手动触发任务轮询，处理了 %d 个任务", count)

	c.JSON(http.StatusOK, gin.H{
		"message": "轮询完成",
		"count":   count,
	})
}

// ==================== 视频生成任务接口 ====================

// CreateVideoTask 创建视频生成任务
// POST /api/task/video
func CreateVideoTask(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt(ctxkey.Id)

	var req struct {
		Model    string `json:"model" binding:"required"` // kling/jimeng
		Prompt   string `json:"prompt" binding:"required"`
		ImageURL string `json:"image_url"`                 // 参考图片（可选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	channelID := c.GetInt(ctxkey.ChannelId)
	group := c.GetString(ctxkey.Group)
	if group == "" {
		group = "default"
	}

	taskService := new(service.TaskService)
	_, taskID, err := taskService.CreateVideoTask(userID, channelID, group, req.Model, req.Prompt, req.ImageURL)
	if err != nil {
		logger.Errorf(ctx, "[TaskController] 创建视频任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 异步调用视频生成 API
	go func() {
		// 获取渠道配置
		channel, err := model.GetChannelById(channelID, true)
		if err != nil {
			logger.Errorf(nil, "[TaskController] 获取渠道配置失败: %v", err)
			return
		}
		
		// 处理视频生成任务
		err = service.ProcessVideoTask(taskID, *channel.BaseURL, channel.Key)
		if err != nil {
			logger.Errorf(nil, "[TaskController] 处理视频任务失败: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  string(model.TaskStatusPending),
		"message": "视频生成任务已创建，正在处理中",
	})
}

// ==================== Suno 音乐生成任务接口 ====================

// CreateSunoTask 创建 Suno 音乐生成任务
// POST /api/task/suno
func CreateSunoTask(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt(ctxkey.Id)

	var req struct {
		Action string `json:"action" binding:"required"` // song/lyrics/description-mode
		Title  string `json:"title"`
		Lyrics string `json:"lyrics"`
		Model  string `json:"model"` // bureaudeep/Breeze|Breeze-2
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	channelID := c.GetInt(ctxkey.ChannelId)
	group := c.GetString(ctxkey.Group)
	if group == "" {
		group = "default"
	}

	taskService := new(service.TaskService)
	_, taskID, err := taskService.CreateSunoTask(userID, channelID, group, req.Action, req.Title, req.Lyrics, req.Model)
	if err != nil {
		logger.Errorf(ctx, "[TaskController] 创建Suno任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 异步调用 Suno API
	go func() {
		// 获取渠道配置
		channel, err := model.GetChannelById(channelID, true)
		if err != nil {
			logger.Errorf(nil, "[TaskController] 获取渠道配置失败: %v", err)
			return
		}
		
		// 处理 Suno 任务
		err = service.ProcessSunoTask(taskID, *channel.BaseURL, channel.Key)
		if err != nil {
			logger.Errorf(nil, "[TaskController] 处理Suno任务失败: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  string(model.TaskStatusPending),
		"message": "音乐生成任务已创建，正在处理中",
	})
}

// ==================== 批量为任务补充配额 ====================

// DeductTaskQuota 扣除任务配额（在任务完成时调用）
// 使用原子 UPDATE 防止并发丢失更新
func DeductTaskQuota(userID int, taskID string, quota int) error {
	// 原子扣款：只有配额足够时才扣
	result := model.DB.Model(&model.User{}).Where("id = ? AND quota >= ?", userID, quota).
		Updates(map[string]interface{}{
			"quota":      gorm.Expr("quota - ?", quota),
			"used_quota": gorm.Expr("used_quota + ?", quota),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("配额不足")
	}

	// 更新任务记录的配额消耗
	taskService := new(service.TaskService)
	return taskService.UpdateTaskQuota(taskID, quota)
}

// ==================== 辅助函数 ====================

// 获取任务平台的中文描述
func getPlatformName(platform model.AsyncTaskPlatform) string {
	switch platform {
	case model.PlatformMidjourney:
		return "Midjourney"
	case model.PlatformVideo:
		return "视频生成"
	case model.PlatformSuno:
		return "Suno音乐"
	default:
		return string(platform)
	}
}

// 获取任务状态的中文描述
func getStatusName(status model.AsyncTaskStatus) string {
	switch status {
	case model.TaskStatusPending:
		return "等待中"
	case model.TaskStatusQueued:
		return "已排队"
	case model.TaskStatusProcessing:
		return "处理中"
	case model.TaskStatusSuccess:
		return "成功"
	case model.TaskStatusFailed:
		return "失败"
	case model.TaskStatusCancelled:
		return "已取消"
	default:
		return string(status)
	}
}
