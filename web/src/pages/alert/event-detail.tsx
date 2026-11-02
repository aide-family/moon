import { RefreshCw } from "lucide-react"
import { type ReactNode, useEffect, useState } from "react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"
import { toast } from "sonner"

import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import {
  ackAlertEvent,
  listAlertEventActions,
  unackAlertEvent,
} from "@/api/alert-event"
import type {
  AlertEventActionItem,
  AlertEventItem,
} from "@/api/alert-event/types"
import { ExprBlock } from "@/components/expr-block"
import { LabelTags } from "@/components/label-tags"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { cn } from "@/lib/class-names"
import { formErrorMessage } from "@/lib/form-error"
import { AlertEventChartBody } from "@/pages/alert/event-chart"
import { formatDuration, formatTime } from "@/pages/alert/format"
import {
  conditionSummary,
  useEventLevelCondition,
} from "@/pages/alert/level-condition"

// 时间线动作 → i18n key。显式映射，避免运行时拼接 key。
const ACTION_KEYS: Record<string, string> = {
  FIRING: "maintenanceWindow.action.FIRING",
  NOTIFIED: "maintenanceWindow.action.NOTIFIED",
  ACKED: "maintenanceWindow.action.ACKED",
  UNACKED: "maintenanceWindow.action.UNACKED",
  INTERVENED: "maintenanceWindow.action.INTERVENED",
  SUPPRESSED: "maintenanceWindow.action.SUPPRESSED",
  MAINTENANCE: "maintenanceWindow.action.MAINTENANCE",
  ESCALATED: "maintenanceWindow.action.ESCALATED",
  RECOVERED: "maintenanceWindow.action.RECOVERED",
  RECOVERED_BY_MANUAL: "maintenanceWindow.action.RECOVERED_BY_MANUAL",
}

function actionLabel(t: TFunction, action: string): string {
  const key = ACTION_KEYS[action]
  return key ? t(key) : action
}

// 升级时间线的 detail 由后端写成英文原文（policy X step N notified M channel(s)），
// 这里转成本地化说明；其他动作与无法识别的 detail 原样展示。
const ESCALATION_DETAIL = /^policy (.+?) step (\d+) notified (\d+) channel\(s\)$/

function actionDetail(t: TFunction, action: string, detail: string): string {
  if (action !== "ESCALATED") {
    return detail
  }
  const match = ESCALATION_DETAIL.exec(detail.trim())
  if (!match) {
    return detail
  }
  return t("escalation.timelineDetail", {
    policy: match[1],
    step: Number(match[2]) + 1,
    channels: Number(match[3]),
  })
}

interface AlertEventDetailProps {
  event: AlertEventItem | null
  // 级别颜色来自告警字典，缺失时用主色。
  level?: AlertDictionaryItem | null
  // 当前时间由列表页统一提供，详情里不再读 Date.now，避免渲染不纯。
  now: number
  open: boolean
  // 只有实时告警且持有干预权限时才显示认领按钮；历史告警只能看时间线。
  canIntervene?: boolean
  // 认领 / 取消认领成功后回传最新事件，父组件据此刷新列表和详情。
  onEventUpdated?: (event: AlertEventItem) => void
  onOpenChange: (open: boolean) => void
}

interface EventDetailField {
  label: string
  value: ReactNode
  full?: boolean
}

/**
 * 实时告警详情抽屉。40vw 宽，字段两列；标签、表达式和描述等长内容独占整行，
 * 只在抽屉内纵向滚动，不允许横向滚动。表达式下方跟命中的级别条件，底部展示事件图表。
 */
export function AlertEventDetail({
  event,
  level,
  now,
  open,
  canIntervene = false,
  onEventUpdated,
  onOpenChange,
}: AlertEventDetailProps) {
  const { t } = useTranslation()
  const ruleUid = event?.ruleUid ?? ""
  const levelUid = event?.levelUid ?? ""
  const levelCondition = useEventLevelCondition(ruleUid, levelUid, open)
  const title = event?.summary || event?.fingerprint || t("alertEvent.title")
  const fields = event ? detailFields(event, level, now, t) : []
  const eventUid = event?.uid ?? ""
  // 时间线按事件 uid 记录一次加载结果：切换事件时先展示空态，避免上一条事件的记录串台。
  const [timeline, setTimeline] = useState<{
    uid: string
    items: AlertEventActionItem[]
  } | null>(null)
  const [timelineAttempt, setTimelineAttempt] = useState(0)
  const [timelineRefreshing, setTimelineRefreshing] = useState(false)
  const [claimPending, setClaimPending] = useState(false)
  const actions = timeline?.uid === eventUid ? timeline.items : []
  const actionsLoading = eventUid !== "" && (!timeline || timeline.uid !== eventUid)

  // 时间线单独加载：失败只提示不影响详情其余字段。
  useEffect(() => {
    if (!open || !eventUid) {
      return
    }
    let alive = true
    listAlertEventActions(eventUid)
      .then((result) => {
        if (alive) {
          setTimeline({ uid: eventUid, items: result.items ?? [] })
          setTimelineRefreshing(false)
        }
      })
      .catch(() => {
        if (alive) {
          setTimeline({ uid: eventUid, items: [] })
          setTimelineRefreshing(false)
          toast.error(t("maintenanceWindow.timelineLoadFailed"))
        }
      })
    return () => {
      alive = false
    }
  }, [open, eventUid, timelineAttempt, t])

  // 认领与取消认领只改事件属性；成功后刷新详情、列表和时间线。
  async function toggleClaim() {
    if (!event || claimPending) {
      return
    }
    const releasing = Boolean(event.ackedBy)
    setClaimPending(true)
    try {
      const result = releasing
        ? await unackAlertEvent(event.uid)
        : await ackAlertEvent(event.uid)
      if (result.event) {
        onEventUpdated?.(result.event)
        setTimelineAttempt((value) => value + 1)
      }
    } catch (err) {
      toast.error(
        formErrorMessage(
          err,
          t(
            releasing
              ? "maintenanceWindow.releaseClaimFailed"
              : "maintenanceWindow.claimFailed"
          )
        )
      )
    } finally {
      setClaimPending(false)
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="min-w-0 overflow-hidden data-[side=right]:sm:w-[40vw]! data-[side=right]:sm:max-w-[40vw]!">
        <SheetHeader className="shrink-0 gap-1 border-b border-border/60 pr-10">
          <SheetTitle className="break-words">{title}</SheetTitle>
          <SheetDescription>
            {event ? <StatusBadge value={event.status} /> : ""}
          </SheetDescription>
          {event ? (
            <div className="flex flex-wrap items-center gap-2 pt-1">
              <StatusBadge
                value={event.ackedBy ? "ACKED" : "UNACKED"}
                tone={event.ackedBy ? "info" : "neutral"}
                label={
                  event.ackedBy
                    ? t("maintenanceWindow.claimed")
                    : t("maintenanceWindow.unclaimed")
                }
              />
              {event.ackedBy ? (
                <span className="min-w-0 text-xs break-words text-muted-foreground">
                  {event.ackedName || event.ackedBy}
                  {event.ackedAt ? ` · ${formatTime(event.ackedAt)}` : ""}
                </span>
              ) : null}
              {canIntervene ? (
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  className="ml-auto shrink-0"
                  disabled={claimPending}
                  onClick={() => void toggleClaim()}
                >
                  {event.ackedBy
                    ? t("maintenanceWindow.releaseClaim")
                    : t("maintenanceWindow.claim")}
                </Button>
              ) : null}
            </div>
          ) : null}
        </SheetHeader>
        {event ? (
          <div className="flex min-h-0 w-full min-w-0 flex-1 flex-col gap-6 overflow-x-hidden overflow-y-auto px-4 pb-6">
            <dl className="grid min-w-0 gap-x-6 gap-y-4 sm:grid-cols-2">
              {fields.map((field) => (
                <div
                  key={field.label}
                  className={
                    field.full
                      ? "grid min-w-0 gap-1 sm:col-span-full"
                      : "grid min-w-0 gap-1"
                  }
                >
                  <dt className="text-sm text-muted-foreground">
                    {field.label}
                  </dt>
                  <dd className="min-w-0 break-words">{field.value || "—"}</dd>
                </div>
              ))}
            </dl>
            <section className="grid min-w-0 gap-2">
              <h3 className="text-sm text-muted-foreground">
                {t("alertEvent.labels")}
              </h3>
              <LabelTags labels={event.labels} />
            </section>
            <section className="grid min-w-0 gap-2">
              <h3 className="text-sm text-muted-foreground">
                {t("alertEvent.expr")}
              </h3>
              {/* 表达式和级别条件放在同一个块里，先看查询语句，再看是哪一档条件命中。 */}
              <ExprBlock
                expr={event.expr}
                footer={
                  levelCondition ? (
                    <p className="min-w-0 border-t border-border/60 pt-2 font-mono text-xs break-words whitespace-pre-wrap text-muted-foreground">
                      <span className="mr-2 font-sans">
                        {t("alertEvent.levelCondition")}
                      </span>
                      {conditionSummary(t, levelCondition, event.levelName)}
                    </p>
                  ) : null
                }
              />
            </section>
            {event.description ? (
              <section className="grid min-w-0 gap-2">
                <h3 className="text-sm text-muted-foreground">
                  {t("alertEvent.descriptionField")}
                </h3>
                <p className="text-sm break-words whitespace-pre-wrap">
                  {event.description}
                </p>
              </section>
            ) : null}
            <section className="grid min-w-0 gap-2">
              <h3 className="text-sm text-muted-foreground">
                {t("alertEvent.eventChart")}
              </h3>
              <AlertEventChartBody
                key={event.uid}
                event={event}
                compact
                levelCondition={levelCondition}
              />
            </section>
            <section className="grid min-w-0 gap-2">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h3 className="text-sm text-muted-foreground">
                  {t("maintenanceWindow.timeline")}
                </h3>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  disabled={actionsLoading || timelineRefreshing}
                  onClick={() => {
                    setTimelineRefreshing(true)
                    setTimelineAttempt((value) => value + 1)
                  }}
                >
                  <RefreshCw
                    aria-hidden
                    data-icon="inline-start"
                    className={cn(
                      (actionsLoading || timelineRefreshing) && "animate-spin"
                    )}
                  />
                  {t("maintenanceWindow.timelineRefresh")}
                </Button>
              </div>
              {actionsLoading && actions.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  {t("state.loading")}
                </p>
              ) : actions.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  {t("maintenanceWindow.timelineEmpty")}
                </p>
              ) : (
                <ol className="grid min-w-0 gap-3">
                  {actions.map((action) => (
                    <li
                      key={action.uid}
                      className="grid min-w-0 gap-0.5 border-l-2 border-border/70 pl-3"
                    >
                      <div className="flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5">
                        <span className="text-sm font-medium">
                          {actionLabel(t, action.action)}
                        </span>
                        <span className="text-xs break-words text-muted-foreground">
                          {action.actorName ||
                            action.actorUid ||
                            t("maintenanceWindow.timelineSystemActor")}
                        </span>
                        <span className="ml-auto font-mono text-xs text-muted-foreground tabular-nums">
                          {formatTime(action.createdAt)}
                        </span>
                      </div>
                      {action.detail ? (
                        <p className="text-xs break-words whitespace-pre-wrap text-muted-foreground">
                          {actionDetail(t, action.action, action.detail)}
                        </p>
                      ) : null}
                    </li>
                  ))}
                </ol>
              )}
            </section>
          </div>
        ) : null}
      </SheetContent>
    </Sheet>
  )
}

function detailFields(
  event: AlertEventItem,
  level: AlertDictionaryItem | null | undefined,
  now: number,
  t: TFunction
): EventDetailField[] {
  const endMs = event.endsAt ? new Date(event.endsAt).getTime() : now
  return [
    {
      label: t("alertEvent.rule"),
      value: event.ruleUid
        ? linkTo(ruleLink(event), event.ruleName || event.ruleUid)
        : event.ruleName || "—",
    },
    {
      label: t("alertEvent.level"),
      value: event.levelUid ? levelValue(event, level) : event.levelName || "—",
    },
    {
      label: t("alertEvent.datasource"),
      value: event.datasourceUid
        ? linkTo(
            `/datasources?uid=${encodeURIComponent(event.datasourceUid)}`,
            event.datasourceName || event.datasourceUid
          )
        : event.datasourceName || "—",
    },
    { label: t("alertEvent.status"), value: <StatusBadge value={event.status} /> },
    { label: t("alertEvent.sample"), value: event.sampleValue },
    {
      label: t("alertEvent.duration"),
      value: formatDuration(event.startsAt, endMs),
    },
    { label: t("alertEvent.startsAt"), value: formatTime(event.startsAt) },
    { label: t("alertEvent.endsAt"), value: formatTime(event.endsAt) },
    {
      label: t("alertEvent.suppressed"),
      value: event.suppressed ? t("alertEvent.yes") : t("alertEvent.no"),
    },
    {
      label: t("alertEvent.suppressedUntil"),
      // 永久抑制没有到期时间，显示「永久」而不是空占位。
      value:
        event.suppressed && !event.suppressedUntil
          ? t("alertEvent.suppressPermanent")
          : formatTime(event.suppressedUntil),
    },
    {
      label: t("alertEvent.intervenedBy"),
      value: event.intervenedName || event.intervenedBy,
    },
    {
      label: t("alertEvent.intervenedAt"),
      value: formatTime(event.intervenedAt),
    },
    {
      label: t("maintenanceWindow.claimedBy"),
      value: event.ackedName || event.ackedBy,
    },
    {
      label: t("maintenanceWindow.claimedAt"),
      value: formatTime(event.ackedAt),
    },
    { label: t("detail.uid"), value: mono(event.uid), full: true },
    {
      label: t("alertEvent.fingerprint"),
      value: mono(event.fingerprint),
      full: true,
    },
  ]
}

// 标签、章节标题和值都由页面传入，这里只放纯展示逻辑，避免在字段数组里重复 t()。
function linkTo(to: string, label: string) {
  return (
    <Link
      to={to}
      className="break-all text-primary-ink underline-offset-4 hover:underline"
    >
      {label}
    </Link>
  )
}

function ruleLink(event: AlertEventItem) {
  const params = new URLSearchParams()
  if (event.ruleGroupUid) {
    params.set("groupUid", event.ruleGroupUid)
  }
  params.set("ruleUid", event.ruleUid)
  return `/alert-rules?${params.toString()}`
}

function levelValue(
  event: AlertEventItem,
  level: AlertDictionaryItem | null | undefined
) {
  return (
    <span className="inline-flex min-w-0 items-center gap-2">
      <span
        aria-hidden
        className="size-2 shrink-0 rounded-full"
        style={{ backgroundColor: level?.color || "#1677ff" }}
      />
      <Link
        to={`/alert-dictionaries?type=ALERT&uid=${encodeURIComponent(event.levelUid)}`}
        className="min-w-0 break-all text-primary-ink underline-offset-4 hover:underline"
      >
        {event.levelName || event.levelUid}
      </Link>
    </span>
  )
}

function mono(value: string) {
  if (!value) {
    return ""
  }
  return <span className="font-mono text-xs break-all">{value}</span>
}
