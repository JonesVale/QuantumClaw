# ============================================================
# QuantumClaw Docker 快速启动脚本（SQLite 轻量模式）
# Windows PowerShell
# ============================================================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  QuantumClaw SQLite 轻量模式启动脚本" -ForegroundColor Cyan
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
    Write-Host ""
    Write-Host "注意：使用 SQLite 模式时，请确保 .env 中 SQL_DSN 为空或已注释掉。"
    Write-Host ""
    $edit = Read-Host "是否现在编辑 .env 文件？(Y/N)"
    if ($edit -eq "Y" -or $edit -eq "y") {
        notepad .env
    }
    Write-Host ""
    Write-Host "请配置完成后重新运行此脚本。" -ForegroundColor Yellow
    exit 0
}

# 检查 Docker 是否运行
try {
    $null = docker info 2>&1
    Write-Host "[成功] Docker 正在运行" -ForegroundColor Green
} catch {
    Write-Host "[错误] Docker 未运行，请启动 Docker Desktop！" -ForegroundColor Red
    exit 1
}

# 构建并启动容器（使用 SQLite docker-compose 文件）
Write-Host ""
Write-Host "[信息] 使用 docker-compose.sqlite.yml 构建并启动..." -ForegroundColor Cyan
Write-Host "（无需 MySQL 和 Redis，单容器轻量运行）" -ForegroundColor Yellow
Write-Host ""

docker compose -f docker-compose.sqlite.yml up -d --build

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  QuantumClaw SQLite 模式启动成功！" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "访问地址："
    Write-Host "  - 主页: http://localhost:3666"
    Write-Host "  - API: http://localhost:3666/api"
    Write-Host ""
    Write-Host "查看日志："
    Write-Host "  - docker compose -f docker-compose.sqlite.yml logs -f"
    Write-Host ""
    Write-Host "停止服务："
    Write-Host "  - docker compose -f docker-compose.sqlite.yml down"
    Write-Host ""
} else {
    Write-Host "[错误] Docker 启动失败，请检查配置！" -ForegroundColor Red
    exit 1
}
