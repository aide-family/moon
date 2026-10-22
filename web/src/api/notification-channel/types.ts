export interface NotificationChannelItem {
  uid: string
  name: string
  type: string
  app: string
  status: string
  endpoint: string
  username: string
  fromAddress: string
  remark: string
  hasPassword: boolean
  hasSecret: boolean
}

export interface ListNotificationChannelsParams {
  page?: number
  pageSize?: number
  name?: string
  type?: string
  app?: string
}

export interface ListNotificationChannelsResult {
  items: NotificationChannelItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateNotificationChannelParams {
  name: string
  type: string
  app: string
  endpoint: string
  username?: string
  password?: string
  secret?: string
  fromAddress?: string
  remark?: string
}

export interface UpdateNotificationChannelParams extends CreateNotificationChannelParams {
  uid: string
}
