import { useState, useMemo } from 'react'
import { Link } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'

export interface CatalogItem { name:string; display_name?:string; description:string; use_case:string; context_window:number; input_modalities?:string[]; series?:string; provider:string; channel_id?:number; channel_name?:string; input_price:number; output_price:number; status:number; group?:string }

const useCaseMeta: Record<string,{label:string;icon:string;gradient:string}> = {
  chat:     { label:'Chat', icon:'💬', gradient:'from-amber-500 to-orange-500' },
  coding:   { label:'Code', icon:'</>', gradient:'from-emerald-500 to-teal-500' },
  reasoning:{ label:'Reasoning', icon:'🧠', gradient:'from-amber-600 to-rose-600' },
  vision:   { label:'Vision', icon:'👁', gradient:'from-orange-400 to-rose-400' },
}

export function useSidebarData(language?: string) {
  const lang = language || 'English'
  const { data } = useQuery({
    queryKey:['model-catalog',lang],
    queryFn:async()=>{const r=await fetch('/api/model-catalog?lang='+encodeURIComponent(lang));if(!r.ok)throw Error();return r.json()},
    staleTime:60_000,
  })
  const all: CatalogItem[] = data?.data || []

  const providers = useMemo(()=>{
    const m=new Map<string,number>()
    all.forEach(mt=>{const p=mt.provider;if(p&&p!=='Unknown'&&!p.startsWith('~'))m.set(p,(m.get(p)||0)+1)})
    return [...m].sort((a,b)=>b[1]-a[1])
  },[all])

  const quantumProviderNames = ['IonQ','IBM','Rigetti','Azure Quantum','Google Quantum','AWS Braket']
  const aiProviders = useMemo(()=>providers.filter(p=>!quantumProviderNames.includes(p[0])),[providers])
  const quantumProviders = useMemo(()=>providers.filter(p=>quantumProviderNames.includes(p[0])),[providers])

  const useCases = useMemo(() => {
    const s = new Set(all.map(m => m.use_case).filter(Boolean))
    const items: {id:string;icon:string;label:string;gradient:string}[] = [
      { id:'all', icon:'⊞', label:'All Models', gradient:'' }
    ]
    s.forEach(uc => {
      const meta = useCaseMeta[uc]
      items.push({ id: uc, icon: meta?.icon || '📦', label: meta?.label || uc, gradient: meta?.gradient || 'from-gray-400 to-gray-500' })
    })
    return items
  },[all])

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
        buckets.push({ v: low+'-'+high, l: fmt(low)+'–'+fmt(high) })
      }
    }
    return buckets
  },[all])

  return { all, aiProviders, quantumProviders, useCases, contextBuckets }
}

export interface SidebarFilters {
  cat: string
  setCat: (v: string) => void
  prov: string
  setProv: (v: string) => void
  ctx: string
  setCtx: (v: string) => void
}

export function ModelFilterSidebar({
  filters, hovered, onEnter, onLeave,
  useCases, aiProviders, quantumProviders, contextBuckets, showContext = true
}: {
  filters: SidebarFilters
  hovered: boolean
  onEnter: () => void
  onLeave: () => void
  useCases: {id:string;icon:string;label:string;gradient:string}[]
  aiProviders: [string,number][]
  quantumProviders: [string,number][]
  contextBuckets: {v:string;l:string}[]
  showContext?: boolean
}) {
  const { t } = useT()
  const { cat, setCat, prov, setProv, ctx, setCtx } = filters
  const [groups, setGroups] = useState({categories:true,providers:true,context:true})
  const toggle = (key: string) => setGroups(p => ({...p, [key]: !(p as any)[key]}))

  return (
    <div
      className="hidden md:block absolute left-0 z-40"
      style={{ top: '0' }}
      onMouseEnter={onEnter}
      onMouseLeave={onLeave}
    >
      <div
        className="bg-white/60 backdrop-blur-xl rounded-2xl border border-border/20 shadow-sm overflow-hidden transition-[width] duration-200 ease-out"
        style={{ width: hovered ? '288px' : '56px' }}
      >
        {!hovered ? (
          <div className="p-2 space-y-1">
            {useCases.map(n=>
              <button key={n.id} onClick={()=>setCat(n.id)}
                className={'w-full flex items-center justify-center h-10 rounded-xl transition-all duration-200 text-lg '+(cat===n.id?'bg-[oklch(0.72_0.18_52)]/10 text-[oklch(0.72_0.18_52)]':'text-muted-foreground hover:bg-muted/50 hover:text-foreground')}
                title={n.label}>{n.icon}</button>
            )}
          </div>
        ) : (
          <div className="p-4 space-y-1">
            {/* Browse */}
            <div>
              <button onClick={()=>toggle('categories')}
                className="w-full flex items-center gap-2 px-3 py-2.5 mb-1 rounded-lg hover:bg-muted/30 transition-colors text-left">
                <div className="w-0.5 h-3 rounded-full bg-accent shrink-0" />
                <svg className={'h-3 w-3 text-muted-foreground/40 transition-transform duration-200 '+(groups.categories?'':'-rotate-90')} viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M6 9l6 6 6-6"/></svg>
                <span className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em]">{t('Browse')}</span>
              </button>
              <div className={'overflow-hidden transition-all duration-200 '+(groups.categories?'max-h-[999px] opacity-100':'max-h-0 opacity-0')}>
                <div className="space-y-0.5 pl-2">
                  {useCases.map(n=>
                    <button key={n.id} onClick={()=>setCat(n.id)}
                      className={'w-full flex items-center gap-4 px-4 py-3 rounded-xl text-xl font-medium transition-all duration-200 '+(cat===n.id?'bg-gradient-to-r from-amber-50 to-orange-50 text-amber-800 shadow-sm':'text-muted-foreground hover:text-foreground hover:bg-muted/40')}>
                      <span className="text-lg w-6 text-center">{n.icon}</span>
                      <span>{n.label}</span>
                    </button>
                  )}
                </div>
              </div>
              {groups.categories && (
                <Link to="/rankings" className="block px-6 py-1.5 text-lg text-muted-foreground/50 hover:text-accent transition-colors">{t('View Rankings')} →</Link>
              )}
            </div>

            <div className="my-3 border-t border-border/10" />

            {/* AI Providers */}
            <div>
              <button onClick={()=>toggle('providers')}
                className="w-full flex items-center gap-2 px-3 py-2.5 mb-1 rounded-lg hover:bg-muted/30 transition-colors text-left">
                <div className="w-0.5 h-3 rounded-full bg-accent shrink-0" />
                <svg className={'h-3 w-3 text-muted-foreground/40 transition-transform duration-200 '+(groups.providers?'':'-rotate-90')} viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M6 9l6 6 6-6"/></svg>
                <span className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em]">{t('AI Providers')}</span>
              </button>
              <div className={'overflow-hidden transition-all duration-200 '+(groups.providers?'max-h-[9999px] opacity-100':'max-h-0 opacity-0')}>
                <div className="space-y-0.5 pl-2">
                  <button onClick={()=>setProv('')}
                    className={'w-full text-left px-4 py-2.5 rounded-lg text-xl transition-all '+(prov===''?'bg-amber-50 text-amber-800 font-medium':'text-muted-foreground hover:text-foreground hover:bg-muted/30')}>
                    <span>{t('All Providers')}</span>
                  </button>
                  {aiProviders.map(([p,c])=>(
                    <button key={p} onClick={()=>setProv(prov===p?'':p)}
                      className={'w-full flex items-center justify-between px-4 py-2.5 rounded-lg text-xl transition-all '+(prov===p?'bg-amber-50 text-amber-800 font-medium':'text-muted-foreground hover:text-foreground hover:bg-muted/30')}>
                      <span>{p}</span>
                      <span className="text-lg text-muted-foreground/40">{c}</span>
                    </button>
                  ))}
                </div>
              </div>
              {groups.providers && quantumProviders.length>0 && (
                <div className="mt-4">
                  <div className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em] px-4 mb-2">{t('量子资源')}</div>
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

            {showContext && (<>
              <div className="my-3 border-t border-border/10" />
              {/* Context */}
              <div>
                <button onClick={()=>toggle('context')}
                  className="w-full flex items-center gap-2 px-3 py-2.5 mb-1 rounded-lg hover:bg-muted/30 transition-colors text-left">
                  <div className="w-0.5 h-3 rounded-full bg-accent shrink-0" />
                  <svg className={'h-3 w-3 text-muted-foreground/40 transition-transform duration-200 '+(groups.context?'':'-rotate-90')} viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M6 9l6 6 6-6"/></svg>
                  <span className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em]">{t('Context')}</span>
                </button>
                <div className={'overflow-hidden transition-all duration-200 '+(groups.context?'max-h-[999px] opacity-100':'max-h-0 opacity-0')}>
                  <div className="space-y-0.5 pl-2">
                    {contextBuckets.map(r=>
                      <button key={r.v} onClick={()=>setCtx(r.v)}
                        className={'w-full text-left px-4 py-2.5 rounded-lg text-xl transition-all '+(ctx===r.v?'bg-amber-50 text-amber-800 font-medium':'text-muted-foreground hover:text-foreground hover:bg-muted/30')}>
                        {r.l}</button>
                    )}
                  </div>
                </div>
              </div>
            </>)}
          </div>
        )}
      </div>
    </div>
  )
}
