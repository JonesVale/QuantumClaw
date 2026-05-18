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

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"

	"github.com/gin-gonic/gin"
)

// ==================== Waffo 支付控制器（安全增强版）====================

const (
	WaffoWebhookMaxBodySize = 1 * 1024 * 1024 // 1MB 最大请求体
	WaffoMaxTopUpAmount     = 10000           // 最大充值数量
)

// zeroDecimalCurrencies 零小数位币种
var zeroDecimalCurrencies = map[string]bool{
	"IDR": true, "JPY": true, "KRW": true, "VND": true,
}

// WaffoPayRequest Waffo支付请求
type WaffoPayRequest struct {
	Amount        int64  `json:"amount" binding:"required,min=1"`
	PayMethodType string `json:"pay_method_type"` // 支付方式类型
}

// WaffoWebhookRequest Waffo webhook请求
type WaffoWebhookRequest struct {
	EventType string `json:"event_type"`
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Signature string `json:"signature"`
}

// @Summary 请求 Waffo 支付
// @Description 创建 Waffo 支付订单
// @Tags User
// @Accept json
// @Produce json
// @Param request body WaffoPayRequest true "支付请求"
// @Success 200 {object} common.Response
// @Router /api/user/topup/waffo [post]
func RequestWaffoTopUp(c *gin.Context) {
	var req WaffoPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "请求参数错误"})
		return
	}

	// 验证功能是否启用
	if !common.IsWaffoEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "Waffo 支付未启用"})
		return
	}

	// 安全增强：验证金额
	minTopUp := common.GetWaffoMinTopUp()
	if minTopUp <= 0 {
		minTopUp = 1
	}
	if req.Amount < int64(minTopUp) {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("充值数量不能小于 %d", minTopUp)})
		return
	}
	if req.Amount > WaffoMaxTopUpAmount {
		c.JSON(http.StatusOK, gin.H{"message": "金额超限"})
		return
	}

	// 获取用户信息
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "获取用户信息失败"})
		return
	}

	// 计算支付金额（服务器端计算）
	payMoney := calculateWaffoPayMoney(req.Amount, user.Group)
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
		UserId:          userId,
		Amount:          req.Amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:    model.PaymentMethodWaffo,
		PaymentProvider:  model.PaymentProviderWaffo,
		UserIP:           c.ClientIP(),
		UserAgent:        c.GetHeader("User-Agent"),
	}

	if err := topUp.Insert(); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("Waffo 创建充值订单失败 user_id=%d trade_no=%s error=%q", userId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "创建订单失败"})
		return
	}

	// 调用 Waffo API 创建订单
	checkoutURL, _, err := common.CreateWaffoOrder(&common.StripeCheckoutParams{
		TradeNo:     tradeNo,
		Amount:      req.Amount,
		PayMoney:    payMoney,
		UserEmail:   user.Email,
		SuccessURL:  common.GetPaymentSetting().PaymentReturnURL,
		ProductName: fmt.Sprintf("Quota x%d", req.Amount),
	})
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("Waffo 创建订单失败 user_id=%d trade_no=%s error=%q", userId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "创建支付链接失败: " + err.Error()})
		return
	}

	logger.Info(c.Request.Context(), fmt.Sprintf("Waffo 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f", 
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

// WaffoWebhook 处理 Waffo webhook 回调
func WaffoWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	if !common.IsWaffoEnabled() {
		logger.Warn(ctx, "Waffo webhook 被拒绝: 功能未启用")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, WaffoWebhookMaxBodySize))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Waffo webhook 读取请求体失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	signature := c.GetHeader("Waffo-Signature")
	if signature == "" {
		logger.Warn(ctx, "Waffo webhook 缺少签名")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	settings := common.GetPaymentSetting()
	apiKey := settings.WaffoApiKey
	if settings.WaffoSandbox {
		apiKey = settings.WaffoSandboxApiKey
	}

	if !verifyWaffoSignature(bodyBytes, signature, apiKey) {
		logger.Warn(ctx, "Waffo webhook 签名验证失败")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var webhookReq WaffoWebhookRequest
	if err := json.Unmarshal(bodyBytes, &webhookReq); err != nil {
		logger.Error(ctx, fmt.Sprintf("Waffo webhook 解析JSON失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	handleWaffoPaymentCompleted(ctx, &webhookReq)

	c.Status(http.StatusOK)
}

// verifyWaffoSignature 验证 Waffo webhook 签名
func verifyWaffoSignature(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// handleWaffoPaymentCompleted 处理 Waffo 支付完成事件
func handleWaffoPaymentCompleted(ctx context.Context, req *WaffoWebhookRequest) {
	if req.EventType != "payment.completed" {
		logger.Info(ctx, fmt.Sprintf("Waffo webhook 收到非支付完成事件: %s", req.EventType))
		return
	}

	topUp, err := model.GetTopUpByTradeNo(req.OrderID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Waffo webhook 查询订单失败: order_id=%s error=%q", req.OrderID, err.Error()))
		return
	}

	if topUp.Status != model.TopUpStatusPending {
		logger.Info(ctx, fmt.Sprintf("Waffo webhook 订单状态非待支付: order_id=%s status=%s", req.OrderID, topUp.Status))
		return
	}

	if err := model.CompleteTopUp(req.OrderID, model.PaymentProviderWaffo, topUp.Amount); err != nil {
		logger.Error(ctx, fmt.Sprintf("Waffo webhook 更新订单状态失败: order_id=%s error=%q", req.OrderID, err.Error()))
		return
	}

	logger.Info(ctx, fmt.Sprintf("Waffo 支付成功: order_id=%s amount=%d money=%.2f", req.OrderID, topUp.Amount, topUp.Money))
}

// calculateWaffoPayMoney 计算 Waffo 支付金额
func calculateWaffoPayMoney(amount int64, group string) float64 {
	settings := common.GetPaymentSetting()
	unitPrice := settings.WaffoUnitPrice
	if unitPrice <= 0 {
		unitPrice = 0.002
	}
	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}
	return float64(amount) * unitPrice * topupGroupRatio
}

// genWaffoCheckoutURL 生成 Waffo 支付链接
// 注意: 需要集成 Waffo SDK 后实现
//       当前返回空字符串表示 SDK 未集成
func genWaffoCheckoutURL(_ string, _ string, _ float64, _ string) string {
	return ""
}

// isWaffoWebhookEnabled 检查 Waffo Webhook 是否启用
func isWaffoWebhookEnabled() bool {
	return common.IsWaffoEnabled()
}
