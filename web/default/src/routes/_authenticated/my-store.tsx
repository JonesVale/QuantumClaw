import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import {
  Store, Save, Upload, Eye, EyeOff, Plus, X, Settings, Package, Tags,
  Image, DollarSign, Percent, RefreshCw, CheckCircle, XCircle, Globe
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'
import {
  getMyStore, saveMyStore, toggleStoreStatus, checkStoreSlug,
  getMyStoreModels, addStoreModel, updateStoreModel, deleteStoreModel,
  getChannels, type ProviderStore, type StoreModelItem
} from '@/lib/api-extended'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/my-store')({
  component: MyStorePage,
})

function MyStorePage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<'info' | 'models'>('info')
  const [showAddModel, setShowAddModel] = useState(false)

  const { data: storeData, isLoading } = useQuery({ queryKey: ['my-store'], queryFn: getMyStore })
  const store = storeData?.data?.store as ProviderStore | undefined
  const modelCount = storeData?.data?.model_count || 0

  // ── Store Info Form ──
  const [form, setForm] = useState({ name: '', description: '', logo: '', banner_url: '', contact_info: '', store_slug: '' })
  const [slugAvailable, setSlugAvailable] = useState<boolean | null>(null)

  useEffect(() => {
    if (store) setForm({
      name: store.name || '',
      description: store.description || '',
      logo: store.logo || '',
      banner_url: store.banner_url || '',
      contact_info: store.contact_info || '',
      store_slug: store.store_slug || '',
    })
  }, [store])

  const checkSlug = async (slug: string) => {
    if (!slug || slug === store?.store_slug) { setSlugAvailable(null); return }
    const res = await checkStoreSlug(slug)
    setSlugAvailable(res.data?.data?.available ?? false)
  }

  const saveMutation = useMutation({
    mutationFn: () => saveMyStore(form),
    onSuccess: () => { toast.success(t('Store saved')); queryClient.invalidateQueries({ queryKey: ['my-store'] }) },
    onError: () => toast.error(t('Failed to save store')),
  })

  const toggleMutation = useMutation({
    mutationFn: toggleStoreStatus,
    onSuccess: (res) => { toast.success(res.data?.data?.status === 1 ? t('Store opened') : t('Store closed')); queryClient.invalidateQueries({ queryKey: ['my-store'] }) },
  })

  // ── Models ──
  const { data: modelsData, refetch: refetchModels } = useQuery({ queryKey: ['my-store-models'], queryFn: () => getMyStoreModels() })
  const models: StoreModelItem[] = modelsData?.data || []
  const { data: channelsData } = useQuery({ queryKey: ['my-channels'], queryFn: () => getChannels(undefined, {}) })
  const channels: { id: number; name: string; models: string }[] = channelsData?.data || []

  const deleteModelMutation = useMutation({
    mutationFn: deleteStoreModel,
    onSuccess: () => { toast.success(t('Model removed')); refetchModels() },
    onError: () => toast.error(t('Failed to remove model')),
  })

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2"><Store className="w-7 h-7 inline mr-2" />{t('My Store')}</h1>
        <p className="text-muted-foreground" style={{maxWidth: 'min(55ch,100%)'}}>{t('Set up your provider store and list your models')}</p>
      </div>

      {/* Tab Switcher */}
      <div className="flex gap-1 mb-4 bg-muted/30 p-1 rounded-xl w-fit">
        <button onClick={() => setTab('info')} className={cn('px-5 py-2.5 rounded-lg text-sm font-medium transition-all', tab === 'info' ? 'bg-white shadow-sm' : 'hover:bg-muted/20')}>
          <Settings className="w-4 h-4 inline mr-1.5" />{t('Store Info')}
        </button>
        <button onClick={() => setTab('models')} className={cn('px-5 py-2.5 rounded-lg text-sm font-medium transition-all', tab === 'models' ? 'bg-white shadow-sm' : 'hover:bg-muted/20')}>
          <Package className="w-4 h-4 inline mr-1.5" />{t('Models')} {models.length > 0 && <Badge variant="secondary" className="ml-1">{models.length}</Badge>}
        </button>
      </div>

      {isLoading ? <div className="space-y-4">{Array.from({length:3}).map((_,i) => <Skeleton key={i} className="h-20 w-full" />)}</div> : (
        <>
          {/* ══════ STORE INFO TAB ══════ */}
          {tab === 'info' && (
            <div className="max-w-2xl mx-auto space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center justify-between">
                    <span><Store className="w-5 h-5 inline mr-2" />{t('Store Profile')}</span>
                    <Badge variant={store?.status === 1 ? 'default' : 'secondary'} className={store?.status === 1 ? 'bg-green-100 text-green-700' : ''}>
                      {store?.status === 1 ? <><CheckCircle className="w-3 h-3 mr-1" />{t('Open')}</> : <><XCircle className="w-3 h-3 mr-1" />{t('Closed')}</>}
                    </Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2"><Label>{t('Store Name')}</Label><Input value={form.name} onChange={e => setForm({...form, name: e.target.value})} placeholder={t('My AI Store')} /></div>
                    <div className="space-y-2">
                      <Label>{t('Store URL')}</Label>
                      <div className="flex items-center gap-2">
                        <span className="text-xs text-muted-foreground shrink-0">/store/</span>
                        <Input value={form.store_slug} onChange={e => { setForm({...form, store_slug: e.target.value}); checkSlug(e.target.value) }} placeholder="my-ai-store" />
                        {slugAvailable === true && <CheckCircle className="w-4 h-4 text-green-500 shrink-0" />}
                        {slugAvailable === false && <XCircle className="w-4 h-4 text-red-500 shrink-0" />}
                      </div>
                    </div>
                  </div>
                  <div className="space-y-2"><Label>{t('Description')}</Label><textarea className="w-full min-h-[80px] px-3 py-2 rounded-lg border bg-background text-sm" value={form.description} onChange={e => setForm({...form, description: e.target.value})} placeholder={t('Describe your store and services...')} /></div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2"><Label>{t('Logo URL')}</Label><Input value={form.logo} onChange={e => setForm({...form, logo: e.target.value})} placeholder="https://..." /></div>
                    <div className="space-y-2"><Label>{t('Contact Info')}</Label><Input value={form.contact_info} onChange={e => setForm({...form, contact_info: e.target.value})} placeholder="email@example.com" /></div>
                  </div>
                  <div className="space-y-2"><Label>{t('Banner URL (optional)')}</Label><Input value={form.banner_url} onChange={e => setForm({...form, banner_url: e.target.value})} placeholder="https://..." /></div>
                  <div className="flex gap-3 pt-2">
                    <Button onClick={() => saveMutation.mutate()} disabled={saveMutation.isPending}><Save className="w-4 h-4 mr-2" />{t('Save')}</Button>
                    <Button variant="outline" onClick={() => toggleMutation.mutate()} disabled={toggleMutation.isPending}>
                      {store?.status === 1 ? <><EyeOff className="w-4 h-4 mr-2" />{t('Close Store')}</> : <><Eye className="w-4 h-4 mr-2" />{t('Open Store')}</>}
                    </Button>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader><CardTitle><BarChart className="w-5 h-5 inline mr-2" />{t('Store Stats')}</CardTitle></CardHeader>
                <CardContent>
                  <div className="grid grid-cols-3 gap-4 text-center">
                    <div><p className="text-2xl font-bold">{modelCount}</p><p className="text-xs text-muted-foreground">{t('Models Listed')}</p></div>
                    <div><p className="text-2xl font-bold">{store?.rating?.toFixed(1) || '-'}</p><p className="text-xs text-muted-foreground">{t('Rating')}</p></div>
                    <div><p className="text-2xl font-bold">{store?.total_sales || 0}</p><p className="text-xs text-muted-foreground">{t('Total Sales')}</p></div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {/* ══════ MODELS TAB ══════ */}
          {tab === 'models' && (
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <p className="text-sm text-muted-foreground">{t('Manage which models are visible in your store')}</p>
                <Button onClick={() => setShowAddModel(true)}><Plus className="w-4 h-4 mr-2" />{t('Add Model')}</Button>
              </div>

              {models.length === 0 ? (
                <Card><CardContent className="py-12 text-center text-muted-foreground">{t('No models in your store yet. Click "Add Model" to start listing.')}</CardContent></Card>
              ) : (
                <Card><CardContent className="p-0">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t('Model')}</TableHead>
                        <TableHead>{t('Channel')}</TableHead>
                        <TableHead>{t('Price Multiplier')}</TableHead>
                        <TableHead>{t('Tags')}</TableHead>
                        <TableHead>{t('Status')}</TableHead>
                        <TableHead className="text-right">{t('Actions')}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {models.map(m => (
                        <TableRow key={m.id}>
                          <TableCell><span className="font-medium">{m.display_name || m.model_name}</span><div className="text-xs text-muted-foreground">{m.model_name}</div></TableCell>
                          <TableCell className="text-xs">#{m.channel_id}</TableCell>
                          <TableCell><Badge variant="outline">{m.price_multiplier?.toFixed(2)}x</Badge></TableCell>
                          <TableCell><div className="flex gap-1 flex-wrap">{m.tags?.split(',').map((tag, i) => <Badge key={i} variant="secondary" className="text-[10px]">{tag.trim()}</Badge>)}</div></TableCell>
                          <TableCell>{m.is_active ? <Badge className="bg-green-100 text-green-700 text-[10px]">{t('Active')}</Badge> : <Badge variant="secondary" className="text-[10px]">{t('Inactive')}</Badge>}</TableCell>
                          <TableCell className="text-right">
                            <Button variant="ghost" size="sm" onClick={() => deleteModelMutation.mutate(m.id)} className="text-red-500"><X className="w-4 h-4" /></Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </CardContent></Card>
              )}
            </div>
          )}
        </>
      )}

      {/* ════ Add Model Dialog ════ */}
      <AddModelDialog open={showAddModel} onOpenChange={setShowAddModel} channels={channels} onSuccess={() => { refetchModels(); setShowAddModel(false) }} />
    </div>
  )
}

function AddModelDialog({ open, onOpenChange, channels, onSuccess }: {
  open: boolean; onOpenChange: (o: boolean) => void; channels: { id: number; name: string; models: string }[]; onSuccess: () => void
}) {
  const { t } = useT()
  const [form, setForm] = useState({ channel_id: 0, model_name: '', display_name: '', description: '', tags: '', price_multiplier: 1 })

  const addMutation = useMutation({
    mutationFn: () => addStoreModel(form),
    onSuccess: () => { toast.success(t('Model listed')); onSuccess(); setForm({ channel_id: 0, model_name: '', display_name: '', description: '', tags: '', price_multiplier: 1 }) },
    onError: (e: any) => toast.error(e?.response?.data?.message || t('Failed to add model')),
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader><DialogTitle><Plus className="w-5 h-5 inline mr-2" />{t('Add Model to Store')}</DialogTitle></DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2"><Label>{t('Channel')}</Label>
            <Select value={String(form.channel_id)} onValueChange={v => setForm({...form, channel_id: Number(v)})}>
              <SelectTrigger><SelectValue placeholder={t('Select channel')} /></SelectTrigger>
              <SelectContent>{channels.map(ch => <SelectItem key={ch.id} value={String(ch.id)}>{ch.name}</SelectItem>)}</SelectContent>
            </Select>
          </div>
          <div className="space-y-2"><Label>{t('Model Name')}</Label><Input value={form.model_name} onChange={e => setForm({...form, model_name: e.target.value})} placeholder="gpt-4o" /></div>
          <div className="space-y-2"><Label>{t('Display Name')}</Label><Input value={form.display_name} onChange={e => setForm({...form, display_name: e.target.value})} placeholder="GPT-4o Latest" /></div>
          <div className="space-y-2"><Label>{t('Description')}</Label><textarea className="w-full min-h-[60px] px-3 py-2 rounded-lg border bg-background text-sm" value={form.description} onChange={e => setForm({...form, description: e.target.value})} placeholder={t('Short model description...')} /></div>
          <div className="space-y-2"><Label>{t('Tags (comma separated)')}</Label><Input value={form.tags} onChange={e => setForm({...form, tags: e.target.value})} placeholder="代码编程,多模态,推理" /></div>
          <div className="space-y-2"><Label>{t('Price Multiplier')}</Label><Input type="number" step="0.1" min="0.1" value={form.price_multiplier} onChange={e => setForm({...form, price_multiplier: parseFloat(e.target.value) || 1})} placeholder="1.0" /></div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>{t('Cancel')}</Button>
          <Button onClick={() => addMutation.mutate()} disabled={addMutation.isPending || !form.channel_id || !form.model_name}>{t('Add to Store')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function BarChart(props: any) { return <svg {...props} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="12" y1="20" x2="12" y2="10"/><line x1="18" y1="20" x2="18" y2="4"/><line x1="6" y1="20" x2="6" y2="16"/></svg> }
