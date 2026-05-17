import { createFileRoute } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { FileQuestion } from 'lucide-react'

export const Route = createFileRoute('/_authenticated/not-found')({
  component: NotFoundPage,
})

function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] p-6">
      <FileQuestion className="h-24 w-24 text-muted-foreground/40 mb-6" />
      <h1 className="text-6xl font-bold text-muted-foreground/30 mb-2">404</h1>
      <h2 className="text-xl font-semibold mb-2">Page Not Found</h2>
      <p className="text-muted-foreground mb-6 text-sm">The page you are looking for does not exist or has been moved.</p>
      <Button onClick={() => window.history.back()} variant="outline">
        <ArrowLeft className="mr-2 h-4 w-4" /> Go Back
      </Button>
    </div>
  )
}
