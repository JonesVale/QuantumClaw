package model

import (
	"encoding/json"
	"errors"
	"testing"
)

// ==================== Channel Methods ====================

func TestChannelGetPriority(t *testing.T) {
	tests := []struct {
		name     string
		channel  Channel
		want     int64
	}{
		{
			name:     "nil priority returns 0",
			channel:  Channel{},
			want:     0,
		},
		{
			name:     "non-nil priority returns value",
			channel:  Channel{Priority: int64Ptr(5)},
			want:     5,
		},
		{
			name:     "zero priority returns 0",
			channel:  Channel{Priority: int64Ptr(0)},
			want:     0,
		},
		{
			name:     "negative priority returns value",
			channel:  Channel{Priority: int64Ptr(-1)},
			want:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetPriority()
			if got != tt.want {
				t.Errorf("Channel.GetPriority() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestChannelGetBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
		want    string
	}{
		{
			name:    "nil base URL returns empty",
			channel: Channel{},
			want:    "",
		},
		{
			name:    "non-nil base URL returns value",
			channel: Channel{BaseURL: stringPtr("https://api.openai.com")},
			want:    "https://api.openai.com",
		},
		{
			name:    "empty base URL returns empty",
			channel: Channel{BaseURL: stringPtr("")},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetBaseURL()
			if got != tt.want {
				t.Errorf("Channel.GetBaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChannelGetModelMapping(t *testing.T) {
	tests := []struct {
		name     string
		channel  Channel
		wantNil  bool
		wantMap  map[string]string
	}{
		{
			name:    "nil model mapping returns nil",
			channel: Channel{},
			wantNil: true,
		},
		{
			name:    "empty model mapping returns nil",
			channel: Channel{ModelMapping: stringPtr("")},
			wantNil: true,
		},
		{
			name:    "empty JSON object returns nil",
			channel: Channel{ModelMapping: stringPtr("{}")},
			wantNil: true,
		},
		{
			name:    "valid JSON mapping returns parsed map",
			channel: Channel{ModelMapping: stringPtr(`{"gpt-3.5":"gpt-3.5-turbo","gpt-4":"gpt-4-turbo"}`)},
			wantNil: false,
			wantMap: map[string]string{"gpt-3.5": "gpt-3.5-turbo", "gpt-4": "gpt-4-turbo"},
		},
		{
			name:    "single entry mapping",
			channel: Channel{ModelMapping: stringPtr(`{"claude":"claude-3"}`)},
			wantNil: false,
			wantMap: map[string]string{"claude": "claude-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetModelMapping()
			if tt.wantNil {
				if got != nil {
					t.Errorf("Channel.GetModelMapping() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("Channel.GetModelMapping() returned nil, want non-nil map")
			}
			for k, v := range tt.wantMap {
				if got[k] != v {
					t.Errorf("Channel.GetModelMapping()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestChannelLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
		want    ChannelConfig
		wantErr bool
	}{
		{
			name:    "empty config returns empty config",
			channel: Channel{},
			want:    ChannelConfig{},
		},
		{
			name:    "valid JSON config",
			channel: Channel{Config: `{"region":"us-east-1","sk":"secret-key","ak":"access-key"}`},
			want:    ChannelConfig{Region: "us-east-1", SK: "secret-key", AK: "access-key"},
		},
		{
			name:    "partial config",
			channel: Channel{Config: `{"user_id":"u123","api_version":"v2"}`},
			want:    ChannelConfig{UserID: "u123", APIVersion: "v2"},
		},
		{
			name:    "config with all fields",
			channel: Channel{Config: `{"region":"us","sk":"sk1","ak":"ak1","user_id":"uid","api_version":"v3","library_id":"lib1","plugin":"plugin1","vertex_ai_project_id":"proj1","vertex_ai_adc":"adc1"}`},
			want:    ChannelConfig{Region: "us", SK: "sk1", AK: "ak1", UserID: "uid", APIVersion: "v3", LibraryID: "lib1", Plugin: "plugin1", VertexAIProjectID: "proj1", VertexAIADC: "adc1"},
		},
		{
			name:    "invalid JSON returns error",
			channel: Channel{Config: `{"invalid json`},
			want:    ChannelConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.channel.LoadConfig()
			if tt.wantErr {
				if err == nil {
					t.Error("Channel.LoadConfig() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Channel.LoadConfig() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Channel.LoadConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// ==================== Token Methods ====================

func TestTokenGetModels(t *testing.T) {
	tests := []struct {
		name  string
		token *Token
		want  string
	}{
		{
			name:  "nil token returns empty",
			token: nil,
			want:  "",
		},
		{
			name:  "nil Models field returns empty",
			token: &Token{},
			want:  "",
		},
		{
			name:  "empty Models field returns empty",
			token: &Token{Models: stringPtr("")},
			want:  "",
		},
		{
			name:  "single model",
			token: &Token{Models: stringPtr("gpt-4")},
			want:  "gpt-4",
		},
		{
			name:  "multiple comma-separated models",
			token: &Token{Models: stringPtr("gpt-4,gpt-3.5-turbo,claude-3")},
			want:  "gpt-4,gpt-3.5-turbo,claude-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.GetModels()
			if got != tt.want {
				t.Errorf("Token.GetModels() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ==================== Error Sentinel Values ====================

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrTokenNotProvided", ErrTokenNotProvided, "未提供令牌"},
		{"ErrTokenInvalid", ErrTokenInvalid, "无效的令牌"},
		{"ErrTokenExpired", ErrTokenExpired, "该令牌已过期"},
		{"ErrTokenExhausted", ErrTokenExhausted, "该令牌额度已用尽"},
		{"ErrQuotaNotEnough", ErrQuotaNotEnough, "配额不足"},
		{"ErrDatabase", ErrDatabase, "数据库错误"},
		{"ErrChannelNotFound", ErrChannelNotFound, "渠道不存在"},
		{"ErrChannelDisabled", ErrChannelDisabled, "渠道已被禁用"},
		{"ErrPermissionDenied", ErrPermissionDenied, "权限不足"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("error is nil, want non-nil")
				return
			}
			if tt.err.Error() != tt.msg {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.msg)
			}
			if !errors.Is(tt.err, tt.err) {
				t.Errorf("errors.Is(%v, %v) should be true", tt.err, tt.err)
			}
		})
	}
}

// ==================== Constants Validation ====================

func TestChannelStatusConstants(t *testing.T) {
	if ChannelStatusUnknown != 0 {
		t.Errorf("ChannelStatusUnknown = %d, want 0", ChannelStatusUnknown)
	}
	if ChannelStatusEnabled != 1 {
		t.Errorf("ChannelStatusEnabled = %d, want 1", ChannelStatusEnabled)
	}
	if ChannelStatusManuallyDisabled != 2 {
		t.Errorf("ChannelStatusManuallyDisabled = %d, want 2", ChannelStatusManuallyDisabled)
	}
	if ChannelStatusAutoDisabled != 3 {
		t.Errorf("ChannelStatusAutoDisabled = %d, want 3", ChannelStatusAutoDisabled)
	}
}

func TestTokenStatusConstants(t *testing.T) {
	if TokenStatusEnabled != 1 {
		t.Errorf("TokenStatusEnabled = %d, want 1", TokenStatusEnabled)
	}
	if TokenStatusDisabled != 2 {
		t.Errorf("TokenStatusDisabled = %d, want 2", TokenStatusDisabled)
	}
	if TokenStatusExpired != 3 {
		t.Errorf("TokenStatusExpired = %d, want 3", TokenStatusExpired)
	}
	if TokenStatusExhausted != 4 {
		t.Errorf("TokenStatusExhausted = %d, want 4", TokenStatusExhausted)
	}
}

func TestUserRoleConstants(t *testing.T) {
	if RoleGuestUser != 0 {
		t.Errorf("RoleGuestUser = %d, want 0", RoleGuestUser)
	}
	if RoleCommonUser != 1 {
		t.Errorf("RoleCommonUser = %d, want 1", RoleCommonUser)
	}
	if RoleAdminUser != 10 {
		t.Errorf("RoleAdminUser = %d, want 10", RoleAdminUser)
	}
	if RoleRootUser != 100 {
		t.Errorf("RoleRootUser = %d, want 100", RoleRootUser)
	}
}

func TestUserStatusConstants(t *testing.T) {
	if UserStatusEnabled != 1 {
		t.Errorf("UserStatusEnabled = %d, want 1", UserStatusEnabled)
	}
	if UserStatusDisabled != 2 {
		t.Errorf("UserStatusDisabled = %d, want 2", UserStatusDisabled)
	}
	if UserStatusDeleted != 3 {
		t.Errorf("UserStatusDeleted = %d, want 3", UserStatusDeleted)
	}
}

func TestLogTypeConstants(t *testing.T) {
	if LogTypeUnknown != 0 {
		t.Errorf("LogTypeUnknown = %d, want 0", LogTypeUnknown)
	}
	if LogTypeTopup != 1 {
		t.Errorf("LogTypeTopup = %d, want 1", LogTypeTopup)
	}
	if LogTypeConsume != 2 {
		t.Errorf("LogTypeConsume = %d, want 2", LogTypeConsume)
	}
	if LogTypeManage != 3 {
		t.Errorf("LogTypeManage = %d, want 3", LogTypeManage)
	}
	if LogTypeSystem != 4 {
		t.Errorf("LogTypeSystem = %d, want 4", LogTypeSystem)
	}
	if LogTypeTest != 5 {
		t.Errorf("LogTypeTest = %d, want 5", LogTypeTest)
	}
}

func TestRedemptionStatusConstants(t *testing.T) {
	if RedemptionCodeStatusEnabled != 1 {
		t.Errorf("RedemptionCodeStatusEnabled = %d, want 1", RedemptionCodeStatusEnabled)
	}
	if RedemptionCodeStatusDisabled != 2 {
		t.Errorf("RedemptionCodeStatusDisabled = %d, want 2", RedemptionCodeStatusDisabled)
	}
	if RedemptionCodeStatusUsed != 3 {
		t.Errorf("RedemptionCodeStatusUsed = %d, want 3", RedemptionCodeStatusUsed)
	}
}

// ==================== Channel Struct JSON ====================

func TestChannelJSONRoundTrip(t *testing.T) {
	inf := int64(42)
	ch := Channel{
		Id:                 1,
		Type:               2,
		Key:                "sk-test",
		Status:             ChannelStatusEnabled,
		Name:               "Test Channel",
		Weight:             uintPtr(10),
		CreatedTime:        1234567890,
		TestTime:           1234567891,
		ResponseTime:       200,
		BaseURL:            stringPtr("https://test.example.com"),
		Balance:            100.50,
		BalanceUpdatedTime: 1234567892,
		Models:             "gpt-4,gpt-3.5",
		Group:              "default",
		UsedQuota:          50000,
		ModelMapping:       stringPtr(`{"gpt-4":"gpt-4-turbo"}`),
		Priority:           &inf,
		Config:             `{"region":"us"}`,
	}

	// Marshal and unmarshal to verify JSON serialization
	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Channel
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Id != ch.Id {
		t.Errorf("Channel.Id = %d, want %d", decoded.Id, ch.Id)
	}
	if decoded.Type != ch.Type {
		t.Errorf("Channel.Type = %d, want %d", decoded.Type, ch.Type)
	}
	if decoded.Name != ch.Name {
		t.Errorf("Channel.Name = %q, want %q", decoded.Name, ch.Name)
	}
	if decoded.Group != ch.Group {
		t.Errorf("Channel.Group = %q, want %q", decoded.Group, ch.Group)
	}
	if decoded.Balance != ch.Balance {
		t.Errorf("Channel.Balance = %f, want %f", decoded.Balance, ch.Balance)
	}
}

// ==================== User Struct JSON ====================

func TestUserJSONRoundTrip(t *testing.T) {
	u := User{
		Id:          1,
		Username:    "testuser",
		Password:    "hashedpw",
		DisplayName: "Test User",
		Role:        RoleCommonUser,
		Status:      UserStatusEnabled,
		Email:       "test@example.com",
		Quota:       1000000,
		UsedQuota:   50000,
		RequestCount: 10,
		Group:       "default",
		AffCode:     "abcd",
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded User
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Username != u.Username {
		t.Errorf("User.Username = %q, want %q", decoded.Username, u.Username)
	}
	if decoded.Role != u.Role {
		t.Errorf("User.Role = %d, want %d", decoded.Role, u.Role)
	}
	if decoded.Status != u.Status {
		t.Errorf("User.Status = %d, want %d", decoded.Status, u.Status)
	}
}

// ==================== Token Struct JSON ====================

func TestTokenJSONRoundTrip(t *testing.T) {
	tok := Token{
		Id:             1,
		UserId:         1,
		Key:            "sk-test-key-123456789012345678901234567890",
		Status:         TokenStatusEnabled,
		Name:           "Default Token",
		CreatedTime:    1234567890,
		AccessedTime:   1234567891,
		ExpiredTime:    -1,
		RemainQuota:    1000000,
		UnlimitedQuota: true,
		UsedQuota:      50000,
		Models:         stringPtr("gpt-4,gpt-3.5"),
		Subnet:         stringPtr(""),
	}

	data, err := json.Marshal(tok)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Token
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Key != tok.Key {
		t.Errorf("Token.Key = %q, want %q", decoded.Key, tok.Key)
	}
	if decoded.UnlimitedQuota != tok.UnlimitedQuota {
		t.Errorf("Token.UnlimitedQuota = %v, want %v", decoded.UnlimitedQuota, tok.UnlimitedQuota)
	}
	if decoded.ExpiredTime != tok.ExpiredTime {
		t.Errorf("Token.ExpiredTime = %d, want %d", decoded.ExpiredTime, tok.ExpiredTime)
	}
}

// ==================== Helpers ====================

func int64Ptr(v int64) *int64 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func uintPtr(v uint) *uint {
	return &v
}
