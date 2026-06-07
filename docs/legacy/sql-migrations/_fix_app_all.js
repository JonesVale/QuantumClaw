const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Replace useTranslation import
c = c.replace("import { useTranslation } from 'react-i18next'", "import { useT } from '@/lib/use-t'");

// 2. Replace all useTranslation() usages
// Pattern: const { t, i18n } = useTranslation()  -> const { t } = useT()
// Pattern: const { t } = useTranslation()  -> const { t } = useT()
c = c.replace(/const \{ t, i18n \} = useTranslation\(\)/g, "const { t } = useT()");
c = c.replace(/const \{ i18n, t \} = useTranslation\(\)/g, "const { t } = useT()");
c = c.replace(/const \{ t \} = useTranslation\(\)/g, "const { t } = useT()");

// 3. Replace i18next import
c = c.replace("import i18next from 'i18next'", "");

// 4. Replace i18n.language -> language
c = c.replace(/i18n\.language/g, 'language');
c = c.replace(/i18n\.changeLanguage/g, 'changeLanguage');

// 5. Add language, changeLanguage, langs to AppHeader useT
c = c.replace(
  'function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t } = useT()',
  'function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t, language, changeLanguage, langs } = useT()'
);

// 6. Remove dbLanguages block + typeToCode + old switchLanguage
const dbStart = c.indexOf('  // Fetch available languages from DB for language selector');
const dbEnd = c.indexOf('  const cycleTheme = useCallback', dbStart);
if (dbStart >= 0 && dbEnd > dbStart) {
  const newSwitch = '  const switchLanguage = useCallback((langType: string) => {\n    changeLanguage(langType)\n  }, [changeLanguage])\n\n';
  c = c.substring(0, dbStart) + newSwitch + c.substring(dbEnd);
}

// 7. Replace dbLanguages ternary with langs.map
// Pattern: dbLanguages.length > 0 ? dbLanguages.map(lang => ( ... )) : ( ... )
// Find the exact range
const ternaryStart = c.indexOf('dbLanguages.length > 0 ? dbLanguages.map(lang =>');
const ternaryEnd = c.indexOf(')}', ternaryStart) + 2;  // end of the map
// Remove the : ( fallback part
const fallbackStart = c.indexOf(' : (', ternaryEnd);
if (fallbackStart >= 0) {
  // Find matching ')'
  let depth = 1;
  let i = fallbackStart + 4;
  while (depth > 0 && i < c.length) {
    if (c[i] === '(') depth++;
    if (c[i] === ')') depth--;
    i++;
  }
  c = c.substring(0, fallbackStart) + c.substring(i);
}

// Replace dbLanguages with langs
c = c.split('dbLanguages.length > 0 ? dbLanguages.map(lang =>').join('langs.map(lang =>');

// 8. Clean up: remove empty lines from removed blocks
c = c.replace(/\n{3,}/g, '\n\n');

fs.writeFileSync(path, c, 'utf-8');

// Verify
const v = fs.readFileSync(path, 'utf-8');
const checks = [
  v.includes('useT'),
  !v.includes('useTranslation'),
  v.includes('language, changeLanguage, langs'),
  !v.includes('dbLanguages'),
  !v.includes('typeToCode'),
  v.includes('changeLanguage(langType)'),
  !v.includes('switchLanguage(\'中文简体\')'),
  !v.includes('switchLanguage(\'English\')'),
];
console.log('Checks:', JSON.stringify(checks));
if (checks.every(Boolean)) console.log('ALL PASSED');
else console.log('FAILED: checks ' + checks.map((c,i) => i).filter(i => !checks[i]).join(','));
