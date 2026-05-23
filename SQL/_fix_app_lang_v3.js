const fs = require('fs');
const { execSync } = require('child_process');

// Get the original file from git
const orig = execSync('git -C H:/AiData/openclaw/workspace/QuantumClaw show HEAD:web/default/src/components/layout/app-layout.tsx', { encoding: 'utf-8' });

let c = orig;

// 1. Add language, changeLanguage, langs to AppHeader's useT
c = c.replace(
  'function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t } = useT()',
  'function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t, language, changeLanguage, langs } = useT()'
);

// 2. Find and replace the entire dbLanguages + typeToCode + switchLanguage block
const start = c.indexOf('  // Fetch available languages from DB for language selector');
const end = c.indexOf('  const cycleTheme = useCallback', start);
if (start >= 0 && end > start) {
  const newBlock = '  const switchLanguage = useCallback((langType: string) => {\n    changeLanguage(langType)\n  }, [changeLanguage])\n\n';
  c = c.substring(0, start) + newBlock + c.substring(end);
}

// 3. Replace dbLanguages with langs in dropdown
c = c.split('dbLanguages.length > 0 ? dbLanguages.map(lang =>').join('langs.map(lang =>');

// 4. Remove hardcoded fallback dropdown items - find by pattern
const cnItem = '                <DropdownMenuItem onClick={() => switchLanguage(\'中文简体\')}>';
const enItem = '                <DropdownMenuItem onClick={() => switchLanguage(\'English\')}>';
const cnIdx = c.indexOf(cnItem);
const enIdx = c.indexOf(enItem);
if (cnIdx >= 0) {
  // Find the end of this item (2 lines after)
  const cnEnd = c.indexOf('\n', c.indexOf('\n', cnIdx) + 1) + 1;
  c = c.substring(0, cnIdx) + c.substring(cnEnd);
}
if (enIdx >= 0) {
  const enEnd = c.indexOf('\n', c.indexOf('\n', enIdx) + 1) + 1;
  c = c.substring(0, enIdx) + c.substring(enEnd);
}

const outPath = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
fs.writeFileSync(outPath, c, 'utf-8');

// Verify by checking key patterns
const verify = fs.readFileSync(outPath, 'utf-8');
const checks = [
  verify.includes('language, changeLanguage, langs'),
  !verify.includes('dbLanguages'),
  !verify.includes('typeToCode'),
  verify.includes('switchLanguage'),
  !verify.includes('switchLanguage(\'中文简体\')'),
  !verify.includes('switchLanguage(\'English\')'),
  verify.includes('langs.map(lang =>'),
];
console.log('Checks passed: ' + checks.filter(Boolean).length + '/' + checks.length);
if (checks.every(Boolean)) console.log('ALL PASSED');
else console.log('SOME FAILED: ' + checks.map((c,i) => i + ':' + c).join(', '));
