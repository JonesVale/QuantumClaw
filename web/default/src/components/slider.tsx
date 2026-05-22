import * as React from 'react'
import { cn } from '@/lib/utils'

interface SliderProps {
  min?: number
  max?: number
  step?: number
  value?: number[]
  onValueChange?: (value: number[]) => void
  className?: string
  disabled?: boolean
}

export function Slider({
  min = 0,
  max = 1,
  step = 0.01,
  value = [0],
  onValueChange,
  className,
  disabled,
}: SliderProps) {
  const percent = ((value[0] - min) / (max - min)) * 100

  return (
    <div className={cn('relative h-2 w-full', className)}>
      <div className="absolute inset-0 rounded-full bg-muted">
        <div
          className="absolute inset-y-0 left-0 rounded-full bg-primary transition-all"
          style={{ width: `${percent}%` }}
        />
      </div>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={value[0]}
        onChange={(e) => onValueChange?.([parseFloat(e.target.value)])}
        disabled={disabled}
        className="absolute inset-0 w-full h-full cursor-pointer opacity-0"
      />
      {/* Custom thumb for visual */}
      <div
        className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 h-4 w-4 rounded-full border-2 border-primary bg-background shadow-sm pointer-events-none transition-all"
        style={{ left: `${percent}%` }}
      />
    </div>
  )
}
