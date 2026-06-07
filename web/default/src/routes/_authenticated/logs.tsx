import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Search,
  Download,
  RefreshCw,
  Filter,
  ScrollText,
  Eye,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { EmptyState } from '@/components/empty-state'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { getLogs, type LogEntry } from '@/lib/api-extended'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/_authenticated/logs')({
  component: LogsPage,
})

function LogsPage() {
  const { t } = useT()
  const [search, setSearch] = useState('')
  const [type, setType] = useState('0')
  const [page, setPage] = useState(1)
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['logs', search, type, page],
    queryFn: () =>
      getLogs({
        search: search || undefined,
        type: type !== '0' ? Number(type) : undefined,
        page: page,
      }),
    staleTime: 10 * 1000,
  })

  const logs: LogEntry[] = data?.data || []

  function openDetail(log: LogEntry) {
    setSelectedLog(log)
    setDetailOpen(true)
  }

  function getTypeLabel(typeValue: string | number | undefined): string {
    const map: Record<number, string> = {
      1: 'Chat Completion',
      2: 'Completion',
      3: 'Embedding',
      4: 'Image',
      5: 'Audio',
    }
    const num = typeof typeValue === 'string' ? Number(typeValue) : typeValue
    return num && map[num] ? map[num] : String(typeValue || '-')
  }

  function formatQuota(q: number | undefined): string {
    if (q === undefined || q === null) return '-'
    return (q / 1000000).toFixed(6)
  }

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Usage Logs')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('API request history and audit logs')}</p>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        <div className="relative w-full sm:max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="pl-9"
            placeholder={t('Search logs...')}
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          />
        </div>
        <Select value={type} onValueChange={(v) => { setType(v); setPage(1) }}>
          <SelectTrigger className="w-36">
            <SelectValue placeholder={t('All Types')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="0">{t('All Types')}</SelectItem>
            <SelectItem value="1">{t('Chat Completion')}</SelectItem>
            <SelectItem value="2">{t('Completion')}</SelectItem>
            <SelectItem value="3">{t('Embedding')}</SelectItem>
            <SelectItem value="4">{t('Image')}</SelectItem>
            <SelectItem value="5">{t('Audio')}</SelectItem>
          </SelectContent>
        </Select>
        <Button variant="outline" size="sm" onClick={() => { refetch() }}>
          <RefreshCw className="h-4 w-4 mr-1" />
          {t('Refresh')}
        </Button>
      </div>

      {!isLoading && logs.length === 0 ? (
        <EmptyState
          icon={ScrollText}
          title={t('No logs found')}
          description={t('API usage logs will appear here once requests are made')}
        />
      ) : (
        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">#</TableHead>
                  <TableHead>{t('User')}</TableHead>
                  <TableHead>{t('Model')}</TableHead>
                  <TableHead>{t('Type')}</TableHead>
                  <TableHead>{t('Prompt')}</TableHead>
                  <TableHead>{t('Completion')}</TableHead>
                  <TableHead>{t('Quota')}</TableHead>
                  <TableHead>{t('Time')}</TableHead>
                  <TableHead className="text-center">{t('明细')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  Array.from({ length: 10 }).map((_, i) => (
                    <TableRow key={i}>
                      {Array.from({ length: 9 }).map((_, j) => (
                        <TableCell key={j}><Skeleton className="h-4 w-20" /></TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  logs.map((log, idx) => (
                    <TableRow key={log.id} className="hover:bg-muted/30 transition-colors">
                      <TableCell className="text-muted-foreground">{idx + 1}</TableCell>
                      <TableCell className="font-medium">{log.username || log.user_id}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className="font-mono text-xs">
                          {log.model_name || log.model || '-'}
                        </Badge>
                      </TableCell>
                      <TableCell>{getTypeLabel(log.type_ || log.type)}</TableCell>
                      <TableCell className="text-sm text-right">
                        {(log.prompt_tokens || 0).toLocaleString()}
                      </TableCell>
                      <TableCell className="text-sm text-right">
                        {(log.completion_tokens || 0).toLocaleString()}
                      </TableCell>
                      <TableCell className="text-sm font-mono">
                        {formatQuota(log.quota)}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-xs">
                        {log.created_at
                          ? dayjs(log.created_at * 1000).format('MM-DD HH:mm:ss')
                          : '-'}
                      </TableCell>
                      <TableCell className="text-center">
                        <Button variant="ghost" size="sm" onClick={() => openDetail(log)}>
                          <Eye className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* Pagination */}
      {logs.length > 0 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {t('Page {{page}}', { page })}
          </p>
          <div className="flex flex-wrap gap-1">
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              {t('Previous')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => p + 1)}
            >
              {t('Next')}
            </Button>
          </div>
        </div>
      )}

      {/* Detail Dialog */}
      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t('日志明细')} #{selectedLog?.id}</DialogTitle>
          </DialogHeader>
          {selectedLog && (
            <div className="space-y-4 py-2 text-sm">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <span className="text-muted-foreground">{t('用户')}：</span>
                  {selectedLog.username || selectedLog.user_id}
                </div>
                <div>
                  <span className="text-muted-foreground">ID：</span>
                  {selectedLog.user_id}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('渠道')}：</span>
                  {selectedLog.channel_name || selectedLog.channel_id}
                  {selectedLog.channel_id ? ` (ID: ${selectedLog.channel_id})` : ''}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('模型')}：</span>
                  {selectedLog.model_name || selectedLog.model || '-'}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('类型')}：</span>
                  {getTypeLabel(selectedLog.type_ || selectedLog.type)}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('时间')}：</span>
                  {selectedLog.created_at
                    ? dayjs(selectedLog.created_at * 1000).format('YYYY-MM-DD HH:mm:ss')
                    : '-'}
                </div>
              </div>

              <div className="border-t pt-3 space-y-2">
                <div className="font-medium text-muted-foreground text-xs uppercase tracking-wide">{t('Token 明细')}</div>
                <div className="grid grid-cols-2 gap-3 font-mono text-sm">
                  <div>{t('Prompt Tokens')}：{(selectedLog.prompt_tokens || 0).toLocaleString()}</div>
                  <div>{t('Completion Tokens')}：{(selectedLog.completion_tokens || 0).toLocaleString()}</div>
                  {(selectedLog.request_tokens || selectedLog.response_tokens) && (
                    <>
                      <div>{t('Request Tokens')}：{(selectedLog.request_tokens || 0).toLocaleString()}</div>
                      <div>{t('Response Tokens')}：{(selectedLog.response_tokens || 0).toLocaleString()}</div>
                    </>
                  )}
                  <div className="col-span-2 font-bold">
                    {t('总计')}：{((selectedLog.prompt_tokens || 0) + (selectedLog.completion_tokens || 0)).toLocaleString()}
                  </div>
                </div>
              </div>

              <div className="border-t pt-3 space-y-2">
                <div className="font-medium text-muted-foreground text-xs uppercase tracking-wide">{t('计费明细')}</div>
                <div className="font-mono text-sm">
                  {t('Quota')}：{formatQuota(selectedLog.quota)}
                </div>
              </div>
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
