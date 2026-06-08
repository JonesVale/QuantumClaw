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
	Link        string    `json:"link" gorm:"type:varchar(512);uniqueIndex:idx_rss_link"`
	Description string    `json:"description" gorm:"type:text"`
	Author      string    `json:"author" gorm:"type:varchar(200)"`
	PublishedAt time.Time `json:"published_at" gorm:"index"`
	Language    string    `json:"language" gorm:"type:varchar(10);default:'zh'"`
	CreatedAt   time.Time `json:"created_at"`
}

func (RssArticle) TableName() string {
	return "rss_articles"
}

// DbRssSource represents an RSS/Atom feed source configuration stored in the database.
// This allows dynamic management of news sources without code changes or recompilation.
type DbRssSource struct {
	Id        int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string     `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"` // Source display name (e.g., "36氪 AI", "雷锋网")
	FeedURL   string     `json:"feed_url" gorm:"type:varchar(512);not null;uniqueIndex"` // RSS/Atom feed URL
	Language  string     `json:"language" gorm:"type:varchar(10);default:'zh'"`         // Language code ("zh", "en")
	Enabled   bool       `json:"enabled" gorm:"type:boolean;default:true;index"`          // Whether this source is active for fetching
	Category  string     `json:"category" gorm:"type:varchar(50);default:'general'"`      // Category tag for grouping (e.g., "ai", "quantum", "tech")
	LastFetch *time.Time `json:"last_fetch" gorm:"type:datetime"`                         // Last successful fetch timestamp (nil if never)
	FetchErr  string     `json:"fetch_err" gorm:"type:varchar(500)"`                      // Last fetch error message (empty = success)
	Articles  int64      `json:"articles" gorm:"type:int;default:0"`                      // Count of articles from this source in rss_articles table
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (DbRssSource) TableName() string {
	return "rss_sources"
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

// ---------- DbRssSource CRUD ----------

// GetEnabledRssSources returns all enabled RSS sources from the database.
func GetEnabledRssSources() ([]DbRssSource, error) {
	var sources []DbRssSource
	err := DB.Where("enabled = ?", true).Order("id ASC").Find(&sources).Error
	return sources, err
}

// GetAllRssSources returns all RSS sources (including disabled ones), paginated.
func GetAllRssSources(limit, offset int) ([]DbRssSource, int64, error) {
	var sources []DbRssSource
	var total int64

	err := DB.Model(&DbRssSource{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Order("id ASC").Limit(limit).Offset(offset).Find(&sources).Error
	if err != nil {
		return nil, 0, err
	}

	return sources, total, nil
}

// GetRssSourceById returns a single RSS source by ID.
func GetRssSourceById(id int) (*DbRssSource, error) {
	var source DbRssSource
	err := DB.First(&source, id).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// CreateRssSource inserts a new RSS source into the database.
func CreateRssSource(source *DbRssSource) error {
	return DB.Create(source).Error
}

// UpdateRssSource updates an existing RSS source.
func UpdateRssSource(source *DbRssSource) error {
	return DB.Save(source).Error
}

// DeleteRssSource deletes an RSS source by ID.
func DeleteRssSource(id int) error {
	return DB.Delete(&DbRssSource{}, id).Error
}

// UpdateRssSourceFetchStatus updates the last_fetch and fetch_err fields after a fetch attempt.
func UpdateRssSourceFetchStatus(id int, lastFetch time.Time, errMsg string) error {
	updates := map[string]interface{}{
		"last_fetch": lastFetch,
		"fetch_err":  errMsg,
	}
	if errMsg == "" {
		// Success: count articles from this source
		var count int64
		DB.Model(&RssArticle{}).Where("source = (SELECT name FROM rss_sources WHERE id = ?)", id).Count(&count)
		updates["articles"] = count
	}
	return DB.Model(&DbRssSource{}).Where("id = ?", id).Updates(updates).Error
}

// SeedDefaultRssSources seeds the database with default RSS sources if the table is empty.
// This is called on startup to ensure there are always some sources available.
//
// 国际化说明 (International):
//   - 中文源 (zh): 国内/国外服务器均可访问
//   - 英文源 (en): 部分源在国内可能超时(标记为disabled)，在国外服务器上可正常访问
//   - 部署到国外服务器后，可通过 Admin API 或后台启用被禁用的英文源
func SeedDefaultRssSources() (int64, error) {
	var count int64
	DB.Model(&DbRssSource{}).Count(&count)
	if count > 0 {
		return 0, nil // Already has data, skip seeding
	}

	defaults := []struct {
		Name     string
		FeedURL  string
		Language string
		Enabled  bool
		Category string
	}{
		// ── 中文源（国内外均可用）──
		{"36氪 AI", "https://36kr.com/feed", "zh", true, "ai"},
		{"雷锋网", "https://www.leiphone.com/feed", "zh", true, "tech"},
		{"AI科技评论", "https://www.aireviewweekly.com/rss.xml", "zh", true, "ai"},
		{"InfoQ AI", "https://www.infoq.cn/feed?tag=人工智能", "zh", true, "ai"},

		// ── 英文源（国外服务器推荐启用）──
		{"MIT Tech Review", "https://www.technologyreview.com/topic/artificial-intelligence/feed/", "en", true, "ai"},
		{"Quanta Magazine", "https://www.quantamagazine.org/feed", "en", true, "quantum"},

		// ── 英文源（国内可能超时，国外服务器可用，初始禁用）──
		{"OpenAI Blog", "https://openai.com/blog/feed.xml", "en", false, "ai"},
		{"Anthropic Blog", "https://www.anthropic.com/blog/rss", "en", false, "ai"},
		{"Google AI Blog", "https://blog.google/technology/ai/rss/", "en", false, "ai"},
		{"ArXiv CS.AI", "https://export.arxiv.org/rss/cs.AI", "en", false, "academic"},
		{"ArXiv Quantum", "https://export.arxiv.org/rss/quant-ph", "en", false, "quantum"},
		{"Hugging Face Blog", "https://huggingface.co/blog/feed.xml", "en", false, "ai"},
		{"Reddit r/artificial", "https://www.reddit.com/r/artificial/.rss", "en", false, "community"},
		{"The Verge AI", "https://www.theverge.com/rss/ai-artificial-intelligence/index.xml", "en", false, "tech"},
		{"VentureBeat AI", "https://venturebeat.com/category/ai/feed/", "en", false, "tech"},
		{"TechCrunch AI", "https://techcrunch.com/category/artificial-intelligence/feed/", "en", false, "tech"},
		{"Wired AI", "https://www.wired.com/feed/tag/ai/latest/rss", "en", false, "tech"},
	}

	var seeded int64
	for _, s := range defaults {
		source := &DbRssSource{
			Name:     s.Name,
			FeedURL:  s.FeedURL,
			Language: s.Language,
			Enabled:  s.Enabled,
			Category: s.Category,
		}
		if err := DB.Create(source).Error; err == nil {
			seeded++
		} else {
			// Ignore duplicate errors (if seed runs twice)
		}
	}
	return seeded, nil
}
