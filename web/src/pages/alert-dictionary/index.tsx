import { useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage, readFormError } from "@/lib/form-error"
import { useSearchParams } from "react-router-dom"

import {
  createAlertDictionary,
  deleteAlertDictionary,
  listAlertDictionaries,
  updateAlertDictionary,
  updateAlertDictionaryStatus,
} from "@/api/alert-dictionary"
import type { AlertDictionaryItem, AlertDictionaryType } from "@/api/alert-dictionary/types"
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
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
} from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { run } from "@/lib/action"

const types = ["ALERT", "DATASOURCE"]

function normalizeColor(value: string): string {
  const trimmed = value.trim()
  if (/^#[0-9a-fA-F]{6}$/.test(trimmed)) {
    return trimmed.toLowerCase()
  }
  if (/^[0-9a-fA-F]{6}$/.test(trimmed)) {
    return `#${trimmed.toLowerCase()}`
  }
  return trimmed
}

function ColorChip({ value }: { value: string }) {
  const hex = normalizeColor(value)
  const swatch = /^#[0-9a-fA-F]{6}$/.test(hex) ? hex : "#1677ff"
  return (
    <span className="inline-flex items-center gap-2">
      <span
        aria-hidden
        className="size-3 shrink-0 rounded-sm border border-border"
        style={{ backgroundColor: swatch }}
      />
      <span>{value || "—"}</span>
    </span>
  )
}

export function AlertDictionaryPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [params, setParams] = useSearchParams()
  const type = params.get("type") || ""
  const [draft, setDraft] = useState(query.name)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AlertDictionaryItem | null>(null)
  const [name, setName] = useState("")
  const [color, setColor] = useState("")
  const [levelType, setLevelType] = useState("ALERT")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AlertDictionaryItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: AlertDictionaryItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("alert_rule:write")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<AlertDictionaryItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${type}`,
    fetch: (page, pageSize) => listAlertDictionaries({ page, pageSize, name: query.name, type }),
  })

  function openCreate() {
    setEditing(null)
    setName("")
    setColor("#d92d20")
    setLevelType("ALERT")
    setError("")
    setFieldErrors({})
    setOpen(true)
  }


  const colorSwatch = /^#[0-9a-fA-F]{6}$/.test(color) ? color : "#1677ff"

  function openEdit(item: AlertDictionaryItem) {
    setEditing(item)
    setName(item.name)
    setColor(item.color)
    setLevelType(item.type)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      [
        "type",
        editing ? null : requireTrimmed(levelType, t("form.required")),
      ],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const nextColor = normalizeColor(color)
      if (editing) {
        await updateAlertDictionary({ uid: editing.uid, name: name.trim(), color: nextColor })
      } else {
        await createAlertDictionary({ name: name.trim(), color: nextColor, type: levelType as AlertDictionaryType })
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
        await deleteAlertDictionary(confirm.item.uid)
      } else {
        await updateAlertDictionaryStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: AlertDictionaryItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => run(updateAlertDictionaryStatus(item.uid, "ACTIVE").then(() => reload()), setError) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("alertDictionary.title")} description={t("alertDictionary.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input className="max-w-sm" value={draft} placeholder={t("alertDictionary.namePlaceholder")} onChange={(event) => setDraft(event.target.value)} />
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
                {enumLabel(t, item)}
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
              <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
              <Button variant="outline" onClick={() => reload()}>
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("alertDictionary.empty")}</p> : null}
          {!loading && !failed && items.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("alertDictionary.name")}</TableHead>
                  <TableHead>{t("alertDictionary.color")}</TableHead>
                  <TableHead>{t("datasource.type")}</TableHead>
                  <TableHead>{t("namespace.status")}</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.uid}>
                    <TableCell>{item.name}</TableCell>
                    <TableCell>
                      <ColorChip value={item.color} />
                    </TableCell>
                    <TableCell>{enumLabel(t, item.type)}</TableCell>
                    <TableCell>{enumLabel(t, item.status)}</TableCell>
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
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4 sm:grid-cols-2" onSubmit={onSave}>
            <FormField
              label={t("alertDictionary.name")}
              htmlFor="level-name"
              required
              error={fieldErrors.name}
            >
              <Input id="level-name" value={name} placeholder={t("alertDictionary.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <FormField
              label={t("alertDictionary.color")}
              htmlFor="level-color"
              hint={t("alertDictionary.colorHint")}
              error={fieldErrors.color}
            >
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  aria-label={t("alertDictionary.colorPicker")}
                  className="size-9 cursor-pointer rounded-md border border-input bg-transparent p-1"
                  value={colorSwatch}
                  onChange={(event) => setColor(event.target.value)}
                />
                <Input
                  id="level-color"
                  value={color}
                  placeholder={t("alertDictionary.colorPlaceholder")}
                  onChange={(event) => setColor(event.target.value)}
                />
              </div>
            </FormField>
            {!editing ? (
              <FormField
                label={t("datasource.type")}
                className="sm:col-span-full"
                required
                error={fieldErrors.type}
              >
                <div className="flex flex-wrap gap-4" role="radiogroup" aria-label={t("datasource.type")}>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="radio"
                      name="alert-dictionary-type"
                      value="ALERT"
                      checked={levelType === "ALERT"}
                      onChange={() => setLevelType("ALERT")}
                    />
                    {t("alertDictionary.typeAlert")}
                  </label>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="radio"
                      name="alert-dictionary-type"
                      value="DATASOURCE"
                      checked={levelType === "DATASOURCE"}
                      onChange={() => setLevelType("DATASOURCE")}
                    />
                    {t("alertDictionary.typeDatasource")}
                  </label>
                </div>
              </FormField>
            ) : null}
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
                { label: t("alertDictionary.color"), value: <ColorChip value={detail.color} /> },
                { label: t("datasource.type"), value: enumLabel(t, detail.type) },
                { label: t("namespace.status"), value: enumLabel(t, detail.status) },
              ]
            : []
        }
      />
      <ConfirmAction
        open={confirm !== null}
        title={confirm?.action === "delete" ? t("confirm.deleteTitle") : t("confirm.disableTitle")}
        description={confirm?.action === "delete" ? t("confirm.deleteDescription") : t("confirm.disableDescription")}
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
