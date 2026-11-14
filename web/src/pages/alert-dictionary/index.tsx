import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Search } from "lucide-react"
import { formErrorMessage, readFormError } from "@/lib/form-error"
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
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { FormField } from "@/components/form-field"
import { ListFilter } from "@/components/list-filter"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { ListFrame } from "@/components/list-frame"
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
import { Field, FieldLabel } from "@/components/ui/field"
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
import { TableActionsHead, TableStatusHead, TableTypeHead } from "@/components/table-heads"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
} from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
import { runAction } from "@/lib/action"

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
  const detailUid = params.get("uid") || ""
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty, reset } = useFilterDraft({ name: query.name, type })
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AlertDictionaryItem | null>(null)
  const [name, setName] = useState("")
  const [color, setColor] = useState("")
  const [sort, setSort] = useState("0")
  const [levelType, setLevelType] = useState("ALERT")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AlertDictionaryItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: AlertDictionaryItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("alert_dictionary:write")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<AlertDictionaryItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${type}`,
    fetch: (page, pageSize) => listAlertDictionaries({ page, pageSize, name: query.name, type }),
  })

  // 实时告警详情的级别名称链接带上 uid，进入本页后直接打开字典项详情。
  useEffect(() => {
    if (!detailUid) {
      return
    }
    let active = true
    async function load() {
      try {
        const result = await listAlertDictionaries({ page: 1, pageSize: 100, type: "ALERT" })
        if (!active) {
          return
        }
        const found = (result.items ?? []).find((item) => item.uid === detailUid)
        if (found) {
          setDetail(found)
        }
      } catch {
        // 字典项已删除或无权访问时保持列表，不打开空详情。
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [detailUid, session.namespaceUid])

  function openCreate() {
    setEditing(null)
    setName("")
    setColor("#d92d20")
    setSort("0")
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
    setSort(String(item.sort ?? 0))
    setLevelType(item.type)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const sortValue = Number(sort)
    const sortInvalid =
      sort.trim() === "" ||
      !Number.isInteger(sortValue) ||
      sortValue < 0 ||
      sortValue > 10000
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      ["sort", sortInvalid ? t("alertDictionary.sortInvalid") : null],
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
        await updateAlertDictionary({ uid: editing.uid, name: name.trim(), color: nextColor, sort: sortValue })
      } else {
        await createAlertDictionary({ name: name.trim(), color: nextColor, type: levelType as AlertDictionaryType, sort: sortValue })
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
      setConfirmError(formErrorMessage(err, t("state.failed")))
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
        : { key: "enable", label: t("action.enable"), onSelect: () => runAction(updateAlertDictionaryStatus(item.uid, "ACTIVE").then(() => reload()), setError) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  const chips: FilterChip[] = []
  if (query.name) {
    chips.push({
      key: "name",
      label: `${t("alertDictionary.name")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (type) {
    chips.push({
      key: "type",
      label: `${t("datasource.type")}: ${enumLabel(t, type)}`,
      onRemove: () => query.update({ type: "" }),
    })
  }

  function applyFilters() {
    query.update({ name: draft.name.trim(), type: draft.type })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "", type: "" })
    reset()
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("alertDictionary.title")} description={t("alertDictionary.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      <ListFilter
        id="alert-dictionary-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={chips}
        onClearAll={clearFilters}
        onClear={clearFilters}
        appliedCount={chips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="min-w-0 w-full max-w-[500px] flex-1 basis-56">
          <FieldLabel
            htmlFor="alert-dictionary-filter-name"
            className="text-xs text-muted-foreground"
          >
            {t("alertDictionary.name")}
          </FieldLabel>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="alert-dictionary-filter-name"
              className="pl-9"
              value={draft.name}
              placeholder={t("alertDictionary.namePlaceholder")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
        </Field>
        <Field className="w-40 shrink-0">
          <FieldLabel
            htmlFor="alert-dictionary-filter-type"
            className="text-xs text-muted-foreground"
          >
            {t("datasource.type")}
          </FieldLabel>
          <Select
            value={draft.type || "ALL"}
            onValueChange={(value) =>
              update("type", value === "ALL" ? "" : value)
            }
          >
            <SelectTrigger id="alert-dictionary-filter-type" className="w-full">
              <SelectValue placeholder={t("datasource.typePlaceholder")}>
                <span className="flex items-center gap-2">
                  <EnumIcon value={draft.type} />
                  <span>
                    {draft.type ? enumLabel(t, draft.type) : t("datasource.typeAll")}
                  </span>
                </span>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">{t("datasource.typeAll")}</SelectItem>
              {types.map((item) => (
                <SelectItem key={item} value={item}>
                  <span className="flex items-center gap-2">
                    <EnumIcon value={item} />
                    <span>{enumLabel(t, item)}</span>
                  </span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Field>
      </ListFilter>
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
                  <TableHead className="whitespace-nowrap">{t("alertDictionary.sort")}</TableHead>
                  <TableHead>{t("alertDictionary.color")}</TableHead>
                  <TableTypeHead>{t("datasource.type")}</TableTypeHead>
                  <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                  <TableActionsHead>{t("action.actions")}</TableActionsHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.uid}>
                    <TableCell>{item.name}</TableCell>
                    <TableCell className="tabular-nums">{item.sort ?? 0}</TableCell>
                    <TableCell>
                      <ColorChip value={item.color} />
                    </TableCell>
                    <TableCell><StatusBadge value={item.type} /></TableCell>
                    <TableCell><StatusBadge value={item.status} /></TableCell>
                    <TableCell>
                      <RowActions nowrap
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
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onSave}>
            <DialogBody className="grid gap-4 sm:grid-cols-2">
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
            <FormField
              label={t("alertDictionary.sort")}
              htmlFor="level-sort"
              hint={t("alertDictionary.sortHint")}
              required
              error={fieldErrors.sort}
            >
              <Input
                id="level-sort"
                type="number"
                inputMode="numeric"
                min={0}
                max={10000}
                step={1}
                value={sort}
                placeholder={t("alertDictionary.sortPlaceholder")}
                onChange={(event) => setSort(event.target.value)}
                required
              />
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
            </DialogBody>
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
            if (detailUid) {
              const next = new URLSearchParams(params)
              next.delete("uid")
              setParams(next)
            }
          }
        }}
        fields={
          detail
            ? [
                { label: t("detail.uid"), value: detail.uid },
                { label: t("alertDictionary.sort"), value: String(detail.sort ?? 0) },
                { label: t("alertDictionary.color"), value: <ColorChip value={detail.color} /> },
                { label: t("datasource.type"), value: <StatusBadge value={detail.type} /> },
                { label: t("namespace.status"), value: <StatusBadge value={detail.status} /> },
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
