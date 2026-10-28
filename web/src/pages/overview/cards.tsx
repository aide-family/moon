import {
  ArrowUpRight,
  BellRing,
  CalendarDays,
  CircleAlert,
  Database,
  MailCheck,
  Server,
  ShieldCheck,
  ShieldHalf,
} from "lucide-react"
import type { ReactNode } from "react"
import type { LucideIcon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"

import type { AlertEventStats } from "@/api/alert-event/types"
import type {
  DashboardAlertSnapshot,
  DashboardDeliverySnapshot,
  DashboardInventory,
} from "@/api/dashboard"
import type { ListWorkerInstancesResult } from "@/api/worker-instances/types"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/class-names"

import { elapsed, formatClock, formatCount, formatPercent, share } from "./lib"

const toneChips = {
  danger: "bg-red-500/10 text-red-600 dark:text-red-400",
  warning: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
  primary: "bg-primary/10 text-primary-ink",
  neutral: "bg-muted text-muted-foreground",
} as const

type Tone = keyof typeof toneChips

/** 卡片头：标题 + 一行口径说明 + 右侧动作。所有卡片共用，避免各写一套间距。 */
function CardHead({
  title,
  description,
  badge,
  actions,
}: {
  title: string
  description?: string
  badge?: ReactNode
  actions?: ReactNode
}) {
  return (
    <div className="flex items-start justify-between gap-3 px-4">
      <div className="min-w-0 space-y-0.5">
        <h2 className="flex items-center gap-2 font-heading text-sm font-medium">
          {title}
          {badge}
        </h2>
        {description ? (
          <p className="text-xs text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {actions ? <div className="shrink-0">{actions}</div> : null}
    </div>
  )
}

function MetricCard({
  label,
  value,
  hint,
  icon,
  tone = "neutral",
  badge,
  loading,
}: {
  label: string
  value: string
  hint: string
  icon: LucideIcon
  tone?: Tone
  badge?: string
  loading: boolean
}) {
  const Icon = icon
  return (
    <Card className="min-w-0">
      <div className="flex items-start justify-between gap-3 px-4">
        <div className="min-w-0 space-y-1">
          <p className="flex items-center gap-2 text-[11px] font-medium tracking-[0.08em] text-muted-foreground uppercase">
            {label}
            {badge ? (
              <span className="rounded border border-border/70 bg-muted/60 px-1 py-px text-[10px] font-normal tracking-normal normal-case">
                {badge}
              </span>
            ) : null}
          </p>
          {loading ? (
            <Skeleton className="h-9 w-20" />
          ) : (
            <p className="font-heading text-3xl leading-tight font-semibold tracking-tight tabular-nums">
              {value}
            </p>
          )}
          <p className="text-xs leading-relaxed text-muted-foreground">{hint}</p>
        </div>
        <span
          className={cn(
            "grid size-9 shrink-0 place-items-center rounded-lg",
            toneChips[tone]
          )}
          aria-hidden="true"
        >
          <Icon className="size-4" />
        </span>
      </div>
    </Card>
  )
}

/**
 * KPI 条：四个数字全部来自已有接口的精确计数。
 * 「累计已恢复」必须显式标注累计，因为它包含归档历史、不是窗口指标。
 */
export function KpiStrip({
  alerts,
  delivery,
  today,
  showAlerts,
  showDelivery,
  loading,
}: {
  alerts: DashboardAlertSnapshot | null
  delivery: DashboardDeliverySnapshot | null
  /** 区间聚合里「今天」那一桶；取不到时为 null，卡片显示占位符。 */
  today: { events: number; closed: number } | null
  showAlerts: boolean
  showDelivery: boolean
  loading: boolean
}) {
  const { t } = useTranslation()
  const stats = alerts?.stats ?? null
  const rate = delivery?.successRate ?? null
  const lowSuccess = rate !== null && rate < 0.95

  return (
    <div className="grid min-w-0 grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {showAlerts ? (
        <>
          <MetricCard
            label={t("overview.kpi.firing")}
            value={stats ? formatCount(stats.firing) : "—"}
            hint={t("overview.kpi.firingHint")}
            icon={BellRing}
            tone={stats && stats.firing > 0 ? "danger" : "primary"}
            loading={loading}
          />
          <MetricCard
            label={t("overview.kpi.suppressed")}
            value={stats ? formatCount(stats.suppressed) : "—"}
            hint={t("overview.kpi.suppressedHint")}
            icon={ShieldHalf}
            tone={stats && stats.suppressed > 0 ? "warning" : "neutral"}
            loading={loading}
          />
          <MetricCard
            label={t("overview.kpi.newToday")}
            value={today ? formatCount(today.events) : "—"}
            hint={
              today
                ? t("overview.kpi.newTodayHint", {
                    n: formatCount(today.closed),
                  })
                : t("overview.kpi.unavailable")
            }
            icon={CalendarDays}
            tone="primary"
            loading={loading}
          />
        </>
      ) : null}
      {showDelivery ? (
        <MetricCard
          label={t("overview.kpi.successRate")}
          value={delivery ? formatPercent(delivery.successRate) : "—"}
          hint={
            delivery
              ? delivery.attempts === 0
                ? t("overview.kpi.noAttempts")
                : t("overview.kpi.successRateHint", {
                    sent: formatCount(delivery.sent),
                    failed: formatCount(delivery.failed),
                    attempts: formatCount(delivery.attempts),
                  })
              : t("overview.kpi.unavailable")
          }
          icon={MailCheck}
          tone={lowSuccess ? "warning" : "primary"}
          loading={loading}
        />
      ) : null}
    </div>
  )
}

function EmptyBlock({
  icon: Icon,
  title,
  hint,
}: {
  icon: LucideIcon
  title: string
  hint: string
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-1 px-4 py-10 text-center">
      <Icon className="size-5 text-muted-foreground/70" aria-hidden="true" />
      <p className="mt-1 text-sm font-medium">{title}</p>
      <p className="max-w-sm text-xs leading-relaxed text-muted-foreground">
        {hint}
      </p>
    </div>
  )
}

/**
 * 触发中清单。只能叫「最近触发」：实时列表按创建时间倒序，
 * 服务端没有严重度优先排序，在截断的几条上再按级别排会制造错误口径。
 */
export function AlertQueueCard({
  snapshot,
  loading,
  now,
}: {
  snapshot: DashboardAlertSnapshot | null
  loading: boolean
  now: number
}) {
  const { t } = useTranslation()
  const items = snapshot?.items ?? []
  const total = snapshot?.total ?? 0

  return (
    <Card className="flex min-h-0 min-w-0 flex-col">
      <CardHead
        title={t("overview.queue.title")}
        description={t("overview.queue.description", { n: items.length })}
        badge={
          total > 0 ? (
            <span className="rounded-full bg-red-500/10 px-2 py-0.5 text-[11px] font-medium text-red-600 tabular-nums dark:text-red-400">
              {formatCount(total)}
            </span>
          ) : null
        }
        actions={
          <Button variant="ghost" size="sm" asChild>
            <Link to="/alerts">
              {t("overview.queue.viewAll")}
              <ArrowUpRight aria-hidden="true" />
            </Link>
          </Button>
        }
      />
      <div className="mt-1 min-h-0">
        {loading && !snapshot ? (
          <div className="space-y-2 px-4 py-3">
            {[0, 1, 2, 3, 4].map((index) => (
              <Skeleton key={index} className="h-9 w-full" />
            ))}
          </div>
        ) : items.length === 0 ? (
          <EmptyBlock
            icon={ShieldCheck}
            title={t("overview.queue.emptyTitle")}
            hint={t("overview.queue.emptyHint")}
          />
        ) : (
          <ul className="divide-y divide-border/60 border-t border-border/60">
            {items.map((item) => (
              <li key={item.uid}>
                <Link
                  to="/alerts"
                  className="flex items-start gap-3 px-4 py-2.5 transition-colors hover:bg-muted/50 focus-visible:bg-muted/50 focus-visible:outline-none"
                >
                  <StatusBadge
                    value={item.levelName || "UNKNOWN"}
                    label={item.levelName || t("overview.queue.noLevel")}
                    className="mt-0.5 shrink-0"
                  />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">
                      {item.summary || item.fingerprint || "—"}
                    </p>
                    <p className="mt-0.5 flex flex-wrap items-center gap-x-1.5 text-xs text-muted-foreground">
                      <span className="max-w-[12rem] truncate">
                        {item.ruleName || "—"}
                      </span>
                      <span aria-hidden="true">·</span>
                      <span className="max-w-[12rem] truncate">
                        {item.datasourceName || "—"}
                      </span>
                    </p>
                  </div>
                  <div className="shrink-0 text-right">
                    <p className="font-mono text-xs tabular-nums">
                      {elapsed(item.startsAt, now, t)}
                    </p>
                    <p className="mt-0.5 text-[11px] text-muted-foreground tabular-nums">
                      {formatClock(item.startsAt)}
                    </p>
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </Card>
  )
}

/** 存量构成：当前触发中的实时快照，明确不进时间序列（抑制中会随时间变小）。 */
export function CompositionCard({
  stats,
  closedTotal,
  days,
  loading,
}: {
  stats: AlertEventStats | null
  /** 区间内已恢复事件数，来自概览区间聚合；没加载成功时为 null。 */
  closedTotal: number | null
  days: number
  loading: boolean
}) {
  const { t } = useTranslation()
  const firing = stats?.firing ?? 0
  const suppressed = stats?.suppressed ?? 0
  const total = firing + suppressed

  return (
    <Card className="min-w-0">
      <CardHead
        title={t("overview.composition.title")}
        description={t("overview.composition.description")}
      />
      <div className="space-y-3 px-4">
        {loading && !stats ? (
          <Skeleton className="h-2.5 w-full rounded-full" />
        ) : (
          <div
            className="flex h-2.5 w-full overflow-hidden rounded-full bg-muted"
            role="img"
            aria-label={t("overview.composition.aria", {
              firing: formatCount(firing),
              suppressed: formatCount(suppressed),
            })}
          >
            <span
              className="h-full bg-red-500 transition-[width] duration-300"
              style={{ width: `${share(firing, total)}%` }}
            />
            <span
              className="h-full bg-amber-500 transition-[width] duration-300"
              style={{ width: `${share(suppressed, total)}%` }}
            />
          </div>
        )}
        <dl className="grid grid-cols-2 gap-2">
          <div className="flex items-center gap-2">
            <span className="size-2 shrink-0 rounded-full bg-red-500" aria-hidden="true" />
            <dt className="text-xs text-muted-foreground">
              {t("overview.composition.firing")}
            </dt>
            <dd className="ml-auto font-heading text-sm font-semibold tabular-nums">
              {formatCount(firing)}
            </dd>
          </div>
          <div className="flex items-center gap-2">
            <span className="size-2 shrink-0 rounded-full bg-amber-500" aria-hidden="true" />
            <dt className="text-xs text-muted-foreground">
              {t("overview.composition.suppressed")}
            </dt>
            <dd className="ml-auto font-heading text-sm font-semibold tabular-nums">
              {formatCount(suppressed)}
            </dd>
          </div>
        </dl>
        <p className="text-xs text-muted-foreground">
          {t("overview.composition.note")}
        </p>
        <p className="border-t border-border/60 pt-2 text-xs text-muted-foreground tabular-nums">
          {closedTotal === null
            ? t("overview.composition.closedUnknown")
            : t("overview.composition.closed", {
                days,
                n: formatCount(closedTotal),
              })}
        </p>
      </div>
    </Card>
  )
}

/** 今日投递：成功率必须连带分母一起展示，避免把「没投递」读成「全失败」。 */
export function DeliveryCard({
  delivery,
  loading,
}: {
  delivery: DashboardDeliverySnapshot | null
  loading: boolean
}) {
  const { t } = useTranslation()
  const sent = delivery?.sent ?? 0
  const failed = delivery?.failed ?? 0
  const attempts = delivery?.attempts ?? 0

  return (
    <Card className="min-w-0">
      <CardHead
        title={t("overview.delivery.title")}
        description={t("overview.delivery.description")}
      />
      <div className="space-y-3 px-4">
        {loading && !delivery ? (
          <Skeleton className="h-8 w-24" />
        ) : (
          <div className="flex items-end justify-between gap-3">
            <p className="font-heading text-2xl font-semibold tabular-nums">
              {delivery ? formatPercent(delivery.successRate) : "—"}
            </p>
            <p className="pb-1 text-xs text-muted-foreground tabular-nums">
              {attempts === 0
                ? t("overview.kpi.noAttempts")
                : t("overview.delivery.attempts", {
                    attempts: formatCount(attempts),
                  })}
            </p>
          </div>
        )}
        <div
          className="flex h-2.5 w-full overflow-hidden rounded-full bg-muted"
          role="img"
          aria-label={t("overview.delivery.aria", {
            sent: formatCount(sent),
            failed: formatCount(failed),
          })}
        >
          <span
            className="h-full bg-primary"
            style={{ width: `${share(sent, attempts)}%` }}
          />
          <span
            className="h-full bg-red-500"
            style={{ width: `${share(failed, attempts)}%` }}
          />
        </div>
        <dl className="flex items-center gap-4 text-xs">
          <div className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-primary" aria-hidden="true" />
            <dt className="text-muted-foreground">{t("overview.delivery.sent")}</dt>
            <dd className="font-semibold tabular-nums">{formatCount(sent)}</dd>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-red-500" aria-hidden="true" />
            <dt className="text-muted-foreground">{t("overview.delivery.failed")}</dt>
            <dd className="font-semibold tabular-nums">{formatCount(failed)}</dd>
          </div>
        </dl>
        <p className="text-xs text-muted-foreground">{t("overview.delivery.note")}</p>
      </div>
    </Card>
  )
}

/** 平台健康：后台进程是唯一能证伪「0 告警 = 健康」的信号，必须常驻首屏。 */
export function PlatformCard({
  workers,
  inventory,
  loading,
}: {
  workers: ListWorkerInstancesResult | null
  inventory: DashboardInventory | null
  loading: boolean
}) {
  const { t } = useTranslation()

  return (
    <Card className="min-w-0">
      <CardHead
        title={t("overview.platform.title")}
        description={t("overview.platform.description")}
      />
      <div className="space-y-3 px-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-2">
            <span
              className={cn(
                "size-2 shrink-0 rounded-full",
                workers && workers.onlineCount > 0
                  ? "bg-emerald-500"
                  : "bg-red-500"
              )}
              aria-hidden="true"
            />
            <span className="text-xs text-muted-foreground">
              {t("overview.platform.workers")}
            </span>
          </div>
          {loading && !workers ? (
            <Skeleton className="h-4 w-32" />
          ) : (
            <p className="text-right text-xs tabular-nums">
              {workers
                ? t("overview.platform.workerSummary", {
                    online: formatCount(workers.onlineCount),
                    busy: formatCount(workers.busyCount),
                    idle: formatCount(workers.idleCount),
                  })
                : t("overview.platform.noWorker")}
            </p>
          )}
        </div>
        <div className="grid grid-cols-2 gap-2 border-t border-border/60 pt-3">
          <div className="flex items-center gap-2">
            <Database className="size-3.5 shrink-0 text-muted-foreground" aria-hidden="true" />
            <span className="text-xs text-muted-foreground">
              {t("overview.platform.datasources")}
            </span>
            <span className="ml-auto font-heading text-sm font-semibold tabular-nums">
              {inventory?.datasources === null || inventory?.datasources === undefined
                ? "—"
                : formatCount(inventory.datasources)}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <CircleAlert className="size-3.5 shrink-0 text-muted-foreground" aria-hidden="true" />
            <span className="text-xs text-muted-foreground">
              {t("overview.platform.rules")}
            </span>
            <span className="ml-auto font-heading text-sm font-semibold tabular-nums">
              {inventory?.rules === null || inventory?.rules === undefined
                ? "—"
                : formatCount(inventory.rules)}
            </span>
          </div>
        </div>
        <p className="flex items-start gap-1.5 text-xs text-muted-foreground">
          <Server className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
          {t("overview.platform.note")}
        </p>
      </div>
    </Card>
  )
}
