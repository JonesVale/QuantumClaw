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
    <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">{t('platform_settings')}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t('platform_settings_desc')}</p>
        </div>
        <Button onClick={() => saveMut.mutate(form)} disabled={saveMut.isPending} className="gap-2">
          <Save className="h-4 w-4" /> {t('save')}
        </Button>
      </div>

      <Card>
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
