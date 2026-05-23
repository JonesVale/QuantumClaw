import { useT } from '@/lib/use-t'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/textarea'
import { Slider } from '@/components/slider'
import { Separator } from '@/components/ui/separator'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface ChatParams {
  temperature: number
  top_p: number
  max_tokens: number
  frequency_penalty: number
  presence_penalty: number
  system_prompt: string
}

export const defaultParams: ChatParams = {
  temperature: 1,
  top_p: 1,
  max_tokens: 4096,
  frequency_penalty: 0,
  presence_penalty: 0,
  system_prompt: '',
}

interface ParameterPanelProps {
  params: ChatParams
  onParamsChange: (params: ChatParams) => void
  onClose: () => void
}

export function ParameterPanel({ params, onParamsChange, onClose }: ParameterPanelProps) {
  const { t } = useT()

  const update = (key: keyof ChatParams, value: number | string) => {
    onParamsChange({ ...params, [key]: value })
  }

  return (
    <div className="w-72 shrink-0 border-l bg-card p-4 overflow-y-auto space-y-5">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">{t('Parameters')}</h3>
        <Button variant="ghost" size="icon" className="h-6 w-6" onClick={onClose}>
          <X className="h-4 w-4" />
        </Button>
      </div>

      {/* Temperature */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label className="text-xs font-medium">{t('Temperature')}</Label>
          <span className="text-xs text-muted-foreground tabular-nums">{params.temperature.toFixed(2)}</span>
        </div>
        <Slider
          min={0}
          max={2}
          step={0.01}
          value={[params.temperature]}
          onValueChange={([v]) => update('temperature', v)}
        />
        <p className="text-[10px] text-muted-foreground leading-tight">{t('What sampling temperature to use')}</p>
      </div>

      {/* Top P */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label className="text-xs font-medium">Top P</Label>
          <span className="text-xs text-muted-foreground tabular-nums">{params.top_p.toFixed(2)}</span>
        </div>
        <Slider
          min={0}
          max={1}
          step={0.01}
          value={[params.top_p]}
          onValueChange={([v]) => update('top_p', v)}
        />
        <p className="text-[10px] text-muted-foreground leading-tight">{t('Nucleus sampling threshold')}</p>
      </div>

      {/* Max Tokens */}
      <div className="space-y-2">
        <Label className="text-xs font-medium">{t('Max Tokens')}</Label>
        <Input
          type="number"
          min={1}
          max={128000}
          value={params.max_tokens}
          onChange={(e) => update('max_tokens', Math.max(1, parseInt(e.target.value) || 1))}
          className="h-8 text-xs"
        />
      </div>

      <Separator />

      {/* Frequency Penalty */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label className="text-xs font-medium">{t('Frequency Penalty')}</Label>
          <span className="text-xs text-muted-foreground tabular-nums">{params.frequency_penalty.toFixed(2)}</span>
        </div>
        <Slider
          min={-2}
          max={2}
          step={0.01}
          value={[params.frequency_penalty]}
          onValueChange={([v]) => update('frequency_penalty', v)}
        />
      </div>

      {/* Presence Penalty */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label className="text-xs font-medium">{t('Presence Penalty')}</Label>
          <span className="text-xs text-muted-foreground tabular-nums">{params.presence_penalty.toFixed(2)}</span>
        </div>
        <Slider
          min={-2}
          max={2}
          step={0.01}
          value={[params.presence_penalty]}
          onValueChange={([v]) => update('presence_penalty', v)}
        />
      </div>

      <Separator />

      {/* System Prompt */}
      <div className="space-y-2">
        <Label className="text-xs font-medium">{t('System Prompt')}</Label>
        <Textarea
          value={params.system_prompt}
          onChange={(e) => update('system_prompt', e.target.value)}
          placeholder={t('You are a helpful assistant...')}
          className="min-h-[80px] text-xs resize-none"
          rows={4}
        />
        <p className="text-[10px] text-muted-foreground leading-tight">
          {t('Sets the context for the conversation')}
        </p>
      </div>
    </div>
  )
}
