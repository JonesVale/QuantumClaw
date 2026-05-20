# 量子算力资源池 — 实现计划

## 概览

```
改动策略：纯增量。不重构任何现有代码，只在存量架构上叠加。
核心模式：量子算力 = 第二类 Channel (type >= 100)
```

---

## 阶段 1 — 类型枚举 + 数据流（无代码风险）

### 1.1 扩展 channeltype 枚举

**文件**: `relay/channeltype/define.go`

```go
const (
    // AI 渠道 (现有, 0~60)
    Unknown = iota     // 0
    OpenAI             // 1
    // ...
    Dummy              // 61 — 保持最后一个

    // 量子算力渠道 (type >= 100)
    QuantumFirst = 100
    IonQ         = 100
    IBMQ         = 101
    Rigetti      = 102
    AWSBraket    = 103
    AzureQuantum = 104
    GoogleSycamore = 105
    QuantumDummy    = 106 // 用于 ChannelBaseURLs 长度检查
)
```

改动量: 10 行 | 风险: 低 — 现有 AI 渠道枚举值不动，QuantumFirst=100 干净分隔

### 1.2 扩展 apitype 枚举

**文件**: `relay/apitype/define.go`

```go
const (
    // 现有 AI 类型 (iota 0~19)
    OpenAI = iota
    // ...
    Dify

    // 量子算力类型
    IONQ
    IBMQ
    RIGETTI
    AWS_BRAKET
    AZURE_QUANTUM
    GOOGLE_SYCAMORE
    QUANTUM_DUMMY
)
```

改动量: 8 行 | 风险: 低

### 1.3 扩展 ChannelBaseURLs

**文件**: `relay/channeltype/url.go`

在现有列表后面追加量子平台 URL，长度对齐 `QuantumDummy`。
同时更新 `ValidateChannelBaseURLs()` 的地址验证逻辑（需兼容 `type >= 100` 的 Dummy 计数）。

改动量: ~10 行 | 风险: 低

### 1.4 实现 ToAPIType 映射

**文件**: `relay/channeltype/api.go`

```go
func ToAPIType(channelType int) int {
    if channelType >= 100 {
        // 量子算力映射
        switch channelType {
        case IonQ:            return apitype.IONQ
        case IBMQ:            return apitype.IBMQ
        case Rigetti:         return apitype.RIGETTI
        case AWSBraket:       return apitype.AWS_BRAKET
        case AzureQuantum:    return apitype.AZURE_QUANTUM
        case GoogleSycamore:  return apitype.GOOGLE_SYCAMORE
        }
    }
    // 现有 AI 映射（不动）
}
```

改动量: 15 行 | 风险: 低

### 1.5 写入约定单元测试

**文件**: `relay/channeltype/helper_test.go`

追加量子类型映射测试，验证每个量子渠道能正确映射到对应 apitype。

改动量: 20 行 | 风险: 零

---

## 阶段 2 — 量子适配器框架（核心）

### 2.1 统一量子请求/响应结构

**新建**: `relay/quantum/types.go`

```go
package quantum

// QuantumTaskRequest — 统一入参
type QuantumTaskRequest struct {
    Provider   string     `json:"provider,omitempty"`   // "auto" | "ionq" | "ibmq"
    Backend    string     `json:"backend,omitempty"`    // "auto" | "ionq_harmony" | "ibm_sherbrooke"
    Circuit    Circuit    `json:"circuit"`
    Shots      int        `json:"shots"`
    OptimizationLevel int `json:"optimization_level,omitempty"`
}

type Circuit struct {
    Qubits int     `json:"qubits"`
    Gates  []Gate  `json:"gates,omitempty"`
    // 也支持直接传 QASM
    QASM   string `json:"qasm,omitempty"`
}

type Gate struct {
    Name     string `json:"name"`
    Target   []int  `json:"targets"`
    Control  []int  `json:"controls,omitempty"`
    Params   []float64 `json:"params,omitempty"`
}

// QuantumTaskResult — 统一出参
type QuantumTaskResult struct {
    TaskID   string            `json:"task_id"`
    Status   string            `json:"status"`         // "queued" | "running" | "completed" | "failed"
    Provider string            `json:"provider"`
    Backend  string            `json:"backend"`
    Results  *TaskResults      `json:"results,omitempty"`
    Error    string            `json:"error,omitempty"`
    ExecTimeMs int64           `json:"execution_time_ms"`
}

type TaskResults struct {
    Counts        map[string]int `json:"counts"`
    Probabilities map[string]float64 `json:"probabilities,omitempty"`
    Shots         int            `json:"shots"`
}
```

改动量: 新建文件 ~80 行 | 风险: 低（纯 struct 定义）

### 2.2 量子适配器接口

**新建**: `relay/quantum/adaptor.go`

```go
type QuantumAdaptor interface {
    // 提交量子任务 — 返回 task_id
    RunTask(ctx context.Context, req *QuantumTaskRequest) (*QuantumTaskResult, error)
    // 查询任务状态
    QueryTask(ctx context.Context, taskID string) (*QuantumTaskResult, error)
    // 取消任务
    CancelTask(ctx context.Context, taskID string) error
    // 列出可用量子后端
    ListBackends(ctx context.Context) ([]string, error)
    // 获取适配器名（用于日志/计费标识）
    ProviderName() string
}
```

改动量: 新建文件 ~30 行 | 风险: 低

### 2.3 各平台适配器实现

**新建目录**: `relay/quantum/{ionq,ibmq,rigetti,braket}/`

每个适配器实现 `QuantumAdaptor` 接口，嵌入 API 密钥管理（复用 `meta.Meta`）：

```
relay/quantum/
├── types.go          ← 统一数据结构
├── adaptor.go        ← QuantumAdaptor 接口
├── converter.go      ← 统一格式 ↔ 各家格式 转换工具
├── ionq/
│   └── adaptor.go    ← IonQ API 实现
├── ibmq/
│   └── adaptor.go    ← IBM Q API 实现
├── rigetti/
│   └── adaptor.go    ← Rigetti API 实现
└── braket/
    └── adaptor.go    ← AWS Braket API 实现
```

每个适配器约 100~150 行，含 HTTP 请求 + 格式转换 + 错误处理。

改动量: 5 个新建文件 × ~120 行 = ~600 行 | 风险: 中 — 需实际 API 文档验证请求格式

### 2.4 格式转换器

**新建**: `relay/quantum/converter.go`

```go
// 统一电路 → IonQ JSON 格式
func CircuitToIonQJSON(circuit *Circuit) ([]byte, error)

// 统一电路 → IBM Q QASM 格式
func CircuitToIBMQQASM(circuit *Circuit) (string, error)

// IonQ 响应 → 统一结果
func IonQResponseToResult(raw []byte) (*QuantumTaskResult, error)

// IBM Q 响应 → 统一结果
func IBMQResponseToResult(raw []byte) (*QuantumTaskResult, error)
```

改动量: 新建文件 ~150 行 | 风险: 低（纯数据转换，无网络/业务依赖）

---

## 阶段 3 — 路由 + 控制器（最小胶水层）

### 3.1 新增 relay mode

**文件**: `relay/relaymode/define.go`

```go
const (
    // 现有...
    // 量子任务
    QuantumRun
    QuantumStatus
    QuantumCancel
    QuantumBackends
)
```

改动量: 4 行 | 风险: 低

### 3.2 新增量子路由

**新建**: `router/quantum.go`

```go
func SetQuantumRouter(router *gin.Engine) {
    // 复用 AI Relay 的 middleware 链
    quantumRouter := router.Group("/v1/quantum")
    quantumRouter.Use(
        middleware.RelayPanicRecover(),
        middleware.TokenAuth(),
        middleware.Distribute(),  // ← 复用同一分发中间件
    )
    {
        quantumRouter.POST("/run", controller.QuantumRelay)
        quantumRouter.GET("/status/:task_id", controller.QuantumRelay)
        quantumRouter.POST("/cancel/:task_id", controller.QuantumRelay)
        quantumRouter.GET("/backends", controller.QuantumRelay)
    }
}
```

改动量: 新建文件 ~25 行 | 风险: 低

### 3.3 注册路由

**文件**: `router/router.go`

在 `SetApiRouter` 或 `main.go` 中增加一行：`SetQuantumRouter(router)`

改动量: 1 行 | 风险: 低

### 3.4 新增量子控制器

**新建**: `controller/quantum.go`

```go
func QuantumRelay(c *gin.Context) {
    // 1. 获取 channel（复用 Distribute 中间件已选中的 channel）
    channel := c.Get(ctxkey.Channel)

    // 2. 获取 QuantumAdaptor（根据 channel type 查找对应适配器）
    adaptor := quantum.GetAdaptor(channel.Type)

    // 3. 解析请求 → 统一 QuantumTaskRequest
    req := parseQuantumRequest(c)

    // 4. 调用适配器
    result, err := adaptor.RunTask(c.Request.Context(), req)

    // 5. 处理结果
    // ...
}
```

复用 `controller/relay.go` 的自动重试逻辑（`shouldRetry` → 切 channel → 重试）。

改动量: 新建文件 ~80 行 | 风险: 低

---

## 阶段 4 — 分发 + 重试复用

### 4.1 middleware/Distribute 筛选量子 Channel

**文件**: `middleware/distributor.go`

现有逻辑：按 `group + model` 从 Ability 表匹配 Channel。
修改：加一个 `channelTypeRange` 判断，AI 请求只选 `type < 100`，量子请求只选 `type >= 100`。

```go
// 在中间件 context 里标记请求类型
c.Set("request_type", "ai")     // 或 "quantum"

// 在筛选时加过滤
if requestType == "quantum" {
    query = query.Where("type >= 100")
} else {
    query = query.Where("type < 100")
}
```

改动量: ~10 行 | 风险: 低

### 4.2 Ability 表支持量子后端

**文件**: `model/ability.go`

Ability 表用的 `model` 字段，量子场景下 `model = backend_name`（如 `ionq_harmony`）。
现有 Ability 模型无需改动，筛选时 `model` 匹配统一字段。

改动量: 0 行 | 风险: 零

---

## 阶段 5 — 计费扩展

### 5.1 量子计费 env 变量扩展

**文件**: `service/tiered_settle.go`

在 `TieredBillingContext` 中已有 `input_tokens`, `output_tokens` 等变量。
量子场景需要新变量：

```go
type TieredBillingContext struct {
    Group        string
    ModelName    string
    ChannelId    int

    // AI tokens (现有)
    InputTokens     int
    OutputTokens    int
    CacheHits       int
    CacheMisses     int

    // 量子算力 (新增)
    Qubits      int     `json:"qubits"`
    Shots       int     `json:"shots"`
    Gates       int     `json:"gates"`
}
```

对应的计费表达式示例：
```
// 按量子比特 × 采样次数 计费
qubits * shots * 0.001

// 按门数量 × 量子比特 × 采样次数 计费
gates * qubits * shots * 0.0001
```

改动量: ~10 行 | 风险: 低（纯字段添加，不影响现有逻辑）

### 5.2 扣费路径判断

**文件**: `service/quota.go`

在消耗 quota 时判断 channel type：

```go
func PreConsumeQuota(...) {
    if channelType >= 100 {
        // 用量子计费公式
        cost = service.CalculateQuantumCost(ctx)
    } else {
        // 用现有 AI 计费公式（不动）
        cost = service.CalculateAICost(ctx)
    }
}
```

改动量: ~5 行 | 风险: 低

---

## 阶段 6 — 管理后台（前端）

### 6.1 Channel 管理 — 量子 tab

**文件**: `web/default/src/.../ChannelManager.tsx`（或对应组件）

加一个 tab 筛选：「AI 渠道 / 量子算力」
- AI 渠道 → 列出现有 `type < 100` 的 channel
- 量子算力 → 列出 `type >= 100` 的 channel

后台 API 加一个查询参数 `?type_range=ai|quantum`。

改动量: 前端 ~50 行, 后端 ~5 行 | 风险: 低

### 6.2 量子计费规则页面

在现有「计费规则」页面加一个量子计费 tab：
- 计费变量可选 `qubits`, `shots`, `gates`
- 与 AI 计费规则共用同一套 `TieredBillingRule` 表

改动量: 前端 ~30 行 | 风险: 低

---

## 文件改动清单汇总

### 新增文件（8 个）

| 文件 | 行数估计 | 内容 |
|------|---------|------|
| `relay/quantum/types.go` | 80 | 统一量子请求/响应结构 |
| `relay/quantum/adaptor.go` | 30 | QuantumAdaptor 接口 |
| `relay/quantum/converter.go` | 150 | 统一 ↔ 各家格式转换 |
| `relay/quantum/ionq/adaptor.go` | 120 | IonQ API 适配器 |
| `relay/quantum/ibmq/adaptor.go` | 120 | IBM Q API 适配器 |
| `relay/quantum/rigetti/adaptor.go` | 120 | Rigetti API 适配器 |
| `relay/quantum/braket/adaptor.go` | 120 | AWS Braket 适配器 |
| `router/quantum.go` | 25 | 量子路由注册 |
| `controller/quantum.go` | 80 | 量子控制器 |

### 修改文件（10 个）

| 文件 | 改动 | 行数 |
|------|------|------|
| `relay/channeltype/define.go` | 加量子枚举 | +10 |
| `relay/apitype/define.go` | 加量子 API 类型 | +8 |
| `relay/channeltype/api.go` | ToAPIType 量子分支 | +15 |
| `relay/channeltype/url.go` | 加量子 BaseURL + 验证 | +10 |
| `relay/channeltype/helper_test.go` | 量子映射测试 | +20 |
| `relay/relaymode/define.go` | 加 QuantumRun 等 | +4 |
| `router/router.go` | 注册量子路由 | +1 |
| `service/tiered_settle.go` | 加 qubits/shots/gates 变量 | +10 |
| `service/quota.go` | 量子/AI 扣费分支 | +5 |
| `middleware/distributor.go` | 请求类型过滤 | +10 |
| 前端 Channel 管理页面 | 量子 tab | +80 |

### 总计

**新增**: ~845 行 Go | **修改**: ~170 行 | **总量**: ~1015 行纯增量，零破坏

---

## 执行顺序建议

1. **Phase 1** (类型枚举 + 数据流) — 30 分钟，可单独测试
2. **Phase 2** (适配器框架 + 1 个实际适配器) — 2 小时
3. **Phase 3** (路由 + 控制器) — 30 分钟
4. **Phase 4** (分发复用) — 15 分钟
5. **Phase 5** (计费) — 30 分钟
6. **Phase 6** (前端) — 1 小时
7. **集成测试 + 修复** — 1 小时

**总预估**: 5~6 小时
