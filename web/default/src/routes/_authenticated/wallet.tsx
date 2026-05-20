import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Wallet, Copy, RefreshCw, CreditCard, TrendingUp, Banknote, History, ArrowUpRight, DollarSign, Building, Landmark, Gift, Users } from 'lucide-react'
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
    <div className=" w-full p-4 sm:p-6 space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
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

      {/* Corporate Transfer */}
      <CompanyPaymentCard />

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
    </div>
  )
}

// ==================== 企业转账组件 ====================

function CompanyPaymentCard() {
  const { t } = useTranslation()
  const { data: optionsData } = useQuery({
    queryKey: ['system-options-company'],
    queryFn: async () => {
      const res = await fetch('/api/option')
      if (!res.ok) return []
      const json = await res.json()
      return json?.data || []
    },
    staleTime: 60 * 1000,
    retry: false,
  })
  const getOption = (key: string, fallback: string) => {
    const opt = (optionsData as any[])?.find((o: any) => o.key === key)
    return opt?.value || fallback
  }
  const companyInfo = {
    company_name: getOption('CompanyName', '深圳市中科劲纬智能有限公司'),
    tax_id: getOption('CompanyTaxId', '91440300MA5GH45W8C'),
    address: getOption('CompanyAddress', '深圳市宝安区石岩街道塘头社区塘头大道33号东海创意园205A'),
    phone: getOption('CompanyPhone', '15920005303'),
    bank_name: getOption('CompanyBank', '深圳农村商业银行股份有限公司应人石支行'),
    bank_account: getOption('CompanyBankAccount', '000396168236'),
    alipay_qr_url: getOption('CompanyAlipayQr', '/payment/alipay-qr.jpg'),
  }
  if (!companyInfo.bank_account) return null

  const copyText = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.success(t('Copied'))
  }

  return (
    <Card className="w-full max-w-2xl border-blue-200 dark:border-blue-800">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Building className="h-4 w-4 text-blue-600" />
          {t('Corporate Transfer')}
        </CardTitle>
        <CardDescription>{t('For B2B payments via bank transfer or Alipay')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 公司开票信息 */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm">
          <div className="flex justify-between gap-2 p-2 rounded bg-muted/50">
            <span className="text-muted-foreground whitespace-nowrap">{t('Company')}</span>
            <span className="font-medium text-right">{companyInfo.company_name}</span>
          </div>
          <div className="flex justify-between gap-2 p-2 rounded bg-muted/50">
            <span className="text-muted-foreground whitespace-nowrap">{t('Tax ID')}</span>
            <span className="font-mono text-right">{companyInfo.tax_id}</span>
          </div>
          <div className="flex justify-between gap-2 p-2 rounded bg-muted/50 sm:col-span-2">
            <span className="text-muted-foreground whitespace-nowrap">{t('Address')}</span>
            <span className="text-right">{companyInfo.address}</span>
          </div>
          <div className="flex justify-between gap-2 p-2 rounded bg-muted/50">
            <span className="text-muted-foreground whitespace-nowrap">{t('Phone')}</span>
            <span className="font-mono text-right">{companyInfo.phone}</span>
          </div>
        </div>

        {/* 银行账户 */}
        <div className="p-3 rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50/50 dark:bg-blue-950/20 space-y-2">
          <div className="flex items-center gap-2 text-sm font-medium text-blue-700 dark:text-blue-300">
            <Landmark className="h-4 w-4" />
            {t('Bank Account')}
          </div>
          <div className="flex items-center justify-between gap-2">
            <div>
              <p className="text-sm">{companyInfo.bank_name}</p>
              <p className="text-lg font-mono font-bold">{companyInfo.bank_account}</p>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => copyText(companyInfo.bank_account)}
            >
              <Copy className="h-3 w-3 mr-1" />
              {t('Copy')}
            </Button>
          </div>
        </div>

        {/* 支付宝收款码 */}
        {companyInfo.alipay_qr_url && (
          <div className="flex flex-col items-center gap-2 p-3 rounded-lg border">
            <span className="text-sm font-medium text-muted-foreground">{t('Alipay Business QR')}</span>
            <img
              src={companyInfo.alipay_qr_url}
              alt="Alipay QR"
              className="w-48 h-48 object-contain rounded"
            />
            <p className="text-xs text-muted-foreground">
              {t('Scan with Alipay to pay')}
            </p>
          </div>
        )}

        <p className="text-xs text-muted-foreground italic">
          {t('After transfer, contact admin with transaction receipt to credit your quota.')}
        </p>
      </CardContent>
    </Card>
  )
}
