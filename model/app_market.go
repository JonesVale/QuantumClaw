package model

import (
	"context"
	"time"
)

// AppMarket 应用市场
type AppMarket struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Author      string `gorm:"size:100" json:"author"`
	AuthorURL   string `gorm:"size:500" json:"author_url"`
	AppURL      string `gorm:"size:500" json:"app_url"`
	IconURL     string `gorm:"size:500" json:"icon_url"`
	Category    string `gorm:"size:50;default:tool" json:"category"` // tool/plugin/integration
	Status      string `gorm:"size:20;default:draft" json:"status"` // draft/published/rejected
	UserID      uint   `json:"user_id"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

func InsertApp(ctx context.Context, app *AppMarket) error {
	now := time.Now().Unix()
	app.CreatedAt = now
	app.UpdatedAt = now
	return DB.WithContext(ctx).Create(app).Error
}

func GetPublishedApps(ctx context.Context, category string, page, pageSize int) ([]AppMarket, int64, error) {
	var items []AppMarket
	var total int64
	query := DB.WithContext(ctx).Model(&AppMarket{}).Where("status = ?", "published")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func GetAppByID(ctx context.Context, id uint) (*AppMarket, error) {
	var app AppMarket
	err := DB.WithContext(ctx).First(&app, id).Error
	return &app, err
}

func UpdateAppStatus(ctx context.Context, id uint, status string) error {
	return DB.WithContext(ctx).Model(&AppMarket{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now().Unix()}).Error
}

func GetUserApps(ctx context.Context, userID uint, page, pageSize int) ([]AppMarket, int64, error) {
	var items []AppMarket
	var total int64
	query := DB.WithContext(ctx).Model(&AppMarket{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func GetAllAppsPaginated(ctx context.Context, page, pageSize int, status string) ([]AppMarket, int64, error) {
	var items []AppMarket
	var total int64
	query := DB.WithContext(ctx).Model(&AppMarket{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func DeleteApp(ctx context.Context, id uint) error {
	return DB.WithContext(ctx).Delete(&AppMarket{}, id).Error
}
