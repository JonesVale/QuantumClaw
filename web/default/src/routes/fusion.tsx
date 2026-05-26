import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { handleFusion, getEnhancedModels, type EnhancedModel } from '@/lib/api-extended'
import { GitCompare, Send, Loader2, RotateCcw, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'

export const Route = createFileRoute('/fusion')({
  component: FusionPage,
})

function FusionPage() {
  const { t } = useT()
  const [modelA, setModelA] = useState('')
  const [modelB, setModelB] = useState('')
  const [prompt, setPrompt] = useState('')
  const [result, setResult] = useState<{ model_a: string; model_b: string; result: string } | null>(null)
  const [loading, setLoading] = useState(false)

  const { data: modelsData } = useQuery({
    queryKey: ['fusion-models'],
    queryFn: getEnhancedModels,
    staleTime: 60_000,
  })

  const models = (modelsData?.data || []) as EnhancedModel[]
  const modelNames = [...new Set(models.map(m => m.name))].sort()

  const runFusion = async () => {
    if (!modelA || !modelB || !prompt.trim()) {
      toast.error(t('Please select two models and enter a prompt'))
      return
    }
    setLoading(true)
    setResult(null)
    try {
      const res = await handleFusion({ model_a: modelA, model_b: modelB, prompt: prompt.trim() })
      if (res.success) {
        setResult(res.data as any)
        toast.success(t('Fusion complete'))
      } else {
        toast.error(res.message || t('Fusion failed'))
      }
    } catch {
      toast.error(t('Fusion failed'))
    }
    setLoading(false)
  }

  const swapModels = () => {
    const tmp = modelA
    setModelA(modelB)
    setModelB(tmp)
  }

  return (
    <div className="min-h-screen bg-background py-8"
      style={{ backgroundImage: 'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)' }}>
      <div className="qc-wrapper space-y-6">
        <div className="text-center">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-amber-200 bg-amber-50 text-amber-700 text-xs font-semibold tracking-wide mb-4">
            <GitCompare className="w-3 h-3" />
            {t('Model Fusion')}
          </div>
          <h1 className="text-3xl font-bold mb-2">{t('Model Fusion')}</h1>
          <p className="text-muted-foreground mb-6">{t('Compare and combine two AI models')}</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* ── Controls ── */}
          <Card>
            <CardHeader>
              <CardTitle>{t('Configure Fusion')}</CardTitle>
              <CardDescription>{t('Select models and enter your prompt')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-[1fr,auto,1fr] gap-2 items-end">
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-1 block">{t('Model A')}</label>
                  <select value={modelA} onChange={e => setModelA(e.target.value)}
                    className="w-full h-10 rounded-lg border border-border bg-background px-3 text-sm focus:border-amber-500 outline-none">
                    <option value="">{t('Select model...')}</option>
                    {modelNames.map(m => <option key={m} value={m}>{m}</option>)}
                  </select>
                </div>
                <Button variant="outline" size="sm" onClick={swapModels} className="mb-0.5" title={t('Swap')}>
                  <RotateCcw className="h-4 w-4" />
                </Button>
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-1 block">{t('Model B')}</label>
                  <select value={modelB} onChange={e => setModelB(e.target.value)}
                    className="w-full h-10 rounded-lg border border-border bg-background px-3 text-sm focus:border-amber-500 outline-none">
                    <option value="">{t('Select model...')}</option>
                    {modelNames.map(m => <option key={m} value={m}>{m}</option>)}
                  </select>
                </div>
              </div>

              <div>
                <label className="text-xs font-medium text-muted-foreground mb-1 block">{t('Prompt')}</label>
                <textarea value={prompt} onChange={e => setPrompt(e.target.value)}
                  placeholder={t('Enter your prompt for both models...')}
                  className="w-full min-h-[120px] rounded-lg border border-border bg-background p-3 text-sm resize-y focus:border-amber-500 outline-none"
                />
              </div>

              <Button onClick={runFusion} disabled={loading || !modelA || !modelB || !prompt.trim()} className="w-full">
                {loading ? <><Loader2 className="h-4 w-4 animate-spin mr-2" />{t('Running...')}</> : <><Send className="h-4 w-4 mr-2" />{t('Run Fusion')}</>}
              </Button>
            </CardContent>
          </Card>

          {/* ── Results ── */}
          <Card>
            <CardHeader>
              <CardTitle>{t('Results')}</CardTitle>
              <CardDescription>{t('Model outputs will appear here')}</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                  <Loader2 className="h-8 w-8 animate-spin mb-4" />
                  <p className="text-sm">{t('Running models in parallel...')}</p>
                </div>
              ) : result ? (
                <div className="space-y-4">
                  <div className="flex items-center gap-2">
                    <Badge variant="default">{result.model_a}</Badge>
                    <Badge variant="secondary">{result.model_b}</Badge>
                    <span className="text-xs text-muted-foreground">{t('Comparison')}</span>
                  </div>
                  <div className="rounded-lg border bg-muted/30 p-4">
                    <pre className="text-sm whitespace-pre-wrap break-words font-sans">{result.result}</pre>
                  </div>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                  <GitCompare className="h-12 w-12 mb-4 opacity-20" />
                  <p className="text-sm font-medium">{t('No results yet')}</p>
                  <p className="text-xs mt-1">{t('Configure models and run fusion')}</p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
