import { cn } from '@/lib/utils'

// Provider type → Emoji mapping
const providerIcons: Record<number, string> = {
  1: '🤖',   // OpenAI
  2: '🦙',   // API2D
  3: '🌐',   // Azure
  4: '☁️',   // CloseAI
  5: '🟦',   // OpenAISB
  6: '🟩',   // OpenAIMAX
  7: '🤯',   // OhMyGPT
  8: '🧮',   // Custom
  9: '🔄',   // AIPROXY
  10: '🕸️',  // API
  11: '🌊',   // AIGC2D
  12: '📡',   // API2GPT
  13: '🧿',   // AIGC2D (old)
  14: '🧠',   // Anthropic
  15: '🔍',   // Baidu
  16: '🦆',   // DashScope / Alibaba
  17: '🔴',   // Xunfei Spark
  18: '☀️',   // Zhipu / ChatGLM
  19: '🌙',   // Tencent
  20: '💎',   // Baichuan
  21: '🐉',   // MiniMax / Hailuo
  22: '💜',   // DeepSeek
  23: '🌿',   // Moonshot / Kimi
  24: '🍊',   // Stepfun / Step
  25: '⚡',   // Groq
  26: '🟣',   // Together AI
  27: '🟢',   // Gemini / Google
  28: '🦄',   // zero-one
  33: '📐',   // Mistral
  34: '🎨',   // Stability AI
  35: '🧩',   // Perplexity
  36: '🔥',   // Fireworks AI
  40: '🦅',   // Claude
  41: '⚡',   // Grok / xAI
  44: '🔮',   // Cohere
  45: '🟠',   // Cloudflare
  46: '🟧',   // Replicate
  47: '📝',   // Notion
  50: '🪐',   // Meta / Llama
  99: '❓',   // Other / Unknown
}

// Fallback: try to infer from provider name
function iconFromName(name: string): string {
  const lower = name.toLowerCase()
  if (lower.includes('openai')) return '🤖'
  if (lower.includes('anthropic') || lower.includes('claude')) return '🧠'
  if (lower.includes('deepseek')) return '💜'
  if (lower.includes('google') || lower.includes('gemini')) return '🟢'
  if (lower.includes('groq')) return '⚡'
  if (lower.includes('azure') || lower.includes('microsoft')) return '☁️'
  if (lower.includes('baidu') || lower.includes('文心')) return '🔍'
  if (lower.includes('zhipu') || lower.includes('glm') || lower.includes('智谱')) return '☀️'
  if (lower.includes('moonshot') || lower.includes('kimi')) return '🌿'
  if (lower.includes('baichuan') || lower.includes('百川')) return '💎'
  if (lower.includes('minimax') || lower.includes('hailuo')) return '🐉'
  if (lower.includes('meta') || lower.includes('llama')) return '🪐'
  if (lower.includes('mistral')) return '📐'
  if (lower.includes('cohere')) return '🔮'
  if (lower.includes('stability')) return '🎨'
  if (lower.includes('perplexity')) return '🧩'
  if (lower.includes('fireworks')) return '🔥'
  if (lower.includes('together')) return '🟣'
  if (lower.includes('replicate')) return '🟧'
  if (lower.includes('cloudflare')) return '🟠'
  if (lower.includes('tencent') || lower.includes('腾讯')) return '🌙'
  if (lower.includes('stepfun') || lower.includes('step')) return '🍊'
  if (lower.includes('xunfei') || lower.includes('讯飞') || lower.includes('spark')) return '🔴'
  if (lower.includes('xiaohongshu') || lower.includes('rednote')) return '📕'
  if (lower.includes('01') || lower.includes('yi-')) return '🦄'
  if (lower.includes('xai') || lower.includes('grok')) return '⚡'
  return '❓'
}

export interface ProviderIconProps {
  type?: number
  name?: string
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

const sizeMap = {
  sm: 'text-lg',
  md: 'text-2xl',
  lg: 'text-4xl',
}

export function ProviderIcon({ type, name, className, size = 'md' }: ProviderIconProps) {
  const icon = type !== undefined && type !== null
    ? (providerIcons[type] || '❓')
    : (name ? iconFromName(name) : '❓')

  return (
    <span
      className={cn(
        'inline-flex items-center justify-center',
        sizeMap[size],
        className
      )}
      role="img"
      aria-label={name || `Provider type ${type}`}
    >
      {icon}
    </span>
  )
}

export { providerIcons }
