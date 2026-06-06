package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// CreateOrgRequest 创建组织请求
type CreateOrgRequest struct {
	Name string `json:"name" binding:"required,max=50"`
	Tier string `json:"tier"` // personal / enterprise / provider
}

// 创建组织
func CreateOrganization(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请先登录"})
		return
	}

	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "组织名称不能为空且不超过50字"})
		return
	}

	tier := req.Tier
	if tier != "enterprise" && tier != "provider" {
		tier = "personal"
	}

	org, err := model.CreateOrganization(userId, req.Name)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建组织失败"})
		return
	}

	// 设置 tier
	if tier != "personal" {
		model.DB.Model(org).Update("tier", tier)
		org.Tier = tier
	}

	// 更新用户的 organization_id
	model.DB.Model(&model.User{}).Where("id = ?", userId).Update("organization_id", org.Id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    org,
	})
}

// 获取当前用户的所有组织
func GetMyOrganizations(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请先登录"})
		return
	}

	orgs, err := model.GetUserOrganizationsWithCount(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取组织列表失败"})
		return
	}
	if orgs == nil {
		orgs = []model.OrganizationWithMemberCount{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    orgs,
	})
}

// 获取组织成员列表
func GetOrganizationMembers(c *gin.Context) {
	userId := c.GetInt("id")
	orgId, err := strconv.Atoi(c.Param("id"))
	if err != nil || orgId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}

	// 校验当前用户是组织成员
	if !model.IsOrganizationMember(userId, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "您不是该组织成员"})
		return
	}

	members, err := model.GetOrganizationMembers(orgId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取成员列表失败"})
		return
	}

	// 填充用户信息
	type MemberInfo struct {
		UserId      int    `json:"user_id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		JoinedAt    string `json:"joined_at"`
	}

	var result []MemberInfo
	for _, m := range members {
		var user model.User
		if model.DB.Where("id = ?", m.UserId).First(&user).Error == nil {
			result = append(result, MemberInfo{
				UserId:      m.UserId,
				Username:    user.Username,
				DisplayName: user.DisplayName,
				Role:        m.Role,
				JoinedAt:    m.JoinedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}
	if result == nil {
		result = []MemberInfo{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    result,
	})
}

type InviteMemberRequest struct {
	Username string `json:"username" binding:"required"`
}

// 邀请用户加入组织
func InviteOrganizationMember(c *gin.Context) {
	userId := c.GetInt("id")
	orgId, err := strconv.Atoi(c.Param("id"))
	if err != nil || orgId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}

	// 校验当前用户是组织管理员
	if !model.IsOrganizationAdmin(userId, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织管理员可邀请成员"})
		return
	}

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请输入要邀请的用户名"})
		return
	}

	// 查找被邀请用户
	var targetUser model.User
	if err := model.DB.Where("username = ?", req.Username).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	// 检查是否已是成员
	if model.IsOrganizationMember(targetUser.Id, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "该用户已是组织成员"})
		return
	}

	// 添加成员（角色为普通成员）
	if err := model.AddOrganizationMember(orgId, targetUser.Id, model.OrgRoleMember); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "邀请失败"})
		return
	}

	// 如果用户没有主组织，自动设为当前组织
	var user model.User
	model.DB.Where("id = ?", targetUser.Id).First(&user)
	if user.OrganizationID == 0 {
		model.DB.Model(&model.User{}).Where("id = ?", targetUser.Id).Update("organization_id", orgId)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已邀请用户加入组织",
	})
}

// 从组织中移除成员
func RemoveOrganizationMember(c *gin.Context) {
	userId := c.GetInt("id")
	orgId, err := strconv.Atoi(c.Param("id"))
	if err != nil || orgId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的组织ID"})
		return
	}

	targetUserId, err := strconv.Atoi(c.Param("userId"))
	if err != nil || targetUserId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的用户ID"})
		return
	}

	// 校验：组织管理员可移除，或者自己退出
	if userId != targetUserId && !model.IsOrganizationAdmin(userId, orgId) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅组织管理员可移除成员"})
		return
	}

	// 不能移除组织创建者
	org, err := model.GetOrganizationByID(orgId)
	if err == nil && org.OwnerId == targetUserId && userId != targetUserId {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "不能移除组织创建者"})
		return
	}

	if err := model.RemoveOrganizationMember(orgId, targetUserId); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "移除成员失败"})
		return
	}

	// 清除用户的 organization_id（如果是当前组织）
	var user model.User
	model.DB.Where("id = ?", targetUserId).First(&user)
	if user.OrganizationID == orgId {
		model.DB.Model(&model.User{}).Where("id = ?", targetUserId).Update("organization_id", 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已移除成员",
	})
}
