import { Check, Copy } from "lucide-react"
import { useEffect, useState } from "react"
import { toast } from "sonner"

import { cn } from "@/lib/class-names"

export interface CopyIconButtonProps {
  /** 复制到剪贴板的原文。 */
  value: string
  /** 未复制时的可访问标签，例如「复制命名空间」。 */
  label: string
  /** 复制成功后的可访问标签，例如「已复制」。 */
  copiedLabel: string
  /** 复制失败时的 toast 文案。 */
  errorLabel: string
  className?: string
  disabled?: boolean
}

/**
 * 复制图标按钮：成功后图标短暂切换为对勾并播报状态，失败时提示，不改变布局宽度。
 * 代码块标题栏和详情字段共用，保证两处的复制反馈一致。
 */
export function CopyIconButton({
  value,
  label,
  copiedLabel,
  errorLabel,
  className,
  disabled = false,
}: CopyIconButtonProps) {
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!copied) {
      return
    }
    const timer = window.setTimeout(() => setCopied(false), 1600)
    return () => window.clearTimeout(timer)
  }, [copied])

  async function copy() {
    if (!value) {
      return
    }
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
    } catch {
      toast.error(errorLabel)
    }
  }

  return (
    <button
      type="button"
      className={cn(
        "grid size-7 shrink-0 place-items-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-40",
        className
      )}
      aria-label={copied ? copiedLabel : label}
      title={copied ? copiedLabel : label}
      data-copied={copied}
      disabled={disabled}
      onClick={() => void copy()}
    >
      <span className="relative grid size-4 place-items-center">
        <Copy
          aria-hidden="true"
          className={cn(
            "absolute size-4 transition-all duration-200",
            copied ? "scale-50 opacity-0" : "scale-100 opacity-100"
          )}
        />
        <Check
          aria-hidden="true"
          className={cn(
            "absolute size-4 text-emerald-600 transition-all duration-200 dark:text-emerald-400",
            copied ? "scale-100 opacity-100" : "scale-50 opacity-0"
          )}
        />
      </span>
      <span aria-live="polite" className="sr-only">
        {copied ? copiedLabel : ""}
      </span>
    </button>
  )
}
