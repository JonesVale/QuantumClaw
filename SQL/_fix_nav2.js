const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. NavItem: add loginRequired
c = c.replace(
  '  labelKey: string\n\n  adminOnly?: boolean\n\n}',
  '  labelKey: string\n\n  adminOnly?: boolean\n  loginRequired?: boolean\n\n}'
);

// 2. isAdmin line: add isLoggedIn below
c = c.replace(
  "  const isAdmin = auth.user?.role === 100\n\n\n\n\n\n  const isActive = useCallback(",
  "  const isAdmin = auth.user?.role === 100\n  const isLoggedIn = !!auth.user\n\n\n\n\n  const isActive = useCallback("
);

// 3. renderNavItem: add loginRequired check
c = c.replace(
  "    if (item.adminOnly && !isAdmin) return null\n\n\n\n    const Icon = item.icon\n    const active = isActive(item.path)",
  "    if (item.adminOnly && !isAdmin) return null\n    if (item.loginRequired && !isLoggedIn) return null\n\n\n\n    const Icon = item.icon\n    const active = isActive(item.path)"
);

// 4. Mark login-required items (add comma before, to preserve trailing comma)
const loginItems = [
  "/dashboard", "/keys", "/logs", "/monitoring", "/reseller", "/reseller-keys",
  "/profile", "/wallet", "/billing", "/checkin", "/subscription", "/api-docs"
];

for (const p of loginItems) {
  // Match: { path: '/xxx', icon: Xxx, labelKey: 'Xxx' },
  // Replace: { path: '/xxx', icon: Xxx, labelKey: 'Xxx', loginRequired: true },
  const regex = new RegExp(`\\{ path: '${p}', icon: (\\w+), labelKey: '([^']+)' \\},`);
  c = c.replace(regex, `{ path: '${p}', icon: $1, labelKey: '$2', loginRequired: true },`);
}

// Verify - count occurrences
const count = (c.match(/loginRequired/g) || []).length;
const commaIssue = (c.match(/loginRequired: true \}/g) || []).length;

fs.writeFileSync(path, c, 'utf-8');
console.log(`Done: ${count} loginRequired found, ${commaIssue} without comma`);
