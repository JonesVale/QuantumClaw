# QuantumClaw Python SDK

**管理工具包，不是 AI 调用库。**

---

## 这个 SDK 是给谁用的

你如果只是想**调 AI 模型**（GPT-4o、Claude、DeepSeek 等），直接用 OpenAI 官方 SDK 就好：

```python
from openai import OpenAI
client = OpenAI(
    api_key=***    # 你的 QuantumClaw API Key
    base_url="https://你的实例地址/v1"   # 改一下 base_url
)
```

这个 SDK 是给**管网关的人**用的——那些需要在代码里自动化管理 Key、查用量、配渠道的管理员。

---

## 适用场景

| 场景 | 适合用这个 SDK | 适合用 OpenAI SDK |
|------|:---:|:---:|
| 调 AI 聊天 | ❌ | ✅ |
| 流式聊天 | ❌ | ✅ |
| 调用 Embedding/Image/Audio | ❌ | ✅ |
| 批量创建 API Key 发给团队 | ✅ | ❌ |
| 查余额、看账单 | ✅ | ❌ |
| 管理上游 AI 厂商渠道 | ✅ | ❌ |
| 创建用户、分配额度 | ✅ | ❌ |
| 查请求日志、用量统计 | ✅ | ❌ |
| 配置系统参数 | ✅ | ❌ |

---

## 快速开始

```python
from quantumclaw import QuantumClaw

client = QuantumClaw(
    api_key=***    # 管理员 API Key（sk-开头）
    base_url="https://你的实例.com"
)

# 查余额
info = client.get_self_info()
balance = client.check_balance()

# 批量创建 API Key 分发给团队成员
team_keys = []
for member in ["alice", "bob", "charlie"]:
    token = client.create_token(
        name=f"{member}-dev",
        remain_quota=100000,           # 每人 10 万额度
        models="gpt-4o,gpt-3.5-turbo"  # 限制可用模型
    )
    team_keys.append({"name": member, "key": token["key"]})

# 查各部门用量
logs = client.get_logs(num=100)

# 列出所有用户（管理员）
users = client.admin_list_users()

# 给某个用户充值
client.admin_add_balance(user_id=2, amount=50000, remark="部门预算")
```

---

## API 列表

### 用户管理

| 方法 | 端点 | 说明 |
|------|------|------|
| `get_self_info()` | GET /api/user/self | 当前用户信息 |
| `check_balance()` | GET /api/user/self/balance | 余额查询 |
| `get_dashboard()` | GET /api/user/self/dashboard | 仪表盘数据 |
| `get_available_models()` | GET /api/user/self/available_models | 当前可用模型 |
| `get_billing_stats()` | GET /api/user/self/billing/stats | 计费统计 |
| `get_billing_records()` | GET /api/user/self/billing/records | 计费记录 |
| `get_transaction_logs()` | GET /api/user/self/transaction_logs | 交易明细 |
| `get_topup_list()` | GET /api/user/self/topup/list | 充值记录 |
| `get_subscription_plans()` | GET /api/user/self/subscription/plans | 订阅套餐 |
| `get_subscription_self()` | GET /api/user/self/subscription/self | 我的订阅 |
| `get_checkin_status()` | GET /api/user/self/checkin | 签到状态 |
| `get_notifications()` | GET /api/user/self/notifications | 通知列表 |

### API Key 管理（核心）

| 方法 | 说明 |
|------|------|
| `list_tokens()` | 列出所有 Key |
| `create_token()` | 创建 Key（支持额度、模型权限、过期时间） |
| `get_token()` | 查看单个 Key |
| `update_token()` | 更新 Key 设置（启停、额度、模型） |
| `delete_token()` | 删除 Key |
| `search_tokens()` | 搜索 Key |

### 渠道管理

| 方法 | 说明 |
|------|------|
| `list_channels()` | 列出所有上游渠道 |
| `get_channel()` | 查看渠道详情 |
| `add_channel()` | 添加上游厂商 |
| `update_channel()` | 更新渠道配置 |
| `delete_channel()` | 删除渠道 |
| `get_channel_types()` | 渠道类型列表 |

### 日志

| 方法 | 说明 |
|------|------|
| `get_logs()` | 自己的请求日志 |
| `get_logs_stat()` | 日志统计 |
| `admin_get_logs()` | 管理员：全部日志 |
| `admin_get_logs_stat()` | 管理员：日志统计 |
| `admin_search_logs()` | 管理员：搜索日志 |

### 管理员

| 方法 | 说明 |
|------|------|
| `admin_list_users()` | 用户列表 |
| `admin_search_users()` | 搜索用户 |
| `admin_get_user()` | 用户详情 |
| `admin_create_user()` | 创建用户 |
| `admin_update_user()` | 更新用户 |
| `admin_delete_user()` | 删除用户 |
| `admin_add_balance()` | 充值额度 |

### 系统

| 方法 | 说明 |
|------|------|
| `get_status()` | 系统状态 |
| `get_site_content()` | 站点配置 |
| `get_models_catalog()` | 模型目录 |

---

## 关于 Relay（AI 调用）

SDK 提供 `chat_completions()` 和 `stream_chat()` 方法，**主要用于测试连通性**。

生产环境调 AI 模型，请使用 OpenAI 官方 SDK：

```python
from openai import OpenAI

client = OpenAI(
    api_key="***",
    base_url="https://你的实例.com/v1"
)

# 一致的用法
response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello"}]
)
```

---

## 错误处理

```python
from quantumclaw import AuthenticationError, RateLimitError, InsufficientQuotaError

try:
    token = client.create_token(name="my-key", remain_quota=1000)
except InsufficientQuotaError:
    print("余额不足，请充值")
except RateLimitError:
    print("频率过高，稍后重试")
except AuthenticationError:
    print("API Key 无效")
```
