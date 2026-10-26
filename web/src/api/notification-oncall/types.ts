/** 值班时段的星期取值，与后端约定为大写英文。 */
export type OnCallWeekday =
  | "MONDAY"
  | "TUESDAY"
  | "WEDNESDAY"
  | "THURSDAY"
  | "FRIDAY"
  | "SATURDAY"
  | "SUNDAY"

/** 单个接收时段；start 晚于 end 表示跨午夜，start 等于 end 不合法。 */
export interface OnCallWindow {
  weekdays: string[]
  start: string
  end: string
}

/** 详情里的通知组绑定：groupName 只读，用于列表和详情展示。 */
export interface OnCallGroupItem {
  groupUid: string
  groupName: string
  windows: OnCallWindow[]
}

/** 创建/更新只提交 groupUid 与时段。 */
export interface OnCallGroupInput {
  groupUid: string
  windows: OnCallWindow[]
}

export interface NotificationOnCallItem {
  uid: string
  name: string
  status: string
  labels: string
  levelUids: string[]
  datasourceUids: string[]
  /** 数据源级别标识，匹配事件的 datasource_level 标签；为空表示不限制。 */
  datasourceLevelUids: string[]
  timezone: string
  groups: OnCallGroupItem[]
  /** 同一告警（相同指纹）重复通知的最小间隔秒数；0 表示不限制，最大 43200。 */
  repeatIntervalSeconds: number
  remark: string
  createdAt: string
}

export interface ListNotificationOnCallsParams {
  page?: number
  pageSize?: number
  name?: string
}

export interface ListNotificationOnCallsResult {
  items: NotificationOnCallItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateNotificationOnCallParams {
  name: string
  labels: string
  levelUids?: string[]
  datasourceUids?: string[]
  datasourceLevelUids?: string[]
  timezone: string
  groups: OnCallGroupInput[]
  repeatIntervalSeconds?: number
  remark?: string
}

export interface UpdateNotificationOnCallParams extends CreateNotificationOnCallParams {
  uid: string
}
