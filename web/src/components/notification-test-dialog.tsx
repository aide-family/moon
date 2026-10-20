import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { testNotificationChannel } from "@/api/notification-channel"
import type { NotificationChannelItem } from "@/api/notification-channel/types"
import { testNotificationTemplate } from "@/api/notification-template"
import type { NotificationTemplateItem } from "@/api/notification-template/types"
import { ChannelLogo } from "@/components/channel-logo"
import { FormField } from "@/components/form-field"
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
import { Textarea } from "@/components/ui/textarea"
import { firstInvalidReceiver } from "@/lib/email"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireSelected,
} from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { readFormError } from "@/lib/form-error"

type TestTarget =
  | { kind: "channel"; item: NotificationChannelItem }
  | { kind: "template"; item: NotificationTemplateItem }

interface NotificationTestDialogProps {
  target: TestTarget
  channels: NotificationChannelItem[]
  onClose: () => void
}

// 渠道试发和模板试发共用同一个弹窗；差异只在“是否需要选渠道/填正文”。
export function NotificationTestDialog({ target, channels, onClose }: NotificationTestDialogProps) {
  const { t } = useTranslation()
  const isChannel = target.kind === "channel"
  const channelType = target.kind === "channel" ? target.item.type : target.item.type === "EMAIL" ? "SMTP" : "WEBHOOK"
  const compatibleChannels = channels.filter((item) => item.status === "ACTIVE" && item.type === channelType)
  const [channelUid, setChannelUid] = useState(() => compatibleChannels[0]?.uid ?? "")
  const [receivers, setReceivers] = useState("")
  const [subject, setSubject] = useState(() => (isChannel && channelType === "SMTP" ? "test" : ""))
  const [body, setBody] = useState("hello")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const selectedChannel = compatibleChannels.find((item) => item.uid === channelUid)

  async function submit(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      [
        "channel",
        isChannel ? null : requireSelected(channelUid, t("form.required")),
      ],
      [
        "receivers",
        channelType === "SMTP" && !receivers.trim()
          ? t("form.required")
          : null,
      ],
      [
        "subject",
        isChannel && channelType === "SMTP" && !subject.trim()
          ? t("form.required")
          : null,
      ],
      [
        "body",
        isChannel && !body.trim() ? t("form.required") : null,
      ],
    ])
    if (
      channelType === "SMTP" &&
      receivers.trim() &&
      firstInvalidReceiver(receivers)
    ) {
      errors.receivers = t("form.invalidEmail")
    }
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      if (target.kind === "channel") {
        await testNotificationChannel(target.item.uid, { receivers: receivers.trim(), subject, body })
      } else {
        await testNotificationTemplate(target.item.uid, { channelUid, receivers: receivers.trim() })
      }
      toast.success(t("channel.testDone"))
      onClose()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {t("action.test")} · {target.item.name}
          </DialogTitle>
        </DialogHeader>
        <form noValidate className="grid gap-4" onSubmit={submit}>
          {target.kind === "template" ? (
            <FormField
              label={t("template.channel")}
              required
              error={fieldErrors.channel}
            >
              <Select value={channelUid || "NONE"} onValueChange={(value) => setChannelUid(value === "NONE" ? "" : value)}>
                <SelectTrigger>
                  <SelectValue placeholder={t("template.channelPlaceholder")}>
                    {selectedChannel ? (
                      <span className="flex min-w-0 items-center gap-2">
                        <ChannelLogo type={selectedChannel.type} app={selectedChannel.app} className="size-4" />
                        <span className="truncate">{selectedChannel.name}</span>
                      </span>
                    ) : (
                      <span className="text-muted-foreground">{t("template.channelPlaceholder")}</span>
                    )}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="NONE">{t("template.channelPlaceholder")}</SelectItem>
                  {compatibleChannels.map((item) => (
                    <SelectItem key={item.uid} value={item.uid}>
                      <span className="flex items-center gap-2">
                        <ChannelLogo type={item.type} app={item.app} className="size-4" />
                        <span>
                          {item.name}（{enumLabel(t, item.type)}）
                        </span>
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          ) : null}
          {channelType === "SMTP" ? (
            <FormField
              label={t("channel.testReceivers")}
              htmlFor="notification-test-receivers"
              required
              error={fieldErrors.receivers}
            >
              <Input
                id="notification-test-receivers"
                value={receivers}
                placeholder={t("channel.testReceiversPlaceholder")}
                onChange={(event) => setReceivers(event.target.value)}
                required
              />
            </FormField>
          ) : null}
          {isChannel && channelType === "SMTP" ? (
            <FormField
              label={t("channel.testSubject")}
              htmlFor="notification-test-subject"
              required
              error={fieldErrors.subject}
            >
              <Input
                id="notification-test-subject"
                value={subject}
                placeholder={t("channel.testSubjectPlaceholder")}
                onChange={(event) => setSubject(event.target.value)}
              />
            </FormField>
          ) : null}
          {isChannel ? (
            <FormField
              label={t("channel.testBody")}
              htmlFor="notification-test-body"
              required
              error={fieldErrors.body}
            >
              <Textarea
                id="notification-test-body"
                value={body}
                placeholder={t("channel.testBodyPlaceholder")}
                onChange={(event) => setBody(event.target.value)}
              />
            </FormField>
          ) : null}
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              {t("action.cancel")}
            </Button>
            <Button type="submit" disabled={pending}>
              {t("action.test")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
