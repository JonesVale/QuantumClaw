import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useMemo, useEffect, useRef } from 'react'
import { PromoCarousel } from '@/components/promo-carousel'
import { ModelFilterSidebar, useSidebarData, type SidebarFilters } from '@/components/model-filter-sidebar'

export const Route = createFileRoute('/pricing')({
  component: PricingPage,
})

interface ModelPricing { name: string; provider: string; input_price: number; output_price: number; status: number }

function PricingPage() {
  const { t, language } = useT()
  const [search, setSearch] = useState('')
  const [prov, setProv] = useState('')
  const [cat, setCat] = useState('all')
  const [ctx, setCtx] = useState('')
  const [activeOnly, setActiveOnly] = useState(true)
  const [hovered, setHovered] = useState(true)
  const [vis, setVis] = useState(20)
  const STEP = 20
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

  const allPrices: ModelPricing[] = all

  const filtered = useMemo(()=>{
    let r=allPrices
    if(search){const q=search.toLowerCase();r=r.filter(m=>m.name.toLowerCase().includes(q)||m.provider?.toLowerCase().includes(q))}
    if(cat!=='all')r=r.filter(m=>m.use_case===cat)
    if(prov)r=r.filter(m=>m.provider===prov)
    if(ctx && ctx.includes('-')){const [cL,cH]=ctx.split('-').map(Number);r=r.filter(m=>(m.context_window||0)>=cL&&(m.context_window||0)<=cH)}
    if(activeOnly)r=r.filter(m=>m.status===1)
    return r
  },[allPrices,search,cat,prov,ctx,activeOnly])

  const shown = useMemo(()=>filtered.slice(0,vis),[filtered,vis])

  return (
    <div className="min-h-screen bg-background overflow-x-hidden" style={{backgroundImage:'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)'}}>
      <div className="qc-wrap qc-section-pad-sm">
        <div className="mb-6">
          <PromoCarousel pageKey="pricing" />
        </div>
        <div className="relative">
          <ModelFilterSidebar filters={filters} hovered={hovered} onEnter={handleSidebarEnter} onLeave={handleSidebarLeave}
            useCases={useCases} aiProviders={aiProviders} quantumProviders={quantumProviders} contextBuckets={contextBuckets} />

          <div className="transition-all duration-200" className='max-w-[100vw] overflow-x-hidden' style={{ paddingLeft: hovered ? '304px' : '72px' }}>
            <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 flex-wrap mb-6">
              <div className="relative flex-1 min-w-[160px] max-w-xs">
                <svg className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/30" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
                <input value={search} onChange={e=>setSearch(e.target.value)} className="w-full h-10 rounded-xl border border-border/30 bg-white/70 px-10 text-sm placeholder:text-muted-foreground/40 outline-none focus:border-[oklch(0.72_0.18_52)]/40 focus:bg-white transition-all" placeholder={`${t('Search')} ${allPrices.length} ${t('models')}...`} />
              </div>
              <label className="flex items-center gap-2 px-3 py-2 rounded-xl border border-border/30 bg-white/70 text-sm cursor-pointer hover:bg-muted/40 transition-all">
                <input type="checkbox" checked={activeOnly} onChange={e=>setActiveOnly(e.target.checked)} className="w-4 h-4 rounded border-2 border-muted-foreground/30 accent-[oklch(0.72_0.18_52)]" />
                {t('Active only')}
              </label>
            </div>

            {filtered.length===0 ? (
              <div className="text-center py-20 text-muted-foreground">
                <p className="text-lg font-medium mb-2">{t('No models found')}</p>
                <button onClick={()=>{setSearch('');setProv('');setCat('all');setCtx('')}} className="mt-4 px-5 py-2.5 rounded-xl border border-border/30 bg-white hover:bg-muted/40 text-sm font-medium transition-all">{t('Reset filters')}</button>
              </div>
            ) : (
              <>
                <p className="text-xs text-muted-foreground/40 font-medium tracking-wide mb-4">{filtered.length} {t('models')}</p>
                <div className="space-y-2">
                  <div className="hidden md:flex items-center gap-4 px-3 sm:px-5 py-2 sm:py-3 text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] bg-muted/20 rounded-xl">
                    <span className="flex-[2]">{t('Model')}</span>
                    <span className="flex-1">{t('Provider')}</span>
                    <span className="flex-1 text-right">{t('Input / 1K tokens')}</span>
                    <span className="flex-1 text-right">{t('Output / 1K tokens')}</span>
                    <span className="w-12 text-center">{t('Status')}</span>
                  </div>
                  {shown.map((m,i)=>(
                    <div key={m.name+i} className="qc-fade-up flex flex-col md:flex-row items-start md:items-center gap-2 md:gap-4 px-3 sm:px-5 py-3 sm:py-4 rounded-2xl bg-white/60 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm" style={{animationDelay:`${(i%10)*0.04}s`}}>
                      <div className="flex-[2] min-w-0"><span className="text-sm font-semibold text-foreground">{m.name}</span></div>
                      <span className="flex-1 text-sm text-muted-foreground/70">{m.provider}</span>
                      <span className="flex-1 text-sm text-right font-medium tabular-nums text-foreground/80">{m.input_price>0?`$${m.input_price.toFixed(6)}`:<span className="text-emerald-600 font-semibold">{t('Free')}</span>}</span>
                      <span className="flex-1 text-sm text-right font-medium tabular-nums text-foreground/80">{m.output_price>0?`$${m.output_price.toFixed(6)}`:<span className="text-emerald-600 font-semibold">{t('Free')}</span>}</span>
                      <div className="w-12 flex justify-center"><span className={`inline-block w-2 h-2 rounded-full ${m.status===1?'bg-emerald-500':'bg-muted-foreground/30'}`} title={m.status===1?t('Active'):t('Inactive')}/></div>
                    </div>
                  ))}
                </div>
                {vis<filtered.length&&<div className="flex justify-center mt-8"><button onClick={()=>setVis(v=>v+STEP)} className="px-8 py-3 rounded-xl border border-border/30 bg-white/70 hover:bg-white hover:shadow-sm text-sm font-medium transition-all hover:-translate-y-0.5">{t('Show more')} <span className="text-muted-foreground/60">({filtered.length-vis})</span></button></div>}
              </>
            )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
