import { useTranslation } from 'react-i18next'
import { useNavigate } from '@tanstack/react-router'
import { X, MessageSquare, Cpu, DollarSign, Box, Layers, Play, ExternalLink } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { ProviderIcon } from '@/components/provider-icon'

// Re-export the CatalogItem type used in models.tsx
export interface CatalogItem {
  name: string
  display_name: string
  description: string
  use_case: string
  context_window: number
  input_modalities: string[]
  series: string
  provider: string
  channel_id: number
  channel_name: string
  input_price: number
  output_price: number
  status: number
  group: string
}

export interface ModelDetailDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  model: CatalogItem | null
}

const useCaseLabels: Record<string, { label: string; color: string }> = {
  'chat': { label: 'Chat & Assistant', color: 'from-blue-500 to-blue-600' },
  'coding': { label: 'Code Generation', color: 'from-green-500 to-green-600' },
  'reasoning': { label: 'Reasoning', color: 'from-purple-500 to-purple-600' },
  'vision': { label: 'Vision', color: 'from-cyan-500 to-cyan-600' },
}

export function ModelDetailDialog({ open, onOpenChange, model }: ModelDetailDialogProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()

  if (!model) return null

  const handleStartChat = () => {
    navigate({ to: '/chat', search: { model: model.name } })
    onOpenChange(false)
  }

  const useCaseInfo = useCaseLabels[model.use_case]

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <div className="flex items-start gap-3">
            <ProviderIcon name={model.provider} size="md" />
            <div className="flex-1 min-w-0">
              <DialogTitle className="text-xl font-bold truncate">
                {model.name}
              </DialogTitle>
              <p className="text-sm text-muted-foreground mt-0.5">
                {model.provider}{model.series ? ` · ${model.series}` : ''}
              </p>
            </div>
          </div>
        </DialogHeader>

        {/* Full Description */}
        <div>
          <p className="text-sm text-muted-foreground leading-relaxed">
            {model.description}
          </p>
        </div>

        {/* Info Cards Grid */}
        <div className="grid grid-cols-2 gap-3">
          {/* Context Window */}
          <div className="rounded-xl border bg-muted/30 p-3.5 space-y-1">
            <div className="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
              <Cpu className="h-3.5 w-3.5" />
              {t('Context')}
            </div>
            <div className="text-lg font-bold">
              {(model.context_window / 1000).toFixed(0)}K
            </div>
            <div className="text-[10px] text-muted-foreground">
              {t('tokens')}
            </div>
          </div>

          {/* Pricing */}
          <div className="rounded-xl border bg-muted/30 p-3.5 space-y-1">
            <div className="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
              <DollarSign className="h-3.5 w-3.5" />
              {t('Pricing')}
            </div>
            <div className="space-y-0.5">
              <div className="flex items-center justify-between text-xs">
                <span className="text-emerald-600 dark:text-emerald-400 font-medium">{t('Input')}</span>
                <span className="font-mono font-semibold">
                  ${model.input_price > 0 ? model.input_price.toFixed(8) : 'Free'}
                </span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-amber-600 dark:text-amber-400 font-medium">{t('Output')}</span>
                <span className="font-mono font-semibold">
                  ${model.output_price > 0 ? model.output_price.toFixed(8) : 'Free'}
                </span>
              </div>
            </div>
          </div>

          {/* Series */}
          <div className="rounded-xl border bg-muted/30 p-3.5 space-y-1">
            <div className="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
              <Box className="h-3.5 w-3.5" />
              {t('Series')}
            </div>
            <div className="text-sm font-semibold">
              {model.series || '—'}
            </div>
            <div className="text-[10px] text-muted-foreground">
              {model.provider}
            </div>
          </div>

          {/* Modalities */}
          <div className="rounded-xl border bg-muted/30 p-3.5 space-y-1">
            <div className="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground">
              <Layers className="h-3.5 w-3.5" />
              {t('Modalities')}
            </div>
            <div className="flex flex-wrap gap-1">
              {model.input_modalities.length > 0 ? (
                model.input_modalities.map(mod => (
                  <span key={mod} className="px-1.5 py-0.5 rounded text-[10px] bg-muted/60 text-muted-foreground font-medium">
                    {mod}
                  </span>
                ))
              ) : (
                <span className="text-xs text-muted-foreground">—</span>
              )}
            </div>
          </div>
        </div>

        {/* Use Case Badge */}
        {useCaseInfo && (
          <div className="flex flex-wrap items-center gap-2">
            <span className={cn(
              'inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-medium text-white bg-gradient-to-r',
              useCaseInfo.color
            )}>
              {t(useCaseInfo.label)}
            </span>
            {model.input_modalities.map(mod => (
              <Badge key={mod} variant="secondary" className="text-[10px]">{mod}</Badge>
            ))}
          </div>
        )}

        {/* Action Button */}
        <Button
          size="default"
          className="w-full gap-2 mt-2"
          onClick={handleStartChat}
        >
          <MessageSquare className="h-4 w-4" />
          {t('Start Chat')}
        </Button>
      </DialogContent>
    </Dialog>
  )
}
