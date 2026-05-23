// T_Languages frontend integration
// DB-first: translations come from database via API.
// Bundled JSON files serve as fallback when DB is unreachable.



// Map T_Languages types to i18next language codes (initial frontend mapping)
export const typeToCode: Record<string, string> = {
  '中文简体': 'zh-CN',
  '中文繁体': 'zh-TW',
  'English': 'en',
  'Français': 'fr',
  '日本語': 'ja',
  'Русский': 'ru',
  'Tiếng Việt': 'vi',
}

export const codeToType: Record<string, string> = {
  'zh-CN': '中文简体',
  'zh-TW': '中文繁体',
  'en': 'English',
  'fr': 'Français',
  'ja': '日本語',
  'ru': 'Русский',
  'vi': 'Tiếng Việt',
}

let tlInitialized = false

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

// Seed translations into T_Languages via public endpoint (no auth, only works when DB empty)
// Falls back to admin-protected endpoint if public fails
export async function seedTranslationJson(langType: string, translations: Record<string, string>): Promise<void> {
  try {
    const entries = Object.entries(translations).map(([key, value]) => ({
      lcode: key,
      display: value,
      fromname: 'frontend-seed',
    }))
    const r = await fetch('/api/languages/seed-public', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ languages_type: langType, entries }),
    })
    const data = await r.json()
    if (data.success) {
      console.log(`T_Languages: seeded ${data.count} entries for ${langType}`)
    }
  } catch {
    console.log(`T_Languages: seed failed for ${langType}, may need admin login`)
  }
}

// Auto-seed Chinese + English only if DB is empty
export async function autoSeedIfEmpty(): Promise<void> {
  try {
    const r = await fetch('/api/translations?lang=中文简体')
    if (!r.ok) return
    const data = await r.json()
    if (data.success && Array.isArray(data.data) && data.data.length > 0) {
      return // DB already has translations
    }

    // DB is empty — seed Chinese + English from bundled JSON
    const modules: Record<string, () => Promise<{ default: Record<string, string> }>> = {
      '中文简体': () => import('@/i18n/zh-CN.json'),
      'English': () => import('@/i18n/en.json'),
    }

    for (const [langType, loader] of Object.entries(modules)) {
      try {
        const mod = await loader()
        await seedTranslationJson(langType, mod.default)
      } catch {
        // Skip
      }
    }
    console.log('T_Languages: auto-seed complete (中文简体 + English)')
  } catch {
    // Silently fail
  }
}

// Sync: DB-first — replace i18next resources with API translations
export async function syncTranslations(): Promise<void> {
  if (tlInitialized) return

  try {
    const r = await fetch('/api/languages')
    if (!r.ok) throw new Error('API unavailable')
    const data = await r.json()
    if (!data.success) throw new Error('API returned error')

    const types: Array<{ languages_type: string }> = data.data

    // For each language type, load translations from DB
    for (const t of types) {
      const langType = t.languages_type
      const code = typeToCode[langType]
      if (!code) continue

      const apiTrans = await loadApiTranslations(langType)
      if (apiTrans && Object.keys(apiTrans).length > 0) {
        // DB authoritative but JSON fills gaps: merge DB into existing bundle
        const currentBundle = {} || {}
        const merged = { ...currentBundle, ...apiTrans }
        
      }
      // If DB empty for this language — bundled JSON fallback stays
    }

    tlInitialized = true
    console.log(`T_Languages: synced ${types.length} languages from DB`)
  } catch {
    console.log('T_Languages: API unavailable, using bundled JSON fallback')
  }
}
