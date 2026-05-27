package model

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// CascadeNode represents a registered slave node.
type CascadeNode struct {
	Id             int    `json:"id" gorm:"primaryKey"`
	Name           string `json:"name" gorm:"type:varchar(64);uniqueIndex"`
	Region         string `json:"region" gorm:"type:varchar(32);index"`
	APIKeyHash     string `json:"-" gorm:"type:char(60)"`                          // bcrypt hash
	APIKeyPrefix   string `json:"api_key_prefix" gorm:"type:char(8)"`              // first 8 chars for admin display
	Status         int    `json:"status" gorm:"default:1"`                         // 1=online 2=offline 3=disabled
	LastHeartbeat  int64  `json:"last_heartbeat" gorm:"bigint"`
	Version        string `json:"version" gorm:"type:varchar(32)"`
	ChannelCount   int    `json:"channel_count" gorm:"default:0"`
	TodayCalls     int64  `json:"today_calls" gorm:"default:0"`
	MasterURL      string `json:"master_url" gorm:"type:varchar(256)"`             // slave uses this to reach master
	CreatedTime    int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime    int64  `json:"updated_time" gorm:"bigint"`
}

// CascadeRegister inserts a new node and returns the generated API key.
func CascadeRegister(name, region, version string) (*CascadeNode, string, error) {
	raw := make([]byte, 16)
	rand.Read(raw) // nolint:errcheck
	apiKey := "qcn_" + hex.EncodeToString(raw)
	hashed, err := common.Password2Hash(apiKey)
	if err != nil {
		return nil, "", err
	}

	now := common.GetTimestamp()
	node := &CascadeNode{
		Name:         name,
		Region:       region,
		APIKeyHash:   hashed,
		APIKeyPrefix: apiKey[:8],
		Status:       1,
		Version:      version,
		CreatedTime:  now,
		UpdatedTime:  now,
	}

	if err := DB.Create(node).Error; err != nil {
		return nil, "", err
	}
	return node, apiKey, nil
}

// CascadeGetNodeByAPIKey looks up a node by the bcrypt-hashed API key.
func CascadeGetNodeByAPIKey(apiKey string) (*CascadeNode, error) {
	// We check all nodes with matching prefix first (limited cardinality)
	var nodes []CascadeNode
	prefix := apiKey[:8]
	if err := DB.Where("api_key_prefix = ? AND status = 1", prefix).Find(&nodes).Error; err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if common.ValidatePasswordAndHash(apiKey, n.APIKeyHash) {
			return &n, nil
		}
	}
	return nil, nil // not found
}

// CascadeGetNodeByID returns a node by its ID.
func CascadeGetNodeByID(id int) (*CascadeNode, error) {
	var node CascadeNode
	err := DB.Where("id = ?", id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// CascadeGetAllNodes returns all registered nodes.
func CascadeGetAllNodes() ([]CascadeNode, error) {
	var nodes []CascadeNode
	err := DB.Order("id desc").Find(&nodes).Error
	return nodes, err
}

// CascadeUpdateHeartbeat updates the node's heartbeat fields.
func CascadeUpdateHeartbeat(id int, channelCount int, todayCalls int64) {
	now := common.GetTimestamp()
	err := DB.Model(&CascadeNode{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_heartbeat": now,
		"updated_time":   now,
		"channel_count":  channelCount,
		"today_calls":    todayCalls,
		"status":         1,
	}).Error
	if err != nil {
		logger.SysError("CascadeUpdateHeartbeat: " + err.Error())
	}
}

// CascadeMarkNodeOffline marks all nodes that haven't heartbeated in 120s as offline.
func CascadeMarkNodeOffline() {
	threshold := common.GetTimestamp() - 120
	err := DB.Model(&CascadeNode{}).
		Where("status = 1 AND last_heartbeat < ?", threshold).
		Update("status", 2).Error
	if err != nil {
		logger.SysError("CascadeMarkNodeOffline: " + err.Error())
	}
}

// CascadeDeleteNode removes a node registration.
func CascadeDeleteNode(id int) error {
	return DB.Delete(&CascadeNode{}, id).Error
}

// 
//  Billing Batch (from slave)
// 

// BillingBatchStatus consts.
const (
	BillingBatchPending   = 0
	BillingBatchConfirmed = 1
	BillingBatchRejected  = 2
)

// CascadeBillingBatch stores a batch of billing records pushed by a slave node.
type CascadeBillingBatch struct {
	Id              int    `json:"id" gorm:"primaryKey"`
	NodeID          int    `json:"node_id" gorm:"index"`
	BatchID         string `json:"batch_id" gorm:"type:varchar(64);uniqueIndex"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
	ConfirmedAt     *int64 `json:"confirmed_at" gorm:"bigint"`
	RecordCount     int    `json:"record_count"`
	TotalAmount     int64  `json:"total_amount"`     // sum(price_cents)
	Status          int    `json:"status" gorm:"default:0"`
}

// CreateBillingBatch inserts a new billing batch from a slave.
func CreateBillingBatch(nodeID int, batchID string, recordCount int, totalAmount int64) (*CascadeBillingBatch, error) {
	now := common.GetTimestamp()
	batch := &CascadeBillingBatch{
		NodeID:      nodeID,
		BatchID:     batchID,
		CreatedAt:   now,
		RecordCount: recordCount,
		TotalAmount: totalAmount,
		Status:      BillingBatchPending,
	}
	err := DB.Create(batch).Error
	return batch, err
}

// ConfirmBillingBatch marks a batch as confirmed (deductions applied).
func ConfirmBillingBatch(id int) {
	now := common.GetTimestamp()
	DB.Model(&CascadeBillingBatch{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       BillingBatchConfirmed,
		"confirmed_at": now,
	})
}

// RejectBillingBatch marks a batch as rejected.
func RejectBillingBatch(id int, reason string) {
	DB.Model(&CascadeBillingBatch{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": BillingBatchRejected,
	})
	logger.SysError("BillingBatch rejected: " + reason)
}
