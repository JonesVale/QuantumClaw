import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { getSelf } from '@/lib/api'
import AppLayout from '@/components/layout/app-layout'

let sessionVerified = false

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ location }) => {
    const { auth } = useAuthStore.getState()
    if (!auth.user) {
      throw redirect({
        to: '/sign-in',
        search: { redirect: location.href },
      })
    }
    if (!sessionVerified) {
      try {
        const res = await getSelf()
        if (res?.success && res.data) {
          auth.setUser(res.data)
          sessionVerified = true
        } else {
          auth.reset()
          throw redirect({
            to: '/sign-in',
            search: { redirect: location.href },
          })
        }
      } catch {
        auth.reset()
        throw redirect({
          to: '/sign-in',
          search: { redirect: location.href },
        })
      }
    }
  },
  component: () => <AppLayout />,
})
