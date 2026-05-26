import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getMyCommissionRecords, getCommissionSetting, saveCommissionSetting,
  requestCommissionWithdrawal, type CommissionRecord, type CommissionSetting
} from '@/lib/api-extended'
import { useAuthStore } from '@/stores/auth-store'
import { DollarSign, History, Settings, RefreshCw, Wallet } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/commission')({
  component: CommissionPage,
})

function CommissionPage() {
  const { t } = useT()
  const { auth } = useAuthStore()
  const qc = useQueryClient()
  const isAdmin = auth.user?.role && auth.user.role >= 10
  const [tab, setTab] = useState<'records' | 'settings'>('records')
  const [amount, setAmount] = useState('')
  const [withdrawing, setWithdrawing] = useState(false)

  const { data: records, isLoading: recLoading } = useQuery({
    queryKey: ['commission-records'],
    queryFn: getMyCommissionRecords,
    staleTime: 30_000,
  })

  const { data: settings } = useQuery({
    queryKey: ['commission-settings'],
    queryFn: getCommissionSetting,
    staleTime: 60_000,
    enabled: isAdmin,
  })

  const handleWithdraw = async () => {
    if (!amount || Number(amount) <= 0) { toast.error(t('Enter valid amount')); return }
    setWithdrawing(true)
    try {
      const res = await requestCommissionWithdrawal(Number(amount))
      if (res.success) {
        toast.success(t('Withdrawal submitted'))
        qc.invalidateQueries({ queryKey: ['commission-records'] })
        setAmount('')
      } else toast.error(res.message || t('Failed'))
    } catch { toast.error(t('Failed')) }
    setWithdrawing(false)
  }

  const recordsList = (records?.data || []) as CommissionRecord[]
  const totalCommission = recordsList.reduce((s, r) => s + r.amount, 0)

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="text-center">
        <h1 className="text-3xl font-bold mb-2">{t('Commission')}</h1>
        <p className="text-muted-foreground">{t('Referral earnings and commission management')}</p>
      </div>

      <Card className="border-green-200">
        <CardHeader className="pb-2">
          <CardTitle className="flex items-center gap-2 text-sm"><Wallet className="h-4 w-4 text-green-600" />{t('Total Earnings')}</CardTitle>
        </CardHeader>
        <CardContent>
          {recLoading ? <Skeleton className="h-10 w-32" /> : (
            <div className="text-3xl font-bold text-green-600">{totalCommission.toLocaleString()}</div>
          )}
        </CardContent>
      </Card>

      <div className="flex gap-2">
        <Button variant={tab === 'records' ? 'default' : 'outline'} size="sm" onClick={() => setTab('records')}>
          <History className="h-4 w-4 mr-1" />{t('Records')}
        </Button>
        {isAdmin && (
          <Button variant={tab === 'settings' ? 'default' : 'outline'} size="sm" onClick={() => setTab('settings')}>
            <Settings className="h-4 w-4 mr-1" />{t('Settings')}
          </Button>
        )}
      </div>

      {tab === 'records' && (
        <>
          <Card>
            <CardHeader><CardTitle className="text-sm">{t('Withdraw')}</CardTitle></CardHeader>
            <CardContent className="flex gap-2">
              <Input type="number" min="1" placeholder={t('Amount')} value={amount} onChange={e => setAmount(e.target.value)} className="flex-1" />
              <Button onClick={handleWithdraw} disabled={withdrawing || !amount}>{withdrawing ? t('Submitting...') : t('Withdraw')}</Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle className="text-sm">{t('Commission Records')}</CardTitle></CardHeader>
            <CardContent>
              {recLoading ? <Skeleton className="h-20" /> : recordsList.length === 0 ? (
                <p className="text-sm text-muted-foreground italic">{t('No commission records yet')}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead><tr className="border-b text-muted-foreground">
                      <th className="text-left py-2 px-2 font-medium">{t('Amount')}</th>
                      <th className="text-left py-2 px-2 font-medium">{t('Status')}</th>
                      <th className="text-right py-2 px-2 font-medium">{t('Date')}</th>
                    </tr></thead>
                    <tbody>
                      {recordsList.slice(0, 50).map(r => (
                        <tr key={r.id} className="border-b hover:bg-muted/30">
                          <td className="py-2 px-2 font-medium">{r.amount.toLocaleString()}</td>
                          <td className="py-2 px-2"><Badge variant={r.status === 'completed' ? 'default' : 'secondary'}>{r.status}</Badge></td>
                          <td className="py-2 px-2 text-right text-muted-foreground">{new Date(r.created_at).toLocaleDateString()}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}

      {tab === 'settings' && isAdmin && (
        <Card>
          <CardHeader><CardTitle className="text-sm">{t('Commission Rate Settings')}</CardTitle></CardHeader>
          <CardContent>
            {!settings?.data ? <Skeleton className="h-20" /> : (
              <div className="space-y-2">
                {(settings.data as CommissionSetting[]).map(s => (
                  <div key={s.id} className="flex items-center justify-between p-3 rounded-lg border text-sm">
                    <span>{s.level}</span>
                    <span className="font-mono">{(s.rate * 100).toFixed(1)}%</span>
                    <Badge variant={s.enabled ? 'default' : 'secondary'}>{s.enabled ? t('Active') : t('Disabled')}</Badge>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
