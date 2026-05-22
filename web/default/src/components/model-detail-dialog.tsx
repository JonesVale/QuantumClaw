import { useTranslation } from 'react-i18next'
import { ExternalLink } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ProviderIcon } from '@/components/provider-icon'
import type { EnhancedModel } from '@/lib/api-extended'

export interface ModelDetailDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  model: EnhancedModel | null
}

export function ModelDetailDialog({ open, onOpenChange, model }: ModelDetailDialogProps) {
  const { t } = useTranslation()

  if (!model) return null

  const isActive = model.status === 1

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          {/* Provider row */}
          <div className="flex items-center gap-3 mb-1">
            <ProviderIcon type={model.provider_type} name={model.provider} size="md" />
            <div>
              <DialogTitle className="text-lg">{model.name}</DialogTitle>
              <DialogDescription className="text-xs mt-0.5">
                {model.provider} · {model.channel_name}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        {/* Details */}
        <div className="space-y-4">
          {/* Pricing */}
          <div>
            <h4 className="text-sm font-semibold text-muted-foreground mb-2">{t('Pricing')}</h4>
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-lg border bg-muted/30 p-3">
                <div className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
                  {t('Input')}
                </div>
                <div className="text-lg font-bold text-emerald-600 dark:text-emerald-400">
                  ${model.input_price.toFixed(4)}
                </div>
                <div className="text-[10px] text-muted-foreground">{t('per 1K tokens')}</div>
              </div>
              <div className="rounded-lg border bg-muted/30 p-3">
                <div className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
                  {t('Output')}
                </div>
                <div className="text-lg font-bold text-amber-600 dark:text-amber-400">
                  ${model.output_price.toFixed(4)}
                </div>
                <div className="text-[10px] text-muted-foreground">{t('per 1K tokens')}</div>
              </div>
            </div>
          </div>

          {/* Info rows */}
          <div className="space-y-2">
            <div className="flex items-center justify-between py-1.5 border-b border-border/40">
              <span className="text-sm text-muted-foreground">{t('Channel')}</span>
              <span className="text-sm font-medium">{model.channel_name}</span>
            </div>
            <div className="flex items-center justify-between py-1.5 border-b border-border/40">
              <span className="text-sm text-muted-foreground">{t('Provider')}</span>
              <span className="text-sm font-medium">{model.provider}</span>
            </div>
            <div className="flex items-center justify-between py-1.5 border-b border-border/40">
              <span className="text-sm text-muted-foreground">{t('Group')}</span>
              <span className="text-sm font-medium">{model.group || '—'}</span>
            </div>
            <div className="flex items-center justify-between py-1.5 border-b border-border/40">
              <span className="text-sm text-muted-foreground">{t('Cost per Unit')}</span>
              <span className="text-sm font-medium font-mono">
                {model.cost_per_unit.toFixed(6)}
              </span>
            </div>
            <div className="flex items-center justify-between py-1.5 border-b border-border/40">
              <span className="text-sm text-muted-foreground">{t('Sell Price Rate')}</span>
              <span className="text-sm font-medium">{model.sell_price_rate}x</span>
            </div>
            <div className="flex items-center justify-between py-1.5">
              <span className="text-sm text-muted-foreground">{t('Status')}</span>
              <Badge variant={isActive ? 'default' : 'secondary'}>
                {isActive ? t('Active') : t('Inactive')}
              </Badge>
            </div>
          </div>

          {/* Test button */}
          <div className="pt-2">
            <Button
              variant="outline"
              size="sm"
              className="w-full gap-2"
              onClick={() => {
                // Navigate to Chat with model pre-selected
                window.location.href = `/chat?model=${encodeURIComponent(model.name)}`
              }}
            >
              <ExternalLink className="h-4 w-4" />
              {t('Chat with Model')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
