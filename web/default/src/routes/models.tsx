import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import {
  Search, X, SlidersHorizontal, ArrowUpDown, RefreshCw,
  ExternalLink
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { getEnhancedModels, type EnhancedModel } from '@/lib/api-extended'

export const Route = createFileRoute('/models')({
  component: ModelsPage,
})

type SortOption = 'provider' | 'name' | 'price-asc' | 'price-desc'

function ModelsPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [providerFilter, setProviderFilter] = useState('all')
  const [sortBy, setSortBy] = useState<SortOption>('provider')
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [selectedModel, setSelectedModel] = useState<EnhancedModel | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['enhanced-models'],
    queryFn: getEnhancedModels,
    staleTime: 60 * 1000,
  })

  const models: EnhancedModel[] = data?.data || []

  // Derive all filter options from actual API data
  const providers = useMemo(() => {
    return [...new Set(models.map(m => m.provider).filter(Boolean))].sort()
  }, [models])

  const priceRange = useMemo(() => {
    if (models.length === 0) return { min: 0, max: 0 }
    const prices = models.map(m => m.input_price).filter(p => p > 0)
    return {
      min: Math.min(...prices),
      max: Math.max(...prices),
    }
  }, [models])

  // Filter and sort
  const filtered = useMemo(() => {
    let result = models

    if (search) {
      const q = search.toLowerCase()
      result = result.filter(m =>
        m.name.toLowerCase().includes(q) ||
        m.provider.toLowerCase().includes(q) ||
        m.channel_name.toLowerCase().includes(q)
      )
    }

    if (providerFilter !== 'all') {
      result = result.filter(m => m.provider === providerFilter)
    }

    switch (sortBy) {
      case 'name': result.sort((a, b) => a.name.localeCompare(b.name)); break
      case 'price-asc': result.sort((a, b) => a.input_price - b.input_price); break
      case 'price-desc': result.sort((a, b) => b.input_price - a.input_price); break
      default: result.sort((a, b) => a.provider.localeCompare(b.provider) || a.name.localeCompare(b.name)); break
    }

    return result
  }, [models, search, providerFilter, sortBy])

  const grouped = useMemo(() => {
    return filtered.reduce<Record<string, EnhancedModel[]>>((acc, m) => {
      const key = m.provider || 'Other'
      if (!acc[key]) acc[key] = []
      acc[key].push(m)
      return acc
    }, {})
  }, [filtered])

  const hasFilters = search || providerFilter !== 'all'

  return (
    <div className="flex min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Sidebar Filters - derived from actual data */}
      <aside className={cn(
        'w-64 shrink-0 border-r bg-card/50 backdrop-blur-sm flex flex-col transition-all',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full fixed z-10 h-full md:relative md:translate-x-0'
      )}>
        <div className="p-4 border-b flex items-center justify-between">
          <span className="font-semibold text-sm flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4" />
            {t('Filters')}
          </span>
          <Button variant="ghost" size="icon" className="h-6 w-6 md:hidden" onClick={() => setSidebarOpen(false)}>
            <X className="h-3.5 w-3.5" />
          </Button>
        </div>
        <ScrollArea className="flex-1 p-4 space-y-5">
          <div>
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
              {t('Provider')}
            </h4>
            <div className="space-y-1">
              <label className="flex items-center gap-2 py-1 cursor-pointer text-xs hover:text-foreground transition-colors">
                <input type="radio" name="provider" checked={providerFilter === 'all'} onChange={() => setProviderFilter('all')} className="accent-primary" />
                {t('All Providers')} ({models.length})
              </label>
              {providers.map(p => (
                <label key={p} className="flex items-center gap-2 py-1 cursor-pointer text-xs hover:text-foreground transition-colors">
                  <input type="radio" name="provider" checked={providerFilter === p} onChange={() => setProviderFilter(p)} className="accent-primary" />
                  {p} ({models.filter(m => m.provider === p).length})
                </label>
              ))}
            </div>
          </div>

          {models.length > 0 && (
            <div>
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                {t('Pricing Range')}
              </h4>
              <p className="text-[10px] text-muted-foreground">
                ${priceRange.min.toFixed(6)} — ${priceRange.max.toFixed(6)} /token
              </p>
            </div>
          )}
        </ScrollArea>
      </aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Top Bar */}
        <div className="flex items-center gap-2 px-4 py-2 border-b bg-card/50 shrink-0">
          <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => setSidebarOpen(!sidebarOpen)}>
            <SlidersHorizontal className="h-4 w-4" />
          </Button>
          <span className="text-xs text-muted-foreground font-medium">
            {models.length} {t('Models')}
          </span>
          {hasFilters && (
            <Badge variant="secondary" className="text-[10px] cursor-pointer" onClick={() => { setSearch(''); setProviderFilter('all') }}>
              <X className="h-2.5 w-2.5 mr-0.5" />
              {t('Clear')}
            </Badge>
          )}
        </div>

        <div className="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
          {/* Header */}
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div>
              <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
                🧠 {t('Models')}
              </h1>
              <p className="text-muted-foreground mt-1 text-sm">
                {models.length} {t('available via configured channels')}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" className="gap-1 h-8 text-xs" onClick={() => refetch()} disabled={isFetching}>
                <RefreshCw className={`h-3.5 w-3.5 ${isFetching ? 'animate-spin' : ''}`} />
                {t('Refresh')}
              </Button>
            </div>
          </div>

          {/* Search + Sort */}
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1 sm:max-w-md">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="pl-9 h-9 text-sm"
                placeholder={t('Search models...')}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <Select value={sortBy} onValueChange={(v) => setSortBy(v as SortOption)}>
              <SelectTrigger className="w-[140px] h-9 text-xs">
                <ArrowUpDown className="h-3.5 w-3.5 mr-1" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="provider">{t('By Provider')}</SelectItem>
                <SelectItem value="name">{t('By Name')}</SelectItem>
                <SelectItem value="price-asc">{t('Price: Low to High')}</SelectItem>
                <SelectItem value="price-desc">{t('Price: High to Low')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Model List */}
          {isLoading ? (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <Card key={i}><CardContent className="p-4 space-y-3">
                  <div className="h-4 w-24 animate-pulse rounded bg-muted" />
                  <div className="h-3 w-32 animate-pulse rounded bg-muted" />
                </CardContent></Card>
              ))}
            </div>
          ) : filtered.length === 0 ? (
            <Card><CardContent className="py-16 text-center">
              <Search className="h-8 w-8 mx-auto mb-3 opacity-30" />
              <p className="text-muted-foreground">{models.length === 0 ? t('No models available. Configure a channel first.') : t('No models found')}</p>
              <a href="/channels" className="text-sm text-blue-500 hover:underline mt-2 inline-block">{t('Go to Channels')}</a>
            </CardContent></Card>
          ) : (
            <div className="space-y-6">
              {Object.entries(grouped).map(([provider, providerModels]) => (
                <section key={provider}>
                  <div className="flex items-center gap-2 mb-3 px-1">
                    <h2 className="text-sm font-semibold">{provider}</h2>
                    <Badge variant="secondary" className="text-[10px]">{providerModels.length}</Badge>
                  </div>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {providerModels.map((m) => (
                      <Card key={`${m.channel_id}-${m.name}`} className="hover:shadow-md transition-all group" onClick={() => { setSelectedModel(m); setDetailOpen(true) }}>
                        <CardContent className="p-3 space-y-1.5">
                          <p className="text-sm font-semibold truncate group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">{m.name}</p>
                          <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
                            <span>{t('Input')}: ${m.input_price?.toFixed(6)}</span>
                            <span>·</span>
                            <span>{t('Output')}: ${m.output_price?.toFixed(6)}</span>
                          </div>
                          <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
                            <Badge variant="outline" className={cn('text-[9px]', m.status === 1 ? 'bg-green-500/10 text-green-600 border-green-500/30' : 'bg-red-500/10 text-red-600 border-red-500/30')}>
                              {m.status === 1 ? '● Active' : '○ Inactive'}
                            </Badge>
                            <span>{m.group}</span>
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                </section>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
