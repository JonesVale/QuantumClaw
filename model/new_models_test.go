package model_test

import (
	"testing"

	"github.com/quantumclaw/quantumclaw/model"
)

// TestModelPackageImports verifies that all model package modules compile
// together. If a new model file has a syntax/import error, this test fails.
// Each model sub-package symbol referenced here ensures the full dependency
// graph resolves correctly.
func TestModelPackageImports(t *testing.T) {
	// Verify failure_tracker types compile
	_ = model.ChannelFailureTracker{}
	t.Log("failure_tracker types OK")

	// Verify fee_config types compile
	_ = model.PlatformFeeConfig{}
	t.Log("fee_config types OK")

	// Verify idempotency types compile
	_ = model.PaymentIdempotencyKey{}
	t.Log("idempotency types OK")

	// Verify consume_record types compile
	_ = model.ConsumeRecord{}
	t.Log("consume_record types OK")

	// Verify store types compile
	_ = model.Store{}
	t.Log("store types OK")

	// Verify pool_agreement types compile
	_ = model.PlatformPoolAgreement{}
	t.Log("pool_agreement types OK")
}
