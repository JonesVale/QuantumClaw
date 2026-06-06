#!/bin/bash
# ============================================================
# QuantumClaw — SSL Certificate Setup (Let's Encrypt)
# Domain: qscl.link
# ============================================================
# 先决条件:
#   1. 域名 qscl.link 的 DNS A 记录已指向服务器公网 IP
#   2. Docker 已安装
#   3. 服务器 80 和 443 端口已开放
# ============================================================
set -euo pipefail

DOMAIN="qscl.link"
EMAIL="admin@qscl.link"           # 可改为你的邮箱
SSL_DIR="$(dirname "$0")/../nginx/ssl"
CERTBOT_DIR="$(dirname "$0")/../nginx/certbot"

echo "=========================================="
echo " QuantumClaw SSL Setup"
echo " Domain: $DOMAIN"
echo "=========================================="

# 1. 创建目录
mkdir -p "$SSL_DIR" "$CERTBOT_DIR/www"

# 2. 临时启动 nginx（仅用于 Let's Encrypt 验证）
echo ""
echo "[1/4] 启动 nginx 用于域名验证..."
docker compose up -d nginx 2>/dev/null || true
sleep 3

# 3. 获取 SSL 证书
echo ""
echo "[2/4] 向 Let's Encrypt 申请证书..."
docker run --rm \
  -v "$(realpath "$CERTBOT_DIR/www")":/var/www/certbot \
  -v "$(realpath "$SSL_DIR")":/etc/letsencrypt \
  certbot/certbot certonly --webroot \
  -w /var/www/certbot \
  -d "$DOMAIN" \
  -d "www.$DOMAIN" \
  --email "$EMAIL" \
  --agree-tos \
  --non-interactive

# 4. 复制证书到 nginx ssl 目录
echo ""
echo "[3/4] 复制证书到 nginx..."
cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" "$SSL_DIR/" 2>/dev/null || \
  cp "$SSL_DIR/live/$DOMAIN/fullchain.pem" "$SSL_DIR/" 2>/dev/null || true
cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" "$SSL_DIR/" 2>/dev/null || \
  cp "$SSL_DIR/live/$DOMAIN/privkey.pem" "$SSL_DIR/" 2>/dev/null || true
cp "/etc/letsencrypt/live/$DOMAIN/chain.pem" "$SSL_DIR/" 2>/dev/null || \
  cp "$SSL_DIR/live/$DOMAIN/chain.pem" "$SSL_DIR/" 2>/dev/null || true

# 5. 重启完整栈
echo ""
echo "[4/4] 重启全部服务..."
docker compose up -d

echo ""
echo "=========================================="
echo " ✅ SSL 配置完成！"
echo " https://$DOMAIN"
echo "=========================================="
echo ""
echo "证书自动续期（每月执行）:"
echo "  docker run --rm -v \$(pwd)/nginx/ssl:/etc/letsencrypt -v \$(pwd)/nginx/certbot/www:/var/www/certbot certbot/certbot renew"
