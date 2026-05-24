package model

import (
	"fmt"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// PromoAd represents a promotional advertisement card in the scrolling marquee.
type PromoAd struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	PageKey   string `gorm:"size:50;not null;default:'all';index" json:"page_key"`
	Icon      string `gorm:"size:50;not null;default:'🚀'" json:"icon"`
	Title     string `gorm:"size:200;not null" json:"title"`
	LinkURL   string `gorm:"size:500;not null;default:'/models'" json:"link_url"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	Enabled   bool   `gorm:"default:true" json:"enabled"`
}

func InitPromoAdsTables() {
	if common.UsingSQLite {
		// SQLite: auto-migrate works fine
		if err := DB.AutoMigrate(&PromoAd{}); err != nil {
			logger.SysError("failed to migrate promo_ads table: " + err.Error())
			return
		}
	} else {
		// MySQL: use raw SQL to avoid auto-migrate issues
		createTableSQL := `CREATE TABLE IF NOT EXISTS promo_ads (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			page_key VARCHAR(50) NOT NULL DEFAULT 'all',
			icon VARCHAR(50) NOT NULL DEFAULT '🚀',
			title VARCHAR(200) NOT NULL,
			link_url VARCHAR(500) NOT NULL DEFAULT '/models',
			sort_order INT NOT NULL DEFAULT 0,
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			INDEX idx_page_key (page_key)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
		if err := DB.Exec(createTableSQL).Error; err != nil {
			logger.SysError("failed to create promo_ads table: " + err.Error())
			return
		}
	}

	// Seed default data if empty
	var count int64
	DB.Model(&PromoAd{}).Count(&count)
	if count == 0 {
		defaults := []PromoAd{
			{PageKey: "all", Icon: "🤖", Title: "GPT-4o — Multimodal Vision & Audio", LinkURL: "/models", SortOrder: 0, Enabled: true},
			{PageKey: "all", Icon: "🧠", Title: "Claude Sonnet 4 — Code & Reasoning", LinkURL: "/models", SortOrder: 1, Enabled: true},
			{PageKey: "all", Icon: "💎", Title: "DeepSeek V3 — 90% Cost Saving", LinkURL: "/models", SortOrder: 2, Enabled: true},
			{PageKey: "all", Icon: "🟢", Title: "Gemini 2.5 Pro — Long Context 1M", LinkURL: "/models", SortOrder: 3, Enabled: true},
			{PageKey: "all", Icon: "⚡", Title: "Groq — Ultra-Fast Inference", LinkURL: "/models", SortOrder: 4, Enabled: true},
			{PageKey: "all", Icon: "📐", Title: "Mistral Large — Precision & Control", LinkURL: "/models", SortOrder: 5, Enabled: true},
			{PageKey: "all", Icon: "🔮", Title: "Quantum Computing API — IonQ & IBM", LinkURL: "/quantum", SortOrder: 6, Enabled: true},
			{PageKey: "all", Icon: "🎲", Title: "Quantum Random Generator — ANU QRNG", LinkURL: "/quantum", SortOrder: 7, Enabled: true},
			{PageKey: "all", Icon: "🔄", Title: "Multi-Model Fusion — Auto Failover", LinkURL: "/fusion", SortOrder: 8, Enabled: true},
			{PageKey: "all", Icon: "⚡", Title: "99.9% Uptime SLA — Enterprise Ready", LinkURL: "/enterprise", SortOrder: 9, Enabled: true},
			{PageKey: "all", Icon: "💰", Title: "Pay Per Token — Only for What You Use", LinkURL: "/pricing", SortOrder: 10, Enabled: true},
			{PageKey: "all", Icon: "🛡", Title: "Enterprise Security — RBAC & Audit", LinkURL: "/enterprise", SortOrder: 11, Enabled: true},
		}
		for _, ad := range defaults {
			DB.Create(&ad)
		}
		logger.SysLog("seeded " + fmt.Sprintf("%d", len(defaults)) + " default promo ads")
	}
}

// GetAllPromoAds returns all promo ads sorted by sort_order.
func GetAllPromoAds() ([]PromoAd, error) {
	var ads []PromoAd
	err := DB.Order("sort_order ASC").Find(&ads).Error
	return ads, err
}

// GetEnabledPromoAds returns enabled ads, optionally filtered by page_key.
func GetEnabledPromoAds(pageKey string) ([]PromoAd, error) {
	var ads []PromoAd
	query := DB.Where("enabled = ?", true)
	if pageKey != "" {
		query = query.Where("page_key IN ?", []string{pageKey, "all"})
	}
	err := query.Order("sort_order ASC").Find(&ads).Error
	return ads, err
}

// CreatePromoAd creates a new promo ad.
func CreatePromoAd(ad *PromoAd) error {
	return DB.Create(ad).Error
}

// UpdatePromoAd updates an existing promo ad.
func UpdatePromoAd(ad *PromoAd) error {
	return DB.Save(ad).Error
}

// DeletePromoAd deletes a promo ad by ID.
func DeletePromoAd(id uint) error {
	return DB.Delete(&PromoAd{}, id).Error
}
