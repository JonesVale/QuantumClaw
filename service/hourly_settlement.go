package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quantumclaw/quantumclaw/model"
)

// StartHourlySettlement 启动每小时对账定时任务
func StartHourlySettlement() {
	// 每小时整点执行
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// 启动后先对上一小时
	model.RunHourlySettlement()

	for range ticker.C {
		model.RunHourlySettlement()
	}
}

// DeductDebtOnTopup 充值成功后自动扣除挂账债务
// 在 CompleteTopUp 中调用
func DeductDebtOnTopup(userId int) (deducted int64, err error) {
	user, err := model.GetUserById(userId, false)
	if err != nil {
		return 0, fmt.Errorf("get user: %w", err)
	}
	if user.Debt <= 0 {
		return 0, nil
	}
	debt := user.Debt
	balance := user.CashBalance

	deduct := debt
	if balance < deduct {
		deduct = balance
	}
	if deduct <= 0 {
		return 0, nil
	}

	err = model.DB.Model(&model.User{}).Where("id = ?", userId).
		Updates(map[string]interface{}{
			"cash_balance": gorm.Expr("cash_balance - ?", deduct),
			"debt":         gorm.Expr("debt - ?", deduct),
		}).Error
	if err != nil {
		return 0, fmt.Errorf("deduct debt: %w", err)
	}

	_ = model.CreateBalanceLog(userId, "debt_deduct", -deduct, balance-deduct, 0, "充值自动抵扣欠费")
	return deduct, nil
}
