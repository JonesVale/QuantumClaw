package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	relaycommon "github.com/quantumclaw/quantumclaw/relay/common"
)

// BillingSource indicates where quota is deducted from.
type BillingSource string

const (
	BillingSourceWallet      BillingSource = "wallet"
	BillingSourceSubscription BillingSource = "subscription"
)

// BillingSession implements relaycommon.BillingSettler.
// It encapsulates pre-consumption, settlement, and refund across
// wallet (token+user) and subscription quota sources.
type BillingSession struct {
	mu sync.Mutex

	TokenId int
	UserId  int

	Source  BillingSource
	SubItem *model.UserSubscription

	preConsumedQuota int
	settled          bool
	refunded         bool
}

func NewBillingSession(tokenId, userId int) *BillingSession {
	return &BillingSession{
		TokenId: tokenId,
		UserId:  userId,
		Source:  BillingSourceWallet,
	}
}

func NewSubscriptionBillingSession(tokenId, userId int, sub *model.UserSubscription) *BillingSession {
	return &BillingSession{
		TokenId: tokenId,
		UserId:  userId,
		Source:  BillingSourceSubscription,
		SubItem: sub,
	}
}

func (s *BillingSession) Reserve(targetQuota int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.settled || s.refunded {
		return errors.New("billing session already settled or refunded")
	}
	if targetQuota <= s.preConsumedQuota {
		return nil
	}
	delta := targetQuota - s.preConsumedQuota

	switch s.Source {
	case BillingSourceWallet:
		if err := model.PreConsumeTokenQuota(s.TokenId, int64(delta)); err != nil {
			return fmt.Errorf("pre-consume token quota: %w", err)
		}
	case BillingSourceSubscription:
		if s.SubItem == nil || s.SubItem.Id == 0 {
			return errors.New("subscription billing session has no subscription item")
		}
		_, err := model.PreConsumeUserSubscription(fmt.Sprintf("bs-%d", s.TokenId), s.UserId, int64(delta))
		if err != nil {
			return fmt.Errorf("pre-consume subscription quota: %w", err)
		}
	}
	s.preConsumedQuota = targetQuota
	return nil
}

func (s *BillingSession) Settle(actualQuota int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.settled {
		return nil
	}
	s.settled = true
	delta := actualQuota - s.preConsumedQuota

	switch s.Source {
	case BillingSourceWallet:
		if err := model.PostConsumeTokenQuota(s.TokenId, int64(delta)); err != nil {
			return fmt.Errorf("post-consume token quota: %w", err)
		}
		_ = model.CacheUpdateUserQuota(nil, s.UserId)
	case BillingSourceSubscription:
		if s.SubItem == nil || s.SubItem.Id == 0 {
			return errors.New("subscription billing session has no subscription item")
		}
		if err := model.PostConsumeUserSubscriptionDelta(s.SubItem.Id, int64(delta)); err != nil {
			return fmt.Errorf("post-consume subscription quota: %w", err)
		}
	}
	return nil
}

func (s *BillingSession) Refund(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.refunded || s.preConsumedQuota == 0 {
		return
	}
	s.refunded = true
	delta := -s.preConsumedQuota

	switch s.Source {
	case BillingSourceWallet:
		go func() {
			if err := model.PostConsumeTokenQuota(s.TokenId, int64(delta)); err != nil {
				logger.SysError(fmt.Sprintf("billing refund error (token): %v", err))
			}
		}()
	case BillingSourceSubscription:
		if s.SubItem != nil {
			go func() {
				if err := model.PostConsumeUserSubscriptionDelta(s.SubItem.Id, int64(delta)); err != nil {
					logger.SysError(fmt.Sprintf("billing refund error (subscription): %v", err))
				}
			}()
		}
	}
}

func (s *BillingSession) NeedsRefund() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.refunded && !s.settled && s.preConsumedQuota > 0
}

func (s *BillingSession) GetPreConsumedQuota() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.preConsumedQuota
}

var _ relaycommon.BillingSettler = (*BillingSession)(nil)
