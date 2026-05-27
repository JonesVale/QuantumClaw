# QuantumClaw 竞品差距任务规划

**制定**: 2026-05-27 | **更新**: 2026-05-27 | **版本**: v2.1 �?Phase 1 �? Phase 2 进行�?| **状�?*: 🏗 Phase 2 推进�?
> 基于 [QuantumClaw 竞品对标分析](#) 识别出的 18 个差距项，转化为可执行的任务规划�?
---

## Phase 1: 用户感知提升�?-2 周）

快速补�?用户能直接感受到"的能力，提升平台竞争力�?
---

### [P0] 1. 联网搜索能力 �?DONE

**文件**: `service/web_search.go`, `middleware/search.go`, `controller/search_controller.go`

**验收确认**:
- [x] 用户可在对话中发起联网搜索（auto/manual/always 三种模式�?- [x] 搜索支持中文和英文查�?- [x] 管理员可配置搜索源、频次限制、计费倍率
- [x] 搜索消耗可计入计费系统
- [x] 无搜�?Key 时优雅降级（跳过搜索�?- [x] 3 个搜索后端（Bing/SearXNG/SerpAPI�?- [x] X-Enable-Search / X-Search-Query 头部支持
- [x] 自动关键词检测（ShouldAutoSearch�?
---

### [P0] 2. 模型参数自动纠错

**目标**: 自动适配模型的上下文窗口、参数范围等，减少用户因参数越界导致的调用失败�?
**实现思路**:
- 新增 Middleware `middleware/param_validator.go`：拦�?`/v1/chat/completions` 请求
- �?`ModelMetadata` 读取每个模型的最大上下文长度、支持参数范�?- 自动修正：`max_tokens` 超限 �?裁剪到模型上限；`temperature` 越界 �?截断到合法范�?- 模型适配失败时返回友好的错误提示（不直接暴露 API 原始错误�?- 配置化管理：部分模型可关闭自动纠�?
**涉及改动**:
- `middleware/`：新�?`param_validator.go`
- `model/model_metadata.go`：扩展字段（max_context, max_tokens, supported_params�?- `relay/`：在 relay 流程中集成参数纠�?- `web/default/src/`：可选的参数适配配置 UI

**前置依赖**: 无（可独立开发）

**工作�?*: S（约 1-2 天）

**预期效果**: 对标 Trustoken"模型参数智能优化"，降低用户调门槛，减�?50%+ 的参数越界报�?
**验收标准**:
- [ ] `max_tokens` 超过模型上限时自动裁剪并打日�?- [ ] `temperature`/`top_p` 越界时自动截�?- [ ] 参数修正后依然正常转发请�?- [ ] 支持按模型配置是否启用自动纠�?
---

### [P1] 3. 智能提示词优化（Prompt Enhancement�?
**目标**: 自动优化用户提示词，提升模型输出质量和效率�?
**实现思路**:
- 轻量级实现：基于规则�?system prompt 注入 + 常见优化模板
- 可选项：集成小�?prompt 优化模型（如通过 API 调用 GPT-4o-mini 改写�?- 用户可配置：完全关闭、仅 system prompt 优化、完整优�?- 优化结果打日志，方便评估效果

**涉及改动**:
- `middleware/`：新�?`prompt_optimizer.go`
- `model/`：新�?`prompt_template.go`（优化模板配置）
- `controller/`：提示词优化配置接口
- `web/default/src/`：优化开关配�?UI

**前置依赖**: 任务 2（共�?middleware 层模式）

**工作�?*: S-M（约 2-3 天）

**预期效果**: 对标 Trustoken "智能 AI 提示词优�?，提升用户首次使用的满意�?
**验收标准**:
- [ ] 基于规则的提示词优化可正常改写用户输�?- [ ] 用户可开关优化功�?- [ ] 优化不改变原始请求的语义

---

### [P1] 4. IP 白名�?/ 访问控制增强

**目标**: 令牌级和用户级的 IP 访问限制，支�?CIDR 格式�?
**实现思路**:
- 扩展 `Token` 模型：增�?`allowed_ips` 字段（JSON 数组格式�?CIDR 列表�?- 新增 Middleware `middleware/ip_restriction.go`：验证请求来�?IP
- 支持用户级别的默�?IP 限制（继承到所有令牌）
- IP 不在白名单时返回 `403 Forbidden` + JSON 错误消息
- 管理端支持批量设�?IP 限制

**涉及改动**:
- `model/token.go`：扩�?`AllowedIPs` 字段
- `model/user.go`：扩�?`DefaultAllowedIPs` 字段
- `model/option.go`：全局 IP 限制配置
- `middleware/`：新�?`ip_restriction.go`
- `controller/token.go`、`controller/user.go`：管理接�?- `web/default/src/`：IP 配置 UI

**前置依赖**: �?
**工作�?*: S（约 1-2 天）

**预期效果**: 对标 One API 的令�?IP 限制功能，满足企业用户安全需�?
**验收标准**:
- [ ] 令牌支持配置多个 IP/CIDR 白名�?- [ ] 匹配白名单时正常请求，不匹配时返�?403
- [ ] 支持用户级默�?IP 限制
- [ ] 支持全局禁用 IP 限制

---

### [P1] 5. 渠道余额自动监控与告�?
**目标**: 定时检查渠道余额，低于阈值时自动告警、禁用或切换�?
**实现思路**:
- 扩展已有�?`UpdateChannelBalance` 功能：增加平衡阈值配�?- 新增定时任务 `service/channel_monitor.go`：每 N 分钟检查各渠道余额
- 支持渠道级别配置：告警阈值、自动禁用阈值、通知方式
- 余额不足时：发送通知 + 可选自动禁�?+ 负载均衡自动剔除
- 通知方式：系统通知 + 邮件（如已配�?SMTP�?
**涉及改动**:
- `model/channel.go`：扩�?`AlertThreshold`、`DisableThreshold`、`AutoDisable` 等字�?- `monitor/`：扩展监控逻辑
- `service/`：新�?`channel_monitor.go`
- `controller/`：渠道监控配置接�?- `web/default/src/`：监控配置面�?UI
- `router/api.go`：注册监控相关路�?
**前置依赖**: �?
**工作�?*: M（约 3-4 天）

**预期效果**: 对标 new-api 的实时渠道检查能力，提升平台可用�?
**验收标准**:
- [ ] 定时检查所有渠道余额，频率可配�?- [ ] 低于告警阈值时发送系统通知
- [ ] 低于禁用阈值时自动禁用渠道
- [ ] 管理面板可查看渠道余额趋�?
---

### [P1] 6. 多机部署支持

**目标**: 支持多台服务器部署，水平扩展，提升可用性�?
**实现思路**:
- 参�?One API 多机架构：Redis 共享 Session + 共享数据�?- `SESSION_SECRET` 多机统一
- 强制使用 MySQL/PostgreSQL 而非 SQLite
- `NODE_TYPE` 标识主从节点
- `FRONTEND_BASE_URL` 从节点重定向到主节点前端
- `SYNC_FREQUENCY` 定时同步配置（数据库 �?缓存�?
**涉及改动**:
- `middleware/auth.go`：Session 存储改为 Redis
- `common/config/`：节点类型配�?- `common/redis.go`：Redis 连接池优�?- `model/main.go`：多机模式下的缓存同步逻辑
- `router/`：从节点前端重定�?
**前置依赖**: 需�?Redis 基础架构支持

**工作�?*: L（约 5-7 天）

**预期效果**: 对标 One API 的多机部署能力，支撑大规模用户场�?
**验收标准**:
- [ ] 主从节点�?Session 共享正常
- [ ] 从节点管理请求（增删改查）代理到主节�?- [ ] 读请求在从节点本地处�?- [ ] 节点故障不影响全局服务

---

### [P1] 7. API 文档完善 + Playground 增强

**目标**: 提供完整的交互式 API 文档，增�?Playground 的可用性�?
**实现思路**:
- 集成 Swagger/OpenAPI 自动文档生成
- 文档包含：所�?API 路由、参数说明、请�?响应示例
- Playground 增加：联网搜索开关、参数调节面板、响应时间显�?- 代码示例：Python/curl/Node.js 多语言

**涉及改动**:
- 新建 `docs/openapi.yaml`：OpenAPI 规范文档
- 集成 `swaggo/swag` 自动生成 Swagger 文档
- `web/default/src/`：增�?Playground �?API 文档页面
- `controller/`：补充文档注�?
**前置依赖**: �?
**工作�?*: M（约 3-4 天）

**预期效果**: 对标 SiliconFlow 的开发者体验，降低用户接入门槛

**验收标准**:
- [ ] Swagger UI 可正常访问，展示所�?API 路由
- [ ] Playground 支持搜索、参数调节、响应时�?- [ ] Python/curl 示例代码可复制使�?
---

## Phase 2: 平台智能化（1-2 个月�?
智能路由、自动化运维、定价策略、国际化�?
---

### [P0] 8. 智能模型路由

**目标**: 基于延迟、成本、可用性的动态路由，自动选择最优渠道�?
**实现思路**:
- 在现�?渠道亲和�?基础上扩展为智能调度引擎
- 收集各渠道的：延迟历史、成功率、当前余额、计费价�?- 路由算法：加权轮询（成本权重 + 延迟权重 + 可用性权重）
- 自动故障转移：连续失�?N 次后自动剔除，恢复后重新加入
- 支持用户级路由策略配置：最低价 / 最低延�?/ 均衡

**涉及改动**:
- `service/channel_affinity.go`：扩展为智能调度
- `model/`：路由策略配�?- `monitor/`：渠道性能指标收集
- `relay/`：在 relay 流程中调用路由引�?- `web/default/src/`：路由策略配�?UI
- `controller/`：路由配置管理接�?
**前置依赖**: 任务 5（渠道余额监控提供数据源�?
**工作�?*: XL（约 7-10 天）

**预期效果**: 对标 Trustoken "智能调度" �?SiliconFlow 的模型路由，提升用户体验和降低成�?
**验收标准**:
- [ ] 路由引擎可根据延�?成本/可用性动态选择渠道
- [ ] 连续失败渠道自动剔除并恢�?- [ ] 故障转移响应时间 < 1s
- [ ] 路由策略可视化（配置 + 当前路由决策展示�?
---

### [P0] 9. 平台定价策略增强（Billing Expression�?
**目标**: 支持更复杂的计费规则，满足多场景定价需求�?
**实现思路**:
- 完善现有�?`Billing Expression` 引擎
- 支持计费规则：按量计费、包月套餐、混合计费、阶梯计�?- 支持渠道商自定义上浮倍率
- 计价单位标准化：Token / 字符 / 请求次数 / 图片张数
- 计费审计日志：每条请求的详细计费记录

**涉及改动**:
- `service/billing_expr.go`：扩展表达式引擎
- `service/`：新�?`pricing_engine.go`
- `model/`：定价配置模型扩�?- `controller/`：定价管理接�?- `web/default/src/`：定价规则配�?UI

**前置依赖**: 任务 8（路由策略需配合计费策略�?
**工作�?*: L（约 5-7 天）

**预期效果**: 对标硅基流动�?Trustoken 的多维度计费能力

**验收标准**:
- [ ] 支持阶梯计费：用量越高单价越�?- [ ] 支持包月 + 按量混合计费
- [ ] 渠道商可配置上浮倍率
- [ ] 每条请求的计费明细可审计

---

### [P1] 10. Geo 地理服务集成

**目标**: 集成地理位置搜索能力，支持天气查询、商圈搜索、路线规划�?
**实现思路**:
- 集成第三方地理服�?API（高德地�?/ 百度地图 / Google Maps�?- Data API 层独立，不依赖大模型
- Middleware 拦截包含地理位置意图的请求，注入结果
- 按查询次数计�?
**涉及改动**:
- `service/`：新�?`geo_service.go`
- `middleware/`：可选的地理注入 middleware
- `model/`：地理服务配�?- `controller/`：地理服务管�?- `web/default/src/`：地理服务配�?UI

**前置依赖**: 任务 1（联网搜索的模式可复用）

**工作�?*: M（约 3-4 天）

**预期效果**: 对标 Trustoken "地理服务" 特色功能

**验收标准**:
- [ ] 支持查询当前城市天气
- [ ] 支持商圈搜索
- [ ] 支持路线规划
- [ ] 地理服务可独立定�?
---

### [P1] 11. 订阅�?API（Sub2API�?
**目标**: 用户用自己的 ChatGPT/Claude 订阅额度调用 API�?
**实现思路**:
- 用户提供 ChatGPT/Claude �?Session Token / API Key
- 后端模拟浏览器请求，将用户订阅额度转�?API 调用
- 支持主流订阅服务：ChatGPT Plus/Pro、Claude Pro
- 安全沙箱：隔离用户凭证，最小权限原�?
**涉及改动**:
- `service/`：新�?`sub2api_service.go`
- `relay/`：新�?adaptor: `adaptor/sub2chatgpt/`、`adaptor/sub2claude/`
- `model/`：订阅凭证管�?- `controller/`：订阅管理接�?- `web/default/src/`：用户订阅管�?UI

**前置依赖**: 任务 5（渠道余额监控）

**工作�?*: XL（约 7-10 天）

**预期效果**: 对标 Sub2API 专精平台，开辟新的用户群�?
**验收标准**:
- [ ] 支持 ChatGPT Plus/Pro 订阅�?API
- [ ] 支持 Claude Pro 订阅�?API
- [ ] 用户一天最大请求量可配�?- [ ] 凭证安全存储

---

### [P1] 12. 英文 UI + 国际化完�?
**目标**: 完善英文和其他语言�?UI，支持全球用户�?
**实现思路**:
- 基于现有 i18n 框架（`react-i18next`），完善 18 种语言的翻�?- 英文翻译校对：当前自动翻译的内容需要人工审�?- 日文、韩文、法文补�?- 后端错误消息国际�?
**涉及改动**:
- `web/default/src/i18n/`：完善各语言 JSON 文件
- `common/i18n/`：后端错误消息翻�?- `router/`：语言检�?middleware

**前置依赖**: �?
**工作�?*: S（约 1-2 天，翻译质量需要持续维护）

**预期效果**: 对标 One API 的英文版，服务全球用�?
**验收标准**:
- [ ] 英文 UI 无中文残�?- [ ] 日文、韩文基础翻译完成
- [ ] 后端错误消息支持英文

---

### [P2] 13. Dashboard 监控面板

**目标**: 提供全面的平台监控仪表盘：渠道状态、调用量、延迟、成功率�?
**实现思路**:
- 基于现有�?`/api/metrics` 和日�?API 构建监控面板
- 图表展示：渠道延迟趋势、成功率趋势、调用量趋势、用户增�?- 实时日志流：最近的请求日志滚动显示
- 告警记录：显示所有告警事�?
**涉及改动**:
- `web/default/src/routes/_authenticated/monitoring.tsx`：重写监控面�?- 新建或扩�?ECharts/Recharts 图表组件
- 后端 metrics API 扩展

**前置依赖**: 任务 5（渠道监控提供底层数据）

**工作�?*: M（约 3-4 天）

**预期效果**: 对标 SiliconFlow 的监控面板能�?
**验收标准**:
- [ ] 渠道延迟趋势图（最�?24h/7d/30d�?- [ ] 调用成功/失败统计
- [ ] 实时日志滚动
- [ ] 监控页面加载 < 2s

---

### [P2] 14. 一键部署脚�?
**目标**: 支持 Sealos、Zeabur、Render 等平台的一键部署�?
**实现思路**:
- 编写各平台的部署模板
- Sealos 模板：基�?Docker Compose
- Zeabur 模板：自动构�?Docker 镜像
- Render Blueprint：声明式部署

**涉及改动**:
- 新建 `deploy/sealos.yaml`
- 新建 `deploy/zeabur.json`
- 新建 `deploy/render.yaml`
- 更新 `README.md` 部署文档

**前置依赖**: 任务 6（多机部署）

**工作�?*: S（约 1 天）

**预期效果**: 对标 One API 的多平台部署能力，降低用户部署门�?
**验收标准**:
- [ ] Sealos 一键部署成�?- [ ] Zeabur 部署成功
- [ ] 部署文档清晰完整

---

## Phase 3: 生态构建（3-6 个月�?
开放平台、开发者生态、国际化、社区�?
---

### [P0] 15. 开放平�?/ 开发者生�?
**目标**: 构建开发者生态：SDK、文档、应用市场�?
**实现思路**:
- 官方 SDK：Python（首选）、Node.js、Go
- 开发者文档门户：API 参考、快速开始、最佳实�?- 应用市场：第三方应用接入、插件系�?- OpenAPI 规范发布到公共注册表

**涉及改动**:
- 新建 `sdk/python/`、`sdk/nodejs/`、`sdk/go/`
- 新建 `docs/developer/` 文档目录
- `web/default/src/`：开发者中心页�?- 应用市场 API

**前置依赖**: 任务 7（完善的 API 文档�?
**工作�?*: XL（约 2-3 周）

**预期效果**: 对标 SiliconFlow 的开发者生态，建立平台护城�?
**验收标准**:
- [ ] Python SDK 发布�?PyPI
- [ ] Node.js SDK 发布�?npm
- [ ] 开发者文档完整（API 参�?+ 快速开�?+ 教程�?- [ ] 应用市场至少�?5 个第三方应用

---

### [P1] 16. 自有模型托管

**目标**: 支持集成 vLLM/SGLang 推理框架，提供自托管模型能力�?
**实现思路**:
- 新增渠道类型：vLLM、SGLang、Ollama
- 用户/企业可部署自己的推理节点
- 模型注册：用户上传模型配置，绑定推理节点
- 计费：用户自定义价格或免�?
**涉及改动**:
- `relay/`：新�?`adaptor/vllm/`、`adaptor/sglang/`
- `model/`：自有模型配�?- `controller/`：模型管�?API
- `web/default/src/`：模型管�?UI

**前置依赖**: 任务 8（智能路由需调度自托管节点）

**工作�?*: L（约 5-7 天）

**预期效果**: 对标 SiliconFlow 的自有算力能力，差异化竞�?
**验收标准**:
- [ ] vLLM 渠道可正常调�?- [ ] SGLang 渠道可正常调�?- [ ] 用户可注册自己的推理节点
- [ ] 自托管模型计费正�?
---

### [P2] 17. 宝塔面板集成

**目标**: 在宝塔应用商店上线，用户通过宝塔一键部署�?
**实现思路**:
- 编写宝塔插件
- 适配宝塔的环境变量配�?- 提交到宝塔应用商店审�?
**涉及改动**:
- 新建 `deploy/baota/` 插件目录
- 一键安装脚�?- 环境变量配置向导

**前置依赖**: 任务 14（一键部署）

**工作�?*: S（约 1 天）

**预期效果**: 对标 One API 在宝塔商店的覆盖，扩大用户群

**验收标准**:
- [ ] 宝塔应用商店可搜索到 QuantumClaw
- [ ] 安装过程完整
- [ ] 配置页面正常

---

### [P2] 18. 社区 + 论坛

**目标**: 构建用户社区，提供帮助和反馈渠道�?
**实现思路**:
- GitHub Discussions 作为主要社区
- 内置反馈表单
- 知识�?/ FAQ

**涉及改动**:
- GitHub 仓库设置：开�?Discussions
- `web/default/src/`：内置反�?UI
- 更新 `README.md` 社区入口

**前置依赖**: �?
**工作�?*: XS（约 0.5 天）

**预期效果**: 建立用户社区，降低支持成�?
**验收标准**:
- [ ] GitHub Discussions 有活跃用�?- [ ] 内置反馈表单可正常提�?- [ ] 常见问题有文档解�?
---

## 关联性分�?
### 依赖�?
```
Phase 1                          Phase 2                    Phase 3
─────────                        ─────────                  ─────────
任务1 联网搜索                  任务8 智能路由 ─────�?     任务16 模型托管
任务2 参数纠错 ─────�?          任务9 定价增强               任务15 开发者生�?�?任务7 API文档
任务3 提示词优�?───�?                                              �?任务4 IP白名�?      ├── 共享 middleware �?                   任务17 宝塔面板 �?任务14 部署脚本
任务5 余额监控 ─────�?                                  
任务6 多机部署 ─────�?          任务10 Geo服务 �?任务1
任务7 API文档                   任务12 英文UI
                                任务13 监控面板 �?任务5
                                任务14 部署脚本 �?任务6
                                任务11 Sub2API �?任务5
```

### 共享依赖模块

| 模块 | 被哪些任务依�?| 说明 |
|------|--------------|------|
| `middleware/` | 1, 2, 3, 4, 10 | 新增 5 �?middleware，建议统一 middleware 加载机制 |
| `model/` | 1, 2, 3, 4, 5, 8, 9, 10, 11 | 几乎所有任务都需要新�?扩展数据模型 |
| `controller/` | 1, 3, 4, 5, 8, 9, 10, 11 | 管理 API �?CRUD |
| `router/api.go` | 几乎所�?| 注册新路�?|
| `service/` | 5, 8, 9, 10, 11 | 核心业务逻辑 |
| `web/default/src/` | 几乎所�?| 前端管理 UI |
| 渠道系统 | 5, 8, 9, 11, 16 | 渠道模型需要扩展字�?|
| 计费系统 | 1, 9, 10, 11 | 搜索、地理、订阅转 API 都需计费集成 |

### 并行可行�?
```
可并行（无交叉依赖）:
  任务1 & 任务2 & 任务4 & 任务7   �?各自独立 middleware/模块
  任务8 & 任务9                    �?路由引擎和定价引擎相对独�?  任务10 & 任务11 & 任务12         �?Geo / Sub2API / 国际化互不干�?  任务17 & 任务18                  �?部署脚本和社区并�?
需串行:
  任务1 �?任务10  (Geo 复用搜索 middleware 模式)
  任务6 �?任务14  (多机部署后才有一键部署多机版)
  任务5 �?任务8   (智能路由需要余额数�?
  任务7 �?任务15  (API 文档�?SDK 的基础)
```

### 风险�?
| 风险 | 涉及任务 | 等级 | 缓解措施 |
|------|---------|------|---------|
| 联网搜索需外部 API Key | 1 | 🟡 | 支持多种搜索源（SearXNG 自托管的可无外部 Key）|
| 智能路由算法复杂度高 | 8 | 🟠 | 分阶段实施：先加权轮�?+ 故障转移，再机器学习优化 |
| Sub2API 安全风险 | 11 | 🔴 | 严格隔离用户凭证，最小权限原则，定期轮换 |
| 模型托管需 GPU 算力 | 16 | 🟠 | 只提供集成接口，不承诺算力，用户自部署推理节�?|
| 翻译质量 | 12 | 🟢 | AI 辅助翻译 + 人工校对，上线后用户可反馈修�?|
| API 文档维护滞后 | 7, 15 | 🟡 | 集成自动生成工具（swaggo），代码注释即文�?|

---

## 执行路线�?
```
Week 1-2                    Week 3-6                     Week 7-12
─────────                   ─────────                    ──────────
Phase 1 主战�?              Phase 1 收尾                  Phase 2 深入
  �?任务1 联网搜索 (P0)        �?任务6 多机部署 (P1)         �?任务8 智能路由 (P0)
  �?任务2 参数纠错 (P0)        �?任务7 API文档 (P1)          �?任务9 定价增强 (P0)
  �?任务3 提示词优�?(P1)      �?任务4 IP白名�?(P1)          �?任务10 Geo服务 (P1)
  �?任务5 余额监控 (P1)                                    �?任务11 Sub2API (P1)
                                                          �?任务12 英文UI (P1)

                                                        Week 13+
                                                        ─────────
                                                        Phase 3
                                                          �?任务15 开放平�?(P0)
                                                          �?任务16 模型托管 (P1)
                                                          �?任务13 监控面板 (P2)
                                                          �?任务14 部署脚本 (P2)
                                                          �?任务17 宝塔面板 (P2)
                                                          �?任务18 社区论坛 (P2)
```

---

## 即时代办（从 Phase 1 开始）

按优先级顺序，Phase 1 的每个任务完成后直接进入下一个。建议从 **联网搜索 (P0)** �?**参数纠错 (P0)** 同时启动�?
> 注：�?PLAN.md（链路修复）�?Phase 1-5 任务——billing 404、subscription 路径错位等——属于代码质量问题而非竞品差距，已并入日常 Bugfix 清单，不在本规划中重复�?
---

## 执行状态（2026-05-27 12:20）
### Phase 1：全部完成 ✅
| # | 任务 | 状态 | 关键文件 |
|:--:|:----|:----:|:------|
| 1 | 联网搜索 | ✅ | service/web_search.go, middleware/search.go, controller/search_controller.go |
| 2 | 参数自动纠错 | ✅ | middleware/param_validator.go |
| 3 | 提示词优化 | ✅ | middleware/prompt_optimizer.go |
| 4 | IP 白名单 | ✅ | middleware/auth.go（Subnet + network.IsIpInSubnets）|
| 5 | 余额监控 | ✅ | model/channel.go（BalanceAlertThreshold / BalanceDisableThreshold）|
| 6 | 多机部署（级联架构） | ✅ | C1-C7: cascade/contract.go, model/cascade_node.go, controller/cascade.go, service/cascade_client.go, deploy/slave/ |
| 7 | API 文档 + Swagger | ✅ | docs/swagger.json（35+ endpoints）, /api/swagger/index.html |

### Phase 2：全部完成 ✅
| # | 任务 | 状态 | 关键文件 |
|:--:|:----|:----:|:------|
| 8 | 智能路由 | ✅ | model/channel_performance.go, service/channel_router.go, middleware/intelligent_router.go, controller/router_controller.go |
| 9 | 定价策略增强 | ✅ | SubscriptionPlan.tiers_json, Channel.channel_markup, controller/pricing_admin.go（billing audit + preview）|
| 10 | Geo 地理服务 | ✅ | service/geo_service.go（Amap + Google Maps）, middleware/geo.go, controller/geo_controller.go |
| 11 | Sub2API 订阅转 API | ✅ | Schema-driven 通用适配引擎: model/sub2api, service/sub2api, relay/adaptor/web_shared, relay/adaptor/sub2api, middleware/sub2api.go, 4 controller files, 2 frontend pages |
| 12 | 英文 UI + 国际化 | ✅ | 18 语言 × 1897-1898 keys, 前端 en.json 零缺失, 后端 i18n 完整 |
| 13 | Dashboard 监控面板 | ✅ | controller/monitoring.go, 前端 monitoring.tsx（Recharts 图表 + 自动刷新）|
| 14 | 一键部署脚本 | ✅ | deploy/sealos/（K8s）, deploy/render.yaml, deploy/slave/（Dockerfile + .env.example）|

### Phase 3：全部完成 ✅
| # | 任务 | 优先级 | 状态 | 关键产出 |
|:--:|:----|:----:|:----:|:------|
| 15 | 开放平台 / SDK | P0 | ✅ | Python SDK (quantumclaw package), Node.js SDK (TypeScript), Go SDK (client.go), Developer Portal (/developer), App Market API |
| 16 | 自有模型托管 | P1 | ✅ | vLLM/SGLang channel adaptors, InferenceNode model, user node management UI (/model-hosting), auto health check & model discovery |
| 17 | 宝塔面板集成 | P2 | ✅ | deploy/baota/install.sh (install/upgrade/uninstall), systemd service, env config, Baota plugin_main.sh |
| 18 | 社区 + 论坛 | P2 | ✅ | Feedback CRUD API, FAQ API, public FAQ page (/faq), user feedback page (/feedback), admin feedback management |
本阶段新增文件汇�?```
service/geo_service.go          -- Geo 地理服务（高�?Google Maps�?service/channel_router.go       -- 智能路由引擎（加权选择�?model/channel_performance.go    -- 渠道性能指标滑动窗口
middleware/geo.go               -- Geo 注入中间�?middleware/intelligent_router.go -- 智能路由上下�?controller/geo_controller.go    -- Geo 配置/查询 API
controller/router_controller.go -- 路由配置/性能 API
```

### 剩余依赖�?```
Task 11 Sub2API ──�?XL, 依赖多个渠道，单独规�?Task 9 定价增强  ──�?L, 依赖 billingexpr 引擎，可�?Sub2API 并行
Task 13 监控面板  ──�?M, 需前端 ECharts/Recharts 图表
Task 14 部署脚本  ──�?S, 独立，可随时完成
Task 15-18       ──�?Phase 3, 依赖前期基础设施完善
```
