import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"

import type { MyOnCallShiftItem } from "@/api/oncall-rotation/types"
import type {
  MyOnCallGroupWindowItem,
  OnCallWindow,
} from "@/api/notification-oncall/types"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"

/**
 * 「我的」视角：按本地日期分组的个人班次列表，外加「我所在通知组的接收时段」。
 *
 * 为什么不把班次画进周/日网格：轮换按绝对日期轮转，同一个星期几在不同周可能属于不同的人；
 * 把个人班次映射到「星期 × 小时」会丢掉日期，两个不同日期的同名时段会被错误合并。
 * 这里用日期分组 + 起止时间，位置只表达「哪一天、从几点到几点」，不做跨周覆盖推断。
 *
 * 通知组接收时段是另一类数据：它按周重复、绑在值班项上，不与具体日期绑定，
 * 因此单独一段展示；两个来源的请求、加载态、失败态、空态互不影响。
 *
 * 时间换算：班次接口给出 RFC3339 绝对时间与轮换时区，列表统一按浏览器本地时间展示，
 * 并在标题旁标注本地时区缩写；每一班再用文字标出它所属轮换的时区，避免把两者混为一谈。
 * 通知组时段沿用后端保存的本地时间（HH:mm），只标注它所属值班项的时区，不做二次换算。
 */

const DAY_MS = 24 * 60 * 60 * 1000

/** 星期按周一到周日排序，与后端保存顺序一致。 */
const WEEKDAYS = [
  "MONDAY",
  "TUESDAY",
  "WEDNESDAY",
  "THURSDAY",
  "FRIDAY",
  "SATURDAY",
  "SUNDAY",
] as const

type Weekday = (typeof WEEKDAYS)[number]

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

/** 把后端返回的星期去重并按周一到周日排序；未知取值直接丢弃。 */
function orderedWeekdays(days: string[]): Weekday[] {
  const unique = new Set(days.map((day) => (day ?? "").trim().toUpperCase()))
  return WEEKDAYS.filter((day) => unique.has(day))
}

/** 解析 HH:mm；非法值返回 null，避免脏数据把跨天时段标错。 */
function clockMinutes(value: string | undefined): number | null {
  const match = /^(\d{1,2}):(\d{2})$/.exec((value ?? "").trim())
  if (!match) {
    return null
  }
  const hour = Number(match[1])
  const minute = Number(match[2])
  if (hour > 23 || minute > 59) {
    return null
  }
  return hour * 60 + minute
}

/** start 晚于 end（或无法解析时按字符串比较）表示跨午夜。 */
function crossesMidnight(window: OnCallWindow): boolean {
  const start = clockMinutes(window.start)
  const end = clockMinutes(window.end)
  if (start !== null && end !== null) {
    return end < start
  }
  return (window.end ?? "") < (window.start ?? "")
}

/** 「周一、周三」「Mon, Wed」这种列表拼接交给 Intl，避免在代码里写中文顿号。 */
function joinDayLabels(labels: string[], locale: string): string {
  try {
    return new Intl.ListFormat(locale, {
      style: "narrow",
      type: "conjunction",
    }).format(labels)
  } catch {
    return labels.join(", ")
  }
}

/** 单个「值班项 × 通知组」卡片：值班项名为主，附通知组、时区与按周重复的时段。 */
function GroupWindowCard({ item }: { item: MyOnCallGroupWindowItem }) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith("zh") ? "zh-CN" : "en-US"
  const windows = item.windows ?? []

  /** 星期几归属于开始那天；跨午夜只改结束时刻的文案，不改变星期归属。 */
  function windowDays(window: OnCallWindow): string {
    const labels = orderedWeekdays(window.weekdays ?? []).map((day) =>
      t(`oncall.weekday.${day}`)
    )
    return labels.length > 0
      ? joinDayLabels(labels, locale)
      : t("oncall.anyDay")
  }

  return (
    <li
      aria-current={item.current ? "true" : undefined}
      className="flex min-w-0 flex-col gap-1.5 rounded-lg border bg-background px-3 py-2"
    >
      <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
        <span className="min-w-0 text-sm font-medium break-words">
          {item.oncallName || item.oncallUid}
        </span>
        {item.current ? (
          <Badge
            variant="secondary"
            className="h-5 rounded-md px-2 font-normal"
          >
            {t("myShifts.groupWindowsCurrent")}
          </Badge>
        ) : null}
        <span className="min-w-0 text-xs break-words text-muted-foreground">
          {t("myShifts.groupName", {
            name: item.groupName || item.groupUid,
          })}
        </span>
        <span className="min-w-0 text-xs break-words text-muted-foreground">
          {t("oncallTimeView.timezone", { zone: item.timezone })}
        </span>
      </div>
      {windows.length === 0 ? (
        <Badge
          variant="secondary"
          className="h-5 w-fit rounded-md px-2 font-normal"
        >
          {t("myShifts.groupWindowsAlways")}
        </Badge>
      ) : (
        <ul className="flex min-w-0 flex-col gap-1">
          {windows.map((window, index) => (
            <li
              key={`${window.start}-${window.end}-${index}`}
              className="flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5 text-xs"
            >
              <span className="min-w-0 break-words text-muted-foreground">
                {windowDays(window)}
              </span>
              <span className="font-mono text-muted-foreground tabular-nums">
                {window.start} –
              </span>
              {crossesMidnight(window) ? (
                <span className="text-muted-foreground">
                  {t("myShifts.nextDay")}
                </span>
              ) : null}
              <span className="font-mono text-muted-foreground tabular-nums">
                {window.end}
              </span>
            </li>
          ))}
        </ul>
      )}
    </li>
  )
}

export interface OnCallMyShiftsProps {
  shifts: MyOnCallShiftItem[]
  /** 展示未来多少天，与接口请求参数一致。 */
  days: number
  loading?: boolean
  failed?: boolean
  onRetry?: () => void
  /** 当前账号所在通知组的接收时段；与班次来自两个独立请求。 */
  groupWindows?: MyOnCallGroupWindowItem[]
  groupWindowsLoading?: boolean
  groupWindowsFailed?: boolean
  onGroupWindowsRetry?: () => void
}

export function OnCallMyShifts({
  shifts,
  days,
  loading = false,
  failed = false,
  onRetry,
  groupWindows = [],
  groupWindowsLoading = false,
  groupWindowsFailed = false,
  onGroupWindowsRetry,
}: OnCallMyShiftsProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith("zh") ? "zh-CN" : "en-US"
  const localZone = browserTimeZoneShort()

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

  const shiftsResolved = !loading && !failed
  const groupsResolved = !groupWindowsLoading && !groupWindowsFailed
  const bothEmpty =
    shiftsResolved &&
    groupsResolved &&
    shifts.length === 0 &&
    groupWindows.length === 0

  if (bothEmpty) {
    return (
      <div className="flex flex-col items-start gap-2">
        <p className="text-sm text-muted-foreground">{t("myShifts.empty")}</p>
        <Button asChild variant="outline" size="sm">
          <Link to="/oncall-rotations">{t("myShifts.goRotations")}</Link>
        </Button>
      </div>
    )
  }

  const shiftGroups = groupByLocalDay(shifts)

  const shiftsSection = loading ? (
    <div role="status" className="flex flex-col gap-2">
      <span className="sr-only">{t("state.loading")}</span>
      {Array.from({ length: 3 }, (_, index) => (
        <Skeleton key={index} className="h-14 w-full rounded-lg" />
      ))}
    </div>
  ) : failed ? (
    <div className="flex flex-wrap items-center gap-2" role="alert">
      <p className="text-sm text-destructive">{t("myShifts.failed")}</p>
      {onRetry ? (
        <Button variant="outline" size="sm" onClick={onRetry}>
          {t("action.retry")}
        </Button>
      ) : null}
    </div>
  ) : shifts.length > 0 ? (
    <section className="flex min-w-0 flex-col gap-3">
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
        {shiftGroups.map((group) => (
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
    </section>
  ) : null

  const groupsSection = groupWindowsLoading ? (
    <div role="status" className="flex flex-col gap-2">
      <span className="sr-only">{t("state.loading")}</span>
      {Array.from({ length: 2 }, (_, index) => (
        <Skeleton key={index} className="h-14 w-full rounded-lg" />
      ))}
    </div>
  ) : groupWindowsFailed ? (
    <div className="flex flex-wrap items-center gap-2" role="alert">
      <p className="text-sm text-destructive">
        {t("myShifts.groupWindowsFailed")}
      </p>
      {onGroupWindowsRetry ? (
        <Button variant="outline" size="sm" onClick={onGroupWindowsRetry}>
          {t("action.retry")}
        </Button>
      ) : null}
    </div>
  ) : groupWindows.length > 0 ? (
    <section className="flex min-w-0 flex-col gap-2">
      <p className="text-sm font-medium">{t("myShifts.groupWindowsTitle")}</p>
      <ul className="flex min-w-0 flex-col gap-1.5">
        {groupWindows.map((item) => (
          <GroupWindowCard
            key={`${item.oncallUid}:${item.groupUid}`}
            item={item}
          />
        ))}
      </ul>
    </section>
  ) : null

  return (
    <div className="flex min-w-0 flex-col gap-4">
      {shiftsSection}
      {groupsSection}
    </div>
  )
}
