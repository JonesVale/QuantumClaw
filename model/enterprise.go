package model

import (
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ── 企业部门 ──

// Department 企业部门（公司下的子组织）
type Department struct {
	Id             int       `json:"id" gorm:"primaryKey;autoIncrement"`
	OrgId          int       `json:"org_id" gorm:"index;not null"`              // 所属企业组织
	Name           string    `json:"name" gorm:"type:varchar(100);not null"`    // 部门名称
	Description    string    `json:"description" gorm:"type:text"`              // 部门描述
	HeadUserId     int       `json:"head_user_id" gorm:"default:0"`             // 部门负责人
	MonthlyBudget  int64     `json:"monthly_budget" gorm:"bigint;default:0"`    // 月度预算限额(分)
	AlertThreshold int64     `json:"alert_threshold" gorm:"bigint;default:0"`   // 预警阈值(用量达此值发通知)
	Status         int       `json:"status" gorm:"default:1"`                   // 1=启用 0=禁用
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ── 企业Token策略 ──

// EnterpriseTokenPolicy 企业Token创建策略（公司级默认值）
type EnterpriseTokenPolicy struct {
	Id               int    `json:"id" gorm:"primaryKey;autoIncrement"`
	OrgId            int    `json:"org_id" gorm:"uniqueIndex;not null"`
	DefaultQuota     int64  `json:"default_quota" gorm:"bigint;default:1000000"`       // 新Token默认额度
	DefaultGroup     string `json:"default_group" gorm:"type:varchar(32);default:'default'"` // 默认分组
	AllowedModels    string `json:"allowed_models" gorm:"type:text"`                   // 允许的模型列表(逗号分隔,空=全部)
	BlockedModels    string `json:"blocked_models" gorm:"type:text"`                   // 禁止的模型列表
	RequireIpWhitelist bool `json:"require_ip_whitelist" gorm:"default:false"`         // 是否强制IP白名单
	MaxKeysPerUser   int    `json:"max_keys_per_user" gorm:"default:50"`               // 每人最多创建多少Key
	AutoExpireDays   int    `json:"auto_expire_days" gorm:"default:0"`                 // 自动过期天数(0=不过期)
	RequireAdminApproval bool `json:"require_admin_approval" gorm:"default:false"`     // 员工创建Key是否需要管理员审批
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ── 企业员工角色 ──

const (
	EnterpriseRoleSuperAdmin = "super_admin"  // 企业超级管理员(owner)
	EnterpriseRoleOrgAdmin   = "org_admin"    // 组织管理员
	EnterpriseRoleDeptHead   = "dept_head"    // 部门负责人
	EnterpriseRoleMember     = "member"       // 普通员工
)

// ── Token扩展（企业专属） ──

// EnterpriseToken 企业Token信息（扩展基础Token）
type EnterpriseToken struct {
	Id           int    `json:"id" gorm:"primaryKey;autoIncrement"`
	TokenId      int    `json:"token_id" gorm:"uniqueIndex;not null"`  // 关联的基础Token
	DepartmentId int    `json:"department_id" gorm:"default:0;index"`  // 所属部门(0=个人)
	CreatedBy    int    `json:"created_by" gorm:"not null"`            // 创建者(管理员或员工本人)
	Purpose      string `json:"purpose" gorm:"type:varchar(200)"`      // 用途说明
	ApprovedBy   int    `json:"approved_by" gorm:"default:0"`          // 审批人(需审批时)
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	Label        string `json:"label" gorm:"type:varchar(50)"`         // 标签(如"生产环境"/"测试环境")
}

// ── 企业用量统计 ──

// EnterpriseUsageStat 企业用量统计（按部门/员工维度）
type EnterpriseUsageStat struct {
	Id           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	OrgId        int       `json:"org_id" gorm:"index;not null"`
	DepartmentId int       `json:"department_id" gorm:"default:0;index"` // 0=全公司
	UserId       int       `json:"user_id" gorm:"default:0;index"`       // 0=部门汇总
	Date         string    `json:"date" gorm:"type:varchar(10);index"`   // 统计日期 2026-06-05
	TokenCount   int64     `json:"token_count" gorm:"bigint;default:0"`  // Token消耗量
	RequestCount int64     `json:"request_count" gorm:"bigint;default:0"`// 请求次数
	Cost         int64     `json:"cost" gorm:"bigint;default:0"`         // 消费金额(分)
	ModelName    string    `json:"model_name" gorm:"type:varchar(100)"`  // 模型名称
}

// ── 企业审批记录 ──

// EnterpriseApproval 企业审批记录
type EnterpriseApproval struct {
	Id         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	OrgId      int       `json:"org_id" gorm:"index;not null"`
	Type       string    `json:"type" gorm:"type:varchar(20)"`       // create_key / change_policy / upgrade_quota
	Status     string    `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending / approved / rejected
	RequestBy  int       `json:"request_by"`                         // 申请人
	ApprovedBy int       `json:"approved_by" gorm:"default:0"`       // 审批人
	Content    string    `json:"content" gorm:"type:text"`           // 申请内容(JSON)
	Reason     string    `json:"reason" gorm:"type:varchar(500)"`    // 申请理由
	Remark     string    `json:"remark" gorm:"type:varchar(500)"`    // 审批备注
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// ── 表初始化 ──

func InitEnterpriseTables() {
	if err := DB.AutoMigrate(&Department{}); err != nil {
		logger.SysError("InitEnterpriseTables Department AutoMigrate failed: " + err.Error())
		return
	}
	if err := DB.AutoMigrate(&EnterpriseTokenPolicy{}); err != nil {
		logger.SysError("InitEnterpriseTables EnterpriseTokenPolicy AutoMigrate failed: " + err.Error())
		return
	}
	if err := DB.AutoMigrate(&EnterpriseToken{}); err != nil {
		logger.SysError("InitEnterpriseTables EnterpriseToken AutoMigrate failed: " + err.Error())
		return
	}
	if err := DB.AutoMigrate(&EnterpriseUsageStat{}); err != nil {
		logger.SysError("InitEnterpriseTables EnterpriseUsageStat AutoMigrate failed: " + err.Error())
		return
	}
	if err := DB.AutoMigrate(&EnterpriseApproval{}); err != nil {
		logger.SysError("InitEnterpriseTables EnterpriseApproval AutoMigrate failed: " + err.Error())
		return
	}
	logger.SysLog("enterprise tables initialized")
}

// ── 部门 CRUD ──

func CreateDepartment(d *Department) error {
	return DB.Create(d).Error
}

func GetDepartmentById(id int) (*Department, error) {
	var d Department
	err := DB.First(&d, id).Error
	return &d, err
}

func GetOrgDepartments(orgId int) ([]Department, error) {
	var list []Department
	err := DB.Where("org_id = ?", orgId).Order("id ASC").Find(&list).Error
	return list, err
}

func UpdateDepartment(d *Department) error {
	return DB.Model(d).Select("name", "description", "head_user_id", "monthly_budget", "alert_threshold", "status").Updates(d).Error
}

func DeleteDepartment(id int) error {
	return DB.Delete(&Department{}, id).Error
}

func GetDepartmentMemberCount(deptId int) (int64, error) {
	var count int64
	err := DB.Model(&User{}).Where("department_id = ?", deptId).Count(&count).Error
	return count, err
}

// ── 策略 CRUD ──

func GetEnterprisePolicy(orgId int) (*EnterpriseTokenPolicy, error) {
	var p EnterpriseTokenPolicy
	err := DB.Where("org_id = ?", orgId).First(&p).Error
	return &p, err
}

func SaveEnterprisePolicy(p *EnterpriseTokenPolicy) error {
	var existing EnterpriseTokenPolicy
	if DB.Where("org_id = ?", p.OrgId).First(&existing).Error != nil {
		return DB.Create(p).Error
	}
	return DB.Model(&existing).Updates(p).Error
}

// ── 企业 Token ──

func CreateEnterpriseToken(et *EnterpriseToken) error {
	return DB.Create(et).Error
}

func GetEnterpriseTokens(orgId int) ([]EnterpriseToken, error) {
	var list []EnterpriseToken
	err := DB.Where("org_id = ?", orgId).Order("id DESC").Find(&list).Error
	return list, err
}

func GetEnterpriseTokenByDept(deptId int) ([]EnterpriseToken, error) {
	var list []EnterpriseToken
	err := DB.Where("department_id = ?", deptId).Order("id DESC").Find(&list).Error
	return list, err
}

// ── 用量统计 ──

func GetOrgUsageSummary(orgId int, dateFrom string) (*EnterpriseUsageStat, error) {
	var result EnterpriseUsageStat
	err := DB.Table("token_transactions").
		Select("org_id, SUM(total_amount) as cost, COUNT(*) as request_count").
		Where("org_id = ? AND created_time >= ?", orgId, dateFrom).
		Scan(&result).Error
	return &result, err
}

// ── 审批 ──

func CreateApproval(a *EnterpriseApproval) error {
	return DB.Create(a).Error
}

func GetOrgApprovals(orgId int, status string) ([]EnterpriseApproval, error) {
	query := DB.Where("org_id = ?", orgId)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var list []EnterpriseApproval
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

func ProcessApproval(id int, status, remark string, approvedBy int) error {
	return DB.Model(&EnterpriseApproval{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       status,
		"remark":       remark,
		"approved_by":  approvedBy,
		"processed_at": time.Now(),
	}).Error
}
