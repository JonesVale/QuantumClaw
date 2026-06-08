package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
	"github.com/gin-gonic/gin"
)

// GetRssArticles returns paginated RSS articles filtered by language.
// GET /api/rss/articles?language=zh&limit=20&offset=0
func GetRssArticles(c *gin.Context) {
	language := c.DefaultQuery("language", "zh")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	articles, total, err := model.GetRssArticles(language, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch articles",
		})
		return
	}

	// If no articles found, return empty response rather than an error
	if articles == nil {
		articles = []model.RssArticle{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"articles": articles,
			"total":    total,
			"limit":    limit,
			"offset":   offset,
		},
	})
}

// ---------- RSS Source Management (Admin) ----------

// AdminGetRssSources returns all RSS sources with pagination.
// GET /api/admin/rss/sources?limit=20&offset=0
func AdminGetRssSources(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit < 1 || limit > 200 {
		limit = 50
	}

	sources, total, err := model.GetAllRssSources(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"sources": sources, "total": total, "limit": limit, "offset": offset},
	})
}

// AdminCreateRssSource creates a new RSS source.
// POST /api/admin/rss/sources
func AdminCreateRssSource(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		FeedURL  string `json:"feed_url" binding:"required"`
		Language string `json:"language"`
		Enabled  *bool  `json:"enabled"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	source := &model.DbRssSource{
		Name:     req.Name,
		FeedURL:  req.FeedURL,
		Language: req.Language,
		Category: req.Category,
	}
	if req.Language == "" {
		source.Language = "zh"
	}
	if req.Enabled != nil {
		source.Enabled = *req.Enabled
	} else {
		source.Enabled = true
	}
	if source.Category == "" {
		source.Category = "general"
	}

	if err := model.CreateRssSource(source); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": source})
}

// AdminUpdateRssSource updates an existing RSS source.
// PUT /api/admin/rss/sources/:id
func AdminUpdateRssSource(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	source, err := model.GetRssSourceById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Source not found"})
		return
	}

	var req struct {
		Name     *string `json:"name"`
		FeedURL  *string `json:"feed_url"`
		Language *string `json:"language"`
		Enabled  *bool   `json:"enabled"`
		Category *string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if req.Name != nil {
		source.Name = *req.Name
	}
	if req.FeedURL != nil {
		source.FeedURL = *req.FeedURL
	}
	if req.Language != nil {
		source.Language = *req.Language
	}
	if req.Enabled != nil {
		source.Enabled = *req.Enabled
	}
	if req.Category != nil {
		source.Category = *req.Category
	}

	if err := model.UpdateRssSource(source); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": source})
}

// AdminDeleteRssSource deletes an RSS source by ID.
// DELETE /api/admin/rss/sources/:id
func AdminDeleteRssSource(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	if err := model.DeleteRssSource(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "RSS source deleted"})
}

// AdminTriggerRssFetch manually triggers a fetch for a specific source (or all sources if id=0).
// POST /api/admin/rss/fetch?id=<source_id>
func AdminTriggerRssFetch(c *gin.Context) {
	idStr := c.DefaultQuery("id", "0")
	id, _ := strconv.Atoi(idStr)

	if id > 0 {
		// Fetch a specific source
		source, err := model.GetRssSourceById(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Source not found"})
			return
		}
		go func() {
			fetchErr := service.FetchSingleSource(source.Name, source.FeedURL, source.Language)
			errMsg := ""
			if fetchErr != nil {
				errMsg = fetchErr.Error()
			}
			model.UpdateRssSourceFetchStatus(id, time.Now(), errMsg)
			logger.SysLog(fmt.Sprintf("Manual RSS fetch [%s]: %s", source.Name, map[bool]string{true: "success", false: "error"}[errMsg == ""]))
		}()
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Fetch triggered for " + source.Name})
	} else {
		// Trigger full cycle via re-importing from DB
		go func() {
			total, success := service.FetchAllEnabledSources()
			logger.SysLog(fmt.Sprintf("Manual RSS fetch all: %d/%d successful", success, total))
		}()
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Full fetch cycle triggered"})
	}
}
