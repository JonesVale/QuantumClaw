/**
 * _fill_translations.cjs — 用本地 Ollama 填补未翻译的键
 *
 * 逻辑：
 *   1. 加载目标语言的当前 .json 文件
 *   2. 找出 identity 键（value === key）— 这些需要翻译
 *   3. 用 zh-CN 中的文本作为源文本
 *   4. 通过本地 Ollama (qwen3.5:4b) 批量翻译
 *   5. 只覆盖 identity 键，保留已有翻译不变
 *   6. 跳过 DB seed（后端语言 API 已移除）
 *
 * 用法: node _fill_translations.cjs [lang_code]
 *       不传参数 = 全部语言，传参 = 只跑指定的一种
 *
 * 依赖: Ollama 运行于 localhost:11434
 */
const http = require('http');
const fs = require('fs');
const path = require('path');

// ========== 目标语言（代码 → 显示名）
const LANGS = [
  ['zh-TW', '繁體中文'],
  ['ja',    '日本語'],
  ['ko',    '한국어'],
  ['ru',    'Русский'],
  ['vi',    'Tiếng Việt'],
  ['fr',    'Français'],
  ['es',    'Español'],
  ['de',    'Deutsch'],
  ['it',    'Italiano'],
  ['pt',    'Português'],
  ['nl',    'Nederlands'],
  ['tr',    'Türkçe'],
  ['th',    'ไทย'],
  ['ar',    'العربية'],
  ['hi',    'हिन्दी'],
  ['id',    'Bahasa Indonesia'],
];

const MODEL = 'qwen3.5:4b';
const BATCH = 5;
const MAX_RETRIES = 3;
const OL_HOST = 'localhost';
const OL_PORT = 11434;

// ========== 加载 JSON（跳过 BOM）
function loadJSON(fp) {
  if (!fs.existsSync(fp)) return {};
  let raw = fs.readFileSync(fp, 'utf8');
  if (raw.charCodeAt(0) === 0xFEFF) raw = raw.slice(1);
  return JSON.parse(raw);
}

// ========== 找出语言的未翻译键（value === key，非 API/数字）
function findUntranslated(langData, sourceData) {
  const result = [];
  for (const [k, v] of Object.entries(langData)) {
    if (v === k || v === '') {
      // 排除 API 路径、数字等无需翻译的键
      if (k.startsWith('/') || k.startsWith('@') || /^\d+$/.test(k) || k === '-' || k === ';' || k === '0' || k === '') continue;
      // 取源文本
      const src = sourceData[k];
      if (src && src !== k && src.length > 0 && src.length < 500) {
        result.push({ key: k, source: src });
      }
    }
  }
  return result;
}

// ========== 调用 Ollama 翻译
function ollamaTranslate(texts, targetLang) {
  return new Promise((resolve, reject) => {
    const prompt = `Translate the following ${texts.length} text(s) to ${targetLang}. Return ONLY a valid JSON array of translated strings in the same order. No markdown, no explanation.

Input: ${JSON.stringify(texts)}

Output:`;
    const body = JSON.stringify({ model: MODEL, prompt, stream: false, options: { temperature: 0.1 } });
    const req = http.request({ hostname: OL_HOST, port: OL_PORT, path: '/api/generate', method: 'POST',
      headers: { 'Content-Type': 'application/json' } }, res => {
      let d = '';
      res.on('data', c => d += c);
      res.on('end', () => {
        try {
          const resp = JSON.parse(d).response;
          const s = resp.indexOf('['), e = resp.lastIndexOf(']');
          if (s >= 0 && e > s + 1) {
            const arr = JSON.parse(resp.substring(s, e + 1));
            if (Array.isArray(arr) && arr.length === texts.length) return resolve(arr);
          }
          reject('bad response: ' + resp.substring(0, 120));
        } catch (err) { reject(err.message); }
      });
    });
    req.on('error', reject);
    req.write(body); req.end();
  });
}

// ========== 写回 JSON（保留原始键顺序，只更新未翻译的）
function patchFile(code, langData, translations) {
  // translations is Map<key, newValue>
  for (const [k, v] of translations) {
    if (v && v !== k && k in langData) langData[k] = v;
  }
  const fp = path.join(__dirname, code + '.json');
  fs.writeFileSync(fp, JSON.stringify(langData, null, 2), 'utf8');
}

// ========== 处理一种语言
async function processLang(code, displayName, sourceData) {
  const langData = loadJSON(path.join(__dirname, code + '.json'));
  const untranslated = findUntranslated(langData, sourceData);

  if (untranslated.length === 0) {
    console.log(`  ${code} (${displayName}): ✅ 全部已翻译`);
    return 0;
  }

  const totalIdentity = Object.entries(langData).filter(([k,v]) => v === k).length;
  const totalKeys = Object.keys(langData).length;
  console.log(`\n${code} (${displayName}): 需译 ${untranslated.length} / 共 ${totalKeys} (identity ${totalIdentity})`);
  console.time('  time');

  let done = 0;
  let errors = 0;

  for (let i = 0; i < untranslated.length; i += BATCH) {
    const batch = untranslated.slice(i, i + BATCH);
    const sourceTexts = batch.map(b => b.source);

    let result = null;
    for (let attempt = 0; attempt < MAX_RETRIES; attempt++) {
      try {
        result = await ollamaTranslate(sourceTexts, displayName);
        break;
      } catch (e) {
        if (attempt < MAX_RETRIES - 1) await new Promise(r => setTimeout(r, 3000));
        else { errors++; /* console.log(`    [${i/BATCH+1}] fail: ${e.substring(0,60)}`); */ }
      }
    }

    if (result) {
      const updates = new Map();
      batch.forEach((b, idx) => {
        if (result[idx] && result[idx] !== b.key) {
          updates.set(b.key, result[idx]);
        }
      });
      if (updates.size > 0) patchFile(code, langData, updates);
      done += updates.size;
    }

    // 每 10 批 + 最后一批 打印进度
    if ((i / BATCH) % 10 === 0 || i + BATCH >= untranslated.length) {
      const pct = ((done / untranslated.length) * 100).toFixed(1);
      console.log(`  ${code}: ${done}/${untranslated.length} (${pct}%)`);
    }
  }

  console.timeEnd('  time');
  const nowTranslated = Object.keys(langData).length - Object.entries(langData).filter(([k,v]) => v === k).length;
  console.log(`  ✅ ${code}: 共 ${nowTranslated}/${Object.keys(langData).length} 已翻译 (+${done})`);
  return done;
}

// ========== 入口
async function main() {
  const targetArg = process.argv[2];
  const sourceData = loadJSON(path.join(__dirname, 'zh-CN.json'));
  const langs = targetArg ? LANGS.filter(([c]) => c === targetArg) : LANGS;

  if (langs.length === 0) { console.log('Language not found:', targetArg); process.exit(1); }

  console.log(`=== 批量翻译 (${MODEL}) [BATCH=${BATCH}] ===\n`);

  let total = 0, totalErrors = 0;
  for (const [code, displayName] of langs) {
    try {
      total += await processLang(code, displayName, sourceData);
    } catch (e) {
      console.error(`\n  ❌ ${code}: ${e.message}`);
      totalErrors++;
    }
  }

  console.log(`\n=== 全部完成 ===\n新增翻译: ${total} 条`);
  if (totalErrors > 0) console.log(`错误语言数: ${totalErrors}`);

  // 自动验证
  console.log(`\n--- 验证 ---`);
  try { require('./_verify_tkeys.cjs'); } catch {}

  process.exit(totalErrors > 0 ? 1 : 0);
}

main().catch(e => { console.error('FATAL:', e.message); process.exit(1); });
