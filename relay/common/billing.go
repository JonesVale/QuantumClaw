package common

import "github.com/gin-gonic/gin"

// BillingSettler abstracts the billing session lifecycle.
// Implemented by service.BillingSession and stored on RelayInfo to avoid circular imports.
type BillingSettler interface {
	// Settle performs final settlement based on actual quota consumed.
	Settle(actualQuota int) error

	// Refund releases all pre-consumed quota asynchronously.
	Refund(c *gin.Context)

	// NeedsRefund returns true if there is pre-consumed quota that hasn't been settled or refunded.
	NeedsRefund() bool

	// GetPreConsumedQuota returns the actual pre-consumed quota value.
	GetPreConsumedQuota() int

	// Reserve increases pre-consumed quota to the target value.
	Reserve(targetQuota int) error
}
