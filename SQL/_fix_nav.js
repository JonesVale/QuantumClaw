const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Add loginRequired to NavItem interface
c = c.replace(
  '  adminOnly?: boolean\n\n}',
  '  adminOnly?: boolean\n  loginRequired?: boolean\n\n}'
);

// 2. Add isLoggedIn after isAdmin  
c = c.replace(
  "  const isAdmin = auth.user?.role === 100\n\n\n\n\n\n",
  "  const isAdmin = auth.user?.role === 100\n  const isLoggedIn = !!auth.user\n\n\n\n\n"
);

// 3. Add loginRequired check in renderNavItem
c = c.replace(
  "    if (item.adminOnly && !isAdmin) return null\n\n\n\n    const Icon = item.icon\n    const active = isActive(item.path)",
  "    if (item.adminOnly && !isAdmin) return null\n    if (item.loginRequired && !isLoggedIn) return null\n\n\n\n    const Icon = item.icon\n    const active = isActive(item.path)"
);

// 4. Mark loginRequired items
const loginItems = [
  "/dashboard", "/keys", "/logs", "/monitoring", "/reseller", "/reseller-keys",
  "/profile", "/wallet", "/billing", "/checkin", "/subscription", "/api-docs"
];

for (const p of loginItems) {
  // Match: { path: '/xxx', icon: Xxx, labelKey: 'Xxx' },
  // Replace with: { path: '/xxx', icon: Xxx, labelKey: 'Xxx', loginRequired: true },
  const regex = new RegExp(`(\\{ path: '${p}', icon: \\w+, labelKey: '[^']+')( \\},)`);
  c = c.replace(regex, '$1, loginRequired: true }');
}

fs.writeFileSync(path, c, 'utf-8');
console.log('app-layout.tsx updated');
