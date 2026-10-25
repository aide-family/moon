import { Layers } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"

import type { AlertInsightLevel } from "@/api/alert-event/types"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

import { formatCount, formatPercent, share } from "./lib"

/**
 * 级别分布。分母固定是区间内的事件总数（eventsTotal），不是当前级别行的合计，
 * 所以占比之和可能小于 100%（未分级、级别已删除的事件都在分母里）。
 */
export function LevelBreakdownCard({
  levels,
  eventsTotal,
  names,
  loading,
  failed,
  days,
}: {
  levels: AlertInsightLevel[] | null
  eventsTotal: number
  names: Record<string, string>
  loading: boolean
  failed: boolean
  days: number
}) {
  const { t } = useTranslation()
  const rows = [...(levels ?? [])].sort((left, right) => right.events - left.events)
  const showEmpty = rows.length === 0 || eventsTotal === 0

  return (
    <Card className="flex min-w-0 flex-col">
      <div className="flex items-start justify-between gap-3 px-4">
        <div className="min-w-0 space-y-0.5">
          <h2 className="font-heading text-sm font-medium">
            {t("overview.level.title")}
          </h2>
          <p className="text-xs leading-relaxed text-muted-foreground">
            {t("overview.level.description", { days })}
          </p>
        </div>
        <span className="shrink-0 rounded-md border border-border/70 bg-muted/40 px-2 py-1 text-[11px] whitespace-nowrap text-muted-foreground tabular-nums">
          {t("overview.level.total", { n: formatCount(eventsTotal) })}
        </span>
      </div>

      <div className="min-h-0 flex-1 px-2">
        {loading && !levels ? (
          <div className="space-y-2 px-2 py-2">
            {[0, 1, 2, 3, 4].map((index) => (
              <Skeleton key={index} className="h-7 w-full" />
            ))}
          </div>
        ) : showEmpty ? (
          <div className="flex h-32 flex-col items-center justify-center gap-1 px-2 text-center">
            <Layers className="size-5 text-muted-foreground/70" aria-hidden="true" />
            <p className="text-sm font-medium">
              {failed
                ? t("overview.level.unavailable")
                : t("overview.level.empty", { days })}
            </p>
            <p className="text-xs text-muted-foreground">
              {failed
                ? t("overview.level.unavailableHint")
                : t("overview.level.emptyHint")}
            </p>
          </div>
        ) : (
          <ul className="space-y-0.5 py-1">
            {rows.map((row) => {
              const name = names[row.levelUid] || t("overview.level.unranked")
              const ratio = eventsTotal > 0 ? row.events / eventsTotal : 0
              const width = share(row.events, eventsTotal)
              return (
                <li key={row.levelUid || "unranked"}>
                  <Link
                    to="/alert-history"
                    className="block rounded-md px-2 py-1.5 transition-colors hover:bg-muted/50 focus-visible:bg-muted/50 focus-visible:outline-none"
                  >
                    <div className="flex items-center gap-2">
                      <span className="min-w-0 flex-1 truncate text-xs font-medium">
                        {name}
                      </span>
                      {row.firing > 0 ? (
                        <span className="shrink-0 rounded border border-amber-500/30 bg-amber-500/10 px-1.5 text-[10px] text-amber-700 tabular-nums dark:text-amber-300">
                          {t("overview.level.firing", {
                            n: formatCount(row.firing),
                          })}
                        </span>
                      ) : null}
                      <span className="w-10 shrink-0 text-right text-xs font-semibold tabular-nums">
                        {formatCount(row.events)}
                      </span>
                      <span className="w-11 shrink-0 text-right text-[11px] text-muted-foreground tabular-nums">
                        {formatPercent(ratio, 1)}
                      </span>
                    </div>
                    <div className="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-muted">
                      <span
                        className="block h-full rounded-full bg-primary"
                        style={{ width: `${width}%` }}
                      />
                    </div>
                  </Link>
                </li>
              )
            })}
          </ul>
        )}
      </div>

      {showEmpty ? null : (
        <p className="px-4 text-xs text-muted-foreground">
          {t("overview.level.note")}
        </p>
      )}
    </Card>
  )
}
