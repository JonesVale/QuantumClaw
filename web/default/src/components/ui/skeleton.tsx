import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const skeletonVariants = cva("relative overflow-hidden bg-primary/10", {
  variants: {
    variant: {
      default: "rounded-md",
      circular: "rounded-full",
    },
  },
  defaultVariants: {
    variant: "default",
  },
})

function Skeleton({
  className,
  variant,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & VariantProps<typeof skeletonVariants>) {
  return (
    <div
      className={cn(
        "animate-pulse",
        skeletonVariants({ variant }),
        className
      )}
      {...props}
    />
  )
}

export { Skeleton, skeletonVariants }
