# QuantumClaw Docker 运行指南

## 快速启动（推荐）

### 方式一：标准部署（MySQL + Redis）

#### Windows (PowerShell)
```powershell
# 1. 复制并编辑配置文件
copy .env.example .env
notepad .env   # 修改必要的配置（见下方"必填配置"）

# 2. 启动 Docker 容器
.\start-docker.ps1
```

#### Linux / WSL / macOS
```bash
# 1. 复制并编辑配置文件
cp .env.example .env
vim .env   # 修改必要的配置（见下方"必填配置"）

# 2. 启动 Docker 容器
docker compose up -d --build

# 3. 查看日志
docker compose logs -f

# 4. 停止服务
docker compose down
```

### 方式二：轻量部署（SQLite + 内存缓存）
适合开发、测试或小规模生产。

```bash
docker compose -f docker-compose.sqlite.yml up -d --build
```

## 必填配置

| 变量 | 说明 |
|------|------|
| `SESSION_SECRET` | 会话加密密钥（必须修改为随机字符串） |
| `INITIAL_ROOT_TOKEN` | 初始 root 用户令牌 |
| `INITIAL_ROOT_ACCESS_TOKEN` | 初始 root 访问令牌 |

## 可选配置

### 日志
- Docker 自带日志轮转，启动时可指定：
```bash
docker compose up -d --log-opt max-size=10m --log-opt max-file=3
```
- 非 Docker 部署建议使用 logrotate：
```
# /etc/logrotate.d/quantumclaw
/app/logs/quantumclaw-*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    copytruncate
}
```

### 安全加固
- 生产环境启用 HTTPS，推荐使用 Nginx 反向代理
- 配置 `ALLOWED_ORIGINS` 为实际域名
- 启用 `CSP_ENABLED=true`、`HSTS_ENABLED=true`

### 数据库
- SQLite 模式适合单机小规模，无需 MySQL/Redis
- 生产环境建议 MySQL 8.0 + Redis

## 端口说明
- 服务端口: 3666（可通过 `PORT` 环境变量修改）
- MySQL 端口: 3306
- Redis 端口: 6379

## 健康检查
```bash
curl http://localhost:3666/api/status
```

## 常见问题
详见 [README.md](./README.md)
