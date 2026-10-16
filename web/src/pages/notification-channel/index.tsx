import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  createNotificationChannel,
  deleteNotificationChannel,
  listNotificationChannels,
  testNotificationChannel,
  updateNotificationChannel,
  updateNotificationChannelStatus,
} from "@/api/notification-channel"
import type { NotificationChannelItem } from "@/api/notification-channel/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
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
import { Textarea } from "@/components/ui/textarea"
import { firstError, requireTrimmed } from "@/lib/form"
import { useListQuery } from "@/lib/list-query"

const types = ["SMTP", "WEBHOOK"]
const apps = ["OTHER", "DINGTALK", "WECHAT", "FEISHU", "SMS_ALICLOUD"]

export function NotificationChannelPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<NotificationChannelItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [open, setOpen] = useState(false)
  const [testOpen, setTestOpen] = useState(false)
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
  const [receivers, setReceivers] = useState("")
  const [subject, setSubject] = useState("")
  const [body, setBody] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NotificationChannelItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: NotificationChannelItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const shown = useRef(confirm)
  if (confirm) {
    shown.current = confirm
  }
  const canWrite = session.can("notification:write")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listNotificationChannels({ page: query.page, pageSize: query.pageSize, name: query.name })
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
  }, [query.page, query.pageSize, query.name, attempt, session.namespaceUid])

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
    setApp(item.app)
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
    setReceivers("")
    setSubject("test")
    setBody("hello")
    setError("")
    setTestOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const invalid = firstError(
      requireTrimmed(name, t("form.required")),
      requireTrimmed(type, t("form.required")),
      requireTrimmed(app, t("form.required")),
      requireTrimmed(endpoint, t("form.required")),
    )
    if (invalid) {
      setError(invalid)
      return
    }
    setPending(true)
    setError("")
    try {
      const body = {
        name: name.trim(),
        type,
        app,
        endpoint: endpoint.trim(),
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
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function onTest(event: React.FormEvent) {
    event.preventDefault()
    if (!testing) {
      return
    }
    const invalid = requireTrimmed(receivers, t("form.required"))
    if (invalid) {
      setError(invalid)
      return
    }
    setPending(true)
    setError("")
    try {
      await testNotificationChannel(testing.uid, { receivers: receivers.trim(), subject, body })
      setTestOpen(false)
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
        await deleteNotificationChannel(confirm.item.uid)
      } else {
        await updateNotificationChannelStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      setConfirmError(err instanceof Error ? err.message : t("state.failed"))
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
      { key: "test", label: t("action.test"), onSelect: () => openTest(item) },
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => void updateNotificationChannelStatus(item.uid, "ACTIVE").then(() => setAttempt((value) => value + 1)) },
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
            <p className="text-sm text-destructive">{t("state.failed")}</p>
            <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
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
                  <TableCell>{item.type}</TableCell>
                  <TableCell>{item.app}</TableCell>
                  <TableCell className="max-w-xs truncate">{item.endpoint}</TableCell>
                  <TableCell>{item.status}</TableCell>
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
              <Select value={type} onValueChange={setType}>
                <SelectTrigger>
                  <SelectValue placeholder={t("channel.typePlaceholder")} />
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
            <FormField label={t("channel.app")}>
              <Select value={app} onValueChange={setApp}>
                <SelectTrigger>
                  <SelectValue placeholder={t("channel.appPlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  {apps.map((item) => (
                    <SelectItem key={item} value={item}>
                      {item}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            <FormField label={t("channel.endpoint")} htmlFor="channel-endpoint" className="sm:col-span-full">
              <Input id="channel-endpoint" value={endpoint} placeholder={t("channel.endpointPlaceholder")} onChange={(event) => setEndpoint(event.target.value)} required />
            </FormField>
            <FormField label={t("channel.username")} htmlFor="channel-username">
              <Input id="channel-username" value={username} placeholder={t("channel.usernamePlaceholder")} onChange={(event) => setUsername(event.target.value)} />
            </FormField>
            <FormField label={t("channel.password")} htmlFor="channel-password">
              <Input id="channel-password" type="password" value={password} placeholder={t("channel.passwordPlaceholder")} onChange={(event) => setPassword(event.target.value)} />
            </FormField>
            <FormField label={t("channel.secret")} htmlFor="channel-secret">
              <Input id="channel-secret" type="password" value={secret} placeholder={t("channel.secretPlaceholder")} onChange={(event) => setSecret(event.target.value)} />
            </FormField>
            <FormField label={t("channel.fromAddress")} htmlFor="channel-from">
              <Input id="channel-from" value={fromAddress} placeholder={t("channel.fromAddressPlaceholder")} onChange={(event) => setFromAddress(event.target.value)} />
            </FormField>
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
      <Dialog open={testOpen} onOpenChange={setTestOpen}>
        <DialogContent className="max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{t("action.test")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={onTest}>
            <FormField label={t("channel.testReceivers")} htmlFor="test-receivers">
              <Input id="test-receivers" value={receivers} placeholder={t("channel.testReceiversPlaceholder")} onChange={(event) => setReceivers(event.target.value)} required />
            </FormField>
            <FormField label={t("channel.testSubject")} htmlFor="test-subject">
              <Input id="test-subject" value={subject} placeholder={t("channel.testSubjectPlaceholder")} onChange={(event) => setSubject(event.target.value)} />
            </FormField>
            <FormField label={t("channel.testBody")} htmlFor="test-body">
              <Textarea id="test-body" value={body} placeholder={t("channel.testBodyPlaceholder")} onChange={(event) => setBody(event.target.value)} />
            </FormField>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending}>
                {t("action.test")}
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
                { label: t("channel.type"), value: detail.type },
                { label: t("channel.app"), value: detail.app },
                { label: t("channel.endpoint"), value: detail.endpoint },
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
