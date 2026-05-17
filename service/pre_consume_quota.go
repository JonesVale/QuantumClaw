package service

import (
	"fmt"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

func PreConsumeQuota(tokenId int, userId int, quota int64) error {
	if quota <= 0 {
		return nil
	}
	err := model.PreConsumeTokenQuota(tokenId, quota)
	if err != nil {
		return fmt.Errorf("pre-consume token quota: %w", err)
	}
	err = model.DecreaseUserQuota(userId, quota)
	if err != nil {
		logger.SysError(fmt.Sprintf("decrease user quota failed: %v", err))
	}
	return nil
}

func PostConsumeQuota(tokenId int, userId int, quotaDelta int64) error {
	return model.PostConsumeTokenQuota(tokenId, quotaDelta)
}

func ReturnPreConsumedQuota(tokenId int, quota int64) {
	if quota <= 0 {
		return
	}
	go func() {
		if err := model.PostConsumeTokenQuota(tokenId, -quota); err != nil {
			logger.SysError("failed to return pre-consumed quota: " + err.Error())
		}
	}()
}
