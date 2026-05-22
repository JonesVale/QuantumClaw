import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { Search, DollarSign, RefreshCw, Filter } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/pricing')({
  component: PricingPage,
})

// ── Types ─────────────────────────────────────────────────────────
interface ModelPricing {
  name: string
  provider: string
  input_price: number
  output_price: number
  status: number
}

// ── Fallback mock data ────────────────────────────────────────────
const fallbackModels: ModelPricing[] = [
  { name: 'gpt-4', provider: 'OpenAI', input_price: 0.03, output_price: 0.06, status: 1 },
  { name: 'gpt-4-turbo', provider: 'OpenAI', input_price: 0.01, output_price: 0.03, status: 1 },
  { name: 'gpt-3.5-turbo', provider: 'OpenAI', input_price: 0.0015, output_price: 0.002, status: 1 },
  { name: 'gpt-4o', provider: 'OpenAI', input_price: 0.005, output_price: 0.015, status: 1 },
  { name: 'gpt-4o-mini', provider: 'OpenAI', input_price: 0.00015, output_price: 0.0006, status: 1 },
  { name: 'claude-opus-4', provider: 'Anthropic', input_price: 0.015, output_price: 0.075, status: 1 },
  { name: 'claude-3.5-sonnet', provider: 'Anthropic', input_price: 0.003, output_price: 0.015, status: 1 },
  { name: 'claude-3-haiku', provider: 'Anthropic', input_price: 0.00025, output_price: 0.00125, status: 1 },
  { name: 'gemini-pro', provider: 'Google', input_price: 0.001, output_price: 0.002, status: 1 },
  { name: 'gemini-1.5-pro', provider: 'Google', input_price: 0.0035, output_price: 0.0105, status: 1 },
  { name: 'gemini-1.5-flash', provider: 'Google', input_price: 0.000075, output_price: 0.0003, status: 1 },
  { name: 'deepseek-chat', provider: 'DeepSeek', input_price: 0.0005, output_price: 0.002, status: 1 },
  { name: 'deepseek-reasoner', provider: 'DeepSeek', input_price: 0.0005, output_price: 0.002, status: 1 },
  { name: 'qwen-max', provider: 'Alibaba', input_price: 0.002, output_price: 0.006, status: 1 },
  { name: 'qwen-plus', provider: 'Alibaba', input_price: 0.0008, output_price: 0.002, status: 1 },
  { name: 'qwen-turbo', provider: 'Alibaba', input_price: 0.0003, output_price: 0.0006, status: 1 },
  { name: 'mistral-large', provider: 'Mistral', input_price: 0.004, output_price: 0.012, status: 1 },
  { name: 'mistral-medium', provider: 'Mistral', input_price: 0.002, output_price: 0.006, status: 1 },
  { name: 'llama-3.1-70b', provider: 'Meta', input_price: 0.00059, output_price: 0.00079, status: 1 },
  { name: 'llama-3.1-405b', provider: 'Meta', input_price: 0.002, output_price: 0.002, status: 1 },
]

// ── Provider icons / colors ───────────────────────────────────────
const PROVIDER_META: Record<string, { color: string; gradient: string }> = {
  OpenAI: { color: '#10a37f', gradient: 'from-emerald-500 to-teal-600' },
  Anthropic: { color: '#d97706', gradient: 'from-amber-500 to-orange-600' },
  Google: { color: '#4285f4', gradient: 'from-blue-500 to-indigo-600' },
  DeepSeek: { color: '#4f46e5', gradient: 'from-indigo-500 to-purple-600' },
  Alibaba: { color: '#ea580c', gradient: 'from-orange-500 to-red-600' },
  Mistral: { color: '#7c3aed', gradient: 'from-violet-500 to-purple-600' },
  Meta: { color: '#0ea5e9', gradient: 'from-sky-500 to-cyan-600' },
}

function getProviderMeta(provider: string): { color: string; gradient: string } {
  return PROVIDER_META[provider] || { color: '#6b7280', gradient: 'from-gray-500 to-gray-600' }
}

function formatPrice(price: number): string {
  if (price === 0) return 'Free'
  if (price < 0.001) return `$${price.toFixed(6)}`
  if (price < 0.01) return `$${price.toFixed(4)}`
  return `$${price.toFixed(3)}`
}

function PricingPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [providerFilter, setProviderFilter] = useState('all')
  const [showActiveOnly, setShowActiveOnly] = useState(true)

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
    return fallbackModels
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

  // Unique providers
  const providers = useMemo(() => {
    const set = new Set(models.map((m) => m.provider))
    return Array.from(set).sort()
  }, [models])

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
    <div className="w-full p-4 sm:p-6 space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            💰 {t('Model Pricing')}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm sm:text-base lg:text-lg">
            {t('Transparent pricing across all providers')}
          </p>
        </div>
        <Button variant="outline" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`mr-2 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('Refresh')}
        </Button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="pl-9"
            placeholder={t('Search models...')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Select value={providerFilter} onValueChange={setProviderFilter}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder={t('All Providers')} />
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
        <Button
          variant={showActiveOnly ? 'default' : 'outline'}
          size="sm"
          onClick={() => setShowActiveOnly(!showActiveOnly)}
          className="gap-2"
        >
          <Filter className="h-4 w-4" />
          {t('Active only')}
        </Button>
      </div>

      {/* Pricing Cards by Provider */}
      {isLoading ? (
        <div className="space-y-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <Skeleton className="h-6 w-32" />
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  {Array.from({ length: 3 }).map((_, j) => (
                    <Skeleton key={j} className="h-8 w-full" />
                  ))}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : grouped.size === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            <DollarSign className="h-12 w-12 mx-auto mb-3 opacity-30" />
            {t('No models found')}
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {Array.from(grouped.entries()).map(([provider, providerModels]) => {
            const meta = getProviderMeta(provider)
            return (
              <Card key={provider} className="overflow-hidden">
                {/* Provider Header */}
                <div className={cn('bg-gradient-to-r p-4 text-white', meta.gradient)}>
                  <div className="flex items-center gap-2">
                    <div className="flex items-center justify-center w-8 h-8 rounded-lg bg-white/20 backdrop-blur">
                      <span className="text-lg font-bold">{provider[0]}</span>
                    </div>
                    <CardTitle className="text-lg">{provider}</CardTitle>
                    <Badge variant="secondary" className="ml-auto bg-white/20 text-white hover:bg-white/30">
                      {providerModels.length} {t('Models')}
                    </Badge>
                  </div>
                </div>

                {/* Price Table */}
                <CardContent className="p-0">
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b bg-muted/30">
                          <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase tracking-wider">
                            {t('Model')}
                          </th>
                          <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase tracking-wider">
                            {t('Input Price')}
                          </th>
                          <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase tracking-wider">
                            {t('Output Price')}
                          </th>
                          <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3 uppercase tracking-wider">
                            {t('Status')}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {providerModels.map((m, idx) => (
                          <tr
                            key={m.name}
                            className={cn(
                              'border-b border-muted/50 transition-colors hover:bg-muted/20',
                              idx % 2 === 0 && 'bg-muted/10'
                            )}
                          >
                            <td className="px-4 py-3">
                              <span className="font-medium text-sm">{m.name}</span>
                            </td>
                            <td className="px-4 py-3 text-right">
                              <code className="text-sm font-mono text-blue-600 dark:text-blue-400">
                                {m.input_price > 0 ? `${formatPrice(m.input_price)}/1K` : '-'}
                              </code>
                            </td>
                            <td className="px-4 py-3 text-right">
                              <code className="text-sm font-mono text-purple-600 dark:text-purple-400">
                                {m.output_price > 0 ? `${formatPrice(m.output_price)}/1K` : '-'}
                              </code>
                            </td>
                            <td className="px-4 py-3 text-center">
                              {m.status === 1 ? (
                                <Badge variant="outline" className="bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-400 border-green-200 dark:border-green-800">
                                  {t('Active')}
                                </Badge>
                              ) : (
                                <Badge variant="outline" className="bg-red-50 text-red-700 dark:bg-red-900/30 dark:text-red-400 border-red-200 dark:border-red-800">
                                  {t('Disabled')}
                                </Badge>
                              )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
