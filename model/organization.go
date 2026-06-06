package model

import (
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// Organization 组织/团队
type Organization struct {
	Id        int       `json:"id"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null" validate:"required,max=50"`
	OwnerId   int       `json:"owner_id" gorm:"type:int;index;not null"` // 组织创建者（超级管理员）
	Tier      string    `json:"tier" gorm:"type:varchar(20);default:'personal'"` // personal / enterprise / provider
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

const (
	OrgRoleAdmin  = "admin"  // 组织管理员
	OrgRoleMember = "member" // 普通成员
)

// OrganizationMember 组织成员关系
type OrganizationMember struct {
	Id       int       `json:"id"`
	UserId   int        `json:"user_id" gorm:"type:int;index;not null;uniqueIndex:idx_org_user"`
	OrgId    int        `json:"org_id" gorm:"type:int;index;not null;uniqueIndex:idx_org_user"`
	Role     string     `json:"role" gorm:"type:varchar(20);default:'member'"` // admin / member
	JoinedAt time.Time  `json:"joined_at" gorm:"autoCreateTime"`
}

// InitOrganizationTables 创建组织相关表
func InitOrganizationTables() {
	if err := DB.AutoMigrate(&Organization{}); err != nil {
		logger.SysError("InitOrganizationTables Organization AutoMigrate failed: " + err.Error())
		return
	}
	if err := DB.AutoMigrate(&OrganizationMember{}); err != nil {
		logger.SysError("InitOrganizationTables OrganizationMember AutoMigrate failed: " + err.Error())
		return
	}
	logger.SysLog("organization tables initialized")
}

// CreateOrganization 创建组织
func CreateOrganization(ownerId int, name string) (*Organization, error) {
	org := &Organization{
		Name:    name,
		OwnerId: ownerId,
	}
	if err := DB.Create(org).Error; err != nil {
		return nil, err
	}
	// 创建者自动成为组织管理员
	member := &OrganizationMember{
		UserId: ownerId,
		OrgId:  org.Id,
		Role:   OrgRoleAdmin,
	}
	if err := DB.Create(member).Error; err != nil {
		return nil, err
	}
	return org, nil
}

// GetOrganizationByID 根据 ID 获取组织
func GetOrganizationByID(orgId int) (*Organization, error) {
	var org Organization
	err := DB.Where("id = ?", orgId).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetOrganizationByMember 查询用户所属的所有组织
func GetOrganizationsByMember(userId int) ([]Organization, error) {
	var orgs []Organization
	err := DB.Raw(`
		SELECT o.* FROM organizations o
		INNER JOIN organization_members om ON om.org_id = o.id
		WHERE om.user_id = ?
		ORDER BY o.created_at DESC
	`, userId).Scan(&orgs).Error
	return orgs, err
}

// GetOrganizationMembers 查询组织所有成员
func GetOrganizationMembers(orgId int) ([]OrganizationMember, error) {
	var members []OrganizationMember
	err := DB.Where("org_id = ?", orgId).Find(&members).Error
	return members, err
}

// GetOrganizationMemberRole 查询用户在组织中的角色
func GetOrganizationMemberRole(userId, orgId int) (string, error) {
	var member OrganizationMember
	err := DB.Where("user_id = ? AND org_id = ?", userId, orgId).First(&member).Error
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

// AddOrganizationMember 添加成员到组织
func AddOrganizationMember(orgId, userId int, role string) error {
	member := &OrganizationMember{
		UserId: userId,
		OrgId:  orgId,
		Role:   role,
	}
	return DB.Create(member).Error
}

// RemoveOrganizationMember 从组织中移除成员
func RemoveOrganizationMember(orgId, userId int) error {
	return DB.Where("org_id = ? AND user_id = ?", orgId, userId).Delete(&OrganizationMember{}).Error
}

// IsOrganizationAdmin 判断用户是否为组织管理员
func IsOrganizationAdmin(userId, orgId int) bool {
	var count int64
	DB.Model(&OrganizationMember{}).Where("user_id = ? AND org_id = ? AND role = ?", userId, orgId, OrgRoleAdmin).Count(&count)
	return count > 0
}

// IsOrganizationMember 判断用户是否为组织成员
func IsOrganizationMember(userId, orgId int) bool {
	var count int64
	DB.Model(&OrganizationMember{}).Where("user_id = ? AND org_id = ?", userId, orgId).Count(&count)
	return count > 0
}

// OrganizationWithMemberCount 带成员数的组织信息
type OrganizationWithMemberCount struct {
	Organization
	MemberCount int64 `json:"member_count"`
}

// GetUserOrganizationsWithCount 获取用户所有组织及其成员数
func GetUserOrganizationsWithCount(userId int) ([]OrganizationWithMemberCount, error) {
	var result []OrganizationWithMemberCount
	err := DB.Raw(`
		SELECT o.*, COUNT(om.id) AS member_count
		FROM organizations o
		INNER JOIN organization_members om ON om.org_id = o.id
		WHERE o.id IN (
			SELECT org_id FROM organization_members WHERE user_id = ?
		)
		GROUP BY o.id
		ORDER BY o.created_at DESC
	`, userId).Scan(&result).Error
	return result, err
}

// GetAdminOrganizationMemberCount 获取管理员管理的组织中的成员数
func GetAdminOrganizationMemberCount(orgId int) (int64, error) {
	var count int64
	err := DB.Model(&OrganizationMember{}).Where("org_id = ?", orgId).Count(&count).Error
	return count, err
}
