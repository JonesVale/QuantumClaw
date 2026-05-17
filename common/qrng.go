package common

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ==================== 量子随机数(QRNG)服务 ====================
// 设计原则：
// 1. 种子一次，本地扩展 — 启动时从 QRNG 源取一次种子，之后用 crypto/rand 本地扩展
// 2. 运行时零外部依赖 — 不影响每次交易延迟
// 3. 源 URL 可配置 — 默认 ANU(国外)，可切换国内量子源
// 4. 优雅降级 — QRNG 不可用时自动回退 crypto/rand
// ====================

var (
	QRNGEnabled = false
	// QRNGSourceURL: 量子随机数源 API 地址
	// 国外(默认澳洲国立大学): https://qrng.anu.edu.au/API/jsonI.php
	// 国内自建服务:  请指向自建或内网量子随机数服务
	// 国内可用服务:  百度AI云 / 阿里云 等商业量子随机数服务
	QRNGSourceURL = "https://qrng.anu.edu.au/API/jsonI.php"

	// qrngSeed 缓存从量子源获取的种子
	qrngSeed     []byte
	qrngSeedOnce sync.Once
	qrngSeedMu   sync.RWMutex
	qrngLogOnce  sync.Once

	// qrngSeedRefreshInterval 种子刷新间隔(默认1小时)
	qrngSeedRefreshInterval = 1 * time.Hour
	qrngLastFetch           time.Time
)

// QRNGResponse represents QRNG API response
type QRNGResponse struct {
	Success bool   `json:"success"`
	Data    []int  `json:"data"`
}

// InitQRNGSeed 初始化量子随机数种子
// 启动时调用一次，异步获取种子，不阻塞启动
func InitQRNGSeed() {
	if !QRNGEnabled {
		return
	}
	go func() {
		seed, err := fetchQuantumSeed()
		if err != nil {
			logger.SysWarn("qrng seed init failed, will fall back to crypto/rand: " + err.Error())
			return
		}
		qrngSeedMu.Lock()
		qrngSeed = seed
		qrngLastFetch = time.Now()
		qrngSeedMu.Unlock()
		logger.SysLog("qrng seed initialized (" + fmt.Sprintf("%d", len(seed)) + " bytes)")
	}()
}

// fetchQuantumSeed 从量子源获取 64 字节种子
func fetchQuantumSeed() ([]byte, error) {
	url := fmt.Sprintf("%s?length=64&type=uint8", QRNGSourceURL)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("qrng api request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("qrng api read failed: %w", err)
	}

	var qrngResp QRNGResponse
	if err := json.Unmarshal(body, &qrngResp); err != nil || !qrngResp.Success {
		return nil, fmt.Errorf("qrng api response invalid")
	}

	buf := make([]byte, len(qrngResp.Data))
	for i, v := range qrngResp.Data {
		buf[i] = byte(v)
	}

	// 混合本地熵增加安全性
	localSeed := make([]byte, 32)
	rand.Read(localSeed)

	hasher := sha256.New()
	hasher.Write(buf)
	hasher.Write(localSeed)
	return hasher.Sum(nil), nil
}

// GetQuantumRandomBytes 返回量子增强的随机字节
// 使用缓存的量子种子 + crypto/rand 本地扩展
// 运行时零外部网络依赖
func GetQuantumRandomBytes(length int) ([]byte, error) {
	if !QRNGEnabled {
		// 未启用 QRNG，使用标准 crypto/rand
		buf := make([]byte, length)
		_, err := rand.Read(buf)
		return buf, err
	}

	// 检查是否需要刷新种子（每小时一次）
	qrngSeedMu.RLock()
	needRefresh := time.Since(qrngLastFetch) > qrngSeedRefreshInterval
	hasSeed := len(qrngSeed) > 0
	seed := make([]byte, len(qrngSeed))
	copy(seed, qrngSeed)
	qrngSeedMu.RUnlock()

	// 异步刷新种子
	if needRefresh && hasSeed {
		go func() {
			newSeed, err := fetchQuantumSeed()
			if err == nil {
				qrngSeedMu.Lock()
				qrngSeed = newSeed
				qrngLastFetch = time.Now()
				qrngSeedMu.Unlock()
			}
		}()
	}

	if !hasSeed {
		// 种子尚未就绪，使用 crypto/rand
		qrngLogOnce.Do(func() {
			logger.SysWarn("qrng seed not ready yet, using crypto/rand")
		})
		buf := make([]byte, length)
		_, err := rand.Read(buf)
		return buf, err
	}

	// 使用量子种子 + crypto/rand 本地扩展随机字节
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	// 量子种子 XOR 本地随机数 → 增强的随机输出
	for i := 0; i < length; i++ {
		buf[i] ^= seed[i%len(seed)]
	}

	return buf, nil
}

// GenerateQuantumHMACKey 生成量子增强的 HMAC 密钥
func GenerateQuantumHMACKey() ([]byte, error) {
	quantumSeed, err := GetQuantumRandomBytes(32)
	if err != nil {
		return nil, err
	}

	localSeed := make([]byte, 32)
	rand.Read(localSeed)

	mac := hmac.New(sha256.New, quantumSeed)
	mac.Write(localSeed)
	return mac.Sum(nil), nil
}

// EnhancePaymentSignature 用量子种子增强支付签名
func EnhancePaymentSignature(secret string, payload []byte) string {
	if !QRNGEnabled {
		return ""
	}

	quantumSeed, err := GetQuantumRandomBytes(16)
	if err != nil {
		return ""
	}

	mac := hmac.New(sha256.New, append([]byte(secret), quantumSeed...))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
