# QuantumClaw 商业运营体系完善计划

## 当前版本: v2.2.0 | 31 commits 未推送

---

## Phase 0 — 推广佣金系统（P0 🔴 高）

完整的推广返佣链路：

```
Admin设置佣金比例 → 推广员分享链接 → 好友注册 → 好友消费 → 自动返佣 → 佣金记录 → 提现
```

### 子任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 0.1 | 数据库: 佣金配置表 `CommissionSetting` | `model/commission.go` | 20 min |
| 0.2 | 数据库: 佣金记录表 `CommissionRecord` | `model/commission.go` | 20 min |
| 0.3 | 数据库: 提现记录表 `WithdrawalRequest` | `model/commission.go` | 15 min |
| 0.4 | 后端: 佣金配置 CRUD API | `controller/commission.go` | 30 min |
| 0.5 | 后端: 用户注册时检查 InviterId → 自动返佣 | `controller/user.go` (修改) | 15 min |
| 0.6 | 后端: 用户消费时按比例返佣给邀请人 | `service/quota.go` (修改) | 20 min |
| 0.7 | 后端: 佣金记录查询 + 提现申请 API | `controller/commission.go` | 30 min |
| 0.8 | 前端: Wallet 页扩展 — 佣金统计 + 提现 | `wallet.tsx` | 40 min |
| 0.9 | 前端: 推广团队/下线列表 | `wallet.tsx` 或新页面 | 30 min |
| 0.10 | i18n: 佣金相关翻译 × 7 语言 | i18n JSON 文件 | 10 min |

**预估: 3.5 小时**

---

## Phase 1 — 渠道利润分析（P1 🟡 中）

```
渠道成本价 → 自动计算利润率 → 管理后台报表
```

### 子任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 1.1 | Channel 模型加 `CostPerUnit` 字段（成本价） | `model/channel.go` | 5 min |
| 1.2 | 后端: 利润计算 API（售价=倍率×单位额度，成本=CostPerUnit） | `controller/channel.go` | 20 min |
| 1.3 | 后端: 按渠道统计总消费 + 总成本 + 总利润 | `controller/channel.go` | 20 min |
| 1.4 | 前端: Channels 页面加利润列 | `channels.tsx` | 20 min |
| 1.5 | i18n 翻译 | i18n JSON | 5 min |

**预估: 1 小时**

---

## Phase 2 — 分销渠道分成（P1 🟡 中）

```
渠道分级定价 → 分销商独立定价 → 自动分账
```

### 子任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 2.1 | 分销商角色 + 分销商表 | `model/distributor.go` | 20 min |
| 2.2 | 分销商独立定价规则（覆盖全局模型倍率） | `controller/distributor.go` | 30 min |
| 2.3 | API: 分销商 CRUD + 定价管理 | `router/api.go` | 20 min |
| 2.4 | 前端: 分销商管理页面 | 新页面 | 40 min |

**预估: 2 小时**

---

## Phase 3 — 语言体系前端集成（P2 🟢 低）

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 3.1 | 前端 useTranslation hook 改为 API 驱动 | hook 文件 | 60 min |
| 3.2 | JSON 翻译数据导入 T_Languages 表 | 迁移脚本 | 20 min |

**预估: 1.5 小时**

---

## Phase 4 — 量子适配器完善（P2 🟢 低）

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 4.1 | Rigetti 适配器完整实现 | `relay/quantum/rigetti/` | 60 min |
| 4.2 | AWS Braket 适配器 | `relay/quantum/braket/` | 60 min |
| 4.3 | Azure Quantum 适配器 | `relay/quantum/azure/` | 60 min |

**预估: 3 小时**

---

## 执行顺序建议

```
Phase 0 (推广佣金) ─── 3.5h ─── 最影响商业价值，优先做
  └── Phase 1 (利润分析) ─ 1h ─── 依赖 Phase 0 的计费数据
      └── Phase 2 (分销分成) ─ 2h ─── 高级功能
          └── Phase 3 (语言统一) ─ 1.5h
              └── Phase 4 (量子适配器) ─ 3h
```

**总预估: 11 小时**

---

从 Phase 0 开始？还是调整优先级？
