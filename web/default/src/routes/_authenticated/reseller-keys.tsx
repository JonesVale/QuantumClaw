import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Plus, Trash2, Play, RefreshCw, Key, Wifi, WifiOff } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import {
  getChannels,
  createChannel,
  deleteChannel,
  testChannel,
  type Channel,
  type ChannelFormData,
} from '@/lib/api-extended'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/reseller-keys')({
  component: ResellerKeysPage,
})

function ResellerKeysPage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [form, setForm] = useState({ name: '', type: '1', key: '', base_url: '', models: '' })
  const [testing, setTesting] = useState<number | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['reseller-channels'],
    queryFn: () => getChannels(undefined, {}),
    staleTime: 15_000,
  })
  const channels = (data?.data as Channel[]) || []

  const createMut = useMutation({
    mutationFn: (data: ChannelFormData) => createChannel(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['reseller-channels'] }); setDialogOpen(false); resetForm(); toast.success(t('saved')) },
    onError: () => toast.error(t('save_failed')),
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => deleteChannel(id),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['reseller-channels'] }); toast.success(t('deleted')) },
    onError: () => toast.error(t('delete_failed')),
  })

  const testChannelFn = async (id: number) => {
    setTesting(id)
    try {
      const res = await testChannel(id)
      if (res.success) {
        toast.success(`${t('test_ok')} ${res.data?.latency_ms || '?'}ms`)
      } else {
        toast.error(res.message || t('test_failed'))
      }
    } catch {
      toast.error(t('test_failed'))
    } finally {
      setTesting(null)
    }
  }

  const resetForm = () => setForm({ name: '', type: '1', key: '', base_url: '', models: '' })

  const handleCreate = () => {
    if (!form.name || !form.key) {
      toast.error(t('name_and_key_required'))
      return
    }
    createMut.mutate({
      name: form.name,
      type: parseInt(form.type),
      key: form.key,
      base_url: form.base_url || undefined,
      models: form.models || undefined,
    } as ChannelFormData)
  }

  const statusMap: Record<number, { label: string; variant: 'default' | 'secondary' | 'destructive' }> = {
    1: { label: t('enabled'), variant: 'default' },
    2: { label: t('disabled'), variant: 'secondary' },
    3: { label: t('auto_disabled'), variant: 'destructive' },
  }

  return (
    <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">{t('my_keys')}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t('my_keys_desc')}</p>
        </div>
        <Button onClick={() => { resetForm(); setDialogOpen(true) }} className="gap-2">
          <Plus className="h-4 w-4" /> {t('add_key')}
        </Button>
      </div>

      <Card>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('name')}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('model')}</th>
                  <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('status')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('actions')}</th>
                </tr>
              </thead>
              <tbody>
                {channels.map((ch) => {
                  const st = statusMap[ch.status] || { label: ch.status.toString(), variant: 'secondary' }
                  return (
                    <tr key={ch.id} className="border-b border-muted/50 hover:bg-muted/30">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <Key className="h-4 w-4 text-muted-foreground shrink-0" />
                          <span className="font-medium text-sm">{ch.name}</span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm text-muted-foreground max-w-[min(30vw,200px)] truncate">{ch.models || '-'}</td>
                      <td className="px-4 py-3 text-center">
                        <Badge variant={st.variant} className="text-[10px]">{st.label}</Badge>
                      </td>
                      <td className="px-4 py-3 text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => testChannelFn(ch.id)} disabled={testing === ch.id}>
                            {testing === ch.id ? <RefreshCw className="h-3.5 w-3.5 animate-spin" /> : <Play className="h-3.5 w-3.5" />}
                          </Button>
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => deleteMut.mutate(ch.id)}>
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
                {channels.length === 0 && (
                  <tr>
                    <td colSpan={4} className="px-4 py-12 text-center text-muted-foreground">
                      <Key className="h-12 w-12 mx-auto mb-3 opacity-30" />
                      {t('no_keys')}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('add_key')}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>{t('name')}</Label>
                <Input value={form.name} onChange={(e) => setForm(f => ({ ...f, name: e.target.value }))} placeholder="My OpenAI Key" />
              </div>
              <div className="space-y-2">
                <Label>{t('type')}</Label>
                <Select value={form.type} onValueChange={(v) => setForm(f => ({ ...f, type: v }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1">OpenAI</SelectItem>
                    <SelectItem value="4">Claude</SelectItem>
                    <SelectItem value="12">DeepSeek</SelectItem>
                    <SelectItem value="14">Google Gemini</SelectItem>
                    <SelectItem value="24">Azure OpenAI</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-2">
              <Label>API Key</Label>
              <Input type="password" value={form.key} onChange={(e) => setForm(f => ({ ...f, key: e.target.value }))} placeholder="sk-..." />
            </div>
            <div className="space-y-2">
              <Label>{t('base_url')}</Label>
              <Input value={form.base_url} onChange={(e) => setForm(f => ({ ...f, base_url: e.target.value }))} placeholder="https://api.openai.com" />
            </div>
            <div className="space-y-2">
              <Label>{t('models')}</Label>
              <Input value={form.models} onChange={(e) => setForm(f => ({ ...f, models: e.target.value }))} placeholder="gpt-4o,gpt-4o-mini" />
              <p className="text-xs text-muted-foreground">{t('models_hint')}</p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('cancel')}</Button>
            <Button onClick={handleCreate} disabled={createMut.isPending}>{t('save')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
