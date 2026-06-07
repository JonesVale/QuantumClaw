const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/models.tsx';
let c = fs.readFileSync(path, 'utf-8');

// === 1. Split providers into aiProviders + quantumProviders ===
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

// === 2. Remove seriesNames useMemo ===
c = c.replace(
  /\n  const seriesNames = useMemo\([\s\S]*?catalog\.length\)\) \/\/ eslint-disable-line\n/,
  "\n"
);

// === 3. Remove seriesFilter state ===
c = c.replace("  const [seriesFilter, setSeriesFilter] = useState('all')\n", "");

// === 4. Remove seriesFilter from filter pipeline ===
c = c.replace("    if (seriesFilter !== 'all') result = result.filter(m => m.series === seriesFilter)\n", "");

// === 5. Remove seriesFilter from filter useMemo dependencies ===
c = c.replace(", seriesFilter", "");

// === 6. Remove seriesFilter from useEffect dependencies ===
c = c.replace(", seriesFilter", "");

// === 7. Remove setSeriesFilter from reset button ===
c = c.replace("setSeriesFilter('all');\n                ", "");

// === 8. Remove series from collapsed initial state ===
c = c.replace(/series: false,\s*/, "");

// === 9. Add aiProvider + quantumProvider to collapsed initial state ===
c = c.replace(
  "    provider: false, categories: false, modalities: false, context: false",
  "    provider: false, aiProvider: false, quantumProvider: false, categories: false, modalities: false, context: false"
);

// === 10. Replace single Provider section with AI + Quantum sections (no borders) ===
const oldProviderSection = `          {/* Provider */}
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

const newProviderSection = `          {/* AI Providers */}
          <section>
            <button onClick={() => setCollapsed(c => ({...c, aiProvider: !c.aiProvider}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('AI Providers')}</span>
              {collapsed.aiProvider ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.aiProvider && <div className="px-4 pb-2 space-y-0.5">
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
          <section>
            <button onClick={() => setCollapsed(c => ({...c, quantumProvider: !c.quantumProvider}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('Quantum Providers')}</span>
              {collapsed.quantumProvider ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.quantumProvider && <div className="px-4 pb-2 space-y-0.5">
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
          </section>
          {/* Categories */}`;

c = c.replace(oldProviderSection, newProviderSection);

// === 11. Remove all border-b from sidebar sections ===
c = c.replace(/className="border-b">\n            <button/g, 'className="">\n            <button');
c = c.replace(/className="border-b">/g, 'className="">');

// === 12. Remove Series section entirely ===
c = c.replace(
  /\n\n          \/\* Series \*\/\n          <section[\s\S]*?<\/section>/,
  ""
);

// === 13. Remove sidebar border/bg ===
c = c.replace(
  '<aside className={cn(\'w-60 shrink-0 border-r bg-card/50 backdrop-blur-sm flex flex-col transition-all\'',
  '<aside className={cn(\'w-60 shrink-0 flex flex-col transition-all\''
);

// === 14. Remove border from sidebar header ===
c = c.replace(
  '<div className="p-4 border-b flex items-center justify-between">',
  '<div className="p-4 flex items-center justify-between">'
);

// === 15. Remove main top bar border ===
c = c.replace(
  '<div className="flex items-center gap-2 px-4 py-2 border-b bg-card/50 shrink-0">',
  '<div className="flex items-center gap-2 px-4 py-2 shrink-0">'
);

fs.writeFileSync(path, c, 'utf-8');
console.log('Done - sidebar cleaned up');
