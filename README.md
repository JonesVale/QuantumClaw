# QuantumClaw （量子灵爪）
全球token聚合平台，使用量子加密支付。保障用户支付安全。
QuantumClaw 是一个**企业级 AI API 网关**（基于 One API 衍生开发），提供多模型代理、令牌管理、计费支付、用户认证等功能。

### 核心技术栈

| 层级 | 技术 |
|------|------|
| **后端** | Go 1.23 + Gin + GORM，支持 MySQL/Postgres/SQLite |
| **前端** | React 19 + Rsbuild + TanStack Router + TailwindCSS 4 + Radix UI |
| **容器** | Docker 多阶段构建 (Node 22 → Go 1.23 → Alpine 3.19) |
| **缓存** | Redis（可选），内置 HybridCache |
| **部署** | Docker Compose（MySQL + Redis 标准版 / SQLite 轻量版）|



### 项目健康度（截至 2026-05-17 16:04）

```
Go 构建:   399 文件, 0 错误 ✅
前端构建:  1.50 MB / 443 KB gzip ✅  24 路由页面
Go 测试:   8/8 包 PASS              ✅
前端测试:  12/12 PASS               ✅
```

### 已完成的增强功能

**P0** — i18n 翻译补齐(7 语言)、API-Extended 类型审计、ErrorBoundary 全局组件、OpenAI Responses API、推理参数支持
**P1** — EmptyState 统一空状态组件、Playground MJ/Suno/Video 面板、Vitest 测试框架、模型级限流(glob/Regex+Redis)、Dify ChatFlow
**P2** — 4 种支付渠道 UI 卡片(Stripe/Creem/Waffo/易支付)、2FA/WebAuthn 前端 UI、9 Tab 设置页、Binance Pay、Telegram/LinuxDO 登录、SQLite 单容器部署、Claude Messages/Batches/Vector Stores API
**P3** — 18 页响应式修复、233MB 主题清理、Logo WebP 95%缩小、编码损坏修复、前端测试 12/12

### 关键配置

- **当前端口**: 3666（docker-compose），3000（Dockerfile 默认）
- **数据库**: MySQL（生产）/ SQLite（轻量测试）
- **.env**: 已配置 QRNG (ANU)、WebAuthn、OAuth（全部禁用）、安全头
