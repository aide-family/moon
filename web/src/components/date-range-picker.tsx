import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  X,
} from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/class-names"
import {
  addDays,
  addMonths,
  dayNumber,
  formatDateValue,
  monthDays,
  parseDateValue,
  startOfMonth,
} from "@/lib/date-range"

export interface DateRange {
  from: string
  to: string
}

interface DateRangePickerProps {
  from: string
  to: string
  onChange: (range: DateRange) => void
  /** 单次可选的最大天数（含首尾），默认 30。 */
  maxDays?: number
  /** 最晚可选的自然日，默认今天；未来日期不允许选择。 */
  maxDate?: string
  className?: string
  placeholder?: string
  ariaLabel?: string
  describedBy?: string
  invalid?: boolean
  disabled?: boolean
  id?: string
}

/**
 * 双月日历的日期范围选择器：按天选择，不产生时分秒。
 * 选满 30 天后更远的日期直接禁用，避免出现超长统计区间。
 */
export function DateRangePicker({
  from,
  to,
  onChange,
  maxDays = 30,
  maxDate,
  className,
  placeholder,
  ariaLabel,
  describedBy,
  invalid = false,
  disabled = false,
  id,
}: DateRangePickerProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith("zh") ? "zh-CN" : "en-US"
  const [open, setOpen] = useState(false)
  const [draftFrom, setDraftFrom] = useState<Date | null>(null)
  const [draftTo, setDraftTo] = useState<Date | null>(null)
  const [hovered, setHovered] = useState<Date | null>(null)
  const [focused, setFocused] = useState<Date | null>(null)
  const [viewMonth, setViewMonth] = useState<Date>(() =>
    startOfMonth(parseDateValue(from) ?? new Date())
  )
  // 内容在 Portal 里延迟挂载，用状态接住节点，节点到位后焦点才能落到日期上。
  const [contentEl, setContentEl] = useState<HTMLDivElement | null>(null)

  const selectedFrom = parseDateValue(from)
  const selectedTo = parseDateValue(to)
  const hasValue = Boolean(from || to)
  const spanDays =
    selectedFrom && selectedTo
      ? Math.abs(dayNumber(selectedTo) - dayNumber(selectedFrom)) + 1
      : 0
  const rangeTooLong = spanDays > maxDays

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
    const start = parseDateValue(from)
    const end = parseDateValue(to)
    setDraftFrom(start ?? null)
    setDraftTo(start && end ? end : null)
    setHovered(null)
    const base = start ?? new Date()
    setViewMonth(startOfMonth(base))
    setFocused(start ?? base)
  }

  // 默认最多选到今天：历史告警和消息日志都不存在未来数据。
  const latestDay = dayNumber(parseDateValue(maxDate ?? "") ?? new Date())

  function isDayDisabled(day: Date) {
    if (dayNumber(day) > latestDay) {
      return true
    }
    if (draftFrom && !draftTo) {
      return Math.abs(dayNumber(day) - dayNumber(draftFrom)) + 1 > maxDays
    }
    return false
  }

  function pickDay(day: Date) {
    if (isDayDisabled(day)) {
      return
    }
    setFocused(day)
    if (!draftFrom || draftTo) {
      setDraftFrom(day)
      setDraftTo(null)
      setHovered(null)
      return
    }
    const start = dayNumber(day) < dayNumber(draftFrom) ? day : draftFrom
    const end = dayNumber(day) < dayNumber(draftFrom) ? draftFrom : day
    setDraftFrom(start)
    setDraftTo(end)
    onChange({ from: formatDateValue(start), to: formatDateValue(end) })
    setOpen(false)
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

  function handleGridKeyDown(event: React.KeyboardEvent<HTMLDivElement>) {
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

  function shiftMonth(months: number) {
    const next = addMonths(viewMonth, months)
    setViewMonth(next)
    // 翻月后让新月份的第一天可聚焦，否则键盘用户会失去落点。
    setFocused(next)
  }

  function renderMonth(month: Date, align: "left" | "right") {
    const days = monthDays(month)
    return (
      <div className="w-[17rem]" key={month.toISOString()}>
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
                  // 超范围的日期不参与预览，否则会画出超过 30 天的区间。
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
        "flex h-8 w-full items-center gap-0.5 rounded-lg border border-input bg-background px-0.5 focus-within:border-ring focus-within:ring-3 focus-within:ring-ring/50 dark:bg-input/30",
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
            aria-label={ariaLabel ?? t("dateRange.open")}
            aria-invalid={invalid || undefined}
            aria-describedby={describedBy}
            className="flex min-w-0 flex-1 items-center gap-1.5 rounded-md px-1 py-1 text-left text-sm outline-none"
          >
            {hasValue ? (
              <span className="min-w-0 truncate font-mono tabular-nums">
                {from || "…"}
                <span className="px-1 text-muted-foreground">→</span>
                {to || "…"}
              </span>
            ) : (
              <span className="truncate text-muted-foreground">
                {placeholder ?? t("dateRange.placeholder")}
              </span>
            )}
          </button>
        </PopoverTrigger>
        <PopoverContent
          ref={setContentEl}
          align="start"
          sideOffset={6}
          className="max-h-[calc(100vh-4rem)] w-auto overflow-y-auto p-3"
          // 焦点交给下面的日期网格；否则原语会把焦点放在第一个按钮（翻月键）上。
          onOpenAutoFocus={(event) => event.preventDefault()}
        >
          <div className="flex flex-col gap-3">
            <div className="grid gap-4 sm:grid-cols-2">
              {renderMonth(viewMonth, "left")}
              {renderMonth(addMonths(viewMonth, 1), "right")}
            </div>
            <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border/60 pt-2 text-xs text-muted-foreground">
              <span>{t("dateRange.maxDaysHint", { count: maxDays })}</span>
              {rangeTooLong ? (
                <span className="text-amber-600 dark:text-amber-400">
                  {t("dateRange.rangeTooLong", { count: maxDays })}
                </span>
              ) : null}
            </div>
          </div>
        </PopoverContent>
      </Popover>
      {hasValue && !disabled ? (
        <Button
          type="button"
          variant="ghost"
          size="icon-xs"
          aria-label={t("dateRange.clear")}
          onClick={() => onChange({ from: "", to: "" })}
        >
          <X aria-hidden />
        </Button>
      ) : null}
    </div>
  )
}
