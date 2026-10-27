import type { ReactNode } from "react"
import { useEffect, useMemo, useRef, useState } from "react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import {
  getAlertAnalysisFingerprints,
  getAlertAnalysisLevel,
  getAlertAnalysisRecovery,
  getAlertAnalysisStatus,
  getAlertAnalysisSummary,
  getAlertAnalysisTrend,
} from "@/api/alert-event"
import type {
  AlertAnalysisGranularity,
  AlertAnalysisRecoveryStats,
  AlertAnalysisStatusFilter,
  AlertAnalysisTrendBucket,
} from "@/api/alert-event"
import { Button } from "@/components/ui/button"
import { DateRangePicker } from "@/components/date-range-picker"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { addDays, formatDateValue } from "@/lib/date-range"
import { formatSpan } from "@/pages/alert/format"
import {
  analysisBlockData,
  analysisBlockIncomplete,
  ANALYSIS_BLOCK_ORDER,
  useAlertAnalysisBlocks,
} from "@/pages/alert/history-analysis-loader"
import type {
  AnalysisBlockStatus,
  AnalysisRequests,
  AnalysisSignatures,
} from "@/pages/alert/history-analysis-loader"
import {
  HISTORY_RANGE_MAX_DAYS,
  localDayWindow,
} from "@/pages/alert/history-range"

// 时间轴最多渲染 180 个时间桶；超出时退化为“只显示有事件的时间桶”。
const MAX_BUCKETS = 180
// 重复指纹接口固定取前 5 个；服务端已按事件数降序排好。
const FINGERPRINT_LIMIT = 5

type Granularity = AlertAnalysisGranularity
// 历史视图的事件终态只有三种：自动恢复、人工恢复、仍在触发（且被抑制）。
// 抑制是触发阶段的属性，不是终态，因此不再作为状态选项。
type StatusFilter = "ALL" | AlertAnalysisStatusFilter

interface AlertHistoryInsightsProps {
  enabled: boolean
  onSelectFingerprint: (fingerprint: string) => void
}

interface FingerprintCount {
  fingerprint: string
  ruleName: string
  count: number
}

interface DurationStats {
  median: number | null
  p90: number | null
  count: number
}

function localTimeZone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC"
  } catch {
    return "UTC"
  }
}

export function AlertHistoryInsights({
  enabled,
  onSelectFingerprint,
}: AlertHistoryInsightsProps) {
  const { t, i18n } = useTranslation()
  const [params, setParams] = useSearchParams()
  const urlFrom = params.get("from") || ""
  const urlTo = params.get("to") || ""
  const status = parseStatusFilter(params.get("status"))
  const granularity = parseGranularity(params.get("interval"))

  // 分析页永远有明确区间：URL 没写 from/to 时，有效区间是最近 30 天（含今天）。
  // 清空筛选会删掉 URL 参数，于是回到这里算出的默认区间，不再有“全部历史”。
  const today = formatDateValue(new Date())
  const defaultFrom = formatDateValue(
    addDays(new Date(), -(HISTORY_RANGE_MAX_DAYS - 1))
  )
  const from = urlFrom || defaultFrom
  const to = urlTo || today
  const invalidDateRange = from > to
  const defaultRange = !urlFrom && !urlTo
  const timeZone = useMemo(() => localTimeZone(), [])
  const statusParam: AlertAnalysisStatusFilter | "" =
    status === "ALL" ? "" : status

  const filterSignature = `${from}|${to}|${status}|${granularity}`
  const [selectedBucket, setSelectedBucket] = useState<{
    key: number
    filterSignature: string
  } | null>(null)
  const candidateBucketKey =
    selectedBucket?.filterSignature === filterSignature
      ? selectedBucket.key
      : null

  // 聚焦窗口 = 所选时间桶与统计区间的交集：周/月桶的起点可能早于 from，
  // 直接拿桶边界会被服务端判成“不在 from/to 窗口内”。
  const focus = useMemo(() => {
    if (candidateBucketKey === null) {
      return null
    }
    const fromWindow = localDayWindow(from)
    const toWindow = localDayWindow(to)
    if (!fromWindow || !toWindow) {
      return null
    }
    const start = Math.max(candidateBucketKey, fromWindow.start)
    const end = Math.min(
      advanceBucket(new Date(candidateBucketKey), granularity).getTime(),
      toWindow.end
    )
    if (!(end > start)) {
      return null
    }
    return {
      start: new Date(start).toISOString(),
      end: new Date(end).toISOString(),
    }
  }, [candidateBucketKey, from, to, granularity])

  const requests = useMemo<AnalysisRequests>(() => {
    const base = { from, to, timeZone, status: statusParam }
    const scoped = focus
      ? { ...base, focusStart: focus.start, focusEnd: focus.end }
      : base
    return {
      // 聚焦时 KPI 也要走聚焦口径：summary 与其余 4 块一样带 focusStart/focusEnd，
      // 只有趋势图（signature 不含聚焦）继续用整段区间。
      summary: () => getAlertAnalysisSummary(scoped),
      status: () => getAlertAnalysisStatus(scoped),
      level: () => getAlertAnalysisLevel(scoped),
      trend: () => getAlertAnalysisTrend({ ...base, granularity }),
      fingerprints: () =>
        getAlertAnalysisFingerprints({ ...scoped, limit: FINGERPRINT_LIMIT }),
      recovery: () => getAlertAnalysisRecovery(scoped),
    }
  }, [from, to, timeZone, statusParam, granularity, focus])

  // 聚焦只改变趋势图以外的 5 个块，因此 trend 的 signature 不含聚焦窗口：
  // 点击时间桶时不重发 trend，只重跑其余块。
  const signatures = useMemo<AnalysisSignatures>(() => {
    const base = `${from}|${to}|${statusParam}|${granularity}|${timeZone}`
    const scoped = focus ? `${base}|${focus.start}|${focus.end}` : `${base}|`
    return {
      summary: scoped,
      status: scoped,
      level: scoped,
      trend: base,
      fingerprints: scoped,
      recovery: scoped,
    }
  }, [from, to, statusParam, granularity, timeZone, focus])

  const { blocks, retry } = useAlertAnalysisBlocks({
    // 分析 tab 未激活或区间非法时一个请求都不发。
    enabled: enabled && !invalidDateRange,
    signatures,
    requests,
  })

  const summary = analysisBlockData(blocks.summary)
  const events = summary?.events ?? 0
  const allReady = ANALYSIS_BLOCK_ORDER.every(
    (block) => blocks[block].status === "ready"
  )
  const anyPending = ANALYSIS_BLOCK_ORDER.some(
    (block) =>
      blocks[block].status === "idle" || blocks[block].status === "loading"
  )
  const busy = enabled && !invalidDateRange && anyPending
  const empty = allReady && events === 0
  // 任意一个接口报告归档不可用时，统一在页面顶部降级提示。
  const degraded = ANALYSIS_BLOCK_ORDER.some((block) =>
    analysisBlockIncomplete(blocks[block])
  )

  const trend = analysisBlockData(blocks.trend)
  const timeline = useMemo(
    () =>
      buildTimeline(trend?.buckets ?? [], granularity, i18n.language, from, to),
    [trend, granularity, i18n.language, from, to]
  )
  // 聚焦徽标只显示当前趋势图里真实存在的时间桶；趋势重新加载后桶集合变化时，
  // 徽标自动消失，但筛选条件本身仍由 filterSignature 控制。
  const focusedPoint =
    candidateBucketKey === null
      ? null
      : (timeline.points.find((point) => point.key === candidateBucketKey) ??
        null)

  const statusTotal = analysisBlockData(blocks.status)?.total ?? 0
  const levelTotal = analysisBlockData(blocks.level)?.total ?? 0

  const statusRows = useMemo(() => {
    const data = analysisBlockData(blocks.status)
    if (!data) {
      return null
    }
    return new Map(data.items.map((item) => [item.status, item.events]))
  }, [blocks.status])

  const levelRows = useMemo(() => {
    const data = analysisBlockData(blocks.level)
    if (!data) {
      return null
    }
    return new Map(data.items.map((item) => [item.levelUid, item.events]))
  }, [blocks.level])

  // 级别分布要显示级别名而不是 UID：分析 tab 自己懒加载一次告警字典（很小的请求），
  // 只在分析 tab 激活时发。请求用 ref 缓存成 Promise：StrictMode 会把 effect 跑两遍，
  // 复用的 Promise 既保证只发一次，又不会像「已请求」布尔标记那样把结果丢掉。
  const [levelNames, setLevelNames] = useState<Map<string, string>>(
    () => new Map()
  )
  const levelDictionary = useRef<Promise<Map<string, string>> | null>(null)
  useEffect(() => {
    if (!enabled) {
      return
    }
    if (!levelDictionary.current) {
      levelDictionary.current = listAlertDictionaries({
        type: "ALERT",
        page: 1,
        pageSize: 100,
      }).then(
        (result) =>
          new Map((result.items ?? []).map((item) => [item.uid, item.name])),
        () => {
          // 字典取不到时退回显示级别 UID；允许下次切回分析 tab 时再试。
          levelDictionary.current = null
          return new Map<string, string>()
        }
      )
    }
    let active = true
    void levelDictionary.current.then((names) => {
      if (active) {
        setLevelNames(names)
      }
    })
    return () => {
      active = false
    }
  }, [enabled])

  const fingerprints = useMemo(() => {
    const data = analysisBlockData(blocks.fingerprints)
    if (!data) {
      return null
    }
    return data.items.map((item): FingerprintCount => ({
      fingerprint: item.fingerprint,
      ruleName: item.ruleName,
      count: item.events,
    }))
  }, [blocks.fingerprints])

  const recovery = analysisBlockData(blocks.recovery)
  const autoRecovery = durationStats(recovery?.auto)
  const manualRecovery = durationStats(recovery?.manual)

  const scopeSummary = describeScope(t, from, to, status)
  const hasFilters =
    Boolean(urlFrom || urlTo) || status !== "ALL" || granularity !== "day"

  function updateAnalysisFilter(key: "status" | "interval", value: string) {
    setSelectedBucket(null)
    const next = new URLSearchParams(params)
    if (value) next.set(key, value)
    else next.delete(key)
    setParams(next, { replace: true })
  }

  // 日期范围一次写入两个参数：选择器只提交完整区间，不会留下半个条件。
  function updateAnalysisRange(range: { from: string; to: string }) {
    setSelectedBucket(null)
    const next = new URLSearchParams(params)
    for (const key of ["from", "to"] as const) {
      const value = range[key]
      if (value) next.set(key, value)
      else next.delete(key)
    }
    setParams(next, { replace: true })
  }

  function clearAnalysisFilters() {
    setSelectedBucket(null)
    const next = new URLSearchParams(params)
    next.delete("from")
    next.delete("to")
    next.delete("status")
    next.delete("interval")
    setParams(next, { replace: true })
  }

  function clearFocusedBucket() {
    setSelectedBucket(null)
  }

  return (
    <section
      className="space-y-4"
      aria-labelledby="alert-history-insights-title"
      aria-busy={busy || undefined}
    >
      <span className="sr-only" role="status" aria-live="polite">
        {busy
          ? t("alertEvent.analysisLoading")
          : t("alertEvent.analysisLoaded")}
      </span>
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="min-w-0">
          <h2
            id="alert-history-insights-title"
            className="text-base font-semibold tracking-tight"
          >
            {t("alertEvent.analysisTitle")}
          </h2>
          <p className="mt-0.5 text-xs text-muted-foreground">
            {t("alertEvent.analysisTabHint")}
          </p>
        </div>
        <div
          role="group"
          className="flex flex-wrap items-center gap-2"
          aria-label={t("alertEvent.analysisFilters")}
        >
          <DateRangePicker
            from={from}
            to={to}
            maxDays={HISTORY_RANGE_MAX_DAYS}
            invalid={invalidDateRange}
            describedBy={
              invalidDateRange ? "alert-history-range-error" : undefined
            }
            ariaLabel={t("alertEvent.eventDateRange")}
            onChange={updateAnalysisRange}
            className="w-64"
          />
          <Select
            value={status}
            onValueChange={(value) =>
              updateAnalysisFilter("status", value === "ALL" ? "" : value)
            }
          >
            <SelectTrigger
              className="w-[10.5rem]"
              aria-label={t("alertEvent.status")}
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">{t("alertEvent.allStatuses")}</SelectItem>
              <SelectItem value="RECOVERED">
                {t("alertEvent.statusRecovered")}
              </SelectItem>
              <SelectItem value="RECOVERED_BY_MANUAL">
                {t("alertEvent.statusRecoveredManual")}
              </SelectItem>
              <SelectItem value="SUPPRESSING">
                {t("alertEvent.statusSuppressing")}
              </SelectItem>
            </SelectContent>
          </Select>
          <Select
            value={granularity}
            onValueChange={(value) =>
              updateAnalysisFilter("interval", value === "day" ? "" : value)
            }
          >
            <SelectTrigger
              className="w-28"
              aria-label={t("alertEvent.granularity")}
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="hour">
                {t("alertEvent.granularityHour")}
              </SelectItem>
              <SelectItem value="day">
                {t("alertEvent.granularityDay")}
              </SelectItem>
              <SelectItem value="week">
                {t("alertEvent.granularityWeek")}
              </SelectItem>
              <SelectItem value="month">
                {t("alertEvent.granularityMonth")}
              </SelectItem>
            </SelectContent>
          </Select>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="px-3"
            disabled={!hasFilters}
            onClick={clearAnalysisFilters}
          >
            {t("alertEvent.clearAnalysisFilters")}
          </Button>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{scopeSummary}</span>
        {defaultRange ? (
          <span>
            {t("alertEvent.defaultRangeHint", {
              count: HISTORY_RANGE_MAX_DAYS,
            })}
          </span>
        ) : null}
        <span>{t("alertEvent.localTimeZone", { timeZone })}</span>
        {invalidDateRange ? (
          <span
            id="alert-history-range-error"
            className="text-sm text-destructive"
            role="alert"
          >
            {t("alertEvent.invalidDateRange")}
          </span>
        ) : null}
        {focusedPoint ? (
          <span className="inline-flex items-center gap-2">
            <span>
              {t("alertEvent.focusedBucket", {
                label: focusedPoint.label,
                count: focusedPoint.count.toLocaleString(i18n.language),
              })}
            </span>
            <Button
              type="button"
              size="sm"
              variant="ghost"
              className="h-7 px-2"
              onClick={clearFocusedBucket}
            >
              {t("alertEvent.clearBucket")}
            </Button>
          </span>
        ) : null}
        <details className="basis-full">
          <summary className="cursor-pointer select-none hover:text-foreground">
            {t("alertEvent.analysisScopeDetails")}
          </summary>
          <p className="mt-1 max-w-4xl text-xs leading-relaxed">
            {t("alertEvent.analysisScopeNote")}
          </p>
        </details>
      </div>

      {degraded ? (
        <p
          className="rounded-lg border border-amber-500/30 bg-amber-500/5 px-4 py-2 text-xs text-amber-700 dark:text-amber-300"
          role="status"
        >
          {t("alertEvent.analysisArchiveDegraded")}
        </p>
      ) : null}

      {invalidDateRange ? (
        <p className="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          {t("alertEvent.analysisRangeBlocked")}
        </p>
      ) : empty ? (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-dashed border-border/80 bg-muted/20 px-4 py-10">
          <div>
            <p className="text-sm font-medium">
              {t("alertEvent.noAnalysisData")}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              {t("alertEvent.noAnalysisDataHint")}
            </p>
          </div>
          {hasFilters ? (
            <Button variant="outline" size="sm" onClick={clearAnalysisFilters}>
              {t("alertEvent.clearAnalysisFilters")}
            </Button>
          ) : null}
        </div>
      ) : (
        <>
          <div className="grid gap-px overflow-hidden rounded-xl border border-border/70 bg-border sm:grid-cols-2 xl:grid-cols-[1.35fr_repeat(3,minmax(0,1fr))]">
            <Metric
              featured
              label={t("alertEvent.eventsInScope")}
              status={blocks.summary.status}
              value={events.toLocaleString(i18n.language)}
              detail={
                focus
                  ? t("alertEvent.focusedRecordsMatching")
                  : t("alertEvent.recordsMatching")
              }
              onRetry={retrySummary(blocks.summary.status, retry)}
            />
            <Metric
              label={t("alertEvent.uniqueFingerprints")}
              status={blocks.summary.status}
              value={(summary?.uniqueFingerprints ?? 0).toLocaleString(
                i18n.language
              )}
              detail={t("alertEvent.averageEpisodes", {
                count: averageEpisodes(
                  events,
                  summary?.uniqueFingerprints ?? 0,
                  i18n.language
                ),
              })}
            />
            <Metric
              label={t("alertEvent.repeatedEvents")}
              status={blocks.summary.status}
              value={(summary?.repeated ?? 0).toLocaleString(i18n.language)}
              detail={t("alertEvent.repeatedEventsHelp")}
            />
            <Metric
              label={t("alertEvent.autoRecoveryMedian")}
              status={blocks.recovery.status}
              value={
                autoRecovery.median === null
                  ? "—"
                  : formatSpan(autoRecovery.median)
              }
              detail={t("alertEvent.durationSample", {
                count: autoRecovery.count.toLocaleString(i18n.language),
              })}
              onRetry={
                blocks.recovery.status === "error"
                  ? () => retry("recovery")
                  : undefined
              }
            />
          </div>

          <div className="grid gap-3 xl:grid-cols-[minmax(0,1.7fr)_minmax(17rem,1fr)]">
            <AnalysisPanel
              title={t("alertEvent.trendTitle")}
              description={t("alertEvent.startsAtTrend")}
              status={blocks.trend.status}
              onRetry={() => retry("trend")}
              skeleton={<ChartSkeleton />}
            >
              {timeline.points.length ? (
                <>
                  {timeline.gapsOmitted ? (
                    <p className="mb-3 text-xs text-muted-foreground">
                      {t("alertEvent.trendSparseNote")}
                    </p>
                  ) : null}
                  <div className="overflow-x-auto pb-1">
                    <div className="flex min-h-36 min-w-max items-end gap-1.5">
                      {timeline.points.map((point) => {
                        const height = point.count
                          ? Math.max(4, (point.count / timeline.maximum) * 100)
                          : 0
                        const selected = point.key === candidateBucketKey
                        return (
                          <button
                            key={point.key}
                            type="button"
                            aria-pressed={selected}
                            aria-label={t("alertEvent.selectTimeBucket", {
                              label: point.label,
                              count: point.count,
                            })}
                            className={`group flex min-h-36 w-11 shrink-0 flex-col items-center justify-end gap-2 rounded-md px-1 pb-1 transition-colors outline-none hover:bg-muted/60 focus-visible:ring-2 focus-visible:ring-ring motion-reduce:transition-none ${selected ? "bg-primary/10" : ""}`}
                            title={`${point.label}: ${point.count}`}
                            onClick={() =>
                              setSelectedBucket(
                                selected
                                  ? null
                                  : { key: point.key, filterSignature }
                              )
                            }
                          >
                            <span className="font-mono text-[10px] text-muted-foreground tabular-nums">
                              {point.count}
                            </span>
                            <span className="flex h-24 w-full items-end rounded-sm bg-muted/70">
                              <span
                                aria-hidden="true"
                                className={`w-full rounded-sm transition-[height,background-color] duration-150 motion-reduce:transition-none ${point.count ? (selected ? "bg-primary" : "bg-primary/70 group-hover:bg-primary") : "bg-transparent"}`}
                                style={{ height: `${height}%` }}
                              />
                            </span>
                            <span className="max-w-14 truncate text-center text-[10px] text-muted-foreground">
                              {point.label}
                            </span>
                          </button>
                        )
                      })}
                    </div>
                  </div>
                  <p className="mt-3 text-xs text-muted-foreground">
                    {t("alertEvent.selectBucketHelp")}
                  </p>
                </>
              ) : (
                <p className="py-12 text-center text-sm text-muted-foreground">
                  {t("alertEvent.noAnalysisData")}
                </p>
              )}
            </AnalysisPanel>

            <AnalysisPanel
              title={t("alertEvent.topRecurringFingerprints")}
              description={t("alertEvent.recurrenceHelp")}
              status={blocks.fingerprints.status}
              onRetry={() => retry("fingerprints")}
              skeleton={<RowsSkeleton rows={4} />}
            >
              <RecurringFingerprintList
                items={fingerprints ?? []}
                onSelect={onSelectFingerprint}
              />
            </AnalysisPanel>
          </div>

          <div className="grid gap-3 lg:grid-cols-3">
            <AnalysisPanel
              title={t("alertEvent.statusDistribution")}
              description={t("alertEvent.statusDistributionHelp")}
              aside={
                <span className="font-mono text-xs text-muted-foreground tabular-nums">
                  {statusRows ? statusTotal.toLocaleString() : ""}
                </span>
              }
              status={blocks.status.status}
              onRetry={() => retry("status")}
              skeleton={<RowsSkeleton rows={3} />}
            >
              <DistributionRows
                data={statusRows ?? new Map()}
                labelFor={(value) => statusLabel(value, t)}
                total={statusTotal}
                selectedValue={
                  status === "ALL"
                    ? ""
                    : status === "SUPPRESSING"
                      ? "FIRING"
                      : status
                }
                onSelect={(value) => {
                  const next = toStatusFilter(value)
                  updateAnalysisFilter(
                    "status",
                    next === status ? "" : next === "ALL" ? "" : next
                  )
                }}
              />
            </AnalysisPanel>
            <AnalysisPanel
              title={t("alertEvent.levelDistribution")}
              description={t("alertEvent.levelDistributionHelp")}
              aside={
                <span className="font-mono text-xs text-muted-foreground tabular-nums">
                  {levelRows ? levelTotal.toLocaleString() : ""}
                </span>
              }
              status={blocks.level.status}
              onRetry={() => retry("level")}
              skeleton={<RowsSkeleton rows={3} />}
            >
              <DistributionRows
                data={levelRows ?? new Map()}
                labelFor={(value) =>
                  levelNames.get(value) ||
                  value ||
                  t("alertEvent.unsetLevel")
                }
                total={levelTotal}
              />
            </AnalysisPanel>
            <AnalysisPanel
              title={t("alertEvent.recoveryDuration")}
              description={t("alertEvent.recoveryDurationHelp")}
              status={blocks.recovery.status}
              onRetry={() => retry("recovery")}
              skeleton={<RowsSkeleton rows={3} />}
            >
              <RecoveryDurations
                autoRecovery={autoRecovery}
                manualRecovery={manualRecovery}
              />
            </AnalysisPanel>
          </div>
        </>
      )}
    </section>
  )
}

function retrySummary(
  status: AnalysisBlockStatus,
  retry: (block: "summary") => void
) {
  return status === "error" ? () => retry("summary") : undefined
}

function durationStats(stats: AlertAnalysisRecoveryStats | undefined) {
  if (!stats || stats.samples <= 0) {
    return { median: null, p90: null, count: 0 }
  }
  return {
    median: stats.medianSeconds * 1000,
    p90: stats.p90Seconds * 1000,
    count: stats.samples,
  }
}

function AnalysisPanel({
  title,
  description,
  aside,
  status,
  onRetry,
  skeleton,
  children,
}: {
  title: string
  description?: string
  aside?: ReactNode
  status: AnalysisBlockStatus
  onRetry: () => void
  skeleton: ReactNode
  children: ReactNode
}) {
  const { t } = useTranslation()
  const pending = status === "idle" || status === "loading"
  return (
    <section
      className="min-w-0 rounded-xl border border-border/70 bg-card p-4"
      aria-label={title}
      aria-busy={pending || undefined}
    >
      <div className="mb-4">
        <div className="flex items-baseline justify-between gap-2">
          <h3 className="text-sm font-semibold">{title}</h3>
          {aside}
        </div>
        {description ? (
          <p className="mt-1 text-xs text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {pending ? (
        skeleton
      ) : status === "error" ? (
        <div
          role="alert"
          className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2"
        >
          <p className="text-xs text-destructive">
            {t("alertEvent.analysisLoadFailed")}
          </p>
          <Button variant="outline" size="sm" onClick={onRetry}>
            {t("action.retry")}
          </Button>
        </div>
      ) : (
        children
      )}
    </section>
  )
}

/** 块内骨架：只占该卡片自己的空间，不移动其他块。 */
function RowsSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="space-y-2" aria-hidden="true">
      {Array.from({ length: rows }, (_, index) => (
        <div key={index} className="space-y-1.5">
          <div className="h-3 w-2/3 animate-pulse rounded bg-muted" />
          <div className="h-1.5 w-full animate-pulse rounded-full bg-muted" />
        </div>
      ))}
    </div>
  )
}

function ChartSkeleton() {
  return (
    <div className="space-y-3" aria-hidden="true">
      <div className="h-3 w-40 animate-pulse rounded bg-muted" />
      <div className="flex h-36 items-end gap-1.5">
        {[38, 62, 44, 78, 52, 88, 66, 72, 48, 58].map((height, index) => (
          <div
            key={index}
            className="w-8 animate-pulse rounded-sm bg-muted"
            style={{ height: `${height}%` }}
          />
        ))}
      </div>
    </div>
  )
}

function Metric({
  label,
  value,
  detail,
  featured = false,
  status,
  onRetry,
}: {
  label: string
  value: string
  detail: string
  featured?: boolean
  status: AnalysisBlockStatus
  onRetry?: () => void
}) {
  const { t } = useTranslation()
  const pending = status === "idle" || status === "loading"
  const failed = status === "error"
  return (
    <div
      className={`min-w-0 bg-card px-4 py-3 ${featured ? "sm:col-span-2 xl:col-span-1" : ""}`}
      aria-busy={pending || undefined}
    >
      <p className="text-xs font-medium text-muted-foreground">{label}</p>
      {pending ? (
        <div aria-hidden="true">
          <div className="mt-1 h-8 w-24 animate-pulse rounded bg-muted" />
          <div className="mt-2 h-3 w-32 animate-pulse rounded bg-muted" />
        </div>
      ) : failed ? (
        <>
          <p className="mt-1 text-2xl font-semibold tracking-tight">—</p>
          <p className="mt-1 flex items-center gap-2 text-xs text-destructive">
            <span>{t("alertEvent.metricLoadFailed")}</span>
            {onRetry ? (
              <Button
                variant="ghost"
                size="sm"
                className="h-6 px-2 text-xs"
                onClick={onRetry}
              >
                {t("action.retry")}
              </Button>
            ) : null}
          </p>
        </>
      ) : (
        <>
          <p
            className={`mt-1 font-semibold tracking-tight tabular-nums ${featured ? "text-3xl" : "text-2xl"}`}
          >
            {value}
          </p>
          <p
            className="mt-1 truncate text-xs text-muted-foreground"
            title={detail}
          >
            {detail}
          </p>
        </>
      )}
    </div>
  )
}

function RecurringFingerprintList({
  items,
  onSelect,
}: {
  items: FingerprintCount[]
  onSelect: (fingerprint: string) => void
}) {
  const { t, i18n } = useTranslation()
  if (!items.length) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        {t("alertEvent.noRecurringFingerprints")}
      </p>
    )
  }
  return (
    <ol className="space-y-1">
      {items.map((item, index) => (
        <li key={item.fingerprint}>
          <button
            type="button"
            className="flex min-h-11 w-full items-center gap-3 rounded-md px-2 text-left transition-colors outline-none hover:bg-muted/60 focus-visible:ring-2 focus-visible:ring-ring motion-reduce:transition-none"
            title={`${item.fingerprint} · ${t("alertEvent.viewFingerprintEvents")}`}
            onClick={() => onSelect(item.fingerprint)}
          >
            <span className="w-5 shrink-0 font-mono text-xs text-muted-foreground tabular-nums">
              {String(index + 1).padStart(2, "0")}
            </span>
            <span className="min-w-0 flex-1">
              <span className="block truncate font-mono text-xs">
                {item.fingerprint}
              </span>
              {item.ruleName ? (
                <span className="mt-0.5 block truncate text-xs text-muted-foreground">
                  {item.ruleName}
                </span>
              ) : null}
            </span>
            <span className="shrink-0 rounded bg-muted px-2 py-1 font-mono text-xs tabular-nums">
              {t("alertEvent.episodeCount", {
                count: item.count.toLocaleString(i18n.language),
              })}
            </span>
          </button>
        </li>
      ))}
    </ol>
  )
}

function RecoveryDurations({
  autoRecovery,
  manualRecovery,
}: {
  autoRecovery: DurationStats
  manualRecovery: DurationStats
}) {
  const { t } = useTranslation()
  return (
    <>
      <DurationRow
        label={t("alertEvent.statusRecovered")}
        stats={autoRecovery}
        showP90
      />
      <div className="my-3 border-t border-border/60" />
      <DurationRow
        label={t("alertEvent.statusRecoveredManual")}
        stats={manualRecovery}
      />
    </>
  )
}

function DurationRow({
  label,
  stats,
  showP90 = false,
}: {
  label: string
  stats: DurationStats
  showP90?: boolean
}) {
  const { t, i18n } = useTranslation()
  return (
    <div className="grid gap-2">
      <div className="flex items-baseline justify-between gap-2">
        <span className="text-xs font-medium">{label}</span>
        <span className="font-mono text-[10px] text-muted-foreground tabular-nums">
          {t("alertEvent.durationSample", {
            count: stats.count.toLocaleString(i18n.language),
          })}
        </span>
      </div>
      {stats.median === null ? (
        <p className="text-sm text-muted-foreground">
          {t("alertEvent.noDurationData")}
        </p>
      ) : (
        <div className="flex flex-wrap items-baseline gap-x-4 gap-y-1">
          <p className="text-sm font-semibold tabular-nums">
            {t("alertEvent.medianLabel")}{" "}
            <span className="font-mono">{formatSpan(stats.median)}</span>
          </p>
          {showP90 && stats.p90 !== null ? (
            <p className="text-xs text-muted-foreground tabular-nums">
              {t("alertEvent.p90Label")}{" "}
              <span className="font-mono">{formatSpan(stats.p90)}</span>
            </p>
          ) : null}
        </div>
      )}
    </div>
  )
}

function DistributionRows({
  data,
  labelFor,
  total,
  selectedValue = "",
  onSelect,
}: {
  data: Map<string, number>
  labelFor: (value: string) => string
  total: number
  selectedValue?: string
  onSelect?: (value: string) => void
}) {
  const { t, i18n } = useTranslation()
  const rows = [...data.entries()].sort((a, b) => b[1] - a[1])
  const max = Math.max(1, ...rows.map(([, count]) => count))
  if (!rows.length) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        {t("alertEvent.noAnalysisData")}
      </p>
    )
  }
  return (
    <ul className="space-y-2">
      {rows.map(([value, count]) => {
        const content = (
          <>
            <span className="mb-1 flex items-center justify-between gap-2 text-xs">
              <span className="min-w-0 truncate" title={labelFor(value)}>
                {labelFor(value)}
              </span>
              <span className="shrink-0 font-mono text-muted-foreground tabular-nums">
                {count.toLocaleString(i18n.language)} ·{" "}
                {total ? Math.round((count / total) * 100) : 0}%
              </span>
            </span>
            <span className="block h-1.5 overflow-hidden rounded-full bg-muted">
              <span
                className="block h-full rounded-full bg-primary/70"
                style={{ width: `${(count / max) * 100}%` }}
              />
            </span>
          </>
        )
        return (
          <li key={value}>
            {onSelect ? (
              <button
                type="button"
                aria-pressed={selectedValue === value}
                className={`w-full rounded-sm p-1 text-left transition-colors outline-none hover:bg-muted/50 focus-visible:ring-2 focus-visible:ring-ring ${selectedValue === value ? "bg-muted/70" : ""}`}
                onClick={() => onSelect(value)}
              >
                {content}
              </button>
            ) : (
              <div className="p-1">{content}</div>
            )}
          </li>
        )
      })}
    </ul>
  )
}

function toStatusFilter(value: string): StatusFilter {
  if (value === "RECOVERED" || value === "RECOVERED_BY_MANUAL") {
    return value
  }
  if (value === "FIRING" || value === "SUPPRESSED" || value === "SUPPRESSING") {
    return "SUPPRESSING"
  }
  return "ALL"
}

function describeScope(
  t: TFunction,
  from: string,
  to: string,
  status: StatusFilter
) {
  // 区间永远有效（URL 缺省时由最近 30 天兜底），所以摘要始终能给出明确范围。
  const range = t("alertEvent.scopeRange", { from, to })
  const label =
    status === "ALL"
      ? null
      : t("alertEvent.scopeStatus", {
          status:
            status === "SUPPRESSING"
              ? t("alertEvent.statusSuppressing")
              : status === "RECOVERED"
                ? t("alertEvent.statusRecovered")
                : t("alertEvent.statusRecoveredManual"),
        })
  return [range, label].filter(Boolean).join(" · ")
}

function averageEpisodes(
  events: number,
  fingerprints: number,
  language: string
) {
  if (fingerprints === 0) {
    return (0).toLocaleString(language)
  }
  return (events / fingerprints).toLocaleString(language, {
    maximumFractionDigits: 1,
  })
}

function parseStatusFilter(value: string | null): StatusFilter {
  // SUPPRESSED 是旧链接里的写法，含义是“仍在触发且被抑制”。
  if (value === "SUPPRESSED" || value === "SUPPRESSING") {
    return "SUPPRESSING"
  }
  if (value === "RECOVERED" || value === "RECOVERED_BY_MANUAL") {
    return value
  }
  return "ALL"
}

function parseGranularity(value: string | null): Granularity {
  if (value === "hour" || value === "week" || value === "month") return value
  return "day"
}

function statusLabel(value: string, t: (key: string) => string) {
  if (value === "FIRING" || value === "SUPPRESSED" || value === "SUPPRESSING") {
    return t("alertEvent.statusSuppressing")
  }
  if (value === "RECOVERED") return t("alertEvent.statusRecovered")
  if (value === "RECOVERED_BY_MANUAL")
    return t("alertEvent.statusRecoveredManual")
  return value || t("alertEvent.unsetLevel")
}

interface TimelinePoint {
  key: number
  count: number
  label: string
}

/**
 * 趋势图的时间轴：输入是服务端返回的非空桶（start 为本地时区桶起，RFC3339）。
 * 仍然在前端做零填充与 MAX_BUCKETS 退化，标签/占位口径与改造前一致。
 */
function buildTimeline(
  buckets: AlertAnalysisTrendBucket[],
  granularity: Granularity,
  language: string,
  from: string,
  to: string
): { points: TimelinePoint[]; gapsOmitted: boolean; maximum: number } {
  const bins = new Map<number, number>()
  for (const bucket of buckets) {
    const date = new Date(bucket.start)
    if (!Number.isFinite(date.getTime())) continue
    const key = bucketStart(date, granularity).getTime()
    bins.set(key, (bins.get(key) ?? 0) + bucket.events)
  }
  if (bins.size === 0) return { points: [], gapsOmitted: false, maximum: 1 }

  const fromWindow = localDayWindow(from)
  const toWindow = localDayWindow(to)
  if (!fromWindow || !toWindow) {
    return { points: [], gapsOmitted: false, maximum: 1 }
  }
  const rangeStart = new Date(fromWindow.start)
  // 上界是下一本地午夜（排除），所以用 end - 1ms 落在最后一个自然日之内。
  const rangeEnd = new Date(toWindow.end - 1)
  if (
    !Number.isFinite(rangeStart.getTime()) ||
    !Number.isFinite(rangeEnd.getTime()) ||
    rangeStart > rangeEnd
  ) {
    return { points: [], gapsOmitted: false, maximum: 1 }
  }

  const first = bucketStart(rangeStart, granularity)
  const last = bucketStart(rangeEnd, granularity)
  const keys: number[] = []
  let cursor = first
  while (cursor <= last && keys.length < MAX_BUCKETS) {
    keys.push(cursor.getTime())
    cursor = advanceBucket(cursor, granularity)
  }
  const gapsOmitted = cursor <= last
  const timelineKeys = gapsOmitted
    ? [...bins.keys()].sort((a, b) => a - b)
    : keys
  const locale = language.startsWith("zh") ? "zh-CN" : "en-US"
  const options: Intl.DateTimeFormatOptions =
    granularity === "hour"
      ? { month: "short", day: "numeric", hour: "2-digit" }
      : granularity === "month"
        ? { year: "numeric", month: "short" }
        : { month: "short", day: "numeric" }
  const points = timelineKeys.map((key) => ({
    key,
    count: bins.get(key) ?? 0,
    label: new Intl.DateTimeFormat(locale, options).format(new Date(key)),
  }))
  return {
    points,
    gapsOmitted,
    maximum: Math.max(1, ...points.map((point) => point.count)),
  }
}

function bucketStart(date: Date, granularity: Granularity) {
  const start = new Date(date)
  if (granularity === "hour") start.setMinutes(0, 0, 0)
  if (granularity === "day") start.setHours(0, 0, 0, 0)
  if (granularity === "week") {
    start.setHours(0, 0, 0, 0)
    start.setDate(start.getDate() - ((start.getDay() + 6) % 7))
  }
  if (granularity === "month") {
    start.setDate(1)
    start.setHours(0, 0, 0, 0)
  }
  return start
}

function advanceBucket(date: Date, granularity: Granularity) {
  const next = new Date(date)
  if (granularity === "hour") next.setHours(next.getHours() + 1)
  if (granularity === "day") next.setDate(next.getDate() + 1)
  if (granularity === "week") next.setDate(next.getDate() + 7)
  if (granularity === "month") next.setMonth(next.getMonth() + 1)
  return next
}
