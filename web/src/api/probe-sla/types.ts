/** 按天拆分的 SLA 明细；day 为 UTC 自然日 YYYY-MM-DD。 */
export interface ProbeSlaDay {
  day: string
  checks: number
  okChecks: number
  failedChecks: number
  /** 可用率，例如 "99.95"；没有数据时为空串。 */
  availability: string
  /** 平均延迟，毫秒。 */
  avgDurationMs: number
}

/** 单个目标组的 SLA 汇总。 */
export interface ProbeSlaGroup {
  groupUid: string
  groupName: string
  checks: number
  okChecks: number
  failedChecks: number
  /** 可用率，例如 "99.95"；没有数据时为空串。 */
  availability: string
  avgDurationMs: number
  p95DurationMs: number
  /** 超出原始结果保留期时 P95 用最大值近似，页面必须标注。 */
  p95Approximate: boolean
  /** 区间内的探测事件数。 */
  eventCount: number
  days: ProbeSlaDay[]
}

export interface ProbeSlaReport {
  from: string
  to: string
  groups: ProbeSlaGroup[]
}

export interface ProbeSlaParams {
  /** 目标组；为空表示全部。 */
  groupUids?: string[]
  /** RFC3339，早于 to。 */
  from: string
  /** RFC3339，跨度不超过 30 天。 */
  to: string
}
