package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// WorldFirstTopUpRequest 万里汇支付请求
type WorldFirstTopUpRequest struct {
	Amount    int64  `json:"amount" binding:"required,min=1"`
}

// RequestWorldFirstTopUp 请求万里汇支付
// POST /api/user/topup/worldfirst
func RequestWorldFirstTopUp(c *gin.Context) {
	var req WorldFirstTopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	minTopUp := 1
	if req.Amount < int64(minTopUp) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("充值数量不能小于 %d", minTopUp)})
		return
	}

	userId := c.GetInt("id")
	payMoney := calculatePayMoney(req.Amount, c.GetString("group"))

	tradeNo, err := model.GenerateSecureTradeNo(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建订单失败"})
		return
	}

	topUp := &model.TopUp{
		UserId:          userId,
		Amount:          req.Amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   "worldfirst",
		PaymentProvider: "worldfirst",
		UserIP:          c.ClientIP(),
		UserAgent:       c.GetHeader("User-Agent"),
	}

	if err := topUp.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建订单失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"trade_no":    tradeNo,
			"amount":      req.Amount,
			"money":       payMoney,
			"payment_url": "", // TODO: 对接万里汇 API 后生成支付链接
		},
	})
}

// WorldFirstWebhook 万里汇异步通知回调
// POST/GET /api/webhook/worldfirst
func WorldFirstWebhook(c *gin.Context) {
	// 1. 取回调参数（支持 GET query 和 POST body）
	var tradeNo string
	if c.Request.Method == "GET" {
		tradeNo = c.Query("trade_no")
	} else {
		c.Request.ParseForm()
		tradeNo = c.PostForm("trade_no")
		if tradeNo == "" {
			tradeNo = c.Query("trade_no")
		}
	}
	if tradeNo == "" {
		c.String(http.StatusOK, "fail")
		return
	}

	// 2. 签名验证 — 防止伪造回调
	webhookKey := common.GetPaymentSetting().WorldFirstWebhookKey
	if webhookKey == "" {
		logger.Error(c.Request.Context(), "万里汇 webhook key 未配置，拒绝回调")
		c.String(http.StatusForbidden, "fail")
		return
	}

	signature := c.GetHeader("X-Signature")
	if signature == "" {
		signature = c.Query("sign")
	}
	if signature == "" {
		logger.Warn(c.Request.Context(), "万里汇回调缺少签名参数")
		c.String(http.StatusForbidden, "fail")
		return
	}

	// 计算期望签名: HMAC-SHA256(trade_no, webhook_key)
	mac := hmac.New(sha256.New, []byte(webhookKey))
	mac.Write([]byte(tradeNo))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		logger.Warn(c.Request.Context(), fmt.Sprintf("万里汇回调签名验证失败 trade_no=%s", tradeNo))
		c.String(http.StatusForbidden, "fail")
		return
	}

	topUp, err := model.GetTopUpByTradeNo(tradeNo)
	if err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	// 3. 完成充值（CompleteTopUp 内部处理状态更新 + 手续费 + 债务抵扣）
	if err := model.CompleteTopUp(tradeNo, model.PaymentProviderWorldFirst, topUp.Amount); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("万里汇充值失败 trade_no=%s error=%q", tradeNo, err.Error()))
		c.String(http.StatusOK, "fail")
		return
	}

	logger.Info(c.Request.Context(), fmt.Sprintf("万里汇充值成功 trade_no=%s user_id=%d amount=%d", tradeNo, topUp.UserId, topUp.Amount))
	c.String(http.StatusOK, "success")
}
