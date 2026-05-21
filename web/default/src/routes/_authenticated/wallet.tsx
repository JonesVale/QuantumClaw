import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Wallet, Copy, RefreshCw, CreditCard, TrendingUp, Banknote, History, ArrowUpRight, DollarSign } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/wallet')({
  component: WalletPage,
})

function WalletPage() {
  const { t } = useTranslation()
  const [redemptionCode, setRedemptionCode] = useState('')
  const [amounts, setAmounts] = useState<Record<string, string>>({})
  const queryClient = useQueryClient()

  const { data: selfData, isLoading } = useQuery({
    queryKey: ['self'],
    queryFn: async () => {
      const res = await fetch('/api/user/self', {
        headers: { 'Cache-Control': 'no-store' },
      })
      if (!res.ok) throw new Error('Failed')
      return res.json()
    },
    retry: false,
    staleTime: 10 * 1000,
  })

  const { data: topupInfo } = useQuery({
    queryKey: ['topup-info'],
    queryFn: async () => {
      const res = await fetch('/api/user/self/topup/info')
      if (!res.ok) return { paymentMethods: [] }
      return res.json()
    },
    retry: false,
    staleTime: 30 * 1000,
  })

  const { data: topupHistory } = useQuery({
    queryKey: ['topup-history'],
    queryFn: async () => {
      const res = await fetch('/api/user/self/topup/list')
      if (!res.ok) return { data: [] }
      return res.json()
    },
    retry: false,
    staleTime: 10 * 1000,
  })

  const { data: myAffCode } = useQuery({
    queryKey: ['aff-code'],
    queryFn: async () => {
      const res = await fetch('/api/user/self/aff')
      if (!res.ok) return null
      return res.json()
    },
    retry: false,
    staleTime: 60 * 1000,
  })

  const { data: myCommission } = useQuery({
    queryKey: ['commission'],
    queryFn: async () => {
      const res = await fetch('/api/user/self/commission')
      if (!res.ok) return null
      return res.json()
    },
    retry: false,
    staleTime: 30 * 1000,
  })

  const { data: myWithdrawals } = useQuery({
    queryKey: ['withdrawals'],
    queryFn: async () => {
      const res = await fetch('/api/user/self/withdrawals')
      if (!res.ok) return null
      return res.json()
    },
    retry: false,
    staleTime: 30 * 1000,
  })

  const paymentMethods: string[] = topupInfo?.paymentMethods || []
  const historyItems: any[] = topupHistory?.data || []

  const methodConfig: Record<string, { label: string; icon: any }> = {
    stripe: { label: 'Stripe', icon: CreditCard },
    epay: { label: 'Epay', icon: Banknote },
    creem: { label: 'Creem', icon: Wallet },
    waffo: { label: 'Waffo', icon: ArrowUpRight },
    binance: { label: t('Binance Pay'), icon: DollarSign },
  }

  const handleTopup = async (method: string) => {
    const amount = amounts[method]
    if (!amount || isNaN(Number(amount)) || Number(amount) <= 0) {
      toast.error(t('Please enter a valid amount'))
      return
    }
    try {
      const res = await fetch(`/api/user/self/topup/${method}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount: Number(amount) }),
      })
      const data = await res.json()
      if (data.success) {
        const paymentUrl = data.payment_url || data.data?.payment_url
        if (paymentUrl) {
          window.open(paymentUrl, '_blank')
        }
        toast.success(t('Top-up initiated'))
        queryClient.invalidateQueries({ queryKey: ['topup-history'] })
        queryClient.invalidateQueries({ queryKey: ['self'] })
      } else {
        toast.error(data.message || t('Top-up failed'))
      }
    } catch {
      toast.error(t('Top-up failed'))
    }
  }

  const user = selfData?.data
  const quota = user?.quota || 0
  const usedQuota = user?.used_quota || 0
  const remaining = quota - usedQuota

  const handleRedeem = async () => {
    if (!redemptionCode.trim()) return
    try {
      const res = await fetch('/api/redemption', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: redemptionCode.trim() }),
      })
      const data = await res.json()
      if (data.success) {
        toast.success(t('Redemption successful! Quota added.'))
        setRedemptionCode('')
        window.location.reload()
      } else {
        toast.error(data.message || t('Redemption failed'))
      }
    } catch {
      toast.error(t('Redemption failed'))
    }
  }

  return (
    <div className="mx-auto max-w-[min(96vw,1600px)] w-full p-4 sm:p-6 space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-xl sm:text-2xl lg:text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('Wallet')}
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            {t('View and manage your quota balance')}
          </p>
        </div>
      </div>

      {/* Balance Card */}
      <Card className="w-full max-w-2xl border-primary/20 bg-gradient-to-br from-primary/5 to-transparent">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Wallet className="h-5 w-5" />
            {t('Quota Balance')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <Skeleton className="h-12 w-48" />
          ) : (
            <div className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight">
              {remaining.toLocaleString()}
            </div>
          )}
          <div className="mt-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            <div>
              <p className="text-xs text-muted-foreground">{t('Total Quota')}</p>
              <p className="text-lg font-semibold">{quota.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">{t('Used Quota')}</p>
              <p className="text-lg font-semibold">{usedQuota.toLocaleString()}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Redemption */}
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CreditCard className="h-4 w-4" />
            {t('Redeem Code')}
          </CardTitle>
          <CardDescription>{t('Enter a redemption code to add quota')}</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={(e) => {
              e.preventDefault()
              handleRedeem()
            }}
            className="flex flex-col sm:flex-row gap-2"
          >
            <Input
              value={redemptionCode}
              onChange={(e) => setRedemptionCode(e.target.value)}
              placeholder={t('Enter redemption code')}
              className="flex-1"
            />
            <Button type="submit" disabled={!redemptionCode.trim()}>
              {t('Redeem')}
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Payment Methods */}
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Banknote className="h-4 w-4" />
            {t('Top-up')}
          </CardTitle>
          <CardDescription>{t('Select a payment method to add quota')}</CardDescription>
        </CardHeader>
        <CardContent>
          {paymentMethods.length === 0 ? (
            <p className="text-sm text-muted-foreground italic">
              {t('No payment methods available')}
            </p>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {paymentMethods.map((method: string) => {
                const config = methodConfig[method]
                if (!config) return null
                const Icon = config.icon
                return (
                  <div
                    key={method}
                    className="flex flex-col p-4 rounded-lg border bg-card hover:bg-accent/50 transition-colors gap-3"
                  >
                    <div className="flex items-center gap-2">
                      <Icon className="h-5 w-5 text-primary" />
                      <span className="font-medium">{config.label}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Input
                        type="number"
                        min="1"
                        placeholder={t('Amount')}
                        value={amounts[method] || ''}
                        onChange={(e) =>
                          setAmounts((prev) => ({ ...prev, [method]: e.target.value }))
                        }
                        className="flex-1"
                      />
                      <Button
                        size="sm"
                        onClick={() => handleTopup(method)}
                        disabled={!amounts[method] || Number(amounts[method]) <= 0}
                      >
                        {t('Pay')}
                      </Button>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Top-up History */}
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <History className="h-4 w-4" />
            {t('Top-up History')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {historyItems.length === 0 ? (
            <p className="text-sm text-muted-foreground italic">
              {t('No top-up history yet')}
            </p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-2 font-medium text-muted-foreground">{t('Method')}</th>
                    <th className="text-right py-2 px-2 font-medium text-muted-foreground">{t('Amount')}</th>
                    <th className="text-right py-2 px-2 font-medium text-muted-foreground">{t('Status')}</th>
                    <th className="text-right py-2 px-2 font-medium text-muted-foreground">{t('Date')}</th>
                  </tr>
                </thead>
                <tbody>
                  {historyItems.map((item: any, idx: number) => (
                    <tr key={item.id || idx} className="border-b last:border-0">
                      <td className="py-2 px-2">{item.method || item.channel || '-'}</td>
                      <td className="py-2 px-2 text-right font-mono">
                        {item.amount?.toLocaleString() || '-'}
                      </td>
                      <td className="py-2 px-2 text-right">
                        <Badge
                          variant={
                            item.status === 'completed' || item.status === 'success'
                              ? 'default'
                              : item.status === 'pending'
                                ? 'secondary'
                                : 'destructive'
                          }
                        >
                          {item.status || '-'}
                        </Badge>
                      </td>
                      <td className="py-2 px-2 text-right text-muted-foreground">
                        {item.created_at
                          ? new Date(item.created_at).toLocaleDateString()
                          : '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Usage Summary */}
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4" />
            {t('Usage Summary')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div className="flex flex-col sm:flex-row justify-between gap-1 sm:gap-4">
            <span className="text-muted-foreground">{t('Total Requests')}</span>
            <span className="font-medium">{user?.request_count?.toLocaleString() || 0}</span>
          </div>
          <div className="flex flex-col sm:flex-row justify-between gap-1 sm:gap-4">
            <span className="text-muted-foreground">{t('Usage Rate')}</span>
            <span className="font-medium">
              {quota > 0 ? ((usedQuota / quota) * 100).toFixed(1) : 0}%
            </span>
          </div>
        </CardContent>
      </Card>

      {/* Commission / Affiliate Section */}
      <div className="grid gap-6 md:grid-cols-3 border-t pt-6">
        <Card className="border-green-200 dark:border-green-800">
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center gap-2 text-sm">
              <TrendingUp className="h-4 w-4 text-green-600" />
              {t('Commission')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-green-600">
              {Number(myCommission?.data?.total_commission || 0).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground mt-1">{t('Earned from referrals')}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">{t('Invite Link')}</CardTitle>
          </CardHeader>
          <CardContent>
            {myAffCode?.data?.aff_code ? (
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <code className="flex-1 rounded bg-muted px-2 py-1 text-xs font-mono truncate">{myAffCode.data.aff_code}</code>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      navigator.clipboard.writeText(`${window.location.origin}/sign-in?aff=${myAffCode.data.aff_code}`)
                      toast.success(t('Copied'))
                    }}
                  >
                    <Copy className="h-3 w-3" />
                  </Button>
                </div>
                <p className="text-xs text-muted-foreground break-all">
                  {window.location.origin}/sign-in?aff={myAffCode.data.aff_code}
                </p>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">{t('Loading...')}</p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">{t('Recent Commissions')}</CardTitle>
          </CardHeader>
          <CardContent className="max-h-36 overflow-y-auto">
            {(myCommission?.data?.records || []).length > 0 ? (
              myCommission.data.records.slice(0, 8).map((r: any) => (
                <div key={r.id} className="flex justify-between items-center text-xs border-b py-1">
                  <span className={r.type === 'register' ? 'text-blue-500' : 'text-green-500'}>
                    {r.type === 'register' ? t('Referral') : t('Usage')}
                  </span>
                  <span className="font-medium">+{Number(r.amount).toLocaleString()}</span>
                </div>
              ))
            ) : (
              <p className="text-xs text-muted-foreground">{t('No records')}</p>
            )}
          </CardContent>
        </Card>
      </div>

      {(myWithdrawals?.data || []).length > 0 && (
        <Card className="w-full max-w-2xl">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">{t('Withdrawals')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-1">
            {(myWithdrawals?.data || []).slice(0, 8).map((w: any, i: number) => (
              <div key={i} className="flex justify-between items-center text-xs border-b pb-1">
                <span className="font-medium">{Number(w.amount).toLocaleString()}</span>
                <span>{new Date(w.created_at).toLocaleDateString()}</span>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
