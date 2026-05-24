import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useAuthStore } from '@/stores/auth-store'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plus,
  Search,
  MoreHorizontal,
  Pencil,
  Trash2,
  CheckCircle,
  XCircle,
  Shield,
  Users,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { EmptyState } from '@/components/empty-state'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  type User, type UserFormData,
  getUsers, createUser, updateUser, deleteUser, manageUser,
} from '@/lib/api-extended'
import { toast } from 'sonner'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/_authenticated/users')({
  component: UsersPage,
})

function UserFormDialog({
  open,
  onOpenChange,
  user,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  user?: User | null
}) {
  const { t } = useT()
  const queryClient = useQueryClient()
  const isEdit = !!user
  const [form, setForm] = useState<UserFormData>({
    username: '',
    password: '',
    display_name: '',
    email: '',
    role: 1,
    group: 'default',
    quota: 0,
  })

  const mutation = useMutation({
    mutationFn: isEdit
      ? (data: UserFormData & { id: number }) => updateUser(data)
      : createUser,
    onSuccess: () => {
      toast.success(isEdit ? t('User updated') : t('User created'))
      queryClient.invalidateQueries({ queryKey: ['users'] })
      onOpenChange(false)
    },
    onError: () => toast.error(t('Operation failed')),
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (isEdit && user) {
      mutation.mutate({ id: user.id, ...form })
    } else {
      mutation.mutate({ id: 0, ...form })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>{isEdit ? t('Edit User') : t('Create User')}</DialogTitle>
          <DialogDescription>
            {isEdit ? t('Update user information') : t('Create a new user account')}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>{t('Username')}</Label>
            <Input
              value={form.username}
              onChange={(e) => setForm({ ...form, username: e.target.value })}
              disabled={isEdit}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>{t('Password')}</Label>
            <Input
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              placeholder={isEdit ? t('Leave empty to keep unchanged') : ''}
              required={!isEdit}
            />
          </div>
          <div className="space-y-2">
            <Label>{t('Display Name')}</Label>
            <Input
              value={form.display_name}
              onChange={(e) => setForm({ ...form, display_name: e.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label>{t('Email')}</Label>
            <Input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label>{t('Role')}</Label>
            <Select
              value={String(form.role)}
              onValueChange={(v) => setForm({ ...form, role: Number(v) })}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1">{t('User')}</SelectItem>
                <SelectItem value="10">{t('Admin')}</SelectItem>
                <SelectItem value="100">{t('Super Admin')}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('Cancel')}
            </Button>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? t('Saving...') : isEdit ? t('Update') : t('Create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function UsersPage() {
  const { t } = useT()
  const { auth } = useAuthStore();
  const isAdmin = auth.user?.role === 100 || auth.user?.role === 10;
  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center min-h-[60vh] p-4">
        <div className="text-center max-w-md">
          <Shield className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
          <h2 className="text-2xl font-bold mb-2">{t('Access Denied')}</h2>
          <p className="text-muted-foreground">{t('You do not have permission to access this page.')}</p>
        </div>
      </div>
    );
  }
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['users', search],
    queryFn: () => getUsers(search ? { keyword: search } : undefined),
    staleTime: 30 * 1000,
  })

  const deleteMutation = useMutation({
    mutationFn: deleteUser,
    onSuccess: () => {
      toast.success(t('User deleted'))
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const manageMutation = useMutation({
    mutationFn: ({ id, action }: { id: number; action: 'disable' | 'enable' | 'reset_quota' | 'reset_used_quota' }) => manageUser({ id, action }),
    onSuccess: () => {
      toast.success(t('User status updated'))
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const users = data?.data || []

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Users')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('Manage user accounts')}</p>
      </div>

      <div className="flex flex-col sm:flex-row gap-3 items-start">
        <div className="relative w-full sm:max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="pl-9"
            placeholder={t('Search users...')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Button
          onClick={() => {
            setEditingUser(null)
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t('Create User')}
        </Button>
      </div>

      {!isLoading && users.length === 0 ? (
        <EmptyState
          icon={Users}
          title={t('No users found')}
          description={t('Create the first user account to get started')}
          action={{ label: t('Create User'), onClick: () => { setEditingUser(null); setDialogOpen(true) } }}
        />
      ) : (
        <Card className="bg-white/80 backdrop-blur-xl rounded-xl border overflow-hidden">
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">#</TableHead>
                  <TableHead>{t('Username')}</TableHead>
                  <TableHead>{t('Display Name')}</TableHead>
                  <TableHead>{t('Role')}</TableHead>
                  <TableHead>{t('Status')}</TableHead>
                  <TableHead>{t('Requests')}</TableHead>
                  <TableHead>{t('Created')}</TableHead>
                  <TableHead className="w-12" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      {Array.from({ length: 8 }).map((_, j) => (
                        <TableCell key={j}><Skeleton className="h-4 w-20" /></TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  users.map((u, idx) => (
                    <TableRow key={u.id}>
                      <TableCell className="text-muted-foreground">{idx + 1}</TableCell>
                      <TableCell className="font-medium">{u.username}</TableCell>
                      <TableCell>{u.display_name || '\u2014'}</TableCell>
                      <TableCell>
                        <Badge variant={u.role >= 100 ? 'default' : u.role >= 10 ? 'secondary' : 'outline'}>
                          {u.role >= 100 ? t('Super Admin') : u.role >= 10 ? t('Admin') : t('User')}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {u.status === 1 ? (
                          <Badge variant="default" className="bg-green-600">
                            <CheckCircle className="mr-1 h-3 w-3" /> {t('Active')}
                          </Badge>
                        ) : (
                          <Badge variant="secondary">
                            <XCircle className="mr-1 h-3 w-3" /> {t('Disabled')}
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell>{u.request_count?.toLocaleString() || 0}</TableCell>
                      <TableCell className="text-muted-foreground text-xs">
                        {dayjs((u.created_time || 0) * 1000).format('YYYY-MM-DD')}
                      </TableCell>
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => {
                              setEditingUser(u)
                              setDialogOpen(true)
                            }}>
                              <Pencil className="mr-2 h-4 w-4" /> {t('Edit')}
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() =>
                              manageMutation.mutate({
                                id: u.id,
                                action: u.status === 1 ? 'disable' : 'enable',
                              })
                            }>
                              {u.status === 1 ? (
                                <XCircle className="mr-2 h-4 w-4" />
                              ) : (
                                <CheckCircle className="mr-2 h-4 w-4" />
                              )}
                              {u.status === 1 ? t('Disable') : t('Enable')}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => {
                                if (confirm(t('Are you sure?'))) deleteMutation.mutate(u.id)
                              }}
                              className="text-destructive"
                            >
                              <Trash2 className="mr-2 h-4 w-4" /> {t('Delete')}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      <UserFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        user={editingUser}
      />
    </div>
  )
}
