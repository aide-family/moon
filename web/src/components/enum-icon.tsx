import { createElement } from "react"
import {
  Archive,
  BadgeCheck,
  Ban,
  BellRing,
  CalendarClock,
  ChartNoAxesCombined,
  CircleCheck,
  CircleDashed,
  CircleOff,
  CirclePause,
  CircleX,
  Clock3,
  Database,
  FilePenLine,
  Flame,
  Globe,
  HeartPulse,
  LoaderCircle,
  Lock,
  Mail,
  MessageSquareText,
  Radar,
  ScrollText,
  Send,
  ShieldCheck,
  UserRound,
  Waypoints,
  Webhook,
  Wifi,
  WifiOff,
  type LucideIcon,
} from "lucide-react"

import { cn } from "@/lib/class-names"

// 枚举筛选共用的图标：状态按语义给勾/叉/时钟，类型按域给数据库/日志/拨测等。
// 只做展示，不参与语义判断；未收录的值不渲染图标，标签文字仍然保留。
const icons: Record<string, LucideIcon> = {
  ACTIVE: CircleCheck,
  ENABLED: CircleCheck,
  OK: CircleCheck,
  ONLINE: Wifi,
  SENT: Send,
  PAID: CircleCheck,
  APPROVED: CircleCheck,
  ACCEPTED: CircleCheck,
  JOINED: CircleCheck,
  RECOVERED: HeartPulse,
  DONE: CircleCheck,
  DISABLED: CircleOff,
  OFFLINE: WifiOff,
  PENDING: Clock3,
  PAUSED: CirclePause,
  SENDING: Clock3,
  WORKING: Clock3,
  RUNNING: LoaderCircle,
  DRAFT: FilePenLine,
  PUBLISHED: BadgeCheck,
  SUCCEEDED: CircleCheck,
  PARTIAL: CircleDashed,
  INVITED: Mail,
  FAILED: CircleX,
  REJECTED: CircleX,
  CANCELLED: CircleX,
  DECLINED: CircleX,
  BANNED: Ban,
  EXPIRED: CalendarClock,
  REVOKED: Ban,
  FIRING: Flame,
  SUPPRESSED: BellRing,
  ARCHIVED: Archive,
  PUBLIC: Globe,
  PRIVATE: Lock,
  ALERT: BellRing,
  DATASOURCE: Database,
  WEBHOOK: Webhook,
  SMTP: Mail,
  EMAIL: Mail,
  SMS: MessageSquareText,
  SMS_ALICLOUD: MessageSquareText,
  SUPER_ADMIN: ShieldCheck,
  SYSTEM_ADMIN: ShieldCheck,
  SYSTEM_READONLY: ShieldCheck,
  USER: UserRound,
  UNKNOWN: CircleDashed,
}

function iconFor(value: string): LucideIcon | null {
  if (icons[value]) {
    return icons[value]
  }
  if (value.startsWith("METRICS")) {
    return ChartNoAxesCombined
  }
  if (value.startsWith("LOGS")) {
    return ScrollText
  }
  if (value.startsWith("TRACE")) {
    return Waypoints
  }
  if (value.startsWith("PROBE")) {
    return Radar
  }
  return null
}

interface EnumIconProps {
  value: string
  className?: string
}

/** 枚举值前的图标；没有对应图标时返回 null，调用方仍显示枚举文案。 */
export function EnumIcon({ value, className }: EnumIconProps) {
  const icon = iconFor(value)
  if (!icon) {
    return null
  }
  // 用 createElement 而不是把 iconFor 的返回值当组件名渲染，
  // 避免 React Compiler 把这里判定为「渲染期间创建组件」。
  return createElement(icon, {
    "aria-hidden": "true",
    className: cn("size-4 shrink-0 text-muted-foreground", className),
  })
}
