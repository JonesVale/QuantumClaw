import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { Code, Mail, Globe, Shield, Server } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export const Route = createFileRoute('/_authenticated/about')({
  component: AboutPage,
})

function AboutPage() {
  const { t } = useT()
  return (
    <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      {/* Header */}
      <div className="text-center py-8">
        <div className="flex items-center justify-center mb-4">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-600 to-purple-600 text-white shadow-lg" />
        </div>
        <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
          QuantumClaw
        </h1>
        <p className="text-muted-foreground mt-2">{t('AI API Gateway & Token Distribution Platform')}</p>
        <p className="text-sm text-muted-foreground mt-1">{t('Version 1.0.0')}</p>
      </div>

      {/* Description */}
      <Card>
        <CardHeader><CardTitle>{t('What is QuantumClaw?')}</CardTitle></CardHeader>
        <CardContent className="text-sm text-muted-foreground space-y-2">
          <p>{t('QuantumClaw is an enterprise-grade AI API gateway that provides unified access to 50+ AI models from providers including OpenAI, Anthropic, Google, DeepSeek, and more.')}</p>
          <p>{t('Key capabilities include token aggregation and distribution, intelligent load balancing, quota management, multi-key rotation, and elastic billing.')}</p>
        </CardContent>
      </Card>

      {/* Features */}
      <Card>
        <CardHeader><CardTitle>{t('Key Features')}</CardTitle></CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {[
              { icon: Shield, title: t('Token Management'), desc: t('Create, manage, and rotate API keys with granular quotas') },
              { icon: Server, title: t('50+ AI Models'), desc: t('Unified access to OpenAI, Claude, Gemini, DeepSeek and more') },
              { icon: Globe, title: t('Load Balancing'), desc: t('Intelligent routing with failover across channels') },
              { icon: Code, title: t('OpenAPI Compatible'), desc: t('Drop-in replacement for any OpenAI-compatible client') },
            ].map((f, i) => (
              <div key={i} className="flex items-start gap-3 p-3 rounded-xl bg-muted/50">
                <f.icon className="h-5 w-5 text-blue-600 shrink-0 mt-0.5" />
                <div>
                  <div className="font-medium text-sm">{f.title}</div>
                  <div className="text-xs text-muted-foreground">{f.desc}</div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Links */}
      <Card>
        <CardHeader><CardTitle>{t('Links')}</CardTitle></CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-4">
            <a href="https://github.com" target="_blank" rel="noopener noreferrer"
              className="flex items-center gap-2 text-sm text-blue-600 hover:underline">
              <Code className="h-4 w-4" /> {t('GitHub Repository')}
            </a>
            <a href="mailto:contact@quantumclaw.ai"
              className="flex items-center gap-2 text-sm text-blue-600 hover:underline">
              <Mail className="h-4 w-4" /> {t('Contact Us')}
            </a>
          </div>
        </CardContent>
      </Card>

      {/* Company Info */}
      <Card>
        <CardHeader>
          <CardTitle>{t('Company Info')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div><span className="font-medium">{t('Company Name')}:</span> 深圳市中科劲纬智能有限公司</div>
          <div><span className="font-medium">{t('Tax ID')}:</span> 91440300MA5GH45W8C</div>
          <div><span className="font-medium">{t('Address')}:</span> 深圳市宝安区石岩街道塘头社区塘头大道33号东海创意园205A</div>
          <div><span className="font-medium">{t('Phone')}:</span> 15920005303</div>
          <div><span className="font-medium">{t('Bank')}:</span> 深圳农村商业银行股份有限公司应人石支行</div>
          <div><span className="font-medium">{t('Bank Account')}:</span> 000396168236</div>
        </CardContent>
      </Card>
    </div>
  )
}
