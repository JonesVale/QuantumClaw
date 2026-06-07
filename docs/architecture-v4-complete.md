# QuantumClaw 整体架构 v4 — 完整分析

> 版本: v4 | 2026-06-06 | 覆盖全部已有模块 + 新增市场体系

---

## 一、角色模型

```
                         ┌──────────────────┐
                         │     平台运营       │
                         │  (管理员 ≥ 100)   │
                         │                  │
                         │ · 定价池配置       │
                         │ · 入驻费阶梯配置    │
                         │ · 店铺管理        │
                         │ · 协议发布        │
                         └──────────────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        ▼                      ▼                      ▼
┌──────────────┐    ┌──────────────────┐    ┌──────────────────┐
│  资源提供方(A) │    │   推广人 (C)      │    │  消费者 (B)       │
│  (Provider)   │    │  (Promoter)      │    │  (Consumer)      │
├──────────────┤    ├──────────────────┤    ├──────────────────┤
│ 两种模式:     │    │ · 生成推广链接     │    │ · 选择资源        │
│              │    │ · 赚取佣金        │    │ · 选择是否同意池   │
│ ① 开店自定价  │    │ · 查看推广记录    │    │ · 使用资源        │
│ ② 入池平台定价 │    │                  │    │ · 查看消费记录    │
└──────────────┘    └──────────────────┘    └──────────────────┘
```

### 三者的互动关系

```
A（供资源）                   平台                    C（推广）
  │                          │                        │
  │── 开店/入池 ───────────►│  ◄── 生成推广链接 ─────│
  │  ① 开店: Store + Listing │                        │
  │  ② 入池: Channel(平台定价)│                        │
  │                          │── 推广链接 → B ────────│
  │                          │                        │
  │◄── CostPrice 结算 ──────│                        │
  │                          │── Commission ─────────►│
  │                          │                        │
  │                    B 使用资源（平台价/店定价）         │
```

---

## 二、两种资源供给模式并存

### 模式 A：店铺模式（自由市场）

```
Provider 开店 → 自主定价 → 用户比价选购
                  │
            平台抽入驻费（按店铺等级）
                  │
            basic 10% / gold 8% / flagship 5%
```

**适用场景：** 想经营品牌、自主定价、积累信誉的 Provider

**数据链路：**
```
Store（店铺）→ Listing（上架）→ price_per_unit（自主定价）
  → 市场行情 API
  → 用户选购
  → Distribute 路由（待接入 Listing 排序）
  → 计费 → Provider 收益 → 入驻费结算 → 提现
```

### 模式 B：池模式（平台定价）

```
Provider 只提供资源 → 平台统一定价
                  │
            平台设 SellPriceRate
                  │
            Provider 拿 CostPrice
```

**适用场景：** 只贡献资源、不参与经营的 Provider（如跨地区资源提供商）

**数据链路：**
```
Channel（平台配置）→ CostPrice（Provider 收入）
  → SellPriceRate（平台售价）
  → Distribute 路由
  → 计费 → Provider 收益 → 入驻费结算 → 提现
```

---

## 三、路由策略

```
用户请求 model = "gpt-4o"
  │
  ├── 指定了 store? → 用该 store 的 listing → 失败 → 有协议 → 池路由
  │                                                        │ 无协议 → 报错
  │
  └── 未指定 → Distribute 四步路由：
        │
        ├── Step 0: 用户指定 channel_id → 直接路由
        │
        ├── Step 1: 首选供应商（PreferredProviderId）
        │     └── 跳过: IsModelPenalized(ch, model)
        │
        ├── Step 2: 推广人渠道（promoterId）
        │     └── 跳过: IsModelPenalized(ch, model)
        │
        ├── Step 3: 国内池兜底
        │     └── 跳过: IsModelPenalized(ch, model)
        │
        ├── Step 4: 海外池兜底
        │     └── 跳过: IsModelPenalized(ch, model)
        │
        └── 全部不可用 → 503

失败 → 用户已同意平台池协议？
  ├── 是 → 重试循环（跳过冷期 + 已失败）
  │         └── 重试成功 → PenalizeModel(失败渠道, 模型)
  │                        进入 30 分钟冷期
  │
  └── 否 → 直接返回错误（不重试）
```

---

## 四、计费链路

```
用户请求 → 计费
  │
  ├── PreConsumeBilling
  │     ├── 先扣订阅剩余（Subscription）
  │     ├── 再扣现金余额（Cash Balance）
  │     ├── 再扣佣金余额（Commission Balance）
  │     ├── 最后扣配额（Quota）
  │     └── 全部不足 → 记录债务（Debt）
  │
  ├── DoRequest → 上游 API 调用
  │     ├── 成功 → DoResponse → PostConsumeDeduct
  │     ├── 失败 → 返回 error（预扣不退）
  │     └── 重试成功 → 取新渠道 PreConsume → DoRequest → ...
  │
  ├── PostConsumeDeduct（三级扣款）
  │     ├── Tier 1: 现金余额（原子 UPDATE WITH WHERE）
  │     ├── Tier 2: 佣金余额
  │     ├── Tier 3: 配额回退
  │     └── Fallback: 挂账 debt
  │
  ├── PostConsumeBilling（退还不准确预扣）
  │     └── quotaDelta = pre - final
  │           ├── positive → 退还用户
  │           └── negative → 补充扣款
  │
  ├── RewardInviterOnConsume（推广佣金）
  │     └── 创建 CommissionRecord
  │
  └── RecordConsume（消费明细）
        └── 写入 consume_records 表
```

---

## 五、推广佣金链路

```
用户 B 注册时带有 promoterId = C
  ├── 系统创建 AffiliateRelation(B, C)
  │
  ├── B 每次消费 → RewardInviterOnConsume(B, amount)
  │     ├── 查 CommissionSetting（费率）
  │     ├── 查 AffiliateRelation（推广人）
  │     ├── 计算 commission = amount × rate
  │     └── 创建 CommissionRecord(C, amount)
  │
  ├── C 查看收益 → GET /commission/self/records
  │
  └── C 提现 → POST /commission/self/withdraw
        └── 自动扣除 pending 入驻费
```

---

## 六、入驻费链路

```
每月 1 号 2:00 AM
  │
  ├── 遍历所有活跃 Store
  ├── 查上月营业额
  │     ├── < ¥100 → 跳过（skipped）
  │     └── ≥ ¥100 → 查阶梯费率
  │           ├── basic    10%
  │           ├── gold     8%
  │           └── flagship 5%
  ├── 创建 PlatformFeeRecord（status = pending）
  │
  ├── 自动升降级检查
  │     ├── 累计销售额 ≥ ¥10,000 → 升 flagship
  │     ├── 累计销售额 ≥ ¥1,000  → 升 gold
  │     └── 3 个月无交易 → 降 basic
  │
  └── 记录 StoreTierLog

Provider 提现时：
  ├── 查询 pending 入驻费
  ├── net = amount - pending
  ├── 标记入驻费为 deducted
  └── 打款
```

---

## 七、冷期（惩罚）机制

```
Relay 重试成功时
  │
  └── PenalizeModel(失败渠道ID, 模型名)
        └── (channelId, modelName) 进入 30 分钟冷期
              │
              ├── 影响该渠道该模型（不影响其他模型）
              ├── 重试循环跳过冷期组合
              └── 新请求路由跳过冷期组合
```

---

## 八、平台池协议

```
用户首次请求 / 协议更新时
  │
  ├── 前端展示协议内容
  ├── 用户选择同意/拒绝
  │     ├── 同意 → UpsertUserPoolConsent(userId, agreed=true)
  │     │          下次请求 relay → IsConsentValid() → true
  │     │          → 失败可重试池内其他资源
  │     └── 拒绝 → UpsertUserPoolConsent(userId, agreed=false)
  │                下次请求 relay → IsConsentValid() → false
  │                → 失败不重试，直接返回错误
  │
  ├── 协议更新后 → 用户需重新同意
  │     └── isConsentValid 检查 agreed_version ≥ latest_version
  │
  └── 管理员发布新协议 → 版本号自动递增
```

---

## 九、安全防护体系

```
┌─────────────────────────────────────────────────────┐
│ 数据库层                                              │
│  ├── 连接池: MaxOpenConns=25, MaxIdleConns=10        │
│  └── ConnMaxLifetime=5m                              │
├─────────────────────────────────────────────────────┤
│ 加密层                                                │
│  ├── API Key 存储: AES-256-GCM + Qpw 后缀            │
│  ├── 密码存储: bcrypt + Qpw 后缀 + AES-256-GCM       │
│  ├── Token 存储: SHA-256 哈希 + AES-256-GCM           │
│  └── CRYPTO_SECRET 配置: 启动时派生 AES 密钥          │
├─────────────────────────────────────────────────────┤
│ Webhook 层                                            │
│  ├── 幂等键: PaymentIdempotencyKey（唯一约束防止双花）  │
│  ├── IP 白名单: WebhookIPWhitelist 中间件              │
│  ├── RSA2 验签: 支付宝异步通知验签                     │
│  └── HMAC 验签: WorldFirst webhook                    │
├─────────────────────────────────────────────────────┤
│ 路由层                                                │
│  ├── 冷期: 30 分钟 (渠道, 模型) 惩罚                   │
│  ├── 自动禁用: ShouldDisableChannel（401/quota/403）   │
│  ├── 自动恢复: ShouldEnableChannel（测试通过）          │
│  └── 健康检测: 定时测试所有渠道                         │
├─────────────────────────────────────────────────────┤
│ 应用层                                                │
│  ├── Goroutine 保护: safeGoWithRestart（panic 恢复）   │
│  ├── Token 认证: TokenAuth 中间件                      │
│  ├── 管理员认证: RootAuth（根用户）/ AdminAuth          │
│  ├── Session: Cookie Store / Redis Store             │
│  ├── 限流: RateLimit 中间件                            │
│  └── 安全启动审计: SESSION_SECRET / 密码检查            │
└─────────────────────────────────────────────────────┘
```

---

## 十、所有业务链路一览

| # | 链路 | 当前状态 | 关键文件 |
|---|------|---------|---------|
| 1 | 用户注册→认证→会话 | ✅ | user.go, auth/ |
| 2 | 渠道→API Key→加密存储 | ✅ | channel.go, encrypt.go |
| 3 | 充值→资金流入 | ✅ | topup*.go, payment_setting.go |
| 4 | 路由→中继→计费（核心） | ✅ | relay.go, distributor.go, billing.go |
| 5 | 回调→充值完成→债务抵扣 | ✅ | topup.go, CompleteTopUp |
| 6 | 异步任务→扣费 | ✅ | task_controller.go, DeductTaskQuota |
| 7 | Provider收益→入驻费→提现 | ✅ | platform_fee.go, commission.go |
| 8 | 订阅计费 | ✅ | subscription.go, PreConsumeBilling |
| 9 | 量子计算 | ✅ | quantum.go, PostConsumeQuantumDeduct |
| 10 | 店铺体系（开店/上架/调价） | ✅ | store.go, market.go |
| 11 | 市场行情与比价 | ✅ | market.go |
| 12 | 入驻费阶梯配置（管理端） | ✅ | admin_store.go, fee_config.go |
| 13 | 推广佣金 | ✅ | commission.go, RewardInviterOnConsume |
| 14 | 推广人路由优先 | ✅ | distributor.go（Step 2） |
| 15 | 冷期机制（渠道+模型） | ✅ | penalty.go, relay.go |
| 16 | 平台池协议（用户同意） | ✅ | pool_agreement.go |
| 17 | 消费明细记录 | ✅ | consume_record.go |
| 18 | Webhook 幂等键 | ✅ | idempotency.go |
| 19 | 健康检查 /health | ✅ | health.go |
| 20 | 数据库连接池 + goroutine保护 | ✅ | main.go |

---

## 十一、数据模型总览

```
users                          ← 用户（消费者/推广人/Provider）
  ├── stores                   ← 店铺（Provider 开店）
  │     ├── listings           ← 上架资源（自主定价）
  │     ├── reviews            ← 评价
  │     ├── store_tier_logs    ← 等级变更日志
  │     └── platform_fee_records ← 入驻费记录
  │
  ├── channels                 ← 渠道（API Key 加密存储）
  │     ├── provider_earnings  ← Provider 分账
  │     └── logs               ← 使用日志
  │
  ├── tokens                   ← API Token
  ├── top_ups                  ← 充值/交易
  ├── consume_records          ← 消费明细
  ├── balance_logs             ← 余额流水
  ├── withdrawal_requests      ← 提现
  ├── commission_records       ← 推广佣金
  ├── affiliate_relations      ← 推广关系
  ├── user_pool_consents       ← 平台池协议同意
  ├── subscriptions            ← 订阅
  └── async_tasks              ← 异步任务

platform_configs               ← 平台全局配置
platform_fee_configs           ← 阶梯入驻费配置
platform_pool_agreements       ← 平台池协议版本

payment_idempotency_keys       ← Webhook 幂等键
```

---

## 十二、当前存在的已知缺口（不改代码）

| 缺口 | 影响 | 触发条件 |
|------|------|---------|
| 路由排序依据 `sell_price_rate`，用户看到的是 `listing.price_per_unit` | 用户预期的最低价比路由实际选的不一致 | 前端市场页面上线后 |
| 预扣在失败时不退还（累积到重试成功时） | 用户可能被多扣少量配额 | 高并发重试场景 |
| `RewardInviterOnConsume` 只走 `commission` 计费源，不走现金 | 池模式 Provider 用现金结算时，佣金可能漏记 | 池模式推广场景 |
