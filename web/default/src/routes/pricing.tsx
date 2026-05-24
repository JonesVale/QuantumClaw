import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'

export const Route = createFileRoute('/pricing')({
  component: PricingPage,
})

interface ModelPricing {
  name: string
  provider: string
  input_price: number
  output_price: number
  status: number
}

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
      const r = await fetch('/api/model-catalog?lang=en')
      if (!r.ok) throw Error()
      return r.json()
    },
    staleTime: 60_000,
  })
  const all: ModelPricing[] = data?.data || []

  const providers = useMemo(() => {
    const s = new Set(all.map(m => m.provider).filter(Boolean))
    return Array.from(s).sort()
  }, [all])

  const filtered = useMemo(() => {
    let r = all
    if (search) { const q = search.toLowerCase(); r = r.filter(m => m.name.toLowerCase().includes(q) || m.provider?.toLowerCase().includes(q)) }
    if (providerFilter !== 'all') r = r.filter(m => m.provider === providerFilter)
    if (showActiveOnly) r = r.filter(m => m.status === 1)
    return r
  }, [all, search, providerFilter, showActiveOnly])

  const shown = useMemo(() => filtered.slice(0, visibleCount), [filtered, visibleCount])

  return (
    <div className="min-h-screen bg-background"
      style={{ backgroundImage: 'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)' }}>
      <div className="qc-wrap qc-section-pad-sm">
        {/* Header */}
        <div className="qc-fade-up mb-10">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-amber-200 bg-amber-50 text-amber-700 text-xs font-semibold tracking-wide mb-5">
            <span className="w-2 h-2 rounded-full bg-amber-500" />
            {t('Transparent Pricing')}
          </div>
          <h1 className="qc-title-hero font-bold tracking-tight text-foreground">
            {t('Model Pricing')}
          </h1>
          <p className="qc-text-body qc-readable-width text-muted-foreground/70 mt-2 leading-relaxed">
            {t('Compare token pricing across all providers. Pay only for what you use.')}
          </p>
        </div>

        {/* Controls */}
        <div className="qc-fade-up flex items-center gap-3 flex-wrap mb-8">
          <div className="relative flex-1 min-w-[180px] max-w-xs">
            <svg className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/30" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
            <input value={search} onChange={e => setSearch(e.target.value)}
              className="w-full h-10 rounded-xl border border-border/30 bg-white/70 px-10 text-sm placeholder:text-muted-foreground/40 outline-none focus:border-[oklch(0.72_0.18_52)]/40 focus:bg-white transition-all"
              placeholder={`${t('Search')} ${all.length} ${t('models')}...`} />
          </div>
          <select value={providerFilter} onChange={e => setProviderFilter(e.target.value)}
            className="h-10 rounded-xl border border-border/30 bg-white/70 px-3 text-sm outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all">
            <option value="all">{t('All Providers')}</option>
            {providers.map(p => <option key={p} value={p}>{p}</option>)}
          </select>
          <label className="flex items-center gap-2 px-3 py-2 rounded-xl border border-border/30 bg-white/70 text-sm cursor-pointer hover:bg-muted/30 transition-all select-none">
            <input type="checkbox" checked={showActiveOnly} onChange={e => setShowActiveOnly(e.target.checked)}
              className="w-4 h-4 rounded border-2 border-muted-foreground/30 accent-[oklch(0.72_0.18_52)]" />
            {t('Active only')}
          </label>
          <button onClick={() => refetch()}
            className="h-10 px-3 rounded-xl border border-border/30 bg-white/70 hover:bg-muted/40 transition-all text-muted-foreground"
            title={t('Refresh')}>
            <svg className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M23 4v6h-6M1 20v-6h6"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>
          </button>
        </div>

        {/* Loading / Empty / Table */}
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <div className="w-6 h-6 rounded-full border-2 border-amber-500/30 border-t-amber-500 animate-spin" />
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-20 text-muted-foreground">
            <p className="text-lg font-medium mb-2">{t('No models found')}</p>
            <button onClick={() => { setSearch(''); setProviderFilter('all') }}
              className="mt-4 px-5 py-2.5 rounded-xl border border-border/30 bg-white hover:bg-muted/40 text-sm font-medium transition-all">
              {t('Reset filters')}
            </button>
          </div>
        ) : (
          <>
            <p className="text-xs text-muted-foreground/40 font-medium tracking-wide mb-4">{filtered.length} {t('models')}</p>
            <div className="space-y-2">
              {/* Table header */}
              <div className="hidden md:flex items-center gap-4 px-5 py-3 text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] bg-muted/20 rounded-xl">
                <span className="flex-[2]">{t('Model')}</span>
                <span className="flex-1">{t('Provider')}</span>
                <span className="flex-1 text-right">{t('Input / 1K tokens')}</span>
                <span className="flex-1 text-right">{t('Output / 1K tokens')}</span>
                <span className="w-16 text-center">{t('Status')}</span>
              </div>
              {shown.map((m, i) => (
                <div key={m.name + i}
                  className="qc-fade-up flex flex-col md:flex-row items-start md:items-center gap-2 md:gap-4 px-5 py-4 rounded-2xl bg-white/60 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm"
                  style={{ animationDelay: `${(i % 10) * 0.05}s` }}>
                  <div className="flex-[2] min-w-0">
                    <span className="text-sm font-semibold text-foreground">{m.name}</span>
                  </div>
                  <span className="flex-1 text-sm text-muted-foreground/70">{m.provider}</span>
                  <span className="flex-1 text-sm text-right font-medium tabular-nums text-foreground/80">
                    {m.input_price > 0 ? `$${m.input_price.toFixed(6)}` : <span className="text-emerald-600 font-semibold">{t('Free')}</span>}
                  </span>
                  <span className="flex-1 text-sm text-right font-medium tabular-nums text-foreground/80">
                    {m.output_price > 0 ? `$${m.output_price.toFixed(6)}` : <span className="text-emerald-600 font-semibold">{t('Free')}</span>}
                  </span>
                  <div className="w-16 flex justify-center">
                    <span className={`inline-block w-2 h-2 rounded-full ${m.status === 1 ? 'bg-emerald-500' : 'bg-muted-foreground/30'}`}
                      title={m.status === 1 ? t('Active') : t('Inactive')} />
                  </div>
                </div>
              ))}
            </div>
            {visibleCount < filtered.length && (
              <div className="flex justify-center mt-8">
                <button onClick={() => setVisibleCount(v => v + PAGE_STEP)}
                  className="px-8 py-3 rounded-xl border border-border/30 bg-white/70 hover:bg-white hover:shadow-sm text-sm font-medium transition-all hover:-translate-y-0.5">
                  {t('Show more')} <span className="text-muted-foreground/60">({filtered.length - visibleCount})</span>
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
