package common

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/dto"
	"github.com/quantumclaw/quantumclaw/pkg/billingexpr"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
	"github.com/quantumclaw/quantumclaw/types"
)

// ThinkingContentInfo tracks thinking/reasoning content in streaming responses.
type ThinkingContentInfo struct {
	IsFirstThinkingContent  bool
	SendLastThinkingContent bool
	HasSentThinkingContent  bool
}

const (
	LastMessageTypeNone     = "none"
	LastMessageTypeText     = "text"
	LastMessageTypeTools    = "tools"
	LastMessageTypeThinking = "thinking"
)

// ClaudeConvertInfo tracks state when converting between Claude and OpenAI formats.
type ClaudeConvertInfo struct {
	LastMessagesType      string
	Index                 int
	Usage                 *dto.Usage
	FinishReason          string
	Done                  bool
	ToolCallBaseIndex     int
	ToolCallMaxIndexOffset int
}

// RerankerInfo holds state for reranking relay.
type RerankerInfo struct {
	Documents       []any
	ReturnDocuments bool
}

// BuildInToolInfo tracks usage of built-in tools (Responses API).
type BuildInToolInfo struct {
	ToolName          string
	CallCount         int
	SearchContextSize string
}

// ResponsesUsageInfo tracks usage for OpenAI Responses API.
type ResponsesUsageInfo struct {
	BuiltInTools map[string]*BuildInToolInfo
}

// ChannelMeta holds channel-level metadata for the relay.
type ChannelMeta struct {
	ChannelType          int
	ChannelID            int
	ChannelIsMultiKey    bool
	ChannelMultiKeyIndex int
	ChannelBaseURL       string
	APIType              int
	APIVersion           string
	APIKey               string
	Organization         string
	ChannelCreateTime    int64
	ParamOverride        map[string]interface{}
	HeadersOverride      map[string]interface{}
	ChannelSetting       dto.ChannelSettings
	ChannelOtherSettings dto.ChannelOtherSettings
	UpstreamModelName    string
	IsModelMapped        bool
	SupportStreamOptions bool
}

// TokenCountMeta holds prompt token estimation metadata.
type TokenCountMeta struct {
	estimatePromptTokens int
}

// TaskRelayInfo holds metadata for asynchronous task relay (MJ, Video, Suno, etc.).
type TaskRelayInfo struct {
	Action       string
	OriginTaskID string
	PublicTaskID string
	ConsumeQuota bool
	LockedChannel any
}

// RelayInfo is the unified relay context for processing a single request.
type RelayInfo struct {
	TokenID           int
	TokenKey          string
	TokenGroup        string
	UserID            int
	UsingGroup        string
	UserGroup         string
	TokenUnlimited    bool
	StartTime         time.Time
	FirstResponseTime time.Time
	isFirstResponse   bool

	IsStream               bool
	UsePrice               bool
	RelayMode              int
	OriginModelName        string

	// 结算系统字段（由 Distribute 中间件写入）
	PromoterId     int
	ChannelOwnerId int
	ChannelId      int
	RequestURLPath         string
	RequestHeaders         map[string]string
	ShouldIncludeUsage     bool
	DisablePing            bool

	InputAudioFormat       string
	OutputAudioFormat      string
	RealtimeTools          []dto.RealTimeTool
	IsFirstRequest         bool
	AudioUsage             bool
	ReasoningEffort        string
	UserSetting            dto.UserSetting
	UserEmail              string
	UserQuota              int
	RelayFormat            types.RelayFormat
	SendResponseCount      int
	ReceivedResponseCount  int
	FinalPreConsumedQuota  int
	ForcePreConsume        bool

	Billing BillingSettler

	BillingSource string
	SubscriptionID           int
	SubscriptionPreConsumed  int64
	SubscriptionPostDelta    int64
	SubscriptionPlanID       int
	SubscriptionPlanTitle    string
	RequestID                string
	SubscriptionAmountTotal               int64
	SubscriptionAmountUsedAfterPreConsume int64

	IsChannelTest bool
	RetryIndex    int

	PriceData types.PriceData

	TieredBillingSnapshot *billingexpr.BillingSnapshot
	BillingRequestInput   *billingexpr.RequestInput

	Request dto.Request

	RequestConversionChain  []types.RelayFormat
	FinalRequestRelayFormat types.RelayFormat

	StreamStatus *StreamStatus

	ThinkingContentInfo
	TokenCountMeta
	*ClaudeConvertInfo
	*RerankerInfo
	*ResponsesUsageInfo
	*ChannelMeta
	*TaskRelayInfo
}

// InitRequestConversionChain initializes the request conversion chain.
func (info *RelayInfo) InitRequestConversionChain() {
	if info == nil {
		return
	}
	if len(info.RequestConversionChain) > 0 {
		return
	}
	if info.RelayFormat == "" {
		return
	}
	info.RequestConversionChain = []types.RelayFormat{info.RelayFormat}
}

// AppendRequestConversion appends a format to the conversion chain.
func (info *RelayInfo) AppendRequestConversion(format types.RelayFormat) {
	if info == nil || format == "" {
		return
	}
	if len(info.RequestConversionChain) == 0 {
		info.RequestConversionChain = []types.RelayFormat{format}
		return
	}
	last := info.RequestConversionChain[len(info.RequestConversionChain)-1]
	if last == format {
		return
	}
	info.RequestConversionChain = append(info.RequestConversionChain, format)
}

// GetFinalRequestRelayFormat returns the final relay format sent to the upstream.
func (info *RelayInfo) GetFinalRequestRelayFormat() types.RelayFormat {
	if info == nil {
		return ""
	}
	if info.FinalRequestRelayFormat != "" {
		return info.FinalRequestRelayFormat
	}
	if n := len(info.RequestConversionChain); n > 0 {
		return info.RequestConversionChain[n-1]
	}
	return info.RelayFormat
}

// SetEstimatePromptTokens sets the estimated prompt token count.
func (info *RelayInfo) SetEstimatePromptTokens(promptTokens int) {
	info.estimatePromptTokens = promptTokens
}

// GetEstimatePromptTokens returns the estimated prompt token count.
func (info *RelayInfo) GetEstimatePromptTokens() int {
	return info.estimatePromptTokens
}

// SetFirstResponseTime records the first response time (once).
func (info *RelayInfo) SetFirstResponseTime() {
	if info.isFirstResponse {
		info.FirstResponseTime = time.Now()
		info.isFirstResponse = false
	}
}

// HasSentResponse returns true if the first response has been sent.
func (info *RelayInfo) HasSentResponse() bool {
	return info.FirstResponseTime.After(info.StartTime)
}

// ToString returns a debug string representation of the RelayInfo.
func (info *RelayInfo) ToString() string {
	if info == nil {
		return "RelayInfo<nil>"
	}
	b := &strings.Builder{}
	b.WriteString("RelayInfo{ ")
	b.WriteString("RelayFormat: ")
	b.WriteString(string(info.RelayFormat))
	b.WriteString(", RelayMode: ")
	b.WriteString(string(rune(info.RelayMode)))
	b.WriteString(", IsStream: ")
	b.WriteString(boolToString(info.IsStream))
	b.WriteString(", OriginModelName: ")
	b.WriteString(info.OriginModelName)
	b.WriteString(", UserID: ")
	b.WriteString(string(rune(info.UserID)))
	b.WriteString(" }")
	return b.String()
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// GenRelayInfoOpenAI creates a RelayInfo for OpenAI-format requests.
func GenRelayInfoOpenAI(c *gin.Context, request dto.Request) *RelayInfo {
	info := genBaseRelayInfo(c, request)
	info.RelayFormat = types.RelayFormatOpenAI
	return info
}

// GenRelayInfoClaude creates a RelayInfo for Claude-format requests.
func GenRelayInfoClaude(c *gin.Context, request dto.Request) *RelayInfo {
	info := genBaseRelayInfo(c, request)
	info.RelayFormat = types.RelayFormatClaude
	info.ShouldIncludeUsage = false
	info.ClaudeConvertInfo = &ClaudeConvertInfo{
		LastMessagesType: LastMessageTypeNone,
	}
	return info
}

// GenRelayInfoGemini creates a RelayInfo for Gemini-format requests.
func GenRelayInfoGemini(c *gin.Context, request dto.Request) *RelayInfo {
	info := genBaseRelayInfo(c, request)
	info.RelayFormat = types.RelayFormatGemini
	info.ShouldIncludeUsage = false
	return info
}

// GenRelayInfoImage creates a RelayInfo for image generation requests.
func GenRelayInfoImage(c *gin.Context, request dto.Request) *RelayInfo {
	info := genBaseRelayInfo(c, request)
	info.RelayFormat = types.RelayFormatOpenAIImage
	return info
}

// GenRelayInfoAudio creates a RelayInfo for audio requests.
func GenRelayInfoAudio(c *gin.Context, request dto.Request) *RelayInfo {
	info := genBaseRelayInfo(c, request)
	info.RelayFormat = types.RelayFormatOpenAIAudio
	return info
}

// GenRelayInfoEmbedding creates a RelayInfo for embedding requests.
func GenRelayInfoEmbedding(c *gin.Context, request dto.Request) *RelayInfo {
	info := genBaseRelayInfo(c, request)
	info.RelayFormat = types.RelayFormatEmbedding
	return info
}

// genBaseRelayInfo creates a RelayInfo from the gin context, populating common fields.
func genBaseRelayInfo(c *gin.Context, request dto.Request) *RelayInfo {
	isStream := false
	if request != nil {
		isStream = request.IsStream(c)
	}

	startTime := time.Now()

	info := &RelayInfo{
		Request: request,

		UserID:     c.GetInt("id"),
		UserGroup:  c.GetString("group"),
		UserQuota:  c.GetInt("quota"),
		UserEmail:  c.GetString("email"),

		TokenID:        c.GetInt("token_id"),
		TokenKey:       c.GetString("token_key"),
		TokenUnlimited: c.GetBool("token_unlimited"),
		TokenGroup:     c.GetString("token_group"),

		OriginModelName: c.GetString("original_model"),
		PromoterId:     c.GetInt("promoter_id"),
		ChannelOwnerId: c.GetInt("channel_owner"),
		ChannelId:      c.GetInt("channel_id"),

		RelayMode:      relaymode.GetByPath(c.Request.URL.Path),
		RequestURLPath: c.Request.URL.String(),
		IsStream:       isStream,

		StartTime:         startTime,
		FirstResponseTime: startTime.Add(-time.Second),
		isFirstResponse:   true,

		ThinkingContentInfo: ThinkingContentInfo{
			IsFirstThinkingContent:  true,
			SendLastThinkingContent: false,
		},
	}

	if info.RelayMode == relaymode.Unknown {
		info.RelayMode = c.GetInt("relay_mode")
	}

	if info.TokenGroup == "" {
		info.TokenGroup = info.UserGroup
	}

	return info
}
