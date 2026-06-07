import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { Store, Star, Server, Tag, Search, ExternalLink, ShoppingCart, DollarSign, TrendingUp, Filter } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { getPublicStore, type StoreModelItem } from '@/lib/api-extended'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/stores/$slug')({
  component: PublicStorePage,
  loader: async ({ params }) => params,
})

function PublicStorePage() {
  const { t } = useT()
  const { slug } = Route.useParams()
  const [tagFilter, setTagFilter] = useState('all')
  const [viewMode, setViewMode] = useState<'price' | 'multiplier' | 'table'>('price')
  const [showAvailableOnly, setShowAvailableOnly] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['public-store', slug],
    queryFn: () => getPublicStore(slug),
    enabled: !!slug,
  })

  const store = data?.data?.store
  const owner = data?.data?.owner
  const models: StoreModelItem[] = data?.data?.models || []

  // Extract all unique tags
  const allTags = useMemo(() => {
    const tagSet = new Set<string>()
    models.forEach(m => (m.tags || '').split(',').filter(Boolean).forEach(t => tagSet.add(t.trim())))
    return Array.from(tagSet)
  }, [models])

  const filtered = useMemo(() => {
    let list = models
    if (tagFilter !== 'all') list = list.filter(m => (m.tags || '').includes(tagFilter))
    if (showAvailableOnly) list = list.filter(m => m.is_active)
    return list
  }, [models, tagFilter, showAvailableOnly])

  const formatPrice = (price: number) => {
    if (!price) return t('Based on multiplier')
    return `¥${(price / 100).toFixed(4)} / 1M tokens`
  }

  // KK-AI style discount label: PriceMultiplier → 官方价X折
  const getDiscountLabel = (multiplier: number): string | null => {
    if (multiplier <= 0) return null
    if (multiplier >= 1) {
      const pct = Math.round(multiplier * 10)
      return `官方价${pct}折`
    }
    const pct = Math.round(multiplier * 10)
    return `官方价${pct}折`
  }

  if (isLoading) return <div className="qc-wrapper py-8 space-y-6"><Skeleton className="h-48 w-full" /><Skeleton className="h-64 w-full" /></div>
  if (!store) return <div className="qc-wrapper py-16 text-center text-muted-foreground">{t('Store not found')}</div>

  return (
    <div className="qc-wrapper py-8 space-y-6">
      {/* Store Header — inspired by KK-AI pricing page layout */}
      <Card className="bg-gradient-to-br from-amber-50/80 to-white border-amber-200/50 overflow-hidden">
        {store.banner_url && (
          <div className="h-32 sm:h-48 w-full overflow-hidden">
            <img src={store.banner_url} className="w-full h-full object-cover" alt="" />
          </div>
        )}
        <CardContent className={cn('p-6', store.banner_url ? '-mt-16 relative' : '')}>
          <div className="flex items-start gap-4">
            <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-amber-500/30 to-orange-600/30 flex items-center justify-center text-2xl font-bold shadow-lg shrink-0 bg-white">
              {store.logo ? <img src={store.logo} className="w-14 h-14 rounded-xl object-cover" /> : (store.name?.charAt(0) || 'S')}
            </div>
            <div className="flex-1 min-w-0">
              <h1 className="text-2xl font-bold mb-1">{store.name || owner?.username}</h1>
              {store.description && <p className="text-sm text-muted-foreground mb-3">{store.description}</p>}
              <div className="flex flex-wrap items-center gap-4 text-sm">
                <span className="flex items-center gap-1"><Star className="w-4 h-4 text-amber-400" />{store.rating?.toFixed(1) || t('New')}</span>
                <span className="flex items-center gap-1"><Server className="w-4 h-4" />{models.length} {t('models')}</span>
                <span className="flex items-center gap-1"><TrendingUp className="w-4 h-4" />{store.total_sales} {t('sales')}</span>
                {owner?.username && <span className="text-xs text-muted-foreground">by @{owner.username}</span>}
              </div>
              {store.contact_info && <p className="text-xs text-muted-foreground mt-2">📧 {store.contact_info}</p>}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Filter Bar — KK-AI style */}
      <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center justify-between">
        <div className="flex gap-2 flex-wrap">
          <Badge
            variant={tagFilter === 'all' ? 'default' : 'outline'}
            className="cursor-pointer"
            onClick={() => setTagFilter('all')}
          >{t('All Models')} ({models.length})</Badge>
          {allTags.map(tag => (
            <Badge
              key={tag}
              variant={tagFilter === tag ? 'default' : 'outline'}
              className="cursor-pointer"
              onClick={() => setTagFilter(tag)}
            >{tag}</Badge>
          ))}
        </div>
        {/* View Mode Toggle */}
        <div className="flex items-center gap-3 flex-wrap">
          <label className="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer">
            <input type="checkbox" checked={showAvailableOnly} onChange={e => setShowAvailableOnly(e.target.checked)} className="rounded" />
            {t('Show available only')}
          </label>
          <div className="flex gap-1 border rounded-lg p-1">
          <Button variant={viewMode === 'price' ? 'default' : 'ghost'} size="sm" className="h-7 text-xs" onClick={() => setViewMode('price')}>
            <DollarSign className="w-3 h-3 mr-1" />{t('Price')}
          </Button>
          <Button variant={viewMode === 'multiplier' ? 'default' : 'ghost'} size="sm" className="h-7 text-xs" onClick={() => setViewMode('multiplier')}>
            <TrendingUp className="w-3 h-3 mr-1" />{t('Multiplier')}
          </Button>
          <Button variant={viewMode === 'table' ? 'default' : 'ghost'} size="sm" className="h-7 text-xs" onClick={() => setViewMode('table')}>
            <Filter className="w-3 h-3 mr-1" />{t('Table')}
          </Button>
        </div>
      </div>

      {/* Model Cards — inspired by KK-AI */}
      {viewMode === 'table' ? (
        <Card><CardContent className="p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/30">
                <th className="text-left px-4 py-3 font-medium">{t('Model')}</th>
                <th className="text-right px-4 py-3 font-medium">{t('Input Price')}</th>
                <th className="text-right px-4 py-3 font-medium">{t('Output Price')}</th>
                <th className="text-right px-4 py-3 font-medium">{t('Cache Read')}</th>
                <th className="text-center px-4 py-3 font-medium">{t('Multiplier')}</th>
                <th className="text-center px-4 py-3 font-medium">{t('Tags')}</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(m => (
                <tr key={m.id} className="border-b border-muted/50 hover:bg-muted/20">
                  <td className="px-4 py-3"><span className="font-medium">{m.display_name || m.model_name}</span><div className="text-xs text-muted-foreground">{m.model_name}</div></td>
                  <td className="px-4 py-3 text-right font-mono text-xs">{formatPrice(m.input_price)}</td>
                  <td className="px-4 py-3 text-right font-mono text-xs">{formatPrice(m.output_price)}</td>
                  <td className="px-4 py-3 text-right font-mono text-xs">{formatPrice(m.cache_read_price)}</td>
                  <td className="px-4 py-3 text-center"><Badge variant="outline">{m.price_multiplier?.toFixed(2)}x</Badge></td>
                  <td className="px-4 py-3"><div className="flex gap-1 flex-wrap justify-center">{m.tags?.split(',').map((t,i) => <Badge key={i} variant="secondary" className="text-[10px]">{t.trim()}</Badge>)}</div></td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent></Card>
      ) : (
        <div className="grid gap-4">
          {filtered.map(m => (
            <Card key={m.id} className="hover:shadow-sm transition-shadow">
              <CardContent className="p-5">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h3 className="font-semibold">{m.display_name || m.model_name}</h3>
                    <p className="text-xs text-muted-foreground font-mono">{m.model_name}</p>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    {m.price_multiplier && m.price_multiplier !== 1 && (
                      <Badge variant="secondary" className="text-[10px] bg-amber-100 text-amber-700 border-amber-200">
                        {getDiscountLabel(m.price_multiplier)}
                      </Badge>
                    )}
                    <Badge variant={viewMode === 'multiplier' ? 'secondary' : 'outline'} className="text-xs">
                      {viewMode === 'multiplier' ? `${m.price_multiplier?.toFixed(2)}x` : t('Pay as you go')}
                    </Badge>
                  </div>
                </div>
                {viewMode === 'price' ? (
                  <div className="grid grid-cols-3 gap-3 mb-3">
                    <div><p className="text-xs text-muted-foreground">{t('Input')}</p><p className="font-mono text-sm font-medium">{formatPrice(m.input_price)}</p></div>
                    <div><p className="text-xs text-muted-foreground">{t('Output')}</p><p className="font-mono text-sm font-medium">{formatPrice(m.output_price)}</p></div>
                    <div><p className="text-xs text-muted-foreground">{t('Cache Read')}</p><p className="font-mono text-sm font-medium">{formatPrice(m.cache_read_price)}</p></div>
                  </div>
                ) : (
                  <div className="grid grid-cols-2 gap-3 mb-3">
                    <div><p className="text-xs text-muted-foreground">{t('Multiplier')}</p><p className="text-lg font-bold text-amber-600">{m.price_multiplier?.toFixed(2)}x</p></div>
                    <div><p className="text-xs text-muted-foreground">{t('Channel ID')}</p><p className="text-sm font-mono">#{m.channel_id}</p></div>
                  </div>
                )}
                {m.description && <p className="text-xs text-muted-foreground mb-3">{m.description}</p>}
                <div className="flex gap-1 flex-wrap">
                  {m.tags?.split(',').filter(Boolean).map((tag, i) => <Badge key={i} variant="secondary" className="text-[10px]">{tag.trim()}</Badge>)}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {filtered.length === 0 && (
        <Card><CardContent className="py-12 text-center text-muted-foreground">{t('No models match this filter')}</CardContent></Card>
      )}
    </div>
    </div>
  )
}
