import json, os

# Force correct absolute path
BASE_DIR = r"H:\AiData\openclaw\workspace\QuantumClaw"
I18N_DIR = os.path.join(BASE_DIR, "web", "default", "src", "i18n")

print(f"I18N DIR: {I18N_DIR}")
print(f"Exists: {os.path.isdir(I18N_DIR)}\n")

files = sorted([f for f in os.listdir(I18N_DIR) if f.endswith('.json')])
print(f"Found {len(files)} language files: {files}\n")

# Load all JSON
data = {}
for f in files:
    path = os.path.join(I18N_DIR, f)
    with open(path, encoding='utf-8') as fp:
        data[f] = json.load(fp)
    print(f"  {f}: {len(data[f])} keys")

# Use zh-CN as reference (has most keys: 2133)
ref = data['zh-CN.json']
ref_keys = set(ref.keys())

print(f"\n{'='*60}")
print("1. MISSING KEYS per language (vs zh-CN.json)")
print('='*60)
missing_report = {}
for f in files:
    if f == 'zh-CN.json':
        continue
    missing = sorted(ref_keys - set(data[f].keys()))
    missing_report[f] = missing
    if missing:
        print(f"\n[{f}] Missing {len(missing)} keys:")
        for k in missing[:30]:
            v = str(ref[k])[:70]
            print(f"  MISSING: {k} = {v}")
        if len(missing) > 30:
            print(f"  ... and {len(missing)-30} more")
    else:
        print(f"[{f}] ✅ All {len(ref_keys)} keys present")

# Check for Chinese text in non-zh languages
print(f"\n{'='*60}")
print("2. CHINESE TEXT in non-zh languages (untranslated)")
print('='*60)
for f in files:
    if f.startswith('zh'):
        continue
    chinese_vals = []
    for k in ref_keys & set(data[f].keys()):
        val = data[f][k]
        if isinstance(val, str) and any('\u4e00' <= c <= '\u9fff' for c in val):
            chinese_vals.append((k, val[:60]))
    if chinese_vals:
        print(f"\n[{f}] {len(chinese_vals)} values still in Chinese:")
        for k, v in chinese_vals[:20]:
            print(f"  HAS_CHINESE: {k} = {v}")
        if len(chinese_vals) > 20:
            print(f"  ... and {len(chinese_vals)-20} more")
    else:
        print(f"[{f}] ✅ No Chinese text found")

# Check for garbled/encoding-broken text
print(f"\n{'='*60}")
print("3. GARBLED/ENCODING-BROKEN text check")
print('='*60)
import re
garble_pat = re.compile(r'[\ufffd\ufffe\uffff\u200b\u2028\u2029]')
for f in files:
    garbled = []
    for k, v in data[f].items():
        if isinstance(v, str) and garble_pat.search(v):
            garbled.append((k, v[:60]))
    if garbled:
        print(f"\n[{f}] {len(garbled)} garbled values:")
        for k, v in garbled[:10]:
            print(f"  GARBLED: {k} = {repr(v[:50])}")
    else:
        print(f"[{f}] ✅ No garbled text found")

print(f"\n{'='*60}")
print("ANALYSIS COMPLETE — results saved to missing_report dict")
print('='*60)

# Save missing keys report to file for next step
report_path = os.path.join(BASE_DIR, "scripts", "i18n_missing_report.txt")
with open(report_path, 'w', encoding='utf-8') as fp:
    for f, keys in missing_report.items():
        fp.write(f"\n=== {f} ({len(keys)} missing) ===\n")
        for k in keys:
            fp.write(f"  {k} = {str(ref[k])[:80]}\n")
print(f"\nReport saved to: {report_path}")
