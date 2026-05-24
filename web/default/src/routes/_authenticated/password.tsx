import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import {
  KeyRound, Shield, Lock, Mail, RefreshCw,
  ArrowLeft, Eye, EyeOff, Loader2, CheckCircle2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import apiClient from '@/lib/api'
import { toast } from 'sonner'

// ---------------------------------------------------------------------------
// Route
// ---------------------------------------------------------------------------

export const Route = createFileRoute('/_authenticated/password')({
  component: PasswordPage,
})

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface ChangePasswordPayload {
  old_password: string
  new_password: string
  confirm_password: string
}

interface EmergencyResetPayload {
  email: string
  token: string
  new_password: string
}

interface PasswordResponse {
  success: boolean
  message?: string
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

function PasswordPage() {
  const { t } = useT()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState('change')

  // --- Change password state ---
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showOld, setShowOld] = useState(false)
  const [showNew, setShowNew] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)

  // --- Emergency reset state ---
  const [resetEmail, setResetEmail] = useState('')
  const [resetToken, setResetToken] = useState('')
  const [resetNewPassword, setResetNewPassword] = useState('')
  const [resetConfirmPassword, setResetConfirmPassword] = useState('')
  const [showResetPassword, setShowResetPassword] = useState(false)
  const [showResetConfirm, setShowResetConfirm] = useState(false)

  // --- Validation helpers ---

  const passwordStrength = (pw: string): { label: string; color: string; score: number } => {
    if (pw.length === 0) return { label: '', color: '', score: 0 }
    let score = 0
    if (pw.length >= 8) score++
    if (pw.length >= 12) score++
    if (/[A-Z]/.test(pw)) score++
    if (/[a-z]/.test(pw)) score++
    if (/[0-9]/.test(pw)) score++
    if (/[^A-Za-z0-9]/.test(pw)) score++
    if (score <= 2) return { label: t('Weak'), color: 'bg-red-500', score }
    if (score <= 4) return { label: t('Medium'), color: 'bg-amber-500', score }
    return { label: t('Strong'), color: 'bg-green-500', score }
  }

  const strength = passwordStrength(newPassword)

  const changePasswordValid =
    oldPassword.length >= 1 &&
    newPassword.length >= 6 &&
    newPassword === confirmPassword

  const resetValid =
    resetEmail.includes('@') &&
    resetToken.length >= 1 &&
    resetNewPassword.length >= 6 &&
    resetNewPassword === resetConfirmPassword

  // --- Mutations ---

  const changeMutation = useMutation({
    mutationFn: async (payload: ChangePasswordPayload) => {
      const res = await apiClient.post('/api/user/self/password/change', payload)
      return res.data as PasswordResponse
    },
    onSuccess: (data) => {
      if (data.success) {
        toast.success(t('Password changed successfully'))
        setOldPassword('')
        setNewPassword('')
        setConfirmPassword('')
      } else {
        toast.error(data.message || t('Failed to change password'))
      }
    },
    onError: (err: any) => {
      const msg = err?.response?.data?.message || err?.message || t('Failed to change password')
      toast.error(msg)
    },
  })

  const resetMutation = useMutation({
    mutationFn: async (payload: EmergencyResetPayload) => {
      const res = await apiClient.post('/api/password/emergency-reset', payload)
      return res.data as PasswordResponse
    },
    onSuccess: (data) => {
      if (data.success) {
        toast.success(t('Password reset successfully'))
        toast.info(t('Please sign in with your new password'))
        setTimeout(() => {
          navigate({ to: '/sign-in' })
        }, 1500)
      } else {
        toast.error(data.message || t('Failed to reset password'))
      }
    },
    onError: (err: any) => {
      const msg = err?.response?.data?.message || err?.message || t('Failed to reset password')
      toast.error(msg)
    },
  })

  // --- Handlers ---

  const handleChangePassword = (e: React.FormEvent) => {
    e.preventDefault()
    if (newPassword !== confirmPassword) {
      toast.error(t('New passwords do not match'))
      return
    }
    if (newPassword.length < 6) {
      toast.error(t('Password must be at least 6 characters'))
      return
    }
    changeMutation.mutate({
      old_password: oldPassword,
      new_password: newPassword,
      confirm_password: confirmPassword,
    })
  }

  const handleEmergencyReset = (e: React.FormEvent) => {
    e.preventDefault()
    if (resetNewPassword !== resetConfirmPassword) {
      toast.error(t('New passwords do not match'))
      return
    }
    if (resetNewPassword.length < 6) {
      toast.error(t('Password must be at least 6 characters'))
      return
    }
    resetMutation.mutate({
      email: resetEmail,
      token: resetToken,
      new_password: resetNewPassword,
    })
  }

  // --- Render ---

  return (
    <div className="qc-wrapper py-8 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">{t('Password')}</h1>
        <p className="text-muted-foreground mt-2">
          {t('Manage your password and account security')}
        </p>
      </div>

      <Card className="w-full max-w-xl mx-auto">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <KeyRound className="h-4 w-4" />
            {t('Password Management')}
          </CardTitle>
          <CardDescription>
            {t('Change your current password or perform an emergency reset')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="grid w-full grid-cols-2 mb-6">
              <TabsTrigger value="change" className="flex items-center gap-2">
                <Lock className="h-3.5 w-3.5" />
                {t('Change Password')}
              </TabsTrigger>
              <TabsTrigger value="reset" className="flex items-center gap-2">
                <RefreshCw className="h-3.5 w-3.5" />
                {t('Forgot Password')}
              </TabsTrigger>
            </TabsList>

            {/* --- Tab: Change Password --- */}
            <TabsContent value="change">
              <form onSubmit={handleChangePassword} className="space-y-5">
                <div className="space-y-2">
                  <Label htmlFor="old-password">{t('Current Password')}</Label>
                  <div className="relative">
                    <Input
                      id="old-password"
                      type={showOld ? 'text' : 'password'}
                      value={oldPassword}
                      onChange={(e) => setOldPassword(e.target.value)}
                      placeholder={t('Enter your current password')}
                      required
                      className="pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowOld(!showOld)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                    >
                      {showOld ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                </div>

                <Separator />

                <div className="space-y-2">
                  <Label htmlFor="new-password">{t('New Password')}</Label>
                  <div className="relative">
                    <Input
                      id="new-password"
                      type={showNew ? 'text' : 'password'}
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      placeholder={t('Enter your new password')}
                      required
                      minLength={6}
                      className="pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowNew(!showNew)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                    >
                      {showNew ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                  {/* Strength indicator */}
                  {newPassword.length > 0 && (
                    <div className="mt-2 space-y-1">
                      <div className="flex gap-1">
                        {[1, 2, 3, 4, 5, 6].map((i) => (
                          <div
                            key={i}
                            className={cn(
                              'h-1 flex-1 rounded-full transition-colors',
                              i <= Math.min(strength.score, 6)
                                ? strength.color
                                : 'bg-muted'
                            )}
                          />
                        ))}
                      </div>
                      <p className="text-xs text-muted-foreground">
                        {strength.label}
                        <span className="ml-2">
                          {newPassword.length < 8
                            ? t('At least 8 characters recommended')
                            : t('Good length')}
                        </span>
                      </p>
                    </div>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="confirm-password">{t('Confirm New Password')}</Label>
                  <div className="relative">
                    <Input
                      id="confirm-password"
                      type={showConfirm ? 'text' : 'password'}
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      placeholder={t('Re-enter your new password')}
                      required
                      className="pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowConfirm(!showConfirm)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                    >
                      {showConfirm ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                  {confirmPassword.length > 0 && newPassword !== confirmPassword && (
                    <p className="text-xs text-red-500 mt-1">{t('Passwords do not match')}</p>
                  )}
                  {confirmPassword.length > 0 && newPassword === confirmPassword && (
                    <p className="text-xs text-green-500 mt-1 flex items-center gap-1">
                      <CheckCircle2 className="h-3 w-3" />
                      {t('Passwords match')}
                    </p>
                  )}
                </div>

                <Button
                  type="submit"
                  className="w-full bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-sm hover:shadow-md"
                  disabled={!changePasswordValid || changeMutation.isPending}
                >
                  {changeMutation.isPending ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      {t('Changing...')}
                    </>
                  ) : (
                    <>
                      <Shield className="mr-2 h-4 w-4" />
                      {t('Change Password')}
                    </>
                  )}
                </Button>
              </form>

              <div className="mt-6 p-4 rounded-xl bg-amber-50 dark:bg-amber-950/10 border border-amber-200 dark:border-amber-800/30">
                <div className="flex items-start gap-3">
                  <Shield className="h-5 w-5 text-amber-600 shrink-0 mt-0.5" />
                  <div>
                    <p className="text-sm font-medium text-amber-800 dark:text-amber-300">
                      {t('Password Tips')}
                    </p>
                    <ul className="text-xs text-amber-700 dark:text-amber-400 mt-1 space-y-1 list-disc list-inside">
                      <li>{t('Use at least 8 characters')}</li>
                      <li>{t('Mix uppercase, lowercase, numbers, and symbols')}</li>
                      <li>{t('Avoid common words or personal information')}</li>
                      <li>{t('Use a different password than other services')}</li>
                    </ul>
                  </div>
                </div>
              </div>
            </TabsContent>

            {/* --- Tab: Forgot / Emergency Reset --- */}
            <TabsContent value="reset">
              <div className="mb-4 p-3 rounded-xl bg-blue-50 dark:bg-blue-950/10 border border-blue-200 dark:border-blue-800/30">
                <p className="text-xs text-blue-700 dark:text-blue-400">
                  {t('Use this option if you have lost access to your account. You will need your email and a reset token.')}
                </p>
              </div>

              <form onSubmit={handleEmergencyReset} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="reset-email">{t('Email')}</Label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/50" />
                    <Input
                      id="reset-email"
                      type="email"
                      value={resetEmail}
                      onChange={(e) => setResetEmail(e.target.value)}
                      placeholder={t('Enter your account email')}
                      required
                      className="pl-10"
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="reset-token">{t('Reset Token')}</Label>
                  <div className="relative">
                    <KeyRound className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/50" />
                    <Input
                      id="reset-token"
                      value={resetToken}
                      onChange={(e) => setResetToken(e.target.value)}
                      placeholder={t('Enter your reset token')}
                      required
                      className="pl-10"
                    />
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {t('Contact support if you don\'t have a reset token')}
                  </p>
                </div>

                <Separator />

                <div className="space-y-2">
                  <Label htmlFor="reset-new-password">{t('New Password')}</Label>
                  <div className="relative">
                    <Input
                      id="reset-new-password"
                      type={showResetPassword ? 'text' : 'password'}
                      value={resetNewPassword}
                      onChange={(e) => setResetNewPassword(e.target.value)}
                      placeholder={t('Enter new password')}
                      required
                      minLength={6}
                      className="pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowResetPassword(!showResetPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                    >
                      {showResetPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="reset-confirm-password">{t('Confirm New Password')}</Label>
                  <div className="relative">
                    <Input
                      id="reset-confirm-password"
                      type={showResetConfirm ? 'text' : 'password'}
                      value={resetConfirmPassword}
                      onChange={(e) => setResetConfirmPassword(e.target.value)}
                      placeholder={t('Re-enter new password')}
                      required
                      className="pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowResetConfirm(!showResetConfirm)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                    >
                      {showResetConfirm ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                  {resetConfirmPassword.length > 0 && resetNewPassword !== resetConfirmPassword && (
                    <p className="text-xs text-red-500 mt-1">{t('Passwords do not match')}</p>
                  )}
                </div>

                <Button
                  type="submit"
                  className="w-full"
                  variant="destructive"
                  disabled={!resetValid || resetMutation.isPending}
                >
                  {resetMutation.isPending ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      {t('Resetting...')}
                    </>
                  ) : (
                    <>
                      <RefreshCw className="mr-2 h-4 w-4" />
                      {t('Reset Password')}
                    </>
                  )}
                </Button>
              </form>

              <div className="mt-6 text-center">
                <Button
                  variant="link"
                  size="sm"
                  className="text-muted-foreground"
                  onClick={() => navigate({ to: '/sign-in' })}
                >
                  <ArrowLeft className="mr-1 h-3 w-3" />
                  {t('Back to sign in')}
                </Button>
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}
