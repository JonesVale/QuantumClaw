package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetSiteInfo 公开站点信息（用于首页 footer）
// GET /api/site-info
func GetSiteInfo(c *gin.Context) {
	// 从 OptionMap 读取公开站点配置
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

	baseURL := "https://qscl.link"
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
			if m.Name != "" {
				slug := strings.ToLower(strings.ReplaceAll(m.Name, " ", "-"))
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
	link := fmt.Sprintf("https://qscl.link/news?article=%s", slug)

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
