import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useMemo, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Search, MoreHorizontal, Pencil, Trash2, Play, CheckCircle, XCircle, RefreshCw, Server, Network, ExternalLink, Wallet } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { EmptyState } from '@/components/empty-state'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { type Channel, type ChannelFormData, getChannels, getChannelTypes, createChannel, updateChannel, deleteChannel, testChannel } from '@/lib/api-extended'
import { toast } from 'sonner'
import dayjs from '@/lib/dayjs'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/channels')({
  component: ChannelsPage,
})

function ChannelFormDialog({ open, onOpenChange, channel, creatingNew }: {
  open: boolean; onOpenChange: (open: boolean) => void; channel?: Channel | null; creatingNew?: boolean
}) {
  const { t } = useT()
  const queryClient = useQueryClient()
  const isEdit = !!channel
  const { data: typeMap } = useQuery({ queryKey: ['channelTypes'], queryFn: getChannelTypes, staleTime: 10 * 60 * 1000 })
  const typeGroups = useMemo(() => {
    const map = typeMap || {}
    const ai: { id: number; name: string }[] = []
    const quantum: { id: number; name: string }[] = []
    Object.entries(map).forEach(([idStr, name]) => {
      const id = Number(idStr)
      if (id <= 0) return
      if (id >= 100) quantum.push({ id, name })
      else ai.push({ id, name })
    })
    return { ai, quantum }
  }, [typeMap])
  const [form, setForm] = useState<ChannelFormData>({ type: 1, key: '', name: '', base_url: '', models: '', group: 'default', model_mapping: '', priority: 0, weight: 1, cache_billing_ratio: 0, cost_per_unit: 0, sell_price_rate: 1, thinking_to_content: false })
  function parseConfig(config?: string): { cache_billing_ratio?: number; thinking_to_content?: boolean } {
    if (!config) return {}
    try { const parsed = JSON.parse(config); return { cache_billing_ratio: typeof parsed.cache_billing_ratio === 'number' ? parsed.cache_billing_ratio : undefined, thinking_to_content: typeof parsed.thinking_to_content === 'boolean' ? parsed.thinking_to_content : undefined } }
    catch { return {} }
  }
  const prevChannel = useRef(channel)
  useEffect(() => {
    if (channel && channel !== prevChannel.current) {
      const cfg = parseConfig(channel.config)
      setForm({ type: channel.type || 1, key: '', name: channel.name || '', base_url: channel.base_url || '', models: channel.models || '', group: channel.group || 'default', model_mapping: '', priority: 0, weight: channel.weight || 1, cache_billing_ratio: cfg.cache_billing_ratio ?? 0, cost_per_unit: channel.cost_per_unit || 0, sell_price_rate: channel.sell_price_rate || 1, thinking_to_content: cfg.thinking_to_content ?? false })
    }
    if (!channel) setForm({ type: 1, key: '', name: '', base_url: '', models: '', group: 'default', model_mapping: '', priority: 0, weight: 1, cache_billing_ratio: 0, cost_per_unit: 0, sell_price_rate: 1, thinking_to_content: false })
    prevChannel.current = channel
  }, [channel])
  const createMutation = useMutation({ mutationFn: createChannel, onSuccess: () => { toast.success(t('Channel created')); queryClient.invalidateQueries({ queryKey: ['channels'] }); onOpenChange(false) }, onError: () => toast.error(t('Failed to create channel')) })
  const updateMutation = useMutation({ mutationFn: (data: ChannelFormData & { id: number }) => updateChannel(data), onSuccess: () => { toast.success(t('Channel updated')); queryClient.invalidateQueries({ queryKey: ['channels'] }); onOpenChange(false) }, onError: () => toast.error(t('Failed to update channel')) })
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (isEdit && channel) {
      const payload: Record<string, unknown> = { id: channel.id }
      if (form.name) payload.name = form.name; if (form.type) payload.type = form.type; if (form.key) payload.key = form.key
      if (form.base_url !== undefined) payload.base_url = form.base_url; if (form.models) payload.models = form.models; if (form.group) payload.group = form.group
      payload.category = form.category || ''; if (form.weight) payload.weight = form.weight
      if (form.cost_per_unit !== undefined) payload.cost_per_unit = form.cost_per_unit
      if (form.sell_price_rate !== undefined) payload.sell_price_rate = form.sell_price_rate
      if (form.cache_billing_ratio !== undefined) payload.cache_billing_ratio = form.cache_billing_ratio
      if (form.thinking_to_content !== undefined) payload.thinking_to_content = form.thinking_to_content
      updateMutation.mutate(payload as unknown as ChannelFormData & { id: number })
    } else createMutation.mutate(form)
  }
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl">
        <DialogHeader><DialogTitle className="flex items-center gap-2">{isEdit ? <><Pencil className="h-5 w-5" />{t('Edit Channel')}</> : <><Plus className="h-5 w-5" />{t('Create Channel')}</>}</DialogTitle>
          <DialogDescription>{isEdit ? t('Update channel configuration') : t('Add a new AI or Quantum computing provider channel')}</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-3"><Label>{t('Channel Name')}</Label><Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="e.g. OpenAI Official" required /></div>
            <div className="space-y-3"><Label>{t('Channel Type')}</Label>
              <Select value={String(form.type)} onValueChange={(v) => setForm({ ...form, type: Number(v) })}>
                <SelectTrigger><SelectValue placeholder={t('Select type')} /></SelectTrigger>
                <SelectContent className="max-h-72">
                  {typeGroups.ai.length > 0 && <><div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">{t('AI Models')}</div>{typeGroups.ai.map((t) => <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>)}</>}
                  {typeGroups.quantum.length > 0 && <><div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider border-t pt-2 mt-1">{t('Quantum Computing')}</div>{typeGroups.quantum.map((t) => <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>)}</>}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="space-y-3"><Label>{t('API Key')}</Label><Input type="password" value={form.key} onChange={(e) => setForm({ ...form, key: e.target.value })} placeholder={isEdit ? t('Leave empty to keep unchanged') : ''} required={!isEdit} /></div>
          <div className="space-y-3"><Label>{t('Base URL')}</Label><Input value={form.base_url} onChange={(e) => setForm({ ...form, base_url: e.target.value })} placeholder="https://api.openai.com/v1" /></div>
          <div className="space-y-3"><Label>{t('Models')}</Label><Input value={form.models} onChange={(e) => setForm({ ...form, models: e.target.value })} placeholder="gpt-4, gpt-3.5-turbo, ..." /></div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-3"><Label>{t('Group')}</Label><Input value={form.group} onChange={(e) => setForm({ ...form, group: e.target.value })} placeholder="default" /></div>
            <div className="space-y-3"><Label>{t('Weight')}</Label><Input type="number" value={form.weight} onChange={(e) => setForm({ ...form, weight: Number(e.target.value) })} min="1" /></div>
          </div>
          <div className="grid gap-2"><Label htmlFor="cost_per_unit">{t('Cost Per Unit')}</Label><Input id="cost_per_unit" type="number" min="0" step="0.0001" value={form.cost_per_unit ?? 0} onChange={(e) => setForm({ ...form, cost_per_unit: parseFloat(e.target.value) || 0 })} placeholder="0" /></div>
          <div className="grid gap-2"><Label htmlFor="sell_price_rate">{t('Sell Price Rate')}</Label><Input id="sell_price_rate" type="number" min="0" step="0.1" value={form.sell_price_rate ?? 1} onChange={(e) => setForm({ ...form, sell_price_rate: parseFloat(e.target.value) || 1 })} placeholder="1.0" /></div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>{t('Cancel')}</Button>
            <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
              {createMutation.isPending || updateMutation.isPending ? t('Saving...') : isEdit ? t('Update') : t('Create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function ChannelsPage() {
  const { t } = useT()
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<string>('all')
  const [typeCategory, setTypeCategory] = useState<string>('all')
  const [channelCategory, setChannelCategory] = useState<string>('all')
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null)
  const [creatingNew, setCreatingNew] = useState(false)
  const queryClient = useQueryClient()

  const { data: selfInfo } = useQuery({ queryKey: ['self'], queryFn: async () => { const res = await fetch('/api/user/self'); if (!res.ok) return null; return res.json() }, retry: false, staleTime: 30 * 1000 })
  const isProvider = selfInfo?.data?.user_type === 'provider' || (selfInfo?.data?.role ?? 0) >= 10
  const { data, isLoading, refetch, isFetching } = useQuery({ queryKey: ['channels'], queryFn: () => getChannels(), staleTime: 30 * 1000 })
  const { data: typeMap } = useQuery({ queryKey: ['channelTypes'], queryFn: getChannelTypes, staleTime: 10 * 60 * 1000 })
  const deleteMutation = useMutation({ mutationFn: deleteChannel, onSuccess: () => { toast.success(t('Channel deleted')); queryClient.invalidateQueries({ queryKey: ['channels'] }) }, onError: () => toast.error(t('Failed to delete channel')) })
  const testMutation = useMutation({ mutationFn: testChannel, onSuccess: (res) => { if (res?.data?.status === 'success') toast.success(t('Channel test successful')); else toast.error(res?.data?.error || t('Channel test failed')) }, onError: () => toast.error(t('Failed to test channel')) })

  const channels: Channel[] = data?.data || []
  const filtered = useMemo(() => channels.filter((ch) => {
    const matchesSearch = !search || ch.name.toLowerCase().includes(search.toLowerCase()) || ch.group.toLowerCase().includes(search.toLowerCase())
    const matchesStatus = status === 'all' || (status === 'enabled' && ch.status === 1) || (status === 'disabled' && ch.status === 2)
    const matchesCategory = typeCategory === 'all' || (typeCategory === 'ai' && Number(ch.type) < 100) || (typeCategory === 'quantum' && Number(ch.type) >= 100)
    const matchesChannelCat = channelCategory === 'all' || (channelCategory === ch.category) || (channelCategory === 'configured' && ch.key && !ch.key.startsWith('PUT_YOUR'))
    return matchesSearch && matchesStatus && matchesCategory && matchesChannelCat
  }), [channels, search, status, typeCategory, channelCategory])

  const getTypeBadge = (type: number) => {
    const typeName = typeMap?.[String(type)]; const isQuantum = type >= 100
    return <Badge variant="outline" className={isQuantum ? 'border-purple-300 text-purple-700 dark:border-purple-700 dark:text-purple-300' : ''}>{typeName || `Type ${type}`}</Badge>
  }

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Channels')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('Manage AI and Quantum computing channels')}</p>
      </div>

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="relative w-full sm:flex-1"><Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" /><Input className="pl-9" placeholder={t('Search channels...')} value={search} onChange={(e) => setSearch(e.target.value)} /></div>
            <Select value={typeCategory} onValueChange={setTypeCategory}>
              <SelectTrigger className="w-full sm:w-32"><SelectValue placeholder={t('All Types')} /></SelectTrigger>
              <SelectContent><SelectItem value="all">{t('All Types')}</SelectItem><SelectItem value="ai">{t('AI Models')}</SelectItem><SelectItem value="quantum">{t('Quantum')}</SelectItem></SelectContent>
            </Select>
            <Select value={channelCategory} onValueChange={setChannelCategory}>
              <SelectTrigger className="w-full sm:w-36"><SelectValue placeholder={t('Category')} /></SelectTrigger>
              <SelectContent><SelectItem value="all">{t('All')}</SelectItem><SelectItem value="free">{t('Free')}</SelectItem><SelectItem value="paid">{t('Paid')}</SelectItem><SelectItem value="configured">{t('Has API Key')}</SelectItem></SelectContent>
            </Select>
            <Select value={status} onValueChange={setStatus}>
              <SelectTrigger className="w-full sm:w-32"><SelectValue placeholder={t('Status')} /></SelectTrigger>
              <SelectContent><SelectItem value="all">{t('All Status')}</SelectItem><SelectItem value="enabled">{t('Enabled')}</SelectItem><SelectItem value="disabled">{t('Disabled')}</SelectItem></SelectContent>
            </Select>
            <Button variant="outline" onClick={() => refetch()} disabled={isFetching}><RefreshCw className={cn('mr-2 h-4 w-4', isFetching && 'animate-spin')} />{t('Refresh')}</Button>
          </div>
        </CardContent>
      </Card>

      {!isLoading && filtered.length === 0 ? (
        <EmptyState icon={Network} title={t('No channels found')} description={t('Add your first channel to get started')} action={{ label: t('Add Channel'), onClick: () => { setEditingChannel(null); setCreatingNew(true) } }} />
      ) : (
        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-12">{t('ID')}</TableHead>
                    <TableHead>{t('Name')}</TableHead>
                    <TableHead>{t('Type')}</TableHead>
                    <TableHead>{t('Group')}</TableHead>
                    <TableHead>{t('Cost')}</TableHead>
                    <TableHead>{t('Status')}</TableHead>
                    <TableHead>{t('Weight')}</TableHead>
                    <TableHead>{t('Created')}</TableHead>
                    <TableHead className="text-right">{t('Actions')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {isLoading ? Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>{Array.from({ length: 9 }).map((_, j) => <TableCell key={j}><Skeleton className="h-4 w-16" /></TableCell>)}</TableRow>
                  )) : filtered.map((channel) => (
                    <TableRow key={channel.id} className="hover:bg-muted/50 transition-colors">
                      <TableCell className="font-medium">{channel.id}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center"><Server className="h-4 w-4 text-white" /></div>
                          <span className="font-medium">{channel.name}</span>
                        </div>
                      </TableCell>
                      <TableCell>{getTypeBadge(channel.type)}</TableCell>
                      <TableCell><Badge variant="secondary">{channel.group}</Badge></TableCell>
                      <TableCell className="font-mono text-xs">{channel.cost_per_unit?.toFixed(4) || '0'}</TableCell>
                      <TableCell>{channel.status === 1 ? <Badge className="bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"><CheckCircle className="w-3 h-3 mr-1" />{t('Enabled')}</Badge> : <Badge variant="secondary" className="bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"><XCircle className="w-3 h-3 mr-1" />{t('Disabled')}</Badge>}</TableCell>
                      <TableCell><Badge variant="outline">{channel.weight}</Badge></TableCell>
                      <TableCell className="text-muted-foreground">{dayjs(channel.created_time * 1000).format('YYYY-MM-DD')}</TableCell>
                      <TableCell className="text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild><Button variant="ghost" size="icon" className="h-8 w-8"><MoreHorizontal className="h-4 w-4" /></Button></DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => setEditingChannel(channel)}><Pencil className="mr-2 h-4 w-4" />{t('Edit')}</DropdownMenuItem>
                            <DropdownMenuItem onClick={() => testMutation.mutate(channel.id)} disabled={testMutation.isPending}><Play className="mr-2 h-4 w-4" />{t('Test')}</DropdownMenuItem>
                            <DropdownMenuItem onClick={() => { if (confirm(t('Are you sure?'))) deleteMutation.mutate(channel.id) }} className="text-red-600"><Trash2 className="mr-2 h-4 w-4" />{t('Delete')}</DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      )}

      <ChannelFormDialog open={creatingNew || !!editingChannel} onOpenChange={(open) => { if (!open) { setEditingChannel(null); setCreatingNew(false) } }} channel={editingChannel} creatingNew={creatingNew} />
    </div>
  )
}
