# QuantumClaw v2.2.0 — 全量审计报告

> 2026-06-07 | 基线: develop (HEAD~0)

---

## 一、消费链完整审计

### 1.1 完整链路图

```
用户请求 (REST/SSE)
  │
  ├── middleware.RelayPanicRecover()      ← 崩溃恢复
  ├── middleware.TokenAuth()              ← ① Token 鉴权
  ├── middleware.SearchMiddleware()       ← ② 搜索/路由
  ├── middleware.GeoMiddleware()          ← ③ 地理路由 (国内/国外)
  ├── middleware.PromptOptimizer()        ← ④ Prompt 优化
  ├── middleware.ParamValidator()         ← ⑤ 参数校验
  ├── middleware.IntelligentRouter()      ← ⑥ 智能路由上下文
  ├── middleware.Sub2APIRouter()          ← ⑦ 订阅 API 路由
  ├── middleware.ModelRateLimit()         ← ⑧ 模型级速率限制
  ├── middleware.Distribute()             ← ⑨ 渠道分发 (5 级路由)
  │                                          Step 0: 指定 channel_id
  │                                          Step 1: 首选供应商
  │                                          Step 2: 推广人渠道
  │                                          Step 3: 国内池兜底
  │                                          Step 4: 国外池兜底
  │                                          Step 5: 全部失败 → 503
  │
  ├── controller.relayHelper()
  │    ├── RelayTextHelper()              ← 文本聊天
  │    ├── RelayImageHelper()             ← 图片生成
  │    ├── RelayAudioHelper()             ← 音频 (TTS/ASR)
  │    ├── RelayFilesHelper()             ← 文件处理
  │    ├── RelayFineTuningHelper()        ← 微调
  │    └── RelayAsyncTask()              ← MJ/视频/Suno(异步)
  │
  ├── relay/common_handler/billing.go
  │    ├── PreConsumeBilling()            ← 预扣 (4 层优先级)
  │    │   Tier 1: 订阅余额
  │    │   Tier 2: CashBalance
  │    │   Tier 3: CommissionBalance
  │    │   Tier 4: Quota
  │    │   → 全部不足 → "insufficient balance"
  │    └── PostConsumeBilling()           ← 后扣/退款
  │
  ├── relay/adaptor/{provider}.go        ← 52 家适配器
  │    ↓
  ├── AI 上游 (API call)
  │    ↓
  └── PostConsumeBilling() + monitor.Emit()
       ↓
  ├── RewardInviterOnConsume()           ← 三级返佣 (5%/2%/1%)
  ├── RecordConsumeLog()                 ← 计费日志
  ├── monitor.Emit(channelId, success)    ← 健康监控
  └── 失败重试 (最多 config.RetryTimes 次)
       ├── GetCheapestSatisfiedChannel()
       ├── monitor.PenalizeModel()       ← 冷期惩罚
       └── 全部失败 → 5xx 错误返回
```

### 1.2 每级验证结果

| 阶段 | 组件 | 验证功能 | 状态 |
|:-----|:-----|:---------|:----:|
| ① | TokenAuth | API Key 鉴权 + Session 降级 | ✅ |
| ② | SearchMiddleware | 搜索/路由上下文 | ✅ |
| ③ | GeoMiddleware | 国内/国外流量分类 | ✅ |
| ④ | PromptOptimizer | Prompt 模板/历史优化 | ✅ |
| ⑤ | ParamValidator | max_tokens/temperature 等校验 | ✅ |
| ⑥ | IntelligentRouter | 智能路由分组 | ✅ |
| ⑦ | Sub2APIRouter | 子模型/订阅路由 | ✅ |
| ⑧ | ModelRateLimit | 每分钟/每小时/每日限制 | ✅ |
| ⑨ | Distribute | 5 级渠道分发 | ✅ |
| ⑩ | PreConsumeBilling | 4 层余额预扣 | ✅ |
| ⑪ | AI Adaptor | 52 家 Provider 适配 | ✅ |
| ⑫ | PostConsumeBilling | 精确结算 + 退款 | ✅ |
| ⑬ | Commission | 三级返佣自动触发 | ✅ |
| ⑭ | Monitor | 成功/失败计数 + 惩罚 | ✅ |
| ⑮ | Retry | 按价格重试 + 冷期 | ✅ |

**消费链结论：功能完整，15 个阶段全部实现。**

---

## 二、资源链完整审计

### 2.1 资源链路

```
Channel 创建
  ├── POST /api/channel (AdminAuth)
  │   ├── type → adaptor 映射
  │   ├── key → AES-GCM+Qpw 加密 (GORM BeforeSave)
  │   ├── model → 模型列表注册
  │   └── group → 用户分组绑定
  │
  ├── 渠道测试
  │   ├── POST /api/channel/test
  │   └── → 发测试请求 → 确认 200
  │
  ├── 健康监控
  │   ├── monitor.Emit(id, success)      ← 成功/失败计数
  │   ├── monitor/penalty.go            ← 失败惩罚
  │   │   ├── 连续 5 次失败 → 停用 5min
  │   │   ├── 1h 窗口 15%+ 失败 → 降权
  │   │   └── 观察期: 新 Provider 前 48h
  │   ├── 自动故障转移
  │   │   └── Distribute Step 3/4 → 池兜底
  │   └── 冷期
  │       └── PenalizeModel → 跳过
  │
  └── 分账
      ├── model/fee_config.go            ← 按阶梯收费
      └── AutoSettleMonthlyFees()        ← 月结
```

### 2.2 资源验证清单

| 资源 | 创建 | 读取 | 更新 | 删除 | 测试 | 状态 |
|:-----|:----:|:----:|:----:|:----:|:----:|:----:|
| Channel | ✅ | ✅ | ✅ | ✅ | ✅ | 完整 |
| API Token | ✅ | ✅ | ✅ | ✅ | ✅ | 完整 |
| User | ✅ | ✅ | ✅ | ✅ | ✅ | 完整 |
| Model | ✅ | ✅ | ✅ | N/A | ❌ | 🟡 无模型测试 |
| Channel Type | ✅ | ✅ | N/A | N/A | ✅ | 完整 |
| Billing Ratio | ✅ | ✅ | ✅ | N/A | ✅ | 完整 |
| Fee Config | ✅ | ✅ | ✅ | ✅ | ❌ | 🟡 新建未测试 |
| Failure Tracker | ✅ | ✅ | N/A | N/A | ❌ | 🟡 新建未测试 |
| Idempotency | ✅ | ✅ | ✅ | ✅ | ❌ | 🟡 新建未测试 |
| Store/Market | ✅ | ✅ | ✅ | ✅ | ✅ | 有 store_test |
| Cascade Node | ✅ | ✅ | ✅ | ✅ | ❌ | 🟡 无测试 |
| Commission | ✅ | ✅ | ✅ | ✅ | ❌ | 🟡 无端到端测试 |

---

## 三、未覆盖区域清单

### 3.1 测试覆盖率严重不足

| 区域 | 生产文件 | 测试文件 | 覆盖率 |
|:-----|:-------:|:--------:|:------:|
| middleware/ | 25 个 | 1 个(middleware_test.go, 单场景) | ~4% |
| relay/adaptor/ | 52 个适配器 | 1 个通用测试 | <2% |
| relay/controller/ | 13 个 handler | 0 | 0% |
| relay/quantum/ | 5 种量子后端 | 2 个测试 | 60% (但量子) |
| monitor/ | 3 个文件 (含 penalty) | 0 | 0% |
| service/ | ~15 个 | 3 个 | ~20% |
| model/ | ~67 个 | 8 个 | ~12% |
| **总计** | **~190 核心文件** | **~30** | **<10%** |

### 3.2 24 个中间件，0 个独立测试

所有关键中间件均无单元测试：

```
auth.go, cache.go, cascade_auth.go, cors.go, distributor.go,
geo.go, gzip.go, https_redirect.go, intelligent_router.go,
logger.go, login_rate_limit.go, model_rate_limit.go,
param_validator.go, payment_auth.go, prompt_optimizer.go,
rate-limit.go, recover.go, request-id.go, search.go,
security_headers.go, ssrf_protection.go, sub2api.go,
turnstile-check.go, webhook_security.go
```

单一 `middleware_test.go`（307 行）只测了 basic auth 流程。

### 3.3 52 个 AI 适配器，51 个无专用测试

唯一有测试的: `relay/adaptor/aws/llama3/main_test.go`

一个测试覆盖 52 个适配器的情况意味着：
- Key 验证错误可能在调用 20 次后才被发现
- 错误映射（各 Provider 返回不同 error 格式）可能断裂
- 新 API 参数变化无法被 CI 捕获

### 3.4 6 个新建模型无测试

| 文件 | 新增内容 | 测试 |
|:-----|:---------|:----:|
| `model/failure_tracker.go` | 渠道失败率滑动窗口 | ❌ |
| `model/fee_config.go` | 阶梯入驻费 | ❌ |
| `model/idempotency.go` | 支付幂等性 | ❌ |
| `model/consume_record.go` | 消费记录 | ❌ |
| `model/market_init.go` | 市场初始化 | ❌ |
| `model/pool_agreement.go` | 池协议 | ❌ |

### 3.5 资源残留/未清理

| 位置 | 问题 |
|:-----|:-----|
| `.env` `.env.bak` | 已在 `.gitignore`，但历史中有 |
| `quantumclaw.exe` (36MB) × 2 版本 | 已 .gitignore，历史中有 |
| `app/node_modules/` | 大量第三方依赖，未在 `.gitignore` |
| `release/*.exe` `release/*.zip` | 已 .gitignore |
| `web/default/dist/` | 已 .gitignore（当前目录有构建产物） |

### 3.6 功能边缘情况

| 场景 | 是否存在 | 说明 |
|:-----|:-------:|:------|
| 并发充值 + 同时消费 → 余额负 | 🟡 | idempotency 已加但无测试验证 |
| 渠道 Key 轮换期间运行请求 | 🟡 | 无 graceful drain |
| Token 过期后宽限期 | ❌ | 过期立即失败 |
| 多实例数据库连接池争用 | ❌ | 无 connection pool tuning |
| 消费竞态（双花） | 🟡 | idempotency 新加 |
| 日结/月度结算精度 | 🟡 | AutoSettleMonthlyFees 无测试 |

---

## 四、风险矩阵

| 风险 | 概率 | 影响 | 紧急度 | 说明 |
|:-----|:----:|:----:|:------:|:-----|
| AI 适配器上游 API 变更 | 高 | 中 | P1 | 无测试捕获，需要手动发现 |
| 新模型配置错误 | 中 | 高 | P1 | 52 个适配器配置全部无验证 |
| 多实例消费竞态 | 中 | 高 | P1 | idempotency 新加无测试 |
| 渠道 Key 加密被破坏 | 低 | 高 | P2 | AES-GCM+Qpw 完整 |
| Session 过期随机踢 | 中 | 中 | P2 | 多实例无 Redis Session |
| .env 历史泄露 | 低 | 高 | P2 | Git history 含敏感数据 |
| API 参数不当导致 AI 账单异常 | 中 | 中 | P2 | ParamValidator 无测试 |

---

## 五、总结

**消费链 ✅ 完整**：15 个阶段全部实现，从 TokenAuth 到 PostConsumeBilling 到 Commission 到 Monitor。

**资源链 ✅ 完整**：Channel 全生命周期 CRUD + 加密 + 测试 + 健康监控 + 自动故障转移。

**测试覆盖 🔴 严重不足**：核心问题是 24 个中间件 + 52 个适配器几乎无测试覆盖。这不是"功能缺失"而是"质量风险"——所有逻辑都写了，但一旦修改，无法通过 CI 发现回归。

**建议后续优先级**：
1. **P1**: 为 52 个适配器添加 smoke test（验证请求构建 + 响应解析 + 错误映射）
2. **P1**: 为 24 个中间件添加单元测试
3. **P2**: 覆盖新建模型 (failure_tracker/idempotency/fee_config) 的测试
4. **P2**: BFG Git history cleanup
