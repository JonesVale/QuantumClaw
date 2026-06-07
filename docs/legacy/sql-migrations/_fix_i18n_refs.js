const fs = require('fs');

const files = [
  'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx',
  'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/index.tsx',
  'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/(auth)/chat.tsx',
  'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/models.tsx',
];

for (const filePath of files) {
  let c = fs.readFileSync(filePath, 'utf-8');
  const orig = c;

  // Replace i18n.language → language (useT returns language)
  c = c.replace(/i18n\.language/g, 'language');

  // Replace i18n.changeLanguage → changeLanguage (useT returns changeLanguage)
  c = c.replace(/i18n\.changeLanguage/g, 'changeLanguage');

  // Replace standalone [i18n] → [language] (React useEffect deps)
  c = c.replace(/\[\s*i18n\s*\]/g, '[language]');

  // Replace codeToType[language] → just language (no more i18n code mapping)
  c = c.replace(/codeToType\[language\]/g, 'language');

  if (c !== orig) {
    fs.writeFileSync(filePath, c, 'utf-8');
    console.log(`✓ ${filePath.split('/').pop()}`);
  }
}
console.log('Done');
