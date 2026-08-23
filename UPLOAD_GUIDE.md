# QuantumClaw Open Source 上传指南

## 版权所有
**深圳市中科劲纬智能有限公司**
**官网**: https://www.ctji.cn
**邮箱**: vale@ctji.cn

---

## ⚠️ 安全警告

**请勿在脚本或配置文件中硬编码 GitHub Token！**

本指南使用环境变量传递 Token，确保安全性。

---

## 上传步骤

### Step 1: 准备环境

```bash
# 进入开源目录
cd H:\AiData\openclaw\workspace\QuantumClaw-Open

# 确认目录结构
ls -la
```

### Step 2: 设置 GitHub Token

**Linux/Mac:**
```bash
export GITHUB_TOKEN=github_pat_xxxxxxxx
```

**Windows PowerShell:**
```powershell
$env:GITHUB_TOKEN = "github_pat_xxxxxxxx"
```

### Step 3: 执行安全检查

```bash
python security_check.py .
```

确保没有敏感信息泄露。

### Step 4: 执行上传脚本

```bash
bash upload_to_github.sh
```

---

## 手动上传（如果脚本失败）

### 1. 初始化仓库
```bash
cd QuantumClaw-Open
git init
git config user.email "opensource@ctji.cn"
git config user.name "QuantumClaw Open Source"
```

### 2. 添加文件
```bash
git add .
```

### 3. 提交
```bash
git commit -m "Initial open source release (AGPL-3.0)

Copyright (C) 2026 深圳市中科劲纬智能有限公司
Website: https://www.ctji.cn
Email: vale@ctji.cn"
```

### 4. 创建远程仓库
在 GitHub 上创建新仓库：
- 名称: `QuantumClaw-Open`
- 可见性: Public
- 不要添加 README（脚本会自动添加）

### 5. 推送
```bash
git remote add origin https://github.com/ctji/QuantumClaw-Open.git
git push -u origin main
```

---

## 开源范围确认

### ✅ 已上传的开源模块

- common/ - 通用工具库
- middleware/ - 中间件
- model/user.go - 用户模型
- model/channel.go - 渠道模型
- model/subscription.go - 订阅模型
- model/async_task.go - 异步任务模型
- model/cascade_node.go - 级联节点模型
- model/enterprise.go - 企业模型
- service/billing.go - 计费服务
- service/subscription.go - 订阅服务
- service/async_task_service.go - 异步任务服务
- controller/user.go - 用户控制器
- controller/channel.go - 渠道控制器
- controller/subscription.go - 订阅控制器
- router/api_open.go - 公开路由
- docs/ - 文档
- migrations/ - 数据库迁移
- deploy/ - 部署配置

### ❌ 已移除的闭源模块

- payment/ - 支付处理
- financial/ - 金融结算
- security/ - 安全加密
- authentication/ - 身份认证
- config/*.env - 敏感配置
- 所有包含密钥的文件

---

## 上传后验证

### 1. 检查仓库
```bash
git remote -v
git log --oneline
```

### 2. 访问 GitHub
打开浏览器访问：
https://github.com/ctji/QuantumClaw-Open

确认：
- [ ] README.md 显示正确
- [ ] LICENSE 文件存在
- [ ] 无敏感信息泄露
- [ ] 闭源模块已移除

---

## 许可证说明

- **开源部分**: AGPL-3.0
- **闭源部分**: 商业许可证（需单独购买）
- **商业用途**: 请联系 vale@ctji.cn

---

## 技术支持

- **官网**: https://www.ctji.cn
- **邮箱**: vale@ctji.cn
- **问题反馈**: https://github.com/ctji/QuantumClaw-Open/issues

---

**版权所有 (C) 2026 深圳市中科劲纬智能有限公司**
**All Rights Reserved.**
