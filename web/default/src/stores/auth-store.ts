/**
 * QuantumClaw - Auth Store (Zustand)
 * 
 * Manages user authentication state with localStorage persistence.
 */

import { create } from 'zustand'

export interface AuthUser {
  id: number
  username: string
  display_name?: string
  email?: string
  avatar_url?: string
  role: number
  status?: number
  group?: string
  quota?: number
  used_quota?: number
  request_count?: number
  unread_count?: number
  identity_verified?: boolean
  identity_name?: string
  identity_number?: string
  balance?: number
}

interface AuthState {
  auth: {
    user: AuthUser | null
    setUser: (user: AuthUser | null) => void
    reset: () => void
  }
}

function loadSavedUser(): AuthUser | null {
  try {
    if (typeof window !== 'undefined') {
      const saved = window.localStorage.getItem('user')
      return saved ? JSON.parse(saved) : null
    }
  } catch {
    if (typeof window !== 'undefined') {
      window.localStorage.removeItem('user')
    }
  }
  return null
}

export const useAuthStore = create<AuthState>()((set) => ({
  auth: {
    user: loadSavedUser(),
    setUser: (user) =>
      set((state) => {
        if (typeof window !== 'undefined') {
          if (user) {
            window.localStorage.setItem('user', JSON.stringify(user))
            if (user.id) {
              window.localStorage.setItem('uid', String(user.id))
            }
          } else {
            window.localStorage.removeItem('user')
            window.localStorage.removeItem('uid')
          }
        }
        return { ...state, auth: { ...state.auth, user } }
      }),
    reset: () =>
      set((state) => {
        if (typeof window !== 'undefined') {
          window.localStorage.removeItem('user')
          window.localStorage.removeItem('uid')
        }
        return { ...state, auth: { ...state.auth, user: null } }
      }),
  },
}))
