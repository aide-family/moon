import {
  useCallback,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type KeyboardEvent,
} from "react"
import { useTranslation } from "react-i18next"

import type {
  NotificationOnCallItem,
  OnCallWindow,
} from "@/api/notification-oncall/types"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { cn } from "@/lib/class-names"

/**
 * 值班时间视图：把「通知组 + 接收时段」画成一周（或单日）的时间覆盖图。
 *
 * 选型理由（为什么不画饼图、环形图或堆叠面积图）：
 * - 时段本质是时间轴上的区间，区间用「位置 + 长度」编码最准确，所以主体是 7×24 的时间格：
 *   横轴是时刻、纵轴是星期，读者只需比较同类格子的位置，不比较角度和面积；
 * - 每个通知组在格子里固定一条横向泳道（纵向位置固定），同色不同纹路，
 *   位置、颜色、纹路三重编码，去掉颜色（色盲、灰度打印）后仍能区分到具体通知组；
 * - 没有被任何通知组覆盖的部分露出「空档底纹」（中性色斜线），和实色泳道强烈对比，
 *   空档不靠颜色也能一眼看出来；
 * - 同时给出「整周空档段」的文字列表与每小时覆盖时长，颜色之外的冗余编码保证
 *   色盲用户、读屏用户和灰度打印都能得到同样的结论。
 *
 * 时段语义与后端 app/alioth/pkg/oncall 一致：
 * - weekdays 归属于开始那一天的星期；start 含该时刻，end 不含该时刻；
 * - start 晚于 end 表示跨午夜，覆盖「开始日的 start→24:00」和「次日 00:00→end」，
 *   次日的星期是否在 weekdays 里不影响这段覆盖；
 * - windows 为空表示全天候接收。
 *
 * 只读组件，不发起请求、不改数据：由列表页把当前页数据传进来即可（见 OnCallTimeViewProps）。
 * 文案走 i18next 的 oncallTimeView.* 词条；词条尚未补进语言包时按当前语言回落到内置
 * 中文/英文字面量，因此可以先接入、后补词条。
 */

const WEEKDAY_KEYS = [
  "MONDAY",
  "TUESDAY",
  "WEDNESDAY",
  "THURSDAY",
  "FRIDAY",
  "SATURDAY",
  "SUNDAY",
] as const

const MINUTES_PER_DAY = 1440
const MINUTES_PER_WEEK = MINUTES_PER_DAY * 7
const HOURS_PER_DAY = 24
/** 单个小时格内最多并排的泳道数，超出部分合并成一条「其他 N 组」。 */
const MAX_LANES = 6
const LANE_HEIGHT = 7
const LANE_BAR_HEIGHT = 6
const DEFAULT_TIMEZONE = "Asia/Shanghai"
const GAP_CHIP_LIMIT = 6

const HOURS = Array.from({ length: HOURS_PER_DAY }, (_, hour) => hour)
const WEEK = Array.from({ length: 7 }, (_, day) => day)

/**
 * 分类色板：Okabe–Ito 色盲友好序列，红绿色盲下仍可区分。
 * 颜色永远和固定泳道位置、纹路一起出现，不单独承载语义。
 */
const GROUP_COLORS = [
  "#0072B2",
  "#E69F00",
  "#009E73",
  "#CC79A7",
  "#56B4E9",
  "#D55E00",
  "#8C6D31",
  "#5B5EA6",
]

interface Pattern {
  image?: string
  size?: string
}

/** 纹路与色板一一对应：实色、45°、135°、竖线、横线、点阵、交叉、宽竖线。 */
const GROUP_PATTERNS: Pattern[] = [
  {},
  {
    image:
      "repeating-linear-gradient(45deg, rgba(255,255,255,0.75) 0 1px, transparent 1px 3px)",
  },
  {
    image:
      "repeating-linear-gradient(-45deg, rgba(255,255,255,0.75) 0 1px, transparent 1px 3px)",
  },
  {
    image:
      "repeating-linear-gradient(90deg, rgba(255,255,255,0.8) 0 1px, transparent 1px 3px)",
  },
  {
    image:
      "repeating-linear-gradient(0deg, rgba(255,255,255,0.8) 0 1px, transparent 1px 3px)",
  },
  {
    image: "radial-gradient(rgba(255,255,255,0.9) 1px, transparent 1.1px)",
    size: "3px 3px",
  },
  {
    image:
      "repeating-linear-gradient(45deg, rgba(255,255,255,0.75) 0 1px, transparent 1px 4px), repeating-linear-gradient(-45deg, rgba(255,255,255,0.75) 0 1px, transparent 1px 4px)",
  },
  {
    image:
      "repeating-linear-gradient(90deg, rgba(255,255,255,0.8) 0 2px, transparent 2px 5px)",
  },
]

/** 空档底纹：中性色 + 斜线，不与任何通知组泳道撞色，也不依赖颜色区分。 */
const GAP_FILL: CSSProperties = {
  backgroundColor: "var(--muted)",
  backgroundImage:
    "repeating-linear-gradient(45deg, color-mix(in srgb, var(--muted-foreground) 32%, transparent) 0 1px, transparent 1px 4px)",
}

/** 超出 MAX_LANES 时合并的泳道：中性深色 + 交叉纹路，明确表示「不是某一个具体的组」。 */
const OTHER_FILL: CSSProperties = {
  backgroundColor: "var(--muted-foreground)",
  backgroundImage:
    "repeating-linear-gradient(45deg, rgba(255,255,255,0.85) 0 1px, transparent 1px 4px), repeating-linear-gradient(-45deg, rgba(255,255,255,0.85) 0 1px, transparent 1px 4px)",
}

function fillStyle(colorIndex: number): CSSProperties {
  const pattern = GROUP_PATTERNS[colorIndex % GROUP_PATTERNS.length]
  return {
    backgroundColor: GROUP_COLORS[colorIndex % GROUP_COLORS.length],
    backgroundImage: pattern.image,
    backgroundSize: pattern.size,
  }
}

/** 周内区间，单位分钟，左闭右开；0 表示周一 00:00，10080 表示周日 24:00。 */
interface Span {
  start: number
  end: number
}

interface Lane {
  /** 通知组标识；全部值班项视图下同一 groupUid 会合并成一条泳道。 */
  key: string
  name: string
  spans: Span[]
  minutes: number
  /** 该组在当前范围内配置的时段文字（空数组表示全天候）。 */
  windows: string[]
  always: boolean
  /** 该组出现在哪些值班项里，用于合并视图的明细说明。 */
  onCallNames: string[]
  colorIndex: number
  /** 合并泳道承载的通知组数量。 */
  other?: number
}

interface Segment {
  lane: number
  /** 相对当前小时格起点的分钟偏移。 */
  from: number
  to: number
}

interface Bucket {
  day: number
  hour: number
  segments: Segment[]
  /** 本小时内没有任何通知组覆盖的时间段；铺满格子高度，横向长度即空档长度。 */
  gaps: Segment[]
  /** 并集覆盖分钟数（0 即整格空档）。 */
  covered: number
  groups: number
}

interface BucketPos {
  day: number
  hour: number
}

/** 解析后端 HH:mm（允许 9:00 这种写法）；非法值返回 null，避免脏数据画出错误的班次。 */
function parseClock(value: string): number | null {
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

/** 合并重叠或首尾相接的区间，保证同一组的泳道不会画出双线。 */
function mergeSpans(spans: Span[]): Span[] {
  const sorted = spans
    .filter((span) => span.end > span.start)
    .map((span) => ({ ...span }))
    .sort((left, right) => left.start - right.start || left.end - right.end)
  const merged: Span[] = []
  for (const span of sorted) {
    const last = merged[merged.length - 1]
    if (last && span.start <= last.end) {
      last.end = Math.max(last.end, span.end)
    } else {
      merged.push(span)
    }
  }
  return merged
}

/**
 * 把一个通知组的接收时段展开成周内区间。
 * 跨午夜：weekdays 归属开始那天，区间跨到次日 end；越过周日 24:00 的部分绕回周一，
 * 因为排班是按周重复的（下周一 00:00–end 由本周日开始的那段覆盖）。
 */
function windowsToSpans(windows: OnCallWindow[]): Span[] {
  const spans: Span[] = []
  for (const window of windows) {
    const start = parseClock(window.start)
    const end = parseClock(window.end)
    if (start === null || end === null || start === end) {
      continue
    }
    const days = (window.weekdays ?? [])
      .map((day) =>
        WEEKDAY_KEYS.indexOf(day.toUpperCase() as (typeof WEEKDAY_KEYS)[number])
      )
      .filter((index) => index >= 0)
    if (days.length === 0) {
      continue
    }
    for (const day of days) {
      const from = day * MINUTES_PER_DAY + start
      const to =
        start < end
          ? day * MINUTES_PER_DAY + end
          : day * MINUTES_PER_DAY + MINUTES_PER_DAY + end
      if (to <= MINUTES_PER_WEEK) {
        spans.push({ start: from, end: to })
      } else {
        spans.push({ start: from, end: MINUTES_PER_WEEK })
        spans.push({ start: 0, end: to - MINUTES_PER_WEEK })
      }
    }
  }
  return mergeSpans(spans)
}

/** 求补集：base 里没有被 covered 覆盖的部分，即空档。 */
function subtractSpans(base: Span[], covered: Span[]): Span[] {
  const result: Span[] = []
  for (const span of base) {
    let cursor = span.start
    for (const cover of covered) {
      if (cover.end <= cursor) {
        continue
      }
      if (cover.start >= span.end) {
        break
      }
      if (cover.start > cursor) {
        result.push({ start: cursor, end: Math.min(cover.start, span.end) })
      }
      cursor = Math.max(cursor, cover.end)
      if (cursor >= span.end) {
        break
      }
    }
    if (cursor < span.end) {
      result.push({ start: cursor, end: span.end })
    }
  }
  return result
}

function intersectSpans(spans: Span[], from: number, to: number): Span[] {
  const result: Span[] = []
  for (const span of spans) {
    const start = Math.max(span.start, from)
    const end = Math.min(span.end, to)
    if (end > start) {
      result.push({ start, end })
    }
  }
  return result
}

/** 把一天的小时格展开成渲染用的桶，避免在渲染函数里重复做区间运算。 */
function buildBuckets(
  days: number[],
  lanes: Span[][],
  union: Span[]
): Bucket[] {
  const buckets: Bucket[] = []
  for (const day of days) {
    for (const hour of HOURS) {
      const from = day * MINUTES_PER_DAY + hour * 60
      const to = from + 60
      const segments: Segment[] = []
      lanes.forEach((spans, lane) => {
        for (const span of intersectSpans(spans, from, to)) {
          segments.push({ lane, from: span.start - from, to: span.end - from })
        }
      })
      const coveredSpans = intersectSpans(union, from, to)
      const gaps: Segment[] = []
      let cursor = 0
      for (const span of coveredSpans) {
        const start = span.start - from
        const end = span.end - from
        if (start > cursor) {
          gaps.push({ lane: -1, from: cursor, to: start })
        }
        cursor = Math.max(cursor, end)
      }
      if (cursor < 60) {
        gaps.push({ lane: -1, from: cursor, to: 60 })
      }
      const covered = coveredSpans.reduce(
        (total, span) => total + (span.end - span.start),
        0
      )
      buckets.push({
        day,
        hour,
        segments,
        gaps,
        covered,
        groups: segments.length,
      })
    }
  }
  return buckets
}

/** HH:mm，24:00 表示当天结束；用于格子和明细的时间文案。 */
function formatClock(minutes: number): string {
  const total = Math.max(0, Math.min(minutes, MINUTES_PER_DAY))
  const hour = Math.floor(total / 60)
  const minute = total % 60
  return `${String(hour).padStart(2, "0")}:${String(minute).padStart(2, "0")}`
}

/** 小时数保留一位小数，用于图例、行标注这类窄空间。 */
function formatHours(minutes: number): string {
  return `${Math.round((minutes / 60) * 10) / 10}h`
}

/** 星期几 + 时钟，跨天时写成「周一 22:00 → 周二 06:00」。 */
function formatSpan(span: Span, weekdayLabel: (day: number) => string): string {
  const startDay = Math.floor(span.start / MINUTES_PER_DAY)
  const endEdge = Math.min(span.end, MINUTES_PER_WEEK)
  const endDay = Math.floor((endEdge - 1) / MINUTES_PER_DAY)
  const startClock = formatClock(span.start % MINUTES_PER_DAY)
  const restOfEnd = endEdge % MINUTES_PER_DAY
  const endClock = restOfEnd === 0 ? "24:00" : formatClock(restOfEnd)
  if (startDay === endDay) {
    return `${weekdayLabel(startDay)} ${startClock}–${endClock}`
  }
  return `${weekdayLabel(startDay)} ${startClock} → ${weekdayLabel(endDay)} ${endClock}`
}

function describeWindow(
  window: OnCallWindow,
  weekdayLabel: (key: string) => string,
  anyDay: string
): string {
  const start = (window.start ?? "").trim()
  const end = (window.end ?? "").trim()
  if (!start || !end) {
    return ""
  }
  const days = (window.weekdays ?? [])
    .map((day) => weekdayLabel(day))
    .filter((label) => label !== "")
  return `${days.length > 0 ? days.join("、") : anyDay} ${start}–${end}`
}

/** 在网格里按扁平序号移动一格，越界时夹住而不是绕回，避免方向键突然跳行。 */
function stepBucket(bucket: BucketPos, delta: number): BucketPos {
  const last = MINUTES_PER_WEEK / 60 - 1
  const flat = Math.min(
    Math.max(bucket.day * HOURS_PER_DAY + bucket.hour + delta, 0),
    last
  )
  return {
    day: Math.floor(flat / HOURS_PER_DAY),
    hour: flat % HOURS_PER_DAY,
  }
}

type OnCallTimeViewCopyKey =
  | "otherGroups"
  | "bucketGap"
  | "bucketCovered"
  | "hourUnit"
  | "minuteUnit"
  | "empty"
  | "rangeLabel"
  | "week"
  | "day"
  | "scopeLabel"
  | "scopeAll"
  | "disabled"
  | "dayLabel"
  | "includeDisabled"
  | "timezone"
  | "timezoneMixed"
  | "coverageSummary"
  | "mixedTimezoneHint"
  | "disabledHint"
  | "gridLabel"
  | "timeAxis"
  | "tooltipGap"
  | "scrollHint"
  | "legendGap"
  | "laneLimit"
  | "detailTitle"
  | "detailGap"
  | "detailCovered"
  | "detailHint"
  | "gapTitle"
  | "noGap"
  | "gapMore"
  | "screenReaderSummary"

/** 时间视图文案：词条在 i18n 的 oncallTimeView 下，插值参数由调用方传入。 */
function useCopy() {
  const { t } = useTranslation()
  return useCallback(
    (key: OnCallTimeViewCopyKey, params?: Record<string, unknown>): string =>
      t(`oncallTimeView.${key}`, params),
    [t]
  )
}

export interface OnCallTimeViewProps {
  /** 列表页当前的一页数据；传入后组件自己负责聚合与切换。 */
  items: NotificationOnCallItem[]
  loading?: boolean
  emptyText?: string
  className?: string
}

/**
 * 用法：`<OnCallTimeView items={items} loading={loading} emptyText={t("oncall.empty")} />`
 * 想默认聚焦某一个值班项时，把该单项单独渲染，或先按 uid 过滤 items 再传入。
 */
export function OnCallTimeView({
  items,
  loading = false,
  emptyText,
  className,
}: OnCallTimeViewProps) {
  const { t, i18n } = useTranslation()
  const copy = useCopy()
  // 「多个时区」提示里的分隔符跟当前语言走，避免英文界面出现中文顿号。
  const zoneSeparator = (i18n.resolvedLanguage ?? i18n.language ?? "zh")
    .toLowerCase()
    .startsWith("zh")
    ? "、"
    : ", "
  const [scopeUid, setScopeUid] = useState("ALL")
  const [range, setRange] = useState<"week" | "day">("week")
  const [day, setDay] = useState(0)
  const [hidden, setHidden] = useState<string[]>([])
  const [includeDisabled, setIncludeDisabled] = useState(false)
  const [selected, setSelected] = useState<BucketPos | null>(null)
  const [tabstop, setTabstop] = useState<BucketPos>({ day: 0, hour: 0 })
  const [hover, setHover] = useState<
    (BucketPos & { x: number; y: number }) | null
  >(null)
  // 只存 DOM 引用，供方向键移动焦点；不在渲染期间读写。
  const cellRefs = useRef<Record<string, HTMLButtonElement | null>>({})
  const gridRef = useRef<HTMLDivElement | null>(null)

  const weekdayName = useCallback(
    (index: number) => t(`oncall.weekday.${WEEKDAY_KEYS[index]}`),
    [t]
  )
  const weekdayNameByKey = useCallback(
    (key: string) => t(`oncall.weekday.${key}`),
    [t]
  )

  const disabledCount = items.filter((item) => item.status !== "ACTIVE").length
  // 选中项被删除或换页时自动回到「全部」，避免出现空视图。
  const scope = items.some((item) => item.uid === scopeUid) ? scopeUid : "ALL"

  const scopeItems = useMemo(() => {
    if (scope === "ALL") {
      // 停用的值班项不会投递通知，默认不计入有效覆盖（可由用户显式打开）。
      return items.filter((item) => includeDisabled || item.status === "ACTIVE")
    }
    return items.filter((item) => item.uid === scope)
  }, [items, scope, includeDisabled])

  const timezones = useMemo(
    () =>
      Array.from(
        new Set(
          scopeItems.map((item) => (item.timezone || DEFAULT_TIMEZONE).trim())
        )
      ),
    [scopeItems]
  )

  const lanes = useMemo<Lane[]>(() => {
    const collected = new Map<
      string,
      {
        name: string
        spans: Span[]
        windows: string[]
        always: boolean
        onCallNames: string[]
      }
    >()
    for (const item of scopeItems) {
      for (const group of item.groups ?? []) {
        const key =
          (group.groupUid ?? "").trim() || (group.groupName ?? "").trim()
        if (!key) {
          continue
        }
        let entry = collected.get(key)
        if (!entry) {
          entry = {
            name: (group.groupName ?? "").trim() || key,
            spans: [],
            windows: [],
            always: false,
            onCallNames: [],
          }
          collected.set(key, entry)
        }
        if (item.name && !entry.onCallNames.includes(item.name)) {
          entry.onCallNames.push(item.name)
        }
        const windows = group.windows ?? []
        if (windows.length === 0) {
          entry.always = true
          continue
        }
        entry.spans.push(...windowsToSpans(windows))
        for (const window of windows) {
          const text = describeWindow(
            window,
            weekdayNameByKey,
            t("oncall.anyDay")
          )
          if (text && !entry.windows.includes(text)) {
            entry.windows.push(text)
          }
        }
      }
    }
    return Array.from(collected.entries())
      .map(([key, entry]) => {
        const spans = entry.always
          ? [{ start: 0, end: MINUTES_PER_WEEK }]
          : mergeSpans(entry.spans)
        return {
          key,
          name: entry.name,
          spans,
          minutes: spans.reduce(
            (total, span) => total + (span.end - span.start),
            0
          ),
          windows: entry.windows,
          always: entry.always,
          onCallNames: entry.onCallNames,
          colorIndex: 0,
        }
      })
      .sort((left, right) => left.name.localeCompare(right.name, "zh-Hans"))
      .map((lane, index) => ({ ...lane, colorIndex: index }))
  }, [scopeItems, t, weekdayNameByKey])

  const visibleLanes = useMemo(
    () => lanes.filter((lane) => !hidden.includes(lane.key)),
    [lanes, hidden]
  )

  // 泳道超过上限时，保留覆盖时长最长的几组，剩下的合并成一条中性色的「其他 N 组」，
  // 避免无限压缩格子。保留最大的几组而不是前几组：全天候/大段覆盖的组被折叠掉，
  // 会让排班图看着几乎没人值班，正好丢掉最该看到的信息。
  const gridLanes = useMemo<Lane[]>(() => {
    if (visibleLanes.length <= MAX_LANES) {
      return visibleLanes
    }
    const ranked = [...visibleLanes].sort(
      (left, right) =>
        right.minutes - left.minutes ||
        left.name.localeCompare(right.name, "zh-Hans")
    )
    const kept = new Set(ranked.slice(0, MAX_LANES - 1).map((lane) => lane.key))
    // 仍然按名称顺序排泳道，保证同一组在整周内始终占同一条横向泳道。
    const head = visibleLanes.filter((lane) => kept.has(lane.key))
    const rest = visibleLanes.filter((lane) => !kept.has(lane.key))
    return [
      ...head,
      {
        key: "__other__",
        name: copy("otherGroups", { count: rest.length }),
        spans: mergeSpans(rest.flatMap((lane) => lane.spans)),
        minutes: 0,
        windows: [],
        always: false,
        onCallNames: [],
        colorIndex: -1,
        other: rest.length,
      },
    ]
  }, [visibleLanes, copy])

  const union = useMemo(
    () => mergeSpans(visibleLanes.flatMap((lane) => lane.spans)),
    [visibleLanes]
  )
  const gaps = useMemo(
    () => subtractSpans([{ start: 0, end: MINUTES_PER_WEEK }], union),
    [union]
  )
  const coveredMinutes = useMemo(
    () => union.reduce((total, span) => total + (span.end - span.start), 0),
    [union]
  )
  const gapMinutes = MINUTES_PER_WEEK - coveredMinutes

  const dailyCoverage = useMemo(
    () =>
      WEEK.map((weekday) =>
        intersectSpans(
          union,
          weekday * MINUTES_PER_DAY,
          (weekday + 1) * MINUTES_PER_DAY
        ).reduce((total, span) => total + (span.end - span.start), 0)
      ),
    [union]
  )

  const days = useMemo(() => (range === "week" ? WEEK : [day]), [range, day])
  const buckets = useMemo(
    () =>
      buildBuckets(
        days,
        gridLanes.map((lane) => lane.spans),
        union
      ),
    [days, gridLanes, union]
  )
  const bucketMap = useMemo(
    () =>
      new Map(
        buckets.map((bucket) => [`${bucket.day}:${bucket.hour}`, bucket])
      ),
    [buckets]
  )

  const activeTabstop = days.includes(tabstop.day)
    ? tabstop
    : { day: days[0], hour: 0 }
  const activeBucket = selected
    ? (bucketMap.get(`${selected.day}:${selected.hour}`) ?? null)
    : null

  const cellHeight =
    Math.max(24, gridLanes.length * LANE_HEIGHT + 4) +
    (range === "day" ? 10 : 0)
  const laneTop = (lane: number) =>
    (cellHeight - gridLanes.length * LANE_HEIGHT) / 2 + lane * LANE_HEIGHT

  function focusCell(bucket: BucketPos) {
    cellRefs.current[`${bucket.day}:${bucket.hour}`]?.focus()
  }

  function selectBucket(bucket: BucketPos) {
    setSelected(bucket)
    setTabstop(bucket)
  }

  /** 切换日视图/星期时把选中和 tab 焦点带到当前可见的那一行，避免明细面板突然空掉。 */
  function changeRange(next: "week" | "day") {
    setRange(next)
    if (next === "day" && selected) {
      setDay(selected.day)
    }
  }

  function changeDay(index: number) {
    const hour = selected?.hour ?? tabstop.hour
    setDay(index)
    setTabstop({ day: index, hour })
    if (selected) {
      setSelected({ day: index, hour })
    }
  }

  /** 从空档列表跳到对应格子：日视图下先切到那一天，否则格子不可见。 */
  function revealBucket(bucket: BucketPos) {
    if (range === "day") {
      setDay(bucket.day)
    }
    selectBucket(bucket)
  }

  function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const anchor = selected ?? activeTabstop
    const week = range === "week"
    let next: BucketPos
    switch (event.key) {
      case "ArrowRight":
        next = week
          ? stepBucket(anchor, 1)
          : {
              day: anchor.day,
              hour: Math.min(HOURS_PER_DAY - 1, anchor.hour + 1),
            }
        break
      case "ArrowLeft":
        next = week
          ? stepBucket(anchor, -1)
          : { day: anchor.day, hour: Math.max(0, anchor.hour - 1) }
        break
      case "ArrowUp":
        next = week
          ? { day: Math.max(0, anchor.day - 1), hour: anchor.hour }
          : anchor
        break
      case "ArrowDown":
        next = week
          ? { day: Math.min(6, anchor.day + 1), hour: anchor.hour }
          : anchor
        break
      case "Home":
        next = { day: anchor.day, hour: 0 }
        break
      case "End":
        next = { day: anchor.day, hour: HOURS_PER_DAY - 1 }
        break
      default:
        return
    }
    event.preventDefault()
    // 键盘移动时收起鼠标 hover 提示，避免两个高亮互相矛盾。
    setHover(null)
    focusCell(next)
  }

  function toggleLane(key: string) {
    setHidden((current) =>
      current.includes(key)
        ? current.filter((value) => value !== key)
        : [...current, key]
    )
  }

  function showHover(
    event: React.MouseEvent<HTMLButtonElement>,
    bucket: BucketPos
  ) {
    const host = gridRef.current?.getBoundingClientRect()
    if (!host) {
      return
    }
    const rect = event.currentTarget.getBoundingClientRect()
    const width = 208
    const x = Math.min(
      Math.max(rect.left - host.left + rect.width / 2 - width / 2, 0),
      Math.max(host.width - width, 0)
    )
    setHover({ ...bucket, x, y: rect.bottom - host.top + 6 })
  }

  function laneFill(lane: Lane): CSSProperties {
    return lane.other ? OTHER_FILL : fillStyle(lane.colorIndex)
  }

  function bucketLabel(bucket: Bucket): string {
    const weekday = weekdayName(bucket.day)
    const clock = `${formatClock(bucket.hour * 60)}–${formatClock(bucket.hour * 60 + 60)}`
    if (bucket.groups === 0) {
      return copy("bucketGap", { weekday, clock })
    }
    const names = bucket.segments
      .map((segment) => gridLanes[segment.lane]?.name)
      .filter(Boolean)
      .join("、")
    return copy("bucketCovered", { weekday, clock, count: bucket.groups, names })
  }

  function duration(minutes: number): string {
    const hours = Math.floor(minutes / 60)
    const rest = minutes % 60
    const hourUnit = copy("hourUnit")
    const minuteUnit = copy("minuteUnit")
    if (hours > 0 && rest > 0) {
      return `${hours} ${hourUnit} ${rest} ${minuteUnit}`
    }
    if (hours > 0) {
      return `${hours} ${hourUnit}`
    }
    return `${rest} ${minuteUnit}`
  }

  const cardClasses = cn(
    "flex min-h-0 min-w-0 flex-col gap-3 rounded-xl border bg-card p-3 sm:p-4",
    className
  )

  if (loading) {
    return (
      <div className={cardClasses}>
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className={cardClasses}>
        <p className="text-sm text-muted-foreground">
          {emptyText ??
            copy("empty")}
        </p>
      </div>
    )
  }

  const gridTemplate = "3.5rem repeat(24, minmax(0, 1fr))"
  const hoverBucket = hover
    ? bucketMap.get(`${hover.day}:${hover.hour}`)
    : undefined
  const sortedGaps = [...gaps].sort(
    (left, right) => right.end - right.start - (left.end - left.start)
  )
  const singleDisabled =
    scope !== "ALL" && scopeItems.every((item) => item.status !== "ACTIVE")

  return (
    <div className={cardClasses}>
      {/* 工具栏：先定范围，再定值班项；图例放在图下面，避免遮挡数据。 */}
      <div className="flex flex-wrap items-center gap-2">
        <div
          role="group"
          aria-label={copy("rangeLabel")}
          className="inline-flex items-center gap-0.5 rounded-lg border bg-muted/40 p-0.5"
        >
          <Button
            size="xs"
            variant={range === "week" ? "secondary" : "ghost"}
            aria-pressed={range === "week"}
            onClick={() => changeRange("week")}
          >
            {copy("week")}
          </Button>
          <Button
            size="xs"
            variant={range === "day" ? "secondary" : "ghost"}
            aria-pressed={range === "day"}
            onClick={() => changeRange("day")}
          >
            {copy("day")}
          </Button>
        </div>

        <Select value={scope} onValueChange={setScopeUid}>
          <SelectTrigger
            size="sm"
            className="min-w-40"
            aria-label={copy("scopeLabel")}
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="ALL">
              {copy("scopeAll")}
            </SelectItem>
            {items.map((item) => (
              <SelectItem key={item.uid} value={item.uid}>
                {item.status === "ACTIVE"
                  ? item.name
                  : `${item.name}（${copy("disabled")}）`}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {range === "day" ? (
          <div
            role="group"
            aria-label={copy("dayLabel")}
            className="inline-flex flex-wrap items-center gap-0.5 rounded-lg border bg-muted/40 p-0.5"
          >
            {WEEK.map((index) => (
              <Button
                key={index}
                size="xs"
                variant={day === index ? "secondary" : "ghost"}
                aria-pressed={day === index}
                onClick={() => changeDay(index)}
              >
                {weekdayName(index)}
              </Button>
            ))}
          </div>
        ) : null}

        {scope === "ALL" && disabledCount > 0 ? (
          <Button
            size="xs"
            variant={includeDisabled ? "secondary" : "outline"}
            aria-pressed={includeDisabled}
            onClick={() => setIncludeDisabled((current) => !current)}
          >
            {copy("includeDisabled", { count: disabledCount })}
          </Button>
        ) : null}

        <div className="ml-auto flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
          <span>
            {timezones.length === 1
              ? copy("timezone", { zone: timezones[0] })
              : copy("timezoneMixed")}
          </span>
          <span className="font-mono">
            {copy("coverageSummary", { covered: formatHours(coveredMinutes) })}
          </span>
        </div>
      </div>

      {timezones.length > 1 ? (
        <p className="rounded-md border border-amber-500/30 bg-amber-500/10 px-2.5 py-1.5 text-xs text-amber-700 dark:text-amber-300">
          {copy("mixedTimezoneHint", { zones: timezones.join(zoneSeparator) })}
        </p>
      ) : null}

      {singleDisabled ? (
        <p className="rounded-md border border-amber-500/30 bg-amber-500/10 px-2.5 py-1.5 text-xs text-amber-700 dark:text-amber-300">
          {copy("disabledHint")}
        </p>
      ) : null}

      <div className="grid min-w-0 gap-3 xl:grid-cols-[minmax(0,1fr)_19rem]">
        <div className="min-w-0">
          <div
            ref={gridRef}
            className="relative min-w-0 rounded-lg border bg-muted/10"
          >
            <div className="overflow-x-auto">
              <div className="min-w-[44rem] p-2">
                <div
                  role="grid"
                  aria-label={copy("gridLabel")}
                  aria-readonly="true"
                  onKeyDown={handleKeyDown}
                  onMouseLeave={() => setHover(null)}
                >
                  {/* 小时轴 */}
                  <div
                    role="row"
                    className="grid items-end gap-x-[2px]"
                    style={{ gridTemplateColumns: gridTemplate }}
                  >
                    <div
                      role="columnheader"
                      aria-label={copy("timeAxis")}
                      className="pb-1 text-[10px] text-muted-foreground"
                    >
                      {copy("timeAxis")}
                    </div>
                    {HOURS.map((hour) => (
                      <div
                        key={hour}
                        role="columnheader"
                        aria-label={formatClock(hour * 60)}
                        className="pb-1 text-center text-[10px] leading-none text-muted-foreground tabular-nums"
                      >
                        {range === "week" && hour % 3 !== 0
                          ? ""
                          : String(hour).padStart(2, "0")}
                      </div>
                    ))}
                  </div>

                  {days.map((weekday) => (
                    <div
                      key={weekday}
                      role="row"
                      className="mt-[2px] grid items-center gap-x-[2px]"
                      style={{ gridTemplateColumns: gridTemplate }}
                    >
                      <div role="rowheader" className="min-w-0 pr-1.5">
                        <span className="block truncate text-xs font-medium">
                          {weekdayName(weekday)}
                        </span>
                        <span
                          className={cn(
                            "block font-mono text-[10px] tabular-nums",
                            dailyCoverage[weekday] === 0
                              ? "font-medium text-destructive"
                              : "text-muted-foreground"
                          )}
                        >
                          {formatHours(dailyCoverage[weekday])}
                        </span>
                      </div>
                      {HOURS.map((hour) => {
                        const bucket = bucketMap.get(`${weekday}:${hour}`)
                        if (!bucket) {
                          return null
                        }
                        const key = `${weekday}:${hour}`
                        const isSelected =
                          selected?.day === weekday && selected?.hour === hour
                        const isTabstop =
                          activeTabstop.day === weekday &&
                          activeTabstop.hour === hour
                        return (
                          <button
                            key={key}
                            ref={(node) => {
                              cellRefs.current[key] = node
                            }}
                            type="button"
                            role="gridcell"
                            tabIndex={isTabstop ? 0 : -1}
                            aria-selected={isSelected}
                            aria-label={bucketLabel(bucket)}
                            data-oncall-hour={key}
                            onClick={() => selectBucket({ day: weekday, hour })}
                            onFocus={() => {
                              setTabstop({ day: weekday, hour })
                              setSelected({ day: weekday, hour })
                            }}
                            onMouseEnter={(event) =>
                              showHover(event, { day: weekday, hour })
                            }
                            className={cn(
                              "relative rounded-[3px] transition-shadow outline-none",
                              "hover:brightness-110 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 focus-visible:ring-offset-background",
                              isSelected &&
                                "ring-2 ring-primary ring-offset-1 ring-offset-background"
                            )}
                            style={{ height: cellHeight }}
                          >
                            {/* 空档只铺在「没有覆盖」的时间段上：横向长度就是空档时长，
                                和有覆盖的泳道共用一条时间轴，不会被误读成「整格都空」。 */}
                            <span
                              aria-hidden="true"
                              className="absolute inset-0 rounded-[3px] bg-muted/50"
                            />
                            {bucket.gaps.map((gap, index) => (
                              <span
                                key={`gap-${index}`}
                                aria-hidden="true"
                                className="absolute inset-y-0 rounded-[2px]"
                                style={{
                                  left: `${(gap.from / 60) * 100}%`,
                                  width: `${((gap.to - gap.from) / 60) * 100}%`,
                                  ...GAP_FILL,
                                }}
                              />
                            ))}
                            {bucket.segments.map((segment, index) => {
                              const lane = gridLanes[segment.lane]
                              if (!lane) {
                                return null
                              }
                              return (
                                <span
                                  key={`${segment.lane}-${index}`}
                                  aria-hidden="true"
                                  className="absolute rounded-[2px]"
                                  style={{
                                    left: `${(segment.from / 60) * 100}%`,
                                    width: `${((segment.to - segment.from) / 60) * 100}%`,
                                    minWidth: 2,
                                    top: laneTop(segment.lane),
                                    height: LANE_BAR_HEIGHT,
                                    ...laneFill(lane),
                                  }}
                                />
                              )
                            })}
                          </button>
                        )
                      })}
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* hover 提示：补充说明，不承载唯一信息（明细面板和 aria-label 都有同样内容）。 */}
            {hover && hoverBucket ? (
              <div
                aria-hidden="true"
                className="pointer-events-none absolute z-20 w-52 rounded-md border bg-popover px-2.5 py-2 text-xs shadow-md"
                style={{ left: hover.x, top: hover.y }}
              >
                <p className="font-medium tabular-nums">
                  {weekdayName(hover.day)} {formatClock(hover.hour * 60)}–
                  {formatClock(hover.hour * 60 + 60)}
                </p>
                {hoverBucket.groups === 0 ? (
                  <p className="mt-1 text-muted-foreground">
                    {copy("tooltipGap")}
                  </p>
                ) : (
                  <ul className="mt-1 flex flex-col gap-1">
                    {hoverBucket.segments.map((segment, index) => {
                      const lane = gridLanes[segment.lane]
                      if (!lane) {
                        return null
                      }
                      return (
                        <li
                          key={`${segment.lane}-${index}`}
                          className="flex min-w-0 items-center gap-1.5"
                        >
                          <span
                            aria-hidden="true"
                            className="size-2.5 shrink-0 rounded-[2px]"
                            style={laneFill(lane)}
                          />
                          <span className="min-w-0 flex-1 truncate">
                            {lane.name}
                          </span>
                        </li>
                      )
                    })}
                  </ul>
                )}
              </div>
            ) : null}
          </div>

          {/* 窄屏上 24 列放不下，告知可以横向滑动，避免误以为只画了半天。 */}
          <p className="mt-1 text-[10px] text-muted-foreground lg:hidden">
            {copy("scrollHint")}
          </p>

          {/* 图例：颜色 + 纹路 + 名称 + 本组覆盖时长，点按可隐藏/显示某组（剔除后的并集就是新空档）。 */}
          <div className="mt-2 flex flex-wrap items-center gap-1.5">
            <span className="flex items-center gap-1.5 rounded-md border bg-background px-2 py-1 text-xs">
              <span
                aria-hidden="true"
                className="size-3 shrink-0 rounded-[3px]"
                style={GAP_FILL}
              />
              {copy("legendGap")}
            </span>
            {lanes.map((lane) => {
              const on = !hidden.includes(lane.key)
              return (
                <button
                  key={lane.key}
                  type="button"
                  aria-pressed={on}
                  onClick={() => toggleLane(lane.key)}
                  className={cn(
                    "flex max-w-full items-center gap-1.5 rounded-md border bg-background px-2 py-1 text-xs transition-colors hover:bg-muted/60 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
                    !on && "opacity-40"
                  )}
                >
                  <span
                    aria-hidden="true"
                    className="size-3 shrink-0 rounded-[3px]"
                    style={laneFill(lane)}
                  />
                  <span className="max-w-36 truncate">{lane.name}</span>
                  {lane.always ? (
                    <span className="shrink-0 rounded-sm bg-muted px-1 text-[10px] text-muted-foreground">
                      {t("oncall.always")}
                    </span>
                  ) : (
                    <span className="shrink-0 font-mono text-[10px] text-muted-foreground tabular-nums">
                      {formatHours(lane.minutes)}
                    </span>
                  )}
                </button>
              )
            })}
            {gridLanes.some((lane) => lane.other) ? (
              <span className="text-[10px] text-muted-foreground">
                {copy("laneLimit", { lanes: MAX_LANES })}
              </span>
            ) : null}
          </div>
        </div>

        <aside className="flex min-w-0 flex-col gap-2 rounded-lg border bg-muted/20 p-2.5">
          <p className="text-xs font-medium text-muted-foreground">
            {copy("detailTitle")}
          </p>
          {/* 只把「选中的时段 + 覆盖情况」放进 live region，
              切换格子时读屏不会连空档列表一起念一遍。 */}
          <div aria-live="polite" className="flex min-w-0 flex-col gap-2">
            {activeBucket ? (
              <>
                <p className="text-sm font-medium tabular-nums">
                  {weekdayName(activeBucket.day)}{" "}
                  {formatClock(activeBucket.hour * 60)}–
                  {formatClock(activeBucket.hour * 60 + 60)}
                </p>
                <p
                  className={cn(
                    "text-xs",
                    activeBucket.groups === 0
                      ? "font-medium text-destructive"
                      : "text-muted-foreground"
                  )}
                >
                  {activeBucket.groups === 0
                    ? copy("detailGap")
                    : copy("detailCovered", { count: activeBucket.groups, duration: duration(activeBucket.covered) })}
                </p>
                <ul className="flex flex-col gap-1.5">
                  {activeBucket.segments.map((segment, index) => {
                    const lane = gridLanes[segment.lane]
                    if (!lane) {
                      return null
                    }
                    return (
                      <li
                        key={`${segment.lane}-${index}`}
                        className="flex min-w-0 flex-col gap-0.5 rounded-md border bg-background px-2 py-1.5"
                      >
                        <div className="flex min-w-0 items-center gap-1.5">
                          <span
                            aria-hidden="true"
                            className="size-3 shrink-0 rounded-[3px]"
                            style={laneFill(lane)}
                          />
                          <span
                            className="min-w-0 flex-1 truncate text-xs font-medium"
                            title={lane.name}
                          >
                            {lane.name}
                          </span>
                          <span className="shrink-0 font-mono text-[10px] text-muted-foreground tabular-nums">
                            {formatClock(activeBucket.hour * 60 + segment.from)}
                            –{formatClock(activeBucket.hour * 60 + segment.to)}
                          </span>
                        </div>
                        {lane.other ? null : (
                          <span className="text-[10px] break-words text-muted-foreground">
                            {lane.always
                              ? t("oncall.always")
                              : lane.windows.length > 0
                                ? lane.windows.join("；")
                                : ""}
                            {scope === "ALL" && lane.onCallNames.length > 0
                              ? ` · ${lane.onCallNames.join("、")}`
                              : ""}
                          </span>
                        )}
                      </li>
                    )
                  })}
                </ul>
              </>
            ) : (
              <p className="text-xs text-muted-foreground">
                {copy("detailHint")}
              </p>
            )}
          </div>

          <div className="mt-1 flex flex-col gap-1 border-t pt-2">
            <p className="text-xs font-medium text-muted-foreground">
              {copy("gapTitle", { count: gaps.length, duration: duration(gapMinutes) })}
            </p>
            {gaps.length === 0 ? (
              <p className="text-[10px] text-muted-foreground">
                {copy("noGap")}
              </p>
            ) : (
              <ul className="flex flex-col gap-1">
                {sortedGaps.slice(0, GAP_CHIP_LIMIT).map((span) => {
                  const bucket = {
                    day: Math.floor(span.start / MINUTES_PER_DAY),
                    hour: Math.floor((span.start % MINUTES_PER_DAY) / 60),
                  }
                  return (
                    <li key={span.start}>
                      <button
                        type="button"
                        onClick={() => revealBucket(bucket)}
                        className="flex w-full min-w-0 items-center gap-2 rounded-md border bg-background px-2 py-1 text-left text-xs hover:bg-muted/60 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
                      >
                        <span className="min-w-0 flex-1 truncate tabular-nums">
                          {formatSpan(span, weekdayName)}
                        </span>
                        <span className="shrink-0 font-mono text-[10px] text-muted-foreground tabular-nums">
                          {formatHours(span.end - span.start)}
                        </span>
                      </button>
                    </li>
                  )
                })}
                {sortedGaps.length > GAP_CHIP_LIMIT ? (
                  <li className="text-[10px] text-muted-foreground">
                    {copy("gapMore", { count: sortedGaps.length - GAP_CHIP_LIMIT })}
                  </li>
                ) : null}
              </ul>
            )}
          </div>
        </aside>
      </div>

      {/* 读屏摘要：把「一周覆盖多少、空档在哪」讲成一句话，避免只能逐个读格子。 */}
      <p className="sr-only">
        {copy("screenReaderSummary", { total: formatHours(MINUTES_PER_WEEK), covered: formatHours(coveredMinutes), count: gaps.length, duration: duration(gapMinutes) })}
      </p>
    </div>
  )
}
