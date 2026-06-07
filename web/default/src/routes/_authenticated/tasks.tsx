import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useAuthStore } from '@/stores/auth-store'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ClipboardList, RefreshCw, Clock, CheckCircle2, AlertCircle, Shield, Eye } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/empty-state'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
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
  const res = await apiClient.get('/api/task/')
  return res.data
}

export const Route = createFileRoute('/_authenticated/tasks')({
  component: TaskLogsPage,
})

function TaskLogsPage() {
  const { t } = useT()
  const { auth } = useAuthStore();
  const isAdmin = auth.user?.role >= 10;
  const [selectedTask, setSelectedTask] = useState<TaskLog | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center min-h-[60vh] p-4">
        <div className="text-center max-w-md">
          <Shield className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
          <h2 className="text-2xl font-bold mb-2">{t('Access Denied')}</h2>
          <p className="text-muted-foreground">{t('You do not have permission to access this page.')}</p>
        </div>
      </div>
    );
  }

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['task-logs'],
    queryFn: getTaskLogs,
    staleTime: 30 * 1000,
  })

  const logs: TaskLog[] = data?.data || []

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success': return 'bg-green-500'
      case 'failed': return 'bg-red-500'
      case 'running': return 'bg-blue-500'
      default: return 'bg-gray-500'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'success': return t('Success')
      case 'failed': return t('Failed')
      case 'running': return t('Running')
      default: return status
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
    if (!timestamp) return '-'
    return new Date(timestamp * 1000).toLocaleString('zh-CN')
  }

  function openDetail(task: TaskLog) {
    setSelectedTask(task)
    setDetailOpen(true)
  }

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Task Logs')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('System task execution logs and status')}</p>
      </div>

      <div className="flex justify-end">
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw className="h-4 w-4 mr-1" />
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
            <Card
              key={log.id}
              className="hover:shadow-md transition-shadow cursor-pointer"
              onClick={() => openDetail(log)}
            >
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
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{getTypeText(log.type)}</span>
                      <Badge
                        variant="secondary"
                        className={`${getStatusColor(log.status)} text-white text-xs`}
                      >
                        {getStatusText(log.status)}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground mt-1 truncate">{log.message}</p>
                    <div className="flex items-center gap-4 mt-2 text-xs text-muted-foreground">
                      <span>{t('Created')}: {formatTime(log.created_at)}</span>
                      {log.updated_at !== log.created_at && (
                        <span>{t('Updated')}: {formatTime(log.updated_at)}</span>
                      )}
                    </div>
                  </div>
                  <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); openDetail(log) }}>
                    <Eye className="h-4 w-4" />
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Detail Dialog */}
      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t('任务明细')} #{selectedTask?.id}</DialogTitle>
          </DialogHeader>
          {selectedTask && (
            <div className="space-y-4 py-2 text-sm">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <span className="text-muted-foreground">{t('任务类型')}：</span>
                  {getTypeText(selectedTask.type)}
                </div>
                <div>
                  <span className="text-muted-foreground">{t('状态')}：</span>
                  <Badge
                    variant="secondary"
                    className={`${getStatusColor(selectedTask.status)} text-white text-xs ml-1`}
                  >
                    {getStatusText(selectedTask.status)}
                  </Badge>
                </div>
                <div className="col-span-2">
                  <span className="text-muted-foreground">{t('创建时间')}：</span>
                  {formatTime(selectedTask.created_at)}
                </div>
                {selectedTask.updated_at !== selectedTask.created_at && (
                  <div className="col-span-2">
                    <span className="text-muted-foreground">{t('更新时间')}：</span>
                    {formatTime(selectedTask.updated_at)}
                  </div>
                )}
              </div>

              <div className="border-t pt-3">
                <div className="font-medium text-muted-foreground text-xs uppercase tracking-wide mb-2">{t('任务消息')}</div>
                <div className="p-3 bg-muted/30 rounded-lg text-sm whitespace-pre-wrap font-mono max-h-60 overflow-y-auto">
                  {selectedTask.message || '-'}
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
