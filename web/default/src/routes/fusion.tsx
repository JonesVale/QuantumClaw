import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'

export const Route = createFileRoute('/fusion')({
  component: FusionPage,
})

function FusionPage() {
  const { t } = useT()
  return (
    <div className="min-h-screen bg-background flex items-center justify-center"
      style={{ backgroundImage: 'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)' }}>
      <div className="qc-wrap qc-section-pad-sm text-center">
        <div className="qc-fade-up max-w-xl mx-auto">
          <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center text-2xl mx-auto mb-6 shadow-sm">⚗</div>
          <h1 className="qc-title-hero font-bold tracking-tight text-foreground mb-4">{t('Model Fusion')}</h1>
          <p className="qc-text-body text-muted-foreground/70 leading-relaxed mb-8">
            {t('Combine multiple AI models into a single pipeline. Route prompts intelligently, compare outputs, and build custom workflows.')}
          </p>
          <div className="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg hover:-translate-y-0.5 transition-all cursor-pointer">
            {t('Coming Soon')}
          </div>
        </div>
      </div>
    </div>
  )
}
