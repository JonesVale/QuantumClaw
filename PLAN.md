# QuantumClaw 业务关联性分析与修复规划

## 一、核心用户流程及关联性

```
    注册 ───→ 登录 ───→ 使用API ───→ 消费扣款
      │                  ↑                │
      │                  │                ├──→ 返佣给邀请人
      │                  │                ├──→ 渠道分账
      │                  │                └──→ 现金扣款
      │                  │
      ├──→ 充值/订阅 ────┘
      ├──→ 签到
      │
      ├──→ 升级供应商 ──→ 配置渠道 ──→ 自己的API Key
      ├──→ 推广 ──→ 邀请注册 ──→ 返佣 ──→ 提现
      └──→ 团队 ──→ 组织管理
```

### 核心关联分组

| 关联组 | 功能模块 | 共享数据 | 风险传染链 |
|--------|---------|---------|-----------|
| **A组：认证体系** | 注册→登录→密码 | User.Password | CryptoSecret 毁掉所有密码 → 所有人都登录不了 |
| **B组：计费体系** | 充值→消费→返佣→提现 | User.Quota/CashBalance/Quota | 充值不成功→消费不扣→返佣断裂→供应商无营收 |
| **C组：API中继** | 渠道配置→Key加密→API转发→消费 | Channel.Key/Ability | 渠道Key加密失败→API转发中断→整个服务不可用 |
| **D组：关键配置** | .env / DB / Session | CryptoSecret/SessionSecret | 配置丢失→认证失效→加密失效 |
| **E组：供应商链** | 升级→渠道→Key→分账 | User.UserType/Channel.UserId | 审核缺失→恶意供应商→平台信用崩塌 |
| **F组：推广体系** | 邀请→返佣→团队→组织 | InviterId/AffCode/AffiliateRelation | 返佣不准→推广体系信任崩塌 |

---

## 二、按关联组分析的详细问题

### A组：认证体系（最高优先级 — 直接影响用户能否使用）

| 问题 | 严重度 | 影响范围 | 逻辑位置 |
|------|:------:|---------|---------|
| CryptoSecret 时序 bug | **致命** | 所有用户密码 | config.go init |
| err:= 变量遮蔽 | **致命** | email登录失败 | model/user.go:279 |
| 错误信息混淆 | **高** | 用户体验 | model/user.go ValidateAndFill |
| 密码强度要求过高(8+大写+数字+特殊) | **中** | 注册转化率 | controller/user.go Register |
| 无 JWT，纯 session | **中** | 多机部署/API调用 | middleware/auth.go |
| LoginRateLimit 3次锁24h | **中** | 用户体验 | middleware/login_rate_limit.go |
| 重置密码流程改进 | **已修** | — | — |

### B组：计费体系（次高优先级 — 影响变现）

| 问题 | 严重度 | 影响范围 | 逻辑位置 |
|------|:------:|---------|---------|
| 新用户注册后 quota=0 无提示 | **高** | 新用户转化率 | controller/user.go Insert |
| 充值+订阅扣费优先级不清 | **高** | 计费准确性 | relay/billing/billing.go |
| 订阅价格用 float64 | **中** | 精度风险 | model/subscription.go |
| 提现流程纯手动审核 | **中** | 运营效率 | controller/commission.go |
| 签到奖励偏低(¥0.002~¥0.02) | **低** | 用户激励 | operation_setting/checkin_setting.go |

### C组：API中继（高优先级 — 影响服务可用性）

| 问题 | 严重度 | 影响范围 | 逻辑位置 |
|------|:------:|---------|---------|
| Distributor.ApiKey 明文存储 | **高** | 安全 | model/distributor.go |
| Channel key 加密用 CryptoSecret | **致命已修** | 渠道Key | common/config.go |
| Channel Type 与 model_metadata provider 不匹配 | **中** | 前端模型列表为空 | controller/channel.go GetChannelTypes |

### D组：关键配置（基础风险）

| 问题 | 严重度 | 影响范围 | 逻辑位置 |
|------|:------:|---------|---------|
| 多个 var 用 `os.Getenv()` 在包初始化时执行 | **中** | 配置加载 | common/config.go 多个 var |
| SessionSecret = uuid.New() 每次启动不同 | **中** | session 持久化 | common/config.go |
| .gitignore 未包含 quantumclaw.db | **低** | 数据库泄露 | .gitignore |

### E组：供应商链（影响平台安全）

| 问题 | 严重度 | 影响范围 | 逻辑位置 |
|------|:------:|---------|---------|
| 供应商升级无审核 | **高** | 所有用户 | controller/user.go UpgradeToProvider |
| 供应商无降级机制 | **中** | 管理灵活性 | controller/user.go |
| `role = 1000` 硬编码 | **低** | 维护性 | controller/distributor.go CreateDistributor |

### F组：推广体系（影响获客）

| 问题 | 严重度 | 影响范围 | 逻辑位置 |
|------|:------:|---------|---------|
| AffCode 4位太短 | **高** | 注册稳定性 | model/user.go Insert |
| 仅一级返佣 | **中** | 推广深度 | model/commission.go RewardInviterOnConsume |
| 返佣直接用 quota 发(非独立佣金池) | **中** | 资金风控 | model/commission.go |
| 推广团队 vs 组织管理 概念混淆 | **中** | 用户体验 | team.tsx |

---

## 三、修复优先级分组（按业务流程相关性）

### P0：立即可做、影响面独立

这些任务不依赖其他模块修改，可以直接修：

| 任务 | 关联组 | 估算 | 说明 |
|------|:-----:|:----:|------|
| 1. AffCode 改为 8 位 | F | 0.5h | `random.GetRandomString(4)` → `(8)` |
| 2. `role = 1000` 硬编码改为常量 | E | 0.5h | `const RoleDistributor = 1000` |
| 3. 供应商升级加管理员审核锁 | E | 1h | `UpgradeToProvider` 改为 `status=pending`，管理员审批 |
| 4. Distributor.ApiKey 用 AES 加密 | C | 0.5h | 复用 `encrypt.Encrypt` |
| 5. SessionSecret 持久化(存DB/文件) | D | 0.5h | `init()` 尝试从 `.session_secret` 文件读取 |
| 6. `.gitignore` 补充 | D | 0.2h | 追加 `*.db *.db-journal` |

### P1：需要按正确顺序做

这些任务的修复顺序有依赖关系：

| 任务 | 前置依赖 | 说明 |
|------|---------|------|
| **注册体验改善** | — | 注册后给试用额度 + 提示，让新用户知道下一步 |
| **充值流程完善** | 注册体验 | 前端清晰展示充值入口 |
| **消费扣费优先级** | 已修 | 明确 `quota → subscription → cash_balance` 顺序 |
| **签到优化** | — | 增签到奖励至有吸引力的范围 |
| **返佣体系** | 消费扣费 | 提现审核改为阈值内自动放行 + 超额审核 |

### P2：非紧急但重要的长期改善

| 任务 | 估算 | 说明 |
|------|:----:|------|
| LoginRateLimit 逐级递增 | 1h | 3次→5min, 5次→30min, 10次→24h |
| 密码强度降低或可配置 | 0.5h | 改为8位包含任意两类 |
| 一级→多级返佣 | 3h | 支持2-3级代理链 |
| JWT 支持 | 8h | 大改动，双auth模式 |

---

## 四、修复实施规划（按批次）

### Batch 1：立即修复（今天做完）
1. `AffCode 4位→8位` — 4行改动
2. `role = 1000` 常量化
3. `Distributor.ApiKey` AES加密
4. `SessionSecret` 持久化到文件
5. `.gitignore` 补充

### Batch 2：连锁修复（明天）
1. 供应商升级加审核 + 管理员审批页面
2. 新用户注册体验改善（试用额度 + 引导）
3. 签到奖励翻倍

### Batch 3：长期规划
1. 计费优先级重构
2. 登录安全增强（逐级递增锁）
3. 多级返佣
4. JWT
