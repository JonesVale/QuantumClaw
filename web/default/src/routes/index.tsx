import { createFileRoute, Link } from '@tanstack/react-router'
import { useState, useEffect, useMemo } from 'react'
import { useT } from '@/lib/use-t'
import { PromoCarousel } from '@/components/promo-carousel'
import { useAuthStore } from '@/stores/auth-store'
import { useSidebarData } from '@/components/model-filter-sidebar'
export const Route = createFileRoute('/')({ component: HomePage })

const icons = {
  bolt: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>,
  layers: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>,
  shield: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 2l7 4v5c0 5-7 9-7 9s-7-4-7-9V6l7-4z"/></svg>,
  cpu: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/><path d="M9 1v3"/><path d="M15 1v3"/><path d="M9 20v3"/><path d="M15 20v3"/><path d="M1 9h3"/><path d="M20 9h3"/><path d="M1 15h3"/><path d="M20 15h3"/></svg>,
  arrowR: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>,
  menu: <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>,
}

interface SiteContent {
  stats: { key: string; value: string; label: string; detail: string }[]
  features: { key: string; title: string; desc: string; icon_name: string }[]
  providers: { name: string; models: string }[]
}

function HomePage() {
  const { t, language } = useT()
  const { auth } = useAuthStore()
  const loggedIn = !!auth.user
  const [companyUrl, setCompanyUrl] = useState('https://www.ctji.cn')
  const [icpBeian, setIcpBeian] = useState('')
  const [content, setContent] = useState<SiteContent | null>(null)

  useEffect(() => {
    // 使用公开的 site-info 端点（无需登录）
    fetch('/api/site-info')
      .then(r => r.json())
      .then((res: {success?: boolean; data?: Record<string, string>}) => {
        if (res?.success && res?.data) {
          if (res.data.company_website_url) setCompanyUrl(res.data.company_website_url)
          if (res.data.icp_beian) setIcpBeian(res.data.icp_beian)
        }
      })
      .catch(() => {})
  }, [])

  useEffect(() => {
    fetch('/api/site-content').then(r => r.json()).then(d => d.data && setContent(d.data)).catch(() => {})
  }, [])

  const stats = content?.stats ?? []
  const features = content?.features ?? []
  const providers = content?.providers ?? []

  // Dynamic provider list from API (fallback when CMS providers are empty)
  const { aiProviders, quantumProviders } = useSidebarData(language)
  const allProviders = useMemo(() => {
    if (providers.length > 0) return providers // CMS overrides
    const merged: typeof providers = []
    // AI brands first
    // Brand order from brand rankings + additional providers from catalog
    const topAI = ['OpenAI','Anthropic','Google','Meta','DeepSeek','Alibaba','Mistral','Baidu','Zhipu AI','Tencent','xAI','Cohere','Together AI','AWS','Groq']
    const seen = new Set<string>()
    // Order: top AI brands + quantum brands
    for (const name of [...topAI, ...['IonQ','IBM','Rigetti','Azure Quantum','Google Quantum']]) {
      const found = [...aiProviders, ...quantumProviders].find(([p]) => p === name)
      if (found) {
        merged.push({ name: found[0], models: found[1] + '+' })
        seen.add(found[0])
      }
    }
    // Any remaining providers not in the curated list
    for (const [name, count] of [...aiProviders, ...quantumProviders]) {
      if (!seen.has(name)) {
        merged.push({ name, models: count + '+' })
      }
    }
    return merged
  }, [providers, aiProviders, quantumProviders])

  return (
    <div className="min-h-screen bg-background">
      {/* ═══ NAV ═══ — provided by __root.tsx */}

      {/* ═══ HERO ═══ */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-amber-50/40 via-white to-orange-50/30" />
        <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[50vw] h-[50vw] rounded-full bg-gradient-to-bl from-amber-200/10 to-transparent blur-3xl pointer-events-none" />
        <div className="fixed top-1/2 left-1/2 -translate-x-1/3 -translate-y-1/3 w-[40vw] h-[40vw] rounded-full bg-gradient-to-tr from-orange-200/10 to-transparent blur-3xl pointer-events-none" />
        <div className="qc-wrap qc-section-pad-lg relative z-10 flex flex-col items-center text-center">
          <div className="qc-fade-up qc-fade-up-1 inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-orange-200 bg-orange-50 text-orange-700 text-xs font-semibold tracking-wide mb-8">
            <span className="w-2 h-2 rounded-full bg-orange-500" />
            {t('AI API Gateway')} · v2.0
          </div>
          <h1 className="qc-fade-up qc-fade-up-2 qc-title-hero font-bold tracking-tight text-foreground w-full">
            {t('Unified Access to')}<br /><span className="qc-gradient-text">{t('400+ AI Models')}</span>
          </h1>
          <p className="qc-fade-up qc-fade-up-3 qc-text-body inline-block max-w-3xl text-muted-foreground mt-6 leading-relaxed">
            {t('Access 400+ AI models and quantum computing resources through a single, unified endpoint. SDK compatible, real-time billing, transparent settlement.')}
          </p>
          <div className="qc-fade-up qc-fade-up-4 flex flex-wrap items-center justify-center gap-4 mt-10">
            <Link to="/models" className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-amber-500 to-orange-500 text-white px-8 py-3.5 text-base font-semibold shadow-lg shadow-orange-500/20 hover:shadow-xl hover:shadow-orange-500/30 hover:-translate-y-0.5 transition-all">{t('Browse Models')} {icons.arrowR}</Link>
            {!loggedIn && <Link to="/sign-in" className="inline-flex items-center gap-2 rounded-xl border border-border bg-white px-8 py-3.5 text-base font-medium text-foreground hover:bg-muted hover:-translate-y-0.5 transition-all">{t('Start Free')}</Link>}
          </div>
          <div className="qc-fade-up qc-fade-up-5 flex flex-wrap items-center justify-center gap-8 mt-16 text-sm text-muted-foreground">
            {[{ icon: icons.bolt, text: t('No GPU Required'), key: 'No GPU Required' }, { icon: icons.layers, text: t('Pay Per Token'), key: 'Pay Per Token' }, { icon: icons.shield, text: t('99.9% Uptime'), key: '99.9% Uptime' }, { icon: icons.cpu, text: t('Reseller Ready'), key: 'Reseller Ready' }].map(item => (
              <span key={item.key} className="flex items-center gap-2.5"><span className="text-orange-500">{item.icon}</span>{t(item.key)}</span>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ PROMO CAROUSEL ═══ */}
      <section className="qc-section-pad-sm">
        <div className="qc-wrap">
          <PromoCarousel large pageKey="home" />
        </div>
      </section>

      {/* ═══ STATS ═══ */}
      <section className="qc-fade-up border-y border-border/30 bg-white">
        <div className="qc-wrap qc-section-pad-sm">
          <div className="qc-grid-auto gap-8">
            {stats.length > 0 ? stats.map(s => (
              <div key={s.key} className="text-center qc-fade-up qc-fade-up-2">
                <div className="qc-stat-number font-bold qc-gradient-text">{s.value}</div>
                <div className="text-sm font-semibold text-foreground mt-2">{t(s.label)}</div>
                <div className="text-xs text-muted-foreground mt-1">{t(s.detail)}</div>
              </div>
            )) : (
              <>
                <div className="text-center"><div className="qc-stat-number font-bold qc-gradient-text">400+</div><div className="text-sm font-semibold mt-2">{t('AI Models')}</div><div className="text-xs text-muted-foreground mt-1">{t('Chat, Code, Vision, Audio')}</div></div>
                <div className="text-center"><div className="qc-stat-number font-bold qc-gradient-text">60+</div><div className="text-sm font-semibold mt-2">{t('Providers')}</div><div className="text-xs text-muted-foreground mt-1">{t('Leading AI providers integrated')}</div></div>
                <div className="text-center"><div className="qc-stat-number font-bold qc-gradient-text">50M+</div><div className="text-sm font-semibold mt-2">{t('Daily Tokens')}</div><div className="text-xs text-muted-foreground mt-1">{t('Processed daily')}</div></div>
                <div className="text-center"><div className="qc-stat-number font-bold qc-gradient-text">99.9%</div><div className="text-sm font-semibold mt-2">{t('Uptime SLA')}</div><div className="text-xs text-muted-foreground mt-1">{t('Enterprise grade')}</div></div>
                <div className="text-center"><div className="qc-stat-number font-bold qc-gradient-text">30K+</div><div className="text-sm font-semibold mt-2">{t('Developers')}</div><div className="text-xs text-muted-foreground mt-1">{t('Active on platform')}</div></div>
              </>
            )}
          </div>
        </div>
      </section>

      {/* ═══ FEATURES ═══ */}
      <section className="qc-fade-up qc-section-pad-lg">
        <div className="qc-wrap">
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full border border-orange-200 bg-orange-50 text-orange-700 text-xs font-semibold mb-5">{t('Core Features')}</div>
            <h2 className="qc-title-section font-bold tracking-tight text-foreground">{t("Quantum Spirit Claw")}, <span className="qc-gradient-text">{t('Every Model')}</span></h2>
            <p className="qc-text-body text-muted-foreground mt-4 max-w-2xl mx-auto leading-relaxed">{t('Access leading AI models through a single unified API endpoint with intelligent routing and real-time analytics.')}</p>
          </div>
          <div className="qc-grid-auto gap-6">
            {(features.length > 0 ? features : [
              { key: 'chat', title: t('Chat & Assistant'), desc: t('Top-tier chat models for every use case, unified in one API.'), icon_name: 'bolt' },
              { key: 'code', title: t('Code Generation'), desc: t('High-performance code models for generation, review, and debugging.'), icon_name: 'cpu' },
              { key: 'reason', title: t('Reasoning & Logic'), desc: t('Advanced reasoning models for complex problem-solving and analysis.'), icon_name: 'layers' },
              { key: 'quantum', title: t('Quantum Computing'), desc: t('Quantum processors accessible via unified API — trapped-ion, superconducting, and more.'), icon_name: 'shield' },
            ]).map((f, i) => (
              <div key={f.key} className={`qc-card-hover rounded-2xl border border-border/50 bg-white p-6 sm:p-8 qc-fade-up qc-fade-up-${i + 1}`}>
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center mb-6 text-white shadow-md shadow-orange-500/10">
                  {f.icon_name === 'bolt' ? icons.bolt : f.icon_name === 'cpu' ? icons.cpu : f.icon_name === 'layers' ? icons.layers : icons.shield}
                </div>
                <h3 className="qc-title-card font-semibold text-foreground mb-3">{t(f.title)}</h3>
                <p className="qc-text-small text-muted-foreground leading-relaxed">{t(f.desc)}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ PROVIDERS ═══ */}
      <section className="qc-fade-up qc-section-pad-lg bg-white border-y border-border/30">
        <div className="qc-wrap">
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full border border-orange-200 bg-orange-50 text-orange-700 text-xs font-semibold mb-5">{t('Ecosystem')}</div>
            <h2 className="qc-title-section font-bold tracking-tight text-foreground">{t('Leading')} <span className="qc-gradient-text">{t('AI Providers')}</span></h2>
            <p className="qc-text-body text-muted-foreground mt-4 max-w-2xl mx-auto leading-relaxed">{t('Browse models from the world\'s leading AI companies and research labs.')}</p>
          </div>
          <div className="qc-grid-auto-sm gap-4">
            {allProviders.map((p, i) => (
              <Link key={p.name} to="/models" search={{ provider: p.name }} className={`qc-card-hover rounded-xl border border-border/50 bg-background p-6 text-center no-underline qc-fade-up qc-fade-up-${i + 1}`}>
                <div className="font-semibold text-foreground group-hover:text-orange-600 transition-colors">{p.name}</div>
                <div className="text-xs text-muted-foreground mt-1.5">{p.models} {t('models')}</div>
                <div className="mt-3 w-10 h-0.5 rounded-full bg-gradient-to-r from-amber-400 to-orange-500 mx-auto opacity-0 group-hover:opacity-100 transition-opacity" />
              </Link>
            ))}
          </div>
          <div className="text-center mt-10">
            <Link to="/models" className="inline-flex items-center gap-2 text-sm font-medium text-muted-foreground hover:text-orange-600 transition-colors no-underline">{t('View All Providers')} {icons.arrowR}</Link>
          </div>
        </div>
      </section>

      {/* ═══ CTA ═══ */}
      <section className="qc-fade-up qc-section-pad-lg">
        <div className="qc-wrap">
          <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-amber-500 via-orange-500 to-rose-500 px-8 sm:px-16 qc-section-pad-sm text-center">
            <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[40vw] h-[40vw] bg-white/10 rounded-full blur-3xl pointer-events-none" />
            <div className="fixed top-1/2 left-1/2 -translate-x-1/3 -translate-y-1/3 w-[30vw] h-[30vw] bg-white/5 rounded-full blur-3xl pointer-events-none" />
            <div className="relative z-10">
              <h2 className="qc-title-section font-bold text-white tracking-tight">{t('Ready to Build?')}</h2>
              <p className="qc-text-body text-white/80 mt-4 inline-block max-w-xl leading-relaxed">{t('A single API key unlocks 400+ models. Start building in minutes, scale to millions.')}</p>
              <div className="flex flex-wrap items-center justify-center gap-4 mt-10">
                <Link to={loggedIn ? '/models' : '/sign-in'} className="inline-flex items-center gap-2 rounded-xl bg-white text-orange-600 px-8 py-3.5 text-base font-semibold hover:bg-white/90 hover:shadow-xl hover:shadow-white/10 transition-all">{loggedIn ? t('Browse Models') : t('Create Free Account')} {icons.arrowR}</Link>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ FOOTER ═══ */}
      <footer className="qc-fade-up border-t border-border/30 bg-white qc-section-pad-sm">
        <div className="qc-wrap">
          <div className="qc-grid-auto gap-10">
            <div className="col-span-2 md:col-span-1">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center"><img src="/logo.webp" alt={t("Quantum Spirit Claw")} className="w-6 h-6 object-contain" /></div>
                <span className="text-base font-bold text-foreground">{t("Quantum Spirit Claw")}</span>
              </div>
              <p className="text-sm text-muted-foreground leading-relaxed max-w-xs">{t('AI API Gateway & Token Distribution Platform')}</p>
            </div>
            {[{ title: t('Platform'), links: [{ to: '/models', label: t('Models') }, { to: '/rankings', label: t('Rankings') }, { to: '/pricing', label: t('Pricing') }, { to: '/download', label: t('Download App') }] },
              { title: t('Resources'), links: [{ to: '/playground', label: t('Playground') }, { to: '/api-docs', label: t('API Docs') }] },
              { title: t('Company'), links: [{ to: '/about', label: t('About') }, { to: '/reseller', label: t('Reseller Program') }] },
            ].map(group => (
              <div key={group.title}>
                <h4 className="text-sm font-semibold text-foreground mb-5">{group.title}</h4>
                <div className="space-y-3.5">{group.links.map(l => (l.external ? <a key={l.to} href={l.to} target="_blank" rel="noopener noreferrer" className="block text-sm text-muted-foreground hover:text-orange-600 transition-colors no-underline">{l.label}</a> : <Link key={l.to} to={l.to} className="block text-sm text-muted-foreground hover:text-orange-600 transition-colors no-underline">{l.label}</Link>))}</div>
              </div>
            ))}
          </div>
          <div className="border-t border-border/30 mt-12 pt-8 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-muted-foreground">
            <div className="flex flex-col sm:flex-row items-center gap-2">
              <a href={companyUrl} target="_blank" rel="noopener noreferrer" className="text-sm text-muted-foreground/40 hover:text-foreground/70 transition-colors">{t('Company Website')}</a>
              <span className="hidden sm:inline text-muted-foreground/20">|</span>
              <span>&copy; {new Date().getFullYear()} {t("Quantum Spirit Claw")}. {t('All rights reserved.')}</span>
            </div>
            <div className="flex items-center gap-6">
              {icpBeian && (
                <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener noreferrer" className="hover:text-foreground cursor-pointer transition-colors text-muted-foreground/60">
                  {icpBeian}
                </a>
              )}
              <span className="hover:text-foreground cursor-pointer transition-colors">{t('Privacy')}</span>
              <span className="hover:text-foreground cursor-pointer transition-colors">{t('Terms')}</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

