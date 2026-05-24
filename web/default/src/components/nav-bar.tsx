import { Link, useLocation } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/api'
import { useState, useRef, useEffect } from 'react'
import { useAuthStore } from '@/stores/auth-store'
import { useNavMenus } from '@/lib/use-menus'

// Hardcoded fallback nav items (used if API is unavailable)
const FALLBACK_NAV_ITEMS = [
  { to: '/models',     label: 'Models',     icon: '☰' },
  { to: '/pricing',    label: 'Pricing',    icon: '¤' },
  { to: '/rankings',   label: 'Rankings',   icon: '≡' },
  { to: '/apps',       label: 'Apps',       icon: '⊞' },
  { to: '/enterprise', label: 'Enterprise', icon: '◈' },
]

export default function NavBar({ variant = 'default' }: { variant?: 'default' | 'transparent' }) {
  const { t, language, langs, changeLanguage } = useT()
  const { auth } = useAuthStore()
  const location = useLocation()
  const loggedIn = !!auth?.user
  const [mobileOpen, setMobileOpen] = useState(false)
  const [langOpen, setLangOpen] = useState(false)
  const langRef = useRef<HTMLDivElement>(null)

  // Fetch nav menus from API
  const { data: navItems = FALLBACK_NAV_ITEMS } = useNavMenus()

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (langRef.current && !langRef.current.contains(e.target as Node)) setLangOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const isActive = (path: string) => location.pathname.startsWith(path)

  return (
    <header
      className={`sticky top-0 z-50 transition-all duration-300 ${
        variant === 'transparent'
          ? 'bg-white/70 backdrop-blur-xl border-b border-border/20'
          : 'bg-white/80 backdrop-blur-xl border-b border-border/20'
      }`}
      style={{ boxShadow: '0 1px 3px oklch(0 0 0 / 0.03)' }}
    >
      <div className="qc-wrap">
        <div className="flex items-center justify-between h-[4.5rem]">
          {/* ─── Logo ─── */}
          <Link to="/" className="flex items-center gap-3 no-underline group shrink-0">
            <div className="relative w-11 h-11 rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center shadow-lg shadow-orange-500/20 transition-transform duration-300 group-hover:scale-105">
              <svg className="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/>
              </svg>
            </div>
            <div className="flex flex-col">
              <span className="text-lg font-bold tracking-tight text-foreground leading-tight">
                Quantum<span className="bg-gradient-to-r from-amber-500 to-orange-500 bg-clip-text text-transparent">Claw</span>
              </span>
            </div>
          </Link>

          {/* ─── Desktop Nav ─── */}
          <nav className="hidden lg:flex items-center bg-muted/30 rounded-2xl border border-border/10 flex-1 max-w-[55%] mx-8">
            {navItems.map(n => (
              <Link
                key={n.to}
                to={n.to}
                className={`relative flex-1 text-center px-3 py-3 text-lg font-semibold rounded-xl transition-all duration-200 no-underline ${
                  isActive(n.to)
                    ? 'text-foreground bg-white shadow-sm'
                    : 'text-muted-foreground/60 hover:text-foreground hover:bg-white/40'
                }`}
              >
                {t(n.label)}
                {isActive(n.to) && (
                  <span className="absolute bottom-0.5 left-1/2 -translate-x-1/2 w-6 h-0.5 bg-gradient-to-r from-amber-400 to-orange-400 rounded-full" />
                )}
              </Link>
            ))}
          </nav>

          {/* ─── Right side ─── */}
          <div className="flex items-center gap-2 sm:gap-3 shrink-0">
            {/* Language selector */}
            <div className="relative" ref={langRef}>
              <button
                onClick={() => setLangOpen(!langOpen)}
                className="flex items-center gap-2 px-3.5 py-2.5 rounded-xl text-sm font-medium text-muted-foreground/60 hover:text-foreground hover:bg-muted/50 transition-all border border-border/20 hover:border-border/40"
                aria-label="Select language"
              >
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                  <circle cx="12" cy="12" r="10"/>
                  <path d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/>
                </svg>
                <span className="hidden sm:inline text-xs tracking-wide uppercase">{language.length > 8 ? language.substring(0, 8) + '…' : language}</span>
                <svg className={`w-3 h-3 transition-transform duration-200 ${langOpen ? 'rotate-180' : ''}`} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 9l6 6 6-6"/></svg>
              </button>
              {langOpen && (
                <div className="absolute right-0 top-full mt-2 w-44 py-2 bg-white rounded-2xl border border-border/20 shadow-xl shadow-black/5 z-50 overflow-hidden">
                  <div className="px-3 pb-2 mb-1 border-b border-border/10">
                    <span className="text-[10px] font-semibold text-muted-foreground/40 uppercase tracking-[0.15em]">{t('Language')}</span>
                  </div>
                  {langs.map((l, i) => (
                    <button
                      key={l}
                      onClick={() => { changeLanguage(l); setLangOpen(false) }}
                      className={`w-full text-left px-4 py-2.5 text-sm transition-all duration-150 flex items-center justify-between ${
                        language === l
                          ? 'text-[oklch(0.66_0.18_52)] font-semibold bg-gradient-to-r from-amber-50/80 to-orange-50/80'
                          : 'text-muted-foreground hover:text-foreground hover:bg-muted/30'
                      }`}
                    >
                      <span>{l}</span>
                      {language === l && (
                        <svg className="w-4 h-4 text-[oklch(0.66_0.18_52)]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="20 6 9 17 4 12"/>
                        </svg>
                      )}
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Notification Bell */}
            {loggedIn && (
              <Link to="/notifications">
                <button className="relative w-10 h-10 rounded-xl hover:bg-muted/50 flex items-center justify-center transition-colors text-muted-foreground" aria-label="Notifications">
                  <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/>
                  </svg>
                  {unreadCount > 0 && (
                    <span className="absolute -top-0.5 -right-0.5 w-4.5 h-4.5 rounded-full bg-red-500 text-white text-[10px] font-bold flex items-center justify-center shadow-md ring-2 ring-white">
                      {unreadCount > 9 ? '9+' : unreadCount}
                    </span>
                  )}
                </button>
              </Link>
            )}

            {/* Auth */}
            {loggedIn ? (
              <Link to="/dashboard">
                <button className="relative inline-flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200">
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/>
                  </svg>
                  {t('Dashboard')}
                </button>
              </Link>
            ) : (
              <div className="flex items-center gap-2">
                <Link to="/sign-in" className="hidden sm:block">
                  <button className="px-4 py-2.5 rounded-xl text-sm font-medium text-muted-foreground hover:text-foreground border border-border/20 hover:border-border/40 bg-transparent transition-all duration-200">
                    {t('Sign In')}
                  </button>
                </Link>
                <Link to="/sign-in">
                  <button className="relative px-5 py-2.5 rounded-xl text-sm font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200">
                    {t('Get Started')}
                  </button>
                </Link>
              </div>
            )}

            {/* Mobile hamburger */}
            <button
              onClick={() => setMobileOpen(!mobileOpen)}
              className="lg:hidden ml-1 relative w-10 h-10 rounded-xl hover:bg-muted/50 flex items-center justify-center transition-colors text-muted-foreground"
              aria-label="Menu"
            >
              <div className="flex flex-col gap-1">
                <span className={`block w-5 h-0.5 bg-current rounded-full transition-all duration-200 ${mobileOpen ? 'rotate-45 translate-y-1.5' : ''}`} />
                <span className={`block w-5 h-0.5 bg-current rounded-full transition-all duration-200 ${mobileOpen ? 'opacity-0' : ''}`} />
                <span className={`block w-5 h-0.5 bg-current rounded-full transition-all duration-200 ${mobileOpen ? '-rotate-45 -translate-y-1.5' : ''}`} />
              </div>
            </button>
          </div>
        </div>

        {/* ─── Mobile nav ─── */}
        {mobileOpen && (
          <div className="lg:hidden pb-6 pt-2 border-t border-border/20 space-y-0.5">
            {navItems.map(n => (
              <Link
                key={n.to}
                to={n.to}
                onClick={() => setMobileOpen(false)}
                className={`flex items-center gap-3 px-4 py-3.5 text-base font-medium rounded-xl no-underline transition-all ${
                  isActive(n.to)
                    ? 'text-foreground font-semibold bg-gradient-to-r from-amber-50/80 to-orange-50/80'
                    : 'text-muted-foreground hover:text-foreground hover:bg-muted/40'
                }`}
              >
                <span className="text-base w-5 text-center">{n.icon}</span>
                {t(n.label)}
              </Link>
            ))}
            <hr className="my-3 border-border/10" />
            <p className="px-4 text-[10px] font-semibold text-muted-foreground/30 uppercase tracking-[0.15em] mb-2">{t('Language')}</p>
            <div className="flex flex-wrap gap-1 px-4">
              {langs.map(l => (
                <button
                  key={l}
                  onClick={() => { changeLanguage(l); setMobileOpen(false) }}
                  className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${
                    language === l
                      ? 'text-white bg-gradient-to-r from-amber-500 to-orange-500'
                      : 'text-muted-foreground bg-muted/40 hover:bg-muted/60'
                  }`}
                >
                  {l}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </header>
  )
}

