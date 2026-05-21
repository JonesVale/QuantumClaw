import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  BarChart3,
  RefreshCw,
  Search,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { getChannelTypes } from '@/lib/api-extended'
import apiClient from '@/lib/api'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/profit')({
  component: ProfitPage,
})

interface ProfitItem {
  id: number
  name: string
  type: number
  used_quota: number
  cost_per_unit: number
  sell_price_rate: number
  sell_price: number
  total_cost: number
  total_revenue: number
  profit: number
  margin: number
}

async function getChannelProfit(): Promise<ProfitItem[]> {
  const res = await apiClient.get('/api/channel/profit')
  return res.data?.data || []
}

function ProfitPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')

  const { data: profitData, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['channelProfit'],
    queryFn: getChannelProfit,
    refetchInterval: 60_000,
  })

  const { data: typeMap } = useQuery({
    queryKey: ['channelTypes'],
    queryFn: getChannelTypes,
    staleTime: 10 * 60 * 1000,
  })

  const filtered = useMemo(() => {
    if (!profitData) return []
    return profitData.filter((item) => {
      if (!search) return true
      return item.name.toLowerCase().includes(search.toLowerCase())
    })
  }, [profitData, search])

  const stats = useMemo(() => {
    if (!profitData || profitData.length === 0) {
      return { totalRevenue: 0, totalCost: 0, totalProfit: 0, avgMargin: 0, profitable: 0, negative: 0 }
    }
    const totalRevenue = profitData.reduce((s, i) => s + i.total_revenue, 0)
    const totalCost = profitData.reduce((s, i) => s + i.total_cost, 0)
    const totalProfit = profitData.reduce((s, i) => s + i.profit, 0)
    const active = profitData.filter((i) => i.total_revenue > 0 || i.total_cost > 0)
    const avgMargin = active.length > 0
      ? active.reduce((s, i) => s + i.margin, 0) / active.length
      : 0
    const profitable = profitData.filter((i) => i.profit > 0).length
    const negative = profitData.filter((i) => i.profit < 0).length
    return { totalRevenue, totalCost, totalProfit, avgMargin, profitable, negative }
  }, [profitData])

  return (
    <div className="p-4 sm:p-6 w-full space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-emerald-600 via-teal-600 to-cyan-600 bg-clip-text text-transparent">
            {t('Channel Profit')}
          </h1>
          <p className="text-muted-foreground mt-1 sm:mt-2 text-sm sm:text-base">
            {t('Revenue, cost and margin analysis per channel')}
          </p>
        </div>
        <Button variant="outline" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={cn('mr-2 h-4 w-4', isFetching && 'animate-spin')} />
          {t('Refresh')}
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground flex items-center gap-2">
              <DollarSign className="h-4 w-4 text-green-500" />
              {t('Total Revenue')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{stats.totalRevenue.toFixed(2)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground flex items-center gap-2">
              <TrendingDown className="h-4 w-4 text-red-500" />
              {t('Total Cost')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-red-600">{stats.totalCost.toFixed(2)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-emerald-500" />
              {t('Total Profit')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className={cn('text-2xl font-bold', stats.totalProfit >= 0 ? 'text-emerald-600' : 'text-red-600')}>
              {stats.totalProfit.toFixed(2)}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground flex items-center gap-2">
              <BarChart3 className="h-4 w-4 text-blue-500" />
              {t('Avg Margin')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{stats.avgMargin.toFixed(1)}%</p>
            <p className="text-xs text-muted-foreground mt-1">
              {stats.profitable} {t('profitable')} · {stats.negative} {t('negative')}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Profit Table */}
      <Card>
        <CardContent className="p-0">
          <div className="p-4 border-b">
            <div className="relative w-full max-w-sm">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="pl-9"
                placeholder={t('Search channels...')}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
          </div>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Channel')}</TableHead>
                  <TableHead>{t('Type')}</TableHead>
                  <TableHead className="text-right">{t('Cost/Unit')}</TableHead>
                  <TableHead className="text-right">{t('Sell Rate')}</TableHead>
                  <TableHead className="text-right">{t('Revenue')}</TableHead>
                  <TableHead className="text-right">{t('Cost')}</TableHead>
                  <TableHead className="text-right">{t('Profit')}</TableHead>
                  <TableHead className="text-right">{t('Margin')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      {Array.from({ length: 8 }).map((_, j) => (
                        <TableCell key={j}><Skeleton className="h-4 w-16" /></TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : filtered.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                      {t('No data')}
                    </TableCell>
                  </TableRow>
                ) : (
                  filtered.map((item) => (
                    <TableRow key={item.id} className="hover:bg-muted/50">
                      <TableCell className="font-medium">{item.name}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className={item.type >= 100 ? 'border-purple-300 text-purple-700' : ''}>
                          {typeMap?.[String(item.type)] || `T${item.type}`}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">{item.cost_per_unit.toFixed(4)}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{item.sell_price_rate.toFixed(1)}x</TableCell>
                      <TableCell className="text-right font-mono text-sm text-green-600">{item.total_revenue.toFixed(2)}</TableCell>
                      <TableCell className="text-right font-mono text-sm text-red-600">{item.total_cost.toFixed(2)}</TableCell>
                      <TableCell className={cn(
                        'text-right font-mono text-sm font-semibold',
                        item.profit >= 0 ? 'text-emerald-600' : 'text-red-600'
                      )}>
                        {item.profit.toFixed(2)}
                      </TableCell>
                      <TableCell className="text-right">
                        <span className={cn(
                          'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium',
                          item.margin >= 30 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' :
                          item.margin >= 0 ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' :
                          'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                        )}>
                          {item.margin.toFixed(1)}%
                        </span>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
