import { ChevronDown } from "lucide-react"
import type { ReactNode } from "react"

import { cn } from "@/lib/class-names"

interface PageHeaderProps {
  title: string
  description?: string
  /** 传入后标题渲染为按钮，用于展开 / 收起标题下方的内容块。 */
  onTitleClick?: () => void
  /** 与 onTitleClick 搭配：内容块当前是否展开。 */
  titleExpanded?: boolean
  /** 标题按钮的无障碍名称，说明点击后会展开还是收起。 */
  titleToggleLabel?: string
  children?: ReactNode
}

export function PageHeader({
  title,
  description,
  onTitleClick,
  titleExpanded,
  titleToggleLabel,
  children,
}: PageHeaderProps) {
  return (
    <div className="flex items-start justify-between gap-4">
      <div>
        {onTitleClick ? (
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            <button
              type="button"
              className="flex items-center gap-1.5 rounded-md text-left transition-colors outline-none focus-visible:ring-3 focus-visible:ring-ring"
              onClick={onTitleClick}
              aria-expanded={titleExpanded}
              aria-label={titleToggleLabel}
            >
              <span>{title}</span>
              <ChevronDown
                aria-hidden="true"
                className={cn(
                  "size-4 shrink-0 text-muted-foreground transition-transform",
                  titleExpanded ? "rotate-180" : ""
                )}
              />
            </button>
          </h1>
        ) : (
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            {title}
          </h1>
        )}
        {description ? (
          <p className="text-sm text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {children}
    </div>
  )
}
