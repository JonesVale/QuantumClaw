import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useCallback, useEffect } from 'react'
import { useAuthStore } from '@/stores/auth-store'

export const Route = createFileRoute('/(auth)/sign-in')({
  component: SignInPage,
})

interface OAuthProvider {
  id: string
  name: string
  loginUrl: string
  enabled: boolean
}

function SignInPage() {
  const { t } = useT()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [oauthLoading, setOauthLoading] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [providers, setProviders] = useState<OAuthProvider[]>([])
  const navigate = useNavigate()
  const auth = useAuthStore(s => s.auth)

  useEffect(() => {
    // Detect enabled OAuth providers from /api/status
    fetch('/api/status').then(r => r.json()).then(data => {
      if (!data.success) return
      const s = data.data
      const list: OAuthProvider[] = []
      
      if (s.github_oauth) {
        list.push({
          id: 'github', name: 'GitHub', enabled: true,
          loginUrl: `https://github.com/login/oauth/authorize?client_id=${s.github_client_id}&redirect_uri=${window.location.origin}/api/oauth/github&state=`,
        })
      }
      if (s.wechat_login) {
        list.push({
          id: 'wechat', name: 'WeChat', enabled: true,
          loginUrl: `/api/oauth/wechat?state=`,
        })
      }
      if (s.oidc) {
        list.push({
          id: 'oidc', name: 'OIDC', enabled: true,
          loginUrl: `/api/oauth/oidc?state=`,
        })
      }
      if (s.lark_client_id) {
        list.push({
          id: 'lark', name: 'Lark', enabled: true,
          loginUrl: `${window.location.origin}/api/oauth/lark?state=`,
        })
      }
      setProviders(list)
    }).catch(() => {})
  }, [])

  const doLogin = useCallback(async () => {
    if (!username || !password) return
    setLoading(true); setError('')
    try {
      const res = await fetch('/api/user/login', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({username, password}) }).then(r=>r.json())
      if (res.success && res.data) {
        auth.setUser(res.data)
        navigate({ to: '/dashboard' })
      } else {
        setError(res.message || t('Login failed'))
      }
    } catch { setError(t('Network error')) }
    setLoading(false)
  }, [username, password, t, navigate, auth])

  const handleOAuthLogin = async (provider: OAuthProvider) => {
    setOauthLoading(provider.id)
    try {
      // Get OAuth state from backend
      const res = await fetch('/api/oauth/state').then(r => r.json())
      if (!res.success || !res.data) {
        setError(t('Failed to initiate OAuth login'))
        setOauthLoading(null)
        return
      }
      // Redirect to provider login URL with state
      window.location.href = provider.loginUrl + res.data
    } catch {
      setError(t('Failed to connect to auth service'))
      setOauthLoading(null)
    }
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center"
      style={{ backgroundImage: 'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)' }}>
      <div className="qc-fade-up w-full max-w-md mx-auto px-6">
        <div className="text-center mb-10">
          <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center shadow-lg shadow-orange-500/20 mx-auto mb-5">
            <img src="/logo.webp" alt="QuantumClaw" className="w-12 h-12 object-contain" />
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-foreground mb-2">{t('Sign In')}</h1>
          <p className="text-base text-muted-foreground/60">{t('Access your QuantumClaw dashboard')}</p>
        </div>

        <div className="rounded-2xl bg-white/80 backdrop-blur-sm border border-border/20 shadow-sm p-8">
          {error && (
            <div className="mb-5 px-5 py-3 rounded-xl bg-red-50 text-red-700 text-sm font-medium border border-red-200/50">
              {error}
            </div>
          )}
          <div className="space-y-5">
            <div>
              <label className="text-sm font-medium text-muted-foreground/70 block mb-2">{t('Username')}</label>
              <input type="text" value={username} onChange={e => setUsername(e.target.value)}
                className="w-full h-12 rounded-xl border border-border/30 bg-white px-5 text-base outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all"
                placeholder="Enter your username" autoComplete="username" />
            </div>
            <div>
              <label className="text-sm font-medium text-muted-foreground/70 block mb-2">{t('Password')}</label>
              <input type="password" value={password} onChange={e => setPassword(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && doLogin()}
                className="w-full h-12 rounded-xl border border-border/30 bg-white px-5 text-base outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all"
                placeholder="&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;" autoComplete="current-password" />
            </div>
            <button onClick={doLogin} disabled={loading || !username || !password}
              className="w-full py-3.5 rounded-xl text-base font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2">
              {loading ? <><div className="w-5 h-5 rounded-full border-2 border-white/30 border-t-white animate-spin" /></> : t('Sign In')}
            </button>
          </div>

          {providers.length > 0 && (
            <>
              <div className="relative my-6">
                <div className="absolute inset-0 flex items-center"><div className="w-full border-t border-border/20" /></div>
                <div className="relative flex justify-center"><span className="bg-white/80 px-3 text-xs text-muted-foreground/60">{t('Or continue with')}</span></div>
              </div>
              <div className="space-y-3">
                {providers.map(p => (
                  <button key={p.id} onClick={() => handleOAuthLogin(p)} disabled={oauthLoading === p.id}
                    className="w-full py-2.5 rounded-xl text-sm font-medium border border-border/20 bg-white hover:bg-gray-50 disabled:opacity-50 transition-all flex items-center justify-center gap-2">
                    {oauthLoading === p.id ? (
                      <div className="w-4 h-4 rounded-full border-2 border-gray-300 border-t-gray-600 animate-spin" />
                    ) : (
                      <OAuthIcon provider={p.id} />
                    )}
                    {t('Continue with')} {p.name}
                  </button>
                ))}
              </div>
            </>
          )}

          <div className="text-center mt-4">
            <p className="text-xs text-muted-foreground/60">
              {t("Don't have an account?")}{' '}
              <Link to="/setup" className="text-blue-600 hover:underline font-medium">{t('Set up now')}</Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

function OAuthIcon({ provider }: { provider: string }) {
  if (provider === 'github') {
    return (
      <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
      </svg>
    )
  }
  if (provider === 'wechat') {
    return (
      <svg className="w-4 h-4 text-green-600" viewBox="0 0 24 24" fill="currentColor">
        <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.045c.134 0 .24-.11.24-.245 0-.06-.024-.12-.04-.178l-.325-1.233a.49.49 0 0 1 .177-.554C23.028 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-7.062-6.122zm-2.18 2.923c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982z"/>
      </svg>
    )
  }
  if (provider === 'oidc') {
    return <span className="w-4 h-4 flex items-center justify-center text-purple-600 text-[10px] font-bold rounded-full border border-purple-300">O</span>
  }
  if (provider === 'lark') {
    return <span className="w-4 h-4 flex items-center justify-center text-blue-600 text-[10px] font-bold rounded-full border border-blue-300">L</span>
  }
  return null
}