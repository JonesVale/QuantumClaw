const fs = require('fs');
const path = require('path');
const SRC = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src';

let changed = 0;
function processFile(filePath) {
  let c = fs.readFileSync(filePath, 'utf-8');
  const original = c;

  // Skip files in node_modules
  if (filePath.includes('node_modules')) return;

  // Replace import
  c = c.replace(
    "import { useTranslation } from 'react-i18next'",
    "import { useT } from '@/lib/use-t'"
  );

  // Also handle: import { useTranslation, Trans } from ...
  c = c.replace(
    /import \{ useTranslation(?:,\s*\w+)? \} from 'react-i18next'/g,
    "import { useT } from '@/lib/use-t'"
  );

  // Replace: const { t, i18n } = useTranslation()  -> const { t } = useT()
  c = c.replace(
    /const \{ t,\s*i18n\s*\}[^;]*= useTranslation\(\)/g,
    "const { t } = useT()"
  );

  // Replace: const { t } = useTranslation()  -> const { t } = useT()
  c = c.replace(
    /const \{ t \}[^;]*= useTranslation\(\)/g,
    "const { t } = useT()"
  );

  // Replace: const [t, i18n] = useTranslation()  -> const { t } = useT()  (rare)
  c = c.replace(
    /const \[t[,\s]*\w+\] = useTranslation\(\)/g,
    "const { t } = useT()"
  );

  // Remove codeToType references if they exist (will be added back if needed)
  c = c.replace(
    /import \{ codeToType \} from ['"]@\/lib\/tlanguages['"]\n/g,
    ''
  );

  // Replace i18n.language usage with 'English' (temporary)
  c = c.replace(/codeToType\[i18n\.language\]/g, "'English'");
  c = c.replace(/i18n\.language/g, "'English'");

  // Replace i18n.changeLanguage -> setLanguage (if useT returns it)
  // This is trickier - need to handle async
  c = c.replace(/i18n\.changeLanguage\(/g, "// i18n.changeLanguage(");

  if (c !== original) {
    fs.writeFileSync(filePath, c, 'utf-8');
    changed++;
  }
}

function walk(dir) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory() && entry.name !== 'node_modules') {
      walk(full);
    } else if (entry.isFile() && /\.(tsx|ts)$/.test(entry.name)) {
      processFile(full);
    }
  }
}

walk(SRC);
console.log(`Processed ${changed} files changed`);
