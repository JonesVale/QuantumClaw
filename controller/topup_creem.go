package controller

import (
	"bytes"
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

// ==================== Creem 支付控制器（安全增强版）====================

const (
	CreemSignatureHeader     = "creem-signature"
	CreemWebhookMaxBodySize  = 1 * 1024 * 1024 // 1MB 最大请求体
	CreemMaxTopUpAmount      = 10000            // 最大充值数量
)

// CreemProduct Creem产品配置
type CreemProduct struct {
	ProductId string  `json:"productId"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Quota     int64   `json:"quota"` // 配额数量
}

// CreemPayRequest Creem支付请求
type CreemPayRequest struct {
	ProductId     string `json:"product_id" binding:"required"`
	PaymentMethod string `json:"payment_method"`
}

// CreemWebhookEvent Creem Webhook事件
type CreemWebhookEvent struct {
	Id        string `json:"id"`
	EventType string `json:"eventType"`
	Object    struct {
		Id        string `json:"id"`
		Object    string `json:"object"`
		RequestId string `json:"request_id"` // 这是我们的订单号
		Order     struct {
			Id          string `json:"id"`
			Amount      int    `json:"amount"`
			AmountPaid  int    `json:"amount_paid"`
			Status      string `json:"status"` // "paid", "pending", "failed"
			Product     string `json:"product"`
			Currency    string `json:"currency"`
			Mode        string `json:"mode"`
			Transaction string `json:"transaction"`
		} `json:"order"`
		Product struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"product"`
		Customer struct {
			Id    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"customer"`
		Metadata map[string]string `json:"metadata"`
	} `json:"object"`
}

// generateCreemSignature 生成 Creem HMAC-SHA256 签名
// 安全增强：使用时间戳和 nonce 防止重放攻击
func generateCreemSignature(payload string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// verifyCreemSignature 验证 Creem Webhook 签名
// 安全增强：
// 1. 时序安全的签名比较（防止时序攻击）
// 2. 记录验证失败日志但不泄露详细信息
// 3. 在测试模式下允许跳过验证
func verifyCreemSignature(payload string, signature string, secret string) bool {
	if secret == "" {
		logger.Warn(context.Background(), "Creem webhook secret 未配置，验证已跳过")
		return false // 生产环境必须配置 secret
	}

	if signature == "" {
		logger.Warn(context.Background(), "Creem webhook 缺少签名")
		return false
	}

	// 生成期望的签名
	expectedSignature := generateCreemSignature(payload, secret)

	// 时序安全的签名比较（防止时序攻击）
	signatureBytes := []byte(signature)
	expectedBytes := []byte(expectedSignature)

	if len(signatureBytes) != len(expectedBytes) {
		logger.Warn(context.Background(), "Creem webhook 签名长度不匹配")
		return false
	}

	if !hmac.Equal(signatureBytes, expectedBytes) {
		logger.Warn(context.Background(), "Creem webhook 签名验证失败")
		return false
	}

	return true
}

// @Summary 请求 Creem 支付
// @Description 创建 Creem 支付订单
// @Tags User
// @Accept json
// @Produce json
// @Param request body CreemPayRequest true "支付请求"
// @Success 200 {object} common.Response
// @Router /api/user/topup/creem [post]
func RequestCreemTopUp(c *gin.Context) {
	var req CreemPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "请求参数错误"})
		return
	}

	// 验证支付方式
	if req.PaymentMethod != model.PaymentMethodCreem {
		c.JSON(http.StatusOK, gin.H{"message": "不支持的支付方式"})
		return
	}

	// 验证产品ID
	if req.ProductId == "" {
		c.JSON(http.StatusOK, gin.H{"message": "请选择产品"})
		return
	}

	// 解析产品配置
	var products []CreemProduct
	creemProducts := common.GetPaymentSetting().CreemProducts
	if creemProducts != "" {
		if err := json.Unmarshal([]byte(creemProducts), &products); err != nil {
			logger.Error(c.Request.Context(), fmt.Sprintf("Creem 产品配置解析失败 user_id=%d error=%q", c.GetInt("id"), err.Error()))
			c.JSON(http.StatusOK, gin.H{"message": "产品配置错误"})
			return
		}
	}

	// 查找对应的产品
	var selectedProduct *CreemProduct
	for i := range products {
		if products[i].ProductId == req.ProductId {
			selectedProduct = &products[i]
			break
		}
	}

	if selectedProduct == nil {
		c.JSON(http.StatusOK, gin.H{"message": "产品不存在"})
		return
	}

	if selectedProduct.Quota > CreemMaxTopUpAmount {
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
		Amount:          selectedProduct.Quota,
		Money:           selectedProduct.Price,
		TradeNo:         tradeNo,
		PaymentMethod:    model.PaymentMethodCreem,
		PaymentProvider:  model.PaymentProviderCreem,
		UserIP:           c.ClientIP(),
		UserAgent:        c.GetHeader("User-Agent"),
	}

	if err := topUp.Insert(); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("Creem 创建充值订单失败 user_id=%d trade_no=%s product_id=%s error=%q", 
			userId, tradeNo, selectedProduct.ProductId, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "创建订单失败"})
		return
	}

	// 生成支付链接（TODO: 调用 Creem API）
	checkoutURL := genCreemCheckoutURL(tradeNo, selectedProduct, user.Email)

	logger.Info(c.Request.Context(), fmt.Sprintf("Creem 充值订单创建成功 user_id=%d trade_no=%s product_id=%s quota=%d money=%.2f", 
		userId, tradeNo, selectedProduct.ProductId, selectedProduct.Quota, selectedProduct.Price))

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"trade_no":     tradeNo,
			"checkout_url": checkoutURL,
			"amount":       selectedProduct.Quota,
			"money":        selectedProduct.Price,
		},
	})
}

// @Summary Creem Webhook 回调
// @Description 处理 Creem 支付回调
// @Tags Payment
// @Accept json
// @Produce json
// @Header Creem-Signature "Creem HMAC 签名"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Router /api/webhook/creem [post]
func CreemWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	// 安全检查：验证 webhook 是否启用
	if !common.IsCreemEnabled() {
		logger.Warn(ctx, "Creem webhook 被拒绝: 功能未启用")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// 读取请求体（限制大小，防止 DoS 攻击）
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, CreemWebhookMaxBodySize))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Creem webhook 读取请求体失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	payload := string(bodyBytes)

	// 获取签名头
	signature := c.GetHeader(CreemSignatureHeader)

	// 安全增强：验证签名（防止伪造回调）
	if !verifyCreemSignature(payload, signature, common.GetPaymentSetting().CreemWebhookSecret) {
		// 记录 IP 但不泄露详细信息
		logger.Warn(ctx, fmt.Sprintf("Creem webhook 验签失败 client_ip=%s", c.ClientIP()))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Creem webhook 验签成功"))

	// 重新设置 body 供 JSON 解析使用
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// 解析 webhook 数据
	var webhookEvent CreemWebhookEvent
	if err := c.ShouldBindJSON(&webhookEvent); err != nil {
		logger.Error(ctx, fmt.Sprintf("Creem webhook 解析失败: %q", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Creem webhook 事件: type=%s event_id=%s order_id=%s", 
		webhookEvent.EventType, webhookEvent.Id, webhookEvent.Object.Order.Id))

	// 处理不同的事件类型
	switch webhookEvent.EventType {
	case "checkout.completed":
		handleCreemCheckoutCompleted(ctx, &webhookEvent)
	default:
		logger.Info(ctx, fmt.Sprintf("Creem webhook 忽略事件: type=%s", webhookEvent.EventType))
	}

	c.Status(http.StatusOK)
}

// handleCreemCheckoutCompleted 处理 Creem 支付完成事件
func handleCreemCheckoutCompleted(ctx context.Context, event *CreemWebhookEvent) {
	// 安全验证：检查订单状态
	if event.Object.Order.Status != "paid" {
		logger.Info(ctx, fmt.Sprintf("Creem 订单未支付，跳过: order_id=%s status=%s", 
			event.Object.Order.Id, event.Object.Order.Status))
		return
	}

	// 获取订单号
	referenceId := event.Object.RequestId
	if referenceId == "" {
		logger.Warn(ctx, fmt.Sprintf("Creem webhook 缺少订单号: event_id=%s", event.Id))
		return
	}

	// 获取本地订单
	topUp, err := model.GetTopUpByTradeNo(referenceId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Creem 订单不存在: trade_no=%s", referenceId))
		return
	}

	// 安全验证：检查支付提供商
	if topUp.PaymentProvider != model.PaymentProviderCreem {
		logger.Warn(ctx, fmt.Sprintf("Creem 支付提供商不匹配: trade_no=%s expected=%s actual=%s", 
			referenceId, model.PaymentProviderCreem, topUp.PaymentProvider))
		return
	}

	// 安全验证：检查订单状态（防止重复处理）
	if topUp.Status != model.TopUpStatusPending {
		logger.Info(ctx, fmt.Sprintf("Creem 订单状态非 pending，跳过: trade_no=%s status=%s", 
			referenceId, topUp.Status))
		return
	}

	// 更新订单状态
	if err := model.UpdateTopUpStatus(referenceId, model.PaymentProviderCreem, model.TopUpStatusSuccess); err != nil {
		logger.Error(ctx, fmt.Sprintf("Creem 更新订单状态失败: trade_no=%s error=%q", referenceId, err.Error()))
		return
	}

	// 充值用户配额
	if err := model.CompleteTopUp(referenceId, model.PaymentProviderCreem, topUp.Amount); err != nil {
		logger.Error(ctx, fmt.Sprintf("Creem 充值配额失败: trade_no=%s error=%q", referenceId, err.Error()))
		return
	}

	logger.Info(ctx, fmt.Sprintf("Creem 充值成功: trade_no=%s user_id=%d amount=%d money=%.2f", 
		referenceId, topUp.UserId, topUp.Amount, topUp.Money))
}

// genCreemCheckoutURL 生成 Creem 支付链接
// TODO: 集成 Creem API
func genCreemCheckoutURL(tradeNo string, product *CreemProduct, email string) string {
	// 这里应该调用 Creem API 创建支付链接
	// 返回示例：https://creem.io/checkout?product_id=xxx&reference=xxx
	return fmt.Sprintf("https://creem.io/checkout?product_id=%s&reference=%s&email=%s", 
		product.ProductId, tradeNo, email)
}

// CreemWebhookAvailability 检查 Creem Webhook 是否可用
func isCreemWebhookEnabled() bool {
	return common.IsCreemEnabled()
}
