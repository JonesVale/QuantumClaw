import { createFileRoute, redirect, useSearch, useRouter } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { useAuthStore } from '@/stores/auth-store'
import { signIn } from '@/lib/api-extended'
import { toast } from 'sonner'
import { getTelegramWidgetInfo } from '@/lib/api-extended'
export const Route = createFileRoute('/(auth)/sign-in')({
  component: SignInPage,
  beforeLoad: () => {
    const { auth } = useAuthStore.getState()
    if (auth.user) {
      throw redirect({ to: '/dashboard' })
    }
  },
  validateSearch: (search) => ({ redirect: (search.redirect as string) || undefined }),
})
function SignInPage() {
  const { t } = useTranslation()
  const { auth } = useAuthStore()
  const router = useRouter()
  const { redirect: redirectUrl } = useSearch({ strict: false })
  const [loading, setLoading] = useState(false)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [botUsername, setBotUsername] = useState<string | null>(null)
  // Load Telegram widget info
  useEffect(() => {
    getTelegramWidgetInfo().then((res) => {
      if (res.success && res.data?.bot_username) {
        setBotUsername(res.data.bot_username)
      }
    }).catch(() => {
      // Telegram login not configured, ignore
    })
  }, [])
  // Telegram auth callback handler - listens for redirected Telegram auth data
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    if (params.has('id') && params.has('hash') && params.has('auth_date')) {
      // Telegram auth data detected in URL - send to backend
      setLoading(true)
      fetch('/api/oauth/telegram' + window.location.search, {
        method: 'GET',
      })
        .then((r) => r.json())
        .then((res) => {
          if (res.success) {
            toast.success(t('Welcome back!'))
            auth.setUser(res.data)
            router.navigate({ to: (redirectUrl as string) || '/dashboard' })
          } else {
            toast.error(res.message || t('Login failed'))
          }
        })
        .catch(() => toast.error(t('Login failed')))
        .finally(() => setLoading(false))
    }
  }, [])
  const handleTelegramLogin = () => {
    if (botUsername) {
      // Open Telegram Login Widget in a popup
      const redirectUri = window.location.origin + '/api/oauth/telegram'
      const tgAuthUrl = `https://oauth.telegram.org/auth?bot_id=${botUsername}&origin=${encodeURIComponent(window.location.origin)}&return_to=${encodeURIComponent(redirectUri)}`
      window.open(tgAuthUrl, 'TelegramLogin', 'width=600,height=500')
    }
  }
  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      const res = await signIn(username, password)
      if (res.success) {
        toast.success(t('Welcome back!'))
        auth.setUser(res.data)
        router.navigate({ to: (redirectUrl as string) || '/dashboard' })
      } else {
        toast.error(res.message || t('Login failed'))
      }
    } catch {
      toast.error(t('Login failed'))
    } finally {
      setLoading(false)
    }
  }
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 via-white to-blue-50 p-4 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950 mx-auto max-w-[min(96vw,1600px)] w-full">
      <Card className="w-full sm:max-w-md">
        <CardHeader className="items-center text-center">
          <img src="/logo.webp" alt="QuantumClaw" className="h-12 w-12 rounded-xl object-cover mb-2" />
          <CardTitle className="text-2xl sm:text-3xl font-bold">{t('QuantumClaw')}</CardTitle>
          <CardDescription>{t('AI API Gateway & Management Platform')}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleLogin} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">{t('Username')}</Label>
              <Input
                id="username"
                placeholder={t('Username or email')}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">{t('Password')}</Label>
              <Input
                id="password"
                type="password"
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? t('Signing in...') : t('Sign In')}
            </Button>
            <p className="mt-4 text-center text-sm text-muted-foreground">
              No account? <a href="/api/user/register" className="text-blue-600 hover:underline font-medium">Register now</a>
            </p>
          </form>
          {botUsername && (
            <>
              <Separator className="my-4" />
              <div className="space-y-2">
                <Button
                  variant="outline"
                  className="w-full"
                  onClick={handleTelegramLogin}
                  disabled={loading}
                >
                  <svg
                    className="mr-2 h-4 w-4"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                  >
                    <path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.127.087.497.131.666.046.178.162.688.185.803a5.268 5.268 0 0 1 .054.644c-.002.052.004.626-.065.98-.028.145-.052.283-.1.392-.02.046-.046.125-.089.16a.365.365 0 0 1-.229.076.899.899 0 0 1-.531-.168c-.042-.032-.503-.348-.789-.545-.092-.064-.19-.113-.299-.07-.082.032-.13.107-.146.188-.02.1-.005.222.004.3.009.06.014.117-.004.187-.065.247-.366.444-.541.516-.11.046-.358.143-.583.178a6.067 6.067 0 0 1-2.313-.027c-.34-.07-.669-.168-.962-.268-.173-.06-.346-.138-.505-.21l-.078-.035c.023-.044.049-.086.077-.126a3.946 3.946 0 0 0 .49-.722c.193-.354.394-.814.516-1.202.03-.095.058-.19.058-.288 0-.006-.002-.402-.03-.47a.33.33 0 0 0-.14-.191.354.354 0 0 0-.214-.047c-.078.01-.258.05-.371.095-.295.117-.731.581-.98.776-.082.064-.185.128-.274.182l-.618.388c-.085.052-.18.113-.243.167-.14.12-.192.214-.23.244-.013.01-.04.019-.064.017-.053-.003-.08-.02-.1-.04a.252.252 0 0 1-.056-.084c-.056-.127-.108-.478-.141-.724-.03-.216-.045-.436-.035-.532.003-.03.01-.058.025-.082a.42.42 0 0 1 .131-.12c.008-.005.288-.176.318-.2.062-.05.137-.12.218-.183.654-.525 1.558-1.058 2.182-1.333.698-.309 1.575-.587 2.413-.537z"/>
                  </svg>
                  {t('Sign in with Telegram')}
                </Button>
              </div>
            </>
          )}
          {/* LinuxDO Login */}
          <Separator className="my-4" />
          <div className="space-y-2">
            <Button
              variant="outline"
              className="w-full"
              onClick={async () => {
                try {
                  const res = await fetch('/api/oauth/linuxdo/generate')
                  const data = await res.json()
                  if (data.success && data.data) {
                    window.location.href = data.data
                  } else {
                    toast.error(t('Failed to initiate LinuxDO login'))
                  }
                } catch {
                  toast.error(t('Failed to initiate LinuxDO login'))
                }
              }}
              disabled={loading}
            >
              <svg className="mr-2 h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/>
              </svg>
              {t('Sign in with LinuxDO')}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
