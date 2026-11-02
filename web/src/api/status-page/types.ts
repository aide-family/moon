/** 状态页状态，与后端约定为大写英文。 */
export type StatusPageStatus = "ACTIVE" | "DISABLED"

/** 状态页配置；列表接口同样返回 token，便于直接复制公开链接。 */
export interface StatusPageItem {
  uid: string
  name: string
  /** 对外展示的标题。 */
  title: string
  /** 对外展示的描述。 */
  description: string
  status: StatusPageStatus
  /** 要展示的探测目标组，至少一个。 */
  targetGroupUids: string[]
  /** 公开访问凭证；重新生成后旧链接立即失效。 */
  token: string
  createdAt: string
  updatedAt: string
}

export interface ListStatusPagesParams {
  page?: number
  pageSize?: number
  name?: string
  /** 空串表示不限制状态。 */
  status?: StatusPageStatus | ""
}

export interface ListStatusPagesResult {
  items: StatusPageItem[]
  total: number
  page: number
  pageSize: number
}

/** 新建与更新共用；更新时 uid 必填。 */
export interface SaveStatusPageParams {
  name: string
  title: string
  description: string
  targetGroupUids: string[]
}

/** 公开页组件状态：正常 / 故障 / 未知。 */
export type PublicComponentStatus = "OK" | "FAULT" | "UNKNOWN"

/** 公开状态页的单个组件（按目标组聚合，不含目标地址与标签）。 */
export interface PublicStatusComponent {
  groupUid: string
  name: string
  status: PublicComponentStatus | string
  /** 可用率字段是字符串，例如 "99.95"；没有数据时为空串。 */
  availability24h: string
  availability7d: string
  availability30d: string
  /** 最近一次检查时间，RFC3339；没有结果时为空。 */
  lastCheckedAt: string
}

export type PublicEventStatus = "FIRING" | "RESOLVED"

/** 公开状态页的事件摘要，不返回目标地址等内部信息。 */
export interface PublicStatusEvent {
  uid: string
  title: string
  status: PublicEventStatus | string
  startedAt: string
  endedAt: string
}

/** 公开状态页数据。 */
export interface PublicStatusPage {
  title: string
  description: string
  updatedAt: string
  components: PublicStatusComponent[]
  activeEvents: PublicStatusEvent[]
  recentEvents: PublicStatusEvent[]
}
