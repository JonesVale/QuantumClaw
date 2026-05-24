import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { PromoCarousel } from '@/components/promo-carousel'

export const Route = createFileRoute('/enterprise')({
  component: EnterprisePage,
})

const FEATURES = [
  { icon: '🛡', title: 'SLA 99.99%', desc: 'Enterprise-grade reliability with multi-region failover and guaranteed uptime.' },
  { icon: '🔐', title: 'Private Deployment', desc: 'Deploy on your infrastructure. Complete data isolation and compliance.' },
  { icon: '📊', title: 'Usage Analytics', desc: 'Real-time dashboards with granular usage, cost, and performance metrics.' },
  { icon: '👥', title: 'Team Management', desc: 'RBAC, API key rotation, usage quotas, and audit logs for your team.' },
  { icon: '🔗', title: 'Custom Integrations', desc: 'Connect to your existing stack with custom SDKs and webhook support.' },
  { icon: '💬', title: 'Priority Support', desc: 'Dedicated support team with 15-minute response SLA. 24/7 coverage.' },
]

function EnterprisePage() {
  const { t } = useT()
  return (
    <div className="min-h-screen bg-background"
      style={{ backgroundImage: 'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)' }}>
      <div className="qc-wrap qc-section-pad-sm">
        <div className="mb-6">
          <PromoCarousel pageKey="enterprise" />
        </div>

        <div className="qc-grid-auto-sm gap-6">
          {FEATURES.map((f, i) => (
            <div key={f.title}
              className="qc-fade-up qc-card-hover rounded-2xl bg-white/70 backdrop-blur-sm p-7 border border-border/10"
              style={{ animationDelay: `${i * 0.08}s` }}>
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-xl mb-4 shadow-sm">{f.icon}</div>
              <h3 className="text-lg font-bold tracking-tight text-foreground mb-2">{t(f.title)}</h3>
              <p className="text-sm text-muted-foreground/70 leading-relaxed">{t(f.desc)}</p>
            </div>
          ))}
        </div>

        <div className="qc-fade-up mt-16 text-center rounded-2xl bg-gradient-to-r from-amber-50/80 via-orange-50/80 to-rose-50/80 p-8 border border-amber-200/30">
          <h3 className="text-xl font-bold tracking-tight mb-2">{t('Ready to scale?')}</h3>
          <p className="text-sm text-muted-foreground/70 qc-readable-width mx-auto mb-5">{t('Contact our sales team for a custom enterprise plan')}</p>
          <a href="mailto:sales@quantumclaw.ai"
            className="inline-flex items-center gap-2 px-6 py-2.5 rounded-xl text-sm font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg hover:-translate-y-0.5 transition-all">
            {t('Contact Sales')}
          </a>
        </div>
      </div>
    </div>
  )
}
