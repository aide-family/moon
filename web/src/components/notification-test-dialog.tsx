import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { testNotificationChannel } from "@/api/notification-channel"
import type { NotificationChannelItem } from "@/api/notification-channel/types"
import { testNotificationTemplate } from "@/api/notification-template"
import type { NotificationTemplateItem } from "@/api/notification-template/types"
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
import { requireSelected } from "@/lib/form"

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
  const [pending, setPending] = useState(false)

  async function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!isChannel) {
      const invalid = requireSelected(channelUid, t("form.required"))
      if (invalid) {
        setError(invalid)
        return
      }
    }
    if (isChannel && !body.trim()) {
      setError(t("form.required"))
      return
    }
    if (isChannel && channelType === "SMTP" && !subject.trim()) {
      setError(t("form.required"))
      return
    }
    if (channelType === "SMTP" && !receivers.trim()) {
      setError(t("form.required"))
      return
    }
    if (channelType === "SMTP" && firstInvalidReceiver(receivers)) {
      setError(t("form.invalidEmail"))
      return
    }
    setPending(true)
    setError("")
    try {
      if (target.kind === "channel") {
        await testNotificationChannel(target.item.uid, { receivers: receivers.trim(), subject, body })
      } else {
        await testNotificationTemplate(target.item.uid, { channelUid, receivers: receivers.trim() })
      }
      toast.success(t("channel.testDone"))
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
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
        <form className="grid gap-4" onSubmit={submit}>
          {target.kind === "template" ? (
            <FormField label={t("template.channel")}>
              <Select value={channelUid || "NONE"} onValueChange={(value) => setChannelUid(value === "NONE" ? "" : value)}>
                <SelectTrigger>
                  <SelectValue placeholder={t("template.channelPlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="NONE">{t("template.channelPlaceholder")}</SelectItem>
                  {compatibleChannels.map((item) => (
                    <SelectItem key={item.uid} value={item.uid}>
                      {item.name}（{item.type}）
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          ) : null}
          {channelType === "SMTP" ? (
            <FormField label={t("channel.testReceivers")} htmlFor="notification-test-receivers">
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
            <FormField label={t("channel.testSubject")} htmlFor="notification-test-subject">
              <Input
                id="notification-test-subject"
                value={subject}
                placeholder={t("channel.testSubjectPlaceholder")}
                onChange={(event) => setSubject(event.target.value)}
              />
            </FormField>
          ) : null}
          {isChannel ? (
            <FormField label={t("channel.testBody")} htmlFor="notification-test-body">
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
