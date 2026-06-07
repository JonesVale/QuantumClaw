const fs = require('fs');

// Fix remaining i18next references using Node.js (avoid PowerShell encoding issues)
const fixes = [
  {
    path: 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/lib/api.ts',
    patches: [
      ["i18next.t('Session expired!')", "'Session expired!'"],
      ["import i18next from 'i18next'\n", ""],
    ]
  },
  {
    path: 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/lib/tlanguages.ts',
    patches: [
      ["import i18next from 'i18next'\n", ""],
      ["i18next.getResourceBundle(code, 'translation')", "{}"],
      ["i18next.addResourceBundle(code, 'translation', merged, true, true)", ""],
    ]
  },
  {
    path: 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx',
    patches: [
      ["localStorage.setItem('i18nextLng', code)", ""],
    ]
  },
  {
    path: 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/index.tsx',
    patches: [
      ["localStorage.setItem('i18nextLng', code)", ""],
    ]
  },
  {
    path: 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/i18n/config.ts',
    patches: []
  },
];

for (const fix of fixes) {
  let c = fs.readFileSync(fix.path, 'utf-8');
  const orig = c;
  for (const [old, replace] of fix.patches) {
    c = c.split(old).join(replace);
  }
  if (c !== orig) {
    fs.writeFileSync(fix.path, c, 'utf-8');
    console.log(`✓ ${fix.path.split('/').pop()}`);
  } else if (fix.patches.length > 0) {
    console.log(`  ${fix.path.split('/').pop()} - no changes needed`);
  }
}

// Check and fix main.tsx
const mainPath = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/main.tsx';
let main = fs.readFileSync(mainPath, 'utf-8');
const mainOrig = main;
main = main.replace("import i18next from 'i18next'\n", "");
main = main.replace("import './i18n/config'\n", "");
main = main.replace("i18next.t('Content not modified!')", "'Content not modified!'");
main = main.replace("i18next.t('Session expired!')", "'Session expired!'");
main = main.replace("i18next.t('Internal Server Error!')", "'Internal Server Error!'");
if (main !== mainOrig) {
  fs.writeFileSync(mainPath, main, 'utf-8');
  console.log('✓ main.tsx');
}

// Replace i18n/config.ts content entirely
const configPath = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/i18n/config.ts';
fs.writeFileSync(configPath, `// T_Languages - replaced i18next
// Translations loaded dynamically via useT() hook
export {}
`, 'utf-8');
console.log('✓ i18n/config.ts (replaced)');

console.log('Done');
