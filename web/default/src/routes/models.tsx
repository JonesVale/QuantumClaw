import { createFileRoute, Link } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo, useEffect } from 'react'
import { useAuthStore } from '@/stores/auth-store'
import { ModelDetailDialog, type CatalogItem } from '@/components/model-detail-dialog'
import { ModelComparisonDialog } from '@/components/model-comparison-dialog'

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

/** Left sidebar: categories + providers + context */
const NAV = [
  { id:'all', icon:'⊞', label:'All Models' },
  { id:'chat', icon:'💬', label:'Chat' },
  { id:'coding', icon:'</>', label:'Code' },
  { id:'reasoning', icon:'🧠', label:'Reasoning' },
  { id:'vision', icon:'👁', label:'Vision' },
]

function ModelsPage() {
  const { t, language } = useT()
  const { auth } = useAuthStore()
  const [q, setQ] = useState('')
  const [cat, setCat] = useState('all')
  const [prov, setProv] = useState('')
  const [ctx, setCtx] = useState('')
  const [sort, setSort] = useState<SortOpt>('name')
  const [collapse, setCollapse] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [dm, setDm] = useState<CatalogItem|null>(null)
  const [do_, setDo_] = useState(false)
  const [vis, setVis] = useState(30)
  const STEP = 30
  const [sel, setSel] = useState<Set<string>>(new Set)
  const [cop, setCop] = useState(false)

  const lang = language || 'English'
  const { data } = useQuery({
    queryKey:['model-catalog',lang],
    queryFn:async()=>{const r=await fetch(`/api/model-catalog?lang=${encodeURIComponent(lang)}`);if(!r.ok)throw Error();return r.json()},
    staleTime:60_000,
  })
  const all: CatalogItem[] = data?.data || []

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
    if(ctx){const c=(m:any)=>m.context_window||0;switch(ctx){
      case'0-8192':r=r.filter(m=>c(m)<=8192);break
      case'8193-32768':r=r.filter(m=>c(m)>8192&&c(m)<=32768);break
      case'32769-131072':r=r.filter(m=>c(m)>32768&&c(m)<=131072);break
      case'131073-999999999':r=r.filter(m=>c(m)>131072);break
    }}
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

  /** Sidebar panel (shared between desktop + mobile overlay) */
  const Sidebar = ({compact=false,onClose}:{compact?:boolean;onClose?:()=>void}) => (
    <div className={`${compact?'p-2':'p-5'} space-y-1`}>
      {compact ? (
        NAV.map(n=>
          <button key={n.id} onClick={()=>{setCat(n.id);onClose?.()}}
            className={`w-full flex items-center justify-center h-10 rounded-xl transition-all duration-200 text-lg ${cat===n.id ? 'bg-[oklch(0.72_0.18_52)]/10 text-[oklch(0.72_0.18_52)]' : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'}`}
            title={n.label}>{n.icon}</button>
        )
      ) : (
        <>
          <div className="flex items-center justify-between mb-4">
            <span className="text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.15em]">{t('Browse')}</span>
            <button onClick={()=>setCollapse(true)} className="w-6 h-6 rounded-lg hover:bg-muted/50 flex items-center justify-center text-muted-foreground/50 text-xs">◀</button>
          </div>
          {NAV.map(n=>
            <button key={n.id} onClick={()=>setCat(n.id)}
              className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${cat===n.id ? 'bg-gradient-to-r from-amber-50 to-orange-50 text-amber-800 shadow-sm' : 'text-muted-foreground hover:text-foreground hover:bg-muted/40'}`}>
              <span className="text-base w-5 text-center">{n.icon}</span>
              <span>{t(n.label)}</span>
            </button>
          )}
          <hr className="my-5 border-border/30" />
          <span className="text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.15em] px-3 block mb-2">{t('Providers')}</span>
          <div className="space-y-0.5 max-h-52 overflow-y-auto thin-scroll">
            {providers.map(([p,c])=>
              <button key={p} onClick={()=>setProv(prov===p?'':p)}
                className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm transition-all ${prov===p ? 'bg-amber-50 text-amber-800 font-medium' : 'text-muted-foreground hover:text-foreground hover:bg-muted/30'}`}>
                <span>{p}</span>
                <span className="text-xs text-muted-foreground/40">{c}</span>
              </button>
            )}
          </div>
          <hr className="my-5 border-border/30" />
          <span className="text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.15em] px-3 block mb-2">{t('Context')}</span>
          {[{v:'',l:'All'},{v:'0-8192',l:'≤ 8K'},{v:'8193-32768',l:'8K–32K'},{v:'32769-131072',l:'32K–128K'},{v:'131073-999999999',l:'≥ 128K'}].map(r=>
            <button key={r.v} onClick={()=>setCtx(r.v)}
              className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-all ${ctx===r.v ? 'bg-amber-50 text-amber-800 font-medium' : 'text-muted-foreground hover:text-foreground hover:bg-muted/30'}`}>
              {r.l}</button>
          )}
        </>
      )}
    </div>
  )

  return (
    <div className="min-h-screen bg-background" style={{backgroundImage:'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)'}}>
      {/* ═══ NAV ═══ — provided by __root.tsx */}
      {/* Mobile overlay */}
      {mobileOpen && (
        <div className="fixed inset-0 z-50 md:hidden" onClick={()=>setMobileOpen(false)}>
          <div className="absolute inset-0 bg-black/15" />
          <div className="absolute left-0 top-0 h-full w-64 bg-white shadow-xl z-50 overflow-y-auto" onClick={e=>e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-border/20">
              <span className="text-sm font-bold tracking-tight">{t('Filters')}</span>
              <button onClick={()=>setMobileOpen(false)} className="w-7 h-7 rounded-lg hover:bg-muted flex items-center justify-center text-muted-foreground text-sm">✕</button>
            </div>
            <Sidebar onClose={()=>setMobileOpen(false)} />
          </div>
        </div>
      )}

      <div className="qc-wrap qc-section-pad-sm">
        {/* ─── Header ─── */}
        <div className="qc-fade-up mb-10">
          <h1 className="qc-title-hero font-bold tracking-tight text-foreground">
            {t('AI Model Catalog')}
          </h1>
          <p className="qc-text-body qc-readable-width text-muted-foreground/70 mt-2 leading-relaxed">
            {t('Browse and compare models from all major providers')}
          </p>
        </div>

        <div className="flex gap-8">
          {/* ─── Desktop sidebar ─── */}
          <div className={`hidden md:block shrink-0 transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] ${collapse?'w-16':'w-56'}`}>
            <div className="sticky top-24 bg-white/60 backdrop-blur-xl rounded-2xl border border-border/20 shadow-sm">
              {collapse ? (
                <div className="p-2 space-y-1">
                  {NAV.map(n=>
                    <button key={n.id} onClick={()=>setCat(n.id)}
                      className={`w-full flex items-center justify-center h-10 rounded-xl transition-all duration-200 text-lg ${cat===n.id ? 'bg-[oklch(0.72_0.18_52)]/10 text-[oklch(0.72_0.18_52)]' : 'text-muted-foreground hover:bg-muted/40 hover:text-foreground'}`}
                      title={n.label}>{n.icon}</button>
                  )}
                  <hr className="my-2 border-border/20" />
                  <button onClick={()=>setCollapse(false)} className="w-full flex items-center justify-center h-8 rounded-lg text-muted-foreground/40 hover:text-muted-foreground text-xs">▶</button>
                </div>
              ) : (
                <Sidebar compact={false} />
              )}
            </div>
          </div>

          {/* ─── Main area ─── */}
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
                  placeholder={`${t('Search')} ${all.length} ${t('models')}...`} />
              </div>
              <select value={sort} onChange={e=>setSort(e.target.value as SortOpt)}
                className="h-10 rounded-xl border border-border/30 bg-white/70 px-3 text-sm outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all">
                <option value="name">{t('Name')}</option>
                <option value="price-asc">{t('Price ↑')}</option>
                <option value="price-desc">{t('Price ↓')}</option>
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
                {ctx&&<Tag onRemove={()=>setCtx('')}>ctx {ctx.replace('-','–')}</Tag>}
                <button onClick={()=>{setProv('');setCat('all');setCtx('')}}
                  className="text-muted-foreground/50 hover:text-muted-foreground underline underline-offset-2 transition-colors ml-1">
                  {t('Clear all')}
                </button>
              </div>
            )}

            {/* ─── Results or empty ─── */}
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
                {/* Result count */}
                <p className="qc-fade-up text-xs text-muted-foreground/40 font-medium tracking-wide">
                  {filtered.length} {t('models')}
                  {filtered.length!==all.length&&<span className="text-muted-foreground/30"> · {t('showing')} {shown.length}</span>}
                </p>

                {/* ─── Model list ─── */}
                <div className="space-y-3">
                  {shown.map((m:any,i:number)=>{
                    const isSel = sel.has(m.name)
                    return (
                      <div key={m.name}
                        className={`qc-fade-up group rounded-2xl p-5 transition-all duration-300 ${isSel ? 'bg-amber-50/70 ring-1 ring-amber-200/50' : 'bg-white/60 hover:bg-white/90 hover:shadow-sm'} border border-border/10`}
                        style={{animationDelay:`${(i%10)*0.05}s`}}>
                        <div className="flex items-start gap-4">
                          {/* Provider badge */}
                          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-orange-600 font-bold text-sm shrink-0 shadow-sm">
                            {(m.provider||'?').charAt(0).toUpperCase()}
                          </div>
                          <div className="flex-1 min-w-0">
                            {/* Top row */}
                            <div className="flex items-start justify-between gap-3">
                              <div className="min-w-0 flex-1">
                                <span className="text-[11px] font-semibold text-muted-foreground/40 uppercase tracking-[0.12em]">{m.provider||'Unknown'}</span>
                                <h3 onClick={()=>{setDm(m);setDo_(true)}}
                                  className="text-lg font-bold tracking-tight mt-0.5 cursor-pointer text-foreground group-hover:text-[oklch(0.72_0.18_52)] transition-colors duration-200">
                                  {m.name}
                                </h3>
                              </div>
                              <div className="flex items-center gap-2 shrink-0 mt-1">
                                <button onClick={()=>{setDm(m);setDo_(true)}}
                                  className="text-xs font-medium text-muted-foreground/50 hover:text-foreground transition-colors md:opacity-0 md:group-hover:opacity-100">
                                  {t('Details')}
                                </button>
                                <Link to={auth?.user?'/chat':'/sign-in'}
                                  className="text-xs font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 px-4 py-1.5 rounded-lg hover:shadow-md hover:-translate-y-0.5 transition-all md:opacity-0 md:group-hover:opacity-100 inline-block">
                                  {t('Call')}
                                </Link>
                                <button onClick={e=>{e.stopPropagation();toggleSel(m.name)}}
                                  className={`w-5 h-5 rounded border-2 flex items-center justify-center shrink-0 transition-all duration-200 ${isSel ? 'bg-[oklch(0.72_0.18_52)] border-[oklch(0.72_0.18_52)] text-white scale-110' : 'border-muted-foreground/20 hover:border-muted-foreground/50 opacity-0 md:group-hover:opacity-100'}`}>
                                  {isSel && <svg className="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><path d="M20 6L9 17l-5-5"/></svg>}
                                </button>
                              </div>
                            </div>

                            {/* Description */}
                            <p className="text-sm text-muted-foreground/60 leading-relaxed mt-2 max-w-prose">
                              {m.description || 'No description'}
                            </p>

                            {/* Tags row */}
                            <div className="flex items-center gap-2 flex-wrap mt-3">
                              <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-[11px] font-medium bg-amber-50 text-amber-700 border border-amber-200/40">
                                {(m.context_window/1000).toFixed(0)}K ctx
                              </span>
                              {m.use_case && useCaseMeta[m.use_case] && (
                                <span className={`px-2.5 py-1 rounded-lg text-[11px] font-medium text-white bg-gradient-to-r ${useCaseMeta[m.use_case].gradient}`}>
                                  {useCaseMeta[m.use_case].icon} {useCaseMeta[m.use_case].label}
                                </span>
                              )}
                              {(m.input_modalities||[]).slice(0,4).map((mod:string)=>(
                                <span key={mod} className="px-2.5 py-1 rounded-lg text-[11px] bg-muted/40 text-muted-foreground/60 border border-border/20 capitalize">{mod}</span>
                              ))}
                              {/* Pricing as tag */}
                              {m.input_price>0
                                ? <span className="px-2.5 py-1 rounded-lg text-[11px] font-medium text-emerald-700 bg-emerald-50 border border-emerald-200/40">
                                    ${m.input_price.toFixed(5)} in
                                  </span>
                                : <span className="px-2.5 py-1 rounded-lg text-[11px] font-medium text-emerald-700 bg-emerald-50 border border-emerald-200/40">Free</span>
                              }
                            </div>
                          </div>
                        </div>
                      </div>
                    )
                  })}
                </div>

                {/* Load more */}
                {vis<filtered.length && (
                  <div className="flex justify-center pt-4">
                    <button onClick={()=>setVis(v=>v+STEP)}
                      className="px-8 py-3 rounded-xl border border-border/30 bg-white/70 hover:bg-white hover:shadow-sm text-sm font-medium transition-all hover:-translate-y-0.5">
                      {t('Show more')} <span className="text-muted-foreground/60">({filtered.length-vis})</span>
                    </button>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>

      {dm && <ModelDetailDialog model={dm} open={do_} onOpenChange={setDo_} />}
      <ModelComparisonDialog models={selList} open={cop} onOpenChange={setCop} />
    </div>
  )
}

/** Filter tag pill */
function Tag({children,onRemove}:{children:React.ReactNode;onRemove:()=>void}) {
  return (
    <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-gradient-to-r from-amber-50 to-orange-50 text-amber-700 border border-amber-200/50 shadow-sm">
      {children}
      <button onClick={onRemove} className="ml-0.5 hover:text-amber-900 transition-colors">✕</button>
    </span>
  )
}
