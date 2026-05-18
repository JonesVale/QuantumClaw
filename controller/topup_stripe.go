package controller

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	// 生成支付链接（需要集成 Stripe SDK）
	checkoutURL := genStripeCheckoutSession(tradeNo, "", user.Email, req.Amount, req.SuccessURL, req.CancelURL)
	if checkoutURL == "" {
		logger.Warn(c.Request.Context(), fmt.Sprintf("Stripe SDK 未集成，订单已创建但无法生成支付链接 trade_no=%s", tradeNo))
		c.JSON(http.StatusOK, gin.H{
			"message": "Stripe SDK 未配置，订单已记录",
			"data": gin.H{
				"trade_no": tradeNo,
				"amount":   req.Amount,
				"money":    payMoney,
			},
		})
		return
	}

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

	// 安全增强：验证 Stripe 签名 (HMAC-SHA256)
	webhookSecret := common.GetPaymentSetting().StripeWebhookSecret
	if webhookSecret != "" {
		// Stripe 签名格式: t=timestamp,v1=signature
		parts := strings.Split(signature, ",")
		var timestamp string
		var sig string
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				switch kv[0] {
				case "t":
					timestamp = kv[1]
				case "v1":
					sig = kv[1]
				}
			}
		}
		if timestamp == "" || sig == "" {
			logger.Warn(ctx, fmt.Sprintf("Stripe webhook 签名格式无效 client_ip=%s", c.ClientIP()))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 构建待签名字符串: timestamp + "." + payload
		signedPayload := timestamp + "." + string(payload)
		mac := hmac.New(sha256.New, []byte(webhookSecret))
		mac.Write([]byte(signedPayload))
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(sig), []byte(expectedSignature)) {
			logger.Warn(ctx, fmt.Sprintf("Stripe webhook 验签失败 client_ip=%s", c.ClientIP()))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		logger.Info(ctx, "Stripe webhook 验签成功")
	} else {
		logger.Warn(ctx, "Stripe webhook secret 未配置，跳过签名验证")
	}

	// 解析 Stripe webhook 事件
	var event StripeWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error(ctx, fmt.Sprintf("Stripe webhook JSON 解析失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Stripe webhook 事件类型: %s", event.Type))

	// 处理不同的事件类型
	switch event.Type {
	case "checkout.session.completed":
		handleStripeCheckoutCompleted(ctx, event)
	case "checkout.session.expired":
		handleStripeSessionExpired(ctx, event)
	default:
		logger.Info(ctx, fmt.Sprintf("Stripe webhook 忽略事件: %s", event.Type))
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

// genStripeCheckoutSession 生成 Stripe Checkout Session
// 注意: 需要集成 github.com/stripe/stripe-go/v81 后实现
//       当前返回空字符串表示 SDK 未集成，请先安装 Stripe SDK
func genStripeCheckoutSession(_ string, _ string, _ string, _ int64, _ string, _ string) string {
	return ""
}

// isStripeWebhookEnabled 检查 Stripe Webhook 是否启用
func isStripeWebhookEnabled() bool {
	return common.IsStripeEnabled()
}

// getStripePayMoney 计算 Stripe 支付金额
func getStripePayMoney(amount float64, group string) float64 {
	return calculateStripePayMoney(int64(amount), group)
}


