import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useListQuery } from "@/lib/list-query"

interface AlertEventListProps {
  view: AlertEventView
}

type ConfirmKind = "suppress" | "recover" | "batch-recover"

export function AlertEventList({ view }: AlertEventListProps) {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<AlertEventItem[]>([])
  const [total, setTotal] = useState(0)
  const [stats, setStats] = useState<AlertEventStats>({ firing: 0, suppressed: 0, recovered: 0 })
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [detail, setDetail] = useState<AlertEventItem | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [pending, setPending] = useState(false)
  const [actionError, setActionError] = useState("")
  const [confirm, setConfirm] = useState<{ kind: ConfirmKind; uids: string[] } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const shown = useRef(confirm)
  if (confirm) {
    shown.current = confirm
  }
  const canIntervene = view === "REALTIME" && session.can("alert:intervene")
  const title = view === "HISTORY" ? t("alertEvent.historyTitle") : t("alertEvent.title")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listAlertEvents({
          page: query.page,
          pageSize: query.pageSize,
          name: query.name,
          view,
        })
        const counts = await alertEventStats()
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
        setStats(counts)
        setSelected([])
      } catch {
        if (active) {
          setFailed(true)
          setItems([])
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
  }, [query.page, query.pageSize, query.name, view, attempt, session.namespaceUid])

  async function intervene(uids: string[]) {
    setPending(true)
    setActionError("")
    try {
      if (uids.length === 1) {
        await interveneAlertEvent(uids[0])
      } else {
        await interveneAlertEvents({ uids })
      }
      setAttempt((value) => value + 1)
    } catch (err) {
      setActionError(err instanceof Error ? err.message : t("state.failed"))
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
      setAttempt((value) => value + 1)
    } catch (err) {
      setConfirmError(err instanceof Error ? err.message : t("state.failed"))
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
      <PageHeader title={title} description={view === "HISTORY" ? t("alertEvent.historyDescription") : t("alertEvent.description")}>
        <div className="flex flex-wrap items-center justify-end gap-2">
          <span className="text-sm text-muted-foreground">
            {t("alertEvent.stats", stats)}
          </span>
          {canIntervene ? (
            <>
              <Button variant="outline" disabled={pending || selected.length === 0} onClick={() => void intervene(selected)}>
                {t("alertEvent.batchIntervene")}
              </Button>
              <Button variant="outline" disabled={pending || selected.length === 0} onClick={() => setConfirm({ kind: "batch-recover", uids: selected })}>
                {t("alertEvent.batchRecover")}
              </Button>
            </>
          ) : null}
        </div>
      </PageHeader>
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input className="max-w-sm" value={draft} placeholder={t("alertEvent.namePlaceholder")} onChange={(event) => setDraft(event.target.value)} />
        <Button type="submit" variant="outline">{t("action.search")}</Button>
      </form>
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
            <p className="text-sm text-destructive">{t("state.failed")}</p>
            <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>{t("action.retry")}</Button>
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
                <TableHead>{t("alertEvent.status")}</TableHead>
                <TableHead>{t("alertEvent.sample")}</TableHead>
                <TableHead>{t("alertEvent.startsAt")}</TableHead>
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
                  <TableCell>{item.status}</TableCell>
                  <TableCell>{item.sampleValue}</TableCell>
                  <TableCell>{formatTime(item.startsAt)}</TableCell>
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
        title={shown.current?.kind === "suppress" ? t("confirm.suppressTitle") : t("confirm.recoverTitle")}
        description={shown.current?.kind === "suppress" ? t("confirm.suppressDescription") : t("confirm.recoverDescription")}
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

function detailFields(item: AlertEventItem, t: (key: string) => string) {
  return [
    { label: t("detail.uid"), value: item.uid },
    { label: t("alertEvent.rule"), value: item.ruleUid },
    { label: t("alertEvent.status"), value: item.status },
    { label: t("alertEvent.fingerprint"), value: item.fingerprint },
    { label: t("alertEvent.sample"), value: item.sampleValue },
    { label: t("alertEvent.expr"), value: item.expr },
    { label: t("alertEvent.labels"), value: item.labels },
    { label: t("alertEvent.descriptionField"), value: item.description },
    { label: t("alertEvent.level"), value: item.levelUid },
    { label: t("alertEvent.startsAt"), value: formatTime(item.startsAt) },
    { label: t("alertEvent.endsAt"), value: formatTime(item.endsAt) },
    { label: t("alertEvent.suppressed"), value: item.suppressed ? t("alertEvent.yes") : t("alertEvent.no") },
    { label: t("alertEvent.intervenedBy"), value: item.intervenedBy },
  ]
}
