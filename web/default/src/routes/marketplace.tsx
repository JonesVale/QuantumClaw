import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { Store, Search, Star, Server, Tag, Filter, Grid3X3, List, DollarSign, TrendingUp } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { listActiveStores, searchStores, type StoreWithOwner } from '@/lib/api-extended'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/marketplace')({
  component: MarketplacePage,
})

function MarketplacePage() {
  const { t } = useT()
  const [query, setQuery] = useState('')
  const [view, setView] = useState<'grid' | 'list'>('grid')
  const [sortBy, setSortBy] = useState('rating')

  const { data, isLoading } = useQuery({
    queryKey: ['marketplace-stores', query],
    queryFn: () => query ? searchStores(query) : listActiveStores(),
    staleTime: 30_000,
  })
  const stores: StoreWithOwner[] = data?.data || []

  const sorted = useMemo(() => {
    const list = [...stores]
    if (sortBy === 'rating') list.sort((a, b) => b.rating - a.rating)
    else if (sortBy === 'sales') list.sort((a, b) => b.total_sales - a.total_sales)
    else if (sortBy === 'models') list.sort((a, b) => b.active_model - a.active_model)
    return list
  }, [stores, sortBy])

  return (
    <div className="qc-wrapper py-8 space-y-6">
      {/* Header */}
      <div className="text-center max-w-2xl mx-auto">
        <h1 className="text-4xl font-bold mb-3"><Store className="w-8 h-8 inline mr-2" />{t('Model Marketplace')}</h1>
        <p className="text-muted-foreground">{t('Browse AI model providers and find the best API for your needs')}</p>
      </div>

      {/* Search & Filters */}
      <div className="flex flex-col sm:flex-row gap-3 items-center max-w-3xl mx-auto w-full">
        <div className="relative flex-1 w-full">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            className="pl-9"
            placeholder={t('Search stores...')}
            value={query}
            onChange={e => setQuery(e.target.value)}
          />
        </div>
        <Select value={sortBy} onValueChange={setSortBy}>
          <SelectTrigger className="w-32"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="rating">{t('Top Rated')}</SelectItem>
            <SelectItem value="sales">{t('Most Sales')}</SelectItem>
            <SelectItem value="models">{t('Most Models')}</SelectItem>
          </SelectContent>
        </Select>
        <div className="flex gap-1 border rounded-lg p-1">
          <Button variant={view === 'grid' ? 'default' : 'ghost'} size="icon" className="h-8 w-8" onClick={() => setView('grid')}>
            <Grid3X3 className="h-4 w-4" />
          </Button>
          <Button variant={view === 'list' ? 'default' : 'ghost'} size="icon" className="h-8 w-8" onClick={() => setView('list')}>
            <List className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Results */}
      {isLoading ? (
        <div className={view === 'grid' ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4' : 'space-y-3'}>
          {Array.from({length:6}).map((_,i) => <Skeleton key={i} className="h-32 w-full" />)}
        </div>
      ) : sorted.length === 0 ? (
        <Card><CardContent className="py-16 text-center text-muted-foreground">{t('No stores found')}</CardContent></Card>
      ) : view === 'grid' ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {sorted.map(store => <StoreCard key={store.id} store={store} />)}
        </div>
      ) : (
        <div className="space-y-3">
          {sorted.map(store => <StoreListItem key={store.id} store={store} />)}
        </div>
      )}
    </div>
  )
}

function StoreCard({ store }: { store: StoreWithOwner }) {
  return (
    <a href={`/stores/${store.store_slug}`} className="block group">
      <Card className="h-full hover:shadow-md transition-shadow">
        <CardContent className="p-5">
          <div className="flex items-start gap-3 mb-3">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-500/20 to-orange-600/20 flex items-center justify-center text-lg font-bold shrink-0">
              {store.logo ? <img src={store.logo} className="w-10 h-10 rounded-lg object-cover" /> : (store.name?.charAt(0) || 'S')}
            </div>
            <div className="min-w-0 flex-1">
              <h3 className="font-semibold truncate group-hover:text-amber-600 transition-colors">{store.name || store.username}</h3>
              <p className="text-xs text-muted-foreground truncate">@{store.store_slug || store.username}</p>
            </div>
          </div>
          {store.description && <p className="text-xs text-muted-foreground line-clamp-2 mb-3">{store.description}</p>}
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span className="flex items-center gap-1"><Star className="w-3 h-3 text-amber-400" />{store.rating?.toFixed(1) || '-'}</span>
            <span className="flex items-center gap-1"><Server className="w-3 h-3" />{store.active_model} {store.active_model > 1 ? 'models' : 'model'}</span>
            <span className="flex items-center gap-1"><TrendingUp className="w-3 h-3" />{store.total_sales} sales</span>
          </div>
          <div className="mt-3 flex gap-1 flex-wrap">
            <Badge variant={store.status === 1 ? 'default' : 'secondary'} className="text-[10px]">
              {store.status === 1 ? 'Open' : 'Closed'}
            </Badge>
          </div>
        </CardContent>
      </Card>
    </a>
  )
}

function StoreListItem({ store }: { store: StoreWithOwner }) {
  return (
    <a href={`/stores/${store.store_slug}`} className="block group">
      <Card className="hover:shadow-sm transition-shadow">
        <CardContent className="p-4 flex items-center gap-4">
          <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-amber-500/20 to-orange-600/20 flex items-center justify-center shrink-0 text-sm font-bold">
            {store.logo ? <img src={store.logo} className="w-8 h-8 rounded object-cover" /> : (store.name?.charAt(0) || 'S')}
          </div>
          <div className="flex-1 min-w-0">
            <p className="font-medium truncate group-hover:text-amber-600 transition-colors">{store.name || store.username}</p>
            <p className="text-xs text-muted-foreground truncate">{store.description}</p>
          </div>
          <div className="flex items-center gap-3 text-xs text-muted-foreground shrink-0">
            <span>⭐ {store.rating?.toFixed(1)}</span>
            <span>{store.active_model} models</span>
            <span>{store.total_sales} sales</span>
          </div>
          <Badge variant={store.status === 1 ? 'default' : 'secondary'} className="text-[10px]">
            {store.status === 1 ? 'Open' : 'Closed'}
          </Badge>
        </CardContent>
      </Card>
    </a>
  )
}
