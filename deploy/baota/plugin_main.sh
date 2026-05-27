#!/bin/bash
# ===============================================
# QuantumClaw Plugin — 宝塔面板应用商店部署文件
# ===============================================
# 路径: /www/server/panel/plugin/quantumclaw/
# ===============================================

PLUGIN_DIR="/www/server/panel/plugin/quantumclaw"
QUANTUM_DIR="/www/quantumclaw"

# 创建插件目录
install_plugin() {
    mkdir -p "$PLUGIN_DIR"
    
    # 主配置文件
    cat > "$PLUGIN_DIR/config.json" <<'EOF'
{
    "name": "quantumclaw",
    "version": "1.0.0",
    "title": "QuantumClaw AI API 网关",
    "description": "聚合 30+ AI 模型渠道，支持多级定价、用户管理、API Key 管理、智能路由",
    "author": "QuantumClaw Team",
    "home": "/quantumclaw",
    "icon": "/quantumclaw/logo.png"
}
EOF

    # 入口文件
    cat > "$PLUGIN_DIR/index.html" <<'HTML'
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>QuantumClaw</title></head>
<body>
<div id="app">
    <h2>QuantumClaw AI API Gateway</h2>
    <p>版本: 1.0.0</p>
    <div class="info-grid">
        <div class="info-item"><label>状态</label><span id="status">检查中...</span></div>
        <div class="info-item"><label>端口</label><span>3666</span></div>
        <div class="info-item"><label>数据目录</label><span>/www/quantumclaw/data</span></div>
    </div>
    <div class="actions">
        <button onclick="action('start')">启动</button>
        <button onclick="action('stop')">停止</button>
        <button onclick="action('restart')">重启</button>
    </div>
    <script>
    function action(cmd) {
        fetch('/quantumclaw/api?action='+cmd).then(r=>r.json()).then(d=>alert(d.msg));
    }
    </script>
</div>
</body>
</html>
HTML

    echo "插件安装完成"
}

# 解析宝塔面板请求
case "$1" in
    install)    install_plugin ;;
    uninstall)  rm -rf "$PLUGIN_DIR"; echo "插件已删除" ;;
    *)
        action="${QUERY_STRING#*=}"
        case "$action" in
            start)   systemctl start quantumclaw; echo '{"status":0,"msg":"已启动"}' ;;
            stop)    systemctl stop quantumclaw; echo '{"status":0,"msg":"已停止"}' ;;
            restart) systemctl restart quantumclaw; echo '{"status":0,"msg":"已重启"}' ;;
            *)       echo '{"status":1,"msg":"未知操作"}' ;;
        esac
    ;;
esac
