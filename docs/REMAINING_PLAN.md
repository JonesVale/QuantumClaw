# QuantumClaw 剩余任务执行规划

## 状态: ✅ 全部完成 (2026-05-21 10:15)

### P0 — 用户分销/推广体系 ✅

| # | 任务 | 状态 |
|---|------|------|
| 0.1 | 注册表单增加「邀请码」输入框 + 传给API | ✅ `aff_code` 传入 register() + 输入框UI |
| 0.2 | Wallet 页增加推广码展示卡片 | ✅ 修复结构bug + myAffCode/myCommission/myWithdrawals查询 |
| 0.3 | 前端调用 GetAffCode API + 复制链接 | ✅ 复制链接+toast提示 |

### P1 — 前端权限守卫 ✅

| # | 任务 | 状态 |
|---|------|------|
| 1.1 | Users 页：非 admin 显示「无权访问」 | ✅ 已有isAdmin守卫 |
| 1.2 | Settings 页：非 admin 显示「无权访问」 | ✅ 已有isAdmin守卫 |
| 1.3 | Redemption 页：非 admin 显示「无权访问」 | ✅ 已有isAdmin守卫 |
| 1.4 | Tasks 页：非 admin 显示「无权访问」 | ✅ 已有isAdmin守卫 |

### P2 — 语言翻译体系前端接入 ✅

| # | 任务 | 状态 |
|---|------|------|
| 2.1 | 前端 useTranslation hook 改用 API 获取翻译 | ✅ syncTranslations() 已实现 |
| 2.2 | 翻译种子数据导入（JSON → DB） | ✅ autoSeedIfEmpty() 自动种子 |

### P3 — 发布

| # | 任务 | 状态 |
|---|------|------|
| 3.1 | 全量测试 + 构建 | ✅ go build 0 error / rsbuild build 0.80s |
