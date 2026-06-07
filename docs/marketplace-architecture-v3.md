# QuantumClaw 市场架构 v3 — 店铺体系 + 阶梯入驻费

> 文档版本: v3 | 最后更新: 2026-06-06 | 作者: 架构组

---

## 一、核心理念

### 1.1 经营模型

```
我们不是 API 代理，我们不是云服务商。

我们是一个资源交易市场：
  · 任何人都可以把 API 算力挂上来卖（开店）
  · 任何人都可以来这里买（搜索比价）
  · 平台保证买卖双方都履约（Escrow + 健康检测 + 兜底池）

平台的收入不是靠抽交易佣金，而是靠卖摊位（入驻费）：
  · 交易手续费（Transaction Fee）→ 支付通道成本，代收代付
  · 入驻费（Platform Fee）→ 平台利润，按月收取
```

### 1.2 入驻费阶梯（管理员配置）

| 店铺等级 | 入驻费率 | 自动升级条件 | 权益 |
|---------|---------|-------------|------|
| 普通店 (basic) | 10% | 开店默认 | 基础路由、T+7 结算 |
| 金牌店 (gold) | 8% | 月营业额 ≥ ¥1,000 | 优先推荐、T+1 结算 |
| 旗舰店 (flagship) | 5% | 月营业额 ≥ ¥10,000 | 最高权重、实时结算、专属客服 |

- 月营业额 < ¥100 → 免入驻费（小卖家保护）
- 等级可由管理员（≥ 100）在控制台手动调整
- 连续 3 个月无交易 → 自动降回 basic

### 1.3 三层路由保障

```
Layer 1: 第三方 Provider 按价格排序 → 选最低价
Layer 2: 无可用第三方 → 选其他第三方（评分优先）
Layer 3: 全部不可用 → 平台兜底池（不断货）
```

---

## 二、数据模型

### 2.1 店铺（Store）

```go
type StoreTier string
const (
    StoreTierBasic    StoreTier = "basic"     // 普通店 10%
    StoreTierGold     StoreTier = "gold"      // 金牌店 8%
    StoreTierFlagship StoreTier = "flagship"  // 旗舰店 5%
)

type Store struct {
    ID           int       `json:"id" gorm:"primaryKey"`
    UserID       int       `json:"user_id" gorm:"uniqueIndex"`
    Name         string    `json:"name"`
    Tier         StoreTier `json:"tier" gorm:"default:'basic'"`
    Status       string    `json:"status" gorm:"default:'active'"`
    Rating       float64   `json:"rating"`
    TotalSales   int64     `json:"total_sales"`
    OpenedAt     int64     `json:"opened_at"`
    LastActiveAt int64     `json:"last_active_at"`
}
```

### 2.2 资源上架（Listing）

```go
type Listing struct {
    ID            string    `json:"id" gorm:"primaryKey"`
    StoreID       int       `json:"store_id" gorm:"index"`
    ProviderID    int       `json:"provider_id"`
    ModelName     string    `json:"model_name" gorm:"index"`
    Region        string    `json:"region"`                    // china / overseas
    PricePerUnit  int64     `json:"price_per_unit"`            // 分/千 token
    Unit          string    `json:"unit"`                      // 1k_tokens / per_shot / per_second
    MinPurchase   int64     `json:"min_purchase"`
    MaxConcurrent int       `json:"max_concurrent"`
    Status        string    `json:"status" gorm:"default:'active'"` // active/paused/archived

    AvgLatencyMs  float64   `json:"avg_latency_ms"`
    Availability  float64   `json:"availability"`
    TotalOrders   int64     `json:"total_orders"`
    AvgRating     float64   `json:"avg_rating"`
    CreatedAt     int64     `json:"created_at"`
}
```

### 2.3 阶梯入驻费配置（PlatformFeeConfig）

```go
type PlatformFeeConfig struct {
    Tier     StoreTier `json:"tier" gorm:"primaryKey"`
    Rate     float64   `json:"rate"`      // 5.0 / 8.0 / 10.0
    MinSkip  int64     `json:"min_skip"`  // 月营业额低于此值免入驻费（分，默认 10000 = ¥100 ）
    UpdatedBy int      `json:"updated_by"`
    UpdatedAt int64    `json:"updated_at"`
}
```

### 2.4 入驻费记录（PlatformFeeRecord）— 已有，扩展

```go
type PlatformFeeRecord struct {
    Id           int     `json:"id"`
    StoreID      int     `json:"store_id"`       // 新增：关联店铺
    UserId       int     `json:"user_id"`
    Period       string  `json:"period"`          // "2026-06"
    TotalRevenue int64   `json:"total_revenue"`
    FeeRate      float64 `json:"fee_rate"`
    FeeAmount    int64   `json:"fee_amount"`
    Status       string  `json:"status"`          // pending / deducted / skipped
    CreatedAt    int64   `json:"created_at"`
    DeductedAt   int64   `json:"deducted_at"`
}
```

### 2.5 交易托管（Escrow / 扩展 TopUp）

```go
// 复用现有 TopUp 表，加 esrow 语义字段
// TopUp 已有字段：
//   TradeNo, UserId, Amount, Money, PaymentMethod, Status, CreatedAt
// 新增：
type EscrowExtension struct {
    TopUpID    int    `gorm:"primaryKey"`
    SellerID   int    `gorm:"index"`              // Provider
    StoreID    int    `gorm:"index"`               // 店铺
    ListingID  string `gorm:"index"`                // 资源
    EscrowStatus string `gorm:"default:'held'"`    // held / released / disputed / refunded
    ReleasedAt int64
}
```

### 2.6 评价（Review）

```go
type Review struct {
    ID         int     `json:"id" gorm:"primaryKey"`
    ListingID  string  `json:"listing_id" gorm:"index"`
    StoreID    int     `json:"store_id" gorm:"index"`
    BuyerID    int     `json:"buyer_id"`
    Rating     int     `json:"rating"`      // 1-5
    Content    string  `json:"content"`
    CreatedAt  int64   `json:"created_at"`
}
```

### 2.7 店铺等级变更记录（StoreTierLog）

```go
type StoreTierLog struct {
    ID        int       `json:"id" gorm:"primaryKey"`
    StoreID   int       `json:"store_id" gorm:"index"`
    FromTier  StoreTier `json:"from_tier"`
    ToTier    StoreTier `json:"to_tier"`
    Reason    string    `json:"reason"`     // auto_upgrade / admin_set / auto_downgrade
    Operator  int       `json:"operator"`   // 管理员 ID（自动变更时为 0）
    CreatedAt int64     `json:"created_at"`
}
```

---

## 三、数据关系图

```
User (1) ────────────────────────────── Provider 身份
  │
  ├── (1)─ Store (1) ──── PlatformFeeConfig (N)  ← 阶梯费率
  │             │
  │             ├── Listing (N) ──── Channel (1)  ← 资源上架 → 执行渠道
  │             │       │
  │             │       └── Review (N)            ← 用户评价
  │             │
  │             └── PlatformFeeRecord (N)         ← 入驻费结算记录
  │
  ├── (N)─ TopUp ──── EscrowExtension (0..1)     ← 交易托管
  │
  └── (N)─ StoreTierLog                          ← 等级变更审计
```

---

## 四、API 设计

### 4.1 市场（Market）— 无需登录

```
GET  /api/market/models                  → 所有可买模型列表 + 最低价
GET  /api/market/price/:model            → 某模型各店铺报价行情
GET  /api/market/stores?model=xxx        → 卖某模型的店铺列表（含评分、价格排序）
GET  /api/market/listing/:id             → 某个上架资源的详情
```

### 4.2 店铺（Store）— Provider 登录

```
POST /api/store/register                 → 开店（创建 Store + 绑定 Provider）
GET  /api/store/profile                  → 我的店铺信息
PUT  /api/store/profile                  → 修改店铺信息

POST /api/store/listings                 → 上架资源
GET  /api/store/listings                 → 我的上架列表
PUT  /api/store/listings/:id/price       → 调价
PUT  /api/store/listings/:id/status      → 暂停/恢复/下架

GET  /api/store/earnings                 → 收益看板（分日/月/总）
GET  /api/store/fees                     → 入驻费明细（pending/deducted/skipped）
POST /api/store/withdraw                 → 提现（自动扣入驻费）
GET  /api/store/withdrawals              → 提现记录
```

### 4.3 用户偏好 — 买家登录

```
PUT  /api/user/preferred_store           → 设置偏好的店铺（路由优先用这家）
GET  /api/user/preferred_store           → 查看当前的偏好
```

### 4.4 管理后台 — Admin ≥ 100

```
GET  /api/admin/store-tiers              → 等级配置列表
PUT  /api/admin/store-tiers/:tier        → 修改某等级的入驻费率/免租门槛

GET  /api/admin/stores                   → 店铺列表（含筛选：等级/状态/活跃度）
GET  /api/admin/stores/:id               → 店铺详情
PUT  /api/admin/stores/:id/tier          → 手动调整店铺等级

GET  /api/admin/platform-fees            → 入驻费汇总统计
GET  /api/admin/listings                 → 全部上架资源审核
PUT  /api/admin/listings/:id/status      → 暂停违规上架
```

---

## 五、核心业务流程

### 5.1 开店流程

```
用户 POST /api/store/register
  │
  ├── 1. 创建 Store（UserID, Name, Tier=basic）
  ├── 2. 检查用户是否有已配置的 Channel
  │     ├── 有 → 自动创建 Listing 关联到该 Channel
  │     └── 无 → 返回"请先配置 API Key"
  ├── 3. 验证 API Key 可用性
  ├── 4. 加入健康检测队列
  └── 5. 店铺状态 = active
```

### 5.2 入驻费月结流程

```
每月 1 号凌晨 2:00（当前 cron 逻辑 + 扩展）
  │
  ├── 1. 遍历所有 active 店铺
  ├── 2. 查询当月该店铺所有已结算收益总和
  ├── 3. 查该店铺等级的 FeeConfig
  │     ├── 月营业额 < MinSkip → 创建 skipped 记录
  │     └── 月营业额 ≥ MinSkip → 计算入驻费 = 营业额 × Rate%
  ├── 4. 创建 PlatformFeeRecord（status = pending）
  ├── 5. 店铺自动升级检查：
  │     ├── 总销售额 ≥ ¥10,000 → 升级 Flagship
  │     ├── 总销售额 ≥ ¥1,000  → 升级 Gold
  │     └── 3 个月无交易 → 降级 Basic
  └── 6. 记录 StoreTierLog
```

### 5.3 路由分发流程

```
用户请求 model = "gpt-4o"
  │
  ├── 检查用户是否指定了 preferred_store
  │     ├── 有 → 该店铺有 active listing → 用他家的
  │     └── 无 → 走自动路由
  │
  ├── Layer 1: 第三方最便宜
  │     SELECT FROM listings WHERE model = ? AND status = active
  │     ORDER BY price_per_unit ASC LIMIT 1
  │     → 找到就用（记录用户本次用的哪家）
  │     → 没有 → 降级
  │
  ├── Layer 2: 其他第三方（按评分排序）
  │     ORDER BY avg_rating DESC
  │     → 找到就用
  │     → 没有 → 降级
  │
  ├── Layer 3: 平台兜底池
  │     → 找到就用（记录告警："市场资源紧缺，已启用兜底"）
  │     → 没有 → 503 + 运营告警
  │
  └── 路由完成后，将对应的 Listing 信息写入请求上下文
      用于 Escrow 记录和后续评价
```

### 5.4 提现扣入驻费流程

```
Provider POST /api/store/withdraw
  │
  ├── 1. 计算可用余额 = 累计已结算收益 - 已提现金额
  ├── 2. 查询待扣入驻费
  │     SELECT SUM(fee_amount) FROM platform_fee_records
  │     WHERE store_id = ? AND status = 'pending'
  ├── 3. 校验：请求金额 ≤ 可用余额 - 待扣入驻费
  ├── 4. 创建提现申请
  │     ├── amount = 请求金额
  │     ├── platform_fee_amount = 待扣入驻费
  │     └── net_amount = amount - platform_fee_amount
  ├── 5. 标记入驻费为 deducted
  ├── 6. 审批（小额自动、大额人工）
  └── 7. 打款
```

---

## 六、新增文件清单

```
controller/
  ├── store.go              # 开店、店铺信息、上架/下架
  ├── market.go             # 市场行情、搜索、比价
  └── admin_store.go        # 后台店铺管理、等级配置

model/
  ├── store.go              # Store / Listing / StoreTierLog
  ├── fee_config.go         # PlatformFeeConfig
  ├── escrow.go             # EscrowExtension（或扩展 TopUp）
  └── review.go             # Review

service/
  ├── store_service.go      # 开店逻辑、等级自动升降
  ├── fee_service.go        # 入驻费计算（阶梯费率版本）
  ├── market_service.go     # 搜索、比价、推荐排序
  └── escrow_service.go     # 托管/释放/超时处理

middleware/
  └── store_context.go      # 将 Store/Listing 信息注入请求上下文

router/
  └── marketplace.go        # 市场相关路由注册（可选拆分）
```

---

## 七、复用与改动清单

### ✅ 完全复用

| 文件 | 说明 |
|------|------|
| `model/topup.go` | 充值 + 交易资金流 |
| `model/platform_fee.go` | 入驻费结算记录（改 StoreID） |
| `model/commission.go` | 提现扣入驻费逻辑 |
| `model/provider_earning.go` | Provider 分账记录 |
| `middleware/distributor.go` | 路由分发框架 |
| `model/settlement.go` | `GetTransactionFeeRate` 费率配置 |

### 🔧 需要改动的

| 文件 | 改动内容 |
|------|----------|
| `model/platform_fee.go` | `UserId` → 增加 `StoreID`；`getPlatformFeeRate()` 改为查 `PlatformFeeConfig` |
| `model/commission.go` | 提现增加 StoreID 关联 |
| `model/channel.go` | 增加 `StoreID` 外键 + `IsPlatformPool` 标记 |
| `model/main.go` | AutoMigrate 新增表 + Seed 默认 FeeConfig |
| `main.go` | AutoSettleMonthlyFees 扩展为按 Store 等级计算 |

### 🆕 纯新增

| 文件 | 预估行数 |
|------|----------|
| `model/store.go` | 80 |
| `model/fee_config.go` | 40 |
| `model/escrow.go` | 50 |
| `model/review.go` | 30 |
| `controller/store.go` | 200 |
| `controller/market.go` | 150 |
| `controller/admin_store.go` | 120 |
| `service/store_service.go` | 100 |
| `service/fee_service.go` | 80 |
| `service/market_service.go` | 60 |
| `service/escrow_service.go` | 80 |
| 前端相关 | 待定 |

**核心新增纯后端代码约 1000 行。** 大多数是 CRUD + 现有逻辑的编排。

---

## 八、落地路线图

### Week 1: 店铺体系

```
目标：任何人都能开店，开店即入驻

任务:
  ├── 1.1 创建 Store 模型 + 数据库迁移
  ├── 1.2 创建 PlatformFeeConfig 模型 + 种子数据
  ├── 1.3 开店 API（POST /api/store/register）
  ├── 1.4 控制台入驻费配置页面
  └── 1.5 创建 StoreTierLog 模型
```

### Week 2: 资源上架 + 市场

```
目标：店铺能上架资源，用户能看到行情

任务:
  ├── 2.1 创建 Listing 模型
  ├── 2.2 上架/下架 API（CRUD）
  ├── 2.3 市场行情 API（搜索/比价/排序）
  ├── 2.4 Listing → Channel 自动关联
  └── 2.5 健康检测接入 Listing
```

### Week 3: 交易 + 结算

```
目标：钱能走通，入驻费能自动收

任务:
  ├── 3.1 Escrow（扩展 TopUp + 托管状态机）
  ├── 3.2 入驻费月结改为按阶梯费率
  ├── 3.3 提现自动扣入驻费（复用 + 改 Store 关联）
  ├── 3.4 店铺自动升降级逻辑
  └── 3.5 平台兜底池标记
```

### Week 4: 评价 + 运营

```
目标：市场能自我运转，运营能看得见

任务:
  ├── 4.1 Review 模型 + API
  ├── 4.2 店铺信誉分计算
  ├── 4.3 运营仪表盘
  ├── 4.4 管理员店铺管理页面
  └── 4.5 Provider 收益看板（前端）
```

---

## 九、当前代码可直接修改的切入点

不需要等架构评审，今天就可以做的事：

### 1. 改 `PlatformFeeRecord.UserId` 为 `StoreID`

```go
// 当前
type PlatformFeeRecord struct {
    UserId int    // 供应商用户 ID
    ...
}

// 改为
type PlatformFeeRecord struct {
    StoreID int    `gorm:"index"` // 店铺 ID
    UserId  int                   // 保留冗余查询
    ...
}
```

### 2. 改 `getPlatformFeeRate()` 为查表

```go
// 当前
func getPlatformFeeRate() float64 {
    var cfg PlatformConfig
    if DB.Where("`key` = ?", "platform_fee_rate_percent").First(&cfg).Error == nil {
        ...
    }
    return 5.0
}

// 改为
func getStoreFeeRate(storeID int) float64 {
    var store Store
    if DB.First(&store, storeID).Error != nil {
        return 10.0 // 找不到就按最高的
    }
    var cfg PlatformFeeConfig
    if DB.Where("tier = ?", store.Tier).First(&cfg).Error != nil {
        return 10.0 // 找不到配的就按普通店 10%
    }
    return cfg.Rate
}
```

### 3. 在现有 cron 中加入自动升降级

```go
// AutoSettleMonthlyFees 末尾加：
func autoUpgradeStoreTier(storeID int, totalSales int64) {
    var store Store
    DB.First(&store, storeID)
    
    newTier := store.Tier
    switch {
    case totalSales >= 1000000:     // ¥10,000
        newTier = StoreTierFlagship // 5%
    case totalSales >= 100000:      // ¥1,000
        newTier = StoreTierGold     // 8%
    case store.LastActiveAt < time.Now().AddDate(0, -3, 0).Unix():
        newTier = StoreTierBasic    // 降回普通
    }
    
    if newTier != store.Tier {
        oldTier := store.Tier
        DB.Model(&store).Update("tier", newTier)
        CreateStoreTierLog(storeID, oldTier, newTier, "auto", 0)
    }
}
```

---

## 十、运营指标定义

```go
// 哪些数字值得每天看
type OpsMetrics struct {
    // 供给端
    TotalStores        int       // 总店铺数
    ActiveStores       int       // 近 7 天有交易的店铺
    NewStoresToday     int       // 今日新开店
    TotalListings      int       // 总上架资源数

    // 交易端
    TotalOrders        int64     // 今日交易笔数
    TotalRevenue       int64     // 今日交易额（分）
    AvgOrderValue      float64   // 平均客单价

    // 平台收入
    MonthlyFeeIncome   int64     // 本月入驻费收入（pending）
    CollectedFees      int64     // 已扣入驻费

    // 健康度
    PoolUtilization    float64   // 平台池使用率（>50% 需补充）
    ProviderDowntime   []Alert   // 最近异常的 Provider

    // 风险
    PendingDisputes    int       // 待处理争议
    OverdueEscrows     int       // 超时未释放的托管
}
```

---

## 附录 A: 现有代码位置索引

| 概念 | 现有位置 |
|------|----------|
| 入驻费模型 | `model/platform_fee.go` |
| 入驻费结算 cron | `main.go:260` → `model.AutoSettleMonthlyFees()` |
| 提现扣入驻费 | `model/commission.go:CreateWithdrawal()` |
| 交易费率配置 | `model/settlement.go:GetTransactionFeeRate()` |
| 渠道路由 | `middleware/distributor.go:Distribute()` |
| 分账记录 | `model/provider_earning.go` |
| 店铺概念（雏形） | `controller/store.go`（已有 Store 相关的路由） |

## 附录 B: 修改对照

| 现有文件 | 改动类型 | 说明 |
|----------|---------|------|
| `model/platform_fee.go` | 🔧 改 | UserId → StoreID；getPlatformFeeRate → getStoreFeeRate |
| `model/commission.go` | 🔧 改 | 提现关联 StoreID |
| `model/channel.go` | 🔧 改 | 加 StoreID + IsPlatformPool |
| `model/main.go` | 🔧 改 | AutoMigrate 新表 + Seed 默认数据 |
| `main.go` | 🔧 改 | cron 扩展 |
| `model/store.go` | 🆕 新增 | Store / Listing / StoreTierLog |
| `model/fee_config.go` | 🆕 新增 | PlatformFeeConfig |
| `model/escrow.go` | 🆕 新增 | EscrowExtension |
| `model/review.go` | 🆕 新增 | Review |
| `controller/store.go` | 🆕 新增 | 开店 API |
| `controller/market.go` | 🆕 新增 | 市场行情 API |
| `controller/admin_store.go` | 🆕 新增 | 后台管理 API |
| `service/store_service.go` | 🆕 新增 | 业务逻辑 |
| `service/fee_service.go` | 🆕 新增 | 入驻费计算 |
| `service/market_service.go` | 🆕 新增 | 市场搜索 |
| `service/escrow_service.go` | 🆕 新增 | 交易托管 |
