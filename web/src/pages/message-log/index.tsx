import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { failureText } from "@/lib/form-error"
import { useSearchParams } from "react-router-dom"

import {
  cancelNotificationTask,
  listMessageLogs,
  listNotificationTasks,
  retryNotificationTask,
} from "@/api/message-log"
import type { MessageLogItem, NotificationTaskItem } from "@/api/message-log/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { enumLabel } from "@/lib/enum-label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useListQuery } from "@/lib/list-query"

function deliveryContent(subject?: string, content?: string): string {
  return [subject?.trim(), content?.trim()].filter((value): value is string => Boolean(value)).join("\n\n")
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
  const [detailTask, setDetailTask] = useState<NotificationTaskItem | null>(null)
  const [confirmAction, setConfirmAction] = useState<{ kind: "retry" | "cancel"; item: NotificationTaskItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const statusFilter = params.get("status") ?? ""
  const taskFilter = params.get("taskUid") ?? ""
  const canWrite = session.can("notification:write")
  const taskStatuses = ["PENDING", "SENDING", "SENT", "FAILED", "CANCELLED"]

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
            status: statusFilter || undefined,
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
            taskUid: taskFilter || undefined,
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
          setLoadError(failureText(err, ""))
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
  }, [tab, query.page, query.pageSize, attempt, session.namespaceUid, statusFilter, taskFilter])

  function setTab(value: "logs" | "tasks") {
    const next = new URLSearchParams(params)
    if (value === "logs") {
      next.delete("tab")
    } else {
      next.set("tab", value)
    }
    next.set("page", "1")
    next.delete("status")
    next.delete("taskUid")
    setParams(next)
  }

  function setFilter(key: "status" | "taskUid", value: string) {
    const next = new URLSearchParams(params)
    if (value) {
      next.set(key, value)
    } else {
      next.delete(key)
    }
    next.set("page", "1")
    setParams(next)
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
    if (item.status !== "SENT" && item.status !== "CANCELLED" && item.status !== "SENDING") {
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
      <PageHeader title={t("messageLog.title")} description={t("messageLog.description")} />
      {actionError ? <p className="text-sm text-destructive">{actionError}</p> : null}
      <div className="flex gap-2">
        <Button variant={tab === "logs" ? "default" : "outline"} onClick={() => setTab("logs")}>
          {t("messageLog.logs")}
        </Button>
        <Button variant={tab === "tasks" ? "default" : "outline"} onClick={() => setTab("tasks")}>
          {t("messageLog.tasks")}
        </Button>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        {tab === "tasks" ? (
          <Select value={statusFilter || "ALL"} onValueChange={(value) => setFilter("status", value === "ALL" ? "" : value)}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder={t("messageLog.statusFilter")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">{t("messageLog.allStatus")}</SelectItem>
              {taskStatuses.map((item) => (
                <SelectItem key={item} value={item}>
                  {item}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : (
          <>
            <Input
              className="max-w-sm"
              value={taskFilter}
              placeholder={t("messageLog.taskFilterPlaceholder")}
              onChange={(event) => setFilter("taskUid", event.target.value)}
            />
            {taskFilter ? (
              <Button variant="outline" onClick={() => setFilter("taskUid", "")}>
                {t("messageLog.clearFilter")}
              </Button>
            ) : null}
          </>
        )}
      </div>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && (tab === "logs" ? logs.length > 0 : tasks.length > 0)}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
            <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && tab === "logs" && logs.length === 0 ? <p className="text-sm text-muted-foreground">{t("messageLog.empty")}</p> : null}
        {!loading && !failed && tab === "tasks" && tasks.length === 0 ? <p className="text-sm text-muted-foreground">{t("messageLog.empty")}</p> : null}
        {!loading && !failed && tab === "logs" && logs.length > 0 ? (
          <Table className="min-w-[56rem]">
            <TableHeader>
              <TableRow>
                <TableHead>{t("messageLog.status")}</TableHead>
                <TableHead>{t("messageLog.sender")}</TableHead>
                <TableHead>{t("messageLog.receivers")}</TableHead>
                <TableHead>{t("messageLog.content")}</TableHead>
                <TableHead>{t("messageLog.taskFilter")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{enumLabel(t, item.status)}</TableCell>
                  <TableCell className="max-w-xs truncate" title={item.sender || undefined}>{item.sender || t("messageLog.notProvided")}</TableCell>
                  <TableCell className="max-w-xs truncate" title={item.receivers || undefined}>{item.receivers || t("messageLog.notProvided")}</TableCell>
                  <TableCell className="max-w-sm truncate" title={deliveryContent(item.subject, item.content) || undefined}>{deliveryContent(item.subject, item.content) || t("messageLog.notProvided")}</TableCell>
                  <TableCell className="font-mono text-xs">{item.taskUid}</TableCell>
                  <TableCell>
                    <RowActions detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => setDetailLog(item)} actions={[]} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
        {!loading && !failed && tab === "tasks" && tasks.length > 0 ? (
          <Table className="min-w-[72rem]">
            <TableHeader>
              <TableRow>
                <TableHead>{t("messageLog.phase")}</TableHead>
                <TableHead>{t("messageLog.status")}</TableHead>
                <TableHead>{t("messageLog.sender")}</TableHead>
                <TableHead>{t("messageLog.receivers")}</TableHead>
                <TableHead>{t("messageLog.content")}</TableHead>
                <TableHead>{t("messageLog.retries")}</TableHead>
                <TableHead>{t("messageLog.detail")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {tasks.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{enumLabel(t, item.phase)}</TableCell>
                  <TableCell>{enumLabel(t, item.status)}</TableCell>
                  <TableCell className="max-w-xs truncate" title={item.sender || undefined}>{item.sender || t("messageLog.notProvided")}</TableCell>
                  <TableCell className="max-w-xs truncate" title={item.receivers || undefined}>{item.receivers || t("messageLog.notProvided")}</TableCell>
                  <TableCell className="max-w-sm truncate" title={deliveryContent(item.subject, item.content) || undefined}>{deliveryContent(item.subject, item.content) || t("messageLog.notProvided")}</TableCell>
                  <TableCell>{item.retriesLeft}</TableCell>
                  <TableCell className="max-w-xs truncate" title={item.lastError || undefined}>{item.lastError || t("messageLog.notProvided")}</TableCell>
                  <TableCell>
                    <RowActions detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => setDetailTask(item)} actions={taskActions(item)} />
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
                { label: t("messageLog.status"), value: enumLabel(t, detailLog.status) },
                { label: t("messageLog.sender"), value: detailLog.sender || t("messageLog.notProvided") },
                { label: t("messageLog.receivers"), value: detailLog.receivers || t("messageLog.notProvided") },
                { label: t("messageLog.subject"), value: detailLog.subject || t("messageLog.notProvided") },
                { label: t("messageLog.content"), value: <span className="whitespace-pre-wrap">{detailLog.content || t("messageLog.notProvided")}</span> },
                { label: t("messageLog.taskFilter"), value: detailLog.taskUid },
                { label: t("messageLog.detail"), value: detailLog.detail || t("messageLog.notProvided") },
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
                { label: t("messageLog.phase"), value: enumLabel(t, detailTask.phase) },
                { label: t("messageLog.status"), value: enumLabel(t, detailTask.status) },
                { label: t("messageLog.sender"), value: detailTask.sender || t("messageLog.notProvided") },
                { label: t("messageLog.receivers"), value: detailTask.receivers || t("messageLog.notProvided") },
                { label: t("messageLog.subject"), value: detailTask.subject || t("messageLog.notProvided") },
                { label: t("messageLog.content"), value: <span className="whitespace-pre-wrap">{detailTask.content || t("messageLog.notProvided")}</span> },
                { label: t("messageLog.retries"), value: detailTask.retriesLeft },
                { label: t("messageLog.detail"), value: detailTask.lastError || t("messageLog.notProvided") },
              ]
            : []
        }
      />
      <ConfirmAction
        open={confirmAction !== null}
        title={confirmAction?.kind === "retry" ? t("messageLog.retryTitle") : t("messageLog.cancelTitle")}
        description={confirmAction?.kind === "retry" ? t("messageLog.retryDescription") : t("messageLog.cancelDescription")}
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
