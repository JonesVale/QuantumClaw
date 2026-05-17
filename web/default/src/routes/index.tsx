import { createFileRoute, Link } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import {
  ArrowRight,
  Brain, MessageSquare, Globe, Code,
  Mail,
  DollarSign, KeyRound, BarChart3, Network,
  Menu, X, Shield, Activity,
  ChevronDown, Check
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAuthStore } from '@/stores/auth-store'

export const Route = createFileRoute('/')({
  component: HomePage,
})

const LANGUAGES = [
  { code: 'zh-CN', name: '简体中文' },
  { code: 'zh-TW', name: '繁體中文' },
  { code: 'en', name: 'English' },
  { code: 'fr', name: 'Français' },
  { code: 'ja', name: '日本語' },
  { code: 'ru', name: 'Русский' },
  { code: 'vi', name: 'Tiếng Việt' },
]



function ImageIcon(p: any) { return <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...p}><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg> }
function VideoIcon(p: any) { return <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...p}><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg> }
function PenIcon(p: any) { return <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...p}><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/><path d="m15 5 4 4"/></svg> }

function HomePage() {
  const { i18n, t } = useTranslation()
  const [menuOpen, setMenuOpen] = useState(false)
  const [currentLang, setCurrentLang] = useState(i18n.language || 'en')
  const { loggedIn } = useAuthStore()

  const getNavItems = (loggedIn: boolean) => [
    { to: '/', label: t('Home') },
    { to: loggedIn ? '/models' : '/sign-in', label: t('Models') },
    { to: loggedIn ? '/playground' : '/sign-in', label: t('Playground') },
    { to: '/sign-in', label: t('Console') },
  ]

  const features = [
    { icon: KeyRound, title: t('API Key Management'), desc: t('Unified create, quota, permission, rotation'), grad: 'from-blue-500 to-cyan-500' },
    { icon: Network, title: t('Smart Channel Distribution'), desc: t('50+ model load balancing and failover'), grad: 'from-purple-500 to-pink-500' },
    { icon: BarChart3, title: t('Usage Monitoring'), desc: t('Real-time quota management, consumption details'), grad: 'from-green-500 to-emerald-500' },
    { icon: Shield, title: t('Security Protection'), desc: t('SSRF protection, IP whitelist, audit logs'), grad: 'from-orange-500 to-red-500' },
    { icon: DollarSign, title: t('Flexible Billing'), desc: t('Pay-as-you-go / Subscription / Rate control'), grad: 'from-indigo-500 to-blue-500' },
    { icon: Activity, title: t('Async Tasks'), desc: t('MJ/Video/Music task scheduling'), grad: 'from-pink-500 to-rose-500' },
  ]

  const heroCards = [
    { title: t('Qwen3.6'), desc: t('Native multimodal, more stable reasoning'), link: 'https://chat.deepseek.com' },
    { title: t('Claude 3.5'), desc: t('Anthropic Strongest Reasoning'), link: 'https://claude.ai' },
    { title: t('GPT-4o'), desc: t('OpenAI Flagship Multimodal'), link: 'https://chat.openai.com' },
    { title: t('Gemini 2.0'), desc: t('Google Fast Reasoning'), link: 'https://gemini.google.com' },
  ]

  const modelApps = [
    { name: t('Chat & Reasoning'), icon: MessageSquare, models: ['GPT-4o', 'Claude 3.5', 'Gemini 2.0', 'DeepSeek V3'], link: 'https://chat.openai.com', grad: 'from-blue-500 to-cyan-500' },
    { name: t('Image Generation'), icon: ImageIcon, models: ['DALL-E 3', 'Midjourney', 'Stable Diffusion'], link: 'https://www.midjourney.com', grad: 'from-purple-500 to-pink-500' },
    { name: t('Programming & Development'), icon: Code, models: ['GPT-4o', 'Claude 3.5', 'DeepSeek Coder'], link: 'https://github.com/features/copilot', grad: 'from-orange-500 to-red-500' },
    { name: t('Audio & Video'), icon: VideoIcon, models: ['Suno', 'Runway', 'ElevenLabs'], link: 'https://suno.com', grad: 'from-pink-500 to-rose-500' },
    { name: t('Text & Writing'), icon: PenIcon, models: ['GPT-4o', 'Claude 3', 'Kimi', 'Moonshot'], link: 'https://www.grammarly.com', grad: 'from-green-500 to-emerald-500' },
    { name: t('Search & Knowledge'), icon: Brain, models: ['Perplexity', 'ArXiv', 'Hugging Face'], link: 'https://www.perplexity.ai', grad: 'from-indigo-500 to-blue-500' },
  ]

  const newsSources = [
    { name: t('Machine Heart'), url: 'https://www.jiqizhixin.com', lang: t('Chinese') },
    { name: t('QbitAI'), url: 'https://www.qbitai.com', lang: t('Chinese') },
    { name: t('36Kr AI'), url: 'https://36kr.com/information/AI/', lang: t('Chinese') },
    { name: t('Leiphone'), url: 'https://www.leiphone.com/category/ai', lang: t('Chinese') },
    { name: t('OpenAI Blog'), url: 'https://openai.com/blog', lang: t('EN') },
    { name: t('Anthropic'), url: 'https://www.anthropic.com/blog', lang: t('EN') },
    { name: t('Google AI'), url: 'https://ai.googleblog.com', lang: t('EN') },
    { name: t('MIT Tech Review'), url: 'https://www.technologyreview.com/topic/artificial-intelligence/', lang: t('EN') },
    { name: t('ArXiv'), url: 'https://arxiv.org', lang: t('EN') },
    { name: t('Hugging Face'), url: 'https://huggingface.co/blog', lang: t('EN') },
    { name: t('Reddit AI'), url: 'https://www.reddit.com/r/artificial/', lang: t('EN') },
  ]

  const stats = [
    { value: '99.9%', label: t('Service Availability') },
    { value: '<50ms', label: t('Average Latency') },
    { value: '50+', label: t('AI Models') },
    { value: '10M+', label: t('Monthly Requests') },
  ]

  useEffect(() => { setCurrentLang(i18n.language) }, [i18n.language])

  const changeLang = (code: string) => {
    i18n.changeLanguage(code)
    localStorage.setItem('i18nextLng', code)
    setCurrentLang(code)
  }
  const langName = LANGUAGES.find(l => l.code === currentLang)?.name || 'English'

  return (
    <>
      <nav className="bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800" style={{ height: 'clamp(44px, 7vh, 56px)' }}>
        <div className="mx-auto flex items-center h-full justify-between" style={{ maxWidth: 'min(92vw, 1400px)', padding: '0 clamp(8px, 3vw, 32px)' }}>
          <Link to="/" className="flex items-center gap-x-1 shrink-0">
            <img src="/logo.webp" alt="" className="rounded-lg object-cover" style={{ width: 'clamp(24px, 3vw, 32px)', height: 'clamp(24px, 3vw, 32px)' }} />
            <span className="font-bold hidden sm:inline" style={{ fontSize: 'clamp(12px, 1.2vw, 14px)' }}>{t('QuantumClaw')}</span>
          </Link>

          <div className="hidden md:flex items-center justify-center flex-1 gap-x-0.5" style={{ fontSize: 'clamp(11px, 1vw, 14px)' }}>
            {getNavItems(loggedIn).map((item) => (
              <Link key={item.to} to={item.to}
                className="font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition-all whitespace-nowrap rounded-md"
                style={{ padding: 'clamp(4px, 0.6vw, 8px) clamp(6px, 1vw, 12px)' }}>
                {t(item.label)}
              </Link>
            ))}
          </div>

          <div className="flex items-center gap-x-1 shrink-0">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="hidden sm:inline-flex items-center gap-1 rounded-md border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors whitespace-nowrap"
                  style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(4px, 0.6vw, 8px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                  <Globe style={{ width: 'clamp(10px, 0.8vw, 14px)', height: 'clamp(10px, 0.8vw, 14px)' }} />
                  <span className="hidden lg:inline">{langName}</span>
                  <ChevronDown style={{ width: 'clamp(8px, 0.7vw, 12px)', height: 'clamp(8px, 0.7vw, 12px)' }} />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" style={{ minWidth: '140px' }}>
                {LANGUAGES.map((lang) => (
                  <DropdownMenuItem key={lang.code} onClick={() => changeLang(lang.code)}
                    className={currentLang === lang.code ? 'font-semibold bg-slate-50 dark:bg-slate-800' : ''}
                    style={{ fontSize: 'clamp(11px, 0.9vw, 13px)' }}>
                    <Globe className="mr-2 shrink-0" style={{ width: 'clamp(10px, 0.8vw, 14px)', height: 'clamp(10px, 0.8vw, 14px)' }} />
                    {lang.name}
                    {currentLang === lang.code && <Check className="ml-auto" style={{ width: 'clamp(10px, 0.8vw, 14px)', height: 'clamp(10px, 0.8vw, 14px)' }} />}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>

            <Link to="/sign-in"
              className="inline-flex items-center justify-center rounded-md bg-gradient-to-r from-blue-600 to-purple-600 text-white font-semibold hover:shadow-md transition-all whitespace-nowrap"
              style={{ padding: 'clamp(4px, 0.5vw, 8px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(11px, 0.9vw, 13px)' }}>
              {t('登录 / 注册')}
            </Link>

            <button onClick={() => setMenuOpen(!menuOpen)}
              className="md:hidden inline-flex items-center justify-center rounded-md hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              style={{ width: 'clamp(24px, 3vw, 32px)', height: 'clamp(24px, 3vw, 32px)' }}>
              {menuOpen ? <X style={{ width: 'clamp(14px, 1.5vw, 18px)', height: 'clamp(14px, 1.5vw, 18px)' }} /> : <Menu style={{ width: 'clamp(14px, 1.5vw, 18px)', height: 'clamp(14px, 1.5vw, 18px)' }} />}
            </button>
          </div>
        </div>
        {menuOpen && (
          <div className="md:hidden border-t border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 shadow-lg">
            <div className="py-1 px-4 space-y-0.5">
              {getNavItems(loggedIn).map((item) => (
                <Link key={item.to} to={item.to} onClick={() => setMenuOpen(false)}
                  className="block px-3 py-2 rounded-md text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
                  {t(item.label)}
                </Link>
              ))}
              <div className="border-t border-slate-100 dark:border-slate-800 pt-2 mt-2">
                <p className="px-3 pb-1 text-[10px] text-slate-400 uppercase tracking-wider">{t('Language')}</p>
                {LANGUAGES.map((lang) => (
                  <button key={lang.code} onClick={() => { changeLang(lang.code); setMenuOpen(false) }}
                    className={'w-full text-left px-3 py-1.5 rounded-md text-sm ' + (currentLang === lang.code ? 'font-semibold bg-slate-100 dark:bg-slate-800' : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800')}>
                    {lang.name}
                  </button>
                ))}
              </div>
              <Link to="/sign-in" onClick={() => setMenuOpen(false)}
                className="block px-3 py-2 rounded-md text-sm font-semibold text-white bg-gradient-to-r from-blue-600 to-purple-600 mt-2">
                {t('登录 / 注册')}
              </Link>
            </div>
          </div>
        )}
      </nav>

      <main className="bg-white dark:bg-slate-950">
        <div className="mx-auto" style={{ maxWidth: 'min(92vw, 1400px)', padding: '0 clamp(8px, 3vw, 32px)' }}>

          {/* HERO */}
          <section style={{ paddingTop: 'clamp(24px, 5vw, 64px)', paddingBottom: 'clamp(24px, 5vw, 64px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(16px, 3vw, 48px)' }}>
              <h1 className="font-extrabold tracking-tight leading-[1.15]" style={{ fontSize: 'clamp(28px, 5vw, 72px)', marginBottom: 'clamp(8px, 1.5vw, 16px)' }}>
                <span className="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">{t('QuantumClaw')}</span>
              </h1>
              <p className="text-slate-500 dark:text-slate-400 mx-auto" style={{ fontSize: 'clamp(13px, 1.5vw, 20px)', maxWidth: 'clamp(280px, 60vw, 600px)' }}>
                <span className="font-semibold text-slate-700 dark:text-slate-300">{t('Key聚合分发')}</span> {t('·')} {t('一个Key聚合调用 50+ 主流 AI 模型')}
              </p>
              <div className="flex flex-wrap gap-y-2 gap-x-3 justify-center" style={{ marginTop: 'clamp(16px, 3vw, 32px)' }}>
                <Link to="/sign-in"
                  className="inline-flex items-center justify-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 text-white font-bold shadow-md hover:shadow-lg transition-all hover:-translate-y-0.5"
                  style={{ padding: 'clamp(8px, 1vw, 16px) clamp(16px, 2.5vw, 28px)', fontSize: 'clamp(12px, 1.1vw, 16px)' }}>
                  <KeyRound style={{ width: 'clamp(14px, 1.3vw, 20px)', height: 'clamp(14px, 1.3vw, 20px)' }} />
                  {t('免费获取 Token')}
                  <ArrowRight style={{ width: 'clamp(12px, 1vw, 16px)', height: 'clamp(12px, 1vw, 16px)' }} />
                </Link>
                <Link to="/playground"
                  className="inline-flex items-center justify-center gap-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-white font-bold shadow-xs hover:shadow-md transition-all hover:-translate-y-0.5"
                  style={{ padding: 'clamp(8px, 1vw, 16px) clamp(16px, 2.5vw, 28px)', fontSize: 'clamp(12px, 1.1vw, 16px)' }}>
                  {t('在线体验 Playground')}
                </Link>
              </div>
            </div>

            <div style={{ maxWidth: 'min(100%, 1000px)', margin: '0 auto', display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(140px, 20vw, 280px), 1fr))', gap: 'clamp(6px, 1vw, 16px)' }}>
              {heroCards.map((card, i) => (
                <a key={i} href={card.link} target="_blank" rel="noopener noreferrer"
                  className="group rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 hover:shadow-lg hover:border-blue-300 dark:hover:border-blue-600 transition-all hover:-translate-y-0.5"
                  style={{ padding: 'clamp(10px, 1.5vw, 24px)' }}>
                  <div className="rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mb-[clamp(6px, 0.8vw, 16px)]"
                    style={{ width: 'clamp(24px, 3vw, 36px)', height: 'clamp(24px, 3vw, 36px)' }}>
                    <img src="/logo.webp" alt="" className="rounded-lg object-cover" style={{ width: '100%', height: '100%' }} />
                  </div>
                  <h3 className="font-bold" style={{ fontSize: 'clamp(11px, 1.2vw, 16px)', marginBottom: 'clamp(2px, 0.3vw, 6px)' }}>{t(card.title)}</h3>
                  <p className="text-slate-500 dark:text-slate-400 leading-relaxed" style={{ fontSize: 'clamp(9px, 1vw, 14px)' }}>{t(card.desc)}</p>
                </a>
              ))}
            </div>
          </section>

          {/* FEATURES */}
          <section style={{ paddingTop: 'clamp(32px, 6vw, 80px)', paddingBottom: 'clamp(32px, 6vw, 80px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(20px, 3vw, 48px)' }}>
              <Badge variant="outline" className="border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 mb-3"
                style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                Core Features
              </Badge>
              <h2 className="font-bold" style={{ fontSize: 'clamp(18px, 2.5vw, 36px)', marginBottom: 'clamp(6px, 1vw, 12px)' }}>
                {t('Token 聚合分发 · 企业级 AI 基础设施')}
              </h2>
              <p className="text-slate-500 dark:text-slate-400 mx-auto" style={{ fontSize: 'clamp(12px, 1.2vw, 16px)', maxWidth: 'clamp(280px, 50vw, 500px)' }}>
                {t('一个平台统一管理所有 AI 密钥、配额、计费')}
              </p>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(200px, 28vw, 380px), 1fr))', gap: 'clamp(8px, 1.2vw, 24px)' }}>
              {features.map((f, i) => (
                <div key={i} className="group rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 hover:shadow-lg hover:border-blue-300 dark:hover:border-blue-700 transition-all"
                  style={{ padding: 'clamp(14px, 2vw, 28px)' }}>
                  <div className={'rounded-xl bg-gradient-to-br ' + f.grad + ' flex items-center justify-center mb-[clamp(10px, 1.5vw, 20px)]'}
                    style={{ width: 'clamp(32px, 4vw, 44px)', height: 'clamp(32px, 4vw, 44px)' }}>
                    <f.icon className="text-white" style={{ width: 'clamp(16px, 2vw, 22px)', height: 'clamp(16px, 2vw, 22px)' }} />
                  </div>
                  <h3 className="font-bold" style={{ fontSize: 'clamp(13px, 1.3vw, 18px)', marginBottom: 'clamp(4px, 0.5vw, 10px)' }}>{t(f.title)}</h3>
                  <p className="text-slate-500 dark:text-slate-400 leading-relaxed" style={{ fontSize: 'clamp(11px, 1.1vw, 15px)' }}>{t(f.desc)}</p>
                </div>
              ))}
            </div>
          </section>

          {/* STATS */}
          <section className="border-t border-slate-100 dark:border-slate-800" style={{ paddingTop: 'clamp(24px, 4vw, 48px)', paddingBottom: 'clamp(24px, 4vw, 48px)' }}>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(100px, 18vw, 200px), 1fr))', gap: 'clamp(16px, 3vw, 40px)', maxWidth: 'clamp(300px, 60vw, 700px)', margin: '0 auto', textAlign: 'center' }}>
              {stats.map((s, i) => (
                <div key={i}>
                  <div className="font-extrabold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent" style={{ fontSize: 'clamp(20px, 3vw, 48px)' }}>{s.value}</div>
                  <div className="text-slate-500 dark:text-slate-400" style={{ fontSize: 'clamp(11px, 1vw, 14px)', marginTop: 'clamp(2px, 0.3vw, 6px)' }}>{t(s.label)}</div>
                </div>
              ))}
            </div>
          </section>

          {/* MODELS */}
          <section className="border-t border-slate-100 dark:border-slate-800" style={{ paddingTop: 'clamp(32px, 5vw, 80px)', paddingBottom: 'clamp(32px, 5vw, 80px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(20px, 3vw, 48px)' }}>
              <Badge variant="outline" className="border-purple-200 dark:border-purple-800 bg-purple-50 dark:bg-purple-900/20 mb-3"
                style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                Ecosystem
              </Badge>
              <h2 className="font-bold" style={{ fontSize: 'clamp(18px, 2.5vw, 36px)', marginBottom: 'clamp(6px, 1vw, 12px)' }}>{t('AI Model Ecosystem')}</h2>
              <p className="text-slate-500 dark:text-slate-400 mx-auto" style={{ fontSize: 'clamp(12px, 1.2vw, 16px)', maxWidth: 'clamp(280px, 50vw, 500px)' }}>
                {t('聚合全球主流 AI 模型，一个接口兼容全部')}
              </p>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(120px, 15vw, 200px), 1fr))', gap: 'clamp(6px, 0.8vw, 20px)' }}>
              {modelApps.map((app, i) => (
                <a key={i} href={app.link} target="_blank" rel="noopener noreferrer"
                  className="group rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 hover:shadow-lg hover:border-blue-300 dark:hover:border-blue-600 transition-all hover:-translate-y-1 text-center"
                  style={{ padding: 'clamp(10px, 1.5vw, 24px)' }}>
                  <div className={'rounded-xl bg-gradient-to-br ' + app.grad + ' flex items-center justify-center mx-auto mb-[clamp(6px, 0.8vw, 14px)]'}
                    style={{ width: 'clamp(28px, 4vw, 44px)', height: 'clamp(28px, 4vw, 44px)' }}>
                    <app.icon className="text-white" style={{ width: 'clamp(14px, 2vw, 22px)', height: 'clamp(14px, 2vw, 22px)' }} />
                  </div>
                  <h3 className="font-bold truncate" style={{ fontSize: 'clamp(11px, 1.2vw, 16px)', marginBottom: 'clamp(4px, 0.5vw, 10px)' }}>{t(app.name)}</h3>
                  <div className="flex flex-wrap gap-1 justify-center">
                    {app.models.slice(0, 2).map((m, j) => (
                      <span key={j} className="px-[clamp(3px, 0.4vw, 6px)] py-[clamp(1px, 0.2vw, 3px)] rounded-full bg-slate-100 dark:bg-slate-800 text-slate-500 truncate"
                        style={{ fontSize: 'clamp(8px, 0.8vw, 11px)', maxWidth: 'clamp(50px, 8vw, 70px)' }}>{m}</span>
                    ))}
                    {app.models.length > 2 && <span style={{ fontSize: 'clamp(8px, 0.8vw, 11px)' }} className="text-slate-400">{'+' + (app.models.length - 2)}</span>}
                  </div>
                </a>
              ))}
            </div>
          </section>

          {/* NEWS */}
          <section className="border-t border-slate-100 dark:border-slate-800" style={{ paddingTop: 'clamp(32px, 5vw, 80px)', paddingBottom: 'clamp(32px, 5vw, 80px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(20px, 3vw, 48px)' }}>
              <Badge variant="outline" className="border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 mb-3"
                style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                AI News
              </Badge>
              <h2 className="font-bold" style={{ fontSize: 'clamp(18px, 2.5vw, 36px)', marginBottom: 'clamp(6px, 1vw, 12px)' }}>{t('AI News Navigation')}</h2>
              <p className="text-slate-500 dark:text-slate-400 mx-auto" style={{ fontSize: 'clamp(12px, 1.2vw, 16px)', maxWidth: 'clamp(280px, 50vw, 500px)' }}>
                {t('精选 AI 资讯源，掌握行业前沿动态')}
              </p>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(160px, 22vw, 300px), 1fr))', gap: 'clamp(6px, 0.8vw, 16px)' }}>
              {newsSources.map((src, i) => (
                <a key={i} href={src.url} target="_blank" rel="noopener noreferrer"
                  className="group flex items-center justify-between rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 hover:shadow-md hover:border-blue-300 dark:hover:border-blue-600 transition-all hover:-translate-y-0.5"
                  style={{ padding: 'clamp(8px, 1vw, 16px) clamp(10px, 1.5vw, 20px)' }}>
                  <span className="font-medium group-hover:text-blue-600 transition-colors truncate" style={{ fontSize: 'clamp(11px, 1vw, 14px)' }}>{t(src.name)}</span>
                  <span className="px-[clamp(3px, 0.3vw, 6px)] py-[clamp(1px, 0.2vw, 3px)] rounded bg-slate-100 dark:bg-slate-800 text-slate-500 shrink-0 ml-1"
                    style={{ fontSize: 'clamp(8px, 0.7vw, 10px)' }}>{t(src.lang)}</span>
                </a>
              ))}
            </div>
          </section>

          {/* CTA */}
          <section className="border-t border-slate-100 dark:border-slate-800" style={{ paddingTop: 'clamp(24px, 4vw, 48px)', paddingBottom: 'clamp(24px, 4vw, 48px)' }}>
            <div className="text-center" style={{ maxWidth: 'clamp(280px, 50vw, 500px)', margin: '0 auto' }}>
              <h2 className="font-bold" style={{ fontSize: 'clamp(18px, 2.5vw, 32px)', marginBottom: 'clamp(6px, 0.8vw, 12px)' }}>
                {t('开始你的 AI 之旅')}
              </h2>
              <p className="text-slate-500 dark:text-slate-400" style={{ fontSize: 'clamp(12px, 1.2vw, 16px)', marginBottom: 'clamp(16px, 2vw, 24px)' }}>
                {t('一个 API Key 接入所有主流 AI 模型')}
              </p>
              <Link to="/sign-in"
                className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 text-white font-bold shadow-md hover:shadow-lg transition-all hover:-translate-y-0.5"
                style={{ padding: 'clamp(8px, 1vw, 16px) clamp(20px, 3vw, 36px)', fontSize: 'clamp(12px, 1.1vw, 16px)' }}>
                {t('免费注册')}
                <ArrowRight style={{ width: 'clamp(12px, 1vw, 16px)', height: 'clamp(12px, 1vw, 16px)' }} />
              </Link>
            </div>
          </section>
        </div>

        {/* FOOTER */}
        <footer className="border-t border-slate-200 dark:border-slate-800">
          <div className="mx-auto" style={{ maxWidth: 'min(92vw, 1400px)', padding: '0 clamp(8px, 3vw, 32px)' }}>
            <div style={{ paddingTop: 'clamp(16px, 2vw, 32px)', paddingBottom: 'clamp(16px, 2vw, 32px)' }}>
              <div className="flex flex-col sm:flex-row items-center justify-between" style={{ gap: 'clamp(8px, 1.5vw, 16px)' }}>
                <div className="flex items-center gap-2">
                  <div className="flex items-center justify-center rounded-md bg-gradient-to-br from-blue-600 to-purple-600 text-white"
                    style={{ width: 'clamp(22px, 2.5vw, 30px)', height: 'clamp(22px, 2.5vw, 30px)' }}>
                    <img src="/logo.webp" alt="" className="rounded-lg object-cover" style={{ width: '100%', height: '100%' }} />
                  </div>
                  <div>
                    <div className="font-bold" style={{ fontSize: 'clamp(11px, 1vw, 14px)' }}>{t('QuantumClaw')}</div>
                    <div className="text-slate-400" style={{ fontSize: 'clamp(9px, 0.8vw, 12px)' }}>{t('AI API Gateway')}</div>
                  </div>
                </div>
                <div className="flex items-center" style={{ gap: 'clamp(8px, 1.5vw, 24px)', fontSize: 'clamp(9px, 0.8vw, 12px)' }}>
                  <a href="https://github.com" target="_blank" rel="noopener noreferrer" className="text-slate-400 hover:text-blue-600 transition-colors flex items-center gap-[clamp(2px, 0.3vw, 6px)]">
                    <Code style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />GitHub
                  </a>
                  <a href="mailto:contact@quantumclaw.ai" className="text-slate-400 hover:text-blue-600 transition-colors flex items-center gap-[clamp(2px, 0.3vw, 6px)]">
                    <Mail style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />{t('联系我们')}
                  </a>
                </div>
              </div>
              <div className="text-center text-slate-400 border-t border-slate-100 dark:border-slate-800"
                style={{ marginTop: 'clamp(12px, 1.5vw, 20px)', paddingTop: 'clamp(8px, 1vw, 16px)', fontSize: 'clamp(9px, 0.8vw, 11px)' }}>
                &copy; {new Date().getFullYear()} {t('QuantumClaw')}
              </div>
            </div>
          </div>
        </footer>
      </main>
    </>
  )
}



