import { useEffect, useState } from "react"
import type { TFunction } from "i18next"
import { useSearchParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { formMessage } from "@/lib/form-error"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import {
  alertEventStats,
  interveneAlertEvent,
  interveneAlertEvents,
  listAlertEvents,
  recoverAlertEvent,
  recoverAlertEvents,
  suppressAlertEvent,
} from "@/api/alert-event"
import type { AlertEventItem, AlertEventStats, AlertEventView } from "@/api/alert-event/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
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

interface AlertEventListProps {
  view: AlertEventView
  mine?: boolean
  pageUid?: string
  hideHeader?: boolean
  // refreshKey 变化时静默刷新列表；由实时页的自动轮询驱动。
  refreshKey?: number
}

type ConfirmKind = "suppress" | "recover" | "batch-recover"

export function AlertEventList({ view, mine = false, pageUid = "", hideHeader = false, refreshKey = 0 }: AlertEventListProps) {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [params, setParams] = useSearchParams()
  const level = view === "REALTIME" ? params.get("level") || "" : ""
  const [draft, setDraft] = useState(query.name)
  const [stats, setStats] = useState<AlertEventStats>({ firing: 0, suppressed: 0, recovered: 0 })
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<AlertEventItem>({
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
      void alertEventStats().then(setStats).catch(() => undefined)
    },
  })
  const [detail, setDetail] = useState<AlertEventItem | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [pending, setPending] = useState(false)
  const [actionError, setActionError] = useState("")
  const [confirm, setConfirm] = useState<{ kind: ConfirmKind; uids: string[] } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [now, setNow] = useState(() => Date.now())
  const canIntervene = view === "REALTIME" && session.can("alert:intervene")
  const title = view === "HISTORY" ? t("alertEvent.historyTitle") : t("alertEvent.title")

  useEffect(() => {
    if (view !== "REALTIME") {
      return
    }
    let active = true
    async function loadLevels() {
      try {
        const result = await listAlertDictionaries({ type: "ALERT", page: 1, pageSize: 100 })
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
      if (confirm.kind === "suppress") {
        await suppressAlertEvent(confirm.uids[0])
      } else if (confirm.uids.length === 1) {
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

  function actions(item: AlertEventItem): RowAction[] {
    if (!canIntervene) {
      return []
    }
    return [
      { key: "intervene", label: t("action.intervene"), onSelect: () => void intervene([item.uid]) },
      { key: "suppress", label: t("action.suppress"), destructive: true, onSelect: () => setConfirm({ kind: "suppress", uids: [item.uid] }) },
      { key: "recover", label: t("action.recover"), destructive: true, onSelect: () => setConfirm({ kind: "recover", uids: [item.uid] }) },
    ]
  }

  function toggle(uid: string) {
    setSelected((current) => (current.includes(uid) ? current.filter((item) => item !== uid) : [...current, uid]))
  }

  const allChecked = items.length > 0 && items.every((item) => selected.includes(item.uid))

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      {hideHeader ? null : (
        <PageHeader
          title={title}
          description={view === "HISTORY" ? t("alertEvent.historyDescription") : t("alertEvent.description")}
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
            className="min-w-0 max-w-sm"
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
              <SelectTrigger className="w-40 shrink-0" aria-label={t("alertEvent.level")}>
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
          <span className="shrink-0 text-sm text-muted-foreground">{t("alertEvent.stats", { firing: stats.firing, suppressed: stats.suppressed, recovered: stats.recovered })}</span>
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
                onClick={() => setConfirm({ kind: "batch-recover", uids: selected })}
              >
                {t("alertEvent.batchRecover")}
              </Button>
            </>
          ) : null}
        </div>
      </div>
      {actionError ? <p className="text-sm text-destructive">{actionError}</p> : null}
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
            <Button variant="outline" onClick={() => reload()}>{t("action.retry")}</Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("alertEvent.empty")}</p> : null}
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
                      onChange={() => setSelected(allChecked ? [] : items.map((item) => item.uid))}
                    />
                  </TableHead>
                ) : null}
                <TableHead>{t("alertEvent.summary")}</TableHead>
                {view === "REALTIME" ? <TableHead>{t("alertEvent.level")}</TableHead> : null}
                {view === "HISTORY" ? <TableHead>{t("alertEvent.status")}</TableHead> : null}
                <TableHead>{t("alertEvent.sample")}</TableHead>
                <TableHead>{t("alertEvent.startsAt")}</TableHead>
                {view === "REALTIME" ? <TableHead>{t("alertEvent.duration")}</TableHead> : null}
                {view === "REALTIME" ? <TableHead>{t("alertEvent.intervened")}</TableHead> : null}
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  {canIntervene ? (
                    <TableCell>
                      <input type="checkbox" aria-label={item.summary || item.uid} checked={selected.includes(item.uid)} onChange={() => toggle(item.uid)} />
                    </TableCell>
                  ) : null}
                  <TableCell>{item.summary || item.fingerprint}</TableCell>
                  {view === "REALTIME" ? <TableCell>{levelCell(item.levelUid, levels)}</TableCell> : null}
                  {view === "HISTORY" ? <TableCell>{enumLabel(t, item.status)}</TableCell> : null}
                  <TableCell>{item.sampleValue}</TableCell>
                  <TableCell>{formatTime(item.startsAt)}</TableCell>
                  {view === "REALTIME" ? <TableCell className="font-mono">{formatDuration(item.startsAt, now)}</TableCell> : null}
                  {view === "REALTIME" ? (
                    <TableCell>
                      {item.intervenedBy ? (
                        <div className="grid">
                          <span>{item.intervenedName || item.intervenedBy}</span>
                          <span className="text-xs text-muted-foreground">{formatTime(item.intervenedAt)}</span>
                        </div>
                      ) : null}
                    </TableCell>
                  ) : null}
                  <TableCell>
                    <RowActions detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => setDetail(item)} actions={actions(item)} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <DetailSheet
        open={detail !== null}
        title={detail?.summary || title}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={detail ? detailFields(detail, t) : []}
      />
      <ConfirmAction
        open={confirm !== null}
        title={confirm?.kind === "suppress" ? t("confirm.suppressTitle") : t("confirm.recoverTitle")}
        description={confirm?.kind === "suppress" ? t("confirm.suppressDescription") : t("confirm.recoverDescription")}
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
      <span className="size-2 shrink-0 rounded-full" style={{ backgroundColor: level.color || "#1677ff" }} />
      {level.name}
    </span>
  )
}

function formatDuration(start: string, now: number) {
  const from = new Date(start).getTime()
  if (Number.isNaN(from)) {
    return ""
  }
  let seconds = Math.max(0, Math.floor((now - from) / 1000))
  const hours = Math.floor(seconds / 3600)
  seconds -= hours * 3600
  const minutes = Math.floor(seconds / 60)
  seconds -= minutes * 60
  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  }
  return `${seconds}s`
}

function formatTime(value: string) {
  if (!value) {
    return ""
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

function detailFields(item: AlertEventItem, t: TFunction) {
  return [
    { label: t("detail.uid"), value: item.uid },
    { label: t("alertEvent.rule"), value: item.ruleUid },
    { label: t("alertEvent.status"), value: enumLabel(t, item.status) },
    { label: t("alertEvent.fingerprint"), value: item.fingerprint },
    { label: t("alertEvent.sample"), value: item.sampleValue },
    { label: t("alertEvent.expr"), value: item.expr },
    { label: t("alertEvent.labels"), value: item.labels },
    { label: t("alertEvent.descriptionField"), value: item.description },
    { label: t("alertEvent.level"), value: item.levelUid },
    { label: t("alertEvent.startsAt"), value: formatTime(item.startsAt) },
    { label: t("alertEvent.endsAt"), value: formatTime(item.endsAt) },
    { label: t("alertEvent.suppressed"), value: item.suppressed ? t("alertEvent.yes") : t("alertEvent.no") },
    { label: t("alertEvent.intervenedBy"), value: item.intervenedName || item.intervenedBy },
    { label: t("alertEvent.intervenedAt"), value: formatTime(item.intervenedAt) },
  ]
}
