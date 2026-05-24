import { createRootRoute, Outlet, useLocation } from '@tanstack/react-router'
import { ThemeProvider } from '@/context/theme-provider'
import NavBar from '@/components/nav-bar'

export const Route = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  const location = useLocation()
  // Show NavBar on public pages only (auth pages have their own app-layout with sidebar)
  const hideNav = location.pathname.match(
    /^\/(dashboard|keys|billing|settings|profile|chat|admin|channels|logs|monitoring|news|distributors|about|profit|redemption|reseller|subscription|settlement|tasks|transactions|users|wallet|checkin|api-docs|platform-settings|not-found)/
  )

  return (
    <ThemeProvider>
      {!hideNav && <NavBar />}
      <Outlet />
    </ThemeProvider>
  )
}
