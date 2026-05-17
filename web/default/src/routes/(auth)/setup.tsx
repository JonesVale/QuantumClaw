import { createFileRoute, useRouter, redirect } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { toast } from 'sonner'
import { Check, Loader2, Shield, Settings, Wrench } from 'lucide-react'
import apiClient from '@/lib/api'

export const Route = createFileRoute('/(auth)/setup')({
  component: SetupPage,
  beforeLoad: async () => {
    // Check if setup is actually needed
    try {
      const res = await apiClient.get('/api/setup/check', {
        skipErrorHandler: true,
      } as Record<string, unknown>)
      const data = res.data as { success: boolean; data: { setup_needed: boolean } }
      if (data.success && !data.data.setup_needed) {
        throw redirect({ to: '/dashboard' })
      }
    } catch {
      // If the endpoint fails (might not exist yet), silently continue
    }
  },
})

function SetupPage() {
  const { t } = useTranslation()
  const router = useRouter()
  const [step, setStep] = useState(0)
  const [loading, setLoading] = useState(false)
  const [checking, setChecking] = useState(true)
  const [setupNeeded, setSetupNeeded] = useState(true)

  // Step 1: Account
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [email, setEmail] = useState('')

  // Step 2: Settings
  const [siteName, setSiteName] = useState('QuantumClaw')
  const [registerEnabled, setRegisterEnabled] = useState(true)
  const [passwordRegister, setPasswordRegister] = useState(true)

  // Check setup status on mount
  useEffect(() => {
    apiClient
      .get('/api/setup/check', {
        skipErrorHandler: true,
      } as Record<string, unknown>)
      .then((res) => {
        const data = res.data as { success: boolean; data: { setup_needed: boolean } }
        if (data.success) {
          setSetupNeeded(data.data.setup_needed)
          if (!data.data.setup_needed) {
            router.navigate({ to: '/dashboard' })
          }
        }
      })
      .catch(() => {
        // Assume setup is needed if endpoint fails
      })
      .finally(() => setChecking(false))
  }, [router])

  const canProceedStep1 =
    username.length >= 3 &&
    password.length >= 8 &&
    password === confirmPassword &&
    email.includes('@')

  const handleNext = () => {
    if (step === 0 && !canProceedStep1) return
    setStep((s) => Math.min(s + 1, 2))
  }

  const handleBack = () => {
    setStep((s) => Math.max(s - 1, 0))
  }

  const handleComplete = async () => {
    setLoading(true)
    try {
      const res = await apiClient.post('/api/setup/complete', {
        username,
        password,
        email,
        site_name: siteName,
        register_enabled: registerEnabled,
        password_register_enabled: passwordRegister,
      })
      const data = res.data as { success: boolean; message?: string }
      if (data.success) {
        toast.success(t('Setup completed! Redirecting...'))
        // Set credentials in localStorage and redirect to sign-in
        setTimeout(() => {
          router.navigate({ to: '/sign-in' })
        }, 1500)
      } else {
        toast.error(data.message || t('Setup failed'))
      }
    } catch (err: unknown) {
      const msg =
        err && typeof err === 'object' && 'response' in err
          ? ((err as { response: { data: { message?: string } } }).response?.data?.message ?? t('Setup failed'))
          : t('Setup failed')
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  if (checking) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 via-white to-blue-50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    )
  }

  if (!setupNeeded) {
    return null
  }

  const steps = [
    { icon: Shield, label: t('Admin Account') },
    { icon: Settings, label: t('Basic Settings') },
    { icon: Check, label: t('Done') },
  ]

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 via-white to-blue-50 p-4 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950">
      <Card className="w-full max-w-lg">
        <CardHeader className="items-center text-center">
          <div className="mb-2 flex h-14 w-14 items-center justify-center rounded-xl bg-gradient-to-br from-blue-600 to-purple-600">
            <Wrench className="h-7 w-7 text-white" />
          </div>
          <CardTitle className="text-2xl font-bold">{t('Welcome to QuantumClaw')}</CardTitle>
          <CardDescription>{t("Let's set up your platform in a few steps")}</CardDescription>
        </CardHeader>

        {/* Step indicators */}
        <div className="px-6">
          <div className="flex items-center justify-center gap-2">
            {steps.map((s, i) => (
              <div key={i} className="flex items-center gap-2">
                <div
                  className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold transition-all ${
                    i === step
                      ? 'bg-blue-600 text-white shadow-md'
                      : i < step
                        ? 'bg-green-500 text-white'
                        : 'bg-slate-200 text-slate-500 dark:bg-slate-700 dark:text-slate-400'
                  }`}
                >
                  {i < step ? <Check className="h-4 w-4" /> : <s.icon className="h-4 w-4" />}
                </div>
                {i < steps.length - 1 && (
                  <div
                    className={`h-0.5 w-12 transition-colors ${
                      i < step ? 'bg-green-500' : 'bg-slate-200 dark:bg-slate-700'
                    }`}
                  />
                )}
              </div>
            ))}
          </div>
          <p className="mt-2 text-center text-sm text-slate-500">
            {steps[step]?.label}
          </p>
        </div>

        <CardContent className="pt-6">
          {step === 0 && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="username">{t('Admin Username')}</Label>
                <Input
                  id="username"
                  placeholder={t('Enter admin username')}
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                  minLength={3}
                  maxLength={12}
                />
                <p className="text-xs text-slate-400">{t('3-12 characters')}</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">{t('Admin Email')}</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder={t('admin@example.com')}
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
              <Separator />
              <div className="space-y-2">
                <Label htmlFor="password">{t('Password')}</Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  minLength={8}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirm-password">{t('Confirm Password')}</Label>
                <Input
                  id="confirm-password"
                  type="password"
                  placeholder="••••••••"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                />
                {confirmPassword && password !== confirmPassword && (
                  <p className="text-xs text-red-500">{t('Passwords do not match')}</p>
                )}
              </div>
            </div>
          )}

          {step === 1 && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="site-name">{t('Site Name')}</Label>
                <Input
                  id="site-name"
                  placeholder={t('QuantumClaw')}
                  value={siteName}
                  onChange={(e) => setSiteName(e.target.value)}
                />
                <p className="text-xs text-slate-400">{t('The name displayed on your platform')}</p>
              </div>
              <Separator />
              <div className="flex items-center justify-between">
                <Label htmlFor="register-enabled" className="cursor-pointer font-normal">
                  {t('Allow user registration')}
                </Label>
                <Switch
                  id="register-enabled"
                  checked={registerEnabled}
                  onCheckedChange={setRegisterEnabled}
                />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="password-register" className="cursor-pointer font-normal">
                  {t('Allow password-based registration')}
                </Label>
                <Switch
                  id="password-register"
                  checked={passwordRegister}
                  onCheckedChange={setPasswordRegister}
                />
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4 text-center">
              <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-900">
                <Check className="h-8 w-8 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <h3 className="text-lg font-semibold">{t('All Set!')}</h3>
                <p className="text-sm text-slate-500 dark:text-slate-400">
                  {t('Your QuantumClaw platform is ready to use. Click "Finish" to complete setup and sign in.')}
                </p>
              </div>
              <div className="rounded-lg bg-slate-50 p-4 text-left dark:bg-slate-800">
                <p className="text-sm font-medium">{t('Summary')}</p>
                <ul className="mt-2 space-y-1 text-sm text-slate-600 dark:text-slate-400">
                  <li>
                    <strong>{t('Admin:')}</strong> {username}
                  </li>
                  <li>
                    <strong>{t('Email:')}</strong> {email}
                  </li>
                  <li>
                    <strong>{t('Site:')}</strong> {siteName}
                  </li>
                  <li>
                    <strong>{t('Registration:')}</strong>{' '}
                    {registerEnabled ? t('Enabled') : t('Disabled')}
                  </li>
                </ul>
              </div>
            </div>
          )}
        </CardContent>

        <CardFooter className="flex justify-between border-t px-6 py-4">
          {step > 0 ? (
            <Button variant="outline" onClick={handleBack} disabled={loading}>
              {t('Back')}
            </Button>
          ) : (
            <div />
          )}
          {step < 2 ? (
            <Button onClick={handleNext} disabled={step === 0 && !canProceedStep1}>
              {t('Next')}
            </Button>
          ) : (
            <Button onClick={handleComplete} disabled={loading}>
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  {t('Finishing...')}
                </>
              ) : (
                t('Finish')
              )}
            </Button>
          )}
        </CardFooter>
      </Card>
    </div>
  )
}
