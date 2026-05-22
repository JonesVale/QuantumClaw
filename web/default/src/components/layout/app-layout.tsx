/**

 * QuantumClaw - Main Application Layout

 * 

 * Provides the sidebar + header + content area structure

 * for all authenticated pages.

 */



import { Suspense, useCallback, useMemo, useState, useEffect } from 'react'

import { Link, Outlet, useLocation, useRouter } from '@tanstack/react-router'

import { useTranslation } from 'react-i18next'

import {

  LayoutDashboard,

  Network,

  Key,

  Users,

  Box,

  ScrollText,

  Settings,

  Ticket,

  Wallet,

  User,

  Moon,

  Sun,

  LogOut,

  ChevronLeft,

  ChevronRight,

  ChevronRight as ChevronRightSmall,

  Menu,

  Zap,

  MessageSquare,

  Globe,

  Gift,

  CreditCard,

  ClipboardList,

  DollarSign,

  Search,

  Bell,

  Home,
  Info,
  Newspaper,
  Activity,
  BookOpen,
  TrendingUp,
  Truck,
  Wrench,
  Atom,
  GitCompare,
  Sparkles,

} from 'lucide-react'

import { Button } from '@/components/ui/button'

import { ScrollArea } from '@/components/ui/scroll-area'

import { Avatar, AvatarFallback } from '@/components/ui/avatar'

import { Badge } from '@/components/ui/badge'

import { Skeleton } from '@/components/ui/skeleton'

import {

  DropdownMenu,

  DropdownMenuContent,

  DropdownMenuItem,

  DropdownMenuLabel,

  DropdownMenuSeparator,

  DropdownMenuTrigger,

} from '@/components/ui/dropdown-menu'

import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'

import { useTheme } from '@/context/theme-provider'

import { useAuthStore } from '@/stores/auth-store'

import { signOut } from '@/lib/api-extended'

import { cn } from '@/lib/utils'
import { CustomerServiceFloating } from '@/components/customer-service'

import i18next from 'i18next'



// ---------------------------------------------------------------------------

// Sidebar Navigation Items

// ---------------------------------------------------------------------------



interface NavItem {

  path: string

  icon: React.ElementType

  labelKey: string

  adminOnly?: boolean

}



// Product pages shown in top navigation bar
const PRODUCT_ITEMS: NavItem[] = [
  { path: '/dashboard', icon: LayoutDashboard, labelKey: 'Dashboard' },
  { path: '/playground', icon: MessageSquare, labelKey: 'Playground' },
  { path: '/models', icon: Box, labelKey: 'Models' },
  { path: '/rankings', icon: TrendingUp, labelKey: 'Rankings' },
  { path: '/pricing', icon: DollarSign, labelKey: 'Pricing' },
  { path: '/quantum', icon: Atom, labelKey: 'Quantum' },
  { path: '/fusion', icon: GitCompare, labelKey: 'Fusion' },
  { path: '/apps', icon: Sparkles, labelKey: 'Apps' },
]

// Admin/management pages (sidebar)
const NAV_ITEMS: NavItem[] = [
  { path: '/channels', icon: Network, labelKey: 'Channels' },
  { path: '/keys', icon: Key, labelKey: 'API Keys' },
  { path: '/users', icon: Users, labelKey: 'Users', adminOnly: true },
  { path: '/logs', icon: ScrollText, labelKey: 'Usage Logs' },
  { path: '/redemption', icon: Ticket, labelKey: 'Redemption Codes', adminOnly: true },
  { path: '/distributors', icon: Truck, labelKey: 'Distributors', adminOnly: true },
  { path: '/admin-tools', icon: Wrench, labelKey: 'Admin Tools', adminOnly: true },
  { path: '/monitoring', icon: Activity, labelKey: 'Monitoring' },
  { path: '/profit', icon: TrendingUp, labelKey: 'Channel Profit', adminOnly: true },
  { path: '/news', icon: Newspaper, labelKey: 'AI News' },
]



const SETTINGS_ITEMS: NavItem[] = [

  { path: '/profile', icon: User, labelKey: 'Profile' },

  { path: '/wallet', icon: Wallet, labelKey: 'Wallet' },

  { path: '/billing', icon: DollarSign, labelKey: 'Billing' },

  { path: '/checkin', icon: Gift, labelKey: 'Daily Check-in' },

  { path: '/subscription', icon: CreditCard, labelKey: 'Subscriptions' },

  { path: '/tasks', icon: ClipboardList, labelKey: 'Task Logs', adminOnly: true },

  { path: '/settings', icon: Settings, labelKey: 'Settings', adminOnly: true },

  { path: '/api-docs', icon: BookOpen, labelKey: 'API Docs' },
  { path: '/about', icon: Info, labelKey: 'About' },

]



// ---------------------------------------------------------------------------

// Sidebar Component

// ---------------------------------------------------------------------------



function SidebarNav({ collapsed, onToggle }: { collapsed: boolean; onToggle: () => void }) {

  const location = useLocation()

  const { t } = useTranslation()

  const { auth } = useAuthStore()

  const isAdmin = auth.user?.role === 100





  const isActive = useCallback(

    (path: string) => location.pathname === path,

    [location.pathname]

  )



  const renderNavItem = (item: NavItem) => {

    if (item.adminOnly && !isAdmin) return null



    const Icon = item.icon

    const active = isActive(item.path)



    const linkContent = (

      <Link

        to={item.path}

        className={cn(

          'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',

          active

            ? 'bg-sidebar-accent text-sidebar-accent-foreground'

            : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'

        )}

      >

        <Icon className="h-4 w-4 shrink-0" />

        {!collapsed && <span>{t(item.labelKey)}</span>}

      </Link>

    )



    if (collapsed) {

      return (

        <Tooltip key={item.path} delayDuration={0}>

          <TooltipTrigger asChild>{linkContent}</TooltipTrigger>

          <TooltipContent side="right">{t(item.labelKey)}</TooltipContent>

        </Tooltip>

      )

    }



    return <div key={item.path}>{linkContent}</div>

  }



  return (

    <div className="flex h-full flex-col border-r bg-sidebar">

      {/* Brand Header */}

      <div className="flex h-14 items-center border-b px-3">

        <Link to="/" className="flex items-center gap-2">

          <img src="/logo.webp" alt="QuantumClaw" className="h-8 w-8 rounded-lg object-cover" />

          {!collapsed && (

            <span className="text-base font-bold tracking-tight">{t('QuantumClaw')}</span>

          )}

        </Link>

      </div>



      {/* Navigation */}

      <ScrollArea className="flex-1 px-3 py-2">

        <div className="space-y-1">

          {NAV_ITEMS.map(renderNavItem)}

          <div className="my-2 border-t" />

          {SETTINGS_ITEMS.map(renderNavItem)}

        </div>

      </ScrollArea>



      {/* Collapse Toggle */}

      <div className="border-t p-2">

        <Button

          variant="ghost"

          size="icon"

          className="w-full"

          onClick={onToggle}

        >

          {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}

        </Button>

      </div>

    </div>

  )

}



// ---------------------------------------------------------------------------

// Header Component

// ---------------------------------------------------------------------------



// ── Breadcrumbs ────────────────────────────────────────────────────

const breadcrumbMap: Record<string, string> = {

  '/dashboard': 'Dashboard',

  '/channels': 'Channels',

  '/keys': 'API Keys',

  '/models': 'Models',
  '/rankings': 'Rankings',
  '/pricing': 'Pricing',

  '/users': 'Users',

  '/logs': 'Usage Logs',

  '/redemption': 'Redemption Codes',

  '/playground': 'Playground',
  '/monitoring': 'Monitoring',
  '/news': 'AI News',

  '/profile': 'Profile',

  '/wallet': 'Wallet',

  '/billing': 'Billing',

  '/checkin': 'Daily Check-in',

  '/subscription': 'Subscriptions',

  '/tasks': 'Task Logs',

  '/settings': 'Settings',

}



function Breadcrumbs({ pathname }: { pathname: string }) {

  const path = breadcrumbMap[pathname]

  if (!path) return null

  return (

    <nav className="flex items-center gap-1 text-xs md:text-sm text-muted-foreground mb-3" aria-label="Breadcrumb">

      <Link to="/" className="hover:text-foreground transition-colors">

        <Home className="h-3.5 w-3.5" />

      </Link>

      <ChevronRightSmall className="h-3 w-3 mx-0.5" />

      <span className="font-medium text-foreground">{path}</span>

    </nav>

  )

}



function AppHeader({ onMobileMenuToggle }: { onMobileMenuToggle: () => void }) {

  const { t, i18n } = useTranslation()

  const { resolvedTheme, setTheme, theme } = useTheme()

  const { auth } = useAuthStore()

  const router = useRouter()

  const location = useLocation()



  const handleSignOut = useCallback(async () => {

    try {

      await signOut()

    } catch {

      /* ignore */

    }

    auth.reset()

    router.navigate({ to: '/sign-in' })

  }, [auth, router])



  // Fetch available languages from DB for language selector
  const [dbLanguages, setDbLanguages] = useState<string[]>([])
  useEffect(() => {
    fetch('/api/languages')
      .then(r => r.json())
      .then(data => {
        if (data.success && Array.isArray(data.data)) {
          setDbLanguages(data.data.map((l: { languages_type: string }) => l.languages_type))
        }
      })
      .catch(() => {})
  }, [])

  // Map T_Languages type to i18next code
  const typeToCode: Record<string, string> = {
    '中文简体': 'zh-CN',
    '中文繁体': 'zh-TW',
    'English': 'en',
    'Français': 'fr',
    '日本語': 'ja',
    'Русский': 'ru',
    'Tiếng Việt': 'vi',
  }

  const switchLanguage = useCallback((langType: string) => {
    const code = typeToCode[langType] || langType
    i18n.changeLanguage(code)
    localStorage.setItem('i18nextLng', code)
  }, [i18n])

  const cycleTheme = useCallback(() => {

    const next = theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light'

    setTheme(next)

  }, [theme, setTheme])



  const ThemeIcon = resolvedTheme === 'dark' ? Moon : Sun



  return (

    <header className="flex h-14 items-center gap-2 border-b bg-background px-4">

      {/* Mobile Menu */}

      <Button

        variant="ghost"

        size="icon"

        className="md:hidden"

        onClick={onMobileMenuToggle}

      >

        <Menu className="h-5 w-5" />

      </Button>



      {/* Desktop: Page title in header */}

      <div className="hidden md:flex items-center gap-2">

        <span className="text-sm font-semibold">{breadcrumbMap[location.pathname] || ''}</span>

      </div>



      {/* Product Top Nav */}
      <div className="hidden md:flex items-center gap-0.5 mx-auto">
        {PRODUCT_ITEMS.map((item) => {
          const active = location.pathname === item.path || location.pathname.startsWith(item.path + '/')
          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                'px-3 py-1.5 text-xs font-medium rounded-md transition-colors',
                active
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'
              )}
            >
              {t(item.labelKey)}
            </Link>
          )
        })}
      </div>

      {/* Right Actions */}
      <div className="ml-auto flex items-center gap-0.5 sm:gap-1">

        {/* Search */}

        <Tooltip delayDuration={0}>

          <TooltipTrigger asChild>

            <Button variant="ghost" size="icon" className="hidden sm:inline-flex">

              <Search className="h-4 w-4" />

            </Button>

          </TooltipTrigger>

          <TooltipContent>{t('Search')}</TooltipContent>

        </Tooltip>



        {/* Notifications */}

        <Tooltip delayDuration={0}>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon">
              <Bell className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>{t('Notifications')}</TooltipContent>
        </Tooltip>

        {/* Language Selector (DB-driven) */}

        <DropdownMenu>

          <DropdownMenuTrigger asChild>

            <Button variant="ghost" size="icon">

              <Globe className="h-4 w-4" />

            </Button>

          </DropdownMenuTrigger>

          <DropdownMenuContent align="end" className="min-w-[140px]">

            <DropdownMenuLabel>{t('Language')}</DropdownMenuLabel>

            <DropdownMenuSeparator />

            {dbLanguages.length > 0 ? dbLanguages.map(lang => (
              <DropdownMenuItem

                key={lang}

                onClick={() => switchLanguage(lang)}

                className={i18n.language === (typeToCode[lang] || lang) ? 'bg-muted font-medium' : ''}

              >

                {lang}

              </DropdownMenuItem>

            )) : (
              <>
                <DropdownMenuItem onClick={() => switchLanguage('中文简体')}>
                  中文简体
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => switchLanguage('English')}>
                  English
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>

        </DropdownMenu>



        {/* Theme Toggle */}

        <Tooltip delayDuration={0}>

          <TooltipTrigger asChild>

            <Button variant="ghost" size="icon" onClick={cycleTheme}>

              <ThemeIcon className="h-4 w-4" />

            </Button>

          </TooltipTrigger>

          <TooltipContent>

            {theme === 'light'

              ? 'Light'

              : theme === 'dark'

                ? 'Dark'

                : 'System'}

          </TooltipContent>

        </Tooltip>



        {/* User Menu */}

        <DropdownMenu>

          <DropdownMenuTrigger asChild>

            <Button variant="ghost" className="flex items-center gap-2 px-2">

              <Avatar className="h-7 w-7">

                <AvatarFallback className="bg-primary/10 text-xs">

                  {auth.user?.display_name?.[0] || auth.user?.username?.[0] || 'U'}

                </AvatarFallback>

              </Avatar>

              <span className="hidden text-sm sm:inline-block">

                {auth.user?.display_name || auth.user?.username || ''}

              </span>

              {auth.user?.role === 100 && (

                <Badge variant="secondary" className="ml-1 text-[10px]">

                  Admin

                </Badge>

              )}

            </Button>

          </DropdownMenuTrigger>

          <DropdownMenuContent align="end" className="w-48">

            <DropdownMenuLabel>

              {auth.user?.display_name || auth.user?.username}

              <p className="text-xs font-normal text-muted-foreground">

                {auth.user?.email || `ID: ${auth.user?.id}`}

              </p>

            </DropdownMenuLabel>

            <DropdownMenuSeparator />

            <DropdownMenuItem onClick={() => router.navigate({ to: '/profile' })}>

              <User className="mr-2 h-4 w-4" />

              {t('Profile')}

            </DropdownMenuItem>

            <DropdownMenuSeparator />

            <DropdownMenuItem onClick={handleSignOut}>

              <LogOut className="mr-2 h-4 w-4" />

              {t('Sign Out')}

            </DropdownMenuItem>

          </DropdownMenuContent>

        </DropdownMenu>

      </div>

    </header>

  )

}



// ---------------------------------------------------------------------------

// Main Layout

// ---------------------------------------------------------------------------



function AppLayout() {

  const { t } = useTranslation()

  const [collapsed, setCollapsed] = useState(true)

  const [mobileOpen, setMobileOpen] = useState(false)



  return (

    <div className="flex h-screen w-full overflow-hidden bg-background">

      {/* Desktop Sidebar */}

      <aside

        className={cn(

          'hidden shrink-0 transition-all duration-200 md:block',

          collapsed ? 'w-0 border-0' : 'w-56 lg:w-60 xl:w-64 2xl:w-72'

        )}

      >

        <SidebarNav collapsed={collapsed} onToggle={() => setCollapsed(!collapsed)} />

      </aside>



      {/* Mobile Sidebar Overlay */}

      {mobileOpen && (

        <div

          className="fixed inset-0 z-40 bg-black/50 md:hidden"

          onClick={() => setMobileOpen(false)}

        />

      )}

      <aside

        className={cn(

          'fixed inset-y-0 left-0 z-50 w-60 transition-transform duration-200 md:hidden',

          mobileOpen ? 'translate-x-0' : '-translate-x-full'

        )}

      >

        <SidebarNav collapsed={false} onToggle={() => setMobileOpen(false)} />

      </aside>



      {/* Main Content Area */}

      <div className="flex flex-1 flex-col overflow-hidden">

        <AppHeader onMobileMenuToggle={() => setMobileOpen(true)} />



        <main className="flex-1 overflow-y-auto p-3 sm:p-4 md:p-6">

          <Breadcrumbs pathname={location.pathname} />

          <Suspense

            fallback={

              <div className="space-y-4">

                <Skeleton className="h-8 w-48" />

                <Skeleton className="h-64 w-full" />

              </div>

            }

          >

            <Outlet />

          </Suspense>

        </main>



        {/* Footer */}

        <footer className="border-t px-4 py-2 text-center text-xs text-muted-foreground">
          <div className="flex items-center justify-center gap-3">
            <p>{t('QuantumClaw')} &copy; {new Date().getFullYear()} {t('AI API Gateway')}</p>
            <span className="text-muted-foreground/40">|</span>
            <button
              onClick={() => { navigator.clipboard.writeText('587600277'); alert('QQ群号已复制: 587600277'); }}
              className="text-blue-500 hover:text-blue-400 transition-colors"
            >
              💬 QQ 群: 587600277
            </button>
          </div>
        </footer>

      </div>

    </div>

  )

}



			<CustomerServiceFloating />

export default AppLayout

