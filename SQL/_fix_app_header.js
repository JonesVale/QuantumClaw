const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
let c = fs.readFileSync(path, 'utf-8');

// Simple string replacement to add language + changeLanguage + langs
const oldStr = '  const { t } = useT()\n\n  const { resolvedTheme, setTheme, theme } = useTheme()';
const newStr = '  const { t, language, changeLanguage, langs } = useT()\n\n  const { resolvedTheme, setTheme, theme } = useTheme()';
c = c.split(oldStr).join(newStr);

fs.writeFileSync(path, c, 'utf-8');

if (fs.readFileSync(path, 'utf-8').includes('language, changeLanguage, langs')) {
  console.log('FIXED: language vars added to AppHeader');
} else {
  console.log('FAILED');
}
