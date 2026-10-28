import { SlidersHorizontal } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"

interface FilterToggleProps {
  open: boolean
  /** 已应用的条件数量，显示在按钮里；为 0 时不显示。 */
  count?: number
  controls?: string
  onClick: () => void
}

/**
 * 筛选面板的收起/展开按钮。历史告警、实时告警和消息日志共用同一个外观，
 * 由各页面把它和右侧的操作（统计、批量操作、视图切换）放在同一行。
 */
export function FilterToggle({
  open,
  count = 0,
  controls,
  onClick,
}: FilterToggleProps) {
  const { t } = useTranslation()
  return (
    <Button
      type="button"
      variant={open ? "secondary" : "outline"}
      aria-expanded={open}
      aria-controls={controls}
      onClick={onClick}
    >
      <SlidersHorizontal aria-hidden="true" data-icon="inline-start" />
      {open ? t("filter.hide") : t("filter.toggle")}
      {count > 0 ? (
        <span className="ml-1 rounded bg-primary/10 px-1.5 font-mono text-xs text-primary-ink tabular-nums">
          {count}
        </span>
      ) : null}
    </Button>
  )
}
