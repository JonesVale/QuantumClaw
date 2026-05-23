import { createFileRoute } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { getSelf } from '@/lib/api'
import AppLayout from '@/components/layout/app-layout'

let sessionVerified = false

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async () => {
    const { auth } = useAuthStore.getState()
    // 已登录用户验证 session 有效性
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
    // 未登录用户可以正常浏览页面，不跳转
  },
  component: () => <AppLayout />,
})
