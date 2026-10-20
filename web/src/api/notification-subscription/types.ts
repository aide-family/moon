export interface NotificationSubscriptionItem {
  uid: string
  name: string
  status: string
  labels: string
  levelUids: string[]
  datasourceUids: string[]
  groupUids: string[]
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
  groupUids: string[]
  remark?: string
}

export interface UpdateNotificationSubscriptionParams extends CreateNotificationSubscriptionParams {
  uid: string
}
