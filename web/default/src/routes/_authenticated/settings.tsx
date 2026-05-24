import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useAuthStore } from '@/stores/auth-store'
import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Save, Settings, DollarSign, Gauge, UserPlus, Mail, Lock, Bell, Zap, Shield, Server } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { getOptions, updateOptions, updateOption, type SystemOption } from '@/lib/api-extended'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/settings')({
  component: SettingsPage,
})

function SettingsPage() {
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
      </Tabs>
    </div>
  )
}
