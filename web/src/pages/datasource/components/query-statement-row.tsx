import { cn } from "@/lib/class-names"
import type { KeyboardEvent, ReactNode } from "react"

interface QueryStatementRowProps {
  title: string
  /** 可选的第二行：语句预览（等宽）。 */
  statement?: string
  /** 可选的第三行：时间、耗时等次要信息。 */
  meta?: string
  /** 有值时才显示状态点：成功为绿色，失败为红色。 */
  ok?: boolean
  /** 悬停时显示的完整内容，长语句被截断后仍可看全。 */
  tooltip?: string
  disabled?: boolean
  onPrimary: () => void
  actions?: ReactNode
}

// 模板行和历史行共用一套视觉：标题 + 语句预览 + 次要信息，
// 悬停或键盘聚焦时右侧才露出次级操作，避免列表被按钮撑乱。
export function QueryStatementRow({
  title,
  statement,
  meta,
  ok,
  tooltip,
  disabled = false,
  onPrimary,
  actions,
}: QueryStatementRowProps) {
  return (
    <div
      role="listitem"
      className="group/row flex items-start gap-1 rounded-md hover:bg-accent"
    >
      <button
        type="button"
        data-row-primary="true"
        disabled={disabled}
        title={tooltip}
        onClick={onPrimary}
        className="flex min-w-0 flex-1 flex-col items-start gap-1 rounded-md px-2 py-1.5 text-left outline-none focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50"
      >
        <span className="flex w-full min-w-0 items-center gap-1.5">
          {ok !== undefined ? (
            <span
              aria-hidden="true"
              className={cn(
                "size-1.5 shrink-0 rounded-full",
                ok ? "bg-emerald-500" : "bg-red-500"
              )}
            />
          ) : null}
          <span className="min-w-0 flex-1 truncate text-sm">{title}</span>
        </span>
        {statement ? (
          <span className="w-full truncate font-mono text-xs text-muted-foreground">
            {statement}
          </span>
        ) : null}
        {meta ? (
          <span className="w-full truncate text-xs text-muted-foreground">
            {meta}
          </span>
        ) : null}
      </button>
      {actions ? (
        <div className="flex shrink-0 items-center gap-0.5 pt-1 opacity-0 group-focus-within/row:opacity-100 group-hover/row:opacity-100">
          {actions}
        </div>
      ) : null}
    </div>
  )
}

/** 列表内 ↑/↓/Home/End 在行之间移动焦点，不拦截其它按键。 */
export function handleRowKeyDown(event: KeyboardEvent<HTMLElement>) {
  if (
    event.key !== "ArrowDown" &&
    event.key !== "ArrowUp" &&
    event.key !== "Home" &&
    event.key !== "End"
  ) {
    return
  }
  const rows = Array.from(
    event.currentTarget.querySelectorAll<HTMLButtonElement>(
      '[data-row-primary="true"]:not(:disabled)'
    )
  )
  if (rows.length === 0) {
    return
  }
  const active = document.activeElement
  const index = active instanceof HTMLButtonElement ? rows.indexOf(active) : -1
  const last = rows.length - 1
  const next =
    event.key === "ArrowDown"
      ? index < 0
        ? 0
        : Math.min(index + 1, last)
      : event.key === "ArrowUp"
        ? index < 0
          ? last
          : Math.max(index - 1, 0)
        : event.key === "Home"
          ? 0
          : last
  event.preventDefault()
  rows[next]?.focus()
}

type Translate = (key: string, options?: { count?: number }) => string

/** 相对时间文案；绝对时间放进 tooltip，不占列表宽度。 */
export function relativeTime(at: number, t: Translate): string {
  const diff = Date.now() - at
  if (diff < 60_000) {
    return t("datasource.query.justNow")
  }
  const minutes = Math.floor(diff / 60_000)
  if (minutes < 60) {
    return t("datasource.query.minutesAgo", { count: minutes })
  }
  const hours = Math.floor(minutes / 60)
  if (hours < 24) {
    return t("datasource.query.hoursAgo", { count: hours })
  }
  const days = Math.floor(hours / 24)
  return t("datasource.query.daysAgo", { count: days })
}
