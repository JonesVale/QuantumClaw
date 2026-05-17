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

### Linux / WSL / macOS
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

### 方式二：SQLite 轻量部署（单容器）

无需 MySQL 和 Redis，适合开发、测试或低负载场景。

#### Windows (PowerShell)
```powershell
# 使用 docker-compose.sqlite.yml 启动
copy .env.example .env
notepad .env   # 修改 SESSION_SECRET 等必要配置
# 使用 SQLite 轻量模式
.\start-docker.sqlite.ps1
```

#### Linux / WSL / macOS
```bash
# 1. 复制并编辑配置文件
cp .env.example .env
vim .env

# 2. 使用 SQLite 模式启动
#    注意：确保 .env 中 SQL_DSN 为空或已注释掉   
docker compose -f docker-compose.sqlite.yml up -d --build
```

> 数据库文件存储在 Docker 卷 `quantumclaw_data` 中，默认路径为 `/app/data/quantumclaw.db`。
> 可通过环境变量 `SQLITE_PATH` 自定义路径。

---

## 必填配置 (.env 文件)

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `SESSION_SECRET` | 会话密钥（随机字符串） | `your-random-secret-key` |
| `INITIAL_ROOT_TOKEN` | 初始 root 令牌 | `root-token-123` |
| `INITIAL_ROOT_ACCESS_TOKEN` | 初始 root 访问令牌 | `root-access-token-456` |
| `MYSQL_ROOT_PASSWORD` | MySQL root 密码 | `quantumclaw` |

---

## 可选配置

### WebAuthn/Passkey（无密码登录）
```env
WEBAUTHN_ENABLED=true
WEBAUTHN_RP_DISPLAY_NAME=QuantumClaw
WEBAUTHN_RP_ID=          # 留空自动从 ServerAddress 提取
WEBAUTHN_ORIGIN=         # 留空自动从 ServerAddress 提取
```

### OAuth 登录
```env
# GitHub
GITHUB_OAUTH_ENABLED=true
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret

# 微信
WECHAT_AUTH_ENABLED=true
WECHAT_CLIENT_ID=your-client-id
WECHAT_CLIENT_SECRET=your-client-secret

# Discord
DISCORD_OAUTH_ENABLED=true
DISCORD_CLIENT_ID=your-client-id
DISCORD_CLIENT_SECRET=your-client-secret
```

### 安全头（CSP/HSTS 等）
```env
CSP_ENABLED=true
HSTS_ENABLED=true
X_FRAME_OPTIONS=DENY
```

---

## 访问地址

启动成功后，访问：
- **主页**: http://localhost:3000
- **API**: http://localhost:3000/api
- **管理后台**: http://localhost:3000/admin

---

## 常用命令

```bash
# 查看运行状态
docker compose ps

# 查看日志
docker compose logs -f

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 重新构建（修改代码后）
docker compose up -d --build
```

---

## 故障排除

### 1. MySQL 启动失败
```bash
# 查看 MySQL 日志
docker compose logs mysql

# 检查端口是否被占用
netstat -ano | findstr :3306   # Windows
lsof -i :3306                   # Linux/macOS
```

### 2.  Redis 启动失败
```bash
# 查看 Redis 日志
docker compose logs redis

# 检查端口是否被占用
netstat -ano | findstr :6379   # Windows
lsof -i :6379                   # Linux/macOS
```

### 3. 应用启动失败
```bash
# 查看应用日志
docker compose logs app

# 进入容器调试
docker compose exec app sh
```

---

## 数据持久化

Docker  volumes：
- `mysql_data`: MySQL 数据文件
- `redis_data`: Redis 数据文件
- `quantumclaw_data`: QuantumClaw 数据文件（SQLite/日志）

手动备份：
```bash
# 备份 MySQL 数据
docker compose exec mysql mysqldump -u root -p quantumclaw > backup.sql

# 备份 volumes
docker run --rm -v quantumclaw_mysql_data:/data -v $(pwd):/backup ubuntu tar czf /backup/mysql_backup.tar.gz /data
```
