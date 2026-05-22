import { createFileRoute, Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import {
  Shield, Lock, Key, Users, Server, Database,
  CheckCircle, ArrowRight, Globe, CreditCard, FileCheck,
  Building2, Network, Sparkles, Cpu
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/enterprise')({
  component: EnterprisePage,
})

function EnterprisePage() {
  const { t } = useTranslation()

  const features = [
    { icon: Shield, title: t('SSO & SAML'), desc: t('Single sign-on with SAML, OIDC, and OAuth providers. Support for Azure AD, Okta, Google Workspace.') },
    { icon: Lock, title: t('Data Residency'), desc: t('Choose which regions your data is processed in. Compliant with GDPR, SOC 2, and HIPAA.') },
    { icon: Key, title: t('Custom Data Policies'), desc: t('Granular controls over which providers handle your data. Block specific models or regions.') },
    { icon: Users, title: t('Team Management'), desc: t('Organize users into teams with role-based access control. Audit logs for all API activity.') },
    { icon: Server, title: t('Dedicated Infrastructure'), desc: t('Private API endpoints with guaranteed throughput. SLAs starting at 99.9% uptime.') },
    { icon: Database, title: t('Usage Analytics'), desc: t('Detailed cost breakdowns by team, model, and provider. Export to your BI tools.') },
  ]

  const plans = [
    {
      name: t('Team'),
      price: '$499',
      period: '/month',
      features: [t('Up to 50 team members'), t('SSO integration'), t('Audit logs'), t('99.9% uptime SLA'), t('Priority support'), t('Custom data policies')],
      cta: t('Contact Sales'),
      highlighted: false,
    },
    {
      name: t('Business'),
      price: '$1,999',
      period: '/month',
      features: [t('Unlimited team members'), t('SSO + SCIM'), t('Advanced audit logs'), t('99.95% uptime SLA'), t('Dedicated account manager'), t('Custom data policies'), t('Data residency controls'), t('API rate limit customization')],
      cta: t('Contact Sales'),
      highlighted: true,
    },
    {
      name: t('Enterprise'),
      price: t('Custom'),
      period: '',
      features: [t('Everything in Business'), t('Dedicated infrastructure'), t('Custom contract terms'), t('On-premise deployment option'), t('24/7 premium support'), t('Custom model routing')],
      cta: t('Contact Us'),
      highlighted: false,
    },
  ]

  return (
    <div className="p-4 sm:p-6 space-y-8 w-full min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Hero */}
      <div className="text-center max-w-3xl mx-auto py-8 sm:py-12">
        <Badge variant="outline" className="mb-4 px-4 py-1.5 gap-2 text-sm">
          <Building2 className="h-4 w-4" />
          {t('Enterprise')}
        </Badge>
        <h1 className="text-3xl sm:text-5xl font-bold mb-4 bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
          {t('AI Infrastructure for Organizations')}
        </h1>
        <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
          {t('Secure, scalable, and compliant AI API gateway for teams and enterprises.')}
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 max-w-4xl mx-auto">
        {[
          { value: '99.9%', label: t('Uptime SLA') },
          { value: '67+', label: t('Providers') },
          { value: '400+', label: t('Models') },
          { value: '7', label: t('Languages') },
        ].map((s, i) => (
          <Card key={i} className="text-center">
            <CardContent className="p-4">
              <p className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">{s.value}</p>
              <p className="text-xs text-muted-foreground mt-1">{s.label}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Features */}
      <div className="max-w-5xl mx-auto">
        <h2 className="text-2xl font-bold text-center mb-8">{t('Enterprise Features')}</h2>
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {features.map((f, i) => (
            <Card key={i} className="group hover:shadow-xl transition-all duration-300 hover:-translate-y-0.5 rounded-xl">
              <CardContent className="p-5">
                <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mb-3 shadow">
                  <f.icon className="h-5 w-5 text-white" />
                </div>
                <h3 className="font-semibold mb-1">{f.title}</h3>
                <p className="text-sm text-muted-foreground">{f.desc}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Unique Advantage */}
      <Card className="max-w-4xl mx-auto bg-gradient-to-r from-purple-500/10 via-blue-500/10 to-cyan-500/10 border-purple-500/20">
        <CardContent className="p-6 flex items-center gap-4">
          <Cpu className="h-10 w-10 text-purple-500 shrink-0" />
          <div>
            <h3 className="font-semibold text-lg">{t('Exclusive: Quantum Computing Access')}</h3>
            <p className="text-sm text-muted-foreground">{t('Only QuantumClaw offers enterprise-grade quantum computing access alongside 400+ AI models. No other platform combines classical AI and quantum resources in a single API.')}</p>
          </div>
        </CardContent>
      </Card>

      {/* Pricing */}
      <div className="max-w-5xl mx-auto">
        <h2 className="text-2xl font-bold text-center mb-8">{t('Plans')}</h2>
        <div className="grid sm:grid-cols-3 gap-4">
          {plans.map((plan, i) => (
            <Card key={i} className={cn('relative rounded-xl', plan.highlighted && 'ring-2 ring-primary shadow-xl scale-105')}>
              {plan.highlighted && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <Badge className="bg-gradient-to-r from-blue-600 to-purple-600">{t('Most Popular')}</Badge>
                </div>
              )}
              <CardContent className="p-6 space-y-4">
                <div>
                  <h3 className="text-lg font-semibold">{plan.name}</h3>
                  <div className="mt-2">
                    <span className="text-3xl font-bold">{plan.price}</span>
                    <span className="text-muted-foreground">{plan.period}</span>
                  </div>
                </div>
                <ul className="space-y-2">
                  {plan.features.map((f, j) => (
                    <li key={j} className="flex items-start gap-2 text-sm">
                      <CheckCircle className="h-4 w-4 text-emerald-500 mt-0.5 shrink-0" />
                      {f}
                    </li>
                  ))}
                </ul>
                <Button className="w-full" variant={plan.highlighted ? 'default' : 'outline'}>
                  {plan.cta}
                  <ArrowRight className="h-4 w-4 ml-2" />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* CTA */}
      <div className="text-center py-8">
        <h2 className="text-2xl font-bold mb-3">{t('Ready to scale?')}</h2>
        <p className="text-muted-foreground mb-6">{t('Talk to our team about your enterprise needs')}</p>
        <div className="flex items-center justify-center gap-4">
          <Link to="/dashboard"><Button size="lg" className="gap-2">{t('Get Started')}<ArrowRight className="h-4 w-4" /></Button></Link>
          <Button variant="outline" size="lg" className="gap-2">
            <Globe className="h-4 w-4" />{t('Schedule Demo')}
          </Button>
        </div>
      </div>
    </div>
  )
}
