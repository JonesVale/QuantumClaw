// _translate.cjs — batch translation with Ollama, context-reset between calls
const http = require('http');
const fs = require('fs');
const path = require('path');

const LANGS = [
  ['zh-TW','繁體中文'], ['ja','日本語'], ['ko','한국어'], ['ru','Русский'],
  ['vi','Tiếng Việt'], ['fr','Français'], ['es','Español'], ['de','Deutsch'],
  ['it','Italiano'], ['pt','Português'], ['nl','Nederlands'], ['tr','Türkçe'],
  ['th','ไทย'], ['ar','العربية'], ['hi','हिन्दी'], ['id','Bahasa Indonesia'],
];

const MODEL = 'qwen3.5:4b';
const BATCH = 10;
const OLLAMA = { hostname: 'localhost', port: 11434 };
const I18N = __dirname;
const PROGRESS = path.join(I18N, '_progress.json');

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

function callOllama(prompt) {
  return new Promise((resolve, reject) => {
    // keep_alive: "0s" = unload model between requests, prevents context accumulation slowdown
    const body = JSON.stringify({ model: MODEL, prompt, stream: false, options: { temperature: 0.1 }, keep_alive: "0s" });
    const req = http.request({ ...OLLAMA, path: '/api/generate', method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(body) }
    }, res => {
      let d = '';
      res.on('data', c => d += c);
      res.on('end', () => { try { const j = JSON.parse(d); resolve(j.response || ''); } catch(e) { reject(e.message); } });
    });
    req.on('error', reject);
    req.write(body); req.end();
  });
}

async function main() {
  let progress = { done: {} };
  try { progress = JSON.parse(fs.readFileSync(PROGRESS, 'utf8')); } catch {}

  console.log(`=== 批量翻译 (${MODEL}, batch=${BATCH}, keep_alive=0s) ===`);

  for (const [code, displayName] of LANGS) {
    const src = loadJSON(path.join(I18N, 'zh-CN.json'));
    const tgt = loadJSON(path.join(I18N, code + '.json'));

    if (progress.done[code]) { console.log(`\n${code}: ✅ 已完成`); continue; }

    const keys = findUntranslated(tgt).filter(k => k in src && src[k] !== k && src[k].length < 500);
    if (keys.length === 0) { console.log(`\n${code}: ✅ 全部已翻译`); progress.done[code] = true; fs.writeFileSync(PROGRESS, JSON.stringify(progress)); continue; }

    const total = keys.length;
    console.log(`\n${code} (${displayName}): ${total} 条待译 / ${Object.keys(tgt).length} 总量`);
    let translated = 0;

    for (let i = 0; i < total; i += BATCH) {
      const batch = keys.slice(i, i + BATCH);
      const input = batch.map((k, idx) => `${idx + 1}. ${src[k] || k}`).join('\n');
      const prompt = `Translate to ${displayName}. Output one line per number.\n\n${input}\n\nTranslations:`;

      process.stdout.write(`  [${Math.floor(i / BATCH) + 1}/${Math.ceil(total / BATCH)}] ${batch.length}条...`);
      let resp;
      try { resp = await callOllama(prompt); } catch (e) { console.log(` ❌ ${e.substring(0, 30)}`); continue; }

      const lines = resp.split('\n').filter(l => l.trim());
      let batchDone = 0;
      for (const line of lines) {
        const m = line.match(/^(\d+)[.、\s:-]\s*(.+)/);
        if (m) {
          const idx = parseInt(m[1]) - 1;
          const val = m[2].trim();
          if (idx >= 0 && idx < batch.length && val && val !== batch[idx] && val !== (src[batch[idx]] || batch[idx])) {
            tgt[batch[idx]] = val;
            batchDone++;
          }
        }
      }

      if (batchDone > 0) fs.writeFileSync(path.join(I18N, code + '.json'), JSON.stringify(tgt, null, 2), 'utf8');
      translated += batchDone;
      console.log(` ✅ ${batchDone}/${batch.length}`);
    }

    progress.done[code] = true;
    fs.writeFileSync(PROGRESS, JSON.stringify(progress));
    const remain = findUntranslated(tgt).length;
    console.log(`  ✅ ${code}: ${Object.keys(tgt).length - remain}/${Object.keys(tgt).length} (+${translated})`);
  }

  fs.writeFileSync(PROGRESS, JSON.stringify({ completed: true, at: new Date().toISOString() }));
  console.log(`\n=== 全部完成 ===`);
  try { require('./_verify_tkeys.cjs'); } catch (e) { console.log('Verify error:', e.message); }
}

main().catch(e => { console.error('FATAL:', e.message); process.exit(1); });
