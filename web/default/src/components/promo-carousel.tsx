/**
 * QuantumClaw - PromoCarousel Component
 *
 * Continuous horizontal scrolling marquee of model/resource ads.
 * Fetches ads from API, falls back to hardcoded defaults.
 * When `large` prop is true (homepage), content is displayed 3x bigger.
 */

import { useRef } from 'react'
import { Link } from '@tanstack/react-router'
import { cn } from '@/lib/utils'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/api'

// ─── Types ──────────────────────────────────────────────────────────
export interface AdItem {
  id?: number
  icon: string
  title: string
  link_url: string
}

export type PromoPageKey =
  | 'home'
  | 'models'
  | 'pricing'
  | 'rankings'
  | 'apps'
  | 'enterprise'
  | 'dashboard'

// ─── Hardcoded fallback ads ─────────────────────────────────────────
const FALLBACK_ADS: AdItem[] = [
  { icon: '🤖', title: 'GPT-4o — Multimodal Vision & Audio', link_url: '/models' },
  { icon: '🧠', title: 'Claude Sonnet 4 — Code & Reasoning', link_url: '/models' },
  { icon: '💎', title: 'DeepSeek V3 — 90% Cost Saving', link_url: '/models' },
  { icon: '🟢', title: 'Gemini 2.5 Pro — Long Context 1M', link_url: '/models' },
  { icon: '⚡', title: 'Groq — Ultra-Fast Inference', link_url: '/models' },
  { icon: '📐', title: 'Mistral Large — Precision & Control', link_url: '/models' },
  { icon: '🔮', title: 'Quantum Computing API — IonQ & IBM', link_url: '/quantum' },
  { icon: '🎲', title: 'Quantum Random Generator — ANU QRNG', link_url: '/quantum' },
  { icon: '🔄', title: 'Multi-Model Fusion — Auto Failover', link_url: '/fusion' },
  { icon: '⚡', title: '99.9% Uptime SLA — Enterprise Ready', link_url: '/enterprise' },
  { icon: '💰', title: 'Pay Per Token — Only for What You Use', link_url: '/pricing' },
  { icon: '🛡', title: 'Enterprise Security — RBAC & Audit', link_url: '/enterprise' },
]

// ─── PromoCarousel Component ───────────────────────────────────────

interface PromoCarouselProps {
  pageKey: PromoPageKey
  large?: boolean
  className?: string
}

export function PromoCarousel({ pageKey, large, className }: PromoCarouselProps) {
  const { t } = useT()
  const ref = useRef<HTMLDivElement>(null)

  const { data: apiAds = [] } = useQuery({
    queryKey: ['promo-ads', pageKey],
    queryFn: async () => {
      try {
        const res = await apiClient.get(`/api/promo-ads?page_key=${pageKey}`, { skipErrorHandler: true } as never)
        if (res.data?.success && Array.isArray(res.data.data)) {
          return res.data.data as AdItem[]
        }
      } catch { /* fallback */ }
      return FALLBACK_ADS
    },
    staleTime: 60_000,
    retry: 1,
  })

  const ads = apiAds.length > 0 ? apiAds : FALLBACK_ADS
  const doubled = [...ads, ...ads]

  return (
    <div
      ref={ref}
      className={cn(
        'relative overflow-hidden rounded-2xl bg-white/50 backdrop-blur-sm border border-border/10 shadow-sm',
        large ? 'py-20 md:py-24' : 'py-14 md:py-16',
        className
      )}
    >
      <div className="absolute inset-0 bg-gradient-to-r from-amber-50/30 via-white/50 to-orange-50/30 rounded-2xl" />
      <div
        className="relative flex overflow-hidden group"
        onMouseEnter={() => { if (ref.current) ref.current.style.setProperty('--play-state', 'paused') }}
        onMouseLeave={() => { if (ref.current) ref.current.style.setProperty('--play-state', 'running') }}
      >
        <div
          className={cn('flex animate-marquee', large ? 'gap-10 md:gap-12' : 'gap-6 md:gap-8')}
          style={{ animationPlayState: 'var(--play-state, running)' }}
        >
          {doubled.map((ad, i) => (
            <Link
              key={i}
              to={ad.link_url}
              className={cn(
                'flex items-center shrink-0 rounded-xl bg-white/80 backdrop-blur border border-border/10 shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 no-underline',
                large
                  ? 'gap-6 md:gap-8 px-12 md:px-14 py-10 md:py-12'
                  : 'gap-4 md:gap-5 px-8 md:px-10 py-6 md:py-7'
              )}
            >
              <span className={cn('leading-none', large ? 'text-5xl md:text-6xl' : 'text-2xl md:text-3xl')}>{ad.icon}</span>
              <span className={cn('font-bold text-foreground whitespace-nowrap', large ? 'text-3xl md:text-4xl' : 'text-lg md:text-xl')}>{ad.title}</span>
            </Link>
          ))}
        </div>
      </div>
      <style>{`
        @keyframes qc-marquee {
          0% { transform: translateX(0); }
          100% { transform: translateX(-50%); }
        }
        .animate-marquee { animation: qc-marquee 50s linear infinite; }
      `}</style>
    </div>
  )
}
