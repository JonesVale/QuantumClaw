const http = require('http');
const fs = require('fs');

const LANGS = [
  ['fr', 'Fran\u00e7ais'], ['ja', '\u65e5\u672c\u8a9e'],
  ['ru', '\u0420\u0443\u0441\u0441\u043a\u0438\u0439'], ['vi', 'Ti\u1ebfng Vi\u1ec7t'],
  ['es', 'Espa\u00f1ol'], ['de', 'Deutsch'], ['it', 'Italiano'],
  ['pt', 'Portugu\u00eas'], ['nl', 'Nederlands'], ['tr', 'T\u00fcrk\u00e7e'],
  ['ko', '\ud55c\uad6d\uc5b4'], ['id', 'Bahasa Indonesia'],
  ['th', '\u0e44\u0e17\u0e22'], ['ar', '\u0627\u0644\u0639\u0631\u0628\u064a\u0629'],
  ['hi', '\u0939\u093f\u0928\u094d\u0926\u0940']
];

let raw = fs.readFileSync('en.json', 'utf8');
if (raw.charCodeAt(0) === 0xFEFF) raw = raw.slice(1);
const en = JSON.parse(raw);
const allKeys = Object.keys(en);

function getExisting(code) {
  const f = code + '.json';
  if (!fs.existsSync(f)) return {};
  let r = fs.readFileSync(f, 'utf8');
  if (r.charCodeAt(0) === 0xFEFF) r = r.slice(1);
  return JSON.parse(r);
}

function translate(texts, lang) {
  return new Promise((resolve, reject) => {
    const body = JSON.stringify({
      model: 'qwen3.5:4b',
      prompt: 'Translate to ' + lang + '. Return JSON array: ' + JSON.stringify(texts),
      stream: false,
      keep_alive: '60m',
      options: { temperature: 0.1 }
    });
    const req = http.request({ hostname: 'localhost', port: 11434, path: '/api/generate', method: 'POST',
      headers: { 'Content-Type': 'application/json' } }, res => {
      let d = '';
      res.on('data', c => d += c);
      res.on('end', () => {
        try {
          const resp = JSON.parse(d).response;
          const s = resp.indexOf('['), e = resp.lastIndexOf(']');
          if (s >= 0 && e > s + 1) resolve(JSON.parse(resp.substring(s, e + 1)));
          else reject('no json');
        } catch (err) { reject(err.message); }
      });
    });
    req.write(body); req.end();
  });
}

function seedDB(lang, data) {
  return new Promise((resolve) => {
    const entries = Object.entries(data).filter(([k, v]) => v && k).map(([k, v]) => ({ lcode: k, display: v }));
    const body = JSON.stringify({ languages_type: lang, entries });
    const req = http.request({ hostname: 'localhost', port: 3666, path: '/api/languages/seed-public', method: 'POST',
      headers: { 'Content-Type': 'application/json' } }, res => { let d='';res.on('data',c=>d+=c);res.on('end',()=>resolve());});
    req.write(body); req.end();
  });
}

const BATCH = 5;
const PF = '_progress.json';
let progress = fs.existsSync(PF) ? JSON.parse(fs.readFileSync(PF, 'utf8')) : {};

(async () => {
  for (const [code, name] of LANGS) {
    let result = getExisting(code);
    const have = Object.keys(result).length;
    const missing = allKeys.filter(k => !result[k]);
    if (missing.length === 0) { console.log(code + ': complete'); continue; }

    console.log('\n' + code + ': ' + have + '/' + allKeys.length + ' (' + name + '), need ' + missing.length);

    for (let i = 0; i < missing.length; i += BATCH) {
      const batch = missing.slice(i, i + BATCH);
      const vals = batch.map(k => en[k]);
      let ok = false;
      for (let a = 0; a < 3 && !ok; a++) {
        try {
          const t = await translate(vals, name);
          if (t) { batch.forEach((k, idx) => { if (t[idx]) result[k] = t[idx]; }); ok = true; }
        } catch (e) { await new Promise(r => setTimeout(r, 3000)); }
      }
      fs.writeFileSync(code + '.json', JSON.stringify(result, null, 2), 'utf8');
      progress[code] = i + BATCH;
      fs.writeFileSync(PF, JSON.stringify(progress));

      const done = Object.keys(result).length - have;
      if (i % 50 === 0 || i + BATCH >= missing.length)
        console.log('  ' + code + ': ' + done + '/' + missing.length);
    }

    console.log('  Seeding DB...');
    await seedDB(name, result);
    console.log('  DB done');
  }
  console.log('\nALL DONE');
})();
