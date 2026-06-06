/**
 * QuantumClaw Mobile - Auth Store (Zustand)
 * 
 * Adapted from web/default/src/stores/auth-store.ts for React Native.
 * Uses expo-secure-store instead of localStorage.
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
  balance?: number
  request_count?: number
  unread_count?: number
}

export type UserRole = 'consumer' | 'provider' | 'enterprise'

export interface AuthState {
  user: AuthUser | null
  role: UserRole  // current active role
  isLoaded: boolean
  setUser: (user: AuthUser | null) => void
  setRole: (role: UserRole) => void
  setLoaded: (loaded: boolean) => void
  reset: () => void
}

export const useAuthStore = create<AuthState>()((set) => ({
  user: null,
  role: 'consumer',
  isLoaded: false,
  setUser: (user) => set({ user, isLoaded: true }),
  setRole: (role) => set({ role }),
  setLoaded: (loaded) => set({ isLoaded: loaded }),
  reset: () => set({ user: null, role: 'consumer', isLoaded: true }),
}))
