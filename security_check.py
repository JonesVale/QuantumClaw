"""
开源安全检查脚本
用于检查代码中是否包含敏感信息
"""

import os
import re
import sys
from pathlib import Path

# 敏感信息模式
SENSITIVE_PATTERNS = {
    'API密钥': r'(?:api[_-]?key|apikey|access[_-]?token|secret[_-]?key)\s*[:=]\s*["\']?[A-Za-z0-9+/]{20,}',
    '数据库密码': r'(?:db[_-]?password|database[_-]?password|mysql[_-]?password)\s*[:=]\s*["\']?[^\s"\']{6,}',
    'JWT密钥': r'(?:jwt[_-]?secret|jwt_secret)\s*[:=]\s*["\']?[A-Za-z0-9+/]{20,}',
    '私钥': r'(?:private[_-]?key|priv[_-]?key)\s*[:=]\s*["\']?[A-Za-z0-9+/=\s]{100,}',
    'Stripe密钥': r'(?:sk_live_|sk_test_|pk_live_|pk_test_)[A-Za-z0-9]{20,}',
    '支付宝密钥': r'(?:alipay[_-]?(?:private|public)_key)\s*[:=]\s*["\']?[A-Za-z0-9+/=\s]{100,}',
    'Base64编码密钥': r'-----BEGIN (?:RSA |EC )?PRIVATE KEY-----',
    '硬编码密码': r'(?:password|passwd|pwd)\s*[:=]\s*["\'][^"\']{6,}["\']',
}

# 禁止提交的文件模式
FORBIDDEN_FILES = [
    r'\.env$',
    r'\.env\..*\.local$',
    r'.*private.*\.key$',
    r'.*certificate.*\.pem$',
    r'data/.*\.db$',
    r'data/.*\.sqlite$',
    r'backup/.*',
]

def check_file_content(filepath):
    """检查文件内容中的敏感信息"""
    issues = []
    
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
            lines = content.split('\n')
            
        for pattern_name, pattern in SENSITIVE_PATTERNS.items():
            matches = re.finditer(pattern, content, re.IGNORECASE)
            for match in matches:
                # 获取行号
                line_num = content[:match.start()].count('\n') + 1
                # 获取行内容（脱敏）
                line = lines[line_num - 1]
                # 脱敏处理
                sensitive_parts = re.findall(r'[A-Za-z0-9+/]{10,}', line)
                if sensitive_parts:
                    line = re.sub(r'[A-Za-z0-9+/]{10,}', '***REDACTED***', line)
                issues.append({
                    'pattern': pattern_name,
                    'line': line_num,
                    'content': line.strip()[:100]
                })
    except Exception as e:
        print(f"  Warning: Could not read {filepath}: {e}")
    
    return issues

def check_filename(filepath):
    """检查文件名是否敏感"""
    for pattern in FORBIDDEN_FILES:
        if re.match(pattern, filepath):
            return True
    return False

def scan_directory(directory):
    """扫描目录"""
    all_issues = {}
    forbidden_files = []
    
    for root, dirs, files in os.walk(directory):
        # 跳过某些目录
        dirs[:] = [d for d in dirs if d not in ['.git', 'node_modules', 'vendor', '__pycache__', '.venv']]
        
        for file in files:
            filepath = os.path.join(root, file)
            rel_path = os.path.relpath(filepath, directory)
            
            # 检查文件名
            if check_filename(filepath):
                forbidden_files.append(rel_path)
                continue
            
            # 检查文件内容
            if file.endswith(('.go', '.py', '.js', '.ts', '.java', '.c', '.cpp', '.h', '.env', '.yaml', '.yml', '.json', '.xml', '.conf', '.ini', '.cfg')):
                issues = check_file_content(filepath)
                if issues:
                    all_issues[rel_path] = issues
    
    return all_issues, forbidden_files

def generate_report(issues, forbidden_files, output_file):
    """生成报告"""
    report = []
    report.append("=" * 80)
    report.append("开源安全检查报告")
    report.append("=" * 80)
    report.append(f"生成时间: {__import__('datetime').datetime.now()}")
    report.append("")
    
    # 禁止文件
    if forbidden_files:
        report.append("❌ 禁止提交的文件:")
        for f in forbidden_files:
            report.append(f"  - {f}")
        report.append("")
    
    # 敏感信息
    if issues:
        report.append("❌ 发现敏感信息:")
        for filepath, file_issues in issues.items():
            report.append(f"\n文件: {filepath}")
            for issue in file_issues:
                report.append(f"  行 {issue['line']}: [{issue['pattern']}]")
                report.append(f"    {issue['content']}")
        report.append("")
    
    # 总结
    total_issues = sum(len(v) for v in issues.values())
    report.append("=" * 80)
    report.append(f"总结: {total_issues} 个敏感信息问题, {len(forbidden_files)} 个禁止文件")
    report.append("=" * 80)
    
    if total_issues == 0 and len(forbidden_files) == 0:
        report.append("✅ 安全检查通过！")
    else:
        report.append("⚠️ 请修复上述问题后再提交开源")
    
    # 保存报告
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write('\n'.join(report))
    
    return total_issues, len(forbidden_files)

def main():
    if len(sys.argv) < 2:
        print("用法: python security_check.py <directory>")
        sys.exit(1)
    
    directory = sys.argv[1]
    output_file = os.path.join(directory, 'security_check_report.txt')
    
    if not os.path.isdir(directory):
        print(f"错误: {directory} 不是有效目录")
        sys.exit(1)
    
    print(f"正在扫描目录: {directory}")
    issues, forbidden_files = scan_directory(directory)
    total_issues, total_forbidden = generate_report(issues, forbidden_files, output_file)
    
    print(f"\n检查完成!")
    print(f"敏感信息问题: {total_issues}")
    print(f"禁止文件: {total_forbidden}")
    print(f"\n详细报告已保存到: {output_file}")
    
    if total_issues > 0 or total_forbidden > 0:
        sys.exit(1)
    else:
        sys.exit(0)

if __name__ == '__main__':
    main()
