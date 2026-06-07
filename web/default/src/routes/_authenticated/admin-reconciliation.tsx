import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
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
  Check,
  X,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { toast } from 'sonner'
import apiClient from '@/lib/api'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/admin-reconciliation')({
  component: AdminReconciliationPage,
})

// ── Types ──────────────────────────────────────────────────────────────────────────

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
}

// ── Helpers ──────────────────────────────────────────────────────────────────────

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
    open: { label: '未处理', variant: 'destructive' },
    resolved: { label: '已处理', variant: 'default' },
    ignored: { label: '已忽略', variant: 'secondary' },
  }
  return map[status] || { label: status, variant: 'outline' }
}

// ── Page Component ───────────────────────────────────────────────────────────────

function AdminReconciliationPage() {
  const { t } = useT()
  const qc = useQueryClient()

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState<string>('open')
  const [search, setSearch] = useState('')
  const [selectedLog, setSelectedLog] = useState<ReconciliationLog | null>(null)
  const [resolveAction, setResolveAction] = useState<'approve_user' | 'approve_channel' | 'adjust_both' | 'ignore'>('approve_user')
  const [resolveRemark, setResolveRemark] = useState('')
  const [detailOpen, setDetailOpen] = useState(false)

  const pageSizeNum = Number(pageSize)

  // ── Queries ───────────────────────────────────────────────────────
  const { data, isLoading, refetch } = useQuery<PaginatedResponse>({
    queryKey: ['adminReconciliations', page, pageSize, statusFilter, search],
    queryFn: async () => {
      const params = new URLSearchParams({
        page: String(page),
        page_size: String(pageSizeNum),
      })
      if (statusFilter !== 'all') params.set('status', statusFilter)
      if (search) params.set('keyword', search)
      const r = await apiClient.get(`/api/admin/reconciliations?${params.toString()}`)
      return r.data?.data || { items: [], total: 0, page: 1, page_size: 20 }
    },
  })

  const { data: discrepancyData, isLoading: discLoading } = useQuery<PaginatedResponse>({
    queryKey: ['adminReconciliationDiscrepancies'],
    queryFn: async () => {
      const r = await apiClient.get('/api/admin/reconciliations/discrepancies?page=1&page_size=100')
      return r.data?.data || { items: [], total: 0 }
    },
  })

  // ── Mutations ─────────────────────────────────────────────────────
  const resolveMut = useMutation({
    mutationFn: async ({ id, action, remark }: { id: number; action: string; remark: string }) => {
      await apiClient.post(`/api/admin/reconciliations/${id}/resolve`, {
        action,
        remark,
        resolved_by: 'admin',
      })
    },
    onSuccess: () => {
      toast.success(t('处理成功'))
      setDetailOpen(false)
      setSelectedLog(null)
      setResolveRemark('')
      qc.invalidateQueries({ queryKey: ['adminReconciliations'] })
      qc.invalidateQueries({ queryKey: ['adminReconciliationDiscrepancies'] })
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || t('处理失败'))
    },
  })

  // ── Derived ───────────────────────────────────────────────────────
  const items = data?.items || []
  const total = data?.total || 0
  const totalPages = Math.max(1, Math.ceil(total / pageSizeNum))
  const openCount = discrepancyData?.total || 0
  const openItems = discrepancyData?.items || []
  const totalDiscrepancyCents = openItems.reduce((s, x) => s + Math.abs(x.discrepancy_cents), 0)

  // ── Handlers ──────────────────────────────────────────────────────
  function openDetail(log: ReconciliationLog) {
    setSelectedLog(log)
    setResolveAction('approve_user')
    setResolveRemark('')
    setDetailOpen(true)
  }

  function handleResolve() {
    if (!selectedLog) return
    resolveMut.mutate({
      id: selectedLog.id,
      action: resolveAction,
      remark: resolveRemark,
    })
  }

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    setPage(1)
    refetch()
  }

  // ── Render ─────────────────────────────────────────────────────────
  return (
    <div className="qc-wrapper py-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">{t('对账管理')}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t('查看和处理计费对账异常')}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={() => { refetch(); qc.invalidateQueries({ queryKey: ['adminReconciliationDiscrepancies'] }) }}>
          <RefreshCw className="h-4 w-4 mr-1" />
          {t('刷新')}
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">{t('未处理异常')}</CardTitle>
            <XCircle className="h-5 w-5 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-destructive">{openCount}</div>
          </CardContent>
        </Card>

        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">{t('异常总金额')}</CardTitle>
            <Scale className="h-5 w-5 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-amber-600">{formatCents(totalDiscrepancyCents)}</div>
          </CardContent>
        </Card>

        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">{t('总记录数')}</CardTitle>
            <FileBarChart className="h-5 w-5 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{total}</div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row gap-4">
            <form onSubmit={handleSearch} className="flex gap-2 flex-1">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder={t('搜索用户ID / 渠道ID')}
                  value={search}
                  onChange={e => setSearch(e.target.value)}
                  className="pl-9"
                />
              </div>
              <Button type="submit" variant="outline" size="sm">{t('搜索')}</Button>
            </form>
            <Select value={statusFilter} onValueChange={v => { setStatusFilter(v); setPage(1) }}>
              <SelectTrigger className="w-[140px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="open">{t('未处理')}</SelectItem>
                <SelectItem value="resolved">{t('已处理')}</SelectItem>
                <SelectItem value="ignored">{t('已忽略')}</SelectItem>
                <SelectItem value="all">{t('全部')}</SelectItem>
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
                    <th className="text-left p-3 font-medium">ID</th>
                    <th className="text-left p-3 font-medium">{t('用户ID')}</th>
                    <th className="text-left p-3 font-medium">{t('渠道ID')}</th>
                    <th className="text-right p-3 font-medium">{t('用户扣款')}</th>
                    <th className="text-right p-3 font-medium">{t('渠道成本')}</th>
                    <th className="text-right p-3 font-medium">{t('平台收入')}</th>
                    <th className="text-right p-3 font-medium text-destructive">{t('差异')}</th>
                    <th className="text-center p-3 font-medium">{t('状态')}</th>
                    <th className="text-left p-3 font-medium">{t('时间')}</th>
                    <th className="text-center p-3 font-medium">{t('操作')}</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map(log => {
                    const sb = statusBadge(log.status)
                    return (
                      <tr key={log.id} className="border-b hover:bg-muted/30 transition-colors">
                        <td className="p-3 font-mono text-xs">{log.id}</td>
                        <td className="p-3">{log.user_id}</td>
                        <td className="p-3">{log.channel_id}</td>
                        <td className="p-3 text-right font-mono">{formatCents(log.user_deduct_cents)}</td>
                        <td className="p-3 text-right font-mono">{formatCents(log.channel_cost_cents)}</td>
                        <td className="p-3 text-right font-mono">{formatCents(log.platform_income_cents)}</td>
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

      {/* Detail / Resolve Dialog */}
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
                  <span className="text-muted-foreground">{t('用户')}：</span>
                  {selectedLog.username || selectedLog.user_id}
                  {selectedLog.username && <span className="text-muted-foreground ml-1">(#{selectedLog.user_id})</span>}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('渠道')}：</span>
                  {selectedLog.channel_name || selectedLog.channel_id}
                  {selectedLog.channel_name && <span className="text-muted-foreground ml-1">(#{selectedLog.channel_id})</span>}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('消费记录ID')}：</span>
                  {selectedLog.consume_log_id}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('状态')}：</span>
                  <Badge variant={statusBadge(selectedLog.status).variant}>{statusBadge(selectedLog.status).label}</Badge>
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

              {/* Already resolved */}
              {selectedLog.status !== 'open' && (
                <div className="p-3 bg-muted/30 rounded-lg text-sm space-y-1">
                  <div><span className="text-muted-foreground">{t('处理人')}：</span>{selectedLog.resolved_by}</div>
                  <div><span className="text-muted-foreground">{t('处理时间')}：</span>{formatTime(selectedLog.resolved_at)}</div>
                  {selectedLog.remark && <div><span className="text-muted-foreground">{t('备注')}：</span>{selectedLog.remark}</div>}
                </div>
              )}

              {/* Resolve Form (only when open) */}
              {selectedLog.status === 'open' && (
                <div className="space-y-3 pt-2 border-t">
                  <div>
                    <Label>{t('处理方式')}</Label>
                    <Select value={resolveAction} onValueChange={v => setResolveAction(v as any)}>
                      <SelectTrigger className="mt-1">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="approve_user">{t('用户扣款正确（平台承担差异）')}</SelectItem>
                        <SelectItem value="approve_channel">{t('渠道成本正确（向用户补扣）')}</SelectItem>
                        <SelectItem value="adjust_both">{t('双方调整（手动修正）')}</SelectItem>
                        <SelectItem value="ignore">{t('忽略此差异')}</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label>{t('处理备注')}</Label>
                    <Textarea
                      value={resolveRemark}
                      onChange={e => setResolveRemark(e.target.value)}
                      placeholder={t('请输入备注...')}
                      className="mt-1"
                      rows={3}
                    />
                  </div>
                </div>
              )}
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDetailOpen(false)}>
              {t('关闭')}
            </Button>
            {selectedLog?.status === 'open' && (
              <Button
                onClick={handleResolve}
                disabled={resolveMut.isPending}
              >
                <Check className="h-4 w-4 mr-1" />
                {resolveMut.isPending ? t('处理中...') : t('确认处理')}
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
