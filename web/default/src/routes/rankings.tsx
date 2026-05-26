import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { ModelRanking } from '@/lib/api-extended'
import { PromoCarousel } from '@/components/promo-carousel'
import { ModelFilterSidebar, useSidebarData, type SidebarFilters } from '@/components/model-filter-sidebar'

export const Route = createFileRoute('/rankings')({
  component: RankingsPage,
})

// Brand ranking from GET /api/brand-rankings
interface BrandEntry {
  brand_name: string
  rank: number
  score: number
  metric: string
  source: string
}

function formatNum(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return n.toLocaleString()
}

function RankingsPage() {
  const { t, language } = useT()
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

  // Brand rankings
  const { data: brandData, isLoading: brandLoading } = useQuery({
    queryKey: ['brand-rankings'],
    queryFn: async () => { const r = await fetch('/api/brand-rankings'); if (!r.ok) throw Error(); return r.json() },
    staleTime: 300_000,
  }) as any
  const brands: BrandEntry[] = brandData?.data || []

  // Model rankings (platform usage data)
  const { data: modelData, isLoading: modelLoading } = useQuery({
    queryKey: ['model-rankings'],
    queryFn: async () => { const r = await fetch('/api/models/rankings'); if (!r.ok) throw Error(); return r.json() },
    staleTime: 60_000,
  }) as any
  const rankings: ModelRanking[] = modelData?.data || []

  const rankLabel = (rank: number) => {
    if (rank === 1) return { text: '#1', cls: 'text-amber-500' }
    if (rank === 2) return { text: '#2', cls: 'text-slate-400' }
    if (rank === 3) return { text: '#3', cls: 'text-amber-700' }
    return { text: `#${rank}`, cls: 'text-muted-foreground/40' }
  }

  return (
    <div className="min-h-screen bg-background overflow-x-hidden" style={{backgroundImage:'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)'}}>
      <div className="qc-wrap qc-section-pad-sm">
        <div className="mb-6">
          <PromoCarousel pageKey="rankings" />
        </div>
        <div className="relative">
          <ModelFilterSidebar filters={filters} hovered={hovered} onEnter={handleSidebarEnter} onLeave={handleSidebarLeave}
            useCases={useCases} aiProviders={aiProviders} quantumProviders={quantumProviders} contextBuckets={contextBuckets} />

          <div className="transition-all duration-200 max-w-[100vw] overflow-x-hidden" style={{ paddingLeft: hovered ? '304px' : '72px' }}>
            <div className="flex-1 min-w-0">

              {/* ── Brand Power Rankings ── */}
              <section className="mb-12">
                <h2 className="text-lg font-bold text-foreground mb-1">🏢 {t('Brand Power Rankings')}</h2>
                <p className="text-xs text-muted-foreground/40 mb-5">AI / Quantum industry-wide brand rankings • Updated monthly</p>

                {brandLoading ? (
                  <div className="flex justify-center py-12"><div className="w-6 h-6 rounded-full border-2 border-violet-500/30 border-t-violet-500 animate-spin"/></div>
                ) : brands.length===0 ? (
                  <div className="text-center py-12 text-muted-foreground"><p className="text-sm">{t('No ranking data yet')}</p></div>
                ) : (
                  <div className="space-y-2">
                    <div className="hidden md:flex items-center gap-4 px-5 py-3 text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] bg-muted/20 rounded-xl">
                      <span className="w-10 text-center">#</span>
                      <span className="flex-[2]">{t('Brand')}</span>
                      <span className="flex-1">{t('Score')}</span>
                      <span className="flex-1">{t('Source')}</span>
                    </div>
                    {brands.map((b,i)=>{
                      const rl = rankLabel(b.rank)
                      return (
                        <div key={b.brand_name} className="qc-fade-up flex flex-col md:flex-row items-start md:items-center gap-2 md:gap-4 px-5 py-4 rounded-2xl bg-white/60 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm" style={{animationDelay:`${(i%10)*0.04}s`}}>
                          <div className="w-10 flex items-center justify-center shrink-0">
                            <span className={`text-lg font-bold ${rl.cls}`}>{rl.text}</span>
                          </div>
                          <div className="flex-[2] min-w-0">
                            <span className="text-sm font-semibold text-foreground">{b.brand_name}</span>
                          </div>
                          <div className="flex-1">
                            <div className="flex items-center gap-2">
                              <div className="h-2 rounded-full bg-gradient-to-r from-violet-400 to-indigo-500" style={{width:`${b.score}%`, maxWidth:'120px'}}/>
                              <span className="text-sm font-medium tabular-nums text-foreground/80">{b.score}</span>
                            </div>
                          </div>
                          <span className="flex-1 text-sm text-muted-foreground/40">{b.source||'\u2014'}</span>
                        </div>
                      )
                    })}
                  </div>
                )}
              </section>

              {/* ── Model Platform Rankings ── */}
              <section>
                <h2 className="text-lg font-bold text-foreground mb-1">📊 {t('Model Platform Rankings')}</h2>
                <p className="text-xs text-muted-foreground/40 mb-5">{t('Top 50 models')} — {t('Usage stats over the last 7 days')}</p>

                {modelLoading ? (
                  <div className="flex justify-center py-12"><div className="w-6 h-6 rounded-full border-2 border-amber-500/30 border-t-amber-500 animate-spin"/></div>
                ) : rankings.length===0 ? (
                  <div className="text-center py-12 text-muted-foreground"><p className="text-sm">{t('No ranking data yet')}</p></div>
                ) : (
                  <div className="space-y-2">
                    {/* Table header — static, not clickable */}
                    <div className="hidden md:flex items-center gap-4 px-5 py-3 text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] bg-muted/20 rounded-xl">
                      <span className="w-10 text-center">#</span>
                      <span className="flex-[2]">{t('Model')}</span>
                      <span className="flex-1">{t('Provider')}</span>
                      <span className="flex-1 text-right">≡ {t('Requests')}</span>
                      <span className="flex-1 text-right">◈ {t('Tokens')}</span>
                      <span className="flex-1 text-right">⚡ {t('Speed')}</span>
                      <span className="flex-1 text-right">$ {t('Price')}</span>
                    </div>
                    {rankings.slice(0, 50).map((m,i)=>{
                      const rl = rankLabel(i+1)
                      return (
                        <div key={m.model_name||i} className="qc-fade-up flex flex-col md:flex-row items-start md:items-center gap-2 md:gap-4 px-5 py-4 rounded-2xl bg-white/60 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm" style={{animationDelay:`${(i%10)*0.04}s`}}>
                          <div className="w-10 flex items-center justify-center shrink-0">
                            <span className={`text-lg font-bold ${rl.cls}`}>{rl.text}</span>
                          </div>
                          <div className="flex-[2] min-w-0"><span className="text-sm font-semibold text-foreground">{m.model_name||m.name}</span></div>
                          <span className="flex-1 text-sm text-muted-foreground/70">{m.provider||'\u2014'}</span>
                          <span className="flex-1 text-sm text-right tabular-nums">{formatNum(m.request_count_7d||0)}</span>
                          <span className="flex-1 text-sm text-right tabular-nums">{formatNum(m.tokens_7d||0)}</span>
                          <span className="flex-1 text-sm text-right tabular-nums">{(m.avg_speed_ms||0).toFixed(0)}ms</span>
                          <span className="flex-1 text-sm text-right tabular-nums">${(m.price_per_1k||0).toFixed(6)}</span>
                        </div>
                      )
                    })}
                  </div>
                )}
              </section>

            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
