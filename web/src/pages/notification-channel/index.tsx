import { useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage } from "@/lib/form-error"

import {
  createNotificationChannel,
  deleteNotificationChannel,
  listNotificationChannels,
  updateNotificationChannel,
  updateNotificationChannelStatus,
} from "@/api/notification-channel"
import type { NotificationChannelItem } from "@/api/notification-channel/types"
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
import { isEmail } from "@/lib/email"
import { firstError, requireTrimmed } from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { run } from "@/lib/action"

const types = ["SMTP", "WEBHOOK"]
const webhookApps = ["OTHER", "DINGTALK", "WECHAT", "FEISHU"]

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
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NotificationChannelItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: NotificationChannelItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("notification:write")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<NotificationChannelItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listNotificationChannels({ page, pageSize, name: query.name }),
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
    setOpen(true)
  }

  function openTest(item: NotificationChannelItem) {
    setTesting(item)
    setError("")
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const invalid = firstError(
      requireTrimmed(name, t("form.required")),
      requireTrimmed(type, t("form.required")),
      requireTrimmed(endpoint, t("form.required")),
    )
    if (invalid) {
      setError(invalid)
      return
    }
    const target = endpoint.trim()
    if (type === "SMTP") {
      if (target.includes("://") || /\s/.test(target)) {
        setError(t("form.invalidEndpoint"))
        return
      }
      if (!fromAddress.trim() && !username.trim()) {
        setError(t("form.required"))
        return
      }
      if (fromAddress.trim() && !isEmail(fromAddress)) {
        setError(t("form.invalidEmail"))
        return
      }
    } else if (!/^https?:\/\//.test(target)) {
      setError(t("form.invalidUrl"))
      return
    }
    setPending(true)
    setError("")
    try {
      const body = {
        name: name.trim(),
        type,
        app: type === "SMTP" ? "OTHER" : app,
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
      if (confirm.action === "delete") {
        await deleteNotificationChannel(confirm.item.uid)
      } else {
        await updateNotificationChannelStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
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
      ...(item.status === "ACTIVE" ? [{ key: "test", label: t("action.test"), onSelect: () => openTest(item) }] : []),
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => run(updateNotificationChannelStatus(item.uid, "ACTIVE").then(() => reload()), setError) },
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
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("channel.empty")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("channel.name")}</TableHead>
                <TableHead>{t("channel.type")}</TableHead>
                <TableHead>{t("channel.app")}</TableHead>
                <TableHead>{t("channel.endpoint")}</TableHead>
                <TableHead>{t("namespace.status")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell>{enumLabel(t, item.type)}</TableCell>
                  <TableCell>{enumLabel(t, item.app)}</TableCell>
                  <TableCell className="max-w-xs truncate">{item.endpoint}</TableCell>
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
          <form className="grid gap-4 sm:grid-cols-2" onSubmit={onSave}>
            <FormField label={t("channel.name")} htmlFor="channel-name">
              <Input id="channel-name" value={name} placeholder={t("channel.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <FormField label={t("channel.type")}>
              <Select
                value={type}
                onValueChange={(value) => {
                  setType(value)
                  setApp("OTHER")
                  setError("")
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t("channel.typePlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  {types.map((item) => (
                    <SelectItem key={item} value={item}>
                      {enumLabel(t, item)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            {type === "WEBHOOK" ? (
              <FormField label={t("channel.app")}>
                <Select value={app} onValueChange={setApp}>
                  <SelectTrigger>
                    <SelectValue placeholder={t("channel.appPlaceholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    {webhookApps.map((item) => (
                      <SelectItem key={item} value={item}>
                        {enumLabel(t, item)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            ) : null}
            <FormField label={t("channel.endpoint")} htmlFor="channel-endpoint" className="sm:col-span-full">
              <Input
                id="channel-endpoint"
                value={endpoint}
                placeholder={type === "SMTP" ? t("channel.endpointSmtpPlaceholder") : t("channel.endpointWebhookPlaceholder")}
                onChange={(event) => setEndpoint(event.target.value)}
                required
              />
            </FormField>
            {type === "SMTP" ? (
              <>
                <FormField label={t("channel.username")} htmlFor="channel-username">
                  <Input id="channel-username" value={username} placeholder={t("channel.usernamePlaceholder")} onChange={(event) => setUsername(event.target.value)} />
                </FormField>
                <FormField label={t("channel.password")} htmlFor="channel-password" hint={editing ? t("form.keepSecret") : undefined}>
                  <Input id="channel-password" type="password" value={password} placeholder={t("channel.passwordPlaceholder")} onChange={(event) => setPassword(event.target.value)} />
                </FormField>
                <FormField label={t("channel.fromAddress")} htmlFor="channel-from">
                  <Input id="channel-from" value={fromAddress} placeholder={t("channel.fromAddressPlaceholder")} onChange={(event) => setFromAddress(event.target.value)} />
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
                { label: t("channel.type"), value: enumLabel(t, detail.type) },
                { label: t("channel.app"), value: enumLabel(t, detail.app) },
                { label: t("channel.endpoint"), value: detail.endpoint },
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
