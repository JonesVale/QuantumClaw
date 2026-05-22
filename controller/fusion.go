package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
)

// FusionRequest — 多模型编排请求
type FusionRequest struct {
	Prompt       string   `json:"prompt"`
	Models       []string `json:"models"`       // 目标模型列表
	SystemPrompt string   `json:"system_prompt,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
	Temperature  float64  `json:"temperature,omitempty"`
	Strategy     string   `json:"strategy"` // "fastest" | "cheapest" | "all"
}

// FusionResult — 单个模型的返回结果
type FusionResult struct {
	Model       string `json:"model"`
	Provider    string `json:"provider"`
	Content     string `json:"content"`
	LatencyMs   int64  `json:"latency_ms"`
	TokenCount  int    `json:"token_count"`
	Cost        float64 `json:"cost"`
	Status      string `json:"status"` // "success" | "error"
	Error       string `json:"error,omitempty"`
}

// FusionResponse — 融合编排响应
type FusionResponse struct {
	Results     []FusionResult `json:"results"`
	BestResult  *FusionResult  `json:"best_result,omitempty"`
	TotalTime   int64          `json:"total_time_ms"`
}

// Fusion 渠道类型名称映射
var fusionTypeNames map[int]string

func init() {
	fusionTypeNames = channeltype.ChannelTypeNames
}

// BuildProviderMap 从渠道列表构建 model→provider 映射
func BuildProviderMap() map[string]string {
	channels, _ := model.GetAllChannels(0, 0, "all")
	providerMap := make(map[string]string)
	for _, ch := range channels {
		if ch.Key == "" || strings.HasPrefix(ch.Key, "PUT_YOUR") {
			continue
		}
		providerName := ""
		if name, ok := fusionTypeNames[ch.Type]; ok {
			providerName = name
		}
		models := strings.Split(ch.Models, ",")
		for _, m := range models {
			m = strings.TrimSpace(m)
			if m != "" {
				providerMap[m] = providerName
			}
		}
	}
	return providerMap
}

// HandleFusion — 多模型融合编排处理器
func HandleFusion(c *gin.Context) {
	var req FusionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	if req.Prompt == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "prompt is required"})
		return
	}
	if len(req.Models) < 1 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "at least 1 model is required"})
		return
	}
	if req.Strategy == "" {
		req.Strategy = "fastest"
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = 1024
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}

	providerMap := BuildProviderMap()

	// 并行向所有模型发送请求
	type resultChan struct {
		index  int
		result FusionResult
	}

	results := make([]FusionResult, len(req.Models))
	ch := make(chan resultChan, len(req.Models))
	var wg sync.WaitGroup
	startTime := time.Now()

	// 需要获取 Bearer token（当前用户的 access_token）
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		// 尝试从 cookie 获取
		token, _ := c.Cookie("token")
		if token != "" {
			authHeader = "Bearer " + token
		}
	}

	for i, modelName := range req.Models {
		wg.Add(1)
		go func(idx int, model string) {
			defer wg.Done()

			modelStart := time.Now()
			result := FusionResult{
				Model:    model,
				Provider: providerMap[model],
				Status:   "error",
			}

			// 构建请求体
			messages := make([]map[string]string, 0)
			if req.SystemPrompt != "" {
				messages = append(messages, map[string]string{"role": "system", "content": req.SystemPrompt})
			}
			messages = append(messages, map[string]string{"role": "user", "content": req.Prompt})

			body := map[string]interface{}{
				"model":       model,
				"messages":    messages,
				"max_tokens":  req.MaxTokens,
				"temperature": req.Temperature,
			}
			bodyJSON, _ := json.Marshal(body)

			// 调用内部 API
			resp, err := http.Post(
				fmt.Sprintf("http://localhost:%d/v1/chat/completions", 3666),
				"application/json",
				strings.NewReader(string(bodyJSON)),
			)
			if err != nil {
				result.Error = err.Error()
				ch <- resultChan{idx, result}
				return
			}
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)

			latency := time.Since(modelStart).Milliseconds()
			result.LatencyMs = latency

			if resp.StatusCode != 200 {
				result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody))
				ch <- resultChan{idx, result}
				return
			}

			var apiResp struct {
				Choices []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
				Usage struct {
					TotalTokens int `json:"total_tokens"`
				} `json:"usage"`
			}
			if err := json.Unmarshal(respBody, &apiResp); err != nil {
				result.Error = "parse error: " + err.Error()
				ch <- resultChan{idx, result}
				return
			}

			if len(apiResp.Choices) > 0 {
				result.Content = apiResp.Choices[0].Message.Content
			}
			result.TokenCount = apiResp.Usage.TotalTokens
			result.Cost = float64(apiResp.Usage.TotalTokens) * 0.00001 // simplified
			result.Status = "success"
			ch <- resultChan{idx, result}
		}(i, modelName)
	}

	// 收集结果
	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		results[r.index] = r.result
	}

	totalTime := time.Since(startTime).Milliseconds()

	// 根据策略选择最佳结果
	var best *FusionResult
	successResults := make([]FusionResult, 0)
	for _, r := range results {
		if r.Status == "success" {
			successResults = append(successResults, r)
		}
	}

	if len(successResults) > 0 {
		switch req.Strategy {
		case "fastest":
			fastest := successResults[0]
			for _, r := range successResults {
				if r.LatencyMs < fastest.LatencyMs {
					fastest = r
				}
			}
			best = &fastest
		case "cheapest":
			cheapest := successResults[0]
			for _, r := range successResults {
				if r.Cost < cheapest.Cost {
					cheapest = r
				}
			}
			best = &cheapest
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": FusionResponse{
			Results:    results,
			BestResult: best,
			TotalTime:  totalTime,
		},
	})
}
