# QuantumClaw 项目规划

**版本**: v2.2.0 | **更新**: 2026-06-02 | **状态**: 🟢 18 项竞品差距已实现，进入发布准备阶段

---

## 📍 当前状态

2026-05-26~29 冲刺完成了 PLAN.md 最初定义的 **全部 18 项竞品差距任务**（P0~P3 四个阶段）。代码库已验证通过。

**剩余工作不再是"补差距"，而是"保质量、修结构、定方向"**。

---

## Phase S0: 基础设施修复（半天）

### S0-1 代码库洁癖清理

**目标**: 停止跟踪二进制/产物文件，降低仓库体积

| 待清理文件 | 大小 | 处理方式 |
|-----------|:----:|---------|
| `quantumclaw.exe` | 64 MB | `git rm --cached` + 追加 `.gitignore` |
| `quantumclaw_new.exe` | 106 MB | `git rm --cached` + 追加 `.gitignore` |
| `quantumclaw_test.exe` | 46 MB | `git rm --cached` + 追加 `.gitignore` |
| `支付宝收款码.jpg` | 172 KB | `git rm --cached` |
| `LoginBackup.png` | 3 MB | `git rm --cached` |
| `coverage.out` / `coverage_full.out` | ~1 MB | `git rm --cached` + 追加 `.gitignore` |
| `quantumclaw.db-journal` | 37 KB | `git rm --cached` + 追加 `.gitignore` |

涉及改动: `.gitignore`（追加 `*.exe` `*.db-journal` `coverage*.out` `payment/*`）

**验收标准**:
- [ ] `git status` 不再显示 `.exe`、`收款码`、`coverage.out`
- [ ] 仓库 `.exe` 文件数变为 0

---

### S0-2 分支同步

**目标**: develop 追上 master，恢复三层分支模型

**操作**: `git checkout develop && git merge master`（fast-forward，零冲突）

**后续跟进**:
- feature/* → develop（不可直接推 master）
- master 仅通过 PR 合并（见 `BRANCH_STRATEGY.md`）
- develop 作为 CI 验证环境

**验收标准**:
- [ ] develop 和 master 指向同一 commit
- [ ] `git log master..develop` 返回空

---

### S0-3 构建检查

- [ ] `go build ./...` 通过
- [ ] `go vet ./...` 零告警
- [ ] `go test ./...` 全部 PASS
- [ ] 前端 `rsbuild build`（如适用）

**验收标准**:
- [ ] 三项工具链命令均通过

---

## Phase S1: 18 项功能验收验证（1~2 天）

代码存在不等于功能可用。逐项跑验收条件。

### P0-1 联网搜索

**代码验证**:
- [ ] `service/web_search.go` ✅ 存在
- [ ] `middleware/search.go` ✅ 存在
- [ ] `controller/search_controller.go` ✅ 存在

**验收条件**:
- [ ] 3 个搜索后端配置可切换（Bing / SearXNG / SerpAPI）
- [ ] Auto / Manual / Always 三种模式生效
- [ ] 无搜索 Key 时优雅降级
- [ ] X-Enable-Search / X-Search-Query 头部支持

### P0-2 模型参数自动纠错

**代码验证**: `middleware/param_validator.go` ✅ 已实现

**验收条件**:
- [ ] max_tokens 超过模型上限时自动裁剪（auto-fix）
- [ ] temperature / top_p 越界时自动截断
- [ ] 勾选了 LogOnly 时仅记录不修改
- [ ] 支持按模型独立配置上下限（o-series fixed temp=1）
- [ ] 修正后的 body 正常透传下游

### P1-3 智能提示词优化

**代码验证**: `middleware/prompt_optimizer.go` ✅ 已实现

**验收条件**:
- [ ] basic 级别注入优化 system prompt
- [ ] advanced 级别改写用户消息
- [ ] 无 system 消息时自动插入
- [ ] 优化不改变原始请求的语义
- [ ] 用户可开关

### P1-4 IP 白名单

**代码验证**: `middleware/auth.go`（token.Subnet + network.IsIpInSubnets）✅

**验收条件**:
- [ ] Token 支持配置 CIDR 白名单
- [ ] 匹配时放行，不匹配返回 403
- [ ] 全局禁用 IP 限制配置

### P1-5 渠道余额自动监控

**代码验证**: `model/channel.go`（BalanceAlertThreshold / BalanceDisableThreshold）✅

**验收条件**:
- [ ] 余额低于 AlertThreshold 时触发通知
- [ ] 余额低于 DisableThreshold 时自动禁用渠道
- [ ] 阈值可在管理端配置

### P1-6 多机部署（级联架构）

**代码验证**:
- [ ] `model/cascade_node.go` ✅
- [ ] `controller/cascade.go` ✅
- [ ] `service/cascade_client.go` ✅
- [ ] `deploy/slave/` ✅

**验收条件**:
- [ ] 主从节点 Session 共享正常
- [ ] 从节点管理请求代理到主节点
- [ ] 读请求在从节点本地处理

### P1-7 API 文档 + Swagger

**代码验证**: `docs/swagger.json` ✅

**验收条件**:
- [ ] Swagger UI 可访问 `/api/swagger/index.html`
- [ ] 覆盖 35+ endpoints
- [ ] 包含请求/响应示例

### P0-8 智能模型路由

**代码验证**:
- [ ] `model/channel_performance.go` ✅
- [ ] `service/channel_router.go` ✅
- [ ] `middleware/intelligent_router.go` ✅
- [ ] `controller/router_controller.go` ✅

**验收条件**:
- [ ] 路由引擎可根据延迟/成本/可用性动态选择渠道
- [ ] 连续失败渠道自动剔除并恢复
- [ ] 路由策略可视化配置

### P1-9 定价策略增强

**代码验证**:
- [ ] `controller/pricing_admin.go` ✅
- [ ] SubscriptionPlan / ChannelMarkup 模型层 ✅

**验收条件**:
- [ ] 支持阶梯计费
- [ ] 支持包月+按量混合计费
- [ ] 渠道商可配置上浮倍率
- [ ] 计费明细可审计

### P1-10 Geo 地理服务

**代码验证**:
- [ ] `service/geo_service.go` ✅
- [ ] `middleware/geo.go` ✅
- [ ] `controller/geo_controller.go` ✅

**验收条件**:
- [ ] 支持天气查询
- [ ] 支持商圈搜索
- [ ] 支持路线规划
- [ ] 地理服务可独立定价

### P1-11 Sub2API

**代码验证**:
- [ ] `relay/adaptor/sub2api/` ✅
- [ ] `middleware/sub2api.go` ✅

**验收条件**:
- [ ] ChatGPT Plus/Pro 订阅转 API
- [ ] Claude Pro 订阅转 API
- [ ] 凭证安全存储
- [ ] 日请求上限可配置

### P1-12 国际化

**代码验证**: `common/i18n/locales/` 14 语种 ✅

**验收条件**:
- [ ] 英文 UI 无中文残留
- [ ] 后端错误消息支持英文
- [ ] 14 语种翻译覆盖

### P2-13 Dashboard 监控面板

**代码验证**: `controller/monitoring.go` ✅

**验收条件**:
- [ ] 渠道延迟趋势图（24h / 7d / 30d）
- [ ] 调用成功/失败统计
- [ ] 页面加载 < 2s

### P2-14 一键部署

**代码验证**:
- [ ] `deploy/sealos/` ✅
- [ ] `deploy/render.yaml` ✅
- [ ] `deploy/slave/Dockerfile` ✅

### P3-15 开放平台 SDK

**代码验证**:
- [ ] Python SDK (`sdk/python/`) ✅
- [ ] Node.js SDK (`sdk/nodejs/`) ✅
- [ ] Go SDK (`sdk/go/`) ✅
- [ ] Developer Portal (`web/default/src/`) ✅

### P3-16 自有模型托管

**代码验证**:
- [ ] vLLM adaptor ✅
- [ ] SGLang adaptor ✅
- [ ] InferenceNode model ✅

### P3-17 宝塔面板

**代码验证**: `deploy/baota/install.sh` ✅

### P3-18 社区反馈

**代码验证**:
- [ ] FAQ API ✅
- [ ] Feedback CRUD ✅
- [ ] `/faq` / `/feedback` 前端页面 ✅

---

## Phase S2: v2.3.0 方向决策（后续）

18 项差距补齐后，下一步需要明确方向。候选方向：

| 方向 | 说明 | 工作量 |
|------|------|--------|
| 生产运营 | 腾讯云上线、监控告警、运维手册 | M（1 周） |
| 性能优化 | relay 链路延迟优化、数据库查询优化 | M（1 周） |
| 测试覆盖率 | 当前 29 个 test file，目标 50+ | L（2 周） |
| CI/CD 自动化 | GitHub Actions 自动构建+测试+部署 | M（1 周） |
| 多租户增强 | 企业级团队管理、RBAC 细化 | L（2 周） |

> 具体方向在 Phase S1 完成后与用户讨论决定。

---

## 附录 A: 分支策略（2026-05-29）

```
master          ← 生产分支，仅通过 PR 合并
  ↑
develop         ← 集成分支，功能分支在此合并验证
  ↑
feature/*       ← 功能分支，从 develop 拉出
fix/*           ← 修复分支
hotfix/*        ← 紧急修复，从 master 拉出，合并回 master + develop
```

**提交规范**: `<type>: <描述>` — 类型: feat / fix / chore / refactor / docs / style / test

---

## 附录 B: 共享依赖模块（参考用）

> 以下来源于原始竞品差距分析中的架构设计，保留供后续开发参考。

| 模块 | 涉及功能 | 说明 |
|------|---------|------|
| `middleware/` | 联网搜索、参数纠错、提示词优化、IP白名单、Geo、Sub2API | 统一 middleware 加载机制 |
| `service/` | 余额监控、智能路由、定价、Geo、Sub2API | 核心业务逻辑 |
| `model/` | 几乎所有功能需要扩展数据模型 | 注意字段兼容性 |
| `controller/` | 管理 API CRUD | RESTful 风格统一 |
| `relay/` | 路由发现、Sub2API、量子适配器 | 适配器接口统一 |
| `web/default/src/` | 前端管理 UI | React + rsbuild |

---

## 附录 C: 已知风险

| 风险 | 等级 | 说明 |
|------|:----:|------|
| 分支习惯 | 🟡 | 直接推 master 的习惯需要改，否则 develop 永远过时 |
| Git 产物污染 | 🟢 | 二进制文件被追踪，方案明确 |
| 验收条件未跑 | 🟡 | 代码存在≠可工作，Phase S1 专门处理 |
| Sub2API 安全 | 🟠 | 用户凭证需隔离存储 |
| 翻译质量 | 🟢 | AI 辅助 + 人工校对 |
| 智能路由算法 | 🟠 | 当前为加权轮询，但复杂场景需验证 |
