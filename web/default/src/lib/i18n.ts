/**
 * i18n 初始化 — 纯文件驱动，替换旧的 DB 驱动 useT()
 *
 * 架构：
 *   i18n/*.json (18个文件) → 构建时打包 → i18next 内存字典
 *   回退链: 当前语言 → zh-CN → 键名
 *   持久化: localStorage('qc_lang')
 *   零网络请求、零竞态
 *
 * 语言代码映射：显示名 (Display Name) ↔ 标准码 (ISO Code)
 *   显示名用于 UI 展示（English, 中文简体, 日本語...）
 *   标准码用于 i18next 内部（en, zh-CN, ja...）
 */
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import zhCN from '@/i18n/zh-CN.json'
import en from '@/i18n/en.json'
import zhTW from '@/i18n/zh-TW.json'
import ja from '@/i18n/ja.json'
import ko from '@/i18n/ko.json'
import ru from '@/i18n/ru.json'
import vi from '@/i18n/vi.json'
import ar from '@/i18n/ar.json'
import hi from '@/i18n/hi.json'
import de from '@/i18n/de.json'
import es from '@/i18n/es.json'
import fr from '@/i18n/fr.json'
import it from '@/i18n/it.json'
import nl from '@/i18n/nl.json'
import pt from '@/i18n/pt.json'
import tr from '@/i18n/tr.json'
import th from '@/i18n/th.json'
import id from '@/i18n/id.json'

// ========== 语言显示名 ↔ 标准码双向映射 ==========

type Resource = Record<string, string>

export const DISPLAY_TO_CODE: Record<string, string> = {
  English: 'en',
  '中文简体': 'zh-CN',
  '中文繁體': 'zh-TW',
  日本語: 'ja',
  한국어: 'ko',
  Русский: 'ru',
  'Tiếng Việt': 'vi',
  Français: 'fr',
  Español: 'es',
  Deutsch: 'de',
  Italiano: 'it',
  Português: 'pt',
  Nederlands: 'nl',
  Türkçe: 'tr',
  ไทย: 'th',
  العربية: 'ar',
  हिन्दी: 'hi',
  'Bahasa Indonesia': 'id',
}

export const CODE_TO_DISPLAY: Record<string, string> = {
  en: 'English',
  'zh-CN': '中文简体',
  'zh-TW': '中文繁體',
  ja: '日本語',
  ko: '한국어',
  ru: 'Русский',
  vi: 'Tiếng Việt',
  fr: 'Français',
  es: 'Español',
  de: 'Deutsch',
  it: 'Italiano',
  pt: 'Português',
  nl: 'Nederlands',
  tr: 'Türkçe',
  th: 'ไทย',
  ar: 'العربية',
  hi: 'हिन्दी',
  id: 'Bahasa Indonesia',
}

/** 所有语言的显示名列表（固定排序，用于语言选择下拉） */
export const ALL_LANG_DISPLAY_NAMES: string[] = [
  'English',
  '中文简体',
  '中文繁體',
  '日本語',
  '한국어',
  'Русский',
  'Tiếng Việt',
  'Français',
  'Español',
  'Deutsch',
  'Italiano',
  'Português',
  'Nederlands',
  'Türkçe',
  'ไทย',
  'العربية',
  'हिन्दी',
  'Bahasa Indonesia',
]

// ========== 构建 i18next resources ==========

/**
 * 构建策略：
 *  - zh-CN: 完整导入（含 identity 键，用作终极回退）
 *  - 其他语言: 只保留 value !== key 的已翻译条目
 *    这样未翻译的键会走 fallbackLng 回退到 zh-CN
 */
const resources: Record<string, { translation: Resource }> = {
  'zh-CN': { translation: zhCN as Resource },
}

function addLang(code: string, data: Resource) {
  const translated = Object.fromEntries(
    Object.entries(data).filter(([k, v]) => v !== k && v !== '')
  )
  if (Object.keys(translated).length > 0) {
    resources[code] = { translation: translated }
  }
}

addLang('en', en as Resource)
addLang('zh-TW', zhTW as Resource)
addLang('ja', ja as Resource)
addLang('ko', ko as Resource)
addLang('ru', ru as Resource)
addLang('vi', vi as Resource)
addLang('ar', ar as Resource)
addLang('hi', hi as Resource)
addLang('de', de as Resource)
addLang('es', es as Resource)
addLang('fr', fr as Resource)
addLang('it', it as Resource)
addLang('nl', nl as Resource)
addLang('pt', pt as Resource)
addLang('tr', tr as Resource)
addLang('th', th as Resource)
addLang('id', id as Resource)

// ========== 初始化 ==========

// 从 localStorage 读取偏好
const saved = localStorage.getItem('qc_lang')
// 无偏好时根据浏览器语言自动检测
function detectLang(): string {
  if (typeof navigator === 'undefined') return 'zh-CN'
  const nl = navigator.language
  // 标准 BCP 47 标签（如 zh-CN、en-US）直接匹配
  if (nl && resources[nl.substring(0, 2)]) {
    if (nl.startsWith('zh')) return nl.substring(0, 5) // zh-CN or zh-TW
    return nl.substring(0, 2)
  }
  return 'zh-CN'
}

const initialCode = saved || detectLang()

i18n.use(initReactI18next).init({
  resources,
  fallbackLng: 'zh-CN',
  lng: initialCode,
  interpolation: { escapeValue: false },
  returnNull: false,
  returnEmptyString: false,
})

// 首次写入 localStorage
if (!saved) {
  localStorage.setItem('qc_lang', initialCode)
}

export default i18n
