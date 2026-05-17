/**
 * QuantumClaw - API Client Layer
 * 
 * Handles all HTTP communication with the backend server.
 * Uses cookie-based authentication with session cookies.
 */

import axios from 'axios'
import i18next from 'i18next'
import { toast } from 'sonner'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

// ---------------------------------------------------------------------------
// Axios Instance
// ---------------------------------------------------------------------------

const apiClient = axios.create({
  baseURL: '',
  withCredentials: true,
  headers: {
    'Cache-Control': 'no-store',
  },
})

// ---------------------------------------------------------------------------
// Request Interceptor - attach user identity header
// ---------------------------------------------------------------------------

function resolveUserId(): string | null {
  try {
    return typeof window !== 'undefined'
      ? window.localStorage.getItem('uid')
      : null
  } catch {
    return null
  }
}

apiClient.interceptors.request.use((config) => {
  const uid = resolveUserId()
  if (uid) {
    config.headers['New-Api-User'] = uid
  }
  return config
})

// ---------------------------------------------------------------------------
// Response Interceptor - unified error handling
// ---------------------------------------------------------------------------

apiClient.interceptors.response.use(
  (response) => {
    const skipBusiness = (response.config as Record<string, unknown>)
      ?.skipBusinessError as boolean
    if (
      !skipBusiness &&
      response?.data &&
      typeof response.data.success === 'boolean' &&
      !response.data.success
    ) {
      toast.error(response.data.message || 'Request failed')
    }
    return response
  },
  (error) => {
    if (!(error?.config as Record<string, unknown>)?.skipErrorHandler) {
      const status = error?.response?.status
      if (status === 401) {
        toast.error(i18next.t('Session expired!'))
        try {
          window.localStorage.removeItem('user')
          window.localStorage.removeItem('uid')
          window.location.href = '/sign-in'
        } catch {
          /* empty */
        }
      } else {
        const msg =
          error?.response?.data?.message || error?.message || 'Request error'
        toast.error(msg)
      }
    }
    return Promise.reject(error)
  },
)

// ---------------------------------------------------------------------------
// Request deduplication for GET
// ---------------------------------------------------------------------------

const inFlightGet = new Map<string, Promise<unknown>>()
const originalGet = apiClient.get.bind(apiClient)

apiClient.get = ((url: string, config = {}) => {
  const disableDuplicate = (config as Record<string, unknown>)
    ?.disableDuplicate as boolean
  if (disableDuplicate) return originalGet(url, config)

  const params = (config as Record<string, unknown>)?.params
    ? JSON.stringify((config as Record<string, unknown>).params)
    : '{}'
  const key = `${url}?${params}`

  if (inFlightGet.has(key)) return inFlightGet.get(key)!

  const req = originalGet(url, config).finally(() => inFlightGet.delete(key))
  inFlightGet.set(key, req)
  return req
}) as typeof apiClient.get

// ---------------------------------------------------------------------------
// Common Headers (for SSE, etc.)
// ---------------------------------------------------------------------------

export function getCommonHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  const uid = resolveUserId()
  if (uid) {
    headers['New-Api-User'] = uid
  }
  return headers
}

// ---------------------------------------------------------------------------
// Common API Functions
// ---------------------------------------------------------------------------

export async function getSelf(): Promise<ApiResponse> {
  const res = await apiClient.get('/api/user/self', {
    skipErrorHandler: true,
  } as Record<string, unknown>)
  return res.data
}

export async function getStatus(): Promise<Record<string, unknown>> {
  const res = await apiClient.get('/api/status')
  return res.data?.data as Record<string, unknown>
}

export async function getNotice(): Promise<ApiResponse<string>> {
  const res = await apiClient.get('/api/notice')
  return res.data
}

export default apiClient
