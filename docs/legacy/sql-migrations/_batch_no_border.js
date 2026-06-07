const fs = require('fs');
const path = require('path');
const SRC = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes';

const pages = ['playground.tsx', 'quantum.tsx', 'fusion.tsx', 'enterprise.tsx', 'apps.tsx', 
  // Also fix chat
  '(auth)/chat.tsx'
];

for (const page of pages) {
  const fp = path.join(SRC, page);
  if (!fs.existsSync(fp)) { console.log(`SKIP ${page}`); continue; }
  
  let c = fs.readFileSync(fp, 'utf-8');
  const orig = c;
  
  // 1. Add border-0 shadow-none to ALL <Card openings that don't already have it
  c = c.replace(/<Card\s+className="([^"]*)"/g, (match, cls) => {
    if (cls.includes('border-0')) return match;
    // Add border-0 and shadow-none
    return `<Card className="border-0 shadow-none ${cls}"`;
  });
  // Also handle <Card className={' ... '} patterns (template literals)
  
  // 2. Remove bg-gradient backgrounds
  c = c.replace(/bg-gradient-to-br from-slate-50 via-white to-blue-50\/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950\/50/g, '');
  
  // 3. Remove border-r from sidebar containers
  c = c.replace(/border-r bg-card\/50 backdrop-blur-sm/g, '');
  
  // 4. Fix gradient text in h1/h2 headers (already should be there)
  
  if (c !== orig) {
    fs.writeFileSync(fp, c, 'utf-8');
    console.log(`✅ ${page} - border-0 added`);
  } else {
    console.log(`  ${page} - no changes`);
  }
}
