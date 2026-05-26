// _translate_deepseek.cjs — 用 DeepSeek API 批量翻译（从 OpenClaw 获取 Key）
const path = require('path');
const fs = require('fs');
const I18N = __dirname;

// Read DeepSeek API key from OpenClaw auth config
const authPath = path.join(process.env.OPENCLAW_ROOT || path.dirname(process.env.HOME || process.env.USERPROFILE) + '\\.openclaw', 'agents', 'main', 'agent', 'auth-profiles.json');
let API_KEY = '';

try {
  const auth = JSON.parse(fs.readFileSync(authPath, 'utf8'));
  if (auth.profiles && auth.profiles['deepseek:default']) {
    API_KEY = auth.profiles['deepseek:default'].key;
  }
} catch (e) {
  // Fallback
  const alt = 'C:\\Users\\jones\\.openclaw\\agents\\main\\agent\\auth-profiles.json';
  try { const auth = JSON.parse(fs.readFileSync(alt, 'utf8')); API_KEY = auth.profiles['deepseek:default'].key; } catch(e2) {}
}

if (!API_KEY) {
  console.error('ERROR: Could not find DeepSeek API key');
  process.exit(1);
}

console.log(`DeepSeek API Key: ${API_KEY.substring(0, 8)}...`);

const API_URL = 'https://api.deepseek.com/v1/chat/completions';
const MODEL = 'deepseek-chat'; // cheaper than v4-flash for simple translation
const BATCH = 50;

const LANGS = [
  ['zh-TW','繁體中文'], ['ja','日本語'], ['ko','한국어'], ['ru','Русский'],
  ['vi','Tiếng Việt'], ['fr','Français'], ['es','Español'], ['de','Deutsch'],
  ['it','Italiano'], ['pt','Português'], ['nl','Nederlands'], ['tr','Türkçe'],
  ['th','ไทย'], ['ar','العربية'], ['hi','हिन्दी'], ['id','Bahasa Indonesia'],
];

function loadJSON(fp) {
  if (!fs.existsSync(fp)) return {};
  let raw = fs.readFileSync(fp, 'utf8');
  if (raw.charCodeAt(0) === 0xFEFF) raw = raw.slice(1);
  return JSON.parse(raw);
}

function findUntranslated(data) {
  const out = [];
  for (const [k, v] of Object.entries(data))
    if ((v === k || v === '') && !k.startsWith('/') && !k.startsWith('@') && !/^\d+$/.test(k) && k !== '-' && k !== ';' && k !== '0')
      out.push(k);
  return out;
}

async function translateBatchViaDeepSeek(keys, src, displayName) {
  // Build input JSON
  const input = {};
  keys.forEach(k => { input[k] = src[k] || k; });

  const systemMsg = `You are a professional translator. Translate the JSON values to ${displayName}. Return ONLY a valid JSON object with the same keys and translated values. Do NOT translate the keys themselves. Do NOT add any explanation or extra text.`;
  const userMsg = JSON.stringify(input, null, 2) + '\n\nOutput JSON:';

  const body = JSON.stringify({
    model: MODEL,
    messages: [
      { role: 'system', content: systemMsg },
      { role: 'user', content: userMsg }
    ],
    temperature: 0.05,
    max_tokens: 4000,
    response_format: { type: 'json_object' }
  });

  const res = await fetch(API_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${API_KEY}` },
    body
  });

  if (!res.ok) {
    const errText = await res.text();
    throw new Error(`HTTP ${res.status}: ${errText.substring(0, 100)}`);
  }

  const data = await res.json();
  const content = data.choices[0].message.content;

  let result;
  try { result = JSON.parse(content); }
  catch (e) { throw new Error(`JSON parse failed: ${e.message.substring(0, 50)}`); }

  return result;
}

async function main() {
  console.log(`=== DeepSeek API 批量翻译 (${MODEL}, batch=${BATCH}) ===\n`);

  for (const [code, displayName] of LANGS) {
    const src = loadJSON(path.join(I18N, 'zh-CN.json'));
    const tgt = loadJSON(path.join(I18N, code + '.json'));

    const keys = findUntranslated(tgt).filter(k => k in src && src[k] !== k && src[k].length < 500);
    if (keys.length === 0) {
      console.log(`${code}: ✅ 全部已翻译`);
      continue;
    }

    const total = keys.length;
    console.log(`${code} (${displayName}): ${total} 条待译`);

    let translated = 0;
    for (let i = 0; i < total; i += BATCH) {
      const batch = keys.slice(i, i + BATCH);
      process.stdout.write(`  [${Math.floor(i / BATCH) + 1}/${Math.ceil(total / BATCH)}] ${batch.length}条...`);

      try {
        const result = await translateBatchViaDeepSeek(batch, src, displayName);
        let batchDone = 0;
        for (const [k, v] of Object.entries(result)) {
          if (v && v !== src[k] && k in src) {
            tgt[k] = v;
            batchDone++;
          }
        }
        if (batchDone > 0) {
          fs.writeFileSync(path.join(I18N, code + '.json'), JSON.stringify(tgt, null, 2), 'utf8');
        }
        translated += batchDone;
        console.log(` ✅ ${batchDone}/${batch.length}`);
      } catch (e) {
        console.log(` ❌ ${e.message.substring(0, 60)}`);
      }
    }

    console.log(`  ✅ ${code}: +${translated}/${total}`);
  }

  console.log(`\n=== 全部完成 ===`);
}

main().catch(e => { console.error('FATAL:', e); process.exit(1); });
