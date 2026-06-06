#!/usr/bin/env python3
"""
QuantumClaw i18n Translation Sync Tool
使用百度翻译 API 自动补全所有语言文件中缺失的翻译键

用法:
  python3 deploy/i18n-sync.py          # 补全所有语言
  python3 deploy/i18n-sync.py --check  # 只检查不翻译

环境变量:
  BAIDU_TRANSLATE_APPID, BAIDU_TRANSLATE_KEY
"""

import json
import hashlib
import random
import time
import os
import sys
import urllib.request
import urllib.parse

# ── 百度翻译配置 ──
BAIDU_APPID = "20260526002620460"
BAIDU_KEY = "btfJiaqZvwsTVbnFpyfi"
BAIDU_API = "https://fanyi-api.baidu.com/api/trans/vip/translate"

# ── 语言映射（zh-CN → 百度语言代码） ──
LANG_MAP = {
    'en.json':    'en',
    'zh-TW.json': 'cht',
    'ja.json':    'jp',
    'ko.json':    'kor',
    'ru.json':    'ru',
    'vi.json':    'vie',
    'fr.json':    'fra',
    'es.json':    'spa',
    'de.json':    'de',
    'it.json':    'it',
    'pt.json':    'pt',
    'nl.json':    'nl',
    'ar.json':    'ara',
}

def baidu_translate(text, to_lang, source_lang='zh'):
    """调用百度翻译 API"""
    if not text.strip():
        return text
    
    salt = str(random.randint(1, 100000))
    sign_str = BAIDU_APPID + text + salt + BAIDU_KEY
    sign = hashlib.md5(sign_str.encode('utf-8')).hexdigest()
    
    params = urllib.parse.urlencode({
        'q': text,
        'from': source_lang,
        'to': to_lang,
        'appid': BAIDU_APPID,
        'salt': salt,
        'sign': sign,
    })
    
    url = f"{BAIDU_API}?{params}"
    
    try:
        req = urllib.request.Request(url, headers={
            'User-Agent': 'Mozilla/5.0',
            'Content-Type': 'application/x-www-form-urlencoded',
        })
        with urllib.request.urlopen(req, timeout=10) as resp:
            result = json.loads(resp.read().decode('utf-8'))
        
        if 'error_code' in result:
            print(f"    ⚠️  API Error {result['error_code']}: {result.get('error_msg', '')}")
            return None
        
        if 'trans_result' in result:
            return result['trans_result'][0]['dst']
        return None
    except Exception as e:
        print(f"    ⚠️  Request failed: {e}")
        return None


def load_json(path):
    """安全加载 JSON（兼容 BOM）"""
    with open(path, 'r', encoding='utf-8-sig') as f:
        return json.load(f)


def save_json(path, data):
    """保存 JSON（UTF-8, indent=2, ensure_ascii=False）"""
    with open(path, 'w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
    # Fix: ensure trailing newline
    with open(path, 'a', encoding='utf-8') as f:
        f.write('\n')


def get_missing_keys(source_data, target_data):
    """返回 target 中缺失的 key 列表"""
    source_keys = set(source_data.keys())
    target_keys = set(target_data.keys())
    return source_keys - target_keys


def main():
    i18n_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 
                           'web/default/src/i18n')
    
    print("="*60)
    print("QuantumClaw i18n Translation Sync")
    print("="*60)
    
    # 1. Load zh-CN (source of truth)
    zh_path = os.path.join(i18n_dir, 'zh-CN.json')
    zh_data = load_json(zh_path)
    print(f"\n📖 Source: zh-CN.json ({len(zh_data)} keys)")
    
    check_only = '--check' in sys.argv
    
    # 2. Process each language
    total_translated = 0
    total_errors = 0
    
    for filename, baidu_code in sorted(LANG_MAP.items()):
        lang_path = os.path.join(i18n_dir, filename)
        lang_data = load_json(lang_path)
        
        missing = get_missing_keys(zh_data, lang_data)
        
        if not missing:
            print(f"\n✅ {filename} - 完整 ({len(lang_data)} keys)")
            continue
        
        print(f"\n📝 {filename} - 缺 {len(missing)} 键 ({baidu_code})")
        
        if check_only:
            for k in sorted(missing):
                print(f"    - {k}")
            continue
        
        # For English: just use the key as value
        if filename == 'en.json':
            for k in missing:
                lang_data[k] = k
            save_json(lang_path, lang_data)
            print(f"    ✅ English: 直接使用键名 (添加 {len(missing)} 键)")
            total_translated += len(missing)
            continue
        
        # For zht-TW: quick conversion
        if filename == 'zh-TW.json':
            for k in missing:
                # Use Baidu Translate from zh to cht
                src_text = zh_data[k]
                result = baidu_translate(src_text, 'cht')
                if result:
                    lang_data[k] = result
                    total_translated += 1
                    print(f"    ✅ {k} → {result}")
                else:
                    lang_data[k] = src_text  # fallback
                    total_errors += 1
                    print(f"    ⚠️ {k} → fallback to source")
                time.sleep(0.3)  # Rate limit
            save_json(lang_path, lang_data)
            continue
        
        # For all other languages: translate from zh-CN
        batch = []
        batch_keys = []
        for k in sorted(missing):
            batch.append(zh_data[k])
            batch_keys.append(k)
        
        # Translate one by one (Baidu has length limit)
        for i, k in enumerate(batch_keys):
            src_text = zh_data[k]
            result = baidu_translate(src_text, baidu_code)
            if result:
                lang_data[k] = result
                total_translated += 1
                print(f"    ✅ {k[:20]:<22} → {result[:40]}")
            else:
                lang_data[k] = src_text  # fallback to Chinese
                total_errors += 1
                print(f"    ⚠️ {k[:20]:<22} → FALLBACK")
            
            # Rate limit: max 1 req/s
            time.sleep(0.5)
        
        save_json(lang_path, lang_data)
    
    # 3. Summary
    print(f"\n{'='*60}")
    print(f"SUMMARY: Translated {total_translated} keys, {total_errors} fallbacks")
    print(f"{'='*60}")
    
    # 4. Verify final state
    print(f"\n{'='*60}")
    print("FINAL VERIFICATION:")
    print(f"{'='*60}")
    zh_keys = set(zh_data.keys())
    for fname in sorted(LANG_MAP.keys()):
        fp = os.path.join(i18n_dir, fname)
        ld = load_json(fp)
        lk = set(ld.keys())
        miss = zh_keys - lk
        status = '✅' if not miss else '❌'
        print(f"  {status} {fname:<12} {len(ld):<6} keys, missing: {len(miss)}")


if __name__ == '__main__':
    main()
