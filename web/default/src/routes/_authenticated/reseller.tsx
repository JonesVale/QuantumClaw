import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Plus, Key, ExternalLink, Copy, RefreshCw, TrendingUp, PieChart, DollarSign } from 'lucide-react'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart as RePieChart, Pie, Cell, Legend
} from 'recharts'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import { getTransactions, type TransactionItem } from '@/lib/api-extended'
import apiClient from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/_authenticated/reseller')({
  component: ResellerPortal,
})

function ResellerPortal() {
  const { t } = useT()
  const { auth } = useAuthStore()
  const queryClient = useQueryClient()
  const [copied, setCopied] = useState(false)

  const userId = auth.user?.id
  const affiliateLink = typeof window !== 'undefined'
    ? `${window.location.origin}/sign-in?aff=${userId}`
    : ''

  const copyLink = () => {
    navigator.clipboard.writeText(affiliateLink)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
    toast.success(t('copied'))
  }

  const { data: promoterTxns, isLoading: promoterLoading } = useQuery({
    queryKey: ['transactions-promoter', userId],
    queryFn: () => getTransactions({ promoter_id: userId, page_size: 10 }),
    enabled: !!userId,
    staleTime: 15_000,
  })

  const { data: ownerTxns, isLoading: ownerLoading } = useQuery({
    queryKey: ['transactions-owner', userId],
    queryFn: () => getTransactions({ channel_owner_id: userId, page_size: 10 }),
    enabled: !!userId,
    staleTime: 15_000,
  })

  const promotedTxns = promoterTxns?.data?.transactions || []
  const ownedTxns = ownerTxns?.data?.transactions || []

  const totalCommission = promotedTxns.reduce((s, t) => s + t.commission_amount, 0)
  const totalRevenue = ownedTxns.reduce((s, t) => s + t.unified_cost, 0)
  const totalFallbacks = promotedTxns.filter(t => t.is_fallback).length

  const [withdrawing, setWithdrawing] = useState(false)
  const handleWithdraw = async () => {
    setWithdrawing(true)
    try {
      const res = await apiClient.post('/api/reseller/withdraw')
      if (res.data?.success) { toast.success(t('withdraw_sent')); refetch() }
      else { toast.error(res.data?.message || t('withdraw_failed')) }
    } catch { toast.error(t('withdraw_failed')) }
    finally { setWithdrawing(false) }
  }

  const { data: balanceData } = useQuery({
    queryKey: ['reseller-balance', userId],
    queryFn: async () => { const r = await apiClient.get('/api/reseller/balance'); return r.data?.data },
    enabled: !!userId,
  })
  const withdrawableBalance = balanceData?.balance || 0

  const { data: statsData } = useQuery({
    queryKey: ['reseller-stats', userId],
    queryFn: async () => { const r = await apiClient.get('/api/reseller/stats?days=30'); return r.data?.data },
    enabled: !!userId,
    staleTime: 60_000,
  })
  const dailyStats = statsData?.daily || []
  const modelStats = statsData?.by_model || []

  const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16']

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('reseller_portal')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('reseller_portal_desc')}</p>
      </div>

      <div className="grid gap-4 grid-cols-1 sm:grid-cols-3">
        <Card>
          <CardContent className="py-4 text-center">
            <p className="text-2xl font-bold text-blue-600">${totalCommission.toFixed(2)}</p>
            <p className="text-xs text-muted-foreground mt-1">{t('commission_earned')}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4 text-center">
            <p className="text-2xl font-bold text-emerald-600">${totalRevenue.toFixed(2)}</p>
            <p className="text-xs text-muted-foreground mt-1">{t('key_revenue')}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4 text-center">
            <p className="text-2xl font-bold text-orange-600">{totalFallbacks}</p>
            <p className="text-xs text-muted-foreground mt-1">{t('fallback_count')}</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 grid-cols-1 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-blue-600" />
              {t('revenue_trend')}
            </CardTitle>
          </CardHeader>
          <CardContent className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={dailyStats}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="date" tick={{ fontSize: 10 }} />
                <YAxis tick={{ fontSize: 10 }} />
                <Tooltip />
                <Bar dataKey="total_amount" fill="#3b82f6" name="Amount" />
                <Bar dataKey="commission" fill="#10b981" name="Commission" />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <PieChart className="h-4 w-4 text-purple-600" />
              {t('model_distribution')}
            </CardTitle>
          </CardHeader>
          <CardContent className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <RePieChart>
                <Pie
                  data={modelStats.slice(0, 8)}
                  dataKey="total_amount"
                  nameKey="model_name"
                  cx="50%"
                  cy="50%"
                  outerRadius={70}
                  label={({ model_name, percent }) => `${model_name} ${(percent * 100).toFixed(0)}%`}
                >
                  {modelStats.slice(0, 8).map((_, idx) => (
                    <Cell key={idx} fill={COLORS[idx % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </RePieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <ExternalLink className="h-4 w-4" />
            {t('affiliate_link')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2">
            <Input value={affiliateLink} readOnly className="font-mono text-xs" />
            <Button variant="outline" size="sm" className="shrink-0 gap-2" onClick={copyLink}>
              <Copy className="h-4 w-4" />
              {copied ? t('copied') : t('copy')}
            </Button>
          </div>
          <p className="text-xs text-muted-foreground mt-2">{t('affiliate_link_desc')}</p>
        </CardContent>
      </Card>

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Key className="h-4 w-4 text-blue-600" />
            {t('promotion_earnings')}
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3">{t('time')}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3">{t('model')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3">{t('amount')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3">{t('commission')}</th>
                  <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3">{t('type')}</th>
                </tr>
              </thead>
              <tbody>
                {promotedTxns.map((tx) => (
                  <tr key={tx.id} className="border-b border-muted/50 text-sm">
                    <td className="px-4 py-3 text-xs text-muted-foreground">{dayjs(tx.created_time * 1000).format('MM-DD HH:mm')}</td>
                    <td className="px-4 py-3 font-medium">{tx.model_name}</td>
                    <td className="px-4 py-3 text-right font-mono">${tx.total_amount.toFixed(4)}</td>
                    <td className="px-4 py-3 text-right font-mono text-blue-600">${tx.commission_amount.toFixed(4)}</td>
                    <td className="px-4 py-3 text-center">
                      {tx.is_fallback
                        ? <Badge variant="secondary" className="text-[10px]">{t('fallback')}</Badge>
                        : <Badge variant="default" className="text-[10px]">{t('direct')}</Badge>
                      }
                    </td>
                  </tr>
                ))}
                {promotedTxns.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-sm text-muted-foreground">{t('no_transactions')}</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Key className="h-4 w-4 text-emerald-600" />
            {t('key_earnings')}
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3">{t('time')}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3">{t('model')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3">{t('amount')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3">{t('unified_cost')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3">{t('cost')}</th>
                </tr>
              </thead>
              <tbody>
                {ownedTxns.map((tx) => (
                  <tr key={tx.id} className="border-b border-muted/50 text-sm">
                    <td className="px-4 py-3 text-xs text-muted-foreground">{dayjs(tx.created_time * 1000).format('MM-DD HH:mm')}</td>
                    <td className="px-4 py-3 font-medium">{tx.model_name}</td>
                    <td className="px-4 py-3 text-right font-mono">${tx.total_amount.toFixed(4)}</td>
                    <td className="px-4 py-3 text-right font-mono text-emerald-600">${tx.unified_cost.toFixed(4)}</td>
                    <td className="px-4 py-3 text-right font-mono text-muted-foreground">${(tx.unified_cost * 0.6).toFixed(4)}</td>
                  </tr>
                ))}
                {ownedTxns.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-sm text-muted-foreground">{t('no_transactions')}</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <DollarSign className="h-4 w-4 text-blue-600" />
            {t('withdraw')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-muted-foreground">{t('withdrawable')}</p>
              <p className="text-2xl font-bold">${withdrawableBalance.toFixed(2)}</p>
            </div>
            <Button
              onClick={handleWithdraw}
              disabled={withdrawing || withdrawableBalance <= 0}
              className="gap-2"
            >
              <DollarSign className="h-4 w-4" />
              {t('withdraw_request')}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
