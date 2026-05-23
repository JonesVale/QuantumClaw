const fs = require('fs');
const path = require('path');
const SRC = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src';

function walk(dir) {
  let count = 0;
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory() && entry.name !== 'node_modules') {
      count += walk(full);
    } else if (entry.isFile() && /\.(tsx|ts)$/.test(entry.name)) {
      if (fixFile(full)) count++;
    }
  }
  return count;
}

function fixFile(filePath) {
  let c = fs.readFileSync(filePath, 'utf-8');
  const orig = c;

  // 1. Replace import line (at start of line)
  c = c.replace(/^import \{ useTranslation \} from 'react-i18next';?$/gm, "import { useT } from '@/lib/use-t'");

  // 2. Replace const { t, i18n } = useTranslation() (indented or not)
  c = c.replace(/^(\s*)const \{ t, i18n \} = useTranslation\(\);?$/gm, "$1const { t } = useT()");

  // 3. Replace const { t } = useTranslation() (indented or not)
  c = c.replace(/^(\s*)const \{ t \} = useTranslation\(\);?$/gm, "$1const { t } = useT()");

  // 4. Replace const { i18n, t } ... variant
  c = c.replace(/^(\s*)const \{ i18n, t \} = useTranslation\(\);?$/gm, "$1const { t } = useT()");

  // 5. Remove standalone i18next imports
  c = c.replace(/^import i18next from 'i18next';?$/gm, '');

  // 6. Remove import { useTranslation, Trans } from 'react-i18next'
  c = c.replace(/^import \{ useTranslation, Trans \} from 'react-i18next';?$/gm, "import { useT } from '@/lib/use-t'\nimport { Trans } from 'react-i18next'");

  if (c !== orig) {
    fs.writeFileSync(filePath, c, 'utf-8');
    const rel = path.relative(SRC, filePath);
    console.log(`  ${rel}`);
    return true;
  }
  return false;
}

console.log('Second pass: replacing useTranslation() usages...');
const count = walk(SRC);
console.log(`\nFixed ${count} files`);

// Final check
console.log('\nVerifying...');
let remaining = 0;
for (const entry of fs.readdirSync(SRC, { withFileTypes: true })) {
  // Just check a few key files
}
const all = require('child_process').execSync(
  `findstr /s /m "useTranslation" "${SRC}\\*.tsx" "${SRC}\\*.ts" 2>nul`,
  { encoding: 'utf-8' }
);
const files = all.split('\n').filter(Boolean).filter(f => !f.includes('node_modules'));
if (files.length > 0) {
  console.log(`⚠️  Remaining ${files.length} files with useTranslation:`);
  files.forEach(f => console.log(`  ${f}`));
} else {
  console.log('✅ All useTranslation references replaced!');
}
