package model

import (
	"time"

	"gorm.io/gorm"
)

// RssArticle represents a fetched RSS/Atom article stored in the database
type RssArticle struct {
	Id          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Source      string    `json:"source" gorm:"index;type:varchar(100)"`
	Title       string    `json:"title" gorm:"type:text"`
	Link        string    `json:"link" gorm:"type:text;uniqueIndex:idx_rss_link"`
	Description string    `json:"description" gorm:"type:text"`
	Author      string    `json:"author" gorm:"type:varchar(200)"`
	PublishedAt time.Time `json:"published_at" gorm:"index"`
	Language    string    `json:"language" gorm:"type:varchar(10);default:'zh'"`
	CreatedAt   time.Time `json:"created_at"`
}

func (RssArticle) TableName() string {
	return "rss_articles"
}

// AddRssArticle inserts a new RSS article, skipping if the link already exists
func AddRssArticle(article *RssArticle) error {
	if article.Link == "" {
		return nil
	}
	// Use FirstOrCreate to ignore duplicates by Link
	result := DB.Where("link = ?", article.Link).FirstOrCreate(article)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpsertRssArticle inserts or updates an article (by link)
func UpsertRssArticle(article *RssArticle) error {
	if article.Link == "" {
		return nil
	}
	var existing RssArticle
	result := DB.Where("link = ?", article.Link).First(&existing)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return DB.Create(article).Error
		}
		return result.Error
	}
	article.Id = existing.Id
	return DB.Model(&existing).Updates(map[string]interface{}{
		"title":       article.Title,
		"description": article.Description,
		"author":      article.Author,
		"published_at": article.PublishedAt,
	}).Error
}

// GetRssArticles retrieves paginated RSS articles filtered by language
func GetRssArticles(language string, limit int, offset int) ([]RssArticle, int64, error) {
	var articles []RssArticle
	var total int64

	query := DB.Model(&RssArticle{})
	if language != "" && language != "all" {
		query = query.Where("language = ?", language)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("published_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&articles).Error
	if err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetRssArticleByLink finds an article by its link
func GetRssArticleByLink(link string) (*RssArticle, error) {
	var article RssArticle
	err := DB.Where("link = ?", link).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// CountRssArticles returns total count of articles matching the language filter
func CountRssArticles(language string) (int64, error) {
	var total int64
	query := DB.Model(&RssArticle{})
	if language != "" && language != "all" {
		query = query.Where("language = ?", language)
	}
	err := query.Count(&total).Error
	return total, err
}
