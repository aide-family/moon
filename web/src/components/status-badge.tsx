import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/class-names"
import { enumLabel } from "@/lib/enum-label"

export type BadgeTone =
  | "success"
  | "warning"
  | "danger"
  | "info"
  | "violet"
  | "primary"
  | "neutral"

// 所有状态/类型标签共用一套底色，避免每个页面各写一种绿、黄、红。
const toneClasses: Record<BadgeTone, string> = {
  success:
    "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
  warning:
    "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300",
  danger:
    "border-red-500/30 bg-red-500/10 text-red-700 dark:text-red-300",
  info: "border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300",
  violet:
    "border-violet-500/30 bg-violet-500/10 text-violet-700 dark:text-violet-300",
  primary: "border-primary/30 bg-primary/10 text-primary",
  neutral: "border-border bg-muted/60 text-muted-foreground",
}

const successValues = new Set([
  "ACTIVE",
  "OK",
  "DONE",
  "SENT",
  "APPROVED",
  "ACCEPTED",
  "JOINED",
  "RECOVERED",
  "ONLINE",
])

const warningValues = new Set([
  "PENDING",
  "SENDING",
  "INVITED",
  "BUSY",
  "WORKING",
  "SUPPRESSED",
])

const dangerValues = new Set([
  "FAILED",
  "REJECTED",
  "DECLINED",
  "BANNED",
  "EXPIRED",
  "REVOKED",
  "FIRING",
])

const infoValues = new Set([
  "RECOVERED_BY_MANUAL",
  "CANCELLED",
  "OFFLINE",
  "IDLE",
])

const typeTones: Record<string, BadgeTone> = {
  METRICS: "info",
  METRICS_PROMETHEUS: "info",
  METRICS_VICTORIA_METRICS: "info",
  LOGS: "violet",
  LOGS_ELASTICSEARCH: "violet",
  LOGS_LOKI: "violet",
  LOGS_VLOGS: "violet",
  TRACE: "primary",
  TRACE_JAEGER: "primary",
  EMAIL: "info",
  SMTP: "info",
  SMS: "warning",
  SMS_ALICLOUD: "warning",
  WEBHOOK: "violet",
  DINGTALK: "violet",
  WECHAT: "violet",
  FEISHU: "violet",
  OTHER: "neutral",
  SUPER_ADMIN: "primary",
  SYSTEM_ADMIN: "info",
  USER: "neutral",
  ALERT: "warning",
  DATASOURCE: "info",
  PUBLIC: "success",
  PRIVATE: "neutral",
}

/** statusTone 把后端状态/类型枚举映射成一致的标签颜色。 */
export function statusTone(value: string): BadgeTone {
  if (typeTones[value]) {
    return typeTones[value]
  }
  if (successValues.has(value)) {
    return "success"
  }
  if (warningValues.has(value)) {
    return "warning"
  }
  if (dangerValues.has(value)) {
    return "danger"
  }
  if (infoValues.has(value)) {
    return "info"
  }
  return "neutral"
}

interface StatusBadgeProps {
  value: string
  // label 用于已有翻译的布尔状态（在线/离线、公开/私有）。
  label?: string
  tone?: BadgeTone
  className?: string
}

/**
 * 状态、类型、阶段等枚举字段统一渲染成 Badge。
 * 空值显示占位符，不隐藏列，避免表格列在行之间跳动。
 */
export function StatusBadge({ value, label, tone, className }: StatusBadgeProps) {
  const { t } = useTranslation()
  const text = label ?? enumLabel(t, value)
  return (
    <Badge
      variant="outline"
      className={cn("font-normal", toneClasses[tone ?? statusTone(value)], className)}
    >
      {text || "—"}
    </Badge>
  )
}

interface LevelBadgeProps {
  name: string
  // 告警字典里的展示色，支持 #rrggbb 与 CSS 颜色。
  color?: string
  className?: string
}

/** 告警级别沿用字典颜色，同时保持和状态标签一致的外形。 */
export function LevelBadge({ name, color, className }: LevelBadgeProps) {
  const style = color
    ? {
        borderColor: color,
        color,
        backgroundColor: `color-mix(in srgb, ${color} 12%, transparent)`,
      }
    : undefined
  return (
    <Badge
      variant="outline"
      style={style}
      className={cn("font-normal", !style && toneClasses.primary, className)}
    >
      {name || "—"}
    </Badge>
  )
}
