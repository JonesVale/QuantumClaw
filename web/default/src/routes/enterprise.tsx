import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { Link } from '@tanstack/react-router'
import { useState, useEffect } from 'react'

export const Route = createFileRoute('/enterprise')({
  component: EnterprisePage,
})

const FEATURES = [
  { icon: '馃洝', title: 'SLA 99.99%', desc: 'Enterprise-grade reliability with multi-region failover and guaranteed uptime. Your AI infrastructure never goes dark.' },
  { icon: '馃攼', title: 'Private Deployment', desc: 'Deploy on your infrastructure 鈥?VPC, on-premises, or air-gapped. Complete data isolation for regulated industries.' },
  { icon: '馃搳', title: 'Real-time Analytics', desc: 'Granular dashboards tracking every token, user, and cost center. Alerts, anomaly detection, and export to your BI tools.' },
  { icon: '馃懃', title: 'Team & RBAC', desc: 'Multi-team access control with API key rotation, usage quotas, per-model permissions, and full audit logging.' },
  { icon: '馃敆', title: 'Custom Integration', desc: 'SDKs for every language, webhook event streams, and dedicated middleware. Drop into your existing stack in hours.' },
  { icon: '馃挰', title: 'Priority Support', desc: 'Dedicated engineer with 15-minute response SLA. 24/7 coverage including on-call escalation and incident management.' },
]

const METRICS = [
  { value: '99.99%', labelKey: 'ent_metric_uptime_label', subKey: 'ent_metric_uptime_sub' },
  { value: '400+', labelKey: 'ent_metric_models_label', subKey: 'ent_metric_models_sub' },
  { value: '<15min', labelKey: 'ent_metric_response_label', subKey: 'ent_metric_response_sub' },
  { value: '50M+', labelKey: 'ent_metric_daily_label', subKey: 'ent_metric_daily_sub' },
]

function EnterprisePage() {
  const { t } = useT()
  const [clients, setClients] = useState<{name:string;industry:string;logo:string;description:string;users:string}[]>([])

  useEffect(() => {
    fetch('/api/enterprise-clients').then(r=>r.json()).then(d=>{if(d.success) setClients(d.data)}).catch(()=>{})
  }, [])

  return (
    <div className="min-h-screen bg-background">
      {/* 鈺愨晲鈺?HERO 鈺愨晲鈺?*/}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-amber-50/60 via-white to-orange-50/40" />
        <div className="qc-wrap qc-section-pad-lg">
            <svg className="absolute top-0 left-0 w-full h-full pointer-events-none opacity-[0.04]" viewBox="0 0 1000 1000" preserveAspectRatio="none">
              <defs>
                <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
                  <path d="M 40 0 L 0 0 0 40" fill="none" stroke="currentColor" strokeWidth="0.5"/>
                </pattern>
              </defs>
              <rect width="100%" height="100%" fill="url(#grid)" />
            </svg>
            <svg className="absolute top-5 left-5 w-48 h-48 pointer-events-none opacity-10" viewBox="0 0 200 200" fill="none">
              <circle cx="100" cy="100" r="80" stroke="#f59e0b" strokeWidth="2" strokeDasharray="8 4" fill="none"/>
              <circle cx="100" cy="100" r="50" stroke="#f97316" strokeWidth="2" fill="none"/>
              <circle cx="100" cy="100" r="20" fill="#f59e0b" opacity="0.5"/>
            </svg>
            <svg className="absolute bottom-5 right-5 w-48 h-48 pointer-events-none opacity-10" viewBox="0 0 200 200" fill="none">
              <rect x="20" y="20" width="160" height="160" rx="20" stroke="#f97316" strokeWidth="2" strokeDasharray="10 5"/>
              <rect x="50" y="50" width="100" height="100" rx="10" stroke="#f59e0b" strokeWidth="2" fill="none"/>
              <circle cx="100" cy="100" r="30" fill="#f59e0b" opacity="0.3"/>
            </svg>
            <div className="relative z-10 flex flex-col items-center w-full">
              <div className="max-w-4xl w-full text-center">
            <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full border border-amber-200 bg-amber-50 text-amber-700 text-sm font-semibold tracking-wide mb-8">
              <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
              {t('Enterprise Plan')}
            </div>
            <h1 className="text-4xl sm:text-5xl lg:text-7xl font-bold tracking-tight text-foreground leading-[1.1]">
              {t('AI Infrastructure')}<br />
              <span className="bg-gradient-to-r from-amber-500 via-orange-500 to-rose-500 bg-clip-text text-transparent">{t('Built for Scale')}</span>
            </h1>
            <p className="text-lg sm:text-xl text-muted-foreground mt-8 max-w-2xl mx-auto leading-relaxed">
              {t('Enterprise Hero Desc')}
            </p>
            <div className="flex flex-wrap items-center justify-center gap-4 mt-10">
              <a href="mailto:sales@quantumclaw.ai"
                className="inline-flex items-center gap-2 px-8 py-4 rounded-xl text-lg font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-lg shadow-orange-500/20 hover:shadow-xl hover:-translate-y-0.5 transition-all">
                {t('Contact Sales')}
                <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
              </a>
              <Link to="/models"
                className="inline-flex items-center gap-2 px-8 py-4 rounded-xl text-lg font-medium text-foreground border border-border bg-white/80 hover:bg-white hover:-translate-y-0.5 transition-all">
                {t('Browse Models')}
              </Link>
            </div>
              </div>
            </div>
          </div>
      </section>

      {/* 鈺愨晲鈺?METRICS 鈺愨晲鈺?*/}
      <section className="border-y border-border/30 bg-white/60">
        <div className="qc-wrap qc-section-pad-sm">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-8">
            {METRICS.map(m => (
              <div key={m.value} className="text-center">
                <div className="text-3xl sm:text-4xl font-bold bg-gradient-to-r from-amber-500 to-orange-500 bg-clip-text text-transparent">{m.value}</div>
                <div className="text-base font-semibold text-foreground mt-2">{t(m.labelKey)}</div>
                <div className="text-sm text-muted-foreground/60 mt-1">{t(m.subKey)}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 鈺愨晲鈺?FEATURES 鈺愨晲鈺?*/}
      <section className="qc-section-pad-lg">
        <div className="qc-wrap">
          <div className="text-center mb-16">
            <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground">
              {t('Everything You Need to')}<br />
              <span className="bg-gradient-to-r from-amber-500 to-orange-500 bg-clip-text text-transparent">{t('Scale AI Across')}</span>
            </h2>
            <p className="text-lg text-muted-foreground mt-4 max-w-2xl mx-auto leading-relaxed">
              {t('Enterprise Features Desc')}
            </p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {FEATURES.map((f, i) => (
              <div key={f.title}
                className="qc-fade-up group rounded-2xl bg-white/80 hover:bg-white border border-border/10 hover:border-amber-200/50 hover:shadow-xl hover:shadow-amber-500/5 transition-all duration-300 p-8 text-center"
                style={{ animationDelay: `${i * 0.1}s` }}>
                <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-2xl mb-6 shadow-sm group-hover:scale-110 transition-transform duration-300 mx-auto">{f.icon}</div>
                <h3 className="text-xl font-bold tracking-tight text-foreground mb-3">{t(f.title)}</h3>
                <p className="text-base text-muted-foreground/70 leading-relaxed max-w-sm mx-auto">{t(f.desc)}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 鈺愨晲鈺?TRUSTED BY 鈺愨晲鈺?*/}
      {clients.length > 0 && (
        <section className="qc-section-pad-lg bg-white/40">
          <div className="qc-wrap">
            <div className="text-center mb-12">
              <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground">
                {t('Trusted by')} <span className="bg-gradient-to-r from-amber-500 to-orange-500 bg-clip-text text-transparent">{t('Innovative Teams')}</span>
              </h2>
              <p className="text-lg text-muted-foreground mt-4 max-w-2xl mx-auto leading-relaxed">
                {t('Enterprise Clients Desc')}
              </p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {clients.map((c, i) => (
                <div key={c.name}
                  className="qc-fade-up group rounded-2xl bg-white border border-border/10 hover:border-amber-200/40 hover:shadow-lg transition-all duration-300 p-6 text-center"
                  style={{ animationDelay: `${i * 0.08}s` }}>
                  <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-sm font-bold text-orange-700 mb-4 shadow-sm mx-auto">
                    {c.logo}
                  </div>
                  <h3 className="text-base font-bold text-foreground mb-1">{c.name}</h3>
                  <span className="text-xs font-medium text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full inline-block">{c.industry}</span>
                  <p className="text-sm text-muted-foreground/70 mt-3 leading-relaxed max-w-xs mx-auto">{c.description}</p>
                  <div className="flex items-center justify-center gap-1.5 mt-4 text-xs text-muted-foreground/50">
                    <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                    {c.users} {t('users')}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* 鈺愨晲鈺?CTA 鈺愨晲鈺?*/}
      <section className="qc-section-pad-lg bg-gradient-to-br from-amber-500 via-orange-500 to-rose-500">
        <div className="qc-wrap text-center relative z-10">
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[40vw] h-[40vw] bg-white/10 rounded-full blur-3xl pointer-events-none" />
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-white tracking-tight leading-[1.1]">
            {t('Ready to Build at Scale?')}
          </h2>
          <p className="text-lg sm:text-xl text-white/80 mt-6 max-w-2xl mx-auto leading-relaxed">
            {t('Enterprise CTA Desc')}
          </p>
          <div className="flex flex-wrap justify-center gap-4 mt-10">
            <a href="mailto:sales@quantumclaw.ai"
              className="inline-flex items-center gap-2 px-8 py-4 rounded-xl text-lg font-semibold text-orange-600 bg-white hover:bg-white/90 hover:shadow-xl hover:-translate-y-0.5 transition-all">
              {t('Talk to Sales')}
              <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
            </a>
            <Link to="/sign-in"
              className="inline-flex items-center gap-2 px-8 py-4 rounded-xl text-lg font-semibold text-white border-2 border-white/30 hover:bg-white/10 hover:-translate-y-0.5 transition-all">
              {t('Start Free')}
            </Link>
          </div>
        </div>
      </section>

      {/* 鈺愨晲鈺?FOOTER 鈺愨晲鈺?*/}
      <footer className="bg-white border-t border-border/30 qc-section-pad-sm">
        <div className="qc-wrap">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center">
                <img src="/logo.webp" alt={t("Quantum Spirit Claw")} className="w-6 h-6 object-contain" />
              </div>
              <span className="font-bold text-foreground">{t("Quantum Spirit Claw Enterprise")}</span>
            </div>
            <div className="flex items-center gap-6">
              <a href="mailto:sales@quantumclaw.ai" className="hover:text-foreground transition-colors">sales@quantumclaw.ai</a>
              <span className="hover:text-foreground cursor-pointer transition-colors">{t('Privacy')}</span>
              <span className="hover:text-foreground cursor-pointer transition-colors">{t('Terms')}</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

