import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import {
  Globe, Plus, Trash2, RefreshCw, CheckCircle2, AlertCircle, Edit3,
  Play, FileJson, Shield, Server, Activity,
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
// Textarea is rendered inline as a styled textarea element
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/sub2api-admin')({
  component: Sub2APIAdminPage,
})

interface Schema {
  id: number
  provider: string
  version: number
  label: string
  endpoint_url: string
  auth_type: string
  auth_key_name: string
  request_template: string
  headers_template: string
  response_path: string
  stream_mode: string
  model_mapping: string
  status: number
  priority: number
  is_builtin: boolean
  last_health_at: number
  last_error: string
  created_time: number
}

interface ProviderHealth {
  provider: string
  total: number
  active: number
  draft: number
  deprecated: number
  broken: number
  highest_version: number
  last_health_at: number
  last_error: string
}

const statusLabels: Record<number, string> = {
  1: 'Active',
  2: 'Draft',
  3: 'Deprecated',
  4: 'Broken',
}

const statusColors: Record<number, string> = {
  1: 'bg-green-500',
  2: 'bg-amber-500',
  3: 'bg-gray-400',
  4: 'bg-red-500',
}

const authLabels: Record<string, string> = {
  cookie: 'Cookie',
  bearer: 'Bearer Token',
  apikey: 'API Key',
  none: 'None',
}

const providerColors: Record<string, string> = {
  chatgpt: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  claude: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
  gemini: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  deepseek: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  grok: 'bg-rose-100 text-rose-800 dark:bg-rose-900/30 dark:text-rose-400',
}

function SchemaEditor({ schema, onClose, onSaved }: { schema?: Schema | null, onClose: () => void, onSaved: () => void }) {
  const { t } = useT()
  const isNew = !schema
  const [form, setForm] = useState<Partial<Schema>>(schema || {
    provider: '', version: 1, label: '', endpoint_url: '',
    auth_type: 'cookie', auth_key_name: '', request_template: '{}',
    headers_template: '{}', response_path: 'message.content',
    stream_mode: 'sse', model_mapping: '{}', status: 2, priority: 50,
  })

  const mutation = useMutation({
    mutationFn: async () => {
      const method = isNew ? 'POST' : 'PUT'
      const url = isNew ? '/api/admin/sub2api/schemas' : `/api/admin/sub2api/schemas/${schema!.id}`
      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      })
      const json = await res.json()
      if (!json.success) throw new Error(json.message)
      return json.data
    },
    onSuccess: () => {
      toast.success(isNew ? t('Schema created') : t('Schema updated'))
      onSaved()
      onClose()
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const testMutation = useMutation({
    mutationFn: async () => {
      if (!schema?.id) throw new Error('Save first')
      const res = await fetch(`/api/admin/sub2api/schemas/${schema.id}/test`)
      const json = await res.json()
      if (!json.success) throw new Error(json.message)
      return json.data
    },
    onSuccess: (data) => {
      toast.success(t('Schema valid'))
      if (data?.model_mapping_test) {
        console.log('Model mapping:', data.model_mapping_test)
      }
    },
    onError: (err: Error) => toast.error(err.message),
  })

  return (
    <DialogContent className="sm:max-w-3xl max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>{isNew ? t('New Schema') : `${t('Edit Schema')}: ${schema?.label || schema?.provider}`}</DialogTitle>
        <DialogDescription>{t('Define how to translate our API calls to the provider\'s web API')}</DialogDescription>
      </DialogHeader>
      <Tabs defaultValue="basic" className="py-4">
        <TabsList>
          <TabsTrigger value="basic">{t('Basic')}</TabsTrigger>
          <TabsTrigger value="auth">{t('Auth')}</TabsTrigger>
          <TabsTrigger value="template">{t('Template')}</TabsTrigger>
          <TabsTrigger value="mapping">{t('Model Mapping')}</TabsTrigger>
        </TabsList>

        <TabsContent value="basic" className="space-y-4 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{t('Provider')}</Label>
              <Input value={form.provider || ''} onChange={e => setForm({...form, provider: e.target.value})} placeholder="chatgpt" />
            </div>
            <div className="space-y-2">
              <Label>{t('Version')}</Label>
              <Input value={form.version || 1} onChange={e => setForm({...form, version: parseInt(e.target.value) || 1})} type="number" min={1} />
            </div>
          </div>
          <div className="space-y-2">
            <Label>{t('Label')}</Label>
            <Input value={form.label || ''} onChange={e => setForm({...form, label: e.target.value})} placeholder="ChatGPT v4 (2026-03)" />
          </div>
          <div className="space-y-2">
            <Label>{t('Endpoint URL')}</Label>
            <Input value={form.endpoint_url || ''} onChange={e => setForm({...form, endpoint_url: e.target.value})} placeholder="https://chatgpt.com/backend-api/conversation" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{t('Status')}</Label>
              <Select value={String(form.status || 2)} onValueChange={v => setForm({...form, status: parseInt(v)})}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="2">{t('Draft')}</SelectItem>
                  <SelectItem value="1">{t('Active')}</SelectItem>
                  <SelectItem value="3">{t('Deprecated')}</SelectItem>
                  <SelectItem value="4">{t('Broken')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>{t('Priority')} (higher=preferred)</Label>
              <Input value={form.priority || 50} onChange={e => setForm({...form, priority: parseInt(e.target.value) || 50})} type="number" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{t('Stream Mode')}</Label>
              <Select value={form.stream_mode || 'sse'} onValueChange={v => setForm({...form, stream_mode: v})}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="sse">SSE</SelectItem>
                  <SelectItem value="websocket">WebSocket</SelectItem>
                  <SelectItem value="poll">Poll</SelectItem>
                  <SelectItem value="none">None</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>{t('Response Path')}</Label>
              <Input value={form.response_path || ''} onChange={e => setForm({...form, response_path: e.target.value})} placeholder="choices.0.message.content" />
            </div>
          </div>
        </TabsContent>

        <TabsContent value="auth" className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>{t('Auth Type')}</Label>
            <Select value={form.auth_type || 'cookie'} onValueChange={v => setForm({...form, auth_type: v})}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="cookie">{t('Cookie')}</SelectItem>
                <SelectItem value="bearer">{t('Bearer Token')}</SelectItem>
                <SelectItem value="apikey">{t('API Key')}</SelectItem>
                <SelectItem value="none">{t('None')}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>{t('Auth Key Name')}</Label>
            <Input value={form.auth_key_name || ''} onChange={e => setForm({...form, auth_key_name: e.target.value})} placeholder="__Secure-next-auth.session-token" />
          </div>
          <div className="space-y-2">
            <Label>{t('Default Headers (JSON)')}</Label>
            <textarea value={form.headers_template || '{}'} onChange={e => setForm({...form, headers_template: e.target.value})} rows={5} className="w-full rounded-md border border-input bg-background px-3 py-2 text-xs font-mono ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" />
          </div>
        </TabsContent>

        <TabsContent value="template" className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>{t('Request Body Template')}</Label>
            <p className="text-xs text-muted-foreground">
              {t('Available placeholders')}: {{model}}, {{messages}}, {{role}}, {{content}}, {{stream}}, {{max_tokens}}, {{temperature}}, {{uuid}}
            </p>
            <textarea value={form.request_template || '{}'} onChange={e => setForm({...form, request_template: e.target.value})} rows={14} className="w-full rounded-md border border-input bg-background px-3 py-2 text-xs font-mono ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" />
          </div>
        </TabsContent>

        <TabsContent value="mapping" className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>{t('Model Mapping (JSON)')}</Label>
            <p className="text-xs text-muted-foreground">
              {t('Maps our API model names to provider internal IDs. Key=our name, Value=provider name. Use * as wildcard.')}
            </p>
            <textarea value={form.model_mapping || '{}'} onChange={e => setForm({...form, model_mapping: e.target.value})} rows={8} className="w-full rounded-md border border-input bg-background px-3 py-2 text-xs font-mono ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" />
          </div>
        </TabsContent>
      </Tabs>

      <DialogFooter className="gap-2">
        <Button variant="outline" onClick={testMutation.mutate} disabled={!schema?.id || testMutation.isPending}>
          <Play className="h-4 w-4 mr-1" />
          {t('Test')}
        </Button>
        <Button variant="outline" onClick={onClose}>{t('Cancel')}</Button>
        <Button onClick={() => mutation.mutate()} disabled={mutation.isPending}>
          {isNew ? t('Create') : t('Save')}
        </Button>
      </DialogFooter>
    </DialogContent>
  )
}

function SchemaCard({ schema, onEdit, onRefresh }: { schema: Schema, onEdit: (s: Schema) => void, onRefresh: () => void }) {
  const { t } = useT()
  const healthDate = schema.last_health_at > 0 ? new Date(schema.last_health_at).toLocaleString() : '-'
  const isHealthy = schema.last_health_at > 0 && !schema.last_error
  const status = statusLabels[schema.status] || 'Unknown'
  const providerClass = providerColors[schema.provider] || 'bg-slate-100'

  const deleteMutation = useMutation({
    mutationFn: async () => {
      const res = await fetch(`/api/admin/sub2api/schemas/${schema.id}`, { method: 'DELETE' })
      const json = await res.json()
      if (!json.success) throw new Error(json.message)
    },
    onSuccess: () => {
      toast.success(t('Schema deleted'))
      onRefresh()
    },
    onError: (err: Error) => toast.error(err.message),
  })

  return (
    <Card className="relative hover:shadow-md transition-all">
      <div className={cn('absolute top-0 left-0 w-full h-1', statusColors[schema.status] || 'bg-gray-400')} />
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <Globe className="h-5 w-5 text-blue-500 shrink-0" />
            <div>
              <CardTitle className="text-sm flex items-center gap-2">
                <Badge variant="outline" className={cn('text-[10px] font-mono', providerClass)}>{schema.provider}</Badge>
                v{schema.version}
                {schema.is_builtin && <Badge variant="secondary" className="text-[10px]">{t('Built-in')}</Badge>}
              </CardTitle>
              <CardDescription className="text-xs mt-1">{schema.label || schema.provider}</CardDescription>
            </div>
          </div>
          <Badge variant="outline" className={cn('text-[10px]', schema.status === 1 ? 'bg-green-50 text-green-700 border-green-200' : '')}>{status}</Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-2 text-xs">
        <div className="flex items-center gap-2 text-muted-foreground">
          <Server className="h-3 w-3 shrink-0" />
          <span className="truncate">{schema.endpoint_url}</span>
        </div>
        <div className="flex items-center gap-2 text-muted-foreground">
          <Shield className="h-3 w-3 shrink-0" />
          <span>{authLabels[schema.auth_type] || schema.auth_type} / {schema.stream_mode}</span>
        </div>
        <div className="flex items-center gap-2">
          <Activity className="h-3 w-3 shrink-0" />
          {isHealthy ? (
            <span className="flex items-center gap-1 text-green-600"><CheckCircle2 className="h-3 w-3" /> {t('Healthy')}</span>
          ) : schema.last_error ? (
            <span className="flex items-center gap-1 text-red-600"><AlertCircle className="h-3 w-3" /> {t('Error')}</span>
          ) : (
            <span className="text-muted-foreground">{t('Not checked')}</span>
          )}
          <span className="text-muted-foreground/50">{healthDate}</span>
        </div>
        <div className="flex gap-1 pt-2">
          <Button size="sm" variant="outline" className="h-7 text-xs gap-1" onClick={() => onEdit(schema)}>
            <Edit3 className="h-3 w-3" /> {t('Edit')}
          </Button>
          {!schema.is_builtin && (
            <Button size="sm" variant="outline" className="h-7 text-xs gap-1 text-red-600 hover:text-red-700"
              onClick={() => { if (confirm(t('Delete this schema?'))) deleteMutation.mutate() }}>
              <Trash2 className="h-3 w-3" /> {t('Delete')}
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

function Sub2APIAdminPage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [editingSchema, setEditingSchema] = useState<Schema | null>(null)
  const [showNewDialog, setShowNewDialog] = useState(false)

  const { data: schemas, isLoading, refetch } = useQuery({
    queryKey: ['admin-sub2api-schemas'],
    queryFn: async () => {
      const res = await fetch('/api/admin/sub2api/schemas')
      const json = await res.json()
      return json.data || []
    },
    refetchInterval: 30_000,
  })

  const { data: health } = useQuery({
    queryKey: ['admin-sub2api-health'],
    queryFn: async () => {
      const res = await fetch('/api/admin/sub2api/health')
      const json = await res.json()
      return json.data || []
    },
    refetchInterval: 15_000,
  })

  const triggerCheck = useMutation({
    mutationFn: async () => {
      const res = await fetch('/api/admin/sub2api/health/check', { method: 'POST' })
      const json = await res.json()
      if (!json.success) throw new Error(json.message)
      return json.data
    },
    onSuccess: (data) => {
      toast.success(t('Health check triggered: ' + data?.schemas_queued + ' schemas'))
      setTimeout(() => refetch(), 5000)
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const providers = schemas ? [...new Set(schemas.map((s: Schema) => s.provider))].sort() : []

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Sub2API Schema Management')}</h1>
        <p className="text-muted-foreground mb-4">{t('Manage provider request templates, auth, and health checking')}</p>
      </div>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Globe className="h-5 w-5 text-indigo-500" />
          <h2 className="text-lg font-semibold">{t('Schemas')} ({schemas?.length || 0})</h2>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" className="gap-1" onClick={() => triggerCheck.mutate()}>
            <RefreshCw className={cn('h-4 w-4', triggerCheck.isPending && 'animate-spin')} />
            {t('Health Check')}
          </Button>
          <Dialog open={showNewDialog} onOpenChange={setShowNewDialog}>
            <DialogTrigger asChild>
              <Button size="sm" className="gap-1"><Plus className="h-4 w-4" /> {t('New Schema')}</Button>
            </DialogTrigger>
            {showNewDialog && (
              <SchemaEditor onClose={() => setShowNewDialog(false)} onSaved={() => refetch()} />
            )}
          </Dialog>
        </div>
      </div>

      {/* Provider health summary */}
      {health && health.length > 0 && (
        <div className="grid gap-3 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5">
          {health.map((h: ProviderHealth) => (
            <Card key={h.provider} className="p-3">
              <div className="flex items-center justify-between mb-2">
                <Badge variant="outline" className={cn('text-xs font-mono', providerColors[h.provider])}>{h.provider}</Badge>
                <Badge variant={h.active > 0 ? 'default' : 'secondary'} className="text-[10px]">{h.active}/{h.total}</Badge>
              </div>
              <div className="flex gap-2 text-xs text-muted-foreground">
                <span className={h.broken > 0 ? 'text-red-500' : ''}>{t('Broken')}: {h.broken}</span>
                <span>{t('Draft')}: {h.draft}</span>
              </div>
              {h.last_error && <p className="text-[10px] text-red-500 mt-1 truncate">{h.last_error}</p>}
            </Card>
          ))}
        </div>
      )}

      {isLoading ? (
        <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
          {[1,2,3].map(i => <Skeleton key={i} className="h-44 rounded-xl" />)}
        </div>
      ) : (
        providers.map(provider => (
          <div key={provider} className="space-y-3">
            <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">{provider}</h3>
            <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
              {schemas?.filter((s: Schema) => s.provider === provider).map((s: Schema) => (
                <SchemaCard key={s.id} schema={s} onEdit={(schema) => setEditingSchema(schema)} onRefresh={() => refetch()} />
              ))}
            </div>
          </div>
        ))
      )}

      {/* Edit dialog */}
      <Dialog open={!!editingSchema} onOpenChange={(o) => { if (!o) setEditingSchema(null) }}>
        {editingSchema && (
          <SchemaEditor schema={editingSchema} onClose={() => setEditingSchema(null)} onSaved={() => refetch()} />
        )}
      </Dialog>
    </div>
  )
}
