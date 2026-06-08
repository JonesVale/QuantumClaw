#!/usr/bin/env python3
"""
QuantumClaw i18n 翻译完善脚本
策略: 只做增量，不删不改已有翻译数据
- Step 1: 给所有语言添加 8 个缺失 key
- Step 2: 填充和 en.json 完全相同的未翻译值（只增不删）
"""
import json, os, re

BASE = r"H:\AiData\openclaw\workspace\QuantumClaw"
I18N = os.path.join(BASE, "web", "default", "src", "i18n")

# ─── 8 个缺失 key 的翻译 ───
MISSING_TRANSLATIONS = {
    "zh-TW": {
        "All Pages": "全部頁面",
        "Apifox Playground": "Apifox 線上調試",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "百度、阿里、DeepSeek、智譜、騰訊...",
        "Cents": "分",
        "Download on the": "下載自",
        "Get it on": "獲取自",
        "Google Play": "Google Play",
    },
    "ja": {
        "All Pages": "すべてのページ",
        "Apifox Playground": "Apifox プレイグラウンド",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "セント",
        "Download on the": "ダウンロード",
        "Get it on": "入手する",
        "Google Play": "Google Play",
    },
    "ko": {
        "All Pages": "전체 페이지",
        "Apifox Playground": "Apifox 플레이그라운드",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "센트",
        "Download on the": "다운로드",
        "Get it on": "다운로드",
        "Google Play": "Google Play",
    },
    "fr": {
        "All Pages": "Toutes les pages",
        "Apifox Playground": "Apifox Playground",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "Cents",
        "Download on the": "Télécharger sur",
        "Get it on": "Disponible sur",
        "Google Play": "Google Play",
    },
    "de": {
        "All Pages": "Alle Seiten",
        "Apifox Playground": "Apifox Playground",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "Cent",
        "Download on the": "Herunterladen im",
        "Get it on": "Jetzt bei",
        "Google Play": "Google Play",
    },
    "es": {
        "All Pages": "Todas las páginas",
        "Apifox Playground": "Apifox Playground",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "Centavos",
        "Download on the": "Descargar en",
        "Get it on": "Disponible en",
        "Google Play": "Google Play",
    },
    "it": {
        "All Pages": "Tutte le pagine",
        "Apifox Playground": "Apifox Playground",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "Centesimi",
        "Download on the": "Scarica su",
        "Get it on": "Disponibile su",
        "Google Play": "Google Play",
    },
    "ru": {
        "All Pages": "Все страницы",
        "Apifox Playground": "Apifox Playground",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "Центов",
        "Download on the": "Скачать в",
        "Get it on": "Доступно в",
        "Google Play": "Google Play",
    },
    "vi": {
        "All Pages": "Tất cả trang",
        "Apifox Playground": "Apifox Playground",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "Xu",
        "Download on the": "Tải xuống trên",
        "Get it on": "Tải trên",
        "Google Play": "Google Play",
    },
    "nl": {
        "All Pages": "Alle pagina's",
        "Apifox Playground": "Apifox Playground",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "Cent",
        "Download on the": "Downloaden in",
        "Get it on": "Verkrijgbaar op",
        "Google Play": "Google Play",
    },
    "pt": {
        "All Pages": "Todas as páginas",
        "Apifox Playground": "Apifox Playground",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "Centavos",
        "Download on the": "Baixar na",
        "Get it on": "Disponível na",
        "Google Play": "Google Play",
    },
    "ar": {
        "All Pages": "جميع الصفحات",
        "Apifox Playground": "Apifox Playground",
        "App Store": "App Store",
        "Baidu, Ali, DeepSeek, Zhipu, Tencent...": "Baidu, Ali, DeepSeek, Zhipu, Tencent...",
        "Cents": "سنت",
        "Download on the": "تنزيل على",
        "Get it on": "احصل عليه من",
        "Google Play": "Google Play",
    },
}

# Load reference files
with open(os.path.join(I18N, "zh-CN.json"), encoding="utf-8") as f:
    zh_cn = json.load(f)
with open(os.path.join(I18N, "en.json"), encoding="utf-8") as f:
    en = json.load(f)

zh_cn_keys = set(zh_cn.keys())
en_keys = set(en.keys())
all_ref_keys = zh_cn_keys | en_keys

stats = {}

for lang, translations in MISSING_TRANSLATIONS.items():
    fn = f"{lang}.json"
    path = os.path.join(I18N, fn)
    with open(path, encoding="utf-8") as f:
        lang_data = json.load(f)

    added_missing = 0
    added_untranslated = 0

    # Step 1: add missing keys
    for k, v in translations.items():
        if k not in lang_data:
            lang_data[k] = v
            added_missing += 1

    # Step 2: fill untranslated values (same as en.json → translate)
    for k in all_ref_keys:
        if k not in lang_data:
            continue
        if not isinstance(lang_data.get(k), str) or not isinstance(en.get(k), str):
            continue
        if lang_data[k] != en[k]:
            continue  # already translated, skip
        if not en[k]:
            continue  # empty value, skip

        # Try to get translation from our dict first
        if k in translations:
            lang_data[k] = translations[k]
            added_untranslated += 1
        else:
            # For keys not in our manual translation list,
            # we need a smarter approach - skip for now
            pass

    # Sort and save
    lang_data = dict(sorted(lang_data.items(), key=lambda x: x[0].lower()))
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        json.dump(lang_data, f, ensure_ascii=False, indent=2)

    stats[lang] = {"missing_added": added_missing, "untranslated_fixed": added_untranslated}
    print(f"[{lang}] +{added_missing} missing keys, +{added_untranslated} untranslated filled")

print("\n=== SUMMARY ===")
total_missing = sum(s["missing_added"] for s in stats.values())
total_untrans = sum(s["untranslated_fixed"] for s in stats.values())
print(f"Total: +{total_missing} missing keys, +{total_untrans} untranslated values fixed")
