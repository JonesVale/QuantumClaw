import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { RefreshCw, Database, ArrowUpDown, Zap, Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import apiClient from '@/lib/api'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/admin-tools')({
  component: AdminToolsPage,
})

function AdminToolsPage() {
  const { t } = useT()

  const [syncCron, setSyncCron] = useState('0 3 * * *')
  const { data: syncStatus } = useQuery({ queryKey: ['modelSyncStatus'], queryFn: async () => { const r = await apiClient.get('/api/admin/model-sync'); return r.data?.data || {} } })
  const syncMut = useMutation({ mutationFn: async () => { await apiClient.post('/api/admin/model-sync/sync'); toast.success(t('Model sync started')) } })
  const saveSyncMut = useMutation({ mutationFn: async () => { await apiClient.put('/api/admin/model-sync', { cron_expr: syncCron }); toast.success(t('Sync config saved')) } })

  const [upstreamCron, setUpstreamCron] = useState('0 4 * * *')
  const checkMut = useMutation({ mutationFn: async () => { const r = await apiClient.post('/api/admin/upstream/check'); return r.data?.data } })
  const saveUpstreamMut = useMutation({ mutationFn: async () => { await apiClient.put('/api/admin/upstream', { cron_expr: upstreamCron }); toast.success(t('Upstream config saved')) } })

  const [affinityEnabled, setAffinityEnabled] = useState(false)
  const [affinityWeight, setAffinityWeight] = useState('2.0')
  const saveAffinityMut = useMutation({
    mutationFn: async () => { await apiClient.put('/api/admin/channel-affinity', { enabled: affinityEnabled, weight_boost: Number(affinityWeight) }); toast.success(t('Affinity saved')) }
  })
  const clearAffinityMut = useMutation({
    mutationFn: async () => { await apiClient.delete('/api/admin/channel-affinity/cache'); toast.success(t('Affinity cache cleared')) }
  })

  const { data: syncStatusData } = useQuery({ queryKey: ['modelSyncStatus'], queryFn: async () => { const r = await apiClient.get('/api/admin/model-sync'); return r.data?.data || {} } })

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-8">{t('Admin Tools')}</h1>
      </div>

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader><CardTitle className="flex items-center gap-2"><Database className="h-5 w-5" />{t('Model Sync')}</CardTitle></CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
            <Label className="min-w-20">{t('Cron Expression')}</Label>
            <Input value={syncCron} onChange={(e) => setSyncCron(e.target.value)} className="font-mono w-full sm:w-auto" placeholder="0 3 * * *" />
            <Button variant="outline" size="sm" onClick={() => saveSyncMut.mutate()} disabled={saveSyncMut.isPending}><Save className="h-4 w-4 mr-1" />{t('Save')}</Button>
          </div>
          <div className="flex items-center gap-4">
            <Button onClick={() => syncMut.mutate()} disabled={syncMut.isPending}><RefreshCw className="h-4 w-4 mr-1" />{t('Sync Now')}</Button>
            {syncStatusData?.last_sync && <span className="text-xs text-muted-foreground">{t('Last sync')}: {syncStatusData.last_sync}</span>}
          </div>
        </CardContent>
      </Card>

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader><CardTitle className="flex items-center gap-2"><ArrowUpDown className="h-5 w-5" />{t('Upstream Update')}</CardTitle></CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
            <Label className="min-w-20">{t('Cron Expression')}</Label>
            <Input value={upstreamCron} onChange={(e) => setUpstreamCron(e.target.value)} className="font-mono w-full sm:w-auto" placeholder="0 4 * * *" />
            <Button variant="outline" size="sm" onClick={() => saveUpstreamMut.mutate()} disabled={saveUpstreamMut.isPending}><Save className="h-4 w-4 mr-1" />{t('Save')}</Button>
          </div>
          <div>
            <Button onClick={() => checkMut.mutate()} disabled={checkMut.isPending}><RefreshCw className="h-4 w-4 mr-1" />{t('Check Updates')}</Button>
          </div>
        </CardContent>
      </Card>

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader><CardTitle className="flex items-center gap-2"><Zap className="h-5 w-5" />{t('Channel Affinity')}</CardTitle></CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between">
            <div><Label>{t('Enable Affinity')}</Label><p className="text-xs text-muted-foreground">{t('Prefer recently used channels')}</p></div>
            <Switch checked={affinityEnabled} onCheckedChange={setAffinityEnabled} />
          </div>
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
            <Label className="min-w-20">{t('Weight Boost')}</Label>
            <Input type="number" step="0.1" value={affinityWeight} onChange={(e) => setAffinityWeight(e.target.value)} className="w-24" />
            <Button variant="outline" size="sm" onClick={() => saveAffinityMut.mutate()} disabled={saveAffinityMut.isPending}><Save className="h-4 w-4 mr-1" />{t('Save')}</Button>
            <Button variant="outline" size="sm" onClick={() => clearAffinityMut.mutate()} disabled={clearAffinityMut.isPending}><Trash2 className="h-4 w-4 mr-1" />{t('Clear Cache')}</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
