const fs = require('fs');

// 1. Fix app-layout.tsx - add language + changeLanguage, remove typeToCode
const path1 = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
let c1 = fs.readFileSync(path1, 'utf-8');

// Add language + changeLanguage to the AppHeader useT() destructure (first useT inside function AppHeader)
// The AppHeader function has: const { t } = useT()
c1 = c1.replace(
  "function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t } = useT()",
  "function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t, language, changeLanguage } = useT()"
);

// Remove the typeToCode mapping (lines with 'typeToCode')
c1 = c1.replace(
  "  const typeToCode: Record<string, string> = {\n    '中文简体': 'zh-CN',\n    '中文繁体': 'zh-TW',\n    'English': 'en',\n    'Français': 'fr',\n    '日本語': 'ja',\n    'Русский': 'ru',\n    'Tiếng Việt': 'vi',\n  }\n\n",
  ""
);

// Fix switchLanguage: use langType directly (no typeToCode)
c1 = c1.replace(
  "    const code = typeToCode[langType] || langType\n    changeLanguage(code)",
  "    changeLanguage(langType)"
);

// Fix className: no typeToCode mapping
c1 = c1.replace(
  "className={language === (typeToCode[lang] || lang)",
  "className={language === lang"
);

fs.writeFileSync(path1, c1, 'utf-8');
console.log('✓ app-layout.tsx');

// 2. Fix index.tsx - add language + changeLanguage 
const path2 = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/index.tsx';
let c2 = fs.readFileSync(path2, 'utf-8');

c2 = c2.replace(
  "  const { t } = useT()",
  "  const { t, language, changeLanguage } = useT()"
);

// Fix the useState: use language directly (it's already the T_Languages type)
c2 = c2.replace(
  "  const [currentLang, setCurrentLang] = useState(language || 'en')",
  "  const [currentLang, setCurrentLang] = useState(language || 'English')"
);

// Fix changeLang function: just pass to changeLanguage
c2 = c2.replace(
  "    changeLanguage(code)\n    \n    setCurrentLang(code)",
  "    changeLanguage(code)\n    setCurrentLang(code)"
);

// Fix langName - find the language name differently (no typeToCode)
c2 = c2.replace(
  "  const langName = (Object.entries(typeToCode).find(([_, code]) => code === currentLang)?.[0]) || currentLang || 'English'",
  "  const langName = currentLang || 'English'"
);

// Fix date locale in news
c2 = c2.replace(
  "new Date(article.published_at).toLocaleDateString(language || 'zh-CN', { month: '2-digit', day: '2-digit' })",
  "new Date(article.published_at).toLocaleDateString('en-US', { month: '2-digit', day: '2-digit' })"
);

fs.writeFileSync(path2, c2, 'utf-8');
console.log('✓ index.tsx');

console.log('Done');
