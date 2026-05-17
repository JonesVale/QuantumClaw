package router

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/middleware"
)

func SetRouter(router *gin.Engine, buildFS embed.FS) {
	router.Use(middleware.HTTPSRedirect())
	SetApiRouter(router)
	SetDashboardRouter(router)
	SetRelayRouter(router)
	frontendBaseUrl := os.Getenv("FRONTEND_BASE_URL")
	if config.IsMasterNode && frontendBaseUrl != "" {
		frontendBaseUrl = ""
		logger.SysLog("FRONTEND_BASE_URL is ignored on master node")
	}
	if frontendBaseUrl == "" {
		SetWebRouter(router, buildFS)
	} else {
		frontendBaseUrl = strings.TrimSuffix(frontendBaseUrl, "/")
		router.NoRoute(func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s%s", frontendBaseUrl, c.Request.RequestURI))
		})
	}
}
