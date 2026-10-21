import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage, readFormError } from "@/lib/form-error"

import { listNotificationChannels } from "@/api/notification-channel"
import type { NotificationChannelItem } from "@/api/notification-channel/types"
import {
  createNotificationTemplate,
  deleteNotificationTemplate,
  listNotificationPresets,
  listNotificationTemplates,
  updateNotificationTemplate,
  updateNotificationTemplateStatus,
} from "@/api/notification-template"
import type { NotificationPresetItem, NotificationTemplateItem } from "@/api/notification-template/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { NotificationTestDialog } from "@/components/notification-test-dialog"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
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
  SelectItem as SelectOption,
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
import { Textarea } from "@/components/ui/textarea"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
} from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { run } from "@/lib/action"

const types = ["EMAIL", "WEBHOOK"]

export function NotificationTemplatePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [channels, setChannels] = useState<NotificationChannelItem[]>([])
  const [open, setOpen] = useState(false)
  const [presetOpen, setPresetOpen] = useState(false)
  const [presets, setPresets] = useState<NotificationPresetItem[]>([])
  const [presetError, setPresetError] = useState("")
  const [editing, setEditing] = useState<NotificationTemplateItem | null>(null)
  const [testing, setTesting] = useState<NotificationTemplateItem | null>(null)
  const [name, setName] = useState("")
  const [type, setType] = useState("EMAIL")
  const [subject, setSubject] = useState("")
  const [body, setBody] = useState("")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NotificationTemplateItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: NotificationTemplateItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("notification:write")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<NotificationTemplateItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listNotificationTemplates({ page, pageSize, name: query.name }),
  })

  useEffect(() => {
    let active = true
    void listNotificationChannels({ page: 1, pageSize: 100 })
      .then((result) => {
        if (active) {
          setChannels(result.items ?? [])
        }
      })
      .catch(() => {
        if (active) setChannels([])
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  function openCreate() {
    setEditing(null)
    setName("")
    setType("EMAIL")
    setSubject("")
    setBody("")
    setRemark("")
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openEdit(item: NotificationTemplateItem) {
    setEditing(item)
    setName(item.name)
    setType(item.type)
    setSubject(item.subject)
    setBody(item.body)
    setRemark(item.remark)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      ["type", requireTrimmed(type, t("form.required"))],
      [
        "subject",
        type === "EMAIL" ? requireTrimmed(subject, t("form.required")) : null,
      ],
      ["body", requireTrimmed(body, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const payload = { name: name.trim(), type, subject, body, remark }
      if (editing) {
        await updateNotificationTemplate({ ...payload, uid: editing.uid })
      } else {
        await createNotificationTemplate(payload)
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

  async function openPresets() {
    setPresetError("")
    setPresetOpen(true)
    try {
      const result = await listNotificationPresets()
      setPresets(result.items ?? [])
    } catch (err) {
      setPresetError(err instanceof Error ? err.message : t("state.failed"))
      setPresets([])
    }
  }

  function applyPreset(preset: NotificationPresetItem) {
    setEditing(null)
    setName(preset.name)
    setType(preset.type)
    setSubject(preset.subject)
    setBody(preset.body)
    setRemark(preset.description)
    setError("")
    setPresetOpen(false)
    setOpen(true)
  }

  async function createFromPreset(preset: NotificationPresetItem) {
    setPending(true)
    setError("")
    try {
      await createNotificationTemplate({
        name: `${preset.name} ${t("template.copySuffix")}`,
        type: preset.type,
        subject: preset.subject,
        body: preset.body,
        remark: preset.description,
      })
      setOpen(false)
      setPresetOpen(false)
      reload()
    } catch (err) {
      setPresetError(err instanceof Error ? err.message : t("state.failed"))
      setPresetOpen(true)
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
        await deleteNotificationTemplate(confirm.item.uid)
      } else {
        await updateNotificationTemplateStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: NotificationTemplateItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      {
        key: "test",
        label: t("action.test"),
        onSelect: () => {
          setTesting(item)
          setError("")
        },
      },
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => run(updateNotificationTemplateStatus(item.uid, "ACTIVE").then(() => reload()), setError) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("template.title")} description={t("template.description")}>
        {canWrite ? (
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => void openPresets()}>
              {t("template.presets")}
            </Button>
            <Button onClick={openCreate}>{t("action.create")}</Button>
          </div>
        ) : null}
      </PageHeader>
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input className="max-w-sm" value={draft} placeholder={t("template.namePlaceholder")} onChange={(event) => setDraft(event.target.value)} />
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
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("template.empty")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("template.name")}</TableHead>
                <TableHead>{t("template.type")}</TableHead>
                <TableHead>{t("template.subject")}</TableHead>
                <TableHead>{t("namespace.status")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell>{enumLabel(t, item.type)}</TableCell>
                  <TableCell>{item.subject}</TableCell>
                  <TableCell>{enumLabel(t, item.status)}</TableCell>
                  <TableCell>
                    <RowActions detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => setDetail(item)} actions={actions(item)} />
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
          <form noValidate className="grid gap-4 sm:grid-cols-2" onSubmit={onSave}>
            <FormField
              label={t("template.name")}
              htmlFor="tpl-name"
              required
              error={fieldErrors.name}
            >
              <Input id="tpl-name" value={name} placeholder={t("template.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <FormField
              label={t("template.type")}
              required
              error={fieldErrors.type}
            >
              <Select
                value={type}
                onValueChange={(value) => {
                  setType(value)
                  if (value !== "EMAIL") {
                    setSubject("")
                  }
                  setFieldErrors({})
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t("template.typePlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  {types.map((item) => (
                    <SelectOption key={item} value={item}>
                      {enumLabel(t, item)}
                    </SelectOption>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            {type === "EMAIL" ? (
              <FormField
                label={t("template.subject")}
                htmlFor="tpl-subject"
                hint={t("template.subjectEmailOnly")}
                className="sm:col-span-full"
                required
                error={fieldErrors.subject}
              >
                <Input id="tpl-subject" value={subject} placeholder={t("template.subjectPlaceholder")} onChange={(event) => setSubject(event.target.value)} required />
              </FormField>
            ) : null}
            <FormField
              label={t("template.body")}
              htmlFor="tpl-body"
              hint={t("template.bodyHint")}
              className="sm:col-span-full"
              required
              error={fieldErrors.body}
            >
              <Textarea id="tpl-body" value={body} placeholder={t("template.bodyPlaceholder")} onChange={(event) => setBody(event.target.value)} />
            </FormField>
            <FormField label={t("template.remark")} htmlFor="tpl-remark" className="sm:col-span-full">
              <Input id="tpl-remark" value={remark} placeholder={t("template.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
            </FormField>
            {error ? <p className="sm:col-span-full text-sm text-destructive">{error}</p> : null}
            <DialogFooter className="sm:col-span-full">
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      {testing ? (
        <NotificationTestDialog target={{ kind: "template", item: testing }} channels={channels} onClose={() => setTesting(null)} />
      ) : null}
      <Dialog open={presetOpen} onOpenChange={setPresetOpen}>
        <DialogContent className="max-h-[80vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t("template.presetTitle")}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">{t("template.presetHint")}</p>
          {presetError ? <p className="text-sm text-destructive">{presetError}</p> : null}
          {presets.length === 0 && !presetError ? (
            <p className="text-sm text-muted-foreground">{t("template.presetEmpty")}</p>
          ) : (
            <div className="grid gap-3">
              {presets.map((preset) => (
                <div key={preset.uid} className="grid gap-2 rounded-lg border p-3">
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div>
                      <p className="font-medium">{preset.name}</p>
                      <p className="text-sm text-muted-foreground">
                        {preset.type} · {preset.description}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <Button variant="outline" size="sm" disabled={pending} onClick={() => applyPreset(preset)}>
                        {t("template.usePresetAndEdit")}
                      </Button>
                      <Button size="sm" disabled={pending} onClick={() => void createFromPreset(preset)}>
                        {t("template.usePreset")}
                      </Button>
                    </div>
                  </div>
                  <pre className="max-h-24 overflow-auto whitespace-pre-wrap text-xs text-muted-foreground">{preset.body}</pre>
                </div>
              ))}
            </div>
          )}
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
                { label: t("template.type"), value: enumLabel(t, detail.type) },
                { label: t("template.subject"), value: detail.subject },
                { label: t("template.body"), value: detail.body },
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
