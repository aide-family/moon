import type { ReactNode } from "react"

import { cn } from "@/lib/class-names"

interface ExprBlockProps {
  // 规则表达式原文，保留换行和空格。
  expr: string
  // 表格里用更紧凑的内边距。
  className?: string
  // 按场景限制可见高度，超出后块内滚动。
  maxHeightClass?: string
  // 表达式下方的补充信息，例如命中的级别条件。
  footer?: ReactNode
}

/**
 * 告警表达式统一展示块。实时告警详情、规则列表和规则详情共用同一套排版：
 * 等宽字体、保留换行、长表达式只在块内纵向滚动，不撑宽容器。
 */
export function ExprBlock({
  expr,
  className,
  maxHeightClass = "max-h-64",
  footer,
}: ExprBlockProps) {
  return (
    <div
      className={cn(
        "grid min-w-0 gap-2 rounded-md border bg-muted/40 p-3",
        className
      )}
    >
      <pre
        className={cn(
          "overflow-x-hidden overflow-y-auto font-mono text-xs break-words whitespace-pre-wrap",
          maxHeightClass
        )}
      >
        {expr || "—"}
      </pre>
      {footer}
    </div>
  )
}
