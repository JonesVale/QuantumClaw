#!/bin/bash
set -e

# QuantumClaw 宝塔面板安装脚本
# 支持: install / upgrade / uninstall

MODE="${1:-install}"
BASE_DIR="/www/quantumclaw"
DATA_DIR="${BASE_DIR}/data"
BIN_DIR="${BASE_DIR}/bin"
RELEASE_URL="https://github.com/quantumclaw/quantumclaw/releases/latest/download/quantumclaw-linux-amd64.tar.gz"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

case "$MODE" in
install)
    info "开始安装 QuantumClaw..."

    # 检查依赖
    command -v curl >/dev/null 2>&1 || error "curl 未安装"
    command -v tar >/dev/null 2>&1  || error "tar 未安装"

    # 创建目录结构
    mkdir -p "$DATA_DIR" "$BIN_DIR"

    # 下载
    TMP_TAR="/tmp/quantumclaw.tar.gz"
    info "下载最新版本..."
    curl -fsSL "$RELEASE_URL" -o "$TMP_TAR" || error "下载失败"

    # 解压
    tar -xzf "$TMP_TAR" -C "$BIN_DIR/" || error "解压失败"
    chmod +x "$BIN_DIR/quantumclaw"
    rm -f "$TMP_TAR"

    # 创建 .env（不覆盖已有的）
    if [ ! -f "$BASE_DIR/.env" ]; then
        SESSION_SECRET=$(openssl rand -hex 32 2>/dev/null || uuidgen)
        cat > "$BASE_DIR/.env" <<EOF
# QuantumClaw 配置 — 自动生成
SESSION_SECRET=${SESSION_SECRET}
SQLITE_PATH=${DATA_DIR}/quantumclaw.db
PORT=3666
REGISTER_ENABLED=true
PASSWORD_REGISTER_ENABLED=true
EOF
        info ".env 文件已创建"
    else
        warn ".env 已存在，跳过创建"
    fi

    # 创建 systemd 服务
    cat > /etc/systemd/system/quantumclaw.service <<'SERVICE'
[Unit]
Description=QuantumClaw AI API Gateway
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/www/quantumclaw
ExecStart=/www/quantumclaw/bin/quantumclaw
Restart=always
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
SERVICE

    systemctl daemon-reload
    systemctl enable quantumclaw
    systemctl start quantumclaw

    info "安装完成！"
    echo ""
    echo "========================================"
    echo "  服务状态: $(systemctl is-active quantumclaw)"
    echo "  管理地址: http://$(curl -s ifconfig.me 2>/dev/null || echo 'localhost'):3666"
    echo "  默认账号: root"
    echo "  默认密码: 123456"
    echo "  请登录后立即修改密码！"
    echo "========================================"
    ;;

upgrade)
    info "开始升级 QuantumClaw..."
    TMP_TAR="/tmp/quantumclaw.tar.gz"
    curl -fsSL "$RELEASE_URL" -o "$TMP_TAR" || error "下载失败"
    tar -xzf "$TMP_TAR" -C "$BIN_DIR/" || error "解压失败"
    chmod +x "$BIN_DIR/quantumclaw"
    rm -f "$TMP_TAR"
    systemctl restart quantumclaw
    info "升级完成！"
    ;;

uninstall)
    warn "此操作将删除 QuantumClaw 的所有文件！"
    read -p "确认卸载? (yes/no): " CONFIRM
    [ "$CONFIRM" = "yes" ] || exit 0

    systemctl stop quantumclaw 2>/dev/null || true
    systemctl disable quantumclaw 2>/dev/null || true
    rm -f /etc/systemd/system/quantumclaw.service
    systemctl daemon-reload
    rm -rf "$BASE_DIR"
    info "已卸载 QuantumClaw"
    ;;

status)
    systemctl status quantumclaw 2>&1 || echo "未安装"
    ;;

*)
    echo "用法: bash install.sh [install|upgrade|uninstall|status]"
    exit 1
    ;;
esac
