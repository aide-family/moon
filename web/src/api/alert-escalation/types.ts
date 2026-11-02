/** 升级策略状态：启用 / 停用。 */
export type AlertEscalationPolicyStatus = "ACTIVE" | "DISABLED"

/** 提交的升级步骤：uid 为空表示新增，服务端按数组顺序重写 stepOrder。 */
export interface AlertEscalationStepInput {
  uid?: string
  /** 距事件触发时间的累计分钟数，必须严格递增，最后一步不得为 0。 */
  delayMinutes: number
  /** 本步投递的通知组，至少一个。 */
  notificationGroupUids: string[]
}

/** 详情里的升级步骤。 */
export interface AlertEscalationStepItem {
  uid: string
  /** 从 0 开始，同一策略内唯一。 */
  stepOrder: number
  delayMinutes: number
  notificationGroupUids: string[]
}

/**
 * 升级策略。匹配条件之间是 AND，空条件表示不限；
 * labels 是 JSON 字符串（标签键 → 完整匹配正则），与后端存储格式一致。
 */
export interface AlertEscalationPolicyItem {
  uid: string
  name: string
  description: string
  labels: string
  levelUids: string[]
  datasourceUids: string[]
  ruleUids: string[]
  status: AlertEscalationPolicyStatus | string
  createdBy: string
  createdAt: string
  updatedAt: string
  steps: AlertEscalationStepItem[]
}

export interface ListAlertEscalationPoliciesParams {
  page?: number
  pageSize?: number
  name?: string
  /** ACTIVE / DISABLED，空值表示不限制。 */
  status?: string
}

export interface ListAlertEscalationPoliciesResult {
  items: AlertEscalationPolicyItem[]
  total: number
  page: number
  pageSize: number
}

/** 创建与更新共用的载荷；更新时由接口函数补上 uid。 */
export interface SaveAlertEscalationPolicyParams {
  name: string
  description?: string
  labels: string
  levelUids?: string[]
  datasourceUids?: string[]
  ruleUids?: string[]
  /** 至少一步，步骤顺序即升级顺序。 */
  steps: AlertEscalationStepInput[]
}

/** 事件升级记录的状态。 */
export type EventEscalationStatus = "PENDING" | "SENT" | "SKIPPED" | "CANCELLED"

/** 一条事件升级记录，供事件详情时间线展示。 */
export interface EventEscalationItem {
  uid: string
  eventUid: string
  policyUid: string
  /** 策略名称；策略已删除时为空。 */
  policyName: string
  /** 从 0 开始。 */
  stepOrder: number
  /** 距事件触发时间的累计分钟数。 */
  delayMinutes: number
  /** 本步通知组数量。 */
  notificationGroupCount: number
  dueAt: string
  status: EventEscalationStatus | string
  /** 实际发送时间；未发送为空。 */
  sentAt: string
}

export interface ListEventEscalationsResult {
  items: EventEscalationItem[]
}
