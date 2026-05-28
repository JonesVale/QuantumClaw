package controller

import (
	"context"
	"encoding/json"
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

// StripeWebhookEvent Stripe webhook 事件结构（仅解析所需字段）
type StripeWebhookEvent struct {
	Type string `json:"type"`
	Data struct {
		Object struct {
			ID              string `json:"id"`
			Status          string `json:"status"`
			PaymentStatus   string `json:"payment_status"`
			ClientReferenceID string `json:"client_reference_id"`
			Metadata        map[string]string `json:"metadata"`
		} `json:"object"`
	} `json:"data"`
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
		c.JSON(http.StatusOK, gin.H{"message": "请求参数错误"})
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

	// 调用 Stripe API 创建 Checkout Session
	checkoutURL, sessionID, err := common.CreateStripeCheckoutSession(&common.PaymentCheckoutParams{
		TradeNo:     tradeNo,
		Amount:      req.Amount,
		PayMoney:    payMoney,
		UserEmail:   user.Email,
		SuccessURL:  req.SuccessURL,
		CancelURL:   req.CancelURL,
		NotifyURL:   common.GetPaymentNotifyURL(),
		ProductName: fmt.Sprintf("Quota x%d", req.Amount),
	})
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("Stripe 创建 Checkout Session 失败 user_id=%d trade_no=%s error=%q", userId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "创建支付链接失败: " + err.Error()})
		return
	}
	_ = sessionID

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

	// 使用 Stripe SDK 验证 webhook 签名
	webhookSecret := common.GetPaymentSetting().StripeWebhookSecret
	eventType, tradeNoFromEvent, verifyErr := common.VerifyStripeWebhook(payload, signature, webhookSecret)
	if verifyErr != nil {
		logger.Warn(ctx, fmt.Sprintf("Stripe webhook 验签失败: %q client_ip=%s", verifyErr.Error(), c.ClientIP()))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	logger.Info(ctx, "Stripe webhook 验签成功")
	_ = tradeNoFromEvent

	// 解析 webhook 事件数据
	var event StripeWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error(ctx, fmt.Sprintf("Stripe webhook JSON 解析失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Stripe webhook 事件类型: %s", eventType))

	// 处理不同的事件类型
	switch eventType {
	case "checkout.session.completed":
		handleStripeCheckoutCompleted(ctx, event)
	case "checkout.session.expired":
		handleStripeSessionExpired(ctx, event)
	default:
		logger.Info(ctx, fmt.Sprintf("Stripe webhook 忽略事件: %s", eventType))
	}

	c.Status(http.StatusOK)
}

// handleStripeCheckoutCompleted 处理 Stripe Checkout 完成事件
func handleStripeCheckoutCompleted(ctx context.Context, event StripeWebhookEvent) {
	obj := event.Data.Object
	tradeNo := obj.ClientReferenceID
	if tradeNo == "" {
		logger.Warn(ctx, "Stripe checkout.completed 缺少订单号 (client_reference_id)")
		return
	}

	// 验证 checkout session 状态
	if obj.Status != "complete" {
		logger.Info(ctx, fmt.Sprintf("Stripe checkout 状态非 complete，跳过: trade_no=%s status=%s", tradeNo, obj.Status))
		return
	}

	// 验证支付状态
	if obj.PaymentStatus != "paid" {
		logger.Info(ctx, fmt.Sprintf("Stripe 支付未完成，等待异步结果: trade_no=%s payment_status=%s", tradeNo, obj.PaymentStatus))
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

	// 充值用户配额（CompleteTopUp 内部处理状态更新 + 手续费 + 债务抵扣）
	if err := model.CompleteTopUp(tradeNo, model.PaymentProviderStripe, topUp.Amount); err != nil {
		logger.Error(ctx, fmt.Sprintf("Stripe 充值配额失败: trade_no=%s error=%q", tradeNo, err.Error()))
		return
	}

	logger.Info(ctx, fmt.Sprintf("Stripe 充值成功: trade_no=%s user_id=%d amount=%d money=%.2f", 
		tradeNo, topUp.UserId, topUp.Amount, topUp.Money))
}

// handleStripeSessionExpired 处理 Stripe 会话过期事件
func handleStripeSessionExpired(ctx context.Context, event StripeWebhookEvent) {
	tradeNo := event.Data.Object.ClientReferenceID
	if tradeNo == "" {
		logger.Warn(ctx, "Stripe checkout.expired 缺少订单号 (client_reference_id)")
		return
	}

	// 更新订单状态为过期
	if err := model.UpdateTopUpStatus(tradeNo, model.PaymentProviderStripe, model.TopUpStatusExpired); err != nil {
		logger.Error(ctx, fmt.Sprintf("Stripe 订单过期处理失败: trade_no=%s error=%q", tradeNo, err.Error()))
		return
	}

	logger.Info(ctx, fmt.Sprintf("Stripe 订单已过期: trade_no=%s", tradeNo))
}


