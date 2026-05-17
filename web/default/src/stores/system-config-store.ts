/**
 * QuantumClaw - System Config Store
 * 
 * Persists system branding (name, logo, etc.) across sessions.
 */

import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { DEFAULT_SYSTEM_NAME, DEFAULT_LOGO } from '@/lib/constants'

export interface CurrencyConfig {
  displayInCurrency: boolean
  quotaDisplayType: 'USD' | 'CNY' | 'TOKENS' | 'CUSTOM'
  quotaPerUnit: number
  usdExchangeRate: number
  customCurrencySymbol: string
  customCurrencyExchangeRate: number
}

export interface SystemConfig {
  systemName: string
  logo: string
  footerHtml?: string
  currency: CurrencyConfig
}

const DEFAULT_CURRENCY: CurrencyConfig = {
  displayInCurrency: true,
  quotaDisplayType: 'USD',
  quotaPerUnit: 500000,
  usdExchangeRate: 1,
  customCurrencySymbol: '\u00A4',
  customCurrencyExchangeRate: 1,
}

interface SystemConfigState {
  config: SystemConfig
  loading: boolean
  setConfig: (config: Partial<SystemConfig>) => void
  setLoading: (loading: boolean) => void
}

export const useSystemConfigStore = create<SystemConfigState>()(
  persist(
    (set) => ({
      config: {
        systemName: DEFAULT_SYSTEM_NAME,
        logo: DEFAULT_LOGO,
        currency: { ...DEFAULT_CURRENCY },
      },
      loading: true,
      setConfig: (newConfig) =>
        set((state) => ({
          config: {
            ...state.config,
            ...newConfig,
            currency: {
              ...state.config.currency,
              ...(newConfig.currency ?? {}),
            },
          },
        })),
      setLoading: (loading) => set({ loading }),
    }),
    {
      name: 'qc-system-config',
      partialize: (state) => ({
        config: state.config,
      }),
    },
  ),
)
