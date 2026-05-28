import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useAuthStore } from '@/stores/auth-store'
import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Save, Settings, DollarSign, Gauge, UserPlus, Mail, Lock, Bell, Zap, Shield, Server, Smartphone, Copy, Clock, Monitor, Globe, QrCode, Key, Plus, Trash2, Fingerprint } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import apiClient from '@/lib/api'
import { getOptions, updateOptions, updateOption, type SystemOption } from '@/lib/api-extended'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  const { t } = useT()
  const { auth } = useAuthStore();
  const isAdmin = auth.user?.role >= 10;
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

  const { data: optionsData, isLoading } = useQuery({
    queryKey: ['system-options'],
    queryFn: getOptions,
    staleTime: 30 * 1000,
  })

  const options: SystemOption[] = optionsData?.data || []
  const getOptionValue = (key: string, fallback: string = '') =>
    options.find((o) => o.key === key)?.value || fallback

  const [systemName, setSystemName] = useState('')
  const [logoUrl, setLogoUrl] = useState('')
  const [footerHtml, setFooterHtml] = useState('')
  const [defaultCurrency, setDefaultCurrency] = useState('')
  const [stripePublicKey, setStripePublicKey] = useState('')
  const [stripeApiSecret, setStripeApiSecret] = useState('')
  const [stripeEnabled, setStripeEnabled] = useState(false)
  const [stripeMinTopUp, setStripeMinTopUp] = useState('')
  const [epayEnabled, setEpayEnabled] = useState(false)
  const [epayId, setEpayId] = useState('')
  const [epayKey, setEpayKey] = useState('')
  const [epayAddress, setEpayAddress] = useState('')
  const [creemEnabled, setCreemEnabled] = useState(false)
  const [creemApiKey, setCreemApiKey] = useState('')
  const [waffoEnabled, setWaffoEnabled] = useState(false)
  const [waffoApiKey, setWaffoApiKey] = useState('')
  const [waffoSandbox, setWaffoSandbox] = useState(false)
  const [binanceEnabled, setBinanceEnabled] = useState(false)
  const [binanceApiKey, setBinanceApiKey] = useState('')
  const [binanceSecretKey, setBinanceSecretKey] = useState('')
  const [binanceMerchantId, setBinanceMerchantId] = useState('')
  const [minTopUp, setMinTopUp] = useState('')
  const [cacheBillingRatio, setCacheBillingRatio] = useState('')
  const [commissionEnabled, setCommissionEnabled] = useState(false)
  const [commissionRate, setCommissionRate] = useState('')
  const [registerReward, setRegisterReward] = useState('')
  const [minWithdraw, setMinWithdraw] = useState('')
  const [globalApiRateLimit, setGlobalApiRateLimit] = useState('')
  const [globalWebRateLimit, setGlobalWebRateLimit] = useState('')
  const [ipRateLimitInput, setIpRateLimitInput] = useState('')
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('')
  const [turnstileSecretKey, setTurnstileSecretKey] = useState('')
  const [registerEnabled, setRegisterEnabled] = useState(true)
  const [emailVerification, setEmailVerification] = useState(false)
  const [turnstileCheck, setTurnstileCheck] = useState(false)
  const [defaultUserGroup, setDefaultUserGroup] = useState('')
  const [allowedEmails, setAllowedEmails] = useState('')
  const [smtpHost, setSmtpHost] = useState('')
  const [smtpPort, setSmtpPort] = useState('')
  const [smtpUsername, setSmtpUsername] = useState('')
  const [smtpPassword, setSmtpPassword] = useState('')
  const [smtpFrom, setSmtpFrom] = useState('')
  const [githubClientId, setGithubClientId] = useState('')
  const [githubClientSecret, setGithubClientSecret] = useState('')
  const [googleClientId, setGoogleClientId] = useState('')
  const [googleClientSecret, setGoogleClientSecret] = useState('')
  const [msClientId, setMsClientId] = useState('')
  const [msClientSecret, setMsClientSecret] = useState('')
  const [oidcProviderUrl, setOidcProviderUrl] = useState('')
  const [oidcClientId, setOidcClientId] = useState('')
  const [oidcClientSecret, setOidcClientSecret] = useState('')
  const [larkAppId, setLarkAppId] = useState('')
  const [larkAppSecret, setLarkAppSecret] = useState('')
  const [wechatAppId, setWechatAppId] = useState('')
  const [wechatAppSecret, setWechatAppSecret] = useState('')
  const [discordClientId, setDiscordClientId] = useState('')
  const [discordClientSecret, setDiscordClientSecret] = useState('')
  const [emailNotification, setEmailNotification] = useState(false)
  const [webhookUrl, setWebhookUrl] = useState('')
  const [notifyCheckin, setNotifyCheckin] = useState(false)
  const [notifyTopup, setNotifyTopup] = useState(false)
  const [notifyUsageThreshold, setNotifyUsageThreshold] = useState(false)
  const [memoryCacheEnabled, setMemoryCacheEnabled] = useState(false)
  const [cacheSyncFrequency, setCacheSyncFrequency] = useState('')
  const [enableMetric, setEnableMetric] = useState(false)
  const [channelTestFrequency, setChannelTestFrequency] = useState('')
  const [batchTestCount, setBatchTestCount] = useState('')
  const [companyName, setCompanyName] = useState('')
  const [companyTaxId, setCompanyTaxId] = useState('')
  const [companyAddress, setCompanyAddress] = useState('')
  const [companyPhone, setCompanyPhone] = useState('')
  const [companyBank, setCompanyBank] = useState('')
  const [companyBankAccount, setCompanyBankAccount] = useState('')
  const [companyAlipayQr, setCompanyAlipayQr] = useState('')
  const [minPasswordLength, setMinPasswordLength] = useState('')
  const [requireSpecialChars, setRequireSpecialChars] = useState(false)
  const [sessionTimeout, setSessionTimeout] = useState('')
  const [enforce2fa, setEnforce2fa] = useState(false)
  const [ipWhitelist, setIpWhitelist] = useState('')
  const [ipBlacklist, setIpBlacklist] = useState('')
  const [ssrfProtection, setSsrfProtection] = useState(false)

  // WebAuthn credentials
  const [webauthnCredentials, setWebauthnCredentials] = useState<any[]>([])
  const [registeringWebAuthn, setRegisteringWebAuthn] = useState(false)

  useQuery({
    queryKey: ['settings-webauthn-credentials'],
    queryFn: async () => {
      try {
        const res = await fetch('/api/user/self/webauthn/credentials')
        const data = await res.json()
        if (data.success) {
          setWebauthnCredentials(data.data || [])
        }
      } catch {}
      return []
    },
    staleTime: 10 * 1000,
  })

  // 2FA states
  const [twoFAStatus, setTwoFAStatus] = useState(false)
  const [twoFALoading, setTwoFALoading] = useState(false)
  const [enabling2FA, setEnabling2FA] = useState(false)
  const [disabling2FA, setDisabling2FA] = useState(false)
  const [qrCodeUrl, setQrCodeUrl] = useState('')
  const [totpSecret, setTotpSecret] = useState('')
  const [verifyCode, setVerifyCode] = useState('')
  const [disableCode, setDisableCode] = useState('')

  // Security activity
  const [securityActivity, setSecurityActivity] = useState<any[]>([])
  const [activityLoading, setActivityLoading] = useState(false)

  useEffect(() => {
    if (!optionsData?.data) return
    setSystemName(getOptionValue('SystemName', 'QuantumClaw'))
    setLogoUrl(getOptionValue('Logo', '/logo.png'))
    setFooterHtml(getOptionValue('Footer', ''))
    setDefaultCurrency(getOptionValue('DefaultCurrency', 'USD'))
    setStripePublicKey(getOptionValue('StripePublicKey', ''))
    setStripeApiSecret(getOptionValue('StripeApiSecret', ''))
    setStripeEnabled(getOptionValue('StripeEnabled', 'false') === 'true')
    setStripeMinTopUp(getOptionValue('StripeMinTopUp', '1'))
    setEpayEnabled(getOptionValue('EpayEnabled', 'false') === 'true')
    setEpayId(getOptionValue('EpayId', ''))
    setEpayKey(getOptionValue('EpayKey', ''))
    setEpayAddress(getOptionValue('EpayAddress', ''))
    setCreemEnabled(getOptionValue('CreemEnabled', 'false') === 'true')
    setCreemApiKey(getOptionValue('CreemApiKey', ''))
    setWaffoEnabled(getOptionValue('WaffoEnabled', 'false') === 'true')
    setWaffoApiKey(getOptionValue('WaffoApiKey', ''))
    setWaffoSandbox(getOptionValue('WaffoSandbox', 'false') === 'true')
    setBinanceEnabled(getOptionValue('BinanceEnabled', 'false') === 'true')
    setBinanceApiKey(getOptionValue('BinanceApiKey', ''))
    setBinanceSecretKey(getOptionValue('BinanceSecretKey', ''))
    setBinanceMerchantId(getOptionValue('BinanceMerchantId', ''))
    setMinTopUp(getOptionValue('MinTopUp', '1'))
    setCacheBillingRatio(getOptionValue('CacheBillingRatio', '1.0'))
    setCommissionEnabled(getOptionValue('CommissionEnabled', 'false') === 'true')
    setCommissionRate(getOptionValue('CommissionRate', '0.1'))
    setRegisterReward(getOptionValue('RegisterReward', '0'))
    setMinWithdraw(getOptionValue('MinWithdraw', '10000'))
    setGlobalApiRateLimit(getOptionValue('GlobalApiRateLimitNum', '480'))
    setGlobalWebRateLimit(getOptionValue('GlobalWebRateLimitNum', '240'))
    setIpRateLimitInput(getOptionValue('CriticalRateLimitNum', '20'))
    setTurnstileSiteKey(getOptionValue('TurnstileSiteKey', ''))
    setTurnstileSecretKey(getOptionValue('TurnstileSecretKey', ''))
    setRegisterEnabled(getOptionValue('RegisterEnabled', 'true') === 'true')
    setEmailVerification(getOptionValue('EmailVerificationEnabled', 'false') === 'true')
    setTurnstileCheck(getOptionValue('TurnstileCheckEnabled', 'false') === 'true')
    setDefaultUserGroup(getOptionValue('DefaultUserGroup', ''))
    setAllowedEmails(getOptionValue('EmailDomainWhitelist', ''))
    setSmtpHost(getOptionValue('SMTPServer', ''))
    setSmtpPort(getOptionValue('SMTPPort', '587'))
    setSmtpUsername(getOptionValue('SMTPAccount', ''))
    setSmtpPassword(getOptionValue('SMTPToken', ''))
    setSmtpFrom(getOptionValue('SMTPFrom', ''))
    setGithubClientId(getOptionValue('GitHubClientId', ''))
    setGithubClientSecret(getOptionValue('GitHubClientSecret', ''))
    setGoogleClientId(getOptionValue('GoogleClientId', ''))
    setGoogleClientSecret(getOptionValue('GoogleClientSecret', ''))
    setMsClientId(getOptionValue('MicrosoftClientId', ''))
    setMsClientSecret(getOptionValue('MicrosoftClientSecret', ''))
    setOidcProviderUrl(getOptionValue('OidcWellKnown', ''))
    setOidcClientId(getOptionValue('OidcClientId', ''))
    setOidcClientSecret(getOptionValue('OidcClientSecret', ''))
    setLarkAppId(getOptionValue('LarkClientId', ''))
    setLarkAppSecret(getOptionValue('LarkClientSecret', ''))
    setWechatAppId(getOptionValue('WeChatServerToken', ''))
    setWechatAppSecret(getOptionValue('WeChatAccountQRCodeImageURL', ''))
    setDiscordClientId(getOptionValue('DiscordClientId', ''))
    setDiscordClientSecret(getOptionValue('DiscordClientSecret', ''))
    setEmailNotification(getOptionValue('EmailNotificationEnabled', 'false') === 'true')
    setWebhookUrl(getOptionValue('WebhookUrl', ''))
    setNotifyCheckin(getOptionValue('NotifyCheckin', 'false') === 'true')
    setNotifyTopup(getOptionValue('NotifyTopup', 'false') === 'true')
    setNotifyUsageThreshold(getOptionValue('NotifyUsageThreshold', 'false') === 'true')
    setMemoryCacheEnabled(getOptionValue('MemoryCacheEnabled', 'false') === 'true')
    setCacheSyncFrequency(getOptionValue('SyncFrequency', '600'))
    setEnableMetric(getOptionValue('EnableMetric', 'false') === 'true')
    setChannelTestFrequency(getOptionValue('AutomaticDisableChannelEnabled', 'false') === 'true' ? getOptionValue('ChannelDisableThreshold', '5') : '')
    setBatchTestCount(getOptionValue('RetryTimes', '3'))
    setMinPasswordLength(getOptionValue('MinPasswordLength', '8'))
    setRequireSpecialChars(getOptionValue('RequireSpecialChars', 'false') === 'true')
    setSessionTimeout(getOptionValue('SessionTimeout', '120'))
    setEnforce2fa(getOptionValue('Enforce2FA', 'false') === 'true')
    setIpWhitelist(getOptionValue('IPWhitelist', ''))
    setIpBlacklist(getOptionValue('IPBlacklist', ''))
    setSsrfProtection(getOptionValue('EnableSSRFProtection', 'true') === 'true')
    setCompanyName(getOptionValue('CompanyName', '\u6df1\u5733\u5e02\u4e2d\u79d1\u52b2\u7eac\u667a\u80fd\u6709\u9650\u516c\u53f8'))
    setCompanyTaxId(getOptionValue('CompanyTaxId', '91440300MA5GH45W8C'))
    setCompanyAddress(getOptionValue('CompanyAddress', '\u6df1\u5733\u5e02\u5b9d\u5b89\u533a\u77f3\u5ca9\u8857\u9053\u5858\u5934\u793e\u533a\u5858\u5934\u5927\u905333\u53f7\u4e1c\u6d77\u521b\u610f\u56ed205A'))
    setCompanyPhone(getOptionValue('CompanyPhone', '15920005303'))
    setCompanyBank(getOptionValue('CompanyBank', '\u6df1\u5733\u519c\u6751\u5546\u4e1a\u94f6\u884c\u80a1\u4efd\u6709\u9650\u516c\u53f8\u5e94\u4eba\u77f3\u652f\u884c'))
    setCompanyBankAccount(getOptionValue('CompanyBankAccount', '000396168236'))
    setCompanyAlipayQr(getOptionValue('CompanyAlipayQr', '/payment/alipay-qr.jpg'))
  }, [optionsData])

  // Fetch 2FA status
  useEffect(() => {
    fetch('/api/user/2fa')
      .then(r => r.json())
      .then(data => { if (data.success) setTwoFAStatus(data.data?.enabled || false) })
      .catch(() => {})
  }, [])

  // Fetch security activity
  useEffect(() => {
    setActivityLoading(true)
    fetch('/api/user/self/security/activity')
      .then(r => r.json())
      .then(data => {
        if (data.success && data.data) {
          setSecurityActivity(data.data.activity_logs || [])
        }
      })
      .catch(() => {})
      .finally(() => setActivityLoading(false))
  }, [])

  const saveMutation = useMutation({
    mutationFn: updateOptions,
    onSuccess: () => { toast.success(t('Settings saved')); queryClient.invalidateQueries({ queryKey: ['system-options'] }) },
    onError: () => toast.error(t('Failed to save settings')),
  })

  const handleSave = (newOptions: Record<string, string>) => saveMutation.mutate(newOptions)

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-center h-64"><p className="text-muted-foreground">{t('Loading...')}</p></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">{t('System Settings')}</h1>
          <p className="text-muted-foreground mt-2">{t('Configure system-wide settings')}</p>
        </div>
      </div>

      <Tabs defaultValue="general" className="w-full">
        <TabsList className="flex-wrap h-auto gap-y-1 text-xs sm:text-sm">
          <TabsTrigger value="general" className="px-2 sm:px-3"><Settings className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />{t('General')}</TabsTrigger>
          <TabsTrigger value="billing" className="px-2 sm:px-3"><DollarSign className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />{t('Billing')}</TabsTrigger>
          <TabsTrigger value="ratelimit" className="px-2 sm:px-3"><Gauge className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />{t('Rate Limit')}</TabsTrigger>
          <TabsTrigger value="registration" className="px-2 sm:px-3"><UserPlus className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />{t('Registration')}</TabsTrigger>
          <TabsTrigger value="smtp" className="px-2 sm:px-3"><Mail className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />{t('SMTP')}</TabsTrigger>
          <TabsTrigger value="oauth" className="px-2 sm:px-3"><Lock className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />{t('OAuth')}</TabsTrigger>
          <TabsTrigger value="notification" className="px-2 sm:px-3"><Bell className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />{t('Notification')}</TabsTrigger>
          <TabsTrigger value="performance" className="px-2 sm:px-3"><Zap className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />{t('Performance')}</TabsTrigger>
          <TabsTrigger value="security" className="px-2 sm:px-3"><Shield className="h-3 w-3 sm:h-4 sm:w-4 mr-1" />{t('Security')}</TabsTrigger>
        </TabsList>

        <TabsContent value="general">
          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2"><Settings className="h-4 w-4" />{t('General Settings')}</CardTitle><CardDescription>{t('Basic system configuration')}</CardDescription></CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2"><Label>{t('System Name')}</Label><Input value={systemName} onChange={(e) => setSystemName(e.target.value)} placeholder="QuantumClaw" /><p className="text-xs text-muted-foreground">{t('Displayed in the browser tab and login page')}</p></div>
              <div className="space-y-2"><Label>{t('Logo URL')}</Label><Input value={logoUrl} onChange={(e) => setLogoUrl(e.target.value)} placeholder="/logo.png" /></div>
              <Button onClick={() => handleSave({ SystemName: systemName, Logo: logoUrl, Footer: footerHtml })} disabled={saveMutation.isPending}>
                <Save className="mr-2 h-4 w-4" />{saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="registration">
          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2"><UserPlus className="h-4 w-4" />{t('Registration Settings')}</CardTitle><CardDescription>{t('User registration and sign-up configuration')}</CardDescription></CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between rounded-xl border p-3">
                <div><Label>{t('Allow Registration')}</Label><p className="text-xs text-muted-foreground">{t('Allow new users to self-register')}</p></div>
                <Switch checked={registerEnabled} onCheckedChange={setRegisterEnabled} />
              </div>
              <div className="flex items-center justify-between rounded-xl border p-3">
                <div><Label>{t('Email Verification')}</Label><p className="text-xs text-muted-foreground">{t('Require email verification for new users')}</p></div>
                <Switch checked={emailVerification} onCheckedChange={setEmailVerification} />
              </div>
              <div className="space-y-2"><Label>{t('Default User Group')}</Label><Input value={defaultUserGroup} onChange={(e) => setDefaultUserGroup(e.target.value)} placeholder="default" /></div>
              <div className="space-y-2"><Label>{t('Allowed Registration Domains')}</Label>
                <textarea className="flex min-h-[80px] w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" value={allowedEmails} onChange={(e) => setAllowedEmails(e.target.value)} placeholder="gmail.com&#10;outlook.com&#10;company.com" />
                <p className="text-xs text-muted-foreground">{t('One domain per line. Leave empty to allow all domains.')}</p>
              </div>
              <Button onClick={() => handleSave({ RegisterEnabled: String(registerEnabled), EmailVerificationEnabled: String(emailVerification), TurnstileCheckEnabled: String(turnstileCheck), DefaultUserGroup: defaultUserGroup, EmailDomainWhitelist: allowedEmails })} disabled={saveMutation.isPending}>
                <Save className="mr-2 h-4 w-4" />{saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="security">
          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2"><Shield className="h-4 w-4" />{t('Security Settings')}</CardTitle><CardDescription>{t('Authentication, password policy, and access control')}</CardDescription></CardHeader>
            <CardContent className="space-y-4">
              <div className="rounded-xl border p-4 space-y-3">
                <h3 className="font-semibold">{t('Password Policy')}</h3>
                <div className="space-y-2"><Label>{t('Minimum Password Length')}</Label><Input type="number" value={minPasswordLength} onChange={(e) => setMinPasswordLength(e.target.value)} placeholder="8" /></div>
                <div className="flex items-center justify-between"><div><Label>{t('Require Special Characters')}</Label><p className="text-xs text-muted-foreground">{t('Require at least one special character in password')}</p></div><Switch checked={requireSpecialChars} onCheckedChange={setRequireSpecialChars} /></div>
              </div>
              <div className="space-y-2"><Label>{t('Session Timeout (minutes)')}</Label><Input type="number" value={sessionTimeout} onChange={(e) => setSessionTimeout(e.target.value)} placeholder="120" /></div>
              <div className="flex items-center justify-between rounded-xl border p-3"><div><Label>{t('Enforce 2FA')}</Label><p className="text-xs text-muted-foreground">{t('Require two-factor authentication for all users')}</p></div><Switch checked={enforce2fa} onCheckedChange={setEnforce2fa} /></div>
              <div className="space-y-2"><Label>{t('IP Whitelist')}</Label>
                <textarea className="flex min-h-[100px] w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" value={ipWhitelist} onChange={(e) => setIpWhitelist(e.target.value)} placeholder="192.168.1.0/24&#10;10.0.0.1" />
                <p className="text-xs text-muted-foreground">{t('One IP or CIDR per line. Whitelist takes precedence.')}</p>
              </div>
              <div className="flex items-center justify-between rounded-xl border p-3"><div><Label>{t('SSRF Protection')}</Label><p className="text-xs text-muted-foreground">{t('Block server-side request forgery attempts')}</p></div><Switch checked={ssrfProtection} onCheckedChange={setSsrfProtection} /></div>
              <Button onClick={() => handleSave({ MinPasswordLength: minPasswordLength || '8', RequireSpecialChars: String(requireSpecialChars), SessionTimeout: sessionTimeout || '120', Enforce2FA: String(enforce2fa), IPWhitelist: ipWhitelist, IPBlacklist: ipBlacklist, EnableSSRFProtection: String(ssrfProtection) })} disabled={saveMutation.isPending}>
                <Save className="mr-2 h-4 w-4" />{saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>

          {/⾯ 2FA Section ⾯/}
          <Card className="mt-4">
            <CardHeader><CardTitle className="flex items-center gap-2"><Smartphone className="h-4 w-4" />{t('Two-Factor Authentication')}</CardTitle><CardDescription>{t('Add an extra layer of security to your account')}</CardDescription></CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between rounded-xl border p-3">
                <div>
                  <div className="flex items-center gap-2"><Smartphone className="h-4 w-4 text-muted-foreground" /><Label>{t('2FA Status')}</Label></div>
                  <p className="text-xs text-muted-foreground mt-1">{t('Two-factor authentication using TOTP')}</p>
                </div>
                <Badge variant={twoFAStatus ? 'default' : 'secondary'}>{twoFAStatus ? t('Enabled') : t('Disabled')}</Badge>
              </div>

              {!disabling2FA && !enabling2FA && (
                <Button
                  variant={twoFAStatus ? 'destructive' : 'default'}
                  size="sm"
                  onClick={async () => {
                    if (twoFAStatus) {
                      setDisabling2FA(true)
                      setDisableCode('')
                    } else {
                      setEnabling2FA(true)
                      setVerifyCode('')
                      setQrCodeUrl('')
                      setTotpSecret('')
                      try {
                        const initRes = await fetch('/api/user/self/2fa/init', { method: 'POST' })
                        const initData = await initRes.json()
                        if (initData.success) {
                          const qrRes = await fetch('/api/user/self/2fa/qrcode')
                          const qrData = await qrRes.json()
                          if (qrData.success) {
                            setQrCodeUrl(qrData.data?.qr_code_url || qrData.data?.url || '')
                            setTotpSecret(qrData.data?.secret || '')
                          } else {
                            toast.error(qrData.message || t('Failed to get QR code'))
                          }
                        } else {
                          toast.error(initData.message || t('Failed to initialize 2FA'))
                        }
                      } catch {
                        toast.error(t('Failed to initialize 2FA'))
                      }
                    }
                  }}
                  disabled={twoFALoading}
                >
                  {twoFAStatus ? t('Disable 2FA') : t('Enable 2FA')}
                </Button>
              )}

              {enabling2FA && (
                <div className="rounded-xl border p-4 space-y-3">
                  <h4 className="text-sm font-medium">{t('Set Up Two-Factor Authentication')}</h4>
                  {qrCodeUrl && (
                    <div className="flex justify-center p-2 bg-white dark:bg-black rounded-lg">
                      <img
                        src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrCodeUrl)}`}
                        alt="TOTP QR Code"
                        className="h-44 w-44"
                      />
                    </div>
                  )}
                  {totpSecret && (
                    <div className="flex items-center gap-2">
                      <code className="flex-1 p-1.5 text-xs rounded border bg-background font-mono break-all">{totpSecret}</code>
                      <Button variant="outline" size="icon" className="h-7 w-7 shrink-0" onClick={() => { navigator.clipboard.writeText(totpSecret); toast.success(t('Copied')) }}>
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                  )}
                  <div className="space-y-2">
                    <Label htmlFor="verify-2fa-code">{t('Enter 6-digit TOTP code')}</Label>
                    <div className="flex gap-2">
                      <Input
                        id="verify-2fa-code"
                        type="text"
                        inputMode="numeric"
                        maxLength={6}
                        placeholder="123456"
                        value={verifyCode}
                        onChange={(e) => setVerifyCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                      />
                      <Button
                        size="sm"
                        disabled={verifyCode.length !== 6}
                        onClick={async () => {
                          const res = await fetch('/api/user/self/2fa/verify', {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ code: verifyCode }),
                          })
                          const data = await res.json()
                          if (data.success) {
                            setTwoFAStatus(true)
                            setEnabling2FA(false)
                            setVerifyCode('')
                            toast.success(t('2FA enabled successfully'))
                          } else {
                            toast.error(data.message || t('Invalid code'))
                          }
                        }}
                      >
                        {t('Verify & Enable')}
                      </Button>
                      <Button variant="outline" size="sm" onClick={() => { setEnabling2FA(false); setVerifyCode(''); setQrCodeUrl(''); setTotpSecret('') }}>
                        {t('Cancel')}
                      </Button>
                    </div>
                  </div>
                  <p className="text-xs text-muted-foreground">{t('Scan the QR code with your authenticator app and enter the 6-digit code to enable 2FA')}</p>
                </div>
              )}

              {disabling2FA && (
                <div className="rounded-xl border border-red-200 p-4 space-y-3 dark:border-red-900/50">
                  <p className="text-sm font-medium text-red-700 dark:text-red-400">{t('Disable Two-Factor Authentication')}</p>
                  <div className="flex gap-2">
                    <Input
                      type="text"
                      inputMode="numeric"
                      maxLength={6}
                      placeholder="123456"
                      value={disableCode}
                      onChange={(e) => setDisableCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                    />
                    <Button
                      variant="destructive"
                      size="sm"
                      disabled={disableCode.length !== 6}
                      onClick={async () => {
                        const res = await fetch('/api/user/self/2fa/disable', {
                          method: 'POST',
                          headers: { 'Content-Type': 'application/json' },
                          body: JSON.stringify({ code: disableCode }),
                        })
                        const data = await res.json()
                        if (data.success) {
                          setTwoFAStatus(false)
                          setDisabling2FA(false)
                          setDisableCode('')
                          toast.success(t('2FA disabled'))
                        } else {
                          toast.error(data.message || t('Failed to disable 2FA'))
                        }
                      }}
                    >
                      {t('Confirm Disable')}
                    </Button>
                    <Button variant="outline" size="sm" onClick={() => { setDisabling2FA(false); setDisableCode('') }}>{t('Cancel')}</Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* WebAuthn Security Keys Section */}
          <Card className="mt-4">
            <CardHeader><CardTitle className="flex items-center gap-2"><Fingerprint className="h-4 w-4" />{t('WebAuthn Security Keys')}</CardTitle><CardDescription>{t('Manage your passkeys and security keys')}</CardDescription></CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-2 mb-2">
                <p className="text-sm text-muted-foreground flex-1">{t('Use biometrics, security keys, or platform authenticators for passwordless sign-in')}</p>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={registeringWebAuthn}
                  onClick={async () => {
                    setRegisteringWebAuthn(true)
                    try {
                      const beginRes = await fetch('/api/webauthn/register/begin', { method: 'POST' })
                      const beginData = await beginRes.json()
                      if (!beginData.success) {
                        toast.error(beginData.message || t('Failed to start registration'))
                        return
                      }
                      const rawData = beginData.data as any
                      if (rawData.challenge) {
                        const challengeStr = rawData.challenge as string
                        rawData.challenge = Uint8Array.from(atob(challengeStr), c => c.charCodeAt(0))
                      }
                      if (rawData.user?.id) {
                        const userIdStr = rawData.user.id as string
                        rawData.user.id = Uint8Array.from(atob(userIdStr), c => c.charCodeAt(0))
                      }
                      const credential = await navigator.credentials.create({ publicKey: rawData as PublicKeyCredentialCreationOptions })
                      if (!credential) {
                        toast.error(t('Registration cancelled'))
                        return
                      }
                      const finishRes = await fetch('/api/webauthn/register/finish', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                          id: credential.id,
                          rawId: credential.id,
                          type: credential.type,
                          response: (credential as PublicKeyCredential).response,
                        }),
                      })
                      const finishData = await finishRes.json()
                      if (finishData.success) {
                        toast.success(t('Security key registered'))
                        queryClient.invalidateQueries({ queryKey: ['settings-webauthn-credentials'] })
                      } else {
                        toast.error(finishData.message || t('Failed to register security key'))
                      }
                    } catch (err: any) {
                      toast.error(err?.message || t('Failed to register security key'))
                    } finally {
                      setRegisteringWebAuthn(false)
                    }
                  }}
                >
                  <Plus className="mr-1 h-3 w-3" />
                  {registeringWebAuthn ? t('Registering...') : t('Register New Key')}
                </Button>
              </div>
              {webauthnCredentials.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground text-sm">
                  <Key className="h-12 w-12 mx-auto mb-3 opacity-30" />
                  <p>{t('No security keys registered yet')}</p>
                </div>
              ) : (
                <div className="space-y-2">
                  {webauthnCredentials.map((cred: any) => (
                    <div
                      key={cred.id || cred.credential_id}
                      className="flex items-center justify-between p-3 rounded-xl border gap-2"
                    >
                      <div className="flex items-center gap-3 min-w-0">
                        <Fingerprint className="h-5 w-5 text-primary shrink-0" />
                        <div className="min-w-0">
                          <p className="text-sm font-medium truncate">{cred.device_name || t('Security Key')}</p>
                          <p className="text-xs text-muted-foreground">
                            {t('Registered')}: {cred.created_time ? new Date(cred.created_time * 1000).toLocaleDateString() : '-'}
                          </p>
                        </div>
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 shrink-0 text-destructive hover:text-destructive"
                        onClick={async () => {
                          const res = await fetch(`/api/user/self/webauthn/credentials/${cred.credential_id}`, {
                            method: 'DELETE',
                          })
                          const data = await res.json()
                          if (data.success) {
                            toast.success(t('Security key removed'))
                            queryClient.invalidateQueries({ queryKey: ['settings-webauthn-credentials'] })
                          } else {
                            toast.error(data.message || t('Failed to remove security key'))
                          }
                        }}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/⾯ Security Activity Section ⾯/}
          <Card className="mt-4">
            <CardHeader><CardTitle className="flex items-center gap-2"><Clock className="h-4 w-4" />{t('Security Activity')}</CardTitle><CardDescription>{t('Recent security events and login activity')}</CardDescription></CardHeader>
            <CardContent>
              {activityLoading ? (
                <div className="flex items-center justify-center py-8"><div className="h-5 w-5 rounded-full border-2 border-amber-500/30 border-t-amber-500 animate-spin" /></div>
              ) : securityActivity.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground text-sm">{t('No security activity recorded yet')}</div>
              ) : (
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t('Time')}</TableHead>
                        <TableHead>{t('Type')}</TableHead>
                        <TableHead>{t('Details')}</TableHead>
                        <TableHead>{t('IP Address')}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {securityActivity.map((entry: any, idx: number) => (
                        <TableRow key={entry.id || idx}>
                          <TableCell className="text-xs font-mono whitespace-nowrap">
                            {entry.created_time
                              ? new Date(entry.created_time * 1000).toLocaleString()
                              : entry.time
                                ? new Date(entry.time).toLocaleString()
                                : '-'
                            }
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline" className="text-xs">
                              {entry.type === 3 ? t('Admin Action') : entry.type === 2 ? t('Login') : entry.type === 1 ? t('Register') : t('Other')}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-xs max-w-[200px] truncate">{entry.content || entry.description || entry.action || '-'}</TableCell>
                          <TableCell className="text-xs font-mono">{entry.ip || entry.ip_address || '-'}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="smtp">
          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2"><Mail className="h-4 w-4" />{t('SMTP Settings')}</CardTitle><CardDescription>{t('Configure email server for sending emails')}</CardDescription></CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2"><Label>{t('SMTP Host')}</Label><Input value={smtpHost} onChange={(e) => setSmtpHost(e.target.value)} placeholder="smtp.gmail.com" /></div>
              <div className="space-y-2"><Label>{t('SMTP Port')}</Label><Input type="number" value={smtpPort} onChange={(e) => setSmtpPort(e.target.value)} placeholder="587" /></div>
              <div className="space-y-2"><Label>{t('SMTP Username')}</Label><Input value={smtpUsername} onChange={(e) => setSmtpUsername(e.target.value)} placeholder="user@gmail.com" /></div>
              <div className="space-y-2"><Label>{t('SMTP Password')}</Label><Input type="password" value={smtpPassword} onChange={(e) => setSmtpPassword(e.target.value)} placeholder={'\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022'} /></div>
              <Button onClick={() => handleSave({ SMTPServer: smtpHost, SMTPPort: smtpPort || '587', SMTPAccount: smtpUsername, SMTPToken: smtpPassword, SMTPFrom: smtpFrom })} disabled={saveMutation.isPending}>
                <Save className="mr-2 h-4 w-4" />{saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="billing">
          <Card>
            <CardHeader><CardTitle className="flex items-center gap-2"><DollarSign className="h-4 w-4" />{t('Payment Merchant')}</CardTitle><CardDescription>{t('Configure payment providers for user top-up')}</CardDescription></CardHeader>
            <CardContent className="space-y-6">

              {/* Epay */}
              <div className="space-y-3 rounded-xl border p-4">
                <div className="flex items-center justify-between">
                  <Label className="text-base">{t('Epay')}</Label>
                  <Switch checked={epayEnabled} onCheckedChange={setEpayEnabled} />
                </div>
                {epayEnabled && (
                  <div className="space-y-3 pl-2 border-l-2 border-amber-200">
                    <div><Label>{t('Merchant ID')}</Label><Input value={epayId} onChange={e => setEpayId(e.target.value)} placeholder="1001" /></div>
                    <div><Label>{t('Merchant Key')}</Label><Input value={epayKey} onChange={e => setEpayKey(e.target.value)} placeholder="xxxxxxxx" type="password" /></div>
                    <div><Label>{t('Gateway URL')}</Label><Input value={epayAddress} onChange={e => setEpayAddress(e.target.value)} placeholder="https://epay.example.com" /></div>
                  </div>
                )}
              </div>

              {/* Stripe */}
              <div className="space-y-3 rounded-xl border p-4">
                <div className="flex items-center justify-between">
                  <Label className="text-base">{t('Stripe')}</Label>
                  <Switch checked={stripeEnabled} onCheckedChange={setStripeEnabled} />
                </div>
                {stripeEnabled && (
                  <div className="space-y-3 pl-2 border-l-2 border-blue-200">
                    <div><Label>{t('API Secret')}</Label><Input value={stripeApiSecret} onChange={e => setStripeApiSecret(e.target.value)} placeholder="sk_test_..." type="password" /></div>
                    <div><Label>{t('Min Top-up')}</Label><Input value={stripeMinTopUp} onChange={e => setStripeMinTopUp(e.target.value)} placeholder="1" type="number" /></div>
                  </div>
                )}
              </div>

              {/* Creem */}
              <div className="space-y-3 rounded-xl border p-4">
                <div className="flex items-center justify-between">
                  <Label className="text-base">{t('Creem')}</Label>
                  <Switch checked={creemEnabled} onCheckedChange={setCreemEnabled} />
                </div>
                {creemEnabled && (
                  <div className="space-y-3 pl-2 border-l-2 border-green-200">
                    <div><Label>{t('API Key')}</Label><Input value={creemApiKey} onChange={e => setCreemApiKey(e.target.value)} placeholder="creem_..." type="password" /></div>
                  </div>
                )}
              </div>

              {/* Waffo */}
              <div className="space-y-3 rounded-xl border p-4">
                <div className="flex items-center justify-between">
                  <Label className="text-base">{t('Waffo')}</Label>
                  <Switch checked={waffoEnabled} onCheckedChange={setWaffoEnabled} />
                </div>
                {waffoEnabled && (
                  <div className="space-y-3 pl-2 border-l-2 border-purple-200">
                    <div><Label>{t('API Key')}</Label><Input value={waffoApiKey} onChange={e => setWaffoApiKey(e.target.value)} placeholder="waffo_..." type="password" /></div>
                    <div className="flex items-center justify-between">
                      <Label>{t('Sandbox Mode')}</Label>
                      <Switch checked={waffoSandbox} onCheckedChange={setWaffoSandbox} />
                    </div>
                  </div>
                )}
              </div>

              {/* Binance */}
              <div className="space-y-3 rounded-xl border p-4">
                <div className="flex items-center justify-between">
                  <Label className="text-base">{t('Binance Pay')}</Label>
                  <Switch checked={binanceEnabled} onCheckedChange={setBinanceEnabled} />
                </div>
                {binanceEnabled && (
                  <div className="space-y-3 pl-2 border-l-2 border-yellow-200">
                    <div><Label>{t('API Key')}</Label><Input value={binanceApiKey} onChange={e => setBinanceApiKey(e.target.value)} placeholder="binance_..." type="password" /></div>
                    <div><Label>{t('Secret Key')}</Label><Input value={binanceSecretKey} onChange={e => setBinanceSecretKey(e.target.value)} placeholder="..." type="password" /></div>
                    <div><Label>{t('Merchant ID')}</Label><Input value={binanceMerchantId} onChange={e => setBinanceMerchantId(e.target.value)} placeholder="..." /></div>
                  </div>
                )}
              </div>

              <Button onClick={() => handleSave({
                EpayEnabled: epayEnabled ? 'true' : 'false',
                EpayId: epayId,
                EpayKey: epayKey,
                EpayAddress: epayAddress,
                StripeEnabled: stripeEnabled ? 'true' : 'false',
                StripeApiSecret: stripeApiSecret,
                StripeMinTopUp: stripeMinTopUp || '1',
                CreemEnabled: creemEnabled ? 'true' : 'false',
                CreemApiKey: creemApiKey,
                WaffoEnabled: waffoEnabled ? 'true' : 'false',
                WaffoApiKey: waffoApiKey,
                WaffoSandbox: waffoSandbox ? 'true' : 'false',
                BinanceEnabled: binanceEnabled ? 'true' : 'false',
                BinanceApiKey: binanceApiKey,
                BinanceSecretKey: binanceSecretKey,
                BinanceMerchantId: binanceMerchantId,
              })} disabled={saveMutation.isPending}>
                <Save className="mr-2 h-4 w-4" />{saveMutation.isPending ? t('Saving...') : t('Save Payment Settings')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
