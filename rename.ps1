# QuantumClaw 项目重命名脚本
# 将 github.com/songquanpeng/one-api 替换为 github.com/quantumclaw/quantumclaw

$oldNamespace = "github.com/songquanpeng/one-api"
$newNamespace = "github.com/quantumclaw/quantumclaw"

Get-ChildItem -Path "." -Recurse -Filter "*.go" | ForEach-Object {
    $filePath = $_.FullName
    Write-Output "Processing: $filePath"
    
    # 读取文件内容
    $content = [System.IO.File]::ReadAllText($filePath, [System.Text.Encoding]::UTF8)
    
    # 替换命名空间
    $newContent = $content -replace [regex]::Escape($oldNamespace), $newNamespace
    
    # 如果有变化，写回文件
    if ($content -ne $newContent) {
        [System.IO.File]::WriteAllText($filePath, $newContent, [System.Text.Encoding]::UTF8)
        Write-Output "  ✅ Updated"
    }
}

# 替换 README.md
if (Test-Path "README.md") {
    $content = [System.IO.File]::ReadAllText("README.md", [System.Text.Encoding]::UTF8)
    $newContent = $content -replace [regex]::Escape($oldNamespace), $newNamespace
    if ($content -ne $newContent) {
        [System.IO.File]::WriteAllText("README.md", $newContent, [System.Text.Encoding]::UTF8)
        Write-Output "✅ Updated README.md"
    }
}

Write-Output "`n✅ 重命名完成！"
