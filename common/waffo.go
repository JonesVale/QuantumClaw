package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ==================== Waffo 支付集成 ====================

const (
	waffoProdAPI = "https://www.waffo.com"
	waffoSandboxAPI = "https://www.waffo.com" // Waffo sandbox 同域名
	// Waffo 创建订单路径
	waffoOrderCreatePath = "/api/v1/order/create"
)

// WaffoOrderCreateRequest Waffo 创建订单请求
type WaffoOrderCreateRequest struct {
	PaymentRequestID string `json:"paymentRequestId"`
	MerchantOrderID  string `json:"merchantOrderId"`
	OrderCurrency    string `json:"orderCurrency"`
	OrderAmount      string `json:"orderAmount"`
	UserCurrency     string `json:"userCurrency,omitempty"`
	UserAmount       string `json:"userAmount,omitempty"`
	PayMethod        string `json:"payMethod,omitempty"`
	ProductName      string `json:"productName"`
	ProductDetail    string `json:"productDetail,omitempty"`
	ReturnURL        string `json:"returnUrl,omitempty"`
	WebhookURL       string `json:"webhookUrl,omitempty"`
	// 以下为扩展字段
	MerchantUserID string `json:"merchantUserId,omitempty"`
	Note           string `json:"note,omitempty"`
}

// WaffoOrderCreateResponse Waffo 创建订单响应
type WaffoOrderCreateResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Data    *struct {
		PaymentID      string `json:"paymentId"`
		TradeNo        string `json:"tradeNo"`
		RedirectURL    string `json:"redirectUrl"`
		OrderStatus    string `json:"orderStatus"`
		ExpireTime     int64  `json:"expireTime"`
	} `json:"data,omitempty"`
}

// WaffoWebhookPayload Waffo webhook 回调结构
type WaffoWebhookPayload struct {
	PaymentID       string `json:"paymentId"`
	MerchantOrderID string `json:"merchantOrderId"`
	TradeNo         string `json:"tradeNo"`
	OrderStatus     string `json:"orderStatus"`
	OrderAmount     string `json:"orderAmount"`
	OrderCurrency   string `json:"orderCurrency"`
	Sign            string `json:"sign"`
	SignType        string `json:"signType"`
	Timestamp       int64  `json:"timestamp"`
}

// CreateWaffoOrder 创建 Waffo 支付订单
func CreateWaffoOrder(params *StripeCheckoutParams) (redirectURL string, paymentID string, err error) {
	ps := GetPaymentSetting()
	if ps.WaffoApiKey == "" {
		return "", "", fmt.Errorf("Waffo API Key 未配置")
	}

	apiBase := waffoProdAPI
	if ps.WaffoSandbox {
		apiBase = waffoSandboxAPI
	}

	notifyURL := GetPaymentNotifyURL()
	if notifyURL == "" {
		notifyURL = ps.PaymentNotifyURL
	}

	// Waffo 金额以字符串形式发送（按 API 文档）
	orderAmount := fmt.Sprintf("%.2f", params.PayMoney)
	orderCurrency := ps.WaffoCurrency
	if orderCurrency == "" {
		orderCurrency = "USD"
	}

	req := WaffoOrderCreateRequest{
		PaymentRequestID: params.TradeNo,
		MerchantOrderID:  params.TradeNo,
		OrderCurrency:    orderCurrency,
		OrderAmount:      orderAmount,
		ProductName:      params.ProductName,
		ProductDetail:    fmt.Sprintf("Quota recharge: %d units", params.Amount),
		ReturnURL:        params.SuccessURL,
		WebhookURL:       notifyURL,
		MerchantUserID:   params.UserEmail,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("waffo request marshal: %w", err)
	}

	httpReq, err := http.NewRequest("POST", apiBase+waffoOrderCreatePath, strings.NewReader(string(body)))
	if err != nil {
		return "", "", fmt.Errorf("waffo request create: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ps.WaffoApiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("waffo api request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("waffo api error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result WaffoOrderCreateResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("waffo response parse: %w", err)
	}

	if result.Code != "SUCCESS" && result.Code != "200" && result.Code != "0" {
		return "", "", fmt.Errorf("waffo order failed: %s (code=%s)", result.Message, result.Code)
	}

	if result.Data == nil || result.Data.RedirectURL == "" {
		return "", "", fmt.Errorf("waffo response missing redirect URL")
	}

	logger.SysLog(fmt.Sprintf("Waffo 订单创建成功: payment_id=%s trade_no=%s", result.Data.PaymentID, params.TradeNo))

	return result.Data.RedirectURL, result.Data.PaymentID, nil
}

// VerifyWaffoWebhook 验证 Waffo webhook 签名
// Waffo 签名规则：按照参数名排序后拼接 value + api_key → SHA256
func VerifyWaffoWebhook(payload WaffoWebhookPayload, apiKey string) bool {
	if apiKey == "" {
		return false
	}

	// Waffo 签名：paymentId + merchantOrderId + tradeNo + orderStatus + orderAmount + orderCurrency + apiKey
	signStr := payload.PaymentID + payload.MerchantOrderID + payload.TradeNo +
		payload.OrderStatus + payload.OrderAmount + payload.OrderCurrency + apiKey

	// Waffo webhook 签名：SHA256(paymentId + merchantOrderId + tradeNo + orderStatus + orderAmount + orderCurrency + apiKey)
	hash := sha256.Sum256([]byte(signStr))
	expected := hex.EncodeToString(hash[:])

	return hmac.Equal([]byte(payload.Sign), []byte(expected))
}
