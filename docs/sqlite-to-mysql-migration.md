# SQLite → MySQL 迁移指南

> QuantumClaw 官方迁移文档  
> 适用版本：v2.2.0+  
> 最后更新：2026-06-08

---

## 1. 为什么要迁移？

| 特性 | SQLite | MySQL |
|------|--------|-------|
| 并发写入 | ❌ 单写锁 | ✅ 行级锁 |
| 多机部署 | ❌ 文件锁 | ✅ 网络访问 |
| 备份热切换 | ❌ 需停服 | ✅ 在线备份 |
| 数据量上限 | ~1TB | TB 级 |
| 适用场景 | 单机开发/测试 | 生产环境 |

**结论**：生产环境强烈建议用 MySQL；开发/测试可继续用 SQLite。

---

## 2. 迁移前准备

### 2.1 备份 SQLite 数据库

```bash
# 找到 SQLite 文件位置（默认 ./sqlite/quantumclaw.db）
cd /path/to/QuantumClaw
ls -lh sqlite/quantumclaw.db

# 备份（必须！）
cp sqlite/quantumclaw.db sqlite/quantumclaw.db.backup-$(date +%Y%m%d)
```

### 2.2 记录当前配置

```bash
# 记录当前 .env 的关键配置
cat .env | grep -E "SQL_DSN|SERVER_ADDRESS|PORT"
```

### 2.3 安装 MySQL（如未安装）

**推荐版本**：MySQL 8.0+ 或 MariaDB 10.6+

```bash
# Ubuntu/Debian
sudo apt install mysql-server -y

# CentOS/RHEL
sudo yum install mysql-server -y

# macOS
brew install mysql
brew services start mysql
```

---

## 3. 迁移方法（推荐：Method 1）

### Method 1：GORM 自动建表 + 数据导入（推荐）

**原理**：利用 QuantumClaw 的 `AutoMigrate` 自动创建 MySQL 表结构，然后只导入数据。

#### Step 1：创建 MySQL 数据库和用户

```sql
-- 登录 MySQL
mysql -u root -p

-- 创建数据库（字符集必须用 utf8mb4）
CREATE DATABASE quantumclaw
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

-- 创建用户
CREATE USER 'quantumclaw'@'%' IDENTIFIED BY 'your_strong_password';
GRANT ALL PRIVILEGES ON quantumclaw.* TO 'quantumclaw'@'%';
FLUSH PRIVILEGES;
```

#### Step 2：配置 `.env` 指向 MySQL

```bash
# 编辑 .env，设置 SQL_DSN（原来不设置或用 SQLite 时留空）
# MySQL DSN 格式：
#   user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local

SQL_DSN=quantumclaw:your_strong_password@tcp(127.0.0.1:3306)/quantumclaw?charset=utf8mb4&parseTime=True&loc=Local
```

> **注意**：设置 `SQL_DSN` 后，QuantumClaw 会自动使用 MySQL（不再用 SQLite）。

#### Step 3：启动 QuantumClaw，自动建表

```bash
# 启动 QuantumClaw（会自动 AutoMigrate 所有表）
go run main.go

# 或者 binary 启动
./quantumclaw
```

检查日志，确认看到：
```
using MySQL as database
database migrated
```

此时 MySQL 数据库已有完整的表结构（但无数据）。

**按 Ctrl+C 停止**。

#### Step 4：导出 SQLite 数据

使用项目自带的迁移脚本（推荐）或手动导出。

```bash
# 使用 sqlite3 CLI 导出（需要 sqlite3 工具）
sqlite3 sqlite/quantumclaw.db ".dump" > /tmp/sqlite_dump.sql
```

#### Step 5：转换并导入数据到 MySQL

**⚠️ 关键**：SQLite 和 MySQL 的 SQL 语法有差异，不能直接导入 `.dump` 文件。

**推荐方式**：使用项目提供的迁移工具（如有）或手动表-by-表导入。

```bash
# 导出为 SQLite 兼容的 INSERT 语句（不含 CREATE TABLE）
sqlite3 sqlite/quantumclaw.db ".mode insert" ".dump" > /tmp/data_only.sql

# 手动编辑 data_only.sql，修复语法差异：
# 1. 移除所有 CREATE TABLE / CREATE INDEX 语句
# 2. 将 SQLite 的布尔值（0/1）保持不变（MySQL TINYINT(1) 兼容）
# 3. 将 Unix timestamp 整数保持不变
# 4. 移除 SQLite 特定的 PRAGMA 语句

# 导入到 MySQL
mysql -u quantumclaw -p quantumclaw < /tmp/data_only_fixed.sql
```

#### Step 6：验证数据

```sql
-- 登录 MySQL 验证
mysql -u quantumclaw -p quantumclaw

-- 检查各表数据量（应与 SQLite 一致）
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM channels;
SELECT COUNT(*) FROM tokens;
SELECT COUNT(*) FROM balance_logs;
```

#### Step 7：重新启动 QuantumClaw

```bash
go run main.go
# 或
./quantumclaw
```

---

### Method 2：使用第三方工具（`sqlite3` → `mysql`）

**适用场景**：数据量较大（>10万行），需要自动化迁移。

```bash
# 使用 sqlite3-to-mysql 工具
pip install sqlite3-to-mysql

sqlite3mysql \
  --sqlite-file sqlite/quantumclaw.db \
  --mysql-host 127.0.0.1 \
  --mysql-port 3306 \
  --mysql-user quantumclaw \
  --mysql-password 'your_password' \
  --mysql-database quantumclaw \
  --mysql-integer-type INT \
  --mysql-charset utf8mb4
```

---

## 4. 配置变更对照

### `.env` 关键变更

| 配置项 | SQLite 时 | MySQL 时 |
|--------|-----------|----------|
| `SQL_DSN` | 不设置或留空 | `user:password@tcp(host:port)/dbname?...` |

**示例**：

```bash
# SQLite（默认）
# SQL_DSN=   # 不设置

# MySQL
SQL_DSN=quantumclaw:password@tcp(127.0.0.1:3306)/quantumclaw?charset=utf8mb4&parseTime=True&loc=Local
```

---

## 5. 代码兼容性说明

QuantumClaw 已内置 SQLite/MySQL 兼容层，`UsingSQLite` 标志位会自动切换：

| 功能 | SQLite | MySQL | 兼容性 |
|------|--------|-------|--------|
| 日期分组统计 | `strftime(...)` | `DATE(FROM_UNIXTIME(...))` | ✅ 自动切换 |
| 表创建 | `AUTO_INCREMENT` | `AUTO_INCREMENT` | ✅ GORM 处理 |
| 布尔值 | `INT(1) 0/1` | `TINYINT(1)` | ✅ 兼容 |
| Unix timestamp | `INT(11)` | `BIGINT` | ✅ 兼容 |
| 分页查询 | `LIMIT/OFFSET` | `LIMIT/OFFSET` | ✅ 标准 SQL |

**结论**：迁移后无需修改任何代码，QuantumClaw 会自动适配 MySQL。

---

## 6. 回滚方案

如果迁移后出现问题，立即回滚：

```bash
# 1. 停止 QuantumClaw
# Ctrl+C 或 kill 进程

# 2. 恢复 .env（移除 SQL_DSN）
vim .env
# 注释或删除 SQL_DSN 行

# 3. 确认 SQLite 文件完好
ls -lh sqlite/quantumclaw.db
ls -lh sqlite/quantumclaw.db.backup-*/quantumclaw.db

# 4. 如果 SQLite 文件损坏，从备份恢复
cp sqlite/quantumclaw.db.backup-20260608/quantumclaw.db sqlite/quantumclaw.db

# 5. 重新启动（会自动使用 SQLite）
go run main.go
```

---

## 7. 常见问题

### Q1：SQLite 和 MySQL 的数据类型差异？

A：QuantumClaw 使用 GORM 建模，数据类型映射如下：

| Go 类型 | SQLite | MySQL |
|---------|--------|-------|
| `int` / `int64` | `INT` | `BIGINT` |
| `string` | `TEXT` | `VARCHAR(255)` / `TEXT` |
| `bool` | `INT(1)` | `TINYINT(1)` |
| `time.Time` / `int64` | `INT(11)` | `BIGINT` |
| `float64` | `REAL` | `DECIMAL` |

GORM 会自动处理这些差异，无需手动转换。

### Q2：迁移后性能有影响吗？

A：MySQL 在多并发、大数据量场景下**显著优于** SQLite。单用户场景差异不大。

### Q3：可以 SQLite 和 MySQL 混用吗？

A：**不可以**。QuantumClaw 启动时必须确定一种数据库，无法混用。

### Q4：如何验证迁移成功？

A：检查以下各项：

```bash
# 1. 检查日志
grep "using MySQL as database" logs/quantumclaw.log

# 2. 检查数据库表
mysql -u quantumclaw -p quantumclaw -e "SHOW TABLES;"

# 3. 检查数据量
mysql -u quantumclaw -p quantumclaw -e "SELECT COUNT(*) as user_count FROM users;"

# 4. 测试登录
curl http://localhost:3666/api/user/self/info -H "Cookie: ..."

# 5. 测试 API 调用（用有效的 Token）
curl http://localhost:3666/v1/models \
  -H "Authorization: Bearer $TOKEN"
```

---

## 8. Docker Compose 部署（推荐生产环境）

新建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: quantumclaw-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: your_root_password
      MYSQL_DATABASE: quantumclaw
      MYSQL_USER: quantumclaw
      MYSQL_PASSWORD: your_password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

  redis:
    image: redis:7-alpine
    container_name: quantumclaw-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  quantumclaw:
    build: .
    container_name: quantumclaw
    restart: unless-stopped
    depends_on:
      - mysql
      - redis
    ports:
      - "3666:3666"
    volumes:
      - ./data:/app/data
      - ./.env:/app/.env
    environment:
      SQL_DSN: "quantumclaw:your_password@tcp(mysql:3306)/quantumclaw?charset=utf8mb4&parseTime=True&loc=Local"
      REDIS_HOST: redis
      REDIS_PORT: 6379

volumes:
  mysql_data:
  redis_data:
```

启动：

```bash
docker-compose up -d
docker-compose logs -f quantumclaw
```

---

## 9. 总结

| 步骤 | 命令/操作 | 预计时间 |
|------|-----------|----------|
| 1. 备份 SQLite | `cp sqlite/quantumclaw.db ...` | 1 min |
| 2. 创建 MySQL 库 | `CREATE DATABASE ...` | 2 min |
| 3. 配置 `.env` | 设置 `SQL_DSN` | 1 min |
| 4. 启动自动建表 | `go run main.go` | 5 min |
| 5. 导入数据 | 手动或工具 | 10-30 min |
| 6. 验证 | 检查日志 + 数据 | 5 min |
| **总计** | | **25-45 min** |

---

## 附录：完整迁移检查清单

- [ ] SQLite 数据库已备份
- [ ] MySQL 已安装并创建 `quantumclaw` 数据库
- [ ] `.env` 已设置 `SQL_DSN`
- [ ] QuantumClaw 启动日志显示 `using MySQL as database`
- [ ] 所有表已自动创建（`SHOW TABLES`）
- [ ] 数据已导入且记录数一致
- [ ] 用户登录正常
- [ ] API 调用正常
- [ ] 旧的 SQLite 备份已安全保存（至少保留 7 天）

---

*如有问题，请提交 Issue 或联系维护团队。*
