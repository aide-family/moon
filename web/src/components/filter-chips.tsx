import { X } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"

export interface FilterChip {
  key: string
  label: string
  onRemove: () => void
}

interface FilterChipsProps {
  items: FilterChip[]
  onClearAll?: () => void
  className?: string
}

/**
 * 已应用筛选的标签行。历史告警、实时告警和消息日志共用同一种外观，
 * 让“现在生效的是哪几个条件”在任何列表页都能一眼看到。
 */
export function FilterChips({
  items,
  onClearAll,
  className,
}: FilterChipsProps) {
  const { t } = useTranslation()
  if (items.length === 0) {
    return null
  }
  return (
    <div className={className ?? "flex flex-wrap items-center gap-1.5"}>
      {items.map((chip) => (
        <button
          key={chip.key}
          type="button"
          aria-label={t("filter.remove", { label: chip.label })}
          className="inline-flex items-center gap-1 rounded-full border border-border/70 bg-muted/40 py-0.5 pr-1.5 pl-2.5 text-xs text-foreground/80 transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          onClick={chip.onRemove}
        >
          <span className="max-w-[18rem] truncate">{chip.label}</span>
          <X aria-hidden="true" className="size-3.5" />
        </button>
      ))}
      {onClearAll ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 px-2 text-xs"
          onClick={onClearAll}
        >
          {t("filter.clearAll")}
        </Button>
      ) : null}
    </div>
  )
}
