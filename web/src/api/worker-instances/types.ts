export interface WorkerInstanceItem {
  service: "MEGREZ" | "MERAK" | string
  instanceUid: string
  hostname: string
  busy: boolean
  taskUid: string
  taskName: string
  taskStartedAt: string
  lastSeenAt: string
  online: boolean
}

export interface ListWorkerInstancesResult {
  items: WorkerInstanceItem[]
  onlineCount: number
  busyCount: number
  idleCount: number
  observedAt: string
}

/** 本轮抢到的「规则 × 数据源」组合及其跑批状态。 */
export interface WorkerTaskItem {
  /** RUNNING / NEW / PENDING / DONE / LEFT；未知值按原样展示。 */
  state: string
  ruleUid: string
  ruleName: string
  datasourceUid: string
  datasourceName: string
  /** 本轮评测时间槽，RFC3339。 */
  evalSlot: string
  /** 服务端算好的跑批幂等键：ruleUID/datasourceUID@evalSlot。 */
  idempotencyKey: string
}

/** 工作进程详情：实例元信息加上本轮领到的组合列表。 */
export interface WorkerInstanceDetail extends WorkerInstanceItem {
  /** 本轮开始时间，空串表示尚未开始。 */
  roundStartedAt: string
  /** 实例下次轮询时间，空串表示无值。 */
  nextPollAt: string
  pollIntervalSeconds: number
  tasks: WorkerTaskItem[]
}

export interface WorkerInstanceDetailResult {
  item: WorkerInstanceDetail
  /** 下一次玉衡评测调度时间，RFC3339。 */
  nextScheduleAt: string
  scheduleIntervalSeconds: number
  observedAt: string
}
