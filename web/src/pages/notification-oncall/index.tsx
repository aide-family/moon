import { Copy } from "lucide-react"
import { useEffect, useState, type FormEvent, type ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import { selectDatasources } from "@/api/datasource"
import type { SelectItem as DatasourceOption } from "@/api/datasource/types"
import { selectNotificationGroups } from "@/api/notification-group"
import type { SelectItem as GroupOption } from "@/api/notification-group/types"
import {
  createNotificationOnCall,
  deleteNotificationOnCall,
  listNotificationOnCalls,
  updateNotificationOnCall,
  updateNotificationOnCallStatus,
} from "@/api/notification-oncall"
import type {
  CreateNotificationOnCallParams,
  NotificationOnCallItem,
  OnCallGroupItem,
  OnCallWindow,
} from "@/api/notification-oncall/types"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet, type DetailSection } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { LabelTags } from "@/components/label-tags"
import { ListFrame } from "@/components/list-frame"
import { MultiSelect } from "@/components/multi-select"
import { OnCallTimeView } from "@/components/oncall-time-view"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { LevelBadge, StatusBadge } from "@/components/status-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatDateTime } from "@/lib/date-time"
import { loadAllPagedItems, loadAllSelectItems } from "@/lib/select-pages"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { formErrorMessage, readFormError } from "@/lib/form-error"
import {
  parseLabelPairs,
  serializeLabelPairs,
  type LabelPair,
} from "@/lib/labels"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

/** 时段编辑器按周一到周日展示并排序，和后端保存顺序一致。 */
const WEEKDAYS = [
  "MONDAY",
  "TUESDAY",
  "WEDNESDAY",
  "THURSDAY",
  "FRIDAY",
  "SATURDAY",
  "SUNDAY",
]

const DEFAULT_TIMEZONE =
  Intl.DateTimeFormat().resolvedOptions().timeZone || "Asia/Shanghai"

/** 重复发送间隔按分钟编辑，后端按秒保存；0 表示不限制。 */
const MAX_REPEAT_INTERVAL_MINUTES = 720

// 时间视图要一次看全所有值班项，不能只拿列表当前页；接口上限是 100。
const TIME_VIEW_PAGE_SIZE = 100

// 常见 IANA 时区，仅作为输入建议；后端仍会校验时区是否合法。
const TIMEZONE_OPTIONS = [
  "UTC",
  "Asia/Shanghai",
  "Asia/Tokyo",
  "Asia/Singapore",
  "Asia/Hong_Kong",
  "Asia/Seoul",
  "Asia/Kolkata",
  "Asia/Dubai",
  "Europe/London",
  "Europe/Paris",
  "Europe/Berlin",
  "Europe/Moscow",
  "America/New_York",
  "America/Chicago",
  "America/Denver",
  "America/Los_Angeles",
  "America/Sao_Paulo",
  "Australia/Sydney",
]

interface GroupBinding {
  groupUid: string
  windows: OnCallWindow[]
}

function sortWeekdays(days: string[]): string[] {
  return [...new Set(days)].sort(
    (left, right) => WEEKDAYS.indexOf(left) - WEEKDAYS.indexOf(right)
  )
}

function toOptions(items: Array<{ uid: string; name: string }>) {
  return items.map((item) => ({ value: item.uid, label: item.name }))
}

/** 把后端返回的秒数换算成分钟展示，非整分钟按四舍五入。 */
function toMinutes(seconds: number | undefined): number {
  return Math.round((seconds ?? 0) / 60)
}

interface GroupWindowsProps {
  binding: OnCallGroupItem
  name: string
}

/**
 * 详情里的单个通知组卡片：没有时段就是全天候，用标签标出来；
 * 跨午夜的时段在结束时间后标 +1，并把完整说明放进 title。
 */
function GroupWindows({ binding, name }: GroupWindowsProps) {
  const { t } = useTranslation()
  const windows = binding.windows ?? []

  return (
    <div className="grid gap-2 rounded-lg border bg-muted/30 px-3 py-2.5">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="min-w-0 text-sm font-medium break-words">{name}</span>
        {windows.length === 0 ? (
          <Badge
            variant="secondary"
            className="h-5 rounded-md px-2 font-normal"
          >
            {t("oncall.always")}
          </Badge>
        ) : null}
      </div>
      {windows.length > 0 ? (
        <ul className="grid gap-1.5">
          {windows.map((window, index) => {
            const days = sortWeekdays(window.weekdays ?? [])
            const crossesMidnight = window.end < window.start
            return (
              <li
                key={`${window.start}-${window.end}-${index}`}
                className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1"
              >
                <span className="flex flex-wrap gap-1">
                  {days.length > 0 ? (
                    days.map((day) => (
                      <Badge
                        key={day}
                        variant="secondary"
                        className="h-5 rounded-md px-1.5 text-[0.7rem] font-normal"
                      >
                        {t(`oncall.weekday.${day}`)}
                      </Badge>
                    ))
                  ) : (
                    <span className="text-xs text-muted-foreground">
                      {t("oncall.anyDay")}
                    </span>
                  )}
                </span>
                <span
                  className="font-mono text-xs text-muted-foreground tabular-nums"
                  title={
                    crossesMidnight ? t("oncall.crossMidnightHint") : undefined
                  }
                >
                  {window.start}–{window.end}
                  {crossesMidnight ? (
                    <span className="ml-1 font-sans text-foreground">+1</span>
                  ) : null}
                </span>
              </li>
            )
          })}
        </ul>
      ) : null}
    </div>
  )
}

export function NotificationOnCallPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [groups, setGroups] = useState<GroupOption[]>([])
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [datasourceLevels, setDatasourceLevels] = useState<
    AlertDictionaryItem[]
  >([])
  const [datasources, setDatasources] = useState<DatasourceOption[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationOnCallItem | null>(null)
  const [name, setName] = useState("")
  const [labels, setLabels] = useState<LabelPair[]>([{ key: "", value: "" }])
  const [levelUids, setLevelUids] = useState<string[]>([])
  const [datasourceUids, setDatasourceUids] = useState<string[]>([])
  const [datasourceLevelUids, setDatasourceLevelUids] = useState<string[]>([])
  const [timezone, setTimezone] = useState(DEFAULT_TIMEZONE)
  const [bindings, setBindings] = useState<GroupBinding[]>([])
  const [repeatIntervalMinutes, setRepeatIntervalMinutes] = useState("0")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NotificationOnCallItem | null>(null)
  // 详情内容与开关分开：关闭动画期间保留上一次的数据，标题不会瞬间变空。
  const [detailOpen, setDetailOpen] = useState(false)
  const [confirm, setConfirm] = useState<{
    action: "disable" | "delete"
    item: NotificationOnCallItem
  } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("notification:write")
  // 列表与时间视图是同一份数据（同一套筛选），只是呈现方式不同。
  // 时间视图是默认视图，URL 只记录非默认的列表视图（?tab=list），地址栏保持干净。
  const [searchParams, setSearchParams] = useSearchParams()
  const view = searchParams.get("tab") === "list" ? "list" : "time"
  function setView(next: "list" | "time") {
    const params = new URLSearchParams(searchParams)
    if (next === "list") {
      params.set("tab", "list")
    } else {
      params.delete("tab")
    }
    // Radix Tabs 在 onMouseDown 里先 focus 再激活，onFocus 与 onMouseDown 各触发一次
    // onValueChange；两次回调发生在同一次渲染内，闭包里的 searchParams 仍是旧值。
    // 只有 window.location 被 pushState 同步更新，用它判断目标是否已经生效。
    if (params.toString() === window.location.search.replace(/^\?/, "")) {
      return
    }
    setSearchParams(params)
  }
  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    reload,
  } = usePagedList<NotificationOnCallItem>({
    // 时间视图按整租户渲染，一页取满；列表仍按 URL 里的分页。
    page: view === "time" ? 1 : query.page,
    pageSize: view === "time" ? TIME_VIEW_PAGE_SIZE : query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) =>
      listNotificationOnCalls({ page, pageSize, name: query.name }),
  })

  useEffect(() => {
    let active = true
    void Promise.all([
      loadAllSelectItems(
        (lastUid) => selectNotificationGroups({ limit: 100, lastUid }),
        100
      ),
      loadAllPagedItems((page, pageSize) =>
        listAlertDictionaries({ page, pageSize, type: "ALERT" })
      ),
      loadAllPagedItems((page, pageSize) =>
        listAlertDictionaries({ page, pageSize, type: "DATASOURCE" })
      ),
      loadAllSelectItems(
        (lastUid) => selectDatasources({ limit: 100, lastUid }),
        100
      ),
    ])
      .then(
        ([
          groupResult,
          levelResult,
          datasourceLevelResult,
          datasourceResult,
        ]) => {
          if (active) {
            setGroups(groupResult)
            setLevels(levelResult)
            setDatasourceLevels(datasourceLevelResult)
            setDatasources(datasourceResult)
          }
        }
      )
      .catch(() => {
        if (active) {
          setGroups([])
          setLevels([])
          setDatasourceLevels([])
          setDatasources([])
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  function toggle(values: string[], uid: string): string[] {
    return values.includes(uid)
      ? values.filter((value) => value !== uid)
      : [...values, uid]
  }

  function selectionSummary(selected: number, total: number): string {
    return selected >= total && total > 0
      ? t("multiSelect.allSelected", { count: selected })
      : t("multiSelect.summary", { count: selected, total })
  }

  function intervalError(): string | null {
    const trimmed = repeatIntervalMinutes.trim()
    if (!/^\d+$/.test(trimmed)) {
      return t("oncall.repeatIntervalInvalid")
    }
    if (Number(trimmed) > MAX_REPEAT_INTERVAL_MINUTES) {
      return t("oncall.repeatIntervalInvalid")
    }
    return null
  }

  /** 列表和详情把秒数展示为分钟；0 表示该告警不限制重复通知。 */
  function intervalSummary(seconds: number | undefined): string {
    const minutes = toMinutes(seconds)
    return minutes > 0
      ? t("oncall.repeatIntervalValue", { minutes })
      : t("oncall.repeatIntervalOff")
  }

  function openCreate() {
    setEditing(null)
    setName("")
    setLabels([{ key: "", value: "" }])
    setLevelUids([])
    setDatasourceUids([])
    setDatasourceLevelUids([])
    setTimezone(DEFAULT_TIMEZONE)
    setBindings([])
    setRepeatIntervalMinutes("0")
    setRemark("")
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openEdit(item: NotificationOnCallItem) {
    setEditing(item)
    setName(item.name)
    setLabels(parseLabelPairs(item.labels))
    setLevelUids(item.levelUids ?? [])
    setDatasourceUids(item.datasourceUids ?? [])
    setDatasourceLevelUids(item.datasourceLevelUids ?? [])
    setTimezone(item.timezone || DEFAULT_TIMEZONE)
    setBindings(
      (item.groups ?? []).map((group) => ({
        groupUid: group.groupUid,
        windows: (group.windows ?? []).map((window) => ({
          weekdays: sortWeekdays(window.weekdays ?? []),
          start: window.start,
          end: window.end,
        })),
      }))
    )
    setRepeatIntervalMinutes(String(toMinutes(item.repeatIntervalSeconds)))
    setRemark(item.remark)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  /**
   * 通知组多选按选择顺序维护时段绑定：取消勾选移除绑定，
   * 新勾选的组追加一个没有时段的绑定（无时段即全天候）。
   */
  function changeGroups(nextUids: string[]) {
    setBindings((current) => {
      const next = new Set(nextUids)
      const kept = current.filter((binding) => next.has(binding.groupUid))
      const keptUids = new Set(kept.map((binding) => binding.groupUid))
      const added = nextUids
        .filter((uid) => !keptUids.has(uid))
        .map((uid) => ({ groupUid: uid, windows: [] }))
      return [...kept, ...added]
    })
  }

  function updateWindows(
    groupUid: string,
    change: (windows: OnCallWindow[]) => OnCallWindow[]
  ) {
    setBindings((current) =>
      current.map((binding) =>
        binding.groupUid === groupUid
          ? { ...binding, windows: change(binding.windows) }
          : binding
      )
    )
  }

  function addWindow(groupUid: string) {
    updateWindows(groupUid, (windows) => [
      ...windows,
      { weekdays: [], start: "09:00", end: "18:00" },
    ])
  }

  function removeWindow(groupUid: string, index: number) {
    updateWindows(groupUid, (windows) =>
      windows.filter((_, current) => current !== index)
    )
  }

  function patchWindow(
    groupUid: string,
    index: number,
    patch: Partial<OnCallWindow>
  ) {
    updateWindows(groupUid, (windows) =>
      windows.map((window, current) =>
        current === index ? { ...window, ...patch } : window
      )
    )
  }

  function toggleWeekday(groupUid: string, index: number, weekday: string) {
    updateWindows(groupUid, (windows) =>
      windows.map((window, current) =>
        current === index
          ? { ...window, weekdays: toggle(window.weekdays, weekday) }
          : window
      )
    )
  }

  function groupName(uid: string): string {
    return groups.find((group) => group.uid === uid)?.name ?? uid
  }

  function windowsSummary(windows: OnCallWindow[]): string {
    if (!windows || windows.length === 0) {
      return t("oncall.always")
    }
    return windows
      .map((window) => {
        const days =
          sortWeekdays(window.weekdays ?? [])
            .map((day) => t(`oncall.weekday.${day}`))
            .join(",") || t("oncall.anyDay")
        return `${days} ${window.start}-${window.end}`
      })
      .join("; ")
  }

  function groupSummary(item: NotificationOnCallItem): string {
    const groupBindings = item.groups ?? []
    if (groupBindings.length === 0) {
      return "—"
    }
    return groupBindings
      .map(
        (binding) =>
          `${binding.groupName || groupName(binding.groupUid)}: ${windowsSummary(binding.windows ?? [])}`
      )
      .join(" · ")
  }

  function conditionSummary(item: NotificationOnCallItem): string {
    const parts: string[] = []
    const pairs = parseLabelPairs(item.labels).filter(
      (pair) => pair.key.trim() !== ""
    )
    if (pairs.length > 0) {
      parts.push(pairs.map((pair) => `${pair.key}=${pair.value}`).join(", "))
    }
    const levelNames = (item.levelUids ?? []).map(
      (uid) => levels.find((level) => level.uid === uid)?.name ?? uid
    )
    if (levelNames.length > 0) {
      parts.push(`${t("oncall.levels")}: ${levelNames.join(", ")}`)
    }
    const datasourceNames = (item.datasourceUids ?? []).map(
      (uid) =>
        datasources.find((datasource) => datasource.uid === uid)?.name ?? uid
    )
    if (datasourceNames.length > 0) {
      parts.push(`${t("oncall.datasources")}: ${datasourceNames.join(", ")}`)
    }
    const datasourceLevelNames = (item.datasourceLevelUids ?? []).map(
      (uid) => datasourceLevels.find((level) => level.uid === uid)?.name ?? uid
    )
    if (datasourceLevelNames.length > 0) {
      parts.push(
        `${t("oncall.datasourceLevels")}: ${datasourceLevelNames.join(", ")}`
      )
    }
    return parts.length > 0 ? parts.join(" · ") : t("oncall.anyEvent")
  }

  /**
   * 打开详情抽屉：只读视图，编辑/停用/删除等写操作只从列表行操作进入。
   * 内容与开关分开维护，关闭动画期间仍显示上一次的数据。
   */
  function openDetail(item: NotificationOnCallItem) {
    setDetail(item)
    setDetailOpen(true)
  }

  async function copyText(value: string) {
    try {
      await navigator.clipboard.writeText(value)
      toast.success(t("action.copied"))
    } catch {
      toast.error(t("action.copyFailed"))
    }
  }

  /** 级别类字段用字典颜色渲染；字典缺失时退回 UID，详情不留空白。 */
  function levelTags(
    uids: string[],
    dictionary: AlertDictionaryItem[]
  ): ReactNode {
    if (uids.length === 0) {
      return null
    }
    return (
      <span className="flex flex-wrap gap-1.5">
        {uids.map((uid) => {
          const level = dictionary.find((entry) => entry.uid === uid)
          return (
            <LevelBadge
              key={uid}
              name={level?.name ?? uid}
              color={level?.color}
            />
          )
        })}
      </span>
    )
  }

  function datasourceTags(uids: string[]): ReactNode {
    if (uids.length === 0) {
      return null
    }
    return (
      <span className="flex flex-wrap gap-1.5">
        {uids.map((uid) => (
          <Badge
            key={uid}
            variant="outline"
            className="h-auto max-w-full rounded-md px-2 py-0.5 font-normal break-all whitespace-normal"
          >
            {datasources.find((item) => item.uid === uid)?.name ?? uid}
          </Badge>
        ))}
      </span>
    )
  }

  /**
   * 详情分区按「匹配什么 → 谁接收、什么时段 → 备注」排列：
   * 匹配条件用左标签右取值的密集排版，接收时段按通知组卡片展开。
   */
  function detailSections(item: NotificationOnCallItem): DetailSection[] {
    const labelEntries = parseLabelPairs(item.labels).filter(
      (pair) => pair.key.trim() !== ""
    )
    const levelUids = item.levelUids ?? []
    const datasourceUids = item.datasourceUids ?? []
    const datasourceLevelUids = item.datasourceLevelUids ?? []
    const groupBindings = item.groups ?? []
    const hasConditions =
      labelEntries.length > 0 ||
      levelUids.length > 0 ||
      datasourceUids.length > 0 ||
      datasourceLevelUids.length > 0

    const sections: DetailSection[] = [
      {
        title: t("oncall.conditions"),
        description: hasConditions ? undefined : t("oncall.anyEvent"),
        layout: "inline",
        fields: [
          {
            label: t("oncall.labels"),
            value: <LabelTags labels={item.labels} />,
          },
          {
            label: t("oncall.levels"),
            value: levelTags(levelUids, levels),
          },
          {
            label: t("oncall.datasources"),
            value: datasourceTags(datasourceUids),
          },
          {
            label: t("oncall.datasourceLevels"),
            value: levelTags(datasourceLevelUids, datasourceLevels),
          },
        ],
      },
      {
        title: t("oncall.receiveWindows"),
        content:
          groupBindings.length > 0 ? (
            <div className="grid gap-2">
              {groupBindings.map((binding) => (
                <GroupWindows
                  key={binding.groupUid}
                  binding={binding}
                  name={binding.groupName || groupName(binding.groupUid)}
                />
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">—</p>
          ),
      },
    ]

    if (item.remark.trim() !== "") {
      sections.push({
        title: t("oncall.remark"),
        content: (
          <p className="text-sm leading-relaxed break-words whitespace-pre-wrap">
            {item.remark}
          </p>
        ),
      })
    }

    return sections
  }

  function windowError(): string | null {
    for (const binding of bindings) {
      for (const window of binding.windows) {
        if ((window.weekdays ?? []).length === 0) {
          return t("oncall.windowWeekdaysRequired")
        }
        if (window.start.trim() === "" || window.end.trim() === "") {
          return t("oncall.windowTimeRequired")
        }
        if (window.start === window.end) {
          return t("oncall.sameTime")
        }
      }
    }
    return null
  }

  async function onSave(event: FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      ["groups", bindings.length === 0 ? t("form.required") : null],
      [
        "labels",
        labels.some(
          (pair) => pair.key.trim() === "" && pair.value.trim() !== ""
        )
          ? t("form.required")
          : null,
      ],
      ["windows", windowError()],
      ["repeatIntervalSeconds", intervalError()],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const payload: CreateNotificationOnCallParams = {
        name: name.trim(),
        labels: serializeLabelPairs(labels),
        levelUids,
        datasourceUids,
        datasourceLevelUids,
        timezone: timezone.trim() || DEFAULT_TIMEZONE,
        groups: bindings.map((binding) => ({
          groupUid: binding.groupUid,
          windows: binding.windows.map((window) => ({
            weekdays: sortWeekdays(window.weekdays ?? []),
            start: window.start,
            end: window.end,
          })),
        })),
        repeatIntervalSeconds: Number(repeatIntervalMinutes.trim()) * 60,
        remark,
      }
      if (editing) {
        await updateNotificationOnCall({ ...payload, uid: editing.uid })
      } else {
        await createNotificationOnCall(payload)
      }
      setOpen(false)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
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
      if (confirm.action === "delete") {
        await deleteNotificationOnCall(confirm.item.uid)
      } else {
        await updateNotificationOnCallStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: NotificationOnCallItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      item.status === "ACTIVE"
        ? {
            key: "disable",
            label: t("action.disable"),
            destructive: true,
            onSelect: () => setConfirm({ action: "disable", item }),
          }
        : {
            key: "enable",
            label: t("action.enable"),
            onSelect: () => {
              // 启停是行内直接生效的操作，没有打开中的弹窗可以显示错误。
              // 403 由 http 层跳 /forbidden，404/5xx 已由 http 层全局 toast；
              // 这里只补它没覆盖的失败（400/409/网络错误），避免重复提示。
              void updateNotificationOnCallStatus(item.uid, "ACTIVE")
                .then(() => reload())
                .catch((err) => {
                  if (
                    err instanceof ApiError &&
                    (err.status === 403 ||
                      err.status === 404 ||
                      err.status >= 500)
                  ) {
                    return
                  }
                  toast.error(formErrorMessage(err, t("state.failed")))
                })
            },
          },
      {
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => setConfirm({ action: "delete", item }),
      },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("oncall.title")}
        description={t("oncall.description")}
      >
        {canWrite ? (
          <Button onClick={openCreate}>{t("action.create")}</Button>
        ) : null}
      </PageHeader>
      {/* 搜索框与按钮固定一行：输入框可缩、按钮不缩，窄屏时整组（而不是按钮）换行。 */}
      <div className="flex min-w-0 flex-wrap items-center justify-between gap-x-3 gap-y-2">
        <form
          className="flex min-w-0 flex-1 items-center gap-2"
          onSubmit={(event) => {
            event.preventDefault()
            query.update({ name: draft })
          }}
        >
          <Input
            className="min-w-0 flex-1 sm:max-w-72"
            value={draft}
            placeholder={t("oncall.namePlaceholder")}
            onChange={(event) => setDraft(event.target.value)}
          />
          <Button type="submit" variant="outline" className="shrink-0">
            {t("action.search")}
          </Button>
        </form>
        <Tabs
          className="w-full shrink-0 sm:w-auto"
          value={view}
          onValueChange={(value) => setView(value as "list" | "time")}
        >
          <TabsList>
            <TabsTrigger value="time">{t("oncall.viewTime")}</TabsTrigger>
            <TabsTrigger value="list">{t("oncall.viewList")}</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>
      {view === "time" ? (
        // app shell 是 h-svh overflow-hidden：这里必须是唯一纵向滚动 owner。
        <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-2 overflow-y-auto pr-1">
          {!loading && !failed && total > items.length ? (
            <p className="text-xs text-muted-foreground">
              {t("oncall.timeViewTruncated", {
                shown: items.length,
                total,
              })}
            </p>
          ) : null}
          <OnCallTimeView
            items={items}
            loading={loading}
            emptyText={
              failed ? loadError || t("state.failed") : t("oncall.empty")
            }
          />
        </div>
      ) : (
        <ListFrame
          total={total}
          page={query.page}
          pageSize={query.pageSize}
          showPager={!loading && !failed && items.length > 0}
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
            <p className="text-sm text-muted-foreground">{t("oncall.empty")}</p>
          ) : null}
          {!loading && !failed && items.length > 0 ? (
            <Table className={TABLE_MIN_WIDTH.default}>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("oncall.name")}</TableHead>
                  <TableHead>{t("oncall.conditions")}</TableHead>
                  <TableHead>{t("oncall.groups")}</TableHead>
                  <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                  <TableActionsHead>{t("action.actions")}</TableActionsHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.uid}>
                    <TableCell className="font-medium break-words">
                      {item.name}
                    </TableCell>
                    <TableCell className="max-w-md text-sm text-muted-foreground">
                      <span className="break-words">
                        {conditionSummary(item)}
                      </span>
                    </TableCell>
                    <TableCell className="max-w-md text-sm">
                      <span className="break-words">{groupSummary(item)}</span>
                    </TableCell>
                    <TableCell className={TABLE_NOWRAP}>
                      <StatusBadge value={item.status} />
                    </TableCell>
                    <TableCell>
                      <RowActions
                        nowrap
                        detailLabel={t("action.detail")}
                        moreLabel={t("action.more")}
                        onDetail={() => openDetail(item)}
                        actions={actions(item)}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : null}
        </ListFrame>
      )}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="gap-0 p-0 sm:max-w-3xl">
          <DialogHeader className="shrink-0 border-b px-4 py-3 pr-12">
            <DialogTitle>
              {editing ? t("action.edit") : t("action.create")}
            </DialogTitle>
          </DialogHeader>
          <form
            noValidate
            className="flex min-h-0 flex-1 flex-col"
            onSubmit={onSave}
          >
            <DialogBody className="grid gap-4 px-4 py-4">
              <FormField
                label={t("oncall.name")}
                htmlFor="oncall-name"
                required
                error={fieldErrors.name}
              >
                <Input
                  id="oncall-name"
                  value={name}
                  placeholder={t("oncall.namePlaceholder")}
                  onChange={(event) => setName(event.target.value)}
                  required
                />
              </FormField>
              <FormField
                label={t("oncall.labels")}
                hint={t("oncall.labelsHint")}
                error={fieldErrors.labels}
              >
                <div className="grid gap-2">
                  {labels.map((pair, index) => (
                    <div key={index} className="flex items-center gap-2">
                      <Input
                        value={pair.key}
                        placeholder={t("oncall.labelKey")}
                        onChange={(event) => {
                          const next = [...labels]
                          next[index] = { ...pair, key: event.target.value }
                          setLabels(next)
                        }}
                      />
                      <Input
                        value={pair.value}
                        placeholder={t("oncall.labelValue")}
                        onChange={(event) => {
                          const next = [...labels]
                          next[index] = { ...pair, value: event.target.value }
                          setLabels(next)
                        }}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() =>
                          setLabels(
                            labels.filter((_, current) => current !== index)
                          )
                        }
                      >
                        {t("action.delete")}
                      </Button>
                    </div>
                  ))}
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() =>
                      setLabels([...labels, { key: "", value: "" }])
                    }
                  >
                    {t("oncall.addLabel")}
                  </Button>
                </div>
              </FormField>
              <FormField
                label={t("oncall.levels")}
                htmlFor="oncall-levels"
                hint={t("oncall.levelsHint")}
              >
                <MultiSelect
                  id="oncall-levels"
                  values={levelUids}
                  options={toOptions(levels)}
                  onChange={setLevelUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("oncall.noLevels")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              <FormField
                label={t("oncall.datasources")}
                htmlFor="oncall-datasources"
                hint={t("oncall.datasourcesHint")}
              >
                <MultiSelect
                  id="oncall-datasources"
                  values={datasourceUids}
                  options={toOptions(datasources)}
                  onChange={setDatasourceUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("oncall.noDatasources")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              <FormField
                label={t("oncall.datasourceLevels")}
                htmlFor="oncall-datasource-levels"
                hint={t("oncall.datasourceLevelsHint")}
              >
                <MultiSelect
                  id="oncall-datasource-levels"
                  values={datasourceLevelUids}
                  options={toOptions(datasourceLevels)}
                  onChange={setDatasourceLevelUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("oncall.noDatasourceLevels")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              <FormField
                label={t("oncall.timezone")}
                htmlFor="oncall-timezone"
                hint={t("oncall.timezoneHint")}
                error={fieldErrors.timezone}
              >
                <Input
                  id="oncall-timezone"
                  list="oncall-timezones"
                  value={timezone}
                  placeholder={t("oncall.timezonePlaceholder")}
                  onChange={(event) => setTimezone(event.target.value)}
                />
                <datalist id="oncall-timezones">
                  {TIMEZONE_OPTIONS.map((zone) => (
                    <option key={zone} value={zone} />
                  ))}
                </datalist>
              </FormField>
              <FormField
                label={t("oncall.groups")}
                htmlFor="oncall-groups"
                hint={t("oncall.groupsHint")}
                required
                error={fieldErrors.groups}
              >
                <MultiSelect
                  id="oncall-groups"
                  invalid={Boolean(fieldErrors.groups)}
                  values={bindings.map((binding) => binding.groupUid)}
                  options={toOptions(groups)}
                  onChange={changeGroups}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("oncall.noGroups")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              {bindings.length > 0 ? (
                <FormField
                  label={t("oncall.receiveWindows")}
                  hint={t("oncall.receiveWindowsHint")}
                  error={fieldErrors.windows}
                >
                  <div className="grid gap-3">
                    {bindings.map((binding) => (
                      <div
                        key={binding.groupUid}
                        className="rounded-md border bg-muted/30 p-3"
                      >
                        <p className="mb-2 text-sm font-medium">
                          {groupName(binding.groupUid)}
                        </p>
                        <div className="grid gap-2">
                          {binding.windows.map((window, index) => (
                            <div
                              key={index}
                              className="grid gap-2 rounded-md border border-dashed bg-background p-2"
                            >
                              <div className="flex flex-wrap gap-1.5">
                                {WEEKDAYS.map((day) => (
                                  <label
                                    key={day}
                                    className="inline-flex cursor-pointer items-center gap-1.5 rounded-md border bg-background px-2 py-1 text-xs"
                                  >
                                    <input
                                      type="checkbox"
                                      checked={window.weekdays.includes(day)}
                                      onChange={() =>
                                        toggleWeekday(
                                          binding.groupUid,
                                          index,
                                          day
                                        )
                                      }
                                    />
                                    {t(`oncall.weekday.${day}`)}
                                  </label>
                                ))}
                              </div>
                              <div className="flex flex-wrap items-center gap-2">
                                <Input
                                  type="time"
                                  className="w-32"
                                  aria-label={t("oncall.windowStart")}
                                  value={window.start}
                                  onChange={(event) =>
                                    patchWindow(binding.groupUid, index, {
                                      start: event.target.value,
                                    })
                                  }
                                />
                                <span className="text-muted-foreground">–</span>
                                <Input
                                  type="time"
                                  className="w-32"
                                  aria-label={t("oncall.windowEnd")}
                                  value={window.end}
                                  onChange={(event) =>
                                    patchWindow(binding.groupUid, index, {
                                      end: event.target.value,
                                    })
                                  }
                                />
                                <Button
                                  type="button"
                                  variant="ghost"
                                  onClick={() =>
                                    removeWindow(binding.groupUid, index)
                                  }
                                >
                                  {t("oncall.removeWindow")}
                                </Button>
                              </div>
                            </div>
                          ))}
                          <div className="flex flex-wrap items-center gap-3">
                            <Button
                              type="button"
                              variant="outline"
                              onClick={() => addWindow(binding.groupUid)}
                            >
                              {t("oncall.addWindow")}
                            </Button>
                            {binding.windows.length === 0 ? (
                              <p className="text-xs text-muted-foreground">
                                {t("oncall.alwaysHint")}
                              </p>
                            ) : (
                              <p className="text-xs text-muted-foreground">
                                {t("oncall.crossMidnightHint")}
                              </p>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </FormField>
              ) : null}
              <FormField
                label={t("oncall.repeatInterval")}
                htmlFor="oncall-repeat-interval"
                hint={t("oncall.repeatIntervalHint")}
                error={fieldErrors.repeatIntervalSeconds}
              >
                <Input
                  id="oncall-repeat-interval"
                  type="number"
                  min={0}
                  max={MAX_REPEAT_INTERVAL_MINUTES}
                  step={1}
                  inputMode="numeric"
                  className="max-w-40"
                  value={repeatIntervalMinutes}
                  onChange={(event) =>
                    setRepeatIntervalMinutes(event.target.value)
                  }
                />
              </FormField>
              <FormField label={t("oncall.remark")} htmlFor="oncall-remark">
                <Input
                  id="oncall-remark"
                  value={remark}
                  placeholder={t("oncall.remarkPlaceholder")}
                  onChange={(event) => setRemark(event.target.value)}
                />
              </FormField>
            </DialogBody>
            {error ? (
              <p className="text-sm text-destructive">{error}</p>
            ) : null}
            <DialogFooter className="mx-0 mb-0 shrink-0 border-t px-4 py-3">
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <DetailSheet
        open={detailOpen}
        size="40vw"
        title={detail?.name ?? ""}
        description={
          detail ? (
            <>
              <StatusBadge value={detail.status} />
              <span className="font-mono text-[0.7rem] break-all">
                {detail.uid}
              </span>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label={`${t("action.copy")} ${t("detail.uid")}`}
                onClick={() => void copyText(detail.uid)}
              >
                <Copy />
              </Button>
            </>
          ) : undefined
        }
        summary={
          detail ? (
            <dl className="grid gap-3 @md:grid-cols-2">
              <div className="grid gap-0.5">
                <dt className="text-xs text-muted-foreground">
                  {t("oncall.repeatInterval")}
                </dt>
                <dd className="text-sm font-medium tabular-nums">
                  {intervalSummary(detail.repeatIntervalSeconds)}
                </dd>
              </div>
              <div className="grid gap-0.5">
                <dt className="text-xs text-muted-foreground">
                  {t("oncall.timezone")}
                </dt>
                <dd className="font-mono text-sm font-medium break-all">
                  {detail.timezone || DEFAULT_TIMEZONE}
                </dd>
              </div>
              <div className="grid gap-0.5 @md:col-span-2">
                <dt className="text-xs text-muted-foreground">
                  {t("oncall.createdAt")}
                </dt>
                <dd className="text-sm font-medium tabular-nums">
                  {formatDateTime(detail.createdAt) || "—"}
                </dd>
              </div>
            </dl>
          ) : undefined
        }
        sections={detail ? detailSections(detail) : []}
        onOpenChange={setDetailOpen}
      />
      <ConfirmAction
        open={confirm !== null}
        title={
          confirm?.action === "delete"
            ? t("confirm.deleteTitle")
            : t("confirm.disableTitle")
        }
        description={
          confirm?.action === "delete"
            ? t("confirm.deleteDescription")
            : t("confirm.disableDescription")
        }
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
