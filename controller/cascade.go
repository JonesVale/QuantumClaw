package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/cascade"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/billing/ratio"
)

// 
//  ?// 

// CascadeRegister ?// @Summary      Register cascade slave node
// @Description  Register a new slave node and get an API key for cascade communication
// @Tags         Cascade
// @Accept       json
// @Produce      json
// @Param        request body cascade.RegisterRequest true "Node registration info"
// @Success      200  {object}  cascade.RegisterResponse
// @Failure      400  {object}  map[string]interface{}
// @Router       /cascade/node/register [post]
func CascadeRegister(c *gin.Context) {
	var req cascade.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	node, apiKey, err := model.CascadeRegister(req.Name, req.Region, req.Version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	logger.SysLog("cascade node registered: " + req.Name + " (" + req.Region + ")")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": cascade.RegisterResponse{
			NodeID: node.Id,
			APIKey: apiKey,
		},
	})
}

// 
//  ?// 

// CascadeHeartbeat handles POST /api/cascade/node/heartbeat
func CascadeHeartbeat(c *gin.Context) {
	nodeID := c.GetInt("cascade_node_id")

	var req cascade.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	model.CascadeUpdateHeartbeat(nodeID, req.ChannelCount, req.TodayCalls)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": cascade.HeartbeatResponse{
			Status: "ok",
		},
	})
}

// 
//  Token 
// 

// CascadeTokenSync handles GET /api/cascade/tokens/sync?since=xxx
func CascadeTokenSync(c *gin.Context) {
	sinceStr := c.Query("since")
	since := int64(0)
	if sinceStr != "" {
		since, _ = strconv.ParseInt(sinceStr, 10, 64)
	}

	var tokens []model.Token
	db := model.DB.Where("updated_time >= ?", since).Order("updated_time asc").Limit(1000)

	// include deleted tokens (slave needs to know to invalidate cache)
	db.Unscoped().Find(&tokens)

	items := make([]cascade.TokenSyncItem, 0, len(tokens))
	for _, t := range tokens {
		// fetch user group/status
		user, err := model.GetUserById(t.UserId, false)
		userGroup := "default"
		userStatus := 1
		if err == nil && user != nil {
			userGroup = user.Group
			userStatus = user.Status
		}

		subnet := ""
		if t.Subnet != nil {
			subnet = *t.Subnet
		}
		models := ""
		if t.Models != nil {
			models = *t.Models
		}

		items = append(items, cascade.TokenSyncItem{
			ID:             t.Id,
			KeyHash:        t.KeyHash,
			UserID:         t.UserId,
			UserGroup:      userGroup,
			UserStatus:     userStatus,
			Status:         t.Status,
			Models:         models,
			Subnet:         subnet,
			RemainQuota:    t.RemainQuota,
			UnlimitedQuota: t.UnlimitedQuota,
			ExpiredTime:    t.ExpiredTime,
			UpdatedTime:    t.UpdatedTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": cascade.TokenSyncResponse{
			Tokens:  items,
			HasMore: len(items) >= 1000,
		},
	})
}

// 
//  
// 

// CascadeUserBatch handles POST /api/cascade/users/batch
func CascadeUserBatch(c *gin.Context) {
	var req cascade.UserBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}
	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": cascade.UserBatchResponse{Users: []cascade.UserBatchItem{}}})
		return
	}

	var users []model.User
	model.DB.Where("id IN ?", req.UserIDs).Find(&users)

	items := make([]cascade.UserBatchItem, 0, len(users))
	for _, u := range users {
		items = append(items, cascade.UserBatchItem{
			ID:     u.Id,
			Group:  u.Group,
			Status: u.Status,
			Quota:  u.Quota,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": cascade.UserBatchResponse{Users: items},
	})
}

// 
//  
// 

// CascadeBillingBatch handles POST /api/cascade/billing/batch
func CascadeBillingBatch(c *gin.Context) {
	nodeID := c.GetInt("cascade_node_id")

	var req cascade.BillingBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}
	if len(req.Records) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": cascade.BillingBatchResponse{Accepted: 0, Rejected: 0}})
		return
	}

	// 1. Check idempotency ?does this batch already exist?
	existing, _ := model.CreateBillingBatch(nodeID, req.BatchID, len(req.Records), req.TotalAmount)
	if existing == nil || existing.Id == 0 {
		// Duplicate batch ?report as already accepted
		logger.Warn(c.Request.Context(), "duplicate billing batch from node "+string(rune(nodeID))+": "+req.BatchID)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": cascade.BillingBatchResponse{
				Accepted: len(req.Records),
				Rejected: 0,
			},
		})
		return
	}

	// 2. Process each record
	accepted := 0
	rejected := 0
	var errors []cascade.BillingError

	for _, rec := range req.Records {
		if err := processBillingRecord(c, rec); err != nil {
			rejected++
			errors = append(errors, cascade.BillingError{
				IdempotencyKey: rec.IdempotencyKey,
				Reason:         err.Error(),
			})
			logger.Warn(c.Request.Context(), "billing record rejected: "+err.Error())
		} else {
			accepted++
		}
	}

	// 3. Mark batch as confirmed/rejected
	if rejected == 0 {
		model.ConfirmBillingBatch(existing.Id)
	} else {
		model.RejectBillingBatch(existing.Id, "rejected records")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": cascade.BillingBatchResponse{
			Accepted: accepted,
			Rejected: rejected,
			Errors:   errors,
		},
	})
}

// processBillingRecord applies a single deduction on the master node.
func processBillingRecord(c *gin.Context, rec cascade.BillingRecord) error {
	// Verify the user exists
	user, err := model.GetUserById(rec.UserID, false)
	if err != nil || user == nil {
		return err
	}

	// Verify the token
	token, err := model.GetTokenById(rec.TokenID)
	if err != nil {
		return err
	}

	// Deduct user quota
	if rec.Quota < 0 {
		if err := model.IncreaseUserQuota(rec.UserID, -rec.Quota); err != nil {
			return err
		}
	} else {
		if err := model.DecreaseUserQuota(rec.UserID, rec.Quota); err != nil {
			return err
		}
	}

	// Deduct token quota
	if !token.UnlimitedQuota {
		if err := model.DecreaseTokenQuota(rec.TokenID, rec.Quota); err != nil {
			return err
		}
	}

	// Record consume log
	model.RecordConsumeLog(c.Request.Context(), &model.Log{
		UserId:           rec.UserID,
		TokenName:        token.Name,
		ModelName:        rec.ModelName,
		PromptTokens:     rec.PromptTokens,
		CompletionTokens: rec.CompletionTokens,
		Quota:            int(rec.Quota),
		Content:          "cascade billing",
		CreatedAt:        rec.Timestamp,
	})

	return nil
}

// 
//  
// 

// CascadeConfigSync handles GET /api/cascade/config
func CascadeConfigSync(c *gin.Context) {
	// Build a snapshot of the global pricing config
	modelRatios := make(map[string]float64)
	for k, v := range ratio.ModelRatio {
		modelRatios[k] = v
	}
	completionRatios := make(map[string]float64)
	for k, v := range ratio.CompletionRatio {
		completionRatios[k] = v
	}

	// Build feature flags relevant to slave nodes
	// (These should match what's defined in common/config)
	features := map[string]bool{
		"enable_search":               true,
		"enable_prompt_optimizer":     true,
		"enable_param_validator":      true,
		"enable_geo_service":          false, // planned
		"enable_model_sync":           true,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": cascade.ConfigSyncResponse{
			Version:          time.Now().Unix(),
			ModelRatios:      modelRatios,
			CompletionRatios: completionRatios,
			Features:         features,
			UpdatedAt:        time.Now().Unix(),
		},
	})
}

// 
//  ?// 

// CascadeListNodes handles GET /api/cascade/nodes
func CascadeListNodes(c *gin.Context) {
	// Auto-mark stale nodes as offline before listing
	model.CascadeMarkNodeOffline()

	nodes, err := model.CascadeGetAllNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": nodes})
}

// CascadeDeleteNode handles DELETE /api/cascade/nodes/:id
func CascadeDeleteNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	if err := model.CascadeDeleteNode(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "deleted"})
}
