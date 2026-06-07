package meta

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
)

type Meta struct {
	Mode         int
	ChannelType  int
	ChannelId    int
	TokenId      int
	TokenName    string
	UserId       int
	Group        string
	ModelMapping map[string]string
	// BaseURL is the proxy url set in the channel config
	BaseURL  string
	APIKey   string
	APIType  int
	Config   model.ChannelConfig
	IsStream bool
	// OriginModelName is the model name from the raw user request
	OriginModelName string
	// ActualModelName is the model name after mapping
	ActualModelName    string
	RequestURLPath     string
	PromptTokens       int // only for DoResponse
	ForcedSystemPrompt string
	StartTime          time.Time

	// Settlement system fields
	PromoterId     int
	ChannelOwnerId int
	IsFallback     bool

	// ── 回退链追踪（Fallback Audit Trail）──
	// 当 relay 层发生 schema/渠道回退时，以下字段会被填充
	OriginalChannelId int      // 分发器最初选择的渠道 ID（回退前）
	FallbackChain     []FbStep // 每次回退的详细记录
	ActualSchemaId    int      // 最终实际使用的 Sub2API Schema ID
}

// FbStep 记录一次回退的详细信息
type FbStep struct {
	FromSchema int `json:"from_schema"` // 回退前的 Schema ID
	ToSchema   int `json:"to_schema"`   // 回退到的 Schema ID
}

func GetByContext(c *gin.Context) *Meta {
	meta := Meta{
		Mode:               relaymode.GetByPath(c.Request.URL.Path),
		ChannelType:        c.GetInt(ctxkey.Channel),
		ChannelId:          c.GetInt(ctxkey.ChannelId),
		TokenId:            c.GetInt(ctxkey.TokenId),
		TokenName:          c.GetString(ctxkey.TokenName),
		UserId:             c.GetInt(ctxkey.Id),
		Group:              c.GetString(ctxkey.Group),
		ModelMapping:       c.GetStringMapString(ctxkey.ModelMapping),
		OriginModelName:    c.GetString(ctxkey.RequestModel),
		PromoterId:         c.GetInt("promoter_id"),
		ChannelOwnerId:     c.GetInt("channel_owner"),
		IsFallback:         c.GetBool("is_fallback"),
		BaseURL:            c.GetString(ctxkey.BaseURL),
		APIKey:             strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		RequestURLPath:     c.Request.URL.String(),
		ForcedSystemPrompt: c.GetString(ctxkey.SystemPrompt),
		StartTime:          time.Now(),
	}

	// ── 回退链追踪字段（可选，仅当回退发生时有值）──
	meta.OriginalChannelId = c.GetInt("original_channel_id")
	if rawChain, exists := c.Get("fallback_chain"); exists {
		if chain, ok := rawChain.([]map[string]int); ok {
			meta.FallbackChain = make([]FbStep, len(chain))
			for i, step := range chain {
				meta.FallbackChain[i] = FbStep{
					FromSchema: step["from_schema"],
					ToSchema:   step["to_schema"],
				}
			}
		}
	}
	meta.ActualSchemaId = c.GetInt("actual_schema_id")

	cfg, ok := c.Get(ctxkey.Config)
	if ok {
		meta.Config = cfg.(model.ChannelConfig)
	}
	if meta.BaseURL == "" {
		meta.BaseURL = channeltype.ChannelBaseURLs[meta.ChannelType]
	}
	meta.APIType = channeltype.ToAPIType(meta.ChannelType)
	return &meta
}
