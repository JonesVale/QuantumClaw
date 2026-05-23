import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Plus, RefreshCw, Pencil, Trash2, DollarSign } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
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

  return (
    <div className="p-4 sm:p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{t('settlement_config')}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t('settlement_config_desc')}</p>
        </div>
        <Button onClick={openCreate} className="gap-2">
          <Plus className="h-4 w-4" /> {t('add_config')}
        </Button>
      </div>

      <Card>
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
