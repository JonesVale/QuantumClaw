/**
 * QuantumClaw Mobile - API Client Layer
 * 
 * Adapted from web/default/src/lib/api.ts for React Native.
 * Key differences:
 *  - Uses expo-secure-store instead of localStorage
 *  - No window/document references
 *  - No toast (sonner) - uses Alert for critical errors
 *  - Navigation reset on 401 instead of window.location
 */

import axios from 'axios'
import * as SecureStore from 'expo-secure-store'
import { Alert } from 'react-native'

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
  baseURL: '',  // Same-origin in dev; configure in production
  withCredentials: true,
  headers: {
    'Cache-Control': 'no-store',
  },
})

// ---------------------------------------------------------------------------
// Token & UID resolution (SecureStore instead of localStorage)
// ---------------------------------------------------------------------------

let _cachedUid: string | null = null
let _cachedToken: string | null = null

export async function initApiClient(): Promise<void> {
  try {
    _cachedUid = await SecureStore.getItemAsync('uid')
    _cachedToken = await SecureStore.getItemAsync('token')
  } catch {
    // SecureStore may fail in Expo Go; fallback silently
  }
}

export async function setStoredCredentials(
  uid: string, 
  token?: string
): Promise<void> {
  _cachedUid = uid
  if (token) _cachedToken = token
  try {
    await SecureStore.setItemAsync('uid', uid)
    if (token) await SecureStore.setItemAsync('token', token)
  } catch { /* noop */ }
}

export async function clearStoredCredentials(): Promise<void> {
  _cachedUid = null
  _cachedToken = null
  try {
    await SecureStore.deleteItemAsync('uid')
    await SecureStore.deleteItemAsync('token')
  } catch { /* noop */ }
}

export function getCachedUid(): string | null {
  return _cachedUid
}

// Navigation reference for 401 redirect (set by App.tsx)
let _navigateToLogin: (() => void) | null = null
export function setNavigateToLogin(fn: () => void): void {
  _navigateToLogin = fn
}

// ---------------------------------------------------------------------------
// Request Interceptor
// ---------------------------------------------------------------------------

apiClient.interceptors.request.use((config) => {
  if (_cachedUid) {
    config.headers['New-Api-User'] = _cachedUid
  }
  // Add Bearer token if available (alternative auth for mobile)
  if (_cachedToken) {
    config.headers['Authorization'] = `Bearer ${_cachedToken}`
  }
  return config
})

// ---------------------------------------------------------------------------
// Response Interceptor
// ---------------------------------------------------------------------------

apiClient.interceptors.response.use(
  (response) => {
    const skipBusiness = (response.config as unknown as Record<string, unknown>)
      ?.skipBusinessError as boolean
    if (
      !skipBusiness &&
      response?.data &&
      typeof response.data.success === 'boolean' &&
      !response.data.success
    ) {
      return Promise.reject(new Error(response.data.message || 'Request failed'))
    }
    return response
  },
  (error) => {
    if (!(error?.config as unknown as Record<string, unknown>)?.skipErrorHandler) {
      const status = error?.response?.status
      if (status === 401) {
        clearStoredCredentials()
        Alert.alert('Session Expired', 'Please login again.')
        _navigateToLogin?.()
      } else {
        const msg =
          error?.response?.data?.message || error?.message || 'Request error'
        Alert.alert('Error', msg)
      }
    }
    return Promise.reject(error)
  },
)

// ---------------------------------------------------------------------------
// Deduplication for GET requests
// ---------------------------------------------------------------------------

const inFlightGet = new Map<string, Promise<unknown>>()
const originalGet = apiClient.get.bind(apiClient)

apiClient.get = ((url: string, config = {}) => {
  const disableDuplicate = (config as unknown as Record<string, unknown>)
    ?.disableDuplicate as boolean
  if (disableDuplicate) return originalGet(url, config)

  const params = (config as unknown as Record<string, unknown>)?.params
    ? JSON.stringify((config as unknown as Record<string, unknown>).params)
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
  if (_cachedUid) {
    headers['New-Api-User'] = _cachedUid
  }
  return headers
}

// ---------------------------------------------------------------------------
// Common API Functions
// ---------------------------------------------------------------------------

export async function getSelf(): Promise<ApiResponse> {
  const res = await apiClient.get('/api/user/self', {
    skipErrorHandler: true,
  } as unknown as Record<string, unknown>)
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
