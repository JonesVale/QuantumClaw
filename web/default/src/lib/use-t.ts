// useT() — T_Languages-driven translation hook
// Pure custom implementation, no i18next/CultureInfo
import { useState, useEffect, useCallback } from 'react'

// === Module-level state (shared across all components) ===
let cachedLangs: string[] = []
let cachedDict: Record<string, string> = {}
let currentLang = 'English'
let version = 0                     // bumped on every language switch
const listeners = new Set<() => void>()  // all active useT instances

async function fetchLangs(): Promise<string[]> {
  try {
    const r = await fetch('/api/languages')
    const data = await r.json()
    if (data.success && Array.isArray(data.data)) {
      return data.data.map((l: any) => l.languages_type)
    }
  } catch { /* ignore */ }
  return ['English', '中文简体']
}

async function fetchTranslations(lang: string): Promise<Record<string, string>> {
  try {
    const r = await fetch(`/api/translations?lang=${encodeURIComponent(lang)}`)
    const data = await r.json()
    if (data.success && data.data) {
      if (typeof data.data === 'object' && !Array.isArray(data.data)) {
        return data.data
      }
      if (Array.isArray(data.data)) {
        const m: Record<string, string> = {}
        for (const item of data.data) {
          m[item.lcode || item.LCode] = item.display || item.Display
        }
        return m
      }
    }
  } catch { /* ignore */ }
  return {}
}

// Preload once on module load
fetchLangs().then(l => { cachedLangs = l })
fetchTranslations(currentLang).then(d => { cachedDict = d })

export function useT() {
  // version in state forces re-render when it changes
  const [, setVer] = useState(0)
  const [langs, setLangs] = useState<string[]>(cachedLangs.length > 0 ? cachedLangs : ['English', '中文简体'])

  // Register/unregister this instance into the global listener pool
  useEffect(() => {
    const cb = () => setVer(v => v + 1)
    listeners.add(cb)
    return () => { listeners.delete(cb) }
  }, [])

  // Sync available languages (one-time)
  useEffect(() => {
    if (cachedLangs.length === 0) {
      fetchLangs().then(l => { cachedLangs = l; setLangs(l); notifyAll() })
    }
  }, [])

  const t = useCallback((key: string): string => {
    return cachedDict[key] ?? key
  }, [])

  const changeLanguage = useCallback(async (lang: string) => {
    if (lang === currentLang) return
    currentLang = lang
    cachedDict = await fetchTranslations(lang)
    notifyAll()
  }, [])

  return { t, language: currentLang, langs, changeLanguage }
}

function notifyAll() {
  version++
  for (const cb of listeners) cb()
}
