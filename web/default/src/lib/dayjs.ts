/**
 * Dayjs setup with common plugins and locale auto-detection.
 */
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import duration from 'dayjs/plugin/duration'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.extend(duration)

// Auto-detect locale from document lang
function updateLocale() {
  try {
    const lang = document.documentElement.lang || navigator.language || 'en'
    const dayjsLocale = lang.toLowerCase().startsWith('zh') ? 'zh-cn' : 'en'
    if (dayjs.locale() !== dayjsLocale) {
      dayjs.locale(dayjsLocale)
    }
  } catch {
    /* ignore */
  }
}

updateLocale()

export default dayjs
