import { BarChart3, List, Search, X } from "lucide-react"
import { useEffect, useLayoutEffect, useRef, useState } from "react"
import { useSearchParams } from "react-router-dom"
import type { ReactNode } from "react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { formErrorMessage } from "@/lib/form-error"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import type { AlertColumnItem } from "@/api/alert-page/types"
import {
  alertEventStats,
  interveneAlertEvent,
  interveneAlertEvents,
  listAlertEvents,
  recoverAlertEvent,
  recoverAlertEvents,
  suppressAlertEvent,
} from "@/api/alert-event"
import type {
  AlertEventItem,
  AlertEventStats,
  AlertEventView,
} from "@/api/alert-event/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { FilterChips, type FilterChip } from "@/components/filter-chips"
import { FilterToggle } from "@/components/filter-toggle"
import { TABLE_COLUMN_WIDTH, TABLE_MIN_WIDTH } from "@/lib/table-layout"
import { LevelBadge, StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { DateRangePicker } from "@/components/date-range-picker"
import { Input } from "@/components/ui/input"
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
import { AlertEventChartDialog } from "@/pages/alert/event-chart"
import { AlertEventDetail } from "@/pages/alert/event-detail"
import { AlertHistoryInsights } from "@/pages/alert/history-insights"
import {
  columnTitle,
  DEFAULT_HISTORY_COLUMNS,
  DEFAULT_REALTIME_COLUMNS,
  parseEventLabels,
} from "@/pages/alert/columns"
import { formatDuration, formatTime } from "@/pages/alert/format"
import { dayEndISO, dayStartISO } from "@/lib/date-range"
import { HISTORY_RANGE_MAX_DAYS } from "@/pages/alert/history-range"
import { AlertEventSuppressDialog } from "@/pages/alert/suppress-dialog"
import type { SuppressInput } from "@/pages/alert/suppress-dialog"

interface AlertEventListProps {
  view: AlertEventView
  mine?: boolean
  pageUid?: string
  hideHeader?: boolean
  // refreshKey 变化时静默刷新列表；由实时页的自动轮询驱动。
  refreshKey?: number
  // 实时页从独立统计接口取数并下发；历史页不展示实时统计。
  stats?: AlertEventStats
  // 页面下发的自定义列；为空时使用视图默认列（实时、历史各一套）。
  columns?: AlertColumnItem[]
  // 列表工具栏右侧的图标操作，实时页和历史明细都用它打开自定义列。
  columnsAction?: ReactNode
}

type ConfirmKind = "recover" | "batch-recover"

type HistoryDetailFilters = {
  name: string
  level: string
  ruleName: string
  datasourceName: string
  from: string
  to: string
  status: string
  suppression: string
}

const HISTORY_DETAIL_QUERY_KEYS = {
  name: "eventName",
  level: "eventLevel",
  ruleName: "eventRule",
  datasourceName: "eventDatasource",
  from: "eventFrom",
  to: "eventTo",
  status: "eventStatus",
  suppression: "eventSuppression",
} as const satisfies Record<keyof HistoryDetailFilters, string>

function activeHistoryFilterCount(filters: HistoryDetailFilters): number {
  return [
    filters.name,
    filters.level,
    filters.ruleName,
    filters.datasourceName,
    filters.from,
    filters.to,
    filters.status,
    filters.suppression,
  ].filter(Boolean).length
}

// 已提交筛选的可移除标签。移除任何一个都回到第 1 页，
// 并且把对应的 URL 参数删掉——筛选状态只有 URL 一处来源。
function historyFilterChips(
  t: TFunction,
  filters: HistoryDetailFilters,
  levelName: string
): Array<{ label: string; keys: string[] }> {
  const chips: Array<{ label: string; keys: string[] }> = []
  const text = (label: string, value: string) =>
    value ? `${label}: ${value}` : ""
  const name = text(t("alertEvent.namePlaceholder"), filters.name)
  if (name) chips.push({ label: name, keys: ["eventName"] })
  const rule = text(t("alertEvent.rule"), filters.ruleName)
  if (rule) chips.push({ label: rule, keys: ["eventRule"] })
  const datasource = text(t("alertEvent.datasource"), filters.datasourceName)
  if (datasource) chips.push({ label: datasource, keys: ["eventDatasource"] })
  if (filters.level) {
    chips.push({
      label: `${t("alertEvent.level")}: ${levelName || filters.level}`,
      keys: ["eventLevel"],
    })
  }
  if (filters.status) {
    const label =
      filters.status === "FIRING"
        ? t("alertEvent.statusFiring")
        : filters.status === "RECOVERED"
          ? t("alertEvent.statusRecovered")
          : t("alertEvent.statusRecoveredManual")
    chips.push({
      label: `${t("alertEvent.eventStatusLabel")}: ${label}`,
      keys: ["eventStatus"],
    })
  }
  if (filters.suppression) {
    const label =
      filters.suppression === "SUPPRESSED"
        ? t("alertEvent.suppressedOnly")
        : t("alertEvent.unsuppressedOnly")
    chips.push({
      label: `${t("alertEvent.suppressionFilter")}: ${label}`,
      keys: ["eventSuppression"],
    })
  }
  if (filters.from || filters.to) {
    chips.push({
      label: `${t("alertEvent.eventDateRange")}: ${filters.from || "…"} ${t(
        "alertEvent.dateTo"
      )} ${filters.to || "…"}`,
      keys: ["eventFrom", "eventTo"],
    })
  }
  return chips
}

export function AlertEventList({
  view,
  mine = false,
  pageUid = "",
  hideHeader = false,
  refreshKey = 0,
  stats: sharedStats,
  columns,
  columnsAction,
}: AlertEventListProps) {
  const { t, i18n } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [params, setParams] = useSearchParams()
  const analysisLevel = params.get("level") || ""
  const historyTab =
    params.get("historyTab") === "analysis" ? "analysis" : "details"
  const committedHistoryFilters: HistoryDetailFilters = {
    name: params.get("eventName") || "",
    level: params.get("eventLevel") || "",
    ruleName: params.get("eventRule") || "",
    datasourceName: params.get("eventDatasource") || "",
    from: params.get("eventFrom") || "",
    to: params.get("eventTo") || "",
    status: params.get("eventStatus") || "",
    suppression: params.get("eventSuppression") || "",
  }
  const historyFilterSignature = JSON.stringify(committedHistoryFilters)
  const draftSignature = `${session.namespaceUid}|${historyFilterSignature}`
  const [historyDraftState, setHistoryDraftState] = useState<{
    signature: string
    filters: HistoryDetailFilters
  } | null>(null)
  const historyDraft =
    historyDraftState?.signature === draftSignature
      ? historyDraftState.filters
      : committedHistoryFilters
  const updateHistoryDraft = <K extends keyof HistoryDetailFilters>(
    key: K,
    value: HistoryDetailFilters[K]
  ) => {
    setHistoryDraftState({
      signature: draftSignature,
      filters: { ...historyDraft, [key]: value },
    })
  }
  // 日期范围是一次选择两个值，单独走一个入口，避免两次 setState 互相覆盖。
  const updateHistoryDraftRange = (range: { from: string; to: string }) => {
    setHistoryDraftState({
      signature: draftSignature,
      filters: { ...historyDraft, from: range.from, to: range.to },
    })
  }
  const invalidHistoryDateRange = Boolean(
    historyDraft.from && historyDraft.to && historyDraft.from > historyDraft.to
  )
  const appliedHistoryFilterCount = activeHistoryFilterCount(
    committedHistoryFilters
  )
  // 筛选面板默认收起：明细列表才是主内容，不能让它常驻占掉表格高度。
  // 首次进入（没有任何已应用筛选）时展开，保证筛选能力可被发现。
  const [filtersOpen, setFiltersOpen] = useState(
    appliedHistoryFilterCount === 0
  )
  // 实时告警同样默认展开，应用后收起；有 URL 条件时直接收起。
  const [realtimeFiltersOpen, setRealtimeFiltersOpen] = useState(
    () => !(query.name || analysisLevel)
  )
  const historyDraftDirty =
    JSON.stringify(historyDraft) !== historyFilterSignature
  const listName =
    view === "HISTORY" ? committedHistoryFilters.name : query.name
  const listLevel =
    view === "HISTORY" ? committedHistoryFilters.level : analysisLevel
  // 实时告警的筛选也是先编辑后应用，和历史告警保持同一种交互。
  // 草稿绑定在“已应用条件”的签名上：URL 变了（后退、清空、外部跳转）就回到 URL 的值。
  const realtimeSignature = `${query.name}|${analysisLevel}`
  const [realtimeDraftState, setRealtimeDraftState] = useState<{
    signature: string
    name: string
    level: string
  } | null>(null)
  const realtimeDraft =
    realtimeDraftState?.signature === realtimeSignature
      ? realtimeDraftState
      : null
  const draft = realtimeDraft?.name ?? query.name
  const draftLevel = realtimeDraft?.level ?? analysisLevel
  const setDraft = (name: string) =>
    setRealtimeDraftState({
      signature: realtimeSignature,
      name,
      level: draftLevel,
    })
  const setDraftLevel = (level: string) =>
    setRealtimeDraftState({ signature: realtimeSignature, name: draft, level })
  const [localStats, setLocalStats] = useState<AlertEventStats>({
    firing: 0,
    suppressed: 0,
    recovered: 0,
  })
  const stats = sharedStats ?? localStats
  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    reload,
  } = usePagedList<AlertEventItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${listName}|${listLevel}|${historyFilterSignature}|${view}|${mine}|${pageUid}`,
    refreshKey,
    fetch: (page, pageSize) =>
      listAlertEvents({
        page,
        pageSize,
        name: listName,
        view,
        mine: mine || undefined,
        pageUid: pageUid || undefined,
        levelUid: listLevel || undefined,
        startsAtFrom:
          view === "HISTORY"
            ? dayStartISO(committedHistoryFilters.from)
            : undefined,
        startsAtTo:
          view === "HISTORY"
            ? dayEndISO(committedHistoryFilters.to)
            : undefined,
        eventStatus:
          view === "HISTORY"
            ? committedHistoryFilters.status || undefined
            : undefined,
        ruleName:
          view === "HISTORY"
            ? committedHistoryFilters.ruleName || undefined
            : undefined,
        datasourceName:
          view === "HISTORY"
            ? committedHistoryFilters.datasourceName || undefined
            : undefined,
        suppression:
          view === "HISTORY"
            ? committedHistoryFilters.suppression || undefined
            : undefined,
      }),
    onLoaded: (data, meta) => {
      // 静默刷新时保留勾选，只丢弃已经离开实时列表的事件。
      setSelected((current) => {
        if (!meta.silent) {
          return current.length === 0 ? current : []
        }
        const visible = new Set((data.items ?? []).map((item) => item.uid))
        const next = current.filter((uid) => visible.has(uid))
        return next.length === current.length ? current : next
      })
      if (!sharedStats && view === "REALTIME") {
        void alertEventStats()
          .then(setLocalStats)
          .catch(() => undefined)
      }
    },
  })
  const [detail, setDetail] = useState<AlertEventItem | null>(null)
  const [chartEvent, setChartEvent] = useState<AlertEventItem | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [pending, setPending] = useState(false)
  const [actionError, setActionError] = useState("")
  const [confirm, setConfirm] = useState<{
    kind: ConfirmKind
    uids: string[]
  } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const [suppressEvent, setSuppressEvent] = useState<AlertEventItem | null>(
    null
  )
  const [suppressError, setSuppressError] = useState("")
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [now, setNow] = useState(() => Date.now())
  // 勾选列的实际宽度，用来定位紧随其后的固定列。
  const checkHeadRef = useRef<HTMLTableCellElement | null>(null)
  const [checkColumnWidth, setCheckColumnWidth] = useState(0)
  const canIntervene = view === "REALTIME" && session.can("alert:intervene")
  const visibleColumns =
    columns && columns.length > 0
      ? columns
      : view === "HISTORY"
        ? DEFAULT_HISTORY_COLUMNS
        : DEFAULT_REALTIME_COLUMNS
  // 横向滚动时钉住左边两列：勾选列（有操作权限时）和排在第一个的展示列。
  // 用户怎么调整列，钉住的都是当前排第一的那一列，而不是写死某个字段。
  // 勾选列的宽度是浏览器按内容算出来的（表格横向溢出时并不等于 2.5rem），
  // 所以第二列的 left 用实测宽度做偏移，否则第二列会整体右移，压住第三列的左侧内容。
  const pin = "sticky bg-background table-pin-left"
  const pinnedHead = `${pin} z-20`
  const pinnedCell = `${pin} group-hover:bg-muted`
  const pinnedCheckHead = `${pin} left-0 z-20`
  const pinnedCheckCell = `${pin} left-0 group-hover:bg-muted`
  const pinnedLeft = canIntervene ? checkColumnWidth : 0
  // 操作列固定在右侧：不换行，宽度跟着两个按钮走（w-px 让它缩到内容宽度）。
  const actionsHead = `sticky right-0 z-20 ${TABLE_COLUMN_WIDTH.actions} bg-background text-center table-pin-right`
  const actionsCell = `sticky right-0 ${TABLE_COLUMN_WIDTH.actions} bg-background text-center table-pin-right group-hover:bg-muted`
  const title =
    view === "HISTORY" ? t("alertEvent.historyTitle") : t("alertEvent.title")

  useEffect(() => {
    let active = true
    async function loadLevels() {
      try {
        const result = await listAlertDictionaries({
          type: "ALERT",
          page: 1,
          pageSize: 100,
        })
        if (active) {
          setLevels(result.items ?? [])
        }
      } catch {
        if (active) {
          setLevels([])
        }
      }
    }
    void loadLevels()
    return () => {
      active = false
    }
  }, [view, session.namespaceUid])

  useLayoutEffect(() => {
    const node = checkHeadRef.current
    if (!node) {
      setCheckColumnWidth(0)
      return
    }
    const sync = () => setCheckColumnWidth(node.getBoundingClientRect().width)
    sync()
    if (typeof ResizeObserver === "undefined") {
      return
    }
    const observer = new ResizeObserver(sync)
    observer.observe(node)
    return () => observer.disconnect()
  }, [canIntervene, loading, failed, items.length])

  useEffect(() => {
    if (view !== "REALTIME") {
      return
    }
    const timer = window.setInterval(() => setNow(Date.now()), 10000)
    return () => window.clearInterval(timer)
  }, [view])

  async function intervene(uids: string[]) {
    setPending(true)
    setActionError("")
    try {
      if (uids.length === 1) {
        await interveneAlertEvent(uids[0])
      } else {
        await interveneAlertEvents({ uids })
      }
      reload()
    } catch (err) {
      setActionError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function runConfirm() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      if (confirm.uids.length === 1) {
        await recoverAlertEvent(confirm.uids[0])
      } else {
        await recoverAlertEvents({ uids: confirm.uids })
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function runSuppress(input: SuppressInput) {
    if (!suppressEvent) {
      return
    }
    setPending(true)
    setSuppressError("")
    try {
      await suppressAlertEvent(
        suppressEvent.uid,
        input.seconds,
        input.permanent
      )
      setSuppressEvent(null)
      reload()
    } catch (err) {
      setSuppressError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: AlertEventItem): RowAction[] {
    const list: RowAction[] = []
    if (view === "REALTIME") {
      list.push({
        key: "chart",
        label: t("action.eventChart"),
        onSelect: () => setChartEvent(item),
      })
    }
    if (canIntervene) {
      list.push(
        {
          key: "intervene",
          label: t("action.intervene"),
          onSelect: () => void intervene([item.uid]),
        },
        {
          key: "suppress",
          label: t("action.suppress"),
          destructive: true,
          onSelect: () => setSuppressEvent(item),
        },
        {
          key: "recover",
          label: t("action.recover"),
          destructive: true,
          onSelect: () => setConfirm({ kind: "recover", uids: [item.uid] }),
        }
      )
    }
    return list
  }

  function toggle(uid: string) {
    setSelected((current) =>
      current.includes(uid)
        ? current.filter((item) => item !== uid)
        : [...current, uid]
    )
  }

  function applyHistoryFilters() {
    const next = new URLSearchParams(params)
    const entries = [
      ["name", historyDraft.name],
      ["level", historyDraft.level],
      ["ruleName", historyDraft.ruleName],
      ["datasourceName", historyDraft.datasourceName],
      ["from", historyDraft.from],
      ["to", historyDraft.to],
      ["status", historyDraft.status],
      ["suppression", historyDraft.suppression],
    ] as const
    for (const [key, value] of entries) {
      const queryKey = HISTORY_DETAIL_QUERY_KEYS[key]
      if (value) next.set(queryKey, value)
      else next.delete(queryKey)
    }
    next.set("page", "1")
    setParams(next)
    setHistoryDraftState(null)
    // 应用后收起面板，把垂直空间还给列表；已应用的筛选改为以标签呈现。
    setFiltersOpen(false)
  }

  function clearHistoryFilters() {
    const next = new URLSearchParams(params)
    for (const queryKey of Object.values(HISTORY_DETAIL_QUERY_KEYS)) {
      next.delete(queryKey)
    }
    next.set("page", "1")
    setParams(next)
    setHistoryDraftState(null)
  }

  function removeHistoryFilter(keys: string[]) {
    const next = new URLSearchParams(params)
    for (const key of keys) {
      next.delete(key)
    }
    next.set("page", "1")
    setParams(next)
    setHistoryDraftState(null)
  }

  // 实时告警只有摘要和级别两个条件：实时列表本身只包含仍在触发的事件，
  // 再加状态、抑制、时间范围会和它重叠，所以这里有意不提供。
  function applyRealtimeFilters() {
    const next = new URLSearchParams(params)
    const name = draft.trim()
    if (name) next.set("name", name)
    else next.delete("name")
    if (draftLevel) next.set("level", draftLevel)
    else next.delete("level")
    next.set("page", "1")
    setParams(next)
    setRealtimeFiltersOpen(false)
  }

  function clearRealtimeFilters() {
    const next = new URLSearchParams(params)
    next.delete("name")
    next.delete("level")
    next.set("page", "1")
    setParams(next)
    setDraft("")
    setDraftLevel("")
  }

  function removeRealtimeFilter(key: "name" | "level") {
    const next = new URLSearchParams(params)
    next.delete(key)
    next.set("page", "1")
    setParams(next)
    if (key === "name") setDraft("")
    else setDraftLevel("")
  }

  // 页码可能来自历史 URL 或被收缩后的结果集：总数更小或结果为空时回到最后一页，
  // 否则 ListFrame 会隐藏分页控件，用户只能手改 URL 才能继续。
  useEffect(() => {
    if (view !== "HISTORY" || loading || failed || query.page <= 1) {
      return
    }
    const lastPage = Math.max(1, Math.ceil(total / query.pageSize))
    if (
      lastPage !== query.page &&
      (items.length === 0 || query.page > lastPage)
    ) {
      query.update({ page: lastPage })
    }
    // query.update 每次渲染都是新函数，这里用页码与总数做真正的依赖。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [view, loading, failed, items.length, total, query.page, query.pageSize])

  const allChecked =
    items.length > 0 && items.every((item) => selected.includes(item.uid))
  const levelName =
    levels.find((item) => item.uid === committedHistoryFilters.level)?.name ??
    ""
  const filterChips = historyFilterChips(t, committedHistoryFilters, levelName)
  const realtimeLevelName =
    levels.find((item) => item.uid === analysisLevel)?.name ?? ""
  const realtimeChips: FilterChip[] = []
  if (view === "REALTIME") {
    if (query.name) {
      realtimeChips.push({
        key: "name",
        label: `${t("alertEvent.namePlaceholder")}: ${query.name}`,
        onRemove: () => removeRealtimeFilter("name"),
      })
    }
    if (analysisLevel) {
      realtimeChips.push({
        key: "level",
        label: `${t("alertEvent.level")}: ${realtimeLevelName || analysisLevel}`,
        onRemove: () => removeRealtimeFilter("level"),
      })
    }
  }

  return (
    <section className="flex min-h-0 min-w-0 flex-1 flex-col gap-6">
      {hideHeader ? null : (
        <PageHeader
          title={title}
          description={
            view === "HISTORY"
              ? t("alertEvent.historyDescription")
              : t("alertEvent.description")
          }
        />
      )}
      <Tabs
        value={view === "HISTORY" ? historyTab : "realtime"}
        onValueChange={(value) => {
          if (value === "analysis" || value === "details") {
            const next = new URLSearchParams(params)
            if (value === "analysis") next.set("historyTab", value)
            else next.delete("historyTab")
            setParams(next)
          }
        }}
        className={
          view === "HISTORY"
            ? "min-h-0 min-w-0 flex-1 gap-4 overflow-hidden"
            : "gap-4"
        }
      >
        {view === "HISTORY" ? (
          <>
            <TabsList
              variant="default"
              className="grid h-8 w-full max-w-[22rem] grid-cols-2 gap-0.5 rounded-lg border border-border/70 bg-muted/40 p-0.5"
            >
              <TabsTrigger
                value="analysis"
                title={t("alertEvent.analysisTabHint")}
                className="h-7 gap-1.5 px-2 text-xs sm:text-sm"
              >
                <BarChart3 aria-hidden="true" />
                {t("alertEvent.analysisTab")}
              </TabsTrigger>
              <TabsTrigger
                value="details"
                title={t("alertEvent.detailTabHint")}
                className="h-7 gap-1.5 px-2 text-xs sm:text-sm"
              >
                <List aria-hidden="true" />
                {t("alertEvent.detailTab")}
              </TabsTrigger>
            </TabsList>
            <TabsContent
              value="analysis"
              forceMount
              // forceMount 让分析结果在切换标签后保留（不必重新聚合），但 Radix 在
              // forceMount 下把面板视为常驻，不会替非激活面板加 hidden；不自己按受控
              // 的 historyTab 隐藏，统计卡片和图表会一直堆在明细列表上方。
              className={
                historyTab === "analysis"
                  ? "min-h-0 min-w-0 flex-1 overflow-y-auto overscroll-contain"
                  : "hidden"
              }
            >
              <AlertHistoryInsights
                key={session.namespaceUid}
                enabled={historyTab === "analysis"}
                onSelectFingerprint={(fingerprint) => {
                  const next = new URLSearchParams(params)
                  next.set("eventName", fingerprint)
                  next.set("page", "1")
                  next.delete("historyTab")
                  setHistoryDraftState(null)
                  setParams(next)
                }}
              />
            </TabsContent>
          </>
        ) : null}
        <TabsContent
          value={view === "HISTORY" ? "details" : "realtime"}
          className={
            view === "HISTORY"
              ? "flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden"
              : "min-w-0"
          }
        >
          <section
            aria-labelledby={
              view === "HISTORY" ? "alert-history-details-title" : undefined
            }
            aria-label={view === "REALTIME" ? title : undefined}
            className={
              view === "HISTORY"
                ? "flex min-h-0 min-w-0 flex-1 flex-col gap-3"
                : // 操作区（筛选/统计/批量按钮）和表格之间留出间距，避免贴在一起。
                  "min-w-0 space-y-4"
            }
          >
            {view === "HISTORY" ? (
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex min-w-0 flex-wrap items-baseline gap-x-2">
                  <h2
                    id="alert-history-details-title"
                    className="text-base font-semibold tracking-tight"
                  >
                    {t("alertEvent.detailTitle")}
                  </h2>
                  <span className="text-xs text-muted-foreground">
                    {t("alertEvent.detailSubtitle")}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  {historyDraftDirty ? (
                    <span
                      className="text-xs text-amber-600 dark:text-amber-400"
                      role="status"
                    >
                      {t("alertEvent.unappliedChanges")}
                    </span>
                  ) : null}
                  <FilterToggle
                    open={filtersOpen}
                    count={filterChips.length}
                    controls="history-detail-filters"
                    onClick={() => setFiltersOpen((open) => !open)}
                  />
                </div>
              </div>
            ) : null}
            {view === "HISTORY" && filterChips.length > 0 ? (
              <FilterChips
                items={filterChips.map((chip) => ({
                  key: chip.label,
                  label: chip.label,
                  onRemove: () => removeHistoryFilter(chip.keys),
                }))}
                onClearAll={clearHistoryFilters}
              />
            ) : null}
            <div
              id={view === "HISTORY" ? "history-detail-filters" : undefined}
              className={
                view === "HISTORY"
                  ? // 面板高度受限：展开筛选时也要给明细表格留出可滚动的高度。
                    `max-h-[min(60vh,32rem)] overflow-y-auto rounded-xl border border-border/70 bg-card/70 p-3 sm:p-4 ${filtersOpen ? "" : "hidden"}`
                  : "flex min-w-0 flex-col gap-2"
              }
            >
              {view === "HISTORY" ? (
                <div className="mb-4">
                  <p className="text-sm font-semibold tracking-tight">
                    {t("alertEvent.detailSearchTitle")}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {t("alertEvent.detailSearchHint")}
                  </p>
                </div>
              ) : null}
              <div
                className={
                  view === "HISTORY"
                    ? "flex flex-wrap items-center justify-between gap-3"
                    : "flex min-w-0 flex-col gap-2"
                }
              >
                {view === "REALTIME" ? (
                  <>
                    <div className="flex flex-wrap items-center justify-end gap-2">
                      <FilterToggle
                        open={realtimeFiltersOpen}
                        count={realtimeChips.length}
                        controls="realtime-filters"
                        onClick={() => setRealtimeFiltersOpen((open) => !open)}
                      />
                      <span className="shrink-0 text-sm text-muted-foreground">
                        {t("alertEvent.stats", {
                          firing: stats.firing,
                          suppressed: stats.suppressed,
                          recovered: stats.recovered,
                        })}
                      </span>
                      {canIntervene ? (
                        <>
                          <Button
                            variant="outline"
                            className="shrink-0"
                            disabled={pending || selected.length === 0}
                            onClick={() => void intervene(selected)}
                          >
                            {t("alertEvent.batchIntervene")}
                          </Button>
                          <Button
                            variant="outline"
                            className="shrink-0"
                            disabled={pending || selected.length === 0}
                            onClick={() =>
                              setConfirm({
                                kind: "batch-recover",
                                uids: selected,
                              })
                            }
                          >
                            {t("alertEvent.batchRecover")}
                          </Button>
                        </>
                      ) : null}
                    </div>
                    {realtimeChips.length > 0 ? (
                      <FilterChips
                        items={realtimeChips}
                        onClearAll={clearRealtimeFilters}
                      />
                    ) : null}
                  </>
                ) : null}
                <form
                  id={view === "REALTIME" ? "realtime-filters" : undefined}
                  role={view === "HISTORY" ? "search" : undefined}
                  aria-label={
                    view === "HISTORY" ? t("alertEvent.detailTitle") : undefined
                  }
                  className={
                    view === "HISTORY"
                      ? "grid w-full min-w-0 gap-4"
                      : `flex min-w-0 flex-wrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3 sm:p-4 ${realtimeFiltersOpen ? "" : "hidden"}`
                  }
                  onSubmit={(event) => {
                    event.preventDefault()
                    if (view === "HISTORY") {
                      if (!invalidHistoryDateRange) applyHistoryFilters()
                    } else {
                      applyRealtimeFilters()
                    }
                  }}
                >
                  {view === "HISTORY" ? (
                    <>
                      <FieldGroup className="grid-cols-2 xl:grid-cols-4">
                        <Field className="col-span-2 xl:col-span-2">
                          <FieldLabel
                            htmlFor="history-detail-name"
                            className="text-xs text-muted-foreground"
                          >
                            {t("alertEvent.namePlaceholder")}
                          </FieldLabel>
                          <div className="relative">
                            <Search
                              aria-hidden="true"
                              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
                            />
                            <Input
                              id="history-detail-name"
                              className="pl-9"
                              value={historyDraft.name}
                              placeholder={t("alertEvent.namePlaceholder")}
                              onChange={(event) =>
                                updateHistoryDraft("name", event.target.value)
                              }
                            />
                          </div>
                        </Field>
                        <Field className="min-w-0">
                          <FieldLabel
                            htmlFor="history-detail-rule"
                            className="text-xs text-muted-foreground"
                          >
                            {t("alertEvent.rule")}
                          </FieldLabel>
                          <Input
                            id="history-detail-rule"
                            value={historyDraft.ruleName}
                            placeholder={t("alertEvent.ruleNamePlaceholder")}
                            onChange={(event) =>
                              updateHistoryDraft("ruleName", event.target.value)
                            }
                          />
                        </Field>
                        <Field className="min-w-0">
                          <FieldLabel
                            htmlFor="history-detail-datasource"
                            className="text-xs text-muted-foreground"
                          >
                            {t("alertEvent.datasource")}
                          </FieldLabel>
                          <Input
                            id="history-detail-datasource"
                            value={historyDraft.datasourceName}
                            placeholder={t(
                              "alertEvent.datasourceNamePlaceholder"
                            )}
                            onChange={(event) =>
                              updateHistoryDraft(
                                "datasourceName",
                                event.target.value
                              )
                            }
                          />
                        </Field>
                        <Field className="min-w-0">
                          <FieldLabel
                            htmlFor="history-detail-level"
                            className="text-xs text-muted-foreground"
                          >
                            {t("alertEvent.level")}
                          </FieldLabel>
                          <Select
                            value={historyDraft.level || "ALL"}
                            onValueChange={(value) =>
                              updateHistoryDraft(
                                "level",
                                value === "ALL" ? "" : value
                              )
                            }
                          >
                            <SelectTrigger
                              id="history-detail-level"
                              className="w-full"
                            >
                              <SelectValue
                                placeholder={t("alertEvent.allLevels")}
                              />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectItem value="ALL">
                                  {t("alertEvent.allLevels")}
                                </SelectItem>
                                {levels.map((item) => (
                                  <SelectItem key={item.uid} value={item.uid}>
                                    {item.name}
                                  </SelectItem>
                                ))}
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        </Field>
                        <Field className="min-w-0">
                          <FieldLabel
                            htmlFor="history-detail-status"
                            className="text-xs text-muted-foreground"
                          >
                            {t("alertEvent.eventStatusLabel")}
                          </FieldLabel>
                          <Select
                            value={historyDraft.status || "ALL"}
                            onValueChange={(value) =>
                              updateHistoryDraft(
                                "status",
                                value === "ALL" ? "" : value
                              )
                            }
                          >
                            <SelectTrigger
                              id="history-detail-status"
                              className="w-full"
                            >
                              <SelectValue
                                placeholder={t("alertEvent.allStatuses")}
                              />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectItem value="ALL">
                                  {t("alertEvent.allStatuses")}
                                </SelectItem>
                                <SelectItem value="FIRING">
                                  {t("alertEvent.statusFiring")}
                                </SelectItem>
                                <SelectItem value="RECOVERED">
                                  {t("alertEvent.statusRecovered")}
                                </SelectItem>
                                <SelectItem value="RECOVERED_BY_MANUAL">
                                  {t("alertEvent.statusRecoveredManual")}
                                </SelectItem>
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        </Field>
                        <Field className="min-w-0">
                          <FieldLabel
                            htmlFor="history-detail-suppression"
                            className="text-xs text-muted-foreground"
                          >
                            {t("alertEvent.suppressionFilter")}
                          </FieldLabel>
                          <Select
                            value={historyDraft.suppression || "ALL"}
                            onValueChange={(value) =>
                              updateHistoryDraft(
                                "suppression",
                                value === "ALL" ? "" : value
                              )
                            }
                          >
                            <SelectTrigger
                              id="history-detail-suppression"
                              className="w-full"
                            >
                              <SelectValue
                                placeholder={t("alertEvent.allSuppression")}
                              />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectItem value="ALL">
                                  {t("alertEvent.allSuppression")}
                                </SelectItem>
                                <SelectItem value="SUPPRESSED">
                                  {t("alertEvent.suppressedOnly")}
                                </SelectItem>
                                <SelectItem value="UNSUPPRESSED">
                                  {t("alertEvent.unsuppressedOnly")}
                                </SelectItem>
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        </Field>
                        <Field
                          // 和旁边其他条件一样只占一格；窄屏下独占一行，避免日期被压断。
                          className="col-span-2 xl:col-span-1"
                          data-invalid={invalidHistoryDateRange || undefined}
                        >
                          <FieldLabel
                            htmlFor="history-detail-range"
                            className="text-xs text-muted-foreground"
                          >
                            {t("alertEvent.eventDateRange")}
                          </FieldLabel>
                          {/* 双月日历一次选完整区间，按天取值，最多 30 天。 */}
                          <DateRangePicker
                            id="history-detail-range"
                            from={historyDraft.from}
                            to={historyDraft.to}
                            maxDays={HISTORY_RANGE_MAX_DAYS}
                            invalid={invalidHistoryDateRange}
                            describedBy={
                              invalidHistoryDateRange
                                ? "history-detail-date-error"
                                : undefined
                            }
                            ariaLabel={t("alertEvent.eventDateRange")}
                            onChange={updateHistoryDraftRange}
                          />
                        </Field>
                      </FieldGroup>
                      <div className="flex flex-col gap-2 border-t border-border/60 pt-3 sm:flex-row sm:items-center sm:justify-between">
                        <div className="min-h-5 flex-1">
                          {invalidHistoryDateRange ? (
                            <p
                              id="history-detail-date-error"
                              role="alert"
                              aria-live="polite"
                              className="text-xs text-destructive"
                            >
                              {t("alertEvent.invalidDateRange")}
                            </p>
                          ) : (
                            <p className="text-xs text-muted-foreground">
                              {historyDraftDirty
                                ? t("alertEvent.unappliedChanges")
                                : filterChips.length > 0
                                  ? t("alertEvent.activeFilterCount", {
                                      count: filterChips.length,
                                    })
                                  : t("alertEvent.detailSearchHint")}
                            </p>
                          )}
                        </div>
                        <div className="flex flex-wrap items-center justify-end gap-2">
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            disabled={appliedHistoryFilterCount === 0}
                            onClick={clearHistoryFilters}
                          >
                            <X aria-hidden="true" data-icon="inline-start" />
                            {t("alertEvent.clearHistoryFilters")}
                          </Button>
                          <Button
                            type="submit"
                            disabled={invalidHistoryDateRange}
                          >
                            <Search
                              aria-hidden="true"
                              data-icon="inline-start"
                            />
                            {t("alertEvent.applyFilters")}
                          </Button>
                        </div>
                      </div>
                    </>
                  ) : (
                    <>
                      <Field className="min-w-0 flex-1 basis-56">
                        <FieldLabel
                          htmlFor="realtime-name"
                          className="text-xs text-muted-foreground"
                        >
                          {t("alertEvent.namePlaceholder")}
                        </FieldLabel>
                        <div className="relative">
                          <Search
                            aria-hidden="true"
                            className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
                          />
                          <Input
                            id="realtime-name"
                            className="pl-9"
                            value={draft}
                            onChange={(event) => setDraft(event.target.value)}
                          />
                        </div>
                      </Field>
                      <Field className="w-44 shrink-0">
                        <FieldLabel
                          htmlFor="realtime-level"
                          className="text-xs text-muted-foreground"
                        >
                          {t("alertEvent.level")}
                        </FieldLabel>
                        <Select
                          value={draftLevel || "ALL"}
                          onValueChange={(value) =>
                            setDraftLevel(value === "ALL" ? "" : value)
                          }
                        >
                          <SelectTrigger id="realtime-level" className="w-full">
                            <SelectValue
                              placeholder={t("alertEvent.allLevels")}
                            />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectGroup>
                              <SelectItem value="ALL">
                                {t("alertEvent.allLevels")}
                              </SelectItem>
                              {levels.map((item) => (
                                <SelectItem key={item.uid} value={item.uid}>
                                  {item.name}
                                </SelectItem>
                              ))}
                            </SelectGroup>
                          </SelectContent>
                        </Select>
                      </Field>
                      <div className="flex shrink-0 items-center gap-2">
                        {realtimeChips.length > 0 ? (
                          <Button
                            type="button"
                            variant="ghost"
                            onClick={clearRealtimeFilters}
                          >
                            {t("filter.clear")}
                          </Button>
                        ) : null}
                        <Button type="submit" variant="outline">
                          <Search aria-hidden="true" data-icon="inline-start" />
                          {t("alertEvent.applyFilters")}
                        </Button>
                      </div>
                    </>
                  )}
                </form>
              </div>
            </div>
            {actionError ? (
              <p className="text-sm text-destructive">{actionError}</p>
            ) : null}
            <ListFrame
              total={total}
              page={query.page}
              pageSize={query.pageSize}
              showPager={!loading && !failed && total > 0}
              actions={columnsAction}
              onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
              onNext={() => query.update({ page: query.page + 1 })}
            >
              {loading ? (
                <p className="text-sm text-muted-foreground">
                  {t("state.loading")}
                </p>
              ) : null}
              {!loading && failed ? (
                <div className="flex items-center gap-2">
                  <p className="text-sm text-destructive">
                    {loadError || t("state.failed")}
                  </p>
                  <Button variant="outline" onClick={() => reload()}>
                    {t("action.retry")}
                  </Button>
                </div>
              ) : null}
              {!loading && !failed && items.length === 0 ? (
                appliedHistoryFilterCount > 0 ? (
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <p className="text-sm text-muted-foreground">
                      {t("alertEvent.emptyFilteredHistory")}
                    </p>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={clearHistoryFilters}
                    >
                      {t("alertEvent.clearHistoryFilters")}
                    </Button>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">
                    {view === "HISTORY"
                      ? t("alertEvent.emptyHistory")
                      : t("alertEvent.empty")}
                  </p>
                )
              ) : null}
              {!loading && !failed && items.length > 0 ? (
                <Table
                  className={
                    view === "HISTORY" ? TABLE_MIN_WIDTH.default : undefined
                  }
                >
                  <TableHeader>
                    <TableRow>
                      {canIntervene ? (
                        <TableHead
                          ref={checkHeadRef}
                          className={`w-10 ${pinnedCheckHead}`}
                        >
                          <input
                            type="checkbox"
                            aria-label={t("alertEvent.batchIntervene")}
                            checked={allChecked}
                            onChange={() =>
                              setSelected(
                                allChecked ? [] : items.map((item) => item.uid)
                              )
                            }
                          />
                        </TableHead>
                      ) : null}
                      {visibleColumns.map((column, index) => (
                        <TableHead
                          key={`${column.type}:${column.key}`}
                          className={
                            index === 0
                              ? `${pinnedHead} ${view === "HISTORY" ? "min-w-[18rem]" : ""}`
                              : undefined
                          }
                          style={index === 0 ? { left: pinnedLeft } : undefined}
                        >
                          {columnTitle(t, i18n.language, column)}
                        </TableHead>
                      ))}
                      <TableHead className={actionsHead}>
                        {t("alertEvent.actions")}
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((item) => (
                      <TableRow key={item.uid} className="group">
                        {canIntervene ? (
                          <TableCell className={pinnedCheckCell}>
                            <input
                              type="checkbox"
                              aria-label={item.summary || item.uid}
                              checked={selected.includes(item.uid)}
                              onChange={() => toggle(item.uid)}
                            />
                          </TableCell>
                        ) : null}
                        {visibleColumns.map((column, index) => (
                          <TableCell
                            key={`${column.type}:${column.key}`}
                            className={index === 0 ? pinnedCell : undefined}
                            style={
                              index === 0 ? { left: pinnedLeft } : undefined
                            }
                          >
                            {eventCell(item, column, levels, now, view, t)}
                          </TableCell>
                        ))}
                        <TableCell className={actionsCell}>
                          <RowActions
                            detailLabel={t("action.detail")}
                            moreLabel={t("action.more")}
                            onDetail={() => setDetail(item)}
                            actions={actions(item)}
                            forceMenu
                            nowrap
                          />
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : null}
            </ListFrame>
          </section>
        </TabsContent>
      </Tabs>
      <AlertEventChartDialog
        event={chartEvent}
        open={chartEvent !== null}
        onOpenChange={(value) => {
          if (!value) {
            setChartEvent(null)
          }
        }}
      />
      <AlertEventDetail
        event={detail}
        level={
          detail ? levels.find((item) => item.uid === detail.levelUid) : null
        }
        now={now}
        open={detail !== null}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
      />
      {suppressEvent ? (
        <AlertEventSuppressDialog
          event={suppressEvent}
          open
          pending={pending}
          error={suppressError}
          onConfirm={(input) => void runSuppress(input)}
          onOpenChange={(value) => {
            if (!value) {
              setSuppressEvent(null)
              setSuppressError("")
            }
          }}
        />
      ) : null}
      <ConfirmAction
        open={confirm !== null}
        title={t("confirm.recoverTitle")}
        description={t("confirm.recoverDescription")}
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void runConfirm()}
        onOpenChange={(value) => {
          if (!value) {
            setConfirm(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}

function eventCell(
  item: AlertEventItem,
  column: AlertColumnItem,
  levels: AlertDictionaryItem[],
  now: number,
  view: AlertEventView,
  t: TFunction
): ReactNode {
  if (column.type === "LABEL") {
    const labels = parseEventLabels(item.labels)
    const value = labels[column.key]
    return value ? (
      <span className="block max-w-[16rem] truncate" title={value}>
        {value}
      </span>
    ) : (
      "—"
    )
  }
  switch (column.key) {
    case "summary": {
      const summary = item.summary || item.fingerprint
      return view === "HISTORY" ? (
        <span
          className="block max-w-[28rem] truncate font-medium text-foreground"
          title={summary}
        >
          {summary}
        </span>
      ) : (
        summary
      )
    }
    case "datasource":
      return (
        <span
          className="block max-w-[12rem] truncate"
          title={item.datasourceName || item.datasourceUid}
        >
          {item.datasourceName || item.datasourceUid || "—"}
        </span>
      )
    case "level":
      return levelCell(item.levelUid, levels)
    case "sample_value":
      return view === "HISTORY" ? (
        <span className="font-mono tabular-nums">{item.sampleValue}</span>
      ) : (
        item.sampleValue
      )
    case "starts_at":
      return view === "HISTORY" ? (
        <span className="font-mono text-xs text-muted-foreground tabular-nums">
          {formatTime(item.startsAt)}
        </span>
      ) : (
        formatTime(item.startsAt)
      )
    case "ends_at":
      return view === "HISTORY" ? (
        <span className="font-mono text-xs text-muted-foreground tabular-nums">
          {formatTime(item.endsAt) || "—"}
        </span>
      ) : (
        formatTime(item.endsAt) || "—"
      )
    case "duration":
      return (
        <span className="font-mono">
          {formatDuration(
            item.startsAt,
            // 历史事件已结束，时长必须止于 endsAt；用 now 会让旧记录一直增长。
            item.endsAt ? new Date(item.endsAt).getTime() : now
          )}
        </span>
      )
    case "intervened":
      return item.intervenedBy ? (
        <div className="grid">
          <span>{item.intervenedName || item.intervenedBy}</span>
          <span className="text-xs text-muted-foreground">
            {formatTime(item.intervenedAt)}
          </span>
        </div>
      ) : null
    case "status":
      return (
        <span className="inline-flex items-center gap-1.5">
          <StatusBadge value={item.status} />
          {/* 历史列表里“触发中”只可能来自仍被抑制的事件，必须标出来，
              否则与实时告警中的触发事件无法区分。 */}
          {item.suppressed ? (
            <StatusBadge
              value="SUPPRESSED"
              tone="warning"
              label={
                item.status === "FIRING"
                  ? t("alertEvent.suppressionActive")
                  : t("alertEvent.suppressionMarked")
              }
            />
          ) : null}
        </span>
      )
    case "description":
      return (
        <span className="block max-w-[20rem] truncate" title={item.description}>
          {item.description || "—"}
        </span>
      )
    case "expr":
      return (
        <span
          className="block max-w-[20rem] truncate font-mono text-xs"
          title={item.expr}
        >
          {item.expr || "—"}
        </span>
      )
    case "fingerprint":
      return (
        <span
          className="block max-w-[12rem] truncate font-mono text-xs"
          title={item.fingerprint}
        >
          {item.fingerprint}
        </span>
      )
    case "rule":
      return item.ruleName || item.ruleUid || "—"
    default:
      return null
  }
}

function levelCell(uid: string, levels: AlertDictionaryItem[]) {
  if (!uid) {
    return ""
  }
  const level = levels.find((item) => item.uid === uid)
  if (!level) {
    return <LevelBadge name={uid} />
  }
  return <LevelBadge name={level.name} color={level.color} />
}
