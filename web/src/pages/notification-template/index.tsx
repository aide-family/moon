import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { selectNotificationChannels } from "@/api/notification-channel"
import type { SelectItem } from "@/api/notification-channel/types"
import {
  createNotificationTemplate,
  deleteNotificationTemplate,
  listNotificationTemplates,
  testNotificationTemplate,
  updateNotificationTemplate,
  updateNotificationTemplateStatus,
} from "@/api/notification-template"
import type { NotificationTemplateItem } from "@/api/notification-template/types"
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
import { useListQuery } from "@/lib/list-query"

const types = ["EMAIL", "WEBHOOK", "SMS_ALICLOUD"]

export function NotificationTemplatePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<NotificationTemplateItem[]>([])
  const [channels, setChannels] = useState<SelectItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [open, setOpen] = useState(false)
  const [testOpen, setTestOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationTemplateItem | null>(null)
  const [testing, setTesting] = useState<NotificationTemplateItem | null>(null)
  const [name, setName] = useState("")
  const [type, setType] = useState("EMAIL")
  const [subject, setSubject] = useState("")
  const [body, setBody] = useState("")
  const [remark, setRemark] = useState("")
  const [channelUid, setChannelUid] = useState("")
  const [receivers, setReceivers] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NotificationTemplateItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: NotificationTemplateItem } | null>(null)
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
        const [result, channelResult] = await Promise.all([
          listNotificationTemplates({ page: query.page, pageSize: query.pageSize, name: query.name }),
          selectNotificationChannels({ limit: 100 }),
        ])
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
        setChannels(channelResult.items ?? [])
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
    setType("EMAIL")
    setSubject("")
    setBody("")
    setRemark("")
    setError("")
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
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      const payload = { name, type, subject, body, remark }
      if (editing) {
        await updateNotificationTemplate({ ...payload, uid: editing.uid })
      } else {
        await createNotificationTemplate(payload)
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
    setPending(true)
    setError("")
    try {
      await testNotificationTemplate(testing.uid, { channelUid, receivers })
      setTestOpen(false)
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
        await deleteNotificationTemplate(confirm.item.uid)
      } else {
        await updateNotificationTemplateStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      setConfirmError(err instanceof Error ? err.message : t("state.failed"))
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
          setChannelUid(channels[0]?.uid ?? "")
          setReceivers("")
          setError("")
          setTestOpen(true)
        },
      },
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => void updateNotificationTemplateStatus(item.uid, "ACTIVE").then(() => setAttempt((value) => value + 1)) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("template.title")} description={t("template.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
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
            <p className="text-sm text-destructive">{t("state.failed")}</p>
            <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
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
                  <TableCell>{item.type}</TableCell>
                  <TableCell>{item.subject}</TableCell>
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
        <DialogContent className="max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={onSave}>
            <FormField label={t("template.name")} htmlFor="tpl-name">
              <Input id="tpl-name" value={name} onChange={(event) => setName(event.target.value)} />
            </FormField>
            <FormField label={t("template.type")}>
              <Select value={type} onValueChange={setType}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {types.map((item) => (
                    <SelectOption key={item} value={item}>
                      {item}
                    </SelectOption>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            <FormField label={t("template.subject")} htmlFor="tpl-subject">
              <Input id="tpl-subject" value={subject} onChange={(event) => setSubject(event.target.value)} />
            </FormField>
            <FormField label={t("template.body")} htmlFor="tpl-body">
              <Textarea id="tpl-body" value={body} onChange={(event) => setBody(event.target.value)} />
            </FormField>
            <FormField label={t("template.remark")} htmlFor="tpl-remark">
              <Input id="tpl-remark" value={remark} onChange={(event) => setRemark(event.target.value)} />
            </FormField>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={testOpen} onOpenChange={setTestOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.test")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={onTest}>
            <FormField label={t("template.channel")}>
              <Select value={channelUid || "NONE"} onValueChange={(value) => setChannelUid(value === "NONE" ? "" : value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectOption value="NONE">{t("template.channel")}</SelectOption>
                  {channels.map((item) => (
                    <SelectOption key={item.uid} value={item.uid}>
                      {item.name}
                    </SelectOption>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            <FormField label={t("channel.testReceivers")} htmlFor="tpl-receivers">
              <Input id="tpl-receivers" value={receivers} onChange={(event) => setReceivers(event.target.value)} />
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
                { label: t("template.type"), value: detail.type },
                { label: t("template.subject"), value: detail.subject },
                { label: t("template.body"), value: detail.body },
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
