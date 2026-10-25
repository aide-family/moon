import { BarChart3 } from "lucide-react"
import { useTranslation } from "react-i18next"

import type { AlertInsight } from "@/api/alert-event/types"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

import {
  bucketDayKey,
  bucketClosed,
  formatCount,
  formatDayLabel,
  formatSeconds,
  formatSigned,
  niceAxisMax,
} from "./lib"

// 新增与恢复用两种颜色区分，但颜色不是唯一线索：图例带数值，数据表给精确值。
const newBarClass = "bg-primary"
const closedBarClass = "bg-emerald-500"

/**
 * 近 N 天告警趋势。选择「并排」而不是「堆叠」：当日恢复的事件不一定当日新增
 * （可能是前几天触发的），堆叠高度并不构成一个有意义的合计，
 * 两根柱各自从 0 起，读数方式写在卡片描述里。
 */
export function InsightTrendCard({
  insight,
  days,
  loading,
  failed,
}: {
  insight: AlertInsight | null
  /** 请求的天数；为空区间时也要能写出「近 N 天」。 */
  days: number
  loading: boolean
  failed: boolean
}) {
  const { t } = useTranslation()
  const buckets = insight?.buckets ?? []
  const eventsTotal = insight?.eventsTotal ?? 0
  const closedTotal = insight?.closedTotal ?? 0
  // 进 - 出 = 存量净变化：新增和闭环是两条独立时间轴，相减才有意义。
  const net = eventsTotal - closedTotal
  const max = niceAxisMax(
    buckets.reduce(
      (peak, bucket) =>
        Math.max(peak, bucket.events, bucketClosed(bucket)),
      0
    )
  )
  const mid = Math.round(max / 2)

  const hasSamples = (insight?.recoverySamples ?? 0) > 0
  const p50 = hasSamples ? formatSeconds(insight?.p50Seconds ?? 0, t) : "—"
  const p90 = hasSamples ? formatSeconds(insight?.p90Seconds ?? 0, t) : "—"

  return (
    <Card className="min-w-0">
      <div className="flex items-start justify-between gap-3 px-4">
        <div className="min-w-0 space-y-0.5">
          <h2 className="font-heading text-sm font-medium">
            {t("overview.insight.title", { days })}
          </h2>
          <p className="text-xs leading-relaxed text-muted-foreground">
            {t("overview.insight.description")}
          </p>
        </div>
        <p
          className="shrink-0 rounded-md border border-border/70 bg-muted/40 px-2 py-1 text-[11px] whitespace-nowrap text-muted-foreground tabular-nums"
          title={t("overview.insight.p50P90Hint", {
            samples: formatCount(insight?.recoverySamples ?? 0),
          })}
        >
          {t("overview.insight.p50P90", { p50, p90 })}
        </p>
      </div>

      <div className="px-4">
        <dl className="mb-2 flex items-center gap-3 text-[11px]">
          <div className="flex items-center gap-1.5">
            <span
              className={`size-2 rounded-full ${newBarClass}`}
              aria-hidden="true"
            />
            <dt className="text-muted-foreground">{t("overview.insight.new")}</dt>
            <dd className="font-semibold tabular-nums">
              {formatCount(eventsTotal)}
            </dd>
          </div>
          <div className="flex items-center gap-1.5">
            <span
              className={`size-2 rounded-full ${closedBarClass}`}
              aria-hidden="true"
            />
            <dt className="text-muted-foreground">
              {t("overview.insight.closed")}
            </dt>
            <dd className="font-semibold tabular-nums">
              {formatCount(closedTotal)}
            </dd>
          </div>
          <div className="flex items-center gap-1.5">
            <dt className="text-muted-foreground">
              {t("overview.insight.netLabel")}
            </dt>
            <dd className="font-semibold tabular-nums">{formatSigned(net)}</dd>
          </div>
        </dl>

        {loading && !insight ? (
          <Skeleton className="h-32 w-full" />
        ) : buckets.length === 0 || (eventsTotal === 0 && closedTotal === 0) ? (
          <div className="flex h-32 flex-col items-center justify-center gap-1 text-center">
            <BarChart3
              className="size-5 text-muted-foreground/70"
              aria-hidden="true"
            />
            <p className="text-sm font-medium">
              {failed
                ? t("overview.insight.unavailable")
                : t("overview.insight.empty", { days })}
            </p>
            <p className="text-xs text-muted-foreground">
              {failed
                ? t("overview.insight.unavailableHint")
                : t("overview.insight.emptyHint")}
            </p>
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
                className="relative flex h-32 items-end gap-1 border-b border-border/60"
                role="img"
                aria-label={t("overview.insight.aria", {
                  days,
                  timeZone: insight?.timeZone ?? "UTC",
                  events: formatCount(eventsTotal),
                  closed: formatCount(closedTotal),
                  net: formatSigned(net),
                  p50,
                  p90,
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
                {buckets.map((bucket) => (
                  <div
                    key={bucket.start}
                    className="group relative flex h-full flex-1 items-end justify-center gap-[2px] rounded-sm transition-colors hover:bg-muted/50"
                    title={t("overview.insight.point", {
                      date: bucketDayKey(bucket.start),
                      events: formatCount(bucket.events),
                      closed: formatCount(bucketClosed(bucket)),
                    })}
                  >
                    <span
                      className={`w-full max-w-4 rounded-t-[2px] ${newBarClass}`}
                      style={{ height: `${(bucket.events / max) * 100}%` }}
                    />
                    <span
                      className={`w-full max-w-4 rounded-t-[2px] ${closedBarClass}`}
                      style={{
                        height: `${(bucketClosed(bucket) / max) * 100}%`,
                      }}
                    />
                  </div>
                ))}
              </div>
              <div className="mt-1.5 flex gap-1">
                {buckets.map((bucket, index) => (
                  <span
                    key={bucket.start}
                    className="flex-1 text-center text-[10px] tabular-nums text-muted-foreground"
                  >
                    <span
                      className={
                        index % 2 === 1 ? "hidden md:inline" : undefined
                      }
                    >
                      {formatDayLabel(bucketDayKey(bucket.start))}
                    </span>
                  </span>
                ))}
              </div>
            </div>

            {insight && !insight.complete ? (
              <p className="mt-2 rounded-md border border-amber-500/30 bg-amber-500/10 px-2.5 py-1.5 text-xs text-amber-700 dark:text-amber-300">
                {t("overview.insight.incomplete")}
              </p>
            ) : null}

            <details className="mt-2">
              <summary className="cursor-pointer text-xs text-muted-foreground select-none hover:text-foreground">
                {t("overview.insight.table")}
              </summary>
              <div className="mt-2 overflow-x-auto">
                <table className="w-full text-left text-xs tabular-nums">
                  <caption className="sr-only">
                    {t("overview.insight.tableCaption", { days })}
                  </caption>
                  <thead className="text-muted-foreground">
                    <tr className="border-b border-border/60">
                      <th scope="col" className="py-1 font-normal">
                        {t("overview.insight.day")}
                      </th>
                      <th scope="col" className="py-1 text-right font-normal">
                        {t("overview.insight.new")}
                      </th>
                      <th scope="col" className="py-1 text-right font-normal">
                        {t("overview.insight.closed")}
                      </th>
                      <th scope="col" className="py-1 text-right font-normal">
                        {t("overview.insight.p50Column")}
                      </th>
                      <th scope="col" className="py-1 text-right font-normal">
                        {t("overview.insight.p90Column")}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {buckets.map((bucket) => (
                      <tr
                        key={bucket.start}
                        className="border-b border-border/40 last:border-0"
                      >
                        <th scope="row" className="py-1 font-normal">
                          {bucketDayKey(bucket.start)}
                        </th>
                        <td className="py-1 text-right">
                          {formatCount(bucket.events)}
                        </td>
                        <td className="py-1 text-right">
                          {formatCount(bucketClosed(bucket))}
                        </td>
                        <td className="py-1 text-right text-muted-foreground">
                          {bucket.recoverySamples > 0
                            ? formatSeconds(bucket.p50Seconds, t)
                            : "—"}
                        </td>
                        <td className="py-1 text-right text-muted-foreground">
                          {bucket.recoverySamples > 0
                            ? formatSeconds(bucket.p90Seconds, t)
                            : "—"}
                        </td>
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
