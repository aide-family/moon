import { useEffect, useMemo, useState } from "react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router-dom"
import { formErrorMessage, readFormError } from "@/lib/form-error"

import { listNotificationChannels } from "@/api/notification-channel"
import type { NotificationChannelItem } from "@/api/notification-channel/types"
import {
  createNotificationTemplate,
  deleteNotificationTemplate,
  listNotificationTemplates,
  updateNotificationTemplate,
  updateNotificationTemplateStatus,
} from "@/api/notification-template"
import type { NotificationTemplateItem } from "@/api/notification-template/types"
import { useSession } from "@/auth/session"
import { authorizedTemplateTypes } from "@/lib/feature-catalog"
import { ChannelLogo } from "@/components/channel-logo"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { NotificationTestDialog } from "@/components/notification-test-dialog"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TemplateEditor } from "@/components/template-editor"
import { TemplatePreviewDialog } from "@/components/template-preview-dialog"
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

const webhookApps = ["OTHER", "DINGTALK", "WECHAT", "FEISHU"]

function templateAppLabel(t: TFunction, type: string, app: string): string {
  if (type !== "WEBHOOK") {
    return enumLabel(t, "OTHER")
  }
  return enumLabel(t, app || "OTHER")
}

export function NotificationTemplatePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const navigate = useNavigate()
  const [draft, setDraft] = useState(query.name)
  const [channels, setChannels] = useState<NotificationChannelItem[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationTemplateItem | null>(null)
  const [testing, setTesting] = useState<NotificationTemplateItem | null>(null)
  const [previewing, setPreviewing] = useState<NotificationTemplateItem | null>(null)
  const [formPreview, setFormPreview] = useState(false)
  const [name, setName] = useState("")
  const [type, setType] = useState("EMAIL")
  const [app, setApp] = useState("OTHER")
  const [subject, setSubject] = useState("")
  const [body, setBody] = useState("")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NotificationTemplateItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: NotificationTemplateItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("notification_template:write")
  // 模板类型跟着渠道类型走：未授权渠道对应的模板不出现在新建和筛选里。
  const allowedTypes = useMemo(
    () => authorizedTemplateTypes(session.features),
    [session.features]
  )
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<NotificationTemplateItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${query.type}|${query.app}`,
    fetch: (page, pageSize) => listNotificationTemplates({ page, pageSize, name: query.name, type: query.type, app: query.app }),
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
    setType(allowedTypes[0] ?? "EMAIL")
    setApp("OTHER")
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
    setApp(item.app || "OTHER")
    setSubject(item.subject)
    setBody(item.body)
    setRemark(item.remark)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function selectType(value: string) {
    setType(value)
    if (value !== "WEBHOOK") {
      setApp("OTHER")
    }
    if (value !== "EMAIL") {
      setSubject("")
    }
    setFieldErrors({})
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
      const payload = {
        name: name.trim(),
        type,
        app: type === "WEBHOOK" ? app : "OTHER",
        subject: type === "EMAIL" ? subject : "",
        body,
        remark,
      }
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
      setConfirmError(formErrorMessage(err, t("state.failed")))
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
      { key: "preview", label: t("template.preview"), onSelect: () => setPreviewing(item) },
      ...(item.type === "SMS_ALICLOUD"
        ? []
        : [{ key: "test", label: t("action.test"), onSelect: () => {
            setTesting(item)
            setError("")
          } }]),
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => runAction(updateNotificationTemplateStatus(item.uid, "ACTIVE").then(() => reload()), setError) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  const formPreviewDialog = useMemo(
    () => (
      <TemplatePreviewDialog
        open={formPreview}
        onClose={() => setFormPreview(false)}
        type={type}
        app={type === "WEBHOOK" ? app : "OTHER"}
        subject={subject}
        body={body}
      />
    ),
    [formPreview, type, app, subject, body],
  )

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("template.title")} description={t("template.description")}>
        {canWrite ? (
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => navigate("/notification-templates/presets")}>
              {t("template.presets")}
            </Button>
            {allowedTypes.length > 0 ? (
              <Button onClick={openCreate}>{t("action.create")}</Button>
            ) : null}
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
        <Select value={query.type || "ALL"} onValueChange={(value) => query.update({ type: value === "ALL" ? "" : value })}>
          <SelectTrigger className="w-44">
            <SelectValue placeholder={t("template.typeFilter")}>
              <span className="flex items-center gap-2">
                <ChannelLogo type={query.type || "EMAIL"} className="size-4" />
                <span>{query.type ? enumLabel(t, query.type) : t("template.allTypes")}</span>
              </span>
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectOption value="ALL">{t("template.allTypes")}</SelectOption>
            {allowedTypes.map((item) => (
              <SelectOption key={item} value={item}>
                <span className="flex items-center gap-2">
                  <ChannelLogo type={item} className="size-4" />
                  <span>{enumLabel(t, item)}</span>
                </span>
              </SelectOption>
            ))}
          </SelectContent>
        </Select>
        <Select value={query.app || "ALL"} onValueChange={(value) => query.update({ app: value === "ALL" ? "" : value })}>
          <SelectTrigger className="w-44">
            <SelectValue placeholder={t("template.appFilter")}>
              <span className="flex items-center gap-2">
                <ChannelLogo type="WEBHOOK" app={query.app || "OTHER"} className="size-4" />
                <span>{query.app ? enumLabel(t, query.app) : t("template.allApps")}</span>
              </span>
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectOption value="ALL">{t("template.allApps")}</SelectOption>
            {webhookApps.map((item) => (
              <SelectOption key={item} value={item}>
                <span className="flex items-center gap-2">
                  <ChannelLogo type="WEBHOOK" app={item} className="size-4" />
                  <span>{enumLabel(t, item)}</span>
                </span>
              </SelectOption>
            ))}
          </SelectContent>
        </Select>
        <Button type="submit" variant="outline" className="shrink-0">
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
                <TableTypeHead>{t("template.type")}</TableTypeHead>
                <TableHead>{t("template.app")}</TableHead>
                <TableHead>{t("template.subject")}</TableHead>
                <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell>
                    <span className="flex min-w-0 items-center gap-2">
                      <ChannelLogo type={item.type} app={item.app} className="size-4" />
                      <StatusBadge value={item.type} />
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="flex min-w-0 items-center gap-2">
                      <ChannelLogo type={item.type} app={item.app} className="size-4" />
                      <StatusBadge value={item.app} label={templateAppLabel(t, item.type, item.app)} />
                    </span>
                  </TableCell>
                  <TableCell className="max-w-xs truncate">{item.subject}</TableCell>
                  <TableCell><StatusBadge value={item.status} /></TableCell>
                  <TableCell>
                    <RowActions nowrap detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => setDetail(item)} actions={actions(item)} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onSave}>
            <DialogBody className="grid gap-4 sm:grid-cols-2">
              <FormField label={t("template.name")} htmlFor="tpl-name" required error={fieldErrors.name}>
                <Input id="tpl-name" value={name} placeholder={t("template.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
              </FormField>
              <FormField label={t("template.type")} required error={fieldErrors.type}>
                <Select value={type} onValueChange={selectType}>
                  <SelectTrigger>
                    <SelectValue placeholder={t("template.typePlaceholder")}>
                      <span className="flex items-center gap-2">
                        <ChannelLogo type={type} app={app} className="size-4" />
                        <span>{enumLabel(t, type)}</span>
                      </span>
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    {allowedTypes.map((item) => (
                      <SelectOption key={item} value={item}>
                        <span className="flex items-center gap-2">
                          <ChannelLogo type={item} className="size-4" />
                          <span>{enumLabel(t, item)}</span>
                        </span>
                      </SelectOption>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
              {type === "WEBHOOK" ? (
                <FormField label={t("template.app")} required error={fieldErrors.app}>
                  <Select
                    value={app}
                    onValueChange={(value) => {
                      setApp(value)
                      setFieldErrors({})
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t("template.appPlaceholder")}>
                        <span className="flex items-center gap-2">
                          <ChannelLogo type="WEBHOOK" app={app} className="size-4" />
                          <span>{enumLabel(t, app)}</span>
                        </span>
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {webhookApps.map((item) => (
                        <SelectOption key={item} value={item}>
                          <span className="flex items-center gap-2">
                            <ChannelLogo type="WEBHOOK" app={item} className="size-4" />
                            <span>{enumLabel(t, item)}</span>
                          </span>
                        </SelectOption>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
              ) : null}
              {type === "EMAIL" ? (
                <FormField
                  label={t("template.subject")}
                  htmlFor="tpl-subject"
                  hint={t("template.subjectHint")}
                  className="sm:col-span-full"
                  required
                  error={fieldErrors.subject}
                >
                  <TemplateEditor
                    value={subject}
                    onChange={setSubject}
                    singleLine
                    placeholder={t("template.subjectPlaceholder")}
                    ariaLabel={t("template.subject")}
                  />
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
                <TemplateEditor
                  value={body}
                  onChange={setBody}
                  height={260}
                  placeholder={t("template.bodyPlaceholder")}
                  ariaLabel={t("template.body")}
                />
              </FormField>
              <FormField label={t("template.remark")} htmlFor="tpl-remark" className="sm:col-span-full">
                <Input id="tpl-remark" value={remark} placeholder={t("template.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
              </FormField>
            </DialogBody>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setFormPreview(true)}>
                {t("template.preview")}
              </Button>
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      {formPreviewDialog}
      {previewing ? (
        <TemplatePreviewDialog
          open
          onClose={() => setPreviewing(null)}
          type={previewing.type}
          app={previewing.app}
          subject={previewing.subject}
          body={previewing.body}
        />
      ) : null}
      {testing ? (
        <NotificationTestDialog target={{ kind: "template", item: testing }} channels={channels} onClose={() => setTesting(null)} />
      ) : null}
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
                {
                  label: t("template.type"),
                  value: (
                    <span className="flex items-center gap-2">
                      <ChannelLogo type={detail.type} app={detail.app} className="size-4" />
                      <StatusBadge value={detail.type} />
                    </span>
                  ),
                },
                { label: t("template.app"), value: templateAppLabel(t, detail.type, detail.app) },
                { label: t("template.subject"), value: detail.subject },
                { label: t("template.body"), value: <span className="whitespace-pre-wrap break-words">{detail.body}</span> },
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
