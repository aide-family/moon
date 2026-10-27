import { get, post } from "@/api/http"

import type {
  AlertEventChart,
  AlertInsight,
  AlertInsightParams,
  AlertEventItem,
  AlertEventRealtimeStats,
  AlertEventStats,
  AlertEventSuppressResult,
  AlertEventUids,
  ListAlertEventsParams,
  ListAlertEventsResult,
} from "./types"

/** GET /v1/alert-events */
export function listAlertEvents(
  params: ListAlertEventsParams = {}
): Promise<ListAlertEventsResult> {
  return get<ListAlertEventsResult>("/alert-events", params)
}

/** GET /v1/alert-events/{uid} */
export function getAlertEvent(uid: string): Promise<{ event: AlertEventItem }> {
  return get<{ event: AlertEventItem }>(`/alert-events/${uid}`)
}

/** GET /v1/alert-events/{uid}/chart */
export function getAlertEventChart(uid: string): Promise<AlertEventChart> {
  return get<AlertEventChart>(`/alert-events/${uid}/chart`)
}

/** GET /v1/alert-event-stats */
export function alertEventStats(): Promise<AlertEventStats> {
  return get<AlertEventStats>("/alert-event-stats")
}

/** GET /v1/alert-event-realtime-stats */
export function alertEventRealtimeStats(): Promise<AlertEventRealtimeStats> {
  return get<AlertEventRealtimeStats>("/alert-event-realtime-stats")
}

/** 浏览器时区；取不到时退回 UTC，不能因为环境异常把分桶算错。 */
function localTimeZone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC"
  } catch {
    return "UTC"
  }
}

/**
 * 计数字段按 number 声明（proto 里是 int32），但 protojson 对 int64/double
 * 的编码习惯在不同版本里不一致，这里统一兜底成有限数字，避免 NaN 穿透到图表。
 */
function toNumber(value: unknown): number {
  const parsed = typeof value === "string" ? Number(value) : value
  return typeof parsed === "number" && Number.isFinite(parsed) ? parsed : 0
}

/**
 * GET /v1/alert-insight
 * 概览用的区间聚合：每日新增与闭环（两条独立时间轴）、区间闭环时长分位数与级别分布。
 * 分桶与分位数都由服务端算好（P90 是最近秩口径），前端只做展示与归一化。
 */
export function getAlertInsight(
  params: AlertInsightParams = {}
): Promise<AlertInsight> {
  const query: AlertInsightParams = {
    timeZone: params.timeZone ?? localTimeZone(),
  }
  if (params.days && params.days > 0) {
    query.days = params.days
  }
  return get<AlertInsight>("/alert-insight", query, undefined, {
    silent: true,
  }).then((data) => ({
    timeZone: data.timeZone || query.timeZone || "UTC",
    from: data.from ?? "",
    to: data.to ?? "",
    buckets: (data.buckets ?? []).map((bucket) => ({
      start: bucket.start ?? "",
      events: toNumber(bucket.events),
      closed: toNumber(bucket.closed),
      closedByManual: toNumber(bucket.closedByManual),
      p50Seconds: toNumber(bucket.p50Seconds),
      p90Seconds: toNumber(bucket.p90Seconds),
      recoverySamples: toNumber(bucket.recoverySamples),
    })),
    levels: (data.levels ?? []).map((level) => ({
      levelUid: level.levelUid ?? "",
      events: toNumber(level.events),
      firing: toNumber(level.firing),
    })),
    p50Seconds: toNumber(data.p50Seconds),
    p90Seconds: toNumber(data.p90Seconds),
    recoverySamples: toNumber(data.recoverySamples),
    eventsTotal: toNumber(data.eventsTotal),
    closedTotal: toNumber(data.closedTotal),
    complete: data.complete !== false,
  }))
}

/** POST /v1/alert-events/{uid}/intervene */
export function interveneAlertEvent(
  uid: string
): Promise<Record<string, never>> {
  return post<Record<string, never>>(`/alert-events/${uid}/intervene`, { uid })
}

/** POST /v1/alert-events/intervene */
export function interveneAlertEvents(
  body: AlertEventUids
): Promise<Record<string, never>> {
  return post<Record<string, never>>("/alert-events/intervene", body)
}

/** POST /v1/alert-events/{uid}/suppress */
export function suppressAlertEvent(
  uid: string,
  suppressedSeconds: number,
  permanent = false
): Promise<AlertEventSuppressResult> {
  return post<AlertEventSuppressResult>(`/alert-events/${uid}/suppress`, {
    uid,
    suppressedSeconds,
    permanent,
  })
}

/** POST /v1/alert-events/{uid}/recover */
export function recoverAlertEvent(uid: string): Promise<Record<string, never>> {
  return post<Record<string, never>>(`/alert-events/${uid}/recover`, { uid })
}

/** POST /v1/alert-events/recover */
export function recoverAlertEvents(
  body: AlertEventUids
): Promise<Record<string, never>> {
  return post<Record<string, never>>("/alert-events/recover", body)
}

export {
  getAlertAnalysisFingerprints,
  getAlertAnalysisLevel,
  getAlertAnalysisRecovery,
  getAlertAnalysisStatus,
  getAlertAnalysisSummary,
  getAlertAnalysisTrend,
} from "./history-analysis"

export type {
  AlertAnalysisFingerprints,
  AlertAnalysisFingerprintItem,
  AlertAnalysisFocusParams,
  AlertAnalysisGranularity,
  AlertAnalysisLevel,
  AlertAnalysisLevelItem,
  AlertAnalysisRangeParams,
  AlertAnalysisRecovery,
  AlertAnalysisRecoveryStats,
  AlertAnalysisStatus,
  AlertAnalysisStatusFilter,
  AlertAnalysisStatusItem,
  AlertAnalysisSummary,
  AlertAnalysisTrend,
  AlertAnalysisTrendBucket,
  AlertAnalysisTrendParams,
} from "./history-analysis"

export type {
  AlertEventChart,
  AlertInsight,
  AlertInsightBucket,
  AlertInsightLevel,
  AlertInsightParams,
  AlertEventItem,
  AlertEventRealtimeStats,
  AlertEventStats,
  AlertEventSuppressResult,
  AlertEventUids,
  AlertEventView,
  ListAlertEventsParams,
  ListAlertEventsResult,
  RealtimeAlertTabStat,
} from "./types"
