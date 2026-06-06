import { createFileRoute } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect, type ReactNode } from 'react'
import { Save, RefreshCw, Key, EyeOff, Eye, Check, X, Settings, DollarSign, ShieldAlert } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import apiClient from '@/lib/api'

export const Route = createFileRoute('/_authenticated/platform-settings')({
  component: PlatformSettingsPage,
})

interface ConfigMap {
  [key: string]: string
}

function PlatformSettingsPage() {
  const { t } = useT()
  const { auth } = useAuthStore()
  const queryClient = useQueryClient()
  const isRoot = (auth.user?.role ?? 0) >= 100
  const [form, setForm] = useState<ConfigMap>({})
  const [saving, setSaving] = useState(false)

  // Show/hide API key text
  const [showKeys, setShowKeys] = useState<Record<string, boolean>>({})

  const { data, isLoading } = useQuery({
    queryKey: ['platform-configs'],
    queryFn: async () => {
      const res = await apiClient.get('/api/platform/config')
      return res.data?.data as ConfigMap || {}
    },
    staleTime: 30_000,
  })

  useEffect(() => {
    if (data) {
      setForm(prev => {
        const merged = { ...prev }
        for (const [k, v] of Object.entries(data)) {
          merged[k] = v || ''
        }
        return merged
      })
    }
  }, [data])

  const freeChatProviders = [
    { key: 'groq_api_key', name: 'Groq', envKey: 'GROQ_API_KEY' },
    { key: 'deepseek_api_key', name: 'DeepSeek', envKey: 'DEEPSEEK_API_KEY' },
    { key: 'gemini_api_key', name: 'Google Gemini', envKey: 'GEMINI_API_KEY' },
    { key: 'siliconflow_api_key', name: 'SiliconFlow', envKey: 'SILICONFLOW_API_KEY' },
    { key: 'mistral_api_key', name: 'Mistral AI', envKey: 'MISTRAL_API_KEY' },
    { key: 'openrouter_api_key', name: 'OpenRouter', envKey: 'OPENROUTER_API_KEY' },
    { key: 'together_api_key', name: 'Together AI', envKey: 'TOGETHER_API_KEY' },
  ]

  const saveAll = async () => {
    setSaving(true)
    try {
      await apiClient.put('/api/platform/config', form)
      queryClient.invalidateQueries({ queryKey: ['platform-configs'] })
      toast.success(t('saved'))
    } catch {
      toast.error(t('save_failed'))
    } finally {
      setSaving(false)
    }
  }



  const promoFields = [
    { key: 'promotion_expire_days', label: t('promotion_expire_days'), hint: t('promotion_expire_days_hint') },
    { key: 'settlement_cycle',      label: t('settlement_cycle'),      hint: t('settlement_cycle_hint') },
    { key: 'min_withdraw_amount',   label: t('min_withdraw_amount'),   hint: t('min_withdraw_amount_hint') },
  ]

  return (
    <div className="qc-wrapper py-8 space-y-8">
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold mb-2">{t('platform_settings')}</h1>
          <p className="text-muted-foreground" style={{maxWidth: 'min(65ch, 100%)'}}>{t('platform_settings_desc')}</p>
        </div>
        <Button onClick={saveAll} disabled={saving || isLoading}>
          {saving ? <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> : <Save className="w-4 h-4 mr-2" />}
          {t('Save All')}
        </Button>
      </div>

      {/* ── Basic Platform Settings ── */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Settings className="w-4 h-4 text-amber-500" />
            {t('platform_basic')}
          </CardTitle>
          <CardDescription>{t('platform_basic_desc')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <Label>{t('server_address')}</Label>
            <Input
              value={form['server_address'] || ''}
              onChange={(e) => setForm(f => ({ ...f, server_address: e.target.value }))}
              placeholder="https://your-domain.com"
            />
            <p className="text-xs text-muted-foreground">{t('server_address_hint')}</p>
          </div>
          <div className="space-y-1.5">
            <Label>{t('system_name')}</Label>
            <Input
              value={form['system_name'] || ''}
              onChange={(e) => setForm(f => ({ ...f, system_name: e.target.value }))}
              placeholder="QuantumClaw"
            />
          </div>
          <div className="space-y-1.5">
            <Label>{t('footer_html')}</Label>
            <textarea
              className="flex min-h-[80px] w-full rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              value={form['footer_html'] || ''}
              onChange={(e) => setForm(f => ({ ...f, footer_html: e.target.value }))}
              placeholder={'<a href="/about">About</a> | &copy; 2026'}
            />
            <p className="text-xs text-muted-foreground">{t('footer_html_hint')}</p>
          </div>
          <div className="space-y-1.5">
            <Label>{t('logo_url')}</Label>
            <Input
              value={form['logo'] || ''}
              onChange={(e) => setForm(f => ({ ...f, logo: e.target.value }))}
              placeholder="/logo.webp"
            />
          </div>
          <div className="space-y-1.5">
            <Label>{t('top_up_link')}</Label>
            <Input
              value={form['top_up_link'] || ''}
              onChange={(e) => setForm(f => ({ ...f, top_up_link: e.target.value }))}
              placeholder="/topup"
            />
          </div>
          <div className="space-y-1.5">
            <Label>{t('icp_beian')}</Label>
            <Input
              value={form['icp_beian'] || ''}
              onChange={(e) => setForm(f => ({ ...f, icp_beian: e.target.value }))}
              placeholder={t('icp_beian_placeholder') || '粤ICP备2025xxxxxx号'}
            />
            <p className="text-xs text-muted-foreground">{t('icp_beian_hint')}</p>
          </div>
        </CardContent>
      </Card>

      {/* ── Free Chat API Keys ── */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base flex items-center gap-2">
              <Key className="w-4 h-4 text-amber-500" />
              {t('Free Chat API Keys')}
            </CardTitle>
            <CardDescription>{t('Configure API keys for the free chat feature. Keys stored in database take priority over environment variables.')}</CardDescription>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowKeys({})}
            className="text-xs text-muted-foreground"
          >
            <EyeOff className="w-3.5 h-3.5 mr-1" />
            {t('Hide All')}
          </Button>
        </CardHeader>
        <CardContent className="space-y-5">
          {freeChatProviders.map((p) => {
            const isVisible = showKeys[p.key]
            const hasValue = !!form[p.key]
            return (
              <div key={p.key}>
                <div className="flex items-center justify-between mb-1.5">
                  <Label className="font-medium flex items-center gap-2">
                    {p.name}
                    {hasValue
                      ? <Badge variant="outline" className="bg-green-50 text-green-600 border-green-200 text-[10px]">
                          <Check className="w-3 h-3 mr-0.5" /> {t('Configured')}
                        </Badge>
                      : <Badge variant="outline" className="bg-muted text-muted-foreground text-[10px]">
                          <X className="w-3 h-3 mr-0.5" /> {t('Not Configured')}
                        </Badge>
                    }
                  </Label>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2 text-xs"
                    onClick={() => setShowKeys(s => ({ ...s, [p.key]: !s[p.key] }))}
                  >
                    {isVisible ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                  </Button>
                </div>
                <div className="flex gap-2">
                  <Input
                    type={isVisible ? 'text' : 'password'}
                    value={form[p.key] || ''}
                    onChange={(e) => setForm(f => ({ ...f, [p.key]: e.target.value }))}
                    placeholder={`${p.envKey} ${t('or set via admin page')}`}
                    className="font-mono text-sm"
                  />
                </div>
                <p className="text-[11px] text-muted-foreground mt-1">
                  {t('Env variable')}: <code className="text-xs bg-muted px-1 py-0.5 rounded">{p.envKey}</code>
                  {' �?'}{t('DB value takes priority')}
                </p>
              </div>
            )
          })}
        </CardContent>
      </Card>

      {/* ── Registration Bonus Settings ── */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <DollarSign className="w-4 h-4 text-green-500" />
            {t('Registration Bonus')}
            {!isRoot && <span className="text-[10px] px-2 py-0.5 rounded-full bg-red-50 text-red-600 border border-red-200 flex items-center gap-1"><ShieldAlert className="w-3 h-3" />{t('Root Only')}</span>}
          </CardTitle>
          <CardDescription>{t('Control trial balance and quota given to new users. Set to 0 to disable.')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <RootField label={t('New User Trial Balance (cents)')} hint={t('Default: 0 (disabled). Only enable for key clients.')} isRoot={isRoot}>
            <Input type="number" min="0" step="100"
              value={form['new_user_trial_balance_cents'] || ''}
              onChange={(e) => setForm(f => ({ ...f, new_user_trial_balance_cents: e.target.value }))}
              placeholder="5000" disabled={!isRoot} />
          </RootField>
          <RootField label={t('New User Quota')} hint={t('Free quota for new users. Set to 0 to disable. Default: 50000')} isRoot={isRoot}>
            <Input type="number" min="0" step="1000"
              value={form['quota_for_new_user'] || ''}
              onChange={(e) => setForm(f => ({ ...f, quota_for_new_user: e.target.value }))}
              placeholder="50000" disabled={!isRoot} />
          </RootField>
        </CardContent>
      </Card>

      {/* ── Platform Fee Settings ── */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <DollarSign className="w-4 h-4 text-blue-500" />
            {t('Platform Fees')}
            {!isRoot && <span className="text-[10px] px-2 py-0.5 rounded-full bg-red-50 text-red-600 border border-red-200 flex items-center gap-1"><ShieldAlert className="w-3 h-3" />{t('Root Only')}</span>}
          </CardTitle>
          <CardDescription>{t('Configure transaction fees and channel partner onboarding fees')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <RootField label={t('Platform Fee Rate (%)')} hint={t('Percentage of monthly revenue charged as onboarding fee. Default: 5.0%')} isRoot={isRoot}>
            <Input type="number" min="0" step="0.1"
              value={form['platform_fee_rate_percent'] || ''}
              onChange={(e) => setForm(f => ({ ...f, platform_fee_rate_percent: e.target.value }))}
              placeholder="5.0" disabled={!isRoot} />
          </RootField>
          <RootField label={t('Min Revenue for Fee (cents)')} hint={t('Minimum monthly revenue before fee applies. Default: 100 (¥1)')} isRoot={isRoot}>
            <Input type="number" min="0" step="100"
              value={form['platform_fee_min_revenue_cents'] || ''}
              onChange={(e) => setForm(f => ({ ...f, platform_fee_min_revenue_cents: e.target.value }))}
              placeholder="100" disabled={!isRoot} />
          </RootField>
          <div className="border-t pt-4 mt-2">
            <p className="text-sm font-medium mb-3">{t('Transaction Fees')}</p>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <RootField label={t('Domestic Rate (%)')} hint={t('Default: 1.0')} isRoot={isRoot}>
                <Input type="number" min="0" step="0.1"
                  value={form['transaction_fee_domestic'] || ''}
                  onChange={(e) => setForm(f => ({ ...f, transaction_fee_domestic: e.target.value }))}
                  placeholder="1.0" disabled={!isRoot} />
              </RootField>
              <RootField label={t('Foreign Rate (%)')} hint={t('Default: 3.0')} isRoot={isRoot}>
                <Input type="number" min="0" step="0.1"
                  value={form['transaction_fee_foreign'] || ''}
                  onChange={(e) => setForm(f => ({ ...f, transaction_fee_foreign: e.target.value }))}
                  placeholder="3.0" disabled={!isRoot} />
              </RootField>
              <RootField label={t('Foreign Min ($)')} hint={t('Default: 5.00')} isRoot={isRoot}>
                <Input type="number" min="0" step="0.01"
                  value={form['transaction_fee_foreign_min'] || ''}
                  onChange={(e) => setForm(f => ({ ...f, transaction_fee_foreign_min: e.target.value }))}
                  placeholder="5.00" disabled={!isRoot} />
              </RootField>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* ── Promotion & Settlement Settings ── */}
      <Card className="bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="text-base">{t('promotion_settings')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {promoFields.map((field) => (
            <div key={field.key} className="space-y-1.5">
              <Label>{field.label}</Label>
              <Input
                value={form[field.key] || ''}
                onChange={(e) => setForm(f => ({ ...f, [field.key]: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground">{field.hint}</p>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}

// RootField 超级管理员专属字段包装器（非root显示锁定且不可编辑）
function RootField({ label, hint, isRoot, children }: { label: string; hint: string; isRoot: boolean; children: ReactNode }) {
  const { t } = useT()
  return (
    <div className="space-y-1.5">
      <Label className="flex items-center gap-2">
        {label}
        {!isRoot && <span className="text-[10px] px-1.5 py-0.5 rounded bg-red-50 text-red-500 border border-red-100 font-normal">{t('Locked')}</span>}
      </Label>
      {children}
      <p className="text-xs text-muted-foreground">{hint}</p>
    </div>
  )
}
