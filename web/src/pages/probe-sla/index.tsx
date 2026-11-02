import { ChartNoAxesColumn, RefreshCw } from "lucide-react"
import { useEffect, useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useSearchParams } from "react-router-dom"

import { getProbeSla } from "@/api/probe-sla"
import type { ProbeSlaGroup, ProbeSlaReport } from "@/api/probe-sla/types"
import { listProbeTargetGroups } from "@/api/probe-target"
import { useSession } from "@/auth/session"
import { DateRangePicker } from "@/components/date-range-picker"
import { MultiSelect, type MultiSelectOption } from "@/components/multi-select"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import {
  addDays,
  dayEndISO,
  dayStartISO,
  diffDays,
  formatDateValue,
  parseDateValue,
} from "@/lib/date-range"
import { failureMessage } from "@/lib/form-error"
import { loadAllPagedItems } from "@/lib/select-pages"

// 默认最近 7 天（含今天）；服务端上限 30 天，选择器与页面都按这个上限拦截。
const DEFAULT_DAYS = 7
const MAX_DAYS = 30

const toneClasses = {
  success: "bg-success-soft text-success",
  danger: "bg-danger-soft text-danger",
  primary: "bg-primary-soft text-primary-ink",
  info: "bg-info-soft text-info",
  warning: "bg-warning-soft text-warning",
} as const

type Tone = keyof typeof toneClasses

/** 汇总卡片：标签 + 数值 + 一行口径说明，右侧图标只起识别作用，颜色不是唯一信号。 */
function StatCard({
  label,
  value,
  hint,
  icon,
  tone,
  badge,
}: {
  label: string
  value: ReactNode
  hint: string
  icon: ReactNode
  tone: Tone
  badge?: ReactNode
}) {
  return (
    <Card className="min-w-0 px-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 space-y-1">
          <p className="flex flex-wrap items-center gap-2 text-[11px] font-medium tracking-[0.08em] text-muted-foreground uppercase">
            {label}
            {badge}
          </p>
          <p className="font-heading text-2xl leading-tight font-semibold tracking-tight tabular-nums">
            {value}
          </p>
          <p className="text-xs leading-relaxed text-muted-foreground">
            {hint}
          </p>
        </div>
        <span
          aria-hidden="true"
          className={`grid size-9 shrink-0 place-items-center rounded-lg ${toneClasses[tone]}`}
        >
          {icon}
        </span>
      </div>
    </Card>
  )
}

/** 可用率字段是字符串；空串表示区间内没有检查，显示占位符而不是 0%。 */
function availabilityText(value: string): string {
  return value ? `${value}%` : "—"
}

/** 单个目标组的汇总 + 按天明细。 */
function GroupSection({ group }: { group: ProbeSlaGroup }) {
  const { t } = useTranslation()
  return (
    <Card className="min-w-0 px-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="min-w-0 font-heading text-sm font-medium break-words">
            {group.groupName || group.groupUid}
          </h2>
          <p className="text-xs text-muted-foreground">
            {t("probeSla.groupSummary", {
              checks: group.checks,
              failed: group.failedChecks,
              events: group.eventCount,
            })}
          </p>
        </div>
        <dl className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs">
          <div className="flex items-center gap-1.5">
            <dt className="text-muted-foreground">
              {t("probeSla.availability")}
            </dt>
            <dd className="font-mono font-medium tabular-nums">
              {availabilityText(group.availability)}
            </dd>
          </div>
          <div className="flex items-center gap-1.5">
            <dt className="text-muted-foreground">
              {t("probeSla.avgDuration")}
            </dt>
            <dd className="font-mono font-medium tabular-nums">
              {group.avgDurationMs} ms
            </dd>
          </div>
          <div className="flex items-center gap-1.5">
            <dt className="text-muted-foreground">{t("probeSla.p95")}</dt>
            <dd className="font-mono font-medium tabular-nums">
              {group.p95DurationMs} ms
              {group.p95Approximate ? (
                <span className="ml-1 rounded border border-warning-border bg-warning-soft px-1 text-[10px] font-normal text-warning">
                  {t("probeSla.approximate")}
                </span>
              ) : null}
            </dd>
          </div>
        </dl>
      </div>
      {group.days.length === 0 ? (
        <p className="mt-3 text-sm text-muted-foreground">
          {t("probeSla.noDailyData")}
        </p>
      ) : (
        <div className="mt-3 overflow-x-auto">
          <table className="w-full min-w-[36rem] text-sm">
            <caption className="sr-only">
              {t("probeSla.tableCaption", { name: group.groupName })}
            </caption>
            <thead>
              <tr className="border-b border-border/70 text-left text-xs text-muted-foreground">
                <th scope="col" className="py-2 pr-3 font-medium">
                  {t("probeSla.day")}
                </th>
                <th scope="col" className="py-2 pr-3 text-right font-medium">
                  {t("probeSla.checks")}
                </th>
                <th scope="col" className="py-2 pr-3 text-right font-medium">
                  {t("probeSla.okChecks")}
                </th>
                <th scope="col" className="py-2 pr-3 text-right font-medium">
                  {t("probeSla.failedChecks")}
                </th>
                <th scope="col" className="py-2 pr-3 text-right font-medium">
                  {t("probeSla.availability")}
                </th>
                <th scope="col" className="py-2 text-right font-medium">
                  {t("probeSla.avgDuration")}
                </th>
              </tr>
            </thead>
            <tbody>
              {group.days.map((day) => (
                <tr
                  key={day.day}
                  className="border-b border-border/60 last:border-0"
                >
                  <th
                    scope="row"
                    className="py-2 pr-3 text-left font-mono text-xs font-normal whitespace-nowrap tabular-nums"
                  >
                    {day.day}
                  </th>
                  <td className="py-2 pr-3 text-right font-mono text-xs tabular-nums">
                    {day.checks}
                  </td>
                  <td className="py-2 pr-3 text-right font-mono text-xs tabular-nums">
                    {day.okChecks}
                  </td>
                  <td className="py-2 pr-3 text-right font-mono text-xs tabular-nums">
                    {day.failedChecks}
                  </td>
                  <td className="py-2 pr-3 text-right font-mono text-xs tabular-nums">
                    {availabilityText(day.availability)}
                  </td>
                  <td className="py-2 text-right font-mono text-xs tabular-nums">
                    {day.avgDurationMs} ms
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Card>
  )
}

/**
 * 探测 SLA：按目标组与日期范围汇总可用率、延迟与失败次数。
 * 日期范围默认最近 7 天，筛选条件全部写进 URL，刷新与分享链接都能复现同一份报表。
 */
export function ProbeSlaPage() {
  const { t } = useTranslation()
  const session = useSession()
  const [searchParams, setSearchParams] = useSearchParams()
  const [groups, setGroups] = useState<MultiSelectOption[]>([])
  const [report, setReport] = useState<ProbeSlaReport | null>(null)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)

  const todayKey = formatDateValue(new Date())
  const defaultFrom = formatDateValue(addDays(new Date(), -DEFAULT_DAYS + 1))
  const fromParam = searchParams.get("from") ?? ""
  const toParam = searchParams.get("to") ?? ""
  const from = parseDateValue(fromParam) ? fromParam : defaultFrom
  const to = parseDateValue(toParam) ? toParam : todayKey
  const selectedGroups = (searchParams.get("groups") ?? "")
    .split(",")
    .map((uid) => uid.trim())
    .filter(Boolean)

  const fromDate = parseDateValue(from)
  const toDate = parseDateValue(to)
  const spanDays = fromDate && toDate ? diffDays(fromDate, toDate) + 1 : 0
  const rangeInvalid = spanDays < 1 || spanDays > MAX_DAYS

  // 目标组下拉数据按租户取一次；选择器的选项不随筛选变化。
  useEffect(() => {
    let active = true
    loadAllPagedItems((page, pageSize) =>
      listProbeTargetGroups({ page, pageSize })
    )
      .then((items) => {
        if (active) {
          setGroups(
            items.map((item) => ({ value: item.uid, label: item.name }))
          )
        }
      })
      .catch(() => {
        if (active) {
          setGroups([])
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  const groupsKey = searchParams.get("groups") ?? ""

  useEffect(() => {
    // 条件不合法时由渲染层直接给出提示，不发请求；这里也不能同步改 loading。
    if (!session.namespaceUid || rangeInvalid) {
      return
    }
    const fromISO = dayStartISO(from)
    const toISO = dayEndISO(to)
    if (!fromISO || !toISO) {
      return
    }
    const groupUids = groupsKey
      .split(",")
      .map((uid) => uid.trim())
      .filter(Boolean)
    let active = true
    getProbeSla({ groupUids, from: fromISO, to: toISO })
      .then((result) => {
        if (!active) {
          return
        }
        setReport(result)
        setFailed(false)
        setError("")
      })
      .catch((err) => {
        if (!active) {
          return
        }
        setFailed(true)
        setError(failureMessage(err, ""))
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid, from, to, rangeInvalid, groupsKey, attempt])

  if (!session.loading && !session.can("probe_sla:read")) {
    return <Navigate to="/forbidden" replace />
  }

  function updateRange(next: { from: string; to: string }) {
    setLoading(true)
    const params = new URLSearchParams(searchParams)
    if (next.from) {
      params.set("from", next.from)
    } else {
      params.delete("from")
    }
    if (next.to) {
      params.set("to", next.to)
    } else {
      params.delete("to")
    }
    setSearchParams(params)
  }

  function updateGroups(values: string[]) {
    setLoading(true)
    const params = new URLSearchParams(searchParams)
    if (values.length > 0) {
      params.set("groups", values.join(","))
    } else {
      params.delete("groups")
    }
    setSearchParams(params)
  }

  function retry() {
    setLoading(true)
    setAttempt((value) => value + 1)
  }

  function selectionSummary(selected: number, total: number): string {
    return selected >= total && total > 0
      ? t("multiSelect.allSelected", { count: selected })
      : t("multiSelect.summary", { count: selected, total })
  }

  const overall =
    report?.groups.reduce(
      (acc, group) => ({
        checks: acc.checks + group.checks,
        okChecks: acc.okChecks + group.okChecks,
        failedChecks: acc.failedChecks + group.failedChecks,
        weightedDuration:
          acc.weightedDuration + group.avgDurationMs * group.checks,
        p95: Math.max(acc.p95, group.p95DurationMs),
        approximate: acc.approximate || group.p95Approximate,
      }),
      {
        checks: 0,
        okChecks: 0,
        failedChecks: 0,
        weightedDuration: 0,
        p95: 0,
        approximate: false,
      }
    ) ?? null
  const availability =
    overall && overall.checks > 0
      ? ((overall.okChecks / overall.checks) * 100).toFixed(2)
      : ""
  const avgDuration =
    overall && overall.checks > 0
      ? Math.round(overall.weightedDuration / overall.checks)
      : 0

  return (
    <div className="min-h-0 flex-1 space-y-5 overflow-y-auto pr-1">
      <PageHeader
        title={t("probeSla.title")}
        description={t("probeSla.description")}
      >
        <div className="flex shrink-0 items-center gap-2">
          <Button variant="outline" disabled={loading} onClick={retry}>
            <RefreshCw
              aria-hidden="true"
              data-icon="inline-start"
              className={loading ? "animate-spin" : undefined}
            />
            {t("action.refresh")}
          </Button>
        </div>
      </PageHeader>
      <Card className="min-w-0 px-4">
        <div className="flex min-w-0 flex-wrap items-end gap-3">
          <div className="min-w-0 flex-1 basis-64">
            <label
              htmlFor="probe-sla-groups"
              className="mb-1.5 block text-xs font-medium text-muted-foreground"
            >
              {t("probeSla.targetGroups")}
            </label>
            <MultiSelect
              id="probe-sla-groups"
              values={selectedGroups}
              options={groups}
              onChange={updateGroups}
              placeholder={t("multiSelect.placeholder")}
              searchPlaceholder={t("multiSelect.searchPlaceholder")}
              emptyText={t("probeSla.noTargetGroups")}
              allLabel={t("multiSelect.all")}
              clearLabel={t("multiSelect.clear")}
              selectedLabel={selectionSummary}
            />
          </div>
          <div className="min-w-0">
            <label className="mb-1.5 block text-xs font-medium text-muted-foreground">
              {t("probeSla.dateRange")}
            </label>
            <DateRangePicker
              id="probe-sla-range"
              ariaLabel={t("probeSla.dateRange")}
              from={from}
              to={to}
              maxDays={MAX_DAYS}
              maxDate={todayKey}
              invalid={rangeInvalid}
              onChange={updateRange}
            />
          </div>
          <p className="min-w-0 basis-full text-xs text-muted-foreground">
            {t("probeSla.rangeHint", { max: MAX_DAYS })}
          </p>
        </div>
      </Card>
      {rangeInvalid ? (
        <Card className="min-w-0 border-danger-border bg-danger-soft px-4 py-3">
          <p className="text-sm text-danger">
            {spanDays < 1
              ? t("probeSla.errors.rangeInvalid")
              : t("probeSla.errors.rangeTooLong", { max: MAX_DAYS })}
          </p>
        </Card>
      ) : loading ? (
        <div className="grid gap-4">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
            {Array.from({ length: 5 }, (_, index) => (
              <Skeleton key={index} className="h-28 w-full" />
            ))}
          </div>
          <Skeleton className="h-64 w-full" />
        </div>
      ) : failed ? (
        <Card className="min-w-0 px-4 py-8 text-center">
          <p className="text-sm font-medium">{t("probeSla.loadFailed")}</p>
          <p className="mt-1 text-xs text-muted-foreground">
            {error || t("probeSla.loadFailedHint")}
          </p>
          <div className="mt-3 flex justify-center">
            <Button variant="outline" onClick={retry}>
              {t("action.retry")}
            </Button>
          </div>
        </Card>
      ) : report && report.groups.length > 0 && overall ? (
        <>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
            <StatCard
              label={t("probeSla.availability")}
              value={availabilityText(availability)}
              hint={t("probeSla.availabilityHint")}
              icon={<ChartNoAxesColumn className="size-4" />}
              tone="success"
            />
            <StatCard
              label={t("probeSla.checks")}
              value={overall.checks.toLocaleString()}
              hint={t("probeSla.checksHint")}
              icon={<ChartNoAxesColumn className="size-4" />}
              tone="primary"
            />
            <StatCard
              label={t("probeSla.failedChecks")}
              value={overall.failedChecks.toLocaleString()}
              hint={t("probeSla.failedChecksHint")}
              icon={<ChartNoAxesColumn className="size-4" />}
              tone="danger"
            />
            <StatCard
              label={t("probeSla.avgDuration")}
              value={`${avgDuration} ms`}
              hint={t("probeSla.avgDurationHint")}
              icon={<ChartNoAxesColumn className="size-4" />}
              tone="info"
            />
            <StatCard
              label={t("probeSla.p95")}
              value={`${overall.p95} ms`}
              hint={t("probeSla.p95Hint")}
              icon={<ChartNoAxesColumn className="size-4" />}
              tone="warning"
              badge={
                overall.approximate ? (
                  <span className="rounded border border-warning-border bg-warning-soft px-1 py-px text-[10px] font-normal tracking-normal text-warning normal-case">
                    {t("probeSla.approximate")}
                  </span>
                ) : null
              }
            />
          </div>
          <p className="text-xs text-muted-foreground">
            {t("probeSla.aggregateNote", {
              from,
              to,
            })}
          </p>
          <div className="grid gap-4">
            {report.groups.map((group) => (
              <GroupSection key={group.groupUid} group={group} />
            ))}
          </div>
        </>
      ) : (
        <Card className="min-w-0 px-6 py-12 text-center">
          <p className="text-sm font-medium">{t("probeSla.empty")}</p>
          <p className="mt-1 text-xs text-muted-foreground">
            {t("probeSla.emptyHint")}
          </p>
        </Card>
      )}
    </div>
  )
}
