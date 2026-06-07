# QuantumClaw 全业务链路分析与架构规划

> 版本: v1.0 | 2026-06-06
> 覆盖: 全部现有业务模块 + 市场架构改造规划

---

## 第一部分：现有业务全景

### 1.1 业务模块总览

```
┌──────────────────────────────────────────────────────────────────┐
│                      QuantumClaw 业务全景                          │
│                                                                   │
│  ┌─────────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐  │
│  │  用户系统     │  │  资源渠道   │  │  交易计费   │  │  平台运营   │  │
│  │             │  │           │  │           │  │           │  │
│  │ · 注册/登录  │  │ · Channel  │  │ · TopUp   │  │ · Admin   │  │
│  │ · OAuth     │  │ · Listing │  │ · Escrow  │  │ · Option  │  │
│  │ · Token     │  │ · Rout    │  │ · 分账    │  │ · Monitor │  │
│  │ · 2FA       │  │ · Relay   │  │ · 结算    │  │ · Cascade │  │
│  │ · Org/Team  │  │ · 中继    │  │ · 入驻费   │  │ · 消息引擎│  │
│  │ · Reseller  │  │ · 适配器   │  │ · 提现    │  │ · 翻译    │  │
│  └──────┬──────┘  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  │
│         │               │               │               │        │
│         └───────────────┼───────────────┼───────────────┘        │
│                         │               │                        │
│                    ┌─────▼───────────────▼─────┐                 │
│                    │       数据支撑层             │                 │
│                    │  · GORM · Redis · 日志     │                 │
│                    │  · 批量更新 · 缓存 · 加密    │                 │
│                    └───────────────────────────┘                 │
└──────────────────────────────────────────────────────────────────┘
```

### 1.2 模块路由数 & 中间件层级

| 路由分组 | 路由数 | 中间件 | 说明 |
|---------|--------|--------|------|
| `apiRouter` | 137 | CORS + RateLimit + Security + TokenAuth | 公开/半公开 API |
| `selfRoute` | 54 | UserAuth | 用户自助 API |
| `relayV1Router` | 58 | RelayRecover + TokenAuth + Search + Geo + ... + Distribute | API 中继核心 |
| `adminRoute` | 14 | AdminAuth | 用户管理 |
| `channelUserRoute` | 9 | UserAuth | 渠道商自管 |
| `channelAdminRoute` | 8 | AdminAuth | 渠道管理 |
| `storeRoute` | 10 | UserAuth | 店铺管理（已有雏形） |
| `orgMgmt` | 12 | UserAuth | 企业组织 |
| `subscriptionRoute` | 2 | UserAuth | 订阅 |
| `taskRoute` | 9 | UserAuth | 异步任务 |
| `tokenRoute` | 7 | UserAuth | API Token |

---

### 1.3 各链路完整数据流

---

#### 链路 A: 用户注册 → 认证 → 会话

```
用户浏览器                        API                           DB
   │                             │                             │
   ├── POST /register ──────────►│                             │
   │                             ├── TurnstileCheck            │
   │                             ├── 密码 bcrypt               │
   │                             ├── 加密 AES-GCM ────────────►│ users
   │                             │◄─── id, role=1 ────────────│
   │◄── token ──────────────────│                             │
   │                             │                             │
   ├── POST /login ─────────────►│                             │
   │                             ├── LoginRateLimit            │
   │                             ├── 解密密码                  │
   │                             ├── 验证 ───────────────────►│ users
   │◄── session cookie ────────│                             │
```

**涉及模型:** User, Option, Log
**涉及文件:** `controller/user.go`, `middleware/auth.go`, `model/user.go`
**当前状态:** ✅ 完整闭环

---

#### 链路 B: 渠道管理 → API Key → 加密存储

```
Admin/渠道商                          API                           DB
   │                             │                             │
   ├── POST /channel ───────────►│                             │
   │   { name, type, key }      ├── encrypt.Encrypt(key) ────►│ channels
   │                             │   AES-256-GCM                │ (key = 加密后)
   │◄── { id, name } ──────────│                             │
   │                             │                             │
   ├── PUT /channel/:id ───────►│                             │
   │   { key: "新key" }         ├── 对比旧key                  │
   │                             ├── getChannelById            │
   │                             ├── encrypt.Encrypt(新key)    │
   │                             ├── DB.Save ─────────────────►│ channels
   │◄── updated ───────────────│                             │
```

**涉及模型:** Channel, ChannelConfig
**涉及文件:** `model/channel.go`, `common/encrypt/encrypt.go`
**当前状态:** ✅ 完整闭环（加密失败已修复为拒绝操作）

---

#### 链路 C: 用户充值 → 资金流入

```
用户                            API                           DB
  │                             │                             │
  ├── POST /topup/alipay ─────►│                             │
  │   { amount }               ├── 读取 PaymentSetting        │
  │                             ├── generateTradeNo           │
  │                             ├── TopUp.Insert ────────────►│ top_ups
  │                             ├── 构建支付宝表单             │
  │◄── payment_url ───────────│                             │
  │                             │                             │
  │ 跳转支付宝页面 → 支付                                     │
  │                             │                             │
  │◄── Alipay 异步通知 ───────│                             │
  │   (Any /webhook/alipay)    ├── RSA2 验签                  │
  │                             ├── CompleteTopUp              │
  │                             │   ├── 事务 + 悲观锁          │
  │                             │   ├── 更新状态               │
  │                             │   ├── 加用户配额            │
  │                             │   ├── 扣手续费 ────────────►│ top_ups
  │                             │   ├── 记余额流水            │
  │                             │   └── 自动还债 ────────────►│ users
  │                             │                                   │
  │                             │                           users.debt
```

**涉及模型:** TopUp, User, ProviderEarning, BalanceLog
**涉及文件:** `controller/topup_alipay.go`, `model/topup.go`, `common/payment_setting.go`
**当前状态:** ✅ 完整闭环（支付宝/Stripe/Binance/Creem/Waffo/WorldFirst 六种方式）
**数据完整性:** 悲观锁 + 事务保护 ✅

---

#### 链路 D: 用户调用资源 → 路由 → 中继 → 计费（核心链路）

```
用户请求                            API                                Provider
  │                                │                                   │
  ├── POST /v1/chat/completions ──►│                                   │
  │                                │                                   │
  │           ┌────────────────────┴────┐                              │
  │           │  Middleware Chain       │                              │
  │           │  · RelayPanicRecover    │                              │
  │           │  · TokenAuth            │  ← 验证用户 token            │
  │           │  · SearchMiddleware     │  ← 搜素增强                  │
  │           │  · GeoMiddleware        │  ← 地理适配                  │
  │           │  · PromptOptimizer      │  ← prompt 优化               │
  │           │  · ModelRateLimit       │  ← 模型限流                  │
  │           │  · Distribute ──────────┼──→ 选择渠道                  │
  │           └────────────────────────┘  │                            │
  │                                │      │                            │
  │                                │  ┌───▼──────────┐                │
  │                                │  │ 渠道选择逻辑    │               │
  │                                │  │ 1. 指定渠道    │               │
  │                                │  │ 2. 首选供应商   │               │
  │                                │  │ 3. 推广人渠道   │               │
  │                                │  │ 4. 国内兜底    │               │
  │                                │  │ 5. 海外兜底    │               │
  │                                │  │ 6. 503        │               │
  │                                │  └──────┬───────┘                │
  │                                │         │                         │
  │                                │  setupAndContinue                  │
  │                                │  · 设置 Authorization: {渠道 Key} │
  │                                │  · 设置 BaseURL / ModelMapping     │
  │                                │         │                         │
  │                                │  ┌──────▼───────┐                 │
  │                                │  │ RelayTextHelper│               │
  │                                │  │ 1. 解析模型+映射               │
  │                                │  │ 2. PreConsumeBilling           │
  │                                │  │   订阅→现金→佣金→配额→拒绝     │
  │                                │  │ 3. adaptor.DoRequest ──────────►│
  │                                │  │ 4. adaptor.DoResponse ◄───────│
  │                                │  │ 5. PostConsumeDeduct           │
  │                                │  │   现金→佣金→配额→挂账          │
  │                                │  │ 6. PostConsumeBilling          │
  │                                │  │   refund/additional charge     │
  │                                │  └──────────┬───────────────────┘ │
  │◄──── response + usage ────────│             │                      │
  │                                │             │                      │
  │                                │   BatchUpdate (异步)              │
  │                                │   · used_quota += quota           │
  │                                │   · request_count += 1            │
```

**涉及模型:** User, Token, Channel, Log, ProviderEarning, BalanceLog, TopUp
**涉及文件:** `controller/relay.go`, `relay/controller/text.go`, `relay/common_handler/billing.go`, `service/cash_billing.go`, `middleware/distributor.go`
**当前状态:** ✅ 完整闭环（最核心也最成熟的链路）

---

#### 链路 E: 第三方支付回调 → 充值完成 → 债务抵扣

```
支付通道回调                          API                           DB
  │                                │                             │
  ├── POST /webhook/stripe ──────►│                             │
  │   (带签名)                    ├── WebhookIPWhitelist        │
  │                                ├── 验签                      │
  │                                ├── matchTopUp(tradeNo)      │
  │                                ├── CompleteTopUp ──────────►│
  │                                │    ├── 事务 BEGIN           │
  │                                │    ├── SELECT...FOR UPDATE  │ top_ups (悲观锁)
  │                                │    ├── 验证状态 pending      │
  │                                │    ├── 更新 quota ─────────►│ users (原子 +)
  │                                │    ├── 扣手续费              │
  │                                │    ├── 自动还债              │
  │                                │    │   if user.Debt > 0     │
  │                                │    │      debt -= cash      │
  │                                │    │      cash 优先还债      │
  │                                │    └── 事务 COMMIT          │
  │                                │                             │
  │                                │    ├── 分账                  │
  │◄── success ──────────────────│    └── 余额流水              │
```

**涉及模型:** TopUp, User, ProviderEarning, BalanceLog
**涉及文件:** `model/topup.go`, `controller/webhook/stripe.go` (各支付方式各自)
**当前状态:** ✅ 完整闭环（6 种支付方式 + 2 种手续费率）

---

#### 链路 F: 异步任务（Midjourney/Suno/Video）→ 提交→ 回调 → 计费

```
用户                            API                           Provider
  │                             │                             │
  ├── POST /task/midjourney ───►│                             │
  │                             ├── UserAuth                  │
  │                             ├── 创建 AsyncTask ──────────►│ tasks
  │◄── { task_id } ───────────│                             │
  │                             │                             │
  │                             ├── go func() {               │
  │                             │     GetChannelById          │
  │                             │     调用 MJ API ────────────►│ Midjourney
  │                             │     ◄── 返回 ──────────────│
  │                             │     DeductTaskQuota         │
  │                             │       · 原子 UPDATE         │
  │                             │         quota -= ?          │
  │                             │         used_quota += ?     │
  │                             │         WHERE quota >= ?    │
  │                             │ }                           │
  │                             │                             │
  ├── GET /task/:id ───────────►│                             │
  │◄── status ────────────────│                             │
```

**涉及模型:** AsyncTask, User, Channel
**涉及文件:** `controller/task_controller.go`
**当前状态:** ⚠️ 有原子 UPDATE 但没有现金计费（用配额老系统）

---

#### 链路 G: Provider 收益 → 入驻费 → 提现

```
系统自动                              API                           DB
  │                                │                             │
  │  每次 relay 完成                │                             │
  ├────────────────────────────────┤  PostConsumeDeduct          │
  │                                ├── 扣用户 → 记 ProviderEarning │
  │                                │                             │
  │  每月 1 号 2:00 cron           │                             │
  ├────────────────────────────────┤  AutoSettleMonthlyFees       │
  │                                ├── 查上月 ProviderEarning     │
  │                                ├── 计算入驻费 ──────────────►│ platform_fee_records
  │                                │   (5% or 阶梯)               │
  │                                │   ≥ ¥100 → pending          │
  │                                │   < ¥100 → skipped          │
  │                                │                             │
  │  Provider 提现                  │                             │
  ├── POST /self/withdraw ────────►│                             │
  │                                ├── 取 pending 入驻费         │
  │                                ├── net = amount - 入驻费     │
  │                                ├── 审批（小额自动/大额手工）    │
  │                                ├── 扣佣金余额                 │
  │                                ├── 标记入驻费 deducted ─────►│
  │◄── { net_amount } ───────────│                             │
```

**涉及模型:** ProviderEarning, PlatformFeeRecord, WithdrawalRequest, User
**涉及文件:** `model/platform_fee.go`, `model/commission.go`, `main.go` (cron)
**当前状态:** ✅ 完整闭环（入驻费已有阶梯改造空间）

---

#### 链路 H: 订阅计费（Subscription）

```
用户                            API                           DB
  │                             │                             │
  ├── GET /subscription/plans ──►│                             │
  │◄── { plans } ─────────────│                             │
  │                             │                             │
  │ 管理员创建/修改 plan          │                             │
  ├── AdminCreateSubscription   │                             │
  │  Plan ────────────────────►│                             │
  │                             ├── DB.Save ────────────────►│ subscription_plans
  │                             │                             │
  │ 用户订阅                      │                             │
  ├── AdminCreateUserSub ──────►│                             │
  │    (后台)                    ├── 创建 UserSubscription ──►│ user_subscriptions
  │                             │                             │
  │ 调用时抵扣                    │                             │
  ├── PreConsumeBilling ───────►│                             │
  │                             ├── 先扣订阅剩余               │
  │                             ├── 不够再扣现金/佣金/配额      │
```

**涉及模型:** SubscriptionPlan, UserSubscription, User
**涉及文件:** `controller/subscription.go`, `relay/common_handler/billing.go`
**当前状态:** ✅ 完整闭环（在 PreConsumeBilling 中作为第一优先级）

---

#### 链路 I: 量子计算任务

```
用户                            API                           DB
  │                             │                             │
  ├── POST /quantum/submit ────►│                             │
  │   { provider, backend,     │    ├── 解析 provider         │
  │     qasm, shots }          │    ├── resolveProviderID     │
  │                             │    ├── GetAllChannels        │
  │                             │    ├── 匹配量子渠道          │
  │                             │    ├── 预扣费（新增）         │
  │                             │    ├── adaptor.RunTask ─────►│ 量子 API
  │                             │    ├── 如果 wait，轮询       │
  │                             │    ├── PostConsumeQuantumDeduct
  │◄── { task_id, results } ───│                             │
```

**涉及模型:** Channel, User
**涉及文件:** `controller/quantum.go`, `service/cash_billing.go` (PostConsumeQuantumDeduct)
**当前状态:** ⚠️ 计费已接入（本轮修复），但量子计算计费模型刚加入，尚未经过生产验证

---

#### 链路 J: 平台运营管理

```
管理员                             API                           DB
  │                                │                             │
  ├── GET /api/admin/monitor ─────►│                             │
  │                                ├── 统计数据                  │
  │◄── 用户/渠道/交易统计 ────────│                             │
  │                                │                             │
  ├── GET /api/option ────────────►│                             │
  │                                ├── 读取 OptionMap            │
  │◄── 系统配置 ─────────────────│                             │
  │                                │                             │
  ├── PUT /api/option ────────────►│                             │
  │                                ├── DB.Save Option ──────────►│ options
  │                                ├── updateOptionMap           │
  │◄── ok ───────────────────────│                             │
  │                                │                             │
  ├── GET /api/admin/model-brands►│                             │
  │                                ├── DB.Find ────────────────►│ model_brands
```

**涉及模型:** Option, User, Channel, Log, PlatformConfig, ModelBrand...
**涉及文件:** `controller/option.go`, `controller/admin.go`, `model/option.go`
**当前状态:** ✅ 完整闭环（根权限保护所有 option 写操作）

---

### 1.4 数据关联图（核心表关系）

```
User
  ├── 1:N Channel          ← 渠道商/Provider 的渠道
  │       ├── StoreID ──── Store（新增）  ← 店铺
  │       └── IsPlatformPool              ← 平台兜底标记
  ├── 1:N Token            ← 用户的 API Token
  ├── 1:N TopUp            ← 充值/交易记录
  │       └── EscrowStatus（新增）         ← 托管状态
  ├── 1:N ProviderEarning ← 作为 Provider 的收益
  ├── 1:N PlatformFeeRecord ← 入驻费
  ├── 1:N WithdrawalRequest ← 提现
  └── 1:N AsyncTask       ← 异步任务

Channel
  ├── N:1 User             ← 渠道所有者
  ├── StoreID ──── Store  ← 所属店铺（新增）
  ├── 1:N Log             ← 使用日志
  └── 1:N ProviderEarning ← 分账

TopUp
  ├── N:1 User             ← 充值用户
  ├── EscrowStatus（新增）  ← 托管状态
  └── 1:N BalanceLog      ← 余额流水

Store（新增）
  ├── 1:1 User             ← 店主
  └── 1:N Channel / Listing ← 资源

PlatformFeeConfig（新增，可配置）
  └── 按 Tier 配置费率

PlatformFeeRecord
  ├── N:1 Store（新增）
  └── N:1 User
```

---

## 第二部分：当前架构痛点

### 2.1 路由层的困惑

当前路由层面对的是 `Channel`（技术实体），不是 `Listing`（商业实体）。

```
Channel 承载了太多含义：
  · 渠道类型（OpenAI / Azure / Quantum...）
  · API Key（加密存储）
  · 渠道所有者（UserId）
  · 定价（CostPerUnit / SellPriceRate）
  · 模型映射
  · 可用性（LastTestPassed）
  · 系统配置（Config JSON）

缺少的语义：
  · 这是谁开的店？→ Store
  · 卖什么价格？→ Listing.PricePerUnit
  · 信誉如何？→ Store.Rating
  · 可不可以选？→ 市场行情
```

**结论：** Channel 是技术执行层，Listing 是商业展示层。需要解耦。

### 2.2 商业模式不清晰

当前系统有入驻费计算逻辑，但没有"店铺"实体来承载这个商业概念。入驻费是挂在 ProviderEarning 上算的，而不是挂在店铺上。

```
现状：入驻费 ≈ User 属性
目标：入驻费 = Store.Tier → PlatformFeeConfig.Rate
```

### 2.3 Provider 没有"开店"体验

现有 `POST /upgrade` 可以升级为 Provider，但没有：
- 店铺名称 / 品牌
- 收益看板
- 自助上架
- 自主定价

### 2.4 用户没有选择权

路由是黑盒——用户不知道自己在用谁的资源、什么价格。`Distribute` 中间件自动选最便宜的，但不告诉用户。

### 2.5 缺少评价与信誉体系

没有评价系统，新用户无法判断一个 Provider 靠不靠谱。平台只能用管理员审核来代替市场信用，这不可规模化。

---

## 第三部分：市场架构改造方案

### 3.1 改造原则

1. **不改现有链路** — 现有支付/计费/路由/中继链路全部保持不动
2. **加新层不拆旧层** — 在 Channel 之上加 Listing，在 User 之上加 Store
3. **兼容运行** — 改造过程中，老路由逻辑（Distribute）和新市场逻辑（Market）可以并存

### 3.2 改造范围图

```
┌──────────────────────────────────────────────────────────────┐
│                       新增市场层                               │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │ Store (店铺)   │  │ Listing (上架)│  │ Review (评价)     │   │
│  │ · 开店/关店    │  │ · 资源上架    │  │ · 评分/评价       │   │
│  │ · 等级/入驻费  │  │ · 自主定价    │  │ · 信誉分计算      │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │
│         │                 │                    │              │
│         ▼                 ▼                    ▼              │
│  ┌──────────────────────────────────────────────────┐        │
│  │             市场路由层（新 Distribute）              │        │
│  │  Layer 1: 第三方 Listing 价格排序 → 选最低价        │        │
│  │  Layer 2: 其他第三方（评分优先）                      │        │
│  │  Layer 3: 平台池兜底（IsPlatformPool）               │        │
│  └──────────────────────┬───────────────────────────┘        │
└─────────────────────────┼────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                       现有执行层                               │
│                                                               │
│  Channel → adaptor.DoRequest → Provider API                   │
│  TopUp → CompleteTopUp → 计费/分账                             │
│  Distribute → 旧路由逻辑（兼容保留）                            │
└──────────────────────────────────────────────────────────────┘
```

### 3.3 核心变更

#### 变更 1：Channel 加两个字段

```go
// model/channel.go
type Channel struct {
    // ... 现有字段不变

    StoreID       int    `json:"store_id" gorm:"index"`    // 新增：所属店铺
    IsPlatformPool bool  `json:"is_platform_pool"`          // 新增：平台兜底池标记
}
```

#### 变更 2：新建 Store / Listing / FeeConfig / Review

```go
// model/store.go — 店铺
type Store struct {
    ID        int       `json:"id" gorm:"primaryKey"`
    UserID    int       `json:"user_id" gorm:"uniqueIndex"`
    Name      string    `json:"name"`
    Tier      StoreTier `json:"tier" gorm:"default:'basic'"`
    Status    string    `json:"status" gorm:"default:'active'"`
    Rating    float64   `json:"rating"`
    Sales     int64     `json:"total_sales"`
    ...timestamps
}

// model/store.go — 上架
type Listing struct {
    ID            string  `json:"id" gorm:"primaryKey"`
    StoreID       int     `json:"store_id" gorm:"index"`
    ChannelID     int     `json:"channel_id"`     // 关联的执行渠道
    ModelName     string  `json:"model_name"`
    PricePerUnit  int64   `json:"price_per_unit"` // 分/千 token
    Region        string  `json:"region"`
    Status        string  `json:"status" gorm:"default:'active'"`
    ...stats
}

// model/fee_config.go — 阶梯入驻费配置
type PlatformFeeConfig struct {
    Tier    StoreTier `json:"tier" gorm:"primaryKey"`
    Rate    float64   `json:"rate"`    // 5/8/10%
    MinSkip int64     `json:"min_skip"`
    ...audit
}

// model/review.go — 评价
type Review struct {
    ID        int    `json:"id" gorm:"primaryKey"`
    ListingID string `json:"listing_id"`
    BuyerID   int    `json:"buyer_id"`
    Rating    int    `json:"rating"`  // 1-5
    Content   string `json:"content"`
}
```

#### 变更 3：入驻费计算改为按 Store 等级

```go
// 改后：查店铺等级 → 查阶梯费率 → 计算
func autoSettleMonthlyFees() {
    stores := getActiveStores()
    for _, s := range stores {
        revenue := sumStoreEarnings(s.ID, prevMonth)
        cfg := getFeeConfig(s.Tier)  // 从 PlatformFeeConfig 查
        if revenue >= cfg.MinSkip {
            fee := int64(float64(revenue) * cfg.Rate / 100)
            createFeeRecord(s.ID, s.UserID, cfg.Rate, fee)
            autoUpgradeTier(s, revenue)
        }
    }
}
```

#### 变更 4：路由层改为三层

```go
func selectChannel(modelName string, userID int) *Channel {
    // Layer 1: 第三方最低价
    if ch := getCheapestThirdParty(modelName); ch != nil {
        return ch
    }
    // Layer 2: 其他第三方（评分优先）
    if ch := getBestRatedThirdParty(modelName); ch != nil {
        return ch
    }
    // Layer 3: 平台兜底
    if ch := getPlatformPool(modelName); ch != nil {
        logWarn("市场资源紧缺，已启用兜底池")
        return ch
    }
    return nil // 503
}
```

### 3.4 路由中间件改造策略

不是重写 `Distribute`，而是**扩展**它：

```
当前：
  Distribute → GetCheapestSatisfiedChannel → setupAndContinue

改造后：
  Distribute → 先用市场路由（Listing 价格排序）
             → 市场无结果再用旧逻辑（Channel 兜底）
             → setupAndContinue
```

保持 `setupAndContinue` 不变，只改渠道选择那一段。

---

## 第四部分：任务分解与实施路径

### 4.1 依赖图

```
Phase 1: 基础设施
  Store 模型 ─────────┬──→ Listing 模型 ───────┬──→ 市场路由
                      │                        │
                      └──→ FeeConfig 模型 ──────┤──→ 阶梯入驻费
                                               │
                                               └──→ Review 模型
                                                        │
Phase 2: API 层                                      │
  开店 API ──────┬──→ 上架 API ───┬──→ 市场行情 API    │
                 │                │                   │
                 ├── 调价 API     ├── 评价 API ────────┘
                 │                │
                 └── 等级配置 API   └── 用户偏好 API
                                    │
Phase 3: 结算改造                     │
  入驻费阶梯 ────┬──→ cron 扩展 ──────┘
                │
                ├── 提现关联 Store
                │
                └── 自动升降级

Phase 4: 运营
  运营面板 ────┬──→ 店铺管理
              │
              └── 数据看板
```

### 4.2 四阶段任务分解

#### Phase 1: 店铺体系（Week 1）

| 任务 | 文件 | 行数 | 前置 |
|------|------|------|------|
| 1.1 Store 模型 + StoreTierLog | `model/store.go` | 80 | — |
| 1.2 PlatformFeeConfig 模型 | `model/fee_config.go` | 40 | — |
| 1.3 Review 模型 | `model/review.go` | 30 | — |
| 1.4 Channel 加字段 StoreID + IsPlatformPool | `model/channel.go` | 20 | — |
| 1.5 DB AutoMigrate + Seed 默认数据 | `model/main.go` | 30 | 1.1-1.4 |
| 1.6 Store 开店 API | `controller/store.go` | 150 | 1.1 |
| 1.7 Admin 入驻费配置 API | `controller/admin_store.go` | 120 | 1.2 |
| 1.8 `getPlatformFeeRate` → `getStoreFeeRate` | `model/platform_fee.go` | 30 | 1.2 |
| 1.9 store_service（开店+升级逻辑） | `service/store_service.go` | 100 | 1.1 |

#### Phase 2: 市场与上架（Week 2）

| 任务 | 文件 | 行数 | 前置 |
|------|------|------|------|
| 2.1 Listing 模型+CRUD | `model/store.go` | 60 | 1.1 |
| 2.2 上架/下架/调价 API | `controller/store.go` | 200 | 2.1 |
| 2.3 市场行情 API（搜索/比价/排序） | `controller/market.go` | 150 | 2.1 |
| 2.4 市场搜索服务 | `service/market_service.go` | 60 | 2.3 |
| 2.5 Listing→Channel 自动关联 | `service/store_service.go` | 60 | 2.1 |
| 2.6 用户偏好店铺 API | `controller/user.go` | 40 | 2.1 |
| 2.7 Listing 接入健康检测 | `controller/channel-test.go` | 40 | 2.1 |

#### Phase 3: 结算与路由（Week 3）

| 任务 | 文件 | 行数 | 前置 |
|------|------|------|------|
| 3.1 阶梯入驻费月结 | `model/platform_fee.go` | 60 | 1.8, 2.1 |
| 3.2 自动升降级 | `service/store_service.go` | 50 | 3.1 |
| 3.3 提现关联 Store | `model/commission.go` | 30 | 1.1 |
| 3.4 入驻费 cron 扩展 | `main.go` | 30 | 3.1 |
| 3.5 市场路由层（三层 spill-over） | `middleware/distributor.go` | 80 | 2.1 |
| 3.6 平台池标记 | `controller/channel-admin.go` | 30 | 1.4 |

#### Phase 4: 评价与运营（Week 4）

| 任务 | 文件 | 行数 | 前置 |
|------|------|------|------|
| 4.1 评价 API（Create/List） | `controller/market.go` | 60 | 1.3 |
| 4.2 店铺信誉分计算 | `service/market_service.go` | 40 | 4.1 |
| 4.3 搜索排序接入信誉分 | `service/market_service.go` | 30 | 4.2 |
| 4.4 运营指标 API | `controller/admin.go` | 80 | 全部 |
| 4.5 StoreTierLog 审计 | `controller/admin_store.go` | 40 | 3.2 |

### 4.3 总计

| 阶段 | 后端新增（行） | 后端改动（行） | 新增文件 |
|------|--------------|--------------|---------|
| Phase 1 | 520 | 50 | 4 |
| Phase 2 | 550 | 40 | 1 |
| Phase 3 | 280 | 80 | 0 |
| Phase 4 | 250 | 30 | 0 |
| **总计** | **~1600** | **~200** | **5** |

### 4.4 不需要碰的部分

| 模块 | 原因 |
|------|------|
| `relay/` 下的适配器代码 | 中继层完全不变，只改路由选择 |
| `contract/{topup,payment}*.go` 支付代码 | 支付链路完全不变 |
| `common/encrypt/` | 加解密逻辑不变 |
| `middleware/auth.go` + `middleware/ratelimit.go` | 认证中间件不变 |
| `i18n/` 翻译文件 | 后期再补 |
| 现有 `storeRoute`（已有店铺雏形） | 保留现有，新增市场 API |
| 现有 `Distribute` 路由逻辑 | 保留为兜底 fallback |

---

## 第五部分：关键设计决策

### 5.1 为什么 Store 不直接用 User，要建新表？

User 是认证实体（邮箱/密码/角色/状态），Store 是商业实体（等级/销售额/评分/开店时间）。一个 User 只能开一个 Store，但语义不同。

### 5.2 为什么 Listing 不直接复用 Channel？

Channel 是技术执行实体（API Key / BaseURL / Config），Listing 是商业展示实体（价格/评分/地区/上架时间）。一对一关联，但语义不同。

### 5.3 为什么路由不改死，保留旧逻辑？

改造期间需要兼容运行。新的市场路由优先，如果新的 Listing 路由无结果（比如还没有任何 Provider 开店），fallback 到旧的 Distribute 逻辑。保证任何时候都不断货。

### 5.4 入驻费为什么在提现时扣而不是实时扣？

提现时扣是"卖货不收费，拿钱出去再收"——友好且确保平台一定能收到（提现是平台控制的闸口）。实时扣的话，如果 Provider 月营业额很大但一直不提现，入驻费会一直挂账，平台收不到现金。

---

## 附录：当前文件索引

```
QuantumClaw
├── main.go                           # 入口 + cron
├── controller/
│   ├── user.go                       # 登录/注册/2FA
│   ├── relay.go                      # 中继主入口
│   ├── relay_text_helper.go          # 文本中继
│   ├── relay_image_helper.go         # 图片中继
│   ├── channel-test.go               # 渠道自测
│   ├── channel_upstream_update.go    # 上游检测
│   ├── task_controller.go            # 异步任务
│   ├── topup.go                      # 充值
│   ├── topup_alipay.go               # 支付宝支付
│   ├── topup_worldfirst.go           # 万里汇支付
│   ├── topup_creem.go / topup_waffo.go / topup_binance.go / topup_stripe.go / topup_epay.go
│   ├── option.go                     # 系统配置
│   ├── quantum.go                    # 量子计算
│   ├── fusion.go                     # 模型编排
│   ├── store.go                      # 店铺（雏形）
│   ├── admin.go                      # 管理后台
│   ├── password.go                   # 密码修改
│   ├── twofa.go                      # 2FA 管理
│   └── auth/                         # OAuth 登录
├── model/
│   ├── user.go                       # 用户模型
│   ├── channel.go                    # 渠道模型
│   ├── token.go                      # API Token
│   ├── topup.go                      # 充值/交易
│   ├── log.go                        # 使用日志
│   ├── option.go                     # 系统配置
│   ├── platform_fee.go               # 入驻费
│   ├── commission.go                 # 佣金/提现
│   ├── model_metadata.go             # 模型元数据
│   ├── provider_earning.go           # Provider 分账
│   ├── subscription.go               # 订阅
│   ├── store.go                      # 店铺模型（雏形）
│   ├── platform_config.go            # 平台配置
│   ├── cache.go                      # 内存缓存
│   └── billing_test.go              # 计费测试
├── middleware/
│   ├── auth.go                       # 认证中间件
│   ├── distributor.go                # 路由分发
│   └── ...                           # 限流/安全/CORS
├── relay/
│   ├── controller/                   # 中继业务
│   ├── common_handler/               # 预扣/后扣
│   ├── billing/                      # 计费钩子
│   └── adaptor/                      # 各 Provider 适配器
├── service/
│   ├── cash_billing.go               # 现金计费
│   ├── pre_consume_quota.go          # 预扣配额
│   ├── http_client.go                # 共享 HTTP 客户端
│   └── ...                           # 其他服务
├── common/
│   ├── encrypt/                      # 加解密
│   ├── config/                       # 全局配置
│   └── payment_setting.go            # 支付配置
├── router/
│   └── api.go                        # 所有路由注册
└── docs/
    └── marketplace-architecture-v3.md # 架构设计文档
```
