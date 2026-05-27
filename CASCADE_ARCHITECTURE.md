# 方案C：级联架构设计分析

## 1. 问题域分析

### 1.1 为什么需要级联

```
用户场景                         单机瓶颈
─────────────────────────────────────────
东南亚用户 → 新加坡节点 → 延迟低  东南亚用户 → 新加坡 → 200ms
欧洲用户   → 法兰克福节点 → 延迟低  欧洲用户   → 新加坡 → 300ms+ 
北美用户   → 硅谷节点 → 延迟低    每节点单机 → 带宽/OOM
```

核心矛盾: **AI API 转发是延迟敏感型业务**。卡在路由和计费上浪费的 50ms，对用户体验影响巨大。

### 1.2 现有基础设施

| 已存在 | 详述 |
|--------|------|
| `NODE_TYPE=slave` | `config.IsMasterNode` 已定义，slave 跳过 DB migration |
| Redis 集群支持 | Sentinel + Cluster 模式，限流已用 Redis |
| Cookie session | `cookie.NewStore()` — 天然无状态 |
| 批量写入 | `BatchUpdateEnabled` — 减少 DB 写入频率 |
| 渠道缓存同步 | `SyncChannelCache` — 定时从 DB 刷缓存 |

### 1.3 现有缺失

| 缺失 | 影响 |
|------|------|
| Token 无 `UpdatedAt` | 无法增量同步 |
| 用户无 `UpdatedAt` | 同上 |
| 计费固定写 DB | 子节点必须直连主库，不能本地处理 |
| 无级联认证 | 子节点无法安全地调用主节点 API |
| 无去重机制 | 批量回传可能重复计费 |

---

## 2. 架构决策

### 2.1 通信方式：HTTP REST → 选择理由

| 方案 | 优势 | 劣势 | 结论 |
|------|------|------|------|
| gRPC Streaming | 实时双向流 | 增加 protobuf 编译链，子节点部署复杂 | ❌ |
| HTTP REST + 轮询 | 零依赖，复用 Gin | 延迟略高，但计费延迟可接受 | ✅ |
| 消息队列 (Kafka/RabbitMQ) | 高吞吐解耦 | 增加运维成本，全球部署需 MQ 集群 | ❌ 留待 v2 |

**决策**：HTTP REST。原因是：
- QuantumClaw 已有完整的 Gin 基础设施
- 子节点就是 QuantumClaw 实例，复用一切
- 计费和配置同步是批量/定时操作，不是实时流

### 2.2 数据同步策略：拉模式

```
主节点  ←── 子节点定时拉取 Token/User 增量
主节点  ←── 子节点批量推送计费记录

为什么不是推模式？
- 主节点不知道子节点什么时候上线
- 子节点数量不确定，主节点推 = 维护连接池
- 拉模式更健壮：子节点断线恢复后自动补全增量
```

### 2.3 计费一致性：最终一致性

```
请求到达子节点
  → 子节点预扣本地缓存中的 Token 配额（乐观扣减）
  → 转发完成
  → 计费记录加入本地队列（内存 + Redis backup）
  → 批量回传给主节点（每 10s 或每 100 条）
  → 主节点验证 + 入库
  → 主节点扣除真实配额

风险：子节点挂了，部分计费记录丢失
缓释：Redis 持久化队列，子节点重启后重发未确认的记录
```

### 2.4 Token 验证流程（子节点版）

```
 ┌── Request ──→
 │
 ├─ Bearer Token → SHA256 → key_hash
 │
 ├─ Redis: token:{key_hash}
 │   ├─ 命中 → 验证状态/过期/配额 ← 本地决策
 │   └─ 未命中 → HTTP GET 主节点 /api/cascade/tokens/sync?since=
 │       → 批量拉取全部增量
 │       → 写入本地 Redis (TTL 300s)
 │       → 重新查询 Redis
 │
 ├─ 验证通过 → c.Set("token_id", ...) → 继续转发
 │
 └─ 验证失败 → 返回 401
```

**关键设计**：
- 子节点本地 Redis 持有 Token 写时缓存（TTL 300s）
- 缓存穿透才回源查主节点
- 主节点批量返回增量（基于 `UpdatedAt`），减少小请求

---

## 3. 模型设计

### 3.1 新增：CascadeNode（主节点表）

```go
type CascadeNode struct {
    Id            int     `json:"id" gorm:"primaryKey"`
    Name          string  `json:"name" gorm:"type:varchar(64);uniqueIndex"`   // "Tokyo-A"
    Region        string  `json:"region" gorm:"type:varchar(32);index"`       // "ap-northeast-1"
    APIKeyHash    string  `json:"-" gorm:"type:char(64);uniqueIndex"`         // bcrypt 哈希
    APIKeyPrefix  string  `json:"api_key_prefix" gorm:"type:char(8)"`         // qcn_xxxx 的可见前缀
    Status        int     `json:"status" gorm:"default:1"`                    // 1=在线 2=离线 3=停用
    LastHeartbeat int64   `json:"last_heartbeat" gorm:"bigint"`
    Version       string  `json:"version" gorm:"type:varchar(32)"`
    ChannelCount  int     `json:"channel_count" gorm:"default:0"`
    TodayCalls    int64   `json:"today_calls" gorm:"default:0"`              // 当日调用量（子节点心跳会上报）
    CreatedTime   int64   `json:"created_time" gorm:"bigint"`
    UpdatedTime   int64   `json:"updated_time" gorm:"bigint"`
}

// 子节点注册时生成 API Key
// 格式: qcn_<random_32_hex>
// 只存 bcrypt 哈希 + 前 8 位可见前缀（给管理员识别）
```

### 3.2 新增：CascadeBillingBatch（主节点表）

```go
type CascadeBillingBatch struct {
    Id              int     `json:"id" gorm:"primaryKey"`
    BatchID         string  `json:"batch_id" gorm:"type:char(36);uniqueIndex"` // UUID
    NodeID          int     `json:"node_id" gorm:"index"`
    CreatedAt       int64   `json:"created_at" gorm:"bigint"`
    ConfirmedAt     *int64  `json:"confirmed_at" gorm:"bigint"`               // 主节点确认时间
    RecordsJSON     string  `json:"records" gorm:"type:longtext"`             // 压缩后的 JSON 数组
    RecordCount     int     `json:"record_count"`
    TotalAmount     float64 `json:"total_amount"`                             // 预汇总金额
    Status          int     `json:"status" gorm:"default:0"`                  // 0=待处理 1=已确认 2=拒绝
    IdempotencyKey string  `json:"idempotency_key" gorm:"type:varchar(64);uniqueIndex"`  // node_id + batch_uuid
}

// 每条批次的记录格式（存在 RecordsJSON 中）
type CascadeBillingRecord struct {
    Idempotency     string `json:"id"`         // sha256(token_id+model+tokens+timestamp) 防重
    TokenId         int    `json:"tid"`
    UserId          int    `json:"uid"`
    ModelName       string `json:"m"`
    PromptTokens    int    `json:"pt"`
    CompletionTokens int   `json:"ct"`
    Quota           int64  `json:"q"`          // 消费的配额（已经乘过 ratio）
    Timestamp       int64  `json:"ts"`
}
```

**为什么用批量而非单条**：
- 子节点每 10s 产生 ~50 条记录
- 单条 HTTP 调用开销太大
- 预汇总 `TotalAmount` 加 `IdempotencyKey` 保证幂等

### 3.3 修改：Token 模型

```go
// 新增字段
type Token struct {
    // ... 已有字段 ...
    UpdatedTime int64 `json:"updated_time" gorm:"bigint"` // 新增：用于增量同步
}

// 所有修改 Token 的地方（状态/配额/删除）都要更新 UpdatedTime
// 修改点：
//   model/token.go: UpdateStatusByKey(), UpdateTokenUsedQuota()
//   controller/token.go:  UpdateToken()
//   controller/user.go:    DeleteToken()
```

### 3.4 修改：User 模型

```go
// 新增字段
type User struct {
    // ... 已有字段 ...
    UpdatedTime int64 `json:"updated_time" gorm:"bigint"` // 新增
}
```

---

## 4. 子节点数据流时序图

```
子节点启动
  │
  ├─ 读取配置 NODE_TYPE=slave, CASCADE_MASTER_URL
  │
  ├─ POST /api/cascade/node/register
  │   → 主节点返回 node_id + api_key
  │   → 写入本地文件: .cascade_key
  │
  ├─ 启动定时器:
  │   ├─ heartbeat (每 30s)
  │   │   POST /api/cascade/node/heartbeat
  │   │   → 主节点更新 LastHeartbeat
  │   │
  │   ├─ token_sync (每 60s, 首次全量)
  │   │   GET /api/cascade/tokens/sync?since=<token.UpdatedTime>
  │   │   → 主节点返回增量的 Token 列表
  │   │   → 写入本地 Redis, TTL=3600s
  │   │
  │   ├─ billing_flush (每 10s 或每 100 条)
  │   │   POST /api/cascade/billing/batch
  │   │   → 主节点返回 accepted/rejected 计数
  │   │   → 清除已确认的队列
  │   │
  │   └─ config_sync (每 300s)
  │       GET /api/cascade/config
  │       → 拉取定价表、模型映射
  │       → 写入本地缓存
  │
  └─ 进入服务模式
      └─ 等待 API 请求
```

---

## 5. 计费流程对比

### 5.1 主节点计费（当前流程）

```
转发完成
  → PostConsumeDeduct()
  → DB 事务:
     1. token.RemainQuota -= quota
     2. token.UsedQuota += quota
     3. 创建 Log 记录
     4. 用户余额变更（如有）
  → 全部在同一个 DB 连接中完成
```

### 5.2 子节点计费（级联流程）

```
转发完成
  → PostConsumeDeduct() [子节点]
  → 仅操作本地 Redis:
     1. token:{hash}.remain_quota -= quota（乐观扣减）
     2. 计费记录加入本地队列
     3. 不写本地 DB
  → 定时回传:
     1. 打包队列 (最多 100 条)
     2. POST /api/cascade/billing/batch
     3. 主节点写入 DB
     4. 主节点返回确认
     5. 子节点标记已确认

冲突处理:
  子节点乐观扣减后，主节点发现配额不足
  → 主节点拒绝该批次
  → 子节点将该用户标记为"配额争议"
  → 下次同步时纠正
```

### 5.3 精度保证

```
问题: 子节点和主节点之间的配额可能不一致
      
场景:
  用户有 1000 配额
  子节点 A 乐观扣减 500（本地）
  子节点 B 乐观扣减 600（本地）
  主节点收到 batches: A=500, B=600
  总额 1100 > 1000 → 超额！

解决方案: 最终一致 + 修正机制
  - 子节点预扣时记录到 Redis（带 TTL）
  - 主节点汇总时发现超额 → 拒绝后面到达的 batch
  - 子节点下次 sync 时修正配额（从主节点拉取最新配额）
  - 用户看到的是最终一致的值
  - 重要: 渠道商场景下，超额通常很少发生（用户购买时已充值）
```

---

## 6. 安全性分析

| 威胁 | 防护 |
|------|------|
| 子节点伪造计费 | 子节点必须提供 `node_id` + API Key 签名 |
| API Key 泄露 | 信道 HTTPS 加密；Key 可撤销；bcrypt 存储 |
| 子节点窃取用户数据 | 级联 API 只返回最小必要字段（key_hash, 状态, 配额），不返回 key 原文 |
| 子节点被反向入侵 | 子节点只有级联的读写权限，无法删除/修改主节点数据 |
| 重放攻击 | `idempotency_key`（含时间戳 + 随机数），主节点 5 分钟内去重 |

---

## 7. 任务拆分与执行顺序

```
          ┌─────────────────────────────────┐
          │  任务 C1: 级联协议 + 模型定义     │ ← 先做，无依赖
          │  model/cascade_node.go           │
          │  model/cascade_billing.go        │
          │  common/cascade/contract.go      │
          │  model/main.go (AutoMigrate)     │
          └────────────┬────────────────────┘
                       │
          ┌────────────┴────────────────────┐
          │  任务 C2: Token/User UpdatedAt   │ ← 与 C1 并行
          │  model/token.go (+字段)          │
          │  model/user.go (+字段)           │
          │  model/main.go (手动迁移)        │
          │  审计所有写 Token 的地方          │
          └────────────┬────────────────────┘
                       │
          ┌────────────┴────────────────────┐
          │  任务 C3: 主节点级联 API          │ ← 依赖 C1, C2
          │  controller/cascade.go           │
          │  router/api.go (路由注册)         │
          │  model/cascade_node.go (CRUD)    │
          │  model/cascade_billing.go (写入)  │
          └────────────┬────────────────────┘
                       │
          ┌────────────┴──────────────────────────────┐
          │  任务 C4: 子节点级联客户端                  │ ← 依赖 C3（核心）
          │  service/cascade_client.go                 │
          │  service/cascade_billing_buffer.go         │ ← 计费缓冲队列
          │  middleware/auth.go (子节点 Token 验证)     │
          │  service/cash_billing.go (计费分支)         │
          └────────────┬──────────────────────────────┘
                       │
          ┌────────────┴─────────────────┐
          │  任务 C5: 子节点启动适配       │ ← 依赖 C4
          │  main.go (NODE_TYPE=slave)   │
          │  common/config/config.go     │
          │  model/main.go (跳过逻辑)     │
          └────────────┬─────────────────┘
                       │
          ┌────────────┴─────────────────┐
          │  任务 C6: 管理前端面板         │ ← 依赖 C3
          │  web/.../cascade/            │
          └────────────┬─────────────────┘
                       │
          ┌────────────┴─────────────────┐
          │  任务 C7: Docker 部署脚本      │ ← 依赖 C5
          │  deploy/slave/               │
          └──────────────────────────────┘
```

### 并行路线

```
Week 1:  C1 + C2 并行 → C3
         ├─ 第一天: 模型 + 契约
         ├─ 第二天: UpdatedAt + 审计写 Token 的地方
         └─ 第三~四天: 主节点 API

Week 2:  C4（核心）
         ├─ 第一天: 客户端框架 + 注册/心跳/Token 同步
         ├─ 第二天: 计费缓冲队列 + 批量回传
         └─ 第三天: 子节点 Token 验证 + auth 分支

Week 3:  C5 + C6 + C7 并行
         ├─ C5: 启动适配 1 天
         ├─ C6: 前端 2 天
         └─ C7: 部署脚本 1 天
```

---

## 8. 风险应对

| 风险 | 概率 | 影响 | 应对 |
|------|:----:|:----:|------|
| 子节点计费丢失 | 中 | 高 | Redis 持久化队列 + 启动重发 + 主从对账 |
| Token 同步延迟 | 中 | 中 | 主节点强制 TTL 短的缓存 + 请求级回查 |
| 跨区域网络不稳定 | 高 | 中 | 心跳超时自动标记离线，子节点独立运行 |
| 配额超分 | 低 | 中 | 最终一致性 + 修正机制 + 告警 |
| 级联 API 性能瓶颈 | 低 | 高 | 主节点按 node_id 分片 + Redis 缓存 Token 快照 |

---

## 9. 验收标准

- [ ] 主节点注册子节点，生成 API Key
- [ ] 子节点心跳上报，主节点显示在线/离线
- [ ] Token 增量同步：子节点创建 Token → 主节点 60s 内可查
- [ ] 计费回传：子节点 100 次 API 调用 → 主节点正确汇总
- [ ] 配额修正：主节点发现超额 → 子节点自动修正
- [ ] 子节点独立运行：主节点下线 → 子节点继续服务（只读状态）
- [ ] 子节点恢复：主节点恢复 → 自动重连 + 补发未确认计费
- [ ] 全球部署：主节点在新加坡，子节点在东京 → 端到端 < 100ms
