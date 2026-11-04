import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"

import type { MyOnCallShiftItem } from "@/api/oncall-rotation/types"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"

/**
 * 「我的」视角：按本地日期分组的个人班次列表。
 *
 * 为什么不画进周/日网格：轮换按绝对日期轮转，同一个星期几在不同周可能属于不同的人；
 * 把个人班次映射到「星期 × 小时」会丢掉日期，两个不同日期的同名时段会被错误合并。
 * 这里用日期分组 + 起止时间，位置只表达「哪一天、从几点到几点」，不做跨周覆盖推断。
 *
 * 时间换算：班次接口给出 RFC3339 绝对时间与轮换时区，列表统一按浏览器本地时间展示，
 * 并在标题旁标注本地时区缩写；每一班再用文字标出它所属轮换的时区，避免把两者混为一谈。
 */

const DAY_MS = 24 * 60 * 60 * 1000

interface DayGroup {
  key: string
  date: Date
  shifts: MyOnCallShiftItem[]
}

/** 浏览器本地时区缩写（例如 GMT+8、CST）；取不到时返回空串，不展示该标注。 */
function browserTimeZoneShort(): string {
  try {
    const parts = new Intl.DateTimeFormat(undefined, {
      timeZoneName: "short",
    }).formatToParts(new Date())
    return parts.find((part) => part.type === "timeZoneName")?.value ?? ""
  } catch {
    return ""
  }
}

/** 本地日历日序号，用于判断班次是否跨天，避免直接比较毫秒在夏令时下算出 23/25 小时。 */
function localDayNumber(date: Date): number {
  return Math.round(
    Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()) / DAY_MS
  )
}

function formatTime(date: Date, locale: string): string {
  return new Intl.DateTimeFormat(locale, {
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).format(date)
}

function formatDay(date: Date, locale: string): string {
  return new Intl.DateTimeFormat(locale, {
    month: "long",
    day: "numeric",
    weekday: "long",
  }).format(date)
}

function groupByLocalDay(shifts: MyOnCallShiftItem[]): DayGroup[] {
  const groups: DayGroup[] = []
  const index = new Map<string, DayGroup>()
  const sorted = [...shifts].sort(
    (left, right) =>
      new Date(left.startAt).getTime() - new Date(right.startAt).getTime()
  )
  for (const shift of sorted) {
    const start = new Date(shift.startAt)
    if (Number.isNaN(start.getTime())) {
      continue
    }
    const key = `${start.getFullYear()}-${start.getMonth()}-${start.getDate()}`
    let group = index.get(key)
    if (!group) {
      group = { key, date: start, shifts: [] }
      index.set(key, group)
      groups.push(group)
    }
    group.shifts.push(shift)
  }
  return groups
}

export interface OnCallMyShiftsProps {
  shifts: MyOnCallShiftItem[]
  /** 展示未来多少天，与接口请求参数一致。 */
  days: number
  loading?: boolean
  failed?: boolean
  onRetry?: () => void
}

export function OnCallMyShifts({
  shifts,
  days,
  loading = false,
  failed = false,
  onRetry,
}: OnCallMyShiftsProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith("zh") ? "zh-CN" : "en-US"
  const localZone = browserTimeZoneShort()

  if (loading) {
    return (
      <div className="flex flex-col gap-2" role="status">
        <span className="sr-only">{t("state.loading")}</span>
        {Array.from({ length: 3 }, (_, index) => (
          <Skeleton key={index} className="h-14 w-full rounded-lg" />
        ))}
      </div>
    )
  }

  if (failed) {
    return (
      <div className="flex flex-wrap items-center gap-2" role="alert">
        <p className="text-sm text-destructive">{t("myShifts.failed")}</p>
        {onRetry ? (
          <Button variant="outline" size="sm" onClick={onRetry}>
            {t("action.retry")}
          </Button>
        ) : null}
      </div>
    )
  }

  if (shifts.length === 0) {
    return (
      <div className="flex flex-col items-start gap-2">
        <p className="text-sm text-muted-foreground">{t("myShifts.empty")}</p>
        <Button asChild variant="outline" size="sm">
          <Link to="/oncall-rotations">{t("myShifts.goRotations")}</Link>
        </Button>
      </div>
    )
  }

  const groups = groupByLocalDay(shifts)

  /** 跨天班次把结束时刻标成「次日 09:00」；更长跨度用「N 天后」，避免只显示一个会误读的时刻。 */
  function shiftRange(shift: MyOnCallShiftItem): string {
    const start = new Date(shift.startAt)
    const end = new Date(shift.endAt)
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
      return `${shift.startAt} – ${shift.endAt}`
    }
    const startTime = formatTime(start, locale)
    const endTime = formatTime(end, locale)
    const offset = localDayNumber(end) - localDayNumber(start)
    if (offset === 0) {
      return `${startTime} – ${endTime}`
    }
    const dayLabel =
      offset === 1
        ? t("myShifts.nextDay")
        : t("myShifts.dayOffset", { days: offset })
    return `${startTime} – ${dayLabel} ${endTime}`
  }

  return (
    <div className="flex min-w-0 flex-col gap-3">
      <div className="flex min-w-0 flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
        <p className="text-sm font-medium">
          {t("myShifts.title")}
          <span className="ml-2 text-xs font-normal text-muted-foreground">
            {t("myShifts.days", { days })}
          </span>
        </p>
        {localZone ? (
          <p className="text-xs text-muted-foreground">
            {t("myShifts.localZone", { zone: localZone })}
          </p>
        ) : null}
      </div>
      <ol className="flex min-w-0 flex-col gap-3">
        {groups.map((group) => (
          <li key={group.key} className="flex min-w-0 flex-col gap-1.5">
            <h3 className="text-xs font-medium text-muted-foreground">
              {formatDay(group.date, locale)}
            </h3>
            <ul className="flex min-w-0 flex-col gap-1.5">
              {group.shifts.map((shift) => (
                <li
                  key={`${shift.rotationUid}:${shift.startAt}`}
                  aria-current={shift.current ? "time" : undefined}
                  className="flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 rounded-lg border bg-background px-3 py-2"
                >
                  <span className="font-mono text-sm font-medium tabular-nums">
                    {shiftRange(shift)}
                  </span>
                  {shift.current ? (
                    <Badge
                      variant="secondary"
                      className="h-5 rounded-md px-2 font-normal"
                    >
                      {t("rotation.currentShift")}
                    </Badge>
                  ) : null}
                  <span className="min-w-0 text-sm break-words">
                    {shift.rotationName || shift.rotationUid}
                  </span>
                  <span className="text-xs break-words text-muted-foreground">
                    {t("oncallTimeView.timezone", { zone: shift.timezone })}
                  </span>
                </li>
              ))}
            </ul>
          </li>
        ))}
      </ol>
    </div>
  )
}
