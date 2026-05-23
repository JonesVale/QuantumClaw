import { useT } from '@/lib/use-t'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { ProviderIcon } from '@/components/provider-icon'
import type { EnhancedModel } from '@/lib/api-extended'

export interface ModelCardProps {
  model: EnhancedModel
  onDetail: (model: EnhancedModel) => void
}

export function ModelCard({ model, onDetail }: ModelCardProps) {
  const { t } = useT()

  const isActive = model.status === 1

  return (
    <Card
      className="group cursor-pointer transition-all duration-200 hover:-translate-y-0.5 border-0 shadow-none bg-transparent"
      onClick={() => onDetail(model)}
    >
      <CardContent className="p-4 sm:p-5">
        {/* Top row: Provider icon + name */}
        <div className="flex items-center gap-2.5 mb-3">
          <ProviderIcon type={model.provider_type} name={model.provider} size="sm" />
          <span className="text-xs font-medium text-muted-foreground truncate">
            {model.provider}
          </span>
        </div>

        {/* Model name */}
        <h3 className="text-base sm:text-lg font-semibold tracking-tight mb-3 truncate">
          {model.name}
        </h3>

        {/* Bottom row: prices + status */}
        <div className="flex items-center justify-between gap-2 flex-wrap">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span className="inline-flex items-center gap-1 px-2 py-0.5">
              <span className="font-medium text-emerald-600 dark:text-emerald-400">IN</span>
              ${model.input_price.toFixed(4)}/1K
            </span>
            <span className="inline-flex items-center gap-1 px-2 py-0.5">
              <span className="font-medium text-amber-600 dark:text-amber-400">OUT</span>
              ${model.output_price.toFixed(4)}/1K
            </span>
          </div>
          <Badge
            variant={isActive ? 'default' : 'secondary'}
            className="text-[10px] px-1.5 py-0 h-5"
          >
            {isActive ? t('Active') : t('Inactive')}
          </Badge>
        </div>
      </CardContent>
    </Card>
  )
}
