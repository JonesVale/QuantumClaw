#!/bin/bash

# QuantumClaw Open Source 上传脚本
# 版权所有 (C) 2026 深圳市中科劲纬智能有限公司
# 官网: https://www.ctji.cn
# 邮箱: vale@ctji.cn

set -e

echo "=========================================="
echo "QuantumClaw Open Source 上传脚本"
echo "=========================================="
echo ""

# 检查是否已设置 Token
if [ -z "$GITHUB_TOKEN" ]; then
    echo "⚠️  请先设置 GITHUB_TOKEN 环境变量"
    echo "   export GITHUB_TOKEN=your_token_here"
    echo ""
    echo "⚠️  警告：请勿在脚本中硬编码 Token！"
    exit 1
fi

# 仓库信息
REPO_NAME="QuantumClaw-Open"
OWNER="ctji"  # 请修改为您的 GitHub 用户名
REMOTE_URL="https://github.com/${OWNER}/${REPO_NAME}.git"

# 进入开源目录
cd QuantumClaw-Open

echo "📁 当前目录: $(pwd)"
echo ""

# 1. 安全检查
echo "🔍 执行安全检查..."
python security_check.py .
if [ $? -ne 0 ]; then
    echo "❌ 安全检查失败，请修复问题后再上传"
    exit 1
fi
echo "✅ 安全检查通过"
echo ""

# 2. 检查开源清单
echo "📋 检查开源清单..."
if [ ! -f "open_source_checklist.md" ]; then
    echo "❌ 缺少开源检查清单"
    exit 1
fi
echo "✅ 开源清单存在"
echo ""

# 3. 初始化 Git
echo "🔧 初始化 Git 仓库..."
if [ ! -d ".git" ]; then
    git init
    git config user.email "opensource@ctji.cn"
    git config user.name "QuantumClaw Open Source"
fi
echo "✅ Git 初始化完成"
echo ""

# 4. 添加所有开源文件
echo "📤 添加文件..."
git add .
echo "✅ 文件添加完成"
echo ""

# 5. 提交更改
echo "💾 提交更改..."
git commit -m "Initial open source release (AGPL-3.0)

Copyright (C) 2026 深圳市中科劲纬智能有限公司
Website: https://www.ctji.cn
Email: vale@ctji.cn

This release includes open-source modules only.
Payment, financial, and security modules are closed-source."

echo "✅ 提交完成"
echo ""

# 6. 创建远程仓库
echo "🌐 创建远程仓库..."
if ! git remote get-url origin > /dev/null 2>&1; then
    git remote add origin "$REMOTE_URL"
fi
echo "✅ 远程仓库配置完成"
echo ""

# 7. 推送代码
echo "🚀 推送到 GitHub..."
git push -u origin main

echo ""
echo "=========================================="
echo "✅ 上传完成！"
echo "=========================================="
echo ""
echo "开源仓库: $REMOTE_URL"
echo ""
echo "重要提示:"
echo "1. 本版本仅包含开源模块（约60%代码）"
echo "2. 支付、金融、安全模块已移除"
echo "3. 商业用途需要商业授权"
echo "4. 联系: vale@ctji.cn"
echo ""
echo "版权所有 (C) 2026 深圳市中科劲纬智能有限公司"
