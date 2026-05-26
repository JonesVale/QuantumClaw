package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetBrandRankings returns industry-wide brand usage rankings.
// GET /api/brand-rankings
func GetBrandRankings(c *gin.Context) {
	var rankings []model.BrandRanking
	if err := model.DB.Order("`rank` ASC").Find(&rankings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rankings,
	})
}
