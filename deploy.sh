#!/bin/bash
set -e

echo "===== QuantumClaw 部署脚本 ====="
cd /root

# 1. 解压
tar xzf quantumclaw_deploy.tar.gz -C /opt/quantumclaw 2>/dev/null || {
    mkdir -p /opt/quantumclaw
    tar xzf quantumclaw_deploy.tar.gz -C /opt/quantumclaw
}
cd /opt/quantumclaw

echo "1/4 代码已解压到 /opt/quantumclaw"

# 2. 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "安装 Docker..."
    curl -fsSL https://get.docker.com | bash
    systemctl enable docker
    systemctl start docker
fi

echo "2/4 Docker: $(docker --version)"

# 3. 启动
export GOPROXY=https://goproxy.cn,direct
docker compose -f docker-compose.sqlite.yml down 2>/dev/null || true
docker compose -f docker-compose.sqlite.yml up -d --build

echo "3/4 构建启动中..."

# 4. 等待健康
for i in $(seq 1 20); do
    sleep 3
    if curl -s http://localhost:3666/api/status > /dev/null 2>&1; then
        echo "4/4 API 就绪!"
        break
    fi
    echo "等待中 ($i)..."
done

# 5. 测试
echo ""
echo "===== 验证 ====="
curl -s http://localhost:3666/api/status
echo ""
curl -s -X POST http://localhost:3666/api/password/emergency-reset \
  -H "Content-Type: application/json" \
  -d '{"reset_token":"qc-reset-3h7mpBUPn6kQaG9H","new_password":"admin123"}'
echo ""
curl -s -X POST http://localhost:3666/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"root","password":"admin123"}'
echo ""
echo "===== 部署完成 ====="
echo "地址: http://139.196.8.90:3666"
echo "账号: root / admin123"
