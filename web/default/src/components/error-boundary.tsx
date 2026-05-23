import { Component, type ReactNode } from 'react'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
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

  render() {
    if (this.state.hasError) {
      return (
        <div className="p-8 max-w-2xl mx-auto mt-8">
          <div className="rounded-xl border border-red-200 bg-red-50 dark:bg-red-950/20 dark:border-red-800 p-6">
            <h2 className="text-lg font-bold text-red-700 dark:text-red-400 mb-2">
              ⚠️ Page Error
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
                  window.location.reload()
                }}
              >
                Reload Page
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => window.history.back()}
              >
                Go Back
              </Button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
