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
  /** 常驻说明：格式、范围、后果，放在控件上方，不依赖 hover。 */
  description?: ReactNode
  /** 有长度上限的字段显示当前计数。 */
  counter?: { value: number; max: number }
  error?: string
  required?: boolean
  className?: string
  /** label 行右侧的次要操作（例如查询模板与历史入口），不额外占用纵向空间。 */
  actions?: ReactNode
  children: ReactNode
}

// 表单字段统一出口：必填项在 label 前显示红色星号，校验失败时给控件加红框并在下方显示原因。
export function FormField({
  label,
  htmlFor,
  hint,
  description,
  counter,
  error,
  required = false,
  className,
  actions,
  children,
}: FormFieldProps) {
  return (
    <div
      data-slot="form-field"
      data-invalid={error ? "true" : undefined}
      className={cn("grid min-w-0 gap-2 [&_[data-slot=select-trigger]]:w-full", className)}
    >
      <div className="flex flex-wrap items-center gap-2">
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
        {counter ? (
          <span className="ml-auto font-mono text-xs tabular-nums text-muted-foreground">
            {counter.value}/{counter.max}
          </span>
        ) : null}
        {actions}
      </div>
      {description ? (
        <p className="text-xs leading-5 text-muted-foreground">{description}</p>
      ) : null}
      {children}
      {error ? (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}
    </div>
  )
}
