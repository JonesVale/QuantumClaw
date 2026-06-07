package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	billingratio "github.com/quantumclaw/quantumclaw/relay/billing/ratio"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	relaymodel "github.com/quantumclaw/quantumclaw/relay/model"
	"gorm.io/gorm"
)

// ==================== 现金计费 — 消费后扣款 ====================

// PostConsumeDeduct 消费成功后扣款 + 分账
// billingSource 来自 PreConsumeBilling 的优先级链结果：
//
//	"subscription" — 订阅已预扣，跳过现金/佣金/配额扣款
//	"cash"         — 现金未预扣，正常走现金→佣金→配额→挂账链
//	"commission"   — 佣金未预扣，跳过现金直接扣佣金→配额→挂账
//	"quota"        — 配额已预扣，跳过现金/佣金，只做配额对齐
//	""             — 兼容旧调用方（audio/image），默认走全链
func PostConsumeDeduct(ctx context.Context, meta *meta.Meta, usage *relaymodel.Usage, textRequest *relaymodel.GeneralOpenAIRequest,
	ratio float64, modelRatio float64, groupRatio float64, preConsumedQuota int64, systemPromptReset bool, billingSource string) error {

	if usage == nil {
		return fmt.Errorf("usage is nil, cannot deduct")
	}

	// ── 详细审计日志 START ──
	logger.Infof(ctx, "[BILLING_AUDIT] PostConsumeDeduct 入口 | user=%d chan=%d(orig:%d) source=%s preQuota=%d usage(prompt=%d,completion=%d) model=%s",
		meta.UserId, meta.ChannelId, meta.OriginalChannelId, billingSource, preConsumedQuota,
		usage.PromptTokens, usage.CompletionTokens, textRequest.Model)
	// ── 详细审计日志 END ──

	// 1. 读取渠道
	channel, err := model.GetChannelById(meta.ChannelId, true)
	if err != nil {
		return fmt.Errorf("get channel %d: %w", meta.ChannelId, err)
	}

	// 2. 计算实际配额（沿用现有逻辑）
	quota := calculateQuota(usage, textRequest, ratio, meta.ChannelType, meta.Config.CacheBillingRatio)
	if quota <= 0 {
		return fmt.Errorf("quota is %d, skip deduction", quota)
	}

	// 3. 配额 → 用户应付价格（分）
	priceCents := QuotaToUserPrice(quota)
	if priceCents < 1 {
		priceCents = 1
	}

	// ── 详细审计日志：价格计算 ──
	logger.Infof(ctx, "[BILLING_AUDIT] 价格计算 | user=%d chan=%d quota=%d priceCents=%d modelRatio=%.4f groupRatio=%.4f ratio=%.4f",
		meta.UserId, meta.ChannelId, quota, priceCents, modelRatio, groupRatio, ratio)
	// ── 详细审计日志 END ──

	// 3.5 成本倒挂硬拦截（扣款前检查！）
	// 如果预计上游成本高于用户付费，直接拒绝，防止平台亏损
	if err := CheckCostInversion(ctx, priceCents, quota, channel, textRequest.Model); err != nil {
		logger.Error(ctx, fmt.Sprintf("[COST_INVERSION_BLOCKED] 请求被拦截: %v", err))
		return fmt.Errorf("request blocked: %w", err)
	}

	// 4. 扣消费者余额（根据 billingSource 决定扣款策略）
	remaining := priceCents
	var deductedFromCash int64   // 实际从现金扣了多少（用于余额流水）
	initialCashBalance, _ := model.GetUserCashBalance(meta.UserId)

	// ── 详细审计日志：扣前状态 ──
	logger.Infof(ctx, "[BILLING_AUDIT] 扣款前 | user=%d 初余额=%d分 preConsumedQuota=%d quota=%d priceCents=%d billingSource=%s",
		meta.UserId, initialCashBalance, preConsumedQuota, quota, priceCents, billingSource)
	// ── 详细审计日志 END ──

	switch billingSource {
	case "subscription":
		// 订阅已预扣，不再重复扣现金/佣金/配额
		logger.Infof(ctx, "[BILLING_AUDIT] 订阅扣款跳过 | user=%d preConsumedQuota=%d", meta.UserId, preConsumedQuota)
		remaining = 0

	case "commission":
		// PreConsumeBilling 仅检查了佣金余额，未实际扣款
		// 直接扣佣金
		if remaining > 0 {
			commBalance, err := model.GetUserCommissionBalance(meta.UserId)
			if err == nil && commBalance > 0 {
				deduct := commBalance
				if deduct > remaining {
					deduct = remaining
				}
				model.MinusUserCommissionBalance(meta.UserId, deduct)
				remaining -= deduct
				logger.Infof(ctx, "[BILLING_AUDIT] 佣金扣款 | user=%d 扣=%d分 剩余=%d分", meta.UserId, deduct, remaining)
			}
		}
		// 佣金不足 → 走配额兜底
		if remaining > 0 {
			userQuota, qErr := model.CacheGetUserQuota(ctx, meta.UserId)
			if qErr == nil && userQuota > 0 {
				deduct := userQuota
				if deduct > remaining {
					deduct = remaining
				}
				model.DecreaseUserQuota(meta.UserId, deduct)
				model.PreConsumeTokenQuota(meta.TokenId, deduct)
				remaining -= deduct
				logger.Infof(ctx, "[BILLING_AUDIT] 配额扣款(commission兜底) | user=%d 扣=%d 剩余=%d分", meta.UserId, deduct, remaining)
			}
		}
		// 仍不足 → 记挂账
		if remaining > 0 {
			logger.Warnf(ctx, "[BILLING_AUDIT] 挂账产生(commission) | user=%d 挂账=%d分", meta.UserId, remaining)
			recordDebt(ctx, meta.UserId, remaining)
			remaining = 0
		}

	case "quota":
		// PreConsumeBilling 已预扣配额，此处只对齐余额
		// 如果还有剩余，记挂账
		if remaining > 0 {
			logger.Warnf(ctx, "[BILLING_AUDIT] 挂账产生(quota) | user=%d 挂账=%d分", meta.UserId, remaining)
			recordDebt(ctx, meta.UserId, remaining)
			remaining = 0
		}

	default:
		// "cash" 或 "" — 正常走全链：现金 → 佣金 → 配额 → 挂账
		// Tier 1: 现金余额
		if remaining > 0 {
			cashBalance, err := model.GetUserCashBalance(meta.UserId)
			if err == nil && cashBalance > 0 {
				deduct := cashBalance
				if deduct > remaining {
					deduct = remaining
				}
				model.MinusUserCashBalance(meta.UserId, deduct)
				deductedFromCash = deduct
				remaining -= deduct
				logger.Infof(ctx, "[BILLING_AUDIT] 现金扣款 | user=%d 扣=%d分 剩余=%d分 新余额=%d分",
					meta.UserId, deduct, remaining, cashBalance-deduct)
			} else {
				logger.Warnf(ctx, "[BILLING_AUDIT] 现金余额为0或读取失败 | user=%d err=%v", meta.UserId, err)
			}
		}

		// Tier 2: 佣金余额
		if remaining > 0 {
			commBalance, err := model.GetUserCommissionBalance(meta.UserId)
			if err == nil && commBalance > 0 {
				deduct := commBalance
				if deduct > remaining {
					deduct = remaining
				}
				model.MinusUserCommissionBalance(meta.UserId, deduct)
				remaining -= deduct
				logger.Infof(ctx, "[BILLING_AUDIT] 佣金扣款 | user=%d 扣=%d分 剩余=%d分", meta.UserId, deduct, remaining)
			}
		}

		// Tier 3: 配额回退
		if remaining > 0 {
			userQuota, qErr := model.CacheGetUserQuota(ctx, meta.UserId)
			if qErr == nil && userQuota > 0 {
				deduct := userQuota
				if deduct > remaining {
					deduct = remaining
				}
				model.DecreaseUserQuota(meta.UserId, deduct)
				model.PreConsumeTokenQuota(meta.TokenId, deduct)
				remaining -= deduct
				logger.Infof(ctx, "[BILLING_AUDIT] 配额扣款 | user=%d 扣=%d 剩余=%d分", meta.UserId, deduct, remaining)
			}
		}

		// Tier 4: 仍不足 → 记追偿挂账
		if remaining > 0 {
			logger.Warnf(ctx, "[BILLING_AUDIT] 挂账产生 | user=%d 挂账=%d分", meta.UserId, remaining)
			recordDebt(ctx, meta.UserId, remaining)
			remaining = 0
		}
	}

	// 5. 记余额流水（仅记录现金余额变动）
	actualDeducted := priceCents - remaining
	newCashBalance := initialCashBalance - deductedFromCash
	if newCashBalance < 0 {
		newCashBalance = 0
	}

	// 构建包含回退链信息的详细备注
	fallbackNote := buildFallbackRemark(meta)
	remark := fmt.Sprintf("模型:%s 配额:%d 价格:%d分 来源:%s 实扣:%d分 挂账:%d分 渠道:%d%s",
		textRequest.Model, quota, priceCents, billingSource, actualDeducted, remaining,
		meta.ChannelId, fallbackNote)
	if err := model.CreateBalanceLog(meta.UserId, model.BalanceLogTypeConsume, -deductedFromCash, newCashBalance, meta.ChannelId, remark); err != nil {
		logger.Error(ctx, fmt.Sprintf("create balance log: %v", err))
	}

	// 5.5 回退安全校验
	// 当检测到回退发生时，验证最终渠道与原始渠道是否一致，记录异常用于事后对账
	if meta.IsFallback && meta.OriginalChannelId > 0 && meta.OriginalChannelId != meta.ChannelId {
		logger.Warnf(ctx,
			"[FALLBACK_AUDIT] ⚠️ 渠道回退 detected! 原渠道:%d → 实际渠道:%d 用户:%d 模型:%s 金额:%d分 回退链:%v",
			meta.OriginalChannelId, meta.ChannelId, meta.UserId, textRequest.Model, priceCents, meta.FallbackChain)
		// 写入系统日志供财务对账使用
		model.RecordLog(ctx, 1, model.LogTypeSystem,
			fmt.Sprintf("[回退审计] 原渠道:%d→实际:%d 用户:%d 模型:%s 金额:%d分 链路:%v",
				meta.OriginalChannelId, meta.ChannelId, meta.UserId, textRequest.Model, priceCents, meta.FallbackChain))
	}

	// 6. 分账
	if channel.UserId > 0 {
		// 使用渠道级分账比例
		split := channel.ProfitSplit
		if split <= 0 || split > 1 {
			split = 0.85 // 默认渠道商 85%
		}
		commissionAmount := int64(float64(priceCents) * (1.0 - split)) // 平台抽成
		netAmount := priceCents - commissionAmount                     // 渠道商净得
		if netAmount < 0 {
			netAmount = 0
		}
		if err := model.CreateProviderEarning(
			int64(channel.UserId),
			int64(meta.ChannelId),
			int64(meta.UserId),
			priceCents,
			commissionAmount,
			netAmount,
			helper.GetCurrentMonth(),
			model.EarningStatusSettled,
		); err != nil {
			logger.Error(ctx, fmt.Sprintf("create provider earning: %v", err))
		}

		// 6.5 平台抽成自动入账
		if commissionAmount > 0 {
			if err := depositPlatformIncome(ctx, commissionAmount, meta.ChannelId, meta.UserId, priceCents, textRequest.Model); err != nil {
				logger.Error(ctx, fmt.Sprintf("platform auto-deposit failed: %v", err))
			}
		}
	} else {
		// channel.UserId == 0 → 平台自有渠道，全部收入归平台
		if priceCents > 0 {
			if err := depositPlatformIncome(ctx, priceCents, meta.ChannelId, meta.UserId, priceCents, textRequest.Model); err != nil {
				logger.Error(ctx, fmt.Sprintf("platform auto-deposit(own channel) failed: %v", err))
			}
		}
	}

	// 7. 保留日志 + 渠道用量统计
	logContent := fmt.Sprintf("倍率：%.2f × %.2f | 扣款：%d分 来源:%s", modelRatio, groupRatio, priceCents, billingSource)
	consumeLog := &model.Log{
		UserId:            meta.UserId,
		ChannelId:         meta.ChannelId,
		PromptTokens:      usage.PromptTokens,
		CompletionTokens:  usage.CompletionTokens,
		ModelName:         textRequest.Model,
		TokenName:         meta.TokenName,
		Quota:             int(quota),
		Content:           logContent,
		IsStream:          meta.IsStream,
		ElapsedTime:       helper.CalcElapsedTime(meta.StartTime),
		SystemPromptReset: systemPromptReset,
		PromoterId:        meta.PromoterId,
		ChannelOwnerId:    meta.ChannelOwnerId,
		IsFallback:        meta.IsFallback,
	}
	model.RecordConsumeLog(ctx, consumeLog)
	consumeLogId := consumeLog.Id // GORM 会自动填充主键
	model.UpdateUserUsedQuotaAndRequestCount(meta.UserId, quota)
	model.UpdateChannelUsedQuota(meta.ChannelId, quota)

	// 8. 异步触发实时对账（不阻塞主请求）
	go TriggerReconciliation(ctx, meta.UserId, meta.ChannelId, consumeLogId, priceCents,
		func() int64 {
			// 计算 platformIncomeCents
			if channel.UserId > 0 {
				split := channel.ProfitSplit
				if split <= 0 || split > 1 {
					split = 0.85
				}
				return int64(float64(priceCents) * (1.0 - split))
			}
			return priceCents
		}())
	return nil
}

// ==================== 辅助函数 ====================

// getEstimatedQuota 估算请求配额（仅 prompt 部分）
// 保留用于单元测试兼容
func getEstimatedQuota(promptTokens int, ratio float64) int64 {
	return int64(float64(config.PreConsumedQuota+int64(promptTokens)) * ratio)
}

// calculateQuota 计算实际消耗配额（沿用现有逻辑）
func calculateQuota(usage *relaymodel.Usage, textRequest *relaymodel.GeneralOpenAIRequest,
	ratio float64, channelType int, cacheBillingRatio float64) int64 {

	completionRatio := billingratio.GetCompletionRatio(textRequest.Model, channelType)
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens

	if cacheBillingRatio > 0 && usage.PromptTokensDetails != nil && usage.PromptTokensDetails.CachedTokens > 0 {
		cachedTokens := usage.PromptTokensDetails.CachedTokens
		if cachedTokens > promptTokens {
			cachedTokens = promptTokens
		}
		nonCachedTokens := promptTokens - cachedTokens
		promptTokens = nonCachedTokens + int(math.Ceil(float64(cachedTokens)*cacheBillingRatio))
		if promptTokens < 0 {
			promptTokens = 0
		}
	}

	quota := int64(math.Ceil((float64(promptTokens) + float64(completionTokens)*completionRatio) * ratio))
	if ratio != 0 && quota <= 0 {
		quota = 1
	}
	totalTokens := promptTokens + completionTokens
	if totalTokens == 0 {
		quota = 0
	}
	return quota
}

// quotaToUserPrice 将用户配额（已含 modelRatio × groupRatio）换算为应付价格（分）
// quota 单位为微额度（÷1_000_000 = USD），×100 = 分
// 这是用户应付价格，与渠道成本（CostPerUnit）无关。
// QuotaToUserPrice 将 quota（已含 modelRatio × groupRatio）换算为用户应付金额（分）
// 换算规则：quota / 1_000_000 = USD，再 × 100 = 分
// 保证最少扣 1 分钱
func QuotaToUserPrice(quota int64) int64 {
	if quota <= 0 {
		return 0
	}
	priceUsd := float64(quota) / 1000000.0
	priceCents := int64(math.Ceil(priceUsd * 100.0))
	if priceCents < 1 {
		priceCents = 1
	}
	return priceCents
}

// quotaToPrice 配额 → 渠道成本价格（分）——— 仅用于计算平台成本/分账，不可用于用户扣费
// price = quota × (CostPerUnit / 1000000) × SellPriceRate × 100（USD→分）
func quotaToPrice(quota int64, costPerUnit, sellPriceRate float64) int64 {
	if quota <= 0 {
		return 0
	}
	costUsd := float64(quota) * costPerUnit / 1000000.0
	priceUsd := costUsd * sellPriceRate
	priceCents := int64(math.Ceil(priceUsd * 100.0))
	if priceCents < 1 {
		priceCents = 1
	}
	return priceCents
}

// recordDebt 记录追偿挂账（仅记账，不封号）
// 用户余额不足时记录欠费用于追踪，不影响登录和使用
// 欠费用户在下次请求时会直接在 API 层面被拒绝，但依然可以登录和充值
func recordDebt(ctx context.Context, userId int, amount int64) {
	model.DB.Model(&model.User{}).Where("id = ?", userId).
		UpdateColumn("debt", gorm.Expr("COALESCE(debt,0) + ?", amount))
	var totalDebt int64
	model.DB.Model(&model.User{}).Where("id = ?", userId).Select("COALESCE(debt,0)").Find(&totalDebt)
	model.RecordLog(ctx, userId, model.LogTypeSystem,
		fmt.Sprintf("消费追偿挂账 %d 分，累计欠费 %d 分", amount, totalDebt))
	logger.Warnf(ctx, "user %d: debt %d added, total debt %d", userId, amount, totalDebt)
}

// ==================== 成本倒挂检测 ====================
// CheckCostInversion 实时检测用户付费是否低于上游成本
// 如果 userChargeCents < upstreamCostCents，说明每卖一单都在亏钱
// 返回 error 时，调用方应中断计费，拒绝本次请求
func CheckCostInversion(ctx context.Context, userChargeCents int64, quota int64, channel *model.Channel, modelName string) error {
	if quota <= 0 || userChargeCents <= 0 {
		return nil
	}

	// 计算上游成本（分）
	// 优先使用渠道配置的 CostPrice（渠道商填写的实际成本）
	var upstreamCostCents float64
	if channel.CostPrice > 0 {
		// 反推 token 数量（千令牌）
		// quota = token数 × modelRatio × groupRatio × 1_000_000 / 1_000
		// 近似：tokenK ≈ quota / (modelRatio × 1_000_000) × 1_000
		// 简化：直接用 quota / 1_000_000 得到 USD，再除以模型单价得到 tokenK
		_ = quota // 保留参数引用
		// CostPrice 单位是 USD/1K-tokens
		// 估算 token 数（近似值，用于成本预估）
		estimatedTokens := float64(quota) / 1000000.0 // 近似 USD 花费
		if estimatedTokens > 0 {
			upstreamCostCents = channel.CostPrice * (estimatedTokens * 1000.0) * 100.0 // 转为分
		}
	}

	// 如果没有 CostPrice，使用 CostPerUnit 作为参考
	if upstreamCostCents <= 0 && channel.CostPerUnit > 0 {
		upstreamCostCents = float64(quota) * channel.CostPerUnit / 10000.0 // 微额度→分
	}

	if upstreamCostCents <= 0 {
		return nil // 无法估算上游成本，跳过检测
	}

	margin := float64(userChargeCents) - upstreamCostCents
	marginPct := 0.0
	if userChargeCents > 0 {
		marginPct = margin / float64(userChargeCents) * 100.0
	}

	if margin < 0 {
		// 🚨 成本倒挂！用户付费 < 上游成本
		logger.Error(ctx,
			fmt.Sprintf("[COST_INVERSION] ⚠️ 模型:%s 用户付:%d分 上游成本约:%.2f分 亏损:%.2f分(%.1f%%) 渠道:%d CostPrice:%.6f CostPerUnit:%.6f",
				modelName, userChargeCents, upstreamCostCents, -margin, marginPct,
				channel.Id, channel.CostPrice, channel.CostPerUnit))

		// 记录到系统日志供后续分析
		model.RecordLog(ctx, 1, model.LogTypeSystem,
			fmt.Sprintf("[成本倒挂] 模型:%s 用户付:%d分 成本约:%.2f分 亏损:%.2f分 渠道ID:%d",
				modelName, userChargeCents, upstreamCostCents, -margin, channel.Id))

		// 硬拦截：亏损超过 5 元（500分）时，拒绝请求，防止平台继续亏损
		if -margin > 500.0 {
			logger.Error(ctx,
				fmt.Sprintf("[COST_INVERSION_BLOCKED] 拒绝请求！模型:%s 预计亏损:%.2f分 用户:%d 渠道:%d",
					modelName, -margin, ctx.Value("user_id"), channel.Id))
			return fmt.Errorf("upstream cost (%.2f) exceeds user charge (%d), request blocked to prevent loss", upstreamCostCents, userChargeCents)
		}
	} else if marginPct < 5.0 {
		// 利润率过低 (<5%)，记录 warning 但不阻断
		logger.Warnf(ctx,
			"[LOW_MARGIN] 模型:%s 用户付:%d分 成本约:%.2f分 利润仅%.2f分(%.1f%%) 渠道:%d",
			modelName, userChargeCents, upstreamCostCents, margin, marginPct, channel.Id)
	}
	return nil
}

// ==================== 平台自动入账 ====================
// depositPlatformIncome 将平台分账收益自动记入平台收入账本
// 每次 PostConsumeDeduct 分账完成后调用
//
// 当前实现：
//   1. 写入 platform_income 表（审计流水）
//   2. 累加 platform_config 中的 platform_balance 字段
//
// TODO 后续可对接真实支付系统（支付宝/微信/银行）实现提现
func depositPlatformIncome(ctx context.Context, commissionAmount int64, channelId int, userId int, priceCents int64, modelName string) error {
	if commissionAmount <= 0 {
		return nil
	}

	// 写入平台收入流水
	err := model.CreatePlatformIncomeRecord(&model.PlatformIncome{
		ChannelId:        channelId,
		ConsumerUserId:   userId,
		TotalAmount:      priceCents,
		CommissionAmount: commissionAmount,
		Source:           "relay_billing",
		ModelName:        modelName,
		Status:           "settled",
		CreatedAt:        helper.GetTimestamp(),
	})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("create platform income record: %v", err))
		return err
	}

	// 累加平台余额（内存+DB 双写）
	model.AddPlatformBalance(commissionAmount)

	logger.Info(ctx, fmt.Sprintf("[平台入账] +%d分 | 来源:渠道%d 用户%d 模型%s",
		commissionAmount, channelId, userId, modelName))
	return nil
}

// PostConsumeQuantumDeduct 量子任务扣款
// 使用与 PostConsumeDeduct 一致的三级扣款链：现金→佣金→配额
func PostConsumeQuantumDeduct(userId, channelId int, costQuota int64) error {
	if costQuota <= 0 {
		return nil
	}
	// 读取渠道获取 CostPerUnit + SellPriceRate
	channel, err := model.GetChannelById(channelId, true)
	if err != nil {
		return fmt.Errorf("get channel %d: %w", channelId, err)
	}
	// 配额 → 用户应付价格（分）——— BUG FIX: 原代码使用 channel.CostPerUnit
	priceCents := QuotaToUserPrice(costQuota)

	remaining := priceCents

	// Tier 1: 现金余额
	if remaining > 0 {
		cashBalance, err := model.GetUserCashBalance(userId)
		if err == nil && cashBalance > 0 {
			deduct := cashBalance
			if deduct > remaining {
				deduct = remaining
			}
			if err := model.MinusUserCashBalance(userId, deduct); err != nil {
				return fmt.Errorf("deduct user %d cash: %w", userId, err)
			}
			remaining -= deduct
		}
	}

	// Tier 2: 佣金余额
	if remaining > 0 {
		commBalance, err := model.GetUserCommissionBalance(userId)
		if err == nil && commBalance > 0 {
			deduct := commBalance
			if deduct > remaining {
				deduct = remaining
			}
			if err := model.MinusUserCommissionBalance(userId, deduct); err != nil {
				return fmt.Errorf("deduct user %d commission: %w", userId, err)
			}
			remaining -= deduct
		}
	}

	// Tier 3: 不足 → 返回错误（量子任务不支持挂账）
	if remaining > 0 {
		return fmt.Errorf("user %d balance + commission %d < price %d", userId, priceCents-remaining, priceCents)
	}

	// 余额流水
	actualDeducted := priceCents - remaining
	_ = model.CreateBalanceLog(userId, model.BalanceLogTypeConsume, -actualDeducted, remaining, channelId, "quantum task")

	// 分账
	if channel.UserId > 0 {
		netAmount := priceCents
		if netAmount < 0 {
			netAmount = 0
		}
		_ = model.CreateProviderEarning(int64(channel.UserId), int64(channelId), int64(userId),
			priceCents, 0, netAmount, helper.GetCurrentMonth(), model.EarningStatusSettled)
	}
	return nil
}

// ==================== 回退链审计辅助函数 ====================

// buildFallbackRemark 构建回退信息备注字符串，用于余额流水和日志
// 如果没有发生回退，返回空字符串
func buildFallbackRemark(meta *meta.Meta) string {
	if !meta.IsFallback {
		return ""
	}
	var b strings.Builder
	b.WriteString(" [FALLBACK]")
	if meta.OriginalChannelId > 0 {
		b.WriteString(fmt.Sprintf(" orig_ch:%d", meta.OriginalChannelId))
	}
	if len(meta.FallbackChain) > 0 {
		b.WriteString(" chain:")
		for _, step := range meta.FallbackChain {
			b.WriteString(fmt.Sprintf("%d→%d,", step.FromSchema, step.ToSchema))
		}
	}
	if meta.ActualSchemaId > 0 {
		b.WriteString(fmt.Sprintf(" schema:%d", meta.ActualSchemaId))
	}
	return b.String()
}
