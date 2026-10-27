import { get } from "@/api/http"

/**
 * 历史告警-统计分析的 6 个服务端聚合接口。
 *
 * 每个图表块一次请求，前端不再翻页拉事件、也不在浏览器里聚合。字段名与后端
 * protojson（camelCase）一一对应；数字字段统一用 toNumber 兜底，避免 int64
 * 被编码成字符串时把 NaN 带进图表。
 */

/** 分析页支持的状态筛选；空串由接口省略，表示全部状态。 */
export type AlertAnalysisStatusFilter =
  "RECOVERED" | "RECOVERED_BY_MANUAL" | "SUPPRESSING"

export type AlertAnalysisGranularity = "hour" | "day" | "week" | "month"

/** summary/status/level/fingerprints/recovery 共用的查询条件。 */
export interface AlertAnalysisRangeParams {
  /** 统计起始自然日（YYYY-MM-DD，包含）。 */
  from: string
  /** 统计结束自然日（YYYY-MM-DD，包含）。 */
  to: string
  /** 用于按本地自然日划分区间的 IANA 时区；空值接口按 UTC 处理。 */
  timeZone?: string
  /** 状态筛选；空串表示全部状态。 */
  status?: AlertAnalysisStatusFilter | ""
}

/**
 * 可选的聚焦窗口：只影响 summary/status/level/fingerprints/recovery，
 * trend 接口忽略它。成对出现且都落在 from/to 窗口内。
 */
export interface AlertAnalysisFocusParams extends AlertAnalysisRangeParams {
  /** 聚焦窗口起点，RFC3339（包含）。 */
  focusStart?: string
  /** 聚焦窗口终点，RFC3339（排除）。 */
  focusEnd?: string
}

export interface AlertAnalysisTrendParams extends AlertAnalysisRangeParams {
  /** 趋势分桶粒度；week 从周一开始。 */
  granularity: AlertAnalysisGranularity
}

export interface AlertAnalysisFingerprintsParams extends AlertAnalysisFocusParams {
  /** 返回的重复指纹条数，1..20，默认 5。 */
  limit?: number
}

/** GET /v1/alert-analysis/summary */
export interface AlertAnalysisSummary {
  from: string
  to: string
  timeZone: string
  /** 为假表示历史分库不可用，结果只含实时表数据。 */
  complete: boolean
  /** 统计范围内的事件数。 */
  events: number
  /** 涉及的指纹数。 */
  uniqueFingerprints: number
  /** 同一指纹第二次及之后触发的事件数。 */
  repeated: number
}

/** trend 的一个非空时间桶。 */
export interface AlertAnalysisTrendBucket {
  /** 本地时区的桶起点，RFC3339。 */
  start: string
  events: number
}

/** GET /v1/alert-analysis/trend */
export interface AlertAnalysisTrend {
  from: string
  to: string
  timeZone: string
  complete: boolean
  /** 只含非空桶，按 start 升序。 */
  buckets: AlertAnalysisTrendBucket[]
}

export interface AlertAnalysisStatusItem {
  /** 服务端状态值；SUPPRESSING 在接口里是 status='FIRING'。 */
  status: string
  events: number
}

/** GET /v1/alert-analysis/status */
export interface AlertAnalysisStatus {
  complete: boolean
  total: number
  items: AlertAnalysisStatusItem[]
}

export interface AlertAnalysisLevelItem {
  levelUid: string
  events: number
}

/** GET /v1/alert-analysis/level */
export interface AlertAnalysisLevel {
  complete: boolean
  total: number
  items: AlertAnalysisLevelItem[]
}

export interface AlertAnalysisFingerprintItem {
  fingerprint: string
  ruleUid: string
  ruleName: string
  events: number
}

/** GET /v1/alert-analysis/fingerprints */
export interface AlertAnalysisFingerprints {
  complete: boolean
  /** 只含 events>1 的指纹，按 events 降序、fingerprint 升序。 */
  items: AlertAnalysisFingerprintItem[]
}

export interface AlertAnalysisRecoveryStats {
  /** 恢复时长中位数（秒）；没有样本时为 0。 */
  medianSeconds: number
  /** 恢复时长 P90（秒，最近秩口径）；没有样本时为 0。 */
  p90Seconds: number
  /** 参与分位数计算的事件数，为 0 时上面两个值没有意义。 */
  samples: number
}

/** GET /v1/alert-analysis/recovery */
export interface AlertAnalysisRecovery {
  complete: boolean
  auto: AlertAnalysisRecoveryStats
  manual: AlertAnalysisRecoveryStats
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
 * 计数字段在 proto 里是 int32/int64，但 protojson 对 64 位整数的编码习惯在
 * 不同版本里不一致，这里统一兜底成有限数字。
 */
function toNumber(value: unknown): number {
  const parsed = typeof value === "string" ? Number(value) : value
  return typeof parsed === "number" && Number.isFinite(parsed) ? parsed : 0
}

function rangeQuery(params: AlertAnalysisRangeParams) {
  return {
    from: params.from,
    to: params.to,
    timeZone: params.timeZone ?? localTimeZone(),
    // 空串会被 get() 跳过，等价于不传 status（全部状态）。
    status: params.status ?? "",
  }
}

/** focus 必须成对出现；只传一个时整体省略，避免服务端按半个窗口聚合。 */
function focusQuery(params: AlertAnalysisFocusParams) {
  if (!params.focusStart || !params.focusEnd) {
    return {}
  }
  return { focusStart: params.focusStart, focusEnd: params.focusEnd }
}

function stats(value: AlertAnalysisRecoveryStats | undefined) {
  return {
    medianSeconds: toNumber(value?.medianSeconds),
    p90Seconds: toNumber(value?.p90Seconds),
    samples: toNumber(value?.samples),
  }
}

/** GET /v1/alert-analysis/summary（支持聚焦窗口，KPI 走聚焦口径） */
export function getAlertAnalysisSummary(
  params: AlertAnalysisFocusParams
): Promise<AlertAnalysisSummary> {
  const query = { ...rangeQuery(params), ...focusQuery(params) }
  return get<AlertAnalysisSummary>(
    "/alert-analysis/summary",
    query,
    undefined,
    {
      silent: true,
    }
  ).then((data) => ({
    from: data.from ?? "",
    to: data.to ?? "",
    timeZone: data.timeZone || query.timeZone,
    complete: data.complete !== false,
    events: toNumber(data.events),
    uniqueFingerprints: toNumber(data.uniqueFingerprints),
    repeated: toNumber(data.repeated),
  }))
}

/** GET /v1/alert-analysis/trend（忽略 focus） */
export function getAlertAnalysisTrend(
  params: AlertAnalysisTrendParams
): Promise<AlertAnalysisTrend> {
  const query = { ...rangeQuery(params), granularity: params.granularity }
  return get<AlertAnalysisTrend>("/alert-analysis/trend", query, undefined, {
    silent: true,
  }).then((data) => ({
    from: data.from ?? "",
    to: data.to ?? "",
    timeZone: data.timeZone || query.timeZone,
    complete: data.complete !== false,
    buckets: (data.buckets ?? [])
      .map((bucket) => ({
        start: bucket.start ?? "",
        events: toNumber(bucket.events),
      }))
      .filter((bucket) => bucket.start),
  }))
}

/** GET /v1/alert-analysis/status */
export function getAlertAnalysisStatus(
  params: AlertAnalysisFocusParams
): Promise<AlertAnalysisStatus> {
  const query = { ...rangeQuery(params), ...focusQuery(params) }
  return get<AlertAnalysisStatus>("/alert-analysis/status", query, undefined, {
    silent: true,
  }).then((data) => ({
    complete: data.complete !== false,
    total: toNumber(data.total),
    items: (data.items ?? []).map((item) => ({
      status: item.status ?? "",
      events: toNumber(item.events),
    })),
  }))
}

/** GET /v1/alert-analysis/level */
export function getAlertAnalysisLevel(
  params: AlertAnalysisFocusParams
): Promise<AlertAnalysisLevel> {
  const query = { ...rangeQuery(params), ...focusQuery(params) }
  return get<AlertAnalysisLevel>("/alert-analysis/level", query, undefined, {
    silent: true,
  }).then((data) => ({
    complete: data.complete !== false,
    total: toNumber(data.total),
    items: (data.items ?? []).map((item) => ({
      levelUid: item.levelUid ?? "",
      events: toNumber(item.events),
    })),
  }))
}

/** GET /v1/alert-analysis/fingerprints */
export function getAlertAnalysisFingerprints(
  params: AlertAnalysisFingerprintsParams
): Promise<AlertAnalysisFingerprints> {
  const query = {
    ...rangeQuery(params),
    ...focusQuery(params),
    limit: params.limit && params.limit > 0 ? params.limit : "",
  }
  return get<AlertAnalysisFingerprints>(
    "/alert-analysis/fingerprints",
    query,
    undefined,
    { silent: true }
  ).then((data) => ({
    complete: data.complete !== false,
    items: (data.items ?? []).map((item) => ({
      fingerprint: item.fingerprint ?? "",
      ruleUid: item.ruleUid ?? "",
      ruleName: item.ruleName ?? "",
      events: toNumber(item.events),
    })),
  }))
}

/** GET /v1/alert-analysis/recovery */
export function getAlertAnalysisRecovery(
  params: AlertAnalysisFocusParams
): Promise<AlertAnalysisRecovery> {
  const query = { ...rangeQuery(params), ...focusQuery(params) }
  return get<AlertAnalysisRecovery>(
    "/alert-analysis/recovery",
    query,
    undefined,
    {
      silent: true,
    }
  ).then((data) => ({
    complete: data.complete !== false,
    auto: stats(data.auto),
    manual: stats(data.manual),
  }))
}
