// QuantumClaw AI Model Catalog 鈥?full redesign (Braket-inspired)
import { type CatalogItem } from '@/components/model-detail-dialog'

interface ModelsPageProps {
  t: (k: string) => string
  catalog: CatalogItem[]
  search: string
  setSearch: (v: string) => void
  providerFilter: string
  setProviderFilter: (v: string) => void
  useCaseFilter: string
  setUseCaseFilter: (v: string) => void
  contextFilter: string
  setContextFilter: (v: string) => void
  modalityFilter: string
  setModalityFilter: (v: string) => void
  sortBy: string
  setSortBy: (v: string) => void
  sidebarOpen: boolean
  setSidebarOpen: (v: boolean) => void
  collapsed: Record<string, boolean>
  setCollapsed: (v: any) => void
  aiProviders: [string, number][]
  quantumProviders: [string, number][]
  useCaseLabels: Record<string, any>
  filtered: CatalogItem[]
  displayed: CatalogItem[]
  selectedModels: Set<string>
  toggleSelect: (name: string) => void
  clearSelection: () => void
  selectedModelsList: CatalogItem[]
  comparisonOpen: boolean
  setComparisonOpen: (v: boolean) => void
  openDetail: (m: CatalogItem) => void
  auth: any
  visibleCount: number
  setVisibleCount: (v: any) => void
  PAGE_STEP: number
}

export function ModelsPageView(props: ModelsPageProps) {
  const {
    t, catalog, search, setSearch,
    providerFilter, setProviderFilter,
    useCaseFilter, setUseCaseFilter,
    contextFilter, setContextFilter,
    modalityFilter, setModalityFilter,
    sortBy, setSortBy,
    sidebarOpen, setSidebarOpen,
    collapsed, setCollapsed,
    aiProviders, quantumProviders,
    useCaseLabels,
    filtered, displayed,
    selectedModels, toggleSelect, clearSelection,
    selectedModelsList, comparisonOpen, setComparisonOpen,
    openDetail, auth,
    visibleCount, setVisibleCount, PAGE_STEP,
  } = props

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10 py-12">
        {/* Hero */}
        <div className="mb-16">
          <h1 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight">
            <span className="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
              {t('AI Model Catalog')}
            </span>
          </h1>
          <p className="text-lg text-muted-foreground mt-4 max-w-2xl leading-relaxed">
            {t('Browse and compare models from all major providers')}
          </p>
        </div>

        {/* Top action bar */}
        <div className="flex items-center gap-4 mb-10 flex-wrap">
          <div className="relative flex-1 min-w-[240px] max-w-sm">
            <svg className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
            <input className="flex h-10 w-full rounded-xl border border-input bg-background px-9 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" placeholder={t('Search models...')} value={search} onChange={(e) => setSearch(e.target.value)} />
          </div>
          <button className="inline-flex items-center justify-center rounded-xl border border-input bg-background h-10 px-4 text-sm font-medium hover:bg-accent hover:text-accent-foreground gap-2" onClick={() => setSidebarOpen(!sidebarOpen)}>
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
            {t('Filters')}
          </button>
          <select value={sortBy} onChange={(e) => setSortBy(e.target.value)} className="h-10 rounded-xl border border-input bg-background px-3 text-sm">
            <option value="name">{t('By Name')}</option>
            <option value="price-asc">{t('Price: Low to High')}</option>
            <option value="price-desc">{t('Price: High to Low')}</option>
          </select>
          {selectedModels.size >= 2 && (
            <div className="flex items-center gap-2 bg-primary/10 rounded-xl px-3 py-1.5">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-primary"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
              <span className="text-xs font-medium text-primary">{selectedModels.size} {t('selected')}</span>
              <button className="inline-flex items-center justify-center rounded-lg bg-primary text-primary-foreground h-7 px-3 text-xs font-medium" onClick={() => setComparisonOpen(true)}>
                {t('Compare')} ({selectedModels.size})
              </button>
              <button className="p-0 rounded-lg hover:bg-muted" onClick={() => { selectedModels.clear(); clearSelection(); }}>
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
              </button>
            </div>
          )}
        </div>

        {/* Active filter tags */}
        {(providerFilter || useCaseFilter !== 'all' || contextFilter || modalityFilter) && (
          <div className="flex items-center gap-3 mb-8 flex-wrap">
            {providerFilter && (
              <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300">
                {providerFilter}
                <button onClick={() => setProviderFilter('')}><svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg></button>
              </span>
            )}
            {useCaseFilter !== 'all' && (
              <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300">
                {useCaseLabels[useCaseFilter]?.label || useCaseFilter}
                <button onClick={() => setUseCaseFilter('all')}><svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg></button>
              </span>
            )}
            <button onClick={() => { setProviderFilter(''); setUseCaseFilter('all'); setContextFilter(''); setModalityFilter(''); }} className="text-xs text-muted-foreground hover:text-foreground underline underline-offset-2">
              {t('Clear filters')}
            </button>
          </div>
        )}

        {/* Sidebar overlay */}
        {sidebarOpen && (
          <div className="fixed inset-0 z-40 md:relative md:inset-auto md:z-auto" onClick={() => setSidebarOpen(false)}>
            <div className="absolute inset-0 bg-black/20 md:hidden" />
            <aside className="absolute left-0 top-0 h-full w-64 bg-background border-r shadow-xl md:shadow-none md:static md:border-0 z-50 overflow-y-auto p-6 space-y-6" onClick={e => e.stopPropagation()}>
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold">{t('Filters')}</h3>
                <button className="h-6 w-6 md:hidden rounded-lg hover:bg-muted flex items-center justify-center" onClick={() => setSidebarOpen(false)}>
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
                </button>
              </div>

              {/* AI Providers */}
              <div>
                <button onClick={() => setCollapsed((c: any) => ({...c, aiProvider: !c.aiProvider}))} className="flex items-center justify-between w-full text-sm font-semibold text-muted-foreground hover:text-foreground py-1.5">
                  <span>{t('AI Providers')}</span>
                  {collapsed.aiProvider
                    ? <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="9 18 15 12 9 6"/></svg>
                    : <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="6 9 12 15 18 9"/></svg>
                  }
                </button>
                {!collapsed.aiProvider && <div className="mt-1 space-y-0.5">
                  <button onClick={() => setProviderFilter('')} className={'w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ' + (!providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                    {t('All Providers')} <span className="text-xs text-muted-foreground">({catalog.filter((m: any) => m.use_case !== 'quantum').length})</span>
                  </button>
                  {aiProviders.slice(0, 20).map(([p, cnt]) => (
                    <button key={p} onClick={() => setProviderFilter(p)} className={'w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center justify-between ' + (providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
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
                      <div className="space-y-0.5">{aiProviders.slice(20).map(([p, cnt]) => (
                        <button key={p} onClick={() => setProviderFilter(p)} className={'w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center justify-between ' + (providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                          <span>{p}</span><span className="text-xs text-muted-foreground">{cnt}</span>
                        </button>
                      ))}</div>
                    </details>
                  )}
                </div>}
              </div>

              {/* Quantum Providers */}
              {quantumProviders.length > 0 && <div>
                <button onClick={() => setCollapsed((c: any) => ({...c, quantumProvider: !c.quantumProvider}))} className="flex items-center justify-between w-full text-sm font-semibold text-muted-foreground hover:text-foreground py-1.5">
                  <span className="flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-purple-500"><circle cx="12" cy="12" r="3"/><path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83"/></svg>
                    {t('Quantum Providers')}
                  </span>
                  {collapsed.quantumProvider
                    ? <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="9 18 15 12 9 6"/></svg>
                    : <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="6 9 12 15 18 9"/></svg>
                  }
                </button>
                {!collapsed.quantumProvider && <div className="mt-1 space-y-0.5">
                  <button onClick={() => setProviderFilter('')} className={'w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ' + (!providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                    {t('All Quantum')} <span className="text-xs text-muted-foreground">({catalog.filter((m: any) => m.use_case === 'quantum').length})</span>
                  </button>
                  {quantumProviders.map(([p, cnt]) => (
                    <button key={p} onClick={() => setProviderFilter(p)} className={'w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center justify-between ' + (providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                      <span className="flex items-center gap-2"><div className="w-2.5 h-2.5 rounded-full bg-gradient-to-br from-violet-500 to-purple-600" />{p}</span>
                      <span className="text-xs text-muted-foreground">{cnt}</span>
                    </button>
                  ))}
                </div>}
              </div>}

              {/* Categories */}
              <div>
                <button onClick={() => setCollapsed((c: any) => ({...c, categories: !c.categories}))} className="flex items-center justify-between w-full text-sm font-semibold text-muted-foreground hover:text-foreground py-1.5">
                  <span>{t('Categories')}</span>
                  {collapsed.categories
                    ? <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="9 18 15 12 9 6"/></svg>
                    : <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="6 9 12 15 18 9"/></svg>
                  }
                </button>
                {!collapsed.categories && <div className="mt-1 space-y-0.5">
                  <button onClick={() => setUseCaseFilter('all')} className={'w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ' + (useCaseFilter === 'all' ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{t('All Models')}</button>
                  {Object.entries(useCaseLabels).map(([k, v]: [string, any]) => (
                    <button key={k} onClick={() => setUseCaseFilter(k)} className={'w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center gap-2 ' + (useCaseFilter === k ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{t(v.label)}</button>
                  ))}
                </div>}
              </div>

              {/* Context Length */}
              <div>
                <button onClick={() => setCollapsed((c: any) => ({...c, context: !c.context}))} className="flex items-center justify-between w-full text-sm font-semibold text-muted-foreground hover:text-foreground py-1.5">
                  <span>{t('Context Length')}</span>
                  {collapsed.context
                    ? <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="9 18 15 12 9 6"/></svg>
                    : <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="6 9 12 15 18 9"/></svg>
                  }
                </button>
                {!collapsed.context && <div className="mt-1 space-y-0.5">
                  {[{v:'',l:'All'},{v:'0-8192',l:'< 8K'},{v:'8193-32768',l:'8K - 32K'},{v:'32769-131072',l:'32K - 128K'},{v:'131073-999999999',l:'> 128K'}].map((r: any) => (
                    <button key={r.v} onClick={() => setContextFilter(r.v)} className={'w-full text-left px-3 py-1.5 text-sm rounded-lg transition-colors ' + (contextFilter === r.v ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{r.l}</button>
                  ))}
                </div>}
              </div>
            </aside>
          </div>
        )}

        {/* Model grid */}
        {filtered.length === 0 ? (
          <div className="text-center py-20 text-muted-foreground">
            <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="mx-auto mb-4 opacity-20"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
            <p className="text-lg mb-2">{t('No models found')}</p>
            <p className="text-sm mb-6">{t('try_adjust_filters')}</p>
            <button className="inline-flex items-center justify-center rounded-xl border border-input bg-background h-10 px-4 text-sm font-medium hover:bg-accent" onClick={() => { setSearch(''); setUseCaseFilter('all'); setContextFilter(''); setProviderFilter(''); setModalityFilter(''); }}>
              {t('reset_filters')}
            </button>
          </div>
        ) : (
          <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {displayed.map((m: any) => {
              const isSelected = selectedModels.has(m.name);
              return (
                <div key={m.name} className={'relative rounded-xl p-6 transition-all duration-200 bg-card hover:shadow-md border-0 ' + (isSelected ? 'ring-2 ring-primary/50' : '')}>
                  {/* Provider row */}
                  <div className="flex items-center justify-between mb-4">
                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">{m.provider || 'Unknown'}</span>
                    <button onClick={(e) => { e.stopPropagation(); toggleSelect(m.name) }}
                      className={'w-4 h-4 rounded border-2 flex items-center justify-center transition-colors shrink-0 ' + (isSelected ? 'bg-primary border-primary text-primary-foreground' : 'border-muted-foreground/30 hover:border-muted-foreground/60')}>
                      {isSelected && <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><polyline points="20 6 9 17 4 12"/></svg>}
                    </button>
                  </div>
                  {/* Model name */}
                  <h3 className="text-lg font-semibold tracking-tight mb-2 cursor-pointer hover:text-primary transition-colors" onClick={() => openDetail(m)}>{m.name}</h3>
                  {/* Tags */}
                  <div className="flex items-center gap-1.5 mb-3 flex-wrap">
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-muted text-muted-foreground">
                      <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/></svg>
                      {(m.context_window / 1000).toFixed(0)}K
                    </span>
                    {(m.input_modalities || []).slice(0, 2).map((mod: string) => (
                      <span key={mod} className="px-1.5 py-0.5 rounded text-xs bg-muted/50 text-muted-foreground">{mod}</span>
                    ))}
                    <div className={'w-2 h-2 rounded-full ' + (m.status === 1 ? 'bg-emerald-500' : 'bg-muted-foreground/30')} />
                  </div>
                  {/* Pricing */}
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-muted-foreground">{m.input_price > 0 ? 'IN $' + m.input_price.toFixed(5) : 'Free'}</span>
                    <span className="text-muted-foreground">{m.output_price > 0 ? 'OUT $' + m.output_price.toFixed(5) : ''}</span>
                  </div>
                  {/* Actions at bottom */}
                  <div className="flex items-center gap-2 mt-4 pt-3 border-t border-border/40">
                    <button className="text-xs text-muted-foreground hover:text-foreground font-medium transition-colors" onClick={() => openDetail(m)}>
                      {t('Details')} →
                    </button>
                    <span className="text-muted-foreground/30">·</span>
                    <button className="text-xs text-primary hover:text-primary/80 font-medium transition-colors" onClick={() => window.location.href = auth?.user ? '/chat' : '/sign-in'}>
                      {t('Call')}
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
          {visibleCount < filtered.length && (
            <div className="flex justify-center mt-10">
              <button className="inline-flex items-center justify-center rounded-xl border border-input bg-background w-full max-w-sm py-3 text-sm font-medium hover:bg-accent" onClick={() => setVisibleCount((v: number) => v + PAGE_STEP)}>
                {t('Show more')} ({filtered.length - visibleCount} {t('remaining')})
              </button>
            </div>
          )}
          </>
        )}
      </div>
    </div>
  );
}
