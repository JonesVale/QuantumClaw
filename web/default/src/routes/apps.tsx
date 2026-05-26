import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { PromoCarousel } from '@/components/promo-carousel'

export const Route = createFileRoute('/apps')({
  component: AppsPage,
})

interface App { name: string; descKey: string; url: string; catKey: string; icon: string; users: string; featured?: boolean }

const apps: App[] = [
  { name:'Cursor', descKey:'app_cursor_desc', url:'https://cursor.sh', catKey:'app_category_Development', icon:'</>', users:'1.2M+' },
  { name:'ChatBox', descKey:'app_chatbox_desc', url:'https://chatbox.app', catKey:'app_category_Chat', icon:'💬', users:'800K+', featured:true },
  { name:'Continue', descKey:'app_continue_desc', url:'https://continue.dev', catKey:'app_category_Development', icon:'{ }', users:'500K+' },
  { name:'LobeChat', descKey:'app_lobechat_desc', url:'https://lobehub.com', catKey:'app_category_Chat', icon:'🤖', users:'300K+', featured:true },
  { name:'Open WebUI', descKey:'app_openwebui_desc', url:'https://openwebui.com', catKey:'app_category_Chat', icon:'🌐', users:'250K+' },
  { name:'Dify', descKey:'app_dify_desc', url:'https://dify.ai', catKey:'app_category_Platform', icon:'🗄', users:'200K+', featured:true },
  { name:'FastGPT', descKey:'app_fastgpt_desc', url:'https://fastgpt.in', catKey:'app_category_Platform', icon:'⚡', users:'150K+' },
  { name:'Cherry Studio', descKey:'app_cherry_desc', url:'https://cherry-ai.com', catKey:'app_category_Chat', icon:'✨', users:'100K+' },
  { name:'AI Toolkit', descKey:'app_aitoolkit_desc', url:'https://aitoolkit.com', catKey:'app_category_Tools', icon:'⭐', users:'80K+' },
]

function AppsPage() {
  const { t } = useT()
  const [search, setSearch] = useState('')

  const filtered = apps.filter(a => !search||a.name.toLowerCase().includes(search.toLowerCase())||t(a.descKey).toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="min-h-screen bg-background overflow-x-hidden" style={{backgroundImage:'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)'}}>
      <div className="qc-wrap qc-section-pad-sm">
        <div className="mb-6">
          <PromoCarousel pageKey="apps" />
        </div>
        <div className="flex items-center gap-3 flex-wrap mb-6">
          <div className="relative flex-1 min-w-[160px] max-w-xs">
            <svg className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/30" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
            <input value={search} onChange={e=>setSearch(e.target.value)} className="w-full h-10 rounded-xl border border-border/30 bg-white/70 px-10 text-sm placeholder:text-muted-foreground/40 outline-none focus:border-[oklch(0.72_0.18_52)]/40 focus:bg-white transition-all" placeholder={t('Search apps...')} />
          </div>
        </div>

        {filtered.length===0?(
          <div className="text-center py-20 text-muted-foreground"><p className="text-lg font-medium mb-2">{t('No apps found')}</p><button onClick={()=>setSearch('')} className="mt-4 px-5 py-2.5 rounded-xl border border-border/30 bg-white hover:bg-muted/40 text-sm font-medium transition-all">{t('Reset filters')}</button></div>
        ):(
          <>
            <p className="text-xs text-muted-foreground/40 font-medium tracking-wide mb-4">{filtered.length} {t('apps')}</p>
            {filtered.filter(a=>a.featured).length>0&&(
              <div className="mb-6">
                <p className="text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] mb-3 px-1">{t('Featured')}</p>
                <div className="space-y-3">
                  {filtered.filter(a=>a.featured).map((app,i)=>(
                    <a key={app.name} href={app.url} target="_blank" rel="noopener noreferrer"
                      className="qc-fade-up group flex items-start gap-4 px-3 sm:px-5 py-3 sm:py-4 rounded-2xl bg-white/70 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm" style={{animationDelay:`${i*0.08}s`}}>
                      <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-lg shadow-sm shrink-0">{app.icon}</div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <h3 className="text-base font-bold tracking-tight group-hover:text-[oklch(0.72_0.18_52)] transition-colors">{app.name}</h3>
                          <svg className="w-4 h-4 text-muted-foreground/30 group-hover:text-[oklch(0.72_0.18_52)] transition-colors shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M7 7h10v10M7 17L17 7"/></svg>
                        </div>
                        <p className="text-sm text-muted-foreground/60 mt-1 leading-relaxed">{t(app.descKey)}</p>
                        <div className="flex items-center gap-3 mt-2"><span className="text-xs font-medium px-2.5 py-1 rounded-lg bg-amber-50 text-amber-700">{t(app.catKey)}</span><span className="text-xs text-muted-foreground/50">{app.users} {t('users')}</span></div>
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
                  className="qc-fade-up group flex items-start gap-4 px-3 sm:px-5 py-3 sm:py-4 rounded-2xl bg-white/60 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm" style={{animationDelay:`${(i%10)*0.04}s`}}>
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-base shadow-sm shrink-0">{app.icon}</div>
                  <div className="flex-1 min-w-0">
                    <h4 className="text-sm font-semibold tracking-tight group-hover:text-[oklch(0.72_0.18_52)] transition-colors">{app.name}</h4>
                    <p className="text-sm text-muted-foreground/60 mt-0.5 leading-relaxed">{t(app.descKey)}</p>
                    <div className="flex items-center gap-3 mt-1.5"><span className="text-xs font-medium px-2.5 py-0.5 rounded-lg bg-muted/50 text-muted-foreground/70">{t(app.catKey)}</span><span className="text-xs text-muted-foreground/50">{app.users}</span></div>
                  </div>
                  <svg className="w-5 h-5 text-muted-foreground/20 group-hover:text-[oklch(0.72_0.18_52)] transition-colors shrink-0 mt-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M7 7h10v10M7 17L17 7"/></svg>
                </a>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
