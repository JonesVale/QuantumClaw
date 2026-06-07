const fs = require('fs');
const path = require('path');

// Extract all t() keys from source code
const dir = 'web/default/src';
const tKeys = new Set();
function scanDir(d) {
  for (const e of fs.readdirSync(d, {withFileTypes:true})) {
    const fp = path.join(d, e.name);
    if (e.isDirectory() && !e.name.startsWith('.')) scanDir(fp);
    else if (e.isFile() && (e.name.endsWith('.tsx') || e.name.endsWith('.ts'))) {
      const c = fs.readFileSync(fp, 'utf-8');
      const re = /t\(['"]([^'"]+)['"]\)/g;
      let m;
      while ((m = re.exec(c)) !== null) tKeys.add(m[1]);
    }
  }
}
scanDir(dir);

const all = [...tKeys].sort();
console.log('=== TOTAL t() KEYS IN CODE: ' + all.length + ' ===');

// Check against existing JSON
function rj(fp) { let r=fs.readFileSync(fp,'utf-8'); if(r.charCodeAt(0)===0xFEFF)r=r.slice(1); return JSON.parse(r); }
const en = rj('web/default/src/i18n/en.json');
const zh = rj('web/default/src/i18n/zh-CN.json');

const notInEn = all.filter(k => en[k] === undefined);
const notInZh = all.filter(k => zh[k] === undefined);

console.log('\nKeys in code but NOT in en.json: ' + notInEn.length);
if (notInEn.length > 0) notInEn.forEach(k => console.log('  MISS en: ' + k));

console.log('\nKeys in code but NOT in zh-CN.json: ' + notInZh.length);
if (notInZh.length > 0 && notInZh.length <= 10) notInZh.forEach(k => console.log('  MISS zh: ' + k));
else if (notInZh.length > 0) {
  notInZh.slice(0,10).forEach(k => console.log('  MISS zh: ' + k));
  console.log('  ... and ' + (notInZh.length-10) + ' more');
}

// Extra keys in JSON but not in code
const inJsonNotCode = Object.keys(en).filter(k => !all.includes(k) && !k.startsWith('/'));
console.log('\nKeys in en.json but NOT in code: ' + inJsonNotCode.length);

console.log('\n=== COVERAGE ===');
console.log('Code keys covered by en.json: ' + (all.length - notInEn.length) + '/' + all.length);
console.log('Code keys covered by zh-CN.json: ' + (all.length - notInZh.length) + '/' + all.length);
