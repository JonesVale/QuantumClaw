import i18next from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

import zhCN from '@/i18n/zh-CN.json'
import zhTW from '@/i18n/zh-TW.json'
import en from '@/i18n/en.json'
import fr from '@/i18n/fr.json'
import ja from '@/i18n/ja.json'
import ru from '@/i18n/ru.json'
import vi from '@/i18n/vi.json'

i18next
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      'zh-CN': { translation: zhCN },
      'zh-TW': { translation: zhTW },
      en: { translation: en },
      fr: { translation: fr },
      ja: { translation: ja },
      ru: { translation: ru },
      vi: { translation: vi },
    },
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false,
    },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
    },
  })

export default i18next