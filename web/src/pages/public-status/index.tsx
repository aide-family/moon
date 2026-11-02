import {
  CircleCheck,
  CircleHelp,
  Clock,
  RefreshCw,
  TriangleAlert,
} from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useParams } from "react-router-dom"

import error404Illustration from "@/assets/undraw/error-404.svg"
import {
  fetchPublicStatusPage,
  PublicStatusPageError,
} from "@/api/status-page/public"
import type {
  PublicStatusComponent,
  PublicStatusEvent,
  PublicStatusPage,
} from "@/api/status-page/types"
import { ErrorPage } from "@/components/error-page"
import { LocaleMenu } from "@/components/locale-menu"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/class-names"
import { formatDateTime } from "@/lib/date-time"

// 公开页没有登录态，也不做实时推送：每 60 秒在可见时补一次即可。
const REFRESH_MS = 60000

type OverallStatus = "OK" | "FAULT" | "UNKNOWN" | "EMPTY"

const bannerTone: Record<OverallStatus, string> = {
  OK: "border-success-border bg-success-soft text-success",
  FAULT: "border-danger-border bg-danger-soft text-danger",
  UNKNOWN: "border-warning-border bg-warning-soft text-warning",
  EMPTY: "border-border bg-muted/40 text-muted-foreground",
}

const componentTone: Record<string, "success" | "danger" | "warning"> = {
  OK: "success",
  FAULT: "danger",
  UNKNOWN: "warning",
}

/** 整体状态：任一组件故障即故障；否则有未知组件也标为未知；全部正常才算正常。 */
function overallStatus(components: PublicStatusComponent[]): OverallStatus {
  if (components.length === 0) {
    return "EMPTY"
  }
  if (components.some((component) => component.status === "FAULT")) {
    return "FAULT"
  }
  if (components.some((component) => component.status !== "OK")) {
    return "UNKNOWN"
  }
  return "OK"
}

function overallIcon(status: OverallStatus) {
  if (status === "OK") {
    return <CircleCheck aria-hidden="true" className="mt-0.5 size-5 shrink-0" />
  }
  if (status === "FAULT") {
    return (
      <TriangleAlert aria-hidden="true" className="mt-0.5 size-5 shrink-0" />
    )
  }
  return <CircleHelp aria-hidden="true" className="mt-0.5 size-5 shrink-0" />
}

/** 可用率字段是后端给的字符串；空串表示没有数据，显示占位符而不是 0%。 */
function availabilityText(value: string): string {
  return value ? `${value}%` : "—"
}

function ComponentCard({ component }: { component: PublicStatusComponent }) {
  const { t } = useTranslation()
  const tone = componentTone[component.status] ?? "warning"
  return (
    <Card className="min-w-0 px-4">
      <div className="flex items-start justify-between gap-3">
        <h3 className="min-w-0 font-medium break-words">{component.name}</h3>
        <StatusBadge
          value={component.status}
          label={t(`publicStatus.componentStatus.${component.status}`, {
            defaultValue: component.status,
          })}
          tone={tone}
        />
      </div>
      <dl className="mt-3 grid grid-cols-3 gap-2">
        {(
          [
            ["publicStatus.availability24h", component.availability24h],
            ["publicStatus.availability7d", component.availability7d],
            ["publicStatus.availability30d", component.availability30d],
          ] as const
        ).map(([label, value]) => (
          <div key={label} className="min-w-0">
            <dt className="text-xs text-muted-foreground">{t(label)}</dt>
            <dd className="font-mono text-sm tabular-nums">
              {availabilityText(value)}
            </dd>
          </div>
        ))}
      </dl>
      <p className="mt-3 flex items-center gap-1.5 text-xs text-muted-foreground">
        <Clock aria-hidden="true" className="size-3.5 shrink-0" />
        <span className="min-w-0 break-words">
          {t("publicStatus.lastChecked")}:{" "}
          {component.lastCheckedAt
            ? formatDateTime(component.lastCheckedAt)
            : t("publicStatus.neverChecked")}
        </span>
      </p>
    </Card>
  )
}

function EventList({
  events,
  emptyText,
}: {
  events: PublicStatusEvent[]
  emptyText: string
}) {
  const { t } = useTranslation()
  if (events.length === 0) {
    return <p className="mt-2 text-sm text-muted-foreground">{emptyText}</p>
  }
  return (
    <Card className="mt-2 min-w-0 divide-y divide-border/70 px-4">
      {events.map((event) => (
        <div
          key={event.uid}
          className="flex flex-wrap items-start justify-between gap-2 py-3"
        >
          <div className="min-w-0">
            <p className="text-sm font-medium break-words">{event.title}</p>
            <p className="mt-0.5 text-xs text-muted-foreground tabular-nums">
              {t("publicStatus.startedAt")}: {formatDateTime(event.startedAt)}
              {event.endedAt
                ? ` · ${t("publicStatus.endedAt")}: ${formatDateTime(event.endedAt)}`
                : ""}
            </p>
          </div>
          <StatusBadge
            value={event.status}
            label={t(`publicStatus.eventStatus.${event.status}`, {
              defaultValue: event.status,
            })}
            tone={event.status === "FIRING" ? "danger" : "success"}
          />
        </div>
      ))}
    </Card>
  )
}

/** 公开状态页：独立布局、不经过登录壳，只调用公开接口。 */
export function PublicStatusPageView() {
  const { t } = useTranslation()
  const { token = "" } = useParams()
  const [page, setPage] = useState<PublicStatusPage | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [failedStatus, setFailedStatus] = useState(0)
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let active = true
    let inFlight = false
    async function load(initial: boolean) {
      if (inFlight) {
        return
      }
      inFlight = true
      if (!initial) {
        setRefreshing(true)
      }
      try {
        const result = await fetchPublicStatusPage(token)
        if (!active) {
          return
        }
        setPage(result)
        setFailedStatus(0)
        setError("")
      } catch (err) {
        if (!active) {
          return
        }
        setFailedStatus(err instanceof PublicStatusPageError ? err.status : 0)
        setError(err instanceof Error ? err.message : "")
      } finally {
        inFlight = false
        if (active) {
          setLoading(false)
          setRefreshing(false)
        }
      }
    }
    void load(true)
    const timer = window.setInterval(() => {
      if (document.visibilityState === "visible") {
        void load(false)
      }
    }, REFRESH_MS)
    return () => {
      active = false
      window.clearInterval(timer)
    }
  }, [token, attempt])

  function reload() {
    setLoading(true)
    setAttempt((value) => value + 1)
  }

  if (loading && !page) {
    return (
      <main className="min-h-svh bg-background">
        <div className="mx-auto grid w-full max-w-5xl gap-4 px-4 py-10 sm:px-6">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-full max-w-xl" />
          <Skeleton className="mt-4 h-20 w-full" />
          <div className="mt-4 grid gap-4 md:grid-cols-2">
            <Skeleton className="h-40 w-full" />
            <Skeleton className="h-40 w-full" />
          </div>
        </div>
      </main>
    )
  }

  if (!page && failedStatus === 404) {
    return (
      <ErrorPage
        fullPage
        status={404}
        label="Not Found"
        title={t("publicStatus.invalidTitle")}
        description={t("publicStatus.invalidHint")}
        image={error404Illustration}
      />
    )
  }

  if (!page) {
    return (
      <main className="flex min-h-svh flex-col bg-background px-4 py-16">
        <div className="m-auto grid w-full max-w-md gap-3 text-center">
          <p className="font-heading text-xl font-semibold">
            {t("publicStatus.loadFailedTitle")}
          </p>
          <p className="text-sm text-muted-foreground">
            {failedStatus === 429
              ? t("publicStatus.rateLimited")
              : error || t("publicStatus.loadFailedHint")}
          </p>
          <div className="mt-2 flex justify-center">
            <Button variant="outline" onClick={reload}>
              <RefreshCw aria-hidden="true" data-icon="inline-start" />
              {t("action.retry")}
            </Button>
          </div>
        </div>
      </main>
    )
  }

  const overall = overallStatus(page.components)

  return (
    <main className="min-h-svh bg-background">
      <header className="border-b border-border/60">
        <div className="mx-auto flex w-full max-w-5xl flex-wrap items-center justify-between gap-2 px-4 py-3 sm:px-6">
          <div className="flex min-w-0 items-center gap-2">
            <img
              src="/alioth.svg"
              alt=""
              className="size-6 shrink-0 rounded-md"
            />
            <span className="font-heading text-sm font-semibold">Alioth</span>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={refreshing}
              onClick={() => setAttempt((value) => value + 1)}
            >
              <RefreshCw
                aria-hidden="true"
                data-icon="inline-start"
                className={refreshing ? "animate-spin" : undefined}
              />
              {t("publicStatus.refresh")}
            </Button>
            <LocaleMenu />
          </div>
        </div>
      </header>
      <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 sm:py-10">
        <div className="max-w-3xl">
          <h1 className="font-heading text-2xl font-semibold tracking-tight break-words sm:text-3xl">
            {page.title}
          </h1>
          {page.description ? (
            <p className="mt-2 text-sm leading-relaxed break-words text-muted-foreground">
              {page.description}
            </p>
          ) : null}
          <p className="mt-2 font-mono text-xs text-muted-foreground tabular-nums">
            {t("publicStatus.updatedAt", {
              time: formatDateTime(page.updatedAt) || "—",
            })}
          </p>
        </div>
        <div
          role="status"
          className={cn(
            "mt-6 flex items-start gap-3 rounded-xl border px-4 py-3",
            bannerTone[overall]
          )}
        >
          {overallIcon(overall)}
          <div className="min-w-0">
            <p className="text-sm font-medium break-words">
              {t(`publicStatus.overall.${overall}.title`)}
            </p>
            <p className="mt-0.5 text-xs break-words opacity-90">
              {t(`publicStatus.overall.${overall}.hint`)}
            </p>
          </div>
        </div>
        <section className="mt-8">
          <h2 className="font-heading text-lg font-medium">
            {t("publicStatus.componentsTitle")}
          </h2>
          {page.components.length === 0 ? (
            <p className="mt-2 text-sm text-muted-foreground">
              {t("publicStatus.empty")}
            </p>
          ) : (
            <div className="mt-3 grid gap-4 md:grid-cols-2">
              {page.components.map((component) => (
                <ComponentCard key={component.groupUid} component={component} />
              ))}
            </div>
          )}
        </section>
        <div className="mt-8 grid gap-6">
          <section>
            <h2 className="font-heading text-lg font-medium">
              {t("publicStatus.activeEvents")}
            </h2>
            <EventList
              events={page.activeEvents}
              emptyText={t("publicStatus.noActiveEvents")}
            />
          </section>
          <section>
            <h2 className="font-heading text-lg font-medium">
              {t("publicStatus.recentEvents")}
            </h2>
            <EventList
              events={page.recentEvents}
              emptyText={t("publicStatus.noRecentEvents")}
            />
          </section>
        </div>
        <p className="mt-10 text-center text-xs text-muted-foreground">
          {t("publicStatus.poweredBy")}
        </p>
      </div>
    </main>
  )
}
