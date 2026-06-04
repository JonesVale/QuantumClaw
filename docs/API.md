# QuantumClaw API 开发者文档

> 本文档面向对接 QuantumClaw 的开发者。
> 所有 API 接口均遵循 RESTful 风格，返回 JSON 格式。

---

## 一、API Key（认证）

### Key 格式

QuantumClaw 遵循 OpenAI API Key 规范：

```
sk-<48位字母数字>
```

例如：`sk-aB3…wP1c...`

### 获取 Key

1. 登录管理后台 → 「令牌」页 → 创建令牌
2. 或调用 API：`POST /api/token/`
3. 新创建的 Key 只在创建时返回一次，请妥善保存

### 使用 Key

所有 API 请求需要在 HTTP Header 中携带：

```
Authorization: Bearer ***<your_key_here>
```

### Key 的属性

| 属性 | 说明 |
|------|------|
| `name` | 自定义名称，便于管理 |
| `status` | 1=启用, 2=禁用, 3=已过期, 4=额度耗尽 |
| `remain_quota` | 剩余额度（分），-1 表示无限 |
| `unlimited_quota` | 是否无限额度 |
| `expired_time` | 过期时间戳，-1=永不过期 |
| `models` | 可用模型列表（逗号分隔），空=全部 |
| `subnet` | IP 白名单（CIDR 格式），空=不限制 |

---

## 二、API 概览

> 📍 **说明**：系统有 **200+ API 端点**，下表列出最常用的部分。
> 完整的 API 清单请访问 Swagger UI：`/api/swagger/index.html`

### 2.1 OpenAI 兼容接口（中继转发）

| 端点 | 方法 | SDK 覆盖 |
|------|------|----------|
| `/v1/chat/completions` | POST | ✅ Python/Node/Go |
| `/v1/completions` | POST | ❌ |
| `/v1/models` | GET | ✅ Python/Node/Go |
| `/v1/models/:model` | GET | ❌ |
| `/v1/embeddings` | POST | ❌ |
| `/v1/images/generations` | POST | ❌ |
| `/v1/audio/transcriptions` | POST | ❌ |
| `/v1/audio/translations` | POST | ❌ |
| `/v1/audio/speech` | POST | ❌ |
| `/v1/moderations` | POST | ❌ |
| `/v1/files` | GET/POST | ❌ |
| `/v1/fine_tuning/jobs` | POST/GET | ❌ |
| `/v1/assistants` | POST/GET | ❌ |
| `/v1/threads` | POST/GET | ❌ |
| `/v1/messages` | POST | ❌ |
| `/v1/batches` | POST/GET | ❌ |
| `/v1/vector_stores` | POST/GET | ❌ |
| `/v1/mj/*` | ANY | ❌ |
| `/v1/video/*` | ANY | ❌ |
| `/v1/suno/*` | ANY | ❌ |

> 以上 OpenAI 兼容接口可直接使用 OpenAI 官方 SDK，无需 QuantumClaw SDK 支持。

### 2.2 用户自有 API（UserAuth）

| 端点 | 方法 | SDK 覆盖 | 说明 |
|------|------|----------|------|
| `/api/user/self` | GET | ✅ | 当前用户信息 |
| `/api/user/self/` | PUT | ❌ | 更新个人信息 |
| `/api/user/self/balance` | GET | ✅ | 查询余额 |
| `/api/user/self/dashboard` | GET | ✅ | 用户仪表盘 |
| `/api/user/self/available_models` | GET | ✅ `get_available_models()` | 可用模型列表 |
| `/api/user/self/token` | GET | ❌ | 生成 Access Token |
| `/api/user/self/billing/stats` | GET | ✅ | 计费统计 |
| `/api/user/self/billing/records` | GET | ✅ | 计费记录 |
| `/api/user/self/transaction_logs` | GET | ✅ `get_transaction_logs()` | 交易明细 |
| `/api/user/self/topup/list` | GET | ✅ `get_topup_list()` | 充值记录 |
| `/api/user/self/subscription/plans` | GET | ✅ | 订阅套餐 |
| `/api/user/self/subscription/self` | GET | ✅ | 我的订阅 |
| `/api/user/self/checkin` | GET | ✅ | 签到状态 |
| `/api/user/self/checkin` | POST | ❌ | 执行签到 |
| `/api/user/self/notifications` | GET | ✅ | 通知列表 |
| `/api/user/self/team` | GET | ❌ | 我的团队 |
| `/api/user/self/aff` | GET | ❌ | 推广码 |
| `/api/user/self/security/activity` | GET | ❌ | 安全活动 |
| `/api/user/self/webauthn/credentials` | GET | ❌ | WebAuthn 凭证 |
| `/api/log/self` | GET | ✅ | 自己的请求日志 |
| `/api/log/self/stat` | GET | ✅ | 日志统计 |

### 2.3 API Key 管理

| 端点 | 方法 | SDK 覆盖 |
|------|------|----------|
| `/api/token/` | GET | ✅ |
| `/api/token/` | POST | ✅ |
| `/api/token/search` | GET | ✅ |
| `/api/token/:id` | GET | ✅ |
| `/api/token/` | PUT | ✅ |
| `/api/token/:id` | DELETE | ✅ |

### 2.4 渠道管理

| 端点 | 方法 | SDK 覆盖 |
|------|------|----------|
| `/api/channel/` | GET | ✅ |
| `/api/channel/` | POST | ✅ |
| `/api/channel/` | PUT | ✅ |
| `/api/channel/:id` | DELETE | ✅ |
| `/api/channel/:id` | GET | ✅ |
| `/api/channel/types` | GET | ✅ |
| `/api/channel/test/:id` | GET | ❌ |
| `/api/channel/test` | GET | ❌ |

### 2.5 管理员接口（AdminAuth）

| 端点 | 方法 | SDK 覆盖 |
|------|------|----------|
| `/api/user/` | GET | ✅ |
| `/api/user/search` | GET | ✅ |
| `/api/user/:id` | GET | ✅ |
| `/api/user/` | POST | ✅ |
| `/api/user/` | PUT | ✅ |
| `/api/user/:id` | DELETE | ✅ |
| `/api/user/add_balance` | POST | ✅ |
| `/api/log/` | GET | ✅ |
| `/api/log/stat` | GET | ✅ |
| `/api/log/search` | GET | ✅ |
| `/api/option/` | GET | ✅ |
| `/api/option/` | PUT | ❌ |
| `/api/redemption/` | GET | ✅ |
| `/api/redemption/` | POST | ✅ |
| `/api/redemption/:id` | DELETE | ❌ |
| `/api/transactions` | GET | ✅ `get_transactions()` |

### 2.6 系统公共接口

| 端点 | 方法 | SDK 覆盖 |
|------|------|----------|
| `/api/status` | GET | ✅ |
| `/api/site-content` | GET | ✅ |
| `/api/model-catalog` | GET | ✅ |
| `/api/models` | GET | ❌（等待 API 文档） |
| `/api/models/rankings` | GET | ❌ |

---

## 三、SDK 覆盖说明

当前 SDK 覆盖约 **40 个核心端点**，占总 API（200+）的 **~20%**。

**覆盖原则**：
- ✅ OpenAI 兼容中继：基础聊天 + 模型列表（其它可直接用 OpenAI 官方 SDK）
- ✅ Key 管理：完整 CRUD
- ✅ 用户信息/余额/计费
- ✅ 渠道管理：完整 CRUD
- ✅ 管理员用户管理
- ✅ 日志查询
- ❌ 异步任务（Midjourney/视频/音乐）：AI 名未封装
- ❌ 高级中继（Assistants/Threads/Batch/Vector Stores）：使用 OpenAI 官方 SDK
- ❌ 预约/WebAuthn/OAuth：使用后台管理

---

## 四、SDK 快速开始

### Python

```python
from quantumclaw import QuantumClaw

# 初始化客户端
client = QuantumClaw(
    api_key=***
    base_url="https://your-instance.com"
)

# 聊天补全
response = client.chat_completions(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response["choices"][0]["message"]["content"])

# 流式聊天
for chunk in client.stream_chat(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hi!"}]
):
    if chunk.get("choices"):
        print(chunk["choices"][0].get("delta", {}).get("content", ""), end="")

# 查询余额
balance = client.check_balance()
print(f"余额: {balance}")

# 创建 API Key
token = client.create_token(name="dev-key", remain_quota=100000)
print(f"新 Key: {token['key']}")

# 列出 API Key
tokens = client.list_tokens()

# 管理员操作
users = client.admin_list_users()
channel = client.add_channel(name="OpenAI", type=1, key="sk-xxx...")
```

### Node.js

```typescript
import { QuantumClaw } from 'quantumclaw';

const client = new QuantumClaw({
    apiKey: '***',
    baseURL: 'https://your-instance.com'
});

// 聊天补全
const response = await client.chatCompletions({
    model: 'gpt-4o',
    messages: [{ role: 'user', content: 'Hello!' }]
});

// 流式聊天
const stream = client.streamChat({
    model: 'gpt-4o',
    messages: [{ role: 'user', content: 'Hi!' }]
});
for await (const chunk of stream) {
    process.stdout.write(chunk.choices?.[0]?.delta?.content || '');
}
```

### Go

```go
import "github.com/quantumclaw/quantumclaw-sdk-go/quantumclaw"

client := quantumclaw.NewClient("***", "https://your-instance.com")
ctx := context.Background()

resp, err := client.ChatCompletion(ctx, &quantumclaw.ChatCompletionRequest{
    Model:    "gpt-4o",
    Messages: []quantumclaw.ChatMessage{{Role: "user", Content: "Hello!"}},
})
```

---

## 五、API 约定

### 请求格式

管理 API 统一使用 JSON body。

### 响应格式

成功：
```json
{ "success": true, "message": "", "data": { ... } }
```

失败：
```json
{ "success": false, "message": "错误描述" }
```

OpenAI 兼容接口的错误格式：
```json
{ "error": { "message": "错误描述", "type": "quantumclaw_error" } }
```

### 错误码

| HTTP | 说明 |
|------|------|
| 200 | 成功（检查 success 字段） |
| 401 | 未认证 |
| 403 | 权限不足 |
| 429 | 频率限制 |
| 500 | 服务器错误 |
| 503 | 无可用渠道 |

---

## 六、流式（SSE）

完全兼容 OpenAI SSE 协议：

```
data: {"id":"...","choices":[{"delta":{"content":"Hello"}}]}
data: {"id":"...","choices":[{"delta":{"content":" world"}}]}
data: [DONE]
```

---

> 完整的 API 映射查看：`http://your-instance/api/swagger/index.html`
> 全部 200+ 端点：`/api/swagger/doc.json`
