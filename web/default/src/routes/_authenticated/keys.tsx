import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plus,
  Search,
  MoreHorizontal,
  Pencil,
  Trash2,
  Copy,
  Eye,
  EyeOff,
  CheckCircle,
  XCircle,
  Key,
  Globe,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
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
import {
  Tooltip, TooltipContent, TooltipTrigger,
} from '@/components/ui/tooltip'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import {
  type Token, type TokenFormData,
  getTokens, createToken, updateToken, deleteToken, manageToken,
  generateJWT,
} from '@/lib/api-extended'
import { toast } from 'sonner'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/_authenticated/keys')({
  component: KeysPage,
})

// ── Helpers ──

/** Mask a key for display: show first 4 + last 4 chars, hide the middle. */
function maskKey(key: string): string {
  if (!key || key.length < 10) return key || ''
  return `${key.slice(0, 4)}${'•'.repeat(Math.min(24, key.length - 8))}${key.slice(-4)}`
}

/** Human-readable quota: e.g. 1,234,567 or "∞" for unlimited. */
function formatQuota(value: number | undefined | null): string {
  if (value === undefined || value === null) return '0'
  return value.toLocaleString()
}

/** Progress percentage for quota usage bar. */
function quotaPercent(used: number | undefined | null, remaining: number | undefined | null): number {
  const total = (used || 0) + (remaining || 0)
  if (total === 0) return 0
  return Math.min(100, Math.round(((used || 0) / total) * 100))
}

// ── TokenFormDialog ──

function TokenFormDialog({
  open,
  onOpenChange,
  token,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  token?: Token | null
}) {
  const { t } = useT()
  const queryClient = useQueryClient()
  const isEdit = !!token
  const [newKey, setNewKey] = useState<string | null>(null)
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
      ? (data: TokenFormData & { id: number }) => updateToken(data as unknown as Partial<Token>)
      : ((data: TokenFormData & { id: number }) => createToken(data as unknown as Partial<Token>)) as unknown as (data: TokenFormData & { id: number }) => Promise<any>,
    onSuccess: (res) => {
      toast.success(isEdit ? t('Token updated') : t('Token created'))
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
      onOpenChange(false)
      if (!isEdit && (res as unknown as { data?: { key?: string } })?.data?.key) {
        setNewKey((res as unknown as { data: { key: string } }).data.key)
      }
    },
    onError: () => toast.error(isEdit ? t('Failed to update token') : t('Failed to create token')),
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (isEdit && token) {
      mutation.mutate({ id: token.id, ...form })
    } else {
      mutation.mutate({ id: 0, ...form })
    }
  }

  return (
    <>
      {/* Create/Edit Key form Dialog */}
      <Dialog open={open && !newKey} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-lg">
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

      {/* Show created Key Dialog */}
      <Dialog open={!!newKey} onOpenChange={(v) => { if (!v) setNewKey(null) }}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-green-600 dark:text-green-400">
              <CheckCircle className="h-5 w-5" />
              {t('API Key Created')}
            </DialogTitle>
            <DialogDescription>
              {t('Please save your API key, it will not be shown again')}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label className="text-sm font-semibold">{t('Your API Key')}</Label>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-3 py-2.5 rounded-lg bg-muted border text-sm font-mono break-all select-all">
                  {newKey}
                </code>
                <Button
                  variant="outline"
                  size="icon"
                  className="h-10 w-10 shrink-0"
                  onClick={() => {
                    navigator.clipboard.writeText(newKey || '')
                    toast.success(t('Copied to clipboard'))
                  }}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>

            <div className="space-y-2">
              <Label className="text-sm font-semibold">{t('Quick Start')}</Label>
              <pre className="px-3 py-2.5 rounded-lg bg-muted border text-xs font-mono overflow-x-auto whitespace-pre-wrap">
{`# AI Chat
curl ${window.location.origin}/v1/chat/completions \\
  -H "Authorization: Bearer ${newKey?.slice(0, 12)}..." \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}'

# Quantum Circuit
curl -X POST ${window.location.origin}/v1/quantum/run \\
  -H "Authorization: Bearer ${newKey?.slice(0, 12)}..." \\
  -H "Content-Type: application/json" \\
  -d '{"backend":"ionq_harmony","shots":1000,"circuit":{"qubits":2,"gates":[{"name":"h","targets":[0]}]}}'`}
              </pre>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setNewKey(null)}>
              {t('Done')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

// ── DeleteConfirmDialog ──

function DeleteConfirmDialog({
  open,
  onOpenChange,
  tokenName,
  onConfirm,
  pending,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  tokenName: string
  onConfirm: () => void
  pending: boolean
}) {
  const { t } = useT()
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-red-600">
            <Trash2 className="h-5 w-5" />
            {t('Delete API Key')}
          </DialogTitle>
          <DialogDescription>
            {t('Are you sure you want to delete')} "<strong>{tokenName}</strong>"? {t('This action cannot be undone.')}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('Cancel')}
          </Button>
          <Button variant="destructive" onClick={onConfirm} disabled={pending}>
            {pending ? t('Deleting...') : t('Delete')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── KeysPage (main) ──

function KeysPage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingToken, setEditingToken] = useState<Token | null>(null)

  // Track which rows have their full key revealed (by token id)
  const [revealedIds, setRevealedIds] = useState<Set<number>>(new Set())

  // Delete confirmation dialog state
  const [deleteTarget, setDeleteTarget] = useState<Token | null>(null)

  // JWT generation state
  const [jwtKey, setJwtKey] = useState<string | null>(null)
  const [jwtLoading, setJwtLoading] = useState(false)

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
      setDeleteTarget(null)
    },
    onError: () => toast.error(t('Failed to delete token')),
  })

  const manageMutation = useMutation({
    mutationFn: ({ id, action }: { id: number; action: string }) => manageToken(id, action),
    onSuccess: () => {
      toast.success(t('Token status updated'))
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
    },
    onError: () => toast.error(t('Failed to update token status')),
  })

  const tokens = data?.data || []

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key)
    toast.success(t('Copied to clipboard'))
  }

  const toggleReveal = (id: number) => {
    setRevealedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  return (
    <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6 min-h-screen bg-background">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight bg-gradient-to-r from-amber-400 to-orange-500 bg-clip-text text-transparent">
            {t('API Keys')}
          </h1>
          <p className="text-muted-foreground text-sm sm:text-base mt-2">
            {t('Manage your API access tokens')}
          </p>
        </div>
      </div>

      {/* Connection Info Card */}
      <Card className="bg-gradient-to-r from-amber-50 to-orange-50 dark:from-amber-950/30 dark:to-orange-950/30 border-amber-200 dark:border-amber-800">
        <CardContent className="p-4 sm:p-6 space-y-5">
          {/* Header */}
          <div className="flex items-start gap-3">
            <div className="hidden sm:flex w-10 h-10 rounded-full bg-amber-100 dark:bg-amber-900 items-center justify-center flex-shrink-0">
              <Globe className="h-5 w-5 text-[oklch(0.72_0.18_52)]" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-amber-800 dark:text-amber-300">
                {t('API Access')}
              </p>
              <p className="text-xs text-amber-600 dark:text-amber-500 mt-0.5">
                {t('All endpoints use the same API Key — no need to create separate tokens')}
              </p>
            </div>
          </div>

          {/* Base URL */}
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs font-medium text-amber-700 dark:text-amber-400">{t('Base URL')}:</span>
            <code className="px-3 py-1.5 rounded bg-amber-100 dark:bg-amber-900 text-amber-800 dark:text-amber-300 font-mono break-all text-sm">
              {window.location.origin}
            </code>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={() => {
                navigator.clipboard.writeText(window.location.origin)
                toast.success(t('Copied to clipboard'))
              }}
            >
              <Copy className="h-3.5 w-3.5" />
            </Button>
            <code className="px-3 py-1.5 rounded bg-amber-100 dark:bg-amber-900 text-amber-800 dark:text-amber-300 font-mono text-sm">
              Authorization: Bearer {'{'}sk-...{'}'}
            </code>
          </div>

          <hr className="border-amber-200 dark:border-amber-800" />

          {/* ── AI Models Section ── */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <div className="w-6 h-6 rounded-md bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center text-white text-[10px] font-bold shrink-0">
                AI
              </div>
              <span className="text-sm font-semibold text-amber-800 dark:text-amber-300">
                {t('AI Models')}
              </span>
            </div>
            <p className="text-xs text-amber-600 dark:text-amber-500 ml-8">
              {t('Call GPT-4, DeepSeek, Claude, Gemini, and all other AI models via OpenAI-compatible API')}
            </p>
            <div className="ml-8 space-y-1.5">
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-amber-500 uppercase tracking-wider shrink-0 w-14">{t('Endpoint')}</span>
                <code className="px-3 py-1.5 rounded bg-amber-100/70 dark:bg-amber-900/50 text-amber-800 dark:text-amber-300 font-mono text-sm">
                  POST {window.location.origin}/v1/chat/completions
                </code>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-xs font-semibold text-amber-500 uppercase tracking-wider shrink-0 w-14 pt-1">{t('Example')}</span>
                <div className="relative w-full group">
                  <pre className="bg-amber-950/10 dark:bg-amber-950/40 rounded-lg p-3 overflow-x-auto text-sm leading-relaxed font-mono text-amber-800 dark:text-amber-300">
{`curl {window.location.origin}/v1/chat/completions \\
  -H "Authorization: Bearer sk-..." \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}]}'`}
                </pre>
                  <Button
                    variant="outline"
                    size="sm"
                    className="absolute top-2 right-2 h-7 gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity bg-amber-50 dark:bg-amber-950 border-amber-300 dark:border-amber-700"
                    onClick={() => {
                      navigator.clipboard.writeText(`curl ${window.location.origin}/v1/chat/completions \\
  -H "Authorization: Bearer sk-..." \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}]}'`)
                      toast.success(t('Copied to clipboard'))
                    }}
                  >
                    <Copy className="h-3.5 w-3.5" />
                    <span className="text-xs">{t('Copy')}</span>
                  </Button>
                </div>
              </div>
            </div>
          </div>

          <hr className="border-amber-200 dark:border-amber-800" />

          {/* ── Quantum Computing Section ── */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <div className="w-6 h-6 rounded-md bg-gradient-to-br from-purple-400 to-purple-600 flex items-center justify-center text-white text-[10px] font-bold shrink-0">
                Q
              </div>
              <span className="text-sm font-semibold text-amber-800 dark:text-amber-300">
                {t('Quantum Computing')}
              </span>
            </div>
            <p className="text-xs text-amber-600 dark:text-amber-500 ml-8">
              {t('Submit quantum computing tasks to IonQ, IBM Q, Rigetti and other quantum backends')}
            </p>
            <div className="ml-8 space-y-1.5">
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-amber-500 uppercase tracking-wider shrink-0 w-14">{t('Endpoint')}</span>
                <code className="px-3 py-1.5 rounded bg-amber-100/70 dark:bg-amber-900/50 text-amber-800 dark:text-amber-300 font-mono text-sm">
                  POST {window.location.origin}/v1/quantum/tasks
                </code>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-xs font-semibold text-amber-500 uppercase tracking-wider shrink-0 w-14 pt-1">{t('Example')}</span>
                <div className="relative w-full group">
                  <pre className="bg-amber-950/10 dark:bg-amber-950/40 rounded-lg p-3 overflow-x-auto text-sm leading-relaxed font-mono text-amber-800 dark:text-amber-300">
{`curl -X POST {window.location.origin}/v1/quantum/tasks \\
  -H "Authorization: Bearer sk-..." \\
  -H "Content-Type: application/json" \\
  -d '{"backend":"ionq-simulator","circuit":{"qubits":2,"gates":[]}}'`}
                </pre>
                  <Button
                    variant="outline"
                    size="sm"
                    className="absolute top-2 right-2 h-7 gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity bg-amber-50 dark:bg-amber-950 border-amber-300 dark:border-amber-700"
                    onClick={() => {
                      navigator.clipboard.writeText(`curl -X POST ${window.location.origin}/v1/quantum/tasks \\
  -H "Authorization: Bearer sk-..." \\
  -H "Content-Type: application/json" \\
  -d '{"backend":"ionq-simulator","circuit":{"qubits":2,"gates":[]}}'`)
                      toast.success(t('Copied to clipboard'))
                    }}
                  >
                    <Copy className="h-3.5 w-3.5" />
                    <span className="text-xs">{t('Copy')}</span>
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Action Bar */}
      <div className="flex flex-col sm:flex-row gap-3 items-start">
        <div className="relative w-full sm:max-w-sm">
          <Input
            className="pr-12 py-6 text-lg"
            placeholder={t('Search tokens...')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <Search className="absolute right-4 top-1/2 h-5 w-5 -translate-y-1/2 text-muted-foreground pointer-events-none" />
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
        <Button
          variant="outline"
          className="gap-2"
          disabled={jwtLoading}
          onClick={async () => {
            setJwtLoading(true)
            try {
              const res = await generateJWT()
              if (res.success && res.data) {
                setJwtKey(res.data as unknown as string)
              } else {
                toast.error(res.message || t('Failed to generate JWT'))
              }
            } catch (e: any) {
              toast.error(e?.message || t('Failed to generate JWT'))
            } finally {
              setJwtLoading(false)
            }
          }}
        >
          {jwtLoading ? (
            <>{t('Generating...')}</>
          ) : (
            <>{t('Generate JWT')}</>
          )}
        </Button>
      </div>

      {/* Token List */}
      {!isLoading && tokens.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
          <Key className="h-12 w-12 mb-4 opacity-20" />
          <p className="text-lg font-medium">{t('No tokens found')}</p>
          <p className="text-sm mt-1 mb-6">{t('Create your first API token to get started')}</p>
          <Button
            onClick={() => {
              setEditingToken(null)
              setDialogOpen(true)
            }}
          >
            {t('Create Token')}
          </Button>
        </div>
      ) : (
        <Card>
          <CardContent className="p-0">
            <div className="rounded-lg border overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-12">#</TableHead>
                    <TableHead>{t('Name')}</TableHead>
                    <TableHead>{t('API Key')}</TableHead>
                    <TableHead>{t('Status')}</TableHead>
                    <TableHead>{t('Usage')}</TableHead>
                    <TableHead>{t('Created')}</TableHead>
                    <TableHead className="w-12" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {isLoading
                    ? Array.from({ length: 5 }).map((_, i) => (
                        <TableRow key={i}>
                          {Array.from({ length: 7 }).map((_, j) => (
                            <TableCell key={j}>
                              <Skeleton className="h-4 w-20" />
                            </TableCell>
                          ))}
                        </TableRow>
                      ))
                    : tokens.map((tk, idx) => {
                        const isRevealed = revealedIds.has(tk.id)
                        const used = tk.used_quota ?? 0
                        const remaining = tk.remaining_quota ?? tk.remain_quota ?? 0
                        const isUnlimited = tk.unlimited_quota

                        return (
                          <TableRow
                            key={tk.id}
                            className="hover:bg-muted/50 transition-colors"
                          >
                            <TableCell className="font-medium">{idx + 1}</TableCell>
                            <TableCell className="font-medium">
                              <div className="flex items-center gap-3">
                                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-amber-500 to-orange-600 flex items-center justify-center">
                                  <Key className="h-4 w-4 text-white" />
                                </div>
                                <span className="truncate max-w-[min(20vw,160px)]" title={tk.name}>
                                  {tk.name}
                                </span>
                              </div>
                            </TableCell>
                            <TableCell>
                              <div className="flex items-center gap-1.5">
                                <code className="font-mono text-xs sm:text-sm bg-muted/60 px-2 py-1 rounded">
                                  {isRevealed ? tk.key : maskKey(tk.key)}
                                </code>
                                <Tooltip>
                                  <TooltipTrigger asChild>
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      className="h-7 w-7 shrink-0"
                                      onClick={() => toggleReveal(tk.id)}
                                    >
                                      {isRevealed ? (
                                        <EyeOff className="h-3.5 w-3.5" />
                                      ) : (
                                        <Eye className="h-3.5 w-3.5" />
                                      )}
                                    </Button>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    {isRevealed ? t('Hide') : t('Show')}
                                  </TooltipContent>
                                </Tooltip>
                                <Tooltip>
                                  <TooltipTrigger asChild>
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      className="h-7 w-7 shrink-0"
                                      onClick={() => copyKey(tk.key)}
                                    >
                                      <Copy className="h-3.5 w-3.5" />
                                    </Button>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    {t('Copy')}
                                  </TooltipContent>
                                </Tooltip>
                              </div>
                            </TableCell>
                            <TableCell>
                              {tk.status === 1 ? (
                                <Badge className="bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 border-0">
                                  <CheckCircle className="mr-1 h-3 w-3" />
                                  {t('Active')}
                                </Badge>
                              ) : (
                                <Badge
                                  variant="outline"
                                  className="bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"
                                >
                                  <XCircle className="mr-1 h-3 w-3" />
                                  {t('Disabled')}
                                </Badge>
                              )}
                            </TableCell>
                            <TableCell>
                              <div className="flex flex-col gap-1 min-w-[max(80px,10vw)]">
                                <div className="flex items-center justify-between text-xs text-muted-foreground">
                                  <span>
                                    {isUnlimited
                                      ? t('Unlimited')
                                      : `${formatQuota(remaining)} ${t('remaining')}`}
                                  </span>
                                  {!isUnlimited && (
                                    <span>{formatQuota(used)} {t('used')}</span>
                                  )}
                                </div>
                                {!isUnlimited && (
                                  <div className="w-full h-1.5 bg-muted rounded-full overflow-hidden">
                                    <div
                                      className="h-full rounded-full transition-all duration-300"
                                      style={{
                                        width: `${quotaPercent(used, remaining)}%`,
                                        background:
                                          quotaPercent(used, remaining) > 80
                                            ? 'linear-gradient(90deg, #f59e0b, #ef4444)'
                                            : quotaPercent(used, remaining) > 50
                                              ? 'linear-gradient(90deg, #f97316, #ea580c)'
                                              : 'linear-gradient(90deg, #22c55e, #06b6d4)',
                                      }}
                                    />
                                  </div>
                                )}
                                {tk.request_count !== undefined && (
                                  <span className="text-[11px] text-muted-foreground">
                                    {tk.request_count.toLocaleString()} {t('requests')}
                                  </span>
                                )}
                              </div>
                            </TableCell>
                            <TableCell className="text-muted-foreground text-sm whitespace-nowrap">
                              {dayjs(tk.created_time * 1000).format('YYYY-MM-DD HH:mm')}
                            </TableCell>
                            <TableCell>
                              <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                  <Button variant="ghost" size="icon" className="h-8 w-8">
                                    <MoreHorizontal className="h-4 w-4" />
                                  </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end" className="min-w-[140px]">
                                  <DropdownMenuItem
                                    onClick={() => {
                                      setEditingToken(tk)
                                      setDialogOpen(true)
                                    }}
                                  >
                                    <Pencil className="mr-2 h-4 w-4" />
                                    {t('Edit')}
                                  </DropdownMenuItem>
                                  <DropdownMenuItem
                                    onClick={() =>
                                      manageMutation.mutate({
                                        id: tk.id,
                                        action: tk.status === 1 ? 'disable' : 'enable',
                                      })
                                    }
                                  >
                                    {tk.status === 1 ? (
                                      <XCircle className="mr-2 h-4 w-4" />
                                    ) : (
                                      <CheckCircle className="mr-2 h-4 w-4" />
                                    )}
                                    {tk.status === 1 ? t('Disable') : t('Enable')}
                                  </DropdownMenuItem>
                                  <Separator className="my-1" />
                                  <DropdownMenuItem
                                    onClick={() => setDeleteTarget(tk)}
                                    className="text-red-600 focus:text-red-600 focus:bg-red-50 dark:focus:bg-red-950/30"
                                  >
                                    <Trash2 className="mr-2 h-4 w-4" />
                                    {t('Delete')}
                                  </DropdownMenuItem>
                                </DropdownMenuContent>
                              </DropdownMenu>
                            </TableCell>
                          </TableRow>
                        )
                      })}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Create / Edit dialog */}
      <TokenFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        token={editingToken}
      />

      {/* Delete confirmation dialog */}
      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => {
          if (!v) setDeleteTarget(null)
        }}
        tokenName={deleteTarget?.name || ''}
        onConfirm={() => {
          if (deleteTarget) deleteMutation.mutate(deleteTarget.id)
        }}
        pending={deleteMutation.isPending}
      />

      {/* JWT Result Dialog */}
      <Dialog open={!!jwtKey} onOpenChange={(v) => { if (!v) setJwtKey(null) }}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-green-600 dark:text-green-400">
              <CheckCircle className="h-5 w-5" />
              {t('JWT Generated')}
            </DialogTitle>
            <DialogDescription>
              {t('Please save your JWT token, it will not be shown again')}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label className="text-sm font-semibold">{t('Your JWT Token')}</Label>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-3 py-2.5 rounded-lg bg-muted border text-sm font-mono break-all select-all">
                  {jwtKey}
                </code>
                <Button
                  variant="outline"
                  size="icon"
                  className="h-10 w-10 shrink-0"
                  onClick={() => {
                    navigator.clipboard.writeText(jwtKey || '')
                    toast.success(t('Copied to clipboard'))
                  }}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>

            <div className="space-y-2">
              <Label className="text-sm font-semibold">{t('Usage')}</Label>
              <pre className="px-3 py-2.5 rounded-lg bg-muted border text-xs font-mono overflow-x-auto whitespace-pre-wrap">
{`# Use JWT for API calls (24h expiry)
curl ${window.location.origin}/v1/chat/completions \\
  -H "Authorization: Bearer ${jwtKey?.slice(0, 20)}..." \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}'`}
              </pre>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setJwtKey(null)}>
              {t('Done')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
