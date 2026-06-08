#!/bin/bash
# ============================================================
# QuantumClaw 恢复脚本
# ============================================================
# 用法：
#   ./scripts/restore.sh <backup_dir>
#
# 参数：
#   <backup_dir>  备份解压后的目录路径
#
# 示例：
#   tar -xzf quantumclaw-backup-20260608-120000.tar.gz -C /tmp
#   ./scripts/restore.sh /tmp/quantumclaw-backup-20260608-120000
#
# 恢复内容：
#   - .env 配置文件
#   - .session_secret（Session 密钥）
#   - 数据库（SQLite 文件或 MySQL dump）
# ============================================================

set -e

# ---- 参数检查 ----
if [ -z "$1" ]; then
    echo "用法: $0 <backup_dir>"
    echo ""
    echo "示例:"
    echo "  tar -xzf quantumclaw-backup-20260608-120000.tar.gz -C /tmp"
    echo "  $0 /tmp/quantumclaw-backup-20260608-120000"
    exit 1
fi

BACKUP_DIR="$1"

if [ ! -d "$BACKUP_DIR" ]; then
    echo "[!] 错误: 备份目录不存在: $BACKUP_DIR"
    exit 1
fi

echo "============================================================"
echo "  QuantumClaw 恢复脚本"
echo "  备份目录: $BACKUP_DIR"
echo "============================================================"

# ---- 停止 QuantumClaw（如果正在运行）----
echo "[0/6] 检查 QuantumClaw 进程..."
if pgrep -f "quantumclaw" > /dev/null; then
    echo "  [!] 警告: QuantumClaw 正在运行，请先停止！"
    echo "  [!] 执行: pkill -f quantumclaw"
    read -p "  按 Enter 继续（或 Ctrl+C 取消）..."
fi

# ---- 备份当前配置（防止恢复失败）----
echo "[1/6] 备份当前配置（安全起见）..."
if [ -f ".env" ]; then
    cp ".env" ".env.backup-$(date +%Y%m%d-%H%M%S)"
    echo "  [✓] 当前 .env 已备份为 .env.backup-*"
fi

# ---- 恢复 .env ----
echo "[2/6] 恢复 .env 配置文件..."

if [ -f "$BACKUP_DIR/.env" ]; then
    cp "$BACKUP_DIR/.env" .env
    echo "  [✓] .env 已恢复"
else
    echo "  [!] 警告: 备份中无 .env 文件"
fi

# ---- 恢复 .session_secret ----
echo "[3/6] 恢复 Session 密钥..."

if [ -f "$BACKUP_DIR/.session_secret" ]; then
    cp "$BACKUP_DIR/.session_secret" .session_secret
    echo "  [✓] .session_secret 已恢复"
else
    echo "  [!] 提示: 备份中无 .session_secret（可能是新版本）"
fi

# ---- 恢复数据库 ----
echo "[4/6] 恢复数据库..."

# 检测备份类型
if [ -f "$BACKUP_DIR/quantumclaw.db" ]; then
    # SQLite
    echo "  [*] 检测到 SQLite 备份"
    mkdir -p sqlite
    cp "$BACKUP_DIR/quantumclaw.db" sqlite/quantumclaw.db
    echo "  [✓] SQLite 数据库已恢复"
elif [ -f "$BACKUP_DIR/mysql_dump.sql" ]; then
    # MySQL
    echo "  [*] 检测到 MySQL 备份"
    echo ""
    echo "  ========================================"
    echo "  请手动执行以下命令恢复 MySQL 数据库："
    echo ""
    echo "  mysql -u <user> -p < <backup_dir>/mysql_dump.sql"
    echo ""
    echo "  或者（如果 .env 已恢复）："
    source "$BACKUP_DIR/.env" 2>/dev/null || true
    if [ -n "$SQL_DSN" ]; then
        echo "  mysql 恢复命令（请根据实际配置调整）："
        echo "    mysql -u root -p quantumclaw < $BACKUP_DIR/mysql_dump.sql"
    fi
    echo "  ========================================"
    echo ""
    read -p "  完成 MySQL 恢复后，按 Enter 继续..."
else
    echo "  [!] 警告: 备份中无数据库文件"
fi

# ---- 恢复密钥 ----
echo "[5/6] 恢复加密密钥..."

if [ -f "$BACKUP_DIR/keys.txt" ]; then
    echo "  [*] 发现 keys.txt，请手动检查并添加到 .env："
    echo ""
    cat "$BACKUP_DIR/keys.txt"
    echo ""
    echo "  [!] 注意: 密钥已打印上方，请确认是否需要添加到 .env"
else
    echo "  [✓] 无额外密钥文件（密钥应在 .env 中）"
fi

# ---- 完成 ----
echo "[6/6] 恢复完成，验证配置..."

# 验证 .env 存在
if [ -f ".env" ]; then
    echo "  [✓] .env 存在"
    # 检查关键配置
    if grep -q "SQL_DSN" .env; then
        echo "  [✓] SQL_DSN 已配置"
    else
        echo "  [!] 警告: SQL_DSN 未配置（将使用 SQLite）"
    fi
else
    echo "  [!] 错误: .env 不存在！"
    exit 1
fi

echo ""
echo "============================================================"
echo "  ✅ 恢复完成！"
echo ""
echo "  下一步："
echo "  1. 检查 .env 配置是否正确"
echo "  2. 启动 QuantumClaw:"
echo "     go run main.go"
echo "     # 或"
echo "     ./quantumclaw"
echo "  3. 检查日志确认启动正常"
echo "============================================================"
