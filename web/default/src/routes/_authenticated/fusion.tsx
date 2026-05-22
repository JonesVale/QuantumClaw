import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useRef, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  GitCompare, Send, Loader2, Check, X, Copy, Clock, DollarSign,
  Activity, ChevronRight, Sparkles, Plus, History, Trash2,
  Search, Paperclip, Globe, Sliders, Zap
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { getModels, type ModelInfo } from '@/lib/api-extended'
import { getCommonHeaders } from '@/lib/api'

export const Route = createFileRoute('/_authenticated/fusion')({
  component: FusionPage,
})

interface FusionResult {
  model: string
  provider: string
  content: string
  latency: number
  tokenCount: number
  cost: number
  status: 'pending' | 'loading' | 'success' | 'error'
  error?: string
}

interface FusionRun {
  id: string
  prompt: string
  modelCount: number
  timestamp: number
  results: FusionResult[]
}

const PRESETS = [
  { id: 'quality', label: 'Quality', desc: 'Best overall quality', icon: Sparkles },
  { id: 'budget', label: 'Budget', desc: 'Lowest cost', icon: DollarSign },
  { id: 'custom', label: 'Custom', desc: 'Manual selection', icon: Sliders },
] as const

type PresetId = typeof PRESETS[number]['id']

function FusionPage() {
  const { t } = useTranslation()
  const [prompt, setPrompt] = useState('')
  const [selectedModels, setSelectedModels] = useState<string[]>([])
  const [results, setResults] = useState<FusionResult[]>([])
  const [running, setRunning] = useState(false)
  const [preset, setPreset] = useState<PresetId>('custom')
  const [strategy, setStrategy] = useState('auto')
  const [enableWebSearch, setEnableWebSearch] = useState(false)
  const [expandedResult, setExpandedResult] = useState<number | null>(null)
  const [history, setHistory] = useState<FusionRun[]>([])
  const [showHistory, setShowHistory] = useState(false)
  const [activeRunId, setActiveRunId] = useState<string | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [attachments, setAttachments] = useState<string[]>([])

  const { data: modelsData } = useQuery({
    queryKey: ['models'],
    queryFn: getModels,
    staleTime: 60 * 1000,
  })
  const allModels: ModelInfo[] = modelsData?.data || []
  const modelsByProvider = allModels.reduce<Record<string, ModelInfo[]>>((acc, m) => {
    const provider = (m as any).provider || m.channel_name || 'Unknown'
    if (!acc[provider]) acc[provider] = []
    acc[provider].push(m)
    return acc
  }, {})

  const topProviders = ['OpenAI', 'Anthropic', 'Google Gemini', 'DeepSeek', 'Groq', 'Mistral AI']

  // Preset model selection
  const applyPreset = (p: PresetId) => {
    setPreset(p)
    if (p === 'quality') {
      const topModels = topProviders.flatMap(prov => 
        (modelsByProvider[prov] || []).slice(0, 1).map(m => m.name)
      ).slice(0, 4)
      setSelectedModels(topModels)
      setStrategy('auto')
    } else if (p === 'budget') {
      const sorted = [...allModels].sort((a, b) => (a as any).input_price - (b as any).input_price)
      setSelectedModels(sorted.slice(0, 3).map(m => m.name))
      setStrategy('cheapest')
    }
  }

  const toggleModel = (name: string) => {
    setSelectedModels(prev =>
      prev.includes(name) ? prev.filter(m => m !== name) : [...prev, name]
    )
  }

  const handleRun = useCallback(async () => {
    if (!prompt.trim() || selectedModels.length < 1 || running) return
    setRunning(true)
    setResults([])
    setExpandedResult(null)

    const newResults: FusionResult[] = selectedModels.map(model => ({
      model,
      provider: allModels.find(m => m.name === model)?.channel_name || '',
      content: '', latency: 0, tokenCount: 0, cost: 0, status: 'loading',
    }))
    setResults(newResults)

    const headers = getCommonHeaders()
    const startTime = Date.now()

    const promises = selectedModels.map(async (model, idx) => {
      try {
        const res = await fetch('/v1/chat/completions', {
          method: 'POST',
          headers: { ...headers, 'Content-Type': 'application/json' },
          body: JSON.stringify({
            model,
            messages: [{ role: 'user', content: prompt }],
            stream: false, max_tokens: 1024,
          }),
        })
        const latency = Date.now() - startTime
        const data = await res.json()
        if (!res.ok) {
          newResults[idx] = { ...newResults[idx], status: 'error', error: data.error?.message || `HTTP ${res.status}`, latency }
          setResults([...newResults])
          return
        }
        const content = data.choices?.[0]?.message?.content || ''
        const tokens = data.usage?.total_tokens || 0
        newResults[idx] = {
          ...newResults[idx], content, latency, tokenCount: tokens,
          cost: tokens * 0.00001, status: 'success',
        }
        setResults([...newResults])
      } catch (err: any) {
        newResults[idx] = { ...newResults[idx], status: 'error', error: err.message }
        setResults([...newResults])
      }
    })
    await Promise.allSettled(promises)
    setRunning(false)

    const sorted = [...newResults].sort((a, b) => {
      if (strategy === 'fastest' || strategy === 'auto') return a.latency - b.latency
      if (strategy === 'cheapest') return a.cost - b.cost
      return 0
    })
    setResults(sorted)
    setExpandedResult(sorted.findIndex(r => r.status === 'success'))

    // Save to history
    const run: FusionRun = {
      id: Date.now().toString(36),
      prompt: prompt.slice(0, 100),
      modelCount: selectedModels.length,
      timestamp: Date.now(),
      results: sorted,
    }
    setHistory(prev => [run, ...prev].slice(0, 20))
    setShowHistory(true)
  }, [prompt, selectedModels, running, strategy, allModels])

  const copyResult = (content: string) => navigator.clipboard.writeText(content)

  const clearHistory = () => { setHistory([]); setActiveRunId(null); setResults([]) }
  const newFusion = () => { setPrompt(''); setResults([]); setActiveRunId(null); setSelectedModels([]) }
  const loadRun = (run: FusionRun) => { setResults(run.results); setActiveRunId(run.id); setShowHistory(false) }

  const handleFileUpload = () => fileInputRef.current?.click()
  const bestResult = results.find(r => r.status === 'success')
  const topModelNames = topProviders.flatMap(prov => 
    (modelsByProvider[prov] || []).slice(0, 1).map(m => m.name)
  ).slice(0, 3)

  return (
    <div className="flex h-full overflow-hidden bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* ── History Sidebar ── */}
      <div className={cn(
        'w-64 shrink-0 border-r bg-card/50 backdrop-blur-sm flex flex-col transition-all',
        showHistory ? 'translate-x-0' : '-translate-x-full fixed z-10 h-full md:relative md:translate-x-0'
      )}>
        <div className="p-3 border-b">
          <Button onClick={newFusion} className="w-full gap-2" size="sm">
            <Plus className="h-4 w-4" /> {t('New Fusion')}
          </Button>
        </div>
        <ScrollArea className="flex-1 p-2">
          {history.length === 0 ? (
            <p className="text-xs text-muted-foreground text-center py-8">{t('No runs yet')}</p>
          ) : (
            <div className="space-y-1">
              {history.map((run) => (
                <div
                  key={run.id}
                  onClick={() => loadRun(run)}
                  className={cn(
                    'p-2 rounded-lg cursor-pointer text-xs transition-colors',
                    activeRunId === run.id ? 'bg-accent' : 'hover:bg-muted/50'
                  )}
                >
                  <p className="font-medium truncate">{run.prompt || t('Untitled')}</p>
                  <p className="text-muted-foreground mt-0.5">
                    {run.modelCount} {t('models')} · {new Date(run.timestamp).toLocaleTimeString()}
                  </p>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
        {history.length > 0 && (
          <div className="p-2 border-t">
            <Button variant="ghost" size="sm" className="w-full text-xs" onClick={clearHistory}>
              <Trash2 className="h-3 w-3 mr-1" /> {t('Clear history')}
            </Button>
          </div>
        )}
      </div>

      {/* ── Main Content ── */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Top Bar */}
        <div className="flex items-center gap-2 px-4 py-2 border-b bg-card/50 shrink-0">
          <Button variant="ghost" size="icon" className="h-7 w-7 md:hidden" onClick={() => setShowHistory(!showHistory)}>
            <History className="h-4 w-4" />
          </Button>
          <h1 className="text-base font-bold tracking-tight bg-gradient-to-r from-emerald-500 via-teal-500 to-cyan-600 bg-clip-text text-transparent flex items-center gap-2">
            <GitCompare className="h-5 w-5 text-emerald-500" />
            {t('Model Fusion')}
            <Badge variant="outline" className="text-[10px] font-normal ml-1">{t('beta')}</Badge>
          </h1>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-4 max-w-5xl mx-auto w-full">
          {/* Description */}
          <p className="text-sm text-muted-foreground">
            {t('Run multiple models side-by-side, run an analysis, and fuse into the best result.')}
          </p>

          {/* Presets */}
          <div className="flex gap-2">
            {PRESETS.map((p) => (
              <Button
                key={p.id}
                variant={preset === p.id ? 'default' : 'outline'}
                size="sm"
                onClick={() => applyPreset(p.id)}
                className="gap-1.5"
              >
                <p.icon className="h-3.5 w-3.5" />
                {t(p.label)}
              </Button>
            ))}
          </div>

          {/* Model Selection */}
          <Card>
            <CardContent className="p-3 space-y-2">
              {topModelNames.map((modelName, idx) => {
                const selected = selectedModels.includes(modelName)
                return (
                  <div key={modelName} className="flex items-center gap-2">
                    <div
                      onClick={() => toggleModel(modelName)}
                      className={cn(
                        'flex-1 flex items-center gap-3 px-3 py-2 rounded-lg cursor-pointer text-sm transition-colors',
                        selected ? 'bg-emerald-500/10 border border-emerald-500/30' : 'bg-muted/30 hover:bg-muted/50'
                      )}
                    >
                      <div className={cn(
                        'w-5 h-5 rounded border-2 flex items-center justify-center shrink-0',
                        selected ? 'bg-emerald-500 border-emerald-500' : 'border-muted-foreground/30'
                      )}>
                        {selected && <Check className="h-3 w-3 text-white" />}
                      </div>
                      <div className="flex-1">
                        <p className="text-sm font-medium">{modelName}</p>
                        <p className="text-xs text-muted-foreground">{topProviders[idx]}</p>
                      </div>
                      <Badge variant="outline" className="text-[10px]">
                        {(allModels.find(m => m.name === modelName) as any)?.input_price?.toFixed(6) || '?'}
                      </Badge>
                    </div>
                  </div>
                )
              })}

              {/* Add Model Button & Strategy */}
              <div className="flex items-center gap-2 pt-1">
                <Select value={strategy} onValueChange={setStrategy}>
                  <SelectTrigger className="w-auto h-8 text-xs gap-1 border-0 bg-muted/50">
                    <Zap className="h-3 w-3" />
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="auto">{t('Auto (first source)')}</SelectItem>
                    <SelectItem value="fastest">{t('Fastest')}</SelectItem>
                    <SelectItem value="cheapest">{t('Cheapest')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          {/* Prompt Input */}
          <Card>
            <CardContent className="p-3 space-y-3">
              <Textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                placeholder={t('Ask anything...')}
                className="min-h-[100px] resize-none border-0 shadow-none p-0 focus-visible:ring-0"
                rows={3}
              />
              <div className="flex items-center gap-2 pt-2 border-t">
                <Button variant="ghost" size="sm" onClick={handleFileUpload} className="gap-1.5 text-xs h-8">
                  <Paperclip className="h-3.5 w-3.5" /> {t('Add attachment')}
                </Button>
                <input ref={fileInputRef} type="file" className="hidden" multiple onChange={(e) => {
                  const files = Array.from(e.target.files || []).map(f => f.name)
                  setAttachments(prev => [...prev, ...files])
                }} />
                <label className="flex items-center gap-2 text-xs cursor-pointer ml-auto">
                  <Switch checked={enableWebSearch} onCheckedChange={setEnableWebSearch} className="h-4 w-7" />
                  <Globe className="h-3.5 w-3.5 text-muted-foreground" />
                  {t('Enable Web Search')}
                </label>
                <Button
                  onClick={handleRun}
                  disabled={selectedModels.length < 1 || !prompt.trim() || running}
                  size="sm"
                  className="gap-1.5"
                >
                  {running ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Send className="h-3.5 w-3.5" />}
                  {t('Run Fusion')}
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Results */}
          {results.length > 0 && (
            <div className="space-y-2" ref={scrollRef as any}>
              {/* Best Result Banner */}
              {bestResult && (
                <Card className="border-emerald-500/30 bg-emerald-500/5">
                  <CardContent className="p-3 flex items-center gap-3">
                    <Sparkles className="h-4 w-4 text-emerald-500 shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                        {t('Best Result')}: {bestResult.model}
                      </p>
                      <p className="text-[10px] text-muted-foreground">
                        {bestResult.latency}ms · {bestResult.tokenCount}t · ~${bestResult.cost.toFixed(6)}
                      </p>
                    </div>
                    <Button variant="outline" size="sm" onClick={() => copyResult(bestResult.content)} className="gap-1 h-7 text-xs">
                      <Copy className="h-3 w-3" /> {t('Copy')}
                    </Button>
                  </CardContent>
                </Card>
              )}

              {/* Result Cards */}
              {results.map((result, idx) => (
                <Card key={idx} className={cn(
                  'transition-all',
                  bestResult === result && 'ring-1 ring-emerald-500/30'
                )}>
                  <div
                    className="p-3 flex items-center justify-between cursor-pointer"
                    onClick={() => setExpandedResult(expandedResult === idx ? null : idx)}
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <Badge variant="outline" className={cn(
                        'text-[10px] shrink-0',
                        result.status === 'success' && 'bg-emerald-500/10 text-emerald-600 border-emerald-500/30',
                        result.status === 'error' && 'bg-red-500/10 text-red-600 border-red-500/30',
                      )}>
                        {result.status === 'loading' && <Loader2 className="h-3 w-3 animate-spin" />}
                        {result.status === 'success' && `${result.latency}ms`}
                        {result.status === 'error' && 'Error'}
                      </Badge>
                      <span className="text-sm font-medium truncate">{result.model}</span>
                      <span className="text-xs text-muted-foreground hidden sm:inline">{result.provider}</span>
                    </div>
                    {result.status === 'success' && (
                      <div className="flex items-center gap-3 text-[10px] text-muted-foreground shrink-0">
                        <span className="hidden sm:inline">{result.tokenCount}t</span>
                        <span className="hidden sm:inline">${result.cost.toFixed(6)}</span>
                        <ChevronRight className={cn('h-3.5 w-3.5 transition-transform', expandedResult === idx && 'rotate-90')} />
                      </div>
                    )}
                  </div>
                  {expandedResult === idx && result.status === 'success' && (
                    <div className="px-3 pb-3">
                      <ScrollArea className="h-[200px] rounded-lg bg-muted/50 p-3">
                        <pre className="text-xs whitespace-pre-wrap font-sans leading-relaxed">{result.content}</pre>
                      </ScrollArea>
                      <div className="flex justify-end mt-1">
                        <Button variant="ghost" size="sm" onClick={() => copyResult(result.content)} className="gap-1 h-6 text-[10px]">
                          <Copy className="h-3 w-3" /> {t('Copy')}
                        </Button>
                      </div>
                    </div>
                  )}
                  {expandedResult === idx && result.status === 'error' && (
                    <div className="px-3 pb-3">
                      <p className="text-xs text-red-500">{result.error}</p>
                    </div>
                  )}
                </Card>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
