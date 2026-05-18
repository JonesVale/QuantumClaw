import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { CreditCard, RefreshCw } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import apiClient from '@/lib/api'
import { type ApiResponse } from '@/lib/api-extended'

interface Subscription {
  id: number
  name: string
  status: string
  current_period_end: number
  cancel_at_period_end: boolean
  created_at: number
}

interface SubscriptionResponse extends ApiResponse<Subscription[]> {}

async function getSubscriptions(): Promise<SubscriptionResponse> {
  const res = await apiClient.get('/api/subscription')
  return res.data
}

export const Route = createFileRoute('/_authenticated/subscription')({
  component: SubscriptionPage,
})

function SubscriptionPage() {
  const { t } = useTranslation()

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['subscriptions'],
    queryFn: getSubscriptions,
    staleTime: 30 * 1000,
  })

  const subscriptions: Subscription[] = data?.data || []

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-green-500'
      case 'past_due':
        return 'bg-yellow-500'
      case 'canceled':
        return 'bg-red-500'
      default:
        return 'bg-gray-500'
    }
  }

  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString()
  }

  return (
    <div className=" w-full p-4 sm:p-6 space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('Subscription Management')}
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            {t('Manage your subscription plans and billing')}
          </p>
        </div>
        <Button variant="outline" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`mr-2 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('Refresh')}
        </Button>
      </div>

      {isLoading ? (
        <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <div className="h-6 w-32 animate-pulse rounded bg-muted" />
                <div className="h-4 w-24 animate-pulse rounded bg-muted" />
              </CardHeader>
              <CardContent>
                <div className="h-20 animate-pulse rounded bg-muted" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : subscriptions.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <CreditCard className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">{t('No subscriptions found')}</p>
            <p className="text-sm text-muted-foreground mt-2">
              {t('You can top up your quota in the Wallet section')}
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
          {subscriptions.map((sub) => (
            <Card key={sub.id} className="hover:shadow-lg transition-shadow">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg">{sub.name}</CardTitle>
                  <Badge
                    variant="secondary"
                    className={`${getStatusColor(sub.status)} text-white`}
                  >
                    {sub.status}
                  </Badge>
                </div>
                <CardDescription>
                  {t('Created')}: {formatDate(sub.created_at)}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">{t('Current Period End')}:</span>
                    <span className="font-medium">{formatDate(sub.current_period_end)}</span>
                  </div>
                  {sub.cancel_at_period_end && (
                    <div className="flex items-center gap-2 text-sm text-yellow-600 dark:text-yellow-400">
                      <span>鈿狅笍</span>
                      <span>{t('Subscription will be canceled at period end')}</span>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>{t('Need a Subscription?')}</CardTitle>
          <CardDescription>
            {t('Contact us to set up a custom subscription plan for your needs')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button asChild>
            <a href="mailto:support@quantumclaw.com">{t('Contact Us')}</a>
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
