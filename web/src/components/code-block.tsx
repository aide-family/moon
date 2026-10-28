import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { CopyIconButton } from "@/components/copy-button"
import { useCodeHighlight } from "@/hooks/use-code-highlight"
import { cn } from "@/lib/class-names"

export interface CodeBlockProps {
  /** 原文。复制的是原文，不是高亮后的 HTML。 */
  code: string
  /** 语种标识，例如 yaml、json、bash、log；未知语种按纯文本展示。 */
  language?: string
  /** 代码块标题，通常放语种名或业务标题。 */
  label?: ReactNode
  /** 是否显示复制按钮，默认不显示。 */
  copyable?: boolean
  /** 复制内容，默认与 code 相同。 */
  copyValue?: string
  copyLabel?: string
  copiedLabel?: string
  copyErrorLabel?: string
  /** 长行折行显示；日志、模板渲染结果等需要折行时打开。 */
  wrap?: boolean
  /** 铺满父级剩余高度：父级必须是 min-h-0 的 flex 列，滚动发生在代码内容区。 */
  fill?: boolean
  /** 非 fill 时的最大高度，例如 max-h-80；滚动发生在代码内容区。 */
  maxHeightClass?: string
  /**
   * card：带边框、标题栏和主题背景的独立代码块；
   * plain：只输出代码内容，用于嵌在已有容器里的表达式、命令等。
   */
  variant?: "card" | "plain"
  className?: string
  /** 标题栏右侧、复制按钮之前的附加操作。 */
  actions?: ReactNode
  ariaLabel?: string
  /** 无内容时显示的占位符。 */
  placeholder?: ReactNode
}

/**
 * 全站统一的代码块：GitHub light/dark 配色、按需语法高亮、复制反馈。
 * 滚动归属固定在代码内容区，外层容器不再为代码块承担滚动。
 */
export function CodeBlock({
  code,
  language,
  label,
  copyable = false,
  copyValue,
  copyLabel,
  copiedLabel,
  copyErrorLabel,
  wrap = false,
  fill = false,
  maxHeightClass,
  variant = "card",
  className,
  actions,
  ariaLabel,
  placeholder,
}: CodeBlockProps) {
  const { t } = useTranslation()
  const html = useCodeHighlight(code, language)
  const plain = variant === "plain"
  const empty = code === ""
  const resolvedLabel = label ?? language?.toUpperCase()
  const showHeader = !plain && Boolean(resolvedLabel || copyable || actions)

  return (
    <div
      data-slot="code-block"
      data-variant={variant}
      className={cn(
        "flex min-w-0 flex-col",
        plain
          ? "overflow-hidden"
          : "overflow-hidden rounded-lg border border-border/70",
        fill && "min-h-0 flex-1",
        className
      )}
    >
      {showHeader ? (
        <div className="flex shrink-0 items-center justify-between gap-2 border-b border-border/60 bg-muted/40 px-3 py-1.5">
          <span className="min-w-0 truncate font-mono text-[11px] tracking-[0.12em] text-muted-foreground uppercase">
            {resolvedLabel}
          </span>
          <div className="flex shrink-0 items-center gap-1">
            {actions}
            {copyable ? (
              <CopyIconButton
                value={copyValue ?? code}
                label={copyLabel ?? t("action.copy")}
                copiedLabel={copiedLabel ?? t("action.copied")}
                errorLabel={copyErrorLabel ?? t("action.copyFailed")}
                disabled={empty}
              />
            ) : null}
          </div>
        </div>
      ) : null}
      <div
        className={cn(
          "code-block min-w-0",
          plain ? "bg-transparent" : "bg-[var(--code-bg)]",
          fill ? "min-h-0 flex-1 overflow-auto" : "overflow-auto",
          !fill && maxHeightClass
        )}
      >
        <pre
          tabIndex={0}
          aria-label={
            ariaLabel ?? (typeof label === "string" ? label : undefined)
          }
          className={cn(
            "min-w-full font-mono text-xs leading-5 text-[var(--code-fg)] outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset",
            plain ? "p-0" : "p-3",
            wrap ? "break-words whitespace-pre-wrap" : "whitespace-pre"
          )}
        >
          {html ? (
            <code dangerouslySetInnerHTML={{ __html: html }} />
          ) : (
            <code>{empty ? (placeholder ?? "—") : code}</code>
          )}
        </pre>
      </div>
    </div>
  )
}
