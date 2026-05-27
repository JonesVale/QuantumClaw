package model

import (
	"context"
	"time"
)

// Feedback 用户反馈
type Feedback struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	UserID        uint   `gorm:"index" json:"user_id"`
	Title         string `gorm:"size:200" json:"title"`
	Content       string `gorm:"type:text" json:"content"`
	Type          string `gorm:"size:20;default:question" json:"type"` // bug / feature / question
	Email         string `gorm:"size:100" json:"email"`
	Status        string `gorm:"size:20;default:pending" json:"status"` // pending / resolved / closed
	AdminResponse string `gorm:"type:text" json:"admin_response"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
}

// FAQ 常见问题
type FAQ struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Question  string `gorm:"size:500" json:"question"`
	Answer    string `gorm:"type:text" json:"answer"`
	Category  string `gorm:"size:50;default:general" json:"category"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// FeedbackWithUser 含用户信息的反馈视图
type FeedbackWithUser struct {
	Feedback
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// ========== Feedback CRUD ==========

func InsertFeedback(ctx context.Context, fb *Feedback) error {
	now := time.Now().Unix()
	fb.CreatedAt = now
	fb.UpdatedAt = now
	return DB.WithContext(ctx).Create(fb).Error
}

func GetFeedbackByID(ctx context.Context, id uint) (*Feedback, error) {
	var fb Feedback
	err := DB.WithContext(ctx).First(&fb, id).Error
	return &fb, err
}

func GetFeedbackPaginated(ctx context.Context, page, pageSize int, status, fbType string) ([]FeedbackWithUser, int64, error) {
	var items []FeedbackWithUser
	var total int64

	query := DB.WithContext(ctx).Table("feedbacks").
		Select("feedbacks.*, users.username, users.display_name").
		Joins("LEFT JOIN users ON users.id = feedbacks.user_id")

	if status != "" {
		query = query.Where("feedbacks.status = ?", status)
	}
	if fbType != "" {
		query = query.Where("feedbacks.type = ?", fbType)
	}

	query.Count(&total)

	err := query.Order("feedbacks.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error

	return items, total, err
}

func UpdateFeedbackStatus(ctx context.Context, id uint, status string) error {
	return DB.WithContext(ctx).Model(&Feedback{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now().Unix()}).Error
}

func RespondToFeedback(ctx context.Context, id uint, response string) error {
	return DB.WithContext(ctx).Model(&Feedback{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"admin_response": response,
			"status":         "resolved",
			"updated_at":     time.Now().Unix(),
		}).Error
}

func GetUserFeedback(ctx context.Context, userID uint, page, pageSize int) ([]Feedback, int64, error) {
	var items []Feedback
	var total int64

	query := DB.WithContext(ctx).Model(&Feedback{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

// ========== FAQ CRUD ==========

func InsertFAQ(ctx context.Context, faq *FAQ) error {
	now := time.Now().Unix()
	faq.CreatedAt = now
	faq.UpdatedAt = now
	return DB.WithContext(ctx).Create(faq).Error
}

func GetFAQs(ctx context.Context, category string) ([]FAQ, error) {
	var items []FAQ
	query := DB.WithContext(ctx).Model(&FAQ{}).Order("sort_order ASC, id ASC")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Find(&items).Error
	return items, err
}

func GetFAQByID(ctx context.Context, id uint) (*FAQ, error) {
	var faq FAQ
	err := DB.WithContext(ctx).First(&faq, id).Error
	return &faq, err
}

func UpdateFAQ(ctx context.Context, faq *FAQ) error {
	faq.UpdatedAt = time.Now().Unix()
	return DB.WithContext(ctx).Model(&FAQ{}).Where("id = ?", faq.ID).Updates(map[string]interface{}{
		"question":   faq.Question,
		"answer":     faq.Answer,
		"category":   faq.Category,
		"sort_order": faq.SortOrder,
		"updated_at": faq.UpdatedAt,
	}).Error
}

func DeleteFAQ(ctx context.Context, id uint) error {
	return DB.WithContext(ctx).Delete(&FAQ{}, id).Error
}

func GetFAQCategories(ctx context.Context) ([]string, error) {
	var cats []string
	err := DB.WithContext(ctx).Model(&FAQ{}).Select("DISTINCT category").Order("category ASC").Pluck("category", &cats).Error
	return cats, err
}
