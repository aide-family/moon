import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { failureText, formMessage } from "@/lib/form-error"

import {
  createAlertPage,
  deleteAlertPage,
  listAlertPages,
  updateAlertPage,
} from "@/api/alert-page"
import type { AlertPageItem } from "@/api/alert-page/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
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
import { firstError, requireTrimmed } from "@/lib/form"
import { useListQuery } from "@/lib/list-query"

interface LabelPair {
  key: string
  value: string
}

function parseLabels(raw: string): LabelPair[] {
  try {
    const data = JSON.parse(raw || "{}") as Record<string, string>
    const pairs = Object.entries(data).map(([key, value]) => ({ key, value: String(value) }))
    return pairs.length > 0 ? pairs : [{ key: "", value: "" }]
  } catch {
    return [{ key: "", value: "" }]
  }
}

function serializeLabels(pairs: LabelPair[]): string {
  const out: Record<string, string> = {}
  for (const pair of pairs) {
    const key = pair.key.trim()
    if (!key) {
      continue
    }
    out[key] = pair.value.trim()
  }
  return JSON.stringify(out)
}

export function AlertPageManagePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<AlertPageItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [attempt, setAttempt] = useState(0)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AlertPageItem | null>(null)
  const [name, setName] = useState("")
  const [labels, setLabels] = useState<LabelPair[]>([{ key: "", value: "" }])
  const [isPublic, setIsPublic] = useState(false)
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AlertPageItem | null>(null)
  const [confirm, setConfirm] = useState<AlertPageItem | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const shown = useRef(confirm)
  if (confirm) {
    shown.current = confirm
  }
  const canWrite = session.can("alert:write")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listAlertPages({
          page: query.page,
          pageSize: query.pageSize,
          name: query.name,
          scope: "manage",
        })
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
      } catch (err) {
        if (active) {
          setFailed(true)
          setLoadError(failureText(err, ""))
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
  }, [query.page, query.pageSize, query.name, attempt, session.namespaceUid])

  function openCreate() {
    setEditing(null)
    setName("")
    setLabels([{ key: "", value: "" }])
    setIsPublic(false)
    setRemark("")
    setError("")
    setOpen(true)
  }

  function openEdit(item: AlertPageItem) {
    setEditing(item)
    setName(item.name)
    setLabels(parseLabels(item.labels))
    setIsPublic(item.public)
    setRemark(item.remark)
    setError("")
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const invalid = firstError(requireTrimmed(name, t("form.required")))
    if (invalid) {
      setError(invalid)
      return
    }
    setPending(true)
    setError("")
    try {
      const payload = {
        name: name.trim(),
        labels: serializeLabels(labels),
        public: isPublic,
        remark: remark.trim(),
      }
      if (editing) {
        await updateAlertPage({ ...payload, uid: editing.uid })
      } else {
        await createAlertPage(payload)
      }
      setOpen(false)
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
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
      await deleteAlertPage(confirm.uid)
      setConfirm(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: AlertPageItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm(item) },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("alertPage.title")} description={t("alertPage.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input
          className="max-w-sm"
          value={draft}
          placeholder={t("alertPage.namePlaceholder")}
          onChange={(event) => setDraft(event.target.value)}
        />
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
            <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
            <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("alertPage.empty")}</p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("alertPage.name")}</TableHead>
                <TableHead>{t("alertPage.labels")}</TableHead>
                <TableHead>{t("alertPage.visibility")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell className="max-w-xs truncate font-mono text-xs">{item.labels}</TableCell>
                  <TableCell>{item.public ? t("alertPage.public") : t("alertPage.private")}</TableCell>
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
        <DialogContent className="max-h-[80vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4 sm:grid-cols-2" onSubmit={onSave}>
            <FormField label={t("alertPage.name")} htmlFor="alert-page-name" className="sm:col-span-full">
              <Input
                id="alert-page-name"
                value={name}
                placeholder={t("alertPage.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
                required
              />
            </FormField>
            <div className="grid gap-2 sm:col-span-full">
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm font-medium">{t("alertPage.labels")}</span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setLabels((current) => [...current, { key: "", value: "" }])}
                >
                  {t("alertPage.addLabel")}
                </Button>
              </div>
              {labels.map((pair, index) => (
                <div key={index} className="grid gap-2 sm:grid-cols-2">
                  <Input
                    value={pair.key}
                    placeholder={t("alertPage.labelKeyPlaceholder")}
                    onChange={(event) => {
                      const next = [...labels]
                      next[index] = { ...pair, key: event.target.value }
                      setLabels(next)
                    }}
                  />
                  <Input
                    value={pair.value}
                    placeholder={t("alertPage.labelValuePlaceholder")}
                    onChange={(event) => {
                      const next = [...labels]
                      next[index] = { ...pair, value: event.target.value }
                      setLabels(next)
                    }}
                  />
                </div>
              ))}
            </div>
            <FormField label={t("alertPage.remark")} htmlFor="alert-page-remark" className="sm:col-span-full">
              <Input
                id="alert-page-remark"
                value={remark}
                placeholder={t("alertPage.remarkPlaceholder")}
                onChange={(event) => setRemark(event.target.value)}
              />
            </FormField>
            <label className="flex items-center gap-2 sm:col-span-full">
              <Checkbox checked={isPublic} onCheckedChange={(value) => setIsPublic(value === true)} />
              <span className="text-sm">{t("alertPage.publicHint")}</span>
            </label>
            {error ? <p className="sm:col-span-full text-sm text-destructive">{error}</p> : null}
            <DialogFooter className="sm:col-span-full">
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
                { label: t("alertPage.labels"), value: detail.labels },
                {
                  label: t("alertPage.visibility"),
                  value: detail.public ? t("alertPage.public") : t("alertPage.private"),
                },
                { label: t("alertPage.remark"), value: detail.remark },
              ]
            : []
        }
      />
      <ConfirmAction
        open={confirm !== null}
        title={t("confirm.deleteTitle")}
        description={t("confirm.deleteDescription")}
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
