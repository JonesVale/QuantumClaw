import { useT } from '@/lib/use-t'
import { useNavigate } from '@tanstack/react-router'
import { Cpu, DollarSign, Box, Layers, Play } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { ProviderIcon } from '@/components/provider-icon'
import type { CatalogItem } from '@/components/model-detail-dialog'

export interface ModelComparisonDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  models: CatalogItem[]
}

const useCaseLabels: Record<string, { label: string; color: string }> = {
  'chat': { label: 'Chat & Assistant', color: 'from-blue-500 to-blue-600' },
  'coding': { label: 'Code Generation', color: 'from-green-500 to-green-600' },
  'reasoning': { label: 'Reasoning', color: 'from-purple-500 to-purple-600' },
  'vision': { label: 'Vision', color: 'from-cyan-500 to-cyan-600' },
}

function ComparisonCell({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1.5">
      <div className="text-[10px] uppercase tracking-wider text-muted-foreground font-semibold">
        {label}
      </div>
      <div className="text-sm">
        {children}
      </div>
    </div>
  )
}

export function ModelComparisonDialog({ open, onOpenChange, models }: ModelComparisonDialogProps) {
  const { t } = useT()
  const navigate = useNavigate()

  if (!models || models.length === 0) return null

  const handleChatWith = (modelName: string) => {
    navigate({ to: '/chat', search: { model: modelName } })
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-4xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold">
            {t('Compare Models')}
          </DialogTitle>
        </DialogHeader>

        {/* Comparison Grid */}
        <div
          className="grid gap-4"
          style={{
            gridTemplateColumns: `140px repeat(${models.length}, 1fr)`,
          }}
        >
          {/* Header row */}
          <div className="sticky top-0 bg-background z-10" />

          {models.map((m) => (
            <div key={m.name} className="text-center space-y-2">
              <div className="flex justify-center">
                <ProviderIcon name={m.provider} size="lg" />
              </div>
              <h3 className="font-semibold text-sm leading-tight">{m.name}</h3>
              <p className="text-[11px] text-muted-foreground">{m.provider}</p>
            </div>
          ))}

          {/* Separator */}
          <div className="col-span-full border-t border-border/40" />

          {/* Use Case / Category row */}
          <ComparisonCell label={t('Use Case')}>
            <span className="text-xs text-muted-foreground">{t('Use Case')}</span>
          </ComparisonCell>
          {models.map((m) => (
            <div key={`uc-${m.name}`}>
              {(() => { const uc = useCaseLabels[m.use_case]; if (!uc) return <span className="text-xs text-muted-foreground">—</span>; return (
                <span className={cn('inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-medium text-white bg-gradient-to-r', uc.color)}>{t(uc.label)}</span>
              ) })()}
            </div>
          ))}

          {/* Description row */}
          <ComparisonCell label={t('Description')}>
            <span className="text-xs text-muted-foreground">{t('Description')}</span>
          </ComparisonCell>
          {models.map((m) => (
            <div key={`desc-${m.name}`}>
              <p className="text-xs text-muted-foreground leading-relaxed line-clamp-4">
                {m.description}
              </p>
            </div>
          ))}

          {/* Context Window row */}
          <ComparisonCell label={t('Context')}>
            <Cpu className="h-3.5 w-3.5 text-muted-foreground" />
          </ComparisonCell>
          {models.map((m) => (
            <div key={`ctx-${m.name}`}>
              <span className="font-semibold">{(m.context_window / 1000).toFixed(0)}K</span>
              <span className="text-xs text-muted-foreground ml-1">{t('tokens')}</span>
            </div>
          ))}

          {/* Pricing row */}
          <ComparisonCell label={t('Pricing')}>
            <DollarSign className="h-3.5 w-3.5 text-muted-foreground" />
          </ComparisonCell>
          {models.map((m) => (
            <div key={`price-${m.name}`} className="space-y-1">
              <div className="flex items-center gap-1.5 text-xs">
                <span className="text-emerald-600 dark:text-emerald-400 font-medium">{t('In')}</span>
                <span className="font-mono font-semibold">
                  {m.input_price > 0 ? `$${m.input_price.toFixed(8)}` : t('Free')}
                </span>
              </div>
              <div className="flex items-center gap-1.5 text-xs">
                <span className="text-amber-600 dark:text-amber-400 font-medium">{t('Out')}</span>
                <span className="font-mono font-semibold">
                  {m.output_price > 0 ? `$${m.output_price.toFixed(8)}` : t('Free')}
                </span>
              </div>
            </div>
          ))}

          {/* Series row */}
          <ComparisonCell label={t('Series')}>
            <Box className="h-3.5 w-3.5 text-muted-foreground" />
          </ComparisonCell>
          {models.map((m) => (
            <div key={`series-${m.name}`}>
              <span className="text-sm">{m.series || '—'}</span>
            </div>
          ))}

          {/* Modalities row */}
          <ComparisonCell label={t('Modalities')}>
            <Layers className="h-3.5 w-3.5 text-muted-foreground" />
          </ComparisonCell>
          {models.map((m) => (
            <div key={`mod-${m.name}`} className="flex flex-wrap gap-1">
              {m.input_modalities.length > 0 ? (
                m.input_modalities.map(mod => (
                  <span key={mod} className="px-1.5 py-0.5 rounded text-[10px] bg-muted/60 text-muted-foreground">
                    {mod}
                  </span>
                ))
              ) : (
                <span className="text-xs text-muted-foreground">—</span>
              )}
            </div>
          ))}

          {/* Provider row */}
          <ComparisonCell label={t('Provider')}>
            <span className="text-xs text-muted-foreground">{t('Provider')}</span>
          </ComparisonCell>
          {models.map((m) => (
            <div key={`prov-${m.name}`}>
              <span className="text-sm">{m.provider}</span>
            </div>
          ))}

          {/* Action buttons */}
          <div />
          {models.map((m) => (
            <div key={`action-${m.name}`} className="pt-1">
              <Button
                size="sm"
                className="w-full h-8 text-xs gap-1"
                onClick={() => handleChatWith(m.name)}
              >
                <Play className="h-3 w-3" />
                {t('Chat')}
              </Button>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  )
}
