package model

import (
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// BrandRanking stores the industry-wide brand usage ranking.
// Updated monthly from external sources (HuggingFace, GitHub, etc.).
type BrandRanking struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	BrandName string  `gorm:"type:varchar(100);not null;uniqueIndex" json:"brand_name"`
	Rank      int     `gorm:"not null;index" json:"rank"`
	Score     float64 `gorm:"type:decimal(12,4);default:0" json:"score"`
	Metric    string  `gorm:"type:varchar(50);default:'composite'" json:"metric"`
	Source    string  `gorm:"type:varchar(200)" json:"source"`
	FetchedAt int64   `json:"fetched_at"`
}

func (BrandRanking) TableName() string {
	return "brand_rankings"
}

// InitBrandRankingTable ensures the table exists.
func InitBrandRankingTable() {
	if err := DB.AutoMigrate(&BrandRanking{}); err != nil {
		logger.SysError("InitBrandRankingTable AutoMigrate failed: " + err.Error())
	}
}
