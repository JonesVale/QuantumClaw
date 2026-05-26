/**
 * QuantumClaw - Menu Permission Hook
 *
 * Fetches menu items from the backend, filtered by the current user's role.
 * Falls back to hardcoded defaults if the API is unavailable.
 */

import { useAuthStore } from '@/stores/auth-store'
import apiClient from '@/lib/api'
import { useQuery } from '@tanstack/react-query'
import { useT } from '@/lib/use-t'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface MenuItemData {
  id: number
  menu_key: string
  parent_key: string
  menu_type: 'nav' | 'sidebar'
  label_key: string
  icon: string
  path: string
  sort_order: number
  roles: string
  group_name: string
  enabled: boolean
}

export interface NavMenuItem {
  to: string
  label: string
  icon: string
}

export interface SidebarMenuItem {
  path: string
  icon: string
  labelKey: string
  groupName: string
  parentKey?: string
}

// ---------------------------------------------------------------------------
// Hardcoded Fallbacks (matching current default config)
// ---------------------------------------------------------------------------

const FALLBACK_NAV: NavMenuItem[] = [
  { to: '/dashboard',  label: 'Dashboard',  icon: '??' },
  { to: '/models',     label: 'Models',     icon: '?' },
  { to: '/pricing',    label: 'Pricing',    icon: 'ก่' },
  { to: '/rankings',   label: 'Rankings',   icon: 'กิ' },
  { to: '/apps',       label: 'Apps',       icon: '?' },
  { to: '/enterprise', label: 'Enterprise', icon: '?' },
  { to: '/news',       label: 'AI News',    icon: '??' },
  { to: '/api-docs',   label: 'API Docs',   icon: '??' },
]


const FALLBACK_SIDEBAR: SidebarMenuItem[] = [
  // Dashboard/management group
  { path: '/dashboard', icon: 'LayoutDashboard', labelKey: 'Dashboard', groupName: '' },
  { path: '/chat', icon: 'MessageSquare', labelKey: 'AI Chat', groupName: '' },
  { path: '/models', icon: 'Box', labelKey: 'Models', groupName: '' },
  { path: '/rankings', icon: 'TrendingUp', labelKey: 'Rankings', groupName: '' },
  { path: '/pricing', icon: 'DollarSign', labelKey: 'Pricing', groupName: '' },
  { path: '/quantum', icon: 'Atom', labelKey: 'Quantum', groupName: '' },
  { path: '/fusion', icon: 'GitCompare', labelKey: 'Fusion', groupName: '' },

  { path: '/apps', icon: 'Sparkles', labelKey: 'Apps', groupName: '' },
  { path: '/enterprise', icon: 'Building2', labelKey: 'Enterprise', groupName: '' },
  // Management group
  { path: '/keys', icon: 'Key', labelKey: 'API Keys', groupName: 'management' },
  { path: '/users', icon: 'Users', labelKey: 'Users', groupName: 'management' },
  { path: '/logs', icon: 'ScrollText', labelKey: 'Usage Logs', groupName: 'management' },
  { path: '/redemption', icon: 'Ticket', labelKey: 'Redemption Codes', groupName: 'management' },
  { path: '/distributors', icon: 'Truck', labelKey: 'Distributors', groupName: 'management' },
  { path: '/admin-tools', icon: 'Wrench', labelKey: 'Admin Tools', groupName: 'management' },
  { path: '/monitoring', icon: 'Activity', labelKey: 'Monitoring', groupName: 'management' },
  { path: '/profit', icon: 'TrendingUp', labelKey: 'Channel Profit', groupName: 'management' },

  { path: '/reseller-admin', icon: 'Store', labelKey: 'Reseller Management', groupName: 'management' },
  { path: '/settlement', icon: 'Percent', labelKey: 'Settlement Config', groupName: 'management' },
  { path: '/transactions', icon: 'Receipt', labelKey: 'Transactions', groupName: 'management' },
  { path: '/promo-ads', icon: 'Megaphone', labelKey: 'Promo Ads', groupName: 'management' },
  { path: '/platform-settings', icon: 'Settings', labelKey: 'Platform Settings', groupName: 'management' },
  { path: '/reseller', icon: 'Store', labelKey: 'Reseller Portal', groupName: 'management' },
  { path: '/reseller-keys', icon: 'Key', labelKey: 'My Keys', groupName: 'management' },
  { path: '/team', icon: 'Users', labelKey: 'My Team', groupName: 'management' },
  { path: '/channels', icon: 'Network', labelKey: 'Channels', groupName: 'management' },
  { path: '/menu-permissions', icon: 'Settings', labelKey: 'Menu Permissions', groupName: 'management' },
  // Account group
  { path: '/profile', icon: 'User', labelKey: 'Profile', groupName: 'account' },
  { path: '/wallet', icon: 'Wallet', labelKey: 'Wallet', groupName: 'account' },
  { path: '/billing', icon: 'DollarSign', labelKey: 'Billing', groupName: 'account' },
  { path: '/checkin', icon: 'Gift', labelKey: 'Daily Check-in', groupName: 'account' },
  { path: '/subscription', icon: 'CreditCard', labelKey: 'Subscriptions', groupName: 'account' },
  { path: '/tasks', icon: 'ClipboardList', labelKey: 'Task Logs', groupName: 'account' },
  { path: '/settings', icon: 'Settings', labelKey: 'Settings', groupName: 'account' },
  { path: '/api-docs', icon: 'BookOpen', labelKey: 'API Docs', groupName: 'account' },
  { path: '/about', icon: 'Info', labelKey: 'About', groupName: 'account' },
  { path: '/connections', icon: 'Link2', labelKey: 'OAuth Connections', groupName: 'account' },
  { path: '/notifications', icon: 'Bell', labelKey: 'Notifications', groupName: 'account' },
  { path: '/password', icon: 'Lock', labelKey: 'Password & Security', groupName: 'account' },
]

// Icon name to Lucide component mapping for fallback icons
const FALLBACK_ICON_MAP: Record<string, string> = {
  'LayoutDashboard': 'LayoutDashboard',
  'MessageSquare': 'MessageSquare',
  'Box': 'Box',
  'TrendingUp': 'TrendingUp',
  'DollarSign': 'DollarSign',
  'Atom': 'Atom',
  'GitCompare': 'GitCompare',
  'Sparkles': 'Sparkles',
  'Building2': 'Building2',
  'Key': 'Key',
  'Users': 'Users',
  'ScrollText': 'ScrollText',
  'Ticket': 'Ticket',
  'Truck': 'Truck',
  'Wrench': 'Wrench',
  'Activity': 'Activity',
  'Newspaper': 'Newspaper',
  'Store': 'Store',
  'Percent': 'Percent',
  'Receipt': 'Receipt',
  'Settings': 'Settings',
  'User': 'User',
  'Wallet': 'Wallet',
  'Gift': 'Gift',
  'CreditCard': 'CreditCard',
  'ClipboardList': 'ClipboardList',
  'BookOpen': 'BookOpen',
  'Info': 'Info',
  'Network': 'Network',
  'Megaphone': 'Megaphone',
  'Lock': 'Lock',
  'Bell': 'Bell',
  'Link2': 'Link2',
}

// ---------------------------------------------------------------------------
// API helpers
// ---------------------------------------------------------------------------

async function fetchMenus(type: string): Promise<MenuItemData[]> {
  const res = await apiClient.get(`/api/menus?type=${type}`, {
    skipErrorHandler: true,
  } as never)
  if (res.data?.success && Array.isArray(res.data.data)) {
    return res.data.data
  }
  return []
}

// ---------------------------------------------------------------------------
// React Query hook for nav menus
// ---------------------------------------------------------------------------

export function useNavMenus() {
  const { auth } = useAuthStore()

  return useQuery({
    queryKey: ['menus', 'nav', auth.user?.role ?? 0],
    queryFn: () => fetchMenus('nav'),
    staleTime: 5 * 60 * 1000,
    retry: 1,
    placeholderData: [],
    select: (data): NavMenuItem[] => {
      if (data && data.length > 0) {
        return data.map(m => ({
          to: m.path,
          label: m.label_key,
          icon: m.icon || '?',
        }))
      }
      // Fallback to hardcoded defaults
      return FALLBACK_NAV
    },
  })
}

// ---------------------------------------------------------------------------
// React Query hook for sidebar menus
// ---------------------------------------------------------------------------

export function useSidebarMenus() {
  const { auth } = useAuthStore()

  return useQuery({
    queryKey: ['menus', 'sidebar', auth.user?.role ?? 0],
    queryFn: () => fetchMenus('sidebar'),
    staleTime: 5 * 60 * 1000,
    retry: 1,
    placeholderData: [],
    select: (data): SidebarMenuItem[] => {
      if (data && data.length > 0) {
        return data.map(m => ({
          path: m.path,
          icon: m.icon || 'Box',
          labelKey: m.label_key,
          groupName: m.group_name,
          parentKey: m.parent_key || undefined,
        }))
      }
      // Fallback to hardcoded defaults (with role-based filtering)
      return getFallbackSidebar(auth.user?.role ?? 0)
    },
  })
}

// ---------------------------------------------------------------------------
// Fallback sidebar with role filtering
// ---------------------------------------------------------------------------

function getFallbackSidebar(role: number): SidebarMenuItem[] {
  const isAdmin = role >= 10
  const isLoggedIn = role >= 1

  return FALLBACK_SIDEBAR.filter(item => {
    // Apply same role-based filtering as the backend
    if (item.groupName === 'management') {
      // Admin-only items
      const adminOnlyPaths = ['/users', '/redemption', '/distributors', '/admin-tools',
        '/profit', '/reseller-admin', '/settlement', '/transactions', '/platform-settings',
        '/promo-ads', '/channels', '/menu-permissions', '/monitoring']
      const loginRequiredPaths = ['/keys', '/logs', '/reseller', '/reseller-keys', '/team']
      const publicPaths = ['/news']

      if (adminOnlyPaths.includes(item.path)) return isAdmin
      if (loginRequiredPaths.includes(item.path)) return isLoggedIn
      if (publicPaths.includes(item.path)) return true
      return true // default management items visible to all
    }
    if (item.groupName === 'account') {
      const adminOnlyPaths = ['/tasks', '/settings']
      const loginRequiredPaths = ['/profile', '/wallet', '/billing', '/checkin', '/subscription', '/api-docs', '/connections', '/notifications', '/password']
      const publicPaths = ['/about']

      if (adminOnlyPaths.includes(item.path)) return isAdmin
      if (loginRequiredPaths.includes(item.path)) return isLoggedIn
      if (publicPaths.includes(item.path)) return true
      return true
    }
    return true
  })
}

// ---------------------------------------------------------------------------
// Group sidebar items by group_name
// ---------------------------------------------------------------------------

export function groupSidebarMenus(items: SidebarMenuItem[]): Record<string, SidebarMenuItem[]> {
  const groups: Record<string, SidebarMenuItem[]> = {}
  for (const item of items) {
    const group = item.groupName || ''
    if (!groups[group]) {
      groups[group] = []
    }
    groups[group].push(item)
  }
  return groups
}

// ---------------------------------------------------------------------------
// Admin API
// ---------------------------------------------------------------------------

export async function adminGetAllMenus(): Promise<MenuItemData[]> {
  const res = await apiClient.get('/api/admin/menus')
  if (res.data?.success && Array.isArray(res.data.data)) {
    return res.data.data
  }
  return []
}

export async function adminSaveMenu(menu: Partial<MenuItemData> & { menu_key: string; label_key: string; path: string }): Promise<MenuItemData> {
  const res = await apiClient.post('/api/admin/menus', menu)
  if (res.data?.success && res.data.data) {
    return res.data.data
  }
  throw new Error(res.data?.message || 'Failed to save menu')
}

export async function adminDeleteMenu(id: number): Promise<void> {
  const res = await apiClient.delete(`/api/admin/menus/${id}`)
  if (!res.data?.success) {
    throw new Error(res.data?.message || 'Failed to delete menu')
  }
}
