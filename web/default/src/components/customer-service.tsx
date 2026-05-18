import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { MessageCircle, X, MessageSquare, Phone, ExternalLink, ChevronDown, Bot, User } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'
import { Link } from '@tanstack/react-router'

const QQ1 = '634165717'
const QQ2 = '108768250'
const QQ3 = '587600277'
const WECHAT = '634165717'

export function CustomerServiceFloating() {
  const [open, setOpen] = useState(false)
  const [expanded, setExpanded] = useState<string | null>(null)
  const { t } = useTranslation()
  const { loggedIn } = useAuthStore()

  const faqItems = [
    {
      key: 'how-to-get-key',
      question: t('How to get an API Key?'),
      answer: t('After logging in, go to the API Keys page and click Create Key.'),
    },
    {
      key: 'how-to-topup',
      question: t('How to top up?'),
      answer: t('Go to Wallet page to choose a payment method (Stripe, Epay, Binance, etc.).'),
    },
    {
      key: 'which-models',
      question: t('Which models are supported?'),
      answer: t('50+ models including GPT-4o, Claude 3.5, Gemini 2.0, DeepSeek V3, Qwen3.6 and more.'),
    },
  ]

  return (
    <>
      {/* Floating button */}
      <div className="fixed bottom-6 right-6 z-50 flex flex-col items-end gap-3">
        {/* Chat panel */}
        {open && (
          <div className="mb-2 w-80 sm:w-96 rounded-2xl border bg-background shadow-2xl overflow-hidden animate-in slide-in-from-bottom-5 duration-200">
            {/* Header */}
            <div className="bg-gradient-to-r from-blue-600 to-purple-600 p-4 text-white">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Bot className="h-5 w-5" />
                  <span className="font-semibold">{t('QuantumClaw Assistant')}</span>
                </div>
                <button onClick={() => setOpen(false)} className="rounded-full p-1 hover:bg-white/20 transition-colors">
                  <X className="h-4 w-4" />
                </button>
              </div>
              <p className="mt-1 text-xs text-white/80">{t('How can we help you today?')}</p>
            </div>

            {/* FAQ */}
            <div className="max-h-80 overflow-y-auto p-4 space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{t('Quick Help')}</p>
              {faqItems.map((item) => (
                <div key={item.key} className="rounded-lg border bg-card">
                  <button
                    onClick={() => setExpanded(expanded === item.key ? null : item.key)}
                    className="flex w-full items-center justify-between p-3 text-left text-sm font-medium hover:bg-muted/50 transition-colors rounded-lg"
                  >
                    {item.question}
                    <ChevronDown className={cn('h-4 w-4 shrink-0 transition-transform', expanded === item.key && 'rotate-180')} />
                  </button>
                  {expanded === item.key && (
                    <div className="border-t px-3 py-2 text-sm text-muted-foreground bg-muted/30 rounded-b-lg">
                      {item.answer}
                      {item.key === 'how-to-get-key' && loggedIn && (
                        <Link to="/keys" className="mt-1 block text-blue-600 hover:underline" onClick={() => setOpen(false)}>
                          {t('Go to API Keys')} &rarr;
                        </Link>
                      )}
                      {item.key === 'how-to-topup' && loggedIn && (
                        <Link to="/wallet" className="mt-1 block text-blue-600 hover:underline" onClick={() => setOpen(false)}>
                          {t('Go to Wallet')} &rarr;
                        </Link>
                      )}
                    </div>
                  )}
                </div>
              ))}

              {/* Contact human */}
              <div className="mt-4 rounded-lg border bg-gradient-to-br from-blue-50 to-purple-50 dark:from-blue-950/30 dark:to-purple-950/30 p-4">
                <div className="flex items-center gap-2 mb-3">
                  <User className="h-4 w-4 text-blue-600" />
                  <span className="text-sm font-semibold">{t('Contact Customer Service')}</span>
                </div>
                <div className="space-y-2 text-sm">
                  <div className="flex items-center gap-2">
                    <MessageSquare className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                    <span>
                      <span className="font-medium">QQ:</span>{' '}
                      <a href={`tencent://message/?uin=${QQ1}&Site=&Menu=yes`} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                        {QQ1}
                      </a>
                      {' / '}
                      <a href={`tencent://message/?uin=${QQ2}&Site=&Menu=yes`} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                        {QQ2}
                      </a>
                      {' / '}
                      <a onClick={() => { navigator.clipboard.writeText(QQ3); alert('QQ群号已复制: ' + QQ3); }}
                         className="text-blue-600 hover:underline cursor-pointer">
                        群{QQ3}
                      </a>
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <MessageCircle className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                    <span>
                      <span className="font-medium">WeChat:</span>{' '}
                      <span className="text-blue-600">{WECHAT}</span>
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Toggle button */}
        <Button
          size="icon"
          className={cn(
            'h-12 w-12 rounded-full shadow-lg transition-all duration-200',
            open ? 'bg-destructive hover:bg-destructive/90 rotate-90' : 'bg-gradient-to-r from-blue-600 to-purple-600 hover:shadow-xl'
          )}
          onClick={() => setOpen(!open)}
        >
          {open ? <X className="h-5 w-5" /> : <MessageCircle className="h-5 w-5" />}
        </Button>
      </div>
    </>
  )
}
