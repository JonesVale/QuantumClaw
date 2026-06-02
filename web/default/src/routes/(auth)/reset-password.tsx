import { createFileRoute, useNavigate, useSearch } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState } from 'react'
import { toast } from 'sonner'

export const Route = createFileRoute('/(auth)/reset-password')({
  component: ResetPasswordPage,
})

function ResetPasswordPage() {
  const { t } = useT()
  const search = useSearch({ from: '/(auth)/reset-password' }) as { email?: string; token?: string }
  const navigate = useNavigate()
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  const doReset = async () => {
    if (!password || password.length < 6) {
      setError(t('Password must be at least 6 characters'))
      return
    }
    if (password !== confirmPassword) {
      setError(t('Passwords do not match'))
      return
    }
    if (!search.email || !search.token) {
      setError(t('Invalid or expired reset link'))
      return
    }

    setLoading(true)
    setError('')
    try {
      const res = await fetch('/api/user/reset', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: search.email,
          token: search.token,
          new_password: password,
        }),
      }).then(r => r.json())

      if (res.success) {
        setSuccess(true)
        toast.success(t('Password reset successfully'))
      } else {
        setError(res.message || t('Failed to reset password'))
      }
    } catch {
      setError(t('Network error'))
    }
    setLoading(false)
  }

  return (
    <div className="min-h-screen flex items-center justify-center relative"
      style={{
        backgroundImage: 'url(/login-backup.png)',
        backgroundSize: '100% 100%',
        backgroundPosition: 'center',
        backgroundRepeat: 'no-repeat',
      }}>
      <div className="qc-fade-up w-full max-w-md mx-auto px-6">
        <div className="rounded-2xl bg-gray-900/95 border border-gray-700 shadow-2xl shadow-black/40 p-8">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-white">{t('Reset Password')}</h1>
            <p className="text-amber-200/80 text-sm mt-2">{t('Enter your new password')}</p>
          </div>

          {!search.email || !search.token ? (
            <div className="px-4 py-6 rounded-xl bg-red-900/80 text-red-200 text-sm border border-red-700 text-center">
              {t('Invalid or expired reset link. Please request a new password reset from the login page.')}
              <button
                onClick={() => navigate({ to: '/sign-in' })}
                className="mt-4 block mx-auto text-amber-400 hover:text-amber-300 underline"
              >
                {t('Back to login')}
              </button>
            </div>
          ) : success ? (
            <div className="space-y-6">
              <div className="px-4 py-6 rounded-xl bg-green-900/80 text-green-200 text-sm border border-green-700 text-center">
                {t('Password has been reset successfully!')}
              </div>
              <button
                onClick={() => navigate({ to: '/sign-in' })}
                className="w-full py-5 rounded-xl text-xl font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg transition-all"
              >
                {t('Login with new password')}
              </button>
            </div>
          ) : (
            <div className="space-y-5">
              {error && (
                <div className="px-4 py-3 rounded-xl bg-red-900/80 text-red-200 text-sm border border-red-700">
                  {error}
                </div>
              )}

              <div>
                <label className="text-lg font-medium text-amber-200 block mb-3">{t('New Password')}</label>
                <input
                  type="password"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  className="w-full h-14 rounded-xl border border-gray-600 bg-gray-800 px-5 text-base text-white outline-none placeholder:text-gray-500 focus:border-amber-500 focus:ring-2 focus:ring-amber-400/30 transition-all"
                  placeholder="&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;"
                  autoComplete="new-password"
                />
              </div>

              <div>
                <label className="text-lg font-medium text-amber-200 block mb-3">{t('Confirm New Password')}</label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={e => setConfirmPassword(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && doReset()}
                  className="w-full h-14 rounded-xl border border-gray-600 bg-gray-800 px-5 text-base text-white outline-none placeholder:text-gray-500 focus:border-amber-500 focus:ring-2 focus:ring-amber-400/30 transition-all"
                  placeholder="&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;"
                  autoComplete="new-password"
                />
              </div>

              <button
                onClick={doReset}
                disabled={loading || !password || password !== confirmPassword}
                className="w-full py-5 rounded-xl text-xl font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2"
              >
                {loading ? (
                  <div className="w-5 h-5 rounded-full border-2 border-white/30 border-t-white animate-spin" />
                ) : (
                  t('Set New Password')
                )}
              </button>

              <div className="text-center">
                <button
                  type="button"
                  onClick={() => navigate({ to: '/sign-in' })}
                  className="text-sm text-amber-400 hover:text-amber-300 transition-colors"
                >
                  {t('Back to login')}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
