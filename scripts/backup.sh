#!/bin/bash
# ============================================================
# QuantumClaw 备份脚本
# ============================================================
# 用法：
#   ./scripts/backup.sh              # 备份到 ./backups/
#   ./scripts/backup.sh /path/to/backups  # 备份到指定目录
#
# 备份内容：
#   - 数据库（SQLite 文件或 MySQL dump）
#   - .env 配置文件
#   - .session_secret（Session 密钥）
#   - CryptoSecret / JWTSecret（从 .env 中）
# ============================================================

set -e

# ---- 配置 ----
BACKUP_DIR="${1:-./backups}"
mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_NAME="quantumclaw-backup-$TIMESTAMP"
TEMP_DIR=$(mktemp -d)
BACKUP_ROOT="$TEMP_DIR/$BACKUP_NAME"

mkdir -p "$BACKUP_ROOT"

echo "============================================================"
echo "  QuantumClaw 备份脚本"
echo "  时间: $TIMESTAMP"
echo "============================================================"

# ---- 加载 .env ----
ENV_FILE=".env"
if [ -f "$ENV_FILE" ]; then
    echo "[1/5] 加载 .env 配置..."
    # 只加载非注释、非空行
    export $(grep -v '^#' "$ENV_FILE" | grep -v '^$' | xargs)
else
    echo "[!] 警告: .env 文件不存在"
fi

# ---- 备份数据库 ----
echo "[2/5] 备份数据库..."

if [ -z "$SQL_DSN" ]; then
    # SQLite
    SQLITE_FILE="./sqlite/quantumclaw.db"
    if [ -f "$SQLITE_FILE" ]; then
        cp "$SQLITE_FILE" "$BACKUP_ROOT/quantumclaw.db"
        DB_SIZE=$(du -h "$SQLITE_FILE" | cut -f1)
        echo "  [✓] SQLite 数据库已备份 ($DB_SIZE)"
    else
        echo "  [!] 警告: SQLite 文件不存在: $SQLITE_FILE"
    fi
else
    # MySQL
    echo "  [*] 检测到 MySQL 数据库，尝试自动备份..."

    # 解析 SQL_DSN: user:pass@tcp(host:port)/dbname?...
    # 简单解析（假设格式标准）
    DSN_NO_PREFIX=${SQL_DSN#*@tcp(}
    DSN_NO_PREFIX=${DSN_NO_PREFIX#*@/}  # 兼容无 tcp() 格式
    MYSQL_HOST=$(echo "$DSN_NO_PREFIX" | grep -oP '^[^:]+' || echo "127.0.0.1")
    MYSQL_PORT=$(echo "$DSN_NO_PREFIX" | grep -oP ':\K\d+' | head -1 || echo "3306")
    MYSQL_USER=$(echo "$SQL_DSN" | grep -oP '//\K[^:]+' || echo "root")
    MYSQL_PASS=$(echo "$SQL_DSN" | grep -oP '://[^:]+:\K[^@]+' || echo "")
    MYSQL_DB=$(echo "$DSN_NO_PREFIX" | grep -oP '/\K[^?]+' || echo "quantumclaw")

    # 尝试用 mysqldump
    if command -v mysqldump &> /dev/null; then
        echo "  [*] 使用 mysqldump 备份..."
        if [ -n "$MYSQL_PASS" ]; then
            mysqldump -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASS" "$MYSQL_DB" > "$BACKUP_ROOT/mysql_dump.sql"
        else
            mysqldump -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" "$MYSQL_DB" > "$BACKUP_ROOT/mysql_dump.sql"
        fi
        DB_SIZE=$(du -h "$BACKUP_ROOT/mysql_dump.sql" | cut -f1)
        echo "  [✓] MySQL 数据库已备份 ($DB_SIZE)"
    else
        echo "  [!] 警告: mysqldump 未安装，跳过 MySQL 备份"
        echo "  [!] 请手动执行: mysqldump -u $MYSQL_USER -p $MYSQL_DB > backup.sql"
    fi
fi

# ---- 备份 .env ----
echo "[3/5] 备份配置文件..."

if [ -f "$ENV_FILE" ]; then
    cp "$ENV_FILE" "$BACKUP_ROOT/.env"
    echo "  [✓] .env 已备份"
else
    echo "  [!] 警告: .env 不存在"
fi

# ---- 备份 .session_secret ----
if [ -f ".session_secret" ]; then
    cp ".session_secret" "$BACKUP_ROOT/.session_secret"
    echo "  [✓] .session_secret 已备份"
fi

# ---- 备份加密密钥（从 .env 提取关键 Secret）----
echo "[4/5] 备份加密密钥..."

KEYS_FILE="$BACKUP_ROOT/keys.txt"
echo "# QuantumClaw 密钥备份 — $(date)" > "$KEYS_FILE"
echo "" >> "$KEYS_FILE"

if [ -n "$CRYPTO_SECRET" ]; then
    echo "CRYPTO_SECRET=$CRYPTO_SECRET" >> "$KEYS_FILE"
    echo "  [✓] CryptoSecret 已记录"
fi

if [ -n "$SESSION_SECRET" ]; then
    echo "SESSION_SECRET=$SESSION_SECRET" >> "$KEYS_FILE"
    echo "  [✓] SessionSecret 已记录"
fi

if [ -n "$JWT_SECRET" ]; then
    echo "JWT_SECRET=$JWT_SECRET" >> "$KEYS_FILE"
    echo "  [✓] JWTSecret 已记录"
fi

if [ -n "$INITIAL_ROOT_PASSWORD" ]; then
    echo "INITIAL_ROOT_PASSWORD=$INITIAL_ROOT_PASSWORD" >> "$KEYS_FILE"
    echo "  [✓] InitialRootPassword 已记录"
fi

# 如果 keys.txt 只有注释，删除它
if [ $(wc -l < "$KEYS_FILE") -le 3 ]; then
    rm "$KEYS_FILE"
    echo "  [!] 无密钥需要备份（可能已存储在 .env 中）"
else
    echo "  [✓] 密钥已保存到 keys.txt"
fi

# ---- 打包 ----
echo "[5/5] 打包备份文件..."

BACKUP_TAR="$BACKUP_DIR/$BACKUP_NAME.tar.gz"
tar -czf "$BACKUP_TAR" -C "$TEMP_DIR" "$BACKUP_NAME"

# 清理临时目录
rm -rf "$TEMP_DIR"

# ---- 完成 ----
BACKUP_SIZE=$(du -h "$BACKUP_TAR" | cut -f1)
echo ""
echo "============================================================"
echo "  ✅ 备份完成！"
echo ""
echo "  备份文件: $BACKUP_TAR"
echo "  文件大小: $BACKUP_SIZE"
echo "  备份内容:"
tar -tzf "$BACKUP_TAR" | head -20 | sed 's/^/    /'
echo ""
echo "  恢复方法:"
echo "    tar -xzf $BACKUP_TAR -C /tmp"
echo "    ./scripts/restore.sh /tmp/$BACKUP_NAME"
echo "============================================================"
