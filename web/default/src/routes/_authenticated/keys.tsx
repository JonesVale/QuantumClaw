import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plus,
  Search,
  MoreHorizontal,
  Pencil,
  Trash2,
  Copy,
  CheckCircle,
  XCircle,
  Key,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
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
import { EmptyState } from '@/components/empty-state'
import {
  type Token, type TokenFormData,
  getTokens, createToken, updateToken, deleteToken, manageToken,
} from '@/lib/api-extended'
import { toast } from 'sonner'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/_authenticated/keys')({
  component: KeysPage,
})

function TokenFormDialog({
  open,
  onOpenChange,
  token,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  token?: Token | null
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!token
  const [form, setForm] = useState<TokenFormData>({
    name: '',
    remain_quota: 500000,
    expired_time: -1,
    unlimited_quota: false,
    models: '',
    subnet: '',
    group: 'default',
  })

  const mutation = useMutation({
    mutationFn: isEdit
      ? (data: TokenFormData & { id: number }) => updateToken(data)
      : createToken,
    onSuccess: (res) => {
      toast.success(isEdit ? t('Token updated') : t('Token created'))
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
      onOpenChange(false)
      if (res?.data?.key) {
        toast.info(t('Please save your API key, it will not be shown again'))
      }
    },
    onError: () => toast.error(isEdit ? t('Failed to update token') : t('Failed to create token')),
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (isEdit && token) {
      mutation.mutate({ id: token.id, ...form })
    } else {
      mutation.mutate(form)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{isEdit ? t('Edit Token') : t('Create Token')}</DialogTitle>
          <DialogDescription>
            {isEdit ? t('Update API key configuration') : t('Create a new API key')}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>{t('Token Name')}</Label>
            <Input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder={t('e.g. Production API Key')}
              required
            />
          </div>
          <div className="flex items-center gap-2">
            <Switch
              checked={form.unlimited_quota}
              onCheckedChange={(v) => setForm({ ...form, unlimited_quota: v })}
            />
            <Label>{t('Unlimited Quota')}</Label>
          </div>
          {!form.unlimited_quota && (
            <div className="space-y-2">
              <Label>{t('Quota')}</Label>
              <Input
                type="number"
                value={form.remain_quota || ''}
                onChange={(e) => setForm({ ...form, remain_quota: Number(e.target.value) })}
                placeholder="500000"
              />
            </div>
          )}
          <div className="space-y-2">
            <Label>{t('Models')}</Label>
            <Input
              value={form.models}
              onChange={(e) => setForm({ ...form, models: e.target.value })}
              placeholder={t('Leave empty for all models')}
            />
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

function KeysPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingToken, setEditingToken] = useState<Token | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['tokens', search],
    queryFn: () => getTokens(search ? { keyword: search } : undefined),
    staleTime: 30 * 1000,
  })

  const deleteMutation = useMutation({
    mutationFn: deleteToken,
    onSuccess: () => {
      toast.success(t('Token deleted'))
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
    },
  })

  const manageMutation = useMutation({
    mutationFn: ({ id, action }: { id: number; action: string }) => manageToken(id, action),
    onSuccess: () => {
      toast.success(t('Token status updated'))
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
    },
  })

  const tokens = data?.data || []

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key)
    toast.success(t('Copied to clipboard'))
  }

  return (
    <div className="p-4 sm:p-6 space-y-6  w-full min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('API Keys')}
          </h1>
          <p className="text-muted-foreground text-sm sm:text-base mt-2">
            {t('Manage your API access tokens')}
          </p>
        </div>
      </div>

      {/* Action Bar */}
      <div className="flex flex-col sm:flex-row gap-3 items-start">
        <div className="relative w-full sm:max-w-sm">
          <Search className="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="pl-12 py-6 text-lg"
            placeholder={t('Search tokens...')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Button
          onClick={() => {
            setEditingToken(null)
            setDialogOpen(true)
          }}
          className="gap-2"
        >
          <Plus className="h-4 w-4" />
          {t('Create Token')}
        </Button>
      </div>

      {/* Token List */}
      {!isLoading && tokens.length === 0 ? (
        <EmptyState
          icon={Key}
          title={t('No tokens found')}
          description={t('Create your first API token to get started')}
          action={{ label: t('Create Token'), onClick: () => { setEditingToken(null); setDialogOpen(true) } }}
        />
      ) : (
        <Card className="hover:shadow-lg transition-shadow">
          <CardContent className="p-0">
            <div className="rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-12">#</TableHead>
                    <TableHead>{t('Name')}</TableHead>
                    <TableHead>{t('Key')}</TableHead>
                    <TableHead>{t('Status')}</TableHead>
                    <TableHead>{t('Quota')}</TableHead>
                    <TableHead>{t('Created')}</TableHead>
                    <TableHead className="w-12" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {isLoading ? (
                    Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        {Array.from({ length: 7 }).map((_, j) => (
                          <TableCell key={j}><Skeleton className="h-4 w-24" /></TableCell>
                        ))}
                      </TableRow>
                    ))
                  ) : (
                    tokens.map((tk, idx) => (
                      <TableRow key={tk.id} className="hover:bg-muted/50 transition-colors">
                        <TableCell className="font-medium">{idx + 1}</TableCell>
                        <TableCell className="font-medium">
                          <div className="flex items-center gap-3">
                            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center">
                              <Key className="h-4 w-4 text-white" />
                            </div>
                            {tk.name}
                          </div>
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          <span className="opacity-60">{tk.key.slice(0, 12)}</span>
                          <span className="text-muted-foreground">...</span>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="inline-flex h-6 w-6 ml-2 hover:bg-accent"
                            onClick={() => copyKey(tk.key)}
                          >
                            <Copy className="h-3 w-3" />
                          </Button>
                        </TableCell>
                        <TableCell>
                          {tk.status === 1 ? (
                            <Badge className="bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400">
                              <CheckCircle className="mr-1 h-3 w-3" /> {t('Active')}
                            </Badge>
                          ) : (
                            <Badge variant="outline" className="bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400">
                              <XCircle className="mr-1 h-3 w-3" /> {t('Disabled')}
                            </Badge>
                          )}
                        </TableCell>
                        <TableCell>
                          {tk.unlimited_quota ? (
                            <Badge variant="secondary" className="text-sm">
                              {t('Unlimited')}
                            </Badge>
                          ) : (
                            <span className="text-sm">
                              {((tk.remaining_quota || 0) / 1000).toFixed(0)}K
                            </span>
                          )}
                        </TableCell>
                        <TableCell className="text-muted-foreground text-sm">
                          {dayjs(tk.created_time * 1000).format('YYYY-MM-DD')}
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
                                setEditingToken(tk)
                                setDialogOpen(true)
                              }}>
                                <Pencil className="mr-2 h-4 w-4" /> {t('Edit')}
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() =>
                                manageMutation.mutate({
                                  id: tk.id,
                                  action: tk.status === 1 ? 'disable' : 'enable',
                                })
                              }>
                                {tk.status === 1 ? (
                                  <XCircle className="mr-2 h-4 w-4" />
                                ) : (
                                  <CheckCircle className="mr-2 h-4 w-4" />
                                )}
                                {tk.status === 1 ? t('Disable') : t('Enable')}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => {
                                  if (confirm(t('Are you sure?'))) deleteMutation.mutate(tk.id)
                                }}
                                className="text-red-600"
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
            </div>
          </CardContent>
        </Card>
      )}

      <TokenFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        token={editingToken}
      />
    </div>
  )
}
