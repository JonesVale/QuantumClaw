# controller/user.go 拆分方案

## 一、文件现状分析

**文件**: `controller/user.go` | **总行数**: 1299 行 | **package**: controller

### 1.1 函数/方法清单及行号范围

| # | 函数名 | 行号范围 | 行数 | 参数 | 功能域 |
|---|--------|----------|------|------|--------|
| 1 | `LoginRequest` (struct) | 26-29 | 4 | - | 认证登录 |
| 2 | `Login` | 31-70 | 40 | `*gin.Context` | 认证登录 |
| 3 | `SetupLogin` | 73-109 | 37 | `*model.User, *gin.Context` | 认证登录(共享) |
| 4 | `Logout` | 111-126 | 16 | `*gin.Context` | 认证登录 |
| 5 | `ValidatePasswordStrength` | 130-178 | 49 | `string → string` | 密码校验(共享) |
| 6 | `Register` | 180-380 | 201 | `*gin.Context` | 认证登录 |
| 7 | `getAdminOrgFilter` | 384-390 | 7 | `*gin.Context → int` | 管理员(共享) |
| 8 | `GetAllUsers` | 392-415 | 24 | `*gin.Context` | 管理员 |
| 9 | `SearchUsers` | 417-434 | 18 | `*gin.Context` | 管理员 |
| 10 | `GetUser` | 436-477 | 42 | `*gin.Context` | 管理员 |
| 11 | `DailyStat` (struct) | 480-485 | 6 | - | 统计面板 |
| 12 | `ModelStat` (struct) | 488-493 | 6 | - | 统计面板 |
| 13 | `ProviderStat` (struct) | 496-501 | 6 | - | 统计面板 |
| 14 | `GetUserDashboard` | 503-605 | 103 | `*gin.Context` | 统计面板 |
| 15 | `buildModelProviderMap` | 608-645 | 38 | `→ map[string]string` | 统计面板(共享) |
| 16 | `GenerateAccessToken` | 647-682 | 36 | `*gin.Context` | Token管理 |
| 17 | `GetAffCode` | 684-710 | 27 | `*gin.Context` | 用户自助 |
| 18 | `GetSelf` | 712-728 | 17 | `*gin.Context` | 用户自助 |
| 19 | `UpdateUser` | 730-793 | 64 | `*gin.Context` | 管理员 |
| 20 | `UpdateSelf` | 795-842 | 48 | `*gin.Context` | 用户自助 |
| 21 | `DeleteUser` | 844-887 | 44 | `*gin.Context` | 管理员 |
| 22 | `DeleteSelf` | 889-914 | 26 | `*gin.Context` | 用户自助 |
| 23 | `CreateUser` | 916-972 | 57 | `*gin.Context` | 管理员 |
| 24 | `ManageRequest` (struct) | 974-977 | 4 | - | 管理员 |
| 25 | `ManageUser` | 980-1113 | 134 | `*gin.Context` | 管理员 |
| 26 | `EmailBind` | 1115-1155 | 41 | `*gin.Context` | 用户资料 |
| 27 | `topUpRequest` (struct) | 1157-1159 | 3 | - | 计费充值 |
| 28 | `TopUp` | 1161-1187 | 27 | `*gin.Context` | 计费充值 |
| 29 | `adminTopUpRequest` (struct) | 1189-1193 | 5 | - | 计费充值 |
| 30 | `AdminTopUp` | 1195-1223 | 29 | `*gin.Context` | 计费充值 |
| 31 | `UpgradeToProvider` | 1228-1257 | 30 | `*gin.Context` | 用户资料 |
| 32 | `GetMyTeam` | 1260-1299 | 40 | `*gin.Context` | 用户资料 |

### 1.2 按功能域汇总行数

| 功能域 | 函数数量 | 总行数 |
|--------|----------|--------|
| 认证登录 (Login/Logout/Register/SetupLogin) | 4 + 1 struct | ~298 |
| 管理员用户管理 (CRUD + ManageUser) | 7 + 1 struct | ~393 |
| 用户自助操作 (GetSelf/UpdateSelf/DeleteSelf/GetAffCode) | 4 | ~118 |
| 统计面板 (Dashboard + stats structs) | 4 structs + 2 funcs | ~159 |
| Token管理 (GenerateAccessToken) | 1 | ~36 |
| 用户资料/邮箱绑定/升级/团队 | 3 | ~111 |
| 计费充值 (TopUp/AdminTopUp) | 2 + 2 structs | ~64 |
| 共享辅助函数 (getAdminOrgFilter, ValidatePasswordStrength, buildModelProviderMap) | 3 | ~94 |

---

## 二、拆分方案

### 2.1 新文件列表

| # | 文件名 | 职责 | 预估行数 |
|---|--------|------|----------|
| 1 | `user_auth.go` | 认证登录：密码登录、注册、登出 | ~300 |
| 2 | `user_admin.go` | 管理员用户管理：CRUD、启禁用、升降级 | ~400 |
| 3 | `user_self.go` | 用户自助操作：查看/更新/删除自己、推广码 | ~120 |
| 4 | `user_dashboard.go` | 统计面板：Dashboard 及统计类型定义 | ~160 |
| 5 | `user_token.go` | Token 管理：生成 Access Token | ~40 |
| 6 | `user_profile.go` | 用户资料：邮箱绑定、升级为渠道商、团队 | ~115 |
| 7 | `user_billing.go` | 计费充值：用户充值、管理员充值 | ~65 |
| 8 | `user_common.go` | 共享逻辑：辅助函数、共享类型 | ~100 |

**合计**: ~1300 行（与原文件一致，但每个文件均在 500 行以内）

### 2.2 每个文件的详细函数清单

#### `user_common.go` — 共享逻辑
```
- type LoginRequest struct                    (原 26-29)
- func SetupLogin(user *model.User, c *gin.Context)   (原 73-109)
- func ValidatePasswordStrength(password string) string (原 130-178)
- func getAdminOrgFilter(c *gin.Context) int           (原 384-390)
- func buildModelProviderMap() map[string]string        (原 608-645)
```
**说明**: 这些函数被多个子文件引用：
- `SetupLogin` → 被 `user_auth.go`(Register) + 外部 `controller/auth/*.go`(8 个 OAuth 文件) 引用
- `ValidatePasswordStrength` → 被 `user_auth.go`(Register) 引用
- `getAdminOrgFilter` → 被 `user_admin.go`(GetAllUsers/SearchUsers/GetUser/DeleteUser/CreateUser/ManageUser) 引用
- `buildModelProviderMap` → 被 `user_dashboard.go`(GetUserDashboard) 引用
- `LoginRequest` → 被 `user_auth.go`(Login) 引用

#### `user_auth.go` — 认证登录
```
- func Login(c *gin.Context)         (原 31-70)
- func Logout(c *gin.Context)        (原 111-126)
- func Register(c *gin.Context)      (原 180-380)
```

#### `user_admin.go` — 管理员用户管理
```
- func GetAllUsers(c *gin.Context)   (原 392-415)
- func SearchUsers(c *gin.Context)   (原 417-434)
- func GetUser(c *gin.Context)       (原 436-477)
- func UpdateUser(c *gin.Context)    (原 730-793)
- func DeleteUser(c *gin.Context)    (原 844-887)
- func CreateUser(c *gin.Context)    (原 916-972)
- type ManageRequest struct          (原 974-977)
- func ManageUser(c *gin.Context)    (原 980-1113)
```

#### `user_self.go` — 用户自助操作
```
- func GetSelf(c *gin.Context)       (原 712-728)
- func UpdateSelf(c *gin.Context)    (原 795-842)
- func DeleteSelf(c *gin.Context)    (原 889-914)
- func GetAffCode(c *gin.Context)    (原 684-710)
```

#### `user_dashboard.go` — 统计面板
```
- type DailyStat struct              (原 480-485)
- type ModelStat struct              (原 488-493)
- type ProviderStat struct           (原 496-501)
- func GetUserDashboard(c *gin.Context) (原 503-605)
```

#### `user_token.go` — Token 管理
```
- func GenerateAccessToken(c *gin.Context) (原 647-682)
```

#### `user_profile.go` — 用户资料/升级/团队
```
- func EmailBind(c *gin.Context)           (原 1115-1155)
- func UpgradeToProvider(c *gin.Context)   (原 1228-1257)
- func GetMyTeam(c *gin.Context)           (原 1260-1299)
```

#### `user_billing.go` — 计费充值
```
- type topUpRequest struct           (原 1157-1159)
- func TopUp(c *gin.Context)         (原 1161-1187)
- type adminTopUpRequest struct      (原 1189-1193)
- func AdminTopUp(c *gin.Context)    (原 1195-1223)
```

---

## 三、依赖关系分析

### 3.1 包内依赖（同 package，无需改 import）

所有拆分文件仍在 `package controller` 内，函数间直接引用，**无需修改任何 import 路径**。

### 3.2 跨包外部引用

| 被引用符号 | 引用方 | 影响评估 |
|-----------|--------|----------|
| `controller.SetupLogin` | `controller/auth/` 下 8 个文件 (alipay, discord, github, lark, linuxdo, oidc, telegram, wechat, webauthn) | **无需修改** — 拆分后 `SetupLogin` 仍在 `controller` 包中，`controller.SetupLogin` 调用不变 |
| `controller.Login` | `router/api.go` | **无需修改** — 包路径不变 |
| `controller.Register` | `router/api.go` | **无需修改** |
| 其他所有 handler | `router/api.go` | **无需修改** — 均通过 `controller.XXX` 引用，包名不变 |

**核心结论**: Go 同一 package 内的文件拆分对包外调用者完全透明。所有 `controller.XXX` 引用路径保持不变，**路由注册文件不需要任何调整**。

### 3.3 包内函数间调用关系图

```
user_auth.go:
  Login → SetupLogin (user_common.go)
  Login → LoginRequest (user_common.go)
  Register → ValidatePasswordStrength (user_common.go)
  Register → SetupLogin (user_common.go)

user_admin.go:
  GetAllUsers → getAdminOrgFilter (user_common.go)
  SearchUsers → getAdminOrgFilter (user_common.go)
  GetUser → getAdminOrgFilter (user_common.go)
  DeleteUser → getAdminOrgFilter (user_common.go)
  CreateUser → getAdminOrgFilter (user_common.go)
  ManageUser → getAdminOrgFilter (user_common.go)

user_dashboard.go:
  GetUserDashboard → buildModelProviderMap (user_common.go)

user_token.go: (无内部依赖)
user_self.go:  (无内部依赖)
user_profile.go: (无内部依赖)
user_billing.go: (无内部依赖)
```

### 3.4 对 `channelId2Models` 的依赖

`buildModelProviderMap`（原 user.go:627）引用了 `controller/channel.go` 中定义的 `channelId2Models` 变量。
拆分后该引用在 `user_common.go` 中，**与现状一致**，无需修改。

---

## 四、路由注册调整建议

**无需任何调整。** 原因：

1. 所有 handler 仍在 `package controller` 中
2. `router/api.go` 中使用 `controller.XXX` 引用，包路径 `github.com/quantumclaw/quantumclaw/controller` 不变
3. 拆分仅影响文件组织，不影响 Go 编译单元（同 package 多文件是 Go 标准做法）

路由对照表（确认无遗漏）：

| 路由 | Handler | 归属子文件 |
|------|---------|-----------|
| POST /api/user/register | Register | user_auth.go |
| POST /api/user/login | Login | user_auth.go |
| GET /api/user/logout | Logout | user_auth.go |
| GET /api/user/self/dashboard | GetUserDashboard | user_dashboard.go |
| GET /api/user/self/ | GetSelf | user_self.go |
| PUT /api/user/self/ | UpdateSelf | user_self.go |
| DELETE /api/user/self/ | DeleteSelf | user_self.go |
| GET /api/user/self/token | GenerateAccessToken | user_token.go |
| GET /api/user/self/aff | GetAffCode | user_self.go |
| POST /api/user/self/topup | TopUp | user_billing.go |
| GET /api/user/self/team | GetMyTeam | user_profile.go |
| POST /api/user/self/upgrade | UpgradeToProvider | user_profile.go |
| GET /api/oauth/email/bind | EmailBind | user_profile.go |
| POST /api/topup | AdminTopUp | user_billing.go |
| GET /api/user/ | GetAllUsers | user_admin.go |
| GET /api/user/search | SearchUsers | user_admin.go |
| GET /api/user/:id | GetUser | user_admin.go |
| POST /api/user/ | CreateUser | user_admin.go |
| POST /api/user/manage | ManageUser | user_admin.go |
| PUT /api/user/ | UpdateUser | user_admin.go |
| DELETE /api/user/:id | DeleteUser | user_admin.go |

---

## 五、风险点和注意事项

### 5.1 高风险

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| `SetupLogin` 被外部 `controller/auth/*.go` 广泛引用 | 8 个 OAuth 文件都调用 `controller.SetupLogin` | 拆分后仍在 `controller` 包，调用路径不变；但必须确保编译通过 |
| `Register` 函数 201 行，逻辑复杂 | 涉及注册 + 邀请奖励 + 推广关系绑定 | 本次仅拆分文件归属，不改动函数内部逻辑 |

### 5.2 中风险

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| `user_common.go` 定位模糊 | 共享函数可能随时间膨胀 | 严格限定：仅放置被 2 个以上子文件引用的函数；仅被 1 个文件引用的辅助函数应放在对应子文件中 |
| import 冗余 | 拆分后各子文件只需部分 import，需各自精简 | 每个子文件独立维护 import 块，用 `goimports` 自动整理 |

### 5.3 低风险

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| 合并冲突 | 多人同时编辑不同子文件时可能冲突 | Git 对多文件变更的合并能力强于单文件，反而降低冲突概率 |
| 文件数增加 | 从 1 个变为 8 个 | controller 目录已有 90+ 文件，8 个 user_*.go 符合现有风格 |

### 5.4 执行步骤建议

1. **先创建 `user_common.go`** — 放置共享函数，确保编译通过
2. **按依赖顺序逐个创建子文件** — user_auth.go → user_admin.go → user_self.go → user_dashboard.go → user_token.go → user_profile.go → user_billing.go
3. **每创建一个子文件后立即 `go build`** — 确保编译通过
4. **最后清空 `user.go`** — 确认所有函数已迁移后，删除原 user.go 或将其变为仅含 package 声明的占位文件
5. **运行 `goimports`** — 清理各文件 import
6. **运行测试** — `go test ./controller/...`

### 5.5 后续优化建议（本次不执行）

- `Register` 函数（201 行）可考虑将"邀请奖励 + 推广关系绑定"逻辑提取为独立函数
- `ManageUser` 函数（134 行）的 switch-case 可考虑拆分为独立的小函数
- `GetUserDashboard` 的聚合排序逻辑可考虑移入 model 层
