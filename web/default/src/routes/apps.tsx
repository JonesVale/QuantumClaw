import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useEffect } from 'react'
import { PromoCarousel } from '@/components/promo-carousel'
import apiClient from '@/lib/api'

export const Route = createFileRoute('/apps')({
  component: AppsPage,
})

interface App {
  name: string
  description?: string
  app_url: string
  category: string
  author: string
  icon_url: string
}

const FALLBACK_APPS: App[] = [
  { name: 'Cursor', description: 'AI-native code editor', app_url: 'https://cursor.sh', category: 'development', author: 'Cursor', icon_url: '' },
  { name: 'ChatBox', description: 'AI chat desktop app', app_url: 'https://chatbox.app', category: 'chat', author: 'ChatBox', icon_url: '', featured: true as any },
  { name: 'Continue', description: 'Open-source code assistant', app_url: 'https://continue.dev', category: 'development', author: 'Continue', icon_url: '' },
  { name: 'Dify', description: 'LLM app development platform', app_url: 'https://dify.ai', category: 'platform', author: 'Dify', icon_url: '' },
]

function AppsPage() {
  const { t } = useT()
  const [search, setSearch] = useState('')
  const [apiApps, setApiApps] = useState<App[] | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apiClient.get('/api/apps')
      .then(res => {
        if (res.data?.success && Array.isArray(res.data.data) && res.data.data.length > 0) {
          setApiApps(res.data.data)
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const apps = apiApps ?? FALLBACK_APPS
  const filtered = apps.filter(a => !search || a.name.toLowerCase().includes(search.toLowerCase()))

  const getAppIcon = (name: string) => {
    const iconMap: Record<string, string> = {
      'cursor': '</>', 'chatbox': '💬', 'continue': '{ }', 'lobechat': '🤖',
      'open webui': '🌐', 'dify': '🗄', 'fastgpt': '⚡', 'cherry studio': '✨',
      'ai toolkit': '⭐',
    }
    return iconMap[name.toLowerCase()] || '🧩'
  }

  if (loading) {
    return (
      <div className="qc-wrap qc-section-pad-sm">
        <div className="flex items-center justify-center py-20 text-muted-foreground">
          <div className="flex flex-col items-center gap-3">
            <div className="w-8 h-8 border-2 border-amber-400 border-t-transparent rounded-full animate-spin" />
            <p className="text-sm">{t('Loading')}...</p>
          </div>
        </div>
      </div>
    )
  }

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
            <div className="space-y-2">
              <p className="text-xs font-semibold text-muted-foreground/40 uppercase tracking-[0.1em] mb-3 px-1">{t('All Apps')}</p>
              {filtered.map((app,i)=>(
                <a key={app.name + app.author} href={app.app_url} target="_blank" rel="noopener noreferrer"
                  className="qc-fade-up group flex items-start gap-4 px-3 sm:px-5 py-3 sm:py-4 rounded-2xl bg-white/60 hover:bg-white/90 transition-all border border-border/10 hover:shadow-sm" style={{animationDelay:`${(i%10)*0.04}s`}}>
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-lg shadow-sm shrink-0">{getAppIcon(app.name)}</div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <h4 className="text-sm font-semibold tracking-tight group-hover:text-[oklch(0.72_0.18_52)] transition-colors">{app.name}</h4>
                      <svg className="w-4 h-4 text-muted-foreground/30 group-hover:text-[oklch(0.72_0.18_52)] transition-colors shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M7 7h10v10M7 17L17 7"/></svg>
                    </div>
                    <p className="text-sm text-muted-foreground/60 mt-0.5 leading-relaxed">{app.description || app.name}</p>
                    <div className="flex items-center gap-3 mt-1.5">
                      <span className="text-xs font-medium px-2.5 py-0.5 rounded-lg bg-muted/50 text-muted-foreground/70">{app.category}</span>
                      <span className="text-xs text-muted-foreground/50">{app.author}</span>
                    </div>
                  </div>
                </a>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
