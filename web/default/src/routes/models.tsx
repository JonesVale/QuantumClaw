import { createFileRoute, Link } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo, useEffect, useRef } from 'react'
import { useAuthStore } from '@/stores/auth-store'
import { PromoCarousel } from '@/components/promo-carousel'
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

function ModelsPage() {
  const { t, language } = useT()
  const { auth } = useAuthStore()
  const [q, setQ] = useState('')
  const [cat, setCat] = useState('all')
  const [prov, setProv] = useState('')
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
  const [sidebarOpenGroups, setSidebarOpenGroups] = useState<Record<string, boolean>>({categories:true,providers:true,context:true})
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

  const allProviders = providers
  const aiProviders = useMemo(()=>providers.filter(p => !["IonQ","IBM","Rigetti"].includes(p[0])),[providers])
  const quantumProviders = useMemo(()=>providers.filter(p => ["IonQ","IBM","Rigetti"].includes(p[0])),[providers])

  // Dynamic use cases from data
  const useCases = useMemo(() => {
    const s = new Set(all.map(m => m.use_case).filter(Boolean))
    const items: {id:string;icon:string;label:string;gradient:string}[] = [
      { id:'all', icon:'\u228e', label:'All Models', gradient:'' }
    ]
    s.forEach(uc => {
      const meta = useCaseMeta[uc]
      items.push({
        id: uc,
        icon: meta?.icon || '📦',
        label: meta?.label || uc,
        gradient: meta?.gradient || 'from-gray-400 to-gray-500'
      })
    })
    return items
  },[all])

  // Dynamic context window buckets from data
  const contextBuckets = useMemo(() => {
    const w = all.map(m => m.context_window || 0).filter(Boolean)
    const buckets: {v:string;l:string}[] = [{v:'', l:'All'}]
    if (w.length > 0) {
      const max = Math.max(...w)
      const steps = 4
      for (let i = 1; i <= steps; i++) {
        const low = Math.round(max * (i-1) / steps)
        const high = Math.round(max * i / steps)
        const fmt = (n:number) => n >= 1000000 ? (n/1000000).toFixed(1)+'M' : (n/1000).toFixed(0)+'K'
        buckets.push({ v: low+'-'+high, l: fmt(low)+'-'+fmt(high) })
      }
    }
    return buckets
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

  /** Sidebar panel (shared between desktop + mobile overlay) */
const SidebarContent = ({expanded=false,onClose}:{expanded?:boolean;onClose?:()=>void}) => (
    <div className="space-y-1">
      {!expanded ? (
        <div className="p-2 space-y-1">
          {useCases.map(n=>
            <button key={n.id} onClick={()=>{setCat(n.id);onClose?.()}}
              className={`w-full flex items-center justify-center h-10 rounded-xl transition-all duration-200 text-xl ${cat===n.id ? 'bg-[oklch(0.72_0.18_52)]/10 text-[oklch(0.72_0.18_52)]' : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'}`}
              title={n.label}>{n.icon}</button>
          )}
        </div>
      ) : (
        <div className="p-4 space-y-1">
          {/* Categories */}
          <div>
            <button onClick={() => setSidebarOpenGroups(prev => ({...prev, categories: !prev.categories}))}
              className="w-full flex items-center gap-2 px-3 py-2.5 mb-1 rounded-lg hover:bg-muted/30 transition-colors text-left">
              <div className="w-0.5 h-3 rounded-full bg-accent shrink-0" />
              <svg className={`h-3 w-3 text-muted-foreground/40 transition-transform duration-200 ${sidebarOpenGroups.categories ? '' : '-rotate-90'}`} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 9l6 6 6-6"/></svg>
              <span className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em]">{t('Browse')}</span>
            </button>
            <div className={`overflow-hidden transition-all duration-200 ${sidebarOpenGroups.categories ? 'max-h-[999px] opacity-100' : 'max-h-0 opacity-0'}`}>
              <div className="space-y-0.5 pl-2">
                {useCases.map(n=>
                  <button key={n.id} onClick={()=>setCat(n.id)}
                    className={`w-full flex items-center gap-4 px-4 py-3 rounded-xl text-xl font-medium transition-all duration-200 ${cat===n.id ? 'bg-gradient-to-r from-amber-50 to-orange-50 text-amber-800 shadow-sm' : 'text-muted-foreground hover:text-foreground hover:bg-muted/40'}`}>
                    <span className="text-lg w-6 text-center">{n.icon}</span>
                    <span>{n.label}</span>
                  </button>
                )}
              </div>
            </div>
            {sidebarOpenGroups.categories && (
              <Link to="/rankings" className="block px-6 py-1.5 text-lg text-muted-foreground/50 hover:text-accent transition-colors">
                {t('View Rankings')} →
              </Link>
            )}
          </div>
          <div className="my-3 border-t border-border/10" />
          {/* Providers */}
          <div>
            <button onClick={() => setSidebarOpenGroups(prev => ({...prev, providers: !prev.providers}))}
              className="w-full flex items-center gap-2 px-3 py-2.5 mb-1 rounded-lg hover:bg-muted/30 transition-colors text-left">
              <div className="w-0.5 h-3 rounded-full bg-accent shrink-0" />
              <svg className={`h-3 w-3 text-muted-foreground/40 transition-transform duration-200 ${sidebarOpenGroups.providers ? '' : '-rotate-90'}`} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 9l6 6 6-6"/></svg>
              <span className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em]">{t('Providers')}</span>
            </button>
            <div className={`overflow-hidden transition-all duration-200 ${sidebarOpenGroups.providers ? 'max-h-[9999px] opacity-100' : 'max-h-0 opacity-0'}`}>
              <div className="space-y-0.5 pl-2">
                <button onClick={()=>setProv('')}
                  className={`w-full text-left px-4 py-2.5 rounded-lg text-xl transition-all ${prov==='' ? 'bg-amber-50 text-amber-800 font-medium' : 'text-muted-foreground hover:text-foreground hover:bg-muted/30'}`}>
                  <span>{t('All Providers')}</span>
                </button>
                {aiProviders.map(([p,c])=>
                <button key={p} onClick={()=>setProv(prov===p?'':p)}
                  className={`w-full flex items-center justify-between px-4 py-2.5 rounded-lg text-xl transition-all ${prov===p ? 'bg-amber-50 text-amber-800 font-medium' : 'text-muted-foreground hover:text-foreground hover:bg-muted/30'}`}>
                  <span>{p}</span>
                  <span className="text-lg text-muted-foreground/40">{c}</span>
                </button>
              )}
              </div>
            </div>
            {sidebarOpenGroups.providers && aiProviders.length > 10 && (
              <Link to="/models" className="block px-6 py-1.5 text-lg text-muted-foreground/50 hover:text-accent transition-colors">
                {t('View All')} →
              </Link>
            )}
                      {sidebarOpenGroups.providers && quantumProviders.length > 0 && (
              <div className="mt-4">
                <div className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em] px-4 mb-2">量子资源</div>
                {quantumProviders.map(([p,c])=>(
                  <button key={p} onClick={()=>setProv(prov===p?'':p)}
                    className={'w-full flex items-center justify-between px-4 py-2.5 rounded-lg text-xl transition-all '+(prov===p?'bg-amber-50 text-amber-800 font-medium':'text-muted-foreground hover:text-foreground hover:bg-muted/30')}>
                    <span>{p}</span>
                    <span className="text-lg text-muted-foreground/40">{c}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
          <div className="my-3 border-t border-border/10" />
          {/* Context */}
          <div>
            <button onClick={() => setSidebarOpenGroups(prev => ({...prev, context: !prev.context}))}
              className="w-full flex items-center gap-2 px-3 py-2.5 mb-1 rounded-lg hover:bg-muted/30 transition-colors text-left">
              <div className="w-0.5 h-3 rounded-full bg-accent shrink-0" />
              <svg className={`h-3 w-3 text-muted-foreground/40 transition-transform duration-200 ${sidebarOpenGroups.context ? '' : '-rotate-90'}`} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 9l6 6 6-6"/></svg>
              <span className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em]">{t('Context')}</span>
            </button>
            <div className={`overflow-hidden transition-all duration-200 ${sidebarOpenGroups.context ? 'max-h-[999px] opacity-100' : 'max-h-0 opacity-0'}`}>
              <div className="space-y-0.5 pl-2">
                {contextBuckets.map(r=>
                  <button key={r.v} onClick={()=>setCtx(r.v)}
                    className={`w-full text-left px-4 py-2.5 rounded-lg text-xl transition-all ${ctx===r.v ? 'bg-amber-50 text-amber-800 font-medium' : 'text-muted-foreground hover:text-foreground hover:bg-muted/30'}`}>
                    {r.l}</button>
                )}
              </div>
            </div>
          </div>
        </div>
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
            <SidebarContent expanded={true} onClose={()=>setMobileOpen(false)} />
          </div>
        </div>
      )}

      <div className="qc-wrap qc-section-pad-sm">
        <div className="mb-6">
          <PromoCarousel pageKey="models" />
        </div>
        <div className="relative">
          {/* ─── Desktop sidebar ─── */}
          <div
            className="hidden md:block absolute left-0 z-40"
            style={{ top: '0' }}
            onMouseEnter={handleSidebarEnter}
            onMouseLeave={handleSidebarLeave}
          >
            <div
              className="bg-white/60 backdrop-blur-xl rounded-2xl border border-border/20 shadow-sm overflow-hidden transition-[width] duration-200 ease-out"
              style={{ width: hovered ? '288px' : '56px' }}
            >
              <SidebarContent expanded={hovered} />
            </div>
          </div>
          </div>

          {/* ─── Main area ─── */}
          <div className="transition-all duration-200" style={{ paddingLeft: hovered ? '304px' : '72px' }}>
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
