import { Mail, MessageSquareText, Webhook } from "lucide-react"

import dingtalk from "@/assets/brands/dingtalk.svg"
import feishu from "@/assets/brands/feishu.svg"
import wecom from "@/assets/brands/wecom.svg"
import { cn } from "@/lib/class-names"

const appLogos: Record<string, string> = {
  DINGTALK: dingtalk,
  WECHAT: wecom,
  FEISHU: feishu,
}

/**
 * 通知渠道的产品标记：Webhook 应用优先展示对应产品 logo，
 * 否则按渠道类型回退到邮件或 Webhook 图标。
 */
export function ChannelLogo({
  type,
  app,
  className,
}: {
  type: string
  app?: string
  className?: string
}) {
  const logo = app ? appLogos[app] : undefined
  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center justify-center",
        className
      )}
      aria-hidden
    >
      {logo ? (
        <img src={logo} alt="" className="size-full object-contain" />
      ) : type === "SMTP" || type === "EMAIL" ? (
        <Mail className="size-full text-primary" />
      ) : type === "SMS" || type === "SMS_ALICLOUD" ? (
        <MessageSquareText className="size-full text-emerald-600 dark:text-emerald-400" />
      ) : (
        <Webhook className="size-full text-muted-foreground" />
      )}
    </span>
  )
}
