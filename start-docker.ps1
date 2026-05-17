# ============================================================
# QuantumClaw Docker 快速启动脚本 (Windows PowerShell)
# ============================================================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  QuantumClaw Docker 快速启动脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 .env 文件是否存在
if (-not (Test-Path ".env")) {
    Write-Host "[信息] .env 文件不存在，从 .env.example 创建..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "[成功] 已创建 .env 文件，请编辑它并配置必要的参数！" -ForegroundColor Green
    Write-Host ""
    Write-Host "必填项："
    Write-Host "  - SESSION_SECRET (会话密钥，请修改为随机字符串)"
    Write-Host "  - INITIAL_ROOT_TOKEN (初始 root 令牌)"
    Write-Host "  - INITIAL_ROOT_ACCESS_TOKEN (初始 root 访问令牌)"
    Write-Host "  - MYSQL_ROOT_PASSWORD (MySQL root 密码)"
    Write-Host ""
    Write-Host "可选项（根据需求配置）："
    Write-Host "  - GITHUB_OAUTH_ENABLED, GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET"
    Write-Host "  - WECHAT_AUTH_ENABLED, WECHAT_CLIENT_ID, WECHAT_CLIENT_SECRET"
    Write-Host "  - DISCORD_OAUTH_ENABLED, DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET"
    Write-Host "  - WEBAUTHN_ENABLED (Passkey/无密码登录)"
    Write-Host ""
    $edit = Read-Host "是否现在编辑 .env 文件？(Y/N)"
    if ($edit -eq "Y" -or $edit -eq "y") {
        notepad .env
    }
    Write-Host ""
    Write-Host "请配置完成后重新运行此脚本。" -ForegroundColor Yellow
    exit 0
}

# 加载 .env 文件
Write-Host "[信息] 加载 .env 配置..." -ForegroundColor Cyan
Get-Content ".env" | ForEach-Object {
    if ($_ -match "^\s*([^#][^=]+)=(.*)$") {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()
        [Environment]::SetEnvironmentVariable($key, $value, "Process")
    }
}

# 检查 Docker 是否运行
try {
    $null = docker info 2>&1
    Write-Host "[成功] Docker 正在运行" -ForegroundColor Green
} catch {
    Write-Host "[错误] Docker 未运行，请启动 Docker Desktop！" -ForegroundColor Red
    exit 1
}

# 构建并启动容器
Write-Host ""
Write-Host "[信息] 构建并启动 Docker 容器..." -ForegroundColor Cyan
Write-Host "（首次运行需要下载镜像，可能需要几分钟）" -ForegroundColor Yellow
Write-Host ""

docker compose up -d --build

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  QuantumClaw 启动成功！" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "访问地址："
    $port = if ($env:PORT) { $env:PORT } else { "3000" }
    Write-Host "  - 主页: http://localhost:$port"
    Write-Host "  - API: http://localhost:$port/api"
    Write-Host ""
    Write-Host "查看日志："
    Write-Host "  - docker compose logs -f"
    Write-Host ""
    Write-Host "停止服务："
    Write-Host "  - docker compose down"
    Write-Host ""
} else {
    Write-Host "[错误] Docker 启动失败，请检查配置！" -ForegroundColor Red
    exit 1
}
