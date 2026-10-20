export interface NotificationGroupItem {
  uid: string
  name: string
  status: string
  channelUid: string
  templateUid: string
  receivers: string
  remark: string
  memberUids: string[]
  webhookChannelUids: string[]
  webhookTemplateUid: string
}

export interface ListNotificationGroupsParams {
  page?: number
  pageSize?: number
  name?: string
}

export interface ListNotificationGroupsResult {
  items: NotificationGroupItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateNotificationGroupParams {
  name: string
  channelUid: string
  templateUid: string
  receivers?: string
  remark?: string
  memberUids?: string[]
  webhookChannelUids?: string[]
  webhookTemplateUid?: string
}

export interface UpdateNotificationGroupParams extends CreateNotificationGroupParams {
  uid: string
}

export interface SelectItem {
  uid: string
  name: string
}
