/**
 * Theme Provider - manages dark/light/system theme.
 * Applies theme class to <html> element and persists via cookie.
 */

import { createContext, useContext, useEffect, useState, useMemo } from 'react'
import { getCookie, setCookie, removeCookie } from '@/lib/cookies'
import { TooltipProvider } from '@/components/ui/tooltip'

type Theme = 'dark' | 'light' | 'system'
type ResolvedTheme = Exclude<Theme, 'system'>

const THEME_COOKIE = 'qc-theme'
const COOKIE_MAX_AGE = 60 * 60 * 24 * 365

interface ThemeState {
  theme: Theme
  resolvedTheme: ResolvedTheme
  setTheme: (theme: Theme) => void
}

const ThemeContext = createContext<ThemeState>({
  theme: 'system',
  resolvedTheme: 'light',
  setTheme: () => {},
})

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, _setTheme] = useState<Theme>(
    () => (getCookie(THEME_COOKIE) as Theme) || 'system',
  )

  const resolvedTheme = useMemo((): ResolvedTheme => {
    if (theme === 'system') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light'
    }
    return theme
  }, [theme])

  useEffect(() => {
    const root = window.document.documentElement
    const mq = window.matchMedia('(prefers-color-scheme: dark)')

    const apply = (t: ResolvedTheme) => {
      root.classList.remove('light', 'dark')
      root.classList.add(t)
    }

    apply(resolvedTheme)

    const handler = () => {
      if (theme === 'system') apply(mq.matches ? 'dark' : 'light')
    }
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [theme, resolvedTheme])

  const setTheme = (t: Theme) => {
    if (t === 'system') removeCookie(THEME_COOKIE)
    else setCookie(THEME_COOKIE, t, COOKIE_MAX_AGE)
    _setTheme(t)
  }

  return (
    <ThemeContext value={{ theme, resolvedTheme, setTheme }}>
      <TooltipProvider>
        {children}
      </TooltipProvider>
    </ThemeContext>
  )
}

export const useTheme = () => useContext(ThemeContext)
