import { Search, X } from "lucide-react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { FilterChips, type FilterChip } from "@/components/filter-chips"
import { FilterToggle } from "@/components/filter-toggle"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/class-names"

interface ListFilterProps {
  /** 筛选面板的 DOM id，FilterToggle 通过 aria-controls 指向它。 */
  id: string
  open: boolean
  onOpenChange: (open: boolean) => void
  /** 已生效的条件数量；缺省时按 chips 数量计算。 */
  count?: number
  chips?: FilterChip[]
  onClearAll?: () => void
  /** 清空已生效条件；缺省时沿用 onClearAll。 */
  onClear?: () => void
  /** 提交草稿条件；筛选面板内的 form 只负责触发这一个入口。 */
  onApply: () => void
  /** 已生效条件数，用于禁用「清除筛选」。 */
  appliedCount?: number
  /** 草稿与已生效条件不一致时，页脚显示「有未应用的改动」。 */
  dirty?: boolean
  /** 没有草稿改动和生效条件时页脚左侧的说明。 */
  hint?: string
  /** 与筛选按钮同一行的右侧操作（统计、批量操作、视图切换）。 */
  actions?: ReactNode
  children: ReactNode
}

/**
 * 列表页统一的筛选区：收起/展开按钮、已生效条件标签、可折叠的条件面板，
 * 以及固定的「清除筛选 / 应用筛选」页脚。交互和实时告警页保持一致，
 * 条件字段由各页面按自己的接口参数传入。
 */
export function ListFilter({
  id,
  open,
  onOpenChange,
  count,
  chips,
  onClearAll,
  onClear,
  onApply,
  appliedCount,
  dirty = false,
  hint,
  actions,
  children,
}: ListFilterProps) {
  const { t } = useTranslation()
  const applied = appliedCount ?? chips?.length ?? count ?? 0

  return (
    <div className="flex min-w-0 flex-col gap-2">
      <div className="flex flex-wrap items-center justify-end gap-2">
        <FilterToggle
          open={open}
          count={count ?? applied}
          controls={id}
          onClick={() => onOpenChange(!open)}
        />
        {actions}
      </div>
      {chips && chips.length > 0 ? (
        <FilterChips items={chips} onClearAll={onClearAll} />
      ) : null}
      <form
        id={id}
        onSubmit={(event) => {
          event.preventDefault()
          onApply()
        }}
        className={cn(
          "min-w-0 rounded-xl border border-border/70 bg-card/70 p-3 sm:p-4",
          !open && "hidden"
        )}
      >
        <div className="flex min-w-0 flex-wrap items-end gap-3">
          {children}
          {/* 清除/应用和条件同一行；一行放不下时整组换到下一行。 */}
          <div className="ml-auto flex flex-wrap items-center gap-2">
            {dirty ? (
              <span
                role="status"
                className="text-xs whitespace-nowrap text-warning"
              >
                {t("filter.unapplied")}
              </span>
            ) : null}
            <Button
              type="button"
              variant="ghost"
              size="sm"
              disabled={applied === 0}
              onClick={onClear ?? onClearAll}
            >
              <X aria-hidden="true" data-icon="inline-start" />
              {t("filter.clear")}
            </Button>
            <Button type="submit">
              <Search aria-hidden="true" data-icon="inline-start" />
              {t("filter.apply")}
            </Button>
          </div>
        </div>
        {hint ? (
          <p className="mt-2 text-xs text-muted-foreground">{hint}</p>
        ) : null}
      </form>
    </div>
  )
}
