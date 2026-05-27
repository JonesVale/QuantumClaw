import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import {
  Key, Plus, Trash2, RefreshCw, AlertCircle, CheckCircle2, Globe, Shield,
  Play, Pause, ExternalLink, Info,
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/sub2api')({
  component: Sub2APIPage,
})

interface Credential {
  id: number
  user_id: number
  provider: string
  label: string
  status: number
  daily_cap: number
  used_today: number
  last_health_at: number
  last_error: string
  expires_at: number
  created_time: number
}

interface Provider {
  provider: string
  name: string
  price: string
  models: string[]
  limits: string
}

const providerColors: Record<string, string> = {
  chatgpt_plus: 'from-green-500 to-emerald-600',
  chatgpt_pro: 'from-purple-500 to-violet-600',
  chatgpt_team: 'from-blue-500 to-indigo-600',
  claude_pro: 'from-orange-500 to-amber-600',
  claude_team: 'from-rose-500 to-pink-600',
}

const providerIcons: Record<string, string> = {
  chatgpt_plus: '💚',
  chatgpt_pro: '💜',
  chatgpt_team: '💙',
  claude_pro: '🟠',
  claude_team: '💗',
}

function AddCredentialDialog({ providers, onSuccess }: { providers: Provider[], onSuccess: () => void }) {
  const { t } = useT()
  const [open, setOpen] = useState(false)
  const [provider, setProvider] = useState('')
  const [token, setToken] = useState('')
  const [label, setLabel] = useState('')
  const [dailyCap, setDailyCap] = useState('100')

  const mutation = useMutation({
    mutationFn: async () => {
      const res = await fetch('/api/sub2api/credentials', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ provider, token, label, daily_cap: parseInt(dailyCap) || 100 }),
      })
      const json = await res.json()
      if (!json.success) throw new Error(json.message)
      return json.data
    },
    onSuccess: () => {
      toast.success(t('Credential added'))
      setOpen(false)
      setProvider('')
      setToken('')
      setLabel('')
      onSuccess()
    },
    onError: (err: Error) => toast.error(err.message),
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className="gap-2">
          <Plus className="h-4 w-4" />
          {t('Add Subscription')}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('Add Subscription Credential')}</DialogTitle>
          <DialogDescription>
            {t('Enter your ChatGPT or Claude subscription session token to expose it via API.')}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label>{t('Provider')}</Label>
            <Select value={provider} onValueChange={setProvider}>
              <SelectTrigger>
                <SelectValue placeholder={t('Select provider')} />
              </SelectTrigger>
              <SelectContent>
                {providers?.map((p: Provider) => (
                  <SelectItem key={p.provider} value={p.provider}>
                    {p.name} ({p.price})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>{t('Label')}</Label>
            <Input value={label} onChange={e => setLabel(e.target.value)} placeholder={t('e.g. My Plus Account')} />
          </div>
          <div className="space-y-2">
            <Label>{t('Session Token')}</Label>
            <Input value={token} onChange={e => setToken(e.target.value)} type="password" placeholder={t('Session token / API key')} />
            <p className="text-xs text-muted-foreground">{t('Your token is encrypted at rest. We never store plaintext.')}</p>
          </div>
          <div className="space-y-2">
            <Label>{t('Daily Request Cap')}</Label>
            <Input value={dailyCap} onChange={e => setDailyCap(e.target.value)} type="number" min={1} placeholder="100" />
          </div>
          {provider && (
            <Card className="bg-muted/50">
              <CardContent className="py-3 space-y-1 text-sm">
                <p className="font-medium">{providers?.find(p => p.provider === provider)?.name}</p>
                <p className="text-xs text-muted-foreground">{t('Models')}: {providers?.find(p => p.provider === provider)?.models?.join(', ')}</p>
                <p className="text-xs text-muted-foreground">{t('Limits')}: {providers?.find(p => p.provider === provider)?.limits}</p>
              </CardContent>
            </Card>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>{t('Cancel')}</Button>
          <Button onClick={() => mutation.mutate()} disabled={!provider || !token || mutation.isPending}>
            {mutation.isPending ? t('Adding...') : t('Add')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function CredentialCard({ cred, onRefresh }: { cred: Credential, onRefresh: () => void }) {
  const { t } = useT()
  const queryClient = useQueryClient()
  const isActive = cred.status === 1

  const toggleMutation = useMutation({
    mutationFn: async () => {
      const res = await fetch(`/api/sub2api/credentials/${cred.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: isActive ? 2 : 1 }),
      })
      const json = await res.json()
      if (!json.success) throw new Error(json.message)
    },
    onSuccess: () => {
      toast.success(isActive ? t('Credential paused') : t('Credential activated'))
      onRefresh()
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const deleteMutation = useMutation({
    mutationFn: async () => {
      const res = await fetch(`/api/sub2api/credentials/${cred.id}`, { method: 'DELETE' })
      const json = await res.json()
      if (!json.success) throw new Error(json.message)
    },
    onSuccess: () => {
      toast.success(t('Credential deleted'))
      onRefresh()
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const providerName = cred.provider.replace('_', ' ').replace(/\b\w/g, c => c.toUpperCase())
  const icon = providerIcons[cred.provider] || '🔑'
  const date = new Date(cred.created_time).toLocaleDateString()

  return (
    <Card className={cn('relative overflow-hidden', !isActive && 'opacity-60')}>
      <div className={cn('absolute top-0 left-0 w-1 h-full', isActive ? 'bg-green-500' : 'bg-gray-400')} />
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="text-2xl">{icon}</div>
            <div>
              <CardTitle className="text-sm flex items-center gap-2">
                {cred.label || providerName}
                <Badge variant={isActive ? 'default' : 'secondary'} className="text-[10px]">
                  {isActive ? t('Active') : t('Paused')}
                </Badge>
              </CardTitle>
              <CardDescription className="text-xs">{providerName} · {t('Added')} {date}</CardDescription>
            </div>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-2">
        <div className="grid grid-cols-2 gap-2 text-sm">
          <div>
            <span className="text-muted-foreground block text-xs">{t('Daily Cap')}</span>
            <span className="font-medium">{cred.daily_cap > 0 ? cred.daily_cap.toLocaleString() : t('Unlimited')}</span>
          </div>
          <div>
            <span className="text-muted-foreground block text-xs">{t('Used Today')}</span>
            <span className="font-medium">{cred.used_today.toLocaleString()}</span>
          </div>
        </div>
        <div className="flex gap-2 pt-1">
          <Button size="sm" variant="outline" className="gap-1 text-xs" onClick={() => toggleMutation.mutate()}>
            {isActive ? <Pause className="h-3 w-3" /> : <Play className="h-3 w-3" />}
            {isActive ? t('Pause') : t('Activate')}
          </Button>
          <Button size="sm" variant="outline" className="gap-1 text-xs" onClick={() => fetch(`/api/sub2api/credentials/${cred.id}/test`).then(r => r.json()).then(j => {
            if (j.success) toast.success(t('Credential is valid'))
            else toast.error(j.message || t('Test failed'))
          })}>
            <RefreshCw className="h-3 w-3" />
            {t('Test')}
          </Button>
          <Button size="sm" variant="outline" className="gap-1 text-xs text-red-600 hover:text-red-700" onClick={() => {
            if (confirm(t('Delete this credential?'))) deleteMutation.mutate()
          }}>
            <Trash2 className="h-3 w-3" />
            {t('Delete')}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

function Sub2APIPage() {
  const { t } = useT()
  const queryClient = useQueryClient()

  const { data: providers, isLoading: provLoading } = useQuery({
    queryKey: ['sub2api-providers'],
    queryFn: async () => {
      const res = await fetch('/api/sub2api/providers')
      const json = await res.json()
      return json.data || []
    },
    staleTime: 300_000,
  })

  const { data: credentials, isLoading: credLoading, refetch } = useQuery({
    queryKey: ['sub2api-credentials'],
    queryFn: async () => {
      const res = await fetch('/api/sub2api/credentials')
      const json = await res.json()
      return json.data || []
    },
    refetchInterval: 30_000,
  })

  const loading = provLoading || credLoading

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Subscription to API')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>
          {t('Use your ChatGPT or Claude subscription as an API endpoint. Add your session token below.')}
        </p>
      </div>

      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <Key className="h-5 w-5 text-blue-500" />
          {t('My Credentials')}
        </h2>
        <div className="flex items-center gap-2">
          {providers && <AddCredentialDialog providers={providers} onSuccess={() => refetch()} />}
        </div>
      </div>

      {loading ? (
        <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
          {[1,2,3].map(i => <Skeleton key={i} className="h-40 rounded-xl" />)}
        </div>
      ) : credentials?.length === 0 ? (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center gap-3 py-12">
            <Key className="h-12 w-12 text-muted-foreground/40" />
            <p className="text-lg font-medium text-muted-foreground">{t('No subscription credentials yet')}</p>
            <p className="text-sm text-muted-foreground/60">{t('Add your ChatGPT Plus, Pro, or Claude Pro subscription to expose it via API.')}</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
          {credentials?.map((cred: Credential) => (
            <CredentialCard key={cred.id} cred={cred} onRefresh={() => refetch()} />
          ))}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-sm">
            <Info className="h-4 w-4 text-blue-500" />
            {t('How to get your session token')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 text-sm">
          <div className="space-y-2">
            <h4 className="font-medium">ChatGPT</h4>
            <ol className="list-decimal list-inside space-y-1 text-muted-foreground">
              <li>{t('Log in to chat.openai.com')}</li>
              <li>{t('Open DevTools (F12) → Application → Cookies')}</li>
              <li>{t('Copy the value of the __Secure-next-auth.session-token cookie')}</li>
              <li>{t('Paste it above as your session token')}</li>
            </ol>
          </div>
          <div className="space-y-2">
            <h4 className="font-medium">{t("Claude")}</h4>
            <ol className="list-decimal list-inside space-y-1 text-muted-foreground">
              <li>{t('Log in to claude.ai')}</li>
              <li>{t('Open DevTools (F12) → Application → Cookies')}</li>
              <li>{t('Copy the value of the sessionKey cookie')}</li>
              <li>{t('Paste it above as your session token')}</li>
            </ol>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
