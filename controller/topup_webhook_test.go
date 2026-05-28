package controller

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormSqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"modernc.org/sqlite"
)

// ── Test Helpers ──────────────────────────────────────────────

var registerWebhookTestOnce sync.Once
var webhookUserCounter int64

func setupWebhookTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	registerWebhookTestOnce.Do(func() {
		sql.Register("sqlite_modernc", &sqlite.Driver{})
	})

	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("qc_webhook_%d.db", time.Now().UnixNano()))
	db, err := gorm.Open(
		gormSqlite.New(gormSqlite.Config{
			DriverName: "sqlite_modernc",
			DSN:        tempFile,
		}),
		&gorm.Config{SkipDefaultTransaction: false},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		os.Remove(tempFile)
	})

	err = db.AutoMigrate(
		&model.User{},
		&model.TopUp{},
		&model.BalanceLog{},
		&model.TransactionLog{},
		&model.PlatformConfig{},
		&model.Notification{},
	)
	require.NoError(t, err)
	return db
}

func createWebhookTestUser(t *testing.T, db *gorm.DB) int {
	t.Helper()
	webhookUserCounter++
	counter := webhookUserCounter
	u := model.User{
		Username:    fmt.Sprintf("wh_%d", counter),
		Password:    "pw",
		DisplayName: "Webhook Test",
		Email:       fmt.Sprintf("wh_%d@test.local", counter),
		AccessToken: fmt.Sprintf("at_wh_%d", counter),
		Phone:       fmt.Sprintf("ph_%d", counter),
		QQ:          fmt.Sprintf("qq_%d", counter),
		Role:        model.RoleCommonUser,
		Status:      model.UserStatusEnabled,
		Quota:       1000000,
		CashBalance: 0,
		AffCode:     fmt.Sprintf("ac_%d", counter),
	}
	result := db.Create(&u)
	require.NoError(t, result.Error)
	return u.Id
}

func createWebhookTestTopUp(t *testing.T, db *gorm.DB, userId int, money float64, provider string) (int64, string) {
	t.Helper()
	tradeNo, err := model.GenerateSecureTradeNo(userId)
	require.NoError(t, err)
	topUp := model.TopUp{
		UserId:          userId,
		Amount:          int64(money * 100),
		Money:           money,
		TradeNo:         tradeNo,
		PaymentMethod:   provider,
		PaymentProvider: provider,
		Status:          model.TopUpStatusPending,
		CreateTime:      helper.GetTimestamp(),
		ExpireTime:      helper.GetTimestamp() + 1800,
	}
	result := db.Create(&topUp)
	require.NoError(t, result.Error)
	return topUp.Id, tradeNo
}

func setModelDB(db *gorm.DB) func() {
	orig := model.DB
	model.DB = db
	return func() { model.DB = orig }
}

// ── Alipay RSA2 Sign/Verify Test ─────────────────────────────

func generateRSAKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	return priv, string(pem.EncodeToMemory(pemBlock))
}

func alipaySign(t *testing.T, priv *rsa.PrivateKey, params map[string]string) string {
	t.Helper()
	var keys []string
	for k := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	var parts []string
	for _, k := range keys {
		if params[k] != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}
	signContent := ""
	for i, p := range parts {
		if i > 0 {
			signContent += "&"
		}
		signContent += p
	}
	hash := sha256.Sum256([]byte(signContent))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(sig)
}

func TestAlipayWebhook_SignVerify(t *testing.T) {
	priv, pubPEM := generateRSAKeyPair(t)
	params := map[string]string{
		"out_trade_no": "QC1234567890abcdef1234abcd",
		"trade_status": "TRADE_SUCCESS",
		"total_amount": "100.00",
		"app_id":       "2021001000000001",
		"charset":      "utf-8",
		"sign_type":    "RSA2",
		"timestamp":    "2026-05-28 20:00:00",
	}
	sig := alipaySign(t, priv, params)
	params["sign"] = sig

	err := verifyAlipaySign(params, pubPEM)
	assert.NoError(t, err, "RSA2 signature should verify correctly")

	params["total_amount"] = "999.99"
	err = verifyAlipaySign(params, pubPEM)
	assert.Error(t, err, "tampered params should fail verification")

	_, wrongPub := generateRSAKeyPair(t)
	err = verifyAlipaySign(params, wrongPub)
	assert.Error(t, err, "wrong public key should fail")
}

func TestAlipayWebhook_FullFlow(t *testing.T) {
	db := setupWebhookTestDB(t)
	defer setModelDB(db)()

	gin.SetMode(gin.TestMode)

	userId := createWebhookTestUser(t, db)
	_, tradeNo := createWebhookTestTopUp(t, db, userId, 100.0, model.PaymentProviderAlipay)

	// Set a valid RSA key for the handler to use
	priv, pubPEM := generateRSAKeyPair(t)
	common.GetPaymentSetting().AlipayPublicKey = pubPEM

	// Build signed params
	params := map[string]string{
		"out_trade_no": tradeNo,
		"trade_status": "TRADE_SUCCESS",
		"total_amount": "100.00",
		"app_id":       "2021001000000001",
		"charset":      "utf-8",
		"sign_type":    "RSA2",
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
	}
	params["sign"] = alipaySign(t, priv, params)

	vals := url.Values{}
	for k, v := range params {
		vals.Set(k, v)
	}

	// First call — should succeed
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/webhook/alipay?"+vals.Encode(), nil)
	AlipayNotify(c)

	t.Logf("First call response: %q (code %d)", w.Body.String(), w.Code)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")

	// Verify user credited
	var user model.User
	err := db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1010000), user.Quota)
	assert.Equal(t, int64(9900), user.CashBalance)

	// Duplicate — should fail
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/api/webhook/alipay?"+vals.Encode(), nil)
	AlipayNotify(c2)

	t.Logf("Duplicate call response: %q (code %d)", w2.Body.String(), w2.Code)
	assert.Contains(t, w2.Body.String(), "fail")
}

func TestAlipayWebhook_MissingSignature(t *testing.T) {
	db := setupWebhookTestDB(t)
	defer setModelDB(db)()

	gin.SetMode(gin.TestMode)

	userId := createWebhookTestUser(t, db)
	_, tradeNo := createWebhookTestTopUp(t, db, userId, 100.0, model.PaymentProviderAlipay)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/webhook/alipay?out_trade_no="+tradeNo+"&trade_status=TRADE_SUCCESS", nil)
	AlipayNotify(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "fail", "missing signature should fail")
}

// ── WorldFirst Webhook Tests ─────────────────────────────────

func TestWorldFirstWebhook_HMACSignVerify(t *testing.T) {
	db := setupWebhookTestDB(t)
	defer setModelDB(db)()

	gin.SetMode(gin.TestMode)

	webhookKey := "test_webhook_secret_key_2026"
	common.GetPaymentSetting().WorldFirstWebhookKey = webhookKey

	t.Run("valid_signature", func(t *testing.T) {
		userId := createWebhookTestUser(t, db)
		_, tradeNo := createWebhookTestTopUp(t, db, userId, 100.0, model.PaymentProviderWorldFirst)

		mac := hmac.New(sha256.New, []byte(webhookKey))
		mac.Write([]byte(tradeNo))
		validSig := fmt.Sprintf("%x", mac.Sum(nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/webhook/worldfirst?trade_no="+tradeNo, nil)
		c.Request.Header.Set("X-Signature", validSig)
		WorldFirstWebhook(c)

		t.Logf("WF valid response: %q (code %d)", w.Body.String(), w.Code)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")

		var user model.User
		err := db.First(&user, userId).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1010000), user.Quota)
		assert.Equal(t, int64(9500), user.CashBalance)
	})

	t.Run("invalid_signature", func(t *testing.T) {
		userId := createWebhookTestUser(t, db)
		_, tradeNo := createWebhookTestTopUp(t, db, userId, 100.0, model.PaymentProviderWorldFirst)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/webhook/worldfirst?trade_no="+tradeNo, nil)
		c.Request.Header.Set("X-Signature", "invalid_sig")
		WorldFirstWebhook(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("missing_signature_header", func(t *testing.T) {
		userId := createWebhookTestUser(t, db)
		_, tradeNo := createWebhookTestTopUp(t, db, userId, 100.0, model.PaymentProviderWorldFirst)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/webhook/worldfirst?trade_no="+tradeNo, nil)
		WorldFirstWebhook(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("missing_webhook_key_config", func(t *testing.T) {
		// 在子测试内保存/恢复全局配置
		origKey := common.GetPaymentSetting().WorldFirstWebhookKey
		defer func() { common.GetPaymentSetting().WorldFirstWebhookKey = origKey }()
		common.GetPaymentSetting().WorldFirstWebhookKey = ""

		userId := createWebhookTestUser(t, db)
		_, tradeNo := createWebhookTestTopUp(t, db, userId, 100.0, model.PaymentProviderWorldFirst)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/webhook/worldfirst?trade_no="+tradeNo, nil)
		c.Request.Header.Set("X-Signature", "any_sig")
		WorldFirstWebhook(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// ── Request Validation ───────────────────────────────────────

func TestAlipayPayRequest_JSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantAmt int64
		wantErr bool
	}{
		{"valid", `{"amount":100}`, 100, false},
		{"empty", `{}`, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req AlipayPayRequest
			err := json.Unmarshal([]byte(tt.input), &req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAmt, req.Amount)
			}
		})
	}
}

func TestWorldFirstTopUpRequest_JSON(t *testing.T) {
	var req WorldFirstTopUpRequest
	err := json.Unmarshal([]byte(`{"amount":100}`), &req)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), req.Amount)
}

func TestEpayWebhook_EpayType(t *testing.T) {
	assert.Equal(t, "alipay", common.EpayTypeAlipay)
	assert.Equal(t, "wxpay", common.EpayTypeWxpay)
	assert.Equal(t, "qqpay", common.EpayTypeQQPay)
}

// ── Concurrency Safety ───────────────────────────────────────

func TestAlipayWebhook_ConcurrentCallbacks(t *testing.T) {
	db := setupWebhookTestDB(t)
	defer setModelDB(db)()

	gin.SetMode(gin.TestMode)

	priv, pubPEM := generateRSAKeyPair(t)
	common.GetPaymentSetting().AlipayPublicKey = pubPEM

	concurrency := 5
	var wg sync.WaitGroup
	errs := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			userId := createWebhookTestUser(t, db)
			_, tradeNo := createWebhookTestTopUp(t, db, userId, 50.0, model.PaymentProviderAlipay)

			params := map[string]string{
				"out_trade_no": tradeNo,
				"trade_status": "TRADE_SUCCESS",
				"total_amount": "50.00",
				"app_id":       "2021001000000001",
				"charset":      "utf-8",
				"sign_type":    "RSA2",
				"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
			}
			params["sign"] = alipaySign(t, priv, params)

			vals := url.Values{}
			for k, v := range params {
				vals.Set(k, v)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/webhook/alipay?"+vals.Encode(), nil)
			AlipayNotify(c)

			if w.Code != http.StatusOK || w.Body.String() != "success" {
				errs <- fmt.Errorf("req %d: got code=%d body=%q", idx, w.Code, w.Body.String())
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
