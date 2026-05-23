import { createFileRoute, Link } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  ArrowRight,
  Brain, MessageSquare, Globe, Code,
  Mail,
  DollarSign, KeyRound, BarChart3, Network,
  Menu, X, Shield, Activity,
  ChevronDown, Check, Newspaper, BookOpen,
  Server, Bot, ChevronRight, Database, Users, Sparkles, Cpu, ShieldCheck, Languages
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAuthStore } from '@/stores/auth-store'
import { useSystemConfigStore } from '@/stores/system-config-store'
import { CustomerServiceFloating } from '@/components/customer-service'

export const Route = createFileRoute('/')({
  component: HomePage,
})

const typeToCode: Record<string, string> = {
  '中文简体': 'zh-CN',
  '中文繁体': 'zh-TW',
  'English': 'en',
  'Français': 'fr',
  '日本語': 'ja',
  'Русский': 'ru',
  'Tiếng Việt': 'vi',
}

function ImageIcon(p: any) { return <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...p}><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg> }
function VideoIcon(p: any) { return <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...p}><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg> }
function PenIcon(p: any) { return <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" {...p}><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/><path d="m15 5 4 4"/></svg> }

// ── API Fetcher for landing page stats ──────────────────────────────────
async function fetchApi<T>(url: string): Promise<T | null> {
  try {
    const res = await fetch(url)
    const json = await res.json()
    return json?.data ?? json ?? null
  } catch {
    return null
  }
}

function HomePage() {
  const [dailyNews, setDailyNews] = useState([])
  const [newsLoading, setNewsLoading] = useState(true)
  const [languages, setLanguages] = useState<string[]>([])
  const [loadingLangs, setLoadingLangs] = useState(true)

  // ── API-driven stats ────────────────────────────────────────────
  const [modelCount, setModelCount] = useState<number | null>(null)
  const [userCount, setUserCount] = useState<number | null>(null)
  const [apiCalls, setApiCalls] = useState<string | null>(null)
  const [topModels, setTopModels] = useState<{ name: string; provider: string; description?: string }[]>([])
  const [statsLoading, setStatsLoading] = useState(true)

  // Load stats from API
  useEffect(() => {
    let cancelled = false

    async function load() {
      // Try to get model count from /api/models
      const models = await fetchApi<unknown[]>('/api/models')
      if (!cancelled && Array.isArray(models)) {
        setModelCount(models.length)
        // Top 6 models for preview
        const preview = models.slice(0, 6).map((m: any) => ({
          name: m.name || m.model_name || m.ModelName || '',
          provider: m.provider || m.Provider || m.provider_name || '',
          description: m.description || m.Description || '',
        }))
        setTopModels(preview)
      }

      // Try to get user count from /api/user (admin only)
      const users = await fetchApi<any>('/api/user')
      if (!cancelled && users) {
        if (Array.isArray(users)) {
          setUserCount(users.length)
        } else if (users.records && Array.isArray(users.records)) {
          setUserCount(users.total || users.records.length)
        }
      }

      // Get total API calls from /api/status or dashboard
      const status = await fetchApi<any>('/api/status')
      if (!cancelled && status) {
        // Try various fields
      }

      setStatsLoading(false)
    }

    load()
    return () => { cancelled = true }
  }, [])

  useEffect(() => {
    fetch('/api/rss/articles?language=zh&limit=20')
      .then(r => r.json())
      .then(res => {
        if (res.success && res.data?.articles) {
          setDailyNews(res.data.articles)
        }
      })
      .catch(() => {})
      .finally(() => setNewsLoading(false))
  }, [])

  // Load available languages from DB
  useEffect(() => {
    fetch('/api/languages')
      .then(r => r.json())
      .then(res => {
        const list = Array.isArray(res) ? res : res?.data || []
        if (Array.isArray(list)) {
          setLanguages(list.map((r: any) => r.languages_type).filter(Boolean))
        }
      })
      .catch(() => {})
      .finally(() => setLoadingLangs(false))
  }, [])

  const { i18n, t } = useTranslation()
  const [menuOpen, setMenuOpen] = useState(false)
  const [currentLang, setCurrentLang] = useState(i18n.language || 'en')
  const { auth } = useAuthStore()
  const loggedIn = !!auth.user
  const { config: sysConfig } = useSystemConfigStore()

  const getNavItems = (loggedIn: boolean) => [
    { to: '/', label: t('Home') },
    { to: '/models', label: t('Models') },
    { to: '/rankings', label: t('Rankings') },
    { to: '/pricing', label: t('Pricing') },
    { to: '/chat', label: t('AI Chat') },
    { to: loggedIn ? '/dashboard' : '/sign-in', label: t('Console') },
  ]

  // ── Feature cards (matching task spec) ───────────────────────────
  const features = [
    { icon: Database, title: t('AI Model Catalog'), desc: t('Browse and discover 50+ AI models with detailed specs, pricing, and performance metrics'), grad: 'from-blue-500 to-cyan-500' },
    { icon: Network, title: t('Smart Routing'), desc: t('Intelligent request distribution with load balancing, failover, and latency-based routing'), grad: 'from-purple-500 to-pink-500' },
    { icon: Languages, title: t('Multilingual Support'), desc: t('Full i18n support with multiple languages for global users and teams'), grad: 'from-green-500 to-emerald-500' },
    { icon: ShieldCheck, title: t('Enterprise Security'), desc: t('SSRF protection, IP whitelist, audit logs, and enterprise-grade access controls'), grad: 'from-orange-500 to-red-500' },
    { icon: KeyRound, title: t('API Key Management'), desc: t('Unified key lifecycle management with quota, permissions, and automatic rotation'), grad: 'from-indigo-500 to-blue-500' },
    { icon: Cpu, title: t('Quantum Computing'), desc: t('IonQ / IBM Q / Rigetti / AWS Braket multi-platform quantum computing aggregation'), grad: 'from-purple-500 to-indigo-500' },
  ]

  const featuresExtras = [
    { icon: BarChart3, title: t('Usage Monitoring'), desc: t('Real-time quota management, consumption details'), grad: 'from-blue-500 to-cyan-500' },
    { icon: DollarSign, title: t('Flexible Billing'), desc: t('Pay-as-you-go / Subscription / Rate control'), grad: 'from-indigo-500 to-blue-500' },
    { icon: Activity, title: t('Async Tasks'), desc: t('MJ/Video/Music task scheduling'), grad: 'from-pink-500 to-rose-500' },
  ]

  const allFeatures = [...features, ...featuresExtras]

  const heroCards = [
    { title: t('Qwen3.6'), desc: t('Native multimodal, more stable reasoning'), link: 'https://qwen.alibaba.com', grad: 'from-blue-500 to-purple-600', logo: '🤖' },
    { title: t('Claude 3.5'), desc: t('Anthropic Strongest Reasoning'), link: 'https://claude.ai', grad: 'from-orange-500 to-amber-600', logo: '🧠' },
    { title: t('GPT-4o'), desc: t('OpenAI Flagship Multimodal'), link: 'https://chat.openai.com', grad: 'from-green-500 to-emerald-600', logo: '✨' },
    { title: t('Gemini 2.0'), desc: t('Google Fast Reasoning'), link: 'https://gemini.google.com', grad: 'from-indigo-500 to-cyan-600', logo: '⭐' },
    { title: t('DeepSeek V3'), desc: t('Most Cost-Effective'), link: 'https://chat.deepseek.com', grad: 'from-blue-500 to-cyan-600', logo: '🔍' },
    { title: t('IonQ'), desc: t('Trapped ion quantum computing with high fidelity'), link: 'https://ionq.com', grad: 'from-purple-500 to-pink-600', logo: '⚛' },
    { title: t('IBM Q'), desc: t('Superconducting qubit quantum processors'), link: 'https://quantum.ibm.com', grad: 'from-blue-600 to-indigo-700', logo: '🔬' },
    { title: t('Rigetti'), desc: t('Hybrid quantum-classical computing platform'), link: 'https://rigetti.com', grad: 'from-emerald-500 to-teal-600', logo: '🔗' },
    { title: t('AWS Braket'), desc: t('Cloud quantum computing service by AWS'), link: 'https://aws.amazon.com/braket', grad: 'from-orange-500 to-red-600', logo: '☁️' },
    { title: t('Azure Quantum'), desc: t('Microsoft quantum computing ecosystem'), link: 'https://quantum.microsoft.com', grad: 'from-cyan-500 to-blue-600', logo: '🟦' },
  ]

  // Preload brand favicons
  useEffect(() => {
    ['qwen.alibaba.com','claude.ai','chat.openai.com','gemini.google.com','chat.deepseek.com','ionq.com'].forEach(d => {
      const img = new Image(); img.src = 'https://www.google.com/s2/favicons?domain='+d+'&sz=64'
    })
  }, [])

  const modelApps = [
    { name: t('Chat & Reasoning'), icon: MessageSquare, models: ['GPT-4o', 'Claude 3.5', 'Gemini 2.0', 'DeepSeek V3'], link: 'https://chat.openai.com', grad: 'from-blue-500 to-cyan-500' },
    { name: t('Image Generation'), icon: ImageIcon, models: ['DALL-E 3', 'Midjourney', 'Stable Diffusion'], link: 'https://www.midjourney.com', grad: 'from-purple-500 to-pink-500' },
    { name: t('Programming & Development'), icon: Code, models: ['GPT-4o', 'Claude 3.5', 'DeepSeek Coder'], link: 'https://github.com/features/copilot', grad: 'from-orange-500 to-red-500' },
    { name: t('Audio & Video'), icon: VideoIcon, models: ['Suno', 'Runway', 'ElevenLabs'], link: 'https://suno.com', grad: 'from-pink-500 to-rose-500' },
    { name: t('Text & Writing'), icon: PenIcon, models: ['GPT-4o', 'Claude 3', 'Kimi', 'Moonshot'], link: 'https://www.grammarly.com', grad: 'from-green-500 to-emerald-500' },
    { name: t('Search & Knowledge'), icon: Brain, models: ['Perplexity', 'ArXiv', 'Hugging Face'], link: 'https://www.perplexity.ai', grad: 'from-indigo-500 to-blue-500' },
    { name: t('Quantum Computing'), icon: Server, models: ['IonQ Harmony', 'IBM Sherbrooke', 'Rigetti Aspen'], link: 'https://ionq.com', grad: 'from-purple-500 to-pink-500' },
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

  // ── API-driven stats ─────────────────────────────────────────────
  const liveStats = [
    { value: modelCount !== null ? `${modelCount}+` : '50+', label: t('AI Models'), loading: statsLoading && modelCount === null },
    { value: userCount !== null ? `${userCount.toLocaleString()}+` : '10K+', label: t('Users'), loading: statsLoading && userCount === null },
    { value: apiCalls || '10M+', label: t('Monthly API Calls'), loading: false },
    { value: '6', label: t('Quantum Platforms'), loading: false },
    { value: '99.9%', label: t('Service Availability'), loading: false },
  ]

  const newsIcons: Record<string, React.ReactNode> = {
    'Machine Heart': <Newspaper style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    'QbitAI': <Newspaper style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    '36Kr AI': <Newspaper style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    'Leiphone': <Newspaper style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    'OpenAI Blog': <Brain style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    'Anthropic': <Brain style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    'Google AI': <Brain style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    'MIT Tech Review': <Globe style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    'ArXiv': <BookOpen style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    'Hugging Face': <Globe style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
    'Reddit AI': <Globe style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />,
  }

  useEffect(() => { setCurrentLang(i18n.language) }, [i18n.language])

  const changeLang = (code: string) => {
    i18n.changeLanguage(code)
    localStorage.setItem('i18nextLng', code)
    setCurrentLang(code)
  }
  const langName = (Object.entries(typeToCode).find(([_, code]) => code === currentLang)?.[0]) || currentLang || 'English'

  return (
    <div className="min-h-screen flex flex-col bg-white dark:bg-slate-950">
      {/* ── NAVIGATION BAR ─────────────────────────────────────────── */}
      <nav className="shrink-0 border-b border-slate-200 dark:border-slate-800" style={{ height: 'clamp(44px, 7vh, 56px)' }}>
        <div className="mx-auto flex items-center h-full justify-between w-full" style={{ padding: '0 clamp(8px, 3vw, 32px)' }}>
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
                {languages.filter(Boolean).map((langType) => (
                  <DropdownMenuItem key={typeToCode[langType] || langType} onClick={() => changeLang(typeToCode[langType] || langType)}
                    className={currentLang === (typeToCode[langType] || langType) ? 'font-semibold bg-slate-50 dark:bg-slate-800' : ''}
                    style={{ fontSize: 'clamp(11px, 0.9vw, 13px)' }}>
                    <Globe className="mr-2 shrink-0" style={{ width: 'clamp(10px, 0.8vw, 14px)', height: 'clamp(10px, 0.8vw, 14px)' }} />
                    {langType}
                    {currentLang === (typeToCode[langType] || langType) && <Check className="ml-auto" style={{ width: 'clamp(10px, 0.8vw, 14px)', height: 'clamp(10px, 0.8vw, 14px)' }} />}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>

            {loggedIn ? (
              <Link to="/dashboard"
                className="inline-flex items-center gap-2 rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-xs hover:shadow-md transition-all"
                style={{ padding: 'clamp(4px, 0.5vw, 10px) clamp(8px, 1.2vw, 16px)' }}>
                <Avatar className="h-7 w-7">
                  <AvatarFallback className="bg-gradient-to-br from-blue-600 to-purple-600 text-white text-xs">
                    {(auth.user?.display_name || auth.user?.username || 'U')[0]}
                  </AvatarFallback>
                </Avatar>
                <span className="text-sm font-medium hidden sm:inline">{auth.user?.display_name || auth.user?.username}</span>
              </Link>
            ) : (
              <Link to="/sign-in" search={{ redirect: undefined }}
                className="inline-flex items-center justify-center rounded-md bg-gradient-to-r from-blue-600 to-purple-600 text-white font-semibold hover:shadow-md transition-all whitespace-nowrap"
                style={{ padding: 'clamp(4px, 0.5vw, 8px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(11px, 0.9vw, 13px)' }}>
                {t('登录 / 注册')}
              </Link>
            )}

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
                {languages.filter(Boolean).map((langType) => (
                  <button key={typeToCode[langType] || langType} onClick={() => { changeLang(typeToCode[langType] || langType); setMenuOpen(false) }}
                    className={'w-full text-left px-3 py-1.5 rounded-md text-sm ' + (currentLang === (typeToCode[langType] || langType) ? 'font-semibold bg-slate-100 dark:bg-slate-800' : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800')}>
                    {langType}
                  </button>
                ))}
              </div>
              {loggedIn ? (
                <Link to="/dashboard" onClick={() => setMenuOpen(false)}
                  className="flex items-center gap-2 px-3 py-2 rounded-md text-sm font-semibold bg-gradient-to-r from-blue-600 to-purple-600 text-white mt-2">
                  <Avatar className="h-6 w-6">
                    <AvatarFallback className="text-white text-[10px] bg-white/20">
                      {(auth.user?.display_name || auth.user?.username || 'U')[0]}
                    </AvatarFallback>
                  </Avatar>
                  <span>{auth.user?.display_name || auth.user?.username || t('Profile')}</span>
                </Link>
              ) : (
                <Link to="/sign-in" search={{ redirect: undefined }} onClick={() => setMenuOpen(false)}
                  className="block px-3 py-2 rounded-md text-sm font-semibold text-white bg-gradient-to-r from-blue-600 to-purple-600 mt-2">
                  {t('登录 / 注册')}
                </Link>
              )}
            </div>
          </div>
        )}
      </nav>

      <main className="flex-1">
        <div className="w-full" style={{ padding: '0 clamp(8px, 3vw, 32px)' }}>

          {/* ── HERO SECTION ───────────────────────────────────────── */}
          <section style={{ paddingTop: 'clamp(24px, 5vw, 64px)', paddingBottom: 'clamp(24px, 5vw, 64px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(16px, 3vw, 48px)' }}>
              <Badge variant="outline" className="border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 mb-4"
                style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                <Sparkles className="mr-1" style={{ width: 'clamp(10px, 0.8vw, 14px)', height: 'clamp(10px, 0.8vw, 14px)' }} />
                {t('Next-Gen AI API Gateway')}
              </Badge>

              <h1 className="font-extrabold tracking-tight leading-[1.15]" style={{ fontSize: 'clamp(28px, 5vw, 72px)', marginBottom: 'clamp(8px, 1.5vw, 16px)' }}>
                <span className="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">{t('QuantumClaw')}</span>
              </h1>
              <p className="text-slate-500 dark:text-slate-400 mx-auto" style={{ fontSize: 'clamp(13px, 1.5vw, 20px)', maxWidth: 'clamp(280px, 60vw, 600px)' }}>
                <span className="font-semibold text-slate-700 dark:text-slate-300">{t('Key聚合分发')}</span> {t('·')} {t('聚合调用 AI 大模型 + 量子算力资源')}
              </p>
              <div className="flex flex-wrap gap-y-2 gap-x-3 justify-center" style={{ marginTop: 'clamp(16px, 3vw, 32px)' }}>
                <Link to="/sign-in" search={{ redirect: undefined }}
                  className="inline-flex items-center justify-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 text-white font-bold shadow-md hover:shadow-lg transition-all hover:-translate-y-0.5"
                  style={{ padding: 'clamp(8px, 1vw, 16px) clamp(16px, 2.5vw, 28px)', fontSize: 'clamp(12px, 1.1vw, 16px)' }}>
                  <KeyRound style={{ width: 'clamp(14px, 1.3vw, 20px)', height: 'clamp(14px, 1.3vw, 20px)' }} />
                  {t('开始使用')}
                  <ArrowRight style={{ width: 'clamp(12px, 1vw, 16px)', height: 'clamp(12px, 1vw, 16px)' }} />
                </Link>
                <Link to="/models"
                  className="inline-flex items-center justify-center gap-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-white font-bold shadow-xs hover:shadow-md transition-all hover:-translate-y-0.5"
                  style={{ padding: 'clamp(8px, 1vw, 16px) clamp(16px, 2.5vw, 28px)', fontSize: 'clamp(12px, 1.1vw, 16px)' }}>
                  {t('了解更多')}
                </Link>
              </div>
            </div>

            {/* ── Model preview cards from API + static hero cards ── */}
            <div style={{ maxWidth: 'min(100%, 1200px)', margin: '0 auto', display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(140px, 20vw, 280px), 1fr))', gap: 'clamp(6px, 1vw, 16px)' }}>
              {heroCards.map((card, i) => (
                <a key={i} href={card.link} target="_blank" rel="noopener noreferrer"
                  className="group rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 hover:shadow-lg hover:border-blue-300 dark:hover:border-blue-600 transition-all hover:-translate-y-0.5"
                  style={{ padding: 'clamp(10px, 1.5vw, 24px)' }}>
                  <div className={"rounded-lg bg-gradient-to-br " + card.grad + " flex items-center justify-center mb-[clamp(6px, 0.8vw, 16px)]"}
                    style={{ width: 'clamp(24px, 3vw, 36px)', height: 'clamp(24px, 3vw, 36px)' }}>
                    <img src={"https://www.google.com/s2/favicons?domain=" + card.link.replace("https://","").split("/")[0] + "&sz=64"} alt=""
                      onError={(e)=>{e.currentTarget.style.display="none"; const p=e.currentTarget.parentElement; if(p){const s=document.createElement("span");s.textContent=card.logo;s.style.fontSize="clamp(12px,1.5vw,18px)";p.appendChild(s);}}}
                      className="rounded-lg object-cover" style={{width:"100%",height:"100%"}} />
                  </div>
                  <h3 className="font-bold" style={{ fontSize: 'clamp(11px, 1.2vw, 16px)', marginBottom: 'clamp(2px, 0.3vw, 6px)' }}>{t(card.title)}</h3>
                  <p className="text-slate-500 dark:text-slate-400 leading-relaxed" style={{ fontSize: 'clamp(9px, 1vw, 14px)' }}>{t(card.desc)}</p>
                </a>
              ))}
            </div>
          </section>

          {/* ── STATISTICS SECTION ──────────────────────────────────── */}
          <section className="border-t border-slate-100 dark:border-slate-800" style={{ paddingTop: 'clamp(24px, 4vw, 48px)', paddingBottom: 'clamp(24px, 4vw, 48px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(16px, 2vw, 32px)' }}>
              <Badge variant="outline" className="border-emerald-200 dark:border-emerald-800 bg-emerald-50 dark:bg-emerald-900/20 mb-3"
                style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                <Activity className="mr-1" style={{ width: 'clamp(10px, 0.8vw, 14px)', height: 'clamp(10px, 0.8vw, 14px)' }} />
                {t('Platform Statistics')}
              </Badge>
              <p className="text-slate-500 dark:text-slate-400 text-sm sm:text-base">
                {t('Trusted by developers and enterprises worldwide')}
              </p>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(100px, 18vw, 200px), 1fr))', gap: 'clamp(16px, 3vw, 40px)', maxWidth: 'clamp(300px, 70vw, 900px)', margin: '0 auto', textAlign: 'center' }}>
              {liveStats.map((s, i) => (
                <div key={i}>
                  {s.loading ? (
                    <Skeleton className="h-10 w-20 mx-auto mb-2" />
                  ) : (
                    <div className="font-extrabold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent" style={{ fontSize: 'clamp(20px, 3vw, 48px)' }}>
                      {s.value}
                    </div>
                  )}
                  <div className="text-slate-500 dark:text-slate-400" style={{ fontSize: 'clamp(11px, 1vw, 14px)', marginTop: 'clamp(2px, 0.3vw, 6px)' }}>{t(s.label)}</div>
                </div>
              ))}
            </div>
          </section>

          {/* ── FEATURES SECTION ────────────────────────────────────── */}
          <section className="border-t border-slate-100 dark:border-slate-800" style={{ paddingTop: 'clamp(32px, 6vw, 80px)', paddingBottom: 'clamp(32px, 6vw, 80px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(20px, 3vw, 48px)' }}>
              <Badge variant="outline" className="border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 mb-3"
                style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                <Sparkles className="mr-1" style={{ width: 'clamp(10px, 0.8vw, 14px)', height: 'clamp(10px, 0.8vw, 14px)' }} />
                {t('Core Features')}
              </Badge>
              <h2 className="font-bold" style={{ fontSize: 'clamp(18px, 2.5vw, 36px)', marginBottom: 'clamp(6px, 1vw, 12px)' }}>
                {t('Everything You Need for AI Development')}
              </h2>
              <p className="text-slate-500 dark:text-slate-400 mx-auto" style={{ fontSize: 'clamp(12px, 1.2vw, 16px)', maxWidth: 'clamp(280px, 50vw, 500px)' }}>
                {t('From model discovery to deployment, we handle the infrastructure')}
              </p>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(200px, 28vw, 380px), 1fr))', gap: 'clamp(8px, 1.2vw, 24px)' }}>
              {allFeatures.map((f, i) => (
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

          {/* ── MODEL PREVIEW SECTION ──────────────────────────────── */}
          <section className="border-t border-slate-100 dark:border-slate-800" style={{ paddingTop: 'clamp(32px, 5vw, 80px)', paddingBottom: 'clamp(32px, 5vw, 80px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(20px, 3vw, 48px)' }}>
              <Badge variant="outline" className="border-purple-200 dark:border-purple-800 bg-purple-50 dark:bg-purple-900/20 mb-3"
                style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                <Bot className="mr-1" style={{ width: 'clamp(10px, 0.8vw, 14px)', height: 'clamp(10px, 0.8vw, 14px)' }} />
                {t('Featured Models')}
              </Badge>
              <h2 className="font-bold" style={{ fontSize: 'clamp(18px, 2.5vw, 36px)', marginBottom: 'clamp(6px, 1vw, 12px)' }}>{t('Popular AI Models')}</h2>
              <p className="text-slate-500 dark:text-slate-400 mx-auto" style={{ fontSize: 'clamp(12px, 1.2vw, 16px)', maxWidth: 'clamp(280px, 50vw, 500px)' }}>
                {t('Discover and compare top AI models available on our platform')}
              </p>
            </div>

            {/* API-driven top models */}
            {statsLoading && topModels.length === 0 ? (
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(160px, 22vw, 280px), 1fr))', gap: 'clamp(8px, 1.2vw, 20px)' }}>
                {[1,2,3,4,5,6].map(i => (
                  <div key={i} className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900" style={{ padding: 'clamp(14px, 2vw, 24px)' }}>
                    <Skeleton className="h-10 w-10 rounded-lg mb-3" />
                    <Skeleton className="h-5 w-28 mb-2" />
                    <Skeleton className="h-4 w-full" />
                  </div>
                ))}
              </div>
            ) : topModels.length > 0 ? (
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(160px, 22vw, 280px), 1fr))', gap: 'clamp(8px, 1.2vw, 20px)' }}>
                {topModels.map((model, i) => (
                  <div key={i}
                    className="group rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 hover:shadow-lg hover:border-blue-300 dark:hover:border-blue-600 transition-all hover:-translate-y-0.5"
                    style={{ padding: 'clamp(14px, 2vw, 24px)' }}>
                    <div className="flex items-center gap-3 mb-3">
                      <div className="rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold"
                        style={{ width: 'clamp(28px, 3vw, 40px)', height: 'clamp(28px, 3vw, 40px)', fontSize: 'clamp(12px, 1.2vw, 16px)' }}>
                        {model.name.charAt(0).toUpperCase()}
                      </div>
                      <div className="min-w-0">
                        <h3 className="font-bold truncate" style={{ fontSize: 'clamp(12px, 1.2vw, 16px)' }}>{model.name}</h3>
                        {model.provider && (
                          <p className="text-xs text-slate-400 truncate">{model.provider}</p>
                        )}
                      </div>
                    </div>
                    {model.description && (
                      <p className="text-slate-500 dark:text-slate-400 leading-relaxed text-sm line-clamp-2">
                        {model.description}
                      </p>
                    )}
                    <Link to="/models" className="inline-flex items-center gap-1 text-xs text-blue-600 hover:text-blue-700 mt-3 group-hover:underline">
                      {t('View Details')} <ChevronRight style={{ width: 'clamp(8px, 0.7vw, 12px)', height: 'clamp(8px, 0.7vw, 12px)' }} />
                    </Link>
                  </div>
                ))}
              </div>
            ) : (
              /* Fallback: ecosystem cards when API unavailable */
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
            )}
          </section>

          {/* ── AI & QUANTUM DAILY NEWS ─────────────────────────────── */}
          <section style={{ paddingTop: 'clamp(32px, 5vw, 80px)', paddingBottom: 'clamp(24px, 3vw, 48px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(20px, 3vw, 48px)' }}>
              <Badge variant="outline" className="border-purple-200 dark:border-purple-800 bg-purple-50 dark:bg-purple-900/20 mb-3"
                style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                {t('Daily Updates')}
              </Badge>
              <h2 className="font-bold" style={{ fontSize: 'clamp(18px, 2.5vw, 36px)', marginBottom: 'clamp(6px, 1vw, 12px)' }}>{t('AI & Quantum Daily News')}</h2>
              <p className="text-slate-500 dark:text-slate-400 mx-auto" style={{ fontSize: 'clamp(12px, 1.2vw, 16px)', maxWidth: 'clamp(280px, 50vw, 500px)' }}>
                {t('Stay updated with the latest AI models and quantum computing breakthroughs')}
              </p>
            </div>
            <div style={{ maxWidth: 'min(100%, 1200px)', margin: '0 auto' }}>
              {dailyNews.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <p style={{ fontSize: 'clamp(12px, 1.1vw, 15px)' }}>{t('News loading, please check back later...')}</p>
                </div>
              ) : (
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(clamp(300px, 45vw, 560px), 1fr))', gap: 'clamp(4px, 0.6vw, 8px)' }}>
                  {dailyNews.slice(0, 20).map((article: any, i: number) => (
                    <a key={i} href={article.link} target="_blank" rel="noopener noreferrer" className="text-decoration-none">
                      <div className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                        style={{ fontSize: 'clamp(11px, 1vw, 14px)' }}>
                        <span className="px-1.5 py-0.5 rounded bg-purple-100 dark:bg-purple-900 text-purple-700 dark:text-purple-300 font-medium shrink-0"
                          style={{ fontSize: 'clamp(8px, 0.7vw, 11px)' }}>
                          {article.source}
                        </span>
                        <span className="text-slate-400 shrink-0" style={{ fontSize: 'clamp(8px, 0.7vw, 11px)', width: 'clamp(36px, 4vw, 50px)' }}>
                          {new Date(article.published_at).toLocaleDateString(i18n.language || 'zh-CN', { month: '2-digit', day: '2-digit' })}
                        </span>
                        <span className="truncate flex-grow min-w-0" style={{ color: '#1e293b', fontSize: 'clamp(11px, 0.9vw, 14px)' }}>
                          {article.title}
                        </span>
                        <svg className="shrink-0 text-purple-400" style={{ width: 'clamp(8px, 0.7vw, 12px)', height: 'clamp(8px, 0.7vw, 12px)', opacity: 0.6 }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                        </svg>
                      </div>
                    </a>
                  ))}
                </div>
              )}
            </div>
          </section>

          {/* ── NEWS SOURCES ────────────────────────────────────────── */}
          <section className="border-t border-slate-100 dark:border-slate-800" style={{ paddingTop: 'clamp(32px, 5vw, 80px)', paddingBottom: 'clamp(32px, 5vw, 80px)' }}>
            <div className="text-center" style={{ marginBottom: 'clamp(20px, 3vw, 48px)' }}>
              <Badge variant="outline" className="border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 mb-3"
                style={{ padding: 'clamp(3px, 0.4vw, 6px) clamp(8px, 1.2vw, 16px)', fontSize: 'clamp(10px, 0.9vw, 13px)' }}>
                {t('AI News')}
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
                  <span className="font-medium group-hover:text-blue-600 transition-colors truncate inline-flex items-center gap-1.5" style={{ fontSize: 'clamp(11px, 1vw, 14px)' }}>
                    {newsIcons[src.name] || <Brain style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />}
                    {t(src.name)}
                  </span>
                  <span className="px-[clamp(3px, 0.3vw, 6px)] py-[clamp(1px, 0.2vw, 3px)] rounded bg-slate-100 dark:bg-slate-800 text-slate-500 shrink-0 ml-1"
                    style={{ fontSize: 'clamp(8px, 0.7vw, 10px)' }}>{t(src.lang)}</span>
                </a>
              ))}
            </div>
          </section>

          {/* ── CTA SECTION ─────────────────────────────────────────── */}
          <section className="border-t border-slate-100 dark:border-slate-800" style={{ paddingTop: 'clamp(24px, 4vw, 48px)', paddingBottom: 'clamp(24px, 4vw, 48px)' }}>
            <div className="text-center" style={{ maxWidth: 'clamp(280px, 50vw, 500px)', margin: '0 auto' }}>
              <h2 className="font-bold" style={{ fontSize: 'clamp(18px, 2.5vw, 32px)', marginBottom: 'clamp(6px, 0.8vw, 12px)' }}>
                {t('开始你的 AI 之旅')}
              </h2>
              <p className="text-slate-500 dark:text-slate-400" style={{ fontSize: 'clamp(12px, 1.2vw, 16px)', marginBottom: 'clamp(16px, 2vw, 24px)' }}>
                {t('一个 API Key 接入所有主流 AI 模型')}
              </p>
              <Link to="/sign-in" search={{ redirect: undefined }}
                className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 text-white font-bold shadow-md hover:shadow-lg transition-all hover:-translate-y-0.5"
                style={{ padding: 'clamp(8px, 1vw, 16px) clamp(20px, 3vw, 36px)', fontSize: 'clamp(12px, 1.1vw, 16px)' }}>
                {t('免费注册')}
                <ArrowRight style={{ width: 'clamp(12px, 1vw, 16px)', height: 'clamp(12px, 1vw, 16px)' }} />
              </Link>
            </div>
          </section>
        </div>

        {/* ── FOOTER ────────────────────────────────────────────────── */}
        <footer className="border-t border-slate-200 dark:border-slate-800">
          <div className="w-full" style={{ padding: '0 clamp(8px, 3vw, 32px)' }}>
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
                  <Link to="/models" className="text-slate-400 hover:text-blue-600 transition-colors flex items-center gap-[clamp(2px, 0.3vw, 6px)]">
                    <Bot style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />{t('Models')}
                  </Link>
                  <Link to="/pricing" className="text-slate-400 hover:text-blue-600 transition-colors flex items-center gap-[clamp(2px, 0.3vw, 6px)]">
                    <DollarSign style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />{t('Pricing')}
                  </Link>
                  <a href="https://github.com" target="_blank" rel="noopener noreferrer" className="text-slate-400 hover:text-blue-600 transition-colors flex items-center gap-[clamp(2px, 0.3vw, 6px)]">
                    <Code style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />{t('GitHub')}
                  </a>
                  <a href="mailto:contact@quantumclaw.ai" className="text-slate-400 hover:text-blue-600 transition-colors flex items-center gap-[clamp(2px, 0.3vw, 6px)]">
                    <Mail style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />{t('联系我们')}
                  </a>
                  <a onClick={() => { navigator.clipboard.writeText('587600277'); alert('QQ群号已复制: 587600277'); }}
                     className="text-slate-400 hover:text-blue-600 transition-colors flex items-center gap-[clamp(2px, 0.3vw, 6px)] cursor-pointer">
                    <MessageSquare style={{ width: 'clamp(10px, 0.9vw, 14px)', height: 'clamp(10px, 0.9vw, 14px)' }} />{t('QQ 群: 587600277')}
                  </a>
                </div>
              </div>
              <div className="text-center text-slate-400 border-t border-slate-100 dark:border-slate-800"
                style={{ marginTop: 'clamp(12px, 1.5vw, 20px)', paddingTop: 'clamp(8px, 1vw, 16px)', fontSize: 'clamp(9px, 0.8vw, 11px)' }}>
                <p>{t('Copyright')} &copy; 2017-2026 {t('Shenzhen Zhongke Jingwei Intelligent Co., Ltd.')} {t('All Rights Reserved')}.</p>
                {sysConfig?.footerHtml ? (
                  <div className="flex justify-center gap-3 mt-1"
                    dangerouslySetInnerHTML={{ __html: sysConfig.footerHtml }} />
                ) : (
                  <p style={{ marginTop: 'clamp(2px, 0.3vw, 6px)' }}>
                    <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener noreferrer"
                      className="text-slate-400 hover:text-blue-600 transition-colors">
                      {t('粤ICP备21033000号')}
                    </a>
                  </p>
                )}
              </div>
            </div>
          </div>
        </footer>
        <CustomerServiceFloating />
      </main>
    </div>
  )
}
