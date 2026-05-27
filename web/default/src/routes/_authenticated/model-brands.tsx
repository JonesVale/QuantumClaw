import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Key, Check, X, RefreshCw, ExternalLink, Cpu, AlertCircle, Info } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import apiClient from '@/lib/api'

export const Route = createFileRoute('/_authenticated/model-brands')({
  component: ModelBrandsPage,
})

interface BrandModel {
  name: string; display_name: string; description: string
  context_window: number; input_price: number; output_price: number
}

interface BrandInfo {
  provider_name: string; channel_type: number; channel_name: string
  default_url: string; auto_fields: string[]; required_user: string[]
  notes: string; is_configured: boolean; channel_id: number
  existing_key: string; models: BrandModel[]; missing_fields: string[] | null
}

function ModelBrandsPage() {
  const { t } = useT()
  const qc = useQueryClient()
  const [configuring, setConfiguring] = useState<BrandInfo | null>(null)
  const [apiKey, setApiKey] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['model-brands'],
    queryFn: async () => {
      const res = await apiClient.get('/api/admin/model-brands')
      return (res.data?.data || []) as BrandInfo[]
    },
    staleTime: 30_000,
  })

  const configMutation = useMutation({
    mutationFn: async ({ provider, key }: { provider: string; key: string }) => {
      const res = await apiClient.post('/api/admin/model-brands/configure', {
        provider_name: provider, api_key: key,
      })
      return res.data
    },
    onSuccess: (res) => {
      if (res?.success) {
        toast.success(res.message || 'Configured successfully')
        setConfiguring(null)
        setApiKey('')
        qc.invalidateQueries({ queryKey: ['model-brands'] })
      } else {
        toast.error(res?.message || 'Configuration failed')
      }
    },
    onError: () => toast.error('Failed to configure provider'),
  })

  const brands = data || []

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold mb-2">{t('Model Brands')}</h1>
          <p className="text-muted-foreground" style={{maxWidth: 'min(65ch, 100%)'}}>
            {t('Configure AI model providers. Select a brand, enter your API key, and we will auto-fill the rest.')}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={() => qc.invalidateQueries({ queryKey: ['model-brands'] })}>
          <RefreshCw className="w-4 h-4 mr-2" />
          {t('Refresh')}
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {brands.map((brand) => {
          const totalModels = brand.models?.length || 0
          const aiModels = brand.models?.filter(m => !m.name.toLowerCase().includes('ionq') && !m.name.toLowerCase().includes('ibm'))?.length || 0
          const quantumModels = totalModels - aiModels
          const isQuantum = brand.channel_type >= 100

          return (
            <Card key={brand.provider_name} className={`bg-white/80 backdrop-blur-xl rounded-xl border transition-all hover:shadow-md ${brand.is_configured ? 'border-green-200' : ''}`}>
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${isQuantum ? 'bg-purple-100 text-purple-700' : 'bg-blue-100 text-blue-700'}`}>
                      <Cpu className="w-4 h-4" />
                    </div>
                    <div>
                      <CardTitle className="text-base">{brand.channel_name || brand.provider_name}</CardTitle>
                      <CardDescription className="text-xs">
                        {totalModels} {t('models')}{quantumModels > 0 ? ` · ${quantumModels} ${t('quantum')}` : ''}
                      </CardDescription>
                    </div>
                  </div>
                  {brand.is_configured ? (
                    <Badge variant="outline" className="bg-green-50 text-green-600 border-green-200 text-[10px] flex items-center gap-1">
                      <Check className="w-3 h-3" />
                      {t('Configured')}
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="bg-muted text-muted-foreground text-[10px]">
                      {t('Not Configured')}
                    </Badge>
                  )}
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                {/* API endpoint */}
                <div className="text-xs">
                  <span className="text-muted-foreground">{t('API URL')}:</span>
                  <code className="ml-1 px-1 py-0.5 bg-muted rounded text-[10px]">{brand.default_url || t('(auto)')}</code>
                </div>

                {/* Models list */}
                {brand.models && brand.models.length > 0 && (
                  <div className="space-y-1">
                    <div className="text-[10px] font-medium text-muted-foreground uppercase tracking-wider">{t('Available Models')}:</div>
                    <div className="flex flex-wrap gap-1">
                      {brand.models.slice(0, 5).map((m) => (
                        <Badge key={m.name} variant="secondary" className="text-[10px] py-0">
                          {m.display_name || m.name}
                        </Badge>
                      ))}
                      {brand.models.length > 5 && (
                        <Badge variant="outline" className="text-[10px] py-0">+{brand.models.length - 5} more</Badge>
                      )}
                    </div>
                  </div>
                )}

                {/* Missing field warnings */}
                {brand.missing_fields && brand.missing_fields.length > 0 && (
                  <div className="bg-amber-50 border border-amber-200 rounded-lg p-2 text-[10px] text-amber-700 flex items-start gap-1.5">
                    <AlertCircle className="w-3 h-3 mt-0.5 flex-shrink-0" />
                    <div>
                      {brand.missing_fields.map((f, i) => (
                        <div key={i}>{f}</div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Action button */}
                <Button
                  size="sm"
                  className="w-full"
                  variant={brand.is_configured ? 'outline' : 'default'}
                  onClick={() => { setConfiguring(brand); setApiKey('') }}
                >
                  <Key className="w-3.5 h-3.5 mr-2" />
                  {brand.is_configured ? t('Update API Key') : t('Configure')}
                </Button>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* Configure Dialog */}
      <Dialog open={!!configuring} onOpenChange={(open) => { if (!open) { setConfiguring(null); setApiKey('') } }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Key className="w-4 h-4" />
              {t('Configure')} {configuring?.channel_name || configuring?.provider_name}
            </DialogTitle>
            <DialogDescription>
              {configuring?.notes}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {/* Auto-fill info */}
            {configuring?.auto_fields && configuring.auto_fields.length > 0 && (
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 text-xs text-blue-700 space-y-1">
                <div className="font-medium">{t('Auto-filled fields')}:</div>
                {configuring.auto_fields.map((f) => (
                  <div key={f} className="flex items-center gap-1.5">
                    <Check className="w-3 h-3 text-blue-500" />
                    <span className="capitalize">{f.replace('_', ' ')}</span>
                    {f === 'base_url' && configuring.default_url && (
                      <code className="ml-1 text-[10px] bg-blue-100 px-1 rounded">{configuring.default_url}</code>
                    )}
                  </div>
                ))}
              </div>
            )}

            {/* API Key input */}
            <div className="space-y-1.5">
              <Label>{t('API Key')} <span className="text-red-500">*</span></Label>
              <Input
                type="password"
                placeholder={configuring?.is_configured ? t('Enter new API key to update...') : t('Enter your API key...')}
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                className="font-mono"
              />
              <p className="text-[10px] text-muted-foreground">
                <Info className="w-3 h-3 inline mr-0.5" />
                {t('Your API key is stored encrypted and never shared.')}
              </p>
            </div>

            {/* Missing field warnings */}
            {configuring?.missing_fields && configuring.missing_fields.length > 0 && (
              <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 text-xs text-amber-700 space-y-1">
                <div className="font-medium">{t('Additional configuration needed')}:</div>
                {configuring.missing_fields.map((f, i) => (
                  <div key={i} className="flex items-start gap-1.5">
                    <AlertCircle className="w-3 h-3 mt-0.5 flex-shrink-0" />
                    <span>{f}</span>
                  </div>
                ))}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => { setConfiguring(null); setApiKey('') }}>
              {t('Cancel')}
            </Button>
            <Button
              onClick={() => {
                if (!apiKey.trim()) { toast.error(t('API Key is required')); return }
                configMutation.mutate({ provider: configuring!.provider_name, key: apiKey.trim() })
              }}
              disabled={configMutation.isPending || !apiKey.trim()}
            >
              {configMutation.isPending ? (
                <><RefreshCw className="w-4 h-4 mr-2 animate-spin" /> {t('Configuring...')}</>
              ) : (
                <><Key className="w-4 h-4 mr-2" /> {configuring?.is_configured ? t('Update') : t('Configure')}</>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
