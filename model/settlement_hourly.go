package model

import (
	"fmt"
	"strconv"
	"time"

	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HourlySettlement 每小时结算对账汇总
type HourlySettlement struct {
	Id              int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	Hour            string  `json:"hour" gorm:"type:varchar(16);uniqueIndex;not null"` // 格式: 2026-05-28 15:00
	TotalRequests   int     `json:"total_requests" gorm:"default:0"`
	TotalTokens     int64   `json:"total_tokens" gorm:"default:0"`
	UserRevenue     float64 `json:"user_revenue" gorm:"type:decimal(16,8);default:0"`     // 用户付费总额
	UpstreamCost    float64 `json:"upstream_cost" gorm:"type:decimal(16,8);default:0"`     // 上游成本
	CommissionPaid  float64 `json:"commission_paid" gorm:"type:decimal(16,8);default:0"`   // 已发佣金
	PlatformFee     float64 `json:"platform_fee" gorm:"type:decimal(16,8);default:0"`      // 平台费
	GrossProfit     float64 `json:"gross_profit" gorm:"type:decimal(16,8);default:0"`      // 毛利
	ProfitMargin    float64 `json:"profit_margin" gorm:"type:decimal(8,4);default:0"`      // 利润率 %
	LossAmount      float64 `json:"loss_amount" gorm:"type:decimal(16,8);default:0"`       // 亏损总额（正=亏损）
	DebtCollected   float64 `json:"debt_collected" gorm:"type:decimal(16,8);default:0"`    // 本轮已追回债务
	AffectedUsers   int     `json:"affected_users" gorm:"default:0"`                       // 被追偿用户数
	Status          string  `json:"status" gorm:"type:varchar(20);default:'pending'"`      // pending / settled / skipped
	CreatedAt       int64   `json:"created_at" gorm:"bigint"`
	SettledAt       int64   `json:"settled_at" gorm:"bigint;default:0"`
}

func (HourlySettlement) TableName() string {
	return "settlement_hourly"
}

// AggregateHourlySettlement 聚合一小时 token_transaction 生成对账记录
func AggregateHourlySettlement(hour string) (*HourlySettlement, error) {
	startTime := parseHourStart(hour)
	endTime := startTime + 3600

	var result struct {
		ReqCount   int
		TotalTok   int64
		UserRev    float64
		UpCost     float64
		CommPaid   float64
		PlatFee    float64
	}
	err := DB.Raw(`
		SELECT
			COUNT(*)                               AS req_count,
			COALESCE(SUM(prompt_tokens + completion_tokens), 0) AS total_tok,
			COALESCE(SUM(total_amount), 0)          AS user_rev,
			COALESCE(SUM(key_provider_cost), 0)     AS up_cost,
			COALESCE(SUM(commission_amount), 0)     AS comm_paid,
			COALESCE(SUM(platform_fee), 0)          AS plat_fee
		FROM token_transaction
		WHERE created_time >= ? AND created_time < ?
	`, startTime, endTime).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("aggregate hourly: %w", err)
	}

	cost := result.UpCost + result.CommPaid
	profit := result.UserRev - cost
	margin := 0.0
	if result.UserRev > 0 {
		margin = profit / result.UserRev * 100.0
	}
	loss := 0.0
	if profit < 0 {
		loss = -profit
	}

	hs := &HourlySettlement{
		Hour:           hour,
		TotalRequests:  result.ReqCount,
		TotalTokens:    result.TotalTok,
		UserRevenue:    result.UserRev,
		UpstreamCost:   result.UpCost,
		CommissionPaid: result.CommPaid,
		PlatformFee:    result.PlatFee,
		GrossProfit:    profit,
		ProfitMargin:   margin,
		LossAmount:     loss,
		Status:         "pending",
		CreatedAt:      helper.GetTimestamp(),
	}
	return hs, nil
}

// CalculateAndRecoverDebt 算亏损 + 追偿 debt
// 按 user_id 分组聚合亏损，从用户余额扣除
func CalculateAndRecoverDebt(hour string) (totalRecovered float64, affected int, err error) {
	startTime := parseHourStart(hour)
	endTime := startTime + 3600

	type UserLoss struct {
		UserId   int
		LossAmt  float64
	}
	var losses []UserLoss

	// 按用户聚合: 用户付了多少 vs 实际成本
	// 直接计算 loss_amt = 实际成本 - 用户付费
	err = DB.Raw(`
		SELECT
			user_id,
			SUM(key_provider_cost + commission_amount) - SUM(total_amount) AS loss_amt
		FROM token_transaction
		WHERE created_time >= ? AND created_time < ?
		GROUP BY user_id
		HAVING loss_amt > 0.001
	`, startTime, endTime).Scan(&losses).Error
	if err != nil {
		return 0, 0, fmt.Errorf("aggregate user loss: %w", err)
	}

	affected = len(losses)
	if affected == 0 {
		return 0, 0, nil
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		for _, loss := range losses {
			lossAmount := loss.LossAmt
			if lossAmount <= 0.001 {
				continue
			}

			// 转分为单位 (1元 = 100分)
			lossCents := int64(lossAmount * 100.0)
			if lossCents < 1 {
				continue
			}

			// 用悲观锁锁定用户行，避免并发扣款竞态
			var user User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", loss.UserId).First(&user).Error; err != nil {
				logger.Warn(nil, fmt.Sprintf("debt recovery: user %d not found, skip", loss.UserId))
				continue
			}

			// 先扣 cash_balance
			deductFromBalance := lossCents
			if user.CashBalance < deductFromBalance {
				deductFromBalance = user.CashBalance
			}
			if deductFromBalance > 0 {
				if err := tx.Model(&User{}).Where("id = ?", loss.UserId).
					Update("cash_balance", gorm.Expr("cash_balance - ?", deductFromBalance)).Error; err != nil {
					return fmt.Errorf("deduct user %d balance: %w", loss.UserId, err)
				}
				_ = CreateBalanceLogTx(tx, loss.UserId, BalanceLogTypeDebtRecover, -deductFromBalance,
					user.CashBalance-deductFromBalance, 0, "对账追偿")
			}

			// 剩余挂 debt（已扣余额后的剩余亏损）
			remaining := lossCents - deductFromBalance
			if remaining > 0 {
				if err := tx.Model(&User{}).Where("id = ?", loss.UserId).
					Update("debt", gorm.Expr("COALESCE(debt,0) + ?", remaining)).Error; err != nil {
					return fmt.Errorf("add debt for user %d: %w", loss.UserId, err)
				}
			}

			totalRecovered += float64(lossCents) / 100.0
		}
		return nil
	})

	return totalRecovered, affected, err
}

// RunHourlySettlement 执行每小时对账
func RunHourlySettlement() {
	now := time.Now()
	prevHour := now.Truncate(time.Hour).Add(-time.Hour)
	hour := prevHour.Format("2006-01-02 15:00")

	logger.Info(nil, fmt.Sprintf("[HourlySettlement] 开始对账: %s", hour))

	// 1. 聚合（无事务，只读）
	hs, err := AggregateHourlySettlement(hour)
	if err != nil {
		logger.Error(nil, fmt.Sprintf("[HourlySettlement] aggregate failed: %v", err))
		return
	}

	// 2. 有亏损 → 追偿（在自身事务内原子执行：FOR UPDATE 锁定用户+debt调整）
	if hs.LossAmount > 0.01 {
		recovered, affected, err := CalculateAndRecoverDebt(hour)
		if err != nil {
			logger.Error(nil, fmt.Sprintf("[HourlySettlement] debt recovery failed: %v", err))
		} else {
			hs.DebtCollected = recovered
			hs.AffectedUsers = affected
		}
	}

	// 3. 写对账记录（事务内 check-then-create 消除竞态）
	err = DB.Transaction(func(tx *gorm.DB) error {
		var existing HourlySettlement
		if err := tx.Where("hour = ?", hour).First(&existing).Error; err == nil {
			return fmt.Errorf("already settled: %s", hour)
		}

		if hs.TotalRequests == 0 {
			hs.Status = "skipped"
		} else {
			hs.Status = "settled"
			hs.SettledAt = helper.GetTimestamp()
		}

		if err := tx.Create(hs).Error; err != nil {
			return fmt.Errorf("save: %w", err)
		}
		return nil
	})

	if err != nil {
		logger.Error(nil, fmt.Sprintf("[HourlySettlement] 对账写入失败 %s: %v", hour, err))
		return
	}

	logger.Info(nil, fmt.Sprintf("[HourlySettlement] 对账完成 %s | 请求=%d 收入=%.2f 成本=%.2f 利润=%.2f(%.1f%%) 亏损=%.2f 追回=%.2f",
		hour, hs.TotalRequests, hs.UserRevenue, hs.UpstreamCost+hs.CommissionPaid,
		hs.GrossProfit, hs.ProfitMargin, hs.LossAmount, hs.DebtCollected))

	// 4. 检查长期欠费用户并自动封号
	suspended, suspendErr := CheckSuspendedDebtUsers()
	if suspendErr != nil {
		logger.Error(nil, fmt.Sprintf("[HourlySettlement] debt suspend check failed: %v", suspendErr))
	} else if suspended > 0 {
		logger.Info(nil, fmt.Sprintf("[HourlySettlement] 长期欠费封号: %d 个用户被禁用", suspended))
	}
}

// GetHourlySettlements 获取对账记录（分页）
func GetHourlySettlements(page, pageSize int) ([]HourlySettlement, int64, error) {
	var list []HourlySettlement
	var total int64
	DB.Model(&HourlySettlement{}).Count(&total)
	err := DB.Order("hour desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// parseHourStart 将 "2026-05-28 15:00" 转为时间戳
// 解析失败时 panic（因上层有 recover 保护，且格式保证来自代码）
func parseHourStart(hour string) int64 {
	t, err := time.Parse("2006-01-02 15:00", hour)
	if err != nil {
		panic(fmt.Sprintf("parseHourStart: 无法解析小时格式 %q: %v", hour, err))
	}
	return t.Unix()
}

// CheckSuspendedDebtUsers 检查长期欠费用户并自动禁用账号
// 欠费超过阈值(默认10000分=$100)且首次欠费时间超过30天的用户，自动禁用
// 返回被禁用的用户数和错误信息
func CheckSuspendedDebtUsers() (suspended int, err error) {
	threshold := int64(10000) // 默认 $100 阈值
	days := int64(30)       // 默认 30 天

	// 从 platform_config 读取配置
	var cfg PlatformConfig
	if DB.Where("`key` = ?", "debt_suspend_threshold_cents").First(&cfg).Error == nil {
		if v, e := strconv.ParseInt(cfg.Value, 10, 64); e == nil && v > 0 {
			threshold = v
		}
	}
	if DB.Where("`key` = ?", "debt_suspend_days").First(&cfg).Error == nil {
		if v, e := strconv.ParseInt(cfg.Value, 10, 64); e == nil && v > 0 {
			days = v
		}
	}

	cutoff := time.Now().Unix() - days*86400

	type debtUser struct {
		Id        int
		Debt      int64
		DebtSince int64
	}
	var users []debtUser

	// 查找欠费超过阈值且首次欠费超过N天的用户
	err = DB.Table("user").
		Where("debt >= ? AND debt_since > 0 AND debt_since <= ?", threshold, cutoff).
		Select("id, debt, debt_since").
		Find(&users).Error
	if err != nil {
		return 0, fmt.Errorf("query debt users: %w", err)
	}

	if len(users) == 0 {
		return 0, nil
	}

	suspended = len(users)
	err = DB.Transaction(func(tx *gorm.DB) error {
		for _, u := range users {
			if err := tx.Model(&User{}).Where("id = ?", u.Id).
				Update("status", UserStatusDisabled).Error; err != nil {
				logger.Warn(nil, fmt.Sprintf("debt suspend: failed to disable user %d: %v", u.Id, err))
				continue
			}
			RecordLog(nil, int(u.Id), LogTypeSystem,
					fmt.Sprintf("长期欠费自动封号：欠费%d分，首次欠费时间%d(%s)",
						u.Debt, u.DebtSince, time.Unix(u.DebtSince, 0).Format("2006-01-02")))
			logger.Info(nil, fmt.Sprintf("[DebtSuspend] user %d suspended, debt=%d cents, since=%s",
					u.Id, u.Debt, time.Unix(u.DebtSince, 0).Format("2006-01-02")))
		}
		return nil
	})

	return suspended, err
}
