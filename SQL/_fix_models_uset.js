const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/models.tsx';
let c = fs.readFileSync(path, 'utf-8');

// Replace import
c = c.replace(
  "import { useTranslation } from 'react-i18next'",
  "import { useT } from '@/lib/use-t'"
);

// Replace destructure
c = c.replace(
  "  const { t, i18n } = useTranslation()",
  "  const { t } = useT()"
);

// Remove i18n.language references
c = c.replace(
  "  const lang = codeToType[i18n.language] || 'English'",
  "  const lang = 'English'"
);

fs.writeFileSync(path, c, 'utf-8');
console.log('models.tsx: useTranslation -> useT');
