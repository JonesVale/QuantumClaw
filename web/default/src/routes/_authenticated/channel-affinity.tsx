import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getChannelAffinitySettings, saveChannelAffinitySettings, clearChannelAffinityCache, getChannelAffinityCacheStats, type ChannelAffinitySetting } from '@/lib/api-extended'
import { RefreshCw, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/channel-affinity')({
  component: ChannelAffinityPage,
})

function ChannelAffinityPage() {
  const { t } = useT()
  const qc = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['channel-affinity'],
    queryFn: getChannelAffinitySettings,
    staleTime: 30_000,
  })

  const { data: cacheStats } = useQuery({
    queryKey: ['channel-affinity-cache-stats'],
    queryFn: getChannelAffinityCacheStats,
    staleTime: 10_000,
  })

  const clearCache = async () => {
    const res = await clearChannelAffinityCache()
    if (res.success) { toast.success(t('Cache cleared')); qc.invalidateQueries({ queryKey: ['channel-affinity'] }) }
    else toast.error(res.message || t('Failed'))
  }

  const settings = (data?.data || []) as ChannelAffinitySetting[]

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex justify-between items-center">
        <div><h1 className="text-3xl font-bold">{t('Channel Affinity')}</h1><p className="text-muted-foreground text-sm">{t('Configure model-to-channel routing preferences')}</p></div>
        <Button variant="outline" size="sm" onClick={clearCache}><Trash2 className="h-4 w-4 mr-1" />{t('Clear Cache')}</Button>
      </div>

      {cacheStats?.success && (
        <Card className="border-blue-200">
          <CardContent className="py-3 text-sm flex gap-4">
            <span>{t('Cache entries')}: {(cacheStats.data as any)?.entries || 0}</span>
            <span>{t('Hit rate')}: {((cacheStats.data as any)?.hit_rate * 100 || 0).toFixed(1)}%</span>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader><CardTitle className="text-sm">{t('Affinity Rules')}</CardTitle></CardHeader>
        <CardContent>
          {isLoading ? <Skeleton className="h-20" /> : settings.length === 0 ? (
            <p className="text-sm text-muted-foreground italic">{t('No affinity rules configured')}</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead><tr className="border-b text-muted-foreground">
                  <th className="text-left py-2 px-2 font-medium">{t('Model')}</th>
                  <th className="text-left py-2 px-2 font-medium">{t('Preferred Channel')}</th>
                  <th className="text-left py-2 px-2 font-medium">{t('Fallbacks')}</th>
                  <th className="text-center py-2 px-2 font-medium">{t('Status')}</th>
                </tr></thead>
                <tbody>
                  {settings.map(s => (
                    <tr key={s.id} className="border-b hover:bg-muted/30">
                      <td className="py-2 px-2 font-medium">{s.model_name}</td>
                      <td className="py-2 px-2">{s.preferred_channel_id}</td>
                      <td className="py-2 px-2 text-muted-foreground">{s.fallback_channel_ids || '-'}</td>
                      <td className="py-2 px-2 text-center"><Badge variant={s.enabled ? 'default' : 'secondary'}>{s.enabled ? t('Active') : t('Disabled')}</Badge></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
