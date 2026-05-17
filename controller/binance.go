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
	"time"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"

	"github.com/gin-gonic/gin"
)

// ==================== Binance Pay 控制器（安全增强版）====================

const (
	BinanceWebhookMaxBodySize = 1 * 1024 * 1024 // 1MB 最大请求体
	BinanceSignatureHeader    = "Binance-Signature"
	BinanceNonceHeader       = "Binance-Nonce"
	BinanceTimestampHeader   = "Binance-Timestamp"
)

// BinancePayRequest Binance支付请求
type BinancePayRequest struct {
	Amount        int64  `json:"amount" binding:"required,min=1"`
	PaymentMethod string `json:"payment_method"`
	SuccessURL    string `json:"success_url,omitempty"`
	CancelURL     string `json:"cancel_url,omitempty"`
}

// BinanceWebhookRequest Binance webhook请求
type BinanceWebhookRequest struct {
	BizType   string          `json:"bizType"`
	Data      json.RawMessage `json:"data"`
	Signature string          `json:"signature"`
}

// BinanceWebhookData Binance回调数据
type BinanceWebhookData struct {
	MerchantTradeNo string `json:"merchantTradeNo"`
	TradeNo         string `json:"tradeNo"`
	Status          string `json:"status"`
	TotalFee        string `json:"totalFee"`
	Currency        string `json:"currency"`
	TransactTime    int64  `json:"transactTime"`
}

// @Summary 请求 Binance Pay 支付
// @Description 创建 Binance Pay 充值订单
// @Tags User
// @Accept json
// @Produce json
// @Param request body BinancePayRequest true "支付请求"
// @Success 200 {object} common.Response
// @Router /api/user/self/topup/binance [post]
func RequestBinanceTopUp(c *gin.Context) {
	var req BinancePayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "参数错误", "data": err.Error()})
		return
	}

	// 验证功能是否启用
	if !common.IsBinanceEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "Binance Pay 未启用"})
		return
	}

	// 验证支付方式
	if req.PaymentMethod != model.PaymentMethodBinance {
		c.JSON(http.StatusOK, gin.H{"message": "不支持的支付方式"})
		return
	}

	// 安全增强：验证金额范围
	minTopUp := common.GetBinanceMinTopUp()
	if minTopUp <= 0 {
		minTopUp = 1
	}
	if req.Amount < int64(minTopUp) {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("充值数量不能小于 %d", minTopUp)})
		return
	}

	// 安全增强：验证重定向URL
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

	// 获取用户信息
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "获取用户信息失败"})
		return
	}

	// 计算支付金额（服务器端计算）
	payMoney := calculateBinancePayMoney(req.Amount, user.Group)
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
		PaymentMethod:    model.PaymentMethodBinance,
		PaymentProvider:  model.PaymentProviderBinance,
		UserIP:           c.ClientIP(),
		UserAgent:        c.GetHeader("User-Agent"),
	}

	if err := topUp.Insert(); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("Binance 创建充值订单失败 user_id=%d trade_no=%s error=%q", userId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "创建订单失败"})
		return
	}

	// 生成支付链接（TODO: 集成 Binance Pay SDK）
	checkoutURL := genBinanceCheckoutURL(tradeNo, user.Email, payMoney, "")

	logger.Info(c.Request.Context(), fmt.Sprintf("Binance 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f",
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

// @Summary Binance Pay Webhook 回调
// @Description 处理 Binance Pay 回调通知
// @Tags Payment
// @Accept json
// @Produce json
// @Header Binance-Signature "Binance HMAC-SHA256 签名"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Router /api/webhook/binance [post]
func BinanceWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	// 安全检查：验证 webhook 是否启用
	if !common.IsBinanceEnabled() {
		logger.Warn(ctx, "Binance webhook 被拒绝: 功能未启用")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// 读取并限制请求体大小（防止 DoS 攻击）
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, BinanceWebhookMaxBodySize))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Binance webhook 读取请求体失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	// 获取签名头
	signature := c.GetHeader(BinanceSignatureHeader)
	if signature == "" {
		// 尝试从JSON body中提取签名
		var event BinanceWebhookRequest
		if err := json.Unmarshal(bodyBytes, &event); err != nil {
			logger.Warn(ctx, "Binance webhook 缺少签名且解析失败")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		signature = event.Signature
	}

	logger.Info(ctx, fmt.Sprintf("Binance webhook 收到: client_ip=%s payload_size=%d",
		c.ClientIP(), len(bodyBytes)))

	// 验证签名
	if !verifyBinanceSignature(bodyBytes, signature) {
		logger.Warn(ctx, fmt.Sprintf("Binance webhook 验签失败 client_ip=%s", c.ClientIP()))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 解析 webhook 事件
	var event BinanceWebhookRequest
	if err := json.Unmarshal(bodyBytes, &event); err != nil {
		logger.Error(ctx, fmt.Sprintf("Binance webhook JSON 解析失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Binance webhook 事件: biz_type=%s", event.BizType))

	// 处理支付完成事件
	if event.BizType == "PAY" || event.BizType == "PAYMENT" {
		handleBinancePaymentCompleted(ctx, event.Data)
	} else {
		logger.Info(ctx, fmt.Sprintf("Binance webhook 忽略事件类型: %s", event.BizType))
	}

	c.Status(http.StatusOK)
}

// verifyBinanceSignature 验证 Binance webhook 签名
// Binance 使用 HMAC-SHA256 对请求体进行签名
func verifyBinanceSignature(payload []byte, signature string) bool {
	settings := common.GetPaymentSetting()
	secret := settings.BinanceSecretKey
	if secret == "" {
		logger.Warn(context.Background(), "Binance secret key 未配置，验签已跳过")
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// handleBinancePaymentCompleted 处理 Binance 支付完成事件
func handleBinancePaymentCompleted(ctx context.Context, data json.RawMessage) {
	var payData BinanceWebhookData
	if err := json.Unmarshal(data, &payData); err != nil {
		logger.Error(ctx, fmt.Sprintf("Binance webhook 解析支付数据失败: %q", err.Error()))
		return
	}

	// 检查支付状态
	if payData.Status != "PAY_SUCCESS" && payData.Status != "SUCCESS" {
		logger.Info(ctx, fmt.Sprintf("Binance 支付未成功，跳过: trade_no=%s status=%s",
			payData.MerchantTradeNo, payData.Status))
		return
	}

	tradeNo := payData.MerchantTradeNo
	if tradeNo == "" {
		logger.Warn(ctx, "Binance webhook 缺少商户订单号")
		return
	}

	// 获取本地订单
	topUp, err := model.GetTopUpByTradeNo(tradeNo)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Binance 订单不存在: trade_no=%s", tradeNo))
		return
	}

	// 安全验证：检查支付提供商
	if topUp.PaymentProvider != model.PaymentProviderBinance {
		logger.Warn(ctx, fmt.Sprintf("Binance 支付提供商不匹配: trade_no=%s expected=%s actual=%s",
			tradeNo, model.PaymentProviderBinance, topUp.PaymentProvider))
		return
	}

	// 安全验证：检查订单状态（防止重复处理）
	if topUp.Status != model.TopUpStatusPending {
		logger.Info(ctx, fmt.Sprintf("Binance 订单状态非 pending，跳过: trade_no=%s status=%s",
			tradeNo, topUp.Status))
		return
	}

	// 完成任务
	if err := model.CompleteTopUp(tradeNo, model.PaymentProviderBinance, topUp.Amount); err != nil {
		logger.Error(ctx, fmt.Sprintf("Binance 充值完成失败: trade_no=%s error=%q", tradeNo, err.Error()))
		return
	}

	logger.Info(ctx, fmt.Sprintf("Binance 充值成功: trade_no=%s user_id=%d amount=%d money=%.2f",
		tradeNo, topUp.UserId, topUp.Amount, topUp.Money))
}

// calculateBinancePayMoney 计算 Binance 支付金额
func calculateBinancePayMoney(amount int64, group string) float64 {
	settings := common.GetPaymentSetting()
	unitPrice := settings.BinanceUnitPrice
	if unitPrice <= 0 {
		unitPrice = 0.002
	}
	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}
	return float64(amount) * unitPrice * topupGroupRatio
}

// genBinanceCheckoutURL 生成 Binance Pay 支付链接
// TODO: 集成 Binance Pay SDK
func genBinanceCheckoutURL(tradeNo string, email string, amount float64, currency string) string {
	settings := common.GetPaymentSetting()
	if currency == "" {
		currency = settings.BinanceCurrency
	}
	if currency == "" {
		currency = "USDT"
	}
	timestamp := time.Now().UnixMilli()
	return fmt.Sprintf("https://pay.binance.com/checkout?merchantTradeNo=%s&amount=%.2f&currency=%s&email=%s&timestamp=%d",
		tradeNo, amount, currency, email, timestamp)
}

// isBinanceWebhookEnabled 检查 Binance Webhook 是否启用
func isBinanceWebhookEnabled() bool {
	return common.IsBinanceEnabled()
}
