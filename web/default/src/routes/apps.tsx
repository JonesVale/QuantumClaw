import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
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
        <div className="qc-fade-up mb-8">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-amber-200 bg-amber-50 text-amber-700 text-xs font-semibold tracking-wide mb-5">{apps.length}+ {t('integrations')}</div>
          <h1 className="qc-title-hero font-bold tracking-tight text-foreground">{t('Apps & Integrations')}</h1>
          <p className="qc-text-body qc-readable-width text-muted-foreground/70 mt-2 leading-relaxed">{t('Applications and tools powered by QuantumClaw API')}</p>
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

            {filtered.filter(a=>a.featured).length>0&&(
              <div className="qc-grid-auto-sm gap-5 mb-8">
                {filtered.filter(a=>a.featured).map((app,i)=>(
                  <a key={app.name} href={app.url} target="_blank" rel="noopener noreferrer"
                    className="qc-fade-up qc-card-hover group block rounded-2xl bg-white/70 backdrop-blur-sm p-6 hover:bg-white/90 hover:shadow-md border border-border/10" style={{animationDelay:`${i*0.1}s`}}>
                    <div className="flex items-start gap-4">
                      <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-lg shadow-sm shrink-0">{app.icon}</div>
                      <div className="flex-1 min-w-0">
                        <h3 className="text-base font-bold tracking-tight group-hover:text-[oklch(0.72_0.18_52)] transition-colors">{app.name}</h3>
                        <p className="text-sm text-muted-foreground/60 mt-1 leading-relaxed">{app.description}</p>
                        <div className="flex items-center gap-3 mt-3"><span className="text-xs font-medium px-2.5 py-1 rounded-lg bg-amber-50 text-amber-700">{app.category}</span><span className="text-xs text-muted-foreground/50">{app.users} {t('users')}</span></div>
                      </div>
                    </div>
                  </a>
                ))}
              </div>
            )}

            {filtered.length===0?(
              <div className="text-center py-20 text-muted-foreground"><p className="text-lg font-medium mb-2">{t('No apps found')}</p><button onClick={()=>{setSearch('');setCat('All')}} className="mt-4 px-5 py-2.5 rounded-xl border border-border/30 bg-white hover:bg-muted/40 text-sm font-medium transition-all">{t('Reset filters')}</button></div>
            ):(
              <div className="qc-grid-auto-sm gap-4">
                {filtered.map((app,i)=>(
                  <a key={app.name} href={app.url} target="_blank" rel="noopener noreferrer"
                    className="qc-fade-up group block rounded-2xl bg-white/60 hover:bg-white/90 p-5 hover:shadow-sm border border-border/10 transition-all" style={{animationDelay:`${(i%12)*0.05}s`}}>
                    <div className="flex items-start gap-3">
                      <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-base shrink-0">{app.icon}</div>
                      <div className="min-w-0 flex-1">
                        <h4 className="text-sm font-semibold tracking-tight group-hover:text-[oklch(0.72_0.18_52)] transition-colors">{app.name}</h4>
                        <p className="text-xs text-muted-foreground/60 mt-0.5 line-clamp-2">{app.description}</p>
                        <div className="flex items-center gap-2 mt-2"><span className="text-[10px] font-medium px-2 py-0.5 rounded bg-muted/50 text-muted-foreground/70">{app.category}</span><span className="text-[10px] text-muted-foreground/50">{app.users}</span></div>
                      </div>
                    </div>
                  </a>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
