package service

// FundingSource abstracts different types of funding sources for quota consumption.
// This allows supporting user balance, subscription allowances, prepaid tokens, etc.
type FundingSource interface {
	// PreConsume attempts to reserve quota from this source.
	// Returns the actually reserved quota (may be less than requested).
	PreConsume(quota int64) (int64, error)

	// PostConsume adjusts the source after actual consumption is known.
	// quota can be positive (consume more) or negative (refund).
	PostConsume(quota int64) error

	// Return releases pre-consumed quota back to the source.
	Return(quota int64)

	// Balance returns the current available balance of this source.
	Balance() int64
}

// UserQuotaSource implements FundingSource using the user's main quota balance.
type UserQuotaSource struct {
	UserID int
}

func (s *UserQuotaSource) PreConsume(quota int64) (int64, error) {
	return quota, PreConsumeQuota(0, s.UserID, quota)
}

func (s *UserQuotaSource) PostConsume(quota int64) error {
	return PostConsumeQuota(0, s.UserID, quota)
}

func (s *UserQuotaSource) Return(quota int64) {
	ReturnPreConsumedQuota(0, quota)
}

func (s *UserQuotaSource) Balance() int64 {
	return 0
}
