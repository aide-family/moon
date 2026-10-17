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
import { DetailSheet } from "@/components/detail-sheet"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { enumLabel } from "@/lib/enum-label"
import { useListQuery } from "@/lib/list-query"

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
  const [attempt, setAttempt] = useState(0)
  const [detailLog, setDetailLog] = useState<MessageLogItem | null>(null)
  const [detailTask, setDetailTask] = useState<NotificationTaskItem | null>(null)
  const canWrite = session.can("notification:write")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        if (tab === "tasks") {
          const result = await listNotificationTasks({ page: query.page, pageSize: query.pageSize })
          if (!active) {
            return
          }
          setTasks(result.items ?? [])
          setTotal(result.total ?? 0)
        } else {
          const result = await listMessageLogs({ page: query.page, pageSize: query.pageSize })
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
  }, [tab, query.page, query.pageSize, attempt, session.namespaceUid])

  function setTab(value: "logs" | "tasks") {
    const next = new URLSearchParams(params)
    if (value === "logs") {
      next.delete("tab")
    } else {
      next.set("tab", value)
    }
    next.set("page", "1")
    setParams(next)
  }

  function taskActions(item: NotificationTaskItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    const actions: RowAction[] = []
    if (item.status === "FAILED") {
      actions.push({
        key: "retry",
        label: t("action.retry"),
        onSelect: () => void retryNotificationTask(item.uid).then(() => setAttempt((value) => value + 1)),
      })
    }
    if (item.status !== "SENT" && item.status !== "CANCELLED") {
      actions.push({
        key: "cancel",
        label: t("action.cancelTask"),
        destructive: true,
        onSelect: () => void cancelNotificationTask(item.uid).then(() => setAttempt((value) => value + 1)),
      })
    }
    return actions
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("messageLog.title")} description={t("messageLog.description")} />
      <div className="flex gap-2">
        <Button variant={tab === "logs" ? "default" : "outline"} onClick={() => setTab("logs")}>
          {t("messageLog.logs")}
        </Button>
        <Button variant={tab === "tasks" ? "default" : "outline"} onClick={() => setTab("tasks")}>
          {t("messageLog.tasks")}
        </Button>
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
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("messageLog.status")}</TableHead>
                <TableHead>{t("messageLog.receivers")}</TableHead>
                <TableHead>{t("messageLog.detail")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{enumLabel(t, item.status)}</TableCell>
                  <TableCell>{item.receivers}</TableCell>
                  <TableCell className="max-w-xs truncate">{item.detail}</TableCell>
                  <TableCell>
                    <RowActions detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => setDetailLog(item)} actions={[]} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
        {!loading && !failed && tab === "tasks" && tasks.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("messageLog.phase")}</TableHead>
                <TableHead>{t("messageLog.status")}</TableHead>
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
                  <TableCell>{item.retriesLeft}</TableCell>
                  <TableCell className="max-w-xs truncate">{item.lastError}</TableCell>
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
                { label: t("messageLog.receivers"), value: detailLog.receivers },
                { label: t("messageLog.detail"), value: detailLog.detail },
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
                { label: t("messageLog.detail"), value: detailTask.lastError },
              ]
            : []
        }
      />
    </section>
  )
}
