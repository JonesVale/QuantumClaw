import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import {
  Search, X, ChevronDown, Cpu, Brain, Image, FileText,
  Music, Video, Code, MessageSquare, Atom, SlidersHorizontal,
  ArrowUpDown, Check, ExternalLink, Zap, Globe
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Slider } from '@/components/slider'
import { cn } from '@/lib/utils'
import { getEnhancedModels, type EnhancedModel } from '@/lib/api-extended'

export const Route = createFileRoute('/models')({
  component: ModelsPage,
})

// ── Model capability metadata (hardcoded, will move to DB later) ──
type InputModality = 'text' | 'image' | 'file' | 'audio' | 'video'
type OutputType = 'text' | 'image' | 'embedding' | 'audio' | 'video' | 'rerank'
type UseCase = 'chat' | 'coding' | 'reasoning' | 'image-gen' | 'embedding' | 'vision' | 'quantum'

interface ModelCapability {
  inputModalities: InputModality[]
  outputTypes: OutputType[]
  contextWindow: number
  useCase: UseCase
  series: string
}

// Map model names to capabilities
const modelCapabilities: Record<string, ModelCapability> = {
  // OpenAI
  'gpt-4o': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4' },
  'gpt-4o-mini': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4' },
  'gpt-4': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 8192, useCase: 'chat', series: 'GPT-4' },
  'gpt-4-turbo': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4' },
  'gpt-3.5-turbo': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 16385, useCase: 'chat', series: 'GPT-3.5' },
  'o1': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 200000, useCase: 'reasoning', series: 'o1' },
  'o1-mini': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 128000, useCase: 'reasoning', series: 'o1' },
  'o3': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 200000, useCase: 'reasoning', series: 'o3' },
  'gpt-4.5': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 128000, useCase: 'chat', series: 'GPT-4' },
  'gpt-5.5': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 256000, useCase: 'chat', series: 'GPT-5' },
  'text-embedding': { inputModalities: ['text'], outputTypes: ['embedding'], contextWindow: 8192, useCase: 'embedding', series: 'Embedding' },

  // Anthropic
  'claude': { inputModalities: ['text', 'image', 'file'], outputTypes: ['text'], contextWindow: 200000, useCase: 'chat', series: 'Claude' },
  'claude-instant': { inputModalities: ['text', 'image', 'file'], outputTypes: ['text'], contextWindow: 100000, useCase: 'chat', series: 'Claude' },
  'claude-3': { inputModalities: ['text', 'image', 'file'], outputTypes: ['text'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3' },
  'claude-3-sonnet': { inputModalities: ['text', 'image', 'file'], outputTypes: ['text'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3' },
  'claude-3-haiku': { inputModalities: ['text', 'image', 'file'], outputTypes: ['text'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3' },
  'claude-3-opus': { inputModalities: ['text', 'image', 'file'], outputTypes: ['text'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3' },
  'claude-3.5-sonnet': { inputModalities: ['text', 'image', 'file'], outputTypes: ['text'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3.5' },
  'claude-3.5-haiku': { inputModalities: ['text', 'image', 'file'], outputTypes: ['text'], contextWindow: 200000, useCase: 'chat', series: 'Claude 3.5' },

  // Google
  'gemini': { inputModalities: ['text', 'image', 'audio', 'video', 'file'], outputTypes: ['text'], contextWindow: 128000, useCase: 'chat', series: 'Gemini' },
  'gemini-1.5': { inputModalities: ['text', 'image', 'audio', 'video', 'file'], outputTypes: ['text'], contextWindow: 1000000, useCase: 'chat', series: 'Gemini 1.5' },
  'gemini-2.0': { inputModalities: ['text', 'image', 'audio', 'video', 'file'], outputTypes: ['text'], contextWindow: 1000000, useCase: 'chat', series: 'Gemini 2.0' },
  'gemini-3.0': { inputModalities: ['text', 'image', 'audio', 'video', 'file'], outputTypes: ['text'], contextWindow: 1000000, useCase: 'chat', series: 'Gemini 3.0' },

  // DeepSeek
  'deepseek': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 64000, useCase: 'coding', series: 'DeepSeek' },
  'deepseek-chat': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 128000, useCase: 'chat', series: 'DeepSeek' },
  'deepseek-coder': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 128000, useCase: 'coding', series: 'DeepSeek' },
  'deepseek-r1': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 128000, useCase: 'reasoning', series: 'DeepSeek' },

  // Meta
  'llama': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 8192, useCase: 'chat', series: 'Llama' },
  'llama-2': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 4096, useCase: 'chat', series: 'Llama 2' },
  'llama-3': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 8192, useCase: 'chat', series: 'Llama 3' },

  // Mistral
  'mistral': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 32000, useCase: 'chat', series: 'Mistral' },
  'mixtral': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 32000, useCase: 'chat', series: 'Mistral' },
  'codestral': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 32000, useCase: 'coding', series: 'Mistral' },

  // Qwen
  'qwen': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 32000, useCase: 'chat', series: 'Qwen' },
  'qwen-2.5': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 128000, useCase: 'chat', series: 'Qwen 2.5' },
  'qwen3': { inputModalities: ['text', 'image'], outputTypes: ['text'], contextWindow: 128000, useCase: 'chat', series: 'Qwen 3' },

  // Groq
  'groq': { inputModalities: ['text'], outputTypes: ['text'], contextWindow: 32768, useCase: 'chat', series: 'Groq' },
}

// Use Case display info
const useCaseInfo: Record<UseCase, { label: string; icon: React.ElementType; color: string }> = {
  'chat': { label: 'Chat & Assistant', icon: MessageSquare, color: 'from-blue-500 to-blue-600' },
  'coding': { label: 'Code Generation', icon: Code, color: 'from-green-500 to-green-600' },
  'reasoning': { label: 'Reasoning', icon: Brain, color: 'from-purple-500 to-purple-600' },
  'image-gen': { label: 'Image Generation', icon: Image, color: 'from-pink-500 to-pink-600' },
  'embedding': { label: 'Embeddings', icon: FileText, color: 'from-amber-500 to-amber-600' },
  'vision': { label: 'Vision', icon: Image, color: 'from-cyan-500 to-cyan-600' },
  'quantum': { label: 'Quantum', icon: Atom, color: 'from-violet-500 to-violet-600' },
}

const inputModalityInfo: Record<InputModality, { label: string; icon: React.ElementType }> = {
  'text': { label: 'Text', icon: MessageSquare },
  'image': { label: 'Image', icon: Image },
  'file': { label: 'File', icon: FileText },
  'audio': { label: 'Audio', icon: Music },
  'video': { label: 'Video', icon: Video },
}

const outputTypeInfo: Record<OutputType, { label: string; color: string }> = {
  'text': { label: 'Text', color: 'bg-blue-500/10 text-blue-600 border-blue-500/30' },
  'image': { label: 'Image', color: 'bg-pink-500/10 text-pink-600 border-pink-500/30' },
  'embedding': { label: 'Embedding', color: 'bg-amber-500/10 text-amber-600 border-amber-500/30' },
  'audio': { label: 'Audio', color: 'bg-green-500/10 text-green-600 border-green-500/30' },
  'video': { label: 'Video', color: 'bg-red-500/10 text-red-600 border-red-500/30' },
  'rerank': { label: 'Rerank', color: 'bg-purple-500/10 text-purple-600 border-purple-500/30' },
}

const contextLengthRanges = [
  { label: '≤ 8K', max: 8192 }, { label: '8K - 32K', min: 8193, max: 32768 },
  { label: '32K - 128K', min: 32769, max: 131072 }, { label: '> 128K', min: 131073 },
]

const seriesOptions = ['GPT-4', 'GPT-3.5', 'Claude 3', 'Gemini', 'DeepSeek', 'Llama', 'Mistral', 'Qwen']

function ModelsPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [inputModalities, setInputModalities] = useState<InputModality[]>([])
  const [selectedUseCases, setSelectedUseCases] = useState<UseCase[]>([])
  const [selectedSeries, setSelectedSeries] = useState<string[]>([])
  const [contextRange, setContextRange] = useState<string>('all')
  const [selectedOutputTypes, setSelectedOutputTypes] = useState<OutputType[]>([])
  const [showInactive, setShowInactive] = useState(false)
  const [sortBy, setSortBy] = useState('provider')
  const [selectedModel, setSelectedModel] = useState<EnhancedModel | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['enhanced-models'],
    queryFn: getEnhancedModels,
    staleTime: 60 * 1000,
  })
  const models: EnhancedModel[] = data?.data || []

  const toggleFilter = <T,>(arr: T[], val: T): T[] =>
    arr.includes(val) ? arr.filter(v => v !== val) : [...arr, val]

  // Get capability for a model
  const getCapability = (name: string): ModelCapability | null => {
    const lower = name.toLowerCase()
    // Match by prefix
    for (const [prefix, cap] of Object.entries(modelCapabilities)) {
      if (lower.startsWith(prefix)) return cap
    }
    return null
  }

  // Filter models
  const filtered = useMemo(() => {
    let result = models

    if (search) {
      const q = search.toLowerCase()
      result = result.filter(m => m.name.toLowerCase().includes(q) || m.provider.toLowerCase().includes(q))
    }

    if (inputModalities.length > 0) {
      result = result.filter(m => {
        const cap = getCapability(m.name)
        return cap ? inputModalities.some(mod => cap.inputModalities.includes(mod)) : true
      })
    }

    if (selectedUseCases.length > 0) {
      result = result.filter(m => {
        const cap = getCapability(m.name)
        if (!cap) return true
        return selectedUseCases.includes(cap.useCase)
      })
    }

    if (selectedSeries.length > 0) {
      result = result.filter(m => {
        const cap = getCapability(m.name)
        return cap ? selectedSeries.includes(cap.series) : true
      })
    }

    if (contextRange !== 'all') {
      const range = contextLengthRanges.find(r => r.label === contextRange)
      if (range) {
        result = result.filter(m => {
          const cap = getCapability(m.name)
          if (!cap) return true
          if (range.min !== undefined && cap.contextWindow < range.min) return false
          if (range.max !== undefined && cap.contextWindow > range.max) return false
          return true
        })
      }
    }

    if (selectedOutputTypes.length > 0) {
      result = result.filter(m => {
        const cap = getCapability(m.name)
        return cap ? selectedOutputTypes.some(ot => cap.outputTypes.includes(ot)) : true
      })
    }

    // Sort
    switch (sortBy) {
      case 'name': result.sort((a, b) => a.name.localeCompare(b.name)); break
      case 'price-asc': result.sort((a, b) => a.input_price - b.input_price); break
      case 'price-desc': result.sort((a, b) => b.input_price - a.input_price); break
      default: result.sort((a, b) => a.provider.localeCompare(b.provider) || a.name.localeCompare(b.name)); break
    }

    return result
  }, [models, search, inputModalities, selectedUseCases, selectedSeries, contextRange, selectedOutputTypes, sortBy])

  const grouped = useMemo(() => {
    return filtered.reduce<Record<string, EnhancedModel[]>>((acc, m) => {
      const key = m.provider || 'Other'
      if (!acc[key]) acc[key] = []
      acc[key].push(m)
      return acc
    }, {})
  }, [filtered])

  const providers = useMemo(() => [...new Set(models.map(m => m.provider).filter(Boolean))].sort(), [models])

  return (
    <div className="flex min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* ── Sidebar Filters ── */}
      <aside className={cn(
        'w-64 shrink-0 border-r bg-card/50 backdrop-blur-sm flex flex-col transition-all',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full fixed z-10 h-full md:relative md:translate-x-0'
      )}>
        <div className="p-4 border-b flex items-center justify-between">
          <div className="flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4" />
            <span className="font-semibold text-sm">{t('Filters')}</span>
          </div>
          <Button variant="ghost" size="icon" className="h-6 w-6 md:hidden" onClick={() => setSidebarOpen(false)}>
            <X className="h-3.5 w-3.5" />
          </Button>
        </div>
        <ScrollArea className="flex-1 p-4 space-y-5">
          {/* Input Modalities */}
          <div>
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
              {t('Input Modalities')}
            </h4>
            {Object.entries(inputModalityInfo).map(([key, info]) => (
              <label key={key} className="flex items-center gap-2 py-1.5 cursor-pointer text-xs hover:text-foreground transition-colors">
                <input
                  type="checkbox"
                  checked={inputModalities.includes(key as InputModality)}
                  onChange={() => setInputModalities(toggleFilter(inputModalities, key as InputModality))}
                  className="rounded border-muted-foreground/30"
                />
                <info.icon className="h-3.5 w-3.5 text-muted-foreground" />
                {t(info.label)}
              </label>
            ))}
          </div>

          {/* Categories (Use Cases) */}
          <div>
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
              {t('Use Cases')}
            </h4>
            {Object.entries(useCaseInfo).map(([key, info]) => (
              <label key={key} className="flex items-center gap-2 py-1.5 cursor-pointer text-xs hover:text-foreground transition-colors">
                <input
                  type="checkbox"
                  checked={selectedUseCases.includes(key as UseCase)}
                  onChange={() => setSelectedUseCases(toggleFilter(selectedUseCases, key as UseCase))}
                  className="rounded border-muted-foreground/30"
                />
                <div className={`w-5 h-5 rounded bg-gradient-to-br ${info.color} flex items-center justify-center`}>
                  <info.icon className="h-3 w-3 text-white" />
                </div>
                {t(info.label)}
              </label>
            ))}
          </div>

          {/* Context Length */}
          <div>
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
              {t('Context Length')}
            </h4>
            <select
              value={contextRange}
              onChange={(e) => setContextRange(e.target.value)}
              className="w-full rounded-md border border-input bg-background px-2 py-1.5 text-xs"
            >
              <option value="all">{t('All')}</option>
              {contextLengthRanges.map(r => (
                <option key={r.label} value={r.label}>{r.label}</option>
              ))}
            </select>
          </div>

          {/* Series */}
          <div>
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
              {t('Series')}
            </h4>
            {seriesOptions.map(s => (
              <label key={s} className="flex items-center gap-2 py-1 cursor-pointer text-xs hover:text-foreground transition-colors">
                <input
                  type="checkbox"
                  checked={selectedSeries.includes(s)}
                  onChange={() => setSelectedSeries(toggleFilter(selectedSeries, s))}
                  className="rounded border-muted-foreground/30"
                />
                {s}
              </label>
            ))}
          </div>
        </ScrollArea>
      </aside>

      {/* ── Main Content ── */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Top Bar - Output Type Filters */}
        <div className="flex items-center gap-2 px-4 py-2 border-b bg-card/50 shrink-0 overflow-x-auto">
          <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => setSidebarOpen(!sidebarOpen)}>
            <SlidersHorizontal className="h-4 w-4" />
          </Button>
          <span className="text-xs text-muted-foreground shrink-0">{t('Output')}:</span>
          {Object.entries(outputTypeInfo).map(([key, info]) => {
            const active = selectedOutputTypes.includes(key as OutputType)
            return (
              <Badge
                key={key}
                variant="outline"
                className={cn(
                  'cursor-pointer text-[10px] shrink-0 transition-colors',
                  active ? info.color : 'hover:bg-muted'
                )}
                onClick={() => setSelectedOutputTypes(toggleFilter(selectedOutputTypes, key as OutputType))}
              >
                {active && <Check className="h-2.5 w-2.5 mr-0.5" />}
                {info.label} 0
              </Badge>
            )
          })}
        </div>

        <div className="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
          {/* Header */}
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div>
              <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
                🧠 {t('Models')}
              </h1>
              <p className="text-muted-foreground mt-1 text-sm">
                {models.length} {t('active models across')} {providers.length} {t('providers')}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" className="gap-1 h-8 text-xs">
                <ExternalLink className="h-3.5 w-3.5" />
                {t('Compare')}
              </Button>
              <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching} className="h-8 text-xs">
                {t('Refresh')}
              </Button>
            </div>
          </div>

          {/* Search + Sort */}
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1 sm:max-w-md">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="pl-9 h-9 text-sm"
                placeholder={t('Search models...')}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <Select value={sortBy} onValueChange={(v) => setSortBy(v)}>
              <SelectTrigger className="w-[140px] h-9 text-xs">
                <ArrowUpDown className="h-3.5 w-3.5 mr-1" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="provider">{t('By Provider')}</SelectItem>
                <SelectItem value="name">{t('By Name')}</SelectItem>
                <SelectItem value="price-asc">{t('Price: Low to High')}</SelectItem>
                <SelectItem value="price-desc">{t('Price: High to Low')}</SelectItem>
              </SelectContent>
            </Select>
            <span className="text-xs text-muted-foreground self-center">
              {filtered.length} / {models.length} {t('models')}
            </span>
          </div>

          {/* Model Cards */}
          {isLoading ? (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {Array.from({ length: 8 }).map((_, i) => (
                <Card key={i}><CardContent className="p-4 space-y-3">
                  <div className="h-4 w-24 animate-pulse rounded bg-muted" />
                  <div className="h-3 w-32 animate-pulse rounded bg-muted" />
                  <div className="flex gap-2"><div className="h-5 w-14 animate-pulse rounded bg-muted" /><div className="h-5 w-14 animate-pulse rounded bg-muted" /></div>
                </CardContent></Card>
              ))}
            </div>
          ) : filtered.length === 0 ? (
            <Card><CardContent className="py-16 text-center">
              <Search className="h-8 w-8 mx-auto mb-3 opacity-30" />
              <p className="text-muted-foreground">{t('No models found')}</p>
            </CardContent></Card>
          ) : (
            <div className="space-y-6">
              {Object.entries(grouped).map(([provider, providerModels]) => (
                <section key={provider}>
                  <div className="flex items-center gap-2 mb-3">
                    <h2 className="text-sm font-semibold">{provider}</h2>
                    <Badge variant="secondary" className="text-[10px]">{providerModels.length}</Badge>
                  </div>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {providerModels.map((m) => {
                      const cap = getCapability(m.name)
                      return (
                        <Card key={`${m.channel_id}-${m.name}`} className="hover:shadow-md transition-all cursor-pointer group" onClick={() => { setSelectedModel(m); setDetailOpen(true) }}>
                          <CardContent className="p-3 space-y-2">
                            <div className="flex items-center justify-between">
                              <p className="text-sm font-semibold truncate group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">{m.name}</p>
                              {cap && (
                                <Badge variant="outline" className={`text-[9px] bg-gradient-to-br ${useCaseInfo[cap.useCase].color} text-white border-0`}>
                                  {t(useCaseInfo[cap.useCase].label)}
                                </Badge>
                              )}
                            </div>
                            <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
                              {cap && (
                                <>
                                  <span>{cap.contextWindow.toLocaleString()} ctx</span>
                                  <span>·</span>
                                </>
                              )}
                              <span>${m.input_price?.toFixed(6)} / {t('token')}</span>
                            </div>
                            {cap && (
                              <div className="flex flex-wrap gap-1">
                                {cap.inputModalities.map(mod => (
                                  <span key={mod} className="text-[9px] px-1.5 py-0.5 rounded bg-muted/50 text-muted-foreground">
                                    {inputModalityInfo[mod].label}
                                  </span>
                                ))}
                              </div>
                            )}
                          </CardContent>
                        </Card>
                      )
                    })}
                  </div>
                </section>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
