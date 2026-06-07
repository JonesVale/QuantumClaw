import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Scale,
  FileBarChart,
  CheckCircle,
  XCircle,
  Clock,
  Search,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  Eye,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import apiClient from '@/lib/api'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/my-reconciliation')({
  component: MyReconciliationPage,
})

// ── Types ──────────────────────────────────────────────────────────────────────
interface ReconciliationLog {
  id: number
  created_at: number
  user_id: number
  username?: string
  channel_id: number
  channel_name?: string
  consume_log_id: number
  user_deduct_cents: number
  channel_cost_cents: number
  platform_income_cents: number
  discrepancy_cents: number
  status: string
  resolved_by: string
  resolved_at: number
  remark: string
}

interface PaginatedResponse {
  items: ReconciliationLog[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// ── Helpers ─────────────────────────────────────────────────────────────────────
function formatCents(cents: number): string {
  const abs = Math.abs(cents)
  const sign = cents < 0 ? '-' : ''
  return `${sign}¥${(abs / 100).toFixed(2)}`
}

function formatTime(ts: number): string {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleString('zh-CN')
}

function statusBadge(status: string) {
  const map: Record<string, { label: string; variant: 'default' | 'destructive' | 'outline' | 'secondary' }> = {
    open:     { label: '未处理', variant: 'destructive' },
    resolved: { label: '已处理', variant: 'default' },
    ignored:  { label: '已忽略', variant: 'secondary' },
  }
  return map[status] || { label: status, variant: 'outline' }
}

// ── Page Component ─────────────────────────────────────────────────────────────
function MyReconciliationPage() {
  const { t } = useT()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [selectedLog, setSelectedLog] = useState<ReconciliationLog | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  // ── Queries ─────────────────────────────────────────────────────
  const { data, isLoading, refetch } = useQuery<PaginatedResponse>({
    queryKey: ['myReconciliations', page, pageSize, statusFilter],
    queryFn: async () => {
      const params = new URLSearchParams({
        page: String(page),
        page_size: String(pageSize),
      })
      if (statusFilter !== 'all') params.set('status', statusFilter)
      const r = await apiClient.get(`/api/reconciliation/my?${params.toString()}`)
      return r.data?.data || { items: [], total: 0, page: 1, page_size: 20, total_pages: 1 }
    },
  })

  const items = data?.items || []
  const total = data?.total || 0
  const totalPages = data?.total_pages || 1

  // ── Handlers ────────────────────────────────────────────────────
  function openDetail(log: ReconciliationLog) {
    setSelectedLog(log)
    setDetailOpen(true)
  }

  // ── Render ──────────────────────────────────────────────────────
  return (
    <div className="qc-wrapper py-6 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold">{t('我的对账')}</h1>
        <p className="text-sm text-muted-foreground mt-1">
          {t('查看您的计费对账明细')}
        </p>
      </div>

      {/* Filters */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row gap-4">
            <Select value={statusFilter} onValueChange={v => { setStatusFilter(v); setPage(1) }}>
              <SelectTrigger className="w-[140px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('全部')}</SelectItem>
                <SelectItem value="open">{t('未处理')}</SelectItem>
                <SelectItem value="resolved">{t('已处理')}</SelectItem>
                <SelectItem value="ignored">{t('已忽略')}</SelectItem>
              </SelectContent>
            </Select>
            <Select value={String(pageSize)} onValueChange={v => { setPageSize(Number(v)); setPage(1) }}>
              <SelectTrigger className="w-[100px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="10">10 / {t('页')}</SelectItem>
                <SelectItem value="20">20 / {t('页')}</SelectItem>
                <SelectItem value="50">50 / {t('页')}</SelectItem>
              </SelectContent>
            </Select>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              <RefreshCw className="h-4 w-4 mr-1" />
              {t('刷新')}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-8 text-center text-muted-foreground">{t('加载中...')}</div>
          ) : items.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground">{t('暂无对账记录')}</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="text-left p-3 font-medium">{t('渠道')}</th>
                    <th className="text-right p-3 font-medium">{t('用户扣款')}</th>
                    <th className="text-right p-3 font-medium">{t('渠道成本')}</th>
                    <th className="text-right p-3 font-medium text-destructive">{t('差异')}</th>
                    <th className="text-center p-3 font-medium">{t('状态')}</th>
                    <th className="text-left p-3 font-medium">{t('时间')}</th>
                    <th className="text-center p-3 font-medium">{t('明细')}</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map(log => {
                    const sb = statusBadge(log.status)
                    return (
                      <tr key={log.id} className="border-b hover:bg-muted/30 transition-colors">
                        <td className="p-3">
                          {log.channel_name || `渠道#${log.channel_id}`}
                        </td>
                        <td className="p-3 text-right font-mono">{formatCents(log.user_deduct_cents)}</td>
                        <td className="p-3 text-right font-mono">{formatCents(log.channel_cost_cents)}</td>
                        <td className={cn('p-3 text-right font-mono font-bold', log.discrepancy_cents > 0 ? 'text-destructive' : 'text-green-600')}>
                          {formatCents(log.discrepancy_cents)}
                        </td>
                        <td className="p-3 text-center">
                          <Badge variant={sb.variant}>{sb.label}</Badge>
                        </td>
                        <td className="p-3 text-xs text-muted-foreground">{formatTime(log.created_at)}</td>
                        <td className="p-3 text-center">
                          <Button variant="ghost" size="sm" onClick={() => openDetail(log)}>
                            <Eye className="h-4 w-4" />
                          </Button>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between p-4 border-t">
              <span className="text-sm text-muted-foreground">
                {t('共')} {total} {t('条')}
              </span>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page <= 1}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <span className="text-sm">
                  {page} / {totalPages}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Detail Dialog */}
      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t('对账详情')} #{selectedLog?.id}</DialogTitle>
          </DialogHeader>
          {selectedLog && (
            <div className="space-y-4 py-2">
              {/* Info Grid */}
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-muted-foreground">{t('渠道')}：</span>
                  {selectedLog.channel_name || `渠道#${selectedLog.channel_id}`}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('消费记录ID')}：</span>
                  {selectedLog.consume_log_id}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('状态')}：</span>
                  <Badge variant={statusBadge(selectedLog.status).variant}>
                    {statusBadge(selectedLog.status).label}
                  </Badge>
                </div>
                <div className="col-span-2">
                  <span className="text-muted-foreground">{t('时间')}：</span>
                  {formatTime(selectedLog.created_at)}
                </div>
              </div>

              {/* Amounts */}
              <div className="grid grid-cols-2 gap-3 p-3 bg-muted/30 rounded-lg text-sm font-mono">
                <div>{t('用户扣款')}：{formatCents(selectedLog.user_deduct_cents)}</div>
                <div>{t('渠道成本')}：{formatCents(selectedLog.channel_cost_cents)}</div>
                <div>{t('平台收入')}：{formatCents(selectedLog.platform_income_cents)}</div>
                <div className="font-bold text-destructive">{t('差异金额')}：{formatCents(selectedLog.discrepancy_cents)}</div>
              </div>

              {/* Resolved info */}
              {selectedLog.status !== 'open' && (
                <div className="p-3 bg-muted/30 rounded-lg text-sm space-y-1">
                  <div><span className="text-muted-foreground">{t('处理人')}：</span>{selectedLog.resolved_by}</div>
                  <div><span className="text-muted-foreground">{t('处理时间')}：</span>{formatTime(selectedLog.resolved_at)}</div>
                  {selectedLog.remark && <div><span className="text-muted-foreground">{t('备注')}：</span>{selectedLog.remark}</div>}
                </div>
              )}
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDetailOpen(false)}>
              {t('关闭')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
