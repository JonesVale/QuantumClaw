package model

import (
	"strconv"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"gorm.io/gorm"
)

// Notification 用户通知
type Notification struct {
	Id        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId    int       `json:"user_id" gorm:"index;not null"`
	Type      string    `json:"type" gorm:"type:varchar(50);index"` // topup, system, alert
	Title     string    `json:"title" gorm:"type:varchar(255);not null"`
	Content   string    `json:"content" gorm:"type:text"`
	Data      string    `json:"data" gorm:"type:text"` // JSON 附加数据
	Read      bool      `json:"read" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty" gorm:"default:null"`
}

func (Notification) TableName() string {
	return "notifications"
}

// CreateNotification 创建通知
func CreateNotification(userId int, notifType string, title string, content string, data string) error {
	notif := &Notification{
		UserId:    userId,
		Type:      notifType,
		Title:     title,
		Content:   content,
		Data:      data,
		Read:      false,
		CreatedAt: time.Now(),
	}
	return DB.Create(notif).Error
}

// GetUserNotifications 获取用户通知列表（分页，最新在前）
func GetUserNotifications(userId int, page int, pageSize int, unreadOnly bool) ([]Notification, int64, error) {
	var notifs []Notification
	var total int64

	query := DB.Model(&Notification{}).Where("user_id = ?", userId)
	if unreadOnly {
		query = query.Where("read = ?", false)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&notifs).Error; err != nil {
		return nil, 0, err
	}

	return notifs, total, nil
}

// MarkNotificationRead 标记通知为已读
func MarkNotificationRead(notifId int, userId int) error {
	return DB.Model(&Notification{}).
		Where("id = ? AND user_id = ?", notifId, userId).
		Updates(map[string]interface{}{
			"read":     true,
			"read_at": time.Now(),
		}).Error
}

// MarkAllNotificationsRead 标记用户所有通知为已读
func MarkAllNotificationsRead(userId int) error {
	return DB.Model(&Notification{}).
		Where("user_id = ? AND read = ?", userId, false).
		Updates(map[string]interface{}{
			"read":     true,
			"read_at": time.Now(),
		}).Error
}

// GetUnreadNotificationCount 获取未读通知数量
func GetUnreadNotificationCount(userId int) (int64, error) {
	var count int64
	err := DB.Model(&Notification{}).Where("user_id = ? AND read = ?", userId, false).Count(&count).Error
	return count, err
}

// DeleteOldNotifications 清理旧通知（默认保留30天）
func DeleteOldNotifications(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	result := DB.Where("created_at < ?", cutoff).Delete(&Notification{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		logger.SysLog("清理了 " + strconv.FormatInt(result.RowsAffected, 10) + " 条旧通知")
	}
	return nil
}

// BeforeCreate GORM hook
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	return nil
}
