# QuantumClaw 部署到騰訊雲操作手冊

## 前置條件

1. 騰訊雲服務器 122.51.221.43 (已付款)
2. 已安裝 Docker + Docker Compose
3. 域名的 DNS 已指向此服務器（可選）

## 一、SSH 連接服務器

```bash
ssh ubuntu@122.51.221.43
# 密碼：請確認您提供的密碼
```

或通過騰訊雲控制台 → 實例 → 登錄 → WebShell (VNC)

## 二、一鍵部署腳本

將以下內容保存為 `deploy.sh` 並執行：

```bash
#!/bin/bash
set -e

echo "=== QuantumClaw 部署開始 ==="

# 1. 安裝 Docker（如未安裝）
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com | bash
    sudo usermod -aG docker ubuntu
fi

# 2. 安裝 Docker Compose（如未安裝）
if ! command -v docker-compose &> /dev/null; then
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

# 3. 克隆或拉取項目
PROJECT_DIR=~/quantumclaw
if [ -d "$PROJECT_DIR" ]; then
    cd $PROJECT_DIR && git pull
else
    git clone <你的倉庫地址> $PROJECT_DIR
    cd $PROJECT_DIR
fi

# 4. 創建環境變量文件
cat > .env << 'ENVEOF'
# ===== 基礎配置 =====
PORT=3666
GIN_MODE=release
SESSION_SECRET=$(openssl rand -hex 32)

# ===== SQLite 數據庫 =====
SQLITE_PATH=/app/data/quantumclaw.db
SQLITE_BUSY_TIMEOUT=3000

# ===== 管理員賬號 =====
INITIAL_ROOT_USERNAME=root
INITIAL_ROOT_PASSWORD=admin123456
INITIAL_ROOT_ACCESS_TOKEN=quantumclaw_token_change_me

# ===== 安全 =====
CRYPTO_SECRET=$(openssl rand -hex 32)
EMERGENCY_RESET_TOKEN=$(openssl rand -hex 16)

# ===== OAuth（可選） =====
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
LARK_CLIENT_ID=
LARK_CLIENT_SECRET=

# ===== 郵件 SMTP（可選） =====
SMTP_SERVER=
SMTP_PORT=587
SMTP_ACCOUNT=
SMTP_FROM=
SMTP_TOKEN=

# ===== 支付（可選） =====
STRIPE_ENABLED=false
EPAY_ENABLED=false
CREEM_ENABLED=false
WAFFO_ENABLED=false

# ===== 安全頭 =====
CSP_ENABLED=true
HSTS_ENABLED=true
ALLOWED_ORIGINS=https://你的域名.com
ENVEOF

# 5. 構建和啟動
docker compose -f docker-compose.sqlite.yml up -d --build

# 6. 等待服務啟動
echo "等待服務啟動..."
sleep 5

# 7. 檢查服務狀態
curl -s http://localhost:3666/api/status | python3 -m json.tool || echo "服務尚在啟動中"

echo "=== 部署完成 ==="
echo "訪問: http://$(curl -s ifconfig.me):3666"
echo "管理員: root / admin123456"
```

## 三、Docker 構建（本地）

如果無法從 GitHub 直接部署，可以從本地構建鏡像後上傳：

```bash
# 在本地項目目錄
docker build -t quantumclaw:latest .
docker save quantumclaw:latest | gzip > quantumclaw.tar.gz
# 上傳到服務器
scp quantumclaw.tar.gz ubuntu@122.51.221.43:~/
# 在服務器載入
docker load < quantumclaw.tar.gz
docker compose -f docker-compose.sqlite.yml up -d
```

## 四、部署後驗證

```bash
# 1. API 狀態檢查
curl http://localhost:3666/api/status
# 預期: {"success":true,"message":"ok","data":{...}}

# 2. 前端頁面
curl -s -o /dev/null -w "%{http_code}" http://localhost:3666/
# 預期: 200

# 3. 管理員登錄
curl -X POST http://localhost:3666/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"root","password":"admin123456"}'
# 預期: { "success": true, "data": { "token": "..." } }

# 4. Docker 容器狀態
docker ps --filter name=quantumclaw

# 5. 日誌檢查
docker logs quantumclaw --tail 50
```

## 五、Nginx 反向代理（如需域名訪問）

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    location / {
        proxy_pass http://127.0.0.1:3666;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }
}
```

## 六、日誌和監控

```bash
# 查看實時日誌
docker logs -f quantumclaw

# 檢查數據庫文件
ls -lh /var/lib/docker/volumes/quantumclaw_quantumclaw_data/

# 備份數據庫
docker run --rm -v quantumclaw_quantumclaw_data:/data -v $(pwd):/backup alpine tar czf /backup/quantumclaw_backup_$(date +%Y%m%d).tar.gz -C /data .
```
