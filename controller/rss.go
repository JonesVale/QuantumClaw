package controller

import (
	"net/http"
	"strconv"

	"github.com/quantumclaw/quantumclaw/model"
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
