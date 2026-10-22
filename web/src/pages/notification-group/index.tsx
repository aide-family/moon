import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage, readFormError } from "@/lib/form-error"

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
import { ChannelLogo } from "@/components/channel-logo"
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
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { run } from "@/lib/action"

const phonePattern = /^\+?[0-9]{6,20}$/

function splitValues(raw: string): string[] {
  return raw
    .split(/[,，\s]+/)
    .map((value) => value.trim())
    .filter((value) => value !== "")
}

function firstInvalidPhone(raw: string): string | null {
  for (const value of splitValues(raw)) {
    if (!phonePattern.test(value)) {
      return value
    }
  }
  return null
}

function channelTypeOf(channel: NotificationChannelItem | undefined): "email" | "sms" | "webhook" {
  if (!channel) {
    return "webhook"
  }
  if (channel.type === "SMTP") {
    return "email"
  }
  if (channel.type === "SMS" || channel.app === "SMS_ALICLOUD") {
    return "sms"
  }
  return "webhook"
}

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
  const [emailChannelUid, setEmailChannelUid] = useState("")
  const [emailTemplateUid, setEmailTemplateUid] = useState("")
  const [smsChannelUid, setSMSChannelUid] = useState("")
  const [smsTemplateUid, setSMSTemplateUid] = useState("")
  const [webhookChannelUids, setWebhookChannelUids] = useState<string[]>([])
  const [webhookTemplates, setWebhookTemplates] = useState<Record<string, string>>({})
  const [emailReceivers, setEmailReceivers] = useState("")
  const [smsReceivers, setSMSReceivers] = useState("")
  const [emailMemberUids, setEmailMemberUids] = useState<string[]>([])
  const [smsMemberUids, setSMSMemberUids] = useState<string[]>([])
  const [memberFilter, setMemberFilter] = useState("")
  const [smsMemberFilter, setSMSMemberFilter] = useState("")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
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
      listMembers({ page: 1, pageSize: 100 }),
    ])
      .then(([channelResult, templateResult, memberResult]) => {
        if (active) {
          setChannels(channelResult.items ?? [])
          setTemplates(templateResult.items ?? [])
          setMembers(memberResult.items ?? [])
        }
      })
      .catch(() => {
        if (active) {
          setChannels([])
          setTemplates([])
          setMembers([])
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  const emailChannels = channels.filter((item) => item.type === "SMTP" && item.status === "ACTIVE")
  const smsChannels = channels.filter((item) => channelTypeOf(item) === "sms" && item.status === "ACTIVE")
  const webhookChannels = channels.filter((item) => channelTypeOf(item) === "webhook" && item.status === "ACTIVE")
  const joinedMembers = members.filter((item) => item.status === "JOINED")

  function emailTemplates(): NotificationTemplateItem[] {
    return templates.filter((item) => item.type === "EMAIL" && item.status === "ACTIVE")
  }

  function smsTemplates(): NotificationTemplateItem[] {
    return templates.filter((item) => item.type === "SMS_ALICLOUD" && item.status === "ACTIVE")
  }

  function webhookTemplatesFor(app: string): NotificationTemplateItem[] {
    return templates.filter(
      (item) =>
        item.type === "WEBHOOK" &&
        item.status === "ACTIVE" &&
        (item.app === "OTHER" || item.app === app),
    )
  }

  // 选中的 Webhook 可能覆盖多个应用；每个应用都需要一个模板，才能满足“模板覆盖所有通知方式”。
  function selectedWebhookApps(): string[] {
    const apps = new Set<string>()
    for (const uid of webhookChannelUids) {
      const channel = channels.find((item) => item.uid === uid)
      apps.add(channel?.app || "OTHER")
    }
    return [...apps]
  }

  function toggleMember(uid: string, kind: "email" | "sms") {
    if (kind === "email") {
      setEmailMemberUids((values) => (values.includes(uid) ? values.filter((value) => value !== uid) : [...values, uid]))
      return
    }
    setSMSMemberUids((values) => (values.includes(uid) ? values.filter((value) => value !== uid) : [...values, uid]))
  }

  function ensureWebhookTemplates(nextChannels: string[], current: Record<string, string>) {
    const next: Record<string, string> = {}
    for (const uid of nextChannels) {
      const channel = channels.find((item) => item.uid === uid)
      const app = channel?.app || "OTHER"
      const options = webhookTemplatesFor(app)
      next[app] = current[app] && options.some((item) => item.uid === current[app]) ? current[app] : options[0]?.uid ?? ""
    }
    return next
  }

  function selectEmailChannel(uid: string) {
    setEmailChannelUid(uid)
    const options = emailTemplates()
    if (!options.some((item) => item.uid === emailTemplateUid)) {
      setEmailTemplateUid(options[0]?.uid ?? "")
    }
    if (!uid) {
      setEmailReceivers("")
      setEmailMemberUids([])
    }
    setFieldErrors({})
  }

  function selectSMSChannel(uid: string) {
    setSMSChannelUid(uid)
    const options = smsTemplates()
    if (!options.some((item) => item.uid === smsTemplateUid)) {
      setSMSTemplateUid(options[0]?.uid ?? "")
    }
    if (!uid) {
      setSMSReceivers("")
      setSMSMemberUids([])
    }
    setFieldErrors({})
  }

  function openCreate() {
    setEditing(null)
    setName("")
    setEmailChannelUid("")
    setEmailTemplateUid("")
    setSMSChannelUid("")
    setSMSTemplateUid("")
    setWebhookChannelUids([])
    setWebhookTemplates({})
    setEmailReceivers("")
    setSMSReceivers("")
    setEmailMemberUids([])
    setSMSMemberUids([])
    setMemberFilter("")
    setSMSMemberFilter("")
    setRemark("")
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openEdit(item: NotificationGroupItem) {
    setEditing(item)
    setName(item.name)
    setEmailChannelUid("")
    setEmailTemplateUid("")
    setSMSChannelUid("")
    setSMSTemplateUid("")
    const webhooks: string[] = []
    const webhookTemplateMap: Record<string, string> = {}
    const channelUIDs = item.channelUids?.length ? item.channelUids : [item.channelUid, ...(item.webhookChannelUids ?? [])].filter(Boolean)
    const templateUIDs = item.templateUids?.length ? item.templateUids : [item.templateUid, ...(item.webhookChannelUids ?? []).map(() => item.webhookTemplateUid)].filter(Boolean)
    channelUIDs.forEach((uid, index) => {
      const channel = channels.find((entry) => entry.uid === uid)
      const kind = channelTypeOf(channel)
      const templateUID = templateUIDs[index] ?? ""
      if (kind === "email") {
        setEmailChannelUid(uid)
        setEmailTemplateUid(templateUID)
      } else if (kind === "sms") {
        setSMSChannelUid(uid)
        setSMSTemplateUid(templateUID)
      } else {
        webhooks.push(uid)
        const app = channel?.app || "OTHER"
        webhookTemplateMap[app] = templateUID
      }
    })
    setWebhookChannelUids(webhooks)
    setWebhookTemplates(webhookTemplateMap)
    setEmailReceivers(item.emailReceivers ?? item.receivers ?? "")
    setSMSReceivers(item.smsReceivers ?? "")
    setEmailMemberUids(item.emailMemberUids?.length ? item.emailMemberUids : (item.memberUids ?? []))
    setSMSMemberUids(item.smsMemberUids ?? [])
    setMemberFilter("")
    setSMSMemberFilter("")
    setRemark(item.remark)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function buildPayload() {
    const channelUids: string[] = []
    const templateUids: string[] = []
    if (emailChannelUid) {
      channelUids.push(emailChannelUid)
      templateUids.push(emailTemplateUid)
    }
    if (smsChannelUid) {
      channelUids.push(smsChannelUid)
      templateUids.push(smsTemplateUid)
    }
    for (const uid of webhookChannelUids) {
      const channel = channels.find((item) => item.uid === uid)
      const app = channel?.app || "OTHER"
      channelUids.push(uid)
      templateUids.push(webhookTemplates[app] ?? "")
    }
    return {
      name: name.trim(),
      channelUids,
      templateUids,
      emailReceivers: emailChannelUid ? emailReceivers : "",
      smsReceivers: smsChannelUid ? smsReceivers : "",
      emailMemberUids: emailChannelUid ? emailMemberUids : [],
      smsMemberUids: smsChannelUid ? smsMemberUids : [],
      remark,
    }
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const payload = buildPayload()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      [
        "channels",
        payload.channelUids.length === 0 ? t("group.channelRequired") : null,
      ],
      [
        "emailTemplate",
        emailChannelUid && !emailTemplateUid ? t("group.templateRequired") : null,
      ],
      [
        "smsTemplate",
        smsChannelUid && !smsTemplateUid ? t("group.templateRequired") : null,
      ],
      [
        "receivers",
        emailChannelUid && !emailReceivers.trim() && emailMemberUids.length === 0
          ? t("group.receiversSmtpHint")
          : null,
      ],
      [
        "smsReceivers",
        smsChannelUid && !smsReceivers.trim() && smsMemberUids.length === 0
          ? t("group.receiversSmsHint")
          : null,
      ],
    ])
    if (emailChannelUid && emailReceivers.trim() && firstInvalidReceiver(emailReceivers)) {
      errors.receivers = t("form.invalidEmail")
    }
    if (smsChannelUid && smsReceivers.trim() && firstInvalidPhone(smsReceivers)) {
      errors.smsReceivers = t("form.invalidPhone")
    }
    for (const app of selectedWebhookApps()) {
      if (!webhookTemplates[app]) {
        errors.webhookTemplate = t("group.webhookTemplateRequired", { app: enumLabel(t, app) })
        break
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
      if (editing) {
        await updateNotificationGroup({ ...payload, uid: editing.uid })
      } else {
        await createNotificationGroup(payload)
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

  function channelItems(uid: string): NotificationChannelItem | undefined {
    return channels.find((item) => item.uid === uid)
  }

  function receiversSummary(item: NotificationGroupItem): string {
    const parts: string[] = []
    const emailCount = (item.emailMemberUids ?? item.memberUids ?? []).length
    if (item.emailReceivers || item.receivers) {
      parts.push(`${t("group.emailReceivers")}: ${item.emailReceivers || item.receivers}`)
    }
    if (emailCount > 0) {
      parts.push(`${t("group.memberCount", { count: emailCount })}`)
    }
    if (item.smsReceivers) {
      parts.push(`${t("group.smsReceivers")}: ${item.smsReceivers}`)
    }
    if ((item.smsMemberUids ?? []).length > 0) {
      parts.push(`${t("group.smsMemberCount", { count: (item.smsMemberUids ?? []).length })}`)
    }
    const webhookCount = (item.channelUids ?? []).filter(
      (uid) => channelTypeOf(channelItems(uid)) === "webhook",
    ).length
    if (webhookCount > 0) {
      parts.push(t("group.webhookCount", { count: webhookCount }))
    }
    return parts.length > 0 ? parts.join(" · ") : t("group.noReceivers")
  }

  const selectedWebhookAppList = selectedWebhookApps()

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
                <TableHead>{t("group.channels")}</TableHead>
                <TableHead>{t("group.receiversSummary")}</TableHead>
                <TableHead>{t("namespace.status")}</TableHead>
                <TableHead className="text-right">{t("action.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-x-3 gap-y-1">
                      {(item.channelUids ?? []).map((uid) => {
                        const channel = channelItems(uid)
                        return (
                          <span key={uid} className="flex min-w-0 items-center gap-1.5 text-sm">
                            <ChannelLogo type={channel?.type ?? "WEBHOOK"} app={channel?.app} className="size-4" />
                            <span className="truncate">{channel?.name ?? uid}</span>
                          </span>
                        )
                      })}
                    </div>
                  </TableCell>
                  <TableCell className="max-w-md text-sm text-muted-foreground">{receiversSummary(item)}</TableCell>
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
        <DialogContent className="max-h-[80vh] overflow-y-auto sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4" onSubmit={onSave}>
            <FormField label={t("group.name")} htmlFor="group-name" required error={fieldErrors.name}>
              <Input id="group-name" value={name} placeholder={t("group.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField label={t("group.emailChannel")} hint={t("group.emailChannelHint")} error={fieldErrors.channels}>
                <Select value={emailChannelUid || "NONE"} onValueChange={(value) => selectEmailChannel(value === "NONE" ? "" : value)}>
                  <SelectTrigger>
                    <SelectValue placeholder={t("group.emailChannelPlaceholder")}>
                      {channelItems(emailChannelUid) ? (
                        <span className="flex min-w-0 items-center gap-2">
                          <ChannelLogo type="SMTP" className="size-4" />
                          <span className="truncate">{channelItems(emailChannelUid)?.name}</span>
                        </span>
                      ) : (
                        <span className="text-muted-foreground">{t("group.emailChannelPlaceholder")}</span>
                      )}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectOption value="NONE">{t("group.channelPlaceholder")}</SelectOption>
                    {emailChannels.map((item) => (
                      <SelectOption key={item.uid} value={item.uid}>
                        <span className="flex items-center gap-2">
                          <ChannelLogo type="SMTP" className="size-4" />
                          <span>{item.name}</span>
                        </span>
                      </SelectOption>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
              {emailChannelUid ? (
                <FormField label={t("group.emailTemplate")} required error={fieldErrors.emailTemplate}>
                  <Select value={emailTemplateUid || "NONE"} onValueChange={(value) => setEmailTemplateUid(value === "NONE" ? "" : value)}>
                    <SelectTrigger>
                      <SelectValue placeholder={t("group.templatePlaceholder")}>
                        {templates.find((item) => item.uid === emailTemplateUid)?.name ?? t("group.templatePlaceholder")}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectOption value="NONE">{t("group.templatePlaceholder")}</SelectOption>
                      {emailTemplates().map((item) => (
                        <SelectOption key={item.uid} value={item.uid}>
                          {item.name}
                        </SelectOption>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
              ) : null}
              <FormField label={t("group.smsChannel")} hint={t("group.smsChannelHint")} error={fieldErrors.channels}>
                <Select value={smsChannelUid || "NONE"} onValueChange={(value) => selectSMSChannel(value === "NONE" ? "" : value)}>
                  <SelectTrigger>
                    <SelectValue placeholder={t("group.smsChannelPlaceholder")}>
                      {channelItems(smsChannelUid) ? (
                        <span className="flex min-w-0 items-center gap-2">
                          <ChannelLogo type="SMS" className="size-4" />
                          <span className="truncate">{channelItems(smsChannelUid)?.name}</span>
                        </span>
                      ) : (
                        <span className="text-muted-foreground">{t("group.smsChannelPlaceholder")}</span>
                      )}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectOption value="NONE">{t("group.channelPlaceholder")}</SelectOption>
                    {smsChannels.map((item) => (
                      <SelectOption key={item.uid} value={item.uid}>
                        <span className="flex items-center gap-2">
                          <ChannelLogo type="SMS" className="size-4" />
                          <span>{item.name}</span>
                        </span>
                      </SelectOption>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
              {smsChannelUid ? (
                <FormField label={t("group.smsTemplate")} required error={fieldErrors.smsTemplate}>
                  <Select value={smsTemplateUid || "NONE"} onValueChange={(value) => setSMSTemplateUid(value === "NONE" ? "" : value)}>
                    <SelectTrigger>
                      <SelectValue placeholder={t("group.templatePlaceholder")}>
                        {templates.find((item) => item.uid === smsTemplateUid)?.name ?? t("group.templatePlaceholder")}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectOption value="NONE">{t("group.templatePlaceholder")}</SelectOption>
                      {smsTemplates().map((item) => (
                        <SelectOption key={item.uid} value={item.uid}>
                          {item.name}
                        </SelectOption>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
              ) : null}
            </div>
            {emailChannels.length === 0 && smsChannels.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("group.noPrimaryChannels")}</p>
            ) : null}
            <div className="grid gap-2">
              <p className="text-sm font-medium">{t("group.webhooks")}</p>
              <p className="text-sm text-muted-foreground">{t("group.webhooksHint")}</p>
              <div className="flex flex-wrap gap-2 rounded-md border p-3">
                {webhookChannels.map((item) => (
                  <label
                    key={item.uid}
                    className="inline-flex cursor-pointer items-center gap-2 rounded-md border bg-background px-3 py-1.5 text-sm"
                  >
                    <input
                      type="checkbox"
                      checked={webhookChannelUids.includes(item.uid)}
                      onChange={() => {
                        const next = webhookChannelUids.includes(item.uid)
                          ? webhookChannelUids.filter((uid) => uid !== item.uid)
                          : [...webhookChannelUids, item.uid]
                        setWebhookChannelUids(next)
                        setWebhookTemplates((current) => ensureWebhookTemplates(next, current))
                      }}
                    />
                    <ChannelLogo type="WEBHOOK" app={item.app} className="size-4" />
                    <span>{item.name}</span>
                    <span className="text-muted-foreground">{enumLabel(t, item.app)}</span>
                  </label>
                ))}
                {webhookChannels.length === 0 ? <p className="text-sm text-muted-foreground">{t("group.webhookEmpty")}</p> : null}
              </div>
            </div>
            {selectedWebhookAppList.length > 0 ? (
              <div className="grid gap-4 sm:grid-cols-2">
                {selectedWebhookAppList.map((app) => (
                  <FormField
                    key={app}
                    label={t("group.webhookTemplateFor", { app: enumLabel(t, app) })}
                    required
                    error={fieldErrors.webhookTemplate}
                  >
                    <Select
                      value={webhookTemplates[app] || "NONE"}
                      onValueChange={(value) =>
                        setWebhookTemplates((current) => ({ ...current, [app]: value === "NONE" ? "" : value }))
                      }
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={t("group.templatePlaceholder")}>
                          <span className="flex min-w-0 items-center gap-2">
                            <ChannelLogo type="WEBHOOK" app={app} className="size-4" />
                            <span className="truncate">
                              {templates.find((item) => item.uid === webhookTemplates[app])?.name ?? t("group.templatePlaceholder")}
                            </span>
                          </span>
                        </SelectValue>
                      </SelectTrigger>
                      <SelectContent>
                        <SelectOption value="NONE">{t("group.templatePlaceholder")}</SelectOption>
                        {webhookTemplatesFor(app).map((item) => (
                          <SelectOption key={item.uid} value={item.uid}>
                            <span className="flex items-center gap-2">
                              <ChannelLogo type="WEBHOOK" app={item.app} className="size-4" />
                              <span>{item.name}</span>
                            </span>
                          </SelectOption>
                        ))}
                      </SelectContent>
                    </Select>
                  </FormField>
                ))}
              </div>
            ) : null}
            {emailChannelUid ? (
              <div className="grid gap-3 rounded-lg border p-3">
                <p className="text-sm font-medium">{t("group.members")}</p>
                <p className="text-sm text-muted-foreground">{t("group.membersHint")}</p>
                <FormField
                  label={t("group.receivers")}
                  htmlFor="group-email-receivers"
                  hint={t("group.receiversSmtpHint")}
                  error={fieldErrors.receivers}
                >
                  <Input
                    id="group-email-receivers"
                    value={emailReceivers}
                    placeholder={t("group.receiversPlaceholder")}
                    onChange={(event) => setEmailReceivers(event.target.value)}
                  />
                </FormField>
                <Input
                  value={memberFilter}
                  placeholder={t("group.memberFilterPlaceholder")}
                  onChange={(event) => setMemberFilter(event.target.value)}
                />
                <div className="flex max-h-44 flex-wrap gap-2 overflow-y-auto">
                  {joinedMembers
                    .filter((item) => {
                      const keyword = memberFilter.trim().toLowerCase()
                      if (!keyword) {
                        return true
                      }
                      return (item.name || item.email).toLowerCase().includes(keyword) || item.email.toLowerCase().includes(keyword)
                    })
                    .map((item) => (
                      <label
                        key={item.uid}
                        className="inline-flex cursor-pointer items-center gap-2 rounded-md border bg-background px-3 py-1.5 text-sm"
                      >
                        <input
                          type="checkbox"
                          checked={emailMemberUids.includes(item.uid)}
                          onChange={() => toggleMember(item.uid, "email")}
                        />
                        <span>{item.name || item.email}</span>
                        <span className="text-muted-foreground">{item.email}</span>
                      </label>
                    ))}
                  {members.length === 0 ? <p className="text-sm text-muted-foreground">{t("group.memberEmpty")}</p> : null}
                </div>
              </div>
            ) : null}
            {smsChannelUid ? (
              <div className="grid gap-3 rounded-lg border p-3">
                <p className="text-sm font-medium">{t("group.smsMembers")}</p>
                <p className="text-sm text-muted-foreground">{t("group.smsMembersHint")}</p>
                <FormField
                  label={t("group.smsReceivers")}
                  htmlFor="group-sms-receivers"
                  hint={t("group.receiversSmsHint")}
                  error={fieldErrors.smsReceivers}
                >
                  <Input
                    id="group-sms-receivers"
                    value={smsReceivers}
                    placeholder={t("group.smsReceiversPlaceholder")}
                    onChange={(event) => setSMSReceivers(event.target.value)}
                  />
                </FormField>
                <Input
                  value={smsMemberFilter}
                  placeholder={t("group.memberFilterPlaceholder")}
                  onChange={(event) => setSMSMemberFilter(event.target.value)}
                />
                <div className="flex max-h-44 flex-wrap gap-2 overflow-y-auto">
                  {joinedMembers
                    .filter((item) => {
                      const keyword = smsMemberFilter.trim().toLowerCase()
                      if (!keyword) {
                        return true
                      }
                      return (item.name || item.email).toLowerCase().includes(keyword) || (item.phone || "").includes(keyword)
                    })
                    .map((item) => (
                      <label
                        key={item.uid}
                        className="inline-flex cursor-pointer items-center gap-2 rounded-md border bg-background px-3 py-1.5 text-sm"
                      >
                        <input
                          type="checkbox"
                          disabled={!item.phone}
                          checked={smsMemberUids.includes(item.uid)}
                          onChange={() => toggleMember(item.uid, "sms")}
                        />
                        <span>{item.name || item.email}</span>
                        <span className="text-muted-foreground">{item.phone || t("group.noPhone")}</span>
                      </label>
                    ))}
                  {members.length === 0 ? <p className="text-sm text-muted-foreground">{t("group.memberEmpty")}</p> : null}
                </div>
              </div>
            ) : null}
            <FormField label={t("group.remark")} htmlFor="group-remark">
              <Input id="group-remark" value={remark} placeholder={t("group.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
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
                  label: t("group.channels"),
                  value: (
                    <div className="flex flex-wrap gap-x-3 gap-y-1">
                      {(detail.channelUids ?? []).map((uid) => {
                        const channel = channelItems(uid)
                        return (
                          <span key={uid} className="flex items-center gap-1.5">
                            <ChannelLogo type={channel?.type ?? "WEBHOOK"} app={channel?.app} className="size-4" />
                            {channel?.name ?? uid}
                          </span>
                        )
                      })}
                    </div>
                  ),
                },
                {
                  label: t("group.receivers"),
                  value: <span className="whitespace-pre-wrap">{receiversSummary(detail)}</span>,
                },
                {
                  label: t("group.members"),
                  value: (detail.emailMemberUids ?? detail.memberUids ?? [])
                    .map((uid) => {
                      const item = members.find((member) => member.uid === uid)
                      return item ? item.name || item.email : uid
                    })
                    .join(", "),
                },
                {
                  label: t("group.smsMembers"),
                  value: (detail.smsMemberUids ?? [])
                    .map((uid) => {
                      const item = members.find((member) => member.uid === uid)
                      return item ? item.name || item.phone : uid
                    })
                    .join(", "),
                },
                { label: t("group.remark"), value: detail.remark },
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
