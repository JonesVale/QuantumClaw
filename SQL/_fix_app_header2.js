const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
let c = fs.readFileSync(path, 'utf-8');

// File has \r\n. Use CRLF-aware replacement
const oldStr = '  const { t } = useT()\r\n\r\n  const { resolvedTheme, setTheme, theme } = useTheme()';
const newStr = '  const { t, language, changeLanguage, langs } = useT()\r\n\r\n  const { resolvedTheme, setTheme, theme } = useTheme()';

if (c.includes(oldStr)) {
  c = c.split(oldStr).join(newStr);
  fs.writeFileSync(path, c, 'utf-8');
  
  const check = fs.readFileSync(path, 'utf-8');
  if (check.includes('language, changeLanguage, langs')) {
    console.log('FIXED');
  } else {
    console.log('Still missing');
  }
} else {
  console.log('String not found - checking actual content...');
  const idx = c.indexOf('const { t } = useT()');
  if (idx >= 0) {
    const snippet = c.substring(idx, idx + 120)
      .replace(/\r/g, '\\r')
      .replace(/\n/g, '\\n');
    console.log('Actual: ' + snippet);
  }
}
