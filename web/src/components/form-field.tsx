import { CircleHelp } from "lucide-react"
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react"

import { Label } from "@/components/ui/label"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { cn } from "@/lib/class-names"

/**
 * 复合控件（标签编辑器、时间范围等）无法自己渲染 label 行，把内部校验结果上报给
 * 最近的 FormField，由 FormField 统一显示在 label 右侧的错误插槽里。
 */
interface FormFieldErrorSlot {
  report: (message: string) => void
}

const FormFieldErrorContext = createContext<FormFieldErrorSlot | null>(null)

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
  /** 分组字段（如升级步骤）只有标题和错误行、控件由子项自行渲染时可以省略。 */
  children?: ReactNode
}

/**
 * 表单字段统一出口：每个字段都在 label 行左侧显示名称（必填带红色星号），
 * 校验失败时把原因显示在同一个 label 行的右侧，而不是控件下方：
 * 错误只占 label 行原有的余量（单个文本节点，nowrap + 省略号 + title），
 * 不换行、不撑高 label 行、不把控件往下挤，也不撑宽栅格列。
 */
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
  const [reportedError, setReportedError] = useState("")
  const slot = useMemo<FormFieldErrorSlot>(
    () => ({
      report: (message) =>
        setReportedError((current) => (current === message ? current : message)),
    }),
    []
  )
  const errorText = error && error.trim() !== "" ? error : reportedError

  return (
    <div
      data-slot="form-field"
      data-invalid={errorText ? "true" : undefined}
      className={cn("grid min-w-0 gap-2 [&_[data-slot=select-trigger]]:w-full", className)}
    >
      <div
        data-slot="form-field-label-row"
        className="flex min-w-0 flex-wrap items-center gap-2"
      >
        <Label htmlFor={htmlFor} aria-required={required || undefined} className="min-w-0">
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
                className="shrink-0 text-muted-foreground"
                aria-label={hint}
              >
                <CircleHelp className="size-4" />
              </button>
            </TooltipTrigger>
            <TooltipContent>{hint}</TooltipContent>
          </Tooltip>
        ) : null}
        <div
          data-slot="form-field-meta"
          className="ml-auto flex min-w-0 flex-1 items-center justify-end gap-2"
        >
          {/* 错误插槽放在 counter/actions 之前：错误出现时只占 label 行剩余空间，
              右侧的计数和操作按钮位置不变，避免整行内容左右跳动。 */}
          {errorText ? (
            <span
              data-slot="form-field-error"
              role="alert"
              title={errorText}
              className="min-w-0 flex-1 truncate text-right text-xs leading-5 text-destructive"
            >
              {errorText}
            </span>
          ) : null}
          {counter ? (
            <span className="shrink-0 font-mono text-xs tabular-nums text-muted-foreground">
              {counter.value}/{counter.max}
            </span>
          ) : null}
          {actions ? (
            <span className="flex min-w-0 shrink-0 items-center gap-2">{actions}</span>
          ) : null}
        </div>
      </div>
      {description ? (
        <p className="text-xs leading-5 text-muted-foreground">{description}</p>
      ) : null}
      <FormFieldErrorContext.Provider value={slot}>
        {children}
      </FormFieldErrorContext.Provider>
    </div>
  )
}

/** 复合控件把当前校验结果上报到所在 FormField 的 label 行错误插槽；不在 FormField 内时什么也不做。 */
export function useReportFormFieldError(message: string): void {
  const slot = useContext(FormFieldErrorContext)
  useEffect(() => {
    slot?.report(message)
  }, [slot, message])
  useEffect(() => {
    if (!slot) {
      return
    }
    return () => slot.report("")
  }, [slot])
}

/**
 * 重复键值行（标签、配置项等）的列名：每一行的键和值都有可见的列 label，
 * 不依赖 placeholder。列宽跟随行栅格，不额外占纵向高度以外的空间。
 */
export function KeyValueRowLabels({
  keyLabel,
  valueLabel,
  className,
  trailing,
  inline = false,
}: {
  keyLabel: string
  valueLabel: string
  className?: string
  trailing?: ReactNode
  /** 行本身是 flex 布局（输入框并排）时用 inline，列宽直接跟随输入框。 */
  inline?: boolean
}) {
  return (
    <div
      aria-hidden="true"
      className={cn(
        inline ? "flex min-w-0 items-center gap-2" : "grid min-w-0 gap-2 sm:grid-cols-2",
        "text-xs leading-5 text-muted-foreground",
        className
      )}
    >
      <span className="min-w-0 flex-1 truncate">{keyLabel}</span>
      <span className="min-w-0 flex-1 truncate">{valueLabel}</span>
      {trailing ?? (inline ? <span className="w-8 shrink-0" /> : null)}
    </div>
  )
}
