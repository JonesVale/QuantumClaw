const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/pricing.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Remove unused imports
c = c.replace("import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'\n", "");
c = c.replace("import { Badge } from '@/components/ui/badge'\n", "");

// 2. Replace the return block
const rIdx = c.indexOf("  return (");
const endIdx = c.lastIndexOf(")");
// Find the component's closing }
const closeBrace = c.indexOf("\n}\n", endIdx);
const afterIdx = closeBrace > 0 ? closeBrace + 2 : c.length - 1;

const beforeRet = c.substring(0, rIdx);
const after = c.substring(afterIdx);

const newRet = `  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10 py-12">
        {/* Hero */}
        <div className="mb-12">
          <h1 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight">
            <span className="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
              {t('Model Pricing')}
            </span>
          </h1>
          <p className="text-lg text-muted-foreground mt-4 max-w-2xl leading-relaxed">
            {t('Browse pricing across providers')}
          </p>
        </div>

        {/* Filter pills */}
        <div className="flex flex-wrap gap-2 mb-10">
          <button onClick={() => setProviderFilter('all')}
            className={\`px-4 py-2 rounded-lg text-sm font-medium transition-all \${
              providerFilter === 'all' ? 'bg-primary text-primary-foreground shadow-sm' : 'bg-muted text-muted-foreground hover:bg-accent'
            }\`}>
            {t('All Providers')}
          </button>
          {providers.slice(0, 15).map(p => (
            <button key={p} onClick={() => setProviderFilter(p)}
              className={\`px-4 py-2 rounded-lg text-sm font-medium transition-all \${
                providerFilter === p ? 'bg-primary text-primary-foreground shadow-sm' : 'bg-muted text-muted-foreground hover:bg-accent'
              }\`}>
              {p}
            </button>
          ))}
          {providers.length > 15 && (
            <details className="group inline-block">
              <summary className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:bg-muted cursor-pointer">
                <span className="group-open:hidden">{t('More')} ({providers.length - 15})</span>
                <span className="hidden group-open:inline">{t('Less')}</span>
              </summary>
            </details>
          )}
        </div>

        {/* Pricing Cards by Provider */}
        <div className="space-y-10">
          {groupedByProvider.map(([provider, models]) => (
            <div key={provider}>
              <h2 className="text-xl font-semibold mb-4">{provider}</h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {models.map((m: any) => (
                  <div key={m.model} className="rounded-xl p-5 bg-card hover:shadow-md transition-all border-0">
                    <h3 className="font-semibold text-sm mb-1">{m.model}</h3>
                    <p className="text-[10px] text-muted-foreground uppercase tracking-wider mb-3">{m.provider}</p>
                    <div className="flex items-center justify-between text-xs">
                      <span className="inline-flex items-center gap-1">
                        <span className="font-medium text-emerald-600 dark:text-emerald-400">IN</span>
                        \${(m.input_price || 0).toFixed(4)}/1K
                      </span>
                      <span className="inline-flex items-center gap-1">
                        <span className="font-medium text-amber-600 dark:text-amber-400">OUT</span>
                        \${(m.output_price || 0).toFixed(4)}/1K
                      </span>
                      <span className="text-xs text-muted-foreground">{(m.context_window / 1000).toFixed(0)}K ctx</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );`;

c = beforeRet + newRet + after;

// Clean up empty lines
c = c.replace(/\n{3,}/g, '\n\n');

fs.writeFileSync(path, c, 'utf-8');
console.log('Pricing page redesigned');
