# QuantumClaw 全量整合规划方案

> 生成日期: 2026-06-07
> 基线版本: v2.2.0 (develop)
> 最后更新: 2026-06-07
>
> ## 📊 执行状态
>
> | Phase | 状态 | Commits |
> |:------|:----:|:--------|
> | Phase 1 构建体系归一化 | ✅ 完成 | `ff3e7c4` `d8c59e2` |
> | Phase 2 app/ 集成 | 🟡 部分完成 | Makefile 已加，页面待实现 |
> | Phase 3 electron/ 集成 | 🟡 部分完成 | Makefile 已加 |
> | Phase 4 SDK 管理 | 🟡 待处理 | — |
> | Phase 5 历史遗留清理 | ✅ 完成 | `d8c59e2` |
> | Phase 6 i18n 整合 | 🟡 待处理 | — |
> | Phase 7 级联架构 | 🟡 待处理 | — |

---

## 一、项目资产全景

| 类型 | 数量 | 代码量 |
|------|:----:|:------:|
| Go 后端 | 572 文件 | ~80,000 行 |
| TS/TSX 前端 | ~150 文件 | ~30,000 行 |
| Git 提交 | 365 | 2 贡献者 |
| 项目目录 | 169（去重 node_modules） | — |

---

## 二、构建体系现状与问题

```
前端构建三套输出：
  web/default/src/         ← 源代码 (rsbuild + TanStack Router + Tailwind)
  web/default/dist/        ← 构建输出 (已 gitignore，history 中曾跟踪)
  web/build/default/       ← 仅 3 个占位文件 (favicon/logo/index.html)
  release/dist/            ← 另一套构建输出 (发布包)

Docker 构建流程（Dockerfile）：
  [Stage 1] rsbuild build → mv default/dist /app/web/build/default
  [Stage 2] go build → 嵌入 web/build/default
```

**根源**：项目经历了多次前端框架迁移（历史遗留），`web/build/` 是旧版构建目标路径。

**✅ 已修复（commit `ff3e7c4` + `d8c59e2`）**：
- `web/build/` 已整体删除（88 个冗余文件）
- Dockerfile 简化：去掉中间 `mv` 步骤，直接 `default/dist → web/default/dist`
- `.gitignore`/`.dockerignore` 统一更新

---

## 三、整合任务规划（共 7 个 Phase）

### Phase 1: 构建体系归一化
**目标**：消除冗余构建输出，统一构建链路

| 任务 | 优先级 | 耗时 | 依赖 |
|------|:------:|:----:|:----:|
| 1.1 清理 web/build/ 占位文件 | P0 | ✅ | commit `d8c59e2` |
| 1.2 修正 Dockerfile 构建输出路径 | P0 | ✅ | commit `ff3e7c4` |
| 1.3 确认 go embed 引用路径一致 | P0 | ✅ | 已验证（`main.go:111` → `router/web.go`） |
| 1.4 清理 release/dist/ 重复构建 | P1 | 🟡 | 非阻塞，可后续处理 |
| 1.5 更新 .gitignore 全面清理 | P1 | ✅ | commit `d8c59e2` |

### Phase 2: app/ 移动端集成
**目标**：移动端纳入统一构建流程 + 补全缺页面

| 任务 | 优先级 | 耗时 | 依赖 |
|------|:------:|:----:|:----:|
| 2.1 Makefile 新增 app 构建目标 | P0 | ✅ | commit `d8c59e2` |
| 2.2 移动端 i18n 对齐后端 | P1 | 🟡 | app/ 无 i18n 模块，待 Feature Sprint |
| 2.3 补全 provider/enterprise 屏幕 | P2 | 🟡 | 8 个 Placeholder 待实现 |

### Phase 3: electron/ 桌面端集成
**目标**：桌面端纳入统一发布流程

| 任务 | 优先级 | 耗时 | 依赖 |
|------|:------:|:----:|:----:|
| 3.1 Makefile 新增 electron 构建目标 | P0 | ✅ | commit `d8c59e2` |
| 3.2 版本号对齐（VERSION → electron） | P1 | 🟡 | electron 通过 API 读取版本 |

### Phase 4: SDK 统一管理
**目标**：三语言 SDK 版本对齐 + 统一发布

| 任务 | 优先级 | 耗时 | 依赖 |
|------|:------:|:----:|:----:|
| 4.1 SDK version 从 VERSION 文件读取 | P0 | 10min | — |
| 4.2 Makefile SDK 发布目标 | P1 | 10min | — |

### Phase 5: 清理历史遗留
**目标**：降低项目噪声，清理安全敏感数据

| 任务 | 优先级 | 耗时 | 依赖 |
|------|:------:|:----:|:----:|
| 5.1 归档 SQL/ 脚本至 docs/legacy/ | P0 | ✅ | 30 JS 脚本移入 docs/legacy/sql-migrations/ |
| 5.2 清理根目录实验性脚本 | P0 | ✅ | 14 个 `_` 文件移入 docs/legacy/ |
| 5.3 data/ 加入 .gitignore | P0 | ✅ | data/ 含 5.7MB SQLite 数据库，已排除 |
| 5.4 BFG 清理 Git history 敏感文件 | P1 | 🟡 | 需协调，已 gitignore 防护

### Phase 6: i18n 整合
**目标**：前后端语言资源统一

| 任务 | 优先级 | 耗时 | 依赖 |
|------|:------:|:----:|:----:|
| 6.1 后端补全 i18n 至 14 语言 | P1 | 1h | — |
| 6.2 前后端翻译 key 对齐 | P2 | 2h | 6.1 |

### Phase 7: 级联架构整合
**目标**：cmd_sync 融入主构建 + 量子计算适配对齐

| 任务 | 优先级 | 耗时 | 依赖 |
|------|:------:|:----:|:----:|
| 7.1 cmd_sync 接入 Makefile | P1 | 10min | — |
| 7.2 确保 quantum/relay 与主构建统一 | P2 | 15min | — |

---

## 四、业务链路与依赖关系分析

### 4.1 构建链路流向

```
源代码 → 构建 → 输出目录 → 嵌入/部署
  │          │        │
  │  web/default/src  │
  │  rsbuild build ───→ web/default/dist ──→ Docker Stage1
  │                                     └── release/dist (发布)
  │
  │  main.go + common/ + controller/ + relay/ + ...
  │  go build ──────────────────────────────────→ quantumclaw 二进制
  │                                             └── embed web/build/default
```

**关键依赖**：
- `router/web.go` 中通过 `embed.FS` 引用前端构建产物
- Dockerfile Stage1 → Stage2 传递依赖

### 4.2 数据流向（API 请求生命周期）

```
用户请求
  │
  ├── nginx (反向代理 / 静态资源)
  │   ├── /api/* ──→ Gin Server (3666)
  │   ├── /mj/*  ──→ Midjourney 代理
  │   ├── /pg/*  ──→ Playground
  │   └── /*     ──→ 前端 SPA
  │
  ├── middleware 层
  │   ├── auth.go           ──→ Session/Cookie/JWT 鉴权
  │   ├── login_rate_limit.go  ──→ 登录逐级锁
  │   ├── rate-limit.go       ──→ API Rate Limit
  │   ├── ssrf_protection.go  ──→ SSRF 防护
  │   ├── geo.go              ──→ Geo 路由
  │   ├── intelligent_router.go  ──→ 智能分组路由
  │   └── ...
  │
  ├── router → controller 层
  │   ├── user.go           ──→ 注册/登录/资料/密码
  │   ├── token.go          ──→ API Key 管理
  │   ├── channel.go        ──→ 渠道管理
  │   ├── topup*.go         ──→ 充值 (Stripe/支付宝/WorldFirst 等)
  │   ├── billing.go        ──→ 账单/消费
  │   ├── commission.go     ──→ 返佣管理
  │   ├── withdrawal.go     ──→ 提现
  │   └── ...
  │
  ├── relay/ 转发层
  │   ├── adaptor/          ──→ 52 家 AI 渠道适配
  │   ├── billing/          ──→ 计费引擎 (预扣/结算)
  │   ├── channeltype/      ──→ 渠道类型映射
  │   └── controller/       ──→ 转发控制器 (文本/图片/音频/视频)
  │
  └── 响应返回
```

### 4.3 计费链资金流向

```
用户充值
  │
  ├── Stripe / 支付宝 / Creem / Waffo / Binance / WorldFirst
  │
  ├── topup_*.go → model/topup.go
  │   └── 资金进入 CashBalance
  │
  ├── 调用 AI 模型时
  │   └── relay/common_handler/billing.go
  │       └── PreConsumeBilling
  │           └── 扣费优先级链：
  │               ① 订阅 (SubscriptionPlan)
  │               ② CashBalance
  │               ③ CommissionBalance (返佣余额)
  │               ④ Quota (免费额度)
  │                   └── 余额不足 → 记 Debt（挂账）
  │
  ├── 返佣链路
  │   └── model/commission.go:RewardInviterOnConsume
  │       └── 递归查找 3 级邀请人
  │           1 级 5% → 2 级 2% → 3 级 1%
  │           └── 进入 CommissionBalance
  │
  └── 提现
      └── withdrawal.go
          └── ≤¥100 自动审批
              └── CommissionBalance 扣除
```

### 4.4 多端共享资源

```
                  ┌─────────────────────┐
                  │    i18n 翻译资源     │
                  │   (7 语言后端 /      │
                  │    14 语言前端)       │
                  └──────────┬──────────┘
             ┌───────────────┼───────────────┐
             │               │               │
             ▼               ▼               ▼
        web/default/     app/ (Expo)    sdk/* (文档)
          (内嵌 JSON)     (内嵌 JSON)    (翻译参考)

                  ┌─────────────────────┐
                  │   API 类型定义       │
                  │   (dto/ 层)          │
                  └──────────┬──────────┘
             ┌───────────────┼───────────────┐
             │               │               │
             ▼               ▼               ▼
        web/default/      app/            sdk/*/
        (useT + API 调用)  (lib/api.ts)    (各语言实现)

                  ┌─────────────────────┐
                  │   版本号 VERSION     │
                  └──────────┬──────────┘
             ┌───────────────┼───────────────┐
             │               │               │
             ▼               ▼               ▼
        Go 后端         electron/         SDK 包
        (-ldflags)      (package.json)    (各语言)

                  ┌─────────────────────┐
                  │   数据模型 Model     │
                  │   (30+ GORM Model)   │
                  └──────────┬──────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
         MySQL 主库      SQLite 单机      Redis 缓存
         (docker)        (开发/小规模)    (限流/会话/渠道缓存)
```

---

## 五、各模块业务关联

| 核心模块 | 依赖方 | 被依赖方 | 说明 |
|---------|--------|---------|------|
| `common/config` | 全部 | — | 配置是整个系统的入口 |
| `model/user.go` | controller | middleware/auth, service 计费 | 用户模型是所有业务基础 |
| `model/channel.go` | controller/channel | relay/adaptor | 渠道配置驱动 AI 转发 |
| `relay/billing/billing.go` | relay/common_handler | model/user, model/commission | 计费引擎是商业核心 |
| `service/channel_router.go` | relay | model/channel | 路由决策 |
| `middleware/auth.go` | router | model/user, model/token | 所有认证 API 必经 |
| `web/default/src/` | 浏览器 | router/web.go + API | 用户交互入口 |
| `app/` | 手机端 | API（与 web 同后端） | 移动入口 |

---

## 六、执行顺序与依赖图

```
Phase 1 (构建)        ── 前置条件，无依赖
  │
  ├──→ Phase 5 (清理)  ── 可并行，但建议在 Phase 1 后
  │
  ├──→ Phase 2 (app)   ── 建议先完成构建统一
  │
  ├──→ Phase 3 (electron) ── 依赖 Phase 1 构建路径
  │
  ├──→ Phase 4 (SDK)   ── 独立，可与 2/3 并行
  │
  ├──→ Phase 6 (i18n)  ── 独立，可随时进行
  │
  └──→ Phase 7 (级联)  ── 依赖 Phase 1 构建
```

**建议执行顺序**：
Phase 1 → Phase 5 → Phase 2/3 (并行) → Phase 4 → Phase 6 → Phase 7
