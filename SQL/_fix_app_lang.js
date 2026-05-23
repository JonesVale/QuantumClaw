const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Add language + changeLanguage + langs to AppHeader useT
c = c.replace(
  'function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t } = useT()',
  'function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {\n\n  const { t, language, changeLanguage, langs } = useT()'
);

// 2. Remove typeToCode mapping and switchLanguage
const oldSwitch = '  // Map T_Languages type to i18next code\n  const typeToCode: Record<string, string> = {\n    \'中文简体\': \'zh-CN\',\n    \'中文繁体\': \'zh-TW\',\n    \'English\': \'en\',\n    \'Français\': \'fr\',\n    \'日本語\': \'ja\',\n    \'Русский\': \'ru\',\n    \'Tiếng Việt\': \'vi\',\n  }\n\n  const switchLanguage = useCallback((langType: string) => {\n    const code = typeToCode[langType] || langType\n    changeLanguage(code)\n    \n  }, [language])';

const newSwitch = '  const switchLanguage = useCallback((langType: string) => {\n    changeLanguage(langType)\n  }, [changeLanguage])';

c = c.replace(oldSwitch, newSwitch);

// 3. Replace full dbLanguages block with using langs from useT
const oldDb = '  // Fetch available languages from DB for language selector\n  const [dbLanguages, setDbLanguages] = useState<string[]>([])\n  useEffect(() => {\n    fetch(\'/api/languages\')\n      .then(r => r.json())\n      .then(data => {\n        if (data.success && Array.isArray(data.data)) {\n          setDbLanguages(data.data.map((l: { languages_type: string }) => l.languages_type))\n        }\n      })\n      .catch(() => {})\n  }, [])';

c = c.replace(oldDb, '');

// 4. Replace dbLanguages usage with langs
c = c.split('dbLanguages.length > 0 ? dbLanguages.map(lang =>').join('langs.map(lang =>');

// 5. Remove hardcoded dropdown items
const oldHardcoded = '                <DropdownMenuItem onClick={() => switchLanguage(\'中文简体\')}>\n                  🇨🇳 中文简体\n                </DropdownMenuItem>\n                <DropdownMenuItem onClick={() => switchLanguage(\'English\')}>\n                  🇬🇧 English\n                </DropdownMenuItem>';

c = c.replace(oldHardcoded, '');

// 6. Clean up comment
c = c.replace('  // Fetch available languages from DB for language selector', '');

fs.writeFileSync(path, c, 'utf-8');
console.log('Done - language system fixed');
