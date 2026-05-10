package controller

import (
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
		c.JSON(http.StatusOK, gin.H{"message": "参数错误", "data": err.Error()})
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

	// 生成支付链接（TODO: 集成 Waffo SDK）
	checkoutURL := genWaffoCheckoutURL(tradeNo, user.Email, payMoney, req.PayMethodType)

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

// @Summary Waffo Webhook 回调
// @Description 处理 Waffo 支付回调
// @Tags Payment
// @Accept json
// @Produce json
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Router /api/webhook/waffo [post]
func WaffoWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	// 安全检查：验证 webhook 是否启用
	if !common.IsWaffoEnabled() {
		logger.Warn(ctx, "Waffo webhook 被拒绝: 功能未启用")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// 安全增强：读取并限制请求体大小
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, WaffoWebhookMaxBodySize))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Waffo webhook 读取请求体失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	_ = bodyBytes // 已读取，TODO: 验证 Waffo 签名时使用

	// TODO: 验证 Waffo 签名
	signature := c.GetHeader("Waffo-Signature")
	if signature == "" {
		logger.Warn(ctx, "Waffo webhook 缺少签名")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Waffo webhook 收到: client_ip=%s", c.ClientIP()))

	// 解析 webhook 数据
	var webhookReq WaffoWebhookRequest
	// 注意：这里需要使用正确的 JSON 解析方法
	// 暂时跳过解析，直接返回成功
	_ = webhookReq

	// TODO: 处理 Waffo 事件
	// handleWaffoPaymentCompleted(ctx, &webhookReq)

	c.Status(http.StatusOK)
}

// calculateWaffoPayMoney 计算 Waffo 支付金额
func calculateWaffoPayMoney(amount int64, group string) float64 {
	// TODO: 集成 Waffo 的价格配置
	// 暂时使用通用计算方式
	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}
	return float64(amount) * 0.002 * topupGroupRatio
}

// genWaffoCheckoutURL 生成 Waffo 支付链接
// TODO: 集成 Waffo SDK
func genWaffoCheckoutURL(tradeNo string, email string, amount float64, payMethodType string) string {
	return fmt.Sprintf("https://waffo.com/checkout/%s?amount=%.2f", tradeNo, amount)
}

// isWaffoWebhookEnabled 检查 Waffo Webhook 是否启用
func isWaffoWebhookEnabled() bool {
	return common.IsWaffoEnabled()
}
