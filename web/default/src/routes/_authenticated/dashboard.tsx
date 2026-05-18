import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { useQuery } from '@tanstack/react-query'
import {
  Users,
  Network,
  Key,
  Activity,
  Zap,
  Server,
  TrendingUp,
  TrendingDown,
  Clock,
  DollarSign,
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
} from 'recharts'

export const Route = createFileRoute('/_authenticated/dashboard')({
  component: DashboardPage,
})

interface StatCardProps {
  title: string
  value: string | number
  description?: string
  icon: React.ElementType
  trend?: { value: number; label: string }
  loading?: boolean
  color?: string
}

const colorMap: Record<string, string> = {
  blue: 'from-blue-500 to-blue-600',
  purple: 'from-purple-500 to-purple-600',
  green: 'from-green-500 to-green-600',
  orange: 'from-orange-500 to-orange-600',
  red: 'from-red-500 to-red-600',
  cyan: 'from-cyan-500 to-cyan-600',
  pink: 'from-pink-500 to-pink-600',
}

function StatCard({ title, value, description, icon: Icon, trend, loading, color = 'blue' }: StatCardProps) {
  const gradientClass = colorMap[color] || colorMap.blue
  
  return (
    <Card className="relative overflow-hidden group hover:shadow-lg transition-all duration-300">
      {/* Gradient Background */}
      <div className={cn(
        'absolute inset-0 bg-gradient-to-br opacity-5 group-hover:opacity-10 transition-opacity',
        gradientClass
      )} />
      
      {/* Icon */}
      <div className={cn(
        'absolute top-4 right-4 w-12 h-12 rounded-xl bg-gradient-to-br flex items-center justify-center shadow-lg',
        gradientClass
      )}>
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
            <div className="text-3xl font-bold tracking-tight mb-2">{value}</div>
            {(description || trend) && (
              <div className="flex items-center gap-2">
                {trend && (
                  <span
                    className={cn(
                      'inline-flex items-center gap-1 text-sm font-medium px-2 py-1 rounded-full',
                      trend.value >= 0 
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' 
                        : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                    )}
                  >
                    {trend.value >= 0 ? (
                      <TrendingUp className="h-3 w-3" />
                    ) : (
                      <TrendingDown className="h-3 w-3" />
                    )}
                    {Math.abs(trend.value)}%
                  </span>
                )}
                {trend?.label || description && (
                  <span className="text-xs text-muted-foreground">
                    {trend?.label || description}
                  </span>
                )}
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}

// 妯℃嫙鍥捐〃鏁版嵁
const chartData = [
  { name: 'Mon', requests: 2400, cost: 12.5 },
  { name: 'Tue', requests: 1398, cost: 8.2 },
  { name: 'Wed', requests: 3800, cost: 18.9 },
  { name: 'Thu', requests: 3908, cost: 19.5 },
  { name: 'Fri', requests: 4800, cost: 24.0 },
  { name: 'Sat', requests: 3800, cost: 19.0 },
  { name: 'Sun', requests: 4300, cost: 21.5 },
]

const modelChartData = [
  { model: 'GPT-4', requests: 1234, percentage: 35 },
  { model: 'Claude-3', requests: 987, percentage: 28 },
  { model: 'Gemini', requests: 654, percentage: 19 },
  { model: 'Others', requests: 543, percentage: 18 },
]

function DashboardPage() {
  const { t } = useTranslation()
  const { auth } = useAuthStore()

  const { data: statsData, isLoading: statsLoading } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      const res = await fetch('/api/user/self/dashboard')
      if (!res.ok) throw new Error('Failed to fetch')
      return res.json()
    },
    retry: false,
    staleTime: 60 * 1000,
  })

  const stats = statsData?.data

  const userName = auth.user?.display_name || auth.user?.username || ''
  const isAdmin = auth.user?.role === 100

  const cards = [
    {
      title: t('Total Requests'),
      value: stats?.total_requests?.toLocaleString() ?? '0',
      description: t('All time'),
      icon: Activity,
      color: 'blue',
    },
    {
      title: t('Active Today'),
      value: stats?.today_requests?.toLocaleString() ?? '0',
      description: t('Today'),
      icon: Zap,
      color: 'orange',
    },
    {
      title: t('Total Cost'),
      value: stats?.total_cost ? `$${parseFloat(stats.total_cost).toFixed(2)}` : '$0',
      description: t('All time'),
      icon: DollarSign,
      color: 'green',
    },
    {
      title: t('Models Used'),
      value: stats?.model_count ?? '0',
      description: t('Unique models'),
      icon: Network,
      color: 'purple',
    },
  ]

  if (isAdmin) {
    cards.unshift(
      {
        title: t('Total Users'),
        value: stats?.user_count?.toLocaleString() ?? '0',
        description: t('Registered'),
        icon: Users,
        color: 'cyan',
      },
      {
        title: t('Total Tokens'),
        value: stats?.total_tokens?.toLocaleString() ?? '0',
        description: t('All time'),
        icon: Key,
        color: 'pink',
      }
    )
  }

  const modelUsage = stats?.model_usage ?? []

  return (
    <div className="p-4 sm:p-6 space-y-6 min-h-screen  w-full bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('Welcome back')}, {userName}
          </h1>
          <p className="text-muted-foreground mt-1 sm:mt-2 text-sm sm:text-base lg:text-lg">
            {t('Here is an overview of your AI usage')}
          </p>
        </div>
        <Badge variant="outline" className="text-sm px-4 py-2 bg-white/50 dark:bg-slate-800/50">
          <Clock className="w-4 h-4 mr-2" />
          {new Date().toLocaleDateString()}
        </Badge>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-3 sm:gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map((card, index) => (
          <StatCard key={index} {...card} loading={statsLoading} />
        ))}
      </div>

      {/* Charts Row */}
      <div className="grid gap-4 sm:gap-6 grid-cols-1 md:grid-cols-2">
        {/* Request Trend Chart */}
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5 text-blue-600" />
              {t('Request Trend')}
            </CardTitle>
            <CardDescription>{t('Daily requests over the past week')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={chartData}>
                  <defs>
                    <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis 
                    dataKey="name" 
                    className="text-xs"
                    tick={{ fill: 'currentColor' }}
                  />
                  <YAxis 
                    className="text-xs"
                    tick={{ fill: 'currentColor' }}
                  />
                  <Tooltip 
                    contentStyle={{ 
                      backgroundColor: 'var(--background)',
                      border: '1px solid var(--border)',
                      borderRadius: '8px',
                      boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                    }}
                  />
                  <Area 
                    type="monotone" 
                    dataKey="requests" 
                    stroke="#3b82f6" 
                    strokeWidth={2}
                    fillOpacity={1} 
                    fill="url(#colorRequests)" 
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Model Distribution Chart */}
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Network className="h-5 w-5 text-purple-600" />
              {t('Model Distribution')}
            </CardTitle>
            <CardDescription>{t('Usage by AI model')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={modelChartData} layout="vertical">
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis type="number" className="text-xs" tick={{ fill: 'currentColor' }} />
                  <YAxis 
                    type="category" 
                    dataKey="model" 
                    className="text-xs"
                    tick={{ fill: 'currentColor' }}
                    width={80}
                  />
                  <Tooltip 
                    contentStyle={{ 
                      backgroundColor: 'var(--background)',
                      border: '1px solid var(--border)',
                      borderRadius: '8px',
                      boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                    }}
                  />
                  <Bar dataKey="requests" fill="#8b5cf6" radius={[0, 4, 4, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Model Usage Table */}
      {modelUsage.length > 0 && (
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5 text-green-600" />
              {t('Model Usage Details')}
            </CardTitle>
            <CardDescription>{t('Top models by usage')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {modelUsage.slice(0, 5).map((model: { model_name: string; request_count: number; total_tokens?: number }, index: number) => (
                <div key={index} className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-3 sm:p-4 rounded-lg bg-muted/50 hover:bg-muted/70 transition-colors gap-2">
                  <div className="flex items-center gap-4">
                    <div className={cn(
                      'flex items-center justify-center w-10 h-10 rounded-xl font-bold text-white',
                      index === 0 && 'bg-gradient-to-br from-blue-500 to-blue-600',
                      index === 1 && 'bg-gradient-to-br from-purple-500 to-purple-600',
                      index === 2 && 'bg-gradient-to-br from-green-500 to-green-600',
                      index === 3 && 'bg-gradient-to-br from-orange-500 to-orange-600',
                      index === 4 && 'bg-gradient-to-br from-pink-500 to-pink-600',
                    )}>
                      {index + 1}
                    </div>
                    <div>
                      <div className="font-semibold text-lg">{model.model_name}</div>
                      {model.total_tokens && (
                        <div className="text-sm text-muted-foreground">
                          {model.total_tokens.toLocaleString()} tokens
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="text-right">
                    <Badge variant="secondary" className="text-sm px-3 py-1">
                      {model.request_count.toLocaleString()} {t('requests')}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
