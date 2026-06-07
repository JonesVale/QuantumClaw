# 三方共赢模型 vs 自由市场模型 — 兼容性分析

> 分析日期：2026-06-06 | 不改代码，仅做逻辑校验

---

## 一、两个模型对比

### 模型 A：自由市场（店铺模式）

```
Provider 开店 → 自主定价 → 用户比价选购
                    │
              平台抽入驻费（5%~10%）
                    │
              平台不干预定价
```

### 模型 B：平台定价池（分销模式）

```
Provider A（只供资源，不开店）
  └── 平台定价售卖（A 不参与定价）
        │
        ├── User C（推广人）推广给 B → 拿佣金
        │     └── CommissionSetting.ConsumeRate
        │
        └── User B（消费者）以平台定价使用
              │
        平台向 A 结算（CostPrice）
```

---

## 二、现有代码能否同时支持两个模型

### 2.1 数据结构检查

| 概念 | 现有代码中的位置 | 模型 A | 模型 B |
|------|-----------------|--------|--------|
| Provider 自主定价 | `Listing.price_per_unit` | ✅ 用 | ❌ 不用 |
| 平台定价 | `Channel.SellPriceRate` | ❌ 不用 | ✅ 用 |
| 成本价（平台付给 Provider） | `Channel.CostPrice` | ❌ 不用 | ✅ 用 |
| 推广佣金 | `CommissionSetting` + `RewardInviterOnConsume` | ❌ 不用 | ✅ 用 |
| 推广关系 | `AffiliateRelation` + `PromoterId` | ❌ 不用 | ✅ 用 |
| 入驻费 | `PlatformFeeRecord` | ✅ 收 | ✅ 收（池内 Provider 也要交） |

**结论：数据结构已经同时支持两个模型，不需要新增字段。**

### 2.2 路由层检查

当前 `Distribute` 中间件的路由逻辑：

```
Step 0: 用户指定 channel_id → 用指定的
Step 1: 首选供应商（PreferredProviderId）
Step 2: 推广人渠道（promoterId）    ← 推广人 C 的渠道在这里被优先选中
Step 3: 国内池兜底
Step 4: 海外池兜底
```

关键点：**Step 2 已经是推广人的渠道优先路由。** 如果 C 推广了 B，C 的渠道会被优先选中。但 C 的渠道里放的是 C 自己配置的 API Key，不是 A 的。

要让 B 用 A 的资源同时 C 拿佣金，需要的是：
1. B 的请求路由到 A 的渠道（最优性价比）
2. 佣金归 C（因为 C 推广了 B）
3. A 拿到资源费用

当前 `RewardInviterOnConsume` 正是在消费后给推广人结算佣金。所以第 2 点和第 3 点已经实现了：
```go
// service/cash_billing.go: PostConsumeDeduct → RewardInviterOnConsume
// → 创建 CommissionRecord → 推广人 C 获得佣金
```

### 2.3 两个模型能否共存

```
同一平台内：

店铺 A（自主定价）      ← 模型 A
  └── gpt-4o 卖 ¥1.50
  └── 入驻费 10%

池 Provider B（平台定价）← 模型 B
  └── gpt-4o CostPrice = ¥0.80
  └── 平台售价 = ¥1.20
  └── B 不露面

推广人 C                ← 佣金 10%
  └── 推广给 D
  └── D 来平台消费 ¥1.20
  └── C 得佣金 ¥0.12
  └── B 得资源费 ¥0.80
  └── 平台得 ¥0.28（含支付成本+入驻费+利润）
```

两个模型不冲突。他们服务的 Provider 类型不同：
- **模型 A**：Provider 想开店卖货，自主定价 → 店铺模式
- **模型 B**：Provider 只贡献资源，不参与经营 → 池模式（平台包办定价、运营、推广）

### 2.4 当前代码的唯一缺口

模型 B 中，平台需要为池内的资源**统一定价**。当前 `Channel` 的定价字段是：

| 字段 | 含义 |
|------|------|
| `CostPrice` | 平台支付给 Provider 的成本价 |
| `SellPriceRate` | 销售加价倍率 |
| `ChannelMarkup` | 渠道加价倍率 |

对于模型 B，平台需要确保池资源的 `SellPriceRate` 和 `ChannelMarkup` 被统一管理（不能由 Provider 自己改），而是由管理员在控制台统一配置。

当前代码已经有管理员控制台可以修改 Channel 的这些字段，所以操作层面没问题——只是需要运营规范：池资源的定价由管理员维护，不开放给 Provider 自改。

---

## 三、完整的三方共赢链路（无需改代码）

```
A（资源提供方）       平台                 C（推广人）
  │                   │                     │
  │──提供 API Key────►│                     │
  │  CostPrice=0.80   │                     │
  │                   │  ◄── 生成推广链接 ──│
  │                   │                     │
  │                   │  ── 推广链接给 ────►│
  │                   │                     │── 给 B
  │                   │                     │
  │                   │◄── B 注册/消费 ─────│
  │                   │     PromoterId = C  │
  │                   │                     │
  │── 扣 CostPrice ──│                     │
  │                   │── Commission ──────►│
  │                   │                     │
  B 以平台定价 ¥1.20 使用 A 的资源
  A 得 ¥0.80（CostPrice）
  C 得 ¥0.12（Commission 10%）
  平台得 ¥0.28（Cover 成本 + 利润）
```

当前代码中：
- `Channel.CostPrice` → A 的收入 ✅
- `RewardInviterOnConsume` → C 的佣金 ✅
- `Distribute Step 2` → 推广人渠道优先 ✅
- `PlatformFeeRecord` → 入驻费 ✅

---

## 四、结论

**两个模型没有冲突，可以并行运行在同一个平台上。**

| 维度 | 自由市场（店铺模式） | 平台定价（池模式） |
|------|-------------------|------------------|
| 定价者 | Provider | 平台管理员 |
| 入驻费 | ✅ 按月收 | ✅ 按月收 |
| 推广佣金 | ❌ 通常不涉及 | ✅ 核心玩法 |
| 用户感知 | 看到店铺名 | 直接使用服务 |
| 数据模型 | Store + Listing | Channel（统一管理） |
| 路由 | 按 Listing.price 排序 | 按 SellPriceRate 排序 |

当前代码两个模式都支持，不需要改动即可同时运行。需要明确的是**运营规范**：池资源的定价由管理员统一维护，不开放给池 Provider 自改。
