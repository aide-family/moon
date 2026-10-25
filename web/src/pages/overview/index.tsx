import { RefreshCw } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import { getAlertInsight } from "@/api/alert-event"
import type { AlertInsight } from "@/api/alert-event/types"
import { fetchDashboard, fetchDeliveryTrend } from "@/api/dashboard"
import type {
  DashboardScope,
  DashboardSnapshot,
  DeliveryTrend,
} from "@/api/dashboard"
import { useSession } from "@/auth/session"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/lib/class-names"

import {
  AlertQueueCard,
  CompositionCard,
  DeliveryCard,
  KpiStrip,
  PlatformCard,
} from "./cards"
import { InsightTrendCard } from "./insight-trend"
import { LevelBreakdownCard } from "./level-breakdown"
import {
  bucketDayKey,
  bucketClosed,
  formatClockSeconds,
  todayKeyIn,
} from "./lib"
import { DeliveryTrendCard } from "./trend"

// 快照面板按「实时告警页」同节奏刷新；趋势属于历史区间，只在手动刷新时重算。
const REFRESH_MS = 30000
// 已在运行中的时长需要跟着时间走，但不需要 1 秒一跳。
const CLOCK_MS = 15000
const TREND_DAYS = 7
// 告警区间聚合的天数：两个卡（趋势 + 级别分布）共用同一次请求。
const INSIGHT_DAYS = 14

export function OverviewPage() {
  const { t } = useTranslation()
  const session = useSession()
  const [snapshot, setSnapshot] = useState<DashboardSnapshot | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [trend, setTrend] = useState<DeliveryTrend | null>(null)
  const [trendLoading, setTrendLoading] = useState(false)
  const [insight, setInsight] = useState<AlertInsight | null>(null)
  // 首屏就先进入加载态，避免直接闪一帧「区间内没有告警」。
  const [insightLoading, setInsightLoading] = useState(true)
  const [insightFailed, setInsightFailed] = useState(false)
  const [levelNames, setLevelNames] = useState<Record<string, string>>({})
  const [attempt, setAttempt] = useState(0)
  const [now, setNow] = useState(() => Date.now())

  const permissions = session.permissions
  const systemPermissions = session.systemPermissions
  const namespaceUid = session.namespaceUid

  const scope = useMemo<DashboardScope>(
    () => ({
      alerts: permissions.includes("alert:read"),
      delivery: permissions.includes("notification:read"),
      workers: systemPermissions.includes("system:user:read"),
      datasources: permissions.includes("datasource:read"),
      rules: permissions.includes("alert_rule:read"),
    }),
    [permissions, systemPermissions]
  )

  const enabled = namespaceUid !== ""

  // 首屏、轮询与手动刷新共用同一套加载逻辑：把它们放进同一个 effect，
  // 用 attempt 作为「再来一次」的触发器，避免在 effect 里同步改状态。
  useEffect(() => {
    if (!enabled) {
      return
    }
    let active = true
    let inFlight = false
    async function refresh(initial: boolean) {
      if (inFlight) {
        return
      }
      inFlight = true
      if (initial) {
        setLoading(true)
      } else {
        setRefreshing(true)
      }
      try {
        const next = await fetchDashboard(scope)
        if (!active) {
          return
        }
        setSnapshot(next)
        setNow(Date.now())
      } catch {
        // 整页加载失败时保留上一次快照，卡片继续显示旧值。
      } finally {
        inFlight = false
        if (active) {
          setLoading(false)
          setRefreshing(false)
        }
      }
    }
    void refresh(true)
    // 页面不可见时停止轮询，切回前台立刻补一次。
    const timer = window.setInterval(() => {
      if (document.visibilityState === "visible") {
        void refresh(false)
      }
    }, REFRESH_MS)
    const onVisible = () => {
      if (document.visibilityState === "visible") {
        void refresh(false)
      }
    }
    document.addEventListener("visibilitychange", onVisible)
    return () => {
      active = false
      window.clearInterval(timer)
      document.removeEventListener("visibilitychange", onVisible)
    }
  }, [enabled, scope, attempt])

  // 趋势属于历史区间，只在首屏和手动刷新时重算，不跟着 30 秒轮询跑。
  useEffect(() => {
    if (!enabled || !scope.delivery) {
      return
    }
    let cancelled = false
    async function loadTrend() {
      setTrendLoading(true)
      try {
        const result = await fetchDeliveryTrend(TREND_DAYS)
        if (!cancelled) {
          setTrend(result)
        }
      } catch {
        if (!cancelled) {
          setTrend(null)
        }
      } finally {
        if (!cancelled) {
          setTrendLoading(false)
        }
      }
    }
    void loadTrend()
    return () => {
      cancelled = true
    }
  }, [enabled, scope, attempt])

  // 区间聚合跟快照同节奏（30s）轮询，但独立成一个请求、独立降级：
  // 它失败时只有告警趋势与级别分布两张卡显示不可用，其余面板照常。
  useEffect(() => {
    if (!enabled || !scope.alerts) {
      return
    }
    let active = true
    let inFlight = false
    async function loadInsight() {
      if (inFlight) {
        return
      }
      inFlight = true
      setInsightLoading(true)
      try {
        const result = await getAlertInsight({ days: INSIGHT_DAYS })
        if (!active) {
          return
        }
        setInsight(result)
        setInsightFailed(false)
      } catch {
        if (active) {
          setInsightFailed(true)
        }
      } finally {
        inFlight = false
        if (active) {
          setInsightLoading(false)
        }
      }
    }
    void loadInsight()
    const timer = window.setInterval(() => {
      if (document.visibilityState === "visible") {
        void loadInsight()
      }
    }, REFRESH_MS)
    return () => {
      active = false
      window.clearInterval(timer)
    }
  }, [enabled, scope.alerts, attempt])

  // 级别名称属于配置数据，按租户取一次；解析不到的名称在卡片里显示「未分级」。
  useEffect(() => {
    if (!enabled || !scope.alerts) {
      return
    }
    let cancelled = false
    listAlertDictionaries({ page: 1, pageSize: 100, type: "ALERT" })
      .then((result) => {
        if (cancelled) {
          return
        }
        const next: Record<string, string> = {}
        for (const item of result.items) {
          if (item.type === "ALERT" && item.uid) {
            next[item.uid] = item.name
          }
        }
        setLevelNames(next)
      })
      .catch(() => {
        // 名称解析失败只影响展示名，不影响统计数字。
      })
    return () => {
      cancelled = true
    }
  }, [enabled, scope.alerts, namespaceUid])

  // 「今日新增」优先取时区里今天那一桶，取不到就退到最后一个桶（今天是最新一天）。
  const todayBucket = useMemo(() => {
    if (!insight || insight.buckets.length === 0) {
      return null
    }
    const key = todayKeyIn(insight.timeZone, new Date(now))
    return (
      insight.buckets.find((bucket) => bucketDayKey(bucket.start) === key) ??
      insight.buckets[insight.buckets.length - 1]
    )
  }, [insight, now])

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), CLOCK_MS)
    return () => window.clearInterval(timer)
  }, [])

  const showAlerts = scope.alerts
  const showDelivery = scope.delivery
  const showPlatform = scope.workers || scope.datasources || scope.rules
  const twoColumns = showAlerts && showDelivery
  const hasAnySection = showAlerts || showDelivery || showPlatform

  function refresh() {
    setAttempt((value) => value + 1)
  }

  if (!enabled || !hasAnySection) {
    return (
      <div className="min-h-0 flex-1 space-y-5 overflow-y-auto pr-1">
        <PageHeader
          title={t("overview.title")}
          description={t("overview.description")}
        />
        <Card className="px-6 py-10 text-center">
          <p className="text-sm font-medium">
            {enabled ? t("overview.noAccess") : t("overview.noNamespace")}
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            {enabled ? t("overview.noAccessHint") : t("overview.noNamespaceHint")}
          </p>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-0 flex-1 space-y-5 overflow-y-auto pr-1">
      <PageHeader
        title={t("overview.title")}
        description={t("overview.description")}
      >
        <div className="flex shrink-0 items-center gap-2">
          <span className="hidden text-xs text-muted-foreground tabular-nums sm:inline">
            {snapshot
              ? t("overview.updatedAt", {
                  time: formatClockSeconds(snapshot.observedAt),
                })
              : t("state.loading")}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={refresh}
            disabled={refreshing}
          >
            <RefreshCw className={refreshing ? "animate-spin" : undefined} />
            {refreshing ? t("overview.refreshing") : t("overview.refresh")}
          </Button>
        </div>
      </PageHeader>

      <KpiStrip
        alerts={showAlerts ? (snapshot?.alerts ?? null) : null}
        delivery={showDelivery ? (snapshot?.delivery ?? null) : null}
        today={
          todayBucket
            ? {
                events: todayBucket.events,
                closed: bucketClosed(todayBucket),
              }
            : null
        }
        showAlerts={showAlerts}
        showDelivery={showDelivery}
        // 今日新增来自区间聚合：只有首次聚合没回来时才显示骨架。
        // 之后的 30 秒轮询不能让整条 KPI 闪回骨架。
        loading={loading || (insightLoading && insight === null)}
      />

      <div
        className={cn(
          "grid min-w-0 gap-5",
          twoColumns && "xl:grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)]"
        )}
      >
        {showAlerts ? (
          <AlertQueueCard
            snapshot={snapshot?.alerts ?? null}
            loading={loading}
            now={now}
          />
        ) : null}
        {showAlerts || showDelivery ? (
          <div className="flex min-w-0 flex-col gap-5">
            {showAlerts ? (
              <CompositionCard
                stats={snapshot?.alerts?.stats ?? null}
                closedTotal={insight?.closedTotal ?? null}
                days={INSIGHT_DAYS}
                loading={loading}
              />
            ) : null}
            {showDelivery ? (
              <DeliveryCard
                delivery={snapshot?.delivery ?? null}
                loading={loading}
              />
            ) : null}
          </div>
        ) : null}
      </div>

      {showAlerts ? (
        <div className="grid min-w-0 gap-5 xl:grid-cols-3">
          <div className="min-w-0 xl:col-span-2">
            <InsightTrendCard
              insight={insight}
              days={INSIGHT_DAYS}
              loading={insightLoading}
              failed={insightFailed}
            />
          </div>
          <div className="min-w-0 xl:col-span-1">
            <LevelBreakdownCard
              levels={insight?.levels ?? null}
              eventsTotal={insight?.eventsTotal ?? 0}
              names={levelNames}
              loading={insightLoading}
              failed={insightFailed}
              days={INSIGHT_DAYS}
            />
          </div>
        </div>
      ) : null}

      {showDelivery || showPlatform ? (
        <div
          className={cn(
            "grid min-w-0 gap-5",
            showDelivery && showPlatform && "xl:grid-cols-3"
          )}
        >
          {showDelivery ? (
            <div className="min-w-0 xl:col-span-2">
              <DeliveryTrendCard trend={trend} loading={trendLoading} />
            </div>
          ) : null}
          {showPlatform ? (
            <div className={cn("min-w-0", showDelivery && "xl:col-span-1")}>
              <PlatformCard
                workers={snapshot?.workers ?? null}
                inventory={snapshot?.inventory ?? null}
                loading={loading}
              />
            </div>
          ) : null}
        </div>
      ) : null}

      <Card className="px-4 py-3">
        <h2 className="font-heading text-xs font-medium">
          {t("overview.boundary.title")}
        </h2>
        <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
          {t("overview.boundary.body")}
        </p>
      </Card>
    </div>
  )
}
