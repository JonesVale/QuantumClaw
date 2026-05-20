import { createFileRoute, Link } from '@tanstack/react-router'
import { Code, Mail, Globe, Shield, Server } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export const Route = createFileRoute('/_authenticated/about')({
  component: AboutPage,
})

function AboutPage() {
  return (
    <div className=" w-full p-4 sm:p-6 space-y-6">
      {/* Header */}
      <div className="text-center py-8">
        <div className="flex items-center justify-center mb-4">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-600 to-purple-600 text-white shadow-lg">
            
          </div>
        </div>
        <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
          QuantumClaw
        </h1>
        <p className="text-muted-foreground mt-2">AI API Gateway & Quantum Computing Distribution Platform</p>
        <p className="text-sm text-muted-foreground mt-1">Version 1.0.0</p>
      </div>

      {/* Description */}
      <Card>
        <CardHeader><CardTitle>What is QuantumClaw?</CardTitle></CardHeader>
        <CardContent className="text-sm text-muted-foreground space-y-2">
          <p>QuantumClaw is an enterprise-grade AI API gateway and quantum computing distribution platform that provides unified access to 50+ AI models and 6 quantum computing platforms.</p>
          <p>Key capabilities include token aggregation and distribution, intelligent load balancing, quota management, multi-key rotation, elastic billing, and quantum computing resource scheduling across IonQ, IBM Q, Rigetti, AWS Braket, and Azure Quantum.</p>
        </CardContent>
      </Card>

      {/* Features */}
      <Card>
        <CardHeader><CardTitle>Key Features</CardTitle></CardHeader>
        <CardContent>
          <div className="grid md:grid-cols-2 gap-3">
            {[
              { icon: Shield, title: 'Token Management', desc: 'Create, manage, and rotate API keys with granular quotas' },
              { icon: Server, title: '50+ AI Models', desc: 'Unified access to OpenAI, Claude, Gemini, DeepSeek and more' },
              { icon: Globe, title: 'Quantum Computing', desc: 'IonQ, IBM Q, Rigetti, AWS Braket multi-platform aggregation' },
              { icon: Code, title: 'OpenAPI Compatible', desc: 'Drop-in replacement for any OpenAI-compatible client' },
            ].map((f, i) => (
              <div key={i} className="flex items-start gap-3 p-3 rounded-lg bg-muted/50">
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
        <CardHeader><CardTitle>Links</CardTitle></CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-4">
            <a href="https://github.com" target="_blank" rel="noopener noreferrer"
              className="flex items-center gap-2 text-sm text-blue-600 hover:underline">
              <Code className="h-4 w-4" /> GitHub Repository
            </a>
            <a href="mailto:contact@quantumclaw.ai"
              className="flex items-center gap-2 text-sm text-blue-600 hover:underline">
              <Mail className="h-4 w-4" /> Contact Us
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

