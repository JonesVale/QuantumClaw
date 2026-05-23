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
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { EmptyState } from '@/components/empty-state'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
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

  return (
    <div className=" w-full p-4 sm:p-6 space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('Usage Logs')}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm sm:text-base lg:text-lg">
            {t('API request history and audit logs')}
          </p>
        </div>
        <Button variant="outline" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`mr-2 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('Refresh')}
        </Button>
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
      </div>

      {!isLoading && logs.length === 0 ? (
        <EmptyState
          icon={ScrollText}
          title={t('No logs found')}
          description={t('API usage logs will appear here once requests are made')}
        />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">#</TableHead>
                  <TableHead>{t('User')}</TableHead>
                  <TableHead>{t('Model')}</TableHead>
                  <TableHead>{t('Type')}</TableHead>
                  <TableHead>{t('Tokens')}</TableHead>
                  <TableHead>{t('Quota')}</TableHead>
                  <TableHead>{t('Time')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  Array.from({ length: 10 }).map((_, i) => (
                    <TableRow key={i}>
                      {Array.from({ length: 7 }).map((_, j) => (
                        <TableCell key={j}><Skeleton className="h-4 w-20" /></TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  logs.map((log, idx) => (
                    <TableRow key={log.id}>
                      <TableCell className="text-muted-foreground">{idx + 1}</TableCell>
                      <TableCell className="font-medium">{log.username}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className="font-mono text-xs">
                          {log.model_name}
                        </Badge>
                      </TableCell>
                      <TableCell>{log.type}</TableCell>
                      <TableCell className="text-sm">
                        {(log.prompt_tokens || 0) + (log.completion_tokens || 0)}
                      </TableCell>
                      <TableCell className="text-sm">{log.quota?.toLocaleString() || 0}</TableCell>
                      <TableCell className="text-muted-foreground text-xs">
                        {log.created_at
                          ? dayjs(log.created_at * 1000).format('MM-DD HH:mm:ss')
                          : '—'}
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
    </div>
  )
}
