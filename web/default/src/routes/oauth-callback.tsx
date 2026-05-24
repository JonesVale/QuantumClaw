import { createFileRoute, useNavigate, useSearch } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useEffect, useState, useRef } from 'react'
import { Loader2, CheckCircle2, XCircle, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import apiClient from '@/lib/api'
import { toast } from 'sonner'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface OAuthCallbackSearch {
  code?: string
  state?: string
  provider?: string
}

interface OAuthCallbackResponse {
  success: boolean
  message?: string
}

// ---------------------------------------------------------------------------
// Route
// ---------------------------------------------------------------------------

export const Route = createFileRoute('/oauth-callback')({
  component: OAuthCallbackPage,
  validateSearch: (search: Record<string, string | undefined>): OAuthCallbackSearch => ({
    code: search.code,
    state: search.state,
    provider: search.provider,
  }),
})

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

function OAuthCallbackPage() {
  const { t } = useT()
  const navigate = useNavigate()
  const search = useSearch({ from: Route.id })
  const [status, setStatus] = useState<'processing' | 'success' | 'error'>('processing')
  const [errorMessage, setErrorMessage] = useState('')
  const calledRef = useRef(false)

  useEffect(() => {
    // Prevent double-call in strict mode
    if (calledRef.current) return
    calledRef.current = true

    const { code, state, provider } = search

    if (!code || !state || !provider) {
      setStatus('error')
      setErrorMessage(t('Missing required OAuth parameters: code, state, or provider'))
      return
    }

    const processOAuth = async () => {
      try {
        const res = await apiClient.post(`/api/oauth/${provider}/callback`, {
          code,
          state,
        })
        const data = res.data as OAuthCallbackResponse

        if (data.success) {
          setStatus('success')
          toast.success(t('OAuth connection successful!'))
          // Redirect to connections page after brief delay
          setTimeout(() => {
            navigate({ to: '/account/connections' })
          }, 1500)
        } else {
          setStatus('error')
          setErrorMessage(data.message || t('OAuth callback failed'))
          toast.error(data.message || t('OAuth callback failed'))
        }
      } catch (err: any) {
        setStatus('error')
        const msg = err?.response?.data?.message || err?.message || t('OAuth callback failed')
        setErrorMessage(msg)
        toast.error(msg)
      }
    }

    processOAuth()
  }, [])

  return (
    <div
      className="min-h-screen bg-background flex items-center justify-center"
      style={{
        backgroundImage:
          'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)',
      }}
    >
      <div className="w-full max-w-md mx-auto px-6">
        <Card className="border-border/20 shadow-sm bg-white/80 backdrop-blur-sm">
          <CardHeader className="text-center pb-4">
            <CardTitle className="text-xl">{t('OAuth Connection')}</CardTitle>
            <CardDescription>
              {status === 'processing'
                ? t('Processing your OAuth callback...')
                : status === 'success'
                  ? t('Connection successful!')
                  : t('Connection failed')}
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4 pb-8">
            {status === 'processing' && (
              <div className="flex flex-col items-center gap-3 py-6">
                <div className="relative">
                  <Loader2 className="h-12 w-12 animate-spin text-amber-500" />
                  <div className="absolute inset-0 flex items-center justify-center">
                    <div className="h-4 w-4 rounded-full bg-amber-500/20 animate-pulse" />
                  </div>
                </div>
                <p className="text-sm text-muted-foreground text-center">
                  {t('Completing your OAuth connection...')}
                </p>
              </div>
            )}

            {status === 'success' && (
              <div className="flex flex-col items-center gap-3 py-6">
                <div className="h-16 w-16 rounded-full bg-green-100 dark:bg-green-950/30 flex items-center justify-center">
                  <CheckCircle2 className="h-8 w-8 text-green-600 dark:text-green-400" />
                </div>
                <p className="text-sm text-green-700 dark:text-green-400 font-medium">
                  {t('OAuth connection successful! Redirecting...')}
                </p>
              </div>
            )}

            {status === 'error' && (
              <div className="flex flex-col items-center gap-3 py-4">
                <div className="h-16 w-16 rounded-full bg-red-100 dark:bg-red-950/30 flex items-center justify-center">
                  <XCircle className="h-8 w-8 text-red-600 dark:text-red-400" />
                </div>
                {errorMessage && (
                  <p className="text-sm text-red-600 dark:text-red-400 text-center max-w-sm">
                    {errorMessage}
                  </p>
                )}
                <div className="flex gap-3 mt-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => navigate({ to: '/account/connections' })}
                  >
                    <ArrowLeft className="mr-1 h-3 w-3" />
                    {t('Back to Connections')}
                  </Button>
                  <Button
                    variant="default"
                    size="sm"
                    className="bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-sm hover:shadow-md"
                    onClick={() => navigate({ to: '/' })}
                  >
                    {t('Go Home')}
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Branding */}
        <div className="text-center mt-6">
          <p className="text-xs text-muted-foreground/50">
            QuantumClaw AI API Gateway
          </p>
        </div>
      </div>
    </div>
  )
}
