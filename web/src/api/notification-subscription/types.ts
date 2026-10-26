export interface NotificationSubscriptionItem {
  uid: string
  name: string
  status: string
  labels: string
  levelUids: string[]
  datasourceUids: string[]
  /** 数据源级别标识，匹配事件的 datasource_level 标签；为空表示不限制。 */
  datasourceLevelUids: string[]
  groupUids: string[]
  /** 同一告警（相同指纹）重复通知的最小间隔秒数；0 表示不限制，最大 43200。 */
  repeatIntervalSeconds: number
  remark: string
  createdAt: string
}

export interface ListNotificationSubscriptionsParams {
  page?: number
  pageSize?: number
  name?: string
}

export interface ListNotificationSubscriptionsResult {
  items: NotificationSubscriptionItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateNotificationSubscriptionParams {
  name: string
  labels: string
  levelUids?: string[]
  datasourceUids?: string[]
  datasourceLevelUids?: string[]
  groupUids: string[]
  repeatIntervalSeconds?: number
  remark?: string
}

export interface UpdateNotificationSubscriptionParams extends CreateNotificationSubscriptionParams {
  uid: string
}
