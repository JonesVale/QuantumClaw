/**
 * QuantumClaw - Application Entry Point
 */

import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import { AxiosError } from 'axios'
import {
  QueryCache,
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import i18next from 'i18next'
import { toast, Toaster } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { getStatus } from '@/lib/api'
import { ThemeProvider } from '@/context/theme-provider'
import '@/lib/dayjs'
import { applyFaviconToDom } from '@/lib/dom-utils'
import './i18n/config'
import { routeTree } from './routeTree.gen'
import './styles/index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        if (failureCount >= 0 && import.meta.env.DEV) return false
        if (failureCount > 3 && import.meta.env.PROD) return false
        return !(
          error instanceof AxiosError &&
          [401, 403].includes(error.response?.status ?? 0)
        )
      },
      refetchOnWindowFocus: import.meta.env.PROD,
      staleTime: 10 * 1000,
    },
    mutations: {
      onError: (error) => {
        if (error instanceof AxiosError && error.response?.status === 304) {
          toast.error(i18next.t('Content not modified!'))
        }
      },
    },
  },
  queryCache: new QueryCache({
    onError: (error) => {
      if (error instanceof AxiosError) {
        if (error.response?.status === 401) {
          toast.error(i18next.t('Session expired!'))
          useAuthStore.getState().auth.reset()
          window.location.href = '/sign-in'
        }
        if (error.response?.status === 500) {
          toast.error(i18next.t('Internal Server Error!'))
        }
      }
    },
  }),
})

const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0,
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

// Brand initialization
;(function initBranding() {
  try {
    if (typeof window === 'undefined') return
    const apply = (name: string) => {
      document.title = name
    }
    try {
      const saved = localStorage.getItem('status')
      if (saved) {
        const s = JSON.parse(saved)
        if (s?.system_name) apply(s.system_name)
        if (s?.logo) applyFaviconToDom(s.logo)
      }
    } catch {
      /* empty */
    }
    getStatus()
      .then((s) => {
        if (s?.system_name) {
          apply(s.system_name as string)
          localStorage.setItem('status', JSON.stringify(s))
        }
        if (s?.logo) applyFaviconToDom(s.logo as string)
      })
      .catch(() => {})
  } catch {
    /* empty */
  }
})()

const rootElement = document.getElementById('root')!
const root = ReactDOM.createRoot(rootElement)
root.render(
  <StrictMode>
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
        <Toaster />
      </QueryClientProvider>
    </ThemeProvider>
  </StrictMode>,
)
