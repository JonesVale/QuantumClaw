import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Users, DollarSign, Activity, Server, TrendingUp, ExternalLink } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import apiClient from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/_authenticated/provider-analytics')({
  component: ProviderAnalytics,
})

function ProviderAnalytics() {
  const { t } = useT()
  const { auth } = useAuthStore()
  const [days, setDays] = useState('30')

  const { data, isLoading } = useQuery({
    queryKey: ['provider-customers', days],
    queryFn: async () => {
      const r = await apiClient.get(`/api/provider/customers?days=${days}`)
      return r.data?.data
    },
    enabled: !!auth.user?.id,
  })

  const customers = data?.customers || []
  const channels = data?.channels || []
  const totalRevenue = customers.reduce((s: number, c: any) => s + c.total_amount, 0)
  const totalCommission = customers.reduce((s: number, c: any) => s + c.commission, 0)

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2"><Users className="w-7 h-7 inline mr-2" />{t('Provider Analytics')}</h1>
        <p className="text-muted-foreground">{t('See who is using your channels and how much you earn')}</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card><CardContent className="py-5 text-center">
          <DollarSign className="w-5 h-5 mx-auto mb-1 text-amber-500" />
          <p className="text-2xl font-bold">${totalRevenue.toFixed(2)}</p>
          <p className="text-xs text-muted-foreground">{t('Total Revenue')}</p>
        </CardContent></Card>
        <Card><CardContent className="py-5 text-center">
          <TrendingUp className="w-5 h-5 mx-auto mb-1 text-green-500" />
          <p className="text-2xl font-bold">${totalCommission.toFixed(2)}</p>
          <p className="text-xs text-muted-foreground">{t('Commission Earned')}</p>
        </CardContent></Card>
        <Card><CardContent className="py-5 text-center">
          <Users className="w-5 h-5 mx-auto mb-1 text-blue-500" />
          <p className="text-2xl font-bold">{customers.length}</p>
          <p className="text-xs text-muted-foreground">{t('Customers')}</p>
        </CardContent></Card>
        <Card><CardContent className="py-5 text-center">
          <Server className="w-5 h-5 mx-auto mb-1 text-purple-500" />
          <p className="text-2xl font-bold">{channels.length}</p>
          <p className="text-xs text-muted-foreground">{t('Active Channels')}</p>
        </CardContent></Card>
      </div>

      {/* Period selector */}
      <div className="flex justify-end">
        <Select value={days} onValueChange={setDays}>
          <SelectTrigger className="w-32"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="7">{t('Last 7 days')}</SelectItem>
            <SelectItem value="30">{t('Last 30 days')}</SelectItem>
            <SelectItem value="90">{t('Last 90 days')}</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Top Customers */}
      <Card>
        <CardHeader><CardTitle><Users className="w-5 h-5 inline mr-2" />{t('Top Customers')}</CardTitle></CardHeader>
        <CardContent className="p-0">
          {isLoading ? <Skeleton className="h-32 m-4" /> : customers.length === 0 ? (
            <p className="text-sm text-muted-foreground p-6 text-center">{t('No customer data yet')}</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Customer')}</TableHead>
                  <TableHead className="text-right">{t('Revenue')}</TableHead>
                  <TableHead className="text-right">{t('Commission')}</TableHead>
                  <TableHead className="text-right">{t('Requests')}</TableHead>
                  <TableHead className="text-right">{t('Last Active')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {customers.map((c: any) => (
                  <TableRow key={c.user_id}>
                    <TableCell className="font-medium">{c.username || `User #${c.user_id}`}</TableCell>
                    <TableCell className="text-right font-mono text-sm">${c.total_amount.toFixed(2)}</TableCell>
                    <TableCell className="text-right font-mono text-sm text-green-600">${c.commission.toFixed(2)}</TableCell>
                    <TableCell className="text-right">{c.request_count.toLocaleString()}</TableCell>
                    <TableCell className="text-right text-xs text-muted-foreground">
                      {c.last_active ? dayjs(c.last_active * 1000).format('MM-DD HH:mm') : '-'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Channel Performance */}
      <Card>
        <CardHeader><CardTitle><Server className="w-5 h-5 inline mr-2" />{t('Channel Performance')}</CardTitle></CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Channel')}</TableHead>
                <TableHead className="text-right">{t('Revenue')}</TableHead>
                <TableHead className="text-right">{t('Requests')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {channels.map((ch: any) => (
                <TableRow key={ch.channel_id}>
                  <TableCell className="font-medium">{ch.channel_name || `#${ch.channel_id}`}</TableCell>
                  <TableCell className="text-right font-mono text-sm">${ch.total_amount.toFixed(2)}</TableCell>
                  <TableCell className="text-right">{ch.request_count.toLocaleString()}</TableCell>
                </TableRow>
              ))}
              {channels.length === 0 && (
                <TableRow><TableCell colSpan={3} className="text-center text-muted-foreground py-8">{t('No channel data')}</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Team Performance (dept_teams) */}
      {data?.dept_teams?.length > 0 && (
        <Card>
          <CardHeader><CardTitle><Users className="w-5 h-5 inline mr-2" />{t('Team Performance')}</CardTitle></CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Department')}</TableHead>
                  <TableHead className="text-right">{t('Revenue')}</TableHead>
                  <TableHead className="text-right">{t('Commission')}</TableHead>
                  <TableHead className="text-right">{t('Requests')}</TableHead>
                  <TableHead className="text-right">{t('Members')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.dept_teams.map((dt: any) => (
                  <TableRow key={dt.department_id}>
                    <TableCell className="font-medium">{dt.dept_name}</TableCell>
                    <TableCell className="text-right font-mono text-sm">${dt.total_amount.toFixed(2)}</TableCell>
                    <TableCell className="text-right font-mono text-sm text-green-600">${dt.commission.toFixed(2)}</TableCell>
                    <TableCell className="text-right">{dt.request_count.toLocaleString()}</TableCell>
                    <TableCell className="text-right">{dt.member_count}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
