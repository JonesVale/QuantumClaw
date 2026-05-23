import { createFileRoute, Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { useQuery } from '@tanstack/react-query'
import {
  Network,
  Key,
  Activity,
  Zap,
  Server,
  TrendingUp,
  TrendingDown,
  Clock,
  DollarSign,
  Wallet,
  Plus,
  List,
  BarChart4,
  Database,
  Inbox,
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/empty-state'
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
  PieChart,
  Pie,
  Cell,
  Legend,
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
      <div className={cn(
        'absolute inset-0 bg-gradient-to-br opacity-5 group-hover:opacity-10 transition-opacity',
        gradientClass
      )} />

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

// ---------------------------------------------------------------------------
// Types matching the Go API response (/api/user/self/dashboard)
// ---------------------------------------------------------------------------
interface DailyRequest {
  date: string
  request_count: number
  token_count: number
  quota_used: number
}

interface ModelBreakdown {
  model_name: string
  request_count: number
  token_count: number
  quota_used: number
}

interface ProviderBreakdown {
  provider: string
  request_count: number
  token_count: number
  quota_used: number
}

interface LogEntry {
  day: string
  model_name: string
  request_count: number
  quota: number
  prompt_tokens: number
  completion_tokens: number
}

interface DashboardData {
  daily_requests: DailyRequest[]
  model_breakdown: ModelBreakdown[]
  provider_breakdown: ProviderBreakdown[]
  logs: LogEntry[]
}

interface BalanceData {
  balance: number
  balance_yuan: number
  logs: unknown[]
}

// Pie chart colors for provider distribution
const PIE_COLORS = ['#3b82f6', '#8b5cf6', '#10b981', '#f59e0b', '#ef4444', '#ec4899', '#06b6d4', '#84cc16']

function DashboardPage() {
  const { t } = useTranslation()
  const { auth } = useAuthStore()

  // ── Fetch dashboard stats ──────────────────────────────────────────────
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

  // ── Fetch balance ──────────────────────────────────────────────────────
  const { data: balanceData, isLoading: balanceLoading } = useQuery({
    queryKey: ['dashboard-balance'],
    queryFn: async () => {
      const res = await fetch('/api/user/self/balance')
      if (!res.ok) throw new Error('Failed to fetch balance')
      return res.json() as Promise<{ success: boolean; data: BalanceData }>
    },
    retry: false,
    staleTime: 30 * 1000,
  })

  // Safe data access
  const rawData = statsData?.data as DashboardData | undefined
  const balData = balanceData?.data as BalanceData | undefined

  const logs: LogEntry[] = rawData?.logs ?? []
  const dailyRequests: DailyRequest[] = rawData?.daily_requests ?? []
  const modelBreakdown: ModelBreakdown[] = rawData?.model_breakdown ?? []
  const providerBreakdown: ProviderBreakdown[] = rawData?.provider_breakdown ?? []

  // ── Calculate summary stats from API data ──────────────────────────────
  const totalRequests = dailyRequests.reduce((sum, d) => sum + d.request_count, 0)
  const todayRequests = dailyRequests
    .filter(d => d.date === new Date().toISOString().slice(0, 10))
    .reduce((sum, d) => sum + d.request_count, 0)
  const totalQuota = dailyRequests.reduce((sum, d) => sum + d.quota_used, 0)
  const totalTokens = dailyRequests.reduce((sum, d) => sum + d.token_count, 0)
  const modelCount = modelBreakdown.length

  // ── Chart data transformations ─────────────────────────────────────────
  const chartData = dailyRequests.map((d) => ({
    name: d.date.slice(5), // "MM-DD"
    requests: d.request_count,
    tokens: d.token_count,
    cost: d.quota_used,
  }))

  const tokenChartData = modelBreakdown
    .slice(0, 10)
    .map((m) => ({
      name: m.model_name.length > 20 ? m.model_name.slice(0, 20) + '…' : m.model_name,
      tokens: m.token_count,
      requests: m.request_count,
    }))
    .sort((a, b) => b.tokens - a.tokens)

  const modelChartData = modelBreakdown
    .slice(0, 10)
    .map((m) => ({
      model: m.model_name,
      requests: m.request_count,
      percentage: totalRequests > 0
        ? Math.round((m.request_count / totalRequests) * 100)
        : 0,
    }))
    .sort((a, b) => b.requests - a.requests)

  // ── Budget forecast ────────────────────────────────────────────────────
  const totalCost7d = chartData.reduce((sum, d) => sum + (d.cost || 0), 0)
  const dailyAvgCost = chartData.length > 0 ? totalCost7d / chartData.length : 0
  const now = new Date()
  const daysInMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0).getDate()
  const daysElapsed = now.getDate()
  const daysRemaining = daysInMonth - daysElapsed
  const projectedMonthly = dailyAvgCost * daysInMonth
  const currentMonthCost = dailyAvgCost * daysElapsed
  const budgetPercent = projectedMonthly > 0 ? Math.min(100, Math.round((currentMonthCost / projectedMonthly) * 100)) : 0

  const hasChartData = chartData.length > 0
  const hasModelData = modelChartData.length > 0
  const hasProviderData = providerBreakdown.length > 0
  const hasLogs = logs.length > 0

  const userName = auth.user?.display_name || auth.user?.username || ''
  const isAdmin = auth.user?.role === 100

  // ── Stat cards ─────────────────────────────────────────────────────────
  const cards = [
    {
      title: t('Total Requests'),
      value: totalRequests.toLocaleString() || '0',
      description: t('All time'),
      icon: Activity,
      color: 'blue' as const,
    },
    {
      title: t('Active Today'),
      value: todayRequests.toLocaleString() || '0',
      description: t('Today'),
      icon: Zap,
      color: 'orange' as const,
    },
    {
      title: t('Total Cost'),
      value: new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 }).format(totalQuota),
      description: t('All time'),
      icon: DollarSign,
      color: 'green' as const,
    },
    {
      title: t('Models Used'),
      value: String(modelCount || '0'),
      description: t('Unique models'),
      icon: Network,
      color: 'purple' as const,
    },
  ]

  if (isAdmin) {
    cards.unshift(
      {
        title: t('Total Tokens'),
        value: totalTokens.toLocaleString() || '0',
        description: t('All time'),
        icon: Key,
        color: 'pink' as unknown as 'blue' | 'green' | 'orange' | 'purple',
      },
    )
  }

  // Most used models for detail list
  const topModels = modelBreakdown
    .slice(0, 5)
    .sort((a, b) => b.request_count - a.request_count)

  // Recent logs (last 10)
  const recentLogs = logs.slice(0, 10)

  return (
    <div className="p-4 sm:p-6 space-y-6 min-h-screen w-full bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* ── Header ─────────────────────────────────────────────────── */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('Welcome back')}, {userName}
          </h1>
          <p className="text-muted-foreground mt-1 sm:mt-2 text-sm sm:text-base lg:text-lg">
            {t('Overview of your AI model and quantum computing usage')}
          </p>
        </div>
        <Badge variant="outline" className="text-sm px-4 py-2 bg-white/50 dark:bg-slate-800/50">
          <Clock className="w-4 h-4 mr-2" />
          {new Date().toLocaleDateString()}
        </Badge>
      </div>

      {/* ── Balance Section ────────────────────────────────────────── */}
      <Card className="relative overflow-hidden hover:shadow-lg transition-shadow">
        <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-green-500/5 to-teal-500/5" />
        <CardContent className="relative p-4 sm:p-6">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="flex items-center gap-3 sm:gap-4">
              <div className="w-12 h-12 sm:w-14 sm:h-14 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-lg shrink-0">
                <Wallet className="h-6 w-6 sm:h-7 sm:w-7 text-white" />
              </div>
              <div>
                <p className="text-xs sm:text-sm text-muted-foreground font-medium">{t('Account Balance')}</p>
                {balanceLoading ? (
                  <Skeleton className="h-8 w-28 mt-1" />
                ) : (
                  <div className="flex items-baseline gap-2">
                    <span className="text-2xl sm:text-3xl font-bold tracking-tight">
                      {new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 }).format(balData?.balance_yuan ?? 0)}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      ≈ {(balData?.balance ?? 0).toLocaleString()} {t('credits')}
                    </span>
                  </div>
                )}
              </div>
            </div>
            <Link to="/wallet">
              <Button className="bg-gradient-to-r from-emerald-500 to-teal-600 hover:from-emerald-600 hover:to-teal-700 text-white shadow-md hover:shadow-lg transition-all">
                <Plus className="h-4 w-4 mr-1" />
                {t('Recharge')}
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>

      {/* ── Stats Grid ─────────────────────────────────────────────── */}
      <div className="grid gap-3 sm:gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map((card, index) => (
          <StatCard key={index} {...card} loading={statsLoading} />
        ))}
      </div>

      {/* ── Budget Forecast ────────────────────────────────────────── */}
      {hasChartData && (
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <DollarSign className="h-5 w-5 text-emerald-500" />
              {t('Budget Forecast')}
            </CardTitle>
            <CardDescription>{t('Estimated monthly cost based on 7-day average')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div>
                <p className="text-xs text-muted-foreground">{t('Daily Avg')}</p>
                <p className="text-xl font-bold">
                  {new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(dailyAvgCost)}
                </p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('This Month')}</p>
                <p className="text-xl font-bold">
                  {new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(currentMonthCost)}
                </p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('Projected')}</p>
                <p className="text-xl font-bold">
                  {new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(projectedMonthly)}
                </p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('Budget Used')}</p>
                <div className="flex items-center gap-2">
                  <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
                    <div
                      className={cn(
                        'h-full rounded-full transition-all',
                        budgetPercent > 80 ? 'bg-red-500' : budgetPercent > 50 ? 'bg-amber-500' : 'bg-emerald-500'
                      )}
                      style={{ width: budgetPercent + '%' }}
                    />
                  </div>
                  <span className="text-sm font-bold">{budgetPercent}%</span>
                </div>
              </div>
            </div>
            <p className="text-xs text-muted-foreground mt-3">
              {daysRemaining} {t('days remaining in this billing period')}
            </p>
          </CardContent>
        </Card>
      )}

      {/* ── Charts Row 1: Daily Requests + Token Consumption ──────── */}
      <div className="grid gap-4 sm:gap-6 grid-cols-1 lg:grid-cols-2">
        {/* Daily Requests Area Chart */}
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5 text-blue-600" />
              {t('Daily Requests')}
            </CardTitle>
            <CardDescription>{t('Daily requests over the past week')}</CardDescription>
          </CardHeader>
          <CardContent>
            {hasChartData ? (
              <div className="h-[300px]">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData}>
                    <defs>
                      <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                    <XAxis dataKey="name" className="text-xs" tick={{ fill: 'currentColor' }} />
                    <YAxis className="text-xs" tick={{ fill: 'currentColor' }} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'var(--background)',
                        border: '1px solid var(--border)',
                        borderRadius: '8px',
                        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                      }}
                      formatter={((value: number, name: string): any => {
                        if (name === 'requests') return [value.toLocaleString(), t('Requests')]
                        if (name === 'cost') return [new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value), t('Cost')]
                        return [value, name]
                      }) as any}
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
            ) : (
              <EmptyState
                icon={Inbox}
                title={t('No request data yet')}
                description={t('Your API requests will appear here once you start using the service')}
              />
            )}
          </CardContent>
        </Card>

        {/* Token Consumption Bar Chart */}
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart4 className="h-5 w-5 text-violet-600" />
              {t('Token Consumption')}
            </CardTitle>
            <CardDescription>{t('Token usage by model')}</CardDescription>
          </CardHeader>
          <CardContent>
            {tokenChartData.length > 0 ? (
              <div className="h-[300px]">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={tokenChartData} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                    <XAxis type="number" className="text-xs" tick={{ fill: 'currentColor' }} />
                    <YAxis
                      type="category"
                      dataKey="name"
                      className="text-xs"
                      tick={{ fill: 'currentColor' }}
                      width={90}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'var(--background)',
                        border: '1px solid var(--border)',
                        borderRadius: '8px',
                        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                      }}
                      formatter={((value: number): any => [value.toLocaleString(), t('Tokens')]) as any}
                    />
                    <Bar dataKey="tokens" fill="#8b5cf6" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <EmptyState
                icon={Database}
                title={t('No token data yet')}
                description={t('Token consumption statistics will appear here once available')}
              />
            )}
          </CardContent>
        </Card>
      </div>

      {/* ── Charts Row 2: Provider Distribution + Model Distribution ── */}
      <div className="grid gap-4 sm:gap-6 grid-cols-1 lg:grid-cols-2">
        {/* Provider Distribution Pie Chart */}
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5 text-cyan-600" />
              {t('Provider Distribution')}
            </CardTitle>
            <CardDescription>{t('Usage by provider')}</CardDescription>
          </CardHeader>
          <CardContent>
            {hasProviderData ? (
              <div className="h-[300px]">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={providerBreakdown}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={100}
                      paddingAngle={3}
                      dataKey="request_count"
                      nameKey="provider"
                      label={(({ provider, request_count }: any) => {
                        const total = providerBreakdown.reduce((s: number, p: ProviderBreakdown) => s + p.request_count, 0)
                        const pct = total > 0 ? ((request_count / total) * 100).toFixed(1) : '0'
                        return `${provider} ${pct}%`
                      }) as any}
                    >
                      {providerBreakdown.map((_: ProviderBreakdown, index: number) => (
                        <Cell key={`cell-${index}`} fill={PIE_COLORS[index % PIE_COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'var(--background)',
                        border: '1px solid var(--border)',
                        borderRadius: '8px',
                      }}
                      formatter={((value: number, name: string): any => [value.toLocaleString(), name]) as any}
                    />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <EmptyState
                icon={Server}
                title={t('No provider data')}
                description={t('Provider distribution will appear once you start making API calls')}
              />
            )}
          </CardContent>
        </Card>

        {/* Model Distribution Bar Chart */}
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Network className="h-5 w-5 text-purple-600" />
              {t('Model Distribution')}
            </CardTitle>
            <CardDescription>{t('Usage by AI model')}</CardDescription>
          </CardHeader>
          <CardContent>
            {hasModelData ? (
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
                      width={90}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'var(--background)',
                        border: '1px solid var(--border)',
                        borderRadius: '8px',
                        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                      }}
                      formatter={((value: number, name: string): any => {
                        if (name === 'requests') return [value.toLocaleString(), t('Requests')]
                        if (name === 'percentage') return [`${value}%`, t('Percentage')]
                        return [value, name]
                      }) as any}
                    />
                    <Bar dataKey="requests" fill="#8b5cf6" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <EmptyState
                icon={Network}
                title={t('No model data yet')}
                description={t('Model usage distribution will appear once you start making API calls')}
              />
            )}
          </CardContent>
        </Card>
      </div>

      {/* ── Top Models Detail List ─────────────────────────────────── */}
      {topModels.length > 0 && (
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart4 className="h-5 w-5 text-green-600" />
              {t('Top Models')}
            </CardTitle>
            <CardDescription>{t('Most used models by request count')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {topModels.map((model, index) => (
                <div
                  key={model.model_name}
                  className="flex items-center justify-between p-3 sm:p-4 rounded-lg bg-muted/50 hover:bg-muted/70 transition-colors gap-3"
                >
                  <div className="flex items-center gap-3 sm:gap-4 min-w-0">
                    <div
                      className={cn(
                        'flex items-center justify-center w-9 h-9 sm:w-10 sm:h-10 rounded-xl font-bold text-white shrink-0',
                        index === 0 && 'bg-gradient-to-br from-blue-500 to-blue-600',
                        index === 1 && 'bg-gradient-to-br from-purple-500 to-purple-600',
                        index === 2 && 'bg-gradient-to-br from-green-500 to-green-600',
                        index === 3 && 'bg-gradient-to-br from-orange-500 to-orange-600',
                        index === 4 && 'bg-gradient-to-br from-pink-500 to-pink-600',
                      )}
                    >
                      {index + 1}
                    </div>
                    <div className="min-w-0">
                      <div className="font-semibold text-sm sm:text-base truncate">{model.model_name}</div>
                      <div className="text-xs sm:text-sm text-muted-foreground">
                        {(model.token_count || 0).toLocaleString()} {t('tokens')} · {(model.quota_used || 0).toFixed(2)} {t('quota')}
                      </div>
                    </div>
                  </div>
                  <Badge variant="secondary" className="text-xs sm:text-sm px-2 sm:px-3 py-1 shrink-0">
                    {(model.request_count || 0).toLocaleString()} {t('requests')}
                  </Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* ── Recent API Calls ───────────────────────────────────────── */}
      <Card className="hover:shadow-lg transition-shadow">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <List className="h-5 w-5 text-amber-600" />
              {t('Recent API Calls')}
            </CardTitle>
            <CardDescription>{t('Last 10 API call records')}</CardDescription>
          </div>
          {hasLogs && (
            <Badge variant="secondary" className="text-xs">
              {t('Latest')} {Math.min(10, logs.length)}/{logs.length}
            </Badge>
          )}
        </CardHeader>
        <CardContent>
          {hasLogs ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-muted-foreground">
                    <th className="text-left py-2 px-2 font-medium">{t('Date')}</th>
                    <th className="text-left py-2 px-2 font-medium">{t('Model')}</th>
                    <th className="text-right py-2 px-2 font-medium">{t('Requests')}</th>
                    <th className="text-right py-2 px-2 font-medium">{t('Tokens')}</th>
                    <th className="text-right py-2 px-2 font-medium">{t('Cost')}</th>
                  </tr>
                </thead>
                <tbody>
                  {recentLogs.map((log, i) => (
                    <tr key={i} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                      <td className="py-2.5 px-2 text-muted-foreground whitespace-nowrap">{log.day}</td>
                      <td className="py-2.5 px-2 font-medium truncate max-w-[200px]">{log.model_name}</td>
                      <td className="py-2.5 px-2 text-right">{(log.request_count || 0).toLocaleString()}</td>
                      <td className="py-2.5 px-2 text-right">{((log.prompt_tokens || 0) + (log.completion_tokens || 0)).toLocaleString()}</td>
                      <td className="py-2.5 px-2 text-right font-mono">
                        {new Intl.NumberFormat('en-US', { style: 'decimal', minimumFractionDigits: 4 }).format(log.quota)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState
              icon={Inbox}
              title={t('No API calls yet')}
              description={t('Your API call history will appear here once you start using the service')}
              action={{
                label: t('Get API Key'),
                onClick: () => window.location.href = '/keys',
              }}
            />
          )}
        </CardContent>
      </Card>
    </div>
  )
}
