package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// AlipayPayRequest 支付宝支付请求
type AlipayPayRequest struct {
	Amount     int64  `json:"amount" binding:"required,min=1"`
	ReturnURL  string `json:"return_url,omitempty"`
}

// RequestAlipayTopUp 请求支付宝支付
// POST /api/user/topup/alipay
func RequestAlipayTopUp(c *gin.Context) {
	var req AlipayPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	minTopUp := common.GetAlipayMinTopUp()
	if req.Amount < int64(minTopUp) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("充值数量不能小于 %d", minTopUp)})
		return
	}

	userId := c.GetInt("id")

	// 计算支付金额
	payMoney := calculatePayMoney(req.Amount, c.GetString("group"))

	// 创建订单
	tradeNo, err := model.GenerateSecureTradeNo(userId)
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("生成订单号失败 user_id=%d error=%q", userId, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建订单失败"})
		return
	}

	topUp := &model.TopUp{
		UserId:          userId,
		Amount:          req.Amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   "alipay",
		PaymentProvider: "alipay",
		UserIP:          c.ClientIP(),
		UserAgent:       c.GetHeader("User-Agent"),
	}

	if err := topUp.Insert(); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("创建充值订单失败 user_id=%d trade_no=%s error=%q", userId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建订单失败"})
		return
	}

	// 构建支付宝支付表单（生成支付链接，让用户跳转）
	payURL := fmt.Sprintf("/payment/alipay/redirect?trade_no=%s", tradeNo)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"trade_no":    tradeNo,
			"amount":      req.Amount,
			"money":       payMoney,
			"payment_url": payURL,
			"qrcode":      "", // TODO: 对接支付宝 SDK 后生成支付二维码
		},
	})
}

// AlipayNotify 支付宝异步通知回调
// POST/GET /api/webhook/alipay
func AlipayNotify(c *gin.Context) {
	// 获取支付宝回调参数
	params := make(map[string]string)
	if c.Request.Method == "GET" {
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
	} else {
		c.Request.ParseForm()
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
	}

	if len(params) == 0 {
		c.String(http.StatusOK, "fail")
		return
	}

	tradeNo := params["out_trade_no"]
	if tradeNo == "" {
		logger.Warn(c.Request.Context(), "支付宝回调缺少订单号")
		c.String(http.StatusOK, "fail")
		return
	}

	tradeStatus := params["trade_status"]
	if tradeStatus != "TRADE_SUCCESS" {
		c.String(http.StatusOK, "success")
		return
	}

	// 验证签名
	alipayPublicKey := common.GetPaymentSetting().AlipayPublicKey
	if alipayPublicKey == "" {
		logger.Warn(c.Request.Context(), "支付宝公钥未配置，无法验签")
		c.String(http.StatusOK, "fail")
		return
	}

	// TODO: 对接支付宝 SDK 后，用 RSA2 验签

	// 更新订单状态
	if err := model.UpdateTopUpStatus(tradeNo, model.PaymentProviderEpay, model.TopUpStatusSuccess); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("支付宝更新订单状态失败 trade_no=%s error=%q", tradeNo, err.Error()))
		c.String(http.StatusOK, "fail")
		return
	}

	// 充值
	topUp, err := model.GetTopUpByTradeNo(tradeNo)
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("支付宝订单不存在 trade_no=%s", tradeNo))
		c.String(http.StatusOK, "fail")
		return
	}

	if err := model.CompleteTopUp(tradeNo, "alipay", topUp.Amount); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("支付宝充值失败 trade_no=%s error=%q", tradeNo, err.Error()))
		c.String(http.StatusOK, "fail")
		return
	}

	logger.Info(c.Request.Context(), fmt.Sprintf("支付宝充值成功 trade_no=%s user_id=%d amount=%d", tradeNo, topUp.UserId, topUp.Amount))
	c.String(http.StatusOK, "success")
}

// AlipayReturn 支付宝同步返回（用户支付完成后跳转）
func AlipayReturn(c *gin.Context) {
	tradeNo := c.Query("out_trade_no")
	if tradeNo == "" {
		c.Redirect(http.StatusFound, "/wallet")
		return
	}
	// 直接跳转到钱包页面
	c.Redirect(http.StatusFound, fmt.Sprintf("/wallet?trade_no=%s&status=success", tradeNo))
}
