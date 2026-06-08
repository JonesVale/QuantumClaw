package controller

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/model"
)

// getSiteURL returns the public site base URL for SEO/link generation.
// Reads from SITE_URL env var, falls back to qscl.link.
// This ensures correct URLs when deploying to overseas servers with different domains.
func getSiteURL() string {
	if url := os.Getenv("SITE_URL"); url != "" {
		return strings.TrimRight(url, "/")
	}
	return "https://qscl.link"
}

// GetSiteInfo 公开站点信息（用于首页 footer）
// GET /api/site-info
func GetSiteInfo(c *gin.Context) {
	config.OptionMapRWMutex.RLock()
	companyURL := config.OptionMap["company_website_url"]
	icpBeian := config.OptionMap["icp_beian"]
	config.OptionMapRWMutex.RUnlock()

	if companyURL == "" {
		companyURL = "https://www.ctji.cn"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"company_website_url": companyURL,
			"icp_beian":           icpBeian,
		},
	})
}

// GetSitemap 生成动态 XML Sitemap
// GET /sitemap.xml
func GetSitemap(c *gin.Context) {
	c.Header("Content-Type", "application/xml")

	baseURL := getSiteURL()
	now := time.Now().Format("2006-01-02")

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)

	// ── 静态页面（优先级高） ──
	staticPages := []struct {
		Path     string
		Priority string
		Changef  string
	}{
		{"/", "1.0", "weekly"},
		{"/models", "0.9", "daily"},
		{"/pricing", "0.9", "weekly"},
		{"/rankings", "0.8", "daily"},
		{"/marketplace", "0.8", "weekly"},
		{"/download", "0.8", "monthly"},
		{"/apps", "0.7", "weekly"},
		{"/faq", "0.7", "monthly"},
		{"/api-docs", "0.7", "monthly"},
		{"/about", "0.5", "monthly"},
		{"/enterprise", "0.7", "weekly"},
		{"/playground", "0.6", "monthly"},
		{"/news", "0.8", "daily"},
		{"/quantum", "0.6", "weekly"},
	}

	for _, p := range staticPages {
		sb.WriteString(fmt.Sprintf(`<url><loc>%s%s</loc><lastmod>%s</lastmod><changefreq>%s</changefreq><priority>%s</priority></url>`,
			baseURL, p.Path, now, p.Changef, p.Priority))
	}

	// ── 模型详情页（动态从数据库获取） ──
	models, err := model.GetAllModelMetadatas()
	if err == nil {
		for _, m := range models {
			if m.ModelName != "" {
				slug := strings.ToLower(strings.ReplaceAll(m.ModelName, " ", "-"))
				sb.WriteString(fmt.Sprintf(`<url><loc>%s/models/%s</loc><lastmod>%s</lastmod><changefreq>weekly</changefreq><priority>0.7</priority></url>`,
					baseURL, slug, now))
			}
		}
	}

	// ── 渠道商店铺页 ──
	stores, err := model.GetAllActiveStores()
	if err == nil {
		for _, s := range stores {
			sb.WriteString(fmt.Sprintf(`<url><loc>%s/stores/%s</loc><lastmod>%s</lastmod><changefreq>weekly</changefreq><priority>0.6</priority></url>`,
				baseURL, s.StoreSlug, now))
		}
	}

	// ── SEO 文章 ──
	articles, _, err := model.GetRssArticles("zh", 200, 0)
	if err == nil {
		for _, a := range articles {
			sb.WriteString(fmt.Sprintf(`<url><loc>%s/news?article=%d</loc><lastmod>%s</lastmod><changefreq>monthly</changefreq><priority>0.6</priority></url>`,
				baseURL, a.Id, a.PublishedAt.Format("2006-01-02")))
		}
	}

	sb.WriteString(`</urlset>`)
	c.String(http.StatusOK, sb.String())
}

// AdminCreateRssArticle 管理员创建 SEO 文章
// POST /api/admin/rss/articles
func AdminCreateRssArticle(c *gin.Context) {
	role := c.GetInt("role")
	if role < 10 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "权限不足"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Author      string `json:"author"`
		Language    string `json:"language"`
		Source      string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	if req.Author == "" {
		req.Author = "QuantumClaw Team"
	}
	if req.Language == "" {
		req.Language = "zh"
	}
	if req.Source == "" {
		req.Source = "seo"
	}

	// 生成唯一链接（基于标题的 slug）
	slug := strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
	link := fmt.Sprintf("%s/news?article=%s", getSiteURL(), slug)

	article := &model.RssArticle{
		Source:      req.Source,
		Title:       req.Title,
		Link:        link,
		Description: req.Content,
		Author:      req.Author,
		Language:    req.Language,
		PublishedAt: time.Now(),
	}

	if err := model.AddRssArticle(article); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建文章失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": article})
}

// GetRobotsTxt returns robots.txt for search engine crawlers.
// GET /robots.txt
func GetRobotsTxt(c *gin.Context) {
	c.Header("Content-Type", "text/plain")
	content := fmt.Sprintf(`User-agent: *
Allow: /
Disallow: /api/
Disallow: /admin/
Disallow: /_authenticated/
Sitemap: %s/sitemap.xml
`, getSiteURL())
	c.String(http.StatusOK, content)
}

// GetNewsFeed returns an RSS/Atom feed of the latest news articles (for feed readers).
// GET /news/feed.xml or /feed.xml
func GetNewsFeed(c *gin.Context) {
	articles, _, err := model.GetRssArticles("all", 50, 0)
	if err != nil || len(articles) == 0 {
		c.String(http.StatusNotFound, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel><title>QuantumClaw News</title><link>%s</link><description>No articles available yet.</description></channel></rss>`, getSiteURL()))
		return
	}

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">`)
	sb.WriteString(`<channel>`)
	sb.WriteString(`<title>QuantumClaw AI News | 量子灵爪 AI 资讯</title>`)
	sb.WriteString(`<link>`)
	sb.WriteString(getSiteURL())
	sb.WriteString(`/news</link>`)
	sb.WriteString(`<description>Latest AI and Quantum Computing industry news from QuantumClaw.</description>`)
	sb.WriteString(`<language>zh-cn</language>`)
	sb.WriteString(fmt.Sprintf(`<atom:link href="%s/news/feed.xml" rel="self" type="application/rss+xml" />`, getSiteURL()))

	for _, a := range articles {
		pubDate := a.PublishedAt.Format(time.RFC1123Z)
		desc := a.Description
		if len(desc) > 300 {
			desc = desc[:300] + "..."
		}
		// Escape XML special characters in description
		desc = strings.ReplaceAll(desc, "&", "&amp;")
		desc = strings.ReplaceAll(desc, "<", "&lt;")
		desc = strings.ReplaceAll(desc, ">", "&gt;")
		title := strings.ReplaceAll(a.Title, "&", "&amp;")
		title = strings.ReplaceAll(title, "<", "&lt;")
		title = strings.ReplaceAll(title, ">", "&gt;")
		sb.WriteString(fmt.Sprintf(`<item><title>%s</title><link>%s</link><guid isPermaLink="false">%s</guid><pubDate>%s</pubDate><description>%s</description></item>`,
			title, a.Link, a.Link, pubDate, desc))
	}

	sb.WriteString(`</channel></rss>`)
	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.String(http.StatusOK, sb.String())
}
