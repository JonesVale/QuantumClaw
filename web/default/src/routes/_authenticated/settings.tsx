import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Save, Settings, DollarSign, Gauge, UserPlus, Mail, Lock, Bell, Zap, Shield } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { getOptions, updateOptions, updateOption, type SystemOption } from '@/lib/api-extended'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const { data: optionsData, isLoading } = useQuery({
    queryKey: ['system-options'],
    queryFn: getOptions,
    staleTime: 30 * 1000,
  })

  const options: SystemOption[] = optionsData?.data || []
  const getOptionValue = (key: string, fallback: string = '') =>
    options.find((o) => o.key === key)?.value || fallback

  // ── General ──
  const [systemName, setSystemName] = useState('')
  const [logoUrl, setLogoUrl] = useState('')
  const [footerHtml, setFooterHtml] = useState('')

  // ── Billing ──
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

  // ── RateLimit ──
  const [globalApiRateLimit, setGlobalApiRateLimit] = useState('')
  const [globalWebRateLimit, setGlobalWebRateLimit] = useState('')
  const [ipRateLimitInput, setIpRateLimitInput] = useState('')
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('')
  const [turnstileSecretKey, setTurnstileSecretKey] = useState('')

  // ── Registration ──
  const [registerEnabled, setRegisterEnabled] = useState(true)
  const [emailVerification, setEmailVerification] = useState(false)
  const [turnstileCheck, setTurnstileCheck] = useState(false)
  const [defaultUserGroup, setDefaultUserGroup] = useState('')
  const [allowedEmails, setAllowedEmails] = useState('')

  // ── SMTP ──
  const [smtpHost, setSmtpHost] = useState('')
  const [smtpPort, setSmtpPort] = useState('')
  const [smtpUsername, setSmtpUsername] = useState('')
  const [smtpPassword, setSmtpPassword] = useState('')
  const [smtpFrom, setSmtpFrom] = useState('')

  // ── OAuth ──
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

  // ── Notification ──
  const [emailNotification, setEmailNotification] = useState(false)
  const [webhookUrl, setWebhookUrl] = useState('')
  const [notifyCheckin, setNotifyCheckin] = useState(false)
  const [notifyTopup, setNotifyTopup] = useState(false)
  const [notifyUsageThreshold, setNotifyUsageThreshold] = useState(false)

  // ── Performance ──
  const [memoryCacheEnabled, setMemoryCacheEnabled] = useState(false)
  const [cacheSyncFrequency, setCacheSyncFrequency] = useState('')
  const [enableMetric, setEnableMetric] = useState(false)
  const [channelTestFrequency, setChannelTestFrequency] = useState('')
  const [batchTestCount, setBatchTestCount] = useState('')

  // ── Company Info ──
  const [companyName, setCompanyName] = useState('')
  const [companyTaxId, setCompanyTaxId] = useState('')
  const [companyAddress, setCompanyAddress] = useState('')
  const [companyPhone, setCompanyPhone] = useState('')
  const [companyBank, setCompanyBank] = useState('')
  const [companyBankAccount, setCompanyBankAccount] = useState('')
  const [companyAlipayQr, setCompanyAlipayQr] = useState('')

  // ── Security ──
  const [minPasswordLength, setMinPasswordLength] = useState('')
  const [requireSpecialChars, setRequireSpecialChars] = useState(false)
  const [sessionTimeout, setSessionTimeout] = useState('')
  const [enforce2fa, setEnforce2fa] = useState(false)
  const [ipWhitelist, setIpWhitelist] = useState('')
  const [ipBlacklist, setIpBlacklist] = useState('')
  const [ssrfProtection, setSsrfProtection] = useState(false)

  // ── Sync local state from loaded options ──
  useEffect(() => {
    if (!optionsData?.data) return

    setSystemName(getOptionValue('SystemName', 'QuantumClaw'))
    setLogoUrl(getOptionValue('Logo', '/logo.png'))
    setFooterHtml(getOptionValue('FooterHTML', ''))

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

    setCompanyName(getOptionValue('CompanyName', '深圳市中科劲纬智能有限公司'))
    setCompanyTaxId(getOptionValue('CompanyTaxId', '91440300MA5GH45W8C'))
    setCompanyAddress(getOptionValue('CompanyAddress', '深圳市宝安区石岩街道塘头社区塘头大道33号东海创意园205A'))
    setCompanyPhone(getOptionValue('CompanyPhone', '15920005303'))
    setCompanyBank(getOptionValue('CompanyBank', '深圳农村商业银行股份有限公司应人石支行'))
    setCompanyBankAccount(getOptionValue('CompanyBankAccount', '000396168236'))
    setCompanyAlipayQr(getOptionValue('CompanyAlipayQr', '/payment/alipay-qr.jpg'))
  }, [optionsData])

  const saveMutation = useMutation({
    mutationFn: updateOptions,
    onSuccess: () => {
      toast.success(t('Settings saved'))
      queryClient.invalidateQueries({ queryKey: ['system-options'] })
    },
    onError: () => toast.error(t('Failed to save settings')),
  })

  const handleSave = (newOptions: Record<string, string>) => {
    saveMutation.mutate(newOptions)
  }

  const handleTestEmail = async () => {
    try {
      const res = await updateOption('smtp_test', smtpHost)
      if (res.success) {
        toast.success(t('Test email sent'))
      } else {
        toast.error(t('Failed to send test email'))
      }
    } catch {
      toast.error(t('Failed to send test email'))
    }
  }

  if (isLoading) {
    return (
      <div className=" w-full p-4 sm:p-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
        <div className="flex items-center justify-center h-64">
          <p className="text-muted-foreground">{t('Loading...')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className=" w-full p-4 sm:p-6 space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-xl sm:text-2xl lg:text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('System Settings')}
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            {t('Configure system-wide settings')}
          </p>
        </div>
      </div>

      <Tabs defaultValue="general" className="w-full">
        <TabsList className="flex-wrap h-auto">
          <TabsTrigger value="general">
            <Settings className="h-4 w-4 mr-2" />
            {t('General')}
          </TabsTrigger>
          <TabsTrigger value="billing">
            <DollarSign className="h-4 w-4 mr-2" />
            {t('Billing')}
          </TabsTrigger>
          <TabsTrigger value="ratelimit">
            <Gauge className="h-4 w-4 mr-2" />
            {t('RateLimit')}
          </TabsTrigger>
          <TabsTrigger value="registration">
            <UserPlus className="h-4 w-4 mr-2" />
            {t('Registration')}
          </TabsTrigger>
          <TabsTrigger value="smtp">
            <Mail className="h-4 w-4 mr-2" />
            {t('SMTP')}
          </TabsTrigger>
          <TabsTrigger value="oauth">
            <Lock className="h-4 w-4 mr-2" />
            {t('OAuth')}
          </TabsTrigger>
          <TabsTrigger value="notification">
            <Bell className="h-4 w-4 mr-2" />
            {t('Notification')}
          </TabsTrigger>
          <TabsTrigger value="performance">
            <Zap className="h-4 w-4 mr-2" />
            {t('Performance')}
          </TabsTrigger>
          <TabsTrigger value="security">
            <Shield className="h-4 w-4 mr-2" />
            {t('Security')}
          </TabsTrigger>
        </TabsList>

        {/* ════════════════════════════════════════════════════ General ═══ */}
        <TabsContent value="general">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="h-4 w-4" />
                {t('General Settings')}
              </CardTitle>
              <CardDescription>{t('Basic system configuration')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>{t('System Name')}</Label>
                <Input
                  value={systemName}
                  onChange={(e) => setSystemName(e.target.value)}
                  placeholder="QuantumClaw"
                />
                <p className="text-xs text-muted-foreground">
                  {t('Displayed in the browser tab and login page')}
                </p>
              </div>
              <div className="space-y-2">
                <Label>{t('Logo URL')}</Label>
                <Input
                  value={logoUrl}
                  onChange={(e) => setLogoUrl(e.target.value)}
                  placeholder="/logo.png"
                />
              </div>
              <div className="space-y-2">
                <Label>{t('Footer HTML')}</Label>
                <Input
                  value={footerHtml}
                  onChange={(e) => setFooterHtml(e.target.value)}
                  placeholder="<a href='...'>...</a>"
                />
              </div>
              <Button
                onClick={() =>
                  handleSave({
                    SystemName: systemName,
                    Logo: logoUrl,
                    FooterHTML: footerHtml,
                  })
                }
                disabled={saveMutation.isPending}
              >
                <Save className="mr-2 h-4 w-4" />
                {saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ════════════════════════════════════════════════════ Billing ═══ */}
        <TabsContent value="billing">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <DollarSign className="h-4 w-4" />
                {t('Billing Settings')}
              </CardTitle>
              <CardDescription>{t('Configure billing and payment options')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* General billing */}
              <div className="space-y-4 rounded-lg border p-4">
                <h3 className="font-semibold">{t('General Billing')}</h3>
                <div className="space-y-2">
                  <Label>{t('Default Currency')}</Label>
                  <Select
                    value={defaultCurrency || 'USD'}
                    onValueChange={setDefaultCurrency}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder={t('Select currency')} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="USD">USD ($)</SelectItem>
                      <SelectItem value="CNY">CNY (¥)</SelectItem>
                      <SelectItem value="EUR">EUR (€)</SelectItem>
                      <SelectItem value="GBP">GBP (£)</SelectItem>
                      <SelectItem value="JPY">JPY (¥)</SelectItem>
                      <SelectItem value="KRW">KRW (₩)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>{t('Minimum Top-up Amount')}</Label>
                  <Input
                    type="number"
                    value={minTopUp}
                    onChange={(e) => setMinTopUp(e.target.value)}
                    placeholder="1"
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('Minimum amount for a single top-up')}
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>{t('Cache Billing Ratio')}</Label>
                  <Input
                    type="number"
                    step="0.1"
                    value={cacheBillingRatio}
                    onChange={(e) => setCacheBillingRatio(e.target.value)}
                    placeholder="1.0"
                  />
                  <p className="text-xs text-muted-foreground">
                    {t('Cost ratio for cache model usage (1.0 = 100%)')}
                  </p>
                </div>
              </div>

              {/* Stripe */}
              <div className="space-y-3 rounded-lg border p-4">
                <h3 className="font-semibold">Stripe</h3>
                <div className="flex flex-row items-center justify-between">
                  <div>
                    <Label>{t('Enable Stripe')}</Label>
                    <p className="text-xs text-muted-foreground">{t('Accept payments via Stripe')}</p>
                  </div>
                  <Switch
                    checked={stripeEnabled}
                    onCheckedChange={setStripeEnabled}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Stripe Public Key')}</Label>
                  <Input
                    value={stripePublicKey}
                    onChange={(e) => setStripePublicKey(e.target.value)}
                    placeholder="pk_live_..."
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Stripe API Secret')}</Label>
                  <Input
                    type="password"
                    value={stripeApiSecret}
                    onChange={(e) => setStripeApiSecret(e.target.value)}
                    placeholder="sk_live_..."
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Stripe Min Top-up')}</Label>
                  <Input
                    type="number"
                    value={stripeMinTopUp}
                    onChange={(e) => setStripeMinTopUp(e.target.value)}
                    placeholder="1"
                  />
                </div>
              </div>

              {/* Epay */}
              <div className="space-y-3 rounded-lg border p-4">
                <h3 className="font-semibold">Epay (易支付)</h3>
                <div className="flex flex-row items-center justify-between">
                  <div>
                    <Label>{t('Enable Epay')}</Label>
                    <p className="text-xs text-muted-foreground">{t('Accept payments via Epay')}</p>
                  </div>
                  <Switch
                    checked={epayEnabled}
                    onCheckedChange={setEpayEnabled}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Epay ID')}</Label>
                  <Input
                    value={epayId}
                    onChange={(e) => setEpayId(e.target.value)}
                    placeholder="1001"
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Epay Key')}</Label>
                  <Input
                    type="password"
                    value={epayKey}
                    onChange={(e) => setEpayKey(e.target.value)}
                    placeholder="••••••••"
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Epay Gateway Address')}</Label>
                  <Input
                    value={epayAddress}
                    onChange={(e) => setEpayAddress(e.target.value)}
                    placeholder="https://epay.example.com"
                  />
                </div>
              </div>

              {/* Creem */}
              <div className="space-y-3 rounded-lg border p-4">
                <h3 className="font-semibold">Creem</h3>
                <div className="flex flex-row items-center justify-between">
                  <div>
                    <Label>{t('Enable Creem')}</Label>
                    <p className="text-xs text-muted-foreground">{t('Accept payments via Creem')}</p>
                  </div>
                  <Switch
                    checked={creemEnabled}
                    onCheckedChange={setCreemEnabled}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Creem API Key')}</Label>
                  <Input
                    type="password"
                    value={creemApiKey}
                    onChange={(e) => setCreemApiKey(e.target.value)}
                    placeholder="••••••••"
                  />
                </div>
              </div>

              {/* Waffo */}
              <div className="space-y-3 rounded-lg border p-4">
                <h3 className="font-semibold">Waffo</h3>
                <div className="flex flex-row items-center justify-between">
                  <div>
                    <Label>{t('Enable Waffo')}</Label>
                    <p className="text-xs text-muted-foreground">{t('Accept payments via Waffo')}</p>
                  </div>
                  <Switch
                    checked={waffoEnabled}
                    onCheckedChange={setWaffoEnabled}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Waffo API Key')}</Label>
                  <Input
                    type="password"
                    value={waffoApiKey}
                    onChange={(e) => setWaffoApiKey(e.target.value)}
                    placeholder="••••••••"
                  />
                </div>
                <div className="flex flex-row items-center justify-between rounded-lg border p-3">
                  <div>
                    <Label>{t('Waffo Sandbox Mode')}</Label>
                    <p className="text-xs text-muted-foreground">{t('Use sandbox environment for testing')}</p>
                  </div>
                  <Switch
                    checked={waffoSandbox}
                    onCheckedChange={setWaffoSandbox}
                  />
                </div>
              </div>

              {/* Binance */}
              <div className="space-y-3 rounded-lg border p-4">
                <h3 className="font-semibold">Binance Pay</h3>
                <div className="flex flex-row items-center justify-between">
                  <div>
                    <Label>{t('Enable Binance Pay')}</Label>
                    <p className="text-xs text-muted-foreground">{t('Accept payments via Binance Pay')}</p>
                  </div>
                  <Switch
                    checked={binanceEnabled}
                    onCheckedChange={setBinanceEnabled}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Binance API Key')}</Label>
                  <Input
                    type="password"
                    value={binanceApiKey}
                    onChange={(e) => setBinanceApiKey(e.target.value)}
                    placeholder="••••••••"
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Binance Secret Key')}</Label>
                  <Input
                    type="password"
                    value={binanceSecretKey}
                    onChange={(e) => setBinanceSecretKey(e.target.value)}
                    placeholder="••••••••"
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Binance Merchant ID')}</Label>
                  <Input
                    value={binanceMerchantId}
                    onChange={(e) => setBinanceMerchantId(e.target.value)}
                    placeholder="m_..."
                  />
                </div>
              </div>

              {/* Company Info */}
              <div className="space-y-3 rounded-lg border p-4">
                <h3 className="font-semibold">{t('Company Payment Info')}</h3>
                <p className="text-xs text-muted-foreground">{t('Shown on Wallet page for B2B transfers')}</p>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label>{t('Company Name')}</Label>
                    <Input value={companyName} onChange={(e) => setCompanyName(e.target.value)} />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('Tax ID')}</Label>
                    <Input value={companyTaxId} onChange={(e) => setCompanyTaxId(e.target.value)} />
                  </div>
                  <div className="space-y-2 sm:col-span-2">
                    <Label>{t('Address')}</Label>
                    <Input value={companyAddress} onChange={(e) => setCompanyAddress(e.target.value)} />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('Phone')}</Label>
                    <Input value={companyPhone} onChange={(e) => setCompanyPhone(e.target.value)} />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('Bank Name')}</Label>
                    <Input value={companyBank} onChange={(e) => setCompanyBank(e.target.value)} />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('Bank Account')}</Label>
                    <Input value={companyBankAccount} onChange={(e) => setCompanyBankAccount(e.target.value)} />
                  </div>
                </div>
              </div>

              <Button
                onClick={() =>
                  handleSave({
                    DefaultCurrency: defaultCurrency || 'USD',
                    StripePublicKey: stripePublicKey,
                    StripeApiSecret: stripeApiSecret,
                    StripeEnabled: String(stripeEnabled),
                    StripeMinTopUp: stripeMinTopUp || '1',
                    EpayEnabled: String(epayEnabled),
                    EpayId: epayId,
                    EpayKey: epayKey,
                    EpayAddress: epayAddress,
                    CreemEnabled: String(creemEnabled),
                    CreemApiKey: creemApiKey,
                    WaffoEnabled: String(waffoEnabled),
                    WaffoApiKey: waffoApiKey,
                    WaffoSandbox: String(waffoSandbox),
                    BinanceEnabled: String(binanceEnabled),
                    BinanceApiKey: binanceApiKey,
                    BinanceSecretKey: binanceSecretKey,
                    BinanceMerchantId: binanceMerchantId,
                    MinTopUp: minTopUp || '1',
                    CacheBillingRatio: cacheBillingRatio || '1.0',
                    CompanyName: companyName,
                    CompanyTaxId: companyTaxId,
                    CompanyAddress: companyAddress,
                    CompanyPhone: companyPhone,
                    CompanyBank: companyBank,
                    CompanyBankAccount: companyBankAccount,
                    CompanyAlipayQr: companyAlipayQr,
                  })
                }
                disabled={saveMutation.isPending}
              >
                <Save className="mr-2 h-4 w-4" />
                {saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ═══════════════════════════════════════════════════ RateLimit ═══ */}
        <TabsContent value="ratelimit">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Gauge className="h-4 w-4" />
                {t('Rate Limit Settings')}
              </CardTitle>
              <CardDescription>{t('Configure API and web rate limits')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>{t('Global API Rate Limit (RPM)')}</Label>
                <Input
                  type="number"
                  value={globalApiRateLimit}
                  onChange={(e) => setGlobalApiRateLimit(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">
                  {t('Max API requests per minute globally')}
                </p>
              </div>
              <div className="space-y-2">
                <Label>{t('Global Web Rate Limit (RPM)')}</Label>
                <Input
                  type="number"
                  value={globalWebRateLimit}
                  onChange={(e) => setGlobalWebRateLimit(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">
                  {t('Max web requests per minute globally')}
                </p>
              </div>
              <div className="space-y-2">
                <Label>{t('IP Rate Limit (RPM)')}</Label>
                <Input
                  type="number"
                  value={ipRateLimitInput}
                  onChange={(e) => setIpRateLimitInput(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">
                  {t('Max requests per minute per IP address')}
                </p>
              </div>
              <div className="space-y-2">
                <Label>{t('Turnstile Site Key')}</Label>
                <Input
                  value={turnstileSiteKey}
                  onChange={(e) => setTurnstileSiteKey(e.target.value)}
                  placeholder="0x4AAAAAA..."
                />
              </div>
              <div className="space-y-2">
                <Label>{t('Turnstile Secret Key')}</Label>
                <Input
                  type="password"
                  value={turnstileSecretKey}
                  onChange={(e) => setTurnstileSecretKey(e.target.value)}
                  placeholder="0x4AAAAAA..."
                />
              </div>
              <Button
                onClick={() =>
                  handleSave({
                    GlobalApiRateLimitNum: globalApiRateLimit || '480',
                    GlobalWebRateLimitNum: globalWebRateLimit || '240',
                    CriticalRateLimitNum: ipRateLimitInput || '20',
                    TurnstileSiteKey: turnstileSiteKey,
                    TurnstileSecretKey: turnstileSecretKey,
                  })
                }
                disabled={saveMutation.isPending}
              >
                <Save className="mr-2 h-4 w-4" />
                {saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ════════════════════════════════════════════ Registration ═══ */}
        <TabsContent value="registration">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <UserPlus className="h-4 w-4" />
                {t('Registration Settings')}
              </CardTitle>
              <CardDescription>{t('User registration and sign-up configuration')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-row items-center justify-between rounded-lg border p-3">
                <div>
                  <Label>{t('Allow Registration')}</Label>
                  <p className="text-xs text-muted-foreground">
                    {t('Allow new users to self-register')}
                  </p>
                </div>
                <Switch
                  checked={registerEnabled}
                  onCheckedChange={setRegisterEnabled}
                />
              </div>
              <div className="flex flex-row items-center justify-between rounded-lg border p-3">
                <div>
                  <Label>{t('Email Verification')}</Label>
                  <p className="text-xs text-muted-foreground">
                    {t('Require email verification for new users')}
                  </p>
                </div>
                <Switch
                  checked={emailVerification}
                  onCheckedChange={setEmailVerification}
                />
              </div>
              <div className="flex flex-row items-center justify-between rounded-lg border p-3">
                <div>
                  <Label>{t('Turnstile Check')}</Label>
                  <p className="text-xs text-muted-foreground">
                    {t('Enable Cloudflare Turnstile captcha during registration')}
                  </p>
                </div>
                <Switch
                  checked={turnstileCheck}
                  onCheckedChange={setTurnstileCheck}
                />
              </div>
              <div className="space-y-2">
                <Label>{t('Default User Group')}</Label>
                <Input
                  value={defaultUserGroup}
                  onChange={(e) => setDefaultUserGroup(e.target.value)}
                  placeholder="default"
                />
              </div>
              <div className="space-y-2">
                <Label>{t('Allowed Registration Domains')}</Label>
                <textarea
                  className="flex min-h-[80px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                  value={allowedEmails}
                  onChange={(e) => setAllowedEmails(e.target.value)}
                  placeholder="gmail.com&#10;outlook.com&#10;company.com"
                />
                <p className="text-xs text-muted-foreground">
                  {t('One domain per line. Leave empty to allow all domains.')}
                </p>
              </div>
              <Button
                onClick={() =>
                  handleSave({
                    RegisterEnabled: String(registerEnabled),
                    EmailVerificationEnabled: String(emailVerification),
                    TurnstileCheckEnabled: String(turnstileCheck),
                    DefaultUserGroup: defaultUserGroup,
                    EmailDomainWhitelist: allowedEmails,
                  })
                }
                disabled={saveMutation.isPending}
              >
                <Save className="mr-2 h-4 w-4" />
                {saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ═════════════════════════════════════════════════════ SMTP ═══ */}
        <TabsContent value="smtp">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Mail className="h-4 w-4" />
                {t('SMTP Settings')}
              </CardTitle>
              <CardDescription>{t('Configure email server for sending emails')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>{t('SMTP Host')}</Label>
                <Input
                  value={smtpHost}
                  onChange={(e) => setSmtpHost(e.target.value)}
                  placeholder="smtp.gmail.com"
                />
              </div>
              <div className="space-y-2">
                <Label>{t('SMTP Port')}</Label>
                <Input
                  type="number"
                  value={smtpPort}
                  onChange={(e) => setSmtpPort(e.target.value)}
                  placeholder="587"
                />
              </div>
              <div className="space-y-2">
                <Label>{t('SMTP Username')}</Label>
                <Input
                  value={smtpUsername}
                  onChange={(e) => setSmtpUsername(e.target.value)}
                  placeholder="user@gmail.com"
                />
              </div>
              <div className="space-y-2">
                <Label>{t('SMTP Password')}</Label>
                <Input
                  type="password"
                  value={smtpPassword}
                  onChange={(e) => setSmtpPassword(e.target.value)}
                  placeholder="••••••••"
                />
              </div>
              <div className="space-y-2">
                <Label>{t('SMTP From Address')}</Label>
                <Input
                  value={smtpFrom}
                  onChange={(e) => setSmtpFrom(e.target.value)}
                  placeholder="noreply@example.com"
                />
              </div>
              <div className="flex flex-col sm:flex-row gap-2">
                <Button
                  onClick={() =>
                    handleSave({
                      SMTPServer: smtpHost,
                      SMTPPort: smtpPort || '587',
                      SMTPAccount: smtpUsername,
                      SMTPToken: smtpPassword,
                      SMTPFrom: smtpFrom,
                    })
                  }
                  disabled={saveMutation.isPending}
                  className="flex-1"
                >
                  <Save className="mr-2 h-4 w-4" />
                  {saveMutation.isPending ? t('Saving...') : t('Save')}
                </Button>
                <Button
                  variant="secondary"
                  onClick={handleTestEmail}
                  disabled={saveMutation.isPending || !smtpHost}
                >
                  <Mail className="mr-2 h-4 w-4" />
                  {t('Test Email')}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ════════════════════════════════════════════════════ OAuth ═══ */}
        <TabsContent value="oauth">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Lock className="h-4 w-4" />
                {t('OAuth Settings')}
              </CardTitle>
              <CardDescription>{t('Configure third-party OAuth providers')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* GitHub */}
              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="font-semibold">GitHub</h3>
                <div className="space-y-2">
                  <Label>{t('Client ID')}</Label>
                  <Input
                    value={githubClientId}
                    onChange={(e) => setGithubClientId(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Client Secret')}</Label>
                  <Input
                    type="password"
                    value={githubClientSecret}
                    onChange={(e) => setGithubClientSecret(e.target.value)}
                  />
                </div>
              </div>

              {/* Google */}
              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="font-semibold">Google</h3>
                <div className="space-y-2">
                  <Label>{t('Client ID')}</Label>
                  <Input
                    value={googleClientId}
                    onChange={(e) => setGoogleClientId(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Client Secret')}</Label>
                  <Input
                    type="password"
                    value={googleClientSecret}
                    onChange={(e) => setGoogleClientSecret(e.target.value)}
                  />
                </div>
              </div>

              {/* Microsoft */}
              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="font-semibold">Microsoft</h3>
                <div className="space-y-2">
                  <Label>{t('Client ID')}</Label>
                  <Input
                    value={msClientId}
                    onChange={(e) => setMsClientId(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Client Secret')}</Label>
                  <Input
                    type="password"
                    value={msClientSecret}
                    onChange={(e) => setMsClientSecret(e.target.value)}
                  />
                </div>
              </div>

              {/* OIDC */}
              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="font-semibold">OIDC</h3>
                <div className="space-y-2">
                  <Label>{t('Provider URL')}</Label>
                  <Input
                    value={oidcProviderUrl}
                    onChange={(e) => setOidcProviderUrl(e.target.value)}
                    placeholder="https://accounts.example.com"
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Client ID')}</Label>
                  <Input
                    value={oidcClientId}
                    onChange={(e) => setOidcClientId(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Client Secret')}</Label>
                  <Input
                    type="password"
                    value={oidcClientSecret}
                    onChange={(e) => setOidcClientSecret(e.target.value)}
                  />
                </div>
              </div>

              {/* Lark */}
              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="font-semibold">Lark / Feishu</h3>
                <div className="space-y-2">
                  <Label>{t('App ID')}</Label>
                  <Input
                    value={larkAppId}
                    onChange={(e) => setLarkAppId(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('App Secret')}</Label>
                  <Input
                    type="password"
                    value={larkAppSecret}
                    onChange={(e) => setLarkAppSecret(e.target.value)}
                  />
                </div>
              </div>

              {/* WeChat */}
              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="font-semibold">WeChat</h3>
                <div className="space-y-2">
                  <Label>{t('App ID')}</Label>
                  <Input
                    value={wechatAppId}
                    onChange={(e) => setWechatAppId(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('App Secret')}</Label>
                  <Input
                    type="password"
                    value={wechatAppSecret}
                    onChange={(e) => setWechatAppSecret(e.target.value)}
                  />
                </div>
              </div>

              {/* Discord */}
              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="font-semibold">Discord</h3>
                <div className="space-y-2">
                  <Label>{t('Client ID')}</Label>
                  <Input
                    value={discordClientId}
                    onChange={(e) => setDiscordClientId(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('Client Secret')}</Label>
                  <Input
                    type="password"
                    value={discordClientSecret}
                    onChange={(e) => setDiscordClientSecret(e.target.value)}
                  />
                </div>
              </div>

              <Button
                onClick={() =>
                  handleSave({
                    GitHubClientId: githubClientId,
                    GitHubClientSecret: githubClientSecret,
                    GoogleClientId: googleClientId,
                    GoogleClientSecret: googleClientSecret,
                    MicrosoftClientId: msClientId,
                    MicrosoftClientSecret: msClientSecret,
                    OidcWellKnown: oidcProviderUrl,
                    OidcClientId: oidcClientId,
                    OidcClientSecret: oidcClientSecret,
                    LarkClientId: larkAppId,
                    LarkClientSecret: larkAppSecret,
                    WeChatServerToken: wechatAppId,
                    WeChatAccountQRCodeImageURL: wechatAppSecret,
                    DiscordClientId: discordClientId,
                    DiscordClientSecret: discordClientSecret,
                  })
                }
                disabled={saveMutation.isPending}
              >
                <Save className="mr-2 h-4 w-4" />
                {saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ════════════════════════════════════════════ Notification ═══ */}
        <TabsContent value="notification">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Bell className="h-4 w-4" />
                {t('Notification Settings')}
              </CardTitle>
              <CardDescription>{t('Configure system notifications and alerts')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-row items-center justify-between rounded-lg border p-3">
                <div>
                  <Label>{t('Email Notifications')}</Label>
                  <p className="text-xs text-muted-foreground">
                    {t('Send notifications via email')}
                  </p>
                </div>
                <Switch
                  checked={emailNotification}
                  onCheckedChange={setEmailNotification}
                />
              </div>
              <div className="space-y-2">
                <Label>{t('Webhook URL')}</Label>
                <Input
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  placeholder="https://hooks.example.com/notify"
                />
              </div>
              <div>
                <Label className="mb-2 block">{t('Notification Events')}</Label>
                <div className="space-y-2 rounded-lg border p-4">
                  <div className="flex flex-row items-center justify-between">
                    <div>
                      <Label>{t('Daily Check-in')}</Label>
                      <p className="text-xs text-muted-foreground">
                        {t('Notify on user daily check-in')}
                      </p>
                    </div>
                    <Switch
                      checked={notifyCheckin}
                      onCheckedChange={setNotifyCheckin}
                    />
                  </div>
                  <div className="flex flex-row items-center justify-between">
                    <div>
                      <Label>{t('Top-up')}</Label>
                      <p className="text-xs text-muted-foreground">
                        {t('Notify on user top-up')}
                      </p>
                    </div>
                    <Switch
                      checked={notifyTopup}
                      onCheckedChange={setNotifyTopup}
                    />
                  </div>
                  <div className="flex flex-row items-center justify-between">
                    <div>
                      <Label>{t('Usage Threshold')}</Label>
                      <p className="text-xs text-muted-foreground">
                        {t('Notify when usage exceeds threshold')}
                      </p>
                    </div>
                    <Switch
                      checked={notifyUsageThreshold}
                      onCheckedChange={setNotifyUsageThreshold}
                    />
                  </div>
                </div>
              </div>
              <Button
                onClick={() =>
                  handleSave({
                    EmailNotificationEnabled: String(emailNotification),
                    WebhookUrl: webhookUrl,
                    NotifyCheckin: String(notifyCheckin),
                    NotifyTopup: String(notifyTopup),
                    NotifyUsageThreshold: String(notifyUsageThreshold),
                  })
                }
                disabled={saveMutation.isPending}
              >
                <Save className="mr-2 h-4 w-4" />
                {saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ════════════════════════════════════════════ Performance ═══ */}
        <TabsContent value="performance">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Zap className="h-4 w-4" />
                {t('Performance Settings')}
              </CardTitle>
              <CardDescription>{t('Configure cache, metrics, and channel testing')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-row items-center justify-between rounded-lg border p-3">
                <div>
                  <Label>{t('Memory Cache')}</Label>
                  <p className="text-xs text-muted-foreground">
                    {t('Enable in-memory caching for faster responses')}
                  </p>
                </div>
                <Switch
                  checked={memoryCacheEnabled}
                  onCheckedChange={setMemoryCacheEnabled}
                />
              </div>
              <div className="space-y-2">
                <Label>{t('Cache Sync Frequency (seconds)')}</Label>
                <Input
                  type="number"
                  value={cacheSyncFrequency}
                  onChange={(e) => setCacheSyncFrequency(e.target.value)}
                  placeholder="600"
                />
              </div>
              <div className="flex flex-row items-center justify-between rounded-lg border p-3">
                <div>
                  <Label>{t('Enable Metrics')}</Label>
                  <p className="text-xs text-muted-foreground">
                    {t('Collect and expose system performance metrics')}
                  </p>
                </div>
                <Switch
                  checked={enableMetric}
                  onCheckedChange={setEnableMetric}
                />
              </div>
              <div className="space-y-2">
                <Label>{t('Channel Test Frequency (seconds)')}</Label>
                <Input
                  type="number"
                  value={channelTestFrequency}
                  onChange={(e) => setChannelTestFrequency(e.target.value)}
                  placeholder="300"
                />
              </div>
              <div className="space-y-2">
                <Label>{t('Batch Test Count')}</Label>
                <Input
                  type="number"
                  value={batchTestCount}
                  onChange={(e) => setBatchTestCount(e.target.value)}
                  placeholder="3"
                />
              </div>
              <Button
                onClick={() =>
                  handleSave({
                    MemoryCacheEnabled: String(memoryCacheEnabled),
                    SyncFrequency: cacheSyncFrequency || '600',
                    EnableMetric: String(enableMetric),
                    ChannelTestFrequency: channelTestFrequency || '300',
                    RetryTimes: batchTestCount || '3',
                  })
                }
                disabled={saveMutation.isPending}
              >
                <Save className="mr-2 h-4 w-4" />
                {saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ════════════════════════════════════════════════ Security ═══ */}
        <TabsContent value="security">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Shield className="h-4 w-4" />
                {t('Security Settings')}
              </CardTitle>
              <CardDescription>{t('Authentication, password policy, and access control')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="font-semibold">{t('Password Policy')}</h3>
                <div className="space-y-2">
                  <Label>{t('Minimum Password Length')}</Label>
                  <Input
                    type="number"
                    value={minPasswordLength}
                    onChange={(e) => setMinPasswordLength(e.target.value)}
                    placeholder="8"
                  />
                </div>
                <div className="flex flex-row items-center justify-between">
                  <div>
                    <Label>{t('Require Special Characters')}</Label>
                    <p className="text-xs text-muted-foreground">
                      {t('Require at least one special character in password')}
                    </p>
                  </div>
                  <Switch
                    checked={requireSpecialChars}
                    onCheckedChange={setRequireSpecialChars}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label>{t('Session Timeout (minutes)')}</Label>
                <Input
                  type="number"
                  value={sessionTimeout}
                  onChange={(e) => setSessionTimeout(e.target.value)}
                  placeholder="120"
                />
              </div>

              <div className="flex flex-row items-center justify-between rounded-lg border p-3">
                <div>
                  <Label>{t('Enforce 2FA')}</Label>
                  <p className="text-xs text-muted-foreground">
                    {t('Require two-factor authentication for all users')}
                  </p>
                </div>
                <Switch
                  checked={enforce2fa}
                  onCheckedChange={setEnforce2fa}
                />
              </div>

              <div className="space-y-2">
                <Label>{t('IP Whitelist')}</Label>
                <textarea
                  className="flex min-h-[100px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                  value={ipWhitelist}
                  onChange={(e) => setIpWhitelist(e.target.value)}
                  placeholder="192.168.1.0/24&#10;10.0.0.1"
                />
                <p className="text-xs text-muted-foreground">
                  {t('One IP or CIDR per line. Whitelist takes precedence.')}
                </p>
              </div>

              <div className="space-y-2">
                <Label>{t('IP Blacklist')}</Label>
                <textarea
                  className="flex min-h-[100px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                  value={ipBlacklist}
                  onChange={(e) => setIpBlacklist(e.target.value)}
                  placeholder="203.0.113.0/24&#10;198.51.100.1"
                />
                <p className="text-xs text-muted-foreground">
                  {t('One IP or CIDR per line.')}
                </p>
              </div>

              <div className="flex flex-row items-center justify-between rounded-lg border p-3">
                <div>
                  <Label>{t('SSRF Protection')}</Label>
                  <p className="text-xs text-muted-foreground">
                    {t('Block server-side request forgery attempts')}
                  </p>
                </div>
                <Switch
                  checked={ssrfProtection}
                  onCheckedChange={setSsrfProtection}
                />
              </div>

              <Button
                onClick={() =>
                  handleSave({
                    MinPasswordLength: minPasswordLength || '8',
                    RequireSpecialChars: String(requireSpecialChars),
                    SessionTimeout: sessionTimeout || '120',
                    Enforce2FA: String(enforce2fa),
                    IPWhitelist: ipWhitelist,
                    IPBlacklist: ipBlacklist,
                    EnableSSRFProtection: String(ssrfProtection),
                  })
                }
                disabled={saveMutation.isPending}
              >
                <Save className="mr-2 h-4 w-4" />
                {saveMutation.isPending ? t('Saving...') : t('Save')}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
