# 部署指南

## 环境要求

- Go 1.24+
- MySQL 5.7+ 或 8.0+
- Redis 6.0+
- (可选) Docker

---

## 快速部署

### 方式一：手动部署

#### 1. 安装依赖

```bash
# 克隆仓库
git clone https://github.com/ctji/QuantumClaw-Open.git
cd QuantumClaw-Open

# 安装 Go 依赖
go mod download
```

#### 2. 配置数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE quantumclaw CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 执行迁移
mysql -u root -p quantumclaw < migrations/001_init.sql
```

#### 3. 配置环境变量

```bash
# 复制配置模板
cp .env.example .env

# 编辑配置文件
vim .env
```

必要配置项：
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=quantumclaw

REDIS_HOST=localhost
REDIS_PORT=6379

JWT_SECRET=your_jwt_secret_min_32_chars
```

#### 4. 编译运行

```bash
# 编译
go build -o quantumclaw ./cmd/server

# 运行
./quantumclaw
```

#### 5. 验证安装

```bash
# 健康检查
curl http://localhost:3666/api/status

# 访问 Swagger 文档
open http://localhost:3666/api/swagger
```

---

### 方式二：Docker 部署

#### 1. 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

#### 2. 验证部署

```bash
# 检查服务状态
docker-compose ps

# 健康检查
curl http://localhost:3666/api/status
```

---

## 配置说明

### 数据库配置

```env
DB_HOST=localhost          # 数据库主机
DB_PORT=3306              # 数据库端口
DB_USER=root              # 数据库用户
DB_PASSWORD=your_password # 数据库密码
DB_NAME=quantumclaw       # 数据库名称
```

### Redis 配置

```env
REDIS_HOST=localhost      # Redis 主机
REDIS_PORT=6379          # Redis 端口
REDIS_PASSWORD=          # Redis 密码（可选）
REDIS_DB=0               # Redis 数据库编号
```

### JWT 配置

```env
JWT_SECRET=your_secret_here  # JWT 密钥（至少32位）
JWT_EXPIRE=24h              # Token 过期时间
```

### 邮件配置（可选）

```env
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=your_smtp_password
SMTP_FROM=noreply@example.com
```

---

## 数据库迁移

### 首次部署

```bash
mysql -u root -p quantumclaw < migrations/001_init.sql
```

### 版本更新

```bash
# 查看当前版本
mysql -u root -p quantumclaw -e "SELECT * FROM system_config WHERE key = 'db_version';"

# 执行迁移
mysql -u root -p quantumclaw < migrations/002_*.sql
```

---

## 日志查看

### 应用日志

```bash
tail -f logs/app.log
```

### 错误日志

```bash
tail -f logs/error.log
```

---

## 服务管理

### 使用 systemd (Linux)

```bash
# 安装服务
sudo cp quantumclaw.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable quantumclaw
sudo systemctl start quantumclaw

# 查看状态
sudo systemctl status quantumclaw

# 重启服务
sudo systemctl restart quantumclaw
```

### 使用 supervisor

```ini
[program:quantumclaw]
command=/path/to/quantumclaw
directory=/path/to/app
autostart=true
autorestart=true
stdout_logfile=/var/log/quantumclaw/access.log
stderr_logfile=/var/log/quantumclaw/error.log
```

---

## 性能优化

### 数据库优化

```sql
-- 启用查询缓存
SET GLOBAL query_cache_type = ON;
SET GLOBAL query_cache_size = 64 * 1024 * 1024;  -- 64MB

-- 调整连接数
SET GLOBAL max_connections = 200;
```

### Redis 优化

```bash
# 配置内存限制
redis-cli CONFIG SET maxmemory 256mb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

---

## 安全建议

1. **使用 HTTPS**: 生产环境必须启用 HTTPS
2. **修改默认端口**: 不要使用默认的 3666 端口
3. **强密码**: 数据库和 Redis 使用强密码
4. **定期备份**: 定期备份数据库
5. **更新依赖**: 定期更新 Go 依赖包

---

## 故障排查

### 常见问题

#### 1. 数据库连接失败

```bash
# 检查数据库服务
systemctl status mysql

# 检查连接
mysql -h localhost -u root -p -e "SELECT 1;"
```

#### 2. Redis 连接失败

```bash
# 检查 Redis 服务
systemctl status redis

# 测试连接
redis-cli ping
```

#### 3. 端口被占用

```bash
# 查看端口占用
netstat -tlnp | grep 3666

# 修改端口
# 编辑 .env 文件，修改 SERVER_ADDR
```

---

## 技术支持

- **官网**: https://www.ctji.cn
- **邮箱**: support@ctji.cn
- **GitHub Issues**: https://github.com/ctji/QuantumClaw-Open/issues

---

*最后更新：2026-08-23*
