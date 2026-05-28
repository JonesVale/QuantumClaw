package model

import (
	"context"
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TopUp 充值订单模型 - 安全增强版
type TopUp struct {
	Id              int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId          int     `json:"user_id" gorm:"index;not null"`
	Amount          int64   `json:"amount" gorm:"not null"`           // 充值额度（配额单位）
	Money           float64 `json:"money" gorm:"type:decimal(10,6);not null"` // 支付金额
	TradeNo         string  `json:"trade_no" gorm:"uniqueIndex;type:varchar(128);not null"` // 订单号
	PaymentMethod   string  `json:"payment_method" gorm:"type:varchar(50)"`
	PaymentProvider string  `json:"payment_provider" gorm:"type:varchar(50);default:'';index"`
	Status          string  `json:"status" gorm:"type:varchar(32);index;not null;default:'pending'"`
	CreateTime      int64   `json:"create_time" gorm:"not null"`
	CompleteTime    int64   `json:"complete_time" gorm:"default:0"`
	ExpireTime      int64   `json:"expire_time" gorm:"default:0"` // 订单过期时间
	UserIP          string  `json:"user_ip" gorm:"type:varchar(45)"` // 记录用户IP，用于风控
	UserAgent      string  `json:"user_agent" gorm:"type:text"`    // 记录User-Agent，用于风控
	NotifyCount     int     `json:"notify_count" gorm:"default:0"` // 通知次数，防止重复通知
	LastNotifyTime  int64   `json:"last_notify_time" gorm:"default:0"`
	ProviderData    string  `json:"provider_data" gorm:"type:text"` // 支付提供商返回的数据（加密存储）
	Signature       string  `json:"-" gorm:"type:varchar(256)"`    // 订单签名，防止篡改
}

// TopUpStatus 订单状态
const (
	TopUpStatusPending   = "pending"   // 待支付
	TopUpStatusSuccess   = "success"   // 支付成功
	TopUpStatusFailed    = "failed"    // 支付失败
	TopUpStatusExpired   = "expired"   // 已过期
	TopUpStatusCancelled = "cancelled" // 已取消
)

// PaymentMethod 支付方式
const (
	PaymentMethodEpay         = "epay"
	PaymentMethodStripe       = "stripe"
	PaymentMethodCreem        = "creem"
	PaymentMethodWaffo        = "waffo"
	PaymentMethodWaffoPancake = "waffo_pancake"
	PaymentMethodBinance      = "binance"
	PaymentMethodAlipay       = "alipay"
	PaymentMethodWorldFirst   = "worldfirst"
)

// PaymentProvider 支付提供商
const (
	PaymentProviderEpay         = "epay"
	PaymentProviderStripe       = "stripe"
	PaymentProviderCreem        = "creem"
	PaymentProviderWaffo        = "waffo"
	PaymentProviderWaffoPancake = "waffo_pancake"
	PaymentProviderBinance      = "binance"
	PaymentProviderAlipay       = "alipay"
	PaymentProviderWorldFirst   = "worldfirst"
)

// 错误信息
var (
	ErrTopUpNotFound          = errors.New("充值订单不存在")
	ErrTopUpStatusInvalid     = errors.New("充值订单状态无效")
	ErrPaymentMethodMismatch  = errors.New("支付方式不匹配")
	ErrTopUpAmountInvalid     = errors.New("充值金额无效")
	ErrTopUpSignatureInvalid  = errors.New("订单签名验证失败")
	ErrTopUpReplayAttack     = errors.New("检测到重放攻击")
	ErrTopUpExpired          = errors.New("订单已过期")
)

// GenerateSecureTradeNo 生成密码学安全的订单号
// 格式：QC + 时间戳(10位) + 随机数(8位十六进制) + 用户ID哈希(4位)
func GenerateSecureTradeNo(userId int) (string, error) {
	// 时间戳（秒级）
	timestamp := time.Now().Unix()
	
	// 生成8字节随机数
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		logger.Error(context.Background(), fmt.Sprintf("生成订单号失败: 随机数生成失败 error=%q", err.Error()))
		return "", fmt.Errorf("生成订单号失败: %w", err)
	}
	randomHex := hex.EncodeToString(randomBytes) // 16位十六进制字符串
	
	// 用户ID的简单哈希（用于防重放）
	userIdHash := fmt.Sprintf("%04x", userId%65536)
	
	// 组合订单号
	tradeNo := fmt.Sprintf("QC%d%s%s", timestamp, randomHex, userIdHash)
	
	return tradeNo, nil
}

// CalculateSignature 计算订单签名（防止篡改）
func (t *TopUp) CalculateSignature(secret string) string {
	// 签名数据：订单号 + 用户ID + 金额 + 时间戳
	data := fmt.Sprintf("%s:%d:%.6f:%d", t.TradeNo, t.UserId, t.Money, t.CreateTime)
	signature := common.HmacSha256(data, secret)
	return signature
}

// VerifySignature 验证订单签名
func (t *TopUp) VerifySignature(secret string) bool {
	expectedSignature := t.CalculateSignature(secret)
	return expectedSignature == t.Signature
}

// BeforeCreate GORM 钩子：创建前自动生成订单号和签名
func (t *TopUp) BeforeCreate(tx *gorm.DB) error {
	// 生成订单号（如果未设置）
	if t.TradeNo == "" {
		tradeNo, err := GenerateSecureTradeNo(t.UserId)
		if err != nil {
			return err
		}
		t.TradeNo = tradeNo
	}
	
	// 设置创建时间
	if t.CreateTime == 0 {
		t.CreateTime = common.GetTimestamp()
	}
	
	// 设置过期时间（默认30分钟）
	if t.ExpireTime == 0 {
		t.ExpireTime = t.CreateTime + 1800 // 30分钟
	}
	
	// 设置默认状态
	if t.Status == "" {
		t.Status = TopUpStatusPending
	}
	
	// 生成签名（使用服务器配置的签名密钥，未配置则不签名）
	if sigSecret := os.Getenv("PAYMENT_SIGNATURE_SECRET"); sigSecret != "" {
		t.Signature = t.CalculateSignature(sigSecret)
	}
	
	return nil
}

// IsExpired 检查订单是否过期
func (t *TopUp) IsExpired() bool {
	if t.ExpireTime == 0 {
		return false
	}
	now := common.GetTimestamp()
	return now > t.ExpireTime
}

// CanProcess 检查订单是否可以处理（防止重放攻击和重复处理）
func (t *TopUp) CanProcess() error {
	// 检查订单是否存在
	if t.Id == 0 {
		return ErrTopUpNotFound
	}
	
	// 检查订单状态
	if t.Status != TopUpStatusPending {
		return fmt.Errorf("%w: 当前状态=%s", ErrTopUpStatusInvalid, t.Status)
	}
	
	// 检查订单是否过期
	if t.IsExpired() {
		// 自动更新为过期状态
		t.Status = TopUpStatusExpired
		DB.Model(&TopUp{}).Where("id = ?", t.Id).Update("status", TopUpStatusExpired)
		return ErrTopUpExpired
	}
	
	// 验证签名
	if common.GetEnvOrDefault("PAYMENT_SIGNATURE_SECRET", "") != "" {
		if !t.VerifySignature(common.GetEnvOrDefault("PAYMENT_SIGNATURE_SECRET", "")) {
			logger.Warn(context.Background(), fmt.Sprintf("订单签名验证失败 trade_no=%s", t.TradeNo))
			return ErrTopUpSignatureInvalid
		}
	}
	
	return nil
}

// Insert 插入新订单（带事务保护）
func (t *TopUp) Insert() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 检查是否存在相同的订单号（防止重复）
		var count int64
		if err := tx.Model(&TopUp{}).Where("trade_no = ?", t.TradeNo).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("订单号已存在")
		}
		
		// 插入订单
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		
		logger.Info(context.Background(), fmt.Sprintf("充值订单创建成功 trade_no=%s user_id=%d amount=%d money=%.2f", 
			t.TradeNo, t.UserId, t.Amount, t.Money))
		
		return nil
	})
}

// Update 更新订单（带事务保护）
func (t *TopUp) Update() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 使用悲观锁锁定订单
		var topUp TopUp
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", t.Id).First(&topUp).Error; err != nil {
			return err
		}
		
		// 验证签名
		if common.GetEnvOrDefault("PAYMENT_SIGNATURE_SECRET", "") != "" {
			if !t.VerifySignature(common.GetEnvOrDefault("PAYMENT_SIGNATURE_SECRET", "")) {
				return ErrTopUpSignatureInvalid
			}
		}
		
		// 更新订单
		return tx.Save(t).Error
	})
}

// GetTopUpById 根据ID获取订单
func GetTopUpById(id int64) (*TopUp, error) {
	var topUp TopUp
	err := DB.Where("id = ?", id).First(&topUp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTopUpNotFound
		}
		return nil, err
	}
	return &topUp, nil
}

// GetTopUpByTradeNo 根据订单号获取订单
func GetTopUpByTradeNo(tradeNo string) (*TopUp, error) {
	if tradeNo == "" {
		return nil, errors.New("订单号不能为空")
	}
	
	var topUp TopUp
	err := DB.Where("trade_no = ?", tradeNo).First(&topUp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTopUpNotFound
		}
		return nil, err
	}
	return &topUp, nil
}

// UpdateTopUpStatus 安全地更新订单状态（带事务和锁）
func UpdateTopUpStatus(tradeNo string, expectedProvider string, newStatus string) error {
	if tradeNo == "" {
		return errors.New("订单号不能为空")
	}
	
	return DB.Transaction(func(tx *gorm.DB) error {
		// 使用悲观锁锁定订单
		var topUp TopUp
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("trade_no = ?", tradeNo).First(&topUp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTopUpNotFound
			}
			return err
		}
		
		// 验证支付提供商（防止混淆攻击）
		if expectedProvider != "" && topUp.PaymentProvider != expectedProvider {
			logger.Warn(context.Background(), fmt.Sprintf("支付提供商不匹配 trade_no=%s expected=%s actual=%s", 
				tradeNo, expectedProvider, topUp.PaymentProvider))
			return ErrPaymentMethodMismatch
		}
		
		// 验证订单状态（防止重复处理）
		if topUp.Status != TopUpStatusPending {
			logger.Warn(context.Background(), fmt.Sprintf("订单状态无效 trade_no=%s current_status=%s", 
				tradeNo, topUp.Status))
			return ErrTopUpStatusInvalid
		}
		
		// 检查订单是否过期
		if topUp.IsExpired() {
			// 更新为过期状态
			tx.Model(&TopUp{}).Where("id = ?", topUp.Id).Update("status", TopUpStatusExpired)
			return ErrTopUpExpired
		}
		
		// 更新状态
		now := common.GetTimestamp()
		updates := map[string]interface{}{
			"status":        newStatus,
			"complete_time": now,
		}
		
		if err := tx.Model(&TopUp{}).Where("id = ?", topUp.Id).Updates(updates).Error; err != nil {
			return err
		}
		
		logger.Info(context.Background(), fmt.Sprintf("订单状态更新成功 trade_no=%s old_status=%s new_status=%s", 
			tradeNo, topUp.Status, newStatus))
		
		return nil
	})
}

// CompleteTopUp 完成充值（带事务和配额更新）
func CompleteTopUp(tradeNo string, provider string, quota int64) error {
	if tradeNo == "" {
		return errors.New("订单号不能为空")
	}
	
	var topUpUserId int
	var topUpAmount float64
	var beforeQuota int64
	
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 使用悲观锁锁定订单
		var topUp TopUp
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("trade_no = ?", tradeNo).First(&topUp).Error; err != nil {
			return err
		}
		
		// 验证订单状态
		if topUp.Status != TopUpStatusPending {
			return fmt.Errorf("%w: 当前状态=%s", ErrTopUpStatusInvalid, topUp.Status)
		}
		
		// 验证支付提供商
		if provider != "" && topUp.PaymentProvider != provider {
			return ErrPaymentMethodMismatch
		}
		
		// 获取当前用户配额（before）
		var user User
		if err := tx.Where("id = ?", topUp.UserId).First(&user).Error; err != nil {
			return err
		}
		beforeQuota = user.Quota
		
		// 更新订单状态
		now := common.GetTimestamp()
		if err := tx.Model(&TopUp{}).Where("id = ?", topUp.Id).Updates(map[string]interface{}{
			"status":         TopUpStatusSuccess,
			"complete_time":  now,
		}).Error; err != nil {
			return err
		}
		
		// 更新用户配额
		if err := tx.Model(&User{}).Where("id = ?", topUp.UserId).Update("quota", gorm.Expr("quota + ?", quota)).Error; err != nil {
			return err
		}

		// 同步更新现金余额：用悲观锁锁定用户行，原子处理手续费+债务扣款
		var userForBalance User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", topUp.UserId).First(&userForBalance).Error; err != nil {
			return err
		}

		cashAmount := int64(math.Ceil(topUp.Money * 100.0))
		if cashAmount > 0 {
			feePctDomestic, feePctForeign, feeMinUsd := GetTransactionFeeRate()
			var fee int64
			switch topUp.PaymentProvider {
			case PaymentProviderEpay, PaymentProviderAlipay:
				fee = int64(math.Ceil(float64(cashAmount) * feePctDomestic / 100.0))
			case PaymentProviderStripe, PaymentProviderCreem,
				PaymentProviderWaffo, PaymentProviderWaffoPancake,
				PaymentProviderBinance, PaymentProviderWorldFirst:
				fee = int64(math.Ceil(float64(cashAmount) * feePctForeign / 100.0))
				minFee := int64(math.Ceil(feeMinUsd * 100.0))
				if fee < minFee {
					fee = minFee
				}
			}
			if fee > cashAmount {
				fee = cashAmount
			}

			// 事务内原子处理：加钱 → 扣手续费 → 自动还债
			// userForBalance 已在事务内加锁读取，基于它的 Debt 计算
			netCashChange := cashAmount - fee
			debtDeduct := int64(0)
			if userForBalance.Debt > 0 {
				available := userForBalance.CashBalance + netCashChange
				if available > 0 {
					debtDeduct = userForBalance.Debt
					if debtDeduct > available {
						debtDeduct = available
					}
					// netCashChange 扣掉债务部分：新余额 = old余额 + cashAmount - fee - debtDeduct
					netCashChange = netCashChange - debtDeduct
				}
			}

			updates := map[string]interface{}{
				"cash_balance": gorm.Expr("cash_balance + ?", netCashChange),
			}
			if debtDeduct > 0 {
				newDebt := userForBalance.Debt - debtDeduct
				if newDebt < 0 {
					newDebt = 0
				}
				updates["debt"] = newDebt
			}
			if err := tx.Model(&User{}).Where("id = ?", topUp.UserId).
				Updates(updates).Error; err != nil {
				return err
			}

			// 写余额变更审计日志（事务内，用 tx 避免 SQLITE_BUSY）
			finalBalance := userForBalance.CashBalance + netCashChange
			_ = CreateBalanceLogTx(tx, topUp.UserId, BalanceLogTypeTopUp, cashAmount-fee,
				finalBalance, 0, fmt.Sprintf("充值入账 provider=%s", topUp.PaymentProvider))
			if debtDeduct > 0 {
				_ = CreateBalanceLogTx(tx, topUp.UserId, BalanceLogTypeDebtDeduct, -debtDeduct,
					finalBalance, 0, "充值自动抵扣欠费")
			}
		}
		
		// 创建交易审计日志
		transactionLog := TransactionLog{
			UserId:      topUp.UserId,
			Action:      "topup",
			Amount:      quota,
			BeforeQuota: beforeQuota,
			AfterQuota:  beforeQuota + quota,
			IP:          topUp.UserIP,
			UserAgent:   topUp.UserAgent,
			TradeNo:     tradeNo,
			Status:      "completed",
			AdminId:     0,
			Remark:      fmt.Sprintf("充值完成，提供商: %s，订单金额: %.2f", provider, topUp.Money),
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(&transactionLog).Error; err != nil {
			logger.Warn(context.Background(), fmt.Sprintf("创建交易审计日志失败 trade_no=%s error=%q", tradeNo, err.Error()))
			// 不阻止交易完成
		}
		
		topUpUserId = topUp.UserId
		topUpAmount = topUp.Money
		
		logger.Info(context.Background(), fmt.Sprintf("充值完成 trade_no=%s user_id=%d quota=%d", 
			tradeNo, topUp.UserId, quota))
		
		return nil
	})
	
	if err == nil && topUpUserId > 0 {
		// 创建应用内通知
		notifData := fmt.Sprintf(`{"trade_no":"%s","quota":%d}`, tradeNo, quota)
		if createErr := CreateNotification(topUpUserId, "topup", "充值成功", fmt.Sprintf("您的账户已成功充值 %s", common.LogQuota(quota)), notifData); createErr != nil {
			logger.Warn(context.Background(), fmt.Sprintf("创建充值通知失败 user_id=%d error=%q", topUpUserId, createErr.Error()))
		}
		
		// 异步发送邮件通知
		go sendTopUpEmailNotification(topUpUserId, tradeNo, quota, topUpAmount)
	}
	
	return err
}

// sendTopUpEmailNotification 发送充值成功邮件通知
func sendTopUpEmailNotification(userId int, tradeNo string, quota int64, amount float64) {
	user, err := GetUserById(userId, false)
	if err != nil {
		logger.Warn(context.Background(), fmt.Sprintf("发送充值邮件失败: 获取用户信息失败 user_id=%d error=%q", userId, err.Error()))
		return
	}
	
	if user.Email == "" {
		return // 用户没有绑定邮箱，跳过邮件通知
	}
	
	if config.SMTPServer == "" {
		return // SMTP 未配置，跳过邮件通知
	}
	
	subject := "充值成功通知 - QuantumClaw"
	body := fmt.Sprintf(
		`<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #10b981;">充值成功 ✓</h2>
			<p>尊敬的 %s，您好！</p>
			<p>您的账户已成功充值，以下是充值详情：</p>
			<table style="border-collapse: collapse; width: 100%%; margin: 16px 0;">
				<tr><td style="padding: 8px; border: 1px solid #e5e7eb; color: #6b7280;">订单号</td><td style="padding: 8px; border: 1px solid #e5e7eb;">%s</td></tr>
				<tr><td style="padding: 8px; border: 1px solid #e5e7eb; color: #6b7280;">充值金额</td><td style="padding: 8px; border: 1px solid #e5e7eb;">¥%.2f</td></tr>
				<tr><td style="padding: 8px; border: 1px solid #e5e7eb; color: #6b7280;">到账额度</td><td style="padding: 8px; border: 1px solid #e5e7eb;">%s</td></tr>
			</table>
			<p>如有任何问题，请联系管理员。</p>
			<p style="color: #9ca3af; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
		</div>`,
		user.DisplayName,
		tradeNo,
		amount,
		common.LogQuota(quota),
	)
	
	if err := common.SendEmail(user.Email, subject, body); err != nil {
		logger.Warn(context.Background(), fmt.Sprintf("发送充值邮件失败 user_id=%d email=%s error=%q", userId, user.Email, err.Error()))
	}
}

// TopUpScan 扫描订单（实现driver.Valuer和sql.Scanner接口，用于加密存储）
type TopUpScan struct {
	Data string `json:"data"`
}

func (ts *TopUpScan) Scan(value interface{}) error {
	if value == nil {
		ts.Data = ""
		return nil
	}
	
	str, ok := value.(string)
	if !ok {
		return errors.New("无效的TopUpScan值")
	}
	
	ts.Data = str
	return nil
}

func (ts TopUpScan) Value() (driver.Value, error) {
	if ts.Data == "" {
		return nil, nil
	}
	return ts.Data, nil
}

// GetUserTopUps 获取用户的充值订单列表（分页）
func GetUserTopUps(userId int, page int, pageSize int) ([]TopUp, int64, error) {
	var topUps []TopUp
	var total int64
	
	// 计数
	if err := DB.Model(&TopUp{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	if err := DB.Where("user_id = ?", userId).Order("id DESC").Offset(offset).Limit(pageSize).Find(&topUps).Error; err != nil {
		return nil, 0, err
	}
	
	return topUps, total, nil
}

// CleanExpiredTopUps 清理过期的订单（定时任务调用）
func CleanExpiredTopUps() error {
	now := common.GetTimestamp()
	
	// 将过期的pending订单标记为expired
	result := DB.Model(&TopUp{}).
		Where("status = ? AND expire_time > 0 AND expire_time < ?", TopUpStatusPending, now).
		Update("status", TopUpStatusExpired)
	
	if result.Error != nil {
		return result.Error
	}
	
	logger.Info(context.Background(), fmt.Sprintf("清理过期订单完成 count=%d", result.RowsAffected))
	
	return nil
}
