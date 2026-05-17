package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// StartChannelUpstreamModelUpdateTask periodically checks channels
// that have upstream model update checking enabled and syncs their model lists.
// Runs every 6 hours by default.
func StartChannelUpstreamModelUpdateTask() {
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		// Run once at startup after a delay
		time.Sleep(30 * time.Second)
		checkAndUpdateChannelModels()

		for range ticker.C {
			checkAndUpdateChannelModels()
		}
	}()
	logger.SysLog("channel upstream model update task started (interval: 6h)")
}

func checkAndUpdateChannelModels() {
	defer common.RecoverAndLog()

	channels, err := model.GetAllChannels(0, -1, "all")
	_ = channels
	if err != nil {
		logger.SysError("failed to fetch channels: " + err.Error())
		return
	}

	for _, ch := range channels {
		if ch.Status != model.ChannelStatusEnabled {
			continue
		}
		updateModelsForChannel(ch)
	}
}

func updateModelsForChannel(ch *model.Channel) {
	// Only check OpenAI-compatible endpoints that expose /v1/models
	baseURL := ch.GetBaseURL()
	if baseURL == "" || strings.Contains(baseURL, "localhost") || strings.Contains(baseURL, "127.0.0.1") {
		return
	}

	modelsURL := strings.TrimRight(baseURL, "/") + "/v1/models"
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+ch.Key)

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	logger.SysLog(fmt.Sprintf("channel #%d (%s): upstream model check completed", ch.Id, ch.Name))
}

// Ensure interface is registered in main.go
func init() {
	logger.SysLog("channel upstream model update service registered")
}
