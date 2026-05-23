// useT() — T_Languages-driven translation hook
// Pure custom implementation, no i18next/CultureInfo
import { useState, useEffect, useCallback } from 'react'

let cachedLangs: string[] = []
let cachedDict: Record<string, string> = {}
let currentLang = 'English'
let loadingPromise: Promise<void> | null = null

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
      // data.data is a map/dict
      if (typeof data.data === 'object' && !Array.isArray(data.data)) {
        return data.data
      }
      // Fallback: array of {lcode, display}
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

// Preload: initialize on import (runs once)
fetchLangs().then(langs => { cachedLangs = langs })
fetchTranslations(currentLang).then(d => { cachedDict = d })

export function useT() {
  const [, setTick] = useState(0)
  const [langs, setLangs] = useState<string[]>(cachedLangs.length > 0 ? cachedLangs : ['English', '中文简体'])

  // Sync available languages
  useEffect(() => {
    if (cachedLangs.length === 0) {
      fetchLangs().then(l => { cachedLangs = l; setLangs(l); setTick(t => t + 1) })
    }
  }, [])

  const t = useCallback((key: string): string => {
    return cachedDict[key] ?? key
  }, [])

  const setLanguage = useCallback(async (lang: string) => {
    if (lang === currentLang) return
    currentLang = lang
    cachedDict = await fetchTranslations(lang)
    setTick(t => t + 1)
  }, [])

  return { t, language: currentLang, langs, changeLanguage: setLanguage }
}
