// _translate_pass2.cjs — 第二轮：用编号行格式补译低覆盖率语言
const path = require('path');
const fs = require('fs');
const I18N = __dirname;

// Read API key
let API_KEY = '';
try {
  const auth = JSON.parse(fs.readFileSync('C:\\Users\\jones\\.openclaw\\agents\\main\\agent\\auth-profiles.json', 'utf8'));
  API_KEY = auth.profiles['deepseek:default'].key;
} catch(e) { console.error('No API key'); process.exit(1); }

const API_URL = 'https://api.deepseek.com/v1/chat/completions';
const MODEL = 'deepseek-chat';
const BATCH = 100; // bigger batch since it's just line-by-line

// Focus on remaining untranslated for each language
const LANGS = [
  ['zh-TW','繁體中文'], ['ja','日本語'], ['ko','한국어'], ['ru','Русский'],
  ['vi','Tiếng Việt'], ['fr','Français'], ['es','Español'], ['de','Deutsch'],
  ['it','Italiano'], ['pt','Português'], ['nl','Nederlands'], ['tr','Türkçe'],
  ['th','ไทย'], ['ar','العربية'], ['hi','हिन्दी'], ['id','Bahasa Indonesia'],
];

function loadJSON(fp) {
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

async function translateLineBatch(keys, src, displayName) {
  const input = keys.map((k, i) => `${i + 1}. ${src[k] || k}`).join('\n');
  const prompt = `Translate to ${displayName}. Output one line per number.\n\n${input}\n\nTranslations:\n`;
  
  const body = JSON.stringify({
    model: MODEL,
    messages: [
      { role: 'system', content: `You translate text to ${displayName}. Return one line per number.` },
      { role: 'user', content: prompt }
    ],
    temperature: 0.05,
    max_tokens: 8000,
  });

  const res = await fetch(API_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${API_KEY}` },
    body
  });
  if (!res.ok) { throw new Error(`HTTP ${res.status}`); }
  
  const data = await res.json();
  const content = data.choices[0].message.content;
  const lines = content.split('\n').filter(l => l.trim());
  
  const result = {};
  for (const line of lines) {
    const m = line.match(/^(\d+)[.、\s:-]\s*(.+)/);
    if (m) {
      const idx = parseInt(m[1]) - 1;
      const val = m[2].trim();
      if (idx >= 0 && idx < keys.length && val && val !== (src[keys[idx]] || keys[idx])) {
        result[keys[idx]] = val;
      }
    }
  }
  return result;
}

async function main() {
  console.log(`=== 第二轮翻译 (${MODEL}, line-format, batch=${BATCH}) ===\n`);
  
  const src = loadJSON(path.join(I18N, 'zh-CN.json'));
  
  for (const [code, displayName] of LANGS) {
    const tgt = loadJSON(path.join(I18N, code + '.json'));
    const keys = findUntranslated(tgt).filter(k => k in src && src[k] !== k && src[k].length < 500);
    
    if (keys.length === 0) { console.log(`${code}: ✅ 全部已翻译`); continue; }
    
    const total = keys.length;
    console.log(`${code} (${displayName}): ${total} 条待译`);
    let translated = 0;
    
    for (let i = 0; i < total; i += BATCH) {
      const batch = keys.slice(i, i + BATCH);
      process.stdout.write(`  [${Math.floor(i / BATCH) + 1}/${Math.ceil(total / BATCH)}] ${batch.length}条...`);
      
      try {
        const result = await translateLineBatch(batch, src, displayName);
        let batchDone = 0;
        for (const [k, v] of Object.entries(result)) {
          tgt[k] = v;
          batchDone++;
        }
        if (batchDone > 0) fs.writeFileSync(path.join(I18N, code + '.json'), JSON.stringify(tgt, null, 2), 'utf8');
        translated += batchDone;
        console.log(` ✅ ${batchDone}/${batch.length}`);
      } catch(e) {
        console.log(` ❌ ${e.message.substring(0, 40)}`);
      }
    }
    
    console.log(`  ✅ ${code}: +${translated}/${total}`);
  }
  
  console.log(`\n=== 第二轮完成 ===`);
  try { require('./_verify_tkeys.cjs'); } catch {}
}

main().catch(e => { console.error('FATAL:', e); process.exit(1); });
