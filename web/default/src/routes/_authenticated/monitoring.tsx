import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import {
  Activity, Clock, Server, Cpu, HardDrive, Zap,
  TrendingUp, TrendingDown, RefreshCw, AlertCircle, CheckCircle2,
  BarChart3, Gauge,
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from 'recharts'

export const Route = createFileRoute('/_authenticated/monitoring')({
  component: MonitoringPage,
})

interface StatCardProps {
  title: string
  value: string | number
  subtitle?: string
  icon: React.ElementType
  color?: string
  loading?: boolean
  status?: 'success' | 'warning' | 'error' | 'info'
}

const colorMap: Record<string, string> = {
  blue: 'from-blue-500 to-blue-600',
  purple: 'from-purple-500 to-purple-600',
  green: 'from-green-500 to-green-600',
  orange: 'from-orange-500 to-orange-600',
  red: 'from-red-500 to-red-600',
  cyan: 'from-cyan-500 to-cyan-600',
  pink: 'from-pink-500 to-pink-600',
  amber: 'from-amber-500 to-amber-600',
  teal: 'from-teal-500 to-teal-600',
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const parts: string[] = []
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  if (minutes > 0) parts.push(`${minutes}m`)
  const secs = seconds % 60
  parts.push(`${secs}s`)
  return parts.join(' ')
}

function StatCard({ title, value, subtitle, icon: Icon, color = 'blue', loading, status }: StatCardProps) {
  const gradientClass = colorMap[color] || colorMap.blue
  const statusIndicator = status === 'success'
    ? 'bg-green-500'
    : status === 'warning'
    ? 'bg-amber-500'
    : status === 'error'
    ? 'bg-red-500'
    : status === 'info'
    ? 'bg-blue-500'
    : undefined

  return (
    <Card className="relative overflow-hidden group hover:shadow-lg transition-all duration-300">
      <div className={cn('absolute inset-0 bg-gradient-to-br opacity-5 group-hover:opacity-10 transition-opacity', gradientClass)} />
      {statusIndicator && (
        <div className={cn('absolute top-3 left-3 w-2.5 h-2.5 rounded-full animate-pulse', statusIndicator)} />
      )}
      <div className={cn('absolute top-4 right-4 w-12 h-12 rounded-xl bg-gradient-to-br flex items-center justify-center shadow-lg', gradientClass)}>
        <Icon className="h-6 w-6 text-white" />
      </div>
      <CardHeader className="relative pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent className="relative">
        {loading ? (
          <Skeleton className="h-10 w-24" />
        ) : (
          <>
            <div className="text-3xl font-bold tracking-tight mb-1">{value}</div>
            {subtitle && <div className="text-xs text-muted-foreground mt-1">{subtitle}</div>}
          </>
        )}
      </CardContent>
    </Card>
  )
}

const chartPlaceholderData = [
  { time: '00:00', value: 1200 }, { time: '00:05', value: 1350 }, { time: '00:10', value: 1100 },
  { time: '00:15', value: 1480 }, { time: '00:20', value: 1220 }, { time: '00:25', value: 1380 },
  { time: '00:30', value: 1050 }, { time: '00:35', value: 1420 }, { time: '00:40', value: 1300 },
  { time: '00:45', value: 980 }, { time: '00:50', value: 1150 }, { time: '00:55', value: 1280 },
]

const latencyData = [
  { time: '00:00', value: 245 }, { time: '00:05', value: 312 }, { time: '00:10', value: 278 },
  { time: '00:15', value: 198 }, { time: '00:20', value: 356 }, { time: '00:25', value: 289 },
  { time: '00:30', value: 267 }, { time: '00:35', value: 301 }, { time: '00:40', value: 234 },
  { time: '00:45', value: 287 }, { time: '00:50', value: 254 }, { time: '00:55', value: 301 },
]

function MonitoringPage() {
  const { t } = useT()
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  const { data: perfData, isLoading: perfLoading, isError: perfError, isRefetching: perfRefetching, refetch: refetchPerf, error: perfQueryError } = useQuery({
    queryKey: ['monitor-performance'],
    queryFn: async () => {
      const res = await fetch('/api/admin/performance')
      if (!res.ok) throw new Error('Failed to fetch performance stats')
      const json = await res.json()
      if (!json.success) throw new Error(json.message || 'API error')
      return json.data
    },
    retry: 2,
    retryDelay: 1000,
    refetchInterval: 10_000,
  })

  const { data: dashData, isLoading: dashLoading } = useQuery({
    queryKey: ['monitor-dashboard'],
    queryFn: async () => {
      const res = await fetch('/api/user/self/dashboard')
      if (!res.ok) throw new Error('Failed to fetch dashboard stats')
      const json = await res.json()
      return json.data
    },
    retry: 2,
    retryDelay: 1000,
    refetchInterval: 10_000,
  })

  const { data: statusData } = useQuery({
    queryKey: ['monitor-status'],
    queryFn: async () => {
      const res = await fetch('/api/status')
      if (!res.ok) throw new Error('Failed to fetch system status')
      const json = await res.json()
      return json.data
    },
    retry: 2,
    retryDelay: 1000,
    staleTime: 60_000,
  })

  useEffect(() => {
    if (perfData) setLastUpdated(new Date())
  }, [perfData])

  const isError = perfError || (!perfLoading && !perfData)
  const isLoading = perfLoading && !perfData

  const memoryAlloc = perfData?.memory?.alloc ?? 0
  const memoryTotal = perfData?.memory?.sys ?? 0
  const memoryMB = formatBytes(memoryAlloc)
  const totalMemoryMB = formatBytes(memoryTotal)
  const goRoutines = perfData?.memory?.goroutines ?? 0
  const numGC = perfData?.memory?.num_gc ?? 0
  const uptimeSeconds = perfData?.uptime?.uptime_seconds ?? 0
  const numCPU = perfData?.runtime?.num_cpu ?? 0
  const totalRequests = dashData?.total_requests ?? 0
  const todayRequests = dashData?.today_requests ?? 0
  const totalCost = dashData?.total_cost ?? 0
  const modelCount = dashData?.model_count ?? 0
  const systemName = statusData?.system_name ?? 'QuantumClaw'
  const version = statusData?.version ?? '-'

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('System Monitoring')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('Real-time performance and health overview')}</p>
      </div>
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div />
        <div className="flex items-center gap-3">
          {isError && (
            <Badge variant="destructive" className="flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              {t('Error')}
            </Badge>
          )}
          {perfData && !error && (
            <Badge variant="outline" className="flex items-center gap-1 text-green-600 border-green-200 bg-green-50 dark:bg-green-900/20 dark:border-green-800">
              <CheckCircle2 className="h-3 w-3" />
              {t('Live')}
            </Badge>
          )}
          <Badge variant="outline" className="text-sm px-4 py-2 bg-white/50 dark:bg-slate-800/50">
            <RefreshCw className={cn('w-4 h-4 mr-2', perfRefetching && 'animate-spin')} />
            {t('Auto-refresh')} 10s
          </Badge>
        </div>
      </div>

      {isError && (
        <Card className="border-red-200 bg-red-50 dark:bg-red-900/10 dark:border-red-800">
          <CardContent className="flex items-center gap-3 py-4">
            <AlertCircle className="h-5 w-5 text-red-600" />
            <div>
              <p className="font-medium text-red-700 dark:text-red-400">{t('Unable to fetch monitoring data')}</p>
              <p className="text-sm text-red-600/70 dark:text-red-400/70">
                {t('The monitoring API may be unavailable. Showing placeholder data.')}
              </p>
            </div>
            <button
              onClick={() => refetchPerf()}
              className="ml-auto px-4 py-2 text-sm font-medium text-red-700 hover:text-red-800 bg-red-100 hover:bg-red-200 rounded-lg transition-colors dark:text-red-400 dark:hover:text-red-300 dark:bg-red-900/30"
            >
              {t('Retry')}
            </button>
          </CardContent>
        </Card>
      )}

      {lastUpdated && (
        <div className="text-xs text-muted-foreground text-right -mb-4">
          {t('Last updated')}: {lastUpdated.toLocaleTimeString()}
        </div>
      )}

      <div>
        <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <Server className="h-5 w-5 text-blue-600" />
          {t('System Status')}
        </h2>
        <div className="grid gap-3 sm:gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
          <StatCard title={t('Uptime')} value={formatUptime(uptimeSeconds)} subtitle={t('Since last restart')} icon={Clock} color="blue" status={uptimeSeconds > 86400 ? 'success' : 'info'} loading={isLoading} />
          <StatCard title={t('Memory (Heap)')} value={memoryMB} subtitle={`${t('System')}: ${totalMemoryMB}`} icon={HardDrive} color="purple"
            status={memoryAlloc > 0 && memoryTotal > 0 && (memoryAlloc / memoryTotal) > 0.8 ? 'error' : memoryAlloc > 0 && memoryTotal > 0 && (memoryAlloc / memoryTotal) > 0.6 ? 'warning' : 'success'} loading={isLoading} />
          <StatCard title={t('Go Routines')} value={goRoutines.toLocaleString()} subtitle={`${numCPU} ${t('CPU cores')}`} icon={Cpu} color="cyan" status={goRoutines > 1000 ? 'warning' : 'success'} loading={isLoading} />
          <StatCard title={t('GC Cycles')} value={numGC.toLocaleString()} subtitle={t('Total garbage collections')} icon={Activity} color="teal" loading={isLoading} />
        </div>
      </div>

      <div>
        <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <Activity className="h-5 w-5 text-green-600" />
          {t('Request Metrics')}
        </h2>
        <div className="grid gap-3 sm:gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
          <StatCard title={t('Total Requests')} value={totalRequests.toLocaleString()} subtitle={t('All time')} icon={BarChart3} color="green" loading={dashLoading} />
          <StatCard title={t('Active Today')} value={todayRequests.toLocaleString()} subtitle={t('Last 24 hours')} icon={Zap} color="orange" loading={dashLoading} />
          <StatCard title={t('Total Cost')} value={`$${parseFloat(totalCost || '0').toFixed(2)}`} subtitle={t('All time')} icon={Server} color="pink" loading={dashLoading} />
          <StatCard title={t('Models Used')} value={modelCount.toLocaleString()} subtitle={t('Unique AI models')} icon={Gauge} color="amber" loading={dashLoading} />
        </div>
      </div>

      <div>
        <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <BarChart3 className="h-5 w-5 text-purple-600" />
          {t('Performance')}
        </h2>
        <div className="grid gap-4 sm:gap-6 grid-cols-1 md:grid-cols-2">
          <Card className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5 text-blue-600" />
                {t('Request Rate (last 60 min)')}
              </CardTitle>
              <CardDescription>{t('requests_per_window')}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="min-h-[200px] h-[30vh] lg:min-h-[250px] lg:h-[35vh]">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartPlaceholderData}>
                    <defs>
                      <linearGradient id="requestRateGrad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                    <XAxis dataKey="time" className="text-xs" tick={{ fill: 'currentColor' }} />
                    <YAxis className="text-xs" tick={{ fill: 'currentColor' }} />
                    <Tooltip contentStyle={{ backgroundColor: 'var(--background)', border: '1px solid var(--border)', borderRadius: '8px' }} />
                    <Area type="monotone" dataKey="value" stroke="#3b82f6" strokeWidth={2} fillOpacity={1} fill="url(#requestRateGrad)" />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          <Card className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Clock className="h-5 w-5 text-orange-600" />
                {t('Average Latency')}
              </CardTitle>
              <CardDescription>{t('response_time_ms')}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="min-h-[200px] h-[30vh] lg:min-h-[250px] lg:h-[35vh]">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={latencyData}>
                    <defs>
                      <linearGradient id="latencyGrad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#f59e0b" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                    <XAxis dataKey="time" className="text-xs" tick={{ fill: 'currentColor' }} />
                    <YAxis className="text-xs" tick={{ fill: 'currentColor' }} />
                    <Tooltip contentStyle={{ backgroundColor: 'var(--background)', border: '1px solid var(--border)', borderRadius: '8px' }} />
                    <Area type="monotone" dataKey="value" stroke="#f59e0b" strokeWidth={2} fillOpacity={1} fill="url(#latencyGrad)" />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <Card className="border-dashed">
        <CardHeader>
          <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
            <Server className="h-4 w-4" />
            {t('System Information')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground block">{t('System')}</span>
              <span className="font-medium">{systemName}</span>
            </div>
            <div>
              <span className="text-muted-foreground block">{t('Version')}</span>
              <span className="font-medium">v{version}</span>
            </div>
            <div>
              <span className="text-muted-foreground block">{t('CPU Cores')}</span>
              <span className="font-medium">{numCPU}</span>
            </div>
            <div>
              <span className="text-muted-foreground block">{t('Memory Sys')}</span>
              <span className="font-medium">{totalMemoryMB}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
