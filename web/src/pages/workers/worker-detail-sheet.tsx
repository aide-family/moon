import { Activity, CircleAlert, Database, Timer } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { getWorkerInstance } from "@/api/worker-instances"
import type {
  WorkerInstanceDetail,
  WorkerInstanceDetailResult,
  WorkerInstanceItem,
  WorkerTaskItem,
} from "@/api/worker-instances"
import { CopyIconButton } from "@/components/copy-button"
import { DetailSectionList, DetailSheet } from "@/components/detail-sheet"
import type { DetailSection } from "@/components/detail-sheet"
import { DetailEmpty } from "@/components/detail-values"
import { StatusBadge } from "@/components/status-badge"
import type { BadgeTone } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"

// 抽屉打开期间详情每 2 秒刷新一次；关闭后立即停止。
const DETAIL_POLL_MS = 2000
// 倒计时只在本地每秒重算，不跟着网络刷新跑。
const COUNTDOWN_TICK_MS = 1000

// 五个状态各用一种语义色，并保留文字标签，色觉障碍用户也能区分。
const taskStateTones: Record<string, BadgeTone> = {
  RUNNING: "warning",
  NEW: "violet",
  PENDING: "info",
  DONE: "success",
  LEFT: "neutral",
}

interface DetailState {
  uid: string
  result: WorkerInstanceDetailResult | null
  failed: boolean
}

export interface WorkerDetailSheetProps {
  // 列表里选中的实例；抽屉关闭后仍保留一帧，用于关闭动画和标题兜底。
  instance: WorkerInstanceItem | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

/**
 * 天权（MEGREZ）实例详情抽屉：顶部固定两个倒计时，内容区列出本轮抢到的
 * 「规则 × 数据源」组合与跑批幂等键。打开期间每 2 秒刷新详情，失败时保留
 * 上一次成功的数据；关闭即停止轮询。
 */
export function WorkerDetailSheet({
  instance,
  open,
  onOpenChange,
}: WorkerDetailSheetProps) {
  const { t } = useTranslation()
  const instanceUid = instance?.instanceUid ?? ""
  const [state, setState] = useState<DetailState | null>(null)
  const [now, setNow] = useState(() => Date.now())
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    if (!open || !instanceUid) {
      return
    }
    let active = true
    let inFlight = false
    async function load() {
      if (inFlight) {
        return
      }
      inFlight = true
      try {
        const result = await getWorkerInstance("MEGREZ", instanceUid)
        if (!active) {
          return
        }
        setState({ uid: instanceUid, result, failed: false })
        setNow(Date.now())
      } catch {
        if (!active) {
          return
        }
        // 失败保留旧数据，只把本次刷新标记为失败，由抽屉内提示条展示。
        setState((previous) =>
          previous && previous.uid === instanceUid
            ? { ...previous, failed: true }
            : { uid: instanceUid, result: null, failed: true }
        )
      } finally {
        inFlight = false
      }
    }
    void load()
    const timer = window.setInterval(() => {
      void load()
    }, DETAIL_POLL_MS)
    return () => {
      active = false
      window.clearInterval(timer)
    }
  }, [open, instanceUid, attempt])

  // 倒计时每秒本地刷新；抽屉关闭后不保留计时器。
  useEffect(() => {
    if (!open) {
      return
    }
    const timer = window.setInterval(() => setNow(Date.now()), COUNTDOWN_TICK_MS)
    return () => window.clearInterval(timer)
  }, [open])

  const current = state && state.uid === instanceUid ? state : null
  const result = current?.result ?? null
  const failed = current?.failed ?? false
  const detail = result?.item ?? null
  const loading = open && instanceUid !== "" && detail === null && !failed
  const offline = detail ? !detail.online : instance?.online === false
  const lastSeenAt = detail?.lastSeenAt || instance?.lastSeenAt || ""
  const title = detail?.hostname || instance?.hostname || instanceUid || t("worker.detailTitle")
  const retry = () => setAttempt((value) => value + 1)

  return (
    <DetailSheet
      open={open}
      size="40vw"
      contentMode="fill"
      title={title}
      titleCopy={
        instanceUid ? { value: instanceUid, label: t("worker.instanceUid") } : undefined
      }
      description={
        <HeaderBadges
          online={detail?.online ?? instance?.online ?? false}
          busy={detail?.busy ?? instance?.busy ?? false}
          lastSeenAt={lastSeenAt}
        />
      }
      onOpenChange={onOpenChange}
    >
      <div className="flex min-h-0 flex-1 flex-col">
        {/* 倒计时区固定在顶部，只有下面的组合列表滚动。 */}
        <div className="shrink-0 space-y-2.5 border-b border-border/60 bg-muted/20 px-4 py-3 sm:px-5">
          {failed && result ? (
            <div
              role="status"
              className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-destructive/30 bg-destructive/5 px-2.5 py-1.5 text-xs text-destructive"
            >
              <span>{t("worker.detailRefreshFailed")}</span>
              <Button type="button" variant="outline" size="sm" onClick={retry}>
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {result && detail ? (
            <CountdownPanel
              online={!offline}
              now={now}
              lastSeenAt={lastSeenAt}
              nextScheduleAt={result.nextScheduleAt}
              scheduleIntervalSeconds={result.scheduleIntervalSeconds}
              nextPollAt={detail.nextPollAt}
              pollIntervalSeconds={detail.pollIntervalSeconds}
              observedAt={result.observedAt}
            />
          ) : loading ? (
            <p className="text-xs text-muted-foreground">{t("worker.detailLoading")}</p>
          ) : null}
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 py-4 sm:px-5">
          {loading ? (
            <DetailSkeleton />
          ) : detail ? (
            <DetailContent detail={detail} />
          ) : (
            <div
              role="status"
              aria-live="polite"
              className="flex flex-col items-center gap-3 py-10 text-center"
            >
              <CircleAlert className="size-6 text-destructive" aria-hidden="true" />
              <div>
                <p className="font-medium">{t("worker.detailLoadFailed")}</p>
                <p className="mt-1 text-sm text-muted-foreground">
                  {t("worker.detailUnavailableHint")}
                </p>
              </div>
              <Button type="button" variant="outline" size="sm" onClick={retry}>
                {t("action.retry")}
              </Button>
            </div>
          )}
        </div>
      </div>
    </DetailSheet>
  )
}

function HeaderBadges({
  online,
  busy,
  lastSeenAt,
}: {
  online: boolean
  busy: boolean
  lastSeenAt: string
}) {
  const { t } = useTranslation()
  return (
    <>
      <StatusBadge
        value={online ? "ONLINE" : "OFFLINE"}
        label={online ? t("worker.online") : t("worker.offline")}
      />
      {online ? (
        <StatusBadge
          value={busy ? "BUSY" : "IDLE"}
          label={busy ? t("worker.working") : t("worker.idle")}
        />
      ) : null}
      <span className="text-xs">
        {t("worker.lastSeen", { time: formatClock(lastSeenAt) || "—" })}
      </span>
    </>
  )
}

interface CountdownPanelProps {
  online: boolean
  now: number
  lastSeenAt: string
  nextScheduleAt: string
  scheduleIntervalSeconds: number
  nextPollAt: string
  pollIntervalSeconds: number
  observedAt: string
}

function CountdownPanel({
  online,
  now,
  lastSeenAt,
  nextScheduleAt,
  scheduleIntervalSeconds,
  nextPollAt,
  pollIntervalSeconds,
  observedAt,
}: CountdownPanelProps) {
  const { t } = useTranslation()
  // 离线实例的心跳与调度时间不可信，两个倒计时都显示不可用。
  const scheduleSeconds = online
    ? scheduleRemaining(nextScheduleAt, scheduleIntervalSeconds, now)
    : null
  const pollSeconds = online
    ? pollRemaining(nextPollAt, pollIntervalSeconds, now)
    : null
  return (
    <>
      <div className="grid grid-cols-2 gap-2">
        <CountdownTile
          label={t("worker.nextSchedule")}
          seconds={scheduleSeconds}
          intervalSeconds={scheduleIntervalSeconds}
          imminentLabel={t("worker.countdownImminent")}
        />
        <CountdownTile
          label={t("worker.nextPoll")}
          seconds={pollSeconds}
          intervalSeconds={pollIntervalSeconds}
          imminentLabel={t("worker.countdownPolling")}
        />
      </div>
      {online ? (
        <p className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
          <span>{t("worker.detailPollHint")}</span>
          {observedAt ? (
            <span>{t("worker.detailObservedAt", { time: formatClock(observedAt) })}</span>
          ) : null}
        </p>
      ) : (
        <p className="flex items-start gap-2 rounded-md border border-warning-border bg-warning-soft px-2.5 py-1.5 text-xs text-warning">
          <CircleAlert className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
          <span>
            {t("worker.countdownOfflineHint", {
              time: formatClock(lastSeenAt) || "—",
            })}
          </span>
        </p>
      )}
    </>
  )
}

function CountdownTile({
  label,
  seconds,
  intervalSeconds,
  imminentLabel,
}: {
  label: string
  seconds: number | null
  intervalSeconds: number
  imminentLabel: string
}) {
  const { t } = useTranslation()
  return (
    <div className="min-w-0 rounded-lg border border-border/60 bg-background/70 px-3 py-2.5">
      <p className="flex items-center gap-1.5 text-[11px] leading-4 font-medium tracking-wide text-muted-foreground">
        <Timer className="size-3.5 shrink-0" aria-hidden="true" />
        <span className="min-w-0 break-words">{label}</span>
      </p>
      {/* 倒计时每秒变化，不给屏幕阅读器实时播报，避免持续打断。 */}
      <p
        aria-live="off"
        className="mt-1 text-xl leading-none font-semibold tabular-nums"
      >
        {seconds === null ? (
          <span className="text-sm font-normal text-muted-foreground">
            {t("worker.countdownUnavailable")}
          </span>
        ) : seconds === 0 ? (
          <span className="text-sm font-medium">{imminentLabel}</span>
        ) : (
          <>
            {seconds}
            <span className="ml-1 text-xs font-normal text-muted-foreground">
              {t("worker.countdownSeconds")}
            </span>
          </>
        )}
      </p>
      {intervalSeconds > 0 ? (
        <p className="mt-1 text-[11px] leading-4 text-muted-foreground">
          {t("worker.countdownEvery", { seconds: intervalSeconds })}
        </p>
      ) : null}
    </div>
  )
}

function DetailContent({ detail }: { detail: WorkerInstanceDetail }) {
  const { t } = useTranslation()
  const tasks = detail.tasks ?? []
  const online = detail.online
  const busy = detail.busy
  const sections: DetailSection[] = [
    {
      title: t("worker.instanceSection"),
      layout: "inline",
      fields: [
        { label: t("worker.instanceHostname"), value: detail.hostname },
        {
          label: t("worker.instanceUid"),
          value: (
            <span className="font-mono text-xs break-all">{detail.instanceUid}</span>
          ),
        },
        {
          label: t("worker.instanceStatus"),
          value: (
            <span className="inline-flex flex-wrap items-center gap-1.5">
              <StatusBadge
                value={online ? "ONLINE" : "OFFLINE"}
                label={online ? t("worker.online") : t("worker.offline")}
              />
              {online ? (
                <StatusBadge
                  value={busy ? "BUSY" : "IDLE"}
                  label={busy ? t("worker.working") : t("worker.idle")}
                />
              ) : null}
            </span>
          ),
        },
        {
          label: t("worker.instanceLastSeen"),
          value: formatClock(detail.lastSeenAt),
        },
        {
          label: t("worker.instanceTask"),
          span: 2,
          value:
            busy && detail.taskUid
              ? detail.taskName || detail.taskUid
              : t("worker.noActiveTask"),
        },
        {
          label: t("worker.instanceTaskStartedAt"),
          value: formatClock(detail.taskStartedAt),
        },
      ],
    },
    {
      title: t("worker.roundSection"),
      layout: "inline",
      fields: [
        {
          label: t("worker.roundStartedAt"),
          value: formatClock(detail.roundStartedAt),
        },
      ],
    },
  ]
  return (
    <div className="grid gap-5">
      <DetailSectionList sections={sections} />
      <TaskListSection tasks={tasks} />
    </div>
  )
}

function TaskListSection({ tasks }: { tasks: WorkerTaskItem[] }) {
  const { t } = useTranslation()
  return (
    <section className="grid min-w-0 gap-2">
      <div className="flex flex-wrap items-baseline justify-between gap-x-2 gap-y-1 border-b pb-1.5">
        <h3 className="font-heading text-xs font-semibold tracking-wide text-foreground">
          {t("worker.tasksTitle")}
        </h3>
        <span className="text-xs text-muted-foreground">
          {t("worker.tasksCount", { count: tasks.length })}
        </span>
      </div>
      {tasks.length === 0 ? (
        <div className="rounded-lg border border-dashed border-border/70 px-4 py-6 text-center">
          <Activity className="mx-auto size-5 text-muted-foreground/70" aria-hidden="true" />
          <p className="mt-2 text-sm font-medium">{t("worker.tasksEmptyTitle")}</p>
          <p className="mt-1 text-xs text-muted-foreground">{t("worker.tasksEmptyHint")}</p>
        </div>
      ) : (
        <ul className="grid gap-2">
          {tasks.map((task, index) => (
            <TaskCard
              key={`${task.idempotencyKey || "task"}:${index}`}
              task={task}
            />
          ))}
        </ul>
      )}
    </section>
  )
}

function TaskCard({ task }: { task: WorkerTaskItem }) {
  const { t } = useTranslation()
  const stateKey = `worker.taskState.${task.state}`
  const stateLabel = t(stateKey)
  const ruleLabel = task.ruleName || task.ruleUid
  const datasourceLabel = task.datasourceName || task.datasourceUid
  return (
    <li className="grid min-w-0 gap-2 rounded-lg border border-border/60 bg-muted/20 p-2.5">
      <div className="flex min-w-0 flex-wrap items-start justify-between gap-x-2 gap-y-1.5">
        <div className="grid min-w-0 gap-0.5">
          <p className="min-w-0 text-sm leading-snug font-medium break-words">
            {ruleLabel || <DetailEmpty />}
          </p>
          {task.ruleUid ? (
            <p className="font-mono text-[11px] leading-4 break-all text-muted-foreground">
              {task.ruleUid}
            </p>
          ) : null}
          <p className="flex flex-wrap items-center gap-1 text-xs text-muted-foreground">
            <Database className="size-3.5 shrink-0" aria-hidden="true" />
            <span className="min-w-0 break-words">
              {datasourceLabel || <DetailEmpty />}
            </span>
          </p>
          {task.datasourceUid ? (
            <p className="font-mono text-[11px] leading-4 break-all text-muted-foreground">
              {task.datasourceUid}
            </p>
          ) : null}
        </div>
        <div className="flex shrink-0 flex-wrap items-center gap-1.5">
          <StatusBadge
            value={task.state}
            // 未知状态不回退到 enum.*，直接展示后端原值。
            label={stateLabel === stateKey ? task.state : stateLabel}
            tone={taskStateTones[task.state] ?? "neutral"}
          />
        </div>
      </div>
      {task.evalSlot ? (
        <p className="flex flex-wrap items-baseline gap-x-1.5 text-[11px] text-muted-foreground">
          <span>{t("worker.taskEvalSlot")}</span>
          <span className="font-mono break-all">{task.evalSlot}</span>
        </p>
      ) : null}
      <div className="grid min-w-0 gap-1">
        <p className="text-[11px] leading-4 font-medium tracking-wide text-muted-foreground">
          {t("worker.taskIdempotencyKey")}
        </p>
        <span className="flex min-w-0 items-start gap-1">
          {task.idempotencyKey ? (
            <>
              <code className="min-w-0 font-mono text-[11px] leading-5 break-all">
                {task.idempotencyKey}
              </code>
              <CopyIconButton
                value={task.idempotencyKey}
                label={t("worker.copyIdempotencyKey")}
                copiedLabel={t("action.copied")}
                errorLabel={t("action.copyFailed")}
                className="-my-0.5 size-6 opacity-60 hover:opacity-100 focus-visible:opacity-100"
              />
            </>
          ) : (
            <DetailEmpty />
          )}
        </span>
      </div>
    </li>
  )
}

function DetailSkeleton() {
  const { t } = useTranslation()
  return (
    <div className="grid gap-4" aria-busy="true" aria-label={t("state.loading")}>
      <Skeleton className="h-24 rounded-lg" />
      <Skeleton className="h-40 rounded-lg" />
    </div>
  )
}

/** 本地时间展示；空值或无法解析时返回空串，交给占位符处理。 */
function formatClock(value: string): string {
  const time = Date.parse(value)
  if (Number.isNaN(time)) {
    return ""
  }
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(time)
}

/**
 * 玉衡评测调度倒计时：以服务端的 nextScheduleAt 为准，到点或过期后显示 0
 * （「即将开始」），不会出现负数。
 */
function scheduleRemaining(
  nextScheduleAt: string,
  intervalSeconds: number,
  now: number
): number | null {
  if (!Number.isFinite(intervalSeconds) || intervalSeconds <= 0) {
    return null
  }
  const target = Date.parse(nextScheduleAt)
  if (Number.isNaN(target)) {
    return null
  }
  return Math.max(0, Math.ceil((target - now) / 1000))
}

/**
 * 本实例轮询倒计时：心跳 5 秒才上报一次，nextPollAt 可能已经过期，
 * 此时按 wall-clock 对齐兜底：剩余 = interval − (Date.now() % interval)。
 */
function pollRemaining(
  nextPollAt: string,
  intervalSeconds: number,
  now: number
): number | null {
  if (!Number.isFinite(intervalSeconds) || intervalSeconds <= 0) {
    return null
  }
  const intervalMs = intervalSeconds * 1000
  const target = Date.parse(nextPollAt)
  if (!Number.isNaN(target)) {
    const remaining = Math.ceil((target - now) / 1000)
    if (remaining > 0) {
      return remaining
    }
    // 刚过期时先显示「轮询中」，超过半个间隔才认为上报值已经失效。
    if (target - now > -intervalMs / 2) {
      return 0
    }
  }
  const toBoundary = intervalMs - (now % intervalMs)
  return Math.max(0, Math.ceil(toBoundary / 1000))
}
