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
  const [confirmPassword, setConfirmPassword] = useState('')
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const navigate = useNavigate()
  const auth = useAuthStore(s => s.auth)
  const refCode = typeof window !== 'undefined'
    ? new URLSearchParams(window.location.search).get('ref') || ''
    : ''

  useEffect(() => {
    fetch('/api/status').then(r => r.json()).then(data => {
      if (!data.success) return
      const s = data.data
      const sv = s.server_address || window.location.origin
      const list: OAuthProvider[] = []

      if (s.github_oauth && s.github_client_id) {
        list.push({
          id: 'github', name: 'GitHub', enabled: true,
          loginUrl: `https://github.com/login/oauth/authorize?client_id=${s.github_client_id}&redirect_uri=${sv}/api/oauth/github&state=`,
        })
      }
      if (s.discord_oauth && s.discord_client_id) {
        list.push({
          id: 'discord', name: 'Discord', enabled: true,
          loginUrl: `https://discord.com/api/oauth2/authorize?client_id=${s.discord_client_id}&redirect_uri=${sv}/api/oauth/discord&response_type=code&scope=identify+email&state=`,
        })
      }
      if (s.linuxdo_oauth && s.linuxdo_client_id) {
        list.push({
          id: 'linuxdo', name: 'Linux Do', enabled: true,
          loginUrl: `https://connect.linux.do/oauth/authorize?client_id=${s.linuxdo_client_id}&redirect_uri=${sv}/api/oauth/linuxdo&response_type=code&scope=read&state=`,
        })
      }
      if (s.wechat_login) {
        list.push({
          id: 'wechat', name: 'WeChat', enabled: true,
          loginUrl: `${sv}/api/oauth/wechat?state=`,
        })
      }
      if (s.lark_client_id) {
        list.push({
          id: 'lark', name: 'Lark', enabled: true,
          loginUrl: `${sv}/api/oauth/lark?state=`,
        })
      }
      if (s.oidc && s.oidc_authorization_endpoint && s.oidc_client_id) {
        list.push({
          id: 'oidc', name: 'OIDC', enabled: true,
          loginUrl: `${s.oidc_authorization_endpoint}?client_id=${s.oidc_client_id}&redirect_uri=${sv}/api/oauth/oidc&response_type=code&scope=openid+profile+email&state=`,
        })
      }
      if (s.telegram_oauth && s.telegram_bot_username) {
        list.push({
          id: 'telegram', name: 'Telegram', enabled: true,
          loginUrl: `https://t.me/${s.telegram_bot_username}?start=auth_`,
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
  }, [username, password, t, navigate, auth, refCode])

  const doRegister = useCallback(async () => {
    if (!username || !password || password !== confirmPassword) return
    setLoading(true); setError('')
    try {
      const body = { username, password }
      if (refCode) body.aff_code = refCode
      const res = await fetch('/api/user/register', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify(body) }).then(r=>r.json())
      if (res.success && res.data) {
        auth.setUser(res.data)
        navigate({ to: '/dashboard' })
      } else {
        setError(res.message || t('Registration failed'))
      }
    } catch { setError(t('Network error')) }
    setLoading(false)
  }, [username, password, confirmPassword, refCode, t, navigate, auth])

  const handleOAuthLogin = async (provider: OAuthProvider) => {
    setOauthLoading(provider.id)
    try {
      const res = await fetch('/api/oauth/state').then(r => r.json())
      if (!res.success || !res.data) {
        setError(t('Failed to initiate OAuth login'))
        setOauthLoading(null)
        return
      }
      window.location.href = provider.loginUrl + res.data
    } catch {
      setError(t('Failed to connect to auth service'))
      setOauthLoading(null)
    }
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
          <div className="flex mb-7 gap-2">
            <button
              onClick={() => setMode('login')}
              className={'flex-1 py-4 text-lg font-medium rounded-xl transition-all ' + (mode === 'login' ? 'bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-sm' : 'text-amber-200 hover:text-white')}
            >
              {t('Sign In')}
            </button>
            <button
              onClick={() => setMode('register')}
              className={'flex-1 py-4 text-lg font-medium rounded-xl transition-all ml-2 ' + (mode === 'register' ? 'bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-sm' : 'text-amber-200 hover:text-white')}
            >
              {t('Register')}
            </button>
          </div>
                    {error && (
            <div className="mb-5 px-5 py-3 rounded-xl bg-red-900/90 text-red-200 text-base font-medium border border-red-700">
              {error}
            </div>
          )}
          <div className="space-y-5">
            <div>
              <label className="text-lg font-medium text-amber-200 block mb-3">{t('Username')}</label>
              <input type="text" value={username} onChange={e => setUsername(e.target.value)}
                className="w-full h-14 rounded-xl border border-gray-600 bg-gray-800 px-5 text-base text-white outline-none placeholder:text-gray-500 focus:border-amber-500 focus:ring-2 focus:ring-amber-400/30 transition-all"
                placeholder="Enter your username" autoComplete="username" />
            </div>
            <div>
              <label className="text-lg font-medium text-amber-200 block mb-3">{t('Password')}</label>
              <input type="password" value={password} onChange={e => setPassword(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && (mode === 'login' ? doLogin() : doRegister())}
                className="w-full h-14 rounded-xl border border-gray-600 bg-gray-800 px-5 text-base text-white outline-none placeholder:text-gray-500 focus:border-amber-500 focus:ring-2 focus:ring-amber-400/30 transition-all"
                placeholder="&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;" autoComplete="current-password" />
            </div>
            {mode === 'register' && (
            <div>
              <label className="text-lg font-medium text-amber-200 block mb-3">{t('Confirm Password')}</label>
              <input type="password" value={confirmPassword} onChange={e => setConfirmPassword(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && doRegister()}
                className="w-full h-14 rounded-xl border border-gray-600 bg-gray-800 px-5 text-base text-white outline-none placeholder:text-gray-500 focus:border-amber-500 focus:ring-2 focus:ring-amber-400/30 transition-all"
                placeholder="&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;&#x2022;" autoComplete="new-password" />
            </div>
            )}
            <button onClick={mode === 'login' ? doLogin : doRegister} disabled={loading || !username || !password || (mode === 'register' && (!confirmPassword || password !== confirmPassword))}
              className="w-full py-5 rounded-xl text-xl font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg hover:shadow-orange-500/30 hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 transition-all duration-200 flex items-center justify-center gap-2">
              {loading ? <><div className="w-5 h-5 rounded-full border-2 border-white/30 border-t-white animate-spin" /></> : mode === 'login' ? t('Sign In') : t('Register')}
            </button>
          </div>

          {providers.length > 0 && (
            <>
              <div className="relative my-8">
                <div className="absolute inset-0 flex items-center"><div className="w-full border-t border-gray-700" /></div>
                <div className="relative flex justify-center"><span className="bg-gray-900 px-4 text-base text-gray-400">{t('Or continue with')}</span></div>
              </div>
              <div className="space-y-3">
                {providers.map(p => (
                  <button key={p.id} onClick={() => handleOAuthLogin(p)} disabled={oauthLoading === p.id}
                    className="w-full py-4 rounded-xl text-lg font-medium border border-gray-700 bg-gray-800 text-white hover:bg-gray-700 disabled:opacity-50 transition-all flex items-center justify-center gap-3">
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
  if (provider === 'discord') {
    return (
      <svg className="w-4 h-4 text-indigo-500" viewBox="0 0 24 24" fill="currentColor">
        <path d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.9555 2.4189-2.1569 2.4189zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.4189-2.1568 2.4189z"/>
      </svg>
    )
  }
  if (provider === 'linuxdo') {
    return <span className="w-4 h-4 flex items-center justify-center text-green-600 text-[9px] font-bold">LD</span>
  }
  if (provider === 'wechat') {
    return (
      <svg className="w-4 h-4 text-green-600" viewBox="0 0 24 24" fill="currentColor">
        <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.045c.134 0 .24-.11.24-.245 0-.06-.024-.12-.04-.178l-.325-1.233a.49.49 0 0 1 .177-.554C23.028 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-7.062-6.122zm-2.18 2.923c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982z"/>
      </svg>
    )
  }
  if (provider === 'lark') {
    return <span className="w-4 h-4 flex items-center justify-center text-blue-600 text-[9px] font-bold">飞</span>
  }
  if (provider === 'oidc') {
    return <span className="w-4 h-4 flex items-center justify-center text-purple-600 text-[10px] font-bold rounded-full border border-purple-300">O</span>
  }
  if (provider === 'telegram') {
    return (
      <svg className="w-4 h-4 text-blue-500" viewBox="0 0 24 24" fill="currentColor">
        <path d="M11.944 0A12 12 0 000 12a12 12 0 0012 12 12 12 0 0012-12A12 12 0 0012 0a12 12 0 00-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 01.171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/>
      </svg>
    )
  }
  return null
}
