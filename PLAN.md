# QuantumClaw 全面完善任务计划

**制定**: 2026-05-24 21:23 | **版本**: v1.0 | **状态**: 📋 待执行

---

## 阶段一：修复断裂（P0 — 预计 30 分钟）

| # | 任务 | 问题 | 方案 | 涉及文件 |
|---|------|------|------|---------|
| 1 | **billing 页面 404** | 调用 `/api/billing/stats`、`/api/billing/records`，后端无此路由 | 创建 billing controller + 注册路由；或前端改调已有的 `transaction_logs` / `log/self` | controller/billing.go(新建), router/api.go |
| 2 | **subscription 路径错位** | 前端调 `/api/subscription`，后端用 `/api/subscription/plans` 和 `/api/subscription/self` | 前端改 URL 路径 | web/.../subscription.tsx |
| 3 | **checkin 路径错位** | 前端调 `/api/checkin`，后端路由是 `selfRoute.POST("/checkin")` | 前端加 `/api/user/self/` 前缀 | web/.../checkin.tsx |
| 4 | **platform-settings 路径错位** | 调用 `/api/platform/config`，后端在 `optionRoute` | 前端改 URL，或后端加平台配置独立路由 | web/.../platform-settings.tsx, controller/platform_config.go |
| 5 | **profit 页面 404** | 调用 `/api/channel/profit`，后端无对应 handler | 创建 profit handler 或前端移除错误调用 | controller/channel-billing.go, web/.../profit.tsx |
| 6 | **wallet 无后端连接** | 页面无 API 调用 | 接入 `transaction_logs` / `balance` / `withdraw/list` 等已有 API | web/.../wallet.tsx |

## 阶段二：核心功能补全（P1 — 预计 2 小时）

| # | 任务 | 说明 | 子步骤 |
|---|------|------|--------|
| 7 | **邀请码注册修复** | 注册时填 `aff_code` 但 `inviter_id` 为 0 | 排查 `LoginRateLimit` 中间件 body 缓存问题，确保注册 handler 能读到 `aff_code` |
| 8 | **渠道商升级后菜单自动显示** | `sidebar-distributors` 当前 roles `[10,100]`，但升级后 user_type=provider 的用户也应看到 | 修改菜单 API 过滤逻辑：检查 user_type |
| 9 | **团队管理完善** | 已有页面 + API，需补充：团队统计 API、注册自动绑定测试 | 后端：团队日活/周统计；前端：测试全流程 |
| 10 | **RSS 新闻页面接入** | DB 有 92 篇文章，news.tsx 页面无 API 调用 | 前端：`GET /api/rss/articles` 已有接口；后端已验证 |
| 11 | **种子测试数据** | 18 张空表需填充基础数据 | 日志、交易、签到记录、佣金记录等 |

## 阶段三：管理页面完善（P1.5 — 预计 3 小时）

| # | 任务 | 说明 | 涉及页面 |
|---|------|------|---------|
| 12 | **用户管理** | users.tsx 已有 CRUD query fn，需验证 API 连接、补充批量操作 UI | users.tsx |
| 13 | **渠道/Key 管理** | channels.tsx、keys.tsx 已有 API wrapper，需验证全流程 CRUD | channels.tsx, keys.tsx |
| 14 | **兑换码管理** | redemption.tsx 已有 query fn，验证 CRUD | redemption.tsx |
| 15 | **结算管理** | settlement.tsx 已有 query fn、mutation，验证全流程 | settlement.tsx |
| 16 | **交易记录** | transactions.tsx 接入已完成的后端 `/api/transactions` | transactions.tsx |
| 17 | **分销商管理** | reseller.tsx、reseller-keys.tsx、reseller-admin.tsx 链路验证 | 3 个 reseller 页面 |
| 18 | **分销 Key 管理** | reseller-keys.tsx 需接入 token API 或渠道 API | reseller-keys.tsx |

## 阶段四：认证体系完善（P2 — 预计 4 小时）

| # | 任务 | 说明 |
|---|------|------|
| 19 | **OAuth 登录 UI** | GitHub/WeChat/Discord/LinuxDO/Telegram 绑定页面 |
| 20 | **WebAuthn 生物认证** | 指纹/人脸注册与验证 UI |
| 21 | **双因素认证 (2FA)** | TOTP 设置与验证 UI |
| 22 | **紧急密码重置 UI** | 忘记密码流程，调用 `/api/password/emergency-reset` |
| 23 | **通知系统** | 通知列表、已读/未读状态、通知设置 UI |

## 阶段五：视觉优化（P2 — 预计 6 小时）

| # | 任务 | 方法 |
|---|------|------|
| 24 | **31 个认证页面视觉 QC** | 逐页截图 → 分析 → 修复（毛玻璃、琥珀色渐变、cubic-bezier 动画） |
| 25 | **NavBar API 驱动渲染** | 当前 useNavMenus 有 fallback，需验证是否走 API |
| 26 | **Sidebar API 驱动渲染** | 当前 useSidebarMenus 有 fallback，需验证分组逻辑 |
| 27 | **模型目录英文版** | model_metadata 只有中文数据，需补英文翻译 |
| 28 | **响应式极限测试** | 320px / 7680px 断点验证 |

---

## 执行策略

```
阶段一 → 阶段二 → 阶段三 + 四（并行）→ 阶段五
  (30min)   (2h)       (7h)            (6h)

总预计: ~15.5 小时
```

**节奏**：
- 每个任务完成后快速验证（`node -e` 测试 API + 前端构建检查）
- 每天批量 commit 一次到 Git
- 阶段五的视觉 QC 按「截图 → 分析 → 修复」三步法逐页做

---

**是否按此计划执行？先从阶段一 P0 修复开始？**
