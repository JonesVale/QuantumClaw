import { createFileRoute, Link } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useMemo, useEffect, useRef } from 'react'
import { useAuthStore } from '@/stores/auth-store'
import { PromoCarousel } from '@/components/promo-carousel'
import { ModelDetailDialog, type CatalogItem } from '@/components/model-detail-dialog'
import { ModelComparisonDialog } from '@/components/model-comparison-dialog'
import { ModelFilterSidebar, useSidebarData, type SidebarFilters } from '@/components/model-filter-sidebar'

export const Route = createFileRoute('/models')({
  component: ModelsPage,
})

const useCaseMeta: Record<string,{label:string;icon:string;gradient:string}> = {
  chat:     { label:'Chat', icon:'💬', gradient:'from-amber-500 to-orange-500' },
  coding:   { label:'Code', icon:'</>', gradient:'from-emerald-500 to-teal-500' },
  reasoning:{ label:'Reasoning', icon:'🧠', gradient:'from-amber-600 to-rose-600' },
  vision:   { label:'Vision', icon:'👁', gradient:'from-orange-400 to-rose-400' },
}

type SortOpt = 'name'|'price-asc'|'price-desc'

function ModelsPage() {
  const { t, language } = useT()
  const { auth } = useAuthStore()
  const [q, setQ] = useState('')
  const [cat, setCat] = useState('all')
  // Read initial provider from URL search params (e.g. /models?provider=OpenAI)
  const [prov, setProv] = useState(() => new URLSearchParams(window.location.search).get('provider') || '')
  const [ctx, setCtx] = useState('')
  const [sort, setSort] = useState<SortOpt>('name')
  const [hovered, setHovered] = useState(true)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [dm, setDm] = useState<CatalogItem|null>(null)
  const [do_, setDo_] = useState(false)
  const [vis, setVis] = useState(30)
  const STEP = 30
  const [sel, setSel] = useState<Set<string>>(new Set)
  const [cop, setCop] = useState(false)
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
    return () => {
      clearTimeout(expandTimer.current)
      clearTimeout(collapseTimer.current)
    }
  }, [])

  // Shared sidebar data
  const { all, aiProviders, quantumProviders, useCases, contextBuckets } = useSidebarData(language)

  const filters: SidebarFilters = { cat, setCat, prov, setProv, ctx, setCtx }

  const providers = useMemo(()=>{
    const m=new Map<string,number>()
    all.forEach(mt=>{const p=mt.provider;if(p&&p!=='Unknown'&&!p.startsWith('~'))m.set(p,(m.get(p)||0)+1)})
    return [...m].sort((a,b)=>b[1]-a[1])
  },[all])

  const filtered = useMemo(()=>{
    let r = all
    if(q){const l=q.toLowerCase();r=r.filter(m=>m.name.toLowerCase().includes(l)||m.description.toLowerCase().includes(l))}
    if(cat!=='all')r=r.filter(m=>m.use_case===cat)
    if(prov)r=r.filter(m=>m.provider===prov)
    if(ctx && ctx.includes('-')){const [cLow,cHigh]=ctx.split('-').map(Number);r=r.filter(m=>(m.context_window||0)>=cLow&&(m.context_window||0)<=cHigh)}
    switch(sort){
      case'price-asc':r.sort((a,b)=>(a.input_price??999)-(b.input_price??999));break
      case'price-desc':r.sort((a,b)=>(b.input_price??0)-(a.input_price??0));break
      default:r.sort((a,b)=>a.name.localeCompare(b.name))
    }
    return r
  },[all,q,cat,prov,ctx,sort])

  useEffect(()=>setVis(STEP),[q,cat,prov,ctx,sort])
  const shown = useMemo(()=>filtered.slice(0,vis),[filtered,vis])

  const toggleSel=(n:string)=>setSel(p=>{const x=new Set(p);x.has(n)?x.delete(n):x.size<4&&x.add(n);return x})
  const selList = useMemo(()=>all.filter(m=>sel.has(m.name)),[all,sel])

  // Group filtered models by provider
  const groupedByProvider = useMemo(() => {
    const groups: Record<string, CatalogItem[]> = {}
    filtered.forEach(m => {
      const p = m.provider || 'Unknown'
      if (!groups[p]) groups[p] = []
      groups[p].push(m)
    })
    return Object.entries(groups).sort((a, b) => a[0].localeCompare(b[0]))
  }, [filtered])

  const [expandedProvs, setExpandedProvs] = useState<Record<string, boolean>>({})
  const prevGroupKeys = useRef('')
  useEffect(() => {
    const keys = groupedByProvider.map(([p]) => p).join(',')
    if (keys !== prevGroupKeys.current) {
      prevGroupKeys.current = keys
      setExpandedProvs(prev => {
        const next: Record<string, boolean> = {}
        groupedByProvider.forEach(([p], i) => { next[p] = i < 3 || (prev[p] === true) })
        return next
      })
    }
  }, [groupedByProvider])

  return (
    <div className="min-h-screen bg-background overflow-x-hidden" style={{backgroundImage:'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)'}}>
      {/* Mobile overlay */}
      {mobileOpen && (
        <div className="fixed inset-0 z-50 md:hidden" onClick={()=>setMobileOpen(false)}>
          <div className="absolute inset-0 bg-black/15" />
          <div className="absolute left-0 top-0 h-full w-64 bg-white shadow-xl z-50 overflow-y-auto" onClick={e=>e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-border/20">
              <span className="text-sm font-bold tracking-tight">{t('Filters')}</span>
              <button onClick={()=>setMobileOpen(false)} className="w-7 h-7 rounded-lg hover:bg-muted flex items-center justify-center text-muted-foreground text-sm">✕</button>
            </div>
            <ModelFilterSidebar filters={filters} hovered={true} onEnter={()=>{}} onLeave={()=>{}}
              useCases={useCases} aiProviders={aiProviders} quantumProviders={quantumProviders} contextBuckets={contextBuckets} />
          </div>
        </div>
      )}

      <div className="qc-wrap qc-section-pad-sm">
        <div className="mb-6">
          <PromoCarousel pageKey="models" />
        </div>
        <div className="relative">
          {/* Desktop sidebar */}
          <ModelFilterSidebar
            filters={filters}
            hovered={hovered}
            onEnter={handleSidebarEnter}
            onLeave={handleSidebarLeave}
            useCases={useCases}
            aiProviders={aiProviders}
            quantumProviders={quantumProviders}
            contextBuckets={contextBuckets}
          />

          {/* Main area */}
          <div className="max-w-[100vw] overflow-x-hidden" style={{ paddingLeft: hovered ? '304px' : '72px' }}>
            <div className="flex-1 min-w-0 space-y-6">
            {/* Search + sort bar */}
            <div className="qc-fade-up flex items-center gap-3 flex-wrap">
              <button onClick={()=>setMobileOpen(true)}
                className="md:hidden w-9 h-9 rounded-xl border border-border/30 hover:bg-muted/40 flex items-center justify-center text-muted-foreground">
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="4" y1="6" x2="20" y2="6"/><line x1="8" y1="12" x2="20" y2="12"/><line x1="12" y1="18" x2="20" y2="18"/></svg>
              </button>
              <div className="relative flex-1 min-w-[160px] max-w-xs">
                <svg className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/30" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
                <input value={q} onChange={e=>setQ(e.target.value)}
                  className="w-full h-10 rounded-xl border border-border/30 bg-white/70 px-10 text-sm placeholder:text-muted-foreground/40 outline-none focus:border-[oklch(0.72_0.18_52)]/40 focus:bg-white transition-all"
                  placeholder={t('Search')+' '+all.length+' '+t('models')+'...'} />
              </div>
              <select value={sort} onChange={e=>setSort(e.target.value as SortOpt)}
                className="h-10 rounded-xl border border-border/30 bg-white/70 px-3 text-sm outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all">
                <option value="name">{t('Name')}</option>
                <option value="price-asc">{t('Price \u2191')}</option>
                <option value="price-desc">{t('Price \u2193')}</option>
              </select>
              {sel.size>=2 && (
                <div className="flex items-center gap-2 bg-amber-50 rounded-xl px-3 py-1.5 border border-amber-200/50">
                  <span className="text-xs font-semibold text-amber-700">{sel.size}</span>
                  <button onClick={()=>setCop(true)}
                    className="px-3 py-1 text-xs font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 rounded-lg hover:shadow-md transition-all">
                    {t('Compare')}
                  </button>
                  <button onClick={()=>setSel(new Set)} className="p-1 text-muted-foreground/50 hover:text-muted-foreground">
                    <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
                  </button>
                </div>
              )}
            </div>

            {/* Active filters */}
            {(prov||cat!=='all'||ctx) && (
              <div className="qc-fade-up flex items-center gap-2 flex-wrap text-xs">
                {prov&&<Tag onRemove={()=>setProv('')}>{prov}</Tag>}
                {cat!=='all'&&<Tag onRemove={()=>setCat('all')}>{useCaseMeta[cat]?.label||cat}</Tag>}
                {ctx&&<Tag onRemove={()=>setCtx('')}>ctx {ctx.replace('-','\u2013')}</Tag>}
                <button onClick={()=>{setProv('');setCat('all');setCtx('')}}
                  className="text-muted-foreground/50 hover:text-muted-foreground underline underline-offset-2 transition-colors ml-1">
                  {t('Clear all')}
                </button>
              </div>
            )}

            {/* Results or empty */}
            {filtered.length===0 ? (
              <div className="text-center py-24 text-muted-foreground">
                <svg className="h-12 w-12 mx-auto mb-4 opacity-10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
                <p className="text-lg font-medium mb-2">{t('No models found')}</p>
                <p className="text-sm mb-6 text-muted-foreground/60">{t('Try adjusting your filters')}</p>
                <button onClick={()=>{setQ('');setCat('all');setProv('');setCtx('')}}
                  className="px-5 py-2.5 rounded-xl border border-border/30 bg-white hover:bg-muted/40 text-sm font-medium transition-all">{t('Reset filters')}</button>
              </div>
            ) : (
              <>
                <p className="qc-fade-up text-xs text-muted-foreground/40 font-medium tracking-wide">{filtered.length} {t('models')}</p>
                <div className="space-y-6">
                  {groupedByProvider.map(([provider, models], gi) => {
                    const isOpen = expandedProvs[provider] !== false
                    return (
                      <div key={provider} className="qc-fade-up" style={{animationDelay:''+((gi%10)*0.03)+'s'}}>
                        <button onClick={() => setExpandedProvs(prev => ({ ...prev, [provider]: !prev[provider] }))}
                          className="w-full flex items-start gap-2 sm:gap-3 px-3 sm:px-4 py-3 rounded-2xl bg-white/40 hover:bg-white/80 border border-border/10 hover:border-border/30 transition-all duration-200 group flex-wrap">
                          <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-orange-600 font-bold text-sm shrink-0 shadow-sm">{provider.charAt(0).toUpperCase()}</div>
                          <div className="flex-1 text-left min-w-0"><span className="text-sm sm:text-base font-bold tracking-tight text-foreground group-hover:text-[oklch(0.72_0.18_52)] transition-colors">{provider}</span></div>
                          <span className="text-xs font-semibold text-muted-foreground/40 mr-3">{models.length} {t('models')}</span>
                          <svg className={'w-4 h-4 text-muted-foreground/40 transition-transform duration-200 '+(isOpen?'rotate-0':'-rotate-90')} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 9l6 6 6-6"/></svg>
                        </button>
                        <div className={'overflow-hidden transition-all duration-200 '+(isOpen?'max-h-[9999px] opacity-100':'max-h-0 opacity-0')}>
                          <div className="space-y-2 mt-3 pl-4">
                            {models.map((m: any) => {
                              const isSel = sel.has(m.name)
                              return (
                                <div key={m.name} className={'rounded-2xl p-3 sm:p-4 transition-all duration-200 '+(isSel?'bg-amber-50/70 ring-1 ring-amber-200/50':'bg-white/50 hover:bg-white/80 hover:shadow-sm')+' border border-border/10'}>
                                  <div className="flex items-start gap-3">
                                    <div className="flex-1 min-w-0">
                                      <div className="flex items-start justify-between gap-2">
                                        <h3 onClick={()=>{setDm(m);setDo_(true)}} className="text-base font-bold tracking-tight cursor-pointer text-foreground hover:text-[oklch(0.72_0.18_52)] transition-colors">{m.name}</h3>
                                        <div className="flex items-center gap-1.5 shrink-0">
                                          <button onClick={()=>{setDm(m);setDo_(true)}} className="text-[11px] font-medium text-muted-foreground/50 hover:text-foreground transition-colors md:opacity-0 md:group-hover:opacity-100">{t('Details')}</button>
                                          <Link to={auth?.user?'/chat':'/sign-in'} className="text-[11px] font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 px-3 py-1 rounded-lg hover:shadow-md hover:-translate-y-0.5 transition-all inline-block">{t('Call')}</Link>
                                          <button onClick={e=>{e.stopPropagation();toggleSel(m.name)}} className={'w-4 h-4 rounded border-2 flex items-center justify-center shrink-0 transition-all duration-200 '+(isSel?'bg-[oklch(0.72_0.18_52)] border-[oklch(0.72_0.18_52)] text-white scale-110':'border-muted-foreground/20 hover:border-muted-foreground/50')}>
                                            {isSel && <svg className="h-2.5 w-2.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><path d="M20 6L9 17l-5-5"/></svg>}
                                          </button>
                                        </div>
                                      </div>
                                      <p className="text-xs text-muted-foreground/60 leading-relaxed mt-1.5 max-w-prose break-words overflow-hidden">{m.description || 'No description'}</p>
                                      <div className="flex items-center gap-1.5 flex-wrap mt-2">
                                        <span className="inline-flex items-center px-2 py-0.5 rounded-lg text-[10px] max-w-full overflow-hidden font-medium bg-amber-50 text-amber-700 border border-amber-200/40">{Math.round(m.context_window/1000)}K ctx</span>
                                        {m.use_case && useCaseMeta[m.use_case] && (
                                          <span className={'px-2 py-0.5 rounded-lg text-[10px] font-medium text-white bg-gradient-to-r '+useCaseMeta[m.use_case].gradient}>{useCaseMeta[m.use_case].icon} {useCaseMeta[m.use_case].label}</span>
                                        )}
                                        {(m.input_modalities||[]).slice(0,3).map((mod:string)=>(
                                          <span key={mod} className="px-2 py-0.5 rounded-lg text-[10px] bg-muted/40 text-muted-foreground/60 border border-border/20 capitalize">{mod}</span>
                                        ))}
                                        {m.input_price>0
                                          ? <span className="px-2 py-0.5 rounded-lg text-[10px] font-medium text-emerald-700 bg-emerald-50 border border-emerald-200/40">${m.input_price.toFixed(5)} in</span>
                                          : <span className="px-2 py-0.5 rounded-lg text-[10px] font-medium text-emerald-700 bg-emerald-50 border border-emerald-200/40">Free</span>
                                        }
                                      </div>
                                    </div>
                                  </div>
                                </div>
                              )
                            })}
                          </div>
                        </div>
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

      {dm && <ModelDetailDialog model={dm} open={do_} onOpenChange={setDo_} />}
      <ModelComparisonDialog models={selList} open={cop} onOpenChange={setCop} />
    </div>
  )
}

function Tag({children,onRemove}:{children:React.ReactNode;onRemove:()=>void}) {
  return (
    <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-gradient-to-r from-amber-50 to-orange-50 text-amber-700 border border-amber-200/50 shadow-sm">
      {children}
      <button onClick={onRemove} className="ml-0.5 hover:text-amber-900 transition-colors">\u2715</button>
    </span>
  )
}
