import { createFileRoute, Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo, useEffect, useRef } from 'react'
import {
  Search, SlidersHorizontal, X, ArrowUpDown,
  ChevronRight, ChevronDown, MessageSquare, Code, Brain,
  Image, Cpu, Atom, Play,
  CheckSquare, BarChart3
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { codeToType } from '@/lib/tlanguages'
import { useAuthStore } from '@/stores/auth-store'
import { getEnhancedModels } from '@/lib/api-extended'
import { ModelDetailDialog, type CatalogItem } from '@/components/model-detail-dialog'
import { ModelComparisonDialog } from '@/components/model-comparison-dialog'

export const Route = createFileRoute('/models')({
  component: ModelsPage,
})

const useCaseLabels: Record<string, { label: string; icon: React.ElementType; color: string }> = {
  'chat': { label: 'Chat & Assistant', icon: MessageSquare, color: 'from-blue-500 to-blue-600' },
  'coding': { label: 'Code Generation', icon: Code, color: 'from-green-500 to-green-600' },
  'reasoning': { label: 'Reasoning', icon: Brain, color: 'from-purple-500 to-purple-600' },
  'vision': { label: 'Vision', icon: Image, color: 'from-cyan-500 to-cyan-600' },
}

type SortOption = 'name' | 'price-asc' | 'price-desc'

function ModelsPage() {
  const { t, i18n } = useTranslation()
  const { auth } = useAuthStore()
  const [search, setSearch] = useState('')
  const [useCaseFilter, setUseCaseFilter] = useState('all')
  const [seriesFilter, setSeriesFilter] = useState('all')
  const [modalityFilter, setModalityFilter] = useState('')
  const [contextFilter, setContextFilter] = useState('')
  const [providerFilter, setProviderFilter] = useState('')
  const [sortBy, setSortBy] = useState<SortOption>('name')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({
    provider: false, categories: false, modalities: false, context: false, series: false
  })

  // Detail dialog state
  const [detailModel, setDetailModel] = useState<CatalogItem | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  // Pagination
  const [visibleCount, setVisibleCount] = useState(30)
  const PAGE_STEP = 30

  // Comparison state
  const [selectedModels, setSelectedModels] = useState<Set<string>>(new Set())
  const [comparisonOpen, setComparisonOpen] = useState(false)

  const lang = codeToType[i18n.language] || 'English'
  const { data } = useQuery({
    queryKey: ['model-catalog', lang],
    queryFn: async () => {
      const r = await fetch(`/api/model-catalog?lang=${encodeURIComponent(lang)}`)
      if (!r.ok) throw new Error('Failed to fetch')
      return r.json()
    },
    staleTime: 60 * 1000,
  })
  const catalog: CatalogItem[] = data?.data || []

  // Derived data - 依赖 catalog.length 避免每次 re-render 重算
  const providers = useMemo(() => {
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
  }, [catalog.length]) // eslint-disable-line

  const filtered = useMemo(() => {
    let result = catalog
    if (search) { const q = search.toLowerCase(); result = result.filter(m => m.name.toLowerCase().includes(q) || m.description.toLowerCase().includes(q)) }
    if (useCaseFilter !== 'all') result = result.filter(m => m.use_case === useCaseFilter)
    if (providerFilter) result = result.filter(m => (m.provider || '') === providerFilter)
    if (seriesFilter !== 'all') result = result.filter(m => m.series === seriesFilter)
    if (modalityFilter) result = result.filter(m => m.input_modalities.some((mod: string) => mod.toLowerCase() === modalityFilter.toLowerCase()))
    if (contextFilter) {
      const ctx = (m: any) => m.context_window || 0
      switch (contextFilter) {
        case '0-8192': result = result.filter(m => ctx(m) <= 8192); break
        case '8193-32768': result = result.filter(m => ctx(m) > 8192 && ctx(m) <= 32768); break
        case '32769-131072': result = result.filter(m => ctx(m) > 32768 && ctx(m) <= 131072); break
        case '131073-999999999': result = result.filter(m => ctx(m) > 131072); break
      }
    }
    switch (sortBy) { case 'price-asc': result.sort((a, b) => (a.input_price ?? 999) - (b.input_price ?? 999)); break; case 'price-desc': result.sort((a, b) => (b.input_price ?? 0) - (a.input_price ?? 0)); break; default: result.sort((a, b) => a.name.localeCompare(b.name)) }
    return result
  }, [catalog, search, useCaseFilter, providerFilter, seriesFilter, modalityFilter, contextFilter, sortBy])

  // Reset pagination when filters change
  useEffect(() => setVisibleCount(PAGE_STEP), [search, useCaseFilter, providerFilter, seriesFilter, modalityFilter, contextFilter, sortBy])

  const displayed = useMemo(() => filtered.slice(0, visibleCount), [filtered, visibleCount])

  const toggleSelect = (name: string) => {
    setSelectedModels(prev => {
      const next = new Set(prev)
      if (next.has(name)) {
        next.delete(name)
      } else {
        if (next.size >= 4) return prev // Max 4
        next.add(name)
      }
      return next
    })
  }

  const openDetail = (m: CatalogItem) => {
    setDetailModel(m)
    setDetailOpen(true)
  }

  const selectedModelsList = useMemo(() => {
    return catalog.filter(m => selectedModels.has(m.name))
  }, [catalog, selectedModels])

  return (
    <div className="flex min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Sidebar */}
      <aside className={cn('w-60 shrink-0 border-r bg-card/50 backdrop-blur-sm flex flex-col transition-all', sidebarOpen ? 'translate-x-0' : '-translate-x-full fixed z-10 h-full md:relative md:translate-x-0')}>
        <div className="p-4 border-b flex items-center justify-between">
          <span className="font-semibold text-sm">{t('Filters')}</span>
          <Button variant="ghost" size="icon" className="h-6 w-6 md:hidden" onClick={() => setSidebarOpen(false)}><X className="h-3.5 w-3.5" /></Button>
        </div>
        <ScrollArea className="flex-1">
          {/* Provider */}
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
          {/* Categories */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, categories: !c.categories}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('Categories')}</span>
              {collapsed.categories ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.categories && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setUseCaseFilter('all')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', useCaseFilter === 'all' ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{t('All Models')}</button>
              <button onClick={() => setUseCaseFilter('quantum')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center gap-2', useCaseFilter === 'quantum' ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                <div className="w-4 h-4 rounded bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shrink-0"><Atom className="h-2.5 w-2.5 text-white" /></div>
                {t('Quantum Computing')}
              </button>
              {Object.entries(useCaseLabels).map(([k, v]) => (
                <button key={k} onClick={() => setUseCaseFilter(k)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center gap-2', useCaseFilter === k ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                  <div className={`w-4 h-4 rounded bg-gradient-to-br ${v.color} flex items-center justify-center shrink-0`}><v.icon className="h-2.5 w-2.5 text-white" /></div>
                  {t(v.label)}
                </button>
              ))}
            </div>}
          </section>

          {/* Input Modalities */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, modalities: !c.modalities}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('Input Modalities')}</span>
              {collapsed.modalities ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.modalities && <div className="px-4 pb-2 space-y-0.5">
              {[{k:'',l:'All'},{k:'Text',l:'Text'},{k:'Image',l:'Image'},{k:'File',l:'File'},{k:'Audio',l:'Audio'},{k:'Video',l:'Video'}].map(mod => (
                <button key={mod.k} onClick={() => setModalityFilter(mod.k)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', modalityFilter === mod.k ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{mod.l}</button>
              ))}
            </div>}
          </section>

          {/* Context Length */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, context: !c.context}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('Context Length')}</span>
              {collapsed.context ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.context && <div className="px-4 pb-2 space-y-0.5">
              {[{v:'',l:'All'},{v:'0-8192',l:'鈮?8K'},{v:'8193-32768',l:'8K - 32K'},{v:'32769-131072',l:'32K - 128K'},{v:'131073-999999999',l:'> 128K'}].map(r => (
                <button key={r.v} onClick={() => setContextFilter(r.v)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', contextFilter === r.v ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{r.l}</button>
              ))}
            </div>}
          </section>

          {/* Series */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, series: !c.series}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('Series')}</span>
              {collapsed.series ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.series && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setSeriesFilter('all')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', seriesFilter === 'all' ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{t('All')}</button>
              {seriesNames.map(s => (
                <button key={s} onClick={() => setSeriesFilter(s)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', seriesFilter === s ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{s}</button>
              ))}
            </div>}
          </section>
        </ScrollArea>
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Top Bar */}
        <div className="flex items-center gap-2 px-4 py-2 border-b bg-card/50 shrink-0">
          <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => setSidebarOpen(!sidebarOpen)}>
            <SlidersHorizontal className="h-4 w-4" />
          </Button>
          
          {/* Comparison Button Bar */}
          {selectedModels.size >= 2 && (
            <div className="flex items-center gap-2 bg-primary/10 rounded-lg px-3 py-1.5 border border-primary/20 ml-1">
              <BarChart3 className="h-4 w-4 text-primary" />
              <span className="text-xs font-medium text-primary">
                {selectedModels.size} {t('selected')}
              </span>
              <Button
                size="sm"
                variant="default"
                className="h-7 text-xs gap-1"
                onClick={() => setComparisonOpen(true)}
              >
                {t('Compare')} ({selectedModels.size})
              </Button>
              <Button
                size="sm"
                variant="ghost"
                className="h-7 w-7 p-0"
                onClick={() => setSelectedModels(new Set())}
              >
                <X className="h-3.5 w-3.5" />
              </Button>
            </div>
          )}

          <span className="text-xs text-muted-foreground font-medium">{catalog.length} {t('models')}</span>
          <div className="ml-auto flex items-center gap-2 overflow-x-auto">
            <Badge variant="outline" className="text-xs cursor-pointer hover:bg-muted">Text</Badge>
            <Badge variant="outline" className="text-xs cursor-pointer hover:bg-muted">Image</Badge>
            <Badge variant="outline" className="text-xs cursor-pointer hover:bg-muted">Audio</Badge>
            <Badge variant="outline" className="text-xs cursor-pointer hover:bg-muted">Video</Badge>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Header */}
          <div>
            <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
              {t('Models')}
            </h1>
            <p className="text-muted-foreground mt-1 text-sm">
              {t('Browse and compare models from all major providers')}
            </p>
          </div>

          {/* Search + Sort */}
          <div className="flex items-center gap-3">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input className="pl-9 h-9 text-sm" placeholder={t('Search models...')} value={search} onChange={(e) => setSearch(e.target.value)} />
            </div>
            <Select value={sortBy} onValueChange={(v) => setSortBy(v as SortOption)}>
              <SelectTrigger className="w-[130px] h-9 text-xs">
                <ArrowUpDown className="h-3.5 w-3.5 mr-1" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="name">{t('By Name')}</SelectItem>
                <SelectItem value="price-asc">{t('Price: Low to High')}</SelectItem>
                <SelectItem value="price-desc">{t('Price: High to Low')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Model Cards */}
          {filtered.length === 0 ? (
            <div className="text-center py-16 text-muted-foreground">
              <Search className="h-8 w-8 mx-auto mb-3 opacity-30" />
              <p className="text-lg mb-2">{t('No models found')}</p>
              <p className="text-sm mb-4">{t('try_adjust_filters')}</p>
              <Button variant="outline" onClick={() => {
                setSearch(''); setUseCaseFilter('all'); setSeriesFilter('all');
                setModalityFilter(''); setContextFilter(''); setProviderFilter('');
              }}>
                {t('reset_filters')}
              </Button>
            </div>
          ) : (
            <>
            <div className="grid grid-cols-1 gap-4">
              {displayed.map((m) => {
                const isSelected = selectedModels.has(m.name)
                return (
                  <div key={m.name} className={cn(
                    'rounded-xl border bg-card hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200 overflow-hidden p-6 space-y-4',
                    isSelected && 'ring-2 ring-primary/50 border-primary'
                  )}>
                    {/* Checkbox + Title row */}
                    <div className="flex items-start gap-3">
                      {/* Checkbox */}
                      <button
                        onClick={(e) => { e.stopPropagation(); toggleSelect(m.name) }}
                        className={cn(
                          'shrink-0 mt-0.5 w-5 h-5 rounded border-2 flex items-center justify-center transition-colors',
                          isSelected
                            ? 'bg-primary border-primary text-primary-foreground'
                            : 'border-muted-foreground/30 hover:border-muted-foreground/60'
                        )}
                      >
                        {isSelected && <CheckSquare className="h-3.5 w-3.5" />}
                      </button>

                      {/* Title area (clickable for detail) */}
                      <div
                        className="flex-1 min-w-0 cursor-pointer"
                        onClick={() => openDetail(m)}
                      >
                        <div className="flex items-center justify-between">
                          <div>
                            <h3 className="font-semibold text-lg">{m.name}</h3>
                            <p className="text-xs text-muted-foreground">{m.series}</p>
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* Description */}
                    <p className="text-base text-muted-foreground leading-relaxed">{m.description || t('no_description')}</p>

                    {/* Tags */}
                    <div className="flex flex-wrap items-center gap-1.5">
                      {(() => { const uc = useCaseLabels[m.use_case]; if (!uc) return null; const Icon = uc.icon; return <span className={cn('inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-medium text-white bg-gradient-to-r', uc.color)}><Icon className="h-3 w-3" />{t(uc.label)}</span> })()}
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] bg-muted text-muted-foreground"><Cpu className="h-3 w-3" />{(m.context_window / 1000).toFixed(0)}K ctx</span>
                      {m.input_modalities.slice(0, 3).map(mod => (
                        <span key={mod} className="px-1.5 py-0.5 rounded text-[10px] bg-muted/50 text-muted-foreground">{mod}</span>
                      ))}
                    </div>

                    {/* Pricing + Action */}
                    <div className="flex items-center justify-between pt-2 border-t border-border/40">
                      <span className="text-xs text-muted-foreground">
                        From {m.input_price > 0 ? `$${m.input_price.toFixed(6)}/token` : 'Free tier'}
                      </span>
                      <div className="flex items-center gap-2">
                        <Button
                          size="sm"
                          variant="ghost"
                          className="h-7 text-xs gap-1"
                          onClick={() => openDetail(m)}
                        >
                          <BarChart3 className="h-3 w-3" />{t('Details')}
                        </Button>
                        <Link to={auth.user ? '/chat' : '/sign-in'} search={auth.user ? undefined : { redirect: '/chat' }}>
                          <Button size="sm" className="h-7 text-xs gap-1">
                            <Play className="h-3 w-3" />{t('Call')}
                          </Button>
                        </Link>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
            {visibleCount < filtered.length && (
              <div className="flex justify-center pt-4">
                <Button
                  variant="outline"
                  size="lg"
                  className="gap-2 px-8"
                  onClick={() => setVisibleCount(c => c + PAGE_STEP)}
                >
                  {t('Load More')} ({filtered.length - visibleCount} {t('remaining')})
                </Button>
              </div>
            )}
            </>
          )}
        </div>
      </div>

      {/* Detail Dialog */}
      <ModelDetailDialog
        open={detailOpen}
        onOpenChange={setDetailOpen}
        model={detailModel}
      />

      {/* Comparison Dialog */}
      <ModelComparisonDialog
        open={comparisonOpen}
        onOpenChange={setComparisonOpen}
        models={selectedModelsList}
      />
    </div>
  )
}
