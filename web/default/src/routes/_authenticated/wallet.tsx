import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  getBalance, getTopUpInfo, getTopUpList, getMyWithdrawals, getMyWithdrawable,
  requestStripeTopUp, requestEpayTopUp, requestCreemTopUp, requestWaffoTopUp, requestBinanceTopUp,
  submitWithdrawal, getMyCommissionRecords, type BalanceInfo
} from '@/lib/api-extended'
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
  const { t } = useT()
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
      const res = await getTopUpInfo()
      return res
    },
    retry: false,
    staleTime: 30 * 1000,
  })

  const { data: topupHistory } = useQuery({
    queryKey: ['topup-history'],
    queryFn: async () => {
      const res = await getTopUpList()
      return res
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
    queryFn: () => getMyCommissionRecords(),
    retry: false,
    staleTime: 30 * 1000,
  })

  const { data: myWithdrawals } = useQuery({
    queryKey: ['withdrawals'],
    queryFn: () => getMyWithdrawals(),
    retry: false,
    staleTime: 30 * 1000,
  })

  const { data: balanceData } = useQuery({
    queryKey: ['balance'],
    queryFn: () => getBalance(),
    retry: false,
    staleTime: 10 * 1000,
  })

  const { data: withdrawableData } = useQuery({
    queryKey: ['withdrawable'],
    queryFn: () => getMyWithdrawable(),
    retry: false,
    staleTime: 10 * 1000,
  })

  const [withdrawAmount, setWithdrawAmount] = useState('')
  const [withdrawAccount, setWithdrawAccount] = useState('')
  const [withdrawing, setWithdrawing] = useState(false)

  // 后端返回 enable_* 布尔标记 + pay_methods 数组
  // 将后端格式转成前端支付方法名数组
  const paymentMethods: string[] = (() => {
    if (!topupInfo?.success || !topupInfo.data) return []
    const info = topupInfo.data as Record<string, unknown>
    const methods: string[] = []
    if (info.enable_online_topup) methods.push('epay')
    if (info.enable_stripe_topup) methods.push('stripe')
    if (info.enable_creem_topup) methods.push('creem')
    if (info.enable_waffo_topup) methods.push('waffo')
    if (info.enable_binance_topup) methods.push('binance')
    return methods
  })()
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
      const apiMap: Record<string, (amt: number) => Promise<any>> = {
        stripe: requestStripeTopUp,
        epay: requestEpayTopUp,
        creem: requestCreemTopUp,
        waffo: requestWaffoTopUp,
        binance: requestBinanceTopUp,
      }
      const fn = apiMap[method]
      if (!fn) { toast.error(t('Unknown payment method')); return }
      const data = await fn(Number(amount))
      if (data.success) {
        const paymentUrl = (data.data as any)?.pay_url || (data.data as any)?.checkout_url || (data.data as any)?.url
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
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Wallet')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('View and manage your quota balance')}</p>
      </div>

      {balanceData?.success && (
        <Card className="w-full border-yellow-200 dark:border-yellow-800 bg-gradient-to-br from-yellow-50/50 to-transparent bg-white/80 backdrop-blur-xl rounded-xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="h-5 w-5 text-yellow-600" />
              {t('Cash Balance')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl sm:text-3xl font-bold tracking-tight text-yellow-600">
              &yen;{Number(balanceData.data.balance_yuan || 0).toFixed(2)}
            </div>
            <div className="mt-2 grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-muted-foreground">{t('Available for API consumption')}</p>
                <p className="text-sm font-mono">{balanceData.data.balance} {t('cents')}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <details className="w-full">
        <summary className="text-xs text-muted-foreground cursor-pointer hover:text-foreground select-none">
          {t('Legacy Quota')} ({remaining.toLocaleString()} {t('remaining')})
        </summary>
        <div className="mt-2 p-3 rounded-xl border bg-card text-xs space-y-1">
          <div className="flex justify-between">
            <span className="text-muted-foreground">{t('Total Quota')}</span>
            <span className="font-mono">{quota.toLocaleString()}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">{t('Used Quota')}</span>
            <span className="font-mono">{usedQuota.toLocaleString()}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">{t('Usage Rate')}</span>
            <span className="font-mono">{quota > 0 ? ((usedQuota / quota) * 100).toFixed(1) : 0}%</span>
          </div>
        </div>
      </details>

      {balanceData?.success && balanceData.data.logs?.length > 0 && (
        <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">{t('Balance History')}</CardTitle>
          </CardHeader>
          <CardContent className="max-h-48 overflow-y-auto">
            <div className="space-y-1">
              {balanceData.data.logs.slice(0, 15).map((log: any, i: number) => (
                <div key={log.id || i} className="flex justify-between items-center text-xs border-b pb-1 last:border-0">
                  <span className={log.amount > 0 ? 'text-green-600' : 'text-red-600'}>
                    {log.amount > 0 ? '+' : ''}{log.amount}&#20998;
                  </span>
                  <span className="text-muted-foreground truncate ml-2 max-w-[min(20vw,200px)]">
                    {log.remark || log.type}
                  </span>
                  <span className="text-muted-foreground ml-2">
                    {log.created_at ? new Date(log.created_at * 1000).toLocaleDateString() : ''}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {user?.user_type !== 'provider' && (
        <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border border-purple-200 dark:border-purple-800">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Wallet className="h-4 w-4 text-purple-600" />
              {t('Become a Provider')}
            </CardTitle>
            <CardDescription>
              {t('Add your own API channels and earn revenue')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-4">
              {t('As a provider, you can add AI/Quantum channels, set your own pricing, and earn from API calls. Platform fee: 5% on revenue + 1% per transaction.')}
            </p>
            <Button onClick={async () => {
              try {
                const res = await fetch('/api/user/self/upgrade', { method: 'POST' })
                const data = await res.json()
                if (data.success) {
                  toast.success(data.message || t('Upgraded successfully'))
                  queryClient.invalidateQueries({ queryKey: ['self'] })
                } else {
                  toast.error(data.message || t('Upgrade failed'))
                }
              } catch {
                toast.error(t('Upgrade failed'))
              }
            }}>
              {t('Upgrade Now')}
            </Button>
          </CardContent>
        </Card>
      )}

      {user?.user_type === 'provider' && withdrawableData?.success && withdrawableData.data.available > 0 && (
        <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border border-blue-200 dark:border-blue-800">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Wallet className="h-4 w-4" />
              {t('Supplier Earnings')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
              <div>
                <p className="text-xs text-muted-foreground">{t('Total Earned')}</p>
                <p className="text-lg font-semibold text-green-600">
                  &yen;{Number(withdrawableData.data.total_earned_yuan || 0).toFixed(2)}
                </p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('Pending Fee')}</p>
                <p className="text-lg font-semibold text-orange-600">
                  &yen;{Number(withdrawableData.data.pending_fee_yuan || 0).toFixed(2)}
                </p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('Available')}</p>
                <p className="text-lg font-semibold text-blue-600">
                  &yen;{Number(withdrawableData.data.available_yuan || 0).toFixed(2)}
                </p>
              </div>
            </div>
            <form onSubmit={async (e) => {
              e.preventDefault()
              if (!withdrawAmount || Number(withdrawAmount) <= 0) {
                toast.error(t('Please enter a valid amount'))
                return
              }
              if (!withdrawAccount.trim()) {
                toast.error(t('Please enter your account info'))
                return
              }
              setWithdrawing(true)
              try {
                const res = await fetch('/api/user/self/withdraw', {
                  method: 'POST',
                  headers: { 'Content-Type': 'application/json' },
                  body: JSON.stringify({
                    amount: Math.round(Number(withdrawAmount) * 100),
                    bank_info: withdrawAccount,
                  }),
                })
                const data = await res.json()
                if (data.success) {
                  toast.success(t('Withdrawal request submitted'))
                  setWithdrawAmount('')
                  setWithdrawAccount('')
                  queryClient.invalidateQueries({ queryKey: ['withdrawable'] })
                  queryClient.invalidateQueries({ queryKey: ['withdrawals'] })
                } else {
                  toast.error(data.message || t('Withdrawal failed'))
                }
              } catch {
                toast.error(t('Withdrawal failed'))
              }
              setWithdrawing(false)
            }} className="flex flex-col sm:flex-row gap-2">
              <div className="flex-1 space-y-2">
                <Input
                  type="number"
                  min="1"
                  step="0.01"
                  placeholder={t('Amount (\u00a5)')}
                  value={withdrawAmount}
                  onChange={(e) => setWithdrawAmount(e.target.value)}
                />
                <Input
                  placeholder={t('Alipay / WeChat / Bank account')}
                  value={withdrawAccount}
                  onChange={(e) => setWithdrawAccount(e.target.value)}
                />
              </div>
              <Button type="submit" disabled={withdrawing || !withdrawAmount || !withdrawAccount}>
                {withdrawing ? t('Submitting...') : t('Withdraw')}
              </Button>
            </form>
            <p className="text-xs text-muted-foreground mt-2">
              {t('Minimum withdrawal: \u00a51. Platform fee (5%) will be deducted automatically on settlement.')}
            </p>
          </CardContent>
        </Card>
      )}

      <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
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

      <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
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
                    className="flex flex-col p-4 rounded-xl border bg-card hover:bg-accent/50 transition-colors gap-3"
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

      <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
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
        <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
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
