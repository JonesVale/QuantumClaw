package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ==================== LoginRequest Struct ====================

func TestLoginRequestJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantUser string
		wantPass string
		wantErr  bool
	}{
		{
			name:     "valid login request",
			input:    `{"username":"admin","password":"secret123"}`,
			wantUser: "admin",
			wantPass: "secret123",
		},
		{
			name:    "empty input",
			input:   `{}`,
			wantErr: false, // Go will just decode empty struct
		},
		{
			name:    "missing fields",
			input:   `{"username":"admin"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req LoginRequest
			err := json.Unmarshal([]byte(tt.input), &req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantUser != "" {
				assert.Equal(t, tt.wantUser, req.Username)
			}
			if tt.wantPass != "" {
				assert.Equal(t, tt.wantPass, req.Password)
			}
		})
	}
}

// ==================== PasswordResetRequest Struct ====================

func TestPasswordResetRequestJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantEmail string
		wantToken string
	}{
		{
			name:      "valid reset request",
			input:     `{"email":"user@example.com","token":"reset-token-123"}`,
			wantEmail: "user@example.com",
			wantToken: "reset-token-123",
		},
		{
			name:      "empty fields",
			input:     `{}`,
			wantEmail: "",
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req PasswordResetRequest
			err := json.Unmarshal([]byte(tt.input), &req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantEmail, req.Email)
			assert.Equal(t, tt.wantToken, req.Token)
		})
	}
}

// ==================== PerformanceStats Struct ====================

func TestPerformanceStatsJSON(t *testing.T) {
	stats := PerformanceStats{}
	stats.Memory.Alloc = 1024
	stats.Memory.TotalAlloc = 2048
	stats.Memory.Sys = 4096
	stats.Memory.NumGC = 5
	stats.Memory.GoRoutines = 10
	stats.Runtime.NumCPU = 8
	stats.Uptime.StartTime = 1234567890
	stats.Uptime.UptimeSeconds = 3600

	data, err := json.Marshal(stats)
	assert.NoError(t, err)

	var decoded PerformanceStats
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, stats.Memory.Alloc, decoded.Memory.Alloc)
	assert.Equal(t, stats.Memory.GoRoutines, decoded.Memory.GoRoutines)
	assert.Equal(t, stats.Runtime.NumCPU, decoded.Runtime.NumCPU)
	assert.Equal(t, stats.Uptime.StartTime, decoded.Uptime.StartTime)
	assert.Equal(t, stats.Uptime.UptimeSeconds, decoded.Uptime.UptimeSeconds)
}

func TestPerformanceStatsZeroValues(t *testing.T) {
	stats := PerformanceStats{}
	data, err := json.Marshal(stats)
	assert.NoError(t, err)

	var decoded PerformanceStats
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, uint64(0), decoded.Memory.Alloc)
	assert.Equal(t, uint32(0), decoded.Memory.NumGC)
	assert.Equal(t, 0, decoded.Memory.GoRoutines)
	assert.Equal(t, 0, decoded.Runtime.NumCPU)
	assert.Equal(t, int64(0), decoded.Uptime.StartTime)
}

// ==================== TwoFASetupResponse Struct ====================

func TestTwoFASetupResponseJSON(t *testing.T) {
	resp := TwoFASetupResponse{
		Secret:    "JBSWY3DPEHPK3PXP",
		QRCodeURL: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var decoded TwoFASetupResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, resp.Secret, decoded.Secret)
	assert.Equal(t, resp.QRCodeURL, decoded.QRCodeURL)
}

func TestTwoFASetupResponseEmpty(t *testing.T) {
	resp := TwoFASetupResponse{}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var decoded TwoFASetupResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "", decoded.Secret)
	assert.Equal(t, "", decoded.QRCodeURL)
}

// ==================== Gin Response Format ====================

func TestStandardJSONResponse(t *testing.T) {
	// Verify the standard response format used by all controllers
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Simulate the typical JSON response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    "test data",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"data":"test data"`)
}

func TestErrorJSONResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "error message",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":false`)
	assert.Contains(t, w.Body.String(), `"message":"error message"`)
}

func TestStatusJSONResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"version": "v1.0.0",
			"enabled": true,
			"count":   42,
		},
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"version":"v1.0.0"`)
	assert.Contains(t, w.Body.String(), `"enabled":true`)
	assert.Contains(t, w.Body.String(), `"count":42`)
}
