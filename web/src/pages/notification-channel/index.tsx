import { useState } from "react"
import { useTranslation } from "react-i18next"
import { formErrorMessage, readFormError } from "@/lib/form-error"

import {
  createNotificationChannel,
  deleteNotificationChannel,
  listNotificationChannels,
  updateNotificationChannel,
  updateNotificationChannelStatus,
} from "@/api/notification-channel"
import type { NotificationChannelItem } from "@/api/notification-channel/types"
import { useSession } from "@/auth/session"
import { ChannelLogo } from "@/components/channel-logo"
import { ConfirmAction } from "@/components/confirm-action"
import { NotificationTestDialog } from "@/components/notification-test-dialog"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
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
import { isEmail } from "@/lib/email"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
import { runAction } from "@/lib/action"

const types = ["SMTP", "SMS", "WEBHOOK"]
const webhookApps = ["OTHER", "DINGTALK", "WECHAT", "FEISHU"]

function appFor(type: string, app: string): string {
  if (type === "SMS") {
    return "SMS_ALICLOUD"
  }
  if (type === "SMTP") {
    return "OTHER"
  }
  return app
}

export function NotificationChannelPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationChannelItem | null>(null)
  const [testing, setTesting] = useState<NotificationChannelItem | null>(null)
  const [name, setName] = useState("")
  const [type, setType] = useState("SMTP")
  const [app, setApp] = useState("OTHER")
  const [endpoint, setEndpoint] = useState("")
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [secret, setSecret] = useState("")
  const [fromAddress, setFromAddress] = useState("")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NotificationChannelItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: NotificationChannelItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("notification:write")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<NotificationChannelItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${query.type}|${query.app}`,
    fetch: (page, pageSize) => listNotificationChannels({ page, pageSize, name: query.name, type: query.type, app: query.app }),
  })

  function openCreate() {
    setEditing(null)
    setName("")
    setType("SMTP")
    setApp("OTHER")
    setEndpoint("")
    setUsername("")
    setPassword("")
    setSecret("")
    setFromAddress("")
    setRemark("")
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openEdit(item: NotificationChannelItem) {
    setEditing(item)
    setName(item.name)
    setType(item.type)
    setApp(item.type === "SMTP" ? "OTHER" : item.app)
    setEndpoint(item.endpoint)
    setUsername(item.username)
    setPassword("")
    setSecret("")
    setFromAddress(item.fromAddress)
    setRemark(item.remark)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openTest(item: NotificationChannelItem) {
    setTesting(item)
    setError("")
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      ["type", requireTrimmed(type, t("form.required"))],
      ["endpoint", requireTrimmed(endpoint, t("form.required"))],
    ])
    const target = endpoint.trim()
    if (target && type === "SMTP") {
      if (target.includes("://") || /\s/.test(target)) {
        errors.endpoint = t("form.invalidEndpoint")
      }
      if (!fromAddress.trim() && !username.trim()) {
        errors.fromAddress = t("channel.senderRequired")
      }
      if (fromAddress.trim() && !isEmail(fromAddress)) {
        errors.fromAddress = t("form.invalidEmail")
      }
    } else if (target && !/^https?:\/\//.test(target)) {
      errors.endpoint = t("form.invalidUrl")
    }
    if (target && type === "SMS") {
      if (!username.trim()) {
        errors.username = t("form.required")
      }
      if (!editing && !password.trim()) {
        errors.password = t("form.required")
      }
      if (!fromAddress.trim()) {
        errors.fromAddress = t("form.required")
      }
    }
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const body = {
        name: name.trim(),
        type,
        app: appFor(type, app),
        endpoint: target,
        username,
        password,
        secret,
        fromAddress,
        remark,
      }
      if (editing) {
        await updateNotificationChannel({ ...body, uid: editing.uid })
      } else {
        await createNotificationChannel(body)
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
        await deleteNotificationChannel(confirm.item.uid)
      } else {
        await updateNotificationChannelStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: NotificationChannelItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      ...(item.status === "ACTIVE" && item.type !== "SMS" ? [{ key: "test", label: t("action.test"), onSelect: () => openTest(item) }] : []),
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => runAction(updateNotificationChannelStatus(item.uid, "ACTIVE").then(() => reload()), setError) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("channel.title")} description={t("channel.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input className="max-w-sm" value={draft} placeholder={t("channel.namePlaceholder")} onChange={(event) => setDraft(event.target.value)} />
        <Select value={query.type || "ALL"} onValueChange={(value) => query.update({ type: value === "ALL" ? "" : value })}>
          <SelectTrigger className="w-44">
            <SelectValue placeholder={t("channel.typeFilter")}>
              <span className="flex items-center gap-2">
                <ChannelLogo type={query.type || "SMTP"} className="size-4" />
                <span>{query.type ? enumLabel(t, query.type) : t("channel.allTypes")}</span>
              </span>
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="ALL">{t("channel.allTypes")}</SelectItem>
            {types.map((item) => (
              <SelectItem key={item} value={item}>
                <span className="flex items-center gap-2">
                  <ChannelLogo type={item} className="size-4" />
                  <span>{enumLabel(t, item)}</span>
                </span>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={query.app || "ALL"} onValueChange={(value) => query.update({ app: value === "ALL" ? "" : value })}>
          <SelectTrigger className="w-44">
            <SelectValue placeholder={t("channel.appFilter")}>
              <span className="flex items-center gap-2">
                <ChannelLogo type="WEBHOOK" app={query.app || "OTHER"} className="size-4" />
                <span>{query.app ? enumLabel(t, query.app) : t("channel.allApps")}</span>
              </span>
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="ALL">{t("channel.allApps")}</SelectItem>
            {["OTHER", ...webhookApps.filter((item) => item !== "OTHER")].map((item) => (
              <SelectItem key={item} value={item}>
                <span className="flex items-center gap-2">
                  <ChannelLogo type="WEBHOOK" app={item} className="size-4" />
                  <span>{enumLabel(t, item)}</span>
                </span>
              </SelectItem>
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
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("channel.empty")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("channel.name")}</TableHead>
                <TableTypeHead>{t("channel.type")}</TableTypeHead>
                <TableHead>{t("channel.app")}</TableHead>
                <TableHead>{t("channel.endpoint")}</TableHead>
                <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>
                    <span className="flex min-w-0 items-center gap-2">
                      <ChannelLogo
                        type={item.type}
                        app={item.app}
                        className="size-5 rounded-md border bg-background p-0.5"
                      />
                      <span className="truncate">{item.name}</span>
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="flex min-w-0 items-center gap-2">
                      <ChannelLogo type={item.type} app={item.app} className="size-4" />
                      <StatusBadge value={item.type} />
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="flex min-w-0 items-center gap-2">
                      <ChannelLogo type={item.type} app={item.app} className="size-4" />
                      <StatusBadge value={item.app} />
                    </span>
                  </TableCell>
                  <TableCell className="max-w-xs truncate">{item.endpoint}</TableCell>
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
        <DialogContent className="max-h-[80vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4 sm:grid-cols-2" onSubmit={onSave}>
            <FormField
              label={t("channel.name")}
              htmlFor="channel-name"
              required
              error={fieldErrors.name}
            >
              <Input id="channel-name" value={name} placeholder={t("channel.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <FormField
              label={t("channel.type")}
              required
              error={fieldErrors.type}
            >
              <Select
                value={type}
                onValueChange={(value) => {
                  setType(value)
                  setApp(value === "SMS" ? "SMS_ALICLOUD" : "OTHER")
                  setError("")
                  setFieldErrors({})
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t("channel.typePlaceholder")}>
                    <span className="flex items-center gap-2">
                      <ChannelLogo type={type} className="size-4" />
                      <span>{enumLabel(t, type)}</span>
                    </span>
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {types.map((item) => (
                    <SelectItem key={item} value={item}>
                      <span className="flex items-center gap-2">
                        <ChannelLogo type={item} className="size-4" />
                        <span>{enumLabel(t, item)}</span>
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            {type === "WEBHOOK" ? (
              <FormField
                label={t("channel.app")}
                required
                error={fieldErrors.app}
              >
                <Select value={app} onValueChange={setApp}>
                  <SelectTrigger>
                    <SelectValue placeholder={t("channel.appPlaceholder")}>
                      <span className="flex items-center gap-2">
                        <ChannelLogo type="WEBHOOK" app={app} className="size-4" />
                        <span>{enumLabel(t, app)}</span>
                      </span>
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    {webhookApps.map((item) => (
                      <SelectItem key={item} value={item}>
                        <span className="flex items-center gap-2">
                          <ChannelLogo type="WEBHOOK" app={item} className="size-4" />
                          <span>{enumLabel(t, item)}</span>
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            ) : null}
            <FormField
              label={t("channel.endpoint")}
              htmlFor="channel-endpoint"
              className="sm:col-span-full"
              required
              error={fieldErrors.endpoint}
            >
              <Input
                id="channel-endpoint"
                value={endpoint}
                placeholder={
                  type === "SMTP"
                    ? t("channel.endpointSmtpPlaceholder")
                    : type === "SMS"
                      ? t("channel.endpointSmsPlaceholder")
                      : t("channel.endpointWebhookPlaceholder")
                }
                onChange={(event) => setEndpoint(event.target.value)}
                required
              />
            </FormField>
            {type === "SMTP" || type === "SMS" ? (
              <>
                <FormField
                  label={type === "SMS" ? t("channel.accessKeyId") : t("channel.username")}
                  htmlFor="channel-username"
                  required={type === "SMS"}
                  error={fieldErrors.username}
                >
                  <Input
                    id="channel-username"
                    value={username}
                    placeholder={type === "SMS" ? t("channel.accessKeyIdPlaceholder") : t("channel.usernamePlaceholder")}
                    onChange={(event) => setUsername(event.target.value)}
                  />
                </FormField>
                <FormField
                  label={type === "SMS" ? t("channel.accessKeySecret") : t("channel.password")}
                  htmlFor="channel-password"
                  required={type === "SMS" && !editing}
                  hint={editing ? t("form.keepSecret") : undefined}
                  error={fieldErrors.password}
                >
                  <Input
                    id="channel-password"
                    type="password"
                    value={password}
                    placeholder={type === "SMS" ? t("channel.accessKeySecretPlaceholder") : t("channel.passwordPlaceholder")}
                    onChange={(event) => setPassword(event.target.value)}
                  />
                </FormField>
                <FormField
                  label={type === "SMS" ? t("channel.smsSignName") : t("channel.fromAddress")}
                  htmlFor="channel-from"
                  required={type === "SMS"}
                  error={fieldErrors.fromAddress}
                >
                  <Input id="channel-from" value={fromAddress} placeholder={type === "SMS" ? t("channel.smsSignNamePlaceholder") : t("channel.fromAddressPlaceholder")} onChange={(event) => setFromAddress(event.target.value)} />
                </FormField>
              </>
            ) : (
              <FormField label={t("channel.secret")} htmlFor="channel-secret" hint={editing ? `${t("channel.secretHint")}；${t("form.keepSecret")}` : t("channel.secretHint")} className="sm:col-span-full">
                <Input id="channel-secret" type="password" value={secret} placeholder={t("channel.secretPlaceholder")} onChange={(event) => setSecret(event.target.value)} />
              </FormField>
            )}
            <FormField label={t("channel.remark")} htmlFor="channel-remark" className="sm:col-span-full">
              <Input id="channel-remark" value={remark} placeholder={t("channel.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
            </FormField>
            {type === "SMS" ? (
              <p className="text-sm text-muted-foreground sm:col-span-full">{t("channel.smsNotImplemented")}</p>
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
      {testing ? (
        <NotificationTestDialog target={{ kind: "channel", item: testing }} channels={[]} onClose={() => setTesting(null)} />
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
                  label: t("channel.type"),
                  value: (
                    <span className="flex items-center gap-2">
                      <ChannelLogo type={detail.type} app={detail.app} className="size-4" />
                      <StatusBadge value={detail.type} />
                    </span>
                  ),
                },
                {
                  label: t("channel.app"),
                  value: (
                    <span className="flex items-center gap-2">
                      <ChannelLogo type={detail.type} app={detail.app} className="size-4" />
                      <StatusBadge value={detail.app} />
                    </span>
                  ),
                },
                { label: t("channel.endpoint"), value: detail.endpoint },
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
