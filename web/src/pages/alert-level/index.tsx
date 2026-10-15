import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"

import {
  createAlertLevel,
  deleteAlertLevel,
  listAlertLevels,
  updateAlertLevel,
  updateAlertLevelStatus,
} from "@/api/alert-level"
import type { AlertLevelItem } from "@/api/alert-level/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { ListFrame } from "@/components/list-frame"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
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
import { useListQuery } from "@/lib/list-query"

const types = ["ALERT", "DATASOURCE"]

export function AlertLevelPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [params, setParams] = useSearchParams()
  const type = params.get("type") || ""
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<AlertLevelItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AlertLevelItem | null>(null)
  const [name, setName] = useState("")
  const [color, setColor] = useState("")
  const [levelType, setLevelType] = useState("ALERT")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AlertLevelItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: AlertLevelItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const shown = useRef(confirm)
  if (confirm) {
    shown.current = confirm
  }
  const canWrite = session.can("alert_rule:write")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listAlertLevels({
          page: query.page,
          pageSize: query.pageSize,
          name: query.name,
          type,
        })
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
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
  }, [query.page, query.pageSize, query.name, type, attempt, session.namespaceUid])

  function openCreate() {
    setEditing(null)
    setName("")
    setColor("")
    setLevelType("ALERT")
    setError("")
    setOpen(true)
  }

  function openEdit(item: AlertLevelItem) {
    setEditing(item)
    setName(item.name)
    setColor(item.color)
    setLevelType(item.type)
    setError("")
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      if (editing) {
        await updateAlertLevel({ uid: editing.uid, name, color })
      } else {
        await createAlertLevel({ name, color, type: levelType })
      }
      setOpen(false)
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
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
        await deleteAlertLevel(confirm.item.uid)
      } else {
        await updateAlertLevelStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      setConfirmError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  function actions(item: AlertLevelItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => void updateAlertLevelStatus(item.uid, "ACTIVE").then(() => setAttempt((value) => value + 1)) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("alertLevel.title")} description={t("alertLevel.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input className="max-w-sm" value={draft} placeholder={t("alertLevel.namePlaceholder")} onChange={(event) => setDraft(event.target.value)} />
        <Select
          value={type || "ALL"}
          onValueChange={(value) => {
            const next = new URLSearchParams(params)
            if (value === "ALL") {
              next.delete("type")
            } else {
              next.set("type", value)
            }
            next.set("page", "1")
            setParams(next)
          }}
        >
          <SelectTrigger className="w-40" aria-label={t("datasource.type")}>
            <SelectValue placeholder={t("datasource.typePlaceholder")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="ALL">{t("datasource.typeAll")}</SelectItem>
            {types.map((item) => (
              <SelectItem key={item} value={item}>
                {item}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button type="submit" variant="outline">
          {t("action.search")}
        </Button>
      </form>
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
              <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("alertLevel.empty")}</p> : null}
          {!loading && !failed && items.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("alertLevel.name")}</TableHead>
                  <TableHead>{t("alertLevel.color")}</TableHead>
                  <TableHead>{t("datasource.type")}</TableHead>
                  <TableHead>{t("namespace.status")}</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.uid}>
                    <TableCell>{item.name}</TableCell>
                    <TableCell>{item.color}</TableCell>
                    <TableCell>{item.type}</TableCell>
                    <TableCell>{item.status}</TableCell>
                    <TableCell>
                      <RowActions
                        detailLabel={t("action.detail")}
                        moreLabel={t("action.more")}
                        onDetail={() => setDetail(item)}
                        actions={actions(item)}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : null}
      </ListFrame>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={onSave}>
            <FormField label={t("alertLevel.name")} htmlFor="level-name">
              <Input id="level-name" value={name} placeholder={t("alertLevel.namePlaceholder")} onChange={(event) => setName(event.target.value)} />
            </FormField>
            <FormField label={t("alertLevel.color")} htmlFor="level-color" hint={t("alertLevel.colorHint")}>
              <Input id="level-color" value={color} placeholder="#d92d20" onChange={(event) => setColor(event.target.value)} />
            </FormField>
            {!editing ? (
              <FormField label={t("datasource.type")}>
                <Select value={levelType} onValueChange={setLevelType}>
                  <SelectTrigger aria-label={t("datasource.type")}>
                    <SelectValue placeholder={t("datasource.typePlaceholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    {types.map((item) => (
                      <SelectItem key={item} value={item}>
                        {item}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            ) : null}
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <DetailSheet
        open={detail !== null}
        title={detail?.name ?? ""}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={
          detail
            ? [
                { label: t("detail.uid"), value: detail.uid },
                { label: t("alertLevel.color"), value: detail.color },
                { label: t("datasource.type"), value: detail.type },
                { label: t("namespace.status"), value: detail.status },
              ]
            : []
        }
      />
      <ConfirmAction
        open={confirm !== null}
        title={shown.current?.action === "delete" ? t("confirm.deleteTitle") : t("confirm.disableTitle")}
        description={shown.current?.action === "delete" ? t("confirm.deleteDescription") : t("confirm.disableDescription")}
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
