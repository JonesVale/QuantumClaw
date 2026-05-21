// T_Languages frontend integration
// Fetches translations from backend API and overlays onto i18next resources.
// Falls back to bundled JSON when API is unavailable.

import i18next from 'i18next'

type TranslationMap = Record<string, Record<string, string>>

// Map T_Languages types to i18next language codes
const typeToCode: Record<string, string> = {
  '中文简体': 'zh-CN',
  '中文繁体': 'zh-TW',
  'English': 'en',
  'Français': 'fr',
  '日本語': 'ja',
  'Русский': 'ru',
  'Tiếng Việt': 'vi',
}

const codeToType: Record<string, string> = {
  'zh-CN': '中文简体',
  'zh-TW': '中文繁体',
  'en': 'English',
  'fr': 'Français',
  'ja': '日本語',
  'ru': 'Русский',
  'vi': 'Tiếng Việt',
}

let tlInitialized = false

// Seed all JSON translations into T_Languages via API (requires admin auth)
export async function seedTranslationJson(langType: string, translations: Record<string, string>): Promise<void> {
  try {
    const entries = Object.entries(translations).map(([key, value]) => ({
      lcode: key,
      display: value,
      fromname: 'frontend-seed',
    }))
    const r = await fetch('/api/languages/seed', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ languages_type: langType, entries }),
    })
    const data = await r.json()
    if (data.success) {
      console.log(`T_Languages: seeded ${data.count} entries for ${langType}`)
    }
  } catch {
    // Silently fail — seed requires admin auth
  }
}

// Auto-seed all bundled translations if DB is empty (requires admin auth)
export async function autoSeedIfEmpty(): Promise<void> {
  try {
    const r = await fetch('/api/translations?lang=中文简体')
    if (!r.ok) return
    const data = await r.json()
    if (data.success && Array.isArray(data.data) && data.data.length > 0) {
      // DB already has translations — skip seeding
      return
    }

    // DB is empty — attempt to seed all bundled languages
    const modules: Record<string, () => Promise<{ default: Record<string, string> }>> = {
      '中文简体': () => import('@/i18n/zh-CN.json'),
      '中文繁体': () => import('@/i18n/zh-TW.json'),
      'English': () => import('@/i18n/en.json'),
      'Français': () => import('@/i18n/fr.json'),
      '日本語': () => import('@/i18n/ja.json'),
      'Русский': () => import('@/i18n/ru.json'),
      'Tiếng Việt': () => import('@/i18n/vi.json'),
    }

    for (const [langType, loader] of Object.entries(modules)) {
      try {
        const mod = await loader()
        await seedTranslationJson(langType, mod.default)
      } catch {
        // Skip languages that fail to load
      }
    }
    console.log('T_Languages: auto-seed complete')
  } catch {
    // Silently fail
  }
}

// Load translations from T_Languages API for the given type
export async function loadApiTranslations(langType: string): Promise<Record<string, string> | null> {
  try {
    const r = await fetch(`/api/translations?lang=${encodeURIComponent(langType)}`)
    if (!r.ok) return null
    const data = await r.json()
    if (!data.success) return null
    const items: Array<{ lcode: string; display: string }> = data.data
    const map: Record<string, string> = {}
    for (const item of items) {
      map[item.lcode] = item.display
    }
    return map
  } catch {
    return null
  }
}

// Sync: load all supported languages from API and add them to i18next
export async function syncTranslations(): Promise<void> {
  if (tlInitialized) return

  try {
    // Fetch supported language types
    const r = await fetch('/api/languages')
    if (!r.ok) throw new Error('API unavailable')
    const data = await r.json()
    if (!data.success) throw new Error('API returned error')

    const types: Array<{ languages_type: string }> = data.data

    // For each type, load translations and add to i18next
    for (const t of types) {
      const langType = t.languages_type
      const code = typeToCode[langType]
      if (!code) continue

      const apiTrans = await loadApiTranslations(langType)
      if (apiTrans && Object.keys(apiTrans).length > 0) {
        // Merge API translations into existing i18next resources
        const existing = i18next.getResourceBundle(code, 'translation') || {}
        const merged = { ...existing, ...apiTrans }
        i18next.addResourceBundle(code, 'translation', merged, true, true)
      }
    }

    tlInitialized = true
    console.log(`T_Languages: synced ${types.length} languages`)
  } catch {
    console.log('T_Languages: API unavailable, using bundled JSON')
  }
}
