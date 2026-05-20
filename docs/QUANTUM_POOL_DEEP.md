# 量子算力资源池 — 深度架构分析

> 读完所有关键代码后的结论。

---

## 一、现有代码的完整请求生命周期

```
Client → POST /v1/chat/completions
  │
  ├─ 1. router/relay.go: SetRelayRouter
  │     注册 middleware 链:
  │     RelayPanicRecover → TokenAuth → ModelRateLimit → Distribute → Relay
  │
  ├─ 2. middleware/auth.go: TokenAuth
  │     ├─ 从 Header 取 Authorization: Bearer sk-xxx
  │     ├─ 查 Token 表 → 用户 ID / Token ID / 可用模型列表
  │     ├─ 调用 getRequestModel(c) ← 解析请求 body 提取 model 字段
  │     │    { "model": "gpt-4" } → 存入 ctxkey.RequestModel
  │     ├─ 校验：token 的 models 限制、子网限制、用户封禁
  │     └─ 设置 ctxkey.Id / TokenId / TokenName
  │
  ├─ 3. middleware/model_rate_limit.go: ModelRateLimit
  │     按 model + group 限流
  │
  ├─ 4. middleware/distributor.go: Distribute
  │     ├─ 从 ctx 取 RequestModel（由 TokenAuth 设置）
  │     ├─ 从 ctx 取 userId → 查询 userGroup
  │     ├─ 调用 model.CacheGetRandomSatisfiedChannel(group, model, false)
  │     │   → 查 Ability 表 → 取 highest priority + enabled channel
  │     ├─ SetupContextForSelectedChannel(c, channel, modelName)
  │     │   → 设置 ctxkey.Channel / ChannelId / ChannelName / BaseURL / Key / Config
  │     │   → 设置 Authorization: Bearer {channel.Key}
  │     └─ c.Next()
  │
  ├─ 5. controller/relay.go: Relay
  │     ├─ relayMode = relaymode.GetByPath(path)
  │     ├─ relayHelper(c, relayMode)
  │     │   └─ case ChatCompletions → controller.RelayTextHelper(c)
  │     │       ├─ 解析 GeneralOpenAIRequest（统一 OpenAI 格式）
  │     │       ├─ relay.GetAdaptor(meta.APIType) → switch 返回适配器
  │     │       ├─ adaptor.DoRequest → adaptor.DoResponse
  │     │       ├─ preConsumeQuota / postConsumeQuota
  │     │       └─ 返回 *model.ErrorWithStatusCode
  │     ├─ 如果 bizErr != nil → 进入重试循环
  │     │   for i := retryTimes; i > 0; i--:
  │     │     CacheGetRandomSatisfiedChannel → 选另一个 channel
  │     │     SetupContextForSelectedChannel → 切换
  │     │     relayHelper(c, relayMode) → 重试
  │     └─ 全部失败 → 返回最终错误
  │
  └─ 6. model/ability.go: CacheGetRandomSatisfiedChannel
        Ability 表: group + model + channel_id + enabled + priority
        WHERE group='{x}' AND model='{y}' AND enabled=1
        ORDER BY priority DESC, RANDOM()  → 加权随机
```

---

## 二、量子请求的生命周期（目标）

```
Client → POST /v1/quantum/run
  │
  ├─ 1. TokenAuth (复用)
  │     → 解析 Header, 查用户, 查 Token, 设 ctx
  │     → 需要扩展 getRequestModel(): 量子请求从 { "backend": "ionq_harmony" } 提取
  │     → shouldCheckModel(): 需要增加 /v1/quantum/* 前缀识别
  │     → Token 的 models 限制: 量子场景下 model = backend name，同样适用
  │
  ├─ 2. Distribute (复用, 0改动)
  │     → 从 RequestModel 查 Ability 表
  │     → Ability 表里 model = "ionq_harmony", group = 用户组
  │     → 选中 type >= 100 的 Channel
  │     → 如果用户 Token 的 models 限制了只能用量子后端，照常拦截
  │
  ├─ 3. controller/relay.go: Relay (复用重试循环, 0改动)
  │     → relayMode = QuantumRun
  │     → relayHelper(c, relayMode) → 进入新分支
  │     → 失败时重试循环自动切换 Channel（完全复用）
  │
  ├─ 4. relayHelper 调度 (新增 case)
  │     case relaymode.QuantumRun → controller.QuantumHelper(c)
  │     case relaymode.QuantumStatus → controller.QuantumStatusHelper(c)
  │
  └─ 5. controller/quantum.go: QuantumHelper (新增)
        ├─ 解析请求体 → QuantumTaskRequest（统一格式）
        ├─ converter.ToIonQJSON() 或 ToIBMQQASM() 适配各平台
        ├─ 构造 HTTP 请求 → 调用量子平台 API
        ├─ 返回 QuantumTaskResult（统一格式）
        ├─ 量子计费: qubits * shots * rate
        └─ 日志记录（复用 model.RecordConsumeLog）
```

---

## 三、关键发现与最优实现策略

### 发现 1: TokenAuth 的 shouldCheckModel() 需要扩展

`middleware/auth.go` 的 `shouldCheckModel()` 当前只识别 AI 路径：
```go
func shouldCheckModel(c *gin.Context) bool {
    // 只有 /v1/chat/completions, /v1/images, /v1/audio 等
}
```
量子路径 `/v1/quantum/run` 不在其中，因此 `getRequestModel()` 的错误会被忽略，RequestModel 为空 → Distribute 找不到 channel。

**最优方案**: 在 shouldCheckModel 中加 `/v1/quantum/*` 识别，或在路由组上设单独的 TokenAuth 中间件包装。

### 发现 2: getRequestModel() 的解析逻辑可以复用

`middleware/utils.go` 的 `getRequestModel()` 解析 `{ "model": "..." }` 字段。
量子请求里字段名不同（`backend: "ionq_harmony"`）。

**最优方案**: 另写 `getQuantumRequestModel()` 解析 `backend` 字段，注入 `ctxkey.RequestModel`。量子路由的 middleware 链用这个替代。

### 发现 3: Distribute 中间件完全通用

`Distribute()` 不关心 channel type。它只按 `group + RequestModel` 查 Ability 表。

只要 Quantum Channel 的 Ability 记录写的是 `model = "ionq_harmony"`，`Distribute()` 就能选中它。

**最优方案**: 0 改动。加个 `QuantumDistribute()` 直接复用其逻辑，或让现有 Distribute 跳过不存在于 AI 路径时的 model 空值检查。

### 发现 4: 重试循环天然支持量子

`controller/relay.go` 的 `Relay()` 函数里：
```go
for i := retryTimes; i > 0; i-- {
    channel, err := model.CacheGetRandomSatisfiedChannel(group, originalModel, ...)
    middleware.SetupContextForSelectedChannel(c, channel, originalModel)
    relayHelper(c, relayMode)  // ← relayMode 是 QuantumRun
}
```
`relayMode` 传到了 `relayHelper`，而 `relayHelper` 会调 `QuantumHelper`。
量子请求重试时会选另一个量子 Channel，完全不用改 `Relay()` 逻辑。

**最优方案**: 0 改动。量子 handler 的生成逻辑跟 AI 完全一样。

### 发现 5: 计费复用度比预期高

当前计费体系：
- `preConsumeQuota`: 预扣（根据 promptTokens + maxTokens × ratio）
- `postConsumeQuota`: 实扣（根据 actual usage）
- `tiered_settle.go`: expr-lang 表达式计算

量子只需在 `TieredBillingContext` 加 3 个变量（qubits, shots, gates），另写一个 `QuantumPreConsumeQuota` 和 `QuantumPostConsumeQuota`，调用相同的 `model.CacheDecreaseUserQuota` / `model.UpdateUserUsedQuotaAndRequestCount`。

**最优方案**: 新增量子计费函数，复用底层 DB 写入逻辑。

### 发现 6: retry 时的 channel 类型隔离是个潜在问题

如果 AI 请求的 retry 选到了一个 type=100+ 的量子 Channel，会崩溃（因为 AI adaptor 不认识量子 URL）。

现有代码用 `CacheGetRandomSatisfiedChannel(group, model, ...)` 来选，其中 `model` 决定了路由。如果能确保 AI 模型名（"gpt-4"）只会匹配 AI Channel 的 Ability，量子后端名只会匹配量子 Channel 的 Ability，那就不会混。

**但实际上**，如果同一个组里同时有 AI 和量子 Channel，且存在 Ability 表中 model 为空的通配记录，就可能串到。

**最优方案**: Distribute 在筛选时加 `channel.type < 100` / `>= 100` 过滤。

---

## 四、改动清单（最优版）

### 改动要求
- **不改**现有的 AI relay 流程
- **不改** `controller/relay.go` 的重试循环
- **不改** `model/channel.go` 的结构
- **不改** `middleware/distributor.go` 的核心逻辑
- **不改** `service/billing.go` 的底层写入
- **不改** `service/tiered_settle.go` 的现有变量

### 纯增量文件

| 文件 | 代码量 | 目的 |
|------|--------|------|
| `relay/quantum/types.go` | 60 | 统一量子请求/响应结构 |
| `relay/quantum/adaptor.go` | 25 | QuantumAdaptor 接口 + 工厂函数 |
| `relay/quantum/converter.go` | 150 | 统一电路 → 各平台格式转换 |
| `relay/quantum/ionq/adaptor.go` | 120 | IonQ 平台适配器 |
| `relay/quantum/ibmq/adaptor.go` | 120 | IBM Q 平台适配器 |
| `controller/quantum.go` | 120 | 量子控制器 + 计费 + 日志 |
| `router/quantum.go` | 30 | 量子路由注册 |

### 修改文件（最小入侵）

| 文件 | 改动内容 | 行数 |
|------|---------|------|
| `relay/channeltype/define.go` | 追加 `IonQ=100` ~ `QuantumDummy` | +7 |
| `relay/apitype/define.go` | 追加 `IONQ` ~ `QUANTUM_DUMMY` | +8 |
| `relay/channeltype/url.go` | 追加量子 BaseURL，更新校验 | +10 |
| `relay/channeltype/api.go` | ToAPIType 新增量子分支 | +15 |
| `relay/relaymode/define.go` | 追加 `QuantumRun` 等 4 个 mode | +4 |
| `router/relay.go` | `relayHelper` 加量子 case 分支 | +5 |
| `router/main.go` | 注册 `SetQuantumRouter` | +1 |
| `middleware/auth.go` | `shouldCheckModel` 识别量子路径 | +2 |
| `middleware/utils.go` | `getRequestModel` 解析量子后端名 | +3 |
| `service/tiered_settle.go` | `TieredBillingContext` 加 3 字段 | +3 |
| `relay/channeltype/helper_test.go` | 追加量子映射测试 | +20 |

### 改动汇总

```
新增: 7 个文件, ~625 行
修改: 11 个文件, ~78 行
总计: ~700 行纯增量

零破坏：不删除、不改写任何现有逻辑
```

---

## 五、执行顺序（最优路径）

```
Phase 0: 枚举 + 数据层 (30 min)
  define.go + apitype + url + api + tests
  ├─ 可单独测试: 编译 + 映射测试
  └─ 安全: 不更改任何现有枚举值

Phase 1: 适配器框架 + IonQ 实现 (2 hr)
  types.go + adaptor.go + converter.go + ionq/adaptor.go
  ├─ 可单独测试: 格式转换 + 接口验证
  └─ 安全: 纯新类型，无引用路径

Phase 2: 路由 + 控制器 + 计费 (1.5 hr)
  relay.go + quantum.go + router + billing
  ├─ 可单独测试: 路由不冲突 + 计费计算
  └─ 安全: relayHelper 加 case, 不影响现有 case

Phase 3: 前端 Channel 管理 (1 hr)
  Type filter tab + 量子计费编辑器
  └─ 安全: 纯 UI 改动

Phase 4: 集成测试 + 修复 (1 hr)
  端到端: TokenAuth → Distribute → QuantumHelper → adapter → billing
```

### 为什么这个顺序最优？

1. **Phase 0** 回答「数据管道通不通」 — 确认 channel type 枚举和数据库层无误
2. **Phase 1** 回答「量子请求能不能发到平台」 — 最核心的风险点，先排除
3. **Phase 2** 回答「整个链路通不通」 — 仅胶水代码
4. **Phase 3** 回答「管理员能不能配置」 — 依赖 Phase 0 的数据通道
5. **Phase 4** 回答「全流程正确性」 — 最终的保障

每阶段都是可独立验证的增量，不存在大爆炸式集成。
