import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useEffect } from 'react'
import { useAuthStore } from '@/stores/auth-store'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import apiClient from '@/lib/api'
import { CheckCircle, ArrowRight, Sparkles, Code, MessageSquare, Gift } from 'lucide-react'

export const Route = createFileRoute('/_authenticated/welcome')({
  component: WelcomePage,
})

function WelcomePage() {
  const { t } = useT()
  const { auth } = useAuthStore()
  const navigate = useNavigate()
  const [step, setStep] = useState(1)
  const [quotaAdded, setQuotaAdded] = useState(false)

  useEffect(() => {
    // On mount, give the new user their signup bonus
    if (!quotaAdded) {
      apiClient.post('/api/user/self/checkin').catch(() => {})
      setQuotaAdded(true)
    }
  }, [quotaAdded])

  const steps = [
    {
      num: 1,
      title: t('onboard_step1_title') || '选择一个模型',
      desc: t('onboard_step1_desc') || '浏览 400+ AI 模型，找到适合你场景的模型',
      icon: <Sparkles className="h-8 w-8" />,
      action: (
        <Link to="/models">
          <Button className="gap-2"><Sparkles className="h-4 w-4" />{t('Browse Models')} <ArrowRight className="h-4 w-4" /></Button>
        </Link>
      ),
    },
    {
      num: 2,
      title: t('onboard_step2_title') || '复制代码到你的项目',
      desc: t('onboard_step2_desc') || '支持 Python、Node.js、Go、cURL — OpenAI 兼容 SDK，一行代码接入',
      icon: <Code className="h-8 w-8" />,
      action: (
        <Link to="/developer">
          <Button className="gap-2"><Code className="h-4 w-4" />{t('View SDK Docs')} <ArrowRight className="h-4 w-4" /></Button>
        </Link>
      ),
    },
    {
      num: 3,
      title: t('onboard_step3_title') || '在线体验效果',
      desc: t('onboard_step3_desc') || '直接在网页上测试模型效果，无需写代码',
      icon: <MessageSquare className="h-8 w-8" />,
      action: (
        <Link to="/chat">
          <Button className="gap-2"><MessageSquare className="h-4 w-4" />{t('Try Online Chat')} <ArrowRight className="h-4 w-4" /></Button>
        </Link>
      ),
    },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-b from-amber-50/30 to-white">
      <div className="max-w-2xl mx-auto px-4 py-12 space-y-8">
        {/* Welcome header */}
        <div className="text-center space-y-4">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-green-50 text-green-700 text-sm font-medium border border-green-200">
            <Gift className="h-4 w-4" />
            {t('onboard_bonus') || '恭喜！你已获得 500,000 体验配额'}
          </div>
          <h1 className="text-3xl font-bold tracking-tight">
            {t('onboard_title') || '欢迎加入 QuantumClaw！'}
          </h1>
          <p className="text-muted-foreground max-w-md mx-auto">
            {t('onboard_subtitle') || '完成以下 3 步，快速上手 AI API 调用'}
          </p>
        </div>

        {/* Progress indicator */}
        <div className="flex items-center justify-center gap-2">
          {[1, 2, 3].map((s) => (
            <div key={s} className="flex items-center gap-2">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-all ${
                step >= s
                  ? 'bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-md'
                  : 'bg-muted text-muted-foreground'
              }`}>
                {step > s ? <CheckCircle className="h-4 w-4" /> : s}
              </div>
              {s < 3 && <div className={`w-12 h-1 rounded-full ${step > s ? 'bg-orange-500' : 'bg-muted'}`} />}
            </div>
          ))}
        </div>

        {/* Step cards */}
        <div className="space-y-4">
          {steps.map((s) => (
            <Card
              key={s.num}
              className={`transition-all cursor-pointer ${
                step === s.num
                  ? 'border-orange-300 shadow-lg shadow-orange-100 ring-1 ring-orange-200'
                  : step > s.num
                    ? 'opacity-60'
                    : 'opacity-40'
              }`}
              onClick={() => step >= s.num && setStep(s.num)}
            >
              <CardContent className="p-6">
                <div className="flex items-start gap-4">
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${
                    step >= s.num
                      ? 'bg-gradient-to-br from-amber-400 to-orange-500 text-white'
                      : 'bg-muted text-muted-foreground'
                  }`}>
                    {step > s.num ? <CheckCircle className="h-6 w-6" /> : s.icon}
                  </div>
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-lg">{s.title}</span>
                      {step === s.num && (
                        <Badge className="bg-orange-100 text-orange-700 border-orange-200">{t('onboard_current') || '当前'}</Badge>
                      )}
                      {step > s.num && (
                        <Badge className="bg-green-100 text-green-700 border-green-200">{t('onboard_done') || '已完成'}</Badge>
                      )}
                    </div>
                    <p className="text-sm text-muted-foreground">{s.desc}</p>
                    {step === s.num && (
                      <div className="pt-3">
                        {s.action}
                      </div>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Navigation */}
        <div className="flex items-center justify-center gap-4">
          {step > 1 && (
            <Button variant="outline" onClick={() => setStep(s => s - 1)}>
              {t('onboard_prev') || '上一步'}
            </Button>
          )}
          {step < 3 ? (
            <Button onClick={() => setStep(s => s + 1)}>
              {t('onboard_next') || '下一步'} <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          ) : (
            <Link to="/dashboard">
              <Button className="bg-gradient-to-r from-amber-500 to-orange-500 text-white">
                {t('onboard_start') || '开始使用'} <ArrowRight className="h-4 w-4 ml-2" />
              </Button>
            </Link>
          )}
        </div>

        {/* Skip */}
        <div className="text-center">
          <Link to="/dashboard" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
            {t('onboard_skip') || '跳过引导，直接进入控制台'}
          </Link>
        </div>
      </div>
    </div>
  )
}
