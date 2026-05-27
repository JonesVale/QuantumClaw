package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/cascade"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

//  Slave Cascade Client 
// Runs on NODE_TYPE=slave instances.
// Maintains connection to the master node, syncs tokens, and batch-pushes billing records.

const (
	cascadeKeyFile   = ".cascade_key"
	heartbeatInterval   = 30 * time.Second
	tokenSyncInterval   = 60 * time.Second
	billingFlushInterval = 10 * time.Second
	configSyncInterval   = 300 * time.Second

	billingBufferMax = 100  // flush when this many records accumulated

	redisTokenPrefix  = "cascade:token:"   // key: cascade:token:<key_hash>
	redisTTLSeconds   = 3600               // 1 hour cache TTL
)

// CascadeConfig is loaded from env for the slave node.
type CascadeConfig struct {
	MasterURL string
	NodeName  string
	Region    string
}

var (
	Cascade *CascadeClient
)

// CascadeClient manages the slave node's connection to the master.
type CascadeClient struct {
	config    CascadeConfig
	nodeID    int
	apiKey    string
	httpClient *http.Client
	startTime  time.Time
	channelCnt int

	// billing buffer
	mu          sync.Mutex
	billingBuf  []cascade.BillingRecord

	// config cache (last known version)
	configVersion int64
	modelRatios   map[string]float64
}

// StartCascadeClient initializes and starts the cascade client on a slave node.
func StartCascadeClient(cfg CascadeConfig) {
	if config.IsMasterNode {
		logger.SysWarn("StartCascadeClient called on master node, skipping")
		return
	}

	Cascade = &CascadeClient{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		startTime:  time.Now(),
	}

	// Step 1: Register or load credentials
	if err := Cascade.initCredentials(); err != nil {
		logger.FatalLog("cascade init failed: " + err.Error())
	}

	logger.SysLog(fmt.Sprintf("cascade client started: node_id=%d master=%s", Cascade.nodeID, cfg.MasterURL))

	// Step 2: Initial full sync
	Cascade.syncTokens(0)
	Cascade.syncConfig()

	// Step 3: Start background goroutines
	go Cascade.heartbeatLoop()
	go Cascade.tokenSyncLoop()
	go Cascade.billingFlushLoop()
	go Cascade.configSyncLoop()

	logger.SysLog("cascade client background loops started")
}

// ==================== Credential Management ====================

func (cc *CascadeClient) initCredentials() error {
	keyPath := filepath.Join(filepath.Dir(os.Args[0]), cascadeKeyFile)
	data, err := os.ReadFile(keyPath)
	if err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(data)), ":", 2)
		if len(parts) == 2 {
			cc.nodeID, _ = strconv.Atoi(parts[0])
			cc.apiKey = parts[1]
			logger.SysLog("cascade: loaded stored credentials, node_id=" + parts[0])
			return nil
		}
	}

	// No stored credentials ?register
	return cc.register()
}

func (cc *CascadeClient) register() error {
	req := cascade.RegisterRequest{
		Name:    cc.config.NodeName,
		Region:  cc.config.Region,
		Version: common.Version,
	}

	var resp cascade.RegisterResponse
	if err := cc.callMaster("POST", "/api/cascade/node/register", req, &resp); err != nil {
		return fmt.Errorf("cascade register: %w", err)
	}

	cc.nodeID = resp.NodeID
	cc.apiKey = resp.APIKey

	// Persist credentials
	keyPath := filepath.Join(filepath.Dir(os.Args[0]), cascadeKeyFile)
	content := fmt.Sprintf("%d:%s", resp.NodeID, resp.APIKey)
	if err := os.WriteFile(keyPath, []byte(content), 0600); err != nil {
		logger.SysWarn("cascade: failed to persist credentials: " + err.Error())
	}

	logger.SysLog(fmt.Sprintf("cascade: registered as node %d (%s)", resp.NodeID, cc.config.NodeName))
	return nil
}

// ==================== HTTP Helper ====================

func (cc *CascadeClient) callMaster(method, path string, reqBody, respDest interface{}) error {
	url := strings.TrimRight(cc.config.MasterURL, "/") + path

	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	httpReq, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if cc.apiKey != "" {
		httpReq.Header.Set("X-Cascade-Key", cc.apiKey)
	}

	resp, err := cc.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var wrapper struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Message string          `json:"message"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return fmt.Errorf("cascade response parse: %w", err)
	}
	if !wrapper.Success {
		return fmt.Errorf("cascade error: %s", wrapper.Message)
	}

	if respDest != nil && wrapper.Data != nil {
		return json.Unmarshal(wrapper.Data, respDest)
	}
	return nil
}

// ==================== Heartbeat ====================

func (cc *CascadeClient) heartbeatLoop() {
	for {
		time.Sleep(heartbeatInterval)
		cc.sendHeartbeat()
	}
}

func (cc *CascadeClient) sendHeartbeat() {
	req := cascade.HeartbeatRequest{
		NodeID:       cc.nodeID,
		ChannelCount: cc.channelCnt,
		TodayCalls:   0,     // TODO: track per-node call stats
		UptimeSec:    int64(time.Since(cc.startTime).Seconds()),
	}

	var resp cascade.HeartbeatResponse
	if err := cc.callMaster("POST", "/api/cascade/node/heartbeat", req, &resp); err != nil {
		logger.Warn(context.Background(), "cascade heartbeat failed: "+err.Error())
	} else if resp.Status == "config_changed" {
		cc.syncConfig()
	}
}

// ==================== Token Sync ====================

func (cc *CascadeClient) tokenSyncLoop() {
	lastSync := int64(0)
	for {
		time.Sleep(tokenSyncInterval)
		lastSync = cc.syncTokens(lastSync)
	}
}

func (cc *CascadeClient) syncTokens(since int64) int64 {
	var resp cascade.TokenSyncResponse
	if err := cc.callMaster("GET", fmt.Sprintf("/api/cascade/tokens/sync?since=%d", since), nil, &resp); err != nil {
		logger.Warn(context.Background(), "cascade token sync failed: "+err.Error())
		return since
	}

	if len(resp.Tokens) == 0 {
		return since
	}

	// Write all tokens into local Redis
	ctx := context.Background()
	latest := since

	for _, t := range resp.Tokens {
		data, _ := json.Marshal(t)
		key := redisTokenPrefix + t.KeyHash

		// Check if token should be cached (deleted tokens get written as empty to signal invalidation)
		if t.Status == 0 { // deleted
			common.RDB.Del(ctx, key)
		} else {
			common.RDB.Set(ctx, key, string(data), redisTTLSeconds*time.Second)
		}

		if t.UpdatedTime > latest {
			latest = t.UpdatedTime
		}
	}

	logger.Debug(context.Background(), fmt.Sprintf("cascade: synced %d tokens (since=%d)", len(resp.Tokens), since))
	return latest
}

// ==================== Billing Buffer & Flush ====================

// AddBillingRecord appends a billing record to the local buffer.
func (cc *CascadeClient) AddBillingRecord(rec cascade.BillingRecord) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.billingBuf = append(cc.billingBuf, rec)
}

func (cc *CascadeClient) billingFlushLoop() {
	for {
		time.Sleep(billingFlushInterval)
		cc.flushBilling()
	}
}

func (cc *CascadeClient) flushBilling() {
	cc.mu.Lock()
	bufLen := len(cc.billingBuf)
	if bufLen == 0 {
		cc.mu.Unlock()
		return
	}
	// Take a snapshot of the buffer
	batch := make([]cascade.BillingRecord, bufLen)
	copy(batch, cc.billingBuf)
	cc.billingBuf = nil
	cc.mu.Unlock()

	// Generate batch ID
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d|%d|%d", cc.nodeID, time.Now().UnixNano(), bufLen)))
	batchID := hex.EncodeToString(hash[:16])

	// Calculate total
	var total int64
	for _, r := range batch {
		total += r.PriceCents
	}

	req := cascade.BillingBatchRequest{
		NodeID:      cc.nodeID,
		BatchID:     batchID,
		Records:     batch,
		TotalAmount: total,
	}

	var resp cascade.BillingBatchResponse
	if err := cc.callMaster("POST", "/api/cascade/billing/batch", req, &resp); err != nil {
		logger.Warn(context.Background(), "cascade billing flush failed: "+err.Error())
		// On failure, push records back to buffer
		cc.mu.Lock()
		cc.billingBuf = append(batch, cc.billingBuf...)
		cc.mu.Unlock()
		return
	}

	if resp.Rejected > 0 {
		logger.Warn(context.Background(), fmt.Sprintf("cascade: %d billing records rejected", resp.Rejected))
		for _, e := range resp.Errors {
			logger.Warn(context.Background(), "  rejected: "+e.IdempotencyKey+" reason="+e.Reason)
		}
	}

	logger.Debug(context.Background(), fmt.Sprintf("cascade: flushed %d billing records (accepted=%d)", len(batch), resp.Accepted))
}

// ==================== Config Sync ====================

func (cc *CascadeClient) configSyncLoop() {
	for {
		time.Sleep(configSyncInterval)
		cc.syncConfig()
	}
}

func (cc *CascadeClient) syncConfig() {
	var resp cascade.ConfigSyncResponse
	if err := cc.callMaster("GET", "/api/cascade/config", nil, &resp); err != nil {
		logger.Warn(context.Background(), "cascade config sync failed: "+err.Error())
		return
	}

	cc.modelRatios = resp.ModelRatios
	cc.configVersion = resp.Version

	logger.Debug(context.Background(), fmt.Sprintf("cascade: synced config version %d", resp.Version))
}

// ==================== Token Auth Helper (for middleware) ====================

// CascadeGetCachedToken looks up a token from the local Redis cache.
// Returns nil if not found or expired.
func CascadeGetCachedToken(keyHash string) *cascade.TokenSyncItem {
	if common.RDB == nil {
		return nil
	}
	ctx := context.Background()
	data, err := common.RDB.Get(ctx, redisTokenPrefix+keyHash).Result()
	if err != nil {
		return nil
	}
	var token cascade.TokenSyncItem
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil
	}
	return &token
}

// CascadeInvalidateToken removes a token from the local cache.
func CascadeInvalidateToken(keyHash string) {
	if common.RDB == nil {
		return
	}
	common.RDB.Del(context.Background(), redisTokenPrefix+keyHash)
}
