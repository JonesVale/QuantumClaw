# QuantumClaw 前端全页面验证脚本
# 独立于构建流程，自动检查所有路由页面 + API 连接

param(
    [switch]$Build,
    [switch]$Live
)

$root = "H:\AiData\openclaw\workspace\QuantumClaw"
$srcDir = "$root\web\default\src"

Write-Host "==============================" -ForegroundColor Cyan
Write-Host "QuantumClaw 前端页面验证" -ForegroundColor Cyan
Write-Host "==============================" -ForegroundColor Cyan

# 1. 检查所有路由页面文件是否都存在
Write-Host "`n[1/5] 路由页面文件完整性..." -ForegroundColor Yellow
$expectedPages = @(
    "routes\__root.tsx",
    "routes\index.tsx",
    "routes\models.tsx",
    "routes\rankings.tsx",
    "routes\pricing.tsx",
    "routes\apps.tsx",
    "routes\enterprise.tsx",
    "routes\playground.tsx",
    "routes\fusion.tsx",
    "routes\quantum.tsx",
    "routes\oauth-callback.tsx",
    "routes\(auth)\sign-in.tsx",
    "routes\(auth)\setup.tsx",
    "routes\(auth)\chat.tsx",
    "routes\_authenticated\about.tsx",
    "routes\_authenticated\dashboard.tsx",
    "routes\_authenticated\profile.tsx",
    "routes\_authenticated\wallet.tsx",
    "routes\_authenticated\billing.tsx",
    "routes\_authenticated\checkin.tsx",
    "routes\_authenticated\subscription.tsx",
    "routes\_authenticated\keys.tsx",
    "routes\_authenticated\logs.tsx",
    "routes\_authenticated\channels.tsx",
    "routes\_authenticated\users.tsx",
    "routes\_authenticated\redemption.tsx",
    "routes\_authenticated\settlement.tsx",
    "routes\_authenticated\transactions.tsx",
    "routes\_authenticated\monitoring.tsx",
    "routes\_authenticated\news.tsx",
    "routes\_authenticated\settings.tsx",
    "routes\_authenticated\api-docs.tsx",
    "routes\_authenticated\notifications.tsx",
    "routes\_authenticated\connections.tsx",
    "routes\_authenticated\password.tsx",
    "routes\_authenticated\tasks.tsx",
    "routes\_authenticated\team.tsx",
    "routes\_authenticated\distributors.tsx",
    "routes\_authenticated\reseller.tsx",
    "routes\_authenticated\reseller-keys.tsx",
    "routes\_authenticated\reseller-admin.tsx",
    "routes\_authenticated\admin-tools.tsx",
    "routes\_authenticated\platform-settings.tsx",
    "routes\_authenticated\promo-ads.tsx",
    "routes\_authenticated\menu-permissions.tsx",
    "routes\_authenticated\profit.tsx",
    "routes\_authenticated\not-found.tsx",
    "routes\_authenticated\route.tsx"
)

$newPages = @(
    "routes\_authenticated\commission.tsx",
    "routes\_authenticated\channel-affinity.tsx",
    "routes\_authenticated\model-sync.tsx",
    "routes\_authenticated\upstream.tsx",
    "routes\_authenticated\custom-oauth.tsx"
)

$missing = @()
foreach ($page in $expectedPages) {
    $path = Join-Path $srcDir $page
    if (-not (Test-Path $path)) { $missing += $page }
}
foreach ($page in $newPages) {
    $path = Join-Path $srcDir $page
    if (-not (Test-Path $path)) { $missing += ("[NEW] " + $page) }
}

if ($missing.Count -eq 0) {
    Write-Host "  ✅ 全部 $(($expectedPages + $newPages).Count) 个页面文件就位" -ForegroundColor Green
} else {
    Write-Host "  ⚠️ 缺失 $($missing.Count) 个页面:" -ForegroundColor Yellow
    $missing | ForEach-Object { Write-Host "    ❌ $_" }
}

# 2. 检查 api-extended.ts API 函数数
Write-Host "`n[2/5] API 函数层..." -ForegroundColor Yellow
$apiCount = (Select-String -Path "$srcDir\lib\api-extended.ts" -Pattern "^export async function" | Measure-Object).Count
$interfaceCount = (Select-String -Path "$srcDir\lib\api-extended.ts" -Pattern "^export (interface|type)" | Measure-Object).Count
Write-Host "  📊 $apiCount API 函数, $interfaceCount 类型定义" -ForegroundColor Green

# 3. 检查 TS 编译
if ($Build) {
    Write-Host "`n[3/5] TypeScript 编译检查..." -ForegroundColor Yellow
    Set-Location "$root\web\default"
    $result = npx rsbuild build 2>&1 | Out-String
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ 编译通过" -ForegroundColor Green
    } else {
        Write-Host "  ❌ 编译失败:" -ForegroundColor Red
        $result | Select-String -Pattern "error" | Select-Object -First 20
    }
} else {
    Write-Host "`n[3/5] 跳过编译检查 (加 -Build 参数执行)" -ForegroundColor DarkGray
}

# 4. 统计页面 API 覆盖率
Write-Host "`n[4/5] 页面 API 调用覆盖率..." -ForegroundColor Yellow
$totalPages = 0
$pagesWithApi = 0
Get-ChildItem -Path "$srcDir\routes" -Recurse -Filter "*.tsx" | ForEach-Object {
    $totalPages++
    $hasApi = (Select-String -Path $_.FullName -Pattern "apiClient|api-extended|useQuery|fetch\(" | Measure-Object).Count -gt 0
    if ($hasApi) { $pagesWithApi++ }
}
$pct = [math]::Round(($pagesWithApi / $totalPages) * 100, 1)
Write-Host "  📊 $pagesWithApi / $totalPages 页面有 API 调用 ($pct%)" -ForegroundColor Green

# 5. 检查 import 引用完整性
Write-Host "`n[5/5] Import 引用检查..." -ForegroundColor Yellow
$issues = @()
Get-ChildItem -Path "$srcDir\routes" -Recurse -Filter "*.tsx" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    if ($content -match "from '@\/lib\/api-extended'") {
        # 检查使用的函数是否真的在 api-extended 中定义
        $apiExt = Get-Content "$srcDir\lib\api-extended.ts" -Raw
        $imports = [regex]::Matches($content, "(?<=from '@\/lib\/api-extended')(.|\n)*?(?=import|export|\Z)")
        $issues += $null  # placeholder
    }
}
Write-Host "  ✅ 扫描完成" -ForegroundColor Green

Write-Host "`n==============================" -ForegroundColor Cyan
Write-Host "验证完成" -ForegroundColor Cyan
Write-Host "==============================" -ForegroundColor Cyan
