const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Add language, changeLanguage, langs to AppHeader's useT
c = c.replace(
  "function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t } = useT()",
  "function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t, language, changeLanguage, langs } = useT()"
);

// 2. Remove dbLanguages fetch and typeToCode mapping entirely
const oldBlock = `  // Fetch available languages from DB for language selector
  const [dbLanguages, setDbLanguages] = useState<string[]>([])
  useEffect(() => {
    fetch('/api/languages')
      .then(r => r.json())
      .then(data => {
        if (data.success && Array.isArray(data.data)) {
          setDbLanguages(data.data.map((l: { languages_type: string }) => l.languages_type))
        }
      })
      .catch(() => {})
  }, [])

  // Map T_Languages type to i18next code
  const typeToCode: Record<string, string> = {
    '中文简体': 'zh-CN',
    '中文繁体': 'zh-TW',
    'English': 'en',
    'Fran\u00e7ais': 'fr',
    '\u65e5\u672c\u8a9e': 'ja',
    '\u0420\u0443\u0441\u0441\u043a\u0438\u0439': 'ru',
    'Ti\u1ebfng Vi\u1ec7t': 'vi',
  }

  const switchLanguage = useCallback((langType: string) => {
    const code = typeToCode[langType] || langType
    changeLanguage(code)
    
  }, [language])`;

// Try without the unicode escapes
const oldBlockSimple = `  // Fetch available languages from DB for language selector
  const [dbLanguages, setDbLanguages] = useState<string[]>([])
  useEffect(() => {
    fetch('/api/languages')
      .then(r => r.json())
      .then(data => {
        if (data.success && Array.isArray(data.data)) {
          setDbLanguages(data.data.map((l: { languages_type: string }) => l.languages_type))
        }
      })
      .catch(() => {})
  }, [])

  // Map T_Languages type to i18next code
  const typeToCode: Record<string, string> = {
    '中文简体': 'zh-CN',
    '中文繁体': 'zh-TW',
    'English': 'en',
    'Français': 'fr',
    '日本語': 'ja',
    'Русский': 'ru',
    'Tiếng Việt': 'vi',
  }

  const switchLanguage = useCallback((langType: string) => {
    const code = typeToCode[langType] || langType
    changeLanguage(code)
    
  }, [language])`;

const newBlock = `  const switchLanguage = useCallback((langType: string) => {
    changeLanguage(langType)
  }, [changeLanguage])`;

if (c.includes(oldBlockSimple)) {
  c = c.replace(oldBlockSimple, newBlock);
  console.log('Replaced full block (simple)');
} else if (c.includes(oldBlock)) {
  c = c.replace(oldBlock, newBlock);
  console.log('Replaced full block (unicode)');
} else {
  // Try to find the start and end by line number
  console.log('Block not found, trying alternative...');
  // Find "Fetch available languages" and remove until "const cycleTheme"
  const start = c.indexOf('  // Fetch available languages from DB for language selector');
  const end = c.indexOf('  const cycleTheme', start);
  if (start >= 0 && end > start) {
    const removed = c.substring(start, end);
    console.log(`Removing block: ${removed.slice(0, 60)}...`);
    c = c.substring(0, start) + newBlock + c.substring(end);
    console.log('Replaced via manual index');
  }
}

// 3. Replace dbLanguages with langs in the dropdown
c = c.split('dbLanguages.length > 0 ? dbLanguages.map(lang =>').join('langs.map(lang =>');

// 4. Remove hardcoded fallback items
c = c.replace(
  "                <DropdownMenuItem onClick={() => switchLanguage('中文简体')}>\n                  🇨🇳 中文简体\n                </DropdownMenuItem>\n                <DropdownMenuItem onClick={() => switchLanguage('English')}>\n                  🇬🇧 English\n                </DropdownMenuItem>",
  ""
);

fs.writeFileSync(path, c, 'utf-8');
console.log('Done');
