import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, memo, useMemo, useEffect, useRef } from 'react'
import { ExternalLink, Newspaper, Loader2, AlertCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export const Route = createFileRoute('/_authenticated/news')({
  component: NewsPage,
})

interface RssArticle {
  id: number
  source: string
  title: string
  link: string
  description: string
  author: string
  published_at: string
  language: string
  created_at: string
}

interface RssApiResponse {
  success: boolean
  data: {
    articles: RssArticle[]
    total: number
    limit: number
    offset: number
  }
}

const LANG_TABS = ['all', 'zh', 'en'] as const
type LangTab = (typeof LANG_TABS)[number]

const LANG_LABELS: Record<string, string> = {
  all: 'All',
  zh: '\u4e2d\u6587',
  en: 'EN',
}

const LANG_VARIANTS: Record<string, 'default' | 'secondary' | 'outline' | 'destructive'> = {
  zh: 'default',
  en: 'secondary',
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  try {
    const date = new Date(dateStr)
    return date.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  } catch {
    return dateStr
  }
}

function truncateText(text: string, maxLen: number): string {
  if (!text) return ''
  const clean = text.replace(/\s+/g, ' ').trim()
  if (clean.length <= maxLen) return clean
  return clean.slice(0, maxLen) + '\u2026'
}

function ArticleCard({ article }: { article: RssArticle }) {
  const { t } = useT()
  return (
    <Card className="flex flex-col h-full">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-2 mb-2">
          <Badge variant={LANG_VARIANTS[article.language] || 'outline'} className="shrink-0 text-[10px]">
            {LANG_LABELS[article.language] || article.language}
          </Badge>
          <span className="text-[10px] text-muted-foreground truncate" title={article.source}>
            {article.source}
          </span>
        </div>
        <CardTitle className="text-sm leading-snug line-clamp-2" title={article.title}>
          {article.title || t('Untitled')}
        </CardTitle>
        {article.author && (
          <p className="text-[11px] text-muted-foreground">{article.author}</p>
        )}
      </CardHeader>
      <CardContent className="flex flex-col gap-3 flex-1">
        {article.description && (
          <p className="text-xs text-muted-foreground leading-relaxed line-clamp-3 flex-1">
            {truncateText(article.description, 200)}
          </p>
        )}
        <div className="flex items-center justify-between gap-2 mt-auto">
          <span className="text-[10px] text-muted-foreground">
            {formatDate(article.published_at)}
          </span>
          <Button
            variant="outline"
            size="sm"
            className="gap-1.5 h-7 text-xs"
            onClick={() => window.open(article.link, '_blank', 'noopener,noreferrer')}
          >
            <ExternalLink className="h-3 w-3" />
            {t('Read')}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

function EmptyState({ message }: { message: string }) {
  const { t } = useT()
  return (
    <div className="flex flex-col items-center justify-center py-16 text-muted-foreground gap-3">
      <Newspaper className="h-12 w-12 opacity-30" />
      <p className="text-sm">{message}</p>
    </div>
  )
}

function LoadingState() {
  const { t } = useT()
  return (
    <div className="flex items-center justify-center py-16 gap-2 text-muted-foreground">
      <Loader2 className="h-5 w-5 animate-spin" />
      <p className="text-sm">{t('Loading articles\u2026')}</p>
    </div>
  )
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  const { t } = useT()
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-3">
      <AlertCircle className="h-10 w-10 text-destructive/60" />
      <p className="text-sm text-muted-foreground">{message}</p>
      <Button variant="outline" size="sm" onClick={onRetry}>
        {t('Retry')}
      </Button>
    </div>
  )
}

function NewsPage() {
  const { t } = useT()
  const [activeLang, setActiveLang] = useState<LangTab>('all')
  const [articles, setArticles] = useState<RssArticle[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchArticles = async (lang: string) => {
    setLoading(true)
    setError(null)
    try {
      const params = new URLSearchParams({ language: lang, limit: '50', offset: '0' })
      const res = await fetch(`/api/rss/articles?${params}`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const json: RssApiResponse = await res.json()
      if (!json.success) throw new Error(json.data as unknown as string || t('Failed to load articles'))
      setArticles(json.data.articles || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : t('An error occurred'))
      setArticles([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchArticles(activeLang === 'all' ? 'all' : activeLang)
  }, [activeLang])

  return (
    <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">{t('AI News')}</h1>
          <p className="text-muted-foreground mt-2 text-sm sm:text-base">
            {t('Latest AI articles from top sources')}
          </p>
        </div>
      </div>

      <div className="flex gap-2 flex-wrap">
        {LANG_TABS.map((lang) => (
          <Button
            key={lang}
            variant={activeLang === lang ? 'default' : 'outline'}
            size="sm"
            onClick={() => setActiveLang(lang)}
            className="text-xs"
          >
            {t(LANG_LABELS[lang])}
          </Button>
        ))}
      </div>

      {loading ? (
        <LoadingState />
      ) : error ? (
        <ErrorState message={error} onRetry={() => fetchArticles(activeLang)} />
      ) : articles.length === 0 ? (
        <EmptyState message={t('No articles yet. RSS feeds are being fetched in the background.')} />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {articles.map((article) => (
            <ArticleCard key={article.id} article={article} />
          ))}
        </div>
      )}
    </div>
  )
}
