# 平台资源池协议体系设计方案

> 版本: v1 | 2026-06-06

---

## 一、业务逻辑

### 1.1 核心场景

用户在 Provider A 的店铺选择了一个资源时，需要做出选择：

```
□ 我同意：当 Provider A 的资源异常时，授权平台自动切换到平台资源池中
  其他 Provider 的同型号资源继续为我服务。

□ 我不同意：只用 Provider A 的资源。A 异常时停止服务。

查看平台资源池协议 →
```

### 1.2 协议内容

协议是一份可供用户查阅的文档，内容由平台管理端编辑维护。内容包括但不限于：

- 平台资源池的定义（所有渠道商资源的集合）
- 切换条件（所选 Provider 异常时自动触发）
- 切换范围（同型号资源，按价格排序）
- 数据说明（切换过程中请求数据由新的 Provider 处理）
- 免责声明（平台尽力保障但不保证 100% 可用）

### 1.3 协议版本管理

- 协议有版本号，每次修改递增
- 用户同意时记录同意的版本号
- 协议更新后，已同意的用户需重新确认
- 不同意协议的用户，平台不做资源池切换

---

## 二、数据模型

### 2.1 协议内容表

```go
type PlatformPoolAgreement struct {
    ID          int    `json:"id" gorm:"primaryKey"`
    Version     int    `json:"version" gorm:"uniqueIndex;not null"`
    Title       string `json:"title" gorm:"type:varchar(255)"`
    Content     string `json:"content" gorm:"type:text;not null"`
    Required    bool   `json:"required" gorm:"default:true"`
    PublishedAt int64  `json:"published_at" gorm:"bigint"`
    CreatedAt   int64  `json:"created_at" gorm:"bigint"`
}
```

### 2.2 用户同意记录表

```go
type UserPoolConsent struct {
    ID            int    `json:"id" gorm:"primaryKey"`
    UserID        int    `json:"user_id" gorm:"uniqueIndex;not null"`
    Agreed        bool   `json:"agreed" gorm:"default:false"`
    AgreedVersion int    `json:"agreed_version" gorm:"int;default:0"`
    AgreedAt      int64  `json:"agreed_at" gorm:"bigint"`
    UpdatedAt     int64  `json:"updated_at" gorm:"bigint"`
}
```

---

## 三、API 设计

### 3.1 公开（无需登录）

```
GET /api/agreements/platform-pool/latest
→ 返回最新版本的协议内容（用于展示）
{
    "version": 3,
    "title": "平台资源池使用协议",
    "content": "...",
    "published_at": 1717564800
}
```

### 3.2 用户（需登录）

```
GET /api/user/consent/platform-pool
→ 返回用户的同意状态
{
    "agreed": true,
    "agreed_version": 3,
    "need_reconsent": false  // 协议更新后需要重新同意
}

POST /api/user/consent/platform-pool
{ "agreed": true }
→ 同意/拒绝协议，记录当前版本
{
    "success": true,
    "message": "已同意"
}
```

### 3.3 管理端（Admin ≥ 100）

```
GET /api/admin/agreements/platform-pool
→ 协议版本列表

PUT /api/admin/agreements/platform-pool
{ "title": "...", "content": "..." }
→ 发布新版本，版本号自动递增
```

---

## 四、Relay 层的集成

### 4.1 当前代码位置

`controller/relay.go:90-96` 已有用户偏好检查逻辑：

```go
if retryTimes > 0 {
    usePool := config.OptionMap[fmt.Sprintf("use_pool_%d", userId)]
    if usePool == "0" || usePool == "" {
        retryTimes = 0
    }
}
```

### 4.2 改为查协议同意表

```go
if retryTimes > 0 {
    // 检查用户是否已同意平台资源池协议
    consent, err := dbmodel.GetUserPoolConsent(userId)
    if err != nil || !consent.Agreed || !consent.IsLatestVersion() {
        retryTimes = 0  // 未同意或协议已更新 → 不重试
    }
}
```

### 4.3 检查逻辑

```
用户请求 relay
  ├── 读取 UserPoolConsent
  ├── agreed == false → retryTimes = 0 → 只试一次
  ├── agreed == true → agreed_version < latest_version → retryTimes = 0（需重新同意）
  └── agreed == true → agreed_version == latest_version → 正常重试
```

---

## 五、管理端协议编辑页面

### 5.1 功能

- 查看当前生效的协议版本
- 编辑协议内容（支持 Markdown）
- 发布新版本（自动递增版本号）
- 查看同意统计（总用户数/已同意/未同意/待重新同意）

### 5.2 数据展示

```
协议版本历史
  v3 | 2026-06-06 | 已发布 | 同意用户: 1,234
  v2 | 2026-06-01 | 已过期 | 同意用户: 987
  v1 | 2026-05-15 | 已过期 | 同意用户: 0（首次发布）

协议同意统计
  总用户: 5,000
  已同意最新版: 1,200
  已同意旧版（需重新同意）: 34
  未同意: 3,766
```

---

## 六、前端交互流程

### 6.1 首次弹出

```
用户登录后、首次发起 relay 请求时：
  ┌──────────────────────────────────────┐
  │  平台资源池使用协议                     │
  │                                        │
  │  当您选择的资源异常时，平台可能会自动    │
  │  切换到其他 Provider 的同型号资源。     │
  │                                        │
  │  [查看完整协议]                         │
  │                                        │
  │  □ 我同意上述协议                       │
  │                                        │
  │  [确定]            [暂不（仅用当前店）]  │
  └──────────────────────────────────────┘
```

### 6.2 协议更新后

```
  ┌──────────────────────────────────────┐
  │  平台资源池协议已更新                   │
  │                                        │
  │  [查看变更内容]                         │
  │  [重新同意]            [暂不同意]       │
  └──────────────────────────────────────┘
```

### 6.3 用户设置页

用户信息页面有一个开关：

```
平台资源池
  当资源异常时自动切换到其他 Provider
  [开关] ← 默认关闭
  [查看协议] ← 随时查看
```

---

## 七、实现顺序

| 步骤 | 内容 | 预估 |
|------|------|------|
| 1 | 建表：PlatformPoolAgreement + UserPoolConsent | ~40 行 |
| 2 | Seed 默认协议内容 | ~10 行 |
| 3 | API：获取最新协议、用户同意/查看 | ~100 行 |
| 4 | API：管理端编辑/发布协议 | ~60 行 |
| 5 | Relay 层改用 UserPoolConsent 检查 | ~10 行 |
| 6 | 前端：协议展示页 | 看前端框架 |
| 7 | 前端：同意弹窗 + 设置页开关 | 看前端框架 |
