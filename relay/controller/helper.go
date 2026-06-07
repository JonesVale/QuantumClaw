package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/quantumclaw/quantumclaw/relay/constant/role"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	relaycommon "github.com/quantumclaw/quantumclaw/relay/common"
	"github.com/quantumclaw/quantumclaw/relay/common_handler"
	"github.com/quantumclaw/quantumclaw/service"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/openai"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"github.com/quantumclaw/quantumclaw/relay/controller/validator"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	relaymodel "github.com/quantumclaw/quantumclaw/relay/model"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
)

func getAndValidateTextRequest(c *gin.Context, relayMode int) (*relaymodel.GeneralOpenAIRequest, error) {
	textRequest := &relaymodel.GeneralOpenAIRequest{}
	err := common.UnmarshalBodyReusable(c, textRequest)
	if err != nil {
		return nil, err
	}
	if relayMode == relaymode.Moderations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayMode == relaymode.Embeddings && textRequest.Model == "" {
	}
	err = validator.ValidateTextRequest(textRequest, relayMode)
	if err != nil {
		return nil, err
	}
	return textRequest, nil
}

func getPromptTokens(textRequest *relaymodel.GeneralOpenAIRequest, relayMode int) int {
	switch relayMode {
	case relaymode.ChatCompletions:
		return openai.CountTokenMessages(textRequest.Messages, textRequest.Model)
	case relaymode.Completions:
		return openai.CountTokenInput(textRequest.Prompt, textRequest.Model)
	case relaymode.Moderations:
		return openai.CountTokenInput(textRequest.Input, textRequest.Model)
	}
	return 0
}

// ==================== 计费优先级链 ====================
// 优先级：订阅 → CashBalance → Quota
// （统一入口，替代旧版仅检查 CashBalance 的逻辑）

// preConsumeBalance — 预扣前检查，使用优先级链
// 返回：(preConsumedQuota, billingSource, error)
func preConsumeBalance(ctx context.Context, textRequest *relaymodel.GeneralOpenAIRequest, promptTokens int, ratio float64, meta *meta.Meta) (int64, string, *relaymodel.ErrorWithStatusCode) {

	// ── 详细审计日志：预扣前 ──
	logger.Infof(ctx, "[BILLING_AUDIT] preConsumeBalance 入口 | user=%d token=%d promptTokens=%d ratio=%.4f model=%s",
		meta.UserId, meta.TokenId, promptTokens, ratio, textRequest.Model)
	// ── 详细审计日志 END ──

	// 构建 RelayInfo 供 common_handler 使用
	relayInfo := &relaycommon.RelayInfo{
		UserID:   meta.UserId,
		TokenID:  meta.TokenId,
	}
	preConsumedQuota, billingSource, err := common_handler.PreConsumeBilling(ctx, relayInfo, promptTokens, ratio)

	// ── 详细审计日志：预扣结果 ──
	if err != nil {
		logger.Warnf(ctx, "[BILLING_AUDIT] preConsumeBalance 失败 | user=%d billingSource=%s preQuota=%d err=%v",
			meta.UserId, billingSource, preConsumedQuota, err)
		return preConsumedQuota, billingSource, openai.ErrorWrapper(err, "insufficient_balance", http.StatusForbidden)
	}
	logger.Infof(ctx, "[BILLING_AUDIT] preConsumeBalance 成功 | user=%d billingSource=%s preQuota=%d",
		meta.UserId, billingSource, preConsumedQuota)

	// ── 成本倒挂预检（在模型调用之前）──
	// 如果预计上游成本高于用户付费，直接拒绝，防止平台亏损
	if preConsumedQuota > 0 {
		channel, chErr := model.GetChannelById(meta.ChannelId, true)
		if chErr == nil {
			estimatedPriceCents := service.QuotaToUserPrice(preConsumedQuota)
			if invErr := service.CheckCostInversion(ctx, estimatedPriceCents, preConsumedQuota, channel, textRequest.Model); invErr != nil {
				logger.Errorf(ctx, "[COST_INVERSION_BLOCKED] 预检拒绝: %v", invErr)
				return preConsumedQuota, billingSource, openai.ErrorWrapper(invErr, "cost_inversion", http.StatusForbidden)
			}
		}
	}
	// ── 详细审计日志 END ──

	return preConsumedQuota, billingSource, nil
}

func postConsumeDeduct(ctx context.Context, usage *relaymodel.Usage, meta *meta.Meta, textRequest *relaymodel.GeneralOpenAIRequest, ratio float64, preConsumedQuota int64, modelRatio float64, groupRatio float64, systemPromptReset bool, billingSource string) {
	if err := service.PostConsumeDeduct(ctx, meta, usage, textRequest, ratio, modelRatio, groupRatio, preConsumedQuota, systemPromptReset, billingSource); err != nil {
		logger.Error(ctx, fmt.Sprintf("post consume deduct failed: %v", err))
	}
}

func getPreConsumedQuota(textRequest *relaymodel.GeneralOpenAIRequest, promptTokens int, ratio float64) int64 {
	preConsumedTokens := config.PreConsumedQuota + int64(promptTokens)
	if textRequest.MaxTokens != 0 {
		preConsumedTokens += int64(textRequest.MaxTokens)
	}
	return int64(float64(preConsumedTokens) * ratio)
}

func getMappedModelName(modelName string, mapping map[string]string) (string, bool) {
	if mapping == nil {
		return modelName, false
	}
	mappedModelName := mapping[modelName]
	if mappedModelName != "" {
		return mappedModelName, true
	}
	return modelName, false
}

func isErrorHappened(meta *meta.Meta, resp *http.Response) bool {
	if resp == nil {
		if meta.ChannelType == channeltype.AwsClaude {
			return false
		}
		return true
	}
	if resp.StatusCode != http.StatusOK &&
		// replicate return 201 to create a task
		resp.StatusCode != http.StatusCreated {
		return true
	}
	if meta.ChannelType == channeltype.DeepL {
		// skip stream check for deepl
		return false
	}

	if meta.IsStream && strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") &&
		// Even if stream mode is enabled, replicate will first return a task info in JSON format,
		// requiring the client to request the stream endpoint in the task info
		meta.ChannelType != channeltype.Replicate &&
		meta.ChannelType != channeltype.Sub2API {
		return true
	}
	return false
}

func setSystemPrompt(ctx context.Context, request *relaymodel.GeneralOpenAIRequest, prompt string) (reset bool) {
	if prompt == "" {
		return false
	}
	if len(request.Messages) == 0 {
		return false
	}
	if request.Messages[0].Role == role.System {
		request.Messages[0].Content = prompt
		logger.Infof(ctx, "rewrite system prompt")
		return true
	}
	request.Messages = append([]relaymodel.Message{{
		Role:    role.System,
		Content: prompt,
	}}, request.Messages...)
	logger.Infof(ctx, "add system prompt")
	return true
}
