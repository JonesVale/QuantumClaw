package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
	"github.com/quantumclaw/quantumclaw/setting"
	"github.com/gin-gonic/gin"
)

// Waffo Pancake 支付方式常量
const PaymentMethodWaffoPancake = "waffo_pancake"

// ==================== 创建 Waffo Pancake 充值订单 ====================

func CreateWaffoPancakeTopUp(c *gin.Context) {
	if !setting.SystemSetting.WaffoPancakeEnabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Waffo Pancake 支付未启用"})
		return
	}
	userId := c.GetInt("id")

	// 解析充值金额
	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的充值金额"})
		return
	}
	amount := req.Amount

	// 生成订单号
	tradeNo, err := model.GenerateSecureTradeNo(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建订单失败"})
		return
	}

	topUp := &model.TopUp{
		UserId:        userId,
		Amount:        amount,
		TradeNo:       tradeNo,
		PaymentMethod: PaymentMethodWaffoPancake,
		Status:        model.TopUpStatusPending,
		CreateTime:    time.Now().Unix(),
	}
	if err := topUp.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建充值订单失败: " + err.Error()})
		return
	}

	successURL := c.Query("success_url")
	if successURL == "" {
		successURL = setting.SystemSetting.BaseURL + "/#/wallet"
	}

	expiresIn := 3600
	params := &service.WaffoPancakeCreateSessionParams{
		StoreID:     setting.SystemSetting.WaffoPancakeStoreID,
		ProductID:   setting.SystemSetting.WaffoPancakeProductID,
		ProductType: "credits",
		Currency:    setting.SystemSetting.WaffoPancakeCurrency,
		PriceSnapshot: &service.WaffoPancakePriceSnapshot{
			Amount:      formatAmountForWaffo(amount),
			TaxIncluded: false,
		},
		BuyerEmail:       getUserEmail(c),
		SuccessURL:       successURL,
		ExpiresInSeconds: &expiresIn,
	}

	session, err := service.CreateWaffoPancakeCheckoutSession(context.Background(), params)
	if err != nil {
		topUp.Status = model.TopUpStatusFailed
		topUp.Update()
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建支付会话失败: " + err.Error()})
		return
	}

	// 更新订单的 session 信息
	topUp.ProviderData = session.SessionID
	topUp.Update()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"trade_no":     tradeNo,
			"checkout_url": session.CheckoutURL,
			"session_id":   session.SessionID,
			"expires_at":   session.ExpiresAt,
		},
	})
}

// ==================== Waffo Pancake Webhook 回调 ====================

func HandleWaffoPancakeWebhook(c *gin.Context) {
	if !setting.SystemSetting.WaffoPancakeEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Waffo Pancake 支付未启用"})
		return
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求体"})
		return
	}

	signatureHeader := c.GetHeader("X-Waffo-Signature")
	event, err := service.VerifyConfiguredWaffoPancakeWebhook(string(payload), signatureHeader)
	if err != nil {
		logger.SysError("Waffo Pancake webhook 验签失败: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "验签失败: " + err.Error()})
		return
	}

	eventType := event.NormalizedEventType()
	if eventType == "checkout.completed" || eventType == "payment.completed" || eventType == "order.completed" {
		tradeNo, err := service.ResolveWaffoPancakeTradeNo(event)
		if err != nil {
			logger.SysError("Waffo Pancake webhook 解析订单失败: " + err.Error())
			c.JSON(http.StatusOK, gin.H{"received": true})
			return
		}
		// 获取订单信息
		topUp, err := model.GetTopUpByTradeNo(tradeNo)
		if err != nil {
			logger.SysError("Waffo Pancake webhook 查询订单失败: " + err.Error())
		} else {
			// 完成充值
			if err := model.CompleteTopUp(tradeNo, model.PaymentProviderWaffoPancake, topUp.Amount); err != nil {
				logger.SysError("Waffo Pancake 完成充值失败: " + err.Error())
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// ==================== Waffo Pancake 设置保存 ====================

type WaffoPancakeSettingsRequest struct {
	Enabled        *bool  `json:"enabled"`
	MerchantID     string `json:"merchant_id"`
	StoreID        string `json:"store_id"`
	ProductID      string `json:"product_id"`
	Currency       string `json:"currency"`
	Sandbox        *bool  `json:"sandbox"`
	WebhookKey     string `json:"webhook_key"`
	WebhookTestKey string `json:"webhook_test_key"`
}

func SaveWaffoPancakeSettings(c *gin.Context) {
	var req WaffoPancakeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if req.Enabled != nil {
		setting.SystemSetting.WaffoPancakeEnabled = *req.Enabled
	}
	setting.SystemSetting.WaffoPancakeMerchantID = strings.TrimSpace(req.MerchantID)
	setting.SystemSetting.WaffoPancakeStoreID = strings.TrimSpace(req.StoreID)
	setting.SystemSetting.WaffoPancakeProductID = strings.TrimSpace(req.ProductID)
	setting.SystemSetting.WaffoPancakeCurrency = strings.TrimSpace(req.Currency)
	if req.Sandbox != nil {
		setting.SystemSetting.WaffoPancakeSandbox = *req.Sandbox
	}
	setting.SystemSetting.WaffoPancakeWebhookPublicKey = strings.TrimSpace(req.WebhookKey)
	setting.SystemSetting.WaffoPancakeWebhookTestKey = strings.TrimSpace(req.WebhookTestKey)
	if err := setting.SaveSetting("SystemSetting", setting.SystemSetting); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存设置失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
}

// ==================== 工具函数 ====================

func formatAmountForWaffo(amount int64) string {
	// Waffo Pancake 金额以分为单位
	return strings.TrimRight(strings.TrimRight(
		strconv.FormatFloat(float64(amount)/100.0, 'f', 2, 64), "0"), ".")
}

func getUserEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if exists {
		if e, ok := email.(string); ok && e != "" {
			return e
		}
	}
	return ""
}

// avoid unused import
var _ = json.Marshal
