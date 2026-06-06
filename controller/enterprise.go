package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/model"
)

// ==================== 部门管理 ====================

// CreateDepartment 创建部门
func CreateDepartment(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil || orgId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}

	// 校验权限：组织管理员或超级管理员
	if !model.IsOrganizationAdmin(userId, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织管理员可创建部门"})
		return
	}

	var req struct {
		Name           string `json:"name" binding:"required"`
		Description    string `json:"description"`
		HeadUserId     int    `json:"head_user_id"`
		MonthlyBudget  int64  `json:"monthly_budget"`
		AlertThreshold int64  `json:"alert_threshold"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "部门名称不能为空"})
		return
	}

	d := &model.Department{
		OrgId:          orgId,
		Name:           req.Name,
		Description:    req.Description,
		HeadUserId:     req.HeadUserId,
		MonthlyBudget:  req.MonthlyBudget,
		AlertThreshold: req.AlertThreshold,
	}
	if err := model.CreateDepartment(d); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建部门失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": d})
}

// ListDepartments 获取组织的部门列表
func ListDepartments(c *gin.Context) {
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil || orgId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}

	depts, err := model.GetOrgDepartments(orgId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	type DeptWithCount struct {
		model.Department
		MemberCount int64 `json:"member_count"`
	}
	result := make([]DeptWithCount, 0, len(depts))
	for _, d := range depts {
		count, _ := model.GetDepartmentMemberCount(d.Id)
		result = append(result, DeptWithCount{Department: d, MemberCount: count})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// UpdateDepartment 编辑部门
func UpdateDepartment(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的部门ID"})
		return
	}

	dept, err := model.GetDepartmentById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "部门不存在"})
		return
	}

	if !model.IsOrganizationAdmin(userId, dept.OrgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织管理员可编辑部门"})
		return
	}

	var req struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		HeadUserId     int    `json:"head_user_id"`
		MonthlyBudget  int64  `json:"monthly_budget"`
		AlertThreshold int64  `json:"alert_threshold"`
		Status         int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if req.Name != "" {
		dept.Name = req.Name
	}
	dept.Description = req.Description
	dept.HeadUserId = req.HeadUserId
	dept.MonthlyBudget = req.MonthlyBudget
	dept.AlertThreshold = req.AlertThreshold
	dept.Status = req.Status

	if err := model.UpdateDepartment(dept); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已更新", "data": dept})
}

// DeleteDepartment 删除部门
func DeleteDepartment(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的部门ID"})
		return
	}

	dept, err := model.GetDepartmentById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "部门不存在"})
		return
	}
	if !model.IsOrganizationAdmin(userId, dept.OrgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织管理员可删除部门"})
		return
	}

	if err := model.DeleteDepartment(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "删除失败"})
		return
	}
	// 清除该部门成员的 department_id
	model.DB.Model(&model.User{}).Where("department_id = ?", id).Update("department_id", 0)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已删除"})
}

// ==================== 组织升级 ====================

// UpgradeOrganization 升级组织层级 (personal → enterprise / provider)
func UpgradeOrganization(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	orgId, err := strconv.Atoi(c.Param("id"))
	if err != nil || orgId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}

	org, err := model.GetOrganizationByID(orgId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "组织不存在"})
		return
	}
	if org.OwnerId != userId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织创建者可升级"})
		return
	}

	var req struct {
		Tier string `json:"tier" binding:"required"` // enterprise / provider
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Tier != "enterprise" && req.Tier != "provider") {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的升级类型"})
		return
	}

	if err := model.DB.Model(org).Update("tier", req.Tier).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "升级失败"})
		return
	}
	model.RecordLog(c.Request.Context(), userId, model.LogTypeSystem, "组织升级为"+req.Tier)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "组织已升级为 " + req.Tier})
}

// ==================== 企业策略 ====================

func GetEnterprisePolicy(c *gin.Context) {
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}
	policy, err := model.GetEnterprisePolicy(orgId)
	if err != nil {
		// 返回默认值
		policy = &model.EnterpriseTokenPolicy{OrgId: orgId}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": policy})
}

func SaveEnterprisePolicy(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}
	if !model.IsOrganizationAdmin(userId, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织管理员可修改策略"})
		return
	}

	var p model.EnterpriseTokenPolicy
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	p.OrgId = orgId
	if err := model.SaveEnterprisePolicy(&p); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "策略已保存"})
}

// ==================== 员工管理（组织内） ====================

// SetDepartment 设置员工所属部门
func SetEmployeeDepartment(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}
	if !model.IsOrganizationAdmin(userId, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织管理员可操作"})
		return
	}

	var req struct {
		UserId       int `json:"user_id" binding:"required"`
		DepartmentId int `json:"department_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", req.UserId).Update("department_id", req.DepartmentId).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已更新"})
}

// ==================== 企业用量查询 ====================

// GetOrgUsage 组织用量概览
func GetOrgUsage(c *gin.Context) {
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}

	dateFrom := c.DefaultQuery("from", "2026-06-01")
	summary, err := model.GetOrgUsageSummary(orgId, dateFrom)
	if err != nil {
		summary = &model.EnterpriseUsageStat{OrgId: orgId}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": summary})
}

// ==================== 企业 Token 创建 ====================

// CreateEnterpriseToken 创建企业Token（含策略约束和审批）
// POST /api/org/:orgId/tokens
func CreateEnterpriseToken(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil || orgId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}

	var req struct {
		UserId       int    `json:"user_id" binding:"required"`  // 持有人
		DepartmentId int    `json:"department_id"`                // 所属部门
		Name         string `json:"name" binding:"required"`      // Token名称
		Purpose      string `json:"purpose"`                      // 用途说明
		Label        string `json:"label"`                        // 标签: production/test/dev
		RemainQuota  int64  `json:"remain_quota"`                 // 额度 (0=用策略默认)
		ExpiredTime  int64  `json:"expired_time"`                 // 过期时间戳(-1=不过期)
		Models       string `json:"models"`                       // 模型白名单
		Subnet       string `json:"subnet"`                       // IP白名单
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	// 权限: 组织管理员可创建Key给任何人，普通成员只能给自己创建
	if userId != req.UserId && !model.IsOrganizationAdmin(userId, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织管理员可为他人创建Key"})
		return
	}

	// 检查目标用户是否在组织中
	if !model.IsOrganizationMember(req.UserId, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "目标用户不是组织成员"})
		return
	}

	// 读取企业策略
	policy, err := model.GetEnterprisePolicy(orgId)
	if err != nil {
		policy = &model.EnterpriseTokenPolicy{OrgId: orgId}
	}

	// 策略约束: 默认额度
	quota := req.RemainQuota
	if quota <= 0 {
		quota = policy.DefaultQuota
	}

	// 策略约束: 自动过期
	expires := req.ExpiredTime
	if expires <= 0 && policy.AutoExpireDays > 0 {
		expires = helper.GetTimestamp() + int64(policy.AutoExpireDays*86400)
	}

	// 策略约束: 模型白名单
	models := req.Models
	if models == "" {
		models = policy.AllowedModels
	}

	// 策略约束: IP白名单强制
	if policy.RequireIpWhitelist && (req.Subnet == "") {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "企业策略要求必须设置IP白名单"})
		return
	}

	// 策略约束: 每人Key数上限
	if policy.MaxKeysPerUser > 0 {
		var count int64
		model.DB.Model(&model.Token{}).Where("user_id = ? AND status = ?", req.UserId, model.TokenStatusEnabled).Count(&count)
		if count >= int64(policy.MaxKeysPerUser) {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "该员工已超过最大Key数限制"})
			return
		}
	}

	// 是否需要审批?
	if policy.RequireAdminApproval && userId != req.UserId {
		// 创建审批记录
		approval := &model.EnterpriseApproval{
			OrgId:     orgId,
			Type:      "create_key",
			Status:    "pending",
			RequestBy: userId,
			Reason:    req.Purpose,
		}
		model.CreateApproval(approval)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Key创建申请已提交，等待管理员审批",
			"data":    gin.H{"approval_id": approval.Id},
		})
		return
	}

	// 不需要审批 → 直接创建Token
	rawKey := common.GetRandomString(32)
	cleanToken := model.Token{
		UserId:         req.UserId,
		Name:           req.Name,
		Key:            rawKey,
		CreatedTime:    helper.GetTimestamp(),
		AccessedTime:   helper.GetTimestamp(),
		ExpiredTime:    expires,
		RemainQuota:    quota,
		UnlimitedQuota: quota <= 0,
		Models:         &models,
		Subnet:         &req.Subnet,
	}
	if err := cleanToken.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建Token失败"})
		return
	}

	// 同时创建 EnterpriseToken 记录
	et := &model.EnterpriseToken{
		TokenId:      cleanToken.Id,
		DepartmentId: req.DepartmentId,
		CreatedBy:    userId,
		Purpose:      req.Purpose,
		Label:        req.Label,
		ApprovedBy:   userId,
	}
	model.CreateEnterpriseToken(et)

	cleanToken.Key = rawKey // 返回原始Key
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "企业Key已创建",
		"data": gin.H{
			"token":     cleanToken,
			"ent_token": et,
		},
	})
}

// GetOrgEnterpriseTokens 查看企业所有Key列表
// GET /api/org/:orgId/tokens
func GetOrgEnterpriseTokens(c *gin.Context) {
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}
	tokens, err := model.GetEnterpriseTokens(orgId)
	if err != nil {
		tokens = []model.EnterpriseToken{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tokens})
}

// ==================== 审批 ====================

// GetOrgApprovals 查看组织的审批列表
// GET /api/org/:orgId/approvals?status=pending
func GetOrgApprovals(c *gin.Context) {
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}
	status := c.Query("status")
	approvals, err := model.GetOrgApprovals(orgId, status)
	if err != nil {
		approvals = []model.EnterpriseApproval{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": approvals})
}

// ProcessApproval 处理审批 (通过/拒绝)
// POST /api/org/:orgId/approvals/:id/process
func ProcessApproval(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	orgId, err := strconv.Atoi(c.Param("orgId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的审批ID"})
		return
	}

	if !model.IsOrganizationAdmin(userId, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织管理员可处理审批"})
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"` // approve / reject
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Action != "approve" && req.Action != "reject") {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	status := "approved"
	if req.Action == "reject" {
		status = "rejected"
	}

	if err := model.ProcessApproval(id, status, req.Remark, userId); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "处理失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已" + status})
}
