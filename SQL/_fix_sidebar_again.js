const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/models.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Remove seriesNames useMemo
c = c.replace(
  "  \n  const seriesNames = useMemo(() => {\n    if (catalog.length === 0) return []\n    const set = new Set<string>()\n    for (const m of catalog) {\n      if (m.series) set.add(m.series)\n    }\n    return Array.from(set).sort()\n  }, [catalog.length]) // eslint-disable-line\n\n",
  "\n\n"
);

// 2. Replace single providers useMemo with AI + Quantum providers
const oldProviders = `  const providers = useMemo(() => {
    if (catalog.length === 0) return []
    const map = new Map<string, number>()
    for (const m of catalog) {
      const p = m.provider || 'Unknown'
      map.set(p, (map.get(p) || 0) + 1)
    }
    return Array.from(map.entries()).sort((a, b) => b[1] - a[1])
  }, [catalog.length]) // eslint-disable-line`;

const newProviders = `  const aiProviders = useMemo(() => {
    if (catalog.length === 0) return []
    const map = new Map<string, number>()
    for (const m of catalog) {
      if (m.use_case === 'quantum') continue
      const p = m.provider || 'Unknown'
      if (p === 'Unknown' || p.startsWith('~')) continue
      map.set(p, (map.get(p) || 0) + 1)
    }
    return Array.from(map.entries()).sort((a, b) => b[1] - a[1])
  }, [catalog.length]) // eslint-disable-line

  const quantumProviders = useMemo(() => {
    if (catalog.length === 0) return []
    const map = new Map<string, number>()
    for (const m of catalog) {
      if (m.use_case !== 'quantum') continue
      const p = m.provider || 'Unknown'
      if (p === 'Unknown' || p.startsWith('~')) continue
      map.set(p, (map.get(p) || 0) + 1)
    }
    return Array.from(map.entries()).sort((a, b) => b[1] - a[1])
  }, [catalog.length]) // eslint-disable-line`;

c = c.replace(oldProviders, newProviders);

// 3. Replace the single Provider sidebar section with AI + Quantum sections
const oldSidebar = `          {/* Provider */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, provider: !c.provider}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('Provider')}</span>
              {collapsed.provider ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.provider && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setProviderFilter('')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', !providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                {t('All Providers')} <span className="text-xs text-muted-foreground">({catalog.length})</span>
              </button>
              {providers.slice(0, 15).map(([p, cnt]) => (
                <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                  <span>{p}</span>
                  <span className="text-xs text-muted-foreground">{cnt}</span>
                </button>
              ))}
              {providers.length > 15 && (
                <details className="group">
                  <summary className="w-full text-left px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted/50 rounded cursor-pointer list-none">
                    <span className="group-open:hidden">{t('Show more...')} ({providers.length - 15})</span>
                    <span className="hidden group-open:inline">{t('Show less')}</span>
                  </summary>
                  <div className="space-y-0.5">
                    {providers.slice(15).map(([p, cnt]) => (
                      <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                        <span>{p}</span>
                        <span className="text-xs text-muted-foreground">{cnt}</span>
                      </button>
                    ))}
                  </div>
                </details>
              )}
            </div>}
          </section>
          {/* Categories */}`;

const newSidebar = `          {/* AI Providers */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, provider: !c.provider}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('AI Providers')}</span>
              {collapsed.provider ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.provider && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setProviderFilter('')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', !providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                {t('All Providers')} <span className="text-xs text-muted-foreground">({catalog.filter(m => m.use_case !== 'quantum').length})</span>
              </button>
              {aiProviders.slice(0, 15).map(([p, cnt]) => (
                <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                  <span>{p}</span>
                  <span className="text-xs text-muted-foreground">{cnt}</span>
                </button>
              ))}
              {aiProviders.length > 15 && (
                <details className="group">
                  <summary className="w-full text-left px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted/50 rounded cursor-pointer list-none">
                    <span className="group-open:hidden">{t('Show more...')} ({aiProviders.length - 15})</span>
                    <span className="hidden group-open:inline">{t('Show less')}</span>
                  </summary>
                  <div className="space-y-0.5">
                    {aiProviders.slice(15).map(([p, cnt]) => (
                      <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                        <span>{p}</span>
                        <span className="text-xs text-muted-foreground">{cnt}</span>
                      </button>
                    ))}
                  </div>
                </details>
              )}
            </div>}
          </section>
          {/* Quantum Providers */}
          {quantumProviders.length > 0 && <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, provider: !c.provider}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('Quantum Providers')}</span>
              {collapsed.provider ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.provider && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setProviderFilter('')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', !providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                {t('All Quantum')} <span className="text-xs text-muted-foreground">({catalog.filter(m => m.use_case === 'quantum').length})</span>
              </button>
              {quantumProviders.map(([p, cnt]) => (
                <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                  <span>{p}</span>
                  <span className="text-xs text-muted-foreground">{cnt}</span>
                </button>
              ))}
            </div>}
          </section>}
          {/* Categories */}`;

c = c.replace(oldSidebar, newSidebar);

// 4. Remove Series section if present
c = c.replace(
  /\n\s*\{\/\* Series \*\/\}\n[\s\S]*?<\/section>/,
  ""
);

// 5. Remove seriesFilter from reset button
c = c.replace("setSeriesFilter('all');\n                ", "");

// 6. Remove seriesFilter from filter pipeline (if still there)
c = c.replace(
  "    if (seriesFilter !== 'all') result = result.filter(m => m.series === seriesFilter)\n",
  ""
);

// 7. Remove seriesFilter from dependencies
c = c.replace(
  "useCaseFilter, providerFilter, seriesFilter, modalityFilter",
  "useCaseFilter, providerFilter, modalityFilter"
);

// 8. Remove collapsed.series from initial state
c = c.replace("context: false, series: false\n  })", "context: false\n  })");

// 9. Replace repeated empty lines
c = c.replace(/\n\n\n\n+/g, "\n\n");

fs.writeFileSync(path, c, 'utf-8');
console.log('models.tsx: restored AI + Quantum provider sections');
