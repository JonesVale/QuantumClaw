/**
 * QuantumClaw - PromoCarousel Component
 *
 * Per-page rotating advertisement carousel with glass-morphism styling,
 * automatic rotation, hover pause, arrows, and pagination dots.
 */

import { useState, useEffect, useCallback, useRef } from 'react'
import { Link } from '@tanstack/react-router'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useT } from '@/lib/use-t'

// ─── Types ──────────────────────────────────────────────────────────
export interface AdItem {
  titleKey: string
  descKey: string
  ctaKey: string
  ctaTo: string
  gradient: string  // e.g. 'from-amber-400 to-orange-500'
  icon?: string     // emoji
}

export type PromoPageKey =
  | 'home'
  | 'models'
  | 'pricing'
  | 'rankings'
  | 'apps'
  | 'enterprise'
  | 'dashboard'

// ─── Per-page Ad Data ──────────────────────────────────────────────
// Keys are translation keys resolved via useT()
const PAGE_ADS: Record<PromoPageKey, AdItem[]> = {
  home: [
    {
      titleKey: 'Quantum API Gateway',
      descKey: 'Enterprise-grade AI gateway with 30+ model providers, real-time billing, and intelligent routing.',
      ctaKey: 'Explore Models',
      ctaTo: '/models',
      gradient: 'from-amber-400 to-orange-500',
      icon: '⚡',
    },
    {
      titleKey: 'Multi-Model Fusion',
      descKey: 'Combine GPT-4o, Claude, Gemini and more into a single unified API with automatic failover.',
      ctaKey: 'View Pricing',
      ctaTo: '/pricing',
      gradient: 'from-violet-500 to-purple-600',
      icon: '🧬',
    },
    {
      titleKey: 'Quantum Random Generator',
      descKey: 'True quantum random number generation via ANU QRNG — provably unpredictable.',
      ctaKey: 'Learn More',
      ctaTo: '/quantum',
      gradient: 'from-emerald-500 to-teal-600',
      icon: '🔮',
    },
  ],
  models: [
    {
      titleKey: 'GPT-4o — Free Trial',
      descKey: 'Experience OpenAI\'s flagship multimodal model with vision, audio, and text capabilities.',
      ctaKey: 'Try Now',
      ctaTo: '/playground',
      gradient: 'from-emerald-400 to-green-500',
      icon: '🤖',
    },
    {
      titleKey: 'Claude Opus 4',
      descKey: 'Anthropic\'s most capable model for complex reasoning, code generation, and analysis.',
      ctaKey: 'View Details',
      ctaTo: '/models',
      gradient: 'from-blue-500 to-indigo-600',
      icon: '🧠',
    },
    {
      titleKey: 'DeepSeek V3 — 90% Cheaper',
      descKey: 'Chinese reasoning model at a fraction of the cost. Ideal for bulk inference.',
      ctaKey: 'Compare Pricing',
      ctaTo: '/pricing',
      gradient: 'from-rose-400 to-pink-500',
      icon: '💰',
    },
  ],
  pricing: [
    {
      titleKey: 'Pay-As-You-Go',
      descKey: 'No monthly commitments. Pay only for what you use with transparent per-token pricing.',
      ctaKey: 'Calculate Cost',
      ctaTo: '/pricing',
      gradient: 'from-amber-400 to-orange-500',
      icon: '💳',
    },
    {
      titleKey: 'Enterprise Plan',
      descKey: 'Volume discounts, dedicated support, custom SLAs, and on-premise deployment options.',
      ctaKey: 'Contact Sales',
      ctaTo: '/enterprise',
      gradient: 'from-violet-500 to-purple-600',
      icon: '🏢',
    },
  ],
  rankings: [
    {
      titleKey: 'Weekly Model Rankings',
      descKey: 'Real-time performance data across all major models — speed, quality, and cost efficiency.',
      ctaKey: 'See Rankings',
      ctaTo: '/rankings',
      gradient: 'from-amber-400 to-orange-500',
      icon: '🏆',
    },
    {
      titleKey: 'Best Value Models',
      descKey: 'Top-performing models ranked by cost-per-token for budget-conscious deployments.',
      ctaKey: 'View Pricing',
      ctaTo: '/pricing',
      gradient: 'from-teal-400 to-emerald-500',
      icon: '⭐',
    },
  ],
  apps: [
    {
      titleKey: 'AI Chat Playground',
      descKey: 'Test any model instantly in our interactive chat interface — no API key required.',
      ctaKey: 'Open Playground',
      ctaTo: '/playground',
      gradient: 'from-amber-400 to-orange-500',
      icon: '💬',
    },
    {
      titleKey: 'API Key Management',
      descKey: 'Generate, rotate, and monitor API keys with granular permission controls.',
      ctaKey: 'Manage Keys',
      ctaTo: '/keys',
      gradient: 'from-blue-500 to-indigo-600',
      icon: '🔑',
    },
  ],
  enterprise: [
    {
      titleKey: 'Dedicated Infrastructure',
      descKey: 'Private clusters, custom model fine-tuning, and 24/7 dedicated support team.',
      ctaKey: 'Contact Us',
      ctaTo: '/enterprise',
      gradient: 'from-amber-400 to-orange-500',
      icon: '🏗️',
    },
    {
      titleKey: 'SLA Guaranteed',
      descKey: '99.9% uptime SLA with multi-region failover and comprehensive monitoring.',
      ctaKey: 'Learn More',
      ctaTo: '/enterprise',
      gradient: 'from-emerald-500 to-teal-600',
      icon: '✅',
    },
  ],
  dashboard: [
    {
      titleKey: 'Welcome to QuantumClaw',
      descKey: 'Your centralized AI API hub. Monitor usage, manage keys, and track costs in real time.',
      ctaKey: 'View API Keys',
      ctaTo: '/keys',
      gradient: 'from-amber-400 to-orange-500',
      icon: '👋',
    },
    {
      titleKey: 'Need Help Getting Started?',
      descKey: 'Browse our API documentation or try out models in the playground for free.',
      ctaKey: 'API Docs',
      ctaTo: '/api-docs',
      gradient: 'from-violet-500 to-purple-600',
      icon: '📖',
    },
  ],
}

// ─── PromoCarousel Component ───────────────────────────────────────

interface PromoCarouselProps {
  pageKey: PromoPageKey
  className?: string
}

export function PromoCarousel({ pageKey, className }: PromoCarouselProps) {
  const { t } = useT()
  const ads = PAGE_ADS[pageKey] || PAGE_ADS.home
  const [current, setCurrent] = useState(0)
  const [isPaused, setIsPaused] = useState(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const total = ads.length
  const prev = useCallback(() => setCurrent(c => (c - 1 + total) % total), [total])
  const next = useCallback(() => setCurrent(c => (c + 1) % total), [total])

  // Auto-rotate every 5s, pause on hover
  useEffect(() => {
    if (isPaused || total <= 1) return
    timerRef.current = setInterval(next, 5000)
    return () => { if (timerRef.current) clearInterval(timerRef.current) }
  }, [isPaused, total, next])

  // Reset timer on manual navigation
  const goTo = useCallback((i: number) => {
    setCurrent(i)
    if (timerRef.current) clearInterval(timerRef.current)
  }, [])

  if (total === 0) return null

  const ad = ads[current]

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl bg-white/50 backdrop-blur-sm border border-border/10 shadow-sm transition-shadow duration-300 hover:shadow-md',
        className
      )}
      onMouseEnter={() => setIsPaused(true)}
      onMouseLeave={() => setIsPaused(false)}
    >
      {/* Gradient background */}
      <div className={`absolute inset-0 bg-gradient-to-br ${ad.gradient} opacity-5 rounded-2xl`} />

      {/* Slide content */}
      <div className="relative flex items-center justify-between px-6 py-5 md:px-8 md:py-6">
        {/* Text area */}
        <div className="flex-1 min-w-0 pr-4">
          <div className="flex items-center gap-2 mb-1.5">
            {ad.icon && <span className="text-xl leading-none">{ad.icon}</span>}
            <h3 className="text-base md:text-lg font-semibold text-foreground tracking-tight">
              {t(ad.titleKey)}
            </h3>
          </div>
          <p
            className="text-sm text-muted-foreground/70 mb-3"
            style={{ maxWidth: 'min(50ch, 100%)' }}
          >
            {t(ad.descKey)}
          </p>
          <Link
            to={ad.ctaTo}
            className={`inline-flex items-center gap-1.5 px-4 py-1.5 rounded-lg text-sm font-medium text-white bg-gradient-to-r ${ad.gradient} shadow-sm hover:shadow-md transition-all duration-200 hover:scale-[1.02] active:scale-[0.98]`}
          >
            {t(ad.ctaKey)}
            <ChevronRight className="h-3.5 w-3.5" />
          </Link>
        </div>

        {/* Pagination dots */}
        {total > 1 && (
          <div className="flex flex-col items-center gap-2 shrink-0">
            {/* Up arrow */}
            <button
              onClick={prev}
              className="p-1 rounded-full hover:bg-muted/50 transition-colors text-muted-foreground/50 hover:text-foreground"
              aria-label="Previous ad"
            >
              <ChevronLeft className="h-3.5 w-3.5" />
            </button>
            {/* Dots */}
            <div className="flex flex-col gap-1.5">
              {ads.map((_, i) => (
                <button
                  key={i}
                  onClick={() => goTo(i)}
                  className={cn(
                    'w-1.5 h-1.5 rounded-full transition-all duration-300',
                    i === current
                      ? 'bg-amber-500 scale-125'
                      : 'bg-muted-foreground/20 hover:bg-muted-foreground/40'
                  )}
                  aria-label={`Go to ad ${i + 1}`}
                />
              ))}
            </div>
            {/* Down arrow */}
            <button
              onClick={next}
              className="p-1 rounded-full hover:bg-muted/50 transition-colors text-muted-foreground/50 hover:text-foreground"
              aria-label="Next ad"
            >
              <ChevronRight className="h-3.5 w-3.5" />
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
