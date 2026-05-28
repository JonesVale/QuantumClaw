package controller

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"sort"
	"strings"

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

	// 验证 RSA2 签名
	alipayPublicKey := common.GetPaymentSetting().AlipayPublicKey
	if alipayPublicKey == "" {
		logger.Warn(c.Request.Context(), "支付宝公钥未配置，无法验签")
		c.String(http.StatusOK, "fail")
		return
	}

	if err := verifyAlipaySign(params, alipayPublicKey); err != nil {
		logger.Warn(c.Request.Context(), fmt.Sprintf("支付宝回调签名验证失败 trade_no=%s error=%v", tradeNo, err))
		c.String(http.StatusOK, "fail")
		return
	}

	// 充值（CompleteTopUp 内部处理状态更新 + 手续费 + 债务抵扣）
	topUp, err := model.GetTopUpByTradeNo(tradeNo)
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("支付宝订单不存在 trade_no=%s", tradeNo))
		c.String(http.StatusOK, "fail")
		return
	}

	if err := model.CompleteTopUp(tradeNo, model.PaymentProviderAlipay, topUp.Amount); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("支付宝充值失败 trade_no=%s error=%q", tradeNo, err.Error()))
		c.String(http.StatusOK, "fail")
		return
	}

	logger.Info(c.Request.Context(), fmt.Sprintf("支付宝充值成功 trade_no=%s user_id=%d amount=%d", tradeNo, topUp.UserId, topUp.Amount))
	c.String(http.StatusOK, "success")
}

// verifyAlipaySign 验证支付宝 RSA2 签名
// 支付宝签名规则：
// 1. 排除 sign 和 sign_type 参数
// 2. 剩余参数按键名升序排序
// 3. 拼接成 key=value&key=value...
// 4. base64 解码 sign
// 5. RSA-SHA256 公钥验证
func verifyAlipaySign(params map[string]string, alipayPublicKey string) error {
	sign := params["sign"]
	if sign == "" {
		return fmt.Errorf("支付宝回调缺少 sign 参数")
	}

	// 1. 排除 sign 和 sign_type，收集待验签参数
	var keys []string
	for k := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 拼接 key=value&key=value...
	var parts []string
	for _, k := range keys {
		if params[k] != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}
	signContent := strings.Join(parts, "&")

	// 3. base64 解码签名
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return fmt.Errorf("base64 解码签名失败: %w", err)
	}

	// 4. 解析 RSA 公钥
	block, _ := pem.Decode([]byte(alipayPublicKey))
	if block == nil {
		// 尝试不带 PEM 头部的纯 Base64 公钥
		pemStr := fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----",
			chunkString(alipayPublicKey, 64))
		block, _ = pem.Decode([]byte(pemStr))
		if block == nil {
			return fmt.Errorf("解析支付宝公钥失败: 无法解码 PEM")
		}
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析 RSA 公钥失败: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("公钥类型不是 RSA")
	}

	// 5. SHA256 哈希 + RSA 验证
	hash := sha256.Sum256([]byte(signContent))
	if err := rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hash[:], signBytes); err != nil {
		return fmt.Errorf("RSA2 签名验证失败: %w", err)
	}

	return nil
}

// chunkString 将字符串每 n 个字符插入换行（用于纯 base64 补 PEM 格式）
func chunkString(s string, n int) string {
	if n <= 0 {
		return s
	}
	var result strings.Builder
	for i := 0; i < len(s); i += n {
		end := i + n
		if end > len(s) {
			end = len(s)
		}
		if result.Len() > 0 {
			result.WriteByte('\n')
		}
		result.WriteString(s[i:end])
	}
	return result.String()
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
