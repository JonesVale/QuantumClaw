import { createFileRoute, Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import {
  ExternalLink, Search, Star, Code, Bot, MessageSquare,
  Braces, Database, Cpu, Globe, Sparkles, ArrowUpRight
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/apps')({
  component: AppsPage,
})

interface App {
  name: string
  description: string
  url: string
  category: string
  icon: React.ElementType
  users: string
  featured?: boolean
}

const apps: App[] = [
  { name: 'Cursor', description: 'AI-first code editor with multi-model support. Integrates QuantumClaw API for intelligent code completion.', url: 'https://cursor.sh', category: 'Development', icon: Code, users: '1.2M+' },
  { name: 'ChatBox', description: 'All-in-one AI desktop client supporting all major LLMs through QuantumClaw gateway.', url: 'https://chatbox.app', category: 'Chat', icon: MessageSquare, users: '800K+', featured: true },
  { name: 'Continue', description: 'Open-source AI code assistant for VS Code & JetBrains, powered by QuantumClaw API.', url: 'https://continue.dev', category: 'Development', icon: Braces, users: '500K+' },
  { name: 'LobeChat', description: 'Modern chat framework with plugin system and multi-model support.', url: 'https://lobehub.com', category: 'Chat', icon: Bot, users: '300K+', featured: true },
  { name: 'Open WebUI', description: 'Self-hosted WebUI for LLMs with QuantumClaw API integration.', url: 'https://openwebui.com', category: 'Chat', icon: Globe, users: '250K+' },
  { name: 'Dify', description: 'Open-source LLM app development platform. Build and deploy AI applications.', url: 'https://dify.ai', category: 'Platform', icon: Database, users: '200K+', featured: true },
  { name: 'FastGPT', description: 'Knowledge-based Q&A system built on LLMs and vector databases.', url: 'https://fastgpt.in', category: 'Platform', icon: Cpu, users: '150K+' },
  { name: 'Cherry Studio', description: 'Desktop client for LLMs with support for multiple AI services.', url: 'https://cherry-ai.com', category: 'Chat', icon: Sparkles, users: '100K+' },
  { name: 'AI Toolkit', description: 'Browser extension for ChatGPT, Gemini, Claude with QuantumClaw API support.', url: 'https://aitoolkit.com', category: 'Tools', icon: Star, users: '80K+' },
]

const categories = ['All', 'Development', 'Chat', 'Platform', 'Tools']

function AppsPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('All')

  const filtered = apps.filter(a => {
    if (category !== 'All' && a.category !== category) return false
    if (search && !a.name.toLowerCase().includes(search.toLowerCase()) && 
        !a.description.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  const featured = apps.filter(a => a.featured)

  return (
    <div className="p-4 sm:p-6 space-y-6 w-full min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-orange-500 via-red-500 to-pink-600 bg-clip-text text-transparent flex items-center gap-3">
            <Sparkles className="h-8 w-8 text-orange-500" />
            {t('Apps')}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm sm:text-base">
            {t('Applications and tools powered by QuantumClaw API')}
          </p>
        </div>
        <Badge variant="outline" className="text-sm px-4 py-2 gap-2">
          <Code className="h-4 w-4" />
          {apps.length}+ {t('integrations')}
        </Badge>
      </div>

      {/* Featured Apps */}
      {featured.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
            <Star className="h-5 w-5 text-amber-500 fill-amber-500" />
            {t('Featured')}
          </h2>
          <div className="grid sm:grid-cols-3 gap-4">
            {featured.map((app) => (
              <a key={app.name} href={app.url} target="_blank" rel="noopener noreferrer" className="block group">
                <Card className="h-full hover:shadow-xl transition-all duration-300 hover:-translate-y-0.5 hover:border-primary/30 relative overflow-hidden rounded-xl">
                  <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
                    <ExternalLink className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <CardContent className="p-5">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-orange-500 to-pink-600 flex items-center justify-center mb-3 shadow">
                      <app.icon className="h-6 w-6 text-white" />
                    </div>
                    <CardTitle className="text-base mb-1 group-hover:text-orange-600 dark:group-hover:text-orange-400 transition-colors">
                      {app.name}
                    </CardTitle>
                    <p className="text-sm text-muted-foreground line-clamp-2">{app.description}</p>
                    <div className="flex items-center gap-2 mt-3">
                      <Badge variant="secondary" className="text-[10px]">{app.category}</Badge>
                      <span className="text-[10px] text-muted-foreground">{app.users} {t('users')}</span>
                    </div>
                  </CardContent>
                </Card>
              </a>
            ))}
          </div>
        </div>
      )}

      {/* Search & Filter */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            className="pl-9"
            placeholder={t('Search apps...')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="flex gap-2 overflow-x-auto">
          {categories.map(cat => (
            <Button
              key={cat}
              variant={category === cat ? 'default' : 'outline'}
              size="sm"
              onClick={() => setCategory(cat)}
              className="shrink-0"
            >
              {t(cat)}
            </Button>
          ))}
        </div>
      </div>

      {/* Apps Grid */}
      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {filtered.map((app) => (
          <a key={app.name} href={app.url} target="_blank" rel="noopener noreferrer" className="block group">
            <Card className="h-full hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5 rounded-xl">
              <CardContent className="p-4">
                <div className="flex items-start gap-3">
                  <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-orange-400 to-pink-500 flex items-center justify-center shrink-0 shadow">
                    <app.icon className="h-5 w-5 text-white" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-semibold truncate">{app.name}</p>
                      <ArrowUpRight className="h-3 w-3 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
                    </div>
                    <p className="text-xs text-muted-foreground line-clamp-2 mt-0.5">{app.description}</p>
                    <div className="flex items-center gap-2 mt-2">
                      <Badge variant="outline" className="text-[10px]">{app.category}</Badge>
                      <span className="text-[10px] text-muted-foreground">{app.users}</span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </a>
        ))}
      </div>

      {filtered.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            <Search className="h-8 w-8 mx-auto mb-3 opacity-30" />
            <p>{t('No apps found')}</p>
          </CardContent>
        </Card>
      )}

      {/* CTA */}
      <Card className="bg-gradient-to-r from-orange-500/10 via-red-500/10 to-pink-500/10 border-orange-500/20">
        <CardContent className="p-6 text-center">
          <h3 className="text-lg font-semibold mb-2">{t('Built something with QuantumClaw?')}</h3>
          <p className="text-sm text-muted-foreground mb-4">{t('Let us know and get featured here')}</p>
          <Button variant="outline" asChild>
            <a href="https://github.com/QuantumClaw" target="_blank" rel="noopener noreferrer" className="gap-2">
              <ExternalLink className="h-4 w-4" />
              {t('Submit your app')}
            </a>
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
