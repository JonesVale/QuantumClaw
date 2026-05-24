import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import type { ModelRanking } from '@/lib/api-extended'

export const Route = createFileRoute('/rankings')({
  component: RankingsPage,
})

type SortKey = 'request_count_7d' | 'tokens_7d' | 'avg_speed_ms' | 'price_per_1k'

function formatNum(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return n.toLocaleString()
}

const TABS: { key: SortKey; label: string; icon: string }[] = [
  { key: 'request_count_7d', label: 'By Requests', icon: '≡' },
  { key: 'tokens_7d', label: 'By Tokens', icon: '◈' },
  { key: 'avg_speed_ms', label: 'By Speed', icon: '⚡' },
  { key: 'price_per_1k', label: 'By Price', icon: '$' },
]

function RankingsPage() {
  const { t } = useT()
  const [sortKey, setSortKey] = useState<SortKey>('request_count_7d')

  const { data, isLoading } = useQuery({
    queryKey: ['model-rankings'],
    queryFn: async () => {
      const r = await fetch('/api/model-rankings')
      if (!r.ok) throw Error()
      return r.json()
    },
    staleTime: 60_000,
  })
  const rankings: ModelRanking[] = data?.data || []

  const sorted = useMemo(() => {
    const r = [...rankings]
    if (sortKey === 'avg_speed_ms') r.sort((a, b) => (a.avg_speed_ms ?? 999) - (b.avg_speed_ms ?? 999))
    else r.sort((a, b) => (b[sortKey] ?? 0) - (a[sortKey] ?? 0))
    return r.slice(0, 50)
  }, [rankings, sortKey])

  const statValue = (m: ModelRanking, k: SortKey): string => {
    switch (k) {
      case 'request_count_7d': return formatNum(m.request_count_7d || 0)
      case 'tokens_7d': return formatNum(m.tokens_7d || 0)
      case 'avg_speed_ms': return (m.avg_speed_ms || 0).toFixed(0) + 'ms'
      case 'price_per_1k': return '$' + (m.price_per_1k || 0).toFixed(6)
    }
  }

  return (
    <div className="min-h-screen bg-background"
      style={{ backgroundImage: 'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)' }}>
      <div className="qc-wrap qc-section-pad-sm">
        {/* Header */}
        <div className="qc-fade-up mb-10">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-amber-200 bg-amber-50 text-amber-700 text-xs font-semibold tracking-wide mb-5">
            <span className="w-2 h-2 rounded-full bg-amber-500" />
            {t('Real-time Rankings')}
          </div>
          <h1 className="qc-title-hero font-bold tracking-tight text-foreground">
            {t('Model Rankings')}
          </h1>
          <p className="qc-text-body qc-readable-width text-muted-foreground/70 mt-2 leading-relaxed">
            {t('See which models are trending. Sorted by real usage across the platform.')}
          </p>
        </div>

        {/* Tabs */}
        <div className="qc-fade-up flex items-center gap-2 mb-8 bg-muted/30 rounded-2xl p-1.5 border border-border/10 w-fit flex-wrap">
          {TABS.map(tab => (
            <button
              key={tab.key}
              onClick={() => setSortKey(tab.key)}
              className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${
                sortKey === tab.key
                  ? 'bg-white shadow-sm text-foreground'
                  : 'text-muted-foreground/60 hover:text-foreground hover:bg-white/50'
              }`}
            >
              <span>{tab.icon}</span>
              {t(tab.label)}
            </button>
          ))}
        </div>

        {/* List */}
        {isLoading ? (
          <div className="flex justify-center py-20">
            <div className="w-6 h-6 rounded-full border-2 border-amber-500/30 border-t-amber-500 animate-spin" />
          </div>
        ) : sorted.length === 0 ? (
          <div className="text-center py-20 text-muted-foreground">
            <p className="text-lg font-medium">{t('No ranking data yet')}</p>
          </div>
        ) : (
          <>
            <p className="text-xs text-muted-foreground/40 font-medium tracking-wide mb-4">{t('Top 50 models')}</p>
            <div className="space-y-2">
              {/* Header */}
              <div className="hidden md:flex items-center gap-4 px-5 py-3 text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] bg-muted/20 rounded-xl">
                <span className="w-10 text-center">#</span>
                <span className="flex-[2]">{t('Model')}</span>
                <span className="flex-1">{t('Provider')}</span>
                <span className="flex-1 text-right">
                  {TABS.find(t => t.key === sortKey)?.label.replace('By ', '') || t('Value')}
                </span>
              </div>
              {sorted.map((m, i) => {
                const rank = i + 1
                return (
                  <div key={m.model_name || i}
                    className="qc-fade-up flex flex-col md:flex-row items-start md:items-center gap-2 md:gap-4 px-5 py-4 rounded-2xl bg-white/60 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm"
                    style={{ animationDelay: `${(i % 10) * 0.04}s` }}>
                    <div className="w-10 flex items-center justify-center shrink-0">
                      {rank <= 3
                        ? <span className={`text-lg font-bold ${rank === 1 ? 'text-amber-500' : rank === 2 ? 'text-slate-400' : 'text-amber-700'}`}>#{rank}</span>
                        : <span className="text-xs font-semibold text-muted-foreground/40">#{rank}</span>
                      }
                    </div>
                    <div className="flex-[2] min-w-0">
                      <span className="text-sm font-semibold text-foreground">{m.model_name || m.name}</span>
                    </div>
                    <span className="flex-1 text-sm text-muted-foreground/70">{m.provider || '—'}</span>
                    <span className="flex-1 text-sm text-right font-medium tabular-nums">
                      {statValue(m, sortKey)}
                    </span>
                  </div>
                )
              })}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
