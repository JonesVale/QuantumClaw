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

## 第二阶段 — 用户体验改善（当前）

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

#### 2A-3：新用户注册引导 + 试用额度（2h）
- **位置**：`model/user.go:Insert`, 前端 pages
- **现状**：注册后 quota=0，用户无所适从
- **目标**：
  - 注册时赠送 `config.QuotaForNewUser`（建议 50000 ≈ $0.1），已有相关配置但被注释/取消
  - 注册完成后跳转引导页/弹窗，提示"试用额度已到账，体验请前往 Playground"
  - Dashboard 显示新手引导卡片
- **文件**：`model/user.go`, `controller/user.go`, `web/.../dashboard.tsx`, `web/.../sign-in.tsx`

### Batch 2B：计费/充值完善

#### 2B-1：充值扣费优先级确认 + 文档化（2h）
- **位置**：`relay/billing/billing.go`, `service/cash_billing.go`
- **现状**：`PreConsumeQuota` 和 `PreConsumeBalance` 是两条并行路径但优先级不明确
- **目标**：
  1. 用户有订阅 → 先扣订阅额度（`SubscriptionPreConsume`）
  2. 订阅额度用完 → 扣 `CashBalance`
  3. `CashBalance` 用完 → 回退到 `Quota`
  4. 没有额度 → 拒绝请求
- **改动**：改造 `relay/common_handler/billing.go` 的消费链，增加优先级判断
- **文件**：`relay/common_handler/billing.go`, `relay/billing/billing.go`

#### 2B-2：签到奖励提升（0.5h）
- **位置**：`setting/operation_setting/checkin_setting.go`
- **现状**：`MinQuota=1000, MaxQuota=10000`（约 ¥0.002~¥0.02）
- **目标**：提升 10 倍 `MinQuota=10000, MaxQuota=50000`（约 ¥0.02~¥0.1），提高用户活跃度
- **文件**：`setting/operation_setting/checkin_setting.go`

#### 2B-3：提现自动审批（2h）
- **位置**：`controller/commission.go:RequestWithdrawal`
- **现状**：纯手动审批
- **目标**：
  - 提现金额 ≤ `AutoWithdrawThreshold`（建议 ¥100）→ 自动通过
  - 超额 → 需管理员审核
  - 前端显示自动/待审核状态
- **文件**：`controller/commission.go`, `model/commission.go`, `web/.../commission.tsx`

### Batch 2C：供应商通道完善

#### 2C-1：供应商升级验证（3h）
- **位置**：`controller/user.go:UpgradeToProvider`
- **现状**：一键升级无审核
- **目标**：
  - 升级条件：用户必须已有有效的 API Key 配置才能申请升级
  - 或升级后状态为 `status=pending`，必须管理员在 Users 页审批
  - 前端 Profile 页升级按钮显示当前状态
- **文件**：`model/user.go`, `controller/user.go`, `web/.../profile.tsx`, `web/.../users.tsx`

#### 2C-2：Channel Type 模型列表修复（2h）
- **位置**：`controller/channel.go:GetChannelTypes`
- **现状**：`model_metadata.provider` 字段与 `channeltype.ChannelTypeNames` 的品牌名不匹配
- **目标**：
  - 添加名映射表，将 channeltype 名称对应到 model_metadata 中的 provider 名称
  - 或修改 model_metadata 的 provider 使用标准名称
- **文件**：`controller/channel.go`, `relay/channeltype/names.go`

---

## 第三阶段 — 安全加固

### Batch 3A：数据安全

#### 3A-1：Distributor.ApiKey AES 加密（0.5h）
- **位置**：`model/distributor.go:Distributor` + `controller/distributor.go:CreateDistributor`
- **现状**：明文 32 位随机串
- **目标**：复用 `common/encrypt` 包的 AES-GCM 加密，和 Channel.Key 加密方案一致
- **文件**：`model/distributor.go`, `controller/distributor.go`

#### 3A-2：SubscriptionPlan 价格 int64 化（0.5h）
- **位置**：`model/subscription.go:SubscriptionPlan`
- **现状**：`PriceAmount float64`
- **目标**：改为 `PriceCents int64`（单位：分），避免浮点精度
- **文件**：`model/subscription.go`, `controller/subscription.go`

### Batch 3B：认证增强

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

#### 4A-1：佣金独立池（3h）
- **位置**：`model/commission.go:RewardInviterOnConsume`
- **现状**：返佣直接加 `quota`（可消费额度），和平台现金混在一起
- **目标**：
  - 新增 `commission_balance` 独立佣金余额字段（用户表 or 独立表）
  - 返佣累加到此余额
  - 提现时从佣金余额扣除
  - 佣金余额不可直接用于消费
- **文件**：`model/user.go`, `model/commission.go`, `controller/commission.go`, `model/billing.go`

#### 4A-2：多级返佣（3-5h）
- **位置**：`model/commission.go:RewardInviterOnConsume`
- **现状**：仅 1 级邀请人
- **目标**：
  - 支持 3 级代理链（1:5%, 2:2%, 3:1%）
  - 从 `InviterId` 递归查找第 2/3 级
  - 在 CommissionRecord 记录代级信息
- **文件**：`model/commission.go`, `model/user.go`

### Batch 4B：消费/结算

#### 4B-1：供应商分账系统优化（3h）
- **位置**：`service/cash_billing.go:PostConsumeDeduct`
- **现状**：分账到渠道所有者，但分账比例硬编码
- **目标**：
  - 渠道级 `profit_split` 分账比例（现状 `channel.UserId > 0` 时全部分给渠道所有者）
  - 平台抽成比例可配置
  - 分账日志/报表
- **文件**：`model/channel.go`, `service/cash_billing.go`, `controller/channel.go`

#### 4B-2：异常结算处理（2h）
- **位置**：`model/billing.go`
- **现状**：`Debt` 字段（追偿挂账）已存在但无完整处理流程
- **目标**：
  - 余额不足但请求已转发完成 → 记 `Debt`（负数余额）
  - 下次充值自动扣还
  - 长期不还 → 禁用账号
- **文件**：`model/user.go`, `relay/billing/billing.go`

---

## 第五阶段 — 长期架构改善

### Batch 5A：核心技术

| # | 任务 | 工时 | 优先级 | 说明 |
|:-:|------|:----:|:------:|------|
| 5A-1 | JWT Token 鉴权 | 8h | 中 | 支持双模式（session + JWT），多机部署必须 |
| 5A-2 | Redis Session 共享 | 3h | 中 | 当前 `REDIS_HOST` 已配但可能未启用 |
| 5A-3 | 计费流水审计日志 | 4h | 低 | 余额变更全量日志，支持回滚 |
| 5A-4 | SQLite → MySQL 迁移文档 | 2h | 低 | 已有双 DB 支持但迁移流程不清晰 |

### Batch 5B：运维/部署

| # | 任务 | 工时 | 优先级 |
|:-:|------|:----:|:------:|
| 5B-1 | Docker Compose 一键部署 | 2h | 低 |
| 5B-2 | nginx 配置示例 | 1h | 低 |
| 5B-3 | 备份/恢复脚本（含密钥） | 2h | 中 |
| 5B-4 | 健康检查端点 + Prometheus 集成 | 3h | 低 |

---

## 执行优先级矩阵

```
               高影响                 中影响                低影响
            ┌────────────────────────────────────────────────────┐
 容易  │  2A-1: 登录逐级锁        2B-3: 提现自动审批    2B-2: 签到奖励
       │  2A-3: 新用户引导        2C-2: 模型列表修复
       │  3A-1: Distributor加密   2B-1: 扣费优先级
       │                        2A-2: 密码强度可配置
       ├────────────────────────────────────────────────────┤
 中等  │  4A-1: 佣金独立池        3A-2: 价格int64化      5A-2: Redis
       │  2C-1: 供应商审批        4B-1: 分账优化
       │                        4A-2: 多级返佣
       ├────────────────────────────────────────────────────┤
 困难  │  5A-1: JWT              4B-2: 异常结算         5B-x: 运维
       │                        5A-3: 审计日志
       └────────────────────────────────────────────────────┘
```

---

## 建议执行顺序

```
本周（Batch 2）
  ├── 2A-1: 登录逐级锁 (1h ↓)
  ├── 2A-2: 密码强度可配置 (1h ↓)
  ├── 2A-3: 新用户注册引导 (2h ↓)
  ├── 2B-2: 签到奖励 (0.5h)
  └── 3A-1: Distributor加密 (0.5h)

下周（Batch 2 + 3）
  ├── 2B-3: 提现自动审批 (2h)
  ├── 2C-1: 供应商审批 (3h)
  ├── 2C-2: 模型列表匹配 (2h)
  └── 2B-1: 扣费优先级 (2h)

后续（Batch 4）
  ├── 4A-1: 佣金独立池 (3h)
  ├── 4B-1: 分账优化 (3h)
  └── 4A-2: 多级返佣 → 按需

长期
  └── 5A-x + 5B-x
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
