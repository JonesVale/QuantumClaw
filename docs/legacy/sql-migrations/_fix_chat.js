const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/(auth)/chat.tsx';
let c = fs.readFileSync(path, 'utf-8');
const orig = c;

// Add border-0 shadow-none to Cards
c = c.replace(/<Card\s+className="([^"]*)"/g, (m, cls) => {
  if (cls.includes('border-0')) return m;
  return '<Card className="border-0 shadow-none ' + cls + '"';
});

// Remove gradient backgrounds
c = c.replace(
  /bg-gradient-to-br from-slate-50 via-white to-blue-50\/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950\/50/g,
  ''
);
c = c.replace(/border-r bg-card\/50 backdrop-blur-sm/g, '');

if (c !== orig) {
  fs.writeFileSync(path, c, 'utf-8');
  console.log('chat.tsx fixed');
} else {
  console.log('chat.tsx no changes');
}
