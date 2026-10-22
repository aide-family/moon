import { CircleHelp } from "lucide-react"
import type { ReactNode } from "react"

import { Label } from "@/components/ui/label"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { cn } from "@/lib/class-names"

interface FormFieldProps {
  label: string
  htmlFor?: string
  hint?: string
  error?: string
  required?: boolean
  className?: string
  children: ReactNode
}

// 表单字段统一出口：必填项在 label 前显示红色星号，校验失败时给控件加红框并在下方显示原因。
export function FormField({
  label,
  htmlFor,
  hint,
  error,
  required = false,
  className,
  children,
}: FormFieldProps) {
  return (
    <div
      data-slot="form-field"
      data-invalid={error ? "true" : undefined}
      className={cn("grid min-w-0 gap-2 [&_[data-slot=select-trigger]]:w-full", className)}
    >
      <div className="flex items-center gap-2">
        <Label htmlFor={htmlFor} aria-required={required || undefined}>
          <span className="inline-flex items-center gap-1">
            {required ? (
              <span aria-hidden="true" className="text-destructive">
                *
              </span>
            ) : null}
            <span>{label}</span>
          </span>
        </Label>
        {hint ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                className="text-muted-foreground"
                aria-label={hint}
              >
                <CircleHelp className="size-4" />
              </button>
            </TooltipTrigger>
            <TooltipContent>{hint}</TooltipContent>
          </Tooltip>
        ) : null}
      </div>
      {children}
      {error ? (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}
    </div>
  )
}
