import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import {
  Users,
  UserPlus,
  UserMinus,
  Activity,
  DollarSign,
  TrendingUp,
  Clock,
  ArrowUpRight,
  Mail,
  Shield,
  Copy,
  ExternalLink,
} from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import apiClient from '@/lib/api'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'

export const Route = createFileRoute('/_authenticated/team')({
  component: TeamPage,
  validateSearch: () => ({}),
})

function TeamPage() {
  const t = useT()
  const [copied, setCopied] = useState(false)

  // Fetch team members
  const { data: teamData, isLoading: teamLoading } = useQuery({
    queryKey: ['my-team'],
    queryFn: async () => {
      const res = await apiClient.get('/api/user/self/team')
      return res.data?.data || []
    },
  })

  // Fetch user self info (for referral link)
  const { data: userSelf } = useQuery({
    queryKey: ['user-self'],
    queryFn: async () => {
      const res = await apiClient.get('/api/user/self')
      return res.data?.data || {}
    },
  })

  // Fetch commission records
  const { data: commissionData } = useQuery({
    queryKey: ['commission-records'],
    queryFn: async () => {
      const res = await apiClient.get('/api/commission/self/records')
      return res.data?.data || []
    },
    retry: false,
  })

  const teamMembers = Array.isArray(teamData) ? teamData : []
  const hasTeam = teamMembers.length > 0
  const affCode = userSelf?.aff_code || ''

  // Calculate team stats
  const totalUsed = teamMembers.reduce((sum: number, m: any) => sum + (m.used_quota || 0), 0)
  const totalQuota = teamMembers.reduce((sum: number, m: any) => sum + (m.quota || 0), 0)

  // Copy referral link
  const referralLink = `${window.location.origin}/register?aff=${affCode}`
  const copyLink = () => {
    navigator.clipboard.writeText(referralLink)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  // Simulated weekly chart data (placeholder until real API exists)
  const chartData = [
    { day: 'Mon', members: 0, usage: 0 },
    { day: 'Tue', members: 0, usage: 0 },
    { day: 'Wed', members: 0, usage: 0 },
    { day: 'Thu', members: 0, usage: 0 },
    { day: 'Fri', members: 0, usage: 0 },
    { day: 'Sat', members: 0, usage: 0 },
    { day: 'Sun', members: 0, usage: 0 },
  ]

  return (
    <div className="qc-wrapper py-8">
      {/* Header */}
      <div className="flex flex-col gap-1 mb-8">
        <h1 className="text-2xl font-bold text-qc-amber-400">My Team</h1>
        <p className="text-qc-warm-400 text-sm max-w-[65ch]">
          Invite others to join QuantumClaw and earn commission from their usage. Share your referral link to grow your team.
        </p>
      </div>

      {/* Referral Link Section */}
      <Card className="mb-6 bg-gradient-to-br from-qc-amber-50/80 to-white border-qc-amber-200/50">
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold text-qc-warm-600 flex items-center gap-2 mb-1">
                <UserPlus className="w-4 h-4 text-qc-amber-400" />
                Your Referral Link
              </h3>
              <p className="text-xs text-qc-warm-400 mb-2">
                Share this link with friends. When they register, they'll automatically join your team.
              </p>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-3 py-1.5 bg-white/80 rounded-lg border text-sm font-mono text-qc-warm-500 truncate">
                  {affCode ? referralLink : 'Loading...'}
                </code>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={copyLink}
                  className="shrink-0 gap-1.5"
                >
                  {copied ? 'Copied!' : <><Copy className="w-3.5 h-3.5" /> Copy</>}
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-qc-warm-400">Team Members</p>
                <p className="text-2xl font-bold text-qc-warm-600 mt-1">
                  {teamLoading ? <Skeleton className="h-8 w-16" /> : teamMembers.length}
                </p>
              </div>
              <div className="w-10 h-10 rounded-full bg-qc-amber-100 flex items-center justify-center">
                <Users className="w-5 h-5 text-qc-amber-400" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-qc-warm-400">Total Usage</p>
                <p className="text-2xl font-bold text-qc-warm-600 mt-1">
                  {teamLoading ? <Skeleton className="h-8 w-20" /> : `${(totalUsed / 1000000).toFixed(2)}M`}
                </p>
              </div>
              <div className="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center">
                <Activity className="w-5 h-5 text-blue-500" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-qc-warm-400">Commission</p>
                <p className="text-2xl font-bold text-qc-warm-600 mt-1">
                  {commissionData?.length ? `${commissionData.length} records` : '0'}
                </p>
              </div>
              <div className="w-10 h-10 rounded-full bg-green-50 flex items-center justify-center">
                <DollarSign className="w-5 h-5 text-green-500" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-qc-warm-400">Active Members</p>
                <p className="text-2xl font-bold text-qc-warm-600 mt-1">
                  {teamLoading ? <Skeleton className="h-8 w-16" /> : teamMembers.filter((m: any) => m.status === 1).length}
                </p>
              </div>
              <div className="w-10 h-10 rounded-full bg-purple-50 flex items-center justify-center">
                <TrendingUp className="w-5 h-5 text-purple-500" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Upgrade to Provider Card */}
      <Card className="mb-6 border-dashed border-qc-amber-300/50 bg-gradient-to-r from-qc-amber-50/40 to-transparent">
        <CardContent className="pt-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h3 className="font-semibold text-qc-warm-600 flex items-center gap-2">
              <Shield className="w-4 h-4 text-qc-amber-400" />
              Upgrade to Channel Partner
            </h3>
            <p className="text-xs text-qc-warm-400 mt-1">
              Become a channel partner to add your own API keys, set pricing, and earn profits. Manage everything from the Distributors dashboard.
            </p>
          </div>
          <Button
            className="shrink-0 bg-qc-amber-400 hover:bg-qc-amber-500 text-white"
            onClick={async () => {
              try {
                const res = await apiClient.post('/api/user/self/upgrade')
                if (res.data?.success) {
                  alert('Upgrade successful! You can now manage API channels.')
                  window.location.reload()
                } else {
                  alert(res.data?.message || 'Upgrade failed')
                }
              } catch (err) {
                alert('Failed to upgrade')
              }
            }}
          >
            <ArrowUpRight className="w-4 h-4 mr-1.5" />
            Upgrade Now
          </Button>
        </CardContent>
      </Card>

      {/* Team Members Table */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base text-qc-warm-600">Team Members</CardTitle>
          <CardDescription>
            {hasTeam
              ? `You have ${teamMembers.length} team member${teamMembers.length > 1 ? 's' : ''}`
              : 'No team members yet. Share your referral link to grow your team.'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {teamLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map(i => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : hasTeam ? (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Username</TableHead>
                    <TableHead>Display Name</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Used Quota</TableHead>
                    <TableHead className="text-right">Requests</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {teamMembers.map((member: any) => (
                    <TableRow key={member.id}>
                      <TableCell className="font-medium">{member.username}</TableCell>
                      <TableCell className="text-qc-warm-400">{member.display_name || '-'}</TableCell>
                      <TableCell>
                        <Badge variant={member.status === 1 ? 'default' : 'secondary'}
                          className={cn(
                            member.status === 1 ? 'bg-green-100 text-green-600' : 'bg-qc-warm-100 text-qc-warm-400'
                          )}
                        >
                          {member.status === 1 ? 'Active' : 'Disabled'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {(member.used_quota / 1000000).toFixed(2)}M
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {member.request_count?.toLocaleString() || 0}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <div className="flex flex-col items-center py-12 text-center">
              <div className="w-16 h-16 rounded-full bg-qc-amber-50 flex items-center justify-center mb-4">
                <UserPlus className="w-8 h-8 text-qc-amber-300" />
              </div>
              <h3 className="text-qc-warm-600 font-medium mb-1">Your Team is Empty</h3>
              <p className="text-qc-warm-400 text-sm max-w-sm">
                Share your referral link above with friends and colleagues. When they register,
                they'll automatically appear here.
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
