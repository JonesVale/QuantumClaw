import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import {
  Building2, Users, Key, Shield, BarChart3, Plus, Edit3, Trash2,
  Save, Settings, DollarSign, Globe, CheckCircle, XCircle,
  Sliders, Eye, EyeOff, Filter
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import apiClient from '@/lib/api'
import {
  getDepartments, createDepartment, updateDepartment, deleteDepartment,
  getEnterprisePolicy, saveEnterprisePolicy,
  getOrgUsage, setEmployeeDepartment,
  createEnterpriseToken, getOrgEnterpriseTokens,
  getOrgApprovals, processApproval,
  getMyStoreModels, type Department, type EnterpriseTokenPolicy
} from '@/lib/api-extended'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/enterprise')({
  component: EnterprisePage,
})

function EnterprisePage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const { auth } = useAuthStore()
  const navigate = useNavigate()
  const [tab, setTab] = useState<'overview' | 'departments' | 'keys' | 'approvals' | 'policy'>('overview')
  const [showCreateDept, setShowCreateDept] = useState(false)

  const orgId = auth.user?.organization_id || 0

  // ── Self info ──
  const { data: selfData } = useQuery({ queryKey: ['self'], queryFn: async () => { const r = await fetch('/api/user/self'); return r.json() } })
  const currentOrgId = selfData?.data?.organization_id || orgId

  // Redirect if no org
  useEffect(() => {
    if (!currentOrgId && selfData) navigate({ to: '/team' })
  }, [currentOrgId, selfData])

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2"><Building2 className="w-7 h-7 inline mr-2" />{t('Enterprise Console')}</h1>
        <p className="text-muted-foreground">{t('Manage your organization, departments, API keys and security policies')}</p>
      </div>

      {/* Tab Switcher */}
      <div className="flex gap-1 mb-4 bg-muted/30 p-1 rounded-xl w-fit flex-wrap">
        {[
          { id: 'overview', icon: BarChart3, label: t('Overview') },
          { id: 'departments', icon: Building2, label: t('Departments') },
          { id: 'keys', icon: Key, label: t('API Keys') },
          { id: 'approvals', icon: Shield, label: t('Approvals') },
          { id: 'policy', icon: Shield, label: t('Security Policy') },
        ].map(tabItem => (
          <button key={tabItem.id} onClick={() => setTab(tabItem.id as any)}
            className={cn('px-4 py-2.5 rounded-lg text-sm font-medium transition-all flex items-center gap-1.5',
              tab === tabItem.id ? 'bg-white shadow-sm' : 'hover:bg-muted/20')}>
            <tabItem.icon className="w-4 h-4" />{tabItem.label}
          </button>
        ))}
      </div>

      {!currentOrgId ? <Card><CardContent className="py-12 text-center text-muted-foreground">{t('Please create or join an organization first')}</CardContent></Card> : (
        <>
          {tab === 'overview' && <OrgOverviewTab orgId={currentOrgId} />}
          {tab === 'departments' && <DepartmentsTab orgId={currentOrgId} />}
          {tab === 'keys' && <EnterpriseKeysTab orgId={currentOrgId} />}
          {tab === 'approvals' && <ApprovalsTab orgId={currentOrgId} />}
          {tab === 'policy' && <PolicyTab orgId={currentOrgId} />}
        </>
      )}
    </div>
  )
}

// ═══════ OVERVIEW TAB ═══════
function OrgOverviewTab({ orgId }: { orgId: number }) {
  const { t } = useT()
  const { data: depts } = useQuery({ queryKey: ['org-depts', orgId], queryFn: () => getDepartments(orgId), enabled: !!orgId })
  const { data: usage } = useQuery({ queryKey: ['org-usage', orgId], queryFn: () => getOrgUsage(orgId), enabled: !!orgId })
  const departments: Department[] = depts?.data || []
  const monthlyUsage = usage?.data?.cost || 0

  return <div className="space-y-6">
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <Card><CardContent className="py-6 text-center"><p className="text-3xl font-bold">{departments.length}</p><p className="text-sm text-muted-foreground mt-1">{t('Departments')}</p></CardContent></Card>
      <Card><CardContent className="py-6 text-center"><p className="text-3xl font-bold">{departments.reduce((s, d) => s + (d.member_count || 0), 0)}</p><p className="text-sm text-muted-foreground mt-1">{t('Members')}</p></CardContent></Card>
      <Card><CardContent className="py-6 text-center"><p className="text-3xl font-bold">¥{(Number(monthlyUsage) / 100).toFixed(2)}</p><p className="text-sm text-muted-foreground mt-1">{t('Monthly Usage')}</p></CardContent></Card>
    </div>

    <Card><CardHeader><CardTitle>{t('Departments')}</CardTitle></CardHeader><CardContent>
      {departments.length === 0 ? <p className="text-sm text-muted-foreground py-4 text-center">{t('No departments yet. Create one to start organizing your team.')}</p> : (
        <Table>
          <TableHeader><TableRow><TableHead>{t('Name')}</TableHead><TableHead>{t('Members')}</TableHead><TableHead>{t('Budget')}</TableHead><TableHead>{t('Usage')}</TableHead><TableHead>{t('Status')}</TableHead></TableRow></TableHeader>
          <TableBody>{departments.map(d => {
            const pct = d.monthly_budget > 0 ? ((monthlyUsage / d.monthly_budget) * 100).toFixed(0) : '-'
            return <TableRow key={d.id}>
              <TableCell className="font-medium">{d.name}</TableCell>
              <TableCell>{d.member_count || 0}</TableCell>
              <TableCell>¥{(d.monthly_budget / 100).toFixed(0)}</TableCell>
              <TableCell>
                <div className="flex items-center gap-2">
                  <div className="usage-bar w-16"><div className="usage-bar-fill bg-amber-500" style={{width: pct !== '-' ? Math.min(Number(pct), 100) + '%' : '0%'}}></div></div>
                  <span className="text-xs">{pct}%</span>
                </div>
              </TableCell>
              <TableCell><Badge variant={d.status === 1 ? 'default' : 'secondary'}>{d.status === 1 ? t('Active') : t('Disabled')}</Badge></TableCell>
            </TableRow>
          })}</TableBody>
        </Table>
      )}
    </CardContent></Card>

    {/* Org Chart - 3.3 */}
    {departments.length > 0 && (
      <Card>
        <CardHeader><CardTitle><Building2 className="w-5 h-5 inline mr-2" />{t('Organization Chart')}</CardTitle></CardHeader>
        <CardContent>
          <div className="flex flex-col items-center">
            <div className="px-6 py-3 rounded-xl bg-amber-50 border border-amber-200 font-semibold text-sm">{t('Organization')}</div>
            <div className="flex gap-6 mt-6 flex-wrap justify-center">
              {departments.map(d => (
                <div key={d.id} className="flex flex-col items-center">
                  <div className="px-4 py-2 rounded-lg bg-amber-50 border border-amber-100 text-sm font-medium">{d.name}</div>
                  <div className="flex gap-1 mt-2">
                    <span className="text-xs px-2 py-1 bg-gray-50 rounded">{d.member_count || 0} {t('members')}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>
    )}
  </div>
}

// ═══════ DEPARTMENTS TAB ═══════
function DepartmentsTab({ orgId }: { orgId: number }) {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [editDept, setEditDept] = useState<Department | null>(null)

  const { data: depts, refetch } = useQuery({ queryKey: ['org-depts', orgId], queryFn: () => getDepartments(orgId), enabled: !!orgId })
  const departments: Department[] = depts?.data || []

  const { data: orgMembers } = useQuery({ queryKey: ['org-members-all', orgId], queryFn: async () => {
    const r = await apiClient.get(`/api/user/self/organizations/${orgId}/members`)
    return r.data?.data || []
  }, enabled: !!orgId })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteDepartment(orgId, id),
    onSuccess: () => { toast.success(t('Department deleted')); refetch() },
    onError: () => toast.error(t('Failed to delete')),
  })

  return <div className="space-y-4">
    <div className="flex justify-between items-center">
      <p className="text-sm text-muted-foreground">{t('Organize your team into departments with budgets')}</p>
      <Button onClick={() => setShowCreate(true)}><Plus className="w-4 h-4 mr-2" />{t('Add Department')}</Button>
    </div>

    {departments.length === 0 ? (
      <Card><CardContent className="py-12 text-center text-muted-foreground">{t('No departments')}</CardContent></Card>
    ) : (
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {departments.map(d => (
          <Card key={d.id} className="hover:shadow-sm transition-shadow">
            <CardContent className="p-5">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <Building2 className="w-5 h-5 text-amber-500" />
                  <span className="font-semibold">{d.name}</span>
                  <Badge variant={d.status === 1 ? 'default' : 'secondary'} className="text-[10px]">{d.status === 1 ? t('Active') : t('Disabled')}</Badge>
                </div>
                <div className="flex gap-1">
                  <Button variant="ghost" size="sm" onClick={() => setEditDept(d)}><Edit3 className="w-4 h-4" /></Button>
                  <Button variant="ghost" size="sm" className="text-red-500" onClick={() => deleteMutation.mutate(d.id)}><Trash2 className="w-4 h-4" /></Button>
                </div>
              </div>
              {d.description && <p className="text-xs text-muted-foreground mb-3">{d.description}</p>}
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div><span className="text-muted-foreground">{t('Members')}</span><span className="ml-2 font-medium">{d.member_count || 0}</span></div>
                <div><span className="text-muted-foreground">{t('Budget')}</span><span className="ml-2 font-medium">¥{(d.monthly_budget / 100).toFixed(0)}</span></div>
              </div>

              {/* Member assignment */}
              <div className="mt-3 pt-3 border-t space-y-2">
                <p className="text-xs font-medium text-muted-foreground">{t('Assign Members')}</p>
                <MemberSelector orgId={orgId} departmentId={d.id} members={orgMembers || []} />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    )}

    <DeptDialog open={showCreate || !!editDept} onOpenChange={(o) => { if (!o) { setShowCreate(false); setEditDept(null) } }} orgId={orgId} edit={editDept} onDone={() => { setShowCreate(false); setEditDept(null); refetch() }} />
  </div>
}

function MemberSelector({ orgId, departmentId, members }: { orgId: number; departmentId: number; members: any[] }) {
  const { t } = useT()
  const [selectedUser, setSelectedUser] = useState('')
  const assignMutation = useMutation({
    mutationFn: (userId: number) => setEmployeeDepartment(orgId, userId, departmentId),
    onSuccess: () => { toast.success(t('Assigned')); setSelectedUser('') },
    onError: () => toast.error(t('Failed')),
  })
  const unassignMutation = useMutation({
    mutationFn: (userId: number) => setEmployeeDepartment(orgId, userId, 0),
    onSuccess: () => toast.success(t('Removed')),
    onError: () => toast.error(t('Failed')),
  })

  const deptMembers = members.filter((m: any) => m.department_id === departmentId)
  const otherMembers = members.filter((m: any) => m.department_id !== departmentId && m.department_id !== 0)

  return <div className="space-y-1">
    <div className="flex gap-1 flex-wrap">
      {deptMembers.slice(0, 5).map((m: any) => (
        <Badge key={m.user_id} variant="secondary" className="text-[10px] gap-1">
          {m.username}
          <button onClick={() => unassignMutation.mutate(m.user_id)} className="text-red-400 hover:text-red-600">×</button>
        </Badge>
      ))}
      {deptMembers.length > 5 && <Badge variant="outline" className="text-[10px]">+{deptMembers.length - 5}</Badge>}
    </div>
    <div className="flex gap-2">
      <select className="text-xs border rounded px-2 py-1 flex-1" value={selectedUser} onChange={e => setSelectedUser(e.target.value)}>
        <option value="">{t('Select member...')}</option>
        {members.filter((m: any) => m.department_id !== departmentId).map((m: any) => (
          <option key={m.user_id} value={m.user_id}>{m.username}</option>
        ))}
      </select>
      <Button size="sm" variant="outline" className="text-xs" disabled={!selectedUser} onClick={() => assignMutation.mutate(Number(selectedUser))}>{t('Add')}</Button>
    </div>
  </div>
}

function DeptDialog({ open, onOpenChange, orgId, edit, onDone }: { open: boolean; onOpenChange: (o: boolean) => void; orgId: number; edit: Department | null; onDone: () => void }) {
  const { t } = useT()
  const [form, setForm] = useState({ name: edit?.name || '', description: edit?.description || '', monthly_budget: edit?.monthly_budget || 0 })

  useEffect(() => {
    if (edit) setForm({ name: edit.name, description: edit.description || '', monthly_budget: edit.monthly_budget })
    else setForm({ name: '', description: '', monthly_budget: 0 })
  }, [edit])

  const mutation = useMutation({
    mutationFn: () => edit
      ? updateDepartment(orgId, edit.id, form)
      : createDepartment(orgId, form),
    onSuccess: () => { toast.success(edit ? t('Department updated') : t('Department created')); onDone(); onOpenChange(false) },
    onError: () => toast.error(t('Failed')),
  })

  return <Dialog open={open} onOpenChange={onOpenChange}>
    <DialogContent><DialogHeader><DialogTitle>{edit ? t('Edit Department') : t('Create Department')}</DialogTitle></DialogHeader>
      <div className="space-y-4">
        <div><Label>{t('Name')}</Label><Input value={form.name} onChange={e => setForm({...form, name: e.target.value})} /></div>
        <div><Label>{t('Description')}</Label><textarea className="w-full min-h-[60px] px-3 py-2 rounded-lg border bg-background text-sm" value={form.description} onChange={e => setForm({...form, description: e.target.value})} /></div>
        <div><Label>{t('Monthly Budget (cents)')}</Label><Input type="number" value={form.monthly_budget} onChange={e => setForm({...form, monthly_budget: Number(e.target.value)})} /></div>
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={() => onOpenChange(false)}>{t('Cancel')}</Button>
        <Button onClick={() => mutation.mutate()} disabled={mutation.isPending || !form.name}>{t('Save')}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
}

// ═══════ KEYS TAB ═══════
function EnterpriseKeysTab({ orgId }: { orgId: number }) {
  const { t } = useT()
  const { auth } = useAuthStore()
  const [showCreate, setShowCreate] = useState(false)
  const { data: tokens, refetch } = useQuery({ queryKey: ['org-enterprise-tokens', orgId], queryFn: () => getOrgEnterpriseTokens(orgId), enabled: !!orgId })
  const { data: depts } = useQuery({ queryKey: ['org-depts', orgId], queryFn: () => getDepartments(orgId), enabled: !!orgId })
  const { data: members } = useQuery({ queryKey: ['org-members-all', orgId], queryFn: async () => {
    const r = await apiClient.get(`/api/user/self/organizations/${orgId}/members`)
    return r.data?.data || []
  }, enabled: !!orgId })

  return <div className="space-y-4">
    <div className="flex justify-between items-center">
      <p className="text-sm text-muted-foreground">{t('Enterprise API keys with department binding and policy controls')}</p>
      <Button onClick={() => setShowCreate(true)}><Plus className="w-4 h-4 mr-2" />{t('Create Key')}</Button>
    </div>

    <Card><CardContent className="p-0">
      <Table>
        <TableHeader><TableRow>
          <TableHead>{t('Name')}</TableHead><TableHead>{t('Holder')}</TableHead><TableHead>{t('Department')}</TableHead>
          <TableHead>{t('Label')}</TableHead><TableHead>{t('Created')}</TableHead>
        </TableRow></TableHeader>
        <TableBody>
          {(tokens?.data || []).map((et: any) => (
            <TableRow key={et.id}>
              <TableCell className="font-medium">#{et.token_id}</TableCell>
              <TableCell>{et.created_by}</TableCell>
              <TableCell>{et.department_id || '-'}</TableCell>
              <TableCell><Badge variant="outline" className="text-[10px]">{et.label || '-'}</Badge></TableCell>
              <TableCell className="text-xs text-muted-foreground">{et.created_at}</TableCell>
            </TableRow>
          ))}
          {(tokens?.data || []).length === 0 && (
            <TableRow><TableCell colSpan={5} className="text-center py-8 text-muted-foreground">{t('No enterprise tokens yet')}</TableCell></TableRow>
          )}
        </TableBody>
      </Table>
    </CardContent></Card>

    <CreateEnterpriseKeyDialog orgId={orgId} open={showCreate} onOpenChange={setShowCreate}
      departments={depts?.data || []} members={members || []} onDone={() => { setShowCreate(false); refetch() }} />
  </div>
}

function CreateEnterpriseKeyDialog({ orgId, open, onOpenChange, departments, members, onDone }: {
  orgId: number; open: boolean; onOpenChange: (o: boolean) => void
  departments: any[]; members: any[]; onDone: () => void
}) {
  const { t } = useT()
  const [form, setForm] = useState({ user_id: 0, name: '', department_id: 0, purpose: '', label: '', remain_quota: 0, models: '', subnet: '' })

  const mutation = useMutation({
    mutationFn: () => createEnterpriseToken(orgId, {
      user_id: form.user_id, name: form.name,
      department_id: form.department_id || undefined,
      purpose: form.purpose, label: form.label,
      remain_quota: form.remain_quota || undefined,
      models: form.models || undefined, subnet: form.subnet || undefined,
    }),
    onSuccess: (res) => { toast.success(res.data?.message || t('Created')); onDone() },
    onError: () => toast.error(t('Failed')),
  })

  return <Dialog open={open} onOpenChange={onOpenChange}>
    <DialogContent className="max-w-lg"><DialogHeader><DialogTitle>{t('Create Enterprise Key')}</DialogTitle></DialogHeader>
      <div className="space-y-4">
        <div><Label>{t('Holder')}</Label>
          <select className="input-field w-full" value={form.user_id} onChange={e => setForm({...form, user_id: Number(e.target.value)})}>
            <option value={0}>{t('Select...')}</option>
            {members.map((m: any) => <option key={m.user_id} value={m.user_id}>{m.username}</option>)}
          </select>
        </div>
        <div><Label>{t('Key Name')}</Label><Input value={form.name} onChange={e => setForm({...form, name: e.target.value})} placeholder="Production Key" /></div>
        <div className="grid grid-cols-2 gap-4">
          <div><Label>{t('Department')}</Label>
            <select className="input-field w-full" value={form.department_id} onChange={e => setForm({...form, department_id: Number(e.target.value)})}>
              <option value={0}>{t('None')}</option>
              {departments.map((d: any) => <option key={d.id} value={d.id}>{d.name}</option>)}
            </select>
          </div>
          <div><Label>{t('Label')}</Label>
            <select className="input-field w-full" value={form.label} onChange={e => setForm({...form, label: e.target.value})}>
              <option value="">{t('None')}</option>
              <option value="production">{t('Production')}</option>
              <option value="test">{t('Test')}</option>
              <option value="dev">{t('Development')}</option>
            </select>
          </div>
        </div>
        <div><Label>{t('Purpose')}</Label><Input value={form.purpose} onChange={e => setForm({...form, purpose: e.target.value})} placeholder="e.g. Backend API server" /></div>
        <div><Label>{t('Quota (0 = policy default)')}</Label><Input type="number" value={form.remain_quota} onChange={e => setForm({...form, remain_quota: Number(e.target.value)})} /></div>
        <div><Label>{t('Models (comma separated, empty = policy default)')}</Label><Input value={form.models} onChange={e => setForm({...form, models: e.target.value})} /></div>
        <div><Label>{t('IP Whitelist')}</Label><Input value={form.subnet} onChange={e => setForm({...form, subnet: e.target.value})} placeholder="192.168.1.0/24" /></div>
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={() => onOpenChange(false)}>{t('Cancel')}</Button>
        <Button onClick={() => mutation.mutate()} disabled={mutation.isPending || !form.user_id || !form.name}>{t('Create')}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
}

// ═══════ APPROVALS TAB ═══════
function ApprovalsTab({ orgId }: { orgId: number }) {
  const { t } = useT()
  const [filter, setFilter] = useState('pending')
  const { data, refetch } = useQuery({ queryKey: ['org-approvals', orgId, filter], queryFn: () => getOrgApprovals(orgId, filter === 'all' ? undefined : filter), enabled: !!orgId })
  const approvals: any[] = data?.data || []

  const processMutation = useMutation({
    mutationFn: ({ id, action }: { id: number; action: 'approve' | 'reject' }) => processApproval(orgId, id, action),
    onSuccess: () => { toast.success(t('Processed')); refetch() },
    onError: () => toast.error(t('Failed')),
  })

  return <div className="space-y-4">
    <div className="flex gap-2">
      {['pending', 'approved', 'rejected', 'all'].map(s => (
        <Button key={s} variant={filter === s ? 'default' : 'outline'} size="sm" onClick={() => setFilter(s)}>
          {s === 'pending' ? t('Pending') : s === 'approved' ? t('Approved') : s === 'rejected' ? t('Rejected') : t('All')}
        </Button>
      ))}
    </div>

    <Card><CardContent className="p-0">
      <Table>
        <TableHeader><TableRow>
          <TableHead>{t('Type')}</TableHead><TableHead>{t('Requester')}</TableHead><TableHead>{t('Reason')}</TableHead>
          <TableHead>{t('Status')}</TableHead><TableHead>{t('Time')}</TableHead><TableHead></TableHead>
        </TableRow></TableHeader>
        <TableBody>
          {approvals.map((a: any) => (
            <TableRow key={a.id}>
              <TableCell><Badge variant="outline">{a.type}</Badge></TableCell>
              <TableCell>#{a.request_by}</TableCell>
              <TableCell className="max-w-40 truncate text-sm">{a.reason || '-'}</TableCell>
              <TableCell>
                {a.status === 'pending' ? <Badge className="bg-yellow-100 text-yellow-700">{t('Pending')}</Badge>
                  : a.status === 'approved' ? <Badge className="bg-green-100 text-green-700">{t('Approved')}</Badge>
                  : <Badge variant="secondary">{t('Rejected')}</Badge>}
              </TableCell>
              <TableCell className="text-xs text-muted-foreground">{a.created_at?.substring(0, 10)}</TableCell>
              <TableCell>
                {a.status === 'pending' && (
                  <div className="flex gap-1">
                    <Button size="sm" variant="outline" className="text-green-600 text-xs"
                      onClick={() => processMutation.mutate({ id: a.id, action: 'approve' })}
                      disabled={processMutation.isPending}>{t('Approve')}</Button>
                    <Button size="sm" variant="outline" className="text-red-600 text-xs"
                      onClick={() => processMutation.mutate({ id: a.id, action: 'reject' })}
                      disabled={processMutation.isPending}>{t('Reject')}</Button>
                  </div>
                )}
              </TableCell>
            </TableRow>
          ))}
          {approvals.length === 0 && <TableRow><TableCell colSpan={6} className="text-center py-8 text-muted-foreground">{t('No approvals')}</TableCell></TableRow>}
        </TableBody>
      </Table>
    </CardContent></Card>
  </div>
}

// ═══════ POLICY TAB ═══════
function PolicyTab({ orgId }: { orgId: number }) {
  const { t } = useT()
  const { data: policyData, refetch } = useQuery({ queryKey: ['org-policy', orgId], queryFn: () => getEnterprisePolicy(orgId), enabled: !!orgId })
  const policy: EnterpriseTokenPolicy | undefined = policyData?.data
  const [form, setForm] = useState<any>({})

  useEffect(() => {
    if (policy) {
      setForm({
        default_quota: policy.default_quota,
        default_group: policy.default_group,
        allowed_models: policy.allowed_models || '',
        blocked_models: policy.blocked_models || '',
        require_ip_whitelist: policy.require_ip_whitelist,
        max_keys_per_user: policy.max_keys_per_user,
        auto_expire_days: policy.auto_expire_days,
        require_admin_approval: policy.require_admin_approval,
      })
    }
  }, [policy])

  const saveMutation = useMutation({
    mutationFn: () => saveEnterprisePolicy(orgId, form),
    onSuccess: () => { toast.success(t('Policy saved')); refetch() },
    onError: () => toast.error(t('Failed to save policy')),
  })

  return <Card>
    <CardHeader><CardTitle><Shield className="w-5 h-5 inline mr-2" />{t('Security Policy')}</CardTitle>
      <CardDescription>{t('Set default token policies for your enterprise')}</CardDescription></CardHeader>
    <CardContent className="space-y-4 max-w-xl">
      <div className="grid grid-cols-2 gap-4">
        <div><Label>{t('Default Quota')}</Label><Input type="number" value={form.default_quota || 0} onChange={e => setForm({...form, default_quota: Number(e.target.value)})} /></div>
        <div><Label>{t('Default Group')}</Label><Input value={form.default_group || 'default'} onChange={e => setForm({...form, default_group: e.target.value})} /></div>
      </div>
      <div><Label>{t('Allowed Models (comma separated, empty = all)')}</Label><Input value={form.allowed_models || ''} onChange={e => setForm({...form, allowed_models: e.target.value})} /></div>
      <div><Label>{t('Blocked Models')}</Label><Input value={form.blocked_models || ''} onChange={e => setForm({...form, blocked_models: e.target.value})} /></div>
      <div className="grid grid-cols-2 gap-4">
        <div><Label>{t('Auto Expire (days)')}</Label><Input type="number" value={form.auto_expire_days || 0} onChange={e => setForm({...form, auto_expire_days: Number(e.target.value)})} /></div>
        <div><Label>{t('Max Keys Per User')}</Label><Input type="number" value={form.max_keys_per_user || 50} onChange={e => setForm({...form, max_keys_per_user: Number(e.target.value)})} /></div>
      </div>
      <div className="space-y-3">
        <label className="flex items-center gap-3 cursor-pointer"><Switch checked={!!form.require_ip_whitelist} onCheckedChange={v => setForm({...form, require_ip_whitelist: v})} /><span className="text-sm">{t('Require IP Whitelist')}</span></label>
        <label className="flex items-center gap-3 cursor-pointer"><Switch checked={!!form.require_admin_approval} onCheckedChange={v => setForm({...form, require_admin_approval: v})} /><span className="text-sm">{t('Require Admin Approval for Key Creation')}</span></label>
      </div>
      <Button onClick={() => saveMutation.mutate()} disabled={saveMutation.isPending}><Save className="w-4 h-4 mr-2" />{t('Save Policy')}</Button>
    </CardContent>
  </Card>
}
