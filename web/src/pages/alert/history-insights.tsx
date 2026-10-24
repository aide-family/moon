import { useEffect, useMemo, useRef, useState } from "react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"

import { listAlertEvents } from "@/api/alert-event"
import type {
  AlertEventItem,
  ListAlertEventsParams,
} from "@/api/alert-event/types"
import { Button } from "@/components/ui/button"
import { DateRangePicker } from "@/components/date-range-picker"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { dayEndISO, dayStartISO } from "@/lib/date-range"
import { formatSpan } from "@/pages/alert/format"
import {
  HISTORY_RANGE_MAX_DAYS,
  localDayWindow,
} from "@/pages/alert/history-range"

// 接口单页上限 100；分析在浏览器侧聚合，必须逐页读完整集合。
const FETCH_PAGE_SIZE = 100
// 聚合上限：50 页 = 5000 条。超过时明确提示统计被截断，而不是悄悄少算。
const MAX_FETCH_PAGES = 50
// 时间轴最多渲染 180 个时间桶；超出时退化为“只显示有事件的时间桶”。
const MAX_BUCKETS = 180
const EMPTY_EVENTS: AlertEventItem[] = []
const EMPTY_PROGRESS = { queryKey: "", loaded: 0, total: 0 }

type Granularity = "hour" | "day" | "week" | "month"
// 历史视图的事件终态只有三种：自动恢复、人工恢复、仍在触发（且被抑制）。
// 抑制是触发阶段的属性，不是终态，因此不再作为状态选项。
type StatusFilter = "ALL" | "RECOVERED" | "RECOVERED_BY_MANUAL" | "SUPPRESSING"

interface AlertHistoryInsightsProps {
  enabled: boolean
  onSelectFingerprint: (fingerprint: string) => void
}

interface AnalysisResult {
  queryKey: string
  events: AlertEventItem[]
  truncated: boolean
  error: string
}

export function AlertHistoryInsights({
  enabled,
  onSelectFingerprint,
}: AlertHistoryInsightsProps) {
  const { t, i18n } = useTranslation()
  const [params, setParams] = useSearchParams()
  const from = params.get("from") || ""
  const to = params.get("to") || ""
  const status = parseStatusFilter(params.get("status"))
  const granularity = parseGranularity(params.get("interval"))
  const invalidDateRange = Boolean(from && to && from > to)
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone

  const [attempt, setAttempt] = useState(0)
  const queryKey = `${from}|${to}|${status}|${attempt}`
  const [result, setResult] = useState<AnalysisResult | null>(null)
  const [progress, setProgress] = useState(EMPTY_PROGRESS)
  const requestId = useRef(0)
  const loadedKey = useRef("")

  const loading = result?.queryKey !== queryKey
  const error = loading ? "" : (result?.error ?? "")
  const events = loading ? EMPTY_EVENTS : (result?.events ?? EMPTY_EVENTS)
  const truncated = !loading && (result?.truncated ?? false)

  const filterSignature = `${from}|${to}|${status}|${granularity}`
  const [selectedBucket, setSelectedBucket] = useState<{
    key: number
    filterSignature: string
  } | null>(null)
  const selectedBucketKey =
    selectedBucket?.filterSignature === filterSignature
      ? selectedBucket.key
      : null

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

  useEffect(() => {
    if (!enabled || invalidDateRange) {
      return
    }
    if (loadedKey.current === queryKey) {
      return
    }
    loadedKey.current = queryKey
    const id = ++requestId.current
    let active = true

    // 统计范围直接下推到列表接口：状态、时间范围在服务端合并归档记录后过滤，
    // 前端只对已受控的结果做聚合，不再“先取全量再自己筛”。
    const scoped: ListAlertEventsParams = {
      view: "HISTORY",
      pageSize: FETCH_PAGE_SIZE,
      startsAtFrom: dayStartISO(from),
      startsAtTo: dayEndISO(to),
      ...statusScope(status),
    }

    async function loadAllHistory() {
      try {
        const collected = new Map<string, AlertEventItem>()
        let total = 0
        let pages = 1
        for (let page = 1; page <= MAX_FETCH_PAGES; page += 1) {
          const data = await listAlertEvents({ ...scoped, page })
          total = data.total ?? collected.size
          for (const item of data.items ?? []) {
            // 归档合并列表在分页期间可能有新记录写入，按 UID 去重避免重复计数。
            collected.set(item.uid, item)
          }
          if (active) {
            setProgress({ queryKey, loaded: collected.size, total })
          }
          pages = Math.max(1, Math.ceil(total / FETCH_PAGE_SIZE))
          if (page >= pages) {
            break
          }
        }
        if (!active || requestId.current !== id) {
          return
        }
        setResult({
          queryKey,
          events: [...collected.values()],
          truncated: pages > MAX_FETCH_PAGES,
          error: "",
        })
      } catch {
        if (!active || requestId.current !== id) {
          return
        }
        setResult({
          queryKey,
          events: [],
          truncated: false,
          error: t("alertEvent.analysisLoadFailed"),
        })
      }
    }

    void loadAllHistory()
    return () => {
      active = false
    }
  }, [enabled, invalidDateRange, queryKey, from, to, status, t])

  // 时间桶聚焦只影响统计口径，不改写明细列表的查询条件。
  const timeline = useMemo(
    () => buildTimeline(events, granularity, i18n.language, from, to),
    [events, granularity, i18n.language, from, to]
  )
  const focusedBucket =
    selectedBucketKey === null
      ? null
      : (timeline.points.find((point) => point.key === selectedBucketKey) ??
        null)
  const scopedEvents = useMemo(
    () =>
      focusedBucket
        ? events.filter(
            (event) =>
              bucketStart(new Date(event.startsAt), granularity).getTime() ===
              focusedBucket.key
          )
        : events,
    [events, focusedBucket, granularity]
  )

  const statusRows = useMemo(
    () => countBy(scopedEvents, (event) => event.status),
    [scopedEvents]
  )
  const levelRows = useMemo(
    () =>
      countBy(
        scopedEvents,
        (event) =>
          event.levelName || event.levelUid || t("alertEvent.unsetLevel")
      ),
    [scopedEvents, t]
  )
  const recurringFingerprints = useMemo(
    () => buildRecurringFingerprints(scopedEvents),
    [scopedEvents]
  )
  const repeatedEpisodes = recurringFingerprints.reduce(
    (count, item) => count + item.count - 1,
    0
  )
  const { uniqueFingerprints, autoRecovery, manualRecovery } = useMemo(() => {
    const fingerprints = new Set(
      scopedEvents.map((event) => event.fingerprint).filter(Boolean)
    )
    return {
      uniqueFingerprints: fingerprints.size,
      autoRecovery: durationSummary(
        recoveryDurations(scopedEvents, "RECOVERED")
      ),
      manualRecovery: durationSummary(
        recoveryDurations(scopedEvents, "RECOVERED_BY_MANUAL")
      ),
    }
  }, [scopedEvents])

  const scopeSummary = describeScope(t, from, to, status)
  const hasFilters =
    Boolean(from || to) || status !== "ALL" || granularity !== "day"

  return (
    <section
      className="space-y-4"
      aria-labelledby="alert-history-insights-title"
      aria-busy={loading || undefined}
    >
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
        {focusedBucket ? (
          <span className="inline-flex items-center gap-2">
            <span>
              {t("alertEvent.focusedBucket", {
                label: focusedBucket.label,
                count: focusedBucket.count.toLocaleString(i18n.language),
              })}
            </span>
            <Button
              type="button"
              size="sm"
              variant="ghost"
              className="h-7 px-2"
              onClick={() => setSelectedBucket(null)}
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

      {invalidDateRange ? (
        <p className="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          {t("alertEvent.analysisRangeBlocked")}
        </p>
      ) : loading ? (
        <AnalysisSkeleton
          label={t("alertEvent.analysisLoading", {
            loaded: progress.queryKey === queryKey ? progress.loaded : 0,
            total: progress.queryKey === queryKey ? progress.total : 0,
          })}
        />
      ) : error ? (
        <div
          className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3"
          role="alert"
        >
          <p className="text-sm text-destructive">{error}</p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAttempt((value) => value + 1)}
          >
            {t("action.retry")}
          </Button>
        </div>
      ) : events.length === 0 ? (
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
          {truncated ? (
            <p
              className="rounded-lg border border-amber-500/30 bg-amber-500/5 px-4 py-2 text-xs text-amber-700 dark:text-amber-300"
              role="status"
            >
              {t("alertEvent.analysisTruncated", {
                count: (MAX_FETCH_PAGES * FETCH_PAGE_SIZE).toLocaleString(
                  i18n.language
                ),
              })}
            </p>
          ) : null}
          <div className="grid gap-px overflow-hidden rounded-xl border border-border/70 bg-border sm:grid-cols-2 xl:grid-cols-[1.35fr_repeat(3,minmax(0,1fr))]">
            <Metric
              featured
              label={t("alertEvent.eventsInScope")}
              value={scopedEvents.length.toLocaleString(i18n.language)}
              detail={
                focusedBucket
                  ? t("alertEvent.focusedRecordsMatching")
                  : t("alertEvent.recordsMatching")
              }
            />
            <Metric
              label={t("alertEvent.uniqueFingerprints")}
              value={uniqueFingerprints.toLocaleString(i18n.language)}
              detail={t("alertEvent.averageEpisodes", {
                count: averageEpisodes(
                  scopedEvents.length,
                  uniqueFingerprints,
                  i18n.language
                ),
              })}
            />
            <Metric
              label={t("alertEvent.repeatedEvents")}
              value={repeatedEpisodes.toLocaleString(i18n.language)}
              detail={t("alertEvent.recurringFingerprintCount", {
                count: recurringFingerprints.length.toLocaleString(
                  i18n.language
                ),
              })}
            />
            <Metric
              label={t("alertEvent.autoRecoveryMedian")}
              value={
                autoRecovery.median === null
                  ? "—"
                  : formatSpan(autoRecovery.median)
              }
              detail={t("alertEvent.durationSample", {
                count: autoRecovery.count.toLocaleString(i18n.language),
              })}
            />
          </div>

          <div className="grid gap-3 xl:grid-cols-[minmax(0,1.7fr)_minmax(17rem,1fr)]">
            <section
              className="min-w-0 rounded-xl border border-border/70 bg-card p-4"
              aria-label={t("alertEvent.trendTitle")}
            >
              <div className="mb-4">
                <h3 className="text-sm font-semibold">
                  {t("alertEvent.trendTitle")}
                </h3>
                <p className="mt-1 text-xs text-muted-foreground">
                  {t("alertEvent.startsAtTrend")}
                </p>
              </div>
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
                        const selected = point.key === selectedBucketKey
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
            </section>

            <RecurringFingerprints
              items={recurringFingerprints.slice(0, 5)}
              onSelect={onSelectFingerprint}
            />
          </div>

          <div className="grid gap-3 lg:grid-cols-3">
            <Distribution
              title={t("alertEvent.statusDistribution")}
              description={t("alertEvent.statusDistributionHelp")}
              data={statusRows}
              labelFor={(value) => statusLabel(value, t)}
              total={scopedEvents.length}
              selectedValue={
                status === "ALL"
                  ? ""
                  : status === "SUPPRESSING"
                    ? "FIRING"
                    : status
              }
              onSelect={(value) =>
                updateAnalysisFilter(
                  "status",
                  toStatusFilter(value) === status ? "" : toStatusFilter(value)
                )
              }
            />
            <Distribution
              title={t("alertEvent.levelDistribution")}
              data={levelRows}
              labelFor={(value) => value}
              total={scopedEvents.length}
            />
            <RecoverySummary
              autoRecovery={autoRecovery}
              manualRecovery={manualRecovery}
            />
          </div>
        </>
      )}
    </section>
  )
}

/** 把分析用的状态值翻译成接口过滤条件；历史里 FIRING 必然是被抑制的事件。 */
function statusScope(
  status: StatusFilter
): Pick<ListAlertEventsParams, "eventStatus" | "suppression"> {
  if (status === "RECOVERED" || status === "RECOVERED_BY_MANUAL") {
    return { eventStatus: status }
  }
  if (status === "SUPPRESSING") {
    return { eventStatus: "FIRING", suppression: "SUPPRESSED" }
  }
  return {}
}

function toStatusFilter(value: string): StatusFilter {
  if (value === "RECOVERED" || value === "RECOVERED_BY_MANUAL") {
    return value
  }
  if (value === "FIRING" || value === "SUPPRESSED") {
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
  const range =
    from || to
      ? t("alertEvent.scopeRange", { from: from || "…", to: to || "…" })
      : null
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
  const parts = [range, label].filter(Boolean)
  return parts.length ? parts.join(" · ") : t("alertEvent.scopeAll")
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

function AnalysisSkeleton({ label }: { label: string }) {
  return (
    <div className="space-y-3" role="status" aria-live="polite">
      <p className="text-xs text-muted-foreground">{label}</p>
      <div className="grid gap-px overflow-hidden rounded-xl border border-border/70 bg-border sm:grid-cols-2 xl:grid-cols-4">
        {[0, 1, 2, 3].map((index) => (
          <div key={index} className="bg-card px-4 py-3">
            <div className="h-3 w-24 animate-pulse rounded bg-muted" />
            <div className="mt-2 h-7 w-16 animate-pulse rounded bg-muted" />
            <div className="mt-2 h-3 w-32 animate-pulse rounded bg-muted" />
          </div>
        ))}
      </div>
      <div className="grid gap-3 xl:grid-cols-[minmax(0,1.7fr)_minmax(17rem,1fr)]">
        <div className="h-48 animate-pulse rounded-xl border border-border/70 bg-card" />
        <div className="h-48 animate-pulse rounded-xl border border-border/70 bg-card" />
      </div>
    </div>
  )
}

function Metric({
  label,
  value,
  detail,
  featured = false,
}: {
  label: string
  value: string
  detail: string
  featured?: boolean
}) {
  return (
    <div
      className={`min-w-0 bg-card px-4 py-3 ${featured ? "sm:col-span-2 xl:col-span-1" : ""}`}
    >
      <p className="text-xs font-medium text-muted-foreground">{label}</p>
      <p
        className={`mt-1 font-semibold tracking-tight tabular-nums ${featured ? "text-3xl" : "text-2xl"}`}
      >
        {value}
      </p>
      <p className="mt-1 truncate text-xs text-muted-foreground" title={detail}>
        {detail}
      </p>
    </div>
  )
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

function RecurringFingerprints({
  items,
  onSelect,
}: {
  items: FingerprintCount[]
  onSelect: (fingerprint: string) => void
}) {
  const { t, i18n } = useTranslation()
  return (
    <section
      className="min-w-0 rounded-xl border border-border/70 bg-card p-4"
      aria-label={t("alertEvent.topRecurringFingerprints")}
    >
      <div className="mb-4">
        <h3 className="text-sm font-semibold">
          {t("alertEvent.topRecurringFingerprints")}
        </h3>
        <p className="mt-1 text-xs text-muted-foreground">
          {t("alertEvent.recurrenceHelp")}
        </p>
      </div>
      {items.length ? (
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
      ) : (
        <p className="py-8 text-center text-sm text-muted-foreground">
          {t("alertEvent.noRecurringFingerprints")}
        </p>
      )}
    </section>
  )
}

function RecoverySummary({
  autoRecovery,
  manualRecovery,
}: {
  autoRecovery: DurationStats
  manualRecovery: DurationStats
}) {
  const { t } = useTranslation()
  return (
    <section
      className="min-w-0 rounded-xl border border-border/70 bg-card p-4"
      aria-label={t("alertEvent.recoveryDuration")}
    >
      <div className="mb-4">
        <h3 className="text-sm font-semibold">
          {t("alertEvent.recoveryDuration")}
        </h3>
        <p className="mt-1 text-xs text-muted-foreground">
          {t("alertEvent.recoveryDurationHelp")}
        </p>
      </div>
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
    </section>
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

function Distribution({
  title,
  description,
  data,
  labelFor,
  total,
  selectedValue = "",
  onSelect,
}: {
  title: string
  description?: string
  data: Map<string, number>
  labelFor: (value: string) => string
  total: number
  selectedValue?: string
  onSelect?: (value: string) => void
}) {
  const rows = [...data.entries()].sort((a, b) => b[1] - a[1])
  const max = Math.max(1, ...rows.map(([, count]) => count))
  return (
    <section
      className="min-w-0 rounded-xl border border-border/70 bg-card p-4"
      aria-label={title}
    >
      <div className="mb-4">
        <div className="flex items-baseline justify-between gap-2">
          <h3 className="text-sm font-semibold">{title}</h3>
          <span className="font-mono text-xs text-muted-foreground tabular-nums">
            {total.toLocaleString()}
          </span>
        </div>
        {description ? (
          <p className="mt-1 text-xs text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {rows.length ? (
        <ul className="space-y-2">
          {rows.map(([value, count]) => {
            const content = (
              <>
                <span className="mb-1 flex items-center justify-between gap-2 text-xs">
                  <span className="min-w-0 truncate" title={labelFor(value)}>
                    {labelFor(value)}
                  </span>
                  <span className="shrink-0 font-mono text-muted-foreground tabular-nums">
                    {count.toLocaleString()} ·{" "}
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
      ) : (
        <p className="py-8 text-center text-sm text-muted-foreground">—</p>
      )}
    </section>
  )
}

function countBy(
  items: AlertEventItem[],
  key: (item: AlertEventItem) => string
) {
  const counts = new Map<string, number>()
  for (const item of items) {
    const value = key(item) || "UNKNOWN"
    counts.set(value, (counts.get(value) ?? 0) + 1)
  }
  return counts
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
  if (value === "FIRING") return t("alertEvent.statusSuppressing")
  if (value === "RECOVERED") return t("alertEvent.statusRecovered")
  if (value === "RECOVERED_BY_MANUAL")
    return t("alertEvent.statusRecoveredManual")
  return value
}

interface TimelinePoint {
  key: number
  count: number
  label: string
}

function buildTimeline(
  items: AlertEventItem[],
  granularity: Granularity,
  language: string,
  from: string,
  to: string
): { points: TimelinePoint[]; gapsOmitted: boolean; maximum: number } {
  const bins = new Map<number, number>()
  const dates: Date[] = []
  for (const item of items) {
    const date = new Date(item.startsAt)
    if (!Number.isFinite(date.getTime())) continue
    dates.push(date)
    const key = bucketStart(date, granularity).getTime()
    bins.set(key, (bins.get(key) ?? 0) + 1)
  }
  if (bins.size === 0) return { points: [], gapsOmitted: false, maximum: 1 }

  const fromWindow = from ? localDayWindow(from) : undefined
  const toWindow = to ? localDayWindow(to) : undefined
  const rangeStart = fromWindow
    ? new Date(fromWindow.start)
    : dates.reduce(
        (earliest, date) => (date < earliest ? date : earliest),
        dates[0]
      )
  // 上界是下一本地午夜（排除），所以用 end - 1ms 落在最后一个自然日之内。
  const rangeEnd = toWindow
    ? new Date(toWindow.end - 1)
    : dates.reduce((latest, date) => (date > latest ? date : latest), dates[0])
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

function buildRecurringFingerprints(
  items: AlertEventItem[]
): FingerprintCount[] {
  const counts = new Map<string, FingerprintCount>()
  for (const event of items) {
    if (!event.fingerprint) continue
    const existing = counts.get(event.fingerprint)
    if (existing) {
      existing.count += 1
      if (!existing.ruleName && event.ruleName)
        existing.ruleName = event.ruleName
    } else {
      counts.set(event.fingerprint, {
        fingerprint: event.fingerprint,
        ruleName: event.ruleName,
        count: 1,
      })
    }
  }
  return [...counts.values()]
    .filter((item) => item.count > 1)
    .sort(
      (a, b) => b.count - a.count || a.fingerprint.localeCompare(b.fingerprint)
    )
}

function recoveryDurations(
  items: AlertEventItem[],
  status: "RECOVERED" | "RECOVERED_BY_MANUAL"
) {
  return items
    .filter((event) => event.status === status)
    .map(
      (event) =>
        new Date(event.endsAt).getTime() - new Date(event.startsAt).getTime()
    )
    .filter((duration) => Number.isFinite(duration) && duration > 0)
}

function durationSummary(values: number[]): DurationStats {
  if (!values.length) return { median: null, p90: null, count: 0 }
  const sorted = [...values].sort((a, b) => a - b)
  const middle = Math.floor(sorted.length / 2)
  const median =
    sorted.length % 2 === 0
      ? (sorted[middle - 1] + sorted[middle]) / 2
      : sorted[middle]
  return {
    median,
    p90: sorted[Math.max(0, Math.ceil(sorted.length * 0.9) - 1)],
    count: sorted.length,
  }
}
