import { createRootRoute, Outlet } from '@tanstack/react-router'
import { ThemeProvider } from '@/context/theme-provider'
import NavBar from '@/components/nav-bar'

export const Route = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  return (
    <ThemeProvider>
      <NavBar />
      <Outlet />
    </ThemeProvider>
  )
}
