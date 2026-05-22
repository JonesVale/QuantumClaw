import { createFileRoute, Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import {
  Search, X, SlidersHorizontal, ArrowUpDown, RefreshCw,
  Zap, Cpu, Brain, Image, FileText, Music, Video,
  Code, MessageSquare, Atom, CheckCircle, Key, ExternalLink,
  Play, Info, ChevronRight
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { getEnhancedModels, type EnhancedModel } from '@/lib/api-extended'

export const Route = createFileRoute('/models')({
  component: ModelsPage,
})

// ── 模型能力元数据（模型目录展示用，与渠道配置无关） ──
interface ModelCapability {
  inputModalities: string[]
  contextWindow: number
  useCase: string
  series: string
  description: string
}

const modelCatalog: Record<string, ModelCapability> = {
  'gpt-4o': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4', description: 'OpenAI flagship multimodal model with vision, high intelligence' },
  'gpt-4o-mini': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4', description: 'Fast, affordable small model for lightweight tasks' },
  'gpt-4': { inputModalities: ['Text'], contextWindow: 8192, useCase: 'chat', series: 'GPT-4', description: 'OpenAI GPT-4 base model, strong reasoning' },
  'gpt-4-turbo': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4', description: 'GPT-4 with vision and longer context' },
  'gpt-3.5-turbo': { inputModalities: ['Text'], contextWindow: 16385, useCase: 'chat', series: 'GPT-3.5', description: 'Fast, cost-effective for simple tasks' },
  'gpt-4.5': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4', description: 'Improved GPT-4 with better instruction following' },
  'gpt-5.5': { inputModalities: ['Text', 'Image'], contextWindow: 256000, useCase: 'chat', series: 'GPT-5', description: 'Next-gen OpenAI model, 2x context, lower verbosity' },
  'o1': { inputModalities: ['Text', 'Image'], contextWindow: 200000, useCase: 'reasoning', series: 'o1', description: 'OpenAI reasoning model for complex problem solving' },
  'o1-mini': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'reasoning', series: 'o1', description: 'Lightweight reasoning for math and coding' },
  'o3': { inputModalities: ['Text', 'Image'], contextWindow: 200000, useCase: 'reasoning', series: 'o3', description: 'Advanced reasoning with extended thinking' },
  'claude-3.5-sonnet': { inputModalities: ['Text', 'Image', 'File'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3.5', description: 'Best balance of speed and intelligence' },
  'claude-3.5-haiku': { inputModalities: ['Text', 'Image', 'File'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3.5', description: 'Fast and affordable for everyday tasks' },
  'claude-opus-4': { inputModalities: ['Text', 'Image', 'File'], contextWindow: 200000, useCase: 'chat', series: 'Claude', description: 'Anthropic most powerful model for complex tasks' },
  'claude-sonnet-4': { inputModalities: ['Text', 'Image', 'File'], contextWindow: 200000, useCase: 'chat', series: 'Claude', description: 'Next-gen Claude with improved capabilities' },
  'gemini-2.0-flash': { inputModalities: ['Text', 'Image', 'Audio', 'Video', 'File'], contextWindow: 1000000, useCase: 'chat', series: 'Gemini 2.0', description: 'Google fast multimodal, 1M context, supports audio/video' },
  'gemini-2.0-pro': { inputModalities: ['Text', 'Image', 'Audio', 'Video', 'File'], contextWindow: 1000000, useCase: 'chat', series: 'Gemini 2.0', description: 'Google high-quality with full multimodal support' },
  'gemini-3.0-pro': { inputModalities: ['Text', 'Image', 'Audio', 'Video', 'File'], contextWindow: 1000000, useCase: 'reasoning', series: 'Gemini 3.0', description: 'Google latest reasoning model' },
  'deepseek-chat': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'DeepSeek', description: 'DeepSeek general chat model, excellent value' },
  'deepseek-reasoner': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'reasoning', series: 'DeepSeek', description: 'DeepSeek reasoning with step-by-step thinking' },
  'deepseek-r1': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'reasoning', series: 'DeepSeek', description: 'Open-source reasoning model, top performance' },
  'deepseek-v3': { inputModalities: ['Text'], contextWindow: 64000, useCase: 'coding', series: 'DeepSeek', description: 'DeepSeek latest for code generation' },
  'qwen-max': { inputModalities: ['Text'], contextWindow: 32000, useCase: 'chat', series: 'Qwen', description: 'Alibaba strongest general model' },
  'qwen-plus': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'Qwen', description: 'Alibaba upgraded with longer context' },
  'qwen2.5-vl': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'vision', series: 'Qwen 2.5', description: 'Ali multimodal vision-language model' },
  'mistral-large': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'Mistral', description: 'Mistral AI flagship for complex tasks' },
  'mixtral-8x7b': { inputModalities: ['Text'], contextWindow: 32000, useCase: 'chat', series: 'Mistral', description: 'Mistral MoE, good quality-to-cost ratio' },
  'codestral': { inputModalities: ['Text'], contextWindow: 32000, useCase: 'coding', series: 'Mistral', description: 'Mistral dedicated code generation model' },
  'llama-3.1-70b': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'Llama 3.1', description: 'Meta open-source, strong general performance' },
  'llama-3.1-405b': { inputModalities: ['Text'], contextWindow: 128000, useCase: 'chat', series: 'Llama 3.1', description: 'Meta largest open-source, near GPT-4 level' },
  'llama-3.2-vision': { inputModalities: ['Text', 'Image'], contextWindow: 128000, useCase: 'vision', series: 'Llama 3.2', description: 'Meta multimodal open-source model' },
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
  const [ctxRange, setCtxRange] = useState('all')
  const [modFilter, setModFilter] = useState('all')
  const [sortBy, setSortBy] = useState<SortOption>('name')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [selectedModel, setSelectedModel] = useState<string | null>(null)

  // 已配置渠道的模型（有真实 API Key）
  const { data, isLoading } = useQuery({
    queryKey: ['enhanced-models'],
    queryFn: getEnhancedModels,
    staleTime: 60 * 1000,
  })
  const configuredModels: EnhancedModel[] = data?.data || []
  const configuredNames = new Set(configuredModels.map(m => m.name.toLowerCase()))

  // 构建完整模型目录：合并硬编码元数据 + 已配置渠道的定价
  const catalog = useMemo(() => {
    const entries = Object.entries(modelCatalog).map(([name, cap]) => {
      const configured = configuredModels.find(m => m.name.toLowerCase() === name.toLowerCase())
      return {
        name,
        ...cap,
        provider: configured?.provider || cap.series.split(' ')[0],
        input_price: configured?.input_price ?? null,
        output_price: configured?.output_price ?? null,
        status: configured ? 1 : 0,
        channel_id: configured?.channel_id ?? null,
      }
    })
    return entries
  }, [configuredModels])

  // Series & use case 从目录数据派生
  const series = useMemo(() => [...new Set(catalog.map(m => m.series))].sort(), [catalog])

  const seriesNames = [...new Set(catalog.map(m => {
    const p = m.series.split(' ')[0]
    return p
  }))].sort()

  // 筛选
  const filtered = useMemo(() => {
    let result = catalog
    if (search) {
      const q = search.toLowerCase()
      result = result.filter(m => m.name.toLowerCase().includes(q) || m.series.toLowerCase().includes(q) || m.description.toLowerCase().includes(q))
    }
    if (useCaseFilter !== 'all') result = result.filter(m => m.useCase === useCaseFilter)
    if (seriesFilter !== 'all') result = result.filter(m => m.series.startsWith(seriesFilter))
    // Context length
    if (ctxRange !== 'all') {
      const [min, max] = ctxRange.split('-').map(Number)
      result = result.filter(m => {
        if (max) return m.contextWindow >= min && m.contextWindow <= max
        return m.contextWindow >= min
      })
    }
    // Input modality
    if (modFilter !== 'all') {
      result = result.filter(m => m.inputModalities.some(mod => mod.toLowerCase() === modFilter.toLowerCase()))
    }
    switch (sortBy) {
      case 'price-asc': result.sort((a, b) => (a.input_price ?? 999) - (b.input_price ?? 999)); break
      case 'price-desc': result.sort((a, b) => (b.input_price ?? 0) - (a.input_price ?? 0)); break
      default: result.sort((a, b) => a.name.localeCompare(b.name)); break
    }
    return result
  }, [catalog, search, useCaseFilter, seriesFilter, sortBy, ctxRange, modFilter])

  return (
    <div className="flex min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Sidebar */}
      <aside className={cn(
        'w-64 shrink-0 border-r bg-card/50 backdrop-blur-sm flex flex-col transition-all',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full fixed z-10 h-full md:relative md:translate-x-0'
      )}>
        <div className="p-4 border-b flex items-center justify-between">
          <span className="font-semibold text-sm flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4" />
            {t('Filters')}
          </span>
          <Button variant="ghost" size="icon" className="h-6 w-6 md:hidden" onClick={() => setSidebarOpen(false)}>
            <X className="h-3.5 w-3.5" />
          </Button>
        </div>
        <ScrollArea className="flex-1 p-4 space-y-5">
          {/* Use Case */}
          <div>
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">{t('Categories')}</h4>
            <div className="space-y-1">
              <label className="flex items-center gap-2 py-1 cursor-pointer text-xs"><input type="radio" name="uc" checked={useCaseFilter === 'all'} onChange={() => setUseCaseFilter('all')} className="accent-primary" /> {t('All Models')}</label>
              {/* Quantum */}
              <label className="flex items-center gap-2 py-1 cursor-pointer text-xs hover:text-foreground transition-colors">
                <input type="radio" name="uc" checked={useCaseFilter === 'quantum'} onChange={() => setUseCaseFilter('quantum')} className="accent-primary" />
                <div className="w-4 h-4 rounded bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shrink-0">
                  <Atom className="h-2.5 w-2.5 text-white" />
                </div>
                <span className={useCaseFilter === 'quantum' ? 'font-medium text-foreground' : 'text-muted-foreground'}>{t('Quantum Computing')}</span>
              </label>
              <div className="border-t pt-1 mt-1">
              </div>
              {Object.entries(useCaseLabels).map(([k, v]) => (
                <label key={k} className="flex items-center gap-2 py-1 cursor-pointer text-xs">
                  <input type="radio" name="uc" checked={useCaseFilter === k} onChange={() => setUseCaseFilter(k)} className="accent-primary" />
                  <div className={`w-4 h-4 rounded bg-gradient-to-br ${v.color} flex items-center justify-center`}>
                    <v.icon className="h-2.5 w-2.5 text-white" />
                  </div>
                  {t(v.label)}
                </label>
              ))}
            </div>
          </div>
          {/* Series */}
          <div>
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">{t('Provider Series')}</h4>
            <div className="space-y-1">
              <label className="flex items-center gap-2 py-1 cursor-pointer text-xs"><input type="radio" name="series" checked={seriesFilter === 'all'} onChange={() => setSeriesFilter('all')} className="accent-primary" /> {t('All')}</label>
              {seriesNames.map(s => (
                <label key={s} className="flex items-center gap-2 py-1 cursor-pointer text-xs"><input type="radio" name="series" checked={seriesFilter === s} onChange={() => setSeriesFilter(s)} className="accent-primary" /> {s}</label>
              ))}
            </div>
          </div>
        </ScrollArea>
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Top Bar */}
        <div className="flex items-center gap-2 px-4 py-2 border-b bg-card/50 shrink-0">
          <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => setSidebarOpen(!sidebarOpen)}>
            <SlidersHorizontal className="h-4 w-4" />
          </Button>
          <span className="text-xs text-muted-foreground font-medium">{catalog.length} {t('models in catalog')}</span>
          <span className="text-xs text-emerald-600 dark:text-emerald-400 font-medium ml-2">
            {configuredModels.length} {t('configured')}
          </span>
        </div>

        <div className="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
          {/* Header */}
          <div>
            <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
              🧠 {t('Model Catalog')}
            </h1>
            <p className="text-muted-foreground mt-1 text-sm">
              {t('Browse models by capability. Configure your API key to use any model.')}
            </p>
          </div>

          {/* Search + Sort */}
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1 sm:max-w-md">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input className="pl-9 h-9 text-sm" placeholder={t('Search models...')} value={search} onChange={(e) => setSearch(e.target.value)} />
            </div>
            <Select value={sortBy} onValueChange={(v) => setSortBy(v as SortOption)}>
              <SelectTrigger className="w-[140px] h-9 text-xs"><ArrowUpDown className="h-3.5 w-3.5 mr-1" /><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="name">{t('By Name')}</SelectItem>
                <SelectItem value="price-asc">{t('Price: Low to High')}</SelectItem>
                <SelectItem value="price-desc">{t('Price: High to Low')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Model Cards */}
          {filtered.length === 0 ? (
            <Card><CardContent className="py-16 text-center text-muted-foreground">
              <Search className="h-8 w-8 mx-auto mb-3 opacity-30" />
              <p>{t('No models found')}</p>
            </CardContent></Card>
          ) : (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {filtered.map((m) => {
                const isConfigured = m.status === 1
                const uc = useCaseLabels[m.useCase]
                return (
                  <Card key={m.name} className={cn(
                    'hover:shadow-lg transition-all overflow-hidden group',
                    isConfigured ? 'border-emerald-500/20' : ''
                  )}>
                    {/* Provider strip */}
                    <div className={cn(
                      'h-1.5',
                      isConfigured ? 'bg-emerald-500' : 'bg-muted'
                    )} />
                    <CardContent className="p-3 space-y-2.5">
                      {/* Title + Status */}
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0">
                          <p className="text-sm font-semibold truncate group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">{m.name}</p>
                          <p className="text-[10px] text-muted-foreground">{m.series}</p>
                        </div>
                        {isConfigured ? (
                          <Badge variant="outline" className="shrink-0 text-[9px] bg-emerald-500/10 text-emerald-600 border-emerald-500/30 gap-0.5">
                            <CheckCircle className="h-2.5 w-2.5" /> {t('Ready')}
                          </Badge>
                        ) : (
                          <Badge variant="outline" className="shrink-0 text-[9px] text-muted-foreground border-dashed">
                            {t('Needs Key')}
                          </Badge>
                        )}
                      </div>

                      {/* Description */}
                      <p className="text-[10px] text-muted-foreground leading-relaxed line-clamp-2">{m.description}</p>

                      {/* Use case + context */}
                      <div className="flex flex-wrap items-center gap-1">
                        {uc && (
                          <Badge variant="secondary" className={cn('text-[9px] text-white', uc.color)}>
                            <uc.icon className="h-2.5 w-2.5 mr-0.5" />
                            {t(uc.label)}
                          </Badge>
                        )}
                        <Badge variant="outline" className="text-[9px]">
                          <Cpu className="h-2.5 w-2.5 mr-0.5" />
                          {(m.contextWindow / 1000).toFixed(0)}K ctx
                        </Badge>
                      </div>

                      {/* Input modalities */}
                      <div className="flex flex-wrap gap-1">
                        {m.inputModalities.map(mod => (
                          <span key={mod} className="text-[8px] px-1.5 py-0.5 rounded bg-muted/50 text-muted-foreground">{mod}</span>
                        ))}
                      </div>

                      {/* Pricing + Action */}
                      <div className="flex items-center justify-between pt-1">
                        <span className="text-[10px] text-muted-foreground">
                          {m.input_price !== null
                            ? `From $${m.input_price.toFixed(6)}/token`
                            : t('No pricing data')}
                        </span>
                        {isConfigured ? (
                          <Link to="/playground" className="shrink-0">
                            <Button variant="outline" size="sm" className="h-6 text-[10px] gap-1">
                              <Play className="h-3 w-3" /> {t('Test')}
                            </Button>
                          </Link>
                        ) : (
                          <Link to="/channels" className="shrink-0">
                            <Button variant="ghost" size="sm" className="h-6 text-[10px] gap-1 text-muted-foreground hover:text-foreground">
                              <Key className="h-3 w-3" /> {t('Configure')}
                            </Button>
                          </Link>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
