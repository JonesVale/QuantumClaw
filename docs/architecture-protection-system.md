# QuantumClaw 架构保障体系 — 从关键问题到防御设计

> 作者: 架构组 | 2026-06-06
> 核心命题: 保障资金、数据、可用性三大底线

---

## 一、最大的危险清单

如果问我"这个项目最怕什么"，按严重程度排：

```
S级（资金直接损失）
  1. 计费扣款多扣 / 漏扣 → 平台赔钱 or 用户跑路
  2. 入驻费该收没收 → 平台收入归零
  3. 支付双花（double-spend）→ 给两次钱

A级（数据完整性与安全）
  4. 数据库损坏 / 数据丢失 → 历史账单全部作废
  5. API Key 泄露 → 第三方 Provider 被拖库
  6. 用户 Token 伪造 → 任意账户可被接管

B级（可用性）
  7. 上游 Provider 全部宕机 → 平台全面不可用
  8. 支付通道故障 → 用户充不了钱
  9. 数据库连接池耗尽 → 请求全部排队超时

C级（体验与口碑）
  10. 用户不知道扣了多少钱 → 信任流失
  11. Provider 不退款 → 争议处理不了
  12. 结算迟迟不到账 → Provider 提桶跑路
```

## 二、针对每个风险的设计与防御

### S-1: 计费扣款一致性

**问题场景：** 用户调用 API，系统先预扣再后扣。如果预扣和后扣之间的逻辑有 bug，可能导致：扣多了（用户投诉）或扣少了（平台亏损）。

**当前代码分析：**
`relay/common_handler/billing.go` 中的 `PreConsumeBilling` 和 `PostConsumeBilling` 构成了一个两阶段计费协议。
`service/cash_billing.go` 的 `PostConsumeDeduct` 是三级扣款链：现金→佣金→配额→挂账。

**防御策略：**

```go
// 第一道防线：后扣金额异常时主动告警，不静默退款
func PostConsumeBilling(ctx context.Context, source string, meta *meta.Meta,
    finalQuota, preConsumed int64) {

    delta := preConsumed - finalQuota

    // ---- 新增保护 ----
    // 如果退费金额超过 10%，告警（可能预扣太少或实际用量异常）
    if preConsumed > 0 && float64(delta)/float64(preConsumed) > 0.10 {
        AlertOps(fmt.Sprintf("unusual billing delta: user=%d pre=%d final=%d delta=%d",
            meta.UserID, preConsumed, finalQuota, delta))
    }

    // 退费超过预扣额 → 表示系统算错了，直接告警 + 记录审计
    if delta < -preConsumed {
        LogAudit(AuditTypeBillingError, meta.UserID,
            map[string]int64{"pre": preConsumed, "final": finalQuota})
        AlertOps("billing error: refund exceeds pre-consumed amount")
    }
    // ---- 保护结束 ----
}
```

**核心原则：预扣是要保证服务可用，后扣是保证金额准确。两个数字不一致时，必须有告警。宁丢用户不可丢信任。**

### S-2: 入驻费该收没收

**问题场景：** `AutoSettleMonthlyFees` 逻辑出错，或者 DB 事务未提交，导致入驻费记录没生成，月底结算时发现少收了几十万。

**防御策略：**

```go
// 第一道：自动结算后验证总数
func VerifyMonthlyFeeSettlement(year int, month time.Month) error {
    // 统计所有 pending 入驻费
    var totalPending int64
    DB.Model(&PlatformFeeRecord{}).
        Where("period = ? AND status = ?", fmt.Sprintf("%04d-%02d", year, month), "pending").
        Select("COALESCE(SUM(fee_amount), 0)").Scan(&totalPending)

    // 统计当月所有 Provider 收益
    start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Unix()
    end := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC).Unix()
    var totalRevenue int64
    DB.Model(&ProviderEarning{}).
        Where("created_at >= ? AND created_at < ? AND status = ?", start, end, "settled").
        Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue)

    // 如果入驻费总额远低于预期，告警
    feeRate := GetAverageFeeRate()
    expectedFee := int64(float64(totalRevenue) * feeRate / 100)
    if float64(totalPending) < float64(expectedFee)*0.5 {
        AlertOps(fmt.Sprintf(
            "suspicious low fee settlement: period=%s pending=%d expected=%d revenue=%d",
            fmt.Sprintf("%04d-%02d", year, month), totalPending, expectedFee, totalRevenue))
    }
    return nil
}
```

**第二道防线：入驻费 cron 配置独立告警**

当前 cron 是 `main.go` 里的一个 goroutine。如果这个 goroutine panicked 而没人知道，系统会安静地停止计算入驻费。

```go
// 改为：带 recover 和独立告警的 cron
func CronMonthlySettlement(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Hour)
    for {
        select {
        case <-ticker.C:
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        AlertOps(fmt.Sprintf("CRON PANIC: monthly settlement: %v", r))
                    }
                }()
                AutoSettleMonthlyFees()
            }()
        case <-ctx.Done():
            return
        }
    }
}
```

**核心原则：所有定期任务必须 panic-safe。一个 goroutine 死了不能影响其他 goroutine。**

### S-3: 支付双花（Double-Spend）

**问题场景：** Stripe/Alipay 的 webhook 同时发送两个通知，或者重试导致同一次充值被处理两次。

**当前代码分析：**
`model/topup.go:CompleteTopUp` 使用了事务 + 悲观锁 `clause.Locking{Strength: "UPDATE"}`。这是正确的做法。

```go
// existing (correct):
err := DB.Transaction(func(tx *gorm.DB) error {
    tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&topUp, id)
    if topUp.Status != TopUpStatusPending { // 已经处理过了
        return nil
    }
    // ... 扣款
})
```

**需要补充：幂等键（Idempotency Key）**

如果 webhook 到达时网络故障，支付已经完成但交易没写入 DB。重试时系统不知道这笔交易处理过，会再扣一次。

```go
// 在 webhook 入口处加幂等键检查
func ProcessPaymentWebhook(provider string, tradeNo string, rawBody []byte) error {
    // 幂等键 = 支付提供商 + tradeNo
    key := provider + ":" + tradeNo

    // 尝试插入幂等记录（唯一约束）
    err := DB.Create(&IdempotencyKey{Key: key, ProcessedAt: time.Now()}).Error
    if err != nil {
        if IsDuplicateKeyError(err) {
            return nil // 已经处理过了，幂等返回
        }
        return err
    }

    // 正常执行 CompleteTopUp
    return CompleteTopUp(tradeNo, provider, ...)
}
```

**核心原则：所有涉及钱的操作，要么不做，要么做一次且仅一次。幂等键是兜底。**

### A-4: 数据库完整性与恢复

**问题场景：** 硬盘故障、SQL 注入（虽然用了 GORM）、手动操作失误删表。

**三层防御：**

```
第一层：备份
  └── 每天凌晨 3:00 全量备份（pg_dump / sqlite3 .backup）
  └── 每 6 小时 WAL 归档
  └── 保留最近 30 天备份

第二层：恢复演练
  └── 每月一次恢复到测试环境验证
  └── 验证：恢复后 CompleteTopUp 能否正常回滚

第三层：GORM 写操作保护
  └── 所有 Update/Delete 操作必须有 WHERE（防止全表更新）
  └── struct 级别的检查：创建时检查必填字段
```

```go
// 避免 .Where("id = ?", id) 写错成无 WHERE 的全表更新
// 建议在 model 层做保护：
func (s *Store) BeforeUpdate(tx *gorm.DB) error {
    if tx.Statement.WhereClause == nil || tx.Statement.WhereClause.Build(tx.Statement) == "" {
        return fmt.Errorf("refusing update without WHERE clause on Store")
    }
    return nil
}
```

**核心原则：备份是最后一道防线，但应该在第一次上线前就配好。不是等出事了才想起。**

### A-5: API Key 泄露

**当前代码分析：**
- 存储：AES-256-GCM 加密 ✅（`common/encrypt/encrypt.go`）
- 传输：仅解密后用于当前请求，不在日志中输出 ✅
- 获取：只有 `GetChannelById(id, true)` 时解密 ✅

**漏掉的：**

```go
// 风险 1: GetAllChannels("all") 会解密所有 Key！
// controller/quantum.go:197 中调用 GetAllChannels(0,0,"all") 后遍历匹配
// 虽然只用了 Key 来创建适配器，但 Key 在内存中存活了整个请求周期

// 风险 2: 错误日志可能泄漏 Key 的前几位
logger.Errorf("channel %s key mismatch", channel.Key[:8])

// 修复：
// 1. GetAllChannels 的 "all" 模式不应无条件解密，改按需解密
// 2. 日志中永远不要输出 Key，即使前 8 位
```

**核心原则：API Key 在系统内部的存活时间越短越好。用完即忘。**

### A-6: Token 伪造

**当前代码分析：**
`middleware/auth.go` 中的 TokenAuth 中间件根据 Token 查用户。

```go
// existing (大致):
func TokenAuth() gin.HandlerFunc {
    token := c.Request.Header.Get("Authorization")
    token = strings.TrimPrefix(token, "Bearer ")
    user := model.GetUserByToken(token)
    // ... 设置上下文
}
```

**风险：Token 是明文存储的。如果 DB 被拖，所有用户的 Token 直接泄露。**

**修复：**

```go
// 1. Token 存储时做 bcrypt hash（与密码相同）
// 2. Token 验证时用 bcrypt.CompareHashAndPassword
// 3. 只在首次生成时返回一次明文 Token
```

**当前代码没有这一步。这是安全性上的一个明显缺口。**

### B-7: 上游全部宕机

**当前代码分析：**
`middleware/distributor.go` 的 `Distribute` 中间件在没有任何可用渠道时返回 503。

**防御的：**
- 平台兜底池（`IsPlatformPool`）
- 健康检测自动暂停不可用 Provider

**漏掉的：**

```go
// 平台兜底池耗尽时的降级策略
func onAllProvidersDown(modelName string) Response {
    // 1. 如果是流式请求，给用户返回一个友好的 JSON，而不是连接中断
    // 2. 记录告警，触发 PagerDuty
    AlertPagerDuty(fmt.Sprintf("ALL PROVIDERS DOWN: %s", modelName))

    // 3. 如果是重要的请求，可以进入排队模式
    if IsCriticalUser(c.GetInt("id")) {
        return EnqueueForLater(modelName)
    }

    // 4. 返回清晰的错误信息，不是裸 503
    return ServiceUnavailable("当前该模型暂无可用资源，请稍后再试", "model_unavailable")
}
```

**核心原则：降级比报错好，排队比拒绝好，友好报错比裸断连好。**

### B-8: 支付通道故障

**防御策略：**

```go
// 支付通道健康检测
func CheckPaymentGateways() []GatewayStatus {
    // 每秒检测一次，失败连续 3 次告警
    // Stripe / Alipay / Binance 各一个 goroutine
    gateways := []string{"stripe", "alipay", "binance"}
    results := []GatewayStatus{}
    for _, g := range gateways {
        status := probe(g)
        if !status.OK {
            consecutiveFailures[g]++
            if consecutiveFailures[g] >= 3 {
                AlertOps(fmt.Sprintf("payment gateway %s down for 3 checks", g))
                // 自动降级：隐藏该支付方式按钮
                DisablePaymentMethod(g)
            }
        } else {
            consecutiveFailures[g] = 0
        }
        results = append(results, status)
    }
    return results
}
```

### B-9: 数据库连接池耗尽

**当前代码分析：**
项目使用了 GORM，默认连接池配置是 GORM 默认值（max open = 0，即不限制）。

```go
// 当前 main.go 中：
// sqlDB, _ := db.DB()
// sqlDB.SetMaxOpenConns(25)  ← 没有这一行！
// sqlDB.SetMaxIdleConns(10)   ← 没有！

// 修复：
sqlDB, err := model.DB.DB()
if err != nil {
    logger.SysError("get db object: " + err.Error())
} else {
    sqlDB.SetMaxOpenConns(25)        // 最多 25 个连接
    sqlDB.SetMaxIdleConns(10)        // 最多 10 个空闲连接
    sqlDB.SetConnMaxLifetime(5 * time.Minute) // 一个连接最多活 5 分钟
}
```

**这是当前代码的一个明显缺口。高并发下连接池不会自动关闭旧连接，可能导致数据库连接数飙升。**

### C-10: 用户不知道扣了多少钱

**当前代码分析：**
用户能看到余额（`GET /self/balance`），但看不到每笔消费的明细。

**修复思路：**
`PostConsumeDeduct` 每次扣款时都应该记录一条可读的消费明细：

```go
type ConsumeRecord struct {
    ID           int    `json:"id"`
    UserID       int    `json:"user_id" gorm:"index"`
    AmountCents  int64  `json:"amount_cents"`
    ModelName    string `json:"model_name"`
    ChannelName  string `json:"channel_name"`
    Source       string `json:"source"`  // cash / commission / quota
    BalanceAfter int64  `json:"balance_after"`
    CreatedAt    int64  `json:"created_at"`
}

// 用户端：
GET /self/consume/records?page=1&page_size=20
→ 查看每笔消费详情，知道"我的钱花哪了"
```

### C-11: 争议处理

当前没有争议系统。Provider 收到钱后不提供服务，用户没有申诉渠道。

**需要新增：**

```go
type Dispute struct {
    ID           int    `json:"id" gorm:"primaryKey"`
    TradeNo      string `json:"trade_no" gorm:"index"`
    BuyerID      int    `json:"buyer_id"`
    SellerID     int    `json:"seller_id"`
    AmountCents  int64  `json:"amount_cents"`
    Reason       string `json:"reason"`
    Status       string `json:"status"` // opened / resolved_for_buyer / resolved_for_seller
    Evidences    JSON   `json:"evidences"` // 聊天记录/日志
    ResolvedBy   int    `json:"resolved_by"`
    ResolvedAt   int64  `json:"resolved_at"`
}
```

争议不是高频事件，但不能没有。没有争议处理机制 = 平台不为交易背书 = 用户不敢通过平台交易。

### C-12: 结算不及时

Provider 看到"已完成"的订单迟迟不到账，就会对平台失去信任。

```go
// 结算 SLA 告警
func CheckSettlementSLA() {
    var stuckCount int64
    DB.Model(&ProviderEarning{}).
        Where("status = ? AND created_at < ?", "settled", time.Now().Add(-72*time.Hour)).
        Count(&stuckCount)
    if stuckCount > 0 {
        AlertOps(fmt.Sprintf("%d earnings settled > 72h but not paid out", stuckCount))
    }
}
```

---

## 三、我是你，我会立刻做的 5 件事

按优先级排列，不需要全部做完，但需要开始做：

### 1. Token 存储加 bcrypt

```go
// model/token.go: 当前 Token 明文存储
// 改为：
func (token *Token) BeforeCreate(tx *gorm.DB) error {
    if !isBcrypt(token.Key) { // bcrypt 以 $2a$/$2b$/$2y$ 开头
        hashed, err := bcrypt.GenerateFromPassword([]byte(token.Key), bcrypt.DefaultCost)
        if err != nil {
            return err
        }
        token.Key = string(hashed)
    }
    return nil
}
```

**改动量：** ~20 行。但这是整个系统的认证基石。不做的话，DB 一泄露所有用户账号全完。

### 2. 数据库连接池配置

```go
// main.go 中 InitDB 之后加：
sqlDB, err := model.DB.DB()
if err == nil {
    sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)
}
```

**改动量：** ~10 行。不改的话，高并发下 DB 会被拖死。

### 3. 支付 webhook 幂等键

```go
// 所有支付 webhook 入口加幂等键表 + 唯一约束
// Stripe / Alipay / Binance / Creem / Waffo 共 6 处
```

**改动量：** ~80 行。不改的话，极端情况可能 double-spend。

### 4. 所有定期 goroutine 加 recover

```go
// main.go 中的每个 go func() 都加：
go func() {
    defer func() {
        if r := recover(); r != nil {
            logger.SysError(fmt.Sprintf("goroutine panicked: %v", r))
            // 重启 goroutine，而不是让它永远停掉
            time.Sleep(5 * time.Second)
            go thisFunction()
        }
    }()
    // ... original logic
}()
```

**改动量：** ~30 行。不改的话，任一 goroutine panic 会导致某个定时功能永久停用而无人知晓。

### 5. ConsumeRecord 消费明细

```go
// PostConsumeDeduct 每次扣款时写一条明细
// 用户端一个 API 查看
```

**改动量：** ~100 行。不改的话，用户看到余额莫名减少会投诉和流失。

---

## 四、监控与告警体系

没有监控的系统等于盲人开车。最少需要：

```
可用性监控 (每 5 分钟)
  ├── /status 端点存活检测
  ├── 上游 API 可用性采样
  └── 平台兜底池余量 (%) < 20% → 告警

业务监控 (每小时)
  ├── 入驻费结算执行成功
  ├── 今日交易笔数 vs 昨日 < 50% → 告警
  ├── 今日新注册用户 vs 昨日 < 50% → 告警
  └── 数据库备份执行成功
```

```go
// /status 端点需要返回的内容：
type HealthStatus struct {
    Status          string `json:"status"`       // ok / degraded / down
    DB              string `json:"db"`           // ok / down
    Redis           string `json:"redis"`        // ok / down / disabled
    ActiveUsers     int    `json:"active_users"`
    ChannelsOnline  int    `json:"channels_online"`
    LastFeeSettle   string `json:"last_fee_settle"`
    LastBackup      string `json:"last_backup"`
    PlatformPoolPct int    `json:"platform_pool_pct"` // 平台池余量 %
}
```

---

## 五、总结

如果让我用一句话概括这个项目的架构保障思路：

> **所有跟钱有关的逻辑，都要有冗余验证和自动告警。所有跟用户有关的数据，都要有加密保护和备份恢复。所有定期执行的代码，都要能 panic 而不死。**

| 等级 | 事项 | 是否已做 | 改动难度 |
|------|------|---------|---------|
| 🔴 立刻 | Token bcrypt 存储 | ❌ | 低 (~20行) |
| 🔴 立刻 | 数据库连接池配置 | ❌ | 低 (~10行) |
| 🟡 紧急 | Webhook 幂等键 | ❌ | 中 (~80行) |
| 🟡 紧急 | Goroutine panic 保护 | ❌ | 低 (~30行) |
| 🟡 紧急 | 消费明细记录 | ❌ | 中 (~100行) |
| 🟢 重要 | 平台池余量告警 | ❌ | 低 (~30行) |
| 🟢 重要 | 入驻费结算验证 | ❌ | 中 (~60行) |
| 🟢 重要 | /status 健康检查 | ❌ | 低 (~40行) |
| 🟢 重要 | 数据库自动备份 | ❌ | 外部工具 |
| 📋 待定 | 争议处理系统 | ❌ | 大 (~300行) |
