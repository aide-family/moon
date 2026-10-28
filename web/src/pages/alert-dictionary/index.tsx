import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Check } from "lucide-react"
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
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
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
import { TableActionsHead, TableStatusHead, TableTypeHead } from "@/components/table-heads"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
} from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
import { runAction } from "@/lib/action"

const types = ["ALERT", "DATASOURCE"]

// 新建字典项时可一键带入的预设。颜色是字典项的数据值（会随字典保存并展示），
// 因此这里允许写 hex，而不是主题 token；不要把这些颜色换成 Tailwind 调色板类。
const presetGroups: Array<{
  type: AlertDictionaryType
  labelKey: string
  items: Array<{ key: string; nameKey: string; color: string }>
}> = [
  {
    type: "ALERT",
    labelKey: "alertDictionary.typeAlert",
    items: [
      { key: "alert-p0", nameKey: "alertDictionary.presetAlert.p0", color: "#e41a0c" },
      { key: "alert-p1", nameKey: "alertDictionary.presetAlert.p1", color: "#f06f05" },
      { key: "alert-p2", nameKey: "alertDictionary.presetAlert.p2", color: "#dfbb07" },
      { key: "alert-p3", nameKey: "alertDictionary.presetAlert.p3", color: "#2ac02d" },
    ],
  },
  {
    type: "DATASOURCE",
    labelKey: "alertDictionary.typeDatasource",
    items: [
      { key: "datasource-high", nameKey: "alertDictionary.presetDatasource.high", color: "#d92d20" },
      { key: "datasource-medium", nameKey: "alertDictionary.presetDatasource.medium", color: "#dfbb07" },
      { key: "datasource-low", nameKey: "alertDictionary.presetDatasource.low", color: "#2ac02d" },
    ],
  },
]

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
    setLevelType("ALERT")
    setError("")
    setFieldErrors({})
    setOpen(true)
  }


  const colorSwatch = /^#[0-9a-fA-F]{6}$/.test(color) ? color : "#1677ff"

  // 选中的预设由当前表单值反推：预设填入后若再手改任一字段，选中态自然消失，
  // 不会出现“标着 P0 实际是别的值”的误导。
  const selectedPresetKey = (() => {
    for (const group of presetGroups) {
      if (group.type !== levelType) {
        continue
      }
      for (const item of group.items) {
        if (name.trim() === t(item.nameKey) && normalizeColor(color) === item.color) {
          return item.key
        }
      }
    }
    return null
  })()

  function applyPreset(preset: { type: AlertDictionaryType; name: string; color: string }) {
    setLevelType(preset.type)
    setName(preset.name)
    setColor(preset.color)
    setError("")
    setFieldErrors({})
  }

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
                  <TableTypeHead>{t("datasource.type")}</TableTypeHead>
                  <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                  <TableActionsHead>{t("action.actions")}</TableActionsHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.uid}>
                    <TableCell>{item.name}</TableCell>
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
        <DialogContent
          className="sm:max-w-2xl"
          onOpenAutoFocus={(event) => {
            // 预设行（含提示按钮）排在名称输入之前，默认焦点会落到预设上；
            // 固定聚焦名称，保持打开对话框后可直接输入，避免提示气泡自动弹出。
            event.preventDefault()
            const content = event.currentTarget as HTMLElement | null
            content?.querySelector<HTMLInputElement>("#level-name")?.focus()
          }}
        >
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid min-h-0 gap-4 overflow-y-auto sm:grid-cols-2" onSubmit={onSave}>
            {!editing ? (
              <FormField
                label={t("alertDictionary.presetLabel")}
                hint={t("alertDictionary.presetHint")}
                className="sm:col-span-full"
              >
                <div className="grid gap-2" role="group" aria-label={t("alertDictionary.presetLabel")}>
                  {presetGroups.map((group) => (
                    <div key={group.type} className="flex flex-wrap items-center gap-2">
                      <span className="min-w-16 text-xs text-muted-foreground">{t(group.labelKey)}</span>
                      {group.items.map((item) => {
                        const presetName = t(item.nameKey)
                        const selected = selectedPresetKey === item.key
                        return (
                          <Button
                            key={item.key}
                            type="button"
                            size="sm"
                            variant={selected ? "default" : "outline"}
                            aria-pressed={selected}
                            onClick={() => applyPreset({ type: group.type, name: presetName, color: item.color })}
                          >
                            <span
                              aria-hidden
                              className="size-3 shrink-0 rounded-sm border border-border"
                              style={{ backgroundColor: item.color }}
                            />
                            <span>{presetName}</span>
                            {selected ? (
                              <>
                                <Check aria-hidden className="size-3.5" />
                                <span className="text-xs">{t("alertDictionary.presetApplied")}</span>
                              </>
                            ) : null}
                          </Button>
                        )
                      })}
                    </div>
                  ))}
                </div>
              </FormField>
            ) : null}
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
