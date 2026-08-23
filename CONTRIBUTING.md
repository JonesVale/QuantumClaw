# 贡献指南

## 欢迎贡献！

我们欢迎所有形式的贡献，包括但不限于：
- 报告 Bug
- 提出新功能建议
- 提交代码修复
- 改进文档
- 翻译文档

---

## 开发环境设置

### 前置要求

- Go 1.24+
- MySQL 5.7+ 或 8.0+
- Redis 6.0+

### 本地开发

```bash
# 1. Fork 本仓库
# 2. Clone 到你的本地
git clone https://github.com/your-username/QuantumClaw-Open.git
cd QuantumClaw-Open

# 3. 配置环境变量
cp .env.example .env
# 编辑 .env 文件

# 4. 创建数据库
mysql -u root -p -e "CREATE DATABASE quantumclaw_open;"

# 5. 执行迁移
mysql -u root -p quantumclaw_open < migrations/001_init.sql

# 6. 编译运行
go build -o quantumclaw ./cmd/server
./quantumclaw

# 7. 运行测试
go test ./... -v
```

---

## 代码规范

### Go 代码规范

- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码风格
- 函数和变量命名遵循 Go 惯例
- 添加必要的注释

### 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
feat: 添加新功能
fix: 修复 bug
docs: 更新文档
style: 代码格式（不影响代码运行的变动）
refactor: 重构（即不是新增功能，也不是修改 bug 的代码变动）
test: 添加测试
chore: 构建过程或辅助工具的变动
```

---

## 贡献流程

1. **Fork** 本仓库
2. **创建分支** `git checkout -b feature/AmazingFeature`
3. **提交更改** `git commit -c 'feat: Add some AmazingFeature'`
4. **推送到分支** `git push origin feature/AmazingFeature`
5. **开启 Pull Request**

---

## 开源范围说明

### ✅ 开源模块

- AI 模型路由系统
- 多级分销系统
- 企业组织架构管理
- 异步任务处理 (Midjourney, Video, Suno)
- 级联节点架构
- 订阅系统
- 渠道管理基础功能

### ❌ 闭源模块 (商业许可)

- 支付处理 (充值、提现)
- 金融结算 (对账、分账)
- 用户身份认证 (实名认证)
- 安全加密模块

---

## 报告问题

请使用 GitHub Issues 报告问题，并提供：

- 问题描述
- 复现步骤
- 预期行为
- 实际行为
- 环境信息 (OS, Go版本, 数据库版本)

---

## 商业许可

如需使用闭源模块，请联系：

- **邮箱**: business@ctji.cn
- **官网**: https://www.ctji.cn
- **公司**: 深圳市中科劲纬智能有限公司

---

## 许可证

本项目采用双许可模式：
- 开源部分: AGPL-3.0
- 闭源部分: 商业许可证

---

*感谢你的贡献！*
