/**
 * useT() — i18next 包装器，完全向后兼容
 *
 * 保持与旧 useT() 一致的 API：{ t, language, langs, changeLanguage }
 * 内部委托给 i18next，语言切换使用显示名 ↔ 标准码双向映射
 *
 * 零网络请求、零竞态、持久化到 localStorage
 */
import './i18n' // 确保 i18n 初始化
import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  DISPLAY_TO_CODE,
  CODE_TO_DISPLAY,
  ALL_LANG_DISPLAY_NAMES,
} from './i18n'

export function useT() {
  const { t: i18nT, i18n } = useTranslation()

  // 强制重渲染触发器（监听 i18next languageChanged 事件）
  const [, setVer] = useState(0)

  useEffect(() => {
    const cb = () => setVer(v => v + 1)
    i18n.on('languageChanged', cb)
    return () => { i18n.off('languageChanged', cb) }
  }, [i18n])

  const t = useCallback((key: string): string => {
    // i18next 内置回退链：当前语言 → zh-CN → key
    return i18nT(key)
  }, [i18nT])

  const changeLanguage = useCallback(async (displayName: string) => {
    const code = DISPLAY_TO_CODE[displayName]
    if (!code) {
      console.warn(`[useT] unknown language display name: ${displayName}`)
      return
    }
    const currentCode = i18n.language
    if (code === currentCode) return
    await i18n.changeLanguage(code)
    localStorage.setItem('qc_lang', code)
  }, [i18n])

  const currentCode = i18n.language
  const language = CODE_TO_DISPLAY[currentCode] || currentCode

  // langs 使用 ALL_LANG_DISPLAY_NAMES 替换旧的 DB 驱动列表
  const langs = ALL_LANG_DISPLAY_NAMES

  return { t, language, langs, changeLanguage }
}
