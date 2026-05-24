import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { PromoCarousel } from '@/components/promo-carousel'
import { useState } from 'react'

export const Route = createFileRoute('/apps')({
  component: AppsPage,
})

interface App { name: string; description: string; url: string; category: string; icon: string; users: string; featured?: boolean }

const apps: App[] = [
  { name:'Cursor', description:'AI-first code editor with multi-model support. Integrates QuantumClaw API for intelligent code completion.', url:'https://cursor.sh', category:'Development', icon:'</>', users:'1.2M+' },
  { name:'ChatBox', description:'All-in-one AI desktop client supporting all major LLMs through QuantumClaw gateway.', url:'https://chatbox.app', category:'Chat', icon:'💬', users:'800K+', featured:true },
  { name:'Continue', description:'Open-source AI code assistant for VS Code & JetBrains, powered by QuantumClaw API.', url:'https://continue.dev', category:'Development', icon:'{ }', users:'500K+' },
  { name:'LobeChat', description:'Modern chat framework with plugin system and multi-model support.', url:'https://lobehub.com', category:'Chat', icon:'🤖', users:'300K+', featured:true },
  { name:'Open WebUI', description:'Self-hosted WebUI for LLMs with QuantumClaw API integration.', url:'https://openwebui.com', category:'Chat', icon:'🌐', users:'250K+' },
  { name:'Dify', description:'Open-source LLM app development platform. Build and deploy AI applications.', url:'https://dify.ai', category:'Platform', icon:'🗄', users:'200K+', featured:true },
  { name:'FastGPT', description:'Knowledge-based Q&A system built on LLMs and vector databases.', url:'https://fastgpt.in', category:'Platform', icon:'⚡', users:'150K+' },
  { name:'Cherry Studio', description:'Desktop client for LLMs with support for multiple AI services.', url:'https://cherry-ai.com', category:'Chat', icon:'✨', users:'100K+' },
  { name:'AI Toolkit', description:'Browser extension for ChatGPT, Gemini, Claude with QuantumClaw API support.', url:'https://aitoolkit.com', category:'Tools', icon:'⭐', users:'80K+' },
]

const categories = ['All', 'Development', 'Chat', 'Platform', 'Tools']

const CAT_ICONS: Record<string,string> = { 'All':'⊞', 'Development':'</>', 'Chat':'💬', 'Platform':'🗄', 'Tools':'🔧' }

function AppsPage() {
  const { t } = useT()
  const [cat, setCat] = useState('All')
  const [search, setSearch] = useState('')
  const [collapse, setCollapse] = useState(false)

  const filtered = apps.filter(a => (cat==='All'||a.category===cat) && (!search||a.name.toLowerCase().includes(search.toLowerCase())||a.description.toLowerCase().includes(search.toLowerCase())))

  return (
    <div className="min-h-screen bg-background" style={{backgroundImage:'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)'}}>
      <div className="qc-wrap qc-section-pad-sm">
        <div className="mb-8">
          <PromoCarousel pageKey="apps" />
        </div>
        <div className="qc-fade-up text-center mb-10">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-amber-200 bg-amber-50 text-amber-700 text-xs font-semibold tracking-wide mb-5">{apps.length}+ {t('integrations')}</div>
          <h1 className="qc-title-hero font-bold tracking-tight text-foreground">{t('Apps & Integrations')}</h1>
          <p className="qc-text-body qc-readable-width text-muted-foreground/70 mt-2 leading-relaxed mx-auto">{t('Applications and tools powered by QuantumClaw API')}</p>
        </div>

        <div className="flex gap-8">
          {/* Sidebar */}
          <div className={`hidden md:block shrink-0 transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] ${collapse?'w-16':'w-56'}`}>
            <div className="sticky top-24 bg-white/60 backdrop-blur-xl rounded-2xl border border-border/20 shadow-sm p-5 space-y-1">
              {collapse ? (
                <div className="space-y-1">
                  <button onClick={()=>setCollapse(false)} className="w-full h-10 rounded-xl hover:bg-muted/40 flex items-center justify-center text-muted-foreground/50 text-xs">▶</button>
                </div>
              ) : (
                <>
                  <div className="flex items-center justify-between mb-4">
                    <span className="text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.15em]">{t('Categories')}</span>
                    <button onClick={()=>setCollapse(true)} className="w-6 h-6 rounded-lg hover:bg-muted/50 flex items-center justify-center text-muted-foreground/50 text-xs">◀</button>
                  </div>
                  {categories.map(c=>(
                    <button key={c} onClick={()=>setCat(c)}
                      className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${cat===c?'bg-gradient-to-r from-amber-50 to-orange-50 text-amber-800 shadow-sm':'text-muted-foreground hover:text-foreground hover:bg-muted/40'}`}>
                      <span className="text-base w-5 text-center">{CAT_ICONS[c]||'⊞'}</span>
                      <span>{t(c)}</span>
                    </button>
                  ))}
                </>
              )}
            </div>
          </div>

          {/* Content */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 flex-wrap mb-6">
              <button onClick={()=>setCollapse(!collapse)} className="hidden md:flex w-9 h-9 rounded-xl border border-border/30 hover:bg-muted/40 items-center justify-center text-muted-foreground">
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="4" y1="6" x2="20" y2="6"/><line x1="8" y1="12" x2="20" y2="12"/><line x1="12" y1="18" x2="20" y2="18"/></svg>
              </button>
              <div className="relative flex-1 min-w-[160px] max-w-xs">
                <svg className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/30" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
                <input value={search} onChange={e=>setSearch(e.target.value)} className="w-full h-10 rounded-xl border border-border/30 bg-white/70 px-10 text-sm placeholder:text-muted-foreground/40 outline-none focus:border-[oklch(0.72_0.18_52)]/40 focus:bg-white transition-all" placeholder={t('Search apps...')} />
              </div>
            </div>

            {filtered.length===0?(
              <div className="text-center py-20 text-muted-foreground"><p className="text-lg font-medium mb-2">{t('No apps found')}</p><button onClick={()=>{setSearch('');setCat('All')}} className="mt-4 px-5 py-2.5 rounded-xl border border-border/30 bg-white hover:bg-muted/40 text-sm font-medium transition-all">{t('Reset filters')}</button></div>
            ):(
              <>
                <p className="text-xs text-muted-foreground/40 font-medium tracking-wide mb-4">{filtered.length} {t('apps')}</p>
                {filtered.filter(a=>a.featured).length>0&&(
                  <div className="mb-6">
                    <p className="text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] mb-3 px-1">{t('Featured')}</p>
                    <div className="space-y-3">
                      {filtered.filter(a=>a.featured).map((app,i)=>(
                        <a key={app.name} href={app.url} target="_blank" rel="noopener noreferrer"
                          className="qc-fade-up group flex items-start gap-4 px-5 py-4 rounded-2xl bg-white/70 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm" style={{animationDelay:`${i*0.08}s`}}>
                          <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-lg shadow-sm shrink-0">{app.icon}</div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                              <h3 className="text-base font-bold tracking-tight group-hover:text-[oklch(0.72_0.18_52)] transition-colors">{app.name}</h3>
                              <svg className="w-4 h-4 text-muted-foreground/30 group-hover:text-[oklch(0.72_0.18_52)] transition-colors shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M7 7h10v10M7 17L17 7"/></svg>
                            </div>
                            <p className="text-sm text-muted-foreground/60 mt-1 leading-relaxed">{app.description}</p>
                            <div className="flex items-center gap-3 mt-2"><span className="text-xs font-medium px-2.5 py-1 rounded-lg bg-amber-50 text-amber-700">{app.category}</span><span className="text-xs text-muted-foreground/50">{app.users} {t('users')}</span></div>
                          </div>
                        </a>
                      ))}
                    </div>
                  </div>
                )}
                <div className="space-y-2">
                  {filtered.filter(a=>!a.featured).length>0&&(
                    <p className="text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] mb-3 px-1">{t('All Apps')}</p>
                  )}
                  {filtered.map((app,i)=>!app.featured&&(
                    <a key={app.name} href={app.url} target="_blank" rel="noopener noreferrer"
                      className="qc-fade-up group flex items-start gap-4 px-5 py-4 rounded-2xl bg-white/60 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm" style={{animationDelay:`${(i%10)*0.04}s`}}>
                      <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-base shadow-sm shrink-0">{app.icon}</div>
                      <div className="flex-1 min-w-0">
                        <h4 className="text-sm font-semibold tracking-tight group-hover:text-[oklch(0.72_0.18_52)] transition-colors">{app.name}</h4>
                        <p className="text-sm text-muted-foreground/60 mt-0.5 leading-relaxed">{app.description}</p>
                        <div className="flex items-center gap-3 mt-1.5"><span className="text-xs font-medium px-2.5 py-0.5 rounded-lg bg-muted/50 text-muted-foreground/70">{app.category}</span><span className="text-xs text-muted-foreground/50">{app.users}</span></div>
                      </div>
                      <svg className="w-5 h-5 text-muted-foreground/20 group-hover:text-[oklch(0.72_0.18_52)] transition-colors shrink-0 mt-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M7 7h10v10M7 17L17 7"/></svg>
                    </a>
                  ))}
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
