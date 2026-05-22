import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useRef, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  GitCompare, Send, Loader2, Check, X, Copy, Clock, DollarSign,
  Activity, ChevronDown, ChevronUp, Info, Zap, Sparkles
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
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

function FusionPage() {
  const { t } = useTranslation()
  const [prompt, setPrompt] = useState('')
  const [selectedModels, setSelectedModels] = useState<string[]>([])
  const [results, setResults] = useState<FusionResult[]>([])
  const [running, setRunning] = useState(false)
  const [strategy, setStrategy] = useState('fastest')
  const [expandedResult, setExpandedResult] = useState<number | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  const { data: modelsData } = useQuery({
    queryKey: ['models'],
    queryFn: getModels,
    staleTime: 60 * 1000,
  })
  const allModels: ModelInfo[] = modelsData?.data || []

  // Group models by provider for the multi-select UI
  const modelsByProvider = allModels.reduce<Record<string, ModelInfo[]>>((acc, m) => {
    const provider = (m as any).provider || m.channel_name || 'Unknown'
    if (!acc[provider]) acc[provider] = []
    acc[provider].push(m)
    return acc
  }, {})

  const toggleModel = (name: string) => {
    setSelectedModels(prev =>
      prev.includes(name) ? prev.filter(m => m !== name) : [...prev, name]
    )
  }

  const handleRun = useCallback(async () => {
    if (!prompt.trim() || selectedModels.length < 2 || running) return

    setRunning(true)
    const newResults: FusionResult[] = selectedModels.map(model => ({
      model,
      provider: allModels.find(m => m.name === model)?.channel_name || '',
      content: '',
      latency: 0,
      tokenCount: 0,
      cost: 0,
      status: 'loading' as const,
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
            stream: false,
            max_tokens: 1024,
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
          ...newResults[idx],
          content,
          latency,
          tokenCount: tokens,
          cost: tokens * 0.00001, // simplified cost calc
          status: 'success',
        }
        setResults([...newResults])
      } catch (err: any) {
        newResults[idx] = { ...newResults[idx], status: 'error', error: err.message }
        setResults([...newResults])
      }
    })

    await Promise.allSettled(promises)
    setRunning(false)

    // Sort by strategy
    const sorted = [...newResults].sort((a, b) => {
      if (strategy === 'fastest') return a.latency - b.latency
      if (strategy === 'cheapest') return a.cost - b.cost
      return 0
    })
    setResults(sorted)
    setExpandedResult(sorted.findIndex(r => r.status === 'success'))
  }, [prompt, selectedModels, running, strategy, allModels])

  const copyResult = (content: string) => {
    navigator.clipboard.writeText(content)
  }

  const bestResult = results.find(r => r.status === 'success')

  return (
    <div className="p-4 sm:p-6 space-y-6 w-full min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-emerald-500 via-teal-500 to-cyan-600 bg-clip-text text-transparent flex items-center gap-3">
            <GitCompare className="h-8 w-8 text-emerald-500" />
            {t('Fusion')}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm sm:text-base">
            {t('Compare models side by side. Send one prompt to multiple models and choose the best result.')}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left: Configuration */}
        <div className="lg:col-span-1 space-y-4">
          {/* Strategy */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm flex items-center gap-2">
                <Zap className="h-4 w-4 text-emerald-500" />
                {t('Routing Strategy')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Select value={strategy} onValueChange={setStrategy}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="fastest">{t('Fastest Response')}</SelectItem>
                  <SelectItem value="cheapest">{t('Cheapest')}</SelectItem>
                  <SelectItem value="all">{t('Show All')}</SelectItem>
                </SelectContent>
              </Select>
            </CardContent>
          </Card>

          {/* Model Selection */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm flex items-center gap-2">
                <GitCompare className="h-4 w-4 text-emerald-500" />
                {t('Select Models')} ({selectedModels.length})
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-[300px] pr-4">
                <div className="space-y-3">
                  {Object.entries(modelsByProvider).map(([provider, models]) => (
                    <div key={provider}>
                      <p className="text-xs font-semibold text-muted-foreground mb-1 uppercase tracking-wider">{provider}</p>
                      {models.map(m => {
                        const selected = selectedModels.includes(m.name)
                        return (
                          <div
                            key={m.name}
                            onClick={() => toggleModel(m.name)}
                            className={cn(
                              'flex items-center gap-2 px-2 py-1.5 rounded-md cursor-pointer text-xs transition-colors mb-0.5',
                              selected ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-medium' : 'hover:bg-muted/50'
                            )}
                          >
                            <div className={cn(
                              'w-4 h-4 rounded border flex items-center justify-center flex-shrink-0',
                              selected ? 'bg-emerald-500 border-emerald-500' : 'border-muted-foreground/30'
                            )}>
                              {selected && <Check className="h-3 w-3 text-white" />}
                            </div>
                            {m.name}
                          </div>
                        )
                      })}
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>

          {/* Run Button */}
          <Button
            onClick={handleRun}
            disabled={selectedModels.length < 2 || !prompt.trim() || running}
            className="w-full gap-2"
            size="lg"
          >
            {running ? (
              <><Loader2 className="h-4 w-4 animate-spin" />{t('Running...')}</>
            ) : (
              <><Send className="h-4 w-4" />{t('Run Comparison')} ({selectedModels.length} {t('models')})</>
            )}
          </Button>
        </div>

        {/* Right: Prompt + Results */}
        <div className="lg:col-span-2 space-y-4">
          {/* Prompt Input */}
          <Card>
            <CardContent className="p-4">
              <Textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                placeholder={t('Enter your prompt to compare across selected models...')}
                className="min-h-[120px] resize-none"
                rows={4}
              />
            </CardContent>
          </Card>

          {/* Results */}
          {results.length > 0 && (
            <div className="space-y-3" ref={scrollRef as any}>
              {/* Best result banner */}
              {bestResult && strategy !== 'all' && (
                <Card className="border-emerald-500/30 bg-emerald-500/5">
                  <CardContent className="p-4 flex items-center gap-3">
                    <Sparkles className="h-5 w-5 text-emerald-500 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-semibold text-emerald-600 dark:text-emerald-400">
                        {t('Best Result')}: {bestResult.model}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {bestResult.latency}ms · {bestResult.tokenCount} tokens · ~${bestResult.cost.toFixed(6)}
                      </p>
                    </div>
                    <Button variant="outline" size="sm" onClick={() => copyResult(bestResult.content)} className="gap-1">
                      <Copy className="h-3 w-3" /> {t('Copy')}
                    </Button>
                  </CardContent>
                </Card>
              )}

              {/* Result cards */}
              {results.map((result, idx) => (
                <Card key={idx} className={cn(
                  'transition-all',
                  bestResult === result && strategy !== 'all' && 'ring-1 ring-emerald-500/30',
                )}>
                  <CardHeader className="pb-2 cursor-pointer" onClick={() => setExpandedResult(expandedResult === idx ? null : idx)}>
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Badge variant={result.status === 'success' ? 'default' : result.status === 'error' ? 'destructive' : 'secondary'}
                          className={cn(
                            'text-[10px]',
                            result.status === 'success' && 'bg-emerald-500/20 text-emerald-600 border-emerald-500/30',
                            result.status === 'error' && 'bg-red-500/20 text-red-600 border-red-500/30',
                          )}
                        >
                          {result.status === 'loading' && <Loader2 className="h-3 w-3 animate-spin mr-1" />}
                          {result.status === 'success' && <Check className="h-3 w-3 mr-1" />}
                          {result.status === 'error' && <X className="h-3 w-3 mr-1" />}
                          {result.status === 'success' ? `${result.latency}ms` : result.status}
                        </Badge>
                        <span className="text-sm font-semibold">{result.model}</span>
                        <span className="text-xs text-muted-foreground">{result.provider}</span>
                      </div>
                      <div className="flex items-center gap-3 text-xs text-muted-foreground">
                        <span className="flex items-center gap-1"><Clock className="h-3 w-3" />{result.latency}ms</span>
                        <span className="flex items-center gap-1"><Activity className="h-3 w-3" />{result.tokenCount}t</span>
                        <span className="flex items-center gap-1"><DollarSign className="h-3 w-3" />${result.cost.toFixed(6)}</span>
                        {expandedResult === idx ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                      </div>
                    </div>
                  </CardHeader>
                  {expandedResult === idx && (
                    <CardContent className="pt-2">
                      <ScrollArea className="h-[200px] rounded-lg bg-muted/50 p-3">
                        {result.status === 'success' ? (
                          <pre className="text-xs whitespace-pre-wrap font-sans leading-relaxed">{result.content}</pre>
                        ) : result.status === 'error' ? (
                          <p className="text-xs text-red-500">{result.error}</p>
                        ) : (
                          <div className="flex items-center justify-center h-full text-muted-foreground">
                            <Loader2 className="h-5 w-5 animate-spin" />
                          </div>
                        )}
                      </ScrollArea>
                      <div className="flex justify-end mt-2">
                        <Button variant="ghost" size="sm" onClick={() => copyResult(result.content)} className="gap-1 h-7 text-xs">
                          <Copy className="h-3 w-3" /> {t('Copy')}
                        </Button>
                      </div>
                    </CardContent>
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
