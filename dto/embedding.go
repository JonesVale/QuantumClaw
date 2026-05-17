package dto

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/types"
)

// EmbeddingRequest represents an OpenAI embedding API request.
type EmbeddingRequest struct {
	Model       string    `json:"model"`
	Input       any       `json:"input"`
	EncodingFormat string `json:"encoding_format,omitempty"`
	Dimensions  *int      `json:"dimensions,omitempty"`
	User        string    `json:"user,omitempty"`
}

func (r *EmbeddingRequest) GetTokenCountMeta() *types.TokenCountMeta {
	var texts []string
	if r.Input != nil {
		switch v := r.Input.(type) {
		case string:
			texts = append(texts, v)
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					texts = append(texts, s)
				}
			}
		}
	}
	return &types.TokenCountMeta{
		CombineText: strings.Join(texts, "\n"),
	}
}

func (r *EmbeddingRequest) IsStream(c *gin.Context) bool {
	return false
}

func (r *EmbeddingRequest) SetModelName(modelName string) {
	if modelName != "" {
		r.Model = modelName
	}
}
