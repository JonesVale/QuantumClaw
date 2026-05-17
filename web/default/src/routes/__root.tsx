import { createRootRoute, Outlet } from '@tanstack/react-router'
import { Toaster } from 'sonner'
import { ThemeProvider } from '@/context/theme-provider'
import { ErrorBoundary } from '@/components/error-boundary'

function RootComponent() {
  return (
    <ThemeProvider>
      <ErrorBoundary>
        <Outlet />
      </ErrorBoundary>
      <Toaster duration={5000} />
    </ThemeProvider>
  )
}

export const Route = createRootRoute({
  component: RootComponent,
})
