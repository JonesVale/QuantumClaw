/**
 * i18n 初始化 — 数据库驱动 + JSON 文件降级
 *
 * 架构：
 *   启动时 → fetch /api/translations/{lang} → 存入 i18next
 *   如果 DB 不可用 → 回退到构建时打包的 JSON 文件
 *   回退链: DB 翻译 → JSON 文件 → zh-CN → 键名
 *   持久化: localStorage('qc_lang')
 *   运行时编辑: 管理员通过平台设置页面修改翻译，前端定时刷新
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
import de from '@/i18n/de.json'
import es from '@/i18n/es.json'
import fr from '@/i18n/fr.json'
import it from '@/i18n/it.json'
import nl from '@/i18n/nl.json'
import pt from '@/i18n/pt.json'

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
  العربية: 'ar',
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
  ar: 'العربية',
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
  'العربية',
]

// ========== 构建 i18next resources ==========

/**
 * 构建策略：
 *  所有语言: 完整导入所有条目（含 identity 键，不过滤）
 *  英文 identity 条目是正确英文原文，过滤后回退到 zh-CN 会导致显示中文
 *  回退链: 当前语言 → zh-CN → 键名
 */
const resources: Record<string, { translation: Resource }> = {
  'zh-CN': { translation: zhCN as Resource },
  'en':     { translation: en as Resource },
  'zh-TW':  { translation: zhTW as Resource },
  'ja':     { translation: ja as Resource },
  'ko':     { translation: ko as Resource },
  'ru':     { translation: ru as Resource },
  'vi':     { translation: vi as Resource },
  'ar':     { translation: ar as Resource },
  'de':     { translation: de as Resource },
  'es':     { translation: es as Resource },
  'fr':     { translation: fr as Resource },
  'it':     { translation: it as Resource },
  'nl':     { translation: nl as Resource },
  'pt':     { translation: pt as Resource },
}

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
