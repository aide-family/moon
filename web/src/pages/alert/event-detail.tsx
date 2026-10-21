import { type ReactNode } from "react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"

import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import type { AlertEventItem } from "@/api/alert-event/types"
import { ExprBlock } from "@/components/expr-block"
import { LabelTags } from "@/components/label-tags"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { enumLabel } from "@/lib/enum-label"
import { EventChartBody } from "@/pages/alert/event-chart"
import { formatDuration, formatTime } from "@/pages/alert/format"
import {
  conditionSummary,
  useEventLevelCondition,
} from "@/pages/alert/level-condition"

interface AlertEventDetailProps {
  event: AlertEventItem | null
  // 级别颜色来自告警字典，缺失时用主色。
  level?: AlertDictionaryItem | null
  // 当前时间由列表页统一提供，详情里不再读 Date.now，避免渲染不纯。
  now: number
  open: boolean
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
  onOpenChange,
}: AlertEventDetailProps) {
  const { t } = useTranslation()
  const ruleUid = event?.ruleUid ?? ""
  const levelUid = event?.levelUid ?? ""
  const levelCondition = useEventLevelCondition(ruleUid, levelUid, open)
  const title = event?.summary || event?.fingerprint || t("alertEvent.title")
  const fields = event ? detailFields(event, level, now, t) : []

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="min-w-0 overflow-hidden data-[side=right]:sm:w-[40vw]! data-[side=right]:sm:max-w-[40vw]!">
        <SheetHeader className="shrink-0 gap-1 border-b border-border/60 pr-10">
          <SheetTitle className="break-words">{title}</SheetTitle>
          <SheetDescription>
            {event ? enumLabel(t, event.status) : ""}
          </SheetDescription>
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
              <EventChartBody
                key={event.uid}
                event={event}
                compact
                levelCondition={levelCondition}
              />
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
    { label: t("alertEvent.status"), value: enumLabel(t, event.status) },
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
      className="break-all text-primary underline-offset-4 hover:underline"
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
        className="min-w-0 break-all text-primary underline-offset-4 hover:underline"
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
