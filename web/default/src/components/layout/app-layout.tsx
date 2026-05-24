/**

 * QuantumClaw - Main Application Layout

 * 

 * Provides the sidebar + header + content area structure

 * for all authenticated pages.

 */



import { Suspense, useCallback, useMemo, useState, useEffect } from 'react'

import { Link, Outlet, useLocation, useRouter } from '@tanstack/react-router'

import { useT } from '@/lib/use-t'

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
  Building2,
  Percent,
  Receipt,
  Store,
  Link2,

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
import { ErrorBoundary } from '@/components/error-boundary'
import { useSidebarMenus, groupSidebarMenus, type SidebarMenuItem } from '@/lib/use-menus'



// ---------------------------------------------------------------------------
// Icon resolver: maps icon name strings from DB to Lucide components
// ---------------------------------------------------------------------------

const ICON_MAP: Record<string, React.ElementType> = {
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
  MessageSquare,
  Gift,
  CreditCard,
  ClipboardList,
  DollarSign,
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
  Building2,
  Percent,
  Receipt,
  Store,
  Zap,
  Home,
}

function resolveIcon(name: string): React.ElementType {
  return ICON_MAP[name] || Box
}

// ---------------------------------------------------------------------------
// Sidebar Navigation Items (hardcoded fallback)
// ---------------------------------------------------------------------------

interface NavItem {

  path: string

  icon: React.ElementType

  labelKey: string

  adminOnly?: boolean
  loginRequired?: boolean

}



// Product pages shown in top navigation bar
const PRODUCT_ITEMS: NavItem[] = [
  { path: '/dashboard', icon: LayoutDashboard, labelKey: 'Dashboard', loginRequired: true },
  { path: '/chat', icon: MessageSquare, labelKey: 'AI Chat' },
  { path: '/models', icon: Box, labelKey: 'Models' },
  { path: '/rankings', icon: TrendingUp, labelKey: 'Rankings' },
  { path: '/pricing', icon: DollarSign, labelKey: 'Pricing' },
  { path: '/quantum', icon: Atom, labelKey: 'Quantum' },
  { path: '/fusion', icon: GitCompare, labelKey: 'Fusion' },
  { path: '/apps', icon: Sparkles, labelKey: 'Apps' },
  { path: '/enterprise', icon: Building2, labelKey: 'Enterprise' },
]

// Admin/management pages (sidebar) - hardcoded fallback
const NAV_ITEMS: NavItem[] = [
  { path: '/keys', icon: Key, labelKey: 'API Keys', loginRequired: true },
  { path: '/users', icon: Users, labelKey: 'Users', adminOnly: true },
  { path: '/logs', icon: ScrollText, labelKey: 'Usage Logs', loginRequired: true },
  { path: '/redemption', icon: Ticket, labelKey: 'Redemption Codes', adminOnly: true },
  { path: '/distributors', icon: Truck, labelKey: 'Distributors', adminOnly: true },
  { path: '/admin-tools', icon: Wrench, labelKey: 'Admin Tools', adminOnly: true },
  { path: '/monitoring', icon: Activity, labelKey: 'Monitoring', loginRequired: true },
  { path: '/profit', icon: TrendingUp, labelKey: 'Channel Profit', adminOnly: true },
  { path: '/news', icon: Newspaper, labelKey: 'AI News' },
  { path: '/reseller-admin', icon: Store, labelKey: 'Reseller Management', adminOnly: true },
  { path: '/settlement', icon: Percent, labelKey: 'Settlement Config', adminOnly: true },
  { path: '/transactions', icon: Receipt, labelKey: 'Transactions', adminOnly: true },
  { path: '/platform-settings', icon: Settings, labelKey: 'Platform Settings', adminOnly: true },
  { path: '/reseller', icon: Store, labelKey: 'Reseller Portal', loginRequired: true },
  { path: '/reseller-keys', icon: Key, labelKey: 'My Keys', loginRequired: true },
]



const SETTINGS_ITEMS: NavItem[] = [

  { path: '/profile', icon: User, labelKey: 'Profile', loginRequired: true },

  { path: '/wallet', icon: Wallet, labelKey: 'Wallet', loginRequired: true },

  { path: '/billing', icon: DollarSign, labelKey: 'Billing', loginRequired: true },

  { path: '/checkin', icon: Gift, labelKey: 'Daily Check-in', loginRequired: true },

  { path: '/subscription', icon: CreditCard, labelKey: 'Subscriptions', loginRequired: true },

  { path: '/tasks', icon: ClipboardList, labelKey: 'Task Logs', adminOnly: true },

  { path: '/settings', icon: Settings, labelKey: 'Settings', adminOnly: true },

  { path: '/connections', icon: Link2, labelKey: 'OAuth Connections', loginRequired: true },

  { path: '/api-docs', icon: BookOpen, labelKey: 'API Docs', loginRequired: true },
  { path: '/notifications', icon: Bell, labelKey: 'Notifications', loginRequired: true },
  { path: '/about', icon: Info, labelKey: 'About' },

]



// Group definitions for fallback sidebar rendering
const FALLBACK_SIDEBAR_GROUPS: Record<string, NavItem[]> = {
  '': [
    { path: '/dashboard', icon: LayoutDashboard, labelKey: 'Dashboard', loginRequired: true },
    { path: '/chat', icon: MessageSquare, labelKey: 'AI Chat' },
    { path: '/models', icon: Box, labelKey: 'Models' },
    { path: '/rankings', icon: TrendingUp, labelKey: 'Rankings' },
    { path: '/pricing', icon: DollarSign, labelKey: 'Pricing' },
    { path: '/quantum', icon: Atom, labelKey: 'Quantum' },
    { path: '/fusion', icon: GitCompare, labelKey: 'Fusion' },
    { path: '/apps', icon: Sparkles, labelKey: 'Apps' },
    { path: '/enterprise', icon: Building2, labelKey: 'Enterprise' },
  ],
  'management': NAV_ITEMS,
  'account': SETTINGS_ITEMS,
}

const GROUP_LABEL_KEYS: Record<string, string> = {
  '': '',
  'management': 'Management',
  'account': 'Account',
}

// ---------------------------------------------------------------------------
// Sidebar Component
// ---------------------------------------------------------------------------



function SidebarNav({ collapsed, onToggle }: { collapsed: boolean; onToggle: () => void }) {

  const location = useLocation()

  const { t } = useT()

  const { auth } = useAuthStore()

  const isAdmin = auth.user?.role === 100
  const isLoggedIn = !!auth.user

  // Fetch sidebar menus from API
  const { data: apiSidebarItems = [] } = useSidebarMenus()

  // Resolve which sidebar items to use (API or fallback)
  const sidebarGroups = useMemo(() => {
    if (apiSidebarItems.length > 0) {
      // Use API data - group by group_name
      const grouped = groupSidebarMenus(apiSidebarItems)
      // Convert to NavItem format for rendering
      const result: Record<string, NavItem[]> = {}
      for (const [group, items] of Object.entries(grouped)) {
        result[group] = items.map(item => ({
          path: item.path,
          icon: resolveIcon(item.icon),
          labelKey: item.labelKey,
        }))
      }
      return result
    }
    // Fallback to hardcoded groups with role filtering
    const fallback: Record<string, NavItem[]> = {}
    for (const [group, items] of Object.entries(FALLBACK_SIDEBAR_GROUPS)) {
      fallback[group] = items.filter(item => {
        if (item.adminOnly && !isAdmin) return false
        if (item.loginRequired && !isLoggedIn) return false
        return true
      })
    }
    return fallback
  }, [apiSidebarItems, isAdmin, isLoggedIn])




  const isActive = useCallback(

    (path: string) => location.pathname === path,

    [location.pathname]

  )



  const renderNavItem = (item: NavItem) => {

    const Icon = item.icon
    const active = isActive(item.path)

    const linkContent = (

      <Link
        to={item.path}
        className={cn(
          'flex items-center gap-3 rounded-xl px-3 py-2 text-sm font-medium transition-colors',
          active
            ? 'bg-gradient-to-r from-amber-50 to-orange-50 text-amber-800 shadow-sm'
            : 'text-muted-foreground/70 hover:text-foreground hover:bg-muted/40'
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
    <div className="sticky top-24 bg-white/60 backdrop-blur-xl rounded-2xl border border-border/20 shadow-sm p-4 w-60">

      {/* Brand Header */}
      <div className="flex h-14 items-center px-3">
        <Link to="/" className="flex items-center gap-2">
          <img src="/logo.webp" alt="QuantumClaw" className="h-8 w-8 rounded-xl object-cover" />
          {!collapsed && (
            <span className="text-base font-bold tracking-tight">{t('QuantumClaw')}</span>
          )}
        </Link>
      </div>

      {/* Navigation */}
      <ScrollArea className="flex-1 px-3 py-2">
        <div className="space-y-1">
          {Object.entries(sidebarGroups).map(([group, items]) => {
            if (items.length === 0) return null
            return (
              <div key={group || '__default'}>
                {/* Group label for non-empty group names */}
                {group && !collapsed && (
                  <div className="px-3 py-1.5 mb-1">
                    <span className="text-[10px] font-semibold text-muted-foreground/40 uppercase tracking-[0.15em]">
                      {t(GROUP_LABEL_KEYS[group] || group)}
                    </span>
                  </div>
                )}
                {items.map(renderNavItem)}
                {/* Separator between groups (except last) */}
                <div className="my-2 border-t border-border/10" />
              </div>
            )
          })}
        </div>
      </ScrollArea>

      {/* Collapse Toggle */}
      <div className="border-t border-border/20 p-2">
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



// ── Breadcrumbs ──────────────
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

  '/chat': 'AI Chat',
  '/monitoring': 'Monitoring',
  '/news': 'AI News',

  '/profile': 'Profile',

  '/wallet': 'Wallet',

  '/billing': 'Billing',

  '/checkin': 'Daily Check-in',

  '/subscription': 'Subscriptions',

  '/tasks': 'Task Logs',

  '/settings': 'Settings',
  '/menu-permissions': 'Menu Permissions',
  '/connections': 'OAuth Connections',

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
  const { t, language, changeLanguage, langs } = useT()
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

  const switchLanguage = useCallback((langType: string) => {
    changeLanguage(langType)
  }, [changeLanguage])

  const cycleTheme = useCallback(() => {
    const next = theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light'
    setTheme(next)
  }, [theme, setTheme])

  const ThemeIcon = resolvedTheme === 'dark' ? Moon : Sun

  return (
    <header className="flex h-14 items-center gap-2 border-b border-border/20 bg-white/80 backdrop-blur-xl px-4">
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
                'px-3 py-1.5 text-xs font-medium rounded-xl transition-colors',
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
            {langs.map(lang => (
              <DropdownMenuItem
                key={lang}
                onClick={() => switchLanguage(lang)}
                className={language === lang ? 'bg-muted font-medium' : ''}
              >
                {lang}
              </DropdownMenuItem>
            ))}
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

function AppLayout() {
  const { t } = useT()
  const location = useLocation()
  const [collapsed, setCollapsed] = useState(true)
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <div className="flex min-h-screen w-full bg-background" style={{backgroundImage:'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)'}}>

      {/* Desktop Sidebar */}
      <div className="hidden md:block shrink-0 pt-4 pl-4">
        <SidebarNav collapsed={collapsed} onToggle={() => setCollapsed(!collapsed)} />
      </div>

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
      <div className="flex-1 min-w-0">

        <main className="p-3 sm:p-4 md:p-6">
          <Breadcrumbs pathname={location.pathname} />
          <Suspense
            fallback={
              <div className="space-y-4">
                <Skeleton className="h-10 w-56" />
                <Skeleton className="h-64 w-full" />
              </div>
            }
          >
            <ErrorBoundary>
              <Outlet />
            </ErrorBoundary>
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
              💰 QQ 群: 587600277
            </button>
          </div>
        </footer>

      </div>

    </div>
  )
}

export default AppLayout
