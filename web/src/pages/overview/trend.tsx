import { BarChart3 } from "lucide-react"
import { useTranslation } from "react-i18next"

import type { DeliveryTrend } from "@/api/dashboard"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

import { formatCount, formatDayLabel } from "./lib"

/** 把最大值向上取整到 1/2/5 × 10^n，坐标轴刻度才是人读得懂的数。 */
function niceMax(value: number): number {
  if (value <= 0) {
    return 1
  }
  const exponent = Math.floor(Math.log10(value))
  const base = 10 ** exponent
  const normalized = value / base
  const step =
    normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 5 ? 5 : 10
  return step * base
}

/**
 * 近 7 天投递趋势。柱图从 0 起（长度编码），成功/失败堆叠只有两段，
 * 每根柱子都能在数据表里读到精确值，颜色不是唯一的信息来源。
 */
export function DeliveryTrendCard({
  trend,
  loading,
}: {
  trend: DeliveryTrend | null
  loading: boolean
}) {
  const { t } = useTranslation()
  const points = trend?.points ?? []
  const max = niceMax(
    points.reduce((peak, point) => Math.max(peak, point.sent + point.failed), 0)
  )
  const mid = Math.round(max / 2)
  const totalSent = points.reduce((sum, point) => sum + point.sent, 0)
  const totalFailed = points.reduce((sum, point) => sum + point.failed, 0)

  return (
    <Card className="min-w-0">
      <div className="flex items-start justify-between gap-3 px-4">
        <div className="min-w-0 space-y-0.5">
          <h2 className="font-heading text-sm font-medium">
            {t("overview.trend.title")}
          </h2>
          <p className="text-xs text-muted-foreground">
            {t("overview.trend.description")}
          </p>
        </div>
        <dl className="flex shrink-0 items-center gap-3 text-[11px]">
          <div className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-primary" aria-hidden="true" />
            <dt className="text-muted-foreground">{t("overview.delivery.sent")}</dt>
            <dd className="font-semibold tabular-nums">{formatCount(totalSent)}</dd>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-red-500" aria-hidden="true" />
            <dt className="text-muted-foreground">{t("overview.delivery.failed")}</dt>
            <dd className="font-semibold tabular-nums">{formatCount(totalFailed)}</dd>
          </div>
        </dl>
      </div>

      <div className="px-4">
        {loading && !trend ? (
          <Skeleton className="h-32 w-full" />
        ) : points.length === 0 ? (
          <div className="flex h-32 flex-col items-center justify-center gap-1 text-center">
            <BarChart3 className="size-5 text-muted-foreground/70" aria-hidden="true" />
            <p className="text-sm font-medium">{t("overview.trend.empty")}</p>
          </div>
        ) : (
          <>
            <div className="relative pl-8">
              <div className="absolute top-0 left-0 flex h-32 w-7 flex-col justify-between text-right text-[10px] tabular-nums text-muted-foreground">
                <span>{formatCount(max)}</span>
                <span>{formatCount(mid)}</span>
                <span>0</span>
              </div>
              <div
                className="relative flex h-32 items-end gap-1.5 border-b border-border/60"
                role="img"
                aria-label={t("overview.trend.aria", {
                  days: points.length,
                  sent: formatCount(totalSent),
                  failed: formatCount(totalFailed),
                })}
              >
                <div
                  className="pointer-events-none absolute inset-x-0 top-0 border-t border-dashed border-border/50"
                  aria-hidden="true"
                />
                <div
                  className="pointer-events-none absolute inset-x-0 top-1/2 border-t border-dashed border-border/50"
                  aria-hidden="true"
                />
                {points.map((point) => {
                  const sentHeight = (point.sent / max) * 100
                  const failedHeight = (point.failed / max) * 100
                  return (
                    <div
                      key={point.date}
                      className="group relative flex h-full flex-1 flex-col justify-end gap-px rounded-sm transition-colors hover:bg-muted/50"
                      title={t("overview.trend.point", {
                        date: point.date,
                        sent: formatCount(point.sent),
                        failed: formatCount(point.failed),
                      })}
                    >
                      <span
                        className="w-full rounded-t-[2px] bg-red-500"
                        style={{ height: `${failedHeight}%` }}
                      />
                      <span
                        className="w-full bg-primary"
                        style={{ height: `${sentHeight}%` }}
                      />
                    </div>
                  )
                })}
              </div>
              <div className="mt-1.5 flex gap-1.5">
                {points.map((point) => (
                  <span
                    key={point.date}
                    className="flex-1 text-center text-[10px] tabular-nums text-muted-foreground"
                  >
                    {formatDayLabel(point.date)}
                  </span>
                ))}
              </div>
            </div>

            {trend && !trend.complete ? (
              <p className="mt-2 rounded-md border border-amber-500/30 bg-amber-500/10 px-2.5 py-1.5 text-xs text-amber-700 dark:text-amber-300">
                {t("overview.trend.incomplete")}
              </p>
            ) : null}

            <details className="mt-2">
              <summary className="cursor-pointer text-xs text-muted-foreground select-none hover:text-foreground">
                {t("overview.trend.table")}
              </summary>
              <div className="mt-2 overflow-x-auto">
                <table className="w-full text-left text-xs tabular-nums">
                  <caption className="sr-only">
                    {t("overview.trend.tableCaption")}
                  </caption>
                  <thead className="text-muted-foreground">
                    <tr className="border-b border-border/60">
                      <th scope="col" className="py-1 font-normal">
                        {t("overview.trend.day")}
                      </th>
                      <th scope="col" className="py-1 text-right font-normal">
                        {t("overview.delivery.sent")}
                      </th>
                      <th scope="col" className="py-1 text-right font-normal">
                        {t("overview.delivery.failed")}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {points.map((point) => (
                      <tr key={point.date} className="border-b border-border/40 last:border-0">
                        <th scope="row" className="py-1 font-normal">
                          {point.date}
                        </th>
                        <td className="py-1 text-right">{formatCount(point.sent)}</td>
                        <td className="py-1 text-right">{formatCount(point.failed)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </details>
          </>
        )}
      </div>
    </Card>
  )
}
