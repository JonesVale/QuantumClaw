package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

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
	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		c.String(http.StatusOK, "fail")
		return
	}

	if err := model.UpdateTopUpStatus(tradeNo, "worldfirst", model.TopUpStatusSuccess); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("万里汇更新订单状态失败 trade_no=%s error=%q", tradeNo, err.Error()))
		c.String(http.StatusOK, "fail")
		return
	}

	topUp, err := model.GetTopUpByTradeNo(tradeNo)
	if err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	if err := model.CompleteTopUp(tradeNo, "worldfirst", topUp.Amount); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("万里汇充值失败 trade_no=%s error=%q", tradeNo, err.Error()))
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}
