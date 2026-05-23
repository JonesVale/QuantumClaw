import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useMemo, useEffect, useRef } from 'react'
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
  Server,
  Zap,
  Network,
  ExternalLink,
  Wallet,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
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
  getChannelTypes,
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
  creatingNew,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  channel?: Channel | null
  creatingNew?: boolean
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!channel

  // 动态获取渠道类型（避免前端硬编码）
  const { data: typeMap } = useQuery({
    queryKey: ['channelTypes'],
    queryFn: getChannelTypes,
    staleTime: 10 * 60 * 1000,
  })

  // 按 AI(<100) / 量子(>=100) 分组
  const typeGroups = useMemo(() => {
    const map = typeMap || {}
    const ai: { id: number; name: string }[] = []
    const quantum: { id: number; name: string }[] = []
    Object.entries(map).forEach(([idStr, name]) => {
      const id = Number(idStr)
      if (id <= 0) return
      if (id >= 100) {
        quantum.push({ id, name })
      } else if (id < 100) {
        ai.push({ id, name })
      }
    })
    return { ai, quantum }
  }, [typeMap])
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
    cost_per_unit: 0,
    sell_price_rate: 1,
    thinking_to_content: false,
  })

  // 从 channel.config JSON 中解析 cache_billing_ratio、thinking_to_content 等字段
  function parseConfig(config?: string): { cache_billing_ratio?: number; thinking_to_content?: boolean } {
    if (!config) return {}
    try {
      const parsed = JSON.parse(config)
      return {
        cache_billing_ratio: typeof parsed.cache_billing_ratio === 'number' ? parsed.cache_billing_ratio : undefined,
        thinking_to_content: typeof parsed.thinking_to_content === 'boolean' ? parsed.thinking_to_content : undefined,
      }
    } catch {
      return {}
    }
  }

  // Populate form from channel data when editing
  const prevChannel = useRef(channel)
  useEffect(() => {
    if (channel && channel !== prevChannel.current) {
      const cfg = parseConfig(channel.config)
      setForm({
        type: channel.type || 1,
        key: '', // never populate key for security
        name: channel.name || '',
        base_url: channel.base_url || '',
        models: channel.models || '',
        group: channel.group || 'default',
        model_mapping: '',
        priority: 0,
        weight: channel.weight || 1,
        cache_billing_ratio: cfg.cache_billing_ratio ?? 0,
        cost_per_unit: channel.cost_per_unit || 0,
        sell_price_rate: channel.sell_price_rate || 1,
        thinking_to_content: cfg.thinking_to_content ?? false,
      })
    }
    if (!channel) {
      setForm({
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
        cost_per_unit: 0,
        sell_price_rate: 1,
        thinking_to_content: false,
      })
    }
    prevChannel.current = channel
  }, [channel])

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
      const payload: Record<string, unknown> = { id: channel.id }
      if (form.name) payload.name = form.name
      if (form.type) payload.type = form.type
      if (form.key) payload.key = form.key
      if (form.base_url !== undefined) payload.base_url = form.base_url
      if (form.models) payload.models = form.models
      if (form.group) payload.group = form.group
      payload.category = form.category || ''
      if (form.weight) payload.weight = form.weight
      if (form.cost_per_unit !== undefined) payload.cost_per_unit = form.cost_per_unit
      if (form.sell_price_rate !== undefined) payload.sell_price_rate = form.sell_price_rate
      if (form.cache_billing_ratio !== undefined) payload.cache_billing_ratio = form.cache_billing_ratio
      if (form.thinking_to_content !== undefined) payload.thinking_to_content = form.thinking_to_content
      updateMutation.mutate(payload as unknown as ChannelFormData & { id: number })
    } else {
      createMutation.mutate(form)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl">
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
              : t('Add a new AI or Quantum computing provider channel')}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-3">
              <Label>{t('Channel Name')}</Label>
              <Input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder={t('e.g. OpenAI Official')}
                required
              />
            </div>
            <div className="space-y-3">
              <Label>{t('Channel Type')}</Label>
              <Select
                value={String(form.type)}
                onValueChange={(v) => setForm({ ...form, type: Number(v) })}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t('Select type')} />
                </SelectTrigger>
                <SelectContent className="max-h-72">
                  {/* AI 大模型渠道 */}
                  {typeGroups.ai.length > 0 && (
                    <>
                      <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                        {t('AI Models')}
                      </div>
                      {typeGroups.ai.map((t) => (
                        <SelectItem key={t.id} value={String(t.id)}>
                          {t.name}
                        </SelectItem>
                      ))}
                    </>
                  )}
                  {/* 量子算力渠道 */}
                  {typeGroups.quantum.length > 0 && (
                    <>
                      <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider border-t pt-2 mt-1">
                        {t('Quantum Computing')}
                      </div>
                      {typeGroups.quantum.map((t) => (
                        <SelectItem key={t.id} value={String(t.id)}>
                          {t.name}
                        </SelectItem>
                      ))}
                    </>
                  )}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-3">
              <Label>{t('Category')}</Label>
              <Select value={form.category || '_none'} onValueChange={(v) => setForm({ ...form, category: v === '_none' ? '' : v })}>
                <SelectTrigger>
                  <SelectValue placeholder={t('Uncategorized')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="_none">{t('Uncategorized')}</SelectItem>
                  <SelectItem value="free" className="text-green-600 font-medium">&#9679; {t('Free')}</SelectItem>
                  <SelectItem value="paid" className="text-amber-600 font-medium">&#9679; {t('Paid')}</SelectItem>
                  <SelectItem value="custom" className="text-purple-600 font-medium">&#9679; {t('Custom')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-3">
            <Label>{t('API Key')}</Label>
            <Input
              type="password"
              value={form.key}
              onChange={(e) => setForm({ ...form, key: e.target.value })}
              placeholder={isEdit ? t('Leave empty to keep unchanged') : ''}
              required={!isEdit}
            />
          </div>

          <div className="space-y-3">
            <Label>{t('Base URL')}</Label>
            <Input
              value={form.base_url}
              onChange={(e) => setForm({ ...form, base_url: e.target.value })}
              placeholder="https://api.openai.com/v1"
            />
          </div>

          <div className="space-y-3">
            <Label>{t('Models')}</Label>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={async () => {
                try {
                  const id = isEdit && channel ? channel.id : 0;
                  if (!id) { toast.error(t('Save the channel first')); return; }
                  const res = await fetch('/api/channel/' + id + '/fetch-models');
                  const data = await res.json();
                  if (data.success && Array.isArray(data.data)) {
                    setForm({ ...form, models: data.data.join(', ') });
                  } else {
                    toast.error(data.message || t('Failed to fetch models'));
                  }
                } catch { toast.error(t('Failed to fetch models')); }
              }}
              className="ml-auto"
            >
              🔄 {t('Fetch Models')}
            </Button>
            <Input
              value={form.models}
              onChange={(e) => setForm({ ...form, models: e.target.value })}
              placeholder="gpt-4, gpt-3.5-turbo, ..."
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-3">
              <Label>{t('Group')}</Label>
              <Input
                value={form.group}
                onChange={(e) => setForm({ ...form, group: e.target.value })}
                placeholder="default"
              />
            </div>
            <div className="space-y-3">
              <Label>{t('Priority')}</Label>
              <Input
                type="number"
                value={form.priority}
                onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })}
                min="0"
              />
            </div>
          </div>

          <div className="space-y-3">
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

          <div className="grid gap-2">
            <Label htmlFor="cost_per_unit">{t('Cost Per Unit')}</Label>
            <Input
              id="cost_per_unit"
              type="number"
              min="0"
              step="0.0001"
              value={form.cost_per_unit ?? 0}
              onChange={(e) => setForm({ ...form, cost_per_unit: parseFloat(e.target.value) || 0 })}
              placeholder="0"
            />
            <p className="text-xs text-muted-foreground">{t('cost_per_unit_hint')}</p>
            {/* 实时价格预览 */}
            <div className="mt-1 rounded bg-muted p-2 text-xs space-y-1">
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t('Cost per 1K tokens')}</span>
                <span className="font-mono">${((form.cost_per_unit ?? 0) / 1000).toFixed(6)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t('Sell price per 1K tokens')}</span>
                <span className="font-mono font-bold">
                  ${((form.cost_per_unit ?? 0) * (form.sell_price_rate ?? 1) / 1000).toFixed(6)}
                </span>
              </div>
            </div>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="sell_price_rate">{t('Sell Price Rate')}</Label>
            <Input
              id="sell_price_rate"
              type="number"
              min="0"
              step="0.1"
              value={form.sell_price_rate ?? 1}
              onChange={(e) => setForm({ ...form, sell_price_rate: parseFloat(e.target.value) || 1 })}
              placeholder="1.0"
            />
            <p className="text-xs text-muted-foreground">{t('sell_price_rate_hint')}</p>
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
  const [typeCategory, setTypeCategory] = useState<string>('all')
  const [channelCategory, setChannelCategory] = useState<string>('all')
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null)
  const [creatingNew, setCreatingNew] = useState(false)
  const queryClient = useQueryClient()

  const { data: selfInfo } = useQuery({
    queryKey: ['self'],
    queryFn: async () => {
      const res = await fetch('/api/user/self')
      if (!res.ok) return null
      return res.json()
    },
    retry: false,
    staleTime: 30 * 1000,
  })

  const isProvider = selfInfo?.data?.user_type === 'provider' || (selfInfo?.data?.role ?? 0) >= 10

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['channels'],
    queryFn: () => getChannels(),
    staleTime: 30 * 1000,
  })

  // 动态渠道类型名称映射
  const { data: typeMap } = useQuery({
    queryKey: ['channelTypes'],
    queryFn: getChannelTypes,
    staleTime: 10 * 60 * 1000,
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
      const matchesCategory = typeCategory === 'all' ||
        (typeCategory === 'ai' && Number(ch.type) < 100) ||
        (typeCategory === 'quantum' && Number(ch.type) >= 100)
      const matchesChannelCat = channelCategory === 'all' ||
        (channelCategory === ch.category) ||
        (channelCategory === 'configured' && ch.key && !ch.key.startsWith('PUT_YOUR'))
      return matchesSearch && matchesStatus && matchesCategory && matchesChannelCat
    })
  }, [channels, search, status, typeCategory, channelCategory])

  const getStatusBadge = (status: number) => {
    if (status === 1) {
      return <Badge className="bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"><CheckCircle className="w-3 h-3 mr-1" /> {t('Enabled')}</Badge>
    }
    return <Badge variant="secondary" className="bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"><XCircle className="w-3 h-3 mr-1" /> {t('Disabled')}</Badge>
  }

  const getTypeBadge = (type: number) => {
    const typeName = typeMap?.[String(type)]
    const isQuantum = type >= 100
    return (
      <Badge variant="outline" className={isQuantum ? 'border-purple-300 text-purple-700 dark:border-purple-700 dark:text-purple-300' : ''}>
        {typeName || `Type ${type}`}
      </Badge>
    )
  }

  // 各渠道充值链接映射
  const getChannelTopupUrl = (type: number, baseUrl?: string): string | null => {
    // OpenAI Compatible -> 使用自定义 base_url
    if (type === 48 && baseUrl) return baseUrl.replace(/\/v1$/, '') + '/billing'
    if (type === 48) return null
    const urls: Record<number, string> = {
      1: 'https://platform.openai.com/settings/organization/billing/overview',
      3: 'https://portal.azure.com/#view/Microsoft_Azure_ProjectOxford/CognitiveServicesHub/~/overview',
      14: 'https://console.anthropic.com/settings/billing',
      15: 'https://console.bce.baidu.com/ai/#/ai/wenxin/price/ticket',
      16: 'https://open.bigmodel.cn/usercenter/my-wallet',
      17: 'https://dashscope.aliyuncs.com/topup',
      18: 'https://www.xfyun.cn/console/charge',
      19: 'https://ai.360.cn/platform/account/balance',
      23: 'https://console.cloud.tencent.com/hunyuan',
      25: 'https://ai.google.dev/pricing',
      26: 'https://platform.moonshot.cn/console/billing',
      27: 'https://platform.baichuan-ai.com/console/balance',
      28: 'https://platform.minimaxi.com/user/balance',
      29: 'https://console.mistral.ai/billing/',
      30: 'https://console.groq.com/settings/billing',
      31: 'https://platform.lingyiwanwu.com/console/payment',
      32: 'https://platform.stepfun.com/account/billing',
      35: 'https://platform.deepseek.com/top_up',
      36: 'https://dash.cloudflare.com/?to=/:account/ai',
      37: 'https://www.deepl.com/pro-api',
      38: 'https://api.together.xyz/settings/billing',
      39: 'https://console.volcengine.com/finance/',
      44: 'https://cloud.siliconflow.cn/account/billing',
      45: 'https://console.x.ai/billing',
      58: 'https://console.volcengine.com/ark/region:ark+cn-beijing/billing',
      // 量子算力
      100: 'https://cloud.ionq.co/account/billing',
      101: 'https://quantum-computing.ibm.com/account',
      102: 'https://dashboard.rigetti.com/billing',
      103: 'https://us-west-1.console.aws.amazon.com/braket/home?region=us-west-1',
      104: 'https://portal.azure.com/#view/Microsoft_Azure_Quantum/QuantumWorkspacesBlade',
      105: 'https://console.cloud.google.com/quantum',
    }
    return urls[type] || null
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
            {t('Manage AI and Quantum computing channels')}
          </p>
        </div>
        {isProvider ? (
          <Button onClick={() => { setEditingChannel(null); setCreatingNew(true); }} className="gap-2">
            <Plus className="h-4 w-4" />
            {t('Add Channel')}
          </Button>
        ) : (
          <Button variant="outline" onClick={() => window.location.href = '/wallet'} className="gap-2">
            <Wallet className="h-4 w-4" />
            {t('Become a Provider')}
          </Button>
        )}
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
            <Select value={typeCategory} onValueChange={setTypeCategory}>
              <SelectTrigger className="w-full sm:w-32">
                <SelectValue placeholder={t('All Types')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('All Types')}</SelectItem>
                <SelectItem value="ai">{t('AI Models')}</SelectItem>
                <SelectItem value="quantum">{t('Quantum')}</SelectItem>
              </SelectContent>
            </Select>
            <Select value={channelCategory} onValueChange={setChannelCategory}>
              <SelectTrigger className="w-full sm:w-36">
                <SelectValue placeholder={t('Category')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('All')}</SelectItem>
                <SelectItem value="free" className="text-green-600 font-medium">&#9679; {t('Free')}</SelectItem>
                <SelectItem value="paid" className="text-amber-600 font-medium">&#9679; {t('Paid')}</SelectItem>
                <SelectItem value="custom" className="text-purple-600 font-medium">&#9679; {t('Custom')}</SelectItem>
                <SelectItem value="configured" className="text-blue-600 font-medium">&#9679; {t('Has API Key')}</SelectItem>
              </SelectContent>
            </Select>
            <Select value={status} onValueChange={setStatus}>
              <SelectTrigger className="w-full sm:w-32">
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
          action={{ label: t('Add Channel'), onClick: () => { setEditingChannel(null); setCreatingNew(true); } }}
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
                    <TableHead>{t('Category')}</TableHead>
                    <TableHead>{t('Group')}</TableHead>
                    <TableHead>{t('Cost')}</TableHead>
                    <TableHead>{t('Price Rate')}</TableHead>
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
                        <TableCell>{getTypeBadge(channel.type)}</TableCell>
                        <TableCell>
                          {channel.category ? (
                            <Badge variant="outline" className={{
                              'free': 'border-green-200 text-green-700 bg-green-50',
                              'paid': 'border-amber-200 text-amber-700 bg-amber-50',
                              'custom': 'border-purple-200 text-purple-700 bg-purple-50',
                            }[channel.category] || ''}>
                              {channel.category}
                            </Badge>
                          ) : (
                            <span className="text-xs text-muted-foreground">—</span>
                          )}
                          {channel.key && !channel.key.startsWith('PUT_YOUR') && (
                            <span className="ml-1 h-2 w-2 inline-block rounded-full bg-green-500" title={t('API Key configured')} />
                          )}
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary">{channel.group}</Badge>
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {channel.cost_per_unit?.toFixed(4) || '0'}
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {channel.sell_price_rate?.toFixed(2) || '1.00'}×
                        </TableCell>
                        <TableCell>{getStatusBadge(channel.status)}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{channel.weight}</Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {dayjs(channel.created_time * 1000).format('YYYY-MM-DD')}
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
                              {(() => {
                                const url = getChannelTopupUrl(channel.type, channel.base_url)
                                if (!url) return null
                                return (
                                  <DropdownMenuItem onClick={() => window.open(url, '_blank')}>
                                    <ExternalLink className="mr-2 h-4 w-4" />
                                    {t('Top Up')}
                                  </DropdownMenuItem>
                                )
                              })()}
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
        open={creatingNew || !!editingChannel}
        onOpenChange={(open) => { if (!open) { setEditingChannel(null); setCreatingNew(false); } }}
        channel={editingChannel}
        creatingNew={creatingNew}
      />
    </div>
  )
}
