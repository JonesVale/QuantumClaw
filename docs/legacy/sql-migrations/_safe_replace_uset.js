const fs = require('fs');
const path = require('path');
const SRC = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src';

function processFile(filePath) {
  let lines = fs.readFileSync(filePath, 'utf-8').split('\n');
  const orig = [...lines];
  let changed = false;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    let newLine = line;

    // Replace import line
    if (/^import \{ useTranslation \} from 'react-i18next'/.test(line)) {
      newLine = line.replace("useTranslation", "useT").replace("'react-i18next'", "'@/lib/use-t'");
    }
    // Handle: import { useTranslation, Trans } from ...
    else if (/^import \{ useTranslation, \w+ \} from 'react-i18next'/.test(line)) {
      // Replace with separate imports
      const extra = line.match(/useTranslation,\s*(\w+)/)?.[1];
      newLine = `import { useT } from '@/lib/use-t'\nimport { ${extra} } from 'react-i18next'`;
    }
    // Replace const { t, i18n } = useTranslation()
    else if (/^const \{ t, i18n \} = useTranslation\(\)/.test(line)) {
      newLine = "const { t } = useT()";
    }
    // Replace const { t } = useTranslation()
    else if (/^const \{ t \} = useTranslation\(\)/.test(line)) {
      newLine = "const { t } = useT()";
    }
    // Replace i18next imports (main.tsx)
    else if (/^import i18next from 'i18next'/.test(line)) {
      newLine = line.replace(/^import i18next from 'i18next'/, '');
      console.log(`  Remove i18next import: ${line}`);
    }

    if (newLine !== line) {
      lines[i] = newLine;
      changed = true;
    }
  }

  if (changed) {
    const out = lines.join('\n');
    // Quick validation: the file should still have export keyword
    if (!/export/.test(out) && !/^\/\*\*/.test(filePath) && !filePath.endsWith('config.ts')) {
      console.log(`⚠️  WARNING: ${filePath} may have lost exports`);
    }
    fs.writeFileSync(filePath, out, 'utf-8');
    console.log(`✓ ${path.relative(SRC, filePath)}`);
    return true;
  }
  return false;
}

function walk(dir) {
  let count = 0;
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory() && entry.name !== 'node_modules') {
      count += walk(full);
    } else if (entry.isFile() && /\.(tsx|ts)$/.test(entry.name)) {
      if (processFile(full)) count++;
    }
  }
  return count;
}

console.log('Replacing useTranslation with useT...');
const c = walk(SRC);
console.log(`\nModified ${c} files`);
