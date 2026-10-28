import { Activity, CircleAlert, Clock3, RefreshCw, Server } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { listWorkerInstances } from "@/api/worker-instances"
import type { ListWorkerInstancesResult, WorkerInstanceItem } from "@/api/worker-instances"
import { PageHeader } from "@/components/page-header"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

const services = [
  { key: "MEGREZ", title: "Megrez", descriptionKey: "megrezDescription" },
  { key: "MERAK", title: "Merak", descriptionKey: "merakDescription" },
] as const

function formatTime(value: string): string {
  const time = Date.parse(value)
  if (Number.isNaN(time)) return "—"
  return new Intl.DateTimeFormat(undefined, { hour: "2-digit", minute: "2-digit", second: "2-digit" }).format(time)
}

function formatDuration(value: string, now: number, t: (key: string, options?: Record<string, number>) => string): string {
  const started = Date.parse(value)
  if (Number.isNaN(started)) return "—"
  const seconds = Math.max(0, Math.floor((now - started) / 1000))
  if (seconds < 60) return t("worker.seconds", { count: seconds })
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return t("worker.minutes", { count: minutes, seconds: seconds % 60 })
  return t("worker.hours", { count: Math.floor(minutes / 60), minutes: minutes % 60 })
}

function WorkerRow({ item, now }: { item: WorkerInstanceItem; now: number }) {
  const { t } = useTranslation()
  const heartbeatAge = formatDuration(item.lastSeenAt, now, t)
  const taskClock = item.online ? now : Date.parse(item.lastSeenAt)
  const taskAge = formatDuration(item.taskStartedAt, taskClock, t)

  return (
    <li className="grid gap-3 rounded-lg border border-border/60 bg-muted/25 p-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <span className="max-w-full truncate font-medium">{item.hostname || item.instanceUid}</span>
          <StatusBadge
            value={item.online ? "ONLINE" : "OFFLINE"}
            label={item.online ? t("worker.online") : t("worker.offline")}
          />
          {item.online ? (
            <StatusBadge
              value={item.busy ? "BUSY" : "IDLE"}
              label={item.busy ? t("worker.working") : t("worker.idle")}
            />
          ) : null}
        </div>
        <p className="mt-1 truncate font-mono text-xs text-muted-foreground" title={item.instanceUid}>{item.instanceUid}</p>
        {item.busy && item.taskUid ? (
          <div className="mt-3 rounded-md bg-background/70 px-3 py-2">
            <p className="truncate text-sm font-medium" title={item.taskName}>{item.taskName || t("worker.unnamedTask")}</p>
            <p className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
              <span className="font-mono">{item.taskUid}</span>
              {item.taskStartedAt ? (
                <span>
                  {item.online
                    ? t("worker.runningFor", { duration: taskAge })
                    : t("worker.lastKnownTask", { time: formatTime(item.lastSeenAt), duration: taskAge })}
                </span>
              ) : null}
            </p>
          </div>
        ) : (
          <p className="mt-2 text-xs text-muted-foreground">{item.online ? t("worker.noActiveTask") : t("worker.lastSeen", { time: formatTime(item.lastSeenAt) })}</p>
        )}
      </div>
      <div className="flex items-center gap-1.5 text-xs tabular-nums text-muted-foreground">
        <Clock3 className="size-3.5" aria-hidden="true" />
        <span className="sr-only">{t("worker.lastHeartbeat", { duration: heartbeatAge })}</span>
        <span>{heartbeatAge}</span>
      </div>
    </li>
  )
}

function ServicePanel({ service, items, now }: { service: (typeof services)[number]; items: WorkerInstanceItem[]; now: number }) {
  const { t } = useTranslation()
  const serviceItems = items.filter((item) => item.service === service.key)
  const online = serviceItems.filter((item) => item.online)
  const busy = online.filter((item) => item.busy).length

  return (
    <Card className="min-w-0">
      <CardHeader className="gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <CardTitle className="flex items-center gap-2 text-lg">
            <span className="grid size-8 place-items-center rounded-md bg-primary/10 text-primary-ink"><Server className="size-4" aria-hidden="true" /></span>
            {service.title}
          </CardTitle>
          <CardDescription>{t(`worker.${service.descriptionKey}`)}</CardDescription>
        </div>
        <div className="flex shrink-0 items-end gap-4 tabular-nums">
          <div>
            <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{t("worker.working")}</p>
            <p className="text-2xl font-semibold leading-none">{busy}<span className="ml-1 text-sm font-normal text-muted-foreground">/ {online.length}</span></p>
          </div>
          <div className="pb-0.5 text-xs text-muted-foreground">{t("worker.onlineUnits", { count: online.length })}</div>
        </div>
      </CardHeader>
      <CardContent>
        {!serviceItems.length ? (
          <div className="rounded-lg border border-dashed border-border/70 px-4 py-8 text-center">
            <Activity className="mx-auto size-5 text-muted-foreground/70" aria-hidden="true" />
            <p className="mt-2 text-sm font-medium">{t("worker.waitingHeartbeat")}</p>
            <p className="mt-1 text-xs text-muted-foreground">{t("worker.waitingHeartbeatHint")}</p>
          </div>
        ) : (
          <>
            <ul className="space-y-2" aria-label={t(`worker.${service.descriptionKey}`)}>
              {serviceItems.map((item) => <WorkerRow key={`${item.service}:${item.instanceUid}`} item={item} now={now} />)}
            </ul>
          </>
        )}
      </CardContent>
    </Card>
  )
}

export function WorkerStatusPage() {
  const { t } = useTranslation()
  const [snapshot, setSnapshot] = useState<ListWorkerInstancesResult | null>(null)
  const [failed, setFailed] = useState(false)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [now, setNow] = useState(0)

  useEffect(() => {
    let active = true
    let inFlight = false
    async function refresh() {
      if (inFlight) return
      inFlight = true
      setRefreshing(true)
      try {
        const result = await listWorkerInstances()
        if (!active) return
        setSnapshot(result)
        setFailed(false)
        setNow(Date.now())
      } catch {
        if (active) setFailed(true)
      } finally {
        inFlight = false
        if (active) {
          setLoading(false)
          setRefreshing(false)
        }
      }
    }
    void refresh()
    const timer = window.setInterval(() => { void refresh() }, 5000)
    return () => {
      active = false
      window.clearInterval(timer)
    }
  }, [attempt])

  const items = snapshot?.items ?? []
  const busyNames = services.map((service) => ({
    title: service.title,
    value: items.filter((item) => item.service === service.key && item.online && item.busy).length,
  }))

  return (
    <div className="flex min-h-0 flex-col gap-5 overflow-y-auto">
      <PageHeader title={t("worker.title")} description={t("worker.description")}>
        <Button type="button" variant="outline" size="sm" onClick={() => setAttempt((value) => value + 1)} disabled={refreshing}>
          <RefreshCw className={refreshing ? "motion-safe:animate-spin" : ""} aria-hidden="true" />
          {t("action.refresh")}
        </Button>
      </PageHeader>

      {failed && snapshot ? <p role="status" aria-live="polite" className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">{t("worker.refreshFailed")}</p> : null}

      {loading ? (
        <div className="grid gap-4 lg:grid-cols-2" aria-busy="true" aria-label={t("state.loading")}>
          <Skeleton className="h-56 rounded-xl" />
          <Skeleton className="h-56 rounded-xl" />
        </div>
      ) : snapshot ? (
        <>
          <Card className="overflow-hidden border-primary/15 bg-primary/[0.035]">
            <CardContent className="grid gap-5 p-5 sm:grid-cols-[1fr_auto] sm:items-center">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.14em] text-muted-foreground">{t("worker.onlineNow")}</p>
                <div className="mt-1 flex items-baseline gap-3 tabular-nums">
                  <span className="text-4xl font-semibold tracking-tight">{snapshot?.onlineCount ?? 0}</span>
                  <span className="text-sm text-muted-foreground">{t("worker.capacityLine", { busy: snapshot?.busyCount ?? 0, idle: snapshot?.idleCount ?? 0 })}</span>
                </div>
                <p className="mt-2 text-xs text-muted-foreground">{t("worker.heartbeatHint")}</p>
              </div>
              <div className="flex gap-5 border-t border-border/50 pt-4 tabular-nums sm:border-l sm:border-t-0 sm:pl-6 sm:pt-0">
                {busyNames.map((item) => (
                  <div key={item.title}>
                    <p className="text-xs text-muted-foreground">{item.title}</p>
                    <p className="mt-0.5 text-lg font-semibold">{item.value}<span className="ml-1 text-xs font-normal text-muted-foreground">{t("worker.workingUnits")}</span></p>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          <div className="grid gap-4 xl:grid-cols-2">
            {services.map((service) => <ServicePanel key={service.key} service={service} items={items} now={now} />)}
          </div>
          <div className="flex flex-wrap items-center justify-between gap-2 px-1 text-xs text-muted-foreground">
            <span>{t("worker.autoRefresh")}</span>
            {snapshot?.observedAt ? <span>{t("worker.lastUpdated", { time: formatTime(snapshot.observedAt) })}</span> : null}
          </div>
        </>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center gap-3 py-10 text-center" role="status" aria-live="polite">
            <CircleAlert className="size-6 text-destructive" aria-hidden="true" />
            <div>
              <p className="font-medium">{t("worker.loadFailed")}</p>
              <p className="mt-1 text-sm text-muted-foreground">{t("worker.unavailableHint")}</p>
            </div>
            <Button type="button" variant="outline" size="sm" onClick={() => setAttempt((value) => value + 1)}>
              {t("action.retry")}
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
