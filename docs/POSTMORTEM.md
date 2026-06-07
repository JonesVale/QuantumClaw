# QuantumClaw v2.2.0 全面分析报告

> 生成: 2026-06-07 | 基线: develop@050dc60

---

## 已完成的整合工作（前序）

| Phase | 内容 | 状态 |
|:------|:------|:----:|
| Phase 1 | 构建体系归一化（web/build/ 删除、Dockerfile 简化） | ✅ |
| Phase 2 | app/ 移动端 Makefile 目标 | 🟡 待补页面 |
| Phase 3 | electron/ 桌面端 Makefile 目标 | 🟡 |
| Phase 4 | SDK 版本对齐 v2.2.0 + Makefile 发布目标 | ✅ |
| Phase 5 | 历史清理（44 个临时脚本归档、gitignore 全面化） | ✅ |
| Phase 6 | i18n 后端 7→14 语言扩容 | ✅ |
| Phase 7 | 级联架构 cmd_sync 编译集成 | ✅ |
| 加密增强 | Option 表 20+ 支付/OAuth 密钥 AES-GCM 加密 | ✅ |

---

## 本期发现的问题（按优先级）

### 🔴 P0 — 立即修复

#### P0-1: 提现身份验证死锁（致命）

**位置**: `controller/withdrawal.go` + `controller/reseller_withdraw.go`
**根因**: 
- `User.IdentityVerified` 在提现入口被检查，要求先实名认证
- **没有 API** 可以设置 `IdentityVerified = true`
- **没有 API** 可以提交 `IdentityName` / `IdentityNumber`
- 前端提示"在「个人资料」中提交身份信息"—该功能不存在

**影响**: 所有渠道商和供应商永远无法提现。能充钱能消费能赚佣金，但一分钱提不出来。

#### P0-2: OAuth CSRF state 多标签竞争

**位置**: `controller/auth/github.go` (所有 OAuth provider 共享 `GenerateOAuthCode`)
**根因**: 
```go
session.Set("oauth_state", state)  // 覆盖上次 state
```
用户同时开两个 OAuth 标签页，先完成的回调拿到的是后覆盖的 state → 验证失败 403

**影响**: 多标签用户 OAuth 登录随机失败，体验极差

### 🟡 P1 — 近期修复

#### P1-1: Session 基于 Cookie、无法多机

**位置**: `middleware/auth.go`
**根因**: Session 默认存储在单机内存中。多实例部署时，用户被轮询到不同节点 → 被强制登出。
**已有基础**: `REDIS_HOST` 已在 `.env` 中配置，但未启用 Redis Session store。

#### P1-2: 禁用用户仍可登录

**位置**: `middleware/auth.go:authHelper()`
```go
if statusInt == model.UserStatusDisabled {
    logger.SysWarnf("disabled user %d accessed route: %s", idInt, c.Request.URL.Path)
    // 没有 return，继续处理！
}
```
注释说"禁用用户仍可登录查看、充值、联系客服"，但这是安全风险。

#### P1-3: 未完成的身份验证功能

**位置**: `model/user.go`（字段存在），`controller/`（API 不存在）
**缺**: 
- `POST /api/user/identity` — 用户提交身份信息
- `POST /api/admin/user/verify` — 管理员审核并设置 IdentityVerified
- 前端「个人资料」页面身份信息填写区

### 🟢 P2 — 后续优化

| ID | 问题 | 说明 |
|:---|:------|:------|
| P2-1 | JWT 鉴权缺失 | 多机部署必须，已有方案但未实现 |
| P2-2 | app provider/enterprise 缺页 | 8 个 Placeholder 页面待实现 |
| P2-3 | app 无 i18n | 硬编码中文 |
| P2-4 | SDK 无自动化发布 | 需 CI/CD 集成 |
| P2-5 | .env/.exe Git history 含敏感数据 | 需 BFG clean |
| P2-6 | 未跟踪代码（store/market/failure-tracker） | 已提交但未测试 |

---

## 执行计划

### Sprint 1：修复 P0（已完成）

| 任务 | 文件 | Commit |
|:-----|:-----|:------:|
| ✅ 用户提交身份信息 API | `controller/identity.go` + `router/api.go` | `a62b277` |
| ✅ 管理员审核身份 API | `controller/identity.go` + `router/api.go` | `a62b277` |
| ✅ OAuth CSRF state 多标签 | `controller/auth/oauth_state.go` + 5 个 OAuth 文件 | `c232a1d` |
| ✅ 禁用用户登录拦截 | `middleware/auth.go` | `c232a1d` |

### Sprint 2：修复 P1（下个周期）

| 任务 | 工作量 |
|:-----|:------:|
| 1. Redis Session 集成 | 2h |
| 2. 前端身份信息提交页面 | 3h |
| 3. 测试未跟踪的 store/market 代码 | 3h |

### Sprint 3：修复 P2（后续）

| 任务 | 工作量 |
|:-----|:------:|
| JWT + BFG + app 页面 + i18n | 视具体情况 |
