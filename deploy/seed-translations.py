#!/usr/bin/env python3
"""
QuantumClaw Translation DB Seeder
读取所有 i18n JSON 文件 → 生成 SQL 插入语句 → 输出到文件

用法:
  python3 deploy/seed-translations.py           # 生成 seed.sql
  python3 deploy/seed-translations.py --apply   # 直接通过 API 写入
"""

import json
import os
import sys
import urllib.request
import urllib.parse
import time

I18N_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 
                       'web/default/src/i18n')

# 语言代码映射（文件名 → 标准代码）
LANG_CODE_MAP = {
    'zh-CN.json': 'zh-CN',
    'zh-TW.json': 'zh-TW',
    'en.json':    'en',
    'ja.json':    'ja',
    'ko.json':    'ko',
    'ru.json':    'ru',
    'vi.json':    'vi',
    'fr.json':    'fr',
    'es.json':    'es',
    'de.json':    'de',
    'it.json':    'it',
    'pt.json':    'pt',
    'nl.json':    'nl',
    'ar.json':    'ar',
}


def generate_sql():
    """读取所有 JSON 文件，生成 SQL 插入语句"""
    sql_lines = [
        "-- ============================================================",
        "-- QuantumClaw Translation Data Seed",
        f"-- Generated: {time.strftime('%Y-%m-%d %H:%M:%S')}",
        "-- Source: 14 language JSON files",
        "-- ============================================================",
        "",
        "START TRANSACTION;",
        "",
        "-- Clear existing translations (safe for first-time seed)",
        "DELETE FROM translations;",
        "",
    ]
    
    total = 0
    for filename, lang_code in sorted(LANG_CODE_MAP.items()):
        filepath = os.path.join(I18N_DIR, filename)
        with open(filepath, 'r', encoding='utf-8-sig') as f:
            data = json.load(f)
        
        batch_count = 0
        for key, value in data.items():
            # Escape single quotes for SQL
            escaped_key = key.replace("'", "''")
            escaped_value = value.replace("'", "''")
            sql_lines.append(
                f"INSERT INTO translations (lang_key, lang_code, value, updated_at) "
                f"VALUES ('{escaped_key}', '{lang_code}', '{escaped_value}', NOW());"
            )
            batch_count += 1
            total += 1
        
        print(f"  ✅ {filename:<15} ({lang_code:<5}) → {batch_count:>5} keys")
    
    sql_lines.append("")
    sql_lines.append("COMMIT;")
    sql_lines.append("")
    sql_lines.append(f"-- Total: {total} translations across {len(LANG_CODE_MAP)} languages")
    
    return '\n'.join(sql_lines), total


def apply_via_api(sql_path, api_base='http://localhost:3666'):
    """通过 API 批量导入翻译（需要管理员 Token）"""
    print("\n📡 通过 API 导入...")
    
    # Read token from env or stdin
    token = os.environ.get('ADMIN_TOKEN', '')
    if not token:
        print("  ⚠️  请设置 ADMIN_TOKEN 环境变量")
        return False
    
    for filename, lang_code in sorted(LANG_CODE_MAP.items()):
        filepath = os.path.join(I18N_DIR, filename)
        with open(filepath, 'r', encoding='utf-8-sig') as f:
            data = json.load(f)
        
        batch = []
        for key, value in data.items():
            batch.append({'lang_key': key, 'lang_code': lang_code, 'value': value})
        
        # Split into chunks of 500
        chunk_size = 500
        for i in range(0, len(batch), chunk_size):
            chunk = batch[i:i+chunk_size]
            payload = json.dumps({'translations': chunk}).encode('utf-8')
            
            req = urllib.request.Request(
                f'{api_base}/api/admin/translations/batch',
                data=payload,
                headers={
                    'Content-Type': 'application/json',
                    'Authorization': f'Bearer {token}',
                },
                method='POST',
            )
            
            try:
                with urllib.request.urlopen(req, timeout=30) as resp:
                    result = json.loads(resp.read().decode('utf-8'))
                if result.get('success'):
                    print(f"  ✅ {filename} chunk {i//chunk_size+1}: {len(chunk)} keys")
                else:
                    print(f"  ⚠️  {filename} chunk {i//chunk_size+1}: {result.get('message')}")
            except Exception as e:
                print(f"  ❌ {filename} chunk {i//chunk_size+1}: {e}")
            
            time.sleep(0.2)  # Rate limit
    
    print("  ✅ API 导入完成")
    return True


def main():
    print("="*60)
    print("QuantumClaw Translation DB Seeder")
    print("="*60)
    print(f"\n📂 Reading from: {I18N_DIR}")
    print()
    
    sql_content, total = generate_sql()
    
    # Output SQL file
    output_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 
                             'deploy')
    sql_path = os.path.join(output_dir, 'seed-translations.sql')
    with open(sql_path, 'w', encoding='utf-8') as f:
        f.write(sql_content)
    
    print(f"\n{'='*60}")
    print(f"📄 SQL 文件: {sql_path}")
    print(f"📊 总计: {total} 条翻译")
    print(f"🌐 语言: {len(LANG_CODE_MAP)} 种")
    print(f"{'='*60}")
    print(f"\n部署方式:")
    print(f"  1. SQL 方式: mysql -u root -p quantumclaw < {sql_path}")
    print(f"  2. API 方式: ADMIN_TOKEN=xxx python3 {sys.argv[0]} --apply")
    
    # Apply if --apply flag
    if '--apply' in sys.argv:
        apply_via_api(sql_path)


if __name__ == '__main__':
    main()
