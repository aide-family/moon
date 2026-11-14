import { useState } from "react"
import { Search } from "lucide-react"
import { useTranslation } from "react-i18next"
import { formErrorMessage, readFormError } from "@/lib/form-error"

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
import type { FilterChip } from "@/components/filter-chips"
import { FormField } from "@/components/form-field"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { SearchableSelect } from "@/components/searchable-select"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { TableActionsHead } from "@/components/table-heads"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
} from "@/lib/form"
import { parseLabelPairs, serializeLabelPairs, type LabelPair } from "@/lib/labels"
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { useLabelOptions } from "@/hooks/use-label-options"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"

export function AlertPageManagePage() {
  const { t } = useTranslation()
  const session = useSession()
  const labelOptions = useLabelOptions()
  const query = useListQuery()
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty, reset } = useFilterDraft({ name: query.name })
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AlertPageItem | null>(null)
  const [name, setName] = useState("")
  const [labels, setLabels] = useState<LabelPair[]>([{ key: "", value: "" }])
  const [isPublic, setIsPublic] = useState(false)
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AlertPageItem | null>(null)
  const [confirm, setConfirm] = useState<AlertPageItem | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("alert_page:write")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<AlertPageItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listAlertPages({ page, pageSize, name: query.name, scope: "manage" }),
  })

  function openCreate() {
    setEditing(null)
    setName("")
    setLabels([{ key: "", value: "" }])
    setIsPublic(false)
    setRemark("")
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openEdit(item: AlertPageItem) {
    setEditing(item)
    setName(item.name)
    setLabels(parseLabelPairs(item.labels))
    setIsPublic(item.public)
    setRemark(item.remark)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const payload = {
        name: name.trim(),
        labels: serializeLabelPairs(labels),
        public: isPublic,
        remark: remark.trim(),
      }
      if (editing) {
        await updateAlertPage({ ...payload, uid: editing.uid })
      } else {
        await createAlertPage(payload)
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
      await deleteAlertPage(confirm.uid)
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
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

  const chips: FilterChip[] = query.name
    ? [
        {
          key: "name",
          label: `${t("alertPage.name")}: ${query.name}`,
          onRemove: clearFilters,
        },
      ]
    : []

  function applyFilters() {
    query.update({ name: draft.name.trim() })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "" })
    reset()
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("alertPage.title")} description={t("alertPage.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      <ListFilter
        id="alert-page-filters"
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
            htmlFor="alert-page-filter-name"
            className="text-xs text-muted-foreground"
          >
            {t("alertPage.name")}
          </FieldLabel>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="alert-page-filter-name"
              className="pl-9"
              value={draft.name}
              placeholder={t("alertPage.namePlaceholder")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
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
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell className="max-w-xs truncate font-mono text-xs">{item.labels}</TableCell>
                  <TableCell>
                    <StatusBadge
                      value={item.public ? "PUBLIC" : "PRIVATE"}
                      label={item.public ? t("alertPage.public") : t("alertPage.private")}
                    />
                  </TableCell>
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
                label={t("alertPage.name")}
                htmlFor="alert-page-name"
                className="sm:col-span-full"
                required
                error={fieldErrors.name}
              >
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
                <p className="text-xs text-muted-foreground">{t("alertPage.labelsHint")}</p>
                {labels.map((pair, index) => {
                  // 编辑旧数据时，历史键/值可能已不在标签集合里；保留为可选项，保存时不会丢数据。
                  const keyOptions = labelOptions.map(({ key }) => ({
                    value: key,
                    label: key,
                  }))
                  if (
                    pair.key !== "" &&
                    !keyOptions.some((option) => option.value === pair.key)
                  ) {
                    keyOptions.push({ value: pair.key, label: pair.key })
                  }
                  const values =
                    labelOptions.find((option) => option.key === pair.key)
                      ?.values ?? []
                  const valueOptions = values.map((value) => ({
                    value,
                    label: value,
                  }))
                  if (pair.value !== "" && !values.includes(pair.value)) {
                    valueOptions.push({ value: pair.value, label: pair.value })
                  }
                  return (
                    <div key={index} className="grid gap-2 sm:grid-cols-2">
                      <SearchableSelect
                        value={pair.key}
                        options={keyOptions}
                        placeholder={t("alertPage.labelKeyPlaceholder")}
                        searchPlaceholder={t(
                          "alertRule.datasourceLabelKeySearch"
                        )}
                        emptyText={t("alertRule.customKeyHint")}
                        allowCustom
                        customLabel={(value) =>
                          t("alertRule.customKey", { value })
                        }
                        customHint={t("alertRule.customKeyHint")}
                        clearLabel={t("alertRule.datasourceLabelClear")}
                        onChange={(key) => {
                          if (key === pair.key) {
                            return
                          }
                          const nextValues =
                            labelOptions.find((option) => option.key === key)
                              ?.values ?? []
                          const next = [...labels]
                          next[index] = {
                            key,
                            value: nextValues.includes(pair.value)
                              ? pair.value
                              : "",
                          }
                          setLabels(next)
                        }}
                      />
                      <SearchableSelect
                        value={pair.value}
                        options={valueOptions}
                        placeholder={t("alertPage.labelValuePlaceholder")}
                        searchPlaceholder={t(
                          "alertRule.datasourceLabelValueSearch"
                        )}
                        emptyText={t("alertRule.datasourceLabelValueEmpty")}
                        allowCustom
                        customLabel={(value) =>
                          t("alertRule.datasourceLabelCustomValue", { value })
                        }
                        customHint={t("alertRule.datasourceLabelCustomHint")}
                        clearLabel={t("alertRule.datasourceLabelClear")}
                        onChange={(value) => {
                          const next = [...labels]
                          next[index] = { ...pair, value }
                          setLabels(next)
                        }}
                      />
                    </div>
                  )
                })}
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
          }
        }}
        fields={
          detail
            ? [
                { label: t("detail.uid"), value: detail.uid },
                { label: t("alertPage.labels"), value: detail.labels },
                {
                  label: t("alertPage.visibility"),
                  value: (
                    <StatusBadge
                      value={detail.public ? "PUBLIC" : "PRIVATE"}
                      label={detail.public ? t("alertPage.public") : t("alertPage.private")}
                    />
                  ),
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
