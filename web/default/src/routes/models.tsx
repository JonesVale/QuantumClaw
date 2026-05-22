import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { Search, RefreshCw, Filter, ChevronDown } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ProviderIcon } from '@/components/provider-icon'
import { ModelCard } from '@/components/model-card'
import { ModelDetailDialog } from '@/components/model-detail-dialog'
import { getEnhancedModels, type EnhancedModel } from '@/lib/api-extended'

export const Route = createFileRoute('/models')({
  component: ModelsPage,
})

type SortOption = 'name' | 'provider' | 'price-asc' | 'price-desc'

function ModelsPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [providerFilter, setProviderFilter] = useState('all')
  const [sortBy, setSortBy] = useState<SortOption>('provider')
  const [selectedModel, setSelectedModel] = useState<EnhancedModel | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['enhanced-models'],
    queryFn: getEnhancedModels,
    staleTime: 60 * 1000,
  })

  const models: EnhancedModel[] = data?.data || []

  // Derive unique providers for filter
  const providers = useMemo(() => {
    const set = new Set(models.map((m) => m.provider).filter(Boolean))
    return Array.from(set).sort()
  }, [models])

  // Derive provider types for tag bar (top N + rest in a dropdown)
  const providerTypes = useMemo(() => {
    const map = new Map<number, { name: string; count: number }>()
    for (const m of models) {
      const key = m.provider_type
      if (!map.has(key)) {
        map.set(key, { name: m.provider, count: 0 })
      }
      map.get(key)!.count++
    }
    return Array.from(map.entries())
      .sort((a, b) => b[1].count - a[1].count)
  }, [models])

  const topProviderTypes = providerTypes.slice(0, 6)

  // Filter & sort
  const filtered = useMemo(() => {
    let result = models

    // Search filter
    if (search) {
      const q = search.toLowerCase()
      result = result.filter(
        (m) =>
          m.name.toLowerCase().includes(q) ||
          m.provider.toLowerCase().includes(q) ||
          m.channel_name.toLowerCase().includes(q)
      )
    }

    // Provider filter
    if (providerFilter !== 'all') {
      result = result.filter((m) => m.provider === providerFilter)
    }

    // Sort
    switch (sortBy) {
      case 'name':
        result.sort((a, b) => a.name.localeCompare(b.name))
        break
      case 'provider':
        result.sort((a, b) => a.provider.localeCompare(b.provider) || a.name.localeCompare(b.name))
        break
      case 'price-asc':
        result.sort((a, b) => a.input_price - b.input_price)
        break
      case 'price-desc':
        result.sort((a, b) => b.input_price - a.input_price)
        break
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

  const totalModels = models.length
  const totalProviders = providers.length

  const handleDetail = (model: EnhancedModel) => {
    setSelectedModel(model)
    setDetailOpen(true)
  }

  return (
    <div className="w-full p-4 sm:p-6 space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            🧠 {t('Models')}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm sm:text-base lg:text-lg">
            {t('{count} active models across {providers} providers', {
              count: totalModels,
              providers: totalProviders,
            })}
          </p>
        </div>
        <Button variant="outline" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`mr-2 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('Refresh')}
        </Button>
      </div>

      {/* Filters */}
      <div className="space-y-3">
        {/* Search + Provider dropdown + Sort */}
        <div className="flex flex-col sm:flex-row gap-3">
          <div className="relative flex-1 sm:max-w-sm">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              className="pl-9"
              placeholder={t('Search models...')}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>

          <div className="flex gap-2 flex-wrap">
            <Select value={providerFilter} onValueChange={(v) => setProviderFilter(v)}>
              <SelectTrigger className="w-[140px]">
                <Filter className="h-3.5 w-3.5 mr-1" />
                <SelectValue placeholder={t('Provider')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('All Providers')}</SelectItem>
                {providers.map((p) => (
                  <SelectItem key={p} value={p}>
                    {p}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select value={sortBy} onValueChange={(v) => setSortBy(v as SortOption)}>
              <SelectTrigger className="w-[130px]">
                <ChevronDown className="h-3.5 w-3.5 mr-1" />
                <SelectValue placeholder={t('Sort')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="provider">{t('By Provider')}</SelectItem>
                <SelectItem value="name">{t('By Name')}</SelectItem>
                <SelectItem value="price-asc">{t('Price: Low to High')}</SelectItem>
                <SelectItem value="price-desc">{t('Price: High to Low')}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Provider type tags */}
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-xs font-medium text-muted-foreground mr-1">{t('Tags')}:</span>
          <Badge
            variant={providerFilter === 'all' ? 'default' : 'outline'}
            className="cursor-pointer text-xs"
            onClick={() => setProviderFilter('all')}
          >
            {t('All')}
          </Badge>
          {topProviderTypes.map(([type, info]) => (
            <Badge
              key={type}
              variant={providerFilter === info.name ? 'default' : 'outline'}
              className="cursor-pointer text-xs gap-1"
              onClick={() => setProviderFilter(info.name)}
            >
              <ProviderIcon type={type} size="sm" className="!text-xs" />
              {info.name}
            </Badge>
          ))}
        </div>
      </div>

      {/* Content */}
      {isLoading ? (
        // Loading skeleton
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <Card key={i}>
              <CardContent className="p-4 sm:p-5 space-y-3">
                <div className="flex items-center gap-2">
                  <div className="h-8 w-8 rounded-full animate-pulse bg-muted" />
                  <div className="h-3 w-16 animate-pulse rounded bg-muted" />
                </div>
                <div className="h-5 w-3/4 animate-pulse rounded bg-muted" />
                <div className="flex items-center justify-between">
                  <div className="flex gap-2">
                    <div className="h-5 w-16 animate-pulse rounded bg-muted" />
                    <div className="h-5 w-16 animate-pulse rounded bg-muted" />
                  </div>
                  <div className="h-5 w-14 animate-pulse rounded bg-muted" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : filtered.length === 0 ? (
        // Empty state
        <Card>
          <CardContent className="py-16 text-center">
            <div className="text-5xl mb-4 opacity-30">🔍</div>
            <h3 className="text-lg font-semibold mb-1">{t('No models found')}</h3>
            <p className="text-sm text-muted-foreground max-w-sm mx-auto">
              {search || providerFilter !== 'all'
                ? t('Try adjusting your search or filter criteria')
                : t('No models are currently available')}
            </p>
            {(search || providerFilter !== 'all') && (
              <Button
                variant="outline"
                size="sm"
                className="mt-4"
                onClick={() => {
                  setSearch('')
                  setProviderFilter('all')
                }}
              >
                {t('Clear filters')}
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        // Grouped model cards
        <div className="space-y-8">
          {Object.entries(grouped).map(([provider, providerModels]) => (
            <section key={provider}>
              {/* Provider group header */}
              <div className="flex items-center gap-2 mb-3">
                <ProviderIcon
                  type={providerModels[0]?.provider_type}
                  name={provider}
                  size="sm"
                />
                <h2 className="text-lg font-semibold tracking-tight">{provider}</h2>
                <Badge variant="secondary" className="text-xs font-normal">
                  {providerModels.length} {providerModels.length === 1 ? t('model') : t('models')}
                </Badge>
              </div>

              {/* Cards grid */}
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {providerModels.map((m) => (
                  <ModelCard key={`${m.channel_id}-${m.name}`} model={m} onDetail={handleDetail} />
                ))}
              </div>
            </section>
          ))}
        </div>
      )}

      {/* Detail Dialog */}
      <ModelDetailDialog
        open={detailOpen}
        onOpenChange={setDetailOpen}
        model={selectedModel}
      />
    </div>
  )
}
