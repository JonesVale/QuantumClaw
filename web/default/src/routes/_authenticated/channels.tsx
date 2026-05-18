import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useCallback, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plus,
  Search,
  MoreHorizontal,
  Pencil,
  Trash2,
  Play,
  CheckCircle,
  XCircle,
  RefreshCw,
  Globe,
  Key,
  Server,
  Zap,
  Network,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { EmptyState } from '@/components/empty-state'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  type Channel,
  type ChannelFormData,
  getChannels,
  createChannel,
  updateChannel,
  deleteChannel,
  testChannel,
} from '@/lib/api-extended'
import { toast } from 'sonner'
import dayjs from '@/lib/dayjs'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/channels')({
  component: ChannelsPage,
})

function ChannelFormDialog({
  open,
  onOpenChange,
  channel,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  channel?: Channel | null
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!channel
  const [form, setForm] = useState<ChannelFormData>({
    type: 1,
    key: '',
    name: '',
    base_url: '',
    models: '',
    group: 'default',
    model_mapping: '',
    priority: 0,
    weight: 1,
    cache_billing_ratio: 0,
  })

  const createMutation = useMutation({
    mutationFn: createChannel,
    onSuccess: () => {
      toast.success(t('Channel created'))
      queryClient.invalidateQueries({ queryKey: ['channels'] })
      onOpenChange(false)
    },
    onError: () => toast.error(t('Failed to create channel')),
  })

  const updateMutation = useMutation({
    mutationFn: (data: ChannelFormData & { id: number }) => updateChannel(data),
    onSuccess: () => {
      toast.success(t('Channel updated'))
      queryClient.invalidateQueries({ queryKey: ['channels'] })
      onOpenChange(false)
    },
    onError: () => toast.error(t('Failed to update channel')),
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (isEdit && channel) {
      updateMutation.mutate({ id: channel.id, ...form })
    } else {
      createMutation.mutate(form)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {isEdit ? (
              <>
                <Pencil className="h-5 w-5" />
                {t('Edit Channel')}
              </>
            ) : (
              <>
                <Plus className="h-5 w-5" />
                {t('Create Channel')}
              </>
            )}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? t('Update channel configuration')
              : t('Add a new API provider channel')}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{t('Channel Name')}</Label>
              <Input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder={t('e.g. OpenAI Official')}
                required
              />
            </div>
            <div className="space-y-2">
              <Label>{t('Channel Type')}</Label>
              <Select
                value={String(form.type)}
                onValueChange={(v) => setForm({ ...form, type: Number(v) })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">OpenAI</SelectItem>
                  <SelectItem value="2">Azure OpenAI</SelectItem>
                  <SelectItem value="3">Google Gemini</SelectItem>
                  <SelectItem value="4">Anthropic</SelectItem>
                  <SelectItem value="5">DeepSeek</SelectItem>
                  <SelectItem value="11">AI Proxy</SelectItem>
                  <SelectItem value="12">API2GPT</SelectItem>
                  <SelectItem value="13">AIGC2D</SelectItem>
                  <SelectItem value="14">SiliconFlow</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label>{t('API Key')}</Label>
            <Input
              type="password"
              value={form.key}
              onChange={(e) => setForm({ ...form, key: e.target.value })}
              placeholder={isEdit ? t('Leave empty to keep unchanged') : ''}
              required={!isEdit}
            />
          </div>

          <div className="space-y-2">
            <Label>{t('Base URL')}</Label>
            <Input
              value={form.base_url}
              onChange={(e) => setForm({ ...form, base_url: e.target.value })}
              placeholder="https://api.openai.com/v1"
            />
          </div>

          <div className="space-y-2">
            <Label>{t('Models')}</Label>
            <Input
              value={form.models}
              onChange={(e) => setForm({ ...form, models: e.target.value })}
              placeholder="gpt-4, gpt-3.5-turbo, ..."
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{t('Group')}</Label>
              <Input
                value={form.group}
                onChange={(e) => setForm({ ...form, group: e.target.value })}
                placeholder="default"
              />
            </div>
            <div className="space-y-2">
              <Label>{t('Priority')}</Label>
              <Input
                type="number"
                value={form.priority}
                onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })}
                min="0"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label>{t('Weight')}</Label>
            <Input
              type="number"
              value={form.weight}
              onChange={(e) => setForm({ ...form, weight: Number(e.target.value) })}
              min="1"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="cache_billing_ratio">{t('cache_billing_ratio')}</Label>
            <Input
              id="cache_billing_ratio"
              type="number"
              min="0"
              max="1"
              step="0.05"
              value={form.cache_billing_ratio ?? ''}
              onChange={(e) => setForm({ ...form, cache_billing_ratio: parseFloat(e.target.value) || 0 })}
              placeholder="0.5"
            />
            <p className="text-xs text-muted-foreground">{t('cache_billing_ratio_hint')}</p>
          </div>

          <div className="grid gap-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="thinking_to_content">{t('thinking_to_content')}</Label>
              <Switch
                id="thinking_to_content"
                checked={form.thinking_to_content ?? false}
                onCheckedChange={(v) => setForm({ ...form, thinking_to_content: v })}
              />
            </div>
            <p className="text-xs text-muted-foreground">{t('thinking_to_content_hint')}</p>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('Cancel')}
            </Button>
            <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
              {createMutation.isPending || updateMutation.isPending
                ? t('Saving...')
                : isEdit
                ? t('Update')
                : t('Create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function ChannelsPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<string>('all')
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null)
  const queryClient = useQueryClient()

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['channels'],
    queryFn: () => getChannels(),
    staleTime: 30 * 1000,
  })

  const deleteMutation = useMutation({
    mutationFn: deleteChannel,
    onSuccess: () => {
      toast.success(t('Channel deleted'))
      queryClient.invalidateQueries({ queryKey: ['channels'] })
    },
    onError: () => toast.error(t('Failed to delete channel')),
  })

  const testMutation = useMutation({
    mutationFn: testChannel,
    onSuccess: (res) => {
      if (res?.data?.status === 'success') {
        toast.success(t('Channel test successful'))
      } else {
        toast.error(res?.data?.error || t('Channel test failed'))
      }
    },
    onError: () => toast.error(t('Failed to test channel')),
  })

  const channels: Channel[] = data?.data || []

  const filtered = useMemo(() => {
    return channels.filter((ch) => {
      const matchesSearch = !search || 
        ch.name.toLowerCase().includes(search.toLowerCase()) ||
        ch.group.toLowerCase().includes(search.toLowerCase())
      const matchesStatus = status === 'all' || 
        (status === 'enabled' && ch.status === 1) ||
        (status === 'disabled' && ch.status === 2)
      return matchesSearch && matchesStatus
    })
  }, [channels, search, status])

  const getStatusBadge = (status: number) => {
    if (status === 1) {
      return <Badge className="bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"><CheckCircle className="w-3 h-3 mr-1" /> {t('Enabled')}</Badge>
    }
    return <Badge variant="secondary" className="bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"><XCircle className="w-3 h-3 mr-1" /> {t('Disabled')}</Badge>
  }

  const getTypeBadge = (type: number) => {
    const types: Record<number, string> = {
      1: 'OpenAI',
      2: 'Azure',
      3: 'Gemini',
      4: 'Claude',
      5: 'DeepSeek',
      11: 'Proxy',
      12: 'API2GPT',
      13: 'AIGC2D',
      14: 'Silicon',
    }
    return <Badge variant="outline">{types[type] || `Type ${type}`}</Badge>
  }

  return (
    <div className="p-4 sm:p-6  w-full space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('Channels')}
          </h1>
          <p className="text-muted-foreground mt-1 sm:mt-2 text-sm sm:text-base lg:text-lg">
            {t('Manage AI service provider channels')}
          </p>
        </div>
        <Button onClick={() => setEditingChannel({} as Channel)} className="gap-2">
          <Plus className="h-4 w-4" />
          {t('Add Channel')}
        </Button>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="relative w-full sm:flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="pl-9"
                placeholder={t('Search channels...')}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <Select value={status} onValueChange={setStatus}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue placeholder={t('Status')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('All Status')}</SelectItem>
                <SelectItem value="enabled">{t('Enabled')}</SelectItem>
                <SelectItem value="disabled">{t('Disabled')}</SelectItem>
              </SelectContent>
            </Select>
            <Button variant="outline" onClick={() => refetch()} disabled={isFetching}>
              <RefreshCw className={cn('mr-2 h-4 w-4', isFetching && 'animate-spin')} />
              {t('Refresh')}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      {!isLoading && filtered.length === 0 ? (
        <EmptyState
          icon={Network}
          title={t('No channels found')}
          description={t('Add your first channel to get started')}
          action={{ label: t('Add Channel'), onClick: () => setEditingChannel({} as Channel) }}
        />
      ) : (
        <Card>
          <CardContent className="p-0">
            <div className="rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-12">{t('ID')}</TableHead>
                    <TableHead>{t('Name')}</TableHead>
                    <TableHead>{t('Type')}</TableHead>
                    <TableHead>{t('Group')}</TableHead>
                    <TableHead>{t('Status')}</TableHead>
                    <TableHead>{t('Weight')}</TableHead>
                    <TableHead>{t('Created')}</TableHead>
                    <TableHead className="text-right">{t('Actions')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {isLoading ? (
                    Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell><Skeleton className="h-4 w-8" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                        <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                      </TableRow>
                    ))
                  ) : (
                    filtered.map((channel) => (
                      <TableRow key={channel.id} className="hover:bg-muted/50 transition-colors">
                        <TableCell className="font-medium">{channel.id}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                              <Server className="h-4 w-4 text-white" />
                            </div>
                            <span className="font-medium">{channel.name}</span>
                          </div>
                        </TableCell>
                        <TableCell>{getTypeBadge(channel.type_)}</TableCell>
                        <TableCell>
                          <Badge variant="secondary">{channel.group}</Badge>
                        </TableCell>
                        <TableCell>{getStatusBadge(channel.status)}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{channel.weight}</Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {dayjs(channel.created_at * 1000).format('YYYY-MM-DD')}
                        </TableCell>
                        <TableCell className="text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8">
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => setEditingChannel(channel)}>
                                <Pencil className="mr-2 h-4 w-4" />
                                {t('Edit')}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => testMutation.mutate(channel.id)}
                                disabled={testMutation.isPending}
                              >
                                <Play className="mr-2 h-4 w-4" />
                                {t('Test')}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => {
                                  if (confirm(t('Are you sure?'))) {
                                    deleteMutation.mutate(channel.id)
                                  }
                                }}
                                className="text-red-600"
                              >
                                <Trash2 className="mr-2 h-4 w-4" />
                                {t('Delete')}
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

      <ChannelFormDialog
        open={!!editingChannel}
        onOpenChange={(open) => !open && setEditingChannel(null)}
        channel={editingChannel}
      />
    </div>
  )
}
