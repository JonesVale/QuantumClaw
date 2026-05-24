import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Bell, BellOff, CheckCheck, ChevronLeft, ChevronRight,
  Loader2, Trash2, Clock,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import apiClient from '@/lib/api'
import { toast } from 'sonner'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface NotificationItem {
  id: number
  type: 'topup' | 'system' | 'alert'
  title: string
  content: string
  is_read: boolean
  created_at: number
  // Some backends use read_at instead; unify internally
  read_at?: number | null
}

interface NotificationListResponse {
  success: boolean
  data?: {
    items: NotificationItem[]
    page: number
    page_size: number
    total: number
  }
}

interface UnreadCountResponse {
  success: boolean
  data?: {
    unread_count: number
  }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const NOTIFICATION_ICONS: Record<string, React.ElementType> = {
  topup: () => <>💰</>,
  system: () => <>🔧</>,
  alert: () => <>⚠️</>,
}

function getNotifIcon(type: string): string {
  switch (type) {
    case 'topup':
      return '💰'
    case 'system':
      return '🔧'
    case 'alert':
      return '⚠️'
    default:
      return '🔔'
  }
}

function getNotifColor(type: string): string {
  switch (type) {
    case 'topup':
      return 'bg-green-100 text-green-700 dark:bg-green-950/30 dark:text-green-400'
    case 'system':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-950/30 dark:text-blue-400'
    case 'alert':
      return 'bg-red-100 text-red-700 dark:bg-red-950/30 dark:text-red-400'
    default:
      return 'bg-muted text-muted-foreground'
  }
}

function formatTime(ts: number): string {
  const d = new Date(ts * 1000)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  if (hours < 24) return `${hours}h ago`
  if (days < 7) return `${days}d ago`
  return d.toLocaleDateString()
}

// ---------------------------------------------------------------------------
// Route
// ---------------------------------------------------------------------------

export const Route = createFileRoute('/_authenticated/notifications')({
  component: NotificationsPage,
})

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

function NotificationsPage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const pageSize = 20

  // --- Fetch notifications ---

  const {
    data: notifData,
    isLoading,
    isError,
    refetch,
    isFetching,
  } = useQuery<NotificationListResponse>({
    queryKey: ['notifications', page],
    queryFn: async () => {
      const res = await apiClient.get('/api/user/self/notifications', {
        params: { page, page_size: pageSize },
      })
      return res.data
    },
    staleTime: 10 * 1000,
  })

  // --- Fetch unread count ---

  const { data: unreadData, refetch: refetchUnread } = useQuery<UnreadCountResponse>({
    queryKey: ['notifications-unread'],
    queryFn: async () => {
      const res = await apiClient.get('/api/user/self/notifications/unread_count')
      return res.data
    },
    staleTime: 5 * 1000,
  })

  const notifications: NotificationItem[] = notifData?.data?.items ?? []
  const total = notifData?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const unreadCount = unreadData?.data?.unread_count ?? 0

  // --- Mark single read ---

  const markReadMutation = useMutation({
    mutationFn: async (id: number) => {
      await apiClient.put(`/api/user/self/notifications/${id}/read`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications', page] })
      refetchUnread()
    },
    onError: () => {
      toast.error(t('Failed to mark notification as read'))
    },
  })

  // --- Mark all read ---

  const markAllReadMutation = useMutation({
    mutationFn: async () => {
      await apiClient.put('/api/user/self/notifications/read_all')
    },
    onSuccess: () => {
      toast.success(t('All notifications marked as read'))
      queryClient.invalidateQueries({ queryKey: ['notifications', page] })
      refetchUnread()
    },
    onError: () => {
      toast.error(t('Failed to mark all as read'))
    },
  })

  // --- Delete (fallback to mark read) ---

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      try {
        await apiClient.delete(`/api/user/self/notifications/${id}`)
      } catch {
        // DELETE may not be available; fallback to mark read
        await apiClient.put(`/api/user/self/notifications/${id}/read`)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications', page] })
      refetchUnread()
      toast.success(t('Notification removed'))
    },
    onError: () => {
      toast.error(t('Failed to remove notification'))
    },
  })

  // --- Handlers ---

  const handleClick = (notif: NotificationItem) => {
    if (!notif.is_read) {
      markReadMutation.mutate(notif.id)
    }
  }

  const handleDelete = (e: React.MouseEvent, id: number) => {
    e.stopPropagation()
    deleteMutation.mutate(id)
  }

  // --- Render ---

  return (
    <div className="qc-wrapper py-8 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">{t('Notifications')}</h1>
          <p className="text-muted-foreground mt-2">
            {t('Stay updated with account activity and system alerts')}
          </p>
        </div>
        <div className="flex items-center gap-2">
          {unreadCount > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => markAllReadMutation.mutate()}
              disabled={markAllReadMutation.isPending}
            >
              {markAllReadMutation.isPending ? (
                <Loader2 className="mr-1 h-3 w-3 animate-spin" />
              ) : (
                <CheckCheck className="mr-1 h-3 w-3" />
              )}
              {t('Mark All Read')}
              <Badge variant="secondary" className="ml-1.5 text-xs">
                {unreadCount}
              </Badge>
            </Button>
          )}
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              refetch()
              refetchUnread()
            }}
            disabled={isFetching}
          >
            {isFetching ? (
              <Loader2 className="h-3 w-3 animate-spin" />
            ) : (
              <Bell className="h-3 w-3" />
            )}
          </Button>
        </div>
      </div>

      {/* Notifications List */}
      <Card className="w-full">
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-base">
            <Bell className="h-4 w-4" />
            {t('Notification History')}
            {unreadCount > 0 && (
              <Badge
                variant="secondary"
                className="bg-amber-100 text-amber-700 dark:bg-amber-950/30 dark:text-amber-400 ml-1"
              >
                {unreadCount} {t('unread')}
              </Badge>
            )}
          </CardTitle>
          <CardDescription>
            {t('Total')}: {total} {t('notifications')}
          </CardDescription>
        </CardHeader>
        <Separator />
        <CardContent className="p-0">
          {isLoading ? (
            /* Loading skeleton */
            <div className="divide-y divide-border/40">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex items-start gap-4 p-4 animate-pulse">
                  <div className="h-9 w-9 rounded-full bg-muted shrink-0" />
                  <div className="flex-1 space-y-2 min-w-0">
                    <div className="h-4 w-3/5 rounded bg-muted" />
                    <div className="h-3 w-full rounded bg-muted" />
                    <div className="h-3 w-1/4 rounded bg-muted" />
                  </div>
                </div>
              ))}
            </div>
          ) : isError ? (
            /* Error state */
            <div className="flex flex-col items-center justify-center py-16 px-4">
              <BellOff className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground mb-2">{t('Failed to load notifications')}</p>
              <Button variant="outline" size="sm" onClick={() => refetch()}>
                {t('Retry')}
              </Button>
            </div>
          ) : notifications.length === 0 ? (
            /* Empty state */
            <div className="flex flex-col items-center justify-center py-16 px-4">
              <BellOff className="h-14 w-14 text-muted-foreground/50 mb-4" />
              <p className="text-base font-medium text-foreground/70">{t('No notifications yet')}</p>
              <p className="text-sm text-muted-foreground mt-1">
                {t('You\'re all caught up!')}
              </p>
            </div>
          ) : (
            /* Notification list */
            <div className="divide-y divide-border/30">
              {notifications.map((notif) => {
                const isUnread = !notif.is_read
                return (
                  <div
                    key={notif.id}
                    onClick={() => handleClick(notif)}
                    className={cn(
                      'flex items-start gap-3 sm:gap-4 p-4 cursor-pointer transition-colors hover:bg-muted/30',
                      isUnread && 'border-l-4 border-l-amber-500 bg-amber-50/30 dark:bg-amber-950/10'
                    )}
                  >
                    {/* Icon */}
                    <div
                      className={cn(
                        'flex h-9 w-9 shrink-0 items-center justify-center rounded-xl text-base',
                        getNotifColor(notif.type)
                      )}
                    >
                      {getNotifIcon(notif.type)}
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <p
                          className={cn(
                            'text-sm font-medium truncate',
                            isUnread && 'font-semibold text-foreground'
                          )}
                        >
                          {notif.title}
                        </p>
                        <div className="flex items-center gap-1 shrink-0">
                          {isUnread && (
                            <span className="h-2 w-2 rounded-full bg-amber-500" />
                          )}
                        </div>
                      </div>
                      <p className="text-sm text-muted-foreground mt-0.5 line-clamp-2">
                        {notif.content}
                      </p>
                      <div className="flex items-center gap-2 mt-1.5">
                        <span className="text-xs text-muted-foreground/60 flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          {formatTime(notif.created_at)}
                        </span>
                        <Badge
                          variant="outline"
                          className="text-[10px] px-1.5 py-0 h-auto capitalize"
                        >
                          {notif.type}
                        </Badge>
                      </div>
                    </div>

                    {/* Actions */}
                    <button
                      onClick={(e) => handleDelete(e, notif.id)}
                      className="shrink-0 p-1.5 rounded-lg text-muted-foreground/40 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/20 transition-colors opacity-0 group-hover:opacity-100"
                      title={t('Remove')}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                )
              })}
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <>
              <Separator />
              <div className="flex items-center justify-between px-4 py-3">
                <p className="text-xs text-muted-foreground">
                  {t('Page')} {page} / {totalPages}
                </p>
                <div className="flex items-center gap-1">
                  <Button
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    disabled={page <= 1}
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  {Array.from({ length: Math.min(totalPages, 5) }).map((_, i) => {
                    // Show pages around current
                    const start = Math.max(1, Math.min(page - 2, totalPages - 4))
                    const p = start + i
                    if (p > totalPages) return null
                    return (
                      <Button
                        key={p}
                        variant={p === page ? 'default' : 'outline'}
                        size="icon"
                        className="h-8 w-8 text-xs"
                        onClick={() => setPage(p)}
                      >
                        {p}
                      </Button>
                    )
                  })}
                  <Button
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    disabled={page >= totalPages}
                    onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
