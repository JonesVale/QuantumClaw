import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useMemo, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { ModelRanking } from '@/lib/api-extended'
import { PromoCarousel } from '@/components/promo-carousel'
import { ModelFilterSidebar, useSidebarData, type SidebarFilters } from '@/components/model-filter-sidebar'

export const Route = createFileRoute('/rankings')({
  component: RankingsPage,
})

type SortKey = 'request_count_7d' | 'tokens_7d' | 'avg_speed_ms' | 'price_per_1k'

const TABS: { key: SortKey; label: string; icon: string }[] = [
  { key: 'request_count_7d', label: 'Requests', icon: '≡' },
  { key: 'tokens_7d', label: 'Tokens', icon: '◈' },
  { key: 'avg_speed_ms', label: 'Speed', icon: '⚡' },
  { key: 'price_per_1k', label: 'Price', icon: '$' },
]

function formatNum(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return n.toLocaleString()
}

function RankingsPage() {
  const { t, language } = useT()
  const [sortKey, setSortKey] = useState<SortKey>('request_count_7d')
  const [cat, setCat] = useState('all')
  const [prov, setProv] = useState('')
  const [ctx, setCtx] = useState('')
  const [hovered, setHovered] = useState(true)
  const expandTimer = useRef<ReturnType<typeof setTimeout>>()
  const collapseTimer = useRef<ReturnType<typeof setTimeout>>()

  const handleSidebarEnter = () => {
    clearTimeout(collapseTimer.current)
    expandTimer.current = setTimeout(() => setHovered(true), 200)
  }
  const handleSidebarLeave = () => {
    clearTimeout(expandTimer.current)
    collapseTimer.current = setTimeout(() => setHovered(false), 150)
  }

  useEffect(() => {
    return () => { clearTimeout(expandTimer.current); clearTimeout(collapseTimer.current) }
  }, [])

  const filters: SidebarFilters = { cat, setCat, prov, setProv, ctx, setCtx }
  const { all, aiProviders, quantumProviders, useCases, contextBuckets } = useSidebarData(language)

  const { data, isLoading } = useQuery({
    queryKey: ['model-rankings'],
    queryFn: async () => { const r = await fetch('/api/models/rankings'); if (!r.ok) throw Error(); return r.json() },
    staleTime: 60_000,
  }) as any
  const rankings: ModelRanking[] = data?.data || []

  const sorted = useMemo(() => {
    let r = [...rankings]
    // Use sidebar filters on rankings data (match by model name)
    if (cat !== 'all') r = r.filter(m => (m as any).use_case === cat)
    if (prov) r = r.filter(m => m.provider === prov)
    if (sortKey === 'avg_speed_ms') r.sort((a, b) => (a.avg_speed_ms ?? 999) - (b.avg_speed_ms ?? 999))
    else r.sort((a, b) => (b[sortKey] ?? 0) - (a[sortKey] ?? 0))
    return r.slice(0, 50)
  }, [rankings, sortKey, cat, prov])

  const statValue = (m: ModelRanking, k: SortKey): string => {
    switch (k) {
      case 'request_count_7d': return formatNum(m.request_count_7d || 0)
      case 'tokens_7d': return formatNum(m.tokens_7d || 0)
      case 'avg_speed_ms': return (m.avg_speed_ms || 0).toFixed(0) + 'ms'
      case 'price_per_1k': return '$' + (m.price_per_1k || 0).toFixed(6)
    }
  }

  return (
    <div className="min-h-screen bg-background" style={{backgroundImage:'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)'}}>
      <div className="qc-wrap qc-section-pad-sm">
        <div className="mb-6">
          <PromoCarousel pageKey="rankings" />
        </div>
        <div className="relative">
          <ModelFilterSidebar filters={filters} hovered={hovered} onEnter={handleSidebarEnter} onLeave={handleSidebarLeave}
            useCases={useCases} aiProviders={aiProviders} quantumProviders={quantumProviders} contextBuckets={contextBuckets} />

          <div className="transition-all duration-200" style={{ paddingLeft: hovered ? '304px' : '72px' }}>
            <div className="flex-1 min-w-0">
            {/* Sort tabs */}
            <div className="flex items-center gap-2 flex-wrap mb-6">
              {TABS.map(tab=>(
                <button key={tab.key} onClick={()=>setSortKey(tab.key)}
                  className={`px-4 py-2 rounded-xl text-sm font-medium transition-all ${sortKey===tab.key?'bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-md':'bg-white/70 border border-border/30 text-muted-foreground hover:text-foreground hover:bg-white'}`}>
                  <span className="mr-1.5">{tab.icon}</span>
                  {t(tab.label)}
                </button>
              ))}
            </div>

            {isLoading ? (
              <div className="flex justify-center py-20"><div className="w-6 h-6 rounded-full border-2 border-amber-500/30 border-t-amber-500 animate-spin"/></div>
            ) : sorted.length===0 ? (
              <div className="text-center py-20 text-muted-foreground"><p className="text-lg font-medium">{t('No ranking data yet')}</p></div>
            ) : (
              <>
                <p className="text-xs text-muted-foreground/40 font-medium tracking-wide mb-4">{t('Top 50 models')}</p>
                <div className="space-y-2">
                  <div className="hidden md:flex items-center gap-4 px-5 py-3 text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] bg-muted/20 rounded-xl">
                    <span className="w-10 text-center">#</span>
                    <span className="flex-[2]">{t('Model')}</span>
                    <span className="flex-1">{t('Provider')}</span>
                    <span className="flex-1 text-right">{TABS.find(t=>t.key===sortKey)?.label||t('Value')}</span>
                  </div>
                  {sorted.map((m,i)=>{
                    const rank=i+1
                    return (
                      <div key={m.model_name||i} className="qc-fade-up flex flex-col md:flex-row items-start md:items-center gap-2 md:gap-4 px-5 py-4 rounded-2xl bg-white/60 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm" style={{animationDelay:`${(i%10)*0.04}s`}}>
                        <div className="w-10 flex items-center justify-center shrink-0">
                          {rank<=3?<span className={`text-lg font-bold ${rank===1?'text-amber-500':rank===2?'text-slate-400':'text-amber-700'}`}>#{rank}</span>:<span className="text-xs font-semibold text-muted-foreground/40">#{rank}</span>}
                        </div>
                        <div className="flex-[2] min-w-0"><span className="text-sm font-semibold text-foreground">{m.model_name||m.name}</span></div>
                        <span className="flex-1 text-sm text-muted-foreground/70">{m.provider||'\u2014'}</span>
                        <span className="flex-1 text-sm text-right font-medium tabular-nums">{statValue(m,sortKey)}</span>
                      </div>
                    )
                  })}
                </div>
              </>
            )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
