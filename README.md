# QuantumClaw Open Source Edition

**版权所有 (C) 2026 深圳市中科劲纬智能有限公司**  
**官网**: https://www.ctji.cn  
**邮箱**: business@ctji.cn

---

## 许可证说明

本项目采用**双许可模式**：

1. **开源许可证**: GNU Affero General Public License v3.0 (AGPL-3.0)
   - 仅适用于开源部分代码
   - 仅限个人学习、研究、交流使用
   - **禁止商业用途**
   - 修改后必须开源

2. **商业许可证**: 需要单独授权
   - 商业应用必须购买商业许可证
   - 包含闭源模块（支付、金融、安全等）
   - 联系: business@ctji.cn

---

## 项目简介

QuantumClaw 是一个高性能 AI API 网关平台，支持多种 AI 模型路由、分销体系、企业客户管理等功能。

**注意**: 本开源版本仅包含通用功能模块，金融支付相关功能为闭源商业模块。

---

## 开源模块

### ✅ 已开源模块（AGPL-3.0）

| 模块 | 说明 |
|------|------|
| AI 模型路由 | 多模型负载均衡、故障转移 |
| 多级分销 | 分销商管理、返佣计算 |
| 企业组织 | 组织架构、部门管理 |
| 异步任务 | Midjourney, Video, Suno 任务处理 |
| 级联架构 | 子节点注册、心跳、Token 同步 |
| 订阅系统 | 订阅计划、额度管理 |
| 渠道管理 | 渠道配置、余额查询 |
| 中间件 | 认证、限流、缓存等 |

### ❌ 闭源模块（需商业许可）

| 模块 | 说明 |
|------|------|
| 支付处理 | 充值、提现、支付回调 |
| 金融结算 | 对账、分账、收益计算 |
| 实名认证 | 身份验证、实名认证 |
| 安全加密 | 密钥管理、签名验证 |
| 配置管理 | 敏感配置、密钥存储 |

---

## 快速开始

### 环境要求

- Go 1.24+
- MySQL 5.7+ / 8.0+
- Redis 6.0+
- (可选) Docker

### 安装

```bash
# 克隆仓库
git clone https://github.com/ctji/QuantumClaw-Open.git
cd QuantumClaw-Open

# 安装依赖
go mod download

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，填写必要配置

# 编译
go build -o quantumclaw ./cmd/server

# 运行
./quantumclaw
```

### Docker 部署

```bash
# 使用 Docker Compose
docker-compose up -d

# 查看日志
docker-compose logs -f
```

---

## 使用说明

### 个人学习

- ✅ 可以下载、学习、研究源代码
- ✅ 可以修改代码用于个人学习
- ✅ 可以分享学习心得
- ⚠️ 不得用于商业用途

### 商业用途

- ❌ **禁止**直接使用开源版本进行商业运营
- ❌ **禁止**将闭源模块用于商业项目
- ⚠️ 需要购买商业许可证

**商业授权联系**:
- 邮箱: business@ctji.cn
- 官网: https://www.ctji.cn
- 公司: 深圳市中科劲纬智能有限公司

---

## API 文档

启动服务后访问: http://localhost:3666/api/swagger

---

## 贡献指南

欢迎提交 Issue 和 Pull Request！

### 贡献流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -c 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

详见 [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 许可证

本项目采用双许可模式：

- 开源部分: AGPL-3.0
- 闭源部分: 商业许可证

详见 [LICENSE](LICENSE) 和 [LICENSE_COMMERCIAL.md](LICENSE_COMMERCIAL.md)

---

## 安全说明

- 本开源版本不包含任何敏感信息
- 所有密钥使用环境变量配置
- 禁止提交 .env 文件到仓库
- 定期更新依赖包

---

## 联系我们

- **公司**: 深圳市中科劲纬智能有限公司
- **官网**: https://www.ctji.cn
- **邮箱**: info@ctji.cn
- **商业**: business@ctji.cn

---

## 免责声明

本软件按"原样"提供，不提供任何明示或暗示的保证。使用者自行承担使用风险。

**Copyright (C) 2026 深圳市中科劲纬智能有限公司**  
**All Rights Reserved.**
