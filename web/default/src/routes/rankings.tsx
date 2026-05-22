import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { RefreshCw, Trophy, TrendingUp, TrendingDown, Zap, Clock, DollarSign, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { ModelRanking } from '@/lib/api-extended'

export const Route = createFileRoute('/rankings')({
  component: RankingsPage,
})

// ── Tab definitions ───────────────────────────────────────────────
type SortKey = 'request_count_7d' | 'tokens_7d' | 'avg_speed_ms' | 'price_per_1k'

interface TabItem {
  key: SortKey
  icon: React.ElementType
  labelKey: string
  sortDir: 'desc' | 'asc'
}

const TABS: TabItem[] = [
  { key: 'request_count_7d', icon: Zap, labelKey: 'Request Count', sortDir: 'desc' },
  { key: 'tokens_7d', icon: Sparkles, labelKey: 'Token Consumption', sortDir: 'desc' },
  { key: 'avg_speed_ms', icon: Clock, labelKey: 'Response Speed', sortDir: 'asc' },
  { key: 'price_per_1k', icon: DollarSign, labelKey: 'Price', sortDir: 'asc' },
]

// ── Fallback mock data ────────────────────────────────────────────
const fallbackRankings: ModelRanking[] = [
  { model: 'gpt-4', provider: 'OpenAI', channel_name: 'OpenAI', tokens_7d: 1200000, trend_percent: 22, avg_speed_ms: 1200, price_per_1k: 0.03, request_count_7d: 45000 },
  { model: 'claude-opus-4', provider: 'Anthropic', channel_name: 'Anthropic', tokens_7d: 987000, trend_percent: 15, avg_speed_ms: 1500, price_per_1k: 0.015, request_count_7d: 32000 },
  { model: 'gemini-pro', provider: 'Google', channel_name: 'Google', tokens_7d: 654000, trend_percent: -3, avg_speed_ms: 800, price_per_1k: 0.001, request_count_7d: 28000 },
  { model: 'deepseek-chat', provider: 'DeepSeek', channel_name: 'DeepSeek', tokens_7d: 520000, trend_percent: 45, avg_speed_ms: 600, price_per_1k: 0.0005, request_count_7d: 22000 },
  { model: 'qwen-max', provider: 'Alibaba', channel_name: 'Alibaba', tokens_7d: 430000, trend_percent: 30, avg_speed_ms: 750, price_per_1k: 0.002, request_count_7d: 18000 },
  { model: 'claude-3-haiku', provider: 'Anthropic', channel_name: 'Anthropic', tokens_7d: 380000, trend_percent: 8, avg_speed_ms: 900, price_per_1k: 0.0025, request_count_7d: 15000 },
  { model: 'gpt-3.5-turbo', provider: 'OpenAI', channel_name: 'OpenAI', tokens_7d: 320000, trend_percent: -12, avg_speed_ms: 600, price_per_1k: 0.0015, request_count_7d: 12000 },
  { model: 'mistral-large', provider: 'Mistral', channel_name: 'Mistral', tokens_7d: 210000, trend_percent: 5, avg_speed_ms: 1100, price_per_1k: 0.004, request_count_7d: 9000 },
]

// ── Rank badge colors ─────────────────────────────────────────────
const RANK_STYLES = [
  'bg-gradient-to-br from-yellow-400 to-yellow-600 shadow-yellow-500/30',  // 1st - Gold
  'bg-gradient-to-br from-gray-300 to-gray-400 shadow-gray-400/30',        // 2nd - Silver
  'bg-gradient-to-br from-amber-600 to-amber-800 shadow-amber-700/30',     // 3rd - Bronze
]

// ── Format helpers ────────────────────────────────────────────────
function formatRequests(n: number): string {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return n.toString()
}

function formatTokens(n: number): string {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return n.toString()
}

function RankingsPage() {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState<SortKey>('request_count_7d')

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['model-rankings'],
    queryFn: async () => {
      const res = await fetch('/api/models/rankings')
      if (!res.ok) throw new Error('Failed to fetch')
      return res.json()
    },
    retry: false,
    staleTime: 60 * 1000,
  })

  const rankings: ModelRanking[] = data?.data ?? fallbackRankings

  // Sort based on active tab
  const activeTabDef = TABS.find((t) => t.key === activeTab)!
  const sorted = [...rankings].sort((a, b) => {
    const aVal = a[activeTab]
    const bVal = b[activeTab]
    return activeTabDef.sortDir === 'desc' ? bVal - aVal : aVal - bVal
  })

  return (
    <div className="flex min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      <aside className="w-56 shrink-0 border-r bg-card/50 backdrop-blur-sm hidden lg:block">
        <div className="p-4 border-b">
          <span className="font-semibold text-sm">{t('Filters')}</span>
        </div>
        <div className="p-4 space-y-4">
          <div>
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">{t('Series')}</h4>
            <div className="space-y-0.5">
              {['All','GPT-4','Claude 3','Gemini','DeepSeek','Mistral','Llama'].map(s => (
                <button key={s} className="w-full text-left px-2 py-1.5 text-xs rounded hover:bg-muted/50 text-muted-foreground transition-colors">{s}</button>
              ))}
            </div>
          </div>
        </div>
      </aside>
      <div className="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            🏆 {t('Model Rankings')}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm sm:text-base lg:text-lg">
            {t('Model performance rankings across providers')}
          </p>
        </div>
        <Button variant="outline" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`mr-2 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('Refresh')}
        </Button>
      </div>

      {/* Tab Bar */}
      <div className="flex flex-wrap gap-2">
        {TABS.map((tab) => {
          const Icon = tab.icon
          const active = activeTab === tab.key
          return (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={cn(
                'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
                active
                  ? 'bg-primary text-primary-foreground shadow-md'
                  : 'bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground border'
              )}
            >
              <Icon className="h-4 w-4" />
              {t(tab.labelKey)}
            </button>
          )
        })}
      </div>

      {/* Rankings List */}
      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i}>
              <CardContent className="p-5">
                <div className="flex items-center gap-4">
                  <Skeleton className="h-10 w-10 rounded-xl" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-5 w-40" />
                    <Skeleton className="h-4 w-24" />
                  </div>
                  <Skeleton className="h-6 w-20" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : sorted.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            <Trophy className="h-12 w-12 mx-auto mb-3 opacity-30" />
            {t('No data available')}
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {sorted.map((item, index) => {
            const rank = index + 1
            const isTop3 = rank <= 3

            return (
              <Card
                key={`${item.provider}-${item.model}`}
                className={cn(
                  'transition-all hover:shadow-md',
                  isTop3 && 'ring-1 ring-yellow-500/20'
                )}
              >
                <CardContent className="p-4">
                  <div className="flex items-center gap-4">
                    {/* Rank Badge */}
                    <div
                      className={cn(
                        'flex items-center justify-center w-10 h-10 rounded-xl font-bold text-white shrink-0',
                        isTop3
                          ? RANK_STYLES[rank - 1]
                          : 'bg-muted text-muted-foreground'
                      )}
                    >
                      {isTop3 ? <Trophy className="h-5 w-5" /> : rank}
                    </div>

                    {/* Model Info */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-semibold text-base">{item.model}</span>
                        <Badge variant="outline" className="text-xs">
                          {item.provider}
                        </Badge>
                      </div>
                      <div className="text-xs text-muted-foreground mt-0.5">
                        {item.channel_name}
                      </div>
                    </div>

                    {/* Stats */}
                    <div className="flex items-center gap-4 sm:gap-6 text-right shrink-0">
                      {/* Request Count / Token / Speed / Price */}
                      {activeTab === 'request_count_7d' && (
                        <div>
                          <div className="text-sm font-semibold">
                            {formatRequests(item.request_count_7d)}
                          </div>
                          <div className="text-[10px] text-muted-foreground uppercase tracking-wider">
                            {t('Requests')}
                          </div>
                        </div>
                      )}
                      {activeTab === 'tokens_7d' && (
                        <div>
                          <div className="text-sm font-semibold">
                            {formatTokens(item.tokens_7d)}
                          </div>
                          <div className="text-[10px] text-muted-foreground uppercase tracking-wider">
                            {t('Tokens')}
                          </div>
                        </div>
                      )}
                      {activeTab === 'avg_speed_ms' && (
                        <div>
                          <div className="text-sm font-semibold">
                            {item.avg_speed_ms}ms
                          </div>
                          <div className="text-[10px] text-muted-foreground uppercase tracking-wider">
                            {t('Avg Speed')}
                          </div>
                        </div>
                      )}
                      {activeTab === 'price_per_1k' && (
                        <div>
                          <div className="text-sm font-semibold">
                            ${item.price_per_1k.toFixed(4)}
                          </div>
                          <div className="text-[10px] text-muted-foreground uppercase tracking-wider">
                            /1K {t('Tokens')}
                          </div>
                        </div>
                      )}

                      {/* Trend */}
                      <span
                        className={cn(
                          'inline-flex items-center gap-0.5 text-xs font-medium px-2 py-1 rounded-full',
                          item.trend_percent >= 0
                            ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                            : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                        )}
                      >
                        {item.trend_percent >= 0 ? (
                          <TrendingUp className="h-3 w-3" />
                        ) : (
                          <TrendingDown className="h-3 w-3" />
                        )}
                        {Math.abs(item.trend_percent)}%
                      </span>
                    </div>
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
