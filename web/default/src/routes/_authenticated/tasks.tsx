import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { ClipboardList, RefreshCw, Clock, CheckCircle2, AlertCircle } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/empty-state'
import apiClient from '@/lib/api'
import { type ApiResponse } from '@/lib/api-extended'

interface TaskLog {
  id: number
  type: string
  status: string
  message: string
  created_at: number
  updated_at: number
}

interface TaskResponse extends ApiResponse<TaskLog[]> {}

async function getTaskLogs(): Promise<TaskResponse> {
  const res = await apiClient.get('/api/task/log')
  return res.data
}

export const Route = createFileRoute('/_authenticated/tasks')({
  component: TaskLogsPage,
})

function TaskLogsPage() {
  const { t } = useTranslation()

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['task-logs'],
    queryFn: getTaskLogs,
    staleTime: 30 * 1000,
  })

  const logs: TaskLog[] = data?.data || []

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'bg-green-500'
      case 'failed':
        return 'bg-red-500'
      case 'running':
        return 'bg-blue-500'
      default:
        return 'bg-gray-500'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'success':
        return t('Success')
      case 'failed':
        return t('Failed')
      case 'running':
        return t('Running')
      default:
        return status
    }
  }

  const getTypeText = (type: string) => {
    const typeMap: Record<string, string> = {
      'model_sync': t('Model Sync'),
      'channel_balance': t('Channel Balance Update'),
      'ratio_sync': t('Ratio Sync'),
      'system_backup': t('System Backup'),
    }
    return typeMap[type] || type
  }

  const formatTime = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString()
  }

  return (
    <div className=" w-full p-4 sm:p-6 space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('Task Logs')}
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            {t('System task execution logs and status')}
          </p>
        </div>
        <Button variant="outline" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`mr-2 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('Refresh')}
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Card key={i}>
              <CardContent className="p-4">
                <div className="flex items-center gap-4">
                  <div className="h-10 w-10 rounded-full bg-muted animate-pulse" />
                  <div className="flex-1 space-y-2">
                    <div className="h-4 w-32 animate-pulse rounded bg-muted" />
                    <div className="h-3 w-48 animate-pulse rounded bg-muted" />
                  </div>
                  <div className="h-8 w-20 animate-pulse rounded bg-muted" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : logs.length === 0 ? (
        <EmptyState
          icon={ClipboardList}
          title={t('No task logs found')}
          description={t('System tasks will appear here when executed')}
        />
      ) : (
        <div className="space-y-3">
          {logs.map((log) => (
            <Card key={log.id} className="hover:shadow-md transition-shadow">
              <CardContent className="p-4">
                <div className="flex items-center gap-4">
                  <div className={`w-10 h-10 rounded-full flex items-center justify-center ${
                    log.status === 'success' ? 'bg-green-50' : 
                    log.status === 'failed' ? 'bg-red-50' : 'bg-blue-50'
                  }`}>
                    {log.status === 'success' ? (
                      <CheckCircle2 className="h-5 w-5 text-green-600" />
                    ) : log.status === 'failed' ? (
                      <AlertCircle className="h-5 w-5 text-red-600" />
                    ) : (
                      <Clock className="h-5 w-5 text-blue-600 animate-spin" />
                    )}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{getTypeText(log.type)}</span>
                      <Badge
                        variant="secondary"
                        className={`${getStatusColor(log.status)} text-white text-xs`}
                      >
                        {getStatusText(log.status)}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground mt-1">{log.message}</p>
                    <div className="flex items-center gap-4 mt-2 text-xs text-muted-foreground">
                      <span>{t('Created')}: {formatTime(log.created_at)}</span>
                      {log.updated_at !== log.created_at && (
                        <span>{t('Updated')}: {formatTime(log.updated_at)}</span>
                      )}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
