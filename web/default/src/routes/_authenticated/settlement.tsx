import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Plus, RefreshCw, Pencil, Trash2, DollarSign, Percent, Globe, Banknote, Save } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import apiClient from '@/lib/api'
import {
  getSettlementConfigs,
  createSettlementConfig,
  updateSettlementConfig,
  deleteSettlementConfig,
  type SettlementConfigItem,
} from '@/lib/api-extended'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'

export const Route = createFileRoute('/_authenticated/settlement')({
  component: SettlementPage,
})

function SettlementPage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<SettlementConfigItem | null>(null)
  const [form, setForm] = useState({ model_name: '', unified_cost: '', commission_rate: '', platform_fee_rate: '' })

  const { data, isLoading } = useQuery({
    queryKey: ['settlement-configs'],
    queryFn: () => getSettlementConfigs(),
    staleTime: 30_000,
  })
  const configs = data?.data || []

  const createMut = useMutation({
    mutationFn: createSettlementConfig,
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['settlement-configs'] }); setDialogOpen(false); resetForm(); toast.success(t('saved')) },
    onError: () => toast.error(t('save_failed')),
  })

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<SettlementConfigItem> }) => updateSettlementConfig(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['settlement-configs'] }); setDialogOpen(false); setEditing(null); resetForm(); toast.success(t('saved')) },
    onError: () => toast.error(t('save_failed')),
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => deleteSettlementConfig(id),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['settlement-configs'] }); toast.success(t('deleted')) },
    onError: () => toast.error(t('delete_failed')),
  })

  const resetForm = () => setForm({ model_name: '', unified_cost: '', commission_rate: '', platform_fee_rate: '' })

  const openCreate = () => { setEditing(null); resetForm(); setDialogOpen(true) }
  const openEdit = (cfg: SettlementConfigItem) => {
    setEditing(cfg)
    setForm({
      model_name: cfg.model_name,
      unified_cost: cfg.unified_cost.toString(),
      commission_rate: (cfg.commission_rate * 100).toString(),
      platform_fee_rate: (cfg.platform_fee_rate * 100).toString(),
    })
    setDialogOpen(true)
  }

  const handleSave = () => {
    const payload = {
      model_name: form.model_name,
      unified_cost: parseFloat(form.unified_cost) || 0.001,
      commission_rate: (parseFloat(form.commission_rate) || 20) / 100,
      platform_fee_rate: (parseFloat(form.platform_fee_rate) || 10) / 100,
    }
    if (editing) {
      updateMut.mutate({ id: editing.id, data: payload })
    } else {
      createMut.mutate(payload)
    }
  }

  // 交易手续费状态
  const [feeForm, setFeeForm] = useState({ domestic: '1.0', foreign: '3.0', foreignMin: '5.00' })
  const [feeSaving, setFeeSaving] = useState(false)

  const { data: platformCfg } = useQuery({
    queryKey: ['platform-configs'],
    queryFn: async () => {
      const res = await apiClient.get('/api/platform/config')
      return res.data?.data || {}
    },
    staleTime: 30_000,
  })

  // Sync fee form when platform config loads
  useState(() => {
    if (platformCfg) {
      setFeeForm({
        domestic: platformCfg['transaction_fee_domestic'] || '1.0',
        foreign: platformCfg['transaction_fee_foreign'] || '3.0',
        foreignMin: platformCfg['transaction_fee_foreign_min'] || '5.00',
      })
    }
  })

  const saveFees = async () => {
    setFeeSaving(true)
    try {
      await apiClient.put('/api/platform/config', {
        'transaction_fee_domestic': feeForm.domestic,
        'transaction_fee_foreign': feeForm.foreign,
        'transaction_fee_foreign_min': feeForm.foreignMin,
      })
      queryClient.invalidateQueries({ queryKey: ['platform-configs'] })
      toast.success(t('saved'))
    } catch {
      toast.error(t('save_failed'))
    } finally {
      setFeeSaving(false)
    }
  }

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('settlement_config')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('settlement_config_desc')}</p>
      </div>

      {/* 交易手续费配置 */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base flex items-center gap-2">
              <Percent className="w-4 h-4 text-amber-500" />
              {t('Transaction Fee')}
            </CardTitle>
            <CardDescription>{t('Transaction fee deducted from each API call. Domestic (China) vs Foreign models use different rates.')}</CardDescription>
          </div>
          <Button onClick={saveFees} disabled={feeSaving} size="sm">
            <Save className="w-4 h-4 mr-2" />
            {t('Save')}
          </Button>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="space-y-2">
              <Label className="flex items-center gap-2">
                <Banknote className="w-4 h-4 text-green-500" />
                {t('Domestic Models')}
              </Label>
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  step="0.1"
                  min="0"
                  max="100"
                  value={feeForm.domestic}
                  onChange={(e) => setFeeForm(f => ({ ...f, domestic: e.target.value }))}
                  className="w-24"
                />
                <span className="text-sm text-muted-foreground">%</span>
              </div>
              <p className="text-xs text-muted-foreground">{t("Baidu, Ali, DeepSeek, Zhipu, Tencent...")}</p>
            </div>
            <div className="space-y-2">
              <Label className="flex items-center gap-2">
                <Globe className="w-4 h-4 text-blue-500" />
                {t('Foreign Models')}
              </Label>
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  step="0.1"
                  min="0"
                  max="100"
                  value={feeForm.foreign}
                  onChange={(e) => setFeeForm(f => ({ ...f, foreign: e.target.value }))}
                  className="w-24"
                />
                <span className="text-sm text-muted-foreground">%</span>
              </div>
              <p className="text-xs text-muted-foreground">OpenAI, Anthropic, Gemini, Groq...</p>
            </div>
            <div className="space-y-2">
              <Label className="flex items-center gap-2">
                <DollarSign className="w-4 h-4 text-amber-500" />
                {t('Foreign Min Fee')}
              </Label>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">$</span>
                <Input
                  type="number"
                  step="0.5"
                  min="0"
                  value={feeForm.foreignMin}
                  onChange={(e) => setFeeForm(f => ({ ...f, foreignMin: e.target.value }))}
                  className="w-24"
                />
              </div>
              <p className="text-xs text-muted-foreground">{t('Minimum fee per transaction for foreign models')}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Separator />

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('model')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('unified_cost')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('commission_rate')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('platform_fee_rate')}</th>
                  <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('status')}</th>
                  <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('actions')}</th>
                </tr>
              </thead>
              <tbody>
                {configs.map((cfg) => (
                  <tr key={cfg.id} className="border-b border-muted/50 hover:bg-muted/30">
                    <td className="px-4 py-3"><span className="font-medium">{cfg.model_name}</span></td>
                    <td className="px-4 py-3 text-right font-mono text-sm">${cfg.unified_cost.toFixed(6)}</td>
                    <td className="px-4 py-3 text-right">{(cfg.commission_rate * 100).toFixed(0)}%</td>
                    <td className="px-4 py-3 text-right">{(cfg.platform_fee_rate * 100).toFixed(0)}%</td>
                    <td className="px-4 py-3 text-center">
                      <Badge variant={cfg.enabled ? 'default' : 'secondary'}>
                        {cfg.enabled ? t('enabled') : t('disabled')}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <div className="flex items-center justify-center gap-1">
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => openEdit(cfg)}>
                          <Pencil className="h-3.5 w-3.5" />
                        </Button>
                        {cfg.model_name !== '*' && (
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => deleteMut.mutate(cfg.id)}>
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? t('edit_config') : t('add_config')}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>{t('model')}</Label>
              <Input value={form.model_name} onChange={(e) => setForm(f => ({ ...f, model_name: e.target.value }))}
                placeholder="gpt-4o / *" disabled={!!editing} />
              <p className="text-xs text-muted-foreground">{t('model_name_hint')}</p>
            </div>
            <div className="space-y-2">
              <Label>{t('unified_cost')} (USD/1K)</Label>
              <Input type="number" step="0.000001" value={form.unified_cost}
                onChange={(e) => setForm(f => ({ ...f, unified_cost: e.target.value }))} />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>{t('commission_rate')} (%)</Label>
                <Input type="number" step="0.1" value={form.commission_rate}
                  onChange={(e) => setForm(f => ({ ...f, commission_rate: e.target.value }))} />
              </div>
              <div className="space-y-2">
                <Label>{t('platform_fee_rate')} (%)</Label>
                <Input type="number" step="0.1" value={form.platform_fee_rate}
                  onChange={(e) => setForm(f => ({ ...f, platform_fee_rate: e.target.value }))} />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('cancel')}</Button>
            <Button onClick={handleSave} disabled={createMut.isPending || updateMut.isPending}>
              {t('save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
