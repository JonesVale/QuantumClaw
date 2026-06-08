import { createFileRoute, Link } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useAuthStore } from '@/stores/auth-store'

export const Route = createFileRoute('/download')({
  component: DownloadPage,
})

const icons = {
  phone: <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>,
  android: <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M17.6 6.4L22 4l-2.4 5.6M3 4l4.4 2.4M6.4 17.6L4 22l5.6-2.4M17.6 6.4A6 6 0 0 0 6.4 17.6 6 6 0 0 0 17.6 6.4Z"/><circle cx="12" cy="12" r="2"/></svg>,
  ios: <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="4" y="2" width="16" height="20" rx="3"/><line x1="9" y1="6" x2="15" y2="6"/></svg>,
  qrcode: <svg width="120" height="120" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><path d="M14 14h3v3h-3zM17 17h3v3h-3z"/></svg>,
  check: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><polyline points="20 6 9 17 4 12"/></svg>,
  arrowR: <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>,
}

const features = [
  { icon: '🔑', title: 'API Key 管理', desc: '创建、管理、监控你的 API Key，随时随地' },
  { icon: '💰', title: '钱包 & 充值', desc: '查看余额、充值、交易记录，一键操作' },
  { icon: '📊', title: '用量统计', desc: '实时查看配额使用、调用趋势、模型消耗' },
  { icon: '📨', title: '邀请好友', desc: '一键邀请通讯录好友，自动绑定推广返佣' },
  { icon: '🏪', title: '店铺管理', desc: '渠道商管理店铺、上架模型、查看业绩' },
  { icon: '🏢', title: '企业控制台', desc: '部门管理、预算控制、审批流程、员工 Key 管控' },
]

const faqs = [
  { q: 'App 支持哪些平台？', a: '支持 iOS 和 Android。iOS 需 iOS 15.0+，Android 需 Android 8.0+。' },
  { q: 'App 收费吗？', a: 'App 完全免费。使用平台 API 时按实际用量计费，与 Web 端一致。' },
  { q: '如何获取帮助？', a: '通过 App 内的「设置-帮助反馈」提交工单，或发送邮件至 support@quantumclaw.com。' },
]

function DownloadPage() {
  const { t } = useT()
  const { auth } = useAuthStore()
  const loggedIn = !!auth.user

  return (
    <div className="min-h-screen bg-white">
      {/* ═══ HERO ═══ */}
      <section className="qc-section-pad-lg bg-gradient-to-br from-amber-50 via-orange-50 to-rose-50">
        <div className="qc-wrap">
          <div className="flex flex-col lg:flex-row items-center gap-12 lg:gap-20">
            {/* Left: Text */}
            <div className="flex-1 text-center lg:text-left">
              <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full border border-orange-200 bg-white/60 text-orange-700 text-xs font-semibold mb-6">
                📱 Mobile App
              </div>
              <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight text-foreground leading-tight">
                {t('Quantum Spirit Claw')}{' '}
                <span className="bg-gradient-to-r from-amber-500 to-orange-500 bg-clip-text text-transparent">{t('Mobile')}</span>
              </h1>
              <p className="text-lg sm:text-xl text-muted-foreground mt-6 max-w-xl leading-relaxed">
                {t('Manage your API keys, monitor usage, invite friends, and run your business — all from your phone.')}
              </p>

              {/* Download Buttons */}
              <div className="flex flex-col sm:flex-row items-center gap-4 mt-10">
                <a
                  href="#"
                  className="inline-flex items-center gap-3 rounded-2xl bg-foreground text-white px-8 py-4 text-base font-semibold hover:bg-foreground/90 transition-all shadow-xl shadow-foreground/10 hover:shadow-foreground/20"
                  onClick={(e) => { e.preventDefault(); alert('iOS 版本即将上线 App Store') }}
                >
                  {icons.ios}
                  <div className="text-left">
                    <div className="text-[10px] text-white/60 font-normal">{t("Download on the")}</div>
                    <div className="text-base font-semibold -mt-0.5">{t("App Store")}</div>
                  </div>
                </a>
                <a
                  href="#"
                  className="inline-flex items-center gap-3 rounded-2xl border-2 border-foreground/20 text-foreground px-8 py-4 text-base font-semibold hover:bg-foreground/5 transition-all"
                  onClick={(e) => { e.preventDefault(); alert('Android APK 正在打包中') }}
                >
                  {icons.android}
                  <div className="text-left">
                    <div className="text-[10px] text-muted-foreground font-normal">{t("Get it on")}</div>
                    <div className="text-base font-semibold -mt-0.5">{t("Google Play")}</div>
                  </div>
                </a>
              </div>

              {/* QR Code */}
              <div className="mt-10 flex items-center gap-4 justify-center lg:justify-start">
                <div className="w-24 h-24 rounded-2xl border-2 border-dashed border-orange-300 bg-white flex items-center justify-center text-orange-300">
                  {icons.qrcode}
                </div>
                <div className="text-left">
                  <div className="text-sm font-semibold text-foreground">{t('Scan to Download')}</div>
                  <div className="text-xs text-muted-foreground mt-1">{t('Use your phone camera to scan')}</div>
                </div>
              </div>
            </div>

            {/* Right: Phone Mockup */}
            <div className="flex-shrink-0">
              <div className="relative w-64 h-[500px] rounded-[2.5rem] border-8 border-foreground/10 bg-white shadow-2xl shadow-amber-500/5 overflow-hidden">
                <div className="absolute top-0 left-1/2 -translate-x-1/2 w-28 h-6 bg-foreground/10 rounded-b-2xl" />
                <div className="h-full flex flex-col items-center justify-center p-6 text-center">
                  <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center mb-4 shadow-lg shadow-orange-500/20">
                    <span className="text-2xl font-bold text-white">QC</span>
                  </div>
                  <h3 className="text-lg font-bold text-foreground">QuantumClaw</h3>
                  <p className="text-xs text-muted-foreground mt-2">AI API 聚合平台</p>
                  <div className="mt-8 w-full space-y-3">
                    {['📨 邀请好友', '🔑 API Key', '💰 钱包', '📊 统计'].map(item => (
                      <div key={item} className="flex items-center gap-3 p-3 rounded-xl bg-amber-50/50 text-sm font-medium text-foreground">
                        <span>{item}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ FEATURES ═══ */}
      <section className="qc-section-pad-lg bg-white">
        <div className="qc-wrap">
          <div className="text-center mb-16">
            <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground">
              {t('Everything on')} <span className="bg-gradient-to-r from-amber-500 to-orange-500 bg-clip-text text-transparent">{t('Mobile')}</span>
            </h2>
            <p className="text-muted-foreground mt-4 max-w-xl mx-auto">{t('All the power of QuantumClaw fits in your pocket.')}</p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((f, i) => (
              <div key={i} className="rounded-2xl border border-border/50 bg-white p-6 hover:shadow-lg hover:border-orange-200/50 transition-all">
                <div className="text-3xl mb-4">{f.icon}</div>
                <h3 className="text-base font-semibold text-foreground mb-2">{f.title}</h3>
                <p className="text-sm text-muted-foreground leading-relaxed">{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ FAQ ═══ */}
      <section className="qc-section-pad-lg bg-amber-50/30 border-y border-border/30">
        <div className="qc-wrap max-w-2xl">
          <h2 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground text-center mb-12">
            {t('Frequently Asked')} <span className="bg-gradient-to-r from-amber-500 to-orange-500 bg-clip-text text-transparent">{t('Questions')}</span>
          </h2>
          <div className="space-y-4">
            {faqs.map((faq, i) => (
              <details key={i} className="group rounded-2xl border border-border/50 bg-white p-6 open:shadow-md transition-all">
                <summary className="flex items-center justify-between cursor-pointer text-foreground font-semibold text-sm">
                  {faq.q}
                  <svg className="w-4 h-4 text-muted-foreground group-open:rotate-180 transition-transform" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 9l6 6 6-6"/></svg>
                </summary>
                <p className="mt-4 text-sm text-muted-foreground leading-relaxed border-t border-border/20 pt-4">{faq.a}</p>
              </details>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ BOTTOM CTA ═══ */}
      <section className="qc-section-pad-sm bg-white">
        <div className="qc-wrap">
          <div className="text-center">
            <p className="text-sm text-muted-foreground">
              {t('Already have an account?')}{' '}
              <Link to={loggedIn ? '/dashboard' : '/sign-in'} className="text-orange-600 font-semibold hover:text-orange-700 transition-colors">
                {loggedIn ? t('Go to Dashboard') : t('Sign In')} {icons.arrowR}
              </Link>
            </p>
          </div>
        </div>
      </section>
    </div>
  )
}
