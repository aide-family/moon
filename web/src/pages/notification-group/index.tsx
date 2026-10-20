import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage } from "@/lib/form-error"

import { listMembers } from "@/api/member"
import type { MemberItem } from "@/api/member/types"
import { listNotificationChannels } from "@/api/notification-channel"
import type { NotificationChannelItem } from "@/api/notification-channel/types"
import {
  createNotificationGroup,
  deleteNotificationGroup,
  listNotificationGroups,
  updateNotificationGroup,
  updateNotificationGroupStatus,
} from "@/api/notification-group"
import type { NotificationGroupItem } from "@/api/notification-group/types"
import { listNotificationTemplates } from "@/api/notification-template"
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
import { firstInvalidReceiver } from "@/lib/email"
import { firstError, requireSelected, requireTrimmed } from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { run } from "@/lib/action"

export function NotificationGroupPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [channels, setChannels] = useState<NotificationChannelItem[]>([])
  const [templates, setTemplates] = useState<NotificationTemplateItem[]>([])
  const [members, setMembers] = useState<MemberItem[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationGroupItem | null>(null)
  const [name, setName] = useState("")
  const [channelUid, setChannelUid] = useState("")
  const [templateUid, setTemplateUid] = useState("")
  const [receivers, setReceivers] = useState("")
  const [memberUids, setMemberUids] = useState<string[]>([])
  const [webhookChannelUids, setWebhookChannelUids] = useState<string[]>([])
  const [webhookTemplateUid, setWebhookTemplateUid] = useState("")
  const [memberFilter, setMemberFilter] = useState("")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NotificationGroupItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; item: NotificationGroupItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("notification:write")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<NotificationGroupItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listNotificationGroups({ page, pageSize, name: query.name }),
  })

  useEffect(() => {
    let active = true
    void Promise.all([
      listNotificationChannels({ page: 1, pageSize: 100 }),
      listNotificationTemplates({ page: 1, pageSize: 100 }),
    ])
      .then(([channelResult, templateResult]) => {
        if (active) {
          setChannels(channelResult.items ?? [])
          setTemplates(templateResult.items ?? [])
        }
      })
      .catch(() => {
        if (active) {
          setChannels([])
          setTemplates([])
        }
      })
    void listMembers({ page: 1, pageSize: 100 })
      .then((result) => {
        if (active) {
          setMembers(result.items ?? [])
        }
      })
      .catch(() => {
        if (active) setMembers([])
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  function compatibleTemplates(kind: string | undefined): NotificationTemplateItem[] {
    const expected = kind === "SMTP" ? "EMAIL" : "WEBHOOK"
    return templates.filter((item) => item.type === expected)
  }

  function webhookTemplates(): NotificationTemplateItem[] {
    return templates.filter((item) => item.type === "WEBHOOK" && item.status === "ACTIVE")
  }

  function selectChannel(uid: string) {
    setChannelUid(uid)
    const channel = channels.find((item) => item.uid === uid)
    if (channel?.type !== "SMTP") {
      setMemberUids([])
    }
    if (channel?.type === "WEBHOOK") {
      setWebhookChannelUids((values) => values.filter((value) => value !== uid))
    }
    const options = compatibleTemplates(channel?.type)
    if (!options.some((item) => item.uid === templateUid)) {
      setTemplateUid(options[0]?.uid ?? "")
    }
  }

  function toggleMember(uid: string) {
    setMemberUids((values) => (values.includes(uid) ? values.filter((value) => value !== uid) : [...values, uid]))
  }

  function toggleWebhook(uid: string) {
    setWebhookChannelUids((values) => (values.includes(uid) ? values.filter((value) => value !== uid) : [...values, uid]))
  }

  function openCreate() {
    setEditing(null)
    setName("")
    const channel = channels.find((item) => item.status === "ACTIVE") ?? channels[0]
    setChannelUid(channel?.uid ?? "")
    setTemplateUid(compatibleTemplates(channel?.type)[0]?.uid ?? "")
    setReceivers("")
    setMemberUids([])
    setWebhookChannelUids([])
    setWebhookTemplateUid("")
    setMemberFilter("")
    setRemark("")
    setError("")
    setOpen(true)
  }

  function openEdit(item: NotificationGroupItem) {
    setEditing(item)
    setName(item.name)
    setChannelUid(item.channelUid)
    setTemplateUid(item.templateUid)
    setReceivers(item.receivers)
    setMemberUids(item.memberUids ?? [])
    setWebhookChannelUids(item.webhookChannelUids ?? [])
    setWebhookTemplateUid(item.webhookTemplateUid ?? "")
    setMemberFilter("")
    setRemark(item.remark)
    setError("")
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const channel = channels.find((item) => item.uid === channelUid)
    const invalid = firstError(
      requireTrimmed(name, t("form.required")),
      requireSelected(channelUid, t("form.required")),
      requireSelected(templateUid, t("form.required")),
      channel?.type === "SMTP" && !receivers.trim() && memberUids.length === 0 ? t("form.required") : null,
      webhookChannelUids.length > 0 && !webhookTemplateUid ? t("form.required") : null,
    )
    if (invalid) {
      setError(invalid)
      return
    }
    if (channel?.type === "SMTP" && firstInvalidReceiver(receivers)) {
      setError(t("form.invalidEmail"))
      return
    }
    setPending(true)
    setError("")
    try {
      const payload = {
        name: name.trim(), channelUid, templateUid, receivers, remark,
        memberUids, webhookChannelUids, webhookTemplateUid,
      }
      if (editing) {
        await updateNotificationGroup({ ...payload, uid: editing.uid })
      } else {
        await createNotificationGroup(payload)
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
        await deleteNotificationGroup(confirm.item.uid)
      } else {
        await updateNotificationGroupStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: NotificationGroupItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm({ action: "disable", item }) }
        : { key: "enable", label: t("action.enable"), onSelect: () => run(updateNotificationGroupStatus(item.uid, "ACTIVE").then(() => reload()), setError) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("group.title")} description={t("group.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input className="max-w-sm" value={draft} placeholder={t("group.namePlaceholder")} onChange={(event) => setDraft(event.target.value)} />
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
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("group.empty")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("group.name")}</TableHead>
                <TableHead>{t("group.channel")}</TableHead>
                <TableHead>{t("group.template")}</TableHead>
                <TableHead>{t("group.receivers")}</TableHead>
                <TableHead>{t("namespace.status")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell>{channels.find((channel) => channel.uid === item.channelUid)?.name ?? item.channelUid}</TableCell>
                  <TableCell>{templates.find((template) => template.uid === item.templateUid)?.name ?? item.templateUid}</TableCell>
                  <TableCell>
                    {item.receivers}
                    {((item.memberUids ?? []).length > 0 ? `${item.receivers ? " · " : ""}${t("group.memberCount", { count: (item.memberUids ?? []).length })}` : "")}
                  </TableCell>
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
            <FormField label={t("group.name")} htmlFor="group-name">
              <Input id="group-name" value={name} placeholder={t("group.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <FormField label={t("group.channel")}>
              <Select value={channelUid || "NONE"} onValueChange={(value) => selectChannel(value === "NONE" ? "" : value)}>
                <SelectTrigger>
                  <SelectValue placeholder={t("group.channelPlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectOption value="NONE">{t("group.channelPlaceholder")}</SelectOption>
                  {channels.map((item) => (
                    <SelectOption key={item.uid} value={item.uid}>
                      {item.name}（{item.type}{item.status === "DISABLED" ? ` · ${t("action.disabled")}` : ""}）
                    </SelectOption>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            <FormField label={t("group.template")}>
              <Select value={templateUid || "NONE"} onValueChange={(value) => setTemplateUid(value === "NONE" ? "" : value)}>
                <SelectTrigger>
                  <SelectValue placeholder={t("group.templatePlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectOption value="NONE">{t("group.templatePlaceholder")}</SelectOption>
                  {compatibleTemplates(channels.find((item) => item.uid === channelUid)?.type).map((item) => (
                    <SelectOption key={item.uid} value={item.uid}>
                      {item.name}
                    </SelectOption>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            {channels.find((item) => item.uid === channelUid)?.type === "SMTP" ? (
              <FormField label={t("group.receivers")} htmlFor="group-receivers" hint={t("group.receiversSmtpHint")} className="sm:col-span-full">
                <Input id="group-receivers" value={receivers} placeholder={t("group.receiversPlaceholder")} onChange={(event) => setReceivers(event.target.value)} />
              </FormField>
            ) : (
              <p className="text-sm text-muted-foreground sm:col-span-full">{t("group.receiversWebhookHint")}</p>
            )}
            {channels.find((item) => item.uid === channelUid)?.type === "SMTP" ? (
              <div className="grid gap-2 sm:col-span-full">
                <p className="text-sm font-medium">{t("group.members")}</p>
                <p className="text-sm text-muted-foreground">{t("group.membersHint")}</p>
                <Input
                  value={memberFilter}
                  placeholder={t("group.memberFilterPlaceholder")}
                  onChange={(event) => setMemberFilter(event.target.value)}
                />
                <div className="grid max-h-44 gap-2 overflow-y-auto rounded-md border p-3">
                  {members
                    .filter((item) => item.status === "JOINED")
                    .filter((item) => {
                      const keyword = memberFilter.trim().toLowerCase()
                      if (!keyword) {
                        return true
                      }
                      return (item.name || item.email).toLowerCase().includes(keyword) || item.email.toLowerCase().includes(keyword)
                    })
                    .map((item) => (
                      <label key={item.uid} className="flex items-center gap-2 text-sm">
                        <input
                          type="checkbox"
                          checked={memberUids.includes(item.uid)}
                          onChange={() => toggleMember(item.uid)}
                        />
                        <span>{item.name || item.email}</span>
                        <span className="text-muted-foreground">{item.email}</span>
                      </label>
                    ))}
                  {members.length === 0 ? <p className="text-sm text-muted-foreground">{t("group.memberEmpty")}</p> : null}
                </div>
              </div>
            ) : null}
            <div className="grid gap-2 sm:col-span-full">
              <p className="text-sm font-medium">{t("group.webhooks")}</p>
              <p className="text-sm text-muted-foreground">{t("group.webhooksHint")}</p>
              <div className="grid max-h-44 gap-2 overflow-y-auto rounded-md border p-3">
                {channels
                  .filter((item) => item.type === "WEBHOOK" && item.status === "ACTIVE" && item.uid !== channelUid)
                  .map((item) => (
                    <label key={item.uid} className="flex items-center gap-2 text-sm">
                      <input
                        type="checkbox"
                        checked={webhookChannelUids.includes(item.uid)}
                        onChange={() => toggleWebhook(item.uid)}
                      />
                      {item.name}
                    </label>
                  ))}
                {channels.filter((item) => item.type === "WEBHOOK" && item.status === "ACTIVE" && item.uid !== channelUid).length === 0 ? (
                  <p className="text-sm text-muted-foreground">{t("group.webhookEmpty")}</p>
                ) : null}
              </div>
            </div>
            {webhookChannelUids.length > 0 ? (
              <FormField label={t("group.webhookTemplate")} className="sm:col-span-full">
                <Select value={webhookTemplateUid || "NONE"} onValueChange={(value) => setWebhookTemplateUid(value === "NONE" ? "" : value)}>
                  <SelectTrigger>
                    <SelectValue placeholder={t("group.webhookTemplatePlaceholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectOption value="NONE">{t("group.webhookTemplatePlaceholder")}</SelectOption>
                    {webhookTemplates().map((item) => (
                      <SelectOption key={item.uid} value={item.uid}>
                        {item.name}
                      </SelectOption>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            ) : null}
            <FormField label={t("group.remark")} htmlFor="group-remark" className="sm:col-span-full">
              <Input id="group-remark" value={remark} placeholder={t("group.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
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
                { label: t("group.channel"), value: channels.find((channel) => channel.uid === detail.channelUid)?.name ?? detail.channelUid },
                { label: t("group.template"), value: templates.find((template) => template.uid === detail.templateUid)?.name ?? detail.templateUid },
                { label: t("group.receivers"), value: detail.receivers },
                {
                  label: t("group.members"),
                  value: (detail.memberUids ?? [])
                    .map((uid) => {
                      const item = members.find((member) => member.uid === uid)
                      return item ? item.name || item.email : uid
                    })
                    .join(", "),
                },
                {
                  label: t("group.webhooks"),
                  value: (detail.webhookChannelUids ?? [])
                    .map((uid) => channels.find((channel) => channel.uid === uid)?.name ?? uid)
                    .join(", "),
                },
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
