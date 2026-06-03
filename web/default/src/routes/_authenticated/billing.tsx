import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { DollarSign, CreditCard, History, RefreshCw, TrendingUp, TrendingDown } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import apiClient from '@/lib/api'
import { type ApiResponse } from '@/lib/api-extended'

interface BillingRecord {
  id: number
  amount: number
  action: string
  status: string
  remark: string
  created_at: number
}

interface BillingStats {
  total_quota: number
  used_quota: number
  remain_quota: number
  request_count: number
  display_in_currency: boolean
  quota_per_unit: number
}

interface BillingStatsResponse extends ApiResponse<BillingStats> {}
interface BillingRecordsResponse extends ApiResponse<BillingRecord[]> {}

async function getBillingStats(): Promise<BillingStatsResponse> {
  const res = await apiClient.get('/api/user/self/billing/stats')
  return res.data
}

async function getBillingRecords(): Promise<BillingRecordsResponse> {
  const res = await apiClient.get('/api/user/self/billing/records')
  return res.data
}

export const Route = createFileRoute('/_authenticated/billing')({
  component: BillingPage,
})

function BillingPage() {
  const { t } = useT()

  const { data: statsData, isLoading: statsLoading } = useQuery({
    queryKey: ['billing-stats'],
    queryFn: getBillingStats,
    staleTime: 30 * 1000,
  })

  const { data: recordsData, isLoading: recordsLoading, refetch, isFetching } = useQuery({
    queryKey: ['billing-records'],
    queryFn: getBillingRecords,
    staleTime: 30 * 1000,
  })

  const stats: BillingStats = statsData?.data || {
    total_quota: 0,
    used_quota: 0,
    remain_quota: 0,
    request_count: 0,
    display_in_currency: false,
    quota_per_unit: 500000,
  }

  const records: BillingRecord[] = recordsData?.data || []

  // 如果开启了货币显示，将配额转换为金额
  const fmtQuota = (q: number) => {
    if (stats.display_in_currency && stats.quota_per_unit > 0) {
      return '$' + (q / stats.quota_per_unit).toFixed(2)
    }
    return q.toLocaleString()
  }

  const getTypeText = (type: string) => {
    const typeMap: Record<string, string> = {
      'topup': t('Top-up'),
      'payment': t('Payment'),
      'refund': t('Refund'),
      'subscription': t('Subscription'),
      'usage': t('Usage'),
    }
    return typeMap[type] || type
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-500'
      case 'pending':
        return 'bg-yellow-500'
      case 'failed':
        return 'bg-red-500'
      default:
        return 'bg-gray-500'
    }
  }

  const formatTime = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString()
  }

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Billing')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('View and manage your billing history and balance')}</p>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">{t('Remaining Quota')}</p>
                <p className="text-2xl font-bold mt-1">{fmtQuota(stats.remain_quota)}</p>
              </div>
              <div className="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center">
                <DollarSign className="h-5 w-5 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">{t('Total Quota')}</p>
                <p className="text-2xl font-bold mt-1 text-green-600">{fmtQuota(stats.total_quota)}</p>
              </div>
              <div className="w-10 h-10 rounded-full bg-green-50 flex items-center justify-center">
                <TrendingUp className="h-5 w-5 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">{t('Used Quota')}</p>
                <p className="text-2xl font-bold mt-1 text-red-600">{fmtQuota(stats.used_quota)}</p>
              </div>
              <div className="w-10 h-10 rounded-full bg-red-50 flex items-center justify-center">
                <TrendingDown className="h-5 w-5 text-red-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">{t('Requests')}</p>
                <p className="text-2xl font-bold mt-1 text-yellow-600">{stats.request_count.toLocaleString()}</p>
              </div>
              <div className="w-10 h-10 rounded-full bg-yellow-50 flex items-center justify-center">
                <CreditCard className="h-5 w-5 text-yellow-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Transaction History */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <div className="flex items-center gap-2">
            <History className="h-6 w-6 text-blue-600" />
            <CardTitle>{t('Transaction History')}</CardTitle>
          </div>
          <CardDescription>
            {t('Recent billing transactions')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {recordsLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4 py-3 border-b border-dashed">
                  <div className="h-10 w-10 rounded-full bg-muted animate-pulse" />
                  <div className="flex-1 space-y-1">
                    <div className="h-4 w-32 animate-pulse rounded bg-muted" />
                    <div className="h-3 w-48 animate-pulse rounded bg-muted" />
                  </div>
                  <div className="h-6 w-24 animate-pulse rounded bg-muted" />
                </div>
              ))}
            </div>
          ) : records.length === 0 ? (
            <div className="text-center py-8">
              <History className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground">{t('No transactions yet')}</p>
            </div>
          ) : (
            <div className="space-y-3">
              {records.map((record) => (
                <div key={record.id} className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4 py-3 border-b border-dashed last:border-0">
                  <div className={`w-10 h-10 rounded-full flex items-center justify-center ${
                    record.amount > 0 ? 'bg-green-50' : 'bg-red-50'
                  }`}>
                    {record.amount > 0 ? (
                      <TrendingUp className="h-5 w-5 text-green-600" />
                    ) : (
                      <TrendingDown className="h-5 w-5 text-red-600" />
                    )}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{record.action}</span>
                      <Badge
                        variant="secondary"
                        className={`${getStatusColor(record.status)} text-white text-xs`}
                      >
                        {record.status}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">{record.remark}</p>
                  </div>
                  <div className="text-left sm:text-right">
                    <p className={`font-medium ${record.amount > 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {record.amount > 0 ? '+' : ''}${record.amount.toFixed(2)}
                    </p>
                    <p className="text-xs text-muted-foreground">{formatTime(record.created_at)}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
