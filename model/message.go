package model

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ── 发送任务 ──

// MessageJob 批量消息发送任务
type MessageJob struct {
	Id               int        `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId           int        `json:"user_id" gorm:"index;not null"`
	Channel          string     `json:"channel" gorm:"type:varchar(20);not null;index"` // sms / wechat / email
	BatchLimit       int        `json:"batch_limit" gorm:"default:20"`
	TotalTargets     int        `json:"total_targets" gorm:"default:0"`
	SentCount        int        `json:"sent_count" gorm:"default:0"`
	FailCount        int        `json:"fail_count" gorm:"default:0"`
	CurrentBatch     int        `json:"current_batch" gorm:"default:0"`
	Status           string     `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending / running / paused / completed / cancelled
	AgreementVersion string     `json:"agreement_version"`
	ConsentTime      int64      `json:"consent_time"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

// ── 发送记录 ──

// MessageLog 单条发送记录
type MessageLog struct {
	Id           int        `json:"id" gorm:"primaryKey;autoIncrement"`
	JobId        int        `json:"job_id" gorm:"index;not null"`
	Batch        int        `json:"batch" gorm:"default:0"`
	UserId       int        `json:"user_id"`
	Target       string     `json:"target"`        // 手机号/微信号
	TargetName   string     `json:"target_name"`   // 联系人名称
	Content      string     `json:"content" gorm:"type:text"` // 实际发送内容
	AffCode      string     `json:"aff_code"`      // 邀请人的推广码
	Status       string     `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending / sent / failed
	ErrorMsg     string     `json:"error_msg" gorm:"type:varchar(500)"`
	DeviceResult string     `json:"device_result" gorm:"type:varchar(50)"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
}

// ── 用户协议 ──

// MessageAgreement 用户批量发送协议版本
type MessageAgreement struct {
	Id        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Version   string    `json:"version" gorm:"type:varchar(20);uniqueIndex;not null"`
	Title     string    `json:"title" gorm:"type:varchar(200)"`
	Content   string    `json:"content" gorm:"type:longtext;not null"`
	Channel   string    `json:"channel" gorm:"type:varchar(20);default:'all'"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ── 表初始化 ──

func InitMessageTables() {
	models := []interface{}{
		&MessageJob{},
		&MessageLog{},
		&MessageAgreement{},
	}
	for _, m := range models {
		if err := DB.AutoMigrate(m); err != nil {
			logger.SysError("InitMessageTables AutoMigrate failed: " + err.Error())
			return
		}
	}
	logger.SysLog("message tables initialized")
}

// ═══════════════════════════════════════════════════════════════════════════
// MessageJob CRUD
// ═══════════════════════════════════════════════════════════════════════════

func CreateMessageJob(j *MessageJob) error {
	return DB.Create(j).Error
}

func GetMessageJob(id int) (*MessageJob, error) {
	var j MessageJob
	err := DB.First(&j, id).Error
	return &j, err
}

func GetUserMessageJobs(userId int, limit int) ([]MessageJob, error) {
	var list []MessageJob
	err := DB.Where("user_id = ?", userId).
		Order("id DESC").Limit(limit).Find(&list).Error
	return list, err
}

func UpdateMessageJobProgress(id, sentCount, failCount, currentBatch int) error {
	return DB.Model(&MessageJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"sent_count":    sentCount,
		"fail_count":    failCount,
		"current_batch": currentBatch,
		"status":        "running",
	}).Error
}

func UpdateMessageJobStatus(id int, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "completed" || status == "cancelled" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return DB.Model(&MessageJob{}).Where("id = ?", id).Updates(updates).Error
}

func UpdateMessageJobStatusWithUserCheck(id, userId int, status string) error {
	updates := map[string]interface{}{"status": status}
	if status == "completed" || status == "cancelled" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	result := DB.Model(&MessageJob{}).
		Where("id = ? AND user_id = ?", id, userId).
		Updates(updates)
	if result.RowsAffected == 0 {
		logger.SysError("no matching job or no permission")
		return fmt.Errorf("无匹配任务或无权限")
	}
	return result.Error
}

// ═══════════════════════════════════════════════════════════════════════════
// MessageLog CRUD
// ═══════════════════════════════════════════════════════════════════════════

func CreateMessageLog(l *MessageLog) error {
	return DB.Create(l).Error
}

func BatchCreateMessageLogs(logs []MessageLog) error {
	if len(logs) == 0 {
		return nil
	}
	return DB.CreateInBatches(logs, 50).Error
}

func GetJobMessageLogs(jobId int, offset, limit int) ([]MessageLog, error) {
	var list []MessageLog
	err := DB.Where("job_id = ?", jobId).
		Order("id ASC").Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}

func GetMessageJobStats(jobId int) (sent int64, failed int64, pending int64, err error) {
	err = DB.Model(&MessageLog{}).Where("job_id = ?", jobId).
		Select(`
			COALESCE(SUM(CASE WHEN status='sent' THEN 1 ELSE 0 END), 0) as sent,
			COALESCE(SUM(CASE WHEN status='failed' THEN 1 ELSE 0 END), 0) as failed,
			COALESCE(SUM(CASE WHEN status='pending' THEN 1 ELSE 0 END), 0) as pending
		`).Row().Scan(&sent, &failed, &pending)
	return
}

// ═══════════════════════════════════════════════════════════════════════════
// MessageAgreement CRUD
// ═══════════════════════════════════════════════════════════════════════════

func GetActiveAgreement() (*MessageAgreement, error) {
	var a MessageAgreement
	err := DB.Where("is_active = ?", true).First(&a).Error
	return &a, err
}

func CreateMessageAgreement(a *MessageAgreement) error {
	// 新版本设为 active，旧版本 deactivate
	DB.Model(&MessageAgreement{}).Where("is_active = ?", true).
		Update("is_active", false)
	return DB.Create(a).Error
}
