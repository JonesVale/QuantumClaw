# 方案C：关键问题与整体架构分析

## 0. 业务全景

```
                      同一域名 (api.quantumclaw.com)
                      DNS 智能解析 / Geo CDN
                              │
         ┌────────────────────┼────────────────────┐
         │                    │                    │
    东京子节点           法兰克夫子节点          硅谷子节点
    (api转发+本地渠道)    (api转发+本地渠道)     (api转发+本地渠道)
         │                    │                    │
         └────────────────────┼────────────────────┘
                              │
                         主节点(新加坡)
                    ┌────────┴────────┐
                    │                  │
                用户管理            财务管理
                Token签发           全局配置
                Dashboard           渠道结算
```

**用户视角**：
1. 注册/登录/充值 → 主节点（一个操作就够了）
2. 创建 API Key → 主节点
3. 用 API Key 调模型 → DNS 自动去最近的子节点
4. 查用量/看统计 → 主节点 Dashboard

---

## 1. 关键问题（必须解决的）

### 🔴 K1：Token 数据的双向同步

这是最核心的问题。

```
        主节点                           子节点
  ┌──────────────────┐          ┌──────────────────┐
  │  用户创建 Token   │  ──→    │  子节点 Redis     │
  │  Token 被禁用     │  ──→    │  Token 缓存更新    │
  │  配额变动         │  ──→    │  缓存失效/刷新     │
  │                   │          │                   │
  │  收取子节点回传    │  ←──    │  本地扣减配额      │
  │  配额修正         │  ←──    │  扣减冲突上报       │
  └──────────────────┘          └──────────────────┘
```

**为什么难**：
- Token 的创建/禁用是**主节点到子节点**（主写从读）
- Token 的配额扣减是**子节点到主节点**（各地都扣）
- 双向流，还涉及**配额一致性**

**解法**：
- 主→子：增量拉取（`UpdatedAt` 字段决定增量集）
- 子→主：批量回传 + 幂等键
- 冲突：最终一致性 + 修正机制

**代价**：Token 表新增 `UpdatedTime` 字段 + 审计所有写 Token 的位置保证都更新这个字段。

### 🔴 K2：计费的原子性 vs 分布式

**主节点（当前）**：
```go
DB.Transaction(func(tx) {
    token.RemainQuota -= amount    // 扣配额
    token.UsedQuota += amount       // 累计用量
    log.Create(...)                  // 写日志
    user.Balance -= cost            // 扣用户余额（在另一张表）
})
```
→ **单机 ACID，完美**

**子节点（级联）**：
```
子节点 Redis:
    token:{hash}.quota -= amount       // 乐观扣减（没有 ACID）

过 10 秒：
    POST /api/cascade/billing/batch    // 发给主节点

主节点:
    token.RemainQuota -= amount        // 真实扣减
    发现配额不足 → 拒绝该 batch
    拒绝通知 → 子节点修正缓存
```
→ **最终一致性，有修正窗口**

**这个窗口会出什么问题？**

| 场景 | 窗口期内 | 修正后 |
|------|---------|--------|
| 用户正常调用 | 子节点成功 | 主节点确认，一切正常 |
| 用户超额调用 | 子节点成功多笔 | 主节点拒绝超额部分 |
| 两个子节点同时扣 | 各自缓存不同步 | 主节点汇总后修正 |

**关键决策：是否接受最终一致性？**

如果谷主的业务涉及**预付费渠道商**（渠道商先充值再消耗），超额会导致渠道商欠费，这不是技术问题能解决的。

如果涉及**普通用户**（先消耗后结算），最终一致性完全够用。

**解法**：给 Token 加一道"配额软上限"——子节点缓存中的配额加倍，主节点才是真实余额。子节点永远不拒绝请求（最多多放行一点），主节点最终扣真实值。

### 🔴 K3：用户信息在哪里管？

用户登录后，需要在子节点做 Token 验证。但 Token 验证需要的用户信息（分组、状态）在哪？

```
子节点 Token 验证需要知道：
  └─ Token 对应的 user.group      ← 决定了模型计费倍率
  └─ Token 对应的 user.status     ← 用户是否被禁用
  └─ Token 本身的 status/quota    ← 是否过期/超配额
```

**选项**：

| 方案 | 做法 | 好坏 |
|------|------|------|
| A. Token 自身携带 user 信息 | Token 缓存里存 user.group + user.status | ✅ 子节点不依赖用户表<br>❌ user 信息改了 Token 感知不到 |
| B. 独立同步用户表 | 子节点也缓存 User 快照 | ✅ 数据独立<br>❌ 多一张表同步 |
| C. 子节点不验证，全量回主 | 每次 Token 验证都 HTTP 问主节点 | ❌ 延迟太高 |

**推荐 A**：Token 缓存里冗余存 user.group + user_status。用户分组变化虽不频繁但存在，方案是 Token 同步时把 user 的 group 也带过来，**Token 缓存 = Token 信息 + User 关键字段**。

### 🔴 K4：子节点要不要有 Web UI？

这是**产品决策**，不是技术决策。

```
设计A：纯 API 网关
  ┌───────────────┐
  │ 东京子节点     │
  │               │
  │ POST /v1/...  │  ← 只用 API
  │ 本地渠道转发   │
  │ 计费回传       │
  │               │
  │ ❌ 无 web 页面  │
  └───────────────┘
  Dashboard 只跑在主节点

设计B：全功能实例
  ┌───────────────┐
  │ 东京子节点     │
  │               │
  │ POST /v1/...  │
  │ 管理页面       │  ← 也有 web
  │ 用户本地登录   │
  │ 本地 Dashboard │
  └───────────────┘
  用户信息从主节点级联查
```

**我建议设计A**，理由：
1. 用户一年操作几次管理页面？主要是**创建/刷新 API Key**和**查看用量**，都在主节点完成
2. 子节点纯 API = 更小镜像（~30MB 无前端）、更快部署、更少攻击面
3. 同一套 Dashboard，不需要维护多区域版本

**代价**：子域名分离 `api.region-qc.com` vs `dashboard.quantumclaw.com`

---

## 2. 关联性问题（改了 A，B 会怎样？）

### 🔗 L1：计费链路改动会影响所有 relay adaptor

当前计费链路：

```
每个 relay adaptor（50+ 个）
  → relay/controller/helper.go:postConsumeDeduct()
  → service/cash_billing.go:PostConsumeDeduct()
  → DB 事务
```

子节点模式下要改成：

```
每个 relay adaptor（50+ 个）
  → relay/controller/helper.go:postConsumeDeduct()
  → service/cash_billing.go:PostConsumeDeduct()
  → if config.IsMasterNode:
      → DB 事务（原逻辑不变）
    else:
      → Redis 缓存扣减 + 入计费队列（新分支）
```

**影响**：只改 `PostConsumeDeduct` 一个函数，50+ adaptor **零改动**。因为 adaptor 只调用 `postConsumeDeduct`，不感知计费细节。✅

### 🔗 L2：认证链路 fork

当前认证：

```
middleware/auth.go:TokenAuth()
  → DB 查 Token (model.GetTokenByKey)
  → DB 查 User (model.GetUserById)
  → c.Set("user_id", ...)  /  c.Set("group", ...)
```

子节点版：

```
middleware/auth.go:TokenAuth()
  → if config.IsMasterNode:
      → DB 查 Token（原逻辑）
    else:
      → Redis 查 Token 缓存
      → 未命中 → 级联 API 拉取 → 写入 Redis
  → c.Set("user_id", ...)  /  c.Set("group", ...)
```

**影响**：只改 `TokenAuth` 一个函数。整个 middleware 链的上下游（Search/Prompt/ParamValidator/Distribute）**不感知** Token 验证来自哪里。✅

### 🔗 L3：配置的分层归属

当前所有配置都在同一套 DB key-value 表：

```
options 表 ← Midjourney 配置、登录设置、通知设置……
platform_configs 表 ← 平台级配置
```

级联后需要分层：

```
全局配置（主节点下发 → 子节点同步）
  ├── 模型定价表（ModelRatio）
  ├── 模型映射（model_mapping）
  ├── 功能开关（EnableMetric, AutomaticDisableChannel…）
  └── 全局限流参数

本地配置（子节点自主）
  ├── 渠道列表（Channels）
  ├── 通知设置（Webhook, Email）
  ├── 本地限流（RateLimit）
  └── 本地日志（Logs）
```

**影响**：
- 子节点需要从配置中区分"全局"和"本地"
- 方法：子节点启动时同步全局配置覆盖本地，本地配置保持独立
- 需要定义一份"全局配置键名单"

### 🔗 L4：日志/统计的归属

当前请求日志直接写 DB，Dashboard 从 DB 查询。

```
设计决策：
  ├── 请求日志 → 子节点本地写
  │   Dashboard 查全局日志时需要跨节点查询
  │   或主节点聚合（代价大）
  │
  └── 汇总统计 → 心跳时上报
      子节点心跳时附带：today_calls, avg_latency, success_rate
      主节点 Dashboard 只展示汇总数据
      详细日志进子节点本地 Dashboard（可选项）
```

---

## 3. 业务逻辑布局（最终建议）

### 3.1 功能矩阵

| 功能 | 主节点 | 子节点 | 备注 |
|------|:------:|:------:|------|
| 用户注册/登录 | ✅ | ❌ | 主节点自建账户系统 |
| Token 签发/管理 | ✅ | ❌ | 主节点集中管理 |
| Token 验证 | ❌ | ✅ | 子节点本地缓存验证 |
| API 转发 | ❌ | ✅ | 子节点本地渠道转发 |
| 计费扣减 | ✅ 汇总 | ✅ 预扣 | 主节点最终确认 |
| 渠道管理 | optional | ✅ | 子节点自有渠道 |
| 模型定价 | ✅ 统一定价 | ❌ 同步执行 | 主节点下发 |
| 配额查询 | ✅ | ✅ 缓存 | 子节点 Redis |
| 用户 Dashboard | ✅ | ❌ | 主节点访问 |
| 节点管理面板 | ✅ | ❌ | 主节点管理员用 |
| 请求日志 | ❌ | ✅ 本地 | 按需聚合 |
| 系统通知 | ✅ | ❌ | 主节点发通知 |

### 3.2 启动流程

```
主节点启动：
  DB migration → Redis → Web 服务
  → 正常提供所有功能
  → 新增：/api/cascade/* 路由

子节点启动：
  1. 连接本地 Redis（无 DB）
  2. POST /api/cascade/node/register → 获取 node_id + api_key
  3. 首次全量同步 Token（GET /api/cascade/tokens/sync?since=0）
  4. 同步全局配置
  5. 标记 master_connected = true
  6. 启动 Gin 服务（仅 /v1/* 路由）
  7. 启动后台定时器（heartbeat/token_sync/billing_flush）

子节点健康检查异常：
  └─ 心跳超时 → 标记 master_connected = false
  └─ Token 同步失败 → 继续服务（使用过期缓存）
  └─ 计费回传失败 → 本地队列继续累积，下次重试
```

### 3.3 API 路由对比

```
主节点：
  /api/*               ← 所有 API（用户管理、Dashboard、级联、转发）
  /api/cascade/*       ← 级联专用路由（子节点调用）
  /v1/chat/completions ← 转发（主节点也能转发）

子节点：
  /v1/chat/completions ← 转发（这是主要功能）
  /v1/*                ← 所有 OpenAI 兼容 API
  /health              ← 健康检查
  ❌ 无 /api/*         ← 不提供管理功能
```

### 3.4 配置示例

```yaml
# 主节点 .env
NODE_TYPE=master
SQL_DSN=root:pass@tcp(master-db:3306)/quantumclaw
REDIS_CONN_STRING=redis://master-redis:6379

# 子节点 .env
NODE_TYPE=slave
CASCADE_MASTER_URL=https://master.quantumclaw.com
CASCADE_NODE_NAME=Tokyo-A
CASCADE_REGION=ap-northeast-1
# ❌ 无 SQL_DSN - 子节点不连 DB
# ✅ 有 Redis - 本地 Token 缓存 + 计费队列
REDIS_CONN_STRING=redis://localhost:6379
# ✅ 有本地渠道 - 子节点自己配 API 供应商
CHANNEL_CONFIG=/etc/qc/channels.yaml
```

---

## 4. 整体实施方案

### 执行阶段

```
Phase 1: 打好地基（C1 + C2）
  修改模型 → 定义契约 → 确保 Token/User 可增量追踪

Phase 2: 主节点先做（C3）
  子节点还没部署，主节点先提供级联 API
  可以先手动测试（curl 模拟子节点）

Phase 3: 子节点核心（C4）
  级联客户端 + 计费缓冲队列 + Token 缓存认证
  最难的部分，需要反复验证

Phase 4: 子节点独立运行（C5）
  启动适配 + 部署脚本
  可以在本地起两个实例模拟测试

Phase 5: 前管后管（C6）
  主节点看到各个子节点状态
  不是必须的，但运营需要
```

### 关键是"自验证"

```
不依赖"等到最后再测试"

Phase 2 完成后:
  curl -X POST https://master/cascade/node/register \
    -H "Content-Type: application/json" \
    -d '{"name":"test-node","region":"dev"}'
  → 拿到 api_key ✓

  curl -X GET https://master/cascade/tokens/sync?since=0 \
    -H "X-Cascade-Key: qcn_a1b2c3..."
  → 拿到 Token 列表 ✓

Phase 3 完成后:
  本地起两个进程（master:8080, slave:9090）
  master 有用户 A，创建 Token T1
  slave 启动 → 注册 → 同步 → Redis 有 T1
  curl -X POST http://slave:9090/v1/chat/completions \
    -H "Authorization: Bearer T1" \
    -d '{"model":"gpt-4", "messages":[...]}'
  → Token 验证通过 → 转发 → 计费回传
```

---

## 5. 几个待决定的点

等待谷主确认：

1. **子节点是否要 Web UI？** 建议不要，纯 API 网关
2. **子节点用谁的渠道？** 子节点自己管本地渠道，渠道不共享
3. **配额超扣的容忍度？** 最终一致性接受吗？（推荐：软上限加倍，不让子节点拒绝请求）
4. **主节点宕机时子节点是否继续服务？** 建议继续，但标为"只读"状态（已缓存的 Token 继续服务，新 Token 拉不到）
5. **开始的部署范围？** 先实现单主+单子验证，再扩展
