import { Component, type ReactNode } from 'react'
import { useRouter } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
  fallback?: ReactNode
  /** 路由 key 变化时自动重置错误状态 */
  routeKey?: string
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: any) {
    console.error('ErrorBoundary caught:', error, info)
  }

  componentDidUpdate(prevProps: Props) {
    // 路由变化时自动重置，避免一次错误崩掉全局
    if (this.state.hasError && this.props.routeKey && this.props.routeKey !== prevProps.routeKey) {
      this.setState({ hasError: false, error: undefined })
    }
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div className="p-8 max-w-2xl mx-auto mt-8">
          <div className="rounded-xl border border-red-200 bg-red-50 dark:bg-red-950/20 dark:border-red-800 p-6">
            <h2 className="text-lg font-bold text-red-700 dark:text-red-400 mb-2">
              Something went wrong
            </h2>
            <p className="text-sm text-red-600 dark:text-red-300 mb-4 font-mono break-all">
              {this.state.error?.message || 'Unknown error'}
            </p>
            <p className="text-xs text-red-500 dark:text-red-400 mb-4 font-mono break-all">
              {this.state.error?.stack?.split('\n').slice(0, 5).join('\n')}
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  this.setState({ hasError: false, error: undefined })
                }}
              >
                Try Again
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => window.location.reload()}
              >
                Reload Page
              </Button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}

/**
 * ErrorBoundaryWithRouter — 自动监听路由变化重置错误状态
 */
export function ErrorBoundaryWithRouter({ children }: { children: ReactNode }) {
  const router = useRouter()
  return <ErrorBoundary routeKey={router.state.location.pathname}>{children}</ErrorBoundary>
}
