package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
)

// ═══════════════════════════════════════════════════════════════════════════
// 用户端：获取当前生效协议
// ═══════════════════════════════════════════════════════════════════════════

// GetActiveAgreement 获取当前生效的批量发送协议
func GetActiveAgreement(c *gin.Context) {
	a, err := model.GetActiveAgreement()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "暂无可用协议"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": a})
}

// ═══════════════════════════════════════════════════════════════════════════
// 用户端：发送任务管理
// ═══════════════════════════════════════════════════════════════════════════

// CreateMessageJob 创建批量发送任务（用户同意协议后调用）
func CreateMessageJob(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)

	var req struct {
		Channel          string `json:"channel" binding:"required"`
		TotalTargets     int    `json:"total_targets" binding:"required"`
		AgreementVersion string `json:"agreement_version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误：channel、total_targets、agreement_version 不能为空"})
		return
	}

	// 校验通道类型
	if req.Channel != "sms" && req.Channel != "wechat" && req.Channel != "email" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "不支持的发送通道"})
		return
	}

	// 校验目标数
	if req.TotalTargets <= 0 || req.TotalTargets > 2000 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "目标数量无效（1-2000）"})
		return
	}

	// 校验协议版本
	if _, err := model.GetActiveAgreement(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请先同意用户协议"})
		return
	}

	job := &model.MessageJob{
		UserId:           userId,
		Channel:          req.Channel,
		BatchLimit:       20, // 默认每批上限
		TotalTargets:     req.TotalTargets,
		Status:           "running",
		AgreementVersion: req.AgreementVersion,
		ConsentTime:      time.Now().Unix(),
	}

	if err := model.CreateMessageJob(job); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": job})
}

// GetMessageJobs 获取当前用户的发送任务列表
func GetMessageJobs(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	jobs, err := model.GetUserMessageJobs(userId, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取任务列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": jobs})
}

// GetMessageJobDetail 获取任务详情（含统计数据）
func GetMessageJobDetail(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	id, _ := strconv.Atoi(c.Param("id"))

	job, err := model.GetMessageJob(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "任务不存在"})
		return
	}

	if job.UserId != userId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无权查看此任务"})
		return
	}

	sent, failed, pending, _ := model.GetMessageJobStats(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"job": job,
			"stats": gin.H{
				"sent":    sent,
				"failed":  failed,
				"pending": pending,
			},
		},
	})
}

// UpdateMessageJobProgress 客户端更新发送进度（断点续传）
func UpdateMessageJobProgress(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		SentCount    int `json:"sent_count"`
		FailCount    int `json:"fail_count"`
		CurrentBatch int `json:"current_batch"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	job, err := model.GetMessageJob(id)
	if err != nil || job.UserId != userId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "任务不存在"})
		return
	}

	if err := model.UpdateMessageJobProgress(id, req.SentCount, req.FailCount, req.CurrentBatch); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "更新进度失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// CompleteMessageJob 客户端标记任务完成
func CompleteMessageJob(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	id, _ := strconv.Atoi(c.Param("id"))

	if err := model.UpdateMessageJobStatusWithUserCheck(id, userId, "completed"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// PauseMessageJob 暂停发送任务
func PauseMessageJob(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	id, _ := strconv.Atoi(c.Param("id"))

	if err := model.UpdateMessageJobStatusWithUserCheck(id, userId, "paused"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// ResumeMessageJob 恢复发送任务
func ResumeMessageJob(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	id, _ := strconv.Atoi(c.Param("id"))

	if err := model.UpdateMessageJobStatusWithUserCheck(id, userId, "running"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// CancelMessageJob 取消发送任务
func CancelMessageJob(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	id, _ := strconv.Atoi(c.Param("id"))

	if err := model.UpdateMessageJobStatusWithUserCheck(id, userId, "cancelled"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// ═══════════════════════════════════════════════════════════════════════════
// 发送记录
// ═══════════════════════════════════════════════════════════════════════════

// BatchCreateMessageLogs 批量记录发送结果
func BatchCreateMessageLogs(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)

	var req struct {
		JobId int              `json:"job_id" binding:"required"`
		Logs  []model.MessageLog `json:"logs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	// 校验任务归属
	job, err := model.GetMessageJob(req.JobId)
	if err != nil || job.UserId != userId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "任务不存在"})
		return
	}

	for i := range req.Logs {
		req.Logs[i].JobId = req.JobId
		req.Logs[i].UserId = userId
		now := time.Now()
		if req.Logs[i].Status == "sent" {
			req.Logs[i].SentAt = &now
		}
	}

	if err := model.BatchCreateMessageLogs(req.Logs); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "记录发送结果失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// GetMessageJobLogs 获取任务的发送记录明细
func GetMessageJobLogs(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	jobId, _ := strconv.Atoi(c.Param("id"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	job, err := model.GetMessageJob(jobId)
	if err != nil || job.UserId != userId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "任务不存在"})
		return
	}

	logs, err := model.GetJobMessageLogs(jobId, offset, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取记录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": logs})
}

// ═══════════════════════════════════════════════════════════════════════════
// 管理员：协议管理
// ═══════════════════════════════════════════════════════════════════════════

// AdminCreateAgreement 管理员创建/更新协议版本
func AdminCreateAgreement(c *gin.Context) {
	role := c.GetInt(ctxkey.Role)
	if role < 100 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "需要超级管理员权限"})
		return
	}

	var req struct {
		Version string `json:"version" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		Channel string `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	a := &model.MessageAgreement{
		Version:  req.Version,
		Title:    req.Title,
		Content:  req.Content,
		Channel:  req.Channel,
		IsActive: true,
	}

	if err := model.CreateMessageAgreement(a); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建协议失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": a})
}
