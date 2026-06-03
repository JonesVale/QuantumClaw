import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Save, Shield, Key, Smartphone, Fingerprint, Plus, Trash2, QrCode, Copy, User, Store } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { useAuthStore } from '@/stores/auth-store'
import { updateSelf } from '@/lib/api-extended'
import apiClient from '@/lib/api'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/profile')({
  component: ProfilePage,
})

function ProfilePage() {
  const { t } = useT()
  const { auth } = useAuthStore()
  const user = auth.user
  const queryClient = useQueryClient()

  const [displayName, setDisplayName] = useState(user?.display_name || '')
  const [email, setEmail] = useState(user?.email || '')
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [totpEnabled, setTotpEnabled] = useState(false)
  const [totpSecret, setTotpSecret] = useState('')
  const [totpQrUri, setTotpQrUri] = useState('')
  const [showSetupKey, setShowSetupKey] = useState(false)
  const [enablingTotp, setEnablingTotp] = useState(false)
  const [totpCode, setTotpCode] = useState('')
  const [verifyingTotp, setVerifyingTotp] = useState(false)
  const [backupCodes, setBackupCodes] = useState<string[]>([])
  const [disablingTotp, setDisablingTotp] = useState(false)
  const [disableCode, setDisableCode] = useState('')
  const [passkeys, setPasskeys] = useState<{ id: string; name: string; created_at: string }[]>([])
  const [registeringPasskey, setRegisteringPasskey] = useState(false)

  const { data: twoFAStatus } = useQuery({
    queryKey: ['2fa-status'],
    queryFn: async () => {
      const res = await fetch('/api/user/2fa', { headers: { 'Cache-Control': 'no-store' } })
      if (!res.ok) return { enabled: false }
      return res.json()
    },
    retry: false,
    staleTime: 10 * 1000,
  })

  useQuery({
    queryKey: ['webauthn-credentials'],
    queryFn: async () => {
      const res = await fetch('/api/user/self/webauthn/credentials', { headers: { 'Cache-Control': 'no-store' } })
      if (!res.ok) return []
      const data = await res.json()
      setPasskeys(data?.data || [])
      return data
    },
    retry: false,
    staleTime: 10 * 1000,
  })

  const updateMutation = useMutation({
    mutationFn: updateSelf,
    onSuccess: () => {
      toast.success(t('Profile updated'))
      queryClient.invalidateQueries({ queryKey: ['self'] })
    },
    onError: () => toast.error(t('Failed to update profile')),
  })

  const handleProfileSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateMutation.mutate({
      display_name: displayName,
      email,
    })
  }

  const handlePasswordSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (newPassword !== confirmPassword) {
      toast.error(t('Passwords do not match'))
      return
    }
    updateMutation.mutate({
      password: newPassword,
      old_password: oldPassword,
    })
    setOldPassword('')
    setNewPassword('')
    setConfirmPassword('')
  }

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Profile')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('Manage your account settings')}</p>
      </div>

      <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <User className="h-4 w-4" />
            {t('Personal Information')}
          </CardTitle>
          <CardDescription>{t('Update your display name and email')}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleProfileSubmit} className="space-y-4">
            <div className="flex flex-col sm:flex-row items-center gap-3 sm:gap-4">
              <Avatar className="h-16 w-16">
                <AvatarFallback className="bg-primary/10 text-lg">
                  {user?.display_name?.[0] || user?.username?.[0] || 'U'}
                </AvatarFallback>
              </Avatar>
              <div>
                <p className="font-medium">{user?.display_name || user?.username}</p>
                <p className="text-sm text-muted-foreground">
                  {user?.role >= 10 ? 'Admin' : `ID: ${user?.id}`}
                </p>
              </div>
            </div>
            <div className="space-y-2">
              <Label>{t('Display Name')}</Label>
              <Input
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder={t('Your display name')}
              />
            </div>
            <div className="space-y-2">
              <Label>{t('Email')}</Label>
              <Input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder={t('Email address')}
              />
            </div>
            <Button type="submit" disabled={updateMutation.isPending}>
              <Save className="mr-2 h-4 w-4" />
              {updateMutation.isPending ? t('Saving...') : t('Save')}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle>{t('become_reseller')}</CardTitle>
          <CardDescription>{t('become_reseller_desc')}</CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            variant="outline"
            className="gap-2"
            onClick={async () => {
              try {
                const r = await apiClient.post('/api/user/self/upgrade')
                if (r.data?.success) { toast.success(t('upgrade_success')); window.location.reload() }
                else { toast.error(r.data?.message || t('upgrade_failed')) }
              } catch { toast.error(t('upgrade_failed')) }
            }}
          >
            <Store className="h-4 w-4" />
            {t('become_reseller')}
          </Button>
        </CardContent>
      </Card>

      <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle>{t('Change Password')}</CardTitle>
          <CardDescription>{t('Update your password to keep your account secure')}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handlePasswordSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label>{t('Current Password')}</Label>
              <Input
                type="password"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label>{t('New Password')}</Label>
              <Input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label>{t('Confirm New Password')}</Label>
              <Input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
              />
            </div>
            <Button type="submit" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? t('Updating...') : t('Update Password')}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            {t('Security')}
          </CardTitle>
          <CardDescription>{t('Manage two-factor authentication and passkeys')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Smartphone className="h-4 w-4 text-muted-foreground" />
                <h3 className="font-medium">{t('Two-Factor Authentication')}</h3>
              </div>
              <Badge variant={twoFAStatus?.enabled ? 'default' : 'secondary'}>
                {twoFAStatus?.enabled ? t('Enabled') : t('Disabled')}
              </Badge>
            </div>
            <p className="text-sm text-muted-foreground">
              {t('Add an extra layer of security to your account using TOTP')}
            </p>
            <Button
              variant={twoFAStatus?.enabled ? 'destructive' : 'default'}
              size="sm"
              onClick={async () => {
                if (twoFAStatus?.enabled) {
                  if (disablingTotp) {
                    setDisablingTotp(false)
                    setDisableCode('')
                  } else {
                    setDisablingTotp(true)
                  }
                } else {
                  setEnablingTotp(true)
                  setShowSetupKey(false)
                  setBackupCodes([])
                  setTotpCode('')
                  const res = await fetch('/api/user/2fa/init', { method: 'POST' })
                  const data = await res.json()
                  if (data.success) {
                    setTotpSecret(data.data?.secret || '')
                    setTotpQrUri(data.data?.qr_code_url || '')
                    setShowSetupKey(true)
                    toast.success(t('Scan the QR code with your authenticator app'))
                  } else {
                    toast.error(data.message || t('Failed to initialize 2FA'))
                  }
                  setEnablingTotp(false)
                }
              }}
              disabled={enablingTotp}
            >
              {enablingTotp ? t('Setting up...') : twoFAStatus?.enabled ? t('Disable') : t('Enable')}
            </Button>
            {disablingTotp && (
              <div className="mt-3 p-3 rounded-xl border border-red-200 bg-red-50 dark:bg-red-950/20 space-y-2">
                <p className="text-sm font-medium text-red-700 dark:text-red-300">
                  {t('Enter your TOTP code to disable 2FA')}
                </p>
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
                      const res = await fetch('/api/user/2fa/disable', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ code: disableCode }),
                      })
                      const data = await res.json()
                      if (data.success) {
                        toast.success(t('2FA disabled'))
                        setDisablingTotp(false)
                        setDisableCode('')
                        queryClient.invalidateQueries({ queryKey: ['2fa-status'] })
                      } else {
                        toast.error(data.message || t('Failed to disable 2FA'))
                      }
                    }}
                  >
                    {t('Confirm Disable')}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => { setDisablingTotp(false); setDisableCode('') }}
                  >
                    {t('Cancel')}
                  </Button>
                </div>
              </div>
            )}
            {showSetupKey && (
              <div className="mt-3 p-3 rounded-xl border bg-muted/50 space-y-2">
                <div className="flex items-center gap-2">
                  <QrCode className="h-4 w-4" />
                  <span className="text-sm font-medium">{t('Setup Key')}</span>
                </div>
                {totpQrUri && (
                  <div className="flex justify-center p-2 bg-white dark:bg-black rounded">
                    <img
                      src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(totpQrUri)}`}
                      alt="TOTP QR Code"
                      className="h-48 w-48"
                    />
                  </div>
                )}
                {totpSecret && (
                  <div className="flex items-center gap-2">
                    <code className="flex-1 p-1.5 text-xs rounded border bg-background font-mono break-all">
                      {totpSecret}
                    </code>
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-7 w-7 shrink-0"
                      onClick={() => {
                        navigator.clipboard.writeText(totpSecret)
                        toast.success(t('Copied'))
                      }}
                    >
                      <Copy className="h-3 w-3" />
                    </Button>
                  </div>
                )}
                <div className="space-y-2 mt-3">
                  <Label htmlFor="totp-code">{t('Enter 6-digit TOTP code')}</Label>
                  <div className="flex gap-2">
                    <Input
                      id="totp-code"
                      type="text"
                      inputMode="numeric"
                      maxLength={6}
                      placeholder="123456"
                      value={totpCode}
                      onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                    />
                    <Button
                      size="sm"
                      disabled={totpCode.length !== 6 || verifyingTotp}
                      onClick={async () => {
                        setVerifyingTotp(true)
                        const res = await fetch('/api/user/2fa/enable', {
                          method: 'POST',
                          headers: { 'Content-Type': 'application/json' },
                          body: JSON.stringify({ code: totpCode }),
                        })
                        const data = await res.json()
                        if (data.success) {
                          setBackupCodes(data.data?.backup_codes || [])
                          setShowSetupKey(false)
                          toast.success(t('2FA enabled'))
                          queryClient.invalidateQueries({ queryKey: ['2fa-status'] })
                        } else {
                          toast.error(data.message || t('Invalid code'))
                        }
                        setVerifyingTotp(false)
                      }}
                    >
                      {verifyingTotp ? t('Verifying...') : t('Verify & Enable')}
                    </Button>
                  </div>
                </div>
                <p className="text-xs text-muted-foreground">
                  {t('Scan QR code, enter the 6-digit code, then click Verify & Enable')}
                </p>
              </div>
            )}
            {backupCodes.length > 0 && (
              <div className="mt-3 p-3 rounded-xl border border-green-200 bg-green-50 dark:bg-green-950/20 space-y-2">
                <p className="text-sm font-medium text-green-700 dark:text-green-300">
                  {t('2FA Enabled! Backup codes:')}
                </p>
                <div className="grid grid-cols-2 gap-1">
                  {backupCodes.map((code, i) => (
                    <code key={i} className="p-1 text-xs rounded border bg-background font-mono">{code}</code>
                  ))}
                </div>
                <p className="text-xs text-muted-foreground">
                  {t('Each code works once. Save them securely.')}
                </p>
              </div>
            )}
          </div>

          <Separator />

          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Fingerprint className="h-4 w-4 text-muted-foreground" />
              <h3 className="font-medium">{t('Passkeys (WebAuthn)')}</h3>
            </div>
            <p className="text-sm text-muted-foreground">
              {t('Use biometrics, security keys, or platform authenticators to sign in')}
            </p>
            <Button
              variant="outline"
              size="sm"
              disabled={registeringPasskey}
              onClick={async () => {
                setRegisteringPasskey(true)
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
                    toast.success(t('Passkey registered'))
                    queryClient.invalidateQueries({ queryKey: ['webauthn-credentials'] })
                  } else {
                    toast.error(finishData.message || t('Failed to register passkey'))
                  }
                } catch (err: any) {
                  toast.error(err?.message || t('Failed to register passkey'))
                } finally {
                  setRegisteringPasskey(false)
                }
              }}
            >
              <Plus className="mr-1 h-3 w-3" />
              {registeringPasskey ? t('Registering...') : t('Register Passkey')}
            </Button>
            {passkeys.length > 0 && (
              <div className="space-y-2 mt-3">
                {passkeys.map((pk) => (
                  <div
                    key={pk.id}
                    className="flex items-center justify-between p-2 rounded-xl border gap-2"
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <Key className="h-4 w-4 text-muted-foreground shrink-0" />
                      <span className="text-sm truncate">{pk.name}</span>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 shrink-0 text-destructive hover:text-destructive"
                      onClick={async () => {
                        const res = await fetch(`/api/user/self/webauthn/credentials/${pk.id}`, {
                          method: 'DELETE',
                        })
                        const data = await res.json()
                        if (data.success) {
                          toast.success(t('Passkey removed'))
                          queryClient.invalidateQueries({ queryKey: ['webauthn-credentials'] })
                        } else {
                          toast.error(data.message || t('Failed to remove passkey'))
                        }
                      }}
                    >
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
            {passkeys.length === 0 && (
              <p className="text-sm text-muted-foreground italic">
                {t('No passkeys registered yet')}
              </p>
            )}
          </div>
        </CardContent>
      </Card>

      <Card className="w-full bg-white/80 backdrop-blur-xl rounded-xl border">
        <CardHeader>
          <CardTitle>{t('Account Details')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div className="flex flex-col sm:flex-row justify-between gap-1 sm:gap-4">
            <span className="text-muted-foreground">{t('Username')}</span>
            <span className="font-mono">{user?.username}</span>
          </div>
          <div className="flex flex-col sm:flex-row justify-between gap-1 sm:gap-4">
            <span className="text-muted-foreground">{t('Group')}</span>
            <span>{user?.group || 'default'}</span>
          </div>
          <div className="flex flex-col sm:flex-row justify-between gap-1 sm:gap-4">
            <span className="text-muted-foreground">{t('Used Quota')}</span>
            <span>{user?.used_quota?.toLocaleString() || 0}</span>
          </div>
          <div className="flex flex-col sm:flex-row justify-between gap-1 sm:gap-4">
            <span className="text-muted-foreground">{t('Remaining Quota')}</span>
            <span>{user?.quota?.toLocaleString() || 0}</span>
          </div>
          <div className="flex flex-col sm:flex-row justify-between gap-1 sm:gap-4">
            <span className="text-muted-foreground">{t('Request Count')}</span>
            <span>{user?.request_count?.toLocaleString() || 0}</span>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
