import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { Save, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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

  const { data, isLoading } = useQuery({
    queryKey: ['platform-configs'],
    queryFn: async () => {
      const res = await apiClient.get('/api/platform/config')
      return res.data?.data as ConfigMap || {}
    },
    staleTime: 30_000,
  })

  useEffect(() => {
    if (data) setForm({ ...data })
  }, [data])

  const saveMut = useMutation({
    mutationFn: async (values: ConfigMap) => {
      await apiClient.put('/api/platform/config', values)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['platform-configs'] })
      toast.success(t('saved'))
    },
    onError: () => toast.error(t('save_failed')),
  })

  const configFields = [
    { key: 'promotion_expire_days', label: t('promotion_expire_days'), hint: t('promotion_expire_days_hint') },
    { key: 'settlement_cycle', label: t('settlement_cycle'), hint: t('settlement_cycle_hint') },
    { key: 'min_withdraw_amount', label: t('min_withdraw_amount'), hint: t('min_withdraw_amount_hint') },
  ]

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('platform_settings')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('platform_settings_desc')}</p>
      </div>

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="text-base">{t('promotion_settings')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {configFields.map((field) => (
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
