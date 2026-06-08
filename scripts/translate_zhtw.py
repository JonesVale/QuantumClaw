#!/usr/bin/env python3
"""
Step 2: zh-TW translation using OpenCC (Simplified → Traditional)
Also counts remaining untranslated for all languages
"""
import json, os
from opencc import OpenCC

BASE = r"H:\AiData\openclaw\workspace\QuantumClaw"
I18N = os.path.join(BASE, "web", "default", "src", "i18n")

# Load references
with open(os.path.join(I18N, "zh-CN.json"), encoding="utf-8") as f:
    zh_cn = json.load(f)
with open(os.path.join(I18N, "en.json"), encoding="utf-8") as f:
    en = json.load(f)

all_keys = set(zh_cn.keys()) | set(en.keys())

# ─── Step 1: zh-TW via OpenCC ───
cc = OpenCC('s2t')
zh_tw_path = os.path.join(I18N, "zh-TW.json")
with open(zh_tw_path, encoding="utf-8") as f:
    zh_tw = json.load(f)

before = len(zh_tw)
filled = 0
skipped = 0

for k in all_keys:
    # Skip if zh-TW already has a proper translation (not same as en)
    if k in zh_tw and isinstance(zh_tw[k], str) and isinstance(en.get(k), str):
        if zh_tw[k] != en[k] and zh_tw[k]:
            continue  # Already translated, skip

    zh_cn_val = zh_cn.get(k, "")
    if isinstance(zh_cn_val, str) and zh_cn_val:
        # Special handling: keep brand names and English text as-is
        # Only convert if value contains Chinese characters
        if any('\u4e00' <= c <= '\u9fff' for c in zh_cn_val):
            tw_val = cc.convert(zh_cn_val)
            # Only write if conversion actually changed something or key didn't exist
            if k not in zh_tw or zh_tw[k] != tw_val:
                zh_tw[k] = tw_val
                filled += 1
            else:
                skipped += 1
        elif k not in zh_tw:
            # No Chinese in value, but key is missing - add as-is
            zh_tw[k] = zh_cn_val
            filled += 1
    elif isinstance(zh_cn_val, (int, float, bool, list, dict)):
        if k not in zh_tw:
            zh_tw[k] = zh_cn_val
            filled += 1

# Sort and save
zh_tw = dict(sorted(zh_tw.items(), key=lambda x: x[0].lower()))
with open(zh_tw_path, "w", encoding="utf-8", newline="\n") as f:
    json.dump(zh_tw, f, ensure_ascii=False, indent=2)

print(f"zh-TW: {before} -> {len(zh_tw)} keys ({filled} filled, {skipped} skipped)")

# ─── Step 2: Count remaining untranslated per language ───
print("\n=== REMAINING UNTRANSLATED (same as en.json) ===")
files = sorted([f for f in os.listdir(I18N) if f.endswith('.json')])
all_data = {}
for f in files:
    with open(os.path.join(I18N, f), encoding="utf-8") as fp:
        all_data[f] = json.load(fp)

en_data = all_data['en.json']
for f in files:
    if f in ('en.json', 'zh-CN.json', 'zh-TW.json'):
        continue
    ld = all_data[f]
    untrans = 0
    for k in all_keys & set(ld.keys()):
        if isinstance(ld.get(k), str) and isinstance(en_data.get(k), str):
            if ld[k] == en_data[k] and en_data[k]:
                untrans += 1
    print(f"  [{f.replace('.json','')}] {untrans} untranslated")

# Count zh-TW remaining
zh_tw_data = all_data['zh-TW.json']
tw_untrans = sum(1 for k in all_keys & set(zh_tw_data.keys())
                 if isinstance(zh_tw_data.get(k), str) and isinstance(en_data.get(k), str)
                 and zh_tw_data[k] == en_data[k] and en_data[k])
print(f"  [zh-TW] {tw_untrans} untranslated (after OpenCC)")

print("\n=== KEY COUNTS ===")
for f in files:
    print(f"  {f.replace('.json','')}: {len(all_data[f])} keys")

print("\nDONE: zh-TW translation complete")
