import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { Save, RefreshCw, Key } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import apiClient from '@/lib/api'

export const Route = createFileRoute('/_authenticated/platform-settings')({
  component: PlatformSettingsPage,
})

interface ConfigMap {
  [key: string]: string
}

function PlatformSettingsPage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [form, setForm] = useState<ConfigMap>({})
  const [saving, setSaving] = useState(false)

  // Show/hide API key text
  const [showKeys, setShowKeys] = useState<Record<string, boolean>>({})

  const { data, isLoading } = useQuery({
    queryKey: ['platform-configs'],
    queryFn: async () => {
      const res = await apiClient.get('/api/platform/config')
      return res.data?.data as ConfigMap || {}
    },
    staleTime: 30_000,
  })

  useEffect(() => {
    if (data) {
      setForm(prev => {
        const merged = { ...prev }
        for (const [k, v] of Object.entries(data)) {
          merged[k] = v || ''
        }
        return merged
      })
    }
  }, [data])

  const saveAll = async () => {
    setSaving(true)
    try {
      await apiClient.put('/api/platform/config', form)
      queryClient.invalidateQueries({ queryKey: ['platform-configs'] })
      toast.success(t('saved'))
    } catch {
      toast.error(t('save_failed'))
    } finally {
      setSaving(false)
    }
  }



  const promoFields = [
    { key: 'promotion_expire_days', label: t('promotion_expire_days'), hint: t('promotion_expire_days_hint') },
    { key: 'settlement_cycle',      label: t('settlement_cycle'),      hint: t('settlement_cycle_hint') },
    { key: 'min_withdraw_amount',   label: t('min_withdraw_amount'),   hint: t('min_withdraw_amount_hint') },
  ]

  return (
    <div className="qc-wrapper py-8 space-y-8">
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold mb-2">{t('platform_settings')}</h1>
          <p className="text-muted-foreground" style={{maxWidth: 'min(65ch, 100%)'}}>{t('platform_settings_desc')}</p>
        </div>
        <Button onClick={saveAll} disabled={saving || isLoading}>
          {saving ? <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> : <Save className="w-4 h-4 mr-2" />}
          {t('Save All')}
        </Button>
      </div>

      {/* ── Free Chat API Keys ── */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base flex items-center gap-2">
              <Key className="w-4 h-4 text-amber-500" />
              {t('Free Chat API Keys')}
            </CardTitle>
            <CardDescription>{t('Configure API keys for the free chat feature. Keys stored in database take priority over environment variables.')}</CardDescription>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowKeys({})}
            className="text-xs text-muted-foreground"
          >
            <EyeOff className="w-3.5 h-3.5 mr-1" />
            {t('Hide All')}
          </Button>
        </CardHeader>
        <CardContent className="space-y-5">
          {freeChatProviders.map((p) => {
            const isVisible = showKeys[p.key]
            const hasValue = !!form[p.key]
            return (
              <div key={p.key}>
                <div className="flex items-center justify-between mb-1.5">
                  <Label className="font-medium flex items-center gap-2">
                    {p.name}
                    {hasValue
                      ? <Badge variant="outline" className="bg-green-50 text-green-600 border-green-200 text-[10px]">
                          <Check className="w-3 h-3 mr-0.5" /> {t('Configured')}
                        </Badge>
                      : <Badge variant="outline" className="bg-muted text-muted-foreground text-[10px]">
                          <X className="w-3 h-3 mr-0.5" /> {t('Not Configured')}
                        </Badge>
                    }
                  </Label>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2 text-xs"
                    onClick={() => setShowKeys(s => ({ ...s, [p.key]: !s[p.key] }))}
                  >
                    {isVisible ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                  </Button>
                </div>
                <div className="flex gap-2">
                  <Input
                    type={isVisible ? 'text' : 'password'}
                    value={form[p.key] || ''}
                    onChange={(e) => setForm(f => ({ ...f, [p.key]: e.target.value }))}
                    placeholder={`${p.envKey} ${t('or set via admin page')}`}
                    className="font-mono text-sm"
                  />
                </div>
                <p className="text-[11px] text-muted-foreground mt-1">
                  {t('Env variable')}: <code className="text-xs bg-muted px-1 py-0.5 rounded">{p.envKey}</code>
                  {' �?'}{t('DB value takes priority')}
                </p>
              </div>
            )
          })}
        </CardContent>
      </Card>

      {/* ── Promotion & Settlement Settings ── */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="text-base">{t('promotion_settings')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {promoFields.map((field) => (
            <div key={field.key} className="space-y-1.5">
              <Label>{field.label}</Label>
              <Input
                value={form[field.key] || ''}
                onChange={(e) => setForm(f => ({ ...f, [field.key]: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground">{field.hint}</p>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
