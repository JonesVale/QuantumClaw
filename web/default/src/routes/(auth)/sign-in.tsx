import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useCallback } from 'react'
import { useAuthStore } from '@/stores/auth-store'
import { setCookie } from '@/lib/cookies'

export const Route = createFileRoute('/(auth)/sign-in')({
  component: SignInPage,
})

function SignInPage() {
  const { t } = useT()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const navigate = useNavigate()
  const auth = useAuthStore(s => s.auth)

  const doLogin = useCallback(async () => {
    if (!email || !password) return
    setLoading(true); setError('')
    try {
      const res = await fetch('/api/auth/login', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({email,password}) }).then(r=>r.json())
      if (res.success && res.data?.token) {
        setCookie(import.meta.env.VITE_TOKEN_KEY || 'token', res.data.token)
        auth.setUser(res.data.user)
        navigate({ to: '/dashboard' })
      } else {
        setError(res.message || t('Login failed'))
      }
    } catch { setError(t('Network error')) }
    setLoading(false)
  }, [email, password, t, navigate, auth])

  return (
    <div className="min-h-screen bg-background flex items-center justify-center"
      style={{ backgroundImage: 'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)' }}>
      <div className="qc-fade-up w-full max-w-sm mx-auto px-4">
        <div className="text-center mb-8">
          <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center shadow-md shadow-orange-500/20 mx-auto mb-4">
            <svg className="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground mb-1">{t('Sign In')}</h1>
          <p className="text-sm text-muted-foreground/60">{t('Access your QuantumClaw dashboard')}</p>
        </div>

        <div className="rounded-2xl bg-white/80 backdrop-blur-sm border border-border/20 shadow-sm p-6">
          {error && (
            <div className="mb-4 px-4 py-2.5 rounded-xl bg-red-50 text-red-700 text-xs font-medium border border-red-200/50">
              {error}
            </div>
          )}
          <div className="space-y-4">
            <div>
              <label className="text-xs font-medium text-muted-foreground/60 block mb-1.5">{t('Email')}</label>
              <input type="email" value={email} onChange={e => setEmail(e.target.value)}
                className="w-full h-10 rounded-xl border border-border/30 bg-white px-4 text-sm outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all"
                placeholder="you@example.com" autoComplete="email" />
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground/60 block mb-1.5">{t('Password')}</label>
              <input type="password" value={password} onChange={e => setPassword(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && doLogin()}
                className="w-full h-10 rounded-xl border border-border/30 bg-white px-4 text-sm outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all"
                placeholder="••••••••" autoComplete="current-password" />
            </div>
            <button onClick={doLogin} disabled={loading || !email || !password}
              className="w-full py-2.5 rounded-xl text-sm font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2">
              {loading ? <><div className="w-4 h-4 rounded-full border-2 border-white/30 border-t-white animate-spin" /></> : t('Sign In')}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
