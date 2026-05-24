import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Link2, Link2Off, Globe, MessageSquare,
  Shield, Box, ExternalLink, Loader2, CheckCircle2, XCircle,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { useAuthStore } from '@/stores/auth-store'
import apiClient from '@/lib/api'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/connections')({
  component: ConnectionsPage,
})

function GithubIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
    </svg>
  )
}

function WeChatIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor">
      <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.045c.134 0 .24-.11.24-.245 0-.06-.024-.12-.04-.178l-.325-1.233a.49.49 0 0 1 .177-.554C23.028 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-7.062-6.122zm-2.18 2.923c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982z"/>
    </svg>
  )
}

function DiscordIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor">
      <path d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.9555 2.4189-2.1569 2.4189zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.4189-2.1568 2.4189z"/>
    </svg>
  )
}

function TelegramIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor">
      <path d="M11.944 0A12 12 0 000 12a12 12 0 0012 12 12 12 0 0012-12A12 12 0 0012 0a12 12 0 00-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 01.171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/>
    </svg>
  )
}

interface OAuthProvider {
  id: string
  name: string
  nameKey: string
  icon: React.ElementType
  field: string // field name in user object
  color: string
  bindEndpoint: string
  initEndpoint?: string
}

const PROVIDERS: OAuthProvider[] = [
  {
    id: 'github',
    name: 'GitHub',
    nameKey: 'GitHub',
    icon: GithubIcon,
    field: 'github_id',
    color: 'text-gray-800 dark:text-gray-200',
    bindEndpoint: '/api/oauth/github',
  },
  {
    id: 'wechat',
    name: 'WeChat',
    nameKey: 'WeChat',
    icon: WeChatIcon,
    field: 'wechat_id',
    color: 'text-green-600',
    bindEndpoint: '/api/oauth/wechat/bind',
  },
  {
    id: 'lark',
    name: 'Lark / Feishu',
    nameKey: 'Lark / Feishu',
    icon: Globe,
    field: 'lark_id',
    color: 'text-blue-600',
    bindEndpoint: '/api/oauth/lark',
  },
  {
    id: 'discord',
    name: 'Discord',
    nameKey: 'Discord',
    icon: DiscordIcon,
    field: 'discord_id',
    color: 'text-indigo-600',
    bindEndpoint: '/api/oauth/discord',
    initEndpoint: '/api/oauth/discord/generate',
  },
  {
    id: 'linuxdo',
    name: 'Linux DO',
    nameKey: 'Linux DO',
    icon: Globe,
    field: 'linuxdo_id',
    color: 'text-amber-600',
    bindEndpoint: '/api/oauth/linuxdo',
    initEndpoint: '/api/oauth/linuxdo/generate',
  },
  {
    id: 'telegram',
    name: 'Telegram',
    nameKey: 'Telegram',
    icon: TelegramIcon,
    field: 'telegram_id',
    color: 'text-sky-600',
    bindEndpoint: '/api/oauth/telegram/bind',
  },
  {
    id: 'oidc',
    name: 'OIDC',
    nameKey: 'OIDC',
    icon: Shield,
    field: 'oidc_id',
    color: 'text-purple-600',
    bindEndpoint: '/api/oauth/oidc',
  },
  {
    id: 'custom_oauth',
    name: 'Custom OAuth',
    nameKey: 'Custom OAuth',
    icon: Box,
    field: 'custom_oauth_id',
    color: 'text-orange-600',
    bindEndpoint: '/api/oauth/custom',
  },
]

function ConnectionsPage() {
  const { t } = useT()
  const { auth } = useAuthStore()
  const user = auth.user
  const queryClient = useQueryClient()
  const [loadingAction, setLoadingAction] = useState<string | null>(null)

  const connectedProviders = new Set<string>()

  if (user) {
    if (user.github_id) connectedProviders.add('github')
    if (user.wechat_id) connectedProviders.add('wechat')
    if (user.lark_id) connectedProviders.add('lark')
    if (user.discord_id) connectedProviders.add('discord')
    if (user.linuxdo_id) connectedProviders.add('linuxdo')
    if (user.telegram_id) connectedProviders.add('telegram')
    if (user.oidc_id) connectedProviders.add('oidc')
    if (user.custom_oauth_id) connectedProviders.add('custom_oauth')
  }

  const handleConnect = async (provider: OAuthProvider) => {
    setLoadingAction(provider.id)

    try {
      if (provider.initEndpoint) {
        // Use the init endpoint to get the authorization URL
        const res = await fetch(provider.initEndpoint)
        const data = await res.json()
        if (data.success && data.data) {
          window.location.href = data.data
          return
        }
      }

      // Fallback: get OAuth state and construct URL
      const stateRes = await fetch('/api/oauth/state')
      const stateData = await stateRes.json()

      if (!stateData.success) {
        toast.error(stateData.message || t('Failed to initiate connection'))
        return
      }

      // For providers without a dedicated init endpoint, use a server redirect
      // Redirect to the backend bind route; backend handles provider redirect
      const bindUrl = provider.bindEndpoint
      window.location.href = bindUrl
    } catch (err) {
      toast.error(t('Failed to connect. Please try again.'))
    } finally {
      setLoadingAction(null)
    }
  }

  const handleDisconnect = async (provider: OAuthProvider) => {
    setLoadingAction(provider.id)
    try {
      const res = await apiClient.post('/api/user/self/unbind-oauth', {
        provider: provider.id,
      })
      const data = res.data as { success: boolean; message?: string }
      if (data.success) {
        toast.success(t('Disconnected successfully'))
        // Refresh user data
        queryClient.invalidateQueries({ queryKey: ['self'] })
        auth.setUser({ ...user, [provider.field]: '' } as typeof user)
      } else {
        toast.error(data.message || t('Failed to disconnect'))
      }
    } catch {
      toast.error(t('Failed to disconnect. Please try again.'))
    } finally {
      setLoadingAction(null)
    }
  }

  return (
    <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">{t('OAuth Connections')}</h1>
          <p className="text-muted-foreground mt-2">
            {t('Connect your account with third-party services for easy login')}
          </p>
        </div>
      </div>

      <Card className="w-full max-w-3xl mx-auto">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Link2 className="h-4 w-4" />
            {t('Connected Services')}
          </CardTitle>
          <CardDescription>
            {t('Manage your OAuth provider connections')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-1">
          {PROVIDERS.map((provider, index) => {
            const isConnected = connectedProviders.has(provider.id)
            const isLoading = loadingAction === provider.id
            const Icon = provider.icon

            return (
              <div key={provider.id}>
                <div className="flex items-center justify-between py-4 px-2">
                  <div className="flex items-center gap-4">
                    <div className={cn(
                      'flex h-10 w-10 items-center justify-center rounded-xl border bg-white/50',
                      isConnected ? 'border-green-200' : 'border-border/30'
                    )}>
                      <Icon className={cn('h-5 w-5', provider.color)} />
                    </div>
                    <div>
                      <p className="text-sm font-medium">{t(provider.nameKey)}</p>
                      <p className="text-xs text-muted-foreground">
                        {isConnected
                          ? t('Connected')
                          : t('Not connected')}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge
                      variant={isConnected ? 'default' : 'secondary'}
                      className={cn(
                        isConnected
                          ? 'bg-green-50 text-green-700 border-green-200 dark:bg-green-950/30 dark:text-green-400 dark:border-green-800'
                          : ''
                      )}
                    >
                      {isConnected ? (
                        <span className="flex items-center gap-1">
                          <CheckCircle2 className="h-3 w-3" />
                          {t('Connected')}
                        </span>
                      ) : (
                        <span className="flex items-center gap-1">
                          <XCircle className="h-3 w-3" />
                          {t('Disconnected')}
                        </span>
                      )}
                    </Badge>
                    {isConnected ? (
                      <Button
                        variant="outline"
                        size="sm"
                        className="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/20 border-red-200 dark:border-red-800/50"
                        onClick={() => handleDisconnect(provider)}
                        disabled={isLoading}
                      >
                        {isLoading ? (
                          <Loader2 className="h-3 w-3 animate-spin mr-1" />
                        ) : (
                          <Link2Off className="h-3 w-3 mr-1" />
                        )}
                        {t('Disconnect')}
                      </Button>
                    ) : (
                      <Button
                        variant="default"
                        size="sm"
                        className="bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-sm hover:shadow-md"
                        onClick={() => handleConnect(provider)}
                        disabled={isLoading}
                      >
                        {isLoading ? (
                          <Loader2 className="h-3 w-3 animate-spin mr-1" />
                        ) : (
                          <Link2 className="h-3 w-3 mr-1" />
                        )}
                        {t('Connect')}
                      </Button>
                    )}
                  </div>
                </div>
                {index < PROVIDERS.length - 1 && <Separator />}
              </div>
            )
          })}
        </CardContent>
      </Card>

      <Card className="w-full max-w-3xl mx-auto">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            {t('Security Note')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            {t('OAuth connections allow you to sign in to QuantumClaw using your third-party accounts. You can connect or disconnect any provider at any time. Your credentials are never shared with third-party services.')}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
