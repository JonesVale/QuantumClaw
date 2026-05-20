package middleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	relaycommon "github.com/quantumclaw/quantumclaw/relay/common"
)

func abortWithMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": helper.MessageWithRequestId(message, c.GetString(helper.RequestIdKey)),
			"type":    "quantumclaw_error",
		},
	})
	c.Abort()
	logger.Error(c.Request.Context(), message)
}

func getRequestModel(c *gin.Context) (string, error) {
	var modelRequest ModelRequest
	err := common.UnmarshalBodyReusable(c, &modelRequest)
	if err != nil {
		return "", fmt.Errorf("common.UnmarshalBodyReusable failed: %w", err)
	}

	// AI 请求的默认模型名
	if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "text-moderation-stable"
		}
	}
	if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
		if modelRequest.Model == "" {
			modelRequest.Model = c.Param("model")
		}
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "dall-e-2"
		}
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") || strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "whisper-1"
		}
	}

	// 量子算力：从 backend 字段提取后端名称作为请求模型
	if strings.HasPrefix(c.Request.URL.Path, "/v1/quantum") {
		if modelRequest.Model == "" {
			var qReq struct {
				Backend string `json:"backend"`
			}
			body, err := common.GetRequestBody(c)
			if err == nil && len(body) > 0 {
				json.Unmarshal(body, &qReq)
				if qReq.Backend != "" {
					modelRequest.Model = qReq.Backend
				}
			}
		}
		// 量子请求不进行后缀剥离
		return modelRequest.Model, nil
	}

	// 常规 AI 请求：剥离 reasoning/thinking 后缀
	if modelRequest.Model != "" {
		modelRequest.Model = relaycommon.StripModelSuffix(modelRequest.Model)
	}
	return modelRequest.Model, nil
}

func isModelInList(modelName string, models string) bool {
	modelList := strings.Split(models, ",")
	for _, model := range modelList {
		if modelName == model {
			return true
		}
	}
	return false
}
