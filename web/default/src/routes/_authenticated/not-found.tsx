import { createFileRoute, Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { Home, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'

export const Route = createFileRoute('/_authenticated/not-found')({
  component: NotFoundPage,
})

function NotFoundPage() {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4 p-8 text-center">
      <div className="text-6xl font-bold bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">404</div>
      <h1 className="text-2xl font-semibold">{t('Page not found')}</h1>
      <p className="text-muted-foreground max-w-md">{t('The page you are looking for does not exist or has been moved.')}</p>
      <div className="flex gap-3 mt-4">
        <Link to="/dashboard">
          <Button className="gap-2"><Home className="h-4 w-4" />{t('Go Home')}</Button>
        </Link>
        <Button variant="outline" className="gap-2" onClick={() => window.history.back()}>
          <ArrowLeft className="h-4 w-4" />{t('Go Back')}
        </Button>
      </div>
    </div>
  )
}
