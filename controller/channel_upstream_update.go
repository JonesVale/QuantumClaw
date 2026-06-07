package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// ChannelUpstreamUpdateSetting 渠道上游模型更新检测设置
type ChannelUpstreamUpdateSetting struct {
	Enabled       bool `json:"enabled"`
	CheckInterval int  `json:"check_interval"` // 分钟
	LastCheckTime int64 `json:"last_check_time"`

	// 是否在检测到更新时自动更新本地模型列表
	AutoUpdateModels bool `json:"auto_update_models"`
	// 只检查这些分组
	WatchGroups []string `json:"watch_groups"`
}

var channelUpstreamUpdateSetting = &ChannelUpstreamUpdateSetting{
	Enabled:           false,
	CheckInterval:     60,
	AutoUpdateModels:  false,
	WatchGroups:       []string{},
}

var upstreamUpdateMu sync.Mutex

// GetUpstreamUpdateSetting 获取设置
func GetUpstreamUpdateSetting(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": channelUpstreamUpdateSetting})
}

// SaveUpstreamUpdateSetting 保存设置
func SaveUpstreamUpdateSetting(c *gin.Context) {
	var req ChannelUpstreamUpdateSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	upstreamUpdateMu.Lock()
	channelUpstreamUpdateSetting.Enabled = req.Enabled
	channelUpstreamUpdateSetting.CheckInterval = req.CheckInterval
	channelUpstreamUpdateSetting.AutoUpdateModels = req.AutoUpdateModels
	channelUpstreamUpdateSetting.WatchGroups = req.WatchGroups
	upstreamUpdateMu.Unlock()
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// CheckUpstreamUpdates 手动触发上游检测
func CheckUpstreamUpdates(c *gin.Context) {
	upstreamUpdateMu.Lock()
	enabled := channelUpstreamUpdateSetting.Enabled
	upstreamUpdateMu.Unlock()

	if !enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "上游检测未启用"})
		return
	}

	result, err := doUpstreamCheck()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "检测失败: " + err.Error()})
		return
	}

	channelUpstreamUpdateSetting.LastCheckTime = time.Now().Unix()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// UpstreamCheckResult 检测结果
type UpstreamCheckResult struct {
	ChannelsChecked int                      `json:"channels_checked"`
	UpdatesFound    int                      `json:"updates_found"`
	ChannelUpdates  []ChannelUpdateInfo     `json:"channel_updates"`
	Errors          []string                 `json:"errors"`
}

type ChannelUpdateInfo struct {
	ChannelId   int      `json:"channel_id"`
	ChannelName string   `json:"channel_name"`
	OldModels   []string `json:"old_models"`
	NewModels   []string `json:"new_models"`
	Added       []string `json:"added"`
	Removed     []string `json:"removed"`
}

func doUpstreamCheck() (*UpstreamCheckResult, error) {
	result := &UpstreamCheckResult{}

	// 获取需要检查的渠道
	upstreamUpdateMu.Lock()
	watchGroups := channelUpstreamUpdateSetting.WatchGroups
	autoUpdate := channelUpstreamUpdateSetting.AutoUpdateModels
	upstreamUpdateMu.Unlock()

	var channels []model.Channel
	db := model.DB.Where("status = ?", model.ChannelStatusEnabled)
	if len(watchGroups) > 0 {
		groupList := strings.Join(watchGroups, "','")
		db = db.Where("`group` IN ('" + groupList + "')")
	}
	db.Find(&channels)

	result.ChannelsChecked = len(channels)

	for _, ch := range channels {
		update, err := checkChannelUpstream(&ch)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("渠道 %d (%s): %v", ch.Id, ch.Name, err))
			continue
		}
		if update != nil && (len(update.Added) > 0 || len(update.Removed) > 0) {
			result.UpdatesFound++
			result.ChannelUpdates = append(result.ChannelUpdates, *update)

			if autoUpdate {
				applyChannelUpdate(update)
			}
		}
	}

	return result, nil
}

// UpstreamModelResponse 上游渠道模型列表响应（格式因渠道而异）
type UpstreamModelResponse struct {
	Data struct {
		Models []struct {
			ID string `json:"id"`
		} `json:"data"`
	} `json:"data"`
}

func checkChannelUpstream(ch *model.Channel) (*ChannelUpdateInfo, error) {
	if ch.BaseURL == nil || *ch.BaseURL == "" || ch.Key == "" {
		return nil, nil
	}

	// 检测支持的模型列表（通用 OpenAI 格式）
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	baseURL := strings.TrimRight(*ch.BaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ch.Key)
	req.Header.Set("User-Agent", "QuantumClaw/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var upstreamResp UpstreamModelResponse
	if err := json.Unmarshal(body, &upstreamResp); err != nil {
		return nil, err
	}

	newModels := make([]string, 0, len(upstreamResp.Data.Models))
	for _, m := range upstreamResp.Data.Models {
		if m.ID != "" {
			newModels = append(newModels, m.ID)
		}
	}

	update := &ChannelUpdateInfo{
		ChannelId:   ch.Id,
		ChannelName: ch.Name,
		OldModels:   strings.Split(ch.Models, ","),
		NewModels:   newModels,
	}

	// 找出新增和删除的模型
	oldSet := make(map[string]bool)
	for _, m := range update.OldModels {
		oldSet[strings.TrimSpace(m)] = true
	}
	for _, m := range newModels {
		if !oldSet[m] {
			update.Added = append(update.Added, m)
		}
	}
	for _, m := range update.OldModels {
		m = strings.TrimSpace(m)
		found := false
		for _, n := range newModels {
			if n == m {
				found = true
				break
			}
		}
		if !found && m != "" {
			update.Removed = append(update.Removed, m)
		}
	}

	return update, nil
}

func applyChannelUpdate(update *ChannelUpdateInfo) {
	if len(update.NewModels) == 0 {
		return
	}
	newModels := strings.Join(update.NewModels, ",")
	model.DB.Model(&model.Channel{}).Where("id = ?", update.ChannelId).
		UpdateColumn("models", newModels)
	logger.SysLogf("渠道 %d (%s) 模型列表已自动更新: 新增=%d 移除=%d",
		update.ChannelId, update.ChannelName, len(update.Added), len(update.Removed))
}

// StartUpstreamUpdateCron 启动上游检测定时任务
func StartUpstreamUpdateCron() {
	upstreamUpdateMu.Lock()
	enabled := channelUpstreamUpdateSetting.Enabled
	interval := channelUpstreamUpdateSetting.CheckInterval
	upstreamUpdateMu.Unlock()

	if !enabled || interval <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			upstreamUpdateMu.Lock()
			stillEnabled := channelUpstreamUpdateSetting.Enabled
			upstreamUpdateMu.Unlock()
			if stillEnabled {
				result, err := doUpstreamCheck()
				if err != nil {
					logger.SysError("上游检测定时任务失败: " + err.Error())
				} else if result.UpdatesFound > 0 {
					logger.SysLogf("上游检测定时任务: 检查 %d 渠道，发现 %d 个更新",
						result.ChannelsChecked, result.UpdatesFound)
				}
			}
		}
	}()
}
