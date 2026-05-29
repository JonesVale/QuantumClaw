package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"gorm.io/gorm"
)

// ── GitHub trending AI apps sync ──────────────────────────
// 每月自动从 GitHub 获取流行 AI 应用，增量入库（不删除已有数据）

var (
	appSyncOnce sync.Once
	appCategoryKeywords = map[string]string{
		"chat":       "ai-chat,llm-ui,chatbot",
		"development": "ai-coding,code-assistant,developer-tool",
		"platform":   "ai-platform,llm-platform,ai-framework",
		"tools":      "ai-tool,llm-tool,ai-utility",
		"agent":      "ai-agent,autonomous-agent,ai-assistant",
	}
)

type appGhSearchResult struct {
	Items []appGhRepo `json:"items"`
}

type appGhRepo struct {
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	HTMLURL     string   `json:"html_url"`
	Stars       int      `json:"stargazers_count"`
	Topics      []string `json:"topics"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// SyncPopularApps 同步流行 AI 应用
// 从 GitHub 搜索热门 AI 相关仓库，增量入库为 published 状态
func SyncPopularApps(ctx context.Context) error {
	logger.SysLog("[AppSync] 开始同步流行 AI 应用...")

	totalAdded := 0
	for category, keywords := range appCategoryKeywords {
		added, err := syncCategory(ctx, category, keywords)
		if err != nil {
			logger.SysError(fmt.Sprintf("[AppSync] %s 同步失败: %v", category, err))
			continue
		}
		totalAdded += added
	}

	logger.SysLog(fmt.Sprintf("[AppSync] 同步完成, 新增 %d 个应用", totalAdded))
	return nil
}

func syncCategory(ctx context.Context, category, keywords string) (int, error) {
	// 搜索 GitHub: stars > 1000, 按 stars 排序, 取前 15
	// 限定 AI 主题仓库
	query := fmt.Sprintf("stars:>1000 topics:>1 topic:%s sort:stars-desc", strings.Split(keywords, ",")[0])

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/search/repositories", nil)
	if err != nil {
		return 0, err
	}

	q := req.URL.Query()
	q.Set("q", query)
	q.Set("sort", "stars")
	q.Set("order", "desc")
	q.Set("per_page", "15")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "QuantumClaw/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("GitHub API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("GitHub API 返回 %d: %s", resp.StatusCode, string(body[:200]))
	}

	var result appGhSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("JSON 解析失败: %w", err)
	}

	added := 0
	for _, repo := range result.Items {
		if repo.Description == "" {
			continue
		}

		// 检查是否已存在（按 name+author 去重）
		var existing model.AppMarket
		err := model.DB.Where("name = ? AND author = ?", repo.FullName, repo.Owner.Login).
			Or("app_url = ?", repo.HTMLURL).
			First(&existing).Error

		if err == nil {
			// 已存在 → 更新用户数和描述
			model.DB.Model(&existing).Updates(map[string]interface{}{
				"description": truncateText(repo.Description, 500),
				"updated_at":  time.Now().Unix(),
			})
			continue
		}

		if err != gorm.ErrRecordNotFound {
			logger.SysError(fmt.Sprintf("[AppSync] 查询出错: %v", err))
			continue
		}

		// 不存在 → 新增
		app := model.AppMarket{
			Name:        repo.FullName,
			Description: truncateText(repo.Description, 500),
			Author:      repo.Owner.Login,
			AuthorURL:   fmt.Sprintf("https://github.com/%s", repo.Owner.Login),
			AppURL:      repo.HTMLURL,
			IconURL:     fmt.Sprintf("https://github.com/%s.png?size=64", repo.Owner.Login),
			Category:    category,
			Status:      "published",
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}

		if err := model.DB.Create(&app).Error; err != nil {
			logger.SysError(fmt.Sprintf("[AppSync] 插入失败 %s: %v", repo.FullName, err))
			continue
		}
		added++
	}

	return added, nil
}

func truncateText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
