import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { Plus, Trash2, Play, RefreshCw, Key, Wifi, WifiOff, Atom, Cpu } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { toast } from 'sonner'
import apiClient from '@/lib/api'
import {
  getChannels,
  createChannel,
  deleteChannel,
  testChannel,
  getChannelTypes,
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
  const [tab, setTab] = useState('ai')
  const [form, setForm] = useState({ name: '', type: '1', key: '', base_url: '', models: '' })
  const [testing, setTesting] = useState<number | null>(null)

  const { data: channelsData, isLoading } = useQuery({
    queryKey: ['reseller-channels'],
    queryFn: () => getChannels(undefined, {}),
    staleTime: 15_000,
  })
  const allChannels = (channelsData?.data as Channel[]) || []

  const { data: typeMap } = useQuery({
    queryKey: ['channelTypes'],
    queryFn: getChannelTypes,
    staleTime: 10 * 60 * 1000,
  })

  const typeGroups = useMemo(() => {
    const list = Array.isArray(typeMap) ? (typeMap as { id: number; name: string; url: string; models: string[] }[]) : []
    const ai: { id: number; name: string }[] = []
    const quantum: { id: number; name: string }[] = []
    list.forEach((t) => {
      if (t.id <= 0) return
      if (t.id >= 100) quantum.push({ id: t.id, name: t.name })
      else ai.push({ id: t.id, name: t.name })
    })
    return { ai, quantum }
  }, [typeMap])

  const aiChannels = useMemo(() => allChannels.filter(ch => ch.type < 100), [allChannels])
  const quantumChannels = useMemo(() => allChannels.filter(ch => ch.type >= 100), [allChannels])

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

  const resetForm = () => setForm({ name: '', type: tab === 'quantum' ? '100' : '1', key: '', base_url: '', models: '' })

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

  const statusMap: Record<number, { label: string; variant: 'default' | 'secondary' | 'destructive'; icon: React.ReactNode }> = {
    1: { label: t('enabled'), variant: 'default', icon: <Wifi className="w-3 h-3" /> },
    2: { label: t('disabled'), variant: 'secondary', icon: <WifiOff className="w-3 h-3" /> },
    3: { label: t('auto_disabled'), variant: 'destructive', icon: <WifiOff className="w-3 h-3" /> },
  }

  const renderChannelTable = (channels: Channel[]) => (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b bg-muted/30">
            <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('name')}</th>
            <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('type')}</th>
            <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('model')}</th>
            <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('status')}</th>
            <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('actions')}</th>
          </tr>
        </thead>
        <tbody>
          {channels.map((ch) => {
            const st = statusMap[ch.status] || { label: ch.status.toString(), variant: 'secondary' as const, icon: <WifiOff className="w-3 h-3" /> }
            const typeName = Array.isArray(typeMap)
              ? (typeMap.find((t: { id: number }) => t.id === ch.type) as { name?: string })?.name || `Type ${ch.type}`
              : `Type ${ch.type}`
            const isQuantum = ch.type >= 100
            return (
              <tr key={ch.id} className="border-b border-muted/50 hover:bg-muted/30">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    {isQuantum ? <Atom className="h-4 w-4 text-blue-500 shrink-0" /> : <Cpu className="h-4 w-4 text-amber-500 shrink-0" />}
                    <span className="font-medium text-sm">{ch.name}</span>
                  </div>
                </td>
                <td className="px-4 py-3 text-sm">
                  <Badge variant="outline" className={isQuantum ? 'border-blue-200 text-blue-600' : 'border-amber-200 text-amber-600'}>{typeName}</Badge>
                </td>
                <td className="px-4 py-3 text-sm text-muted-foreground max-w-[min(30vw,200px)] truncate">{ch.models || '-'}</td>
                <td className="px-4 py-3 text-center">
                  <Badge variant={st.variant} className="text-[10px] flex items-center gap-1 w-fit mx-auto">{st.icon}{st.label}</Badge>
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
              <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                <Key className="h-12 w-12 mx-auto mb-3 opacity-30" />
                {t('no_keys')}
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold mb-2">{t('my_keys')}</h1>
          <p className="text-muted-foreground mb-2" style={{maxWidth: 'min(65ch, 100%)'}}>{t('my_keys_desc')}</p>
        </div>
        <Button onClick={() => { resetForm(); setDialogOpen(true) }}>
          <Plus className="w-4 h-4 mr-2" />{t('add_key')}
        </Button>
      </div>

      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          <TabsTrigger value="ai" className="flex items-center gap-1.5">
            <Cpu className="w-4 h-4" />{t('AI Models')}
            <Badge variant="secondary" className="ml-1 text-[10px]">{aiChannels.length}</Badge>
          </TabsTrigger>
          <TabsTrigger value="quantum" className="flex items-center gap-1.5">
            <Atom className="w-4 h-4" />{t('Quantum Computing')}
            <Badge variant="secondary" className="ml-1 text-[10px]">{quantumChannels.length}</Badge>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="ai" className="mt-4">
          <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
            <CardContent className="p-0">{renderChannelTable(aiChannels)}</CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="quantum" className="mt-4">
          <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
            <CardContent className="p-0">{renderChannelTable(quantumChannels)}</CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('add_key')}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>{t('name')}</Label>
                <Input value={form.name} onChange={(e) => setForm(f => ({ ...f, name: e.target.value }))} placeholder={tab === 'quantum' ? "My IonQ Key" : "My OpenAI Key"} />
              </div>
              <div className="space-y-2">
                <Label>{t('type')}</Label>
                <Select value={form.type} onValueChange={(v) => setForm(f => ({ ...f, type: v }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {tab === 'ai' && typeGroups.ai.length > 0 && <>
                      <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">{t('AI Models')}</div>
                      {typeGroups.ai.map((t) => <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>)}
                    </>}
                    {tab === 'quantum' && typeGroups.quantum.length > 0 && <>
                      <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">{t('Quantum Computing')}</div>
                      {typeGroups.quantum.map((t) => <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>)}
                    </>}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-2">
              <Label>{t("API Key")}</Label>
              <Input type="password" value={form.key} onChange={(e) => setForm(f => ({ ...f, key: e.target.value }))} placeholder="sk-... or quantum API token" />
            </div>
            <div className="space-y-2">
              <Label>{t('base_url')}</Label>
              <Input value={form.base_url} onChange={(e) => setForm(f => ({ ...f, base_url: e.target.value }))} placeholder="https://api.openai.com" />
            </div>
            <div className="space-y-2">
              <Label>{t('models')}</Label>
              <Input value={form.models} onChange={(e) => setForm(f => ({ ...f, models: e.target.value }))} placeholder={tab === 'quantum' ? "ionq/harmony,ionq/aria-1" : "gpt-4o,gpt-4o-mini"} />
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