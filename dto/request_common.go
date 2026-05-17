package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/types"
)

// Request is the common interface for all request DTOs.
type Request interface {
	GetTokenCountMeta() *types.TokenCountMeta
	IsStream(c *gin.Context) bool
	SetModelName(modelName string)
}

// BaseRequest provides default implementations for the Request interface.
type BaseRequest struct{}

func (b *BaseRequest) GetTokenCountMeta() *types.TokenCountMeta {
	return &types.TokenCountMeta{
		TokenType: types.TokenTypeTokenizer,
	}
}

func (b *BaseRequest) IsStream(c *gin.Context) bool {
	return false
}

func (b *BaseRequest) SetModelName(modelName string) {}
