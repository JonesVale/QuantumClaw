package common

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ==================== Binance Pay 支付集成 ====================

const (
	binancePayAPI           = "https://bpay.binanceapi.com"
	binanceOrderCreatePath  = "/binancepay/openapi/v2/order"
)

// BinanceOrderRequest Binance Pay 创建订单请求
type BinanceOrderRequest struct {
	MerchantTradeNo string  `json:"merchantTradeNo"`
	TotalFee        float64 `json:"totalFee"`
	Currency        string  `json:"currency"`
	ProductType     string  `json:"productType"`
	ProductName     string  `json:"productName"`
	NotifyURL       string  `json:"notifyUrl,omitempty"`
	ReturnURL       string  `json:"returnUrl,omitempty"`
	WebhookUrl      string  `json:"webhookUrl,omitempty"`
}

// BinanceOrderResponse Binance Pay 创建订单响应
type BinanceOrderResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Error   string `json:"error,omitempty"`
	ErrorMsg string `json:"errorMessage,omitempty"`
	Data    *struct {
		PrepayID    string `json:"prepayId"`
		TradeNo     string `json:"tradeNo"`
		Currency    string `json:"currency"`
		TotalFee    string `json:"totalFee"`
		MerchantID  string `json:"merchantId"`
		ExpireTime  int64  `json:"expireTime"`
		CheckoutURL string `json:"checkoutUrl"`
	} `json:"data,omitempty"`
}

// BinanceWebhookPayload Binance Pay webhook payload
type BinanceWebhookPayload struct {
	BizType   string `json:"bizType"`
	Data      string `json:"data"`
	SignType  string `json:"signType"`
	Sign      string `json:"sign"`
	Timestamp int64  `json:"timestamp"`
}

// CreateBinancePayOrder 创建 Binance Pay 订单（REST API v2）
func CreateBinancePayOrder(params *PaymentCheckoutParams) (checkoutURL string, prepayID string, err error) {
	ps := GetPaymentSetting()
	if ps.BinanceApiKey == "" || ps.BinanceSecretKey == "" {
		return "", "", fmt.Errorf("Binance Pay API Key 未配置")
	}

	notifyURL := GetPaymentNotifyURL()
	if notifyURL == "" {
		notifyURL = ps.PaymentNotifyURL
	}

	currency := ps.BinanceCurrency
	if currency == "" {
		currency = "USDT"
	}

	req := BinanceOrderRequest{
		MerchantTradeNo: params.TradeNo,
		TotalFee:        params.PayMoney,
		Currency:        currency,
		ProductType:     "Product",
		ProductName:     params.ProductName,
		NotifyURL:       notifyURL,
		WebhookUrl:      notifyURL,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("binance request marshal: %w", err)
	}

	// Binance Pay 签名: timestamp + nonce + body
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	payloadToSign := timestamp + "\n" + nonce + "\n" + string(body) + "\n"

	mac := hmac.New(sha512.New, []byte(ps.BinanceSecretKey))
	mac.Write([]byte(payloadToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	httpReq, err := http.NewRequest("POST", binancePayAPI+binanceOrderCreatePath, strings.NewReader(string(body)))
	if err != nil {
		return "", "", fmt.Errorf("binance request create: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("BinancePay-Timestamp", timestamp)
	httpReq.Header.Set("BinancePay-Nonce", nonce)
	httpReq.Header.Set("BinancePay-Certificate-SN", ps.BinanceApiKey)
	httpReq.Header.Set("BinancePay-Signature", signature)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("binance api request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result BinanceOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("binance response parse: %w", err)
	}

	if result.Status != "SUCCESS" || result.Data == nil {
		errMsg := result.ErrorMsg
		if errMsg == "" {
			errMsg = result.Error
		}
		return "", "", fmt.Errorf("binance order failed: %s (code=%s)", errMsg, result.Code)
	}

	logger.SysLog(fmt.Sprintf("Binance Pay order created: trade_no=%s prepay_id=%s", params.TradeNo, result.Data.PrepayID))
	return result.Data.CheckoutURL, result.Data.PrepayID, nil
}

// VerifyBinanceWebhook 验证 Binance Pay webhook 签名
func VerifyBinanceWebhook(payload []byte, signature string) (bool, error) {
	ps := GetPaymentSetting()
	if ps.BinanceSecretKey == "" {
		return false, fmt.Errorf("Binance Secret Key 未配置")
	}

	mac := hmac.New(sha512.New, []byte(ps.BinanceSecretKey))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected)), nil
}
