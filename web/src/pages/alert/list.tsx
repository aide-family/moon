import { useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"
import type { TFunction } from "i18next"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { formMessage } from "@/lib/form-error"

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
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
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
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { EventChartDialog } from "@/pages/alert/event-chart"
import { AlertEventDetail } from "@/pages/alert/event-detail"
import {
  DEFAULT_HISTORY_COLUMNS,
  DEFAULT_REALTIME_COLUMNS,
  fieldLabel,
  parseEventLabels,
} from "@/pages/alert/columns"
import { formatDuration, formatTime } from "@/pages/alert/format"
import { SuppressDialog } from "@/pages/alert/suppress-dialog"
import type { SuppressInput } from "@/pages/alert/suppress-dialog"

interface AlertEventListProps {
  view: AlertEventView
  mine?: boolean
  pageUid?: string
  hideHeader?: boolean
  // refreshKey 变化时静默刷新列表；由实时页的自动轮询驱动。
  refreshKey?: number
  // 实时页从独立统计接口取数并下发，列表不再为统计条额外请求；
  // 历史页不传，仍由列表自己加载。
  stats?: AlertEventStats
  // 实时页下发的自定义列；为空时使用视图默认列。
  columns?: AlertColumnItem[]
  // 表头右侧的图标操作，实时页用来打开自定义列。
  columnsAction?: ReactNode
}

type ConfirmKind = "recover" | "batch-recover"

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
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [params, setParams] = useSearchParams()
  const level = view === "REALTIME" ? params.get("level") || "" : ""
  const [draft, setDraft] = useState(query.name)
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
    filterKey: `${session.namespaceUid}|${query.name}|${level}|${view}|${mine}|${pageUid}`,
    refreshKey,
    fetch: (page, pageSize) =>
      listAlertEvents({
        page,
        pageSize,
        name: query.name,
        view,
        mine: mine || undefined,
        pageUid: pageUid || undefined,
        levelUid: level || undefined,
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
      if (!sharedStats) {
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
  const canIntervene = view === "REALTIME" && session.can("alert:intervene")
  const visibleColumns =
    columns && columns.length > 0
      ? columns
      : view === "HISTORY"
        ? DEFAULT_HISTORY_COLUMNS
        : DEFAULT_REALTIME_COLUMNS
  const title =
    view === "HISTORY" ? t("alertEvent.historyTitle") : t("alertEvent.title")

  useEffect(() => {
    if (view !== "REALTIME") {
      return
    }
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
      setActionError(formMessage(err, t("state.failed")))
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
      setConfirmError(formMessage(err, t("state.failed")))
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
      setSuppressError(formMessage(err, t("state.failed")))
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

  const allChecked =
    items.length > 0 && items.every((item) => selected.includes(item.uid))

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
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
      <div className="flex flex-wrap items-center justify-between gap-2">
        <form
          className="flex flex-nowrap items-center gap-2"
          onSubmit={(event) => {
            event.preventDefault()
            query.update({ name: draft })
          }}
        >
          <Input
            className="max-w-sm min-w-0"
            value={draft}
            placeholder={t("alertEvent.namePlaceholder")}
            onChange={(event) => setDraft(event.target.value)}
          />
          {view === "REALTIME" ? (
            <Select
              value={level || "ALL"}
              onValueChange={(value) => {
                const next = new URLSearchParams(params)
                if (value === "ALL") {
                  next.delete("level")
                } else {
                  next.set("level", value)
                }
                next.set("page", "1")
                setParams(next)
              }}
            >
              <SelectTrigger
                className="w-40 shrink-0"
                aria-label={t("alertEvent.level")}
              >
                <SelectValue placeholder={t("alertEvent.allLevels")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ALL">{t("alertEvent.allLevels")}</SelectItem>
                {levels.map((item) => (
                  <SelectItem key={item.uid} value={item.uid}>
                    {item.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          ) : null}
          <Button type="submit" variant="outline" className="shrink-0">
            {t("action.search")}
          </Button>
        </form>
        <div className="flex flex-nowrap items-center justify-end gap-2">
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
                  setConfirm({ kind: "batch-recover", uids: selected })
                }
              >
                {t("alertEvent.batchRecover")}
              </Button>
            </>
          ) : null}
        </div>
      </div>
      {actionError ? (
        <p className="text-sm text-destructive">{actionError}</p>
      ) : null}
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        actions={columnsAction}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? (
          <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
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
          <p className="text-sm text-muted-foreground">
            {t("alertEvent.empty")}
          </p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                {canIntervene ? (
                  <TableHead className="w-10">
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
                {visibleColumns.map((column) => (
                  <TableHead key={`${column.type}:${column.key}`}>
                    {column.label.trim()
                      ? column.label
                      : column.type === "FIELD"
                        ? fieldLabel(t, column.key)
                        : column.key}
                  </TableHead>
                ))}
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  {canIntervene ? (
                    <TableCell>
                      <input
                        type="checkbox"
                        aria-label={item.summary || item.uid}
                        checked={selected.includes(item.uid)}
                        onChange={() => toggle(item.uid)}
                      />
                    </TableCell>
                  ) : null}
                  {visibleColumns.map((column) => (
                    <TableCell key={`${column.type}:${column.key}`}>
                      {eventCell(item, column, levels, now, t)}
                    </TableCell>
                  ))}
                  <TableCell>
                    <RowActions
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(item)}
                      actions={actions(item)}
                      forceMenu
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <EventChartDialog
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
        <SuppressDialog
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
    case "summary":
      return item.summary || item.fingerprint
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
      return item.sampleValue
    case "starts_at":
      return formatTime(item.startsAt)
    case "duration":
      return <span className="font-mono">{formatDuration(item.startsAt, now)}</span>
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
      return enumLabel(t, item.status)
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
    return uid
  }
  return (
    <span className="inline-flex items-center gap-2">
      <span
        className="size-2 shrink-0 rounded-full"
        style={{ backgroundColor: level.color || "#1677ff" }}
      />
      {level.name}
    </span>
  )
}
