import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter,
  DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Search, RotateCcw, AlertTriangle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { getAuditLogs, rollbackAuditLog, type BalanceLogItem } from '@/lib/api-extended'
import dayjs from '@/lib/dayjs'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/audit-logs')({
  component: AuditLogsPage,
})

const typeColors: Record<string, string> = {
  recharge: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  consume: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  refund: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  admin: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  topup: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400',
  debt_recovery: 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400',
  debt_deduct: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
}

const typeLabels: Record<string, string> = {
  recharge: 'Recharge',
  consume: 'Consume',
  refund: 'Refund',
  admin: 'Admin Adjust',
  topup: 'TopUp',
  debt_recovery: 'Debt Recovery',
  debt_deduct: 'Debt Deduct',
}

function canRollback(log: BalanceLogItem): boolean {
  // Can only rollback if it's NOT a refund and NOT already a rollback record
  if (log.type === 'refund') return false
  if (log.type === 'debt_recovery') return false
  if (log.type === 'debt_deduct') return false
  if (log.related_log_id > 0) return false
  return true
}

function AuditLogsPage() {
  const { t } = useT()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [userIdFilter, setUserIdFilter] = useState('')
  const [appliedUserId, setAppliedUserId] = useState<number>(0)
  const [rollbackTarget, setRollbackTarget] = useState<BalanceLogItem | null>(null)
  const [rollbackRemark, setRollbackRemark] = useState('')

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['auditLogs', page, appliedUserId],
    queryFn: () => getAuditLogs({ page, page_size: 30, user_id: appliedUserId || undefined }),
    staleTime: 10_000,
  })

  const rollbackMut = useMutation({
    mutationFn: async (logId: number) => {
      return rollbackAuditLog(logId, rollbackRemark || undefined)
    },
    onSuccess: (res) => {
      if (res.success) {
        toast.success(t('Rollback successful'))
      } else {
        toast.error(res.message || t('Rollback failed'))
      }
      setRollbackTarget(null)
      setRollbackRemark('')
      qc.invalidateQueries({ queryKey: ['auditLogs'] })
    },
    onError: (e: any) => {
      toast.error(e?.message || t('Rollback failed'))
    },
  })

  const result = data?.data
  const logs: BalanceLogItem[] = result?.logs || []
  const total = result?.total || 0
  const totalPages = Math.ceil(total / 30)

  const handleSearch = () => {
    const uid = parseInt(userIdFilter, 10)
    setAppliedUserId(isNaN(uid) ? 0 : uid)
    setPage(1)
  }

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Audit Logs') || 'Audit Logs'}</h1>
        <p className="text-muted-foreground mb-8" style={{ maxWidth: 'min(65ch, 100%)' }}>
          {t('Full balance change audit trail with rollback support') || 'Full balance change audit trail with rollback support'}
        </p>
      </div>

      {/* Search bar */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            className="pl-10"
            placeholder={t('Filter by user ID') || 'Filter by user ID'}
            value={userIdFilter}
            onChange={(e) => setUserIdFilter(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
        <Button variant="outline" size="sm" onClick={handleSearch}>
          {t('Search') || 'Search'}
        </Button>
        <Button variant="ghost" size="sm" onClick={() => refetch()}>
          <RotateCcw className="h-4 w-4 mr-1" />
          {t('Refresh') || 'Refresh'}
        </Button>
      </div>

      {/* Table */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/30">
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">#</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('User ID') || 'User ID'}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('Type') || 'Type'}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('Amount') || 'Amount'}</th>
                  <th className="text-right text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('Balance') || 'Balance'}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('Channel') || 'Channel'}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('Remark') || 'Remark'}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('Time') || 'Time'}</th>
                  <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 uppercase">{t('Action') || 'Action'}</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  Array.from({ length: 8 }).map((_, i) => (
                    <tr key={i} className="border-b">
                      {Array.from({ length: 9 }).map((_, j) => (
                        <td key={j} className="px-4 py-3"><Skeleton className="h-4 w-16" /></td>
                      ))}
                    </tr>
                  ))
                ) : logs.length === 0 ? (
                  <tr>
                    <td colSpan={9} className="px-4 py-12 text-center text-muted-foreground">
                      {t('No audit logs found') || 'No audit logs found'}
                    </td>
                  </tr>
                ) : (
                  logs.map((log) => (
                    <tr key={log.id} className={cn('border-b hover:bg-muted/20 transition-colors', log.related_log_id > 0 && 'bg-blue-50/30 dark:bg-blue-900/10')}>
                      <td className="px-4 py-3 text-sm font-mono tabular-nums">{log.id}</td>
                      <td className="px-4 py-3 text-sm font-mono tabular-nums">{log.user_id}</td>
                      <td className="px-4 py-3">
                        <Badge className={cn('text-xs font-normal', typeColors[log.type] || 'bg-gray-100')}>
                          {typeLabels[log.type] || log.type}
                        </Badge>
                        {log.related_log_id > 0 && (
                          <span className="text-xs text-muted-foreground ml-2" title={t('Related to') + ` #${log.related_log_id}`}>
                            ← #{log.related_log_id}
                          </span>
                        )}
                      </td>
                      <td className={cn('px-4 py-3 text-sm text-right font-mono tabular-nums', log.amount > 0 ? 'text-green-600' : 'text-red-600')}>
                        {log.amount > 0 ? '+' : ''}{log.amount}
                      </td>
                      <td className="px-4 py-3 text-sm text-right font-mono tabular-nums">{log.balance}</td>
                      <td className="px-4 py-3 text-sm font-mono">{log.channel_id > 0 ? `#${log.channel_id}` : '-'}</td>
                      <td className="px-4 py-3 text-sm max-w-xs truncate" title={log.remark}>
                        {log.remark || '-'}
                      </td>
                      <td className="px-4 py-3 text-sm text-muted-foreground whitespace-nowrap">
                        {log.created_at ? dayjs.unix(log.created_at).format('YYYY-MM-DD HH:mm') : '-'}
                      </td>
                      <td className="px-4 py-3">
                        {canRollback(log) ? (
                          <Button
                            variant="outline"
                            size="sm"
                            className="text-xs h-7 text-amber-600 hover:text-amber-700 border-amber-200"
                            onClick={() => {
                              setRollbackTarget(log)
                              setRollbackRemark('')
                            }}
                          >
                            <AlertTriangle className="h-3 w-3 mr-1" />
                            {t('Rollback') || 'Rollback'}
                          </Button>
                        ) : null}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between px-4 py-3 border-t">
              <span className="text-sm text-muted-foreground">
                {t('Total') || 'Total'}: {total} | {t('Page') || 'Page'} {page}/{totalPages}
              </span>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
                  {t('Previous') || 'Previous'}
                </Button>
                <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
                  {t('Next') || 'Next'}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Rollback confirmation dialog */}
      <Dialog open={!!rollbackTarget} onOpenChange={(v) => { if (!v) setRollbackTarget(null) }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-amber-600">
              <AlertTriangle className="h-5 w-5" />
              {t('Confirm Rollback') || 'Confirm Rollback'}
            </DialogTitle>
            <DialogDescription>
              {rollbackTarget && (
                <>
                  <p className="mb-2">{t('Are you sure you want to rollback this transaction?') || 'Are you sure you want to rollback this transaction?'}</p>
                  <div className="space-y-1 text-sm bg-muted p-3 rounded-lg">
                    <p><strong>ID:</strong> #{rollbackTarget.id}</p>
                    <p><strong>{t('User ID') || 'User ID'}:</strong> {rollbackTarget.user_id}</p>
                    <p><strong>{t('Type') || 'Type'}:</strong> {rollbackTarget.type}</p>
                    <p><strong>{t('Amount') || 'Amount'}:</strong> <span className={rollbackTarget.amount > 0 ? 'text-green-600' : 'text-red-600'}>{rollbackTarget.amount > 0 ? '+' : ''}{rollbackTarget.amount} cents</span></p>
                    <p><strong>{t('Balance after') || 'Balance after'}:</strong> {rollbackTarget.balance}</p>
                  </div>
                  <p className="mt-2 text-xs text-muted-foreground">
                    {t('This will reverse the amount and create a refund record') || 'This will reverse the amount and create a refund record'}
                  </p>
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label className="text-sm">{t('Remark (optional)') || 'Remark (optional)'}</Label>
            <Input
              placeholder={t('Reason for rollback') || 'Reason for rollback'}
              value={rollbackRemark}
              onChange={(e) => setRollbackRemark(e.target.value)}
            />
          </div>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setRollbackTarget(null)} disabled={rollbackMut.isPending}>
              {t('Cancel') || 'Cancel'}
            </Button>
            <Button
              variant="default"
              className="bg-amber-600 hover:bg-amber-700 text-white"
              onClick={() => { if (rollbackTarget) rollbackMut.mutate(rollbackTarget.id) }}
              disabled={rollbackMut.isPending}
            >
              {rollbackMut.isPending ? (t('Rolling back...') || 'Rolling back...') : (t('Confirm Rollback') || 'Confirm Rollback')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
