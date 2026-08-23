# 项目结构

## 目录说明

```
QuantumClaw-Open/
│
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go         # 主入口文件
│
├── common/                 # 通用工具库 (开源)
│   ├── config/             # 配置管理
│   ├── helper/             # 工具函数
│   ├── logger/             # 日志系统
│   ├── random/             # 随机数生成
│   └── email/              # 邮件服务
│
├── middleware/             # 中间件 (开源)
│   ├── auth.go             # 认证中间件
│   ├── cache.go            # 缓存中间件
│   ├── cors.go             # CORS 处理
│   ├── gzip.go             # GZIP 压缩
│   ├── rate_limit.go       # 限流中间件
│   └── security_headers.go # 安全 headers
│
├── model/                  # 数据模型 (部分开源)
│   ├── user.go             # 用户模型
│   ├── channel.go          # 渠道模型
│   ├── subscription.go     # 订阅模型
│   ├── async_task.go       # 异步任务模型
│   ├── cascade_node.go     # 级联节点模型
│   ├── enterprise.go       # 企业模型
│   │
│   └── payment/            # ❌ 支付模型 (闭源)
│   └── financial/          # ❌ 金融模型 (闭源)
│   └── security/           # ❌ 安全模型 (闭源)
│
├── service/                # 业务服务 (部分开源)
│   ├── billing.go          # 计费服务
│   ├── subscription.go     # 订阅服务
│   ├── async_task_service.go # 异步任务服务
│   ├── cascade_service.go  # 级联服务
│   ├── enterprise_service.go # 企业服务
│   │
│   └── payment_service.go  # ❌ 支付服务 (闭源)
│   └── financial_service.go # ❌ 金融服务 (闭源)
│   └── security_service.go # ❌ 安全服务 (闭源)
│
├── controller/             # 控制器 (部分开源)
│   ├── user.go             # 用户控制器
│   ├── channel.go          # 渠道控制器
│   ├── subscription.go     # 订阅控制器
│   ├── async_task.go       # 异步任务控制器
│   ├── cascade.go          # 级联控制器
│   ├── enterprise.go       # 企业控制器
│   │
│   └── payment_controller.go   # ❌ 支付控制器 (闭源)
│   └── financial_controller.go # ❌ 金融控制器 (闭源)
│   └── security_controller.go  # ❌ 安全控制器 (闭源)
│
├── router/                 # 路由 (部分开源)
│   ├── api_open.go         # 公开路由
│   ├── api_public.go       # 公共路由
│   │
│   └── api_payment.go      # ❌ 支付路由 (闭源)
│   └── api_financial.go    # ❌ 金融路由 (闭源)
│
├── pkg/                    # 第三方包封装 (开源)
│   ├── ai/                 # AI SDK 封装
│   ├── email/              # 邮件服务
│   └── queue/              # 消息队列
│
├── migrations/             # 数据库迁移 (开源)
│   └── 001_init.sql        # 初始化脚本
│
├── docs/                   # 文档 (开源)
│   └── DEPLOYMENT.md       # 部署文档
│
├── deploy/                 # 部署配置 (开源)
│   ├── docker/
│   ├── k8s/
│   └── nginx/
│
├── test/                   # 测试 (开源)
│   ├── unit/
│   └── integration/
│
├── .env.example            # 环境变量模板
├── .gitignore              # Git 忽略配置
├── Dockerfile              # Docker 配置
├── docker-compose.yml      # Docker Compose
├── fly.toml                # Fly.io 配置
├── go.mod                  # Go 模块定义
├── go.sum                  # Go 依赖校验
├── LICENSE                 # AGPL-3.0 许可证
├── LICENSE_COMMERCIAL.md   # 商业授权协议
├── README.md               # 项目说明
├── CONTRIBUTING.md         # 贡献指南
├── CHANGELOG.md            # 版本历史
└── SECURITY.md             # 安全说明
```

## 开源范围

### ✅ 开源模块 (约60%)

- common/
- middleware/
- model/user.go
- model/channel.go
- model/subscription.go
- model/async_task.go
- model/cascade_node.go
- model/enterprise.go
- service/billing.go
- service/subscription.go
- service/async_task_service.go
- service/cascade_service.go
- service/enterprise_service.go
- controller/user.go
- controller/channel.go
- controller/subscription.go
- controller/async_task.go
- controller/cascade.go
- controller/enterprise.go
- router/api_open.go
- router/api_public.go
- pkg/
- migrations/
- docs/
- deploy/
- test/

### ❌ 闭源模块 (约40%)

- model/payment/
- model/financial/
- model/security/
- service/payment_service.go
- service/financial_service.go
- service/security_service.go
- controller/payment_controller.go
- controller/financial_controller.go
- controller/security_controller.go
- router/api_payment.go
- router/api_financial.go
- config/*.env

---

*深圳市中科劲纬智能有限公司*
*官网: https://www.ctji.cn*
