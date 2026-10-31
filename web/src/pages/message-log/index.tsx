import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { failureMessage } from "@/lib/form-error"
import { useSearchParams } from "react-router-dom"

import {
  cancelNotificationTask,
  listMessageLogs,
  listNotificationTasks,
  retryNotificationTask,
} from "@/api/message-log"
import type {
  MessageLogItem,
  NotificationTaskItem,
} from "@/api/message-log/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DateRangePicker } from "@/components/date-range-picker"
import { FilterChips, type FilterChip } from "@/components/filter-chips"
import { FilterToggle } from "@/components/filter-toggle"
import { DetailSheet } from "@/components/detail-sheet"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import { Field, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { dayStartSeconds, nextDayStartSeconds } from "@/lib/date-range"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"
import { formatDateTime } from "@/lib/date-time"
import { enumLabel } from "@/lib/enum-label"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Search } from "lucide-react"
import { useListQuery } from "@/hooks/use-list-query"

interface MessageLogFilters {
  status: string
  taskUid: string
  from: string
  to: string
}

function deliveryContent(subject?: string, content?: string): string {
  return [subject?.trim(), content?.trim()]
    .filter((value): value is string => Boolean(value))
    .join("\n\n")
}

export function MessageLogPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [params, setParams] = useSearchParams()
  const tab = params.get("tab") === "tasks" ? "tasks" : "logs"
  const [logs, setLogs] = useState<MessageLogItem[]>([])
  const [tasks, setTasks] = useState<NotificationTaskItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [actionError, setActionError] = useState("")
  const [pending, setPending] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [detailLog, setDetailLog] = useState<MessageLogItem | null>(null)
  const [detailTask, setDetailTask] = useState<NotificationTaskItem | null>(
    null
  )
  const [confirmAction, setConfirmAction] = useState<{
    kind: "retry" | "cancel"
    item: NotificationTaskItem
  } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  // 筛选条件同样以 URL 为唯一来源：面板里编辑的是草稿，点“应用筛选”才生效。
  const appliedFilters: MessageLogFilters = {
    status: params.get("status") ?? "",
    taskUid: params.get("taskUid") ?? "",
    from: params.get("from") ?? "",
    to: params.get("to") ?? "",
  }
  const appliedSignature = JSON.stringify(appliedFilters)
  const [draftState, setDraftState] = useState<{
    signature: string
    filters: MessageLogFilters
  } | null>(null)
  const filters =
    draftState?.signature === appliedSignature
      ? draftState.filters
      : appliedFilters
  const filtersDirty = JSON.stringify(filters) !== appliedSignature
  const appliedCount = Object.values(appliedFilters).filter(Boolean).length
  // 默认无筛选时展开，应用后收起；与历史告警明细一致。
  const [filtersOpen, setFiltersOpen] = useState(appliedCount === 0)
  const canWrite = session.can("message_log:write")
  const taskStatuses = ["PENDING", "SENDING", "SENT", "FAILED", "CANCELLED"]
  const logStatuses = ["SENT", "FAILED"]

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        if (tab === "tasks") {
          const result = await listNotificationTasks({
            page: query.page,
            pageSize: query.pageSize,
            status: appliedFilters.status || undefined,
            startTime: dayStartSeconds(appliedFilters.from),
            endTime: nextDayStartSeconds(appliedFilters.to),
          })
          if (!active) {
            return
          }
          setTasks(result.items ?? [])
          setTotal(result.total ?? 0)
        } else {
          const result = await listMessageLogs({
            page: query.page,
            pageSize: query.pageSize,
            taskUid: appliedFilters.taskUid || undefined,
            status: appliedFilters.status || undefined,
            startTime: dayStartSeconds(appliedFilters.from),
            endTime: nextDayStartSeconds(appliedFilters.to),
          })
          if (!active) {
            return
          }
          setLogs(result.items ?? [])
          setTotal(result.total ?? 0)
        }
      } catch (err) {
        if (active) {
          setFailed(true)
          setLoadError(failureMessage(err, ""))
          setLogs([])
          setTasks([])
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [
    tab,
    query.page,
    query.pageSize,
    attempt,
    session.namespaceUid,
    appliedFilters.status,
    appliedFilters.taskUid,
    appliedFilters.from,
    appliedFilters.to,
  ])

  function setTab(value: "logs" | "tasks") {
    const next = new URLSearchParams(params)
    if (value === "logs") {
      next.delete("tab")
    } else {
      next.set("tab", value)
    }
    next.set("page", "1")
    // 两个视图的筛选条件不通用（日志有任务标识，任务没有），切换时清空。
    for (const key of ["status", "taskUid", "from", "to"]) {
      next.delete(key)
    }
    setParams(next)
    setDraftState(null)
  }

  function updateDraft(key: keyof MessageLogFilters, value: string) {
    setDraftState({
      signature: appliedSignature,
      filters: { ...filters, [key]: value },
    })
  }

  function applyFilters() {
    const next = new URLSearchParams(params)
    for (const key of ["status", "taskUid", "from", "to"] as const) {
      const value = filters[key]
      if (value) {
        next.set(key, value)
      } else {
        next.delete(key)
      }
    }
    next.set("page", "1")
    setParams(next)
    setDraftState(null)
    setFiltersOpen(false)
  }

  function removeFilters(keys: string[]) {
    const next = new URLSearchParams(params)
    for (const key of keys) {
      next.delete(key)
    }
    next.set("page", "1")
    setParams(next)
    setDraftState(null)
  }

  const filterChips: FilterChip[] = []
  if (appliedFilters.status) {
    filterChips.push({
      key: "status",
      label: `${t("messageLog.statusFilter")}: ${enumLabel(t, appliedFilters.status)}`,
      onRemove: () => removeFilters(["status"]),
    })
  }
  if (appliedFilters.taskUid) {
    filterChips.push({
      key: "taskUid",
      label: `${t("messageLog.taskFilter")}: ${appliedFilters.taskUid}`,
      onRemove: () => removeFilters(["taskUid"]),
    })
  }
  if (appliedFilters.from || appliedFilters.to) {
    filterChips.push({
      key: "range",
      label: `${t("filter.timeRange")}: ${appliedFilters.from || "…"} → ${appliedFilters.to || "…"}`,
      onRemove: () => removeFilters(["from", "to"]),
    })
  }

  function taskActions(item: NotificationTaskItem): RowAction[] {
    const actions: RowAction[] = [
      {
        key: "logs",
        label: t("messageLog.viewLogs"),
        onSelect: () => {
          const next = new URLSearchParams(params)
          next.set("tab", "logs")
          next.set("taskUid", item.uid)
          next.set("page", "1")
          setParams(next)
        },
      },
    ]
    if (!canWrite) {
      return actions
    }
    if (item.status === "FAILED") {
      actions.push({
        key: "retry",
        label: t("action.retry"),
        onSelect: () => {
          setConfirmError("")
          setConfirmAction({ kind: "retry", item })
        },
      })
    }
    if (
      item.status !== "SENT" &&
      item.status !== "CANCELLED" &&
      item.status !== "SENDING"
    ) {
      actions.push({
        key: "cancel",
        label: t("action.cancelTask"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setConfirmAction({ kind: "cancel", item })
        },
      })
    }
    return actions
  }

  async function runConfirm() {
    if (!confirmAction) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      if (confirmAction.kind === "retry") {
        await retryNotificationTask(confirmAction.item.uid)
      } else {
        await cancelNotificationTask(confirmAction.item.uid)
      }
      setConfirmAction(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      const message = err instanceof Error ? err.message : t("state.failed")
      setConfirmError(message)
      setActionError(message)
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("messageLog.title")}
        description={t("messageLog.description")}
      />
      {actionError ? (
        <p className="text-sm text-destructive">{actionError}</p>
      ) : null}
      <div className="flex min-w-0 flex-col gap-2">
        <div className="flex flex-wrap items-center justify-end gap-2">
          <FilterToggle
            open={filtersOpen}
            count={appliedCount}
            controls="message-log-filters"
            onClick={() => setFiltersOpen((open) => !open)}
          />
          <Tabs
            value={tab}
            onValueChange={(value) => setTab(value as "logs" | "tasks")}
          >
            <TabsList>
              <TabsTrigger value="logs">{t("messageLog.logs")}</TabsTrigger>
              <TabsTrigger value="tasks">{t("messageLog.tasks")}</TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
        {filterChips.length > 0 ? (
          <FilterChips
            items={filterChips}
            onClearAll={() =>
              removeFilters(["status", "taskUid", "from", "to"])
            }
          />
        ) : null}
        <form
          id="message-log-filters"
          onSubmit={(event) => {
            event.preventDefault()
            applyFilters()
          }}
          className={`flex min-w-0 flex-wrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3 sm:p-4 ${filtersOpen ? "" : "hidden"}`}
        >
          {tab === "tasks" ? (
            <Field className="w-44 shrink-0">
              <FieldLabel
                htmlFor="log-status"
                className="text-xs text-muted-foreground"
              >
                {t("messageLog.statusFilter")}
              </FieldLabel>
              <Select
                value={filters.status || "ALL"}
                onValueChange={(value) =>
                  updateDraft("status", value === "ALL" ? "" : value)
                }
              >
                <SelectTrigger id="log-status" className="w-full">
                  <SelectValue placeholder={t("messageLog.allStatus")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="ALL">
                      {t("messageLog.allStatus")}
                    </SelectItem>
                    {taskStatuses.map((item) => (
                      <SelectItem key={item} value={item}>
                        {enumLabel(t, item)}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
          ) : (
            <>
              <Field className="w-40 shrink-0">
                <FieldLabel
                  htmlFor="log-status"
                  className="text-xs text-muted-foreground"
                >
                  {t("messageLog.statusFilter")}
                </FieldLabel>
                <Select
                  value={filters.status || "ALL"}
                  onValueChange={(value) =>
                    updateDraft("status", value === "ALL" ? "" : value)
                  }
                >
                  <SelectTrigger id="log-status" className="w-full">
                    <SelectValue placeholder={t("messageLog.allStatus")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="ALL">
                        {t("messageLog.allStatus")}
                      </SelectItem>
                      {logStatuses.map((item) => (
                        <SelectItem key={item} value={item}>
                          {enumLabel(t, item)}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
              <Field className="min-w-0 flex-1 basis-56">
                <FieldLabel
                  htmlFor="log-task"
                  className="text-xs text-muted-foreground"
                >
                  {t("messageLog.taskFilter")}
                </FieldLabel>
                <Input
                  id="log-task"
                  value={filters.taskUid}
                  placeholder={t("messageLog.taskFilterPlaceholder")}
                  onChange={(event) =>
                    updateDraft("taskUid", event.target.value)
                  }
                />
              </Field>
            </>
          )}
          <Field className="w-64 shrink-0">
            <FieldLabel
              htmlFor="log-range"
              className="text-xs text-muted-foreground"
            >
              {t("filter.timeRange")}
            </FieldLabel>
            {/* 默认上限就是 30 天，且不能选未来日期。 */}
            <DateRangePicker
              id="log-range"
              from={filters.from}
              to={filters.to}
              onChange={(range) =>
                setDraftState({
                  signature: appliedSignature,
                  filters: { ...filters, from: range.from, to: range.to },
                })
              }
            />
          </Field>
          <div className="flex items-center gap-2">
            {filtersDirty ? (
              <span
                className="text-xs text-amber-600 dark:text-amber-400"
                role="status"
              >
                {t("filter.unapplied")}
              </span>
            ) : null}
            {appliedCount > 0 ? (
              <Button
                type="button"
                variant="ghost"
                onClick={() =>
                  removeFilters(["status", "taskUid", "from", "to"])
                }
              >
                {t("filter.clear")}
              </Button>
            ) : null}
            <Button type="submit" variant="outline">
              <Search aria-hidden="true" data-icon="inline-start" />
              {t("filter.apply")}
            </Button>
          </div>
        </form>
      </div>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={
          !loading &&
          !failed &&
          (tab === "logs" ? logs.length > 0 : tasks.length > 0)
        }
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
            <Button
              variant="outline"
              onClick={() => setAttempt((value) => value + 1)}
            >
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && tab === "logs" && logs.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            {t("messageLog.empty")}
          </p>
        ) : null}
        {!loading && !failed && tab === "tasks" && tasks.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            {t("messageLog.empty")}
          </p>
        ) : null}
        {!loading && !failed && tab === "logs" && logs.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableStatusHead>{t("messageLog.status")}</TableStatusHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                >
                  {t("messageLog.createdAt")}
                </TableHead>
                <TableHead>{t("messageLog.sender")}</TableHead>
                <TableHead>{t("messageLog.receivers")}</TableHead>
                <TableHead>{t("messageLog.content")}</TableHead>
                <TableHead>{t("messageLog.group")}</TableHead>
                <TableHead>{t("messageLog.subscription")}</TableHead>
                <TableHead>{t("messageLog.taskFilter")}</TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>
                    <StatusBadge value={item.status} />
                  </TableCell>
                  <TableCell
                    className={`font-mono text-xs tabular-nums ${TABLE_NOWRAP}`}
                  >
                    {formatDateTime(item.createdAt)}
                  </TableCell>
                  <TableCell
                    className="max-w-xs truncate"
                    title={item.sender || undefined}
                  >
                    {item.sender || t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell
                    className="max-w-xs truncate"
                    title={item.receivers || undefined}
                  >
                    {item.receivers || t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell
                    className="max-w-sm truncate"
                    title={
                      deliveryContent(item.subject, item.content) || undefined
                    }
                  >
                    {deliveryContent(item.subject, item.content) ||
                      t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell
                    className="max-w-xs truncate"
                    title={item.groupName || undefined}
                  >
                    {item.groupName || t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell
                    className="max-w-xs truncate"
                    title={item.subscriptionName || undefined}
                  >
                    {item.subscriptionName || t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {item.taskUid}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetailLog(item)}
                      actions={[]}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
        {!loading && !failed && tab === "tasks" && tasks.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("messageLog.phase")}</TableHead>
                <TableStatusHead>{t("messageLog.status")}</TableStatusHead>
                <TableHead>{t("messageLog.sender")}</TableHead>
                <TableHead>{t("messageLog.receivers")}</TableHead>
                <TableHead>{t("messageLog.group")}</TableHead>
                <TableHead>{t("messageLog.subscription")}</TableHead>
                <TableHead>{t("messageLog.content")}</TableHead>
                <TableHead>{t("messageLog.retries")}</TableHead>
                <TableHead>{t("messageLog.detail")}</TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tasks.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>
                    <StatusBadge value={item.phase} />
                  </TableCell>
                  <TableCell>
                    <StatusBadge value={item.status} />
                  </TableCell>
                  <TableCell
                    className="max-w-xs truncate"
                    title={item.sender || undefined}
                  >
                    {item.sender || t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell
                    className="max-w-xs truncate"
                    title={item.receivers || undefined}
                  >
                    {item.receivers || t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell
                    className="max-w-xs truncate"
                    title={item.groupName || item.groupUid}
                  >
                    {item.groupName || item.groupUid}
                  </TableCell>
                  <TableCell
                    className="max-w-xs truncate"
                    title={item.subscriptionName || item.subscriptionUid}
                  >
                    {item.subscriptionName ||
                      item.subscriptionUid ||
                      t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell
                    className="max-w-sm truncate"
                    title={
                      deliveryContent(item.subject, item.content) || undefined
                    }
                  >
                    {deliveryContent(item.subject, item.content) ||
                      t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell>{item.retriesLeft}</TableCell>
                  <TableCell
                    className="max-w-xs truncate"
                    title={item.lastError || undefined}
                  >
                    {item.lastError || t("messageLog.notProvided")}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetailTask(item)}
                      actions={taskActions(item)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <DetailSheet
        open={detailLog !== null}
        title={detailLog?.uid ?? ""}
        onOpenChange={(value) => {
          if (!value) {
            setDetailLog(null)
          }
        }}
        fields={
          detailLog
            ? [
                { label: t("detail.uid"), value: detailLog.uid },
                {
                  label: t("messageLog.status"),
                  value: <StatusBadge value={detailLog.status} />,
                },
                {
                  label: t("messageLog.createdAt"),
                  value: formatDateTime(detailLog.createdAt),
                },
                {
                  label: t("messageLog.sender"),
                  value: detailLog.sender || t("messageLog.notProvided"),
                },
                {
                  label: t("messageLog.receivers"),
                  value: detailLog.receivers || t("messageLog.notProvided"),
                },
                {
                  label: t("messageLog.subject"),
                  value: detailLog.subject || t("messageLog.notProvided"),
                },
                {
                  label: t("messageLog.content"),
                  value: (
                    <span className="whitespace-pre-wrap">
                      {detailLog.content || t("messageLog.notProvided")}
                    </span>
                  ),
                },
                {
                  label: t("messageLog.group"),
                  value: detailLog.groupName || t("messageLog.notProvided"),
                },
                {
                  label: t("messageLog.subscription"),
                  value:
                    detailLog.subscriptionName || t("messageLog.notProvided"),
                },
                { label: t("messageLog.taskFilter"), value: detailLog.taskUid },
                {
                  label: t("messageLog.detail"),
                  value: detailLog.detail || t("messageLog.notProvided"),
                },
              ]
            : []
        }
      />
      <DetailSheet
        open={detailTask !== null}
        title={detailTask?.uid ?? ""}
        onOpenChange={(value) => {
          if (!value) {
            setDetailTask(null)
          }
        }}
        fields={
          detailTask
            ? [
                { label: t("detail.uid"), value: detailTask.uid },
                {
                  label: t("messageLog.phase"),
                  value: <StatusBadge value={detailTask.phase} />,
                },
                {
                  label: t("messageLog.status"),
                  value: <StatusBadge value={detailTask.status} />,
                },
                {
                  label: t("messageLog.sender"),
                  value: detailTask.sender || t("messageLog.notProvided"),
                },
                {
                  label: t("messageLog.receivers"),
                  value: detailTask.receivers || t("messageLog.notProvided"),
                },
                {
                  label: t("messageLog.subject"),
                  value: detailTask.subject || t("messageLog.notProvided"),
                },
                {
                  label: t("messageLog.content"),
                  value: (
                    <span className="whitespace-pre-wrap">
                      {detailTask.content || t("messageLog.notProvided")}
                    </span>
                  ),
                },
                {
                  label: t("messageLog.group"),
                  value: detailTask.groupName || detailTask.groupUid,
                },
                {
                  label: t("messageLog.subscription"),
                  value:
                    detailTask.subscriptionName ||
                    detailTask.subscriptionUid ||
                    t("messageLog.notProvided"),
                },
                {
                  label: t("messageLog.channel"),
                  value: detailTask.channelName || detailTask.channelUid,
                },
                {
                  label: t("messageLog.template"),
                  value: detailTask.templateName || detailTask.templateUid,
                },
                {
                  label: t("messageLog.retries"),
                  value: detailTask.retriesLeft,
                },
                {
                  label: t("messageLog.detail"),
                  value: detailTask.lastError || t("messageLog.notProvided"),
                },
              ]
            : []
        }
      />
      <ConfirmAction
        open={confirmAction !== null}
        title={
          confirmAction?.kind === "retry"
            ? t("messageLog.retryTitle")
            : t("messageLog.cancelTitle")
        }
        description={
          confirmAction?.kind === "retry"
            ? t("messageLog.retryDescription")
            : t("messageLog.cancelDescription")
        }
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void runConfirm()}
        onOpenChange={(value) => {
          if (!value) {
            setConfirmAction(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
