import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Gift, Calendar, CheckCircle2, Clock } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import apiClient from '@/lib/api'
import { type ApiResponse } from '@/lib/api-extended'
import { toast } from 'sonner'

interface CheckinRecord {
  id: number
  user_id: number
  date: string
  amount: number
  created_at: number
}

interface CheckinResponse extends ApiResponse<{
  checkin: boolean
  consecutive_days: number
  total_reward: number
  last_checkin_time: number
}> {}

interface CheckinHistoryResponse extends ApiResponse<CheckinRecord[]> {}

async function getCheckinStatus(): Promise<CheckinResponse> {
  const res = await apiClient.get('/api/checkin')
  return res.data
}

async function getCheckinHistory(): Promise<CheckinHistoryResponse> {
  const res = await apiClient.get('/api/checkin/history')
  return res.data
}

async function doCheckin(): Promise<ApiResponse<{ amount: number }>> {
  const res = await apiClient.post('/api/checkin')
  return res.data
}

export const Route = createFileRoute('/_authenticated/checkin')({
  component: CheckinPage,
})

function CheckinPage() {
  const { t } = useT()
  const queryClient = useQueryClient()

  const { data: statusData, isLoading, refetch } = useQuery({
    queryKey: ['checkin-status'],
    queryFn: getCheckinStatus,
    staleTime: 30 * 1000,
  })

  const { data: historyData } = useQuery({
    queryKey: ['checkin-history'],
    queryFn: getCheckinHistory,
    staleTime: 60 * 1000,
  })

  const checkinMutation = useMutation({
    mutationFn: doCheckin,
    onSuccess: (data) => {
      if (data.success) {
        toast.success(t('Check-in successful!') + ` +${data.data?.amount} ${t('Quota')}`)
        queryClient.invalidateQueries({ queryKey: ['checkin-status'] })
        queryClient.invalidateQueries({ queryKey: ['checkin-history'] })
      }
    },
    onError: (error: any) => {
      toast.error(error?.response?.data?.message || t('Check-in failed'))
    },
  })

  const status = statusData?.data
  const history: CheckinRecord[] = historyData?.data || []

  const canCheckin = !status?.checkin
  const consecutiveDays = status?.consecutive_days || 0
  const totalReward = status?.total_reward || 0
  const lastCheckinTime = status?.last_checkin_time
    ? new Date(status.last_checkin_time * 1000)
    : null

  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString()
  }

  return (
    <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">{t('Daily Check-in')}</h1>
          <p className="text-muted-foreground mt-2">
            {t('Check in daily to earn rewards and maintain your streak')}
          </p>
        </div>
      </div>

      <div className="grid gap-6 grid-cols-1 md:grid-cols-2">
        {/* Check-in Card */}
        <Card className="relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-purple-500/10 to-pink-500/10" />
          <CardHeader className="relative">
            <div className="flex items-center gap-2">
              <Gift className="h-6 w-6 text-purple-600" />
              <CardTitle>{t('Daily Check-in')}</CardTitle>
            </div>
            <CardDescription>
              {t('Check in every day to earn quota rewards')}
            </CardDescription>
          </CardHeader>
          <CardContent className="relative space-y-6">
            {canCheckin ? (
              <div className="text-center space-y-4">
                <div className="mx-auto w-32 h-32 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center">
                  <Gift className="h-16 w-16 text-white" />
                </div>
                <div>
                  <p className="text-lg font-medium">{t('Ready to check in?')}</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    {t('You can claim your daily reward now')}
                  </p>
                </div>
                <Button
                  size="lg"
                  onClick={() => checkinMutation.mutate()}
                  disabled={checkinMutation.isPending}
                  className="w-full sm:w-auto"
                >
                  {checkinMutation.isPending ? (
                    <>
                      <Clock className="mr-2 h-4 w-4 animate-spin" />
                      {t('Checking in...')}
                    </>
                  ) : (
                    <>
                      <Gift className="mr-2 h-4 w-4" />
                      {t('Check In Now')}
                    </>
                  )}
                </Button>
              </div>
            ) : (
              <div className="text-center space-y-4">
                <div className="mx-auto w-32 h-32 rounded-full bg-green-500/20 flex items-center justify-center">
                  <CheckCircle2 className="h-16 w-16 text-green-600" />
                </div>
                <div>
                  <p className="text-lg font-medium text-green-600">{t('Checked in today!')}</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    {lastCheckinTime && (
                      <>
                        {t('Last check-in')}: {lastCheckinTime.toLocaleString()}
                      </>
                    )}
                  </p>
                </div>
                <Button size="lg" disabled className="w-full sm:w-auto">
                  <CheckCircle2 className="mr-2 h-4 w-4" />
                  {t('Already Checked In')}
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Stats Card */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Calendar className="h-6 w-6 text-blue-600" />
              <CardTitle>{t('Check-in Stats')}</CardTitle>
            </div>
            <CardDescription>
              {t('Track your check-in streak and rewards')}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div className="text-center p-4 rounded-xl bg-gradient-to-br from-purple-500/10 to-pink-500/10">
                <p className="text-3xl font-bold text-purple-600">{consecutiveDays}</p>
                <p className="text-sm text-muted-foreground mt-1">{t('Consecutive Days')}</p>
              </div>
              <div className="text-center p-4 rounded-xl bg-gradient-to-br from-green-500/10 to-emerald-500/10">
                <p className="text-3xl font-bold text-green-600">{totalReward}</p>
                <p className="text-sm text-muted-foreground mt-1">{t('Total Rewards')}</p>
              </div>
            </div>

            <div className="space-y-2">
              <h4 className="font-medium text-sm">{t('Consecutive Check-in Rewards')}</h4>
              <div className="space-y-2">
                {[3, 7, 30].map((days) => (
                  <div key={days} className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">
                      {t('{{days}} days', { days })}
                    </span>
                    <Badge variant={consecutiveDays >= days ? 'default' : 'outline'}>
                      {days === 3 ? '+10' : days === 7 ? '+30' : '+100'} {t('Quota')}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* History Card */}
      <Card>
        <CardHeader>
          <CardTitle>{t('Check-in History')}</CardTitle>
          <CardDescription>
            {t('Recent check-in records')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {history.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Calendar className="mx-auto h-12 w-12 mb-3 opacity-50" />
              <p>{t('No check-in history yet')}</p>
              <p className="text-sm mt-1">{t('Start checking in to see your history here')}</p>
            </div>
          ) : (
            <div className="space-y-3">
              {history.slice(0, 10).map((record) => (
                <div
                  key={record.id}
                  className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-3 rounded-xl border"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-green-500/20 flex items-center justify-center">
                      <CheckCircle2 className="h-5 w-5 text-green-600" />
                    </div>
                    <div>
                      <p className="font-medium">{formatDate(record.created_at)}</p>
                      <p className="text-sm text-muted-foreground">{record.date}</p>
                    </div>
                  </div>
                  <Badge variant="outline" className="text-green-600">
                    +{record.amount} {t('Quota')}
                  </Badge>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
