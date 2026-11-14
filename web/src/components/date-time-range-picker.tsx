import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Clock,
  X,
} from "lucide-react"
import {
  useEffect,
  useId,
  useMemo,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
} from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/class-names"
import {
  addDays,
  addMonths,
  combineDateTime,
  dayNumber,
  formatDateValue,
  monthDays,
  normalizeTimeValue,
  parseDateValue,
  parseDateTimeValue,
  startOfMonth,
  timePartOf,
} from "@/lib/date-range"

/**
 * 值的格式是本地时区的 YYYY-MM-DDTHH:mm，与 <input type="datetime-local"> 完全一致：
 * 不带时区、不带秒，交接给 new Date() 时也按本地时间解释，不会出现来回转换的时区漂移。
 * 需要提交 RFC3339 的调用方自行 toISOString()。
 */
export interface DateTimeRange {
  from: string
  to: string
}

interface DateTimeRangePickerProps {
  from: string
  to: string
  onChange: (range: DateTimeRange) => void
  /** 可选的最大天数（含首尾）；不传表示不限制。 */
  maxDays?: number
  className?: string
  placeholder?: string
  ariaLabel?: string
  describedBy?: string
  invalid?: boolean
  disabled?: boolean
  id?: string
}

// 只选了日期还没调时间时的落点：开始取当天 00:00、结束取当天 23:59，
// 相当于「一整天」，运维再按需要改成具体时刻。
const DEFAULT_START_TIME = "00:00"
const DEFAULT_END_TIME = "23:59"

function startOfDay(value: Date): Date {
  return new Date(value.getFullYear(), value.getMonth(), value.getDate())
}

// 自然日 + 时刻拼成本地 YYYY-MM-DDTHH:mm；时刻不合法就退回整天落点。
function startValueOf(day: Date | null, time: string): string {
  return day
    ? combineDateTime(day, normalizeTimeValue(time) || DEFAULT_START_TIME)
    : ""
}

function endValueOf(day: Date | null, time: string): string {
  return day
    ? combineDateTime(day, normalizeTimeValue(time) || DEFAULT_END_TIME)
    : ""
}

/** 触发器里把 YYYY-MM-DDTHH:mm 显示成更易读的 YYYY-MM-DD HH:mm。 */
function displayValue(value: string): string {
  return value ? value.replace("T", " ") : ""
}

/**
 * 双月日历 + 时/分输入的日期时间范围选择器。
 *
 * 与按天选择的 DateRangePicker 分开维护：
 * - 允许未来日期（维护窗口通常安排在以后的某个时刻）；
 * - 不默认限制跨度，需要时用 maxDays 收口；
 * - 面板里可编辑开始/结束的 HH:mm，选日期不会丢掉已选时间；
 * - 日期与时间都在本地时区，输出 YYYY-MM-DDTHH:mm。
 */
export function DateTimeRangePicker({
  from,
  to,
  onChange,
  maxDays,
  className,
  placeholder,
  ariaLabel,
  describedBy,
  invalid = false,
  disabled = false,
  id,
}: DateTimeRangePickerProps) {
  const { t, i18n } = useTranslation()
  const fieldId = useId()
  const locale = i18n.language.startsWith("zh") ? "zh-CN" : "en-US"
  const [open, setOpen] = useState(false)
  // 草稿按自然日保存，时间单独存 HH:mm：改日期时时间保持不变。
  const [draftFrom, setDraftFrom] = useState<Date | null>(null)
  const [draftTo, setDraftTo] = useState<Date | null>(null)
  const [startTime, setStartTime] = useState(DEFAULT_START_TIME)
  const [endTime, setEndTime] = useState(DEFAULT_END_TIME)
  const [hovered, setHovered] = useState<Date | null>(null)
  const [focused, setFocused] = useState<Date | null>(null)
  // 维护窗口通常安排在未来，打开时直接落在开始日所在月，而不是固定的本月。
  const [viewMonth, setViewMonth] = useState<Date>(() =>
    startOfMonth(new Date())
  )
  // 内容在 Portal 里延迟挂载，用状态接住节点，节点到位后焦点才能落到日期上。
  const [contentEl, setContentEl] = useState<HTMLDivElement | null>(null)

  const hasValue = Boolean(from || to)
  const spannedDays =
    maxDays !== undefined && draftFrom && draftTo
      ? Math.abs(dayNumber(draftTo) - dayNumber(draftFrom)) + 1
      : 0
  const rangeTooLong = maxDays !== undefined && spannedDays > maxDays

  const weekdays = useMemo(() => {
    const format = new Intl.DateTimeFormat(locale, {
      weekday: locale === "zh-CN" ? "narrow" : "short",
    })
    // 2024-01-01 是周一，按周一起始取七天表头。
    const monday = new Date(2024, 0, 1)
    return Array.from({ length: 7 }, (_, index) =>
      format.format(addDays(monday, index))
    )
  }, [locale])

  const dayFormat = useMemo(
    () => new Intl.DateTimeFormat(locale, { dateStyle: "long" }),
    [locale]
  )
  const monthFormat = useMemo(
    () => new Intl.DateTimeFormat(locale, { year: "numeric", month: "long" }),
    [locale]
  )

  useEffect(() => {
    if (!open || !focused || !contentEl) {
      return
    }
    const node = contentEl.querySelector<HTMLButtonElement>(
      `[data-day="${formatDateValue(focused)}"]`
    )
    node?.focus()
  }, [open, focused, contentEl])

  function handleOpenChange(next: boolean) {
    setOpen(next)
    if (!next) {
      return
    }
    const start = parseDateTimeValue(from)
    const end = parseDateTimeValue(to)
    const startDay = start ? startOfDay(start) : null
    const endDay = end ? startOfDay(end) : null
    setDraftFrom(startDay)
    setDraftTo(startDay && endDay ? endDay : null)
    setStartTime(timePartOf(from) || DEFAULT_START_TIME)
    setEndTime(timePartOf(to) || DEFAULT_END_TIME)
    setHovered(null)
    const anchor = startDay ?? new Date()
    setViewMonth(startOfMonth(anchor))
    setFocused(startDay ?? startOfMonth(anchor))
  }

  // 可见两个月是 [viewMonth 的第一天, 右月最后一天]。
  function visibleBounds(month: Date) {
    const first = dayNumber(startOfMonth(month))
    const last = dayNumber(addDays(addMonths(startOfMonth(month), 2), -1))
    return { first, last }
  }

  // 不限制未来日期；只有调用方传了 maxDays 才在选好起点后收口。
  function isDayDisabled(day: Date) {
    if (maxDays === undefined || !draftFrom || draftTo) {
      return false
    }
    return Math.abs(dayNumber(day) - dayNumber(draftFrom)) + 1 > maxDays
  }

  // 打开/翻月时的键盘落点：优先用 preferred（已选开始日或今天），但它必须
  // 落在可见两个月内且可选；否则退到可见范围内第一个可选日，保证方向键
  // 永远落在真实可选的日期上。
  function focusableDay(month: Date, preferred: Date): Date {
    const { first, last } = visibleBounds(month)
    const value = dayNumber(preferred)
    if (value >= first && value <= last && !isDayDisabled(preferred)) {
      return preferred
    }
    for (let offset = 0; offset <= last - first; offset += 1) {
      const candidate = addDays(startOfMonth(month), offset)
      if (!isDayDisabled(candidate)) {
        return candidate
      }
    }
    return preferred
  }

  /**
   * 把草稿同步给父级。时间输入框可能被清空或停在半截输入，这时按整天落点
   * （开始 00:00 / 结束 23:59）补齐，避免把空时间拼进提交值；
   * 还没选开始日则整体清空。
   */
  function emit(
    nextFrom: Date | null,
    nextTo: Date | null,
    nextStartTime = startTime,
    nextEndTime = endTime
  ) {
    onChange({
      from: startValueOf(nextFrom, nextStartTime),
      to: endValueOf(nextFrom && nextTo ? nextTo : null, nextEndTime),
    })
  }

  function pickDay(day: Date) {
    if (isDayDisabled(day)) {
      return
    }
    setFocused(day)
    if (!draftFrom || draftTo) {
      // 重新开一段：先只定开始日，等第二次点击补齐结束日。
      setDraftFrom(day)
      setDraftTo(null)
      setHovered(null)
      emit(day, null)
      return
    }
    const start = dayNumber(day) < dayNumber(draftFrom) ? day : draftFrom
    const end = dayNumber(day) < dayNumber(draftFrom) ? draftFrom : day
    setDraftFrom(start)
    setDraftTo(end)
    setHovered(null)
    emit(start, end)
    // 面板保持打开，方便紧接着调时分。
  }

  function handleStartTime(next: string) {
    setStartTime(next)
    if (draftFrom) {
      emit(draftFrom, draftTo, next, endTime)
    }
  }

  function handleEndTime(next: string) {
    setEndTime(next)
    if (draftFrom) {
      emit(draftFrom, draftTo, startTime, next)
    }
  }

  function moveFocus(current: Date, days: number) {
    const next = addDays(current, days)
    // 禁用的日子不能聚焦，焦点停在边界上，避免状态偷偷跳过一段不可选区间。
    if (isDayDisabled(next)) {
      return
    }
    setFocused(next)
    const monthOffset =
      (next.getFullYear() - viewMonth.getFullYear()) * 12 +
      (next.getMonth() - viewMonth.getMonth())
    if (monthOffset < 0) {
      setViewMonth(startOfMonth(next))
    } else if (monthOffset > 1) {
      setViewMonth(addMonths(startOfMonth(next), -1))
    }
  }

  function handleGridKeyDown(event: ReactKeyboardEvent<HTMLDivElement>) {
    const target = (event.target as HTMLElement).closest<HTMLElement>(
      "[data-day]"
    )
    const value = target?.dataset.day
    if (!value) {
      return
    }
    const current = parseDateValue(value)
    if (!current) {
      return
    }
    const deltas: Record<string, number> = {
      ArrowLeft: -1,
      ArrowRight: 1,
      ArrowUp: -7,
      ArrowDown: 7,
    }
    const delta = deltas[event.key]
    if (delta === undefined) {
      return
    }
    event.preventDefault()
    moveFocus(current, delta)
  }

  const startNum = draftFrom ? dayNumber(draftFrom) : null
  const previewEnd = draftFrom && !draftTo && hovered ? hovered : draftTo
  const endNum = previewEnd ? dayNumber(previewEnd) : null
  const rangeLow =
    startNum !== null && endNum !== null ? Math.min(startNum, endNum) : startNum
  const rangeHigh =
    startNum !== null && endNum !== null ? Math.max(startNum, endNum) : startNum

  // 就地提示：同一天里开始时刻不早于结束时刻时立即可见。
  // 跨天时日期本身已经按先后规整过，所以这里只需比较拼好的本地字符串。
  const draftFromValue = startValueOf(draftFrom, startTime)
  const draftToValue = endValueOf(draftFrom && draftTo ? draftTo : null, endTime)
  const orderInvalid = Boolean(
    draftFromValue && draftToValue && draftToValue <= draftFromValue
  )

  function shiftMonth(months: number) {
    const next = addMonths(viewMonth, months)
    setViewMonth(next)
    // 翻月后把落点放到新可见范围内的可选日期，否则键盘用户会失去落点。
    setFocused(focusableDay(next, next))
  }

  function reset() {
    setDraftFrom(null)
    setDraftTo(null)
    setStartTime(DEFAULT_START_TIME)
    setEndTime(DEFAULT_END_TIME)
    setHovered(null)
    onChange({ from: "", to: "" })
  }

  function renderMonth(month: Date, align: "left" | "right") {
    const days = monthDays(month)
    return (
      <div className="w-full min-w-0 sm:w-[17rem]" key={month.toISOString()}>
        <div className="mb-2 flex items-center justify-between">
          {align === "left" ? (
            <div className="flex items-center">
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                aria-label={t("dateRange.prevYear")}
                onClick={() => shiftMonth(-12)}
              >
                <ChevronsLeft aria-hidden />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                aria-label={t("dateRange.prevMonth")}
                onClick={() => shiftMonth(-1)}
              >
                <ChevronLeft aria-hidden />
              </Button>
            </div>
          ) : (
            <span className="w-12" aria-hidden />
          )}
          <span className="text-sm font-semibold">
            {monthFormat.format(month)}
          </span>
          {align === "right" ? (
            <div className="flex items-center">
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                aria-label={t("dateRange.nextMonth")}
                onClick={() => shiftMonth(1)}
              >
                <ChevronRight aria-hidden />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                aria-label={t("dateRange.nextYear")}
                onClick={() => shiftMonth(12)}
              >
                <ChevronsRight aria-hidden />
              </Button>
            </div>
          ) : (
            <span className="w-12" aria-hidden />
          )}
        </div>
        <div className="grid grid-cols-7 gap-0.5 pb-1" aria-hidden="true">
          {weekdays.map((weekday) => (
            <span
              key={weekday}
              className="flex h-7 items-center justify-center text-[11px] font-medium text-muted-foreground"
            >
              {weekday}
            </span>
          ))}
        </div>
        <div
          className="grid grid-cols-7 gap-0.5"
          role="group"
          aria-label={monthFormat.format(month)}
          onKeyDown={handleGridKeyDown}
        >
          {days.map((day) => {
            const num = dayNumber(day)
            const outside = day.getMonth() !== month.getMonth()
            const dayDisabled = isDayDisabled(day)
            const isEndpoint =
              startNum !== null &&
              (num === startNum || (endNum !== null && num === endNum))
            const inRange =
              !isEndpoint &&
              rangeLow !== null &&
              rangeHigh !== null &&
              num > rangeLow &&
              num < rangeHigh
            const isFocused =
              focused !== null && dayNumber(focused) === num && !dayDisabled
            return (
              <button
                key={`${month.getMonth()}-${num}`}
                type="button"
                data-day={formatDateValue(day)}
                tabIndex={isFocused ? 0 : -1}
                disabled={dayDisabled}
                aria-pressed={isEndpoint}
                aria-label={dayFormat.format(day)}
                onMouseEnter={() => {
                  // 超范围的日期不参与预览，否则会画出超过 maxDays 的区间。
                  if (draftFrom && !draftTo && !dayDisabled) {
                    setHovered(day)
                  }
                }}
                onMouseLeave={() => {
                  if (draftFrom && !draftTo) {
                    setHovered(null)
                  }
                }}
                onClick={() => pickDay(day)}
                className={cn(
                  "flex h-9 items-center justify-center rounded-md text-sm tabular-nums transition-colors outline-none",
                  "focus-visible:ring-2 focus-visible:ring-ring",
                  outside && "text-muted-foreground/40",
                  !dayDisabled && !isEndpoint && !inRange && "hover:bg-muted",
                  inRange && "bg-primary/15 text-foreground",
                  isEndpoint &&
                    "bg-primary font-semibold text-primary-foreground hover:bg-primary",
                  dayDisabled &&
                    "cursor-not-allowed text-muted-foreground/30 hover:bg-transparent",
                  num === dayNumber(new Date()) && !isEndpoint && !dayDisabled
                    ? "ring-1 ring-border"
                    : ""
                )}
              >
                {day.getDate()}
              </button>
            )
          })}
        </div>
      </div>
    )
  }

  return (
    <div
      className={cn(
        "flex h-8 w-full items-center gap-0.5 rounded-lg border border-input bg-background px-0.5 focus-within:border-ring focus-within:ring-3 focus-within:ring-ring dark:bg-input/30",
        invalid && "border-destructive ring-3 ring-destructive/20",
        disabled && "pointer-events-none opacity-50",
        className
      )}
    >
      <Popover open={open} onOpenChange={handleOpenChange}>
        <PopoverTrigger asChild>
          <button
            type="button"
            id={id}
            disabled={disabled}
            aria-label={ariaLabel ?? t("dateRange.openDateTime")}
            aria-invalid={invalid || undefined}
            aria-describedby={describedBy}
            className="flex min-w-0 flex-1 items-center gap-1.5 rounded-md px-1 py-1 text-left text-sm outline-none"
          >
            {hasValue ? (
              <span className="min-w-0 truncate font-mono tabular-nums">
                {displayValue(from) || "…"}
                <span className="px-1 text-muted-foreground">→</span>
                {displayValue(to) || "…"}
              </span>
            ) : (
              <span className="truncate text-muted-foreground">
                {placeholder ?? t("dateRange.placeholderDateTime")}
              </span>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent
          ref={setContentEl}
          align="start"
          sideOffset={6}
          collisionPadding={12}
          className="z-[60] max-h-[calc(100vh-4rem)] w-[min(18.5rem,calc(100vw-2rem))] overflow-x-hidden overflow-y-auto p-3 sm:w-[36.5rem]"
          // 焦点交给下面的日期网格；否则原语会把焦点放在第一个按钮（翻月键）上。
          onOpenAutoFocus={(event) => event.preventDefault()}
        >
          <div className="flex flex-col gap-3">
            {/* 窄屏单列堆叠，双月时并排；固定面板宽度避免弹层横向溢出视口。 */}
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              {renderMonth(viewMonth, "left")}
              {renderMonth(addMonths(viewMonth, 1), "right")}
            </div>
            {/* 窄屏一行一个，避免 h-8 的时间控件被挤到看不见分钟。 */}
            <div className="grid grid-cols-1 gap-2 border-t border-border/60 pt-2 sm:grid-cols-2">
              <div className="grid min-w-0 gap-1">
                <Label
                  htmlFor={`${fieldId}-start`}
                  className="text-xs font-normal text-muted-foreground"
                >
                  <Clock aria-hidden className="size-3.5" />
                  {t("dateRange.startTime")}
                </Label>
                <Input
                  id={`${fieldId}-start`}
                  type="time"
                  value={startTime}
                  aria-invalid={orderInvalid || undefined}
                  onChange={(event) => handleStartTime(event.target.value)}
                />
              </div>
              <div className="grid min-w-0 gap-1">
                <Label
                  htmlFor={`${fieldId}-end`}
                  className="text-xs font-normal text-muted-foreground"
                >
                  <Clock aria-hidden className="size-3.5" />
                  {t("dateRange.endTime")}
                </Label>
                <Input
                  id={`${fieldId}-end`}
                  type="time"
                  value={endTime}
                  aria-invalid={orderInvalid || undefined}
                  onChange={(event) => handleEndTime(event.target.value)}
                />
              </div>
            </div>
            {orderInvalid ? (
              <p role="alert" className="text-xs text-destructive">
                {t("dateRange.timeOrderInvalid")}
              </p>
            ) : null}
            {maxDays !== undefined ? (
              <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border/60 pt-2 text-xs text-muted-foreground">
                <span>{t("dateRange.maxDaysHint", { count: maxDays })}</span>
                {rangeTooLong ? (
                  <span className="text-warning">
                    {t("dateRange.rangeTooLong", { count: maxDays })}
                  </span>
                ) : null}
              </div>
            ) : null}
          </div>
        </PopoverContent>
      </Popover>
      {hasValue && !disabled ? (
        <Button
          type="button"
          variant="ghost"
          size="icon-xs"
          aria-label={t("dateRange.clear")}
          onClick={reset}
        >
          <X aria-hidden />
        </Button>
      ) : null}
    </div>
  )
}
