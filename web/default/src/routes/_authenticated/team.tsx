import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import {
  Users,
  UserPlus,
  UserMinus,
  Activity,
  DollarSign,
  TrendingUp,
  Copy,
  ArrowUpRight,
  Shield,
  Building2,
  Plus,
  LogOut,
  Mail,
} from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import apiClient from '@/lib/api'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/team')({
  component: TeamPage,
  validateSearch: () => ({}),
})

function TeamPage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<'referral' | 'org'>('referral')
  const [copied, setCopied] = useState(false)
  const [newOrgName, setNewOrgName] = useState('')
  const [inviteUsername, setInviteUsername] = useState('')
  const [inviteOrgId, setInviteOrgId] = useState<number | null>(null)
  const [showCreateOrg, setShowCreateOrg] = useState(false)

  // ── Fetch team members (inviter-based) ──
  const { data: teamData, isLoading: teamLoading } = useQuery({
    queryKey: ['my-team'],
    queryFn: async () => {
      const res = await apiClient.get('/api/user/self/team')
      return res.data?.data || []
    },
  })

  // ── Fetch user self info ──
  const { data: userSelf } = useQuery({
    queryKey: ['user-self'],
    queryFn: async () => {
      const res = await apiClient.get('/api/user/self')
      return res.data?.data || {}
    },
  })

  // ── Fetch organizations ──
  const { data: orgsData, isLoading: orgsLoading } = useQuery({
    queryKey: ['my-orgs'],
    queryFn: async () => {
      const res = await apiClient.get('/api/user/self/organizations')
      return res.data?.data || []
    },
  })

  // ── Fetch members for a specific org ──
  const fetchOrgMembers = (orgId: number) => ({
    queryKey: ['org-members', orgId],
    queryFn: async () => {
      const res = await apiClient.get(`/api/user/self/organizations/${orgId}/members`)
      return res.data?.data || []
    },
  })

  // ── Fetch commission records ──
  const { data: commissionData } = useQuery({
    queryKey: ['commission-records'],
    queryFn: async () => {
      const res = await apiClient.get('/api/commission/self/records')
      return res.data?.data || []
    },
    retry: false,
  })

  // ── Create organization mutation ──
  const createOrgMutation = useMutation({
    mutationFn: async (name: string) => {
      const res = await apiClient.post('/api/user/self/organizations', { name })
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-orgs'] })
      queryClient.invalidateQueries({ queryKey: ['user-self'] })
      setShowCreateOrg(false)
      setNewOrgName('')
    },
  })

  // ── Invite member mutation ──
  const inviteMutation = useMutation({
    mutationFn: async ({ orgId, username }: { orgId: number; username: string }) => {
      const res = await apiClient.post(`/api/user/self/organizations/${orgId}/invite`, { username })
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-members'] })
      setInviteUsername('')
      setInviteOrgId(null)
    },
  })

  // ── Remove member mutation ──
  const removeMutation = useMutation({
    mutationFn: async ({ orgId, userId }: { orgId: number; userId: number }) => {
      const res = await apiClient.delete(`/api/user/self/organizations/${orgId}/members/${userId}`)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-members'] })
    },
  })

  const teamMembers = Array.isArray(teamData) ? teamData : []
  const orgs = Array.isArray(orgsData) ? orgsData : []
  const hasTeam = teamMembers.length > 0
  const affCode = userSelf?.aff_code || ''
  const currentOrgId = userSelf?.organization_id || 0

  // Stats
  const totalUsed = teamMembers.reduce((sum: number, m: any) => sum + (m.used_quota || 0), 0)

  // ── Copy referral link ──
  const referralLink = `${window.location.origin}/sign-in?ref=${affCode}`
  const copyLink = () => {
    navigator.clipboard.writeText(referralLink)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  // ── Create Org ──
  const handleCreateOrg = () => {
    if (!newOrgName.trim()) return
    createOrgMutation.mutate(newOrgName.trim())
  }

  // ── Invite ──
  const handleInvite = (orgId: number) => {
    if (!inviteUsername.trim()) return
    inviteMutation.mutate({ orgId, username: inviteUsername.trim() })
  }

  return (
    <div className="qc-wrapper py-8">
      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-qc-warm-50 p-1 rounded-xl w-fit">
        <button
          onClick={() => setTab('referral')}
          className={cn(
            'px-5 py-2.5 rounded-lg text-sm font-medium transition-all',
            tab === 'referral'
              ? 'bg-white shadow-sm text-qc-warm-600'
              : 'text-qc-warm-400 hover:text-qc-warm-500'
          )}
        >
          <Users className="w-4 h-4 inline mr-1.5" />
          {t('Referral Team')}
        </button>
        <button
          onClick={() => setTab('org')}
          className={cn(
            'px-5 py-2.5 rounded-lg text-sm font-medium transition-all',
            tab === 'org'
              ? 'bg-white shadow-sm text-qc-warm-600'
              : 'text-qc-warm-400 hover:text-qc-warm-500'
          )}
        >
          <Building2 className="w-4 h-4 inline mr-1.5" />
          {t('Organizations')}
        </button>
      </div>

      {/* ══════ REFERRAL TAB ══════ */}
      {tab === 'referral' && (
        <>
          {/* Referral Link */}
          <Card className="mb-6 bg-gradient-to-br from-qc-amber-50/80 to-white border-qc-amber-200/50">
            <CardContent className="pt-6">
              <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold text-qc-warm-600 flex items-center gap-2 mb-1">
                    <UserPlus className="w-4 h-4 text-qc-amber-400" />
                    {t("Your Referral Link")}
                  </h3>
                  <p className="text-xs text-qc-warm-400 mb-2">
                    {t("Share this link with friends. When they register, they'll automatically join your team.")}
                  </p>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 px-3 py-1.5 bg-white/80 rounded-lg border text-sm font-mono text-qc-warm-500 truncate">
                      {affCode ? referralLink : t("Loading")}
                    </code>
                    <Button variant="outline" size="sm" onClick={copyLink} className="shrink-0 gap-1.5">
                      {copied ? t('Copied') : <><Copy className="w-3.5 h-3.5" /> {t('Copy')}</>}
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Stats */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs text-qc-warm-400">{t("Team Members")}</p>
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
                    <p className="text-xs text-qc-warm-400">{t("Total Usage")}</p>
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
                    <p className="text-xs text-qc-warm-400">{t("Commission")}</p>
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
                    <p className="text-xs text-qc-warm-400">{t("Active Members")}</p>
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

          {/* Upgrade to Provider */}
          <Card className="mb-6 border-dashed border-qc-amber-300/50 bg-gradient-to-r from-qc-amber-50/40 to-transparent">
            <CardContent className="pt-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              <div>
                <h3 className="font-semibold text-qc-warm-600 flex items-center gap-2">
                  <Shield className="w-4 h-4 text-qc-amber-400" />
                  {t("Upgrade to Channel Partner")}
                </h3>
                <p className="text-xs text-qc-warm-400 mt-1">
                  {t("Become a channel partner to add your own API keys, set pricing, and earn profits.")}
                </p>
              </div>
              <Button
                className="shrink-0 bg-qc-amber-400 hover:bg-qc-amber-500 text-white"
                onClick={async () => {
                  try {
                    const res = await apiClient.post('/api/user/self/upgrade')
                    if (res.data?.success) {
                      alert(t('Upgrade successful!'))
                      window.location.reload()
                    } else {
                      alert(res.data?.message || t('Upgrade failed'))
                    }
                  } catch { alert(t('Failed to upgrade')) }
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
              <CardTitle className="text-base text-qc-warm-600">{t("Team Members")}</CardTitle>
              <CardDescription>
                {hasTeam
                  ? `You have ${teamMembers.length} team member${teamMembers.length > 1 ? 's' : ''}`
                  : 'No team members yet. Share your referral link to grow your team.'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {teamLoading ? (
                <div className="space-y-3">
                  {[1, 2, 3].map(i => <Skeleton key={i} className="h-12 w-full" />)}
                </div>
              ) : hasTeam ? (
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t("Username")}</TableHead>
                        <TableHead>{t("Display Name")}</TableHead>
                        <TableHead>{t("Status")}</TableHead>
                        <TableHead className="text-right">{t("Used Quota")}</TableHead>
                        <TableHead className="text-right">{t("Requests")}</TableHead>
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
                              {member.status === 1 ? t('Active') : t('Disabled')}
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
                  <h3 className="text-qc-warm-600 font-medium mb-1">{t("Your Team is Empty")}</h3>
                  <p className="text-qc-warm-400 text-sm max-w-sm">
                    Share your referral link above. When they register, they'll appear here.
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}

      {/* ══════ ORGANIZATIONS TAB ══════ */}
      {tab === 'org' && (
        <>
          {/* Create org button */}
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-lg font-semibold text-qc-warm-600">{t("My Organizations")}</h2>
              <p className="text-sm text-qc-warm-400">{t("Organizations let you manage team members and permissions together.")}</p>
            </div>
            <Dialog open={showCreateOrg} onOpenChange={setShowCreateOrg}>
              <DialogTrigger asChild>
                <Button className="bg-qc-amber-400 hover:bg-qc-amber-500 text-white gap-1.5">
                  <Plus className="w-4 h-4" /> {t("Create Organization")}
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>{t("Create Organization")}</DialogTitle>
                </DialogHeader>
                <div className="flex flex-col gap-4 pt-2">
                  <Input
                    placeholder={t("Organization name")}
                    value={newOrgName}
                    onChange={e => setNewOrgName(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && handleCreateOrg()}
                  />
                  <Button onClick={handleCreateOrg} disabled={!newOrgName.trim() || createOrgMutation.isPending}>
                    {createOrgMutation.isPending ? t('Creating...') : t('Create')}
                  </Button>
                  {createOrgMutation.isError && (
                    <p className="text-red-500 text-sm">{(createOrgMutation.error as any)?.message || t('Failed to create')}</p>
                  )}
                </div>
              </DialogContent>
            </Dialog>
          </div>

          {/* Organization list */}
          {orgsLoading ? (
            <div className="space-y-4">
              {[1, 2].map(i => <Skeleton key={i} className="h-40 w-full" />)}
            </div>
          ) : orgs.length === 0 ? (
            <div className="flex flex-col items-center py-16 text-center">
              <div className="w-16 h-16 rounded-full bg-qc-amber-50 flex items-center justify-center mb-4">
                <Building2 className="w-8 h-8 text-qc-amber-300" />
              </div>
              <h3 className="text-qc-warm-600 font-medium mb-1">{t("No Organizations Yet")}</h3>
              <p className="text-qc-warm-400 text-sm max-w-sm mb-4">
                {t("Create an organization to manage your team members and their permissions.")}
              </p>
              <Button
                className="bg-qc-amber-400 hover:bg-qc-amber-500 text-white gap-1.5"
                onClick={() => setShowCreateOrg(true)}
              >
                <Plus className="w-4 h-4" /> {t("Create Your First Organization")}
              </Button>
            </div>
          ) : (
            <div className="space-y-6">
              {orgs.map((org: any) => (
                <OrgCard
                  key={org.id}
                  org={org}
                  isActive={org.id === currentOrgId}
                  fetchMembers={fetchOrgMembers}
                  onInvite={handleInvite}
                  onRemove={(userId) => removeMutation.mutate({ orgId: org.id, userId })}
                  inviteUsername={inviteOrgId === org.id ? inviteUsername : ''}
                  onInviteUsernameChange={(val) => { setInviteUsername(val); setInviteOrgId(org.id) }}
                  invitePending={inviteMutation.isPending && inviteOrgId === org.id}
                />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}

// ── Org Card Component ──
function OrgCard({
  org,
  isActive,
  fetchMembers,
  onInvite,
  onRemove,
  inviteUsername,
  onInviteUsernameChange,
  invitePending,
}: {
  org: any
  isActive: boolean
  fetchMembers: (id: number) => any
  onInvite: (orgId: number) => void
  onRemove: (userId: number) => void
  inviteUsername: string
  onInviteUsernameChange: (val: string) => void
  invitePending: boolean
}) {
  const { t } = useT()
  const [showMembers, setShowMembers] = useState(false)

  const { data: membersData, isLoading: membersLoading } = useQuery({
    ...fetchMembers(org.id),
    enabled: showMembers,
  })

  const members = Array.isArray(membersData) ? membersData : []
  const isAdmin = members.find((m: any) => m.role === 'admin' && m.user_id === undefined) // place-holder; proper check from API

  return (
    <Card className={cn('border', isActive ? 'border-qc-amber-300' : '')}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div>
            <CardTitle className="text-base text-qc-warm-600 flex items-center gap-2">
              <Building2 className="w-4 h-4 text-qc-amber-400" />
              {org.name}
              {isActive && (
                <Badge className="bg-qc-amber-100 text-qc-amber-600 text-xs">{t('Current')}</Badge>
              )}
            </CardTitle>
            <CardDescription>
              {org.member_count || 0} {t('members')} · {t('Created')} {new Date(org.created_at).toLocaleDateString()}
            </CardDescription>
          </div>
          <Button variant="outline" size="sm" onClick={() => setShowMembers(!showMembers)}>
            {showMembers ? t('Hide Members') : t('View Members')}
          </Button>
        </div>
      </CardHeader>

      {showMembers && (
        <CardContent>
          {membersLoading ? (
            <div className="space-y-2">
              {[1, 2].map(i => <Skeleton key={i} className="h-10 w-full" />)}
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("Username")}</TableHead>
                    <TableHead>{t("Role")}</TableHead>
                    <TableHead>{t("Joined")}</TableHead>
                    <TableHead className="w-20"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {members.map((member: any) => (
                    <TableRow key={member.user_id}>
                      <TableCell className="font-medium">{member.username}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className={cn(
                          member.role === 'admin' ? 'border-qc-amber-300 text-qc-amber-600' : 'text-qc-warm-400',
                        )}>
                          {member.role === 'admin' ? t('Admin') : t('Member')}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm text-qc-warm-400">{member.joined_at}</TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-red-400 hover:text-red-500 hover:bg-red-50"
                          onClick={() => onRemove(member.user_id)}
                        >
                          <UserMinus className="w-4 h-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {/* Invite form */}
              <div className="flex items-center gap-2 mt-4 pt-4 border-t">
                <Input
                  placeholder={t("Enter username to invite")}
                  value={inviteUsername}
                  onChange={e => onInviteUsernameChange(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && onInvite(org.id)}
                  className="flex-1"
                />
                <Button
                  size="sm"
                  className="bg-qc-amber-400 hover:bg-qc-amber-500 text-white gap-1"
                  onClick={() => onInvite(org.id)}
                  disabled={!inviteUsername.trim() || invitePending}
                >
                  <UserPlus className="w-3.5 h-3.5" />
                  {invitePending ? t('Inviting...') : t('Invite')}
                </Button>
              </div>
            </>
          )}
        </CardContent>
      )}
    </Card>
  )
}
