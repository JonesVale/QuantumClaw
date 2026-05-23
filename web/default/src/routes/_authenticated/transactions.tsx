import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { RefreshCw, DollarSign, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { getTransactions, type TransactionItem } from '@/lib/api-extended'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/_authenticated/transactions')({
  component: TransactionsPage,
})

function TransactionsPage() {
  const { t } = useT()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['transactions', page],
    queryFn: () => getTransactions({ page, page_size: 20, model: search || undefined }),
    staleTime: 10_000,
  })
  const result = data?.data
  const txns: TransactionItem[] = result?.transactions || []
  const total = result?.total || 0
  const totalPages = Math.ceil(total / 20)

  return (
    <div className="p-4 sm:p-6 space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">{t('transactions')}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t('transactions_desc')}</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input className="pl-9 w-48" placeholder={t('search_model')} value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1) }} />
          </div>
          <Button variant="outline" onClick={() => refetch()}><RefreshCw className="h-4 w-4" /></Button>
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('time')}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('model')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('amount')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('unified_cost')}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('commission')}</th>
                  <th className="text-center text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('type')}</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <tr key={i} className="border-b border-muted/50">
                      {Array.from({ length: 6 }).map((_, j) => (
                        <td key={j} className="px-4 py-3"><Skeleton className="h-5 w-full" /></td>
                      ))}
                    </tr>
                  ))
                ) : txns.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                      <DollarSign className="h-12 w-12 mx-auto mb-2 opacity-30" />
                      {t('no_transactions')}
                    </td>
                  </tr>
                ) : (
                  txns.map((tx) => (
                    <tr key={tx.id} className="border-b border-muted/50 hover:bg-muted/30">
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {dayjs(tx.created_time * 1000).format('MM-DD HH:mm')}
                      </td>
                      <td className="px-4 py-3 font-medium text-sm">{tx.model_name}</td>
                      <td className="px-4 py-3 text-right font-mono text-sm">
                        ${tx.total_amount.toFixed(4)}
                      </td>
                      <td className="px-4 py-3 text-right font-mono text-xs text-emerald-600">
                        ${tx.unified_cost.toFixed(4)}
                      </td>
                      <td className="px-4 py-3 text-right font-mono text-xs text-blue-600">
                        ${tx.commission_amount.toFixed(4)}
                      </td>
                      <td className="px-4 py-3 text-center">
                        {tx.is_fallback ? (
                          <Badge variant="secondary" className="text-[10px]">{t('fallback')}</Badge>
                        ) : (
                          <Badge variant="default" className="text-[10px]">{t('direct')}</Badge>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
            {t('prev')}
          </Button>
          <span className="text-sm text-muted-foreground">
            {page} / {totalPages}
          </span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
            {t('next')}
          </Button>
        </div>
      )}
    </div>
  )
}
