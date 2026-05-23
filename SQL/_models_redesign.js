const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/models.tsx';
let c = fs.readFileSync(path, 'utf-8');
const orig = c;

// ===== IMPORTS =====
// Remove unused imports
c = c.replace(/^import \* as Icons from.*\n/m, '');
c = c.replace("import { codeToType } from '@/lib/tlanguages'\n", "");

// Add Atom to lucide imports if missing
c = c.replace(
  "CheckSquare, BarChart3\n} from 'lucide-react'",
  "CheckSquare, BarChart3, Atom, ListFilter\n} from 'lucide-react'"
);

c = c.replace(
  "CheckSquare, BarChart3 } from 'lucide-react'",
  "CheckSquare, BarChart3, Atom, ListFilter } from 'lucide-react'"
);

// ===== STATE DECLARATIONS =====
// Remove seriesFilter state
c = c.replace("  const [seriesFilter, setSeriesFilter] = useState('all')\n", "");

// Remove series from collapsed initial state
c = c.replace(", series: false", "");

// Clean up deprecated provider in collapsed (use aiProvider/quantumProvider now)
c = c.replace(
  "    provider: false, aiProvider: false, quantumProvider: false, categories: false, modalities: false, context: false",
  "    aiProvider: false, quantumProvider: false, categories: false, modalities: false, context: false"
);

// ===== DATA COMPUTATIONS =====
// Replace single providers + seriesNames with aiProviders + quantumProviders
const oldProvidersAndSeries = `  const providers = useMemo(() => {
    if (catalog.length === 0) return []
    const map = new Map<string, number>()
    for (const m of catalog) {
      const p = m.provider || 'Unknown'
      map.set(p, (map.get(p) || 0) + 1)
    }
    return Array.from(map.entries()).sort((a, b) => b[1] - a[1])
  }, [catalog.length]) // eslint-disable-line
  
  const seriesNames = useMemo(() => {
    if (catalog.length === 0) return []
    const set = new Set<string>()
    for (const m of catalog) {
      if (m.series) set.add(m.series)
    }
    return Array.from(set).sort()
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

c = c.replace(oldProvidersAndSeries, newProviders);

// Remove series from filter pipeline
c = c.replace("    if (seriesFilter !== 'all') result = result.filter(m => m.series === seriesFilter)\n", "");
c = c.replace(", seriesFilter", "");
c = c.replace("setSeriesFilter('all');\n                ", "");

// ===== COMPLETE REWRITE OF THE RETURN BLOCK =====
// Match from "return (" to the end
const returnStart = c.indexOf("  return (");
const returnEnd = c.lastIndexOf("</div>");

if (returnStart < 0 || returnEnd < 0) {
  console.error("Could not find return block");
  process.exit(1);
}

const beforeReturn = c.substring(0, returnStart);
const afterReturn = c.substring(returnEnd + 6).trim();

// New JSX layout
const newReturn = `  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* ===== HERO ===== */}
        <div className="mb-10">
          <h1 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight">
            <span className="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
              {t('AI Model Catalog')}
            </span>
          </h1>
          <p className="text-base text-muted-foreground mt-3 max-w-2xl leading-relaxed">
            {t('Browse and compare models from all major providers')}
          </p>
        </div>

        {/* ===== TOP ACTION BAR ===== */}
        <div className="flex items-center gap-3 mb-8 flex-wrap">
          {/* Search */}
          <div className="relative flex-1 min-w-[240px] max-w-sm">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input className="pl-9 h-10 text-sm rounded-xl" placeholder={t('Search models...')} value={search} onChange={(e) => setSearch(e.target.value)} />
          </div>

          {/* Sidebar toggle */}
          <Button variant="outline" size="sm" className="h-10 gap-2 rounded-xl" onClick={() => setSidebarOpen(!sidebarOpen)}>
            <ListFilter className="h-4 w-4" />
            {t('Filters')}
          </Button>

          {/* Sort */}
          <Select value={sortBy} onValueChange={(v) => setSortBy(v as SortOption)}>
            <SelectTrigger className="w-[150px] h-10 text-sm rounded-xl">
              <ArrowUpDown className="h-3.5 w-3.5 mr-1" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="name">{t('By Name')}</SelectItem>
              <SelectItem value="price-asc">{t('Price: Low to High')}</SelectItem>
              <SelectItem value="price-desc">{t('Price: High to Low')}</SelectItem>
            </SelectContent>
          </Select>

          {/* Comparison button */}
          {selectedModels.size >= 2 && (
            <div className="flex items-center gap-2 bg-primary/10 rounded-xl px-3 py-1.5">
              <BarChart3 className="h-4 w-4 text-primary" />
              <span className="text-xs font-medium text-primary">{selectedModels.size} {t('selected')}</span>
              <Button size="sm" variant="default" className="h-7 text-xs gap-1 rounded-lg" onClick={() => setComparisonOpen(true)}>
                {t('Compare')} ({selectedModels.size})
              </Button>
              <Button size="sm" variant="ghost" className="h-7 w-7 p-0" onClick={() => setSelectedModels(new Set())}>
                <X className="h-3.5 w-3.5" />
              </Button>
            </div>
          )}
        </div>

        {/* ===== ACTIVE FILTER TAGS ===== */}
        {(providerFilter || useCaseFilter !== 'all' || contextFilter || modalityFilter) && (
          <div className="flex items-center gap-2 mb-6 flex-wrap">
            {providerFilter && (
              <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300">
                {providerFilter}
                <button onClick={() => setProviderFilter('')} className="hover:text-blue-900 dark:hover:text-blue-100"><X className="h-3 w-3" /></button>
              </span>
            )}
            {useCaseFilter !== 'all' && (
              <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300">
                {useCaseLabels[useCaseFilter]?.label || useCaseFilter}
                <button onClick={() => setUseCaseFilter('all')} className="hover:text-purple-900 dark:hover:text-purple-100"><X className="h-3 w-3" /></button>
              </span>
            )}
            {contextFilter && (
              <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300">
                {contextFilter}
                <button onClick={() => setContextFilter('')} className="hover:text-amber-900 dark:hover:text-amber-100"><X className="h-3 w-3" /></button>
              </span>
            )}
            <button onClick={() => { setProviderFilter(''); setUseCaseFilter('all'); setContextFilter(''); setModalityFilter(''); }} className="text-xs text-muted-foreground hover:text-foreground underline underline-offset-2">
              {t('Clear filters')}
            </button>
          </div>
        )}

        {/* ===== SIDEBAR (overlay panel) ===== */}
        {sidebarOpen && (
          <div className="fixed inset-0 z-40 md:relative md:inset-auto md:z-auto" onClick={() => setSidebarOpen(false)}>
            <div className="absolute inset-0 bg-black/20 md:hidden" />
            <aside className="absolute left-0 top-0 h-full w-64 bg-background border-r shadow-xl md:shadow-none md:static md:border-0 z-50 overflow-y-auto p-4 space-y-4" onClick={e => e.stopPropagation()}>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-semibold">{t('Filters')}</h3>
                <Button variant="ghost" size="icon" className="h-6 w-6 md:hidden" onClick={() => setSidebarOpen(false)}><X className="h-3.5 w-3.5" /></Button>
              </div>

              {/* AI Providers */}
              <div>
                <button onClick={() => setCollapsed(c => ({...c, aiProvider: !c.aiProvider}))} className="flex items-center justify-between w-full text-sm font-semibold text-muted-foreground hover:text-foreground transition-colors py-1.5">
                  <span>{t('AI Providers')}</span>
                  {!collapsed.aiProvider ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                {!collapsed.aiProvider && <div className="mt-1 space-y-0.5 pl-1">
                  <button onClick={() => setProviderFilter('')} className={cn('w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors', !providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                    {t('All Providers')} <span className="text-xs text-muted-foreground">({catalog.filter(m => m.use_case !== 'quantum').length})</span>
                  </button>
                  {aiProviders.slice(0, 20).map(([p, cnt]) => (
                    <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                      <span>{p}</span>
                      <span className="text-xs text-muted-foreground">{cnt}</span>
                    </button>
                  ))}
                  {aiProviders.length > 20 && (
                    <details className="group">
                      <summary className="w-full text-left px-3 py-1 text-xs text-muted-foreground hover:bg-muted/50 rounded-lg cursor-pointer list-none">
                        <span className="group-open:hidden">{t('Show more...')} ({aiProviders.length - 20})</span>
                        <span className="hidden group-open:inline">{t('Show less')}</span>
                      </summary>
                      <div className="space-y-0.5">
                        {aiProviders.slice(20).map(([p, cnt]) => (
                          <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                            <span>{p}</span>
                            <span className="text-xs text-muted-foreground">{cnt}</span>
                          </button>
                        ))}
                      </div>
                    </details>
                  )}
                </div>}
              </div>

              {/* Quantum Providers */}
              {quantumProviders.length > 0 && <div>
                <button onClick={() => setCollapsed(c => ({...c, quantumProvider: !c.quantumProvider}))} className="flex items-center justify-between w-full text-sm font-semibold text-muted-foreground hover:text-foreground transition-colors py-1.5">
                  <span className="flex items-center gap-2">
                    <Atom className="h-3.5 w-3.5 text-purple-500" />
                    {t('Quantum Providers')}
                  </span>
                  {!collapsed.quantumProvider ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                {!collapsed.quantumProvider && <div className="mt-1 space-y-0.5 pl-1">
                  <button onClick={() => setProviderFilter('')} className={cn('w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors', !providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                    {t('All Quantum')} <span className="text-xs text-muted-foreground">({catalog.filter(m => m.use_case === 'quantum').length})</span>
                  </button>
                  {quantumProviders.map(([p, cnt]) => (
                    <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                      <span className="flex items-center gap-2">
                        <div className="w-2.5 h-2.5 rounded-full bg-gradient-to-br from-violet-500 to-purple-600" />
                        {p}
                      </span>
                      <span className="text-xs text-muted-foreground">{cnt}</span>
                    </button>
                  ))}
                </div>}
              </div>}

              {/* Categories */}
              <div>
                <button onClick={() => setCollapsed(c => ({...c, categories: !c.categories}))} className="flex items-center justify-between w-full text-sm font-semibold text-muted-foreground hover:text-foreground transition-colors py-1.5">
                  <span>{t('Categories')}</span>
                  {!collapsed.categories ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                {!collapsed.categories && <div className="mt-1 space-y-0.5 pl-1">
                  <button onClick={() => setUseCaseFilter('all')} className={cn('w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors', useCaseFilter === 'all' ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{t('All Models')}</button>
                  {Object.entries(useCaseLabels).map(([k, v]) => {
                    const Icon = v.icon;
                    return (
                      <button key={k} onClick={() => setUseCaseFilter(k)} className={cn('w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center gap-2', useCaseFilter === k ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                        <div className={\`w-4 h-4 rounded bg-gradient-to-br \${v.color} flex items-center justify-center shrink-0\`}><Icon className="h-2.5 w-2.5 text-white" /></div>
                        {t(v.label)}
                      </button>
                    );
                  })}
                </div>}
              </div>

              {/* Context Length */}
              <div>
                <button onClick={() => setCollapsed(c => ({...c, context: !c.context}))} className="flex items-center justify-between w-full text-sm font-semibold text-muted-foreground hover:text-foreground transition-colors py-1.5">
                  <span>{t('Context Length')}</span>
                  {!collapsed.context ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </button>
                {!collapsed.context && <div className="mt-1 space-y-0.5 pl-1">
                  {[{v:'',l:'All'},{v:'0-8192',l:'< 8K'},{v:'8193-32768',l:'8K - 32K'},{v:'32769-131072',l:'32K - 128K'},{v:'131073-999999999',l:'> 128K'}].map(r => (
                    <button key={r.v} onClick={() => setContextFilter(r.v)} className={cn('w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors', contextFilter === r.v ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{r.l}</button>
                  ))}
                </div>}
              </div>
            </aside>
          </div>
        )}

        {/* ===== MODEL GRID ===== */}
        {filtered.length === 0 ? (
          <div className="text-center py-20 text-muted-foreground">
            <Search className="h-10 w-10 mx-auto mb-4 opacity-20" />
            <p className="text-lg mb-2">{t('No models found')}</p>
            <p className="text-sm mb-6">{t('try_adjust_filters')}</p>
            <Button variant="outline" onClick={() => { setSearch(''); setUseCaseFilter('all'); setContextFilter(''); setProviderFilter(''); setModalityFilter(''); }}>
              {t('reset_filters')}
            </Button>
          </div>
        ) : (
          <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {displayed.map((m) => {
              const isSelected = selectedModels.has(m.name);
              const provider = m.provider || 'Unknown';
              return (
                <div key={m.name} className={cn(
                  'group relative rounded-xl p-5 transition-all duration-200 bg-card hover:shadow-md border-0',
                  isSelected && 'ring-2 ring-primary/50'
                )}>
                  {/* Top: provider + checkbox */}
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">{provider}</span>
                    <button
                      onClick={(e) => { e.stopPropagation(); toggleSelect(m.name) }}
                      className={cn(
                        'w-4 h-4 rounded border-2 flex items-center justify-center transition-colors shrink-0',
                        isSelected ? 'bg-primary border-primary text-primary-foreground' : 'border-muted-foreground/30 hover:border-muted-foreground/60'
                      )}
                    >
                      {isSelected && <CheckSquare className="h-3 w-3" />}
                    </button>
                  </div>

                  {/* Model name */}
                  <h3 className="text-base font-semibold tracking-tight mb-2 cursor-pointer hover:text-primary transition-colors" onClick={() => openDetail(m)}>
                    {m.name}
                  </h3>

                  {/* Tags row */}
                  <div className="flex items-center gap-1.5 mb-3 flex-wrap">
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-medium bg-muted text-muted-foreground">
                      <Cpu className="h-2.5 w-2.5" />{(m.context_window / 1000).toFixed(0)}K
                    </span>
                    {m.input_modalities?.slice(0, 2).map(mod => (
                      <span key={mod} className="px-1.5 py-0.5 rounded text-[10px] bg-muted/50 text-muted-foreground">{mod}</span>
                    ))}
                    <div className={\`w-2 h-2 rounded-full \${m.status === 1 ? 'bg-emerald-500' : 'bg-muted-foreground/30'}\`} />
                  </div>

                  {/* Pricing */}
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-muted-foreground">
                      {m.input_price > 0 ? \`IN $\${m.input_price.toFixed(5)}\` : 'Free'}
                    </span>
                    <span className="text-muted-foreground">
                      {m.output_price > 0 ? \`OUT $\${m.output_price.toFixed(5)}\` : ''}
                    </span>
                  </div>

                  {/* Hover actions */}
                  <div className="absolute inset-0 rounded-xl bg-background/80 backdrop-blur-[1px] opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                    <Button size="sm" variant="outline" className="h-8 text-xs rounded-lg" onClick={() => openDetail(m)}>
                      {t('Details')}
                    </Button>
                    <Link to={auth.user ? '/chat' : '/sign-in'} search={auth.user ? undefined : { redirect: '/chat' }}>
                      <Button size="sm" className="h-8 text-xs gap-1 rounded-lg">
                        <Play className="h-3 w-3" />{t('Call')}
                      </Button>
                    </Link>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Load More */}
          {visibleCount < filtered.length && (
            <div className="flex justify-center mt-10">
              <Button variant="outline" className="w-full max-w-sm rounded-xl py-5 text-sm" onClick={() => setVisibleCount(v => v + PAGE_STEP)}>
                {t('Show more')} ({filtered.length - visibleCount} {t('remaining')})
              </Button>
            </div>
          )}
          </>
        )}
      </div>

      {/* Detail Dialog */}
      {detailOpen && detailModel && (
        <ModelDetailDialog
          model={detailModel}
          open={detailOpen}
          onOpenChange={(open) => { setDetailOpen(open); if (!open) setDetailModel(null); }}
        />
      )}

      {/* Comparison Dialog */}
      {comparisonOpen && selectedModelsList.length >= 2 && (
        <ModelComparisonDialog
          models={selectedModelsList}
          open={comparisonOpen}
          onOpenChange={(open) => { if (!open) setComparisonOpen(false); }}
        />
      )}
    </div>
  );`;

c = beforeReturn + newReturn;

// Clean up the end (after return block)
if (c.endsWith('\n\n\n\n')) c = c.replace(/\n*$/, '\n');

fs.writeFileSync(path, c, 'utf-8');
console.log('Done - models.tsx fully redesigned');
