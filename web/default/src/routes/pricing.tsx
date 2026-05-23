import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo, useEffect, useRef } from 'react'
import { Search, DollarSign, RefreshCw, Filter } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { cn } from '@/lib/utils'

// 鈹€鈹€ Helper functions 鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€
const PROVIDER_META: Record<string, { gradient: string }> = {
  OpenAI: { gradient: 'from-green-600 to-emerald-700' },
  Anthropic: { gradient: 'from-orange-600 to-amber-700' },
  Google: { gradient: 'from-blue-600 to-indigo-700' },
  DeepSeek: { gradient: 'from-blue-600 to-cyan-700' },
  Meta: { gradient: 'from-indigo-600 to-blue-800' },
  Mistral: { gradient: 'from-cyan-600 to-teal-700' },
  Microsoft: { gradient: 'from-azure-600 to-blue-700' },
  Amazon: { gradient: 'from-orange-600 to-yellow-700' },
  Cohere: { gradient: 'from-purple-600 to-pink-700' },
  Stability: { gradient: 'from-green-600 to-emerald-700' },
}

function getProviderMeta(provider: string): { gradient: string } {
  return PROVIDER_META[provider] || { gradient: 'from-slate-600 to-slate-700' }
}

function formatPrice(price: number): string {
  if (price === 0) return 'Free'
  if (price < 0.000001) return `$${price.toExponential(2)}`
  if (price < 0.001) return `$${price.toFixed(6)}`
  return `$${price.toFixed(4)}`
}

export const Route = createFileRoute('/pricing')({
  component: PricingPage,
})

// 鈹€鈹€ Types 鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€
interface ModelPricing {
  name: string
  provider: string
  input_price: number
  output_price: number
  status: number
}

// 鈹€鈹€ Fallback mock data 鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€鈹€
function PricingPage() {
  const { t } = useT()
  const [search, setSearch] = useState('')
  const [providerFilter, setProviderFilter] = useState('all')
  const [showActiveOnly, setShowActiveOnly] = useState(true)
  const [visibleCount, setVisibleCount] = useState(20)
  const PAGE_STEP = 20

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['model-pricing', 'models'],
    queryFn: async () => {
      const res = await fetch('/api/models')
      if (!res.ok) throw new Error('Failed to fetch')
      return res.json()
    },
    retry: false,
    staleTime: 60 * 1000,
  })

  // Transform API response or use fallback
  const models: ModelPricing[] = useMemo(() => {
    const raw = data?.data
    if (Array.isArray(raw)) return raw as ModelPricing[]
    if (raw && typeof raw === 'object') {
      // Handle dict format: { "1": ["gpt-4", ...] }
      const dict = raw as Record<string, string[]>
      const entries = Object.entries(dict)
      if (entries.length > 0 && Array.isArray(entries[0][1])) {
        return entries.flatMap(([channelId, names]) =>
          (names as string[]).map((name) => ({
            name,
            provider: `Channel ${channelId}`,
            input_price: 0,
            output_price: 0,
            status: 1,
          }))
        )
      }
    }
    return []
  }, [data])

  // Filter
  const filtered = useMemo(() => {
    let list = models
    if (search) {
      const q = search.toLowerCase()
      list = list.filter((m) => m.name.toLowerCase().includes(q) || m.provider.toLowerCase().includes(q))
    }
    if (providerFilter !== 'all') {
      list = list.filter((m) => m.provider === providerFilter)
    }
    if (showActiveOnly) {
      list = list.filter((m) => m.status === 1)
    }
    return list
  }, [models, search, providerFilter, showActiveOnly])

  // Reset pagination when filters change
  useEffect(() => setVisibleCount(PAGE_STEP), [search, providerFilter, showActiveOnly])

  // Unique providers - 鐢?useRef 缂撳瓨锛屽噺灏戦噸澶嶉亶鍘?
  const cachedProviders = useRef<string[]>([])
  const cachedKey = useRef('')
  const currentKey = `${(models||[]).length}-${showActiveOnly}`
  if (cachedKey.current !== currentKey) {
    const pool = showActiveOnly ? models.filter(m => m.status === 1) : models
    const set = new Set(pool.map((m) => m.provider))
    cachedProviders.current = Array.from(set).sort()
    cachedKey.current = currentKey
  }
  const providers = cachedProviders.current

  // Group by provider
  const grouped = useMemo(() => {
    const map = new Map<string, ModelPricing[]>()
    for (const m of filtered) {
      if (!map.has(m.provider)) map.set(m.provider, [])
      map.get(m.provider)!.push(m)
    }
    return map
  }, [filtered])

  return (
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
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              providerFilter === 'all' ? 'bg-primary text-primary-foreground shadow-sm' : 'bg-muted text-muted-foreground hover:bg-accent'
            }`}>
            {t('All Providers')}
          </button>
          {providers.slice(0, 15).map(p => (
            <button key={p} onClick={() => setProviderFilter(p)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                providerFilter === p ? 'bg-primary text-primary-foreground shadow-sm' : 'bg-muted text-muted-foreground hover:bg-accent'
              }`}>
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
                        ${(m.input_price || 0).toFixed(4)}/1K
                      </span>
                      <span className="inline-flex items-center gap-1">
                        <span className="font-medium text-amber-600 dark:text-amber-400">OUT</span>
                        ${(m.output_price || 0).toFixed(4)}/1K
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
  );

}
