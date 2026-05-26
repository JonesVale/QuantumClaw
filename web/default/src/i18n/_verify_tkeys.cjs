/**
 * _verify_tkeys.cjs — 验证所有语言的翻译覆盖率
 * 输出每个语言的已翻译/总键数和百分比
 * 在 _fill_translations 运行完毕后自动调用
 * 
 * 用法: node _verify_tkeys.cjs
 */
const fs = require('fs');
const path = require('path');

const I18N_DIR = __dirname;

const LANGS = [
  'zh-CN', 'en', 'zh-TW', 'ja', 'ko', 'ru', 'vi',
  'fr', 'es', 'de', 'it', 'pt', 'nl', 'tr',
  'th', 'ar', 'hi', 'id'
];

function getCoverage(code) {
  const f = path.join(I18N_DIR, code + '.json');
  if (!fs.existsSync(f)) return { total: 0, translated: 0, pct: 0 };
  let raw = fs.readFileSync(f, 'utf8');
  if (raw.charCodeAt(0) === 0xFEFF) raw = raw.slice(1);
  const data = JSON.parse(raw);
  const entries = Object.entries(data);
  const total = entries.length;
  const translated = entries.filter(([k, v]) => v && v !== k).length;
  // API paths and numbers that are intentionally identity
  const isCN = code === 'zh-CN';
  const exempt = entries.filter(([k, v]) => v === k && (isCN || k.startsWith('/') || k.match(/^\d+$/) || k === '-' || k === ';')).length;
  const meaningful = total - exempt;
  const done = translated;
  return {
    total,
    translated,
    meaningful,
    done,
    pct: meaningful > 0 ? ((done / meaningful) * 100).toFixed(1) : '100.0',
    identity: total - translated,
    exempt
  };
}

console.log('=== 语言翻译覆盖率 ===\n');
console.log('语言代码'.padEnd(12) + '总量'.padEnd(6) + '已翻译'.padEnd(8) + '无需翻译'.padEnd(8) + '待翻译'.padEnd(8) + '百分比');
console.log('-'.repeat(50));

let totalDone = 0;
let totalMeaningful = 0;

for (const code of LANGS) {
  const c = getCoverage(code);
  const need = c.meaningful - c.done;
  totalDone += c.done;
  totalMeaningful += c.meaningful;
  const status = c.pct === '100.0' ? '✅' : '❌';
  console.log(`${code.padEnd(12)}${String(c.total).padEnd(6)}${String(c.translated).padEnd(8)}${String(c.exempt).padEnd(8)}${String(Math.max(0, need)).padEnd(8)}${c.pct}% ${status}`);
}

console.log('-'.repeat(50));
const overall = totalMeaningful > 0 ? ((totalDone / totalMeaningful) * 100).toFixed(1) : 'N/A';
console.log(`总计: ${totalDone}/${totalMeaningful} = ${overall}%`);

// Return exit code based on all languages being 100%
const allDone = LANGS.every(code => {
  const c = getCoverage(code);
  return c.done >= c.meaningful;
});

if (allDone) {
  console.log('\n✅ 全部语言 100% 覆盖');
  process.exit(0);
} else {
  console.log('\n❌ 部分语言仍需翻译');
  process.exit(1);
}
