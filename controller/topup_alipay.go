package controller

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/config"
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

	// 生成支付宝支付表单参数
	setting := common.GetPaymentSetting()
	if setting.AlipayAppId == "" || setting.AlipayPrivateKey == "" {
		logger.Error(c.Request.Context(), "支付宝支付未配置 AppId 或 PrivateKey")
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "支付宝支付未配置完成"})
		return
	}

	notifyURL := config.ServerAddress + "/api/webhook/alipay"
	returnURL := config.ServerAddress + "/payment/alipay/return"
	if req.ReturnURL != "" {
		returnURL = req.ReturnURL
	}

	isMobile := isMobileDevice(c.GetHeader("User-Agent"))
	subject := common.GetPaymentSetting().AlipaySubject
	if subject == "" {
		subject = "QuantumClaw 充值"
	}
	bizContent := buildAlipayBizContent(tradeNo, payMoney, subject, isMobile)
	payFormURL, err := buildAlipayPayForm(setting.AlipayAppId, setting.AlipayPrivateKey,
		setting.AlipayGatewayUrl, notifyURL, returnURL, bizContent, isMobile)
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("生成支付宝支付表单失败 trade_no=%s error=%q", tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "生成支付链接失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"trade_no":    tradeNo,
			"amount":      req.Amount,
			"money":       payMoney,
			"payment_url": payFormURL,
		},
	})
}

// isMobileDevice 通过 User-Agent 判断是否为移动设备
func isMobileDevice(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	mobileKeywords := []string{"mobile", "android", "iphone", "ipad", "ipod", "phone", "windows phone"}
	for _, kw := range mobileKeywords {
		if strings.Contains(ua, kw) {
			return true
		}
	}
	return false
}

// buildAlipayBizContent 构建支付宝 biz_content
// 聚合支付：自动适配 PC 网页支付和移动端 H5 支付
func buildAlipayBizContent(tradeNo string, totalAmount float64, subject string, isMobile bool) string {
	biz := map[string]interface{}{
		"out_trade_no":  tradeNo,
		"total_amount":  fmt.Sprintf("%.2f", totalAmount),
		"subject":       subject,
		"product_code":  "FAST_INSTANT_TRADE_PAY",
	}
	if isMobile {
		// 移动端：使用快捷支付（wap），支付宝收银台自适应
		biz["product_code"] = "QUICK_WAP_PAY"
		biz["quit_url"] = "" // 前端传递退出回跳 URL
	}
	b, _ := json.Marshal(biz)
	return string(b)
}

// buildAlipayPayForm 构建支付宝聚合支付表单（PC端 page.pay / 移动端 wap.pay）
// 返回值: 支付跳转 URL，前端可直接跳转或展示 iframe
func buildAlipayPayForm(appId, privateKey, gatewayUrl, notifyURL, returnURL, bizContent string, isMobile bool) (string, error) {
	params := url.Values{}
	method := "alipay.trade.page.pay"
	if isMobile {
		method = "alipay.trade.wap.pay"
	}
	params.Set("app_id", appId)
	params.Set("method", method)
	params.Set("format", "JSON")
	params.Set("charset", "utf-8")
	params.Set("sign_type", "RSA2")
	params.Set("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	params.Set("version", "1.0")
	params.Set("notify_url", notifyURL)
	params.Set("return_url", returnURL)
	params.Set("biz_content", bizContent)

	// 签名：对除 sign 本身外的所有参数排序后拼接 + RSA2 签名
	signStr := buildAlipaySignString(params)
	sign, err := rsa2Sign(signStr, privateKey)
	if err != nil {
		return "", fmt.Errorf("RSA2 签名失败: %w", err)
	}
	params.Set("sign", sign)

	// 支付宝新网关（聚合支付统一入口）
	if gatewayUrl == "" {
		gatewayUrl = "https://openapi.alipay.com/gateway.do"
	}
	return gatewayUrl + "?" + params.Encode(), nil
}

// buildAlipaySignString 构造待签名字符串：key=value&key=value...（按键名升序）
func buildAlipaySignString(params url.Values) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := params.Get(k)
		if v != "" {
			parts = append(parts, k+"="+v)
		}
	}
	return strings.Join(parts, "&")
}

// rsa2Sign 使用 RSA2 (SHA256) 签名
func rsa2Sign(signStr, privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		// 尝试不带 PEM 头的纯 base64 私钥
		pemStr := "-----BEGIN RSA PRIVATE KEY-----\n" + chunkString(privateKeyPEM, 64) + "\n-----END RSA PRIVATE KEY-----"
		block, _ = pem.Decode([]byte(pemStr))
		if block == nil {
			return "", fmt.Errorf("无法解码 RSA 私钥")
		}
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试 PKCS8 格式
		key, parseErr := x509.ParsePKCS8PrivateKey(block.Bytes)
		if parseErr != nil {
			return "", fmt.Errorf("解析 RSA 私钥失败: %w", err)
		}
		var ok bool
		privKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("私钥类型不是 RSA")
		}
	}

	hash := sha256.Sum256([]byte(signStr))
	signBytes, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signBytes), nil
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
