import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "../lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors",
  {
    variants: {
      variant: {
        default: "border-[var(--color-volt)]/30 bg-[var(--color-volt)]/10 text-[var(--color-volt)]",
        success:
          "border-[var(--color-convert-green)]/30 bg-[var(--color-convert-green)]/10 text-[var(--color-convert-green)]",
        destructive:
          "border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/10 text-[var(--color-signal-red)]",
        outline: "border-white/20 text-white/60",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}

export { Badge, badgeVariants };
