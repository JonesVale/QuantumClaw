const fs = require('fs');

// Step 1: Fix main.tsx - remove i18next import, replace direct t() calls
const mainPath = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/main.tsx';
let c = fs.readFileSync(mainPath, 'utf-8');
c = c.replace("import i18next from 'i18next'\n", '');
c = c.replace("import './i18n/config'\n", '');
c = c.replace("i18next.t('Content not modified!')", "'Content not modified!'");
c = c.replace("i18next.t('Session expired!')", "'Session expired!'");
c = c.replace("i18next.t('Internal Server Error!')", "'Internal Server Error!'");
fs.writeFileSync(mainPath, c, 'utf-8');
console.log('main.tsx: removed i18next');

// Step 2: Replace i18n/config.ts with placeholder
const configPath = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/i18n/config.ts';
fs.writeFileSync(configPath, `// T_Languages config - replaced i18next
// Translations loaded dynamically via useT() hook
export {}
`, 'utf-8');
console.log('i18n/config.ts: replaced with stub');

// Step 3: Remove package dependencies later - handled in package.json
console.log('Done - rebuild to verify');
