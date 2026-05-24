import { createFileRoute, lazyRouteComponent } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { getSelf } from '@/lib/api'
import AppLayout from '@/components/layout/app-layout'

let sessionVerified = false

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async () => {
    const { auth } = useAuthStore.getState()
    // Verify logged-in user session validity
    if (auth.user && !sessionVerified) {
      try {
        const res = await getSelf()
        if (res?.success && res.data) {
          auth.setUser(res.data as import('@/stores/auth-store').AuthUser)
          sessionVerified = true
        } else {
          auth.reset()
        }
      } catch {
        auth.reset()
      }
    }
    // Non-logged-in users can browse normally without redirect
  },
  component: () => <AppLayout />,
})
