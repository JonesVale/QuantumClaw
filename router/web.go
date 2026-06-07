package router

import (
	"embed"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/controller"
	"github.com/quantumclaw/quantumclaw/middleware"
	"net/http"
	"os"
	"strings"
)

func SetWebRouter(router *gin.Engine, buildFS embed.FS) {
	fmt.Println("[SetWebRouter] ENTERED")
	// 开发模式：从文件系统读取（支持热重载）
	// 生产模式：从 embed.FS 读取（单二进制部署）
	distPath := "web/default/dist"
	// 检查 DEBUG 或 GIN_MODE 环境变量
	debugMode := strings.ToLower(os.Getenv("GIN_MODE")) == "debug" || strings.ToLower(os.Getenv("DEBUG")) == "true"
	fmt.Println("[SetWebRouter] debugMode:", debugMode, "GIN_MODE:", os.Getenv("GIN_MODE"), "DEBUG:", os.Getenv("DEBUG"))
	
	if debugMode {
		// 开发模式：从文件系统读取 index.html
		fmt.Println("[DEBUG] Serving static files from filesystem:", distPath)
		router.Use(gzip.Gzip(gzip.DefaultCompression))
		router.Use(middleware.GlobalWebRateLimit())
		router.Use(middleware.Cache())
		router.Use(middleware.SecurityHeaders())
		router.Use(static.Serve("/", static.LocalFile(distPath, false)))
		router.NoRoute(func(c *gin.Context) {
			uri := c.Request.RequestURI
			if strings.HasPrefix(uri, "/v1") || strings.HasPrefix(uri, "/api/") || uri == "/api" {
				controller.RelayNotFound(c)
				return
			}
			// 从文件系统读取最新的 index.html
			data, err := os.ReadFile(distPath + "/index.html")
			if err != nil {
				c.String(http.StatusInternalServerError, "index.html not found: %v", err)
				return
			}
			c.Header("Cache-Control", "no-cache")
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})
	} else {
		// 生产模式：从 embed.FS 读取
		fmt.Println("[PROD] Serving static files from embed.FS")
		indexPageData, _ := buildFS.ReadFile(fmt.Sprintf("web/default/dist/index.html"))
		router.Use(gzip.Gzip(gzip.DefaultCompression))
		router.Use(middleware.GlobalWebRateLimit())
		router.Use(middleware.Cache())
		router.Use(middleware.SecurityHeaders())
		router.Use(static.Serve("/", common.EmbedFolder(buildFS, "web/default/dist")))
		router.NoRoute(func(c *gin.Context) {
			uri := c.Request.RequestURI
			if strings.HasPrefix(uri, "/v1") || strings.HasPrefix(uri, "/api/") || uri == "/api" {
				controller.RelayNotFound(c)
				return
			}
			c.Header("Cache-Control", "no-cache")
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexPageData)
		})
	}
}