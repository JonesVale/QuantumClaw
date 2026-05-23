// Batch redesign: strip Card/border from all remaining public pages
// Keep logic intact, replace layout with Braket-style divs
const fs = require('fs');
const path = require('path');
const SRC = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes';

const pages = [
  'playground.tsx', 'quantum.tsx', 'enterprise.tsx', 'apps.tsx', 'fusion.tsx'
];

for (const page of pages) {
  const fp = path.join(SRC, page);
  if (!fs.existsSync(fp)) { console.log(`SKIP ${page}: not found`); continue; }
  
  let c = fs.readFileSync(fp, 'utf-8');
  const orig = c;
  
  // 1. Replace Card imports with nothing (or keep just types)
  c = c.replace(/import \{ Card, CardContent, CardHeader, CardTitle, CardDescription \} from ['"]@\/components\/ui\/card['"]\n/g, '');
  c = c.replace(/import \{ Card, CardContent, CardHeader, CardTitle \} from ['"]@\/components\/ui\/card['"]\n/g, '');
  c = c.replace(/import \{ Card, CardHeader, CardContent \} from ['"]@\/components\/ui\/card['"]\n/g, '');
  c = c.replace(/import \{ Card, CardContent \} from ['"]@\/components\/ui\/card['"]\n/g, '');
  c = c.replace(/import \{ Card \} from ['"]@\/components\/ui\/card['"]\n/g, '');
  c = c.replace(/import \{ Badge \} from ['"]@\/components\/ui\/badge['"]\n/g, '');
  c = c.replace(/import \{ ScrollArea \} from ['"]@\/components\/ui\/scroll-area['"]\n/g, '');
  c = c.replace(/import \{ Skeleton \} from ['"]@\/components\/ui\/skeleton['"]\n/g, '');
  
  // 2. Replace <Card> wrappers with borderless divs
  // Match: <Card className="something"> ... </Card>
  // But we need to be careful with nested Cards
  // Strategy: replace opening/closing Card tags with div
  c = c.replace(/<Card\s*/g, '<div ');
  c = c.replace(/<\/Card>/g, '</div>');
  c = c.replace(/<CardContent\s*/g, '<div ');
  c = c.replace(/<\/CardContent>/g, '</div>');
  c = c.replace(/<CardHeader\s*/g, '<div ');
  c = c.replace(/<\/CardHeader>/g, '</div>');
  c = c.replace(/<CardTitle\s*/g, '<h3 ');
  c = c.replace(/<\/CardTitle>/g, '</h3>');
  c = c.replace(/<CardDescription\s*/g, '<p ');
  c = c.replace(/<\/CardDescription>/g, '</p>');
  
  // 3. Remove border/background classes from card-like elements
  // Replace Card className patterns that use borders
  c = c.replace(/className="([^"]*)border([^"]*)"(\s*\/>|>)/g, (match, before, after, closer) => {
    let cls = before + after;
    // Remove border-related classes
    cls = cls.replace(/border-\w+/g, '').replace(/bg-card/g, 'bg-background').replace(/shadow-\w+/g, '');
    cls = cls.replace(/\s+/g, ' ').trim();
    return 'className="' + cls + '"' + (closer === '/>' ? ' />' : '>');
  });
  
  // 4. Remove bg-gradient backgrounds (keep Braket-style clean)
  c = c.replace(/from-slate-50 via-white to-blue-50\/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950\/50/g, '');
  
  // 5. Remove border-r from sidebar-like elements
  c = c.replace(/border-r bg-card\/50 backdrop-blur-sm/g, '');
  
  if (c !== orig) {
    fs.writeFileSync(fp, c, 'utf-8');
    console.log(`✅ ${page} redesigned`);
  } else {
    console.log(`  ${page} no changes needed`);
  }
}

console.log('Done');
