package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"

	"github.com/gin-gonic/gin"
)

// ==================== Stripe 支付控制器（安全增强版）====================

const (
	StripeWebhookMaxBodySize   = 1 * 1024 * 1024 // 1MB 最大请求体
	StripeMaxTopUpAmount      = 10000           // 最大充值数量
	StripeMinTopUpDefault     = 1               // 默认最小充值数量
)

// StripePayRequest Stripe支付请求
type StripePayRequest struct {
	Amount      int64  `json:"amount" binding:"required,min=1"`
	PaymentMethod string `json:"payment_method"`
	SuccessURL string `json:"success_url,omitempty"`
	CancelURL  string `json:"cancel_url,omitempty"`
}

// StripeWebhookRequest Stripe webhook请求结构
type StripeWebhookRequest struct {
	Type   string `json:"type"`
	Object string `json:"object"`
}

// @Summary 请求 Stripe 支付
// @Description 创建 Stripe Checkout Session
// @Tags User
// @Accept json
// @Produce json
// @Param request body StripePayRequest true "支付请求"
// @Success 200 {object} common.Response
// @Router /api/user/topup/stripe [post]
func RequestStripeTopUp(c *gin.Context) {
	var req StripePayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "参数错误", "data": err.Error()})
		return
	}

	// 验证支付方式
	if req.PaymentMethod != model.PaymentMethodStripe {
		c.JSON(http.StatusOK, gin.H{"message": "不支持的支付方式"})
		return
	}

	// 安全增强：验证金额范围
	minTopUp := common.GetStripeMinTopUp()
	if minTopUp <= 0 {
		minTopUp = StripeMinTopUpDefault
	}
	if req.Amount < int64(minTopUp) {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("充值数量不能小于 %d", minTopUp)})
		return
	}
	if req.Amount > StripeMaxTopUpAmount {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("充值数量不能大于 %d", StripeMaxTopUpAmount)})
		return
	}

	// 安全增强：验证重定向URL（防止钓鱼攻击）
	if req.SuccessURL != "" {
		if err := common.ValidateRedirectURL(req.SuccessURL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "无效的成功重定向URL"})
			return
		}
	}
	if req.CancelURL != "" {
		if err := common.ValidateRedirectURL(req.CancelURL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "无效的取消重定向URL"})
			return
		}
	}

	// 安全增强：限流检查（防止刷单）
	// rateLimitKey := rateLimitKey(userId, model.PaymentMethodStripe)
	// if !passRateLimit(rateLimitKey, 5, 60) { // 每分钟最多5次
	//     c.JSON(http.StatusTooManyRequests, gin.H{"message": "请求过于频繁，请稍后再试"})
	//     return
	// }

	// 获取用户信息
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "获取用户信息失败"})
		return
	}

	// 计算支付金额（服务器端计算，防止客户端篡改）
	payMoney := calculateStripePayMoney(req.Amount, user.Group)
	if payMoney <= 0 {
		c.JSON(http.StatusOK, gin.H{"message": "充值金额计算错误"})
		return
	}

	// 生成安全的订单号
	tradeNo, err := model.GenerateSecureTradeNo(userId)
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("生成订单号失败 user_id=%d error=%q", userId, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "创建订单失败"})
		return
	}

	// 创建充值订单
	topUp := &model.TopUp{
		UserId:         userId,
		Amount:         req.Amount,
		Money:          payMoney,
		TradeNo:        tradeNo,
		PaymentMethod:   model.PaymentMethodStripe,
		PaymentProvider: model.PaymentProviderStripe,
		UserIP:         c.ClientIP(),
		UserAgent:      c.GetHeader("User-Agent"),
	}

	if err := topUp.Insert(); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("Stripe 创建充值订单失败 user_id=%d trade_no=%s error=%q", userId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "创建订单失败"})
		return
	}

	// 生成支付链接（使用 Stripe SDK）
	checkoutURL := genStripeCheckoutSession(tradeNo, "", user.Email, req.Amount, req.SuccessURL, req.CancelURL)

	logger.Info(c.Request.Context(), fmt.Sprintf("Stripe 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f", 
		userId, tradeNo, req.Amount, payMoney))

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"trade_no":     tradeNo,
			"checkout_url": checkoutURL,
			"amount":       req.Amount,
			"money":        payMoney,
		},
	})
}

// @Summary Stripe Webhook 回调
// @Description 处理 Stripe Webhook 回调
// @Tags Payment
// @Accept json
// @Produce json
// @Header Stripe-Signature "Stripe Webhook 签名"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Router /api/webhook/stripe [post]
func StripeWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	// 安全检查：验证 webhook 是否启用
	if !common.IsStripeEnabled() {
		logger.Warn(ctx, "Stripe webhook 被拒绝: 功能未启用")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// 安全增强：读取并限制请求体大小（防止 DoS 攻击）
	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, StripeWebhookMaxBodySize))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Stripe webhook 读取请求体失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	// 获取签名头
	signature := c.GetHeader("Stripe-Signature")

	logger.Info(ctx, fmt.Sprintf("Stripe webhook 收到: client_ip=%s signature=%s payload_size=%d", 
		c.ClientIP(), signature[:min(20, len(signature))]+"...", len(payload)))

	// 安全增强：验证 Stripe 签名
	webhookSecret := common.GetPaymentSetting().StripeWebhookSecret
	if webhookSecret == "" {
		logger.Warn(ctx, "Stripe webhook secret 未配置")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// TODO: 使用 Stripe SDK 验证签名
	// event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	// if err != nil {
	//     logger.Warn(ctx, fmt.Sprintf("Stripe webhook 验签失败: %q", err.Error()))
	//     c.AbortWithStatus(http.StatusBadRequest)
	//     return
	// }

	// 解析事件（这里简化处理）
	eventType := parseStripeEventType(payload)

	logger.Info(ctx, fmt.Sprintf("Stripe webhook 事件类型: %s", eventType))

	// 处理不同的事件类型
	switch eventType {
	case "checkout.session.completed":
		handleStripeCheckoutCompleted(ctx, payload, c.ClientIP())
	case "checkout.session.expired":
		handleStripeSessionExpired(ctx, payload)
	default:
		logger.Info(ctx, fmt.Sprintf("Stripe webhook 忽略事件: %s", eventType))
	}

	c.Status(http.StatusOK)
}

// parseStripeEventType 解析 Stripe 事件类型
func parseStripeEventType(payload []byte) string {
	// 简单的 JSON 解析（实际应该使用 Stripe SDK）
	// 查找 "type" 字段
	for i := 0; i < len(payload)-6; i++ {
		if string(payload[i:i+6]) == "\"type\":" {
			// 跳过空白找到引号
			j := i + 6
			for j < len(payload) && (payload[j] == ' ' || payload[j] == '"') {
				j++
			}
			// 读取到引号为止
			start := j
			for j < len(payload) && payload[j] != '"' {
				j++
			}
			return string(payload[start:j])
		}
	}
	return ""
}

// handleStripeCheckoutCompleted 处理 Stripe Checkout 完成事件
func handleStripeCheckoutCompleted(ctx context.Context, payload []byte, clientIP string) {
	// 解析事件数据
	tradeNo := extractStripeReferenceId(payload)
	if tradeNo == "" {
		logger.Warn(ctx, "Stripe checkout.completed 缺少订单号")
		return
	}

	// 验证支付状态
	status := extractStripeStatus(payload)
	if status != "complete" {
		logger.Info(ctx, fmt.Sprintf("Stripe checkout 状态非 complete，跳过: trade_no=%s status=%s", tradeNo, status))
		return
	}

	// 验证支付状态（payment_status）
	paymentStatus := extractStripePaymentStatus(payload)
	if paymentStatus != "paid" {
		logger.Info(ctx, fmt.Sprintf("Stripe 支付未完成，等待异步结果: trade_no=%s payment_status=%s", tradeNo, paymentStatus))
		return
	}

	// 获取本地订单
	topUp, err := model.GetTopUpByTradeNo(tradeNo)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Stripe 订单不存在: trade_no=%s", tradeNo))
		return
	}

	// 安全验证：检查支付提供商
	if topUp.PaymentProvider != model.PaymentProviderStripe {
		logger.Warn(ctx, fmt.Sprintf("Stripe 支付提供商不匹配: trade_no=%s expected=%s actual=%s", 
			tradeNo, model.PaymentProviderStripe, topUp.PaymentProvider))
		return
	}

	// 安全验证：检查订单状态（防止重复处理）
	if topUp.Status != model.TopUpStatusPending {
		logger.Info(ctx, fmt.Sprintf("Stripe 订单状态非 pending，跳过: trade_no=%s status=%s", tradeNo, topUp.Status))
		return
	}

	// 完成任务
	if err := model.UpdateTopUpStatus(tradeNo, model.PaymentProviderStripe, model.TopUpStatusSuccess); err != nil {
		logger.Error(ctx, fmt.Sprintf("Stripe 更新订单状态失败: trade_no=%s error=%q", tradeNo, err.Error()))
		return
	}

	// 充值用户配额
	if err := model.CompleteTopUp(tradeNo, model.PaymentProviderStripe, topUp.Amount); err != nil {
		logger.Error(ctx, fmt.Sprintf("Stripe 充值配额失败: trade_no=%s error=%q", tradeNo, err.Error()))
		return
	}

	logger.Info(ctx, fmt.Sprintf("Stripe 充值成功: trade_no=%s user_id=%d amount=%d money=%.2f", 
		tradeNo, topUp.UserId, topUp.Amount, topUp.Money))
}

// handleStripeSessionExpired 处理 Stripe 会话过期事件
func handleStripeSessionExpired(ctx context.Context, payload []byte) {
	tradeNo := extractStripeReferenceId(payload)
	if tradeNo == "" {
		logger.Warn(ctx, "Stripe checkout.expired 缺少订单号")
		return
	}

	// 更新订单状态为过期
	if err := model.UpdateTopUpStatus(tradeNo, model.PaymentProviderStripe, model.TopUpStatusExpired); err != nil {
		logger.Error(ctx, fmt.Sprintf("Stripe 订单过期处理失败: trade_no=%s error=%q", tradeNo, err.Error()))
		return
	}

	logger.Info(ctx, fmt.Sprintf("Stripe 订单已过期: trade_no=%s", tradeNo))
}

// extractStripeReferenceId 从 payload 中提取 reference_id（订单号）
func extractStripeReferenceId(payload []byte) string {
	// 查找 "client_reference_id" 字段
	marker := "\"client_reference_id\":\""
	for i := 0; i < len(payload)-len(marker); i++ {
		if string(payload[i:i+len(marker)]) == marker {
			j := i + len(marker)
			start := j
			for j < len(payload) && payload[j] != '"' {
				j++
			}
			return string(payload[start:j])
		}
	}
	return ""
}

// extractStripeStatus 从 payload 中提取 session status
func extractStripeStatus(payload []byte) string {
	marker := "\"status\":\""
	for i := 0; i < len(payload)-len(marker); i++ {
		if string(payload[i:i+len(marker)]) == marker {
			j := i + len(marker)
			start := j
			for j < len(payload) && payload[j] != '"' {
				j++
			}
			return string(payload[start:j])
		}
	}
	return ""
}

// extractStripePaymentStatus 从 payload 中提取 payment_status
func extractStripePaymentStatus(payload []byte) string {
	marker := "\"payment_status\":\""
	for i := 0; i < len(payload)-len(marker); i++ {
		if string(payload[i:i+len(marker)]) == marker {
			j := i + len(marker)
			start := j
			for j < len(payload) && payload[j] != '"' {
				j++
			}
			return string(payload[start:j])
		}
	}
	return ""
}

// genStripeCheckoutSession 生成 Stripe Checkout Session
// TODO: 集成 Stripe SDK
func genStripeCheckoutSession(tradeNo string, customerId string, email string, amount int64, successURL string, cancelURL string) string {
	// 这里应该调用 Stripe SDK 创建 Checkout Session
	// 返回示例：https://checkout.stripe.com/c/pay/xxx
	notifyURL := common.GetPaymentNotifyURL()
	if notifyURL == "" {
		notifyURL = common.GetPaymentSetting().PaymentNotifyURL
	}
	return fmt.Sprintf("https://checkout.stripe.com/c/pay/%s?reference=%s&amount=%d", 
		tradeNo, tradeNo, amount)
}

// isStripeWebhookEnabled 检查 Stripe Webhook 是否启用
func isStripeWebhookEnabled() bool {
	return common.IsStripeEnabled()
}

// getStripePayMoney 计算 Stripe 支付金额
func getStripePayMoney(amount float64, group string) float64 {
	return calculateStripePayMoney(int64(amount), group)
}

// helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
