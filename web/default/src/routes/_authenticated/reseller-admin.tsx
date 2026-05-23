import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { toast } from 'sonner'
import apiClient from '@/lib/api'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/_authenticated/reseller-admin')({
  component: ResellerAdminPage,
})

function ResellerAdminPage() {
  const { t } = useT()
  const queryClient = useQueryClient()

  const { data: resellers } = useQuery({
    queryKey: ['admin-resellers'],
    queryFn: async () => { const r = await apiClient.get('/api/admin/resellers'); return r.data?.data || [] },
    staleTime: 15_000,
  })

  const { data: withdrawals, refetch: refetchW } = useQuery({
    queryKey: ['admin-withdrawals'],
    queryFn: async () => { const r = await apiClient.get('/api/admin/withdrawals?status=pending'); return r.data?.data || [] },
    staleTime: 10_000,
  })

  const approveMut = useMutation({
    mutationFn: async (id: number) => { await apiClient.post(`/api/admin/withdrawals/${id}/approve`) },
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['admin-withdrawals'] }); toast.success(t('approved')) },
  })

  return (
    <div className="p-4 sm:p-6 space-y-8">
      {/* Resellers list */}
      <div>
        <h1 className="text-2xl font-bold mb-2">{t('reseller_management')}</h1>
        <p className="text-sm text-muted-foreground mb-4">{t('reseller_management_desc')}</p>
        <Card>
          <CardContent className="p-0">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3">{t('username')}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3">{t('store_name')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3">{t('balance')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3">{t('total_earned')}</th>
                  <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3">{t('status')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3">{t('created')}</th>
                </tr>
              </thead>
              <tbody>
                {(resellers || []).map((r: any) => (
                  <tr key={r.id} className="border-b border-muted/50">
                    <td className="px-4 py-3 font-medium">{r.username || '-'}</td>
                    <td className="px-4 py-3">{r.name || '-'}</td>
                    <td className="px-4 py-3 text-right font-mono">${(r.balance || 0).toFixed(2)}</td>
                    <td className="px-4 py-3 text-right font-mono">${(r.total_earned || 0).toFixed(2)}</td>
                    <td className="px-4 py-3 text-center">
                      <Badge variant={r.status === 1 ? 'default' : 'secondary'}>
                        {r.status === 1 ? t('active') : t('disabled')}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground text-right">
                      {dayjs(r.created_time * 1000).format('YYYY-MM-DD')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      </div>

      {/* Pending withdrawals */}
      <div>
        <h2 className="text-lg font-semibold mb-3">{t('pending_withdrawals')}</h2>
        <Card>
          <CardContent className="p-0">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3">{t('user_id')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3">{t('amount')}</th>
                  <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3">{t('status')}</th>
                  <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3">{t('actions')}</th>
                </tr>
              </thead>
              <tbody>
                {(withdrawals || []).map((w: any) => (
                  <tr key={w.id} className="border-b border-muted/50">
                    <td className="px-4 py-3">{w.user_id}</td>
                    <td className="px-4 py-3 text-right font-mono">${(w.amount / 100).toFixed(2)}</td>
                    <td className="px-4 py-3 text-center">
                      <Badge variant={w.status === 'pending' ? 'secondary' : 'default'}>
                        {w.status === 'pending' ? t('pending') : t('approved')}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-center">
                      {w.status === 'pending' && (
                        <Button size="sm" variant="outline" onClick={() => approveMut.mutate(w.id)}>
                          {t('approve')}
                        </Button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
