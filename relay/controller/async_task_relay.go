package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
	"github.com/quantumclaw/quantumclaw/service"
)

// RelayAsyncTask 异步任务中继入口
// 处理 /mj/ /video/ /suno/ 开头的异步任务请求
func RelayAsyncTask(c *gin.Context) {
	relayMode := relaymode.GetByPath(c.Request.URL.Path)

	switch relayMode {
	case relaymode.Midjourney:
		RelayMidjourneyTask(c)
	case relaymode.VideoGeneration:
		RelayVideoTask(c)
	case relaymode.Suno:
		RelaySunoTask(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的异步任务类型"})
	}
}

// RelayMidjourneyTask 处理 Midjourney 异步任务
func RelayMidjourneyTask(c *gin.Context) {
	ctx := c.Request.Context()
	path := c.Request.URL.Path
	userID := c.GetInt("id") // 从TokenAuth中间件获取
	channelID := c.GetInt("channel_id")
	group := c.GetString("group")
	if group == "" {
		group = "default"
	}

	// 判断动作类型
	action := parseMidjourneyAction(path)

	// 读取请求体
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败: " + err.Error()})
		return
	}

	// 解析提示词
	var req map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "解析请求体失败: " + err.Error()})
			return
		}
	}

	prompt := ""
	if v, ok := req["prompt"].(string); ok {
		prompt = v
	}
	promptEn := ""
	if v, ok := req["prompt_en"].(string); ok {
		promptEn = v
	}
	if promptEn == "" {
		// 未提供英文 prompt，使用中文 prompt 直接传给 Midjourney
		// Midjourney 会尝试理解，但中文 prompt 可能影响出图质量
		// 如需翻译，可接入 TranslateService 接口（待实现）
		promptEn = prompt
	}

	// 创建任务
	taskService := service.NewTaskService()
	_, taskID, err := taskService.CreateMidjourneyTask(userID, channelID, group, action, prompt, promptEn)
	if err != nil {
		logger.Errorf(ctx, "[RelayMidjourneyTask] 创建任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 保存请求体到任务的 RequestData
	task, _ := model.GetAsyncTaskByTaskID(taskID)
	if task != nil {
		task.SetRequestData(req)
		model.DB.Model(&model.AsyncTask{}).Where("task_id = ?", taskID).Update("request_data", string(body))
	}

	// 异步处理任务（在后台goroutine中调用Midjourney API）
	go processMidjourneyTaskAsync(taskID, channelID, req)

	// 立即返回任务ID给客户端
	c.JSON(http.StatusOK, gin.H{
		"task_id":  taskID,
		"status":    string(model.TaskStatusPending),
		"message":  "任务已提交，请使用 task_id 轮询状态",
	})
}

// RelayVideoTask 处理视频生成异步任务
func RelayVideoTask(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")
	channelID := c.GetInt("channel_id")
	group := c.GetString("group")
	if group == "" {
		group = "default"
	}

	// 读取请求体
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败: " + err.Error()})
		return
	}

	var req map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "解析请求体失败: " + err.Error()})
			return
		}
	}

	modelName := ""
	if v, ok := req["model"].(string); ok {
		modelName = v
	}
	prompt := ""
	if v, ok := req["prompt"].(string); ok {
		prompt = v
	}
	imageURL := ""
	if v, ok := req["image_url"].(string); ok {
		imageURL = v
	}

	// 创建任务
	taskService := service.NewTaskService()
	_, taskID, err := taskService.CreateVideoTask(userID, channelID, group, modelName, prompt, imageURL)
	if err != nil {
		logger.Errorf(ctx, "[RelayVideoTask] 创建任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 异步处理任务
	go processVideoTaskAsync(taskID, channelID, modelName, req)

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  string(model.TaskStatusPending),
		"message": "视频生成任务已提交，请轮询状态",
	})
}

// RelaySunoTask 处理 Suno 音乐生成异步任务
func RelaySunoTask(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")
	channelID := c.GetInt("channel_id")
	group := c.GetString("group")
	if group == "" {
		group = "default"
	}

	// 读取请求体
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败: " + err.Error()})
		return
	}

	var req map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "解析请求体失败: " + err.Error()})
			return
		}
	}

	action := "song"
	if v, ok := req["action"].(string); ok {
		action = v
	}
	title := ""
	if v, ok := req["title"].(string); ok {
		title = v
	}
	lyrics := ""
	if v, ok := req["lyrics"].(string); ok {
		lyrics = v
	}
	modelName := "bureaudeep/Breeze"
	if v, ok := req["model"].(string); ok {
		modelName = v
	}

	// 创建任务
	taskService := service.NewTaskService()
	_, taskID, err := taskService.CreateSunoTask(userID, channelID, group, action, title, lyrics, modelName)
	if err != nil {
		logger.Errorf(ctx, "[RelaySunoTask] 创建任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 异步处理任务
	go processSunoTaskAsync(taskID, channelID, modelName, req)

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  string(model.TaskStatusPending),
		"message": "音乐生成任务已提交，请轮询状态",
	})
}

// ==================== 后台异步处理 ====================

// processMidjourneyTaskAsync 后台异步处理 Midjourney 任务
func processMidjourneyTaskAsync(taskID string, channelID int, req map[string]interface{}) {
	ctx := context.Background()
	taskService := service.NewTaskService()

	taskService.MarkTaskProcessing(taskID)

	channel, err := model.GetChannelById(channelID, true)
	if err != nil {
		taskService.MarkTaskFailed(taskID, "获取渠道配置失败: "+err.Error())
		return
	}

	mjService := service.NewMidjourneyService(*channel.BaseURL, channel.Key, channelID)

	mjTask, err := model.GetMidjourneyTaskByTaskID(taskID)
	if err != nil {
		taskService.MarkTaskFailed(taskID, "获取MJ任务记录失败: "+err.Error())
		return
	}

	action := parseMidjourneyActionFromReq(req)
	var mjTaskID string

	switch action {
	case "imagine":
		mjTaskID, err = mjService.SubmitImagine(mjTask.Prompt, mjTask.PromptEn, nil)
	case "upscale":
		parentTaskID := ""
		index := 1
		if v, ok := req["parent_task_id"].(string); ok {
			parentTaskID = v
		}
		if v, ok := req["index"].(float64); ok {
			index = int(v)
		}
		mjTaskID, err = mjService.SubmitUpscale(parentTaskID, index)
	case "vary":
		parentTaskID := ""
		index := 1
		if v, ok := req["parent_task_id"].(string); ok {
			parentTaskID = v
		}
		if v, ok := req["index"].(float64); ok {
			index = int(v)
		}
		mjTaskID, err = mjService.SubmitVary(parentTaskID, index, "")
	default:
		taskService.MarkTaskFailed(taskID, "不支持的动作: "+action)
		return
	}

	if err != nil {
		taskService.MarkTaskFailed(taskID, "提交Midjourney任务失败: "+err.Error())
		return
	}

	model.UpdateMidjourneyTask(mjTaskID, map[string]interface{}{
		"mj_id": mjTaskID,
	})

	logger.Infof(ctx, "[processMidjourneyTaskAsync] Midjourney任务已提交: taskID=%s, mjTaskID=%s, action=%s", taskID, mjTaskID, action)
}

// processVideoTaskAsync 后台异步处理视频生成任务
func processVideoTaskAsync(taskID string, channelID int, modelName string, req map[string]interface{}) {
	ctx := context.Background()
	taskService := service.NewTaskService()

	taskService.MarkTaskProcessing(taskID)

	channel, err := model.GetChannelById(channelID, true)
	if err != nil {
		taskService.MarkTaskFailed(taskID, "获取渠道配置失败: "+err.Error())
		return
	}

	videoService := service.NewVideoService(*channel.BaseURL, channel.Key, modelName)

	videoTask, err := model.GetVideoTaskByTaskID(taskID)
	if err != nil {
		taskService.MarkTaskFailed(taskID, "获取视频任务记录失败: "+err.Error())
		return
	}

	prompt := videoTask.Prompt
	if v, ok := req["prompt"].(string); ok && v != "" {
		prompt = v
	}

	imageURL := videoTask.ImageURL
	if v, ok := req["image_url"].(string); ok && v != "" {
		imageURL = v
	}

	platformTaskID, err := videoService.SubmitVideoTask(prompt, imageURL)
	if err != nil {
		taskService.MarkTaskFailed(taskID, "提交视频生成任务失败: "+err.Error())
		return
	}

	model.UpdateVideoTask(taskID, map[string]interface{}{
		"platform_task_id": platformTaskID,
	})

	logger.Infof(ctx, "[processVideoTaskAsync] 视频生成任务已提交: taskID=%s, platformTaskID=%s, model=%s", taskID, platformTaskID, modelName)
}

// processSunoTaskAsync 后台异步处理 Suno 音乐生成任务
func processSunoTaskAsync(taskID string, channelID int, modelName string, req map[string]interface{}) {
	ctx := context.Background()
	taskService := service.NewTaskService()

	taskService.MarkTaskProcessing(taskID)

	channel, err := model.GetChannelById(channelID, true)
	if err != nil {
		taskService.MarkTaskFailed(taskID, "获取渠道配置失败: "+err.Error())
		return
	}

	sunoService := service.NewSunoService(*channel.BaseURL, channel.Key, modelName)

	sunoTask, err := model.GetSunoTaskByTaskID(taskID)
	if err != nil {
		taskService.MarkTaskFailed(taskID, "获取Suno任务记录失败: "+err.Error())
		return
	}

	action := sunoTask.Action
	title := sunoTask.Title
	lyrics := sunoTask.Lyrics

	if v, ok := req["action"].(string); ok && v != "" {
		action = v
	}
	if v, ok := req["title"].(string); ok && v != "" {
		title = v
	}
	if v, ok := req["lyrics"].(string); ok && v != "" {
		lyrics = v
	}

	sunoTaskID, err := sunoService.SubmitSunoTask(action, title, lyrics)
	if err != nil {
		taskService.MarkTaskFailed(taskID, "提交Suno任务失败: "+err.Error())
		return
	}

	model.UpdateSunoTask(taskID, map[string]interface{}{
		"suno_task_id": sunoTaskID,
	})

	logger.Infof(ctx, "[processSunoTaskAsync] Suno音乐生成任务已提交: taskID=%s, sunoTaskID=%s, action=%s", taskID, sunoTaskID, action)
}

// ==================== 辅助函数 ====================

// parseMidjourneyAction 根据请求路径解析 Midjourney 动作类型
func parseMidjourneyAction(path string) string {
	if strings.Contains(path, "/submit/imagine") {
		return "imagine"
	}
	if strings.Contains(path, "/submit/upscale") {
		return "upscale"
	}
	if strings.Contains(path, "/submit/vary") {
		return "vary"
	}
	if strings.Contains(path, "/submit/describe") {
		return "describe"
	}
	if strings.Contains(path, "/submit/blend") {
		return "blend"
	}
	return "imagine" // 默认
}

// parseMidjourneyActionFromReq 从请求数据中解析Midjourney动作类型
func parseMidjourneyActionFromReq(req map[string]interface{}) string {
	action := ""
	if v, ok := req["action"].(string); ok {
		action = v
	}
	if action != "" {
		return action
	}
	return "imagine" // 默认
}
