const fs = require('fs');
const path = require('path');
const dir = 'web/default/src';
const tKeys = new Set();
function scanDir(d) {
  const entries = fs.readdirSync(d, {withFileTypes:true});
  for (const e of entries) {
    const fp = path.join(d, e.name);
    if (e.isDirectory() && !e.name.startsWith('.')) scanDir(fp);
    else if (e.isFile() && (e.name.endsWith('.tsx') || e.name.endsWith('.ts'))) {
      const c = fs.readFileSync(fp, 'utf-8');
      const matches = c.matchAll(/t\(['"]([^'"]+)['"]\)/g);
      for (const m of matches) tKeys.add(m[1]);
    }
  }
}
scanDir(dir);
const arr = [...tKeys].sort();
console.log('=== Code t() keys: ' + arr.length + ' ===');
arr.forEach(k => console.log(k));
