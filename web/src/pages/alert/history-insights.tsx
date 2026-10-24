import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"

import { listAlertEvents } from "@/api/alert-event"
import type { AlertEventItem } from "@/api/alert-event/types"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { formatDuration } from "@/pages/alert/format"

const FETCH_PAGE_SIZE = 100
const EMPTY_EVENTS: AlertEventItem[] = []

type Granularity = "hour" | "day" | "week" | "month"
type StatusFilter =
  "ALL" | "FIRING" | "RECOVERED" | "RECOVERED_BY_MANUAL" | "SUPPRESSED"

interface AlertHistoryInsightsProps {
  name: string
  levelUid: string
  enabled: boolean
  onSelectFingerprint: (fingerprint: string) => void
}

export function AlertHistoryInsights({
  name,
  levelUid,
  enabled,
  onSelectFingerprint,
}: AlertHistoryInsightsProps) {
  const { t, i18n } = useTranslation()
  const [attempt, setAttempt] = useState(0)
  const queryKey = `${name}|${levelUid}|${attempt}`
  const [result, setResult] = useState<{
    queryKey: string
    events: AlertEventItem[]
    error: string
  } | null>(null)
  const [progress, setProgress] = useState({
    queryKey: "",
    loaded: 0,
    total: 0,
  })
  const loading = result?.queryKey !== queryKey
  const error = loading ? "" : (result?.error ?? "")
  const events = loading ? EMPTY_EVENTS : (result?.events ?? EMPTY_EVENTS)
  const [params, setParams] = useSearchParams()
  const from = params.get("from") || ""
  const to = params.get("to") || ""
  const status = parseStatusFilter(params.get("status"))
  const granularity = parseGranularity(params.get("interval"))
  const invalidDateRange = Boolean(from && to && from > to)
  const filterSignature = `${from}|${to}|${status}|${granularity}`
  const [selectedBucket, setSelectedBucket] = useState<{
    key: number
    filterSignature: string
  } | null>(null)
  const selectedBucketKey =
    selectedBucket?.filterSignature === filterSignature
      ? selectedBucket.key
      : null

  function updateAnalysisFilter(
    key: "from" | "to" | "status" | "interval",
    value: string
  ) {
    setSelectedBucket(null)
    const next = new URLSearchParams(params)
    if (value) next.set(key, value)
    else next.delete(key)
    setParams(next, { replace: true })
  }

  useEffect(() => {
    if (!enabled || result?.queryKey === queryKey) {
      return
    }
    let active = true

    async function loadAllHistory() {
      try {
        const first = await listAlertEvents({
          page: 1,
          pageSize: FETCH_PAGE_SIZE,
          view: "HISTORY",
          name,
          levelUid: levelUid || undefined,
        })
        const all = [...(first.items ?? [])]
        const total = first.total ?? all.length
        if (active) {
          setProgress({ queryKey, loaded: all.length, total })
        }
        const pageCount = Math.ceil(total / FETCH_PAGE_SIZE)
        for (let page = 2; page <= pageCount && active; page += 1) {
          const result = await listAlertEvents({
            page,
            pageSize: FETCH_PAGE_SIZE,
            view: "HISTORY",
            name,
            levelUid: levelUid || undefined,
          })
          all.push(...(result.items ?? []))
          if (active) {
            setProgress({ queryKey, loaded: all.length, total })
          }
        }
        if (active) {
          setResult({ queryKey, events: all, error: "" })
        }
      } catch {
        if (active) {
          setResult({
            queryKey,
            events: [],
            error: t("alertEvent.analysisLoadFailed"),
          })
        }
      }
    }

    void loadAllHistory()
    return () => {
      active = false
    }
  }, [enabled, name, levelUid, queryKey, attempt, result, t])

  const filtered = useMemo(() => {
    const fromMs = from
      ? new Date(`${from}T00:00:00`).getTime()
      : Number.NEGATIVE_INFINITY
    const toMs = to
      ? new Date(`${to}T23:59:59.999`).getTime()
      : Number.POSITIVE_INFINITY
    return events.filter((event) => {
      const started = new Date(event.startsAt).getTime()
      if (from && (!Number.isFinite(started) || started < fromMs)) return false
      if (to && (!Number.isFinite(started) || started > toMs)) return false
      if (status === "SUPPRESSED") return event.suppressed
      return status === "ALL" || event.status === status
    })
  }, [events, from, to, status])

  const timeline = useMemo(
    () => buildTimeline(filtered, granularity, i18n.language, from, to),
    [filtered, granularity, i18n.language, from, to]
  )
  const focusedBucket = timeline.points.find(
    (point) => point.key === selectedBucketKey
  )
  const scopedEvents = useMemo(
    () =>
      focusedBucket
        ? filtered.filter(
            (event) =>
              bucketStart(new Date(event.startsAt), granularity).getTime() ===
              focusedBucket.key
          )
        : filtered,
    [filtered, focusedBucket, granularity]
  )
  const statuses = useMemo(
    () => countBy(scopedEvents, (event) => event.status),
    [scopedEvents]
  )
  const levels = useMemo(
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
    (total, item) => total + item.count - 1,
    0
  )
  const {
    uniqueFingerprints,
    suppressedCount,
    autoRecovery,
    manualRecovery,
    recoveredCount,
  } = useMemo(() => {
    const fingerprints = new Set(
      scopedEvents.map((event) => event.fingerprint).filter(Boolean)
    )
    const automaticDurations = recoveryDurations(scopedEvents, "RECOVERED")
    const manualDurations = recoveryDurations(
      scopedEvents,
      "RECOVERED_BY_MANUAL"
    )
    return {
      uniqueFingerprints: fingerprints.size,
      suppressedCount: scopedEvents.filter((event) => event.suppressed).length,
      autoRecovery: durationSummary(automaticDurations),
      manualRecovery: durationSummary(manualDurations),
      recoveredCount: scopedEvents.filter(
        (event) =>
          event.status === "RECOVERED" || event.status === "RECOVERED_BY_MANUAL"
      ).length,
    }
  }, [scopedEvents])
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone

  return (
    <section
      className="space-y-4"
      aria-labelledby="alert-history-insights-title"
    >
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-xs font-semibold tracking-[0.12em] text-muted-foreground uppercase">
            {t("alertEvent.analysisEyebrow")}
          </p>
          <h2
            id="alert-history-insights-title"
            className="mt-1 text-lg font-semibold tracking-tight"
          >
            {t("alertEvent.analysisTitle")}
          </h2>
        </div>
        <div
          role="group"
          className="flex flex-wrap items-center gap-2"
          aria-label={t("alertEvent.analysisFilters")}
        >
          <Input
            aria-label={t("alertEvent.fromDate")}
            className="h-11 w-[9.5rem] tabular-nums"
            type="date"
            max={to || undefined}
            aria-invalid={invalidDateRange}
            aria-describedby={
              invalidDateRange ? "alert-history-range-error" : undefined
            }
            value={from}
            onChange={(event) =>
              updateAnalysisFilter("from", event.target.value)
            }
          />
          <span className="text-xs text-muted-foreground">
            {t("alertEvent.dateTo")}
          </span>
          <Input
            aria-label={t("alertEvent.toDate")}
            className="h-11 w-[9.5rem] tabular-nums"
            type="date"
            min={from || undefined}
            aria-invalid={invalidDateRange}
            aria-describedby={
              invalidDateRange ? "alert-history-range-error" : undefined
            }
            value={to}
            onChange={(event) => updateAnalysisFilter("to", event.target.value)}
          />
          <Select
            value={status}
            onValueChange={(value) =>
              updateAnalysisFilter("status", value === "ALL" ? "" : value)
            }
          >
            <SelectTrigger
              className="h-11 w-44"
              aria-label={t("alertEvent.status")}
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">{t("alertEvent.allStatuses")}</SelectItem>
              <SelectItem value="FIRING">
                {t("alertEvent.statusFiring")}
              </SelectItem>
              <SelectItem value="RECOVERED">
                {t("alertEvent.statusRecovered")}
              </SelectItem>
              <SelectItem value="RECOVERED_BY_MANUAL">
                {t("alertEvent.statusRecoveredManual")}
              </SelectItem>
              <SelectItem value="SUPPRESSED">
                {t("alertEvent.suppressed")}
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
              className="h-11 w-32"
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
            className="h-11 px-3"
            disabled={!from && !to && status === "ALL" && granularity === "day"}
            onClick={() => {
              setSelectedBucket(null)
              const next = new URLSearchParams(params)
              next.delete("from")
              next.delete("to")
              next.delete("status")
              next.delete("interval")
              setParams(next, { replace: true })
            }}
          >
            {t("alertEvent.clearAnalysisFilters")}
          </Button>
        </div>
      </div>

      <p className="text-xs text-muted-foreground">
        {invalidDateRange ? (
          <span
            id="alert-history-range-error"
            className="mr-2 text-sm text-destructive"
            role="alert"
          >
            {t("alertEvent.invalidDateRange")}
          </span>
        ) : null}
        {t("alertEvent.analysisScopeNote")}{" "}
        {t("alertEvent.localTimeZone", { timeZone })}
      </p>

      {loading ? (
        <div
          className="rounded-lg border border-border/70 bg-muted/30 px-4 py-3 text-sm text-muted-foreground"
          role="status"
        >
          {t("alertEvent.analysisLoading", {
            loaded: progress.queryKey === queryKey ? progress.loaded : 0,
            total: progress.queryKey === queryKey ? progress.total : 0,
          })}
        </div>
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
      ) : (
        <>
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
              detail={t("alertEvent.repeatEpisodes", {
                count: repeatedEpisodes.toLocaleString(i18n.language),
              })}
            />
            <Metric
              label={t("alertEvent.suppressedCount")}
              value={suppressedCount.toLocaleString(i18n.language)}
              detail={t("alertEvent.suppressedHelp")}
            />
            <Metric
              label={t("alertEvent.recoveredCount")}
              value={recoveredCount.toLocaleString(i18n.language)}
              detail={t("alertEvent.recoveredHelp")}
            />
          </div>

          <div className="grid gap-3 xl:grid-cols-[minmax(0,1.7fr)_minmax(17rem,1fr)]">
            <section
              className="min-w-0 rounded-xl border border-border/70 bg-card p-4"
              aria-label={t("alertEvent.trendTitle")}
            >
              <div className="mb-4 flex flex-wrap items-baseline justify-between gap-3">
                <div>
                  <h3 className="text-sm font-semibold">
                    {t("alertEvent.trendTitle")}
                  </h3>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {t("alertEvent.startsAtTrend")}
                  </p>
                </div>
                {focusedBucket ? (
                  <div className="flex items-center gap-2 text-xs">
                    <span className="text-muted-foreground">
                      {t("alertEvent.focusedBucket", {
                        label: focusedBucket.label,
                        count: focusedBucket.count.toLocaleString(
                          i18n.language
                        ),
                      })}
                    </span>
                    <Button
                      type="button"
                      size="sm"
                      variant="ghost"
                      onClick={() => setSelectedBucket(null)}
                    >
                      {t("alertEvent.clearBucket")}
                    </Button>
                  </div>
                ) : null}
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
                          ? Math.max(
                              4,
                              (point.count / timelineMaximum(timeline.points)) *
                                100
                            )
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
              data={statuses}
              labelFor={(value) => statusLabel(value, t)}
              total={scopedEvents.length}
              selectedValue={status === "ALL" ? "" : status}
              onSelect={(value) =>
                updateAnalysisFilter("status", value === status ? "" : value)
              }
            />
            <Distribution
              title={t("alertEvent.levelDistribution")}
              data={levels}
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
                title={item.fingerprint}
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
            <span className="font-mono">
              {formatDuration("1970-01-01T00:00:00Z", stats.median)}
            </span>
          </p>
          {showP90 && stats.p90 !== null ? (
            <p className="text-xs text-muted-foreground tabular-nums">
              {t("alertEvent.p90Label")}{" "}
              <span className="font-mono">
                {formatDuration("1970-01-01T00:00:00Z", stats.p90)}
              </span>
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
  if (
    value === "FIRING" ||
    value === "RECOVERED" ||
    value === "RECOVERED_BY_MANUAL" ||
    value === "SUPPRESSED"
  ) {
    return value
  }
  return "ALL"
}

function parseGranularity(value: string | null): Granularity {
  if (value === "hour" || value === "week" || value === "month") return value
  return "day"
}

function statusLabel(value: string, t: (key: string) => string) {
  if (value === "FIRING") return t("alertEvent.statusFiring")
  if (value === "RECOVERED") return t("alertEvent.statusRecovered")
  if (value === "RECOVERED_BY_MANUAL")
    return t("alertEvent.statusRecoveredManual")
  return value
}

function buildTimeline(
  items: AlertEventItem[],
  granularity: Granularity,
  language: string,
  from: string,
  to: string
) {
  const bins = new Map<number, number>()
  const dates: Date[] = []
  for (const item of items) {
    const date = new Date(item.startsAt)
    if (!Number.isFinite(date.getTime())) continue
    dates.push(date)
    const key = bucketStart(date, granularity).getTime()
    bins.set(key, (bins.get(key) ?? 0) + 1)
  }
  if (bins.size === 0) return { points: [], gapsOmitted: false }

  const rangeStart = from
    ? new Date(`${from}T00:00:00`)
    : dates.reduce(
        (earliest, date) => (date < earliest ? date : earliest),
        dates[0]
      )
  const rangeEnd = to
    ? new Date(`${to}T23:59:59.999`)
    : dates.reduce((latest, date) => (date > latest ? date : latest), dates[0])
  if (
    !Number.isFinite(rangeStart.getTime()) ||
    !Number.isFinite(rangeEnd.getTime()) ||
    rangeStart > rangeEnd
  ) {
    return { points: [], gapsOmitted: false }
  }

  const first = bucketStart(rangeStart, granularity)
  const last = bucketStart(rangeEnd, granularity)
  const keys: number[] = []
  let cursor = first
  while (cursor <= last && keys.length <= 180) {
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
  return {
    points: timelineKeys.map((key) => ({
      key,
      count: bins.get(key) ?? 0,
      label: new Intl.DateTimeFormat(locale, options).format(new Date(key)),
    })),
    gapsOmitted,
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

function timelineMaximum(points: Array<{ count: number }>) {
  return Math.max(1, ...points.map((point) => point.count))
}
