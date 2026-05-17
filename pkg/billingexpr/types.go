package billingexpr

import (
	"crypto/sha256"
	"fmt"
)

// RequestInput captures the raw HTTP request for billing expression evaluation.
type RequestInput struct {
	Headers map[string]string
	Body    []byte
}

// BillingContext holds runtime parameters for billing expression evaluation.
type BillingContext struct {
	PromptTokens     float64
	CompletionTokens float64
	TotalTokens      float64
	ImageCount       float64
	AudioSeconds     float64
	CustomRatio      float64
	ChannelRatio     float64
	GroupRatio       float64
	ModelRatio       float64
	APIKey           string
	ModelName        string
	GroupName        string
	ChannelName      string
}

// TokenParams holds all token dimensions for billing expression evaluation.
type TokenParams struct {
	P    float64 // prompt tokens (text)
	C    float64 // completion tokens (text)
	Len  float64 // total input context length
	CR   float64 // cache read (hit) tokens
	CC   float64 // cache creation tokens (generic)
	CC1h float64 // cache creation tokens (1-hour TTL)
	Img  float64 // image input tokens
	ImgO float64 // image output tokens
	AI   float64 // audio input tokens
	AO   float64 // audio output tokens
}

// TraceResult holds side-channel info from tier() function during Expr execution.
type TraceResult struct {
	MatchedTier string  `json:"matched_tier"`
	Cost        float64 `json:"cost"`
}

// BillingSnapshot captures the billing rule state frozen at pre-consume time.
type BillingSnapshot struct {
	BillingMode               string  `json:"billing_mode"`
	ModelName                 string  `json:"model_name"`
	ExprString                string  `json:"expr_string"`
	ExprHash                  string  `json:"expr_hash"`
	GroupRatio                float64 `json:"group_ratio"`
	EstimatedPromptTokens     int     `json:"estimated_prompt_tokens"`
	EstimatedCompletionTokens int     `json:"estimated_completion_tokens"`
	EstimatedQuotaBeforeGroup float64 `json:"estimated_quota_before_group"`
	EstimatedQuotaAfterGroup  int     `json:"estimated_quota_after_group"`
	EstimatedTier             string  `json:"estimated_tier"`
	QuotaPerUnit              float64 `json:"quota_per_unit"`
	ExprVersion               int     `json:"expr_version"`
}

// TieredResult holds everything needed after running tiered settlement.
type TieredResult struct {
	ActualQuotaBeforeGroup float64 `json:"actual_quota_before_group"`
	ActualQuotaAfterGroup  int     `json:"actual_quota_after_group"`
	MatchedTier            string  `json:"matched_tier"`
	CrossedTier            bool    `json:"crossed_tier"`
}

type ExprVariables struct {
	P     float64 `expr:"p"`     // prompt tokens
	C     float64 `expr:"c"`     // completion tokens
	Len   float64 `expr:"len"`   // total tokens
	Cr    float64 `expr:"cr"`    // channel ratio
	Cc    float64 `expr:"cc"`    // custom ratio
	Cc1h  float64 `expr:"cc1h"`  // custom ratio for 1h
	Img   float64 `expr:"img"`   // image count
	ImgO  float64 `expr:"img_o"` // original image count
	Ai    float64 `expr:"ai"`    // audio input seconds
	Ao    float64 `expr:"ao"`    // audio output seconds
}

func (v *ExprVariables) FromContext(ctx *BillingContext) {
	v.P = ctx.PromptTokens
	v.C = ctx.CompletionTokens
	v.Len = ctx.TotalTokens
	v.Cr = ctx.ChannelRatio
	v.Cc = ctx.CustomRatio
	v.Img = ctx.ImageCount
	v.Ai = ctx.AudioSeconds
	v.Ao = ctx.AudioSeconds
}

func ExprHashString(exprStr string) string {
	h := sha256.New()
	h.Write([]byte(exprStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}
