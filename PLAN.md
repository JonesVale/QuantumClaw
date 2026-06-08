# QuantumClaw 全量任务规划

## 项目总览

| 维度 | 现状 |
|------|------|
| Go 后端路由 | 300+ 端点，27 个路由组 |
| 前端页面 | 58 个 TSX，含 2 层布局 |
| Controller | 30+ 文件，60+ Controller 函数 |
| 支付渠道 | 7 种：Stripe/Epay/Creem/Waffo/Binance/Alipay/WorldFirst |
| OAuth 登录 | 7 种：GitHub/Discord/LinuxDo/WeChat/Lark/OIDC/Telegram |
| DB 数据模型 | 30+ GORM Model |
| 已完成修复 | 10 commits: 认证修复 + Channel 表单 + 密码重置 + Batch1 基础设施 |

---

## 第一阶段 — 认证/安全完善（已完成）

| # | 任务 | 状态 | commit |
|:-:|------|:----:|--------|
| 1 | CryptoSecret 加载时序修复 | ✅ | `04c78ba` `5fb35ca` |
| 2 | `err:=` 变量遮蔽修复 | ✅ | `04c78ba` |
| 3 | 错误信息三路分离 | ✅ | `04c78ba` |
| 4 | 登录页"忘记密码"入口 + 邮箱重置流程 | ✅ | `c11b3bf` `c6efa87` |
| 5 | Admin Users 页 Reset Password | ✅ | `5fb35ca` `c6efa87` |
| 6 | 重置密码自定义 `new_password` | ✅ | `c6efa87` |
| 7 | 重置密码前台页面 `reset-password.tsx` | ✅ | `c6efa87` |
| 8 | resetUserPassword API URL 修正 | ✅ | `c6efa87` |
| 9 | AffCode 4→8 位 | ✅ | `a8d5d90` |
| 10 | RoleDistributor 常量化 | ✅ | `a8d5d90` |
| 11 | SessionSecret 持久化到 `.session_secret` | ✅ | `a8d5d90` |
| 12 | `.gitignore` 补充 | ✅ | `a8d5d90` |
| 13 | Channel Type 自动填充 URL + Models | ✅ | `63b2d59` `36bb072` |

---

## 第二阶段 — 用户体验改善（已完成）

### Batch 2A：登录/注册体验

#### 2A-1：LoginRateLimit 逐级递增（1h）
- **位置**：`middleware/login_rate_limit.go`
- **现状**：3 次连续失败 → 锁 24h
- **目标**：3 次→5min, 5 次→30min, 10 次→24h
- **改动**：修改 `loginRateLimiter` 的 `lockedUntil` 计算逻辑，按 `consecutiveFails` 阶梯计算锁定期
- **文件**：`middleware/login_rate_limit.go`

#### 2A-2：密码强度可配置（1h）
- **位置**：`controller/user.go:ValidatePasswordStrength`
- **现状**：硬编码 8 位 + 大写 + 数字 + 特殊字符
- **目标**：改为 config 驱动，`config.PasswordMinLength` / `config.PasswordRequireUpper` / `config.PasswordRequireNumber` / `config.PasswordRequireSpecial`，前端注册页显示强度要求
- **文件**：`common/config/config.go`, `controller/user.go`, `web/.../sign-in.tsx`

#### 2A-3：新用户注册引导 + 试用额度 ✅（`8542fc2`）
- **位置**：`model/user.go:Insert`, 前端 pages
- **现状**：注册后赠送 QuotaForNewUser=50000 + NewUserTrialBalance=¥50
- **改动**：
  - 启用 `config.QuotaForNewUser=50000`, `QuotaForInviter=10000`, `QuotaForInvitee=5000`
  - `SetupLogin` 新增 `quota_for_new_user` / `trial_balance` 字段
  - Dashboard 新增新手引导卡片（零用量时显示，含 Playground/Key/定价入口）

### Batch 2B：计费/充值完善

#### 2B-1：充值扣费优先级确认 + 文档化 ✅（`94a0cfd`）
- **位置**：`relay/common_handler/billing.go`
- **现状**：三级优先级链已实现：订阅 → CashBalance → CommissionBalance → Quota
- **改动**：PreConsumeBilling / PostConsumeBilling 统一入口

#### 2B-2：签到奖励提升 ✅（`527a048` Min 1000→10000, Max 10000→50000）

#### 2B-3：提现自动审批 ✅（`9109acb`）
- **位置**：`model/commission.go:CreateWithdrawal`
- **现状**：金额 ≤ AutoWithdraw（默认¥100）→ 自动审批通过 + 即时扣除佣金余额

### Batch 2C：供应商通道完善

#### 2C-1：供应商升级验证 ✅（已修正）
- **位置**：`controller/user.go:UpgradeToProvider`
- **现状**：用户添加至少一个 API 渠道后即可立即升级为供应商（无需审核）
- **身份审核**：提现时检查 `IdentityVerified` 字段，必须完成实名认证才能提现
- **改动**：
  - 升级仅检查是否有 Channel（不检查 Token/API Key）
  - 直接设置 `user_type=provider`, `role=supplier`，即时生效
  - 提现前强制身份信息审核（见 `withdrawal.go`）

#### 2C-2：Channel Type 模型列表修复 ✅（`7e0cc8d`）
- **位置**：`controller/channel.go:GetChannelTypes`
- **现状**：名称映射表桥接 `ChannelTypeNames` 显示名 → `model_metadata.Provider` 技术名
- **改动**：
  - 新增 `channeltype.ChannelTypeNameToProvider` 映射表（70+ 条目）
  - 例如：Google Gemini→Google, Ali (Qwen)→Alibaba, Mistral AI→Mistral

---

## 第三阶段 — 安全加固

### Batch 3A：数据安全

#### 3A-1：Distributor.ApiKey AES 加密（0.5h）
- **位置**：`model/distributor.go:Distributor` + `controller/distributor.go:CreateDistributor`
- **现状**：明文 32 位随机串
- **目标**：复用 `common/encrypt` 包的 AES-GCM 加密，和 Channel.Key 加密方案一致
- **文件**：`model/distributor.go`, `controller/distributor.go`

#### 3A-2：SubscriptionPlan 价格 int64 化 ✅（`b35bdef`）
- **位置**：`model/subscription.go:SubscriptionPlan`
- **现状**：`PriceAmount float64` → `PriceCents int64`（单位：美分）
- **改动**：数据库字段 price_amount→price_cents, JSON 同步更改, 验证上界从 9999 改为 999900

### Batch 3B：认证增强 ✅（已验证）

#### 3B-1：WebAuthn/Passkey 完善（如已启用只需验证）
- **位置**：已有 WebAuthn 注册/登录路由
- **现状**：4 个 WebAuthn 端点（register_begin/finish, login_begin/finish）
- **目标**：验证 Passkey 登录流程是否完整

#### 3B-2：Two-Factor Auth (2FA) 支持（已有实现）
- **位置**：`controller/twofa.go` + 前端 `password.tsx`
- **现状**：双因素认证已有 TOTP 实现
- **目标**：验证流程完整性

---

## 第四阶段 — 计费体系重构

### Batch 4A：佣金/返佣体系

#### 4A-1：佣金独立池 ✅（`2c84170` + `3cce18d` 修正可消费）
- **位置**：`model/user.go:CommissionBalance`
- **现状**：返佣计入独立 CommissionBalance，既可消费也可提现
- **消费优先级**：订阅 → CashBalance → CommissionBalance → Quota
- **提现**：ApproveWithdrawal 从 CommissionBalance 扣除

#### 4A-2：多级返佣 ✅（`c33350c`）
- **位置**：`model/commission.go:RewardInviterOnConsume`
- **现状**：递归查找 3 级邀请人，1级5% / 2级2% / 3级1%
- **改动**：CommissionSetting 新增 Level2Rate/Level3Rate, CommissionRecord 新增 Level 字段

### Batch 4B：消费/结算

#### 4B-1：供应商分账系统优化 ✅（已实现，含前端 UI）
- **位置**：`service/cash_billing.go:PostConsumeDeduct` + 前端 `channels.tsx`
- **现状**：`channel.ProfitSplit` 字段已存在（model/channel.go），`PostConsumeDeduct` 已使用分账逻辑
- **已完成**：
  - 渠道级 `profit_split` 分账比例（后端 ✅ + 前端表单输入框 ✅ + 表格展示 ✅）
  - 前端在 Channel 编辑/创建表单中可设置 Profit Split（0-1），表格显示为百分比
- **文件**：`model/channel.go`, `service/cash_billing.go`, `api-extended.ts`, `channels.tsx`, `zh-CN.json`, `en.json`

#### 4B-2：异常结算处理 ✅（已实现）
- **位置**：`service/cash_billing.go:recordDebt` + `model/settlement_hourly.go`
- **已完成**：
  - 余额不足但请求已转发完成 → 记 `Debt` + `DebtSince` ✅
  - 下次充值自动扣还 ✅（`model/settlement_hourly.go:CalculateAndRecoverDebt`）
  - 长期不还 → 自动禁用账号 ✅（`CheckSuspendedDebtUsers` 已接入 `RunHourlySettlement`）
- **文件**：`model/user.go`, `service/cash_billing.go`, `model/settlement_hourly.go`, `model/main.go`

---

## 第五阶段 — 长期架构改善

### Batch 5A：核心技术

| # | 任务 | 工时 | 优先级 | 说明 |
|:-:|------|:----:|:------:|------|
| 5A-1 | JWT Token 鉴权 | 8h | 中 | 支持双模式（session + JWT），多机部署必须 | ✅
| 5A-2 | Redis Session 共享 | 3h | 中 | ✅ 已实现（main.go:461-475），增加错误处理 + 告警日志 |
| 5A-3 | 计费流水审计日志 | 4h | 低 | 余额变更全量日志，支持回滚 | ✅
| 5A-4 | SQLite → MySQL 迁移文档 | 2h | 低 | 已有双 DB 支持但迁移流程不清晰 | ✅

### Batch 5B：运维/部署

| # | 任务 | 工时 | 优先级 |
|:-:|------|:----:|:------:|
| 5B-1 | Docker Compose 一键部署 | 2h | 低 | ✅
| 5B-2 | nginx 配置示例 | 1h | 低 | ✅
| 5B-3 | 备份/恢复脚本（含密钥） | 2h | 中 | ✅
| 5B-4 | 健康检查端点 + Prometheus 集成 | 3h | 低 | ✅

---

## 执行优先级矩阵

```
               高影响                 中影响                低影响
            ┌────────────────────────────────────────────────────┐
 容易  │  ✅ 2A-1: 登录逐级锁     ✅ 2B-3: 提现自动审批    ✅ 2B-2: 签到奖励
       │  ✅ 2A-3: 新用户引导     ✅ 2C-2: 模型列表修复
       │  ✅ 3A-1: Dist加密        ✅ 2B-1: 扣费优先级
       │                         ✅ 2A-2: 密码强度可配置
       ├────────────────────────────────────────────────────┤
 中等  │  ✅ 4A-1: 佣金独立池     ✅ 3A-2: 价格int64化     ✅ 5A-2: Redis
       │  ✅ 2C-1: 供应商审批     ✅ 4B-1: 分账优化
       │  ✅ 4A-2: 多级返佣       ✅ 4B-2: 异常结算
       ├────────────────────────────────────────────────────┤
 困难  │  ❌ 5A-1: JWT            ✅ Batch D: 前端高价       ❌ 5B-x: 运维
       │                         ❌ 5A-3: 审计日志
       └────────────────────────────────────────────────────┘
```

---

## 建议执行顺序

```
✅ 已完成
  ├── 4B-1: 分账比例配置化 (渠道级 profit_split + 前端 UI)
  ├── 4B-2: 异常结算处理 (Debt 追偿 + 自动封号)
  ├── 5A-2: Redis Session 共享 (多机部署 + 错误告警)
  └── Batch D: 前端高价提示 (Playground 价格对比 + 会话费用追踪)

已完成（本轮）
  ├── 5A-1: JWT Token 鉴权 ✅
  ├── 5A-3: 计费流水审计日志 ✅
  └── 5A-4: SQLite → MySQL 迁移文档 ✅

✅ PLAN.md 所有任务已完成！

可选后续工作：
  ├── 提交当前 23 个文件
  ├── 测试 Docker Compose 生产部署
  ├── 更新 README.md 和 API 文档
  ├── 性能压测（Redis Session + JWT）
  └── 发版 v2.3.0

长期
  └── 前端高价提示 (Playground/模型选择器) + 前端首选供应商设置
```

---

## 依赖关系图

```
2A-1 逐级锁        ───  无依赖
2A-2 密码强度       ───  无依赖
2A-3 新用户引导     ───  依赖 2A-2（密码强度决定注册表单修改范围）
2B-1 扣费优先级     ───  依赖 4A-1（佣金池拆分后扣费逻辑不同）
2B-2 签到奖励       ───  无依赖
2B-3 提现自动审批   ───  依赖 4A-1（佣金池）
2C-1 供应商审批     ───  无依赖
2C-2 模型列表       ───  无依赖
3A-1 Distributor加密 ─── 无依赖
4A-1 佣金独立池     ───  无依赖（但要改动 model/user.go）
4A-2 多级返佣       ───  依赖 4A-1
4B-1 分账优化       ───  依赖 4A-1
```



