import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

interface ListFrameProps {
  total: number
  page: number
  pageSize: number
  showPager: boolean
  onPrev: () => void
  onNext: () => void
  children: ReactNode
}

export function ListFrame({
  total,
  page,
  pageSize,
  showPager,
  onPrev,
  onNext,
  children,
}: ListFrameProps) {
  const { t } = useTranslation()
  const pages = Math.max(1, Math.ceil(total / Math.max(pageSize, 1)))

  return (
    <div className={cn("glass-card flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl")}>
      <div className="data-table-control-bar justify-between gap-2">
        <span className="font-mono text-sm text-muted-foreground">
          {t("list.total", { count: total })}
        </span>
        {showPager ? (
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={onPrev}
            >
              {t("action.prev")}
            </Button>
            <span className="font-mono text-sm text-muted-foreground">
              {page} / {pages}
            </span>
            <Button
              variant="outline"
              size="sm"
              disabled={page * pageSize >= total}
              onClick={onNext}
            >
              {t("action.next")}
            </Button>
          </div>
        ) : null}
      </div>
      <div className="flex min-h-0 flex-1 flex-col overflow-hidden p-4">
        {children}
      </div>
    </div>
  )
}
