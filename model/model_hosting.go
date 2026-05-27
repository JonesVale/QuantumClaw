package model

import (
	"context"
	"time"
)

// InferenceNode 推理节点（用户自托管）
type InferenceNode struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	UserID          uint   `gorm:"index" json:"user_id"`
	Name            string `gorm:"size:100" json:"name"`
	NodeType        string `gorm:"size:20" json:"node_type"`          // vllm / sglang / ollama
	BaseURL         string `gorm:"size:500" json:"base_url"`          // http://192.168.1.100:8000
	APIKey          string `gorm:"size:500" json:"api_key"`           // stored as-is, user provides
	Models          string `gorm:"type:text" json:"models"`           // JSON array: ["model1","model2"]
	ModelCount      int    `gorm:"default:0" json:"model_count"`
	Status          string `gorm:"size:20;default:inactive" json:"status"` // active / inactive / error
	LastHealthCheck int64  `json:"last_health_check"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

func InsertInferenceNode(ctx context.Context, node *InferenceNode) error {
	now := time.Now().Unix()
	node.CreatedAt = now
	node.UpdatedAt = now
	return DB.WithContext(ctx).Create(node).Error
}

func GetUserInferenceNodes(ctx context.Context, userID uint) ([]InferenceNode, error) {
	var nodes []InferenceNode
	err := DB.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&nodes).Error
	return nodes, err
}

func GetInferenceNodeByID(ctx context.Context, id uint) (*InferenceNode, error) {
	var node InferenceNode
	err := DB.WithContext(ctx).First(&node, id).Error
	return &node, err
}

func UpdateInferenceNode(ctx context.Context, node *InferenceNode) error {
	node.UpdatedAt = time.Now().Unix()
	return DB.WithContext(ctx).Model(&InferenceNode{}).Where("id = ?", node.ID).Updates(map[string]interface{}{
		"name":     node.Name,
		"base_url": node.BaseURL,
		"api_key":  node.APIKey,
		"node_type": node.NodeType,
		"updated_at": node.UpdatedAt,
	}).Error
}

func DeleteInferenceNode(ctx context.Context, id uint) error {
	return DB.WithContext(ctx).Delete(&InferenceNode{}, id).Error
}

func GetAllActiveInferenceNodes(ctx context.Context) ([]InferenceNode, error) {
	var nodes []InferenceNode
	err := DB.WithContext(ctx).Where("status = ?", "active").Find(&nodes).Error
	return nodes, err
}

func UpdateInferenceNodeStatus(ctx context.Context, id uint, status string) error {
	return DB.WithContext(ctx).Model(&InferenceNode{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now().Unix()}).Error
}

func UpdateInferenceNodeModels(ctx context.Context, id uint, models string, count int) error {
	return DB.WithContext(ctx).Model(&InferenceNode{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"models":      models,
			"model_count": count,
			"updated_at":  time.Now().Unix(),
		}).Error
}

func GetAllInferenceNodes(ctx context.Context, page, pageSize int) ([]InferenceNode, int64, error) {
	var nodes []InferenceNode
	var total int64
	query := DB.WithContext(ctx).Model(&InferenceNode{})
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&nodes).Error
	return nodes, total, err
}
