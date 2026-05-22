import { createFileRoute, Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import {
  Search, SlidersHorizontal, X, ArrowUpDown, RefreshCw,
  ChevronRight, ChevronDown, MessageSquare, Code, Brain,
  Image, Cpu, Atom, CheckCircle, Key, Play, ExternalLink,
  Zap
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { getEnhancedModels, type EnhancedModel } from '@/lib/api-extended'

export const Route = createFileRoute('/models')({
  component: ModelsPage,
})

// ── Model catalog metadata ──
interface ModelCapability {
  inputModalities: string[]
  contextWindow: number
  useCase: string
  series: string
  description: string
}

const modelCatalog: Record<string, ModelCapability> = {
  'gpt-4o': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4', description: 'OpenAI flagship multimodal model with vision, high intelligence and fast responses.' },
  'gpt-4o-mini': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4', description: 'Fast, affordable small model for lightweight tasks with strong performance.' },
  'gpt-4': { inputModalities: ['Text'], contextWindow: 8192, useCase: 'chat', series: 'GPT-4', description: 'OpenAI GPT-4 base model with strong reasoning capabilities.' },
  'gpt-4-turbo': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4', description: 'GPT-4 with vision support and extended context window.' },
  'gpt-3.5-turbo': { inputModalities: ['Text'], contextWindow: 16385, useCase: 'chat', series: 'GPT-3.5', description: 'Fast and cost-effective for simple conversational tasks.' },
  'gpt-4.5': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4', description: 'Improved GPT-4 with better instruction following and reliability.' },
  'gpt-5.5': { inputModalities: ['Text', 'Image'], contextWindow: 256000, useCase: 'chat', series: 'GPT-5', description: 'Next-gen OpenAI model with 2x context window and lower verbosity.' },
  'o1': { inputModalities: ['Text', 'Image'], contextWindow: 200000, useCase: 'reasoning', series: 'o1', description: 'OpenAI reasoning model designed for complex problem solving and analysis.' },
  'o1-mini': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'reasoning', series: 'o1', description: 'Lightweight reasoning model optimized for math and coding tasks.' },
  'o3': { inputModalities: ['Text', 'Image'], contextWindow: 200000, useCase: 'reasoning', series: 'o3', description: 'Advanced reasoning model with extended thinking capabilities.' },
  'claude-3.5-sonnet': { inputModalities: ['Text', 'Image', 'File'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3.5', description: 'Best balance of speed and intelligence from Anthropic.' },
  'claude-3.5-haiku': { inputModalities: ['Text', 'Image', 'File'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3.5', description: 'Fast and affordable Claude model for everyday tasks.' },
  'claude-opus-4': { inputModalities: ['Text', 'Image', 'File'], contextWindow: 200000, useCase: 'chat', series: 'Claude', description: 'Anthropic most powerful model for complex enterprise workloads.' },
  'gemini-2.0-flash': { inputModalities: ['Text', 'Image', 'Audio', 'Video', 'File'], contextWindow: 1000000, useCase: 'chat', series: 'Gemini 2.0', description: 'Google fast multimodal model with 1M context, supports audio/video.' },
  'gemini-2.0-pro': { inputModalities: ['Text', 'Image', 'Audio', 'Video', 'File'], contextWindow: 1000000, useCase: 'chat', series: 'Gemini 2.0', description: 'Google high-quality model with full multimodal input support.' },
  'gemini-3.0-pro': { inputModalities: ['Text', 'Image', 'Audio', 'Video', 'File'], contextWindow: 1000000, useCase: 'reasoning', series: 'Gemini 3.0', description: 'Google latest reasoning model with enhanced analytical capabilities.' },
  'deepseek-chat': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'DeepSeek', description: 'DeepSeek general chat model offering excellent value for its price.' },
  'deepseek-reasoner': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'reasoning', series: 'DeepSeek', description: 'DeepSeek reasoning model with transparent step-by-step thinking.' },
  'deepseek-r1': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'reasoning', series: 'DeepSeek', description: 'Open-source reasoning model with top-tier performance benchmarks.' },
  'deepseek-v3': { inputModalities: ['Text'], contextWindow: 64000, useCase: 'coding', series: 'DeepSeek', description: 'DeepSeek latest model optimized for code generation tasks.' },
  'qwen-max': { inputModalities: ['Text'], contextWindow: 32000, useCase: 'chat', series: 'Qwen', description: 'Alibaba strongest general-purpose language model.' },
  'qwen-plus': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'Qwen', description: 'Alibaba upgraded model with extended context understanding.' },
  'qwen2.5-vl': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'vision', series: 'Qwen 2.5', description: 'Alibaba multimodal vision-language model for image understanding.' },
  'mistral-large': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'Mistral', description: 'Mistral AI flagship model for complex reasoning tasks.' },
  'mixtral-8x7b': { inputModalities: ['Text'], contextWindow: 32000, useCase: 'chat', series: 'Mistral', description: 'Mistral MoE architecture offering good quality-to-cost ratio.' },
  'codestral': { inputModalities: ['Text'], contextWindow: 32000, useCase: 'coding', series: 'Mistral', description: 'Mistral dedicated code generation model for developers.' },
  'llama-3.1-70b': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'Llama 3.1', description: 'Meta open-source model with strong general performance.' },
  'llama-3.1-405b': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'Llama 3.1', description: 'Meta largest open-source model approaching GPT-4 level quality.' },
  'llama-3.2-vision': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'vision', series: 'Llama 3.2', description: 'Meta multimodal open-source model for vision-language tasks.' },
}

const useCaseLabels: Record<string, { label: string; icon: React.ElementType; color: string }> = {
  'chat': { label: 'Chat & Assistant', icon: MessageSquare, color: 'from-blue-500 to-blue-600' },
  'coding': { label: 'Code Generation', icon: Code, color: 'from-green-500 to-green-600' },
  'reasoning': { label: 'Reasoning', icon: Brain, color: 'from-purple-500 to-purple-600' },
  'vision': { label: 'Vision', icon: Image, color: 'from-cyan-500 to-cyan-600' },
}

type SortOption = 'name' | 'price-asc' | 'price-desc'

function ModelsPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [useCaseFilter, setUseCaseFilter] = useState('all')
  const [seriesFilter, setSeriesFilter] = useState('all')
  const [sortBy, setSortBy] = useState<SortOption>('name')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({
    categories: false, modalities: false, context: false, series: false
  })

  const { data } = useQuery({
    queryKey: ['enhanced-models'],
    queryFn: getEnhancedModels,
    staleTime: 60 * 1000,
  })
  const configuredModels: EnhancedModel[] = data?.data || []
  const configuredNames = new Set(configuredModels.map(m => m.name.toLowerCase()))

  const catalog = useMemo(() => {
    return Object.entries(modelCatalog).map(([name, cap]) => {
      const configured = configuredModels.find(m => m.name.toLowerCase() === name.toLowerCase())
      return { name, ...cap, provider: configured?.provider || cap.series.split(' ')[0],
        input_price: configured?.input_price ?? null, output_price: configured?.output_price ?? null,
        status: configured ? 1 : 0, channel_id: configured?.channel_id ?? null }
    })
  }, [configuredModels])

  const seriesNames = [...new Set(catalog.map(m => m.series.split(' ')[0]))].sort()

  const filtered = useMemo(() => {
    let result = catalog
    if (search) { const q = search.toLowerCase(); result = result.filter(m => m.name.toLowerCase().includes(q) || m.description.toLowerCase().includes(q)) }
    if (useCaseFilter !== 'all') result = result.filter(m => m.useCase === useCaseFilter)
    if (seriesFilter !== 'all') result = result.filter(m => m.series.startsWith(seriesFilter))
    switch (sortBy) { case 'price-asc': result.sort((a, b) => (a.input_price ?? 999) - (b.input_price ?? 999)); break; case 'price-desc': result.sort((a, b) => (b.input_price ?? 0) - (a.input_price ?? 0)); break; default: result.sort((a, b) => a.name.localeCompare(b.name)) }
    return result
  }, [catalog, search, useCaseFilter, seriesFilter, sortBy])

  return (
    <div className="flex min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Sidebar */}
      <aside className={cn('w-64 shrink-0 border-r bg-card/50 backdrop-blur-sm flex flex-col transition-all', sidebarOpen ? 'translate-x-0' : '-translate-x-full fixed z-10 h-full md:relative md:translate-x-0')}>
        <div className="p-4 border-b flex items-center justify-between">
          <span className="font-semibold text-sm">{t('Filters')}</span>
          <Button variant="ghost" size="icon" className="h-6 w-6 md:hidden" onClick={() => setSidebarOpen(false)}><X className="h-3.5 w-3.5" /></Button>
        </div>
        <ScrollArea className="flex-1">
          {/* Categories */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, categories: !c.categories}))} className="flex items-center justify-between w-full px-4 py-2.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider hover:bg-muted/30 transition-colors">
              <span>{t('Categories')}</span>
              {collapsed.categories ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.categories && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setUseCaseFilter('all')} className={cn('w-full text-left px-2 py-1.5 text-xs rounded transition-colors', useCaseFilter === 'all' ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{t('All Models')}</button>
              <button onClick={() => setUseCaseFilter('quantum')} className={cn('w-full text-left px-2 py-1.5 text-xs rounded transition-colors flex items-center gap-2', useCaseFilter === 'quantum' ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                <div className="w-4 h-4 rounded bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shrink-0"><Atom className="h-2.5 w-2.5 text-white" /></div>
                {t('Quantum Computing')}
              </button>
              {Object.entries(useCaseLabels).map(([k, v]) => (
                <button key={k} onClick={() => setUseCaseFilter(k)} className={cn('w-full text-left px-2 py-1.5 text-xs rounded transition-colors flex items-center gap-2', useCaseFilter === k ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                  <div className={`w-4 h-4 rounded bg-gradient-to-br ${v.color} flex items-center justify-center shrink-0`}><v.icon className="h-2.5 w-2.5 text-white" /></div>
                  {t(v.label)}
                </button>
              ))}
            </div>}
          </section>

          {/* Input Modalities */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, modalities: !c.modalities}))} className="flex items-center justify-between w-full px-4 py-2.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider hover:bg-muted/30 transition-colors">
              <span>{t('Input Modalities')}</span>
              {collapsed.modalities ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.modalities && <div className="px-4 pb-2 space-y-0.5">
              {['Text','Image','File','Audio','Video'].map(mod => (
                <button key={mod} className="w-full text-left px-2 py-1.5 text-xs text-muted-foreground hover:bg-muted/50 rounded transition-colors">{mod}</button>
              ))}
            </div>}
          </section>

          {/* Context Length */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, context: !c.context}))} className="flex items-center justify-between w-full px-4 py-2.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider hover:bg-muted/30 transition-colors">
              <span>{t('Context Length')}</span>
              {collapsed.context ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.context && <div className="px-4 pb-2 space-y-0.5">
              {[{v:'all',l:'All'},{v:'0-8192',l:'≤ 8K'},{v:'8193-32768',l:'8K - 32K'},{v:'32769-131072',l:'32K - 128K'},{v:'131073-999999999',l:'> 128K'}].map(r => (
                <button key={r.v} className="w-full text-left px-2 py-1.5 text-xs text-muted-foreground hover:bg-muted/50 rounded transition-colors">{r.l}</button>
              ))}
            </div>}
          </section>

          {/* Series */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, series: !c.series}))} className="flex items-center justify-between w-full px-4 py-2.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider hover:bg-muted/30 transition-colors">
              <span>{t('Series')}</span>
              {collapsed.series ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.series && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setSeriesFilter('all')} className={cn('w-full text-left px-2 py-1.5 text-xs rounded transition-colors', seriesFilter === 'all' ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{t('All')}</button>
              {seriesNames.map(s => (
                <button key={s} onClick={() => setSeriesFilter(s)} className={cn('w-full text-left px-2 py-1.5 text-xs rounded transition-colors', seriesFilter === s ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{s}</button>
              ))}
            </div>}
          </section>
        </ScrollArea>
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Top Bar */}
        <div className="flex items-center gap-2 px-4 py-2 border-b bg-card/50 shrink-0">
          <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => setSidebarOpen(!sidebarOpen)}>
            <SlidersHorizontal className="h-4 w-4" />
          </Button>
          <span className="text-xs text-muted-foreground font-medium">{catalog.length} {t('models')}</span>
          <div className="ml-auto flex items-center gap-2 overflow-x-auto">
            <Badge variant="outline" className="text-[10px] cursor-pointer hover:bg-muted">Text</Badge>
            <Badge variant="outline" className="text-[10px] cursor-pointer hover:bg-muted">Image</Badge>
            <Badge variant="outline" className="text-[10px] cursor-pointer hover:bg-muted">Audio</Badge>
            <Badge variant="outline" className="text-[10px] cursor-pointer hover:bg-muted">Video</Badge>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Header */}
          <div>
            <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
              {t('Models')}
            </h1>
            <p className="text-muted-foreground mt-1 text-sm">
              {t('Browse and compare models from all major providers')}
            </p>
          </div>

          {/* Search + Sort */}
          <div className="flex items-center gap-3">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input className="pl-9 h-9 text-sm" placeholder={t('Search models...')} value={search} onChange={(e) => setSearch(e.target.value)} />
            </div>
            <Select value={sortBy} onValueChange={(v) => setSortBy(v as SortOption)}>
              <SelectTrigger className="w-[130px] h-9 text-xs">
                <ArrowUpDown className="h-3.5 w-3.5 mr-1" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="name">{t('By Name')}</SelectItem>
                <SelectItem value="price-asc">{t('Price: Low to High')}</SelectItem>
                <SelectItem value="price-desc">{t('Price: High to Low')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Model Cards */}
          {filtered.length === 0 ? (
            <div className="text-center py-16 text-muted-foreground">
              <Search className="h-8 w-8 mx-auto mb-3 opacity-30" />
              <p>{t('No models found')}</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4">
              {filtered.map((m) => {
                const isConfigured = m.status === 1
                const uc = useCaseLabels[m.useCase]
                return (
                  <div key={m.name} className="rounded-xl border border-border/60 bg-card hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200 overflow-hidden p-6 space-y-4">
                    {/* Title row */}
                    <div className="flex items-center justify-between">
                        <div>
                          <h3 className="font-semibold text-lg">{m.name}</h3>
                          <p className="text-xs text-muted-foreground">{m.series}</p>
                        </div>
                        <Badge variant={isConfigured ? 'default' : 'outline'} className={cn('shrink-0 text-[10px]', isConfigured ? 'bg-emerald-500/10 text-emerald-600 border-emerald-500/30' : 'text-muted-foreground border-dashed')}>
                          {isConfigured ? t('Ready') : t('Needs Key')}
                        </Badge>
                      </div>

                      {/* Description */}
                      <p className="text-base text-muted-foreground leading-relaxed">{m.description}</p>

                      {/* Tags */}
                      <div className="flex flex-wrap items-center gap-1.5">
                        {uc && <span className={cn('inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-medium text-white bg-gradient-to-r', uc.color)}><uc.icon className="h-3 w-3" />{t(uc.label)}</span>}
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] bg-muted text-muted-foreground"><Cpu className="h-3 w-3" />{(m.contextWindow / 1000).toFixed(0)}K ctx</span>
                        {m.inputModalities.slice(0, 3).map(mod => (
                          <span key={mod} className="px-1.5 py-0.5 rounded text-[10px] bg-muted/50 text-muted-foreground">{mod}</span>
                        ))}
                      </div>

                      {/* Pricing + Action */}
                      <div className="flex items-center justify-between pt-2 border-t border-border/40">
                        <span className="text-xs text-muted-foreground">
                          {m.input_price !== null ? `From $${m.input_price.toFixed(6)}/token` : t('No pricing data')}
                        </span>
                        {isConfigured ? (
                          <Link to="/playground"><Button size="sm" className="h-7 text-xs gap-1"><Play className="h-3 w-3" />{t('Test')}</Button></Link>
                        ) : (
                          <Link to="/channels"><Button variant="outline" size="sm" className="h-7 text-xs gap-1"><Key className="h-3 w-3" />{t('Configure')}</Button></Link>
                        )}
                      </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
