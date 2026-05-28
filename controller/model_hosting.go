package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/i18n"
	"github.com/quantumclaw/quantumclaw/model"
)

// ListInferenceNodes 用户列出推理节点 — GET /api/user/inference-nodes
func ListInferenceNodes(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")
	nodes, err := model.GetUserInferenceNodes(ctx, uint(userID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": nodes})
}

// CreateInferenceNode 创建推理节点 — POST /api/user/inference-nodes
func CreateInferenceNode(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")

	var req struct {
		Name     string `json:"name" validate:"required,max=100"`
		NodeType string `json:"node_type" validate:"required,oneof=vllm sglang ollama"`
		BaseURL  string `json:"base_url" validate:"required,max=500"`
		APIKey   string `json:"api_key" validate:"max=500"`
	}
	if err := common.Validate.Struct(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}

	node := &model.InferenceNode{
		UserID:   uint(userID),
		Name:     req.Name,
		NodeType: req.NodeType,
		BaseURL:  strings.TrimRight(req.BaseURL, "/"),
		APIKey:   req.APIKey,
		Status:   "inactive",
	}
	if err := model.InsertInferenceNode(ctx, node); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": node})
}

// TestInferenceNode 测试推理节点连接 — POST /api/user/inference-nodes/:id/test
func TestInferenceNode(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_parameter")})
		return
	}

	node, err := model.GetInferenceNodeByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if node.UserID != uint(userID) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "permission_denied")})
		return
	}

	// Test connection by calling /v1/models
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, "GET", node.BaseURL+"/v1/models", nil)
	if node.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+node.APIKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		model.UpdateInferenceNodeStatus(ctx, node.ID, "error")
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Cannot connect: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		model.UpdateInferenceNodeStatus(ctx, node.ID, "error")
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Invalid response format"})
		return
	}

	modelNames := make([]string, 0, len(modelsResp.Data))
	for _, m := range modelsResp.Data {
		modelNames = append(modelNames, m.ID)
	}

	modelsJSON, _ := json.Marshal(modelNames)
	model.UpdateInferenceNodeModels(ctx, node.ID, string(modelsJSON), len(modelNames))
	model.UpdateInferenceNodeStatus(ctx, node.ID, "active")
	model.UpdateInferenceNode(ctx, &model.InferenceNode{
		ID:              node.ID,
		LastHealthCheck: time.Now().Unix(),
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    gin.H{"models": modelNames, "count": len(modelNames)},
	})
}

// DeleteInferenceNode 删除推理节点 — DELETE /api/user/inference-nodes/:id
func DeleteInferenceNode(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_parameter")})
		return
	}
	node, err := model.GetInferenceNodeByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if node.UserID != uint(userID) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "permission_denied")})
		return
	}
	if err := model.DeleteInferenceNode(ctx, uint(id)); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// UpdateInferenceNode 更新推理节点 — PUT /api/user/inference-nodes/:id
func UpdateInferenceNode(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_parameter")})
		return
	}
	node, err := model.GetInferenceNodeByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if node.UserID != uint(userID) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "permission_denied")})
		return
	}

	var req struct {
		Name    string `json:"name"`
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}
	if req.Name != "" {
		node.Name = req.Name
	}
	if req.BaseURL != "" {
		node.BaseURL = strings.TrimRight(req.BaseURL, "/")
	}
	if req.APIKey != "" {
		node.APIKey = req.APIKey
	}
	if err := model.UpdateInferenceNode(ctx, node); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": node})
}

// AdminListInferenceNodes 管理员查看所有推理节点 — GET /api/admin/inference-nodes
func AdminListInferenceNodes(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, total, err := model.GetAllInferenceNodes(ctx, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": items, "total": total})
}
